package video

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/go-resty/resty/v2"
	mclient "github.com/nicoxiang/geektime-downloader/internal/client"
	cfile "github.com/nicoxiang/geektime-downloader/internal/pkg/file"
)

const syncByte = uint8(71) //0x47

// DownloadVideo ...
func DownloadVideo(ctx context.Context, m3u8url, title, downloadProjectFolder string, size int64, concurrency int) {
	i := strings.LastIndex(m3u8url, "/")
	tsURLPrefix := m3u8url[:i+1]
	filenamifyTitle := cfile.Filenamify(title)

	// Stage1: Make m3u8 URL call and resolve
	decryptkmsURL, tsFileNames := readM3U8File(ctx, m3u8url)
	if decryptkmsURL == "" || len(tsFileNames) == 0 {
		return
	}

	// Stage2: Get decrypt key
	key := getDecryptKey(ctx, decryptkmsURL)
	if key == nil {
		return
	}

	// Stage3: Make temp ts folder and download temp ts files
	tempVideoDir := filepath.Join(downloadProjectFolder, filenamifyTitle)
	if err := os.MkdirAll(tempVideoDir, os.ModePerm); err != nil {
		panic(err)
	}
	defer func() {
		if err := os.RemoveAll(tempVideoDir); err != nil {
			panic(err)
		}
	}()

	// classic bounded work pooling pattern
	var wg sync.WaitGroup
	wg.Add(concurrency)
	ch := make(chan string, concurrency)

	bar := newBar(size, fmt.Sprintf("[正在下载 %s] ", title))
	bar.Start()

	for i := 0; i < concurrency; i++ {
		go func() {
			writeToTempVideoFile(ctx, ch, bar, &wg, tsURLPrefix, tempVideoDir)
		}()
	}

	for _, tsFileName := range tsFileNames {
		ch <- tsFileName
	}
	close(ch)
	wg.Wait()
	bar.Finish()

	// Stage4: Read temp ts files, decrypt and merge into the one final video file
	mergeTSFiles(ctx, tempVideoDir, filenamifyTitle, downloadProjectFolder, key)
}

func writeToTempVideoFile(ctx context.Context, ch chan string, bar *pb.ProgressBar, wg *sync.WaitGroup, tsURLPrefix, tempVideoDir string) {
	defer wg.Done()

	for tsFileName := range ch {
		c := resty.New()
		c.SetOutputDirectory(tempVideoDir)
		mclient.SetGeekBangHeaders(c)

		resp, err := c.R().
			SetContext(ctx).
			SetOutput(tsFileName).
			Get(tsURLPrefix + tsFileName)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				continue
			}
			panic(err)
		}
		addBarValue(bar, resp.Size())
	}
}

func readM3U8File(ctx context.Context, url string) (decryptkmsURL string, tsFileNames []string) {
	resp, err := mclient.New().R().SetContext(ctx).Get(url)
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			panic(err)
		}
		return "", nil
	}
	data := resp.Body()
	s := string(data)
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "#EXT-X-KEY") {
			i := strings.LastIndex(line, "URI=")
			decryptkmsURL = line[i+5 : len(line)-1]
		}
		if !strings.HasPrefix(line, "#") && strings.HasSuffix(line, ".ts") {
			tsFileNames = append(tsFileNames, line)
		}
	}
	return
}

func mergeTSFiles(ctx context.Context, tempVideoDir, filenamifyTitle, downloadProjectFolder string, key []byte) {
	if cancelled(ctx) {
		return
	}

	tempTSFiles, err := ioutil.ReadDir(tempVideoDir)
	if err != nil {
		panic(err)
	}
	sort.Sort(cfile.ByNumericalFilename(tempTSFiles))
	fullPath := filepath.Join(downloadProjectFolder, filenamifyTitle+".ts")
	finalVideoFile, err := os.OpenFile(fullPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	for _, tempTSFile := range tempTSFiles {
		f, err := ioutil.ReadFile(filepath.Join(tempVideoDir, tempTSFile.Name()))
		if err != nil {
			panic(err)
		}
		aes128 := aesDecryptCBC(f, key, make([]byte, 16))
		for j := 0; j < len(aes128); j++ {
			if aes128[j] == syncByte {
				aes128 = aes128[j:]
				break
			}
		}
		for j := 0; j < len(aes128); j++ {
			if aes128[j] == syncByte {
				aes128 = aes128[j:]
				break
			}
		}
		if _, err := finalVideoFile.Write(aes128); err != nil {
			panic(err)
		}
	}
}

func getDecryptKey(ctx context.Context, decryptkmsURL string) []byte {
	if cancelled(ctx) {
		return nil
	}
	keyResp, err := mclient.New().R().SetContext(ctx).Get(decryptkmsURL)
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			panic(err)
		}
		return nil
	}
	return keyResp.Body()
}

func aesDecryptCBC(encrypted, key, iv []byte) (decrypted []byte) {
	block, _ := aes.NewCipher(key)
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, iv[:blockSize])
	decrypted = make([]byte, len(encrypted))
	blockMode.CryptBlocks(decrypted, encrypted)
	decrypted = pkcs5UnPadding(decrypted)
	return decrypted
}

func pkcs5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func newBar(size int64, prefix string) *pb.ProgressBar {
	bar := pb.New64(size)
	bar.SetRefreshRate(time.Second)
	bar.Set(pb.Bytes, true)
	bar.Set(pb.SIBytesPrefix, true)
	bar.SetTemplate(pb.Simple)
	bar.Set("prefix", prefix)
	return bar
}

// total bytes may greater than expected
func addBarValue(bar *pb.ProgressBar, written int64) {
	if bar.Current()+written > bar.Total() {
		bar.SetCurrent(bar.Total())
	} else {
		bar.Add64(written)
	}
}

func cancelled(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
