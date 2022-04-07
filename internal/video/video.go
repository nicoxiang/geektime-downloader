package video

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/nicoxiang/geektime-downloader/internal/client"
	cfile "github.com/nicoxiang/geektime-downloader/internal/pkg/file"
)

const syncByte = uint8(71) //0x47

// DownloadVideo ...
func DownloadVideo(ctx context.Context, m3u8url, title, downloadProjectFolder string, size int64, concurrency int) {
	i := strings.LastIndex(m3u8url, "/")
	tsURLPrefix := m3u8url[:i+1]
	filenamifyTitle := cfile.Filenamify(title)

	// Stage1: Make m3u8 URL call and resolve
	decryptkmsURL, tsFileNames, err := readM3U8File(ctx, m3u8url)
	if err != nil && errors.Is(err, context.Canceled) {
		return
	}

	// Stage2: Get decrypt key
	key, err := getDecryptKey(ctx, decryptkmsURL)
	if err != nil && errors.Is(err, context.Canceled) {
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

	c := len(tsFileNames)
	sem := make(chan bool, concurrency)
	count := make(chan struct{}, c)

	bar := newBar(size, fmt.Sprintf("[正在下载 %s] ", title))
	bar.Start()
	for _, tsFileName := range tsFileNames {
		go func(tsFileName string) {
			writeToTempVideoFile(ctx, sem, count, bar, tsURLPrefix, tempVideoDir, tsFileName)
		}(tsFileName)
	}

	for c > 0 {
		<-count
		c--
	}
	bar.Finish()

	// Stage4: Read temp ts files, decrypt and merge into the one final video file
	mergeTSFiles(ctx, tempVideoDir, filenamifyTitle, downloadProjectFolder, key)
}

func writeToTempVideoFile(ctx context.Context, sem chan bool, count chan struct{}, bar *pb.ProgressBar, tsURLPrefix, tempVideoDir, tsFileName string) {
	defer func() {
		<-sem
		count <- struct{}{}
	}()
	sem <- true
	{
		if cancelled(ctx) {
			return
		}
		tsURL := tsURLPrefix + tsFileName
		resp, err := client.NewNoParseResponseRestyClient().R().SetContext(ctx).Get(tsURL)
		if err != nil && !errors.Is(err, context.Canceled) {
			panic(err)
		}
		if resp.RawResponse != nil && resp.RawResponse.StatusCode == 200 && resp.RawResponse.ContentLength > 0 {
			t := filepath.Join(tempVideoDir, tsFileName)
			f, err := os.OpenFile(t, os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				panic(err)
			}
			written, err := io.Copy(f, resp.RawBody())
			if err != nil && !errors.Is(err, context.Canceled) {
				panic(err)
			}
			addBar(bar, written)
			if err := f.Close(); err != nil {
				panic(err)
			}
		}
	}
}

func readM3U8File(ctx context.Context, url string) (decryptkmsURL string, tsFileNames []string, err error) {
	if cancelled(ctx) {
		return "", nil, context.Canceled
	}
	resp, err := client.New().R().SetContext(ctx).Get(url)
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			panic(err)
		}
		return "", nil, context.Canceled
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
		f, err := os.ReadFile(filepath.Join(tempVideoDir, tempTSFile.Name()))
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

func getDecryptKey(ctx context.Context, decryptkmsURL string) ([]byte, error) {
	if cancelled(ctx) {
		return nil, context.Canceled
	}
	keyResp, err := client.New().R().SetContext(ctx).Get(decryptkmsURL)
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			panic(err)
		}
		return nil, context.Canceled
	}
	return keyResp.Body(), nil
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
func addBar(bar *pb.ProgressBar, written int64) {
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
