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
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/go-resty/resty/v2"
	iclient "github.com/nicoxiang/geektime-downloader/internal/client"
	ifile "github.com/nicoxiang/geektime-downloader/internal/pkg/file"
	"golang.org/x/sync/errgroup"
)

const (
	syncByte = uint8(71) //0x47
	tsExt    = ".ts"
)

// ErrUnexpectedM3U8Format ...
var ErrUnexpectedM3U8Format = errors.New("unexpected m3u8 response format")

// ErrUnexpectedDecryptKeyResponse ...
var ErrUnexpectedDecryptKeyResponse = errors.New("unexpected decrypt key response")

// DownloadVideo ...
func DownloadVideo(ctx context.Context, m3u8url, title, downloadProjectFolder string, size int64, concurrency int) (err error) {
	i := strings.LastIndex(m3u8url, "/")
	tsURLPrefix := m3u8url[:i+1]
	filenamifyTitle := ifile.Filenamify(title)

	// Stage1: Make m3u8 URL call and resolve
	decryptkmsURL, tsFileNames, err := readM3U8File(ctx, m3u8url)
	if err != nil {
		return
	}
	if decryptkmsURL == "" || len(tsFileNames) == 0 {
		return ErrUnexpectedM3U8Format
	}

	// Stage2: Get decrypt key
	key, err := getDecryptKey(ctx, decryptkmsURL)
	if err != nil {
		return
	}
	if key == nil {
		return ErrUnexpectedDecryptKeyResponse
	}

	// Stage3: Make temp ts folder and download temp ts files
	tempVideoDir := filepath.Join(downloadProjectFolder, filenamifyTitle)
	if err = os.MkdirAll(tempVideoDir, os.ModePerm); err != nil {
		return
	}
	// temp folder cleanup
	defer func() {
		os.RemoveAll(tempVideoDir)
	}()

	// classic bounded work pooling pattern
	g := new(errgroup.Group)
	ch := make(chan string, concurrency)

	bar := newBar(size, fmt.Sprintf("[正在下载 %s] ", title))
	bar.Start()

	for i := 0; i < concurrency; i++ {
		g.Go(func() error {
			return writeToTempVideoFile(ctx, ch, bar, tsURLPrefix, tempVideoDir)
		})
	}

	for _, tsFileName := range tsFileNames {
		ch <- tsFileName
	}
	close(ch)
	err = g.Wait()
	bar.Finish()
	if err != nil {
		return
	}

	// Stage4: Read temp ts files, decrypt and merge into the one final video file
	err = mergeTSFiles(tempVideoDir, filenamifyTitle, downloadProjectFolder, key)

	return
}

func writeToTempVideoFile(ctx context.Context, tsFileNames chan string, bar *pb.ProgressBar, tsURLPrefix, tempVideoDir string) (err error) {
	var es []error
loop:
	for {
		select {
		case <-ctx.Done():
			// Drain tsFileNames to allow existing goroutines to finish.
			for range tsFileNames {
			}
		case tsFileName, ok := <-tsFileNames:
			if !ok {
				break loop
			}
			c := resty.New()
			c.SetOutputDirectory(tempVideoDir).
				SetTimeout(time.Minute)

			iclient.SetGeekBangHeaders(c)

			resp, err := c.R().
				SetContext(ctx).
				SetOutput(tsFileName).
				Get(tsURLPrefix + tsFileName)
			if err != nil {
				es = append(es, err)
				continue
			}
			addBarValue(bar, resp.Size())
		}
	}
	if len(es) > 0{
		return es[0]
	} 
	return nil
}

func readM3U8File(ctx context.Context, url string) (decryptkmsURL string, tsFileNames []string, err error) {
	resp, err := iclient.New().R().SetContext(ctx).Get(url)
	if err != nil {
		return
	}
	s := string(resp.Body())
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "#EXT-X-KEY") {
			i := strings.LastIndex(line, "URI=")
			decryptkmsURL = line[i+5 : len(line)-1]
		}
		if !strings.HasPrefix(line, "#") && strings.HasSuffix(line, tsExt) {
			tsFileNames = append(tsFileNames, line)
		}
	}
	return
}

func mergeTSFiles(tempVideoDir, filenamifyTitle, downloadProjectFolder string, key []byte) error {
	tempTSFiles, err := ioutil.ReadDir(tempVideoDir)
	if err != nil {
		return err
	}
	sort.Sort(ifile.ByNumericalFilename(tempTSFiles))
	fullPath := filepath.Join(downloadProjectFolder, filenamifyTitle+tsExt)
	finalVideoFile, err := os.OpenFile(fullPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	for _, tempTSFile := range tempTSFiles {
		f, err := ioutil.ReadFile(filepath.Join(tempVideoDir, tempTSFile.Name()))
		if err != nil {
			return err
		}
		aes128 := aesDecryptCBC(f, key, make([]byte, 16))
		// https://en.wikipedia.org/wiki/MPEG_transport_stream
		for j := 0; j < len(aes128); j++ {
			if aes128[j] == syncByte {
				aes128 = aes128[j:]
				break
			}
		}
		if _, err := finalVideoFile.Write(aes128); err != nil {
			return err
		}
	}
	return nil
}

func getDecryptKey(ctx context.Context, decryptkmsURL string) (key []byte, err error) {
	keyResp, err := iclient.New().R().SetContext(ctx).Get(decryptkmsURL)
	if err != nil {
		return
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
func addBarValue(bar *pb.ProgressBar, written int64) {
	if bar.Current()+written > bar.Total() {
		bar.SetCurrent(bar.Total())
	} else {
		bar.Add64(written)
	}
}
