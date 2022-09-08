package video

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/nicoxiang/geektime-downloader/internal/geektime"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/crypto"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/filenamify"
	pgt "github.com/nicoxiang/geektime-downloader/internal/pkg/geektime"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/logger"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/m3u8"
	"github.com/nicoxiang/geektime-downloader/internal/video/vod"
	"golang.org/x/sync/errgroup"
)

const (
	syncByte = uint8(71) //0x47
	// TSExtension ...
	TSExtension = ".ts"
)

// EncryptType enum
type EncryptType int

const (
	// AliyunVodEncrypt ...
	AliyunVodEncrypt EncryptType = iota
	// HLSStandardEncrypt ...
	HLSStandardEncrypt
)

var (
	// make simple api call
	client *resty.Client
	// used to download video
	downloadClient *resty.Client
)

func init() {
	client = resty.New().
		SetHeader(pgt.UserAgentHeaderName, pgt.UserAgentHeaderValue).
		SetRetryCount(1).
		SetTimeout(10 * time.Second).
		SetLogger(logger.DiscardLogger{})

	downloadClient = resty.New().
		SetHeader(pgt.UserAgentHeaderName, pgt.UserAgentHeaderValue).
		SetRetryCount(0).
		SetTimeout(time.Minute).
		SetLogger(logger.DiscardLogger{})
}

// GetPlayInfoResponse is the response struct for api GetPlayInfo
type GetPlayInfoResponse struct {
	RequestID    string                        `json:"RequestId" xml:"RequestId"`
	VideoBase    vod.VideoBase                 `json:"VideoBase" xml:"VideoBase"`
	PlayInfoList vod.PlayInfoListInGetPlayInfo `json:"PlayInfoList" xml:"PlayInfoList"`
}

// DownloadAliyunVodEncryptVideo ...
func DownloadAliyunVodEncryptVideo(ctx context.Context,
	articleID int,
	currentProduct geektime.Product,
	projectDir string,
	quality string,
	concurrency int) error {
	playAuthInfo, err := geektime.GetPlayAuth(articleID, currentProduct.ID)
	if err != nil {
		return err
	}
	clientRand := uuid.NewString()
	playInfoURL, err := vod.BuildVodGetPlayInfoURL(playAuthInfo.PlayAuth, playAuthInfo.VideoID, clientRand)
	if err != nil {
		return err
	}
	playInfo, err := getPlayInfo(playInfoURL, quality)
	if err != nil {
		return err
	}
	tsURLPrefix := extractTSURLPrefix(playInfo.PlayURL)
	// just ignore keyURI in m3u8, aliyun private vod use another decrypt method
	tsFileNames, _, err := m3u8.Parse(playInfo.PlayURL)
	if err != nil {
		return err
	}
	decryptKey := crypto.GetAESDecryptKey(clientRand, playInfo.Rand, playInfo.Plaintext)
	title := getUniversityVideoTitle(articleID, currentProduct)
	return download(ctx, tsURLPrefix, title, projectDir, tsFileNames, []byte(decryptKey), playInfo.Size, AliyunVodEncrypt, concurrency)
}

// DownloadHLSStandardEncryptVideo ...
func DownloadHLSStandardEncryptVideo(ctx context.Context, m3u8url, title, projectDir string, size int64, concurrency int) (err error) {
	tsURLPrefix := extractTSURLPrefix(m3u8url)
	tsFileNames, keyURI, err := m3u8.Parse(m3u8url)
	if err != nil {
		return err
	}
	var decryptKey []byte
	// Old version keyURI
	// https://misc.geekbang.org/serv/v1/decrypt/decryptkms/?Ciphertext=longlongstring
	if strings.HasPrefix(keyURI, "https://") || strings.HasPrefix(keyURI, "http://") {
		resp, err := client.R().
			SetContext(ctx).
			SetHeader(pgt.OriginHeaderName, pgt.GeekBang).
			Get(keyURI)
		decryptKey = resp.Body()
		if err != nil {
			return err
		}
	} else {
		return errors.New("unexpected m3u8 keyURI")
	}

	return download(ctx, tsURLPrefix, title, projectDir, tsFileNames, decryptKey, size, HLSStandardEncrypt, concurrency)
}

func download(ctx context.Context,
	tsURLPrefix,
	title,
	projectDir string,
	tsFileNames []string,
	decryptKey []byte,
	size int64,
	videoEncryptType EncryptType,
	concurrency int) (err error) {

	// Make temp ts folder and download temp ts files
	filenamifyTitle := filenamify.Filenamify(title)
	tempVideoDir := filepath.Join(projectDir, filenamifyTitle)
	if err = os.MkdirAll(tempVideoDir, os.ModePerm); err != nil {
		return
	}
	// temp folder cleanup
	defer func() {
		_ = os.RemoveAll(tempVideoDir)
	}()

	// classic bounded work pooling pattern
	g, ctx := errgroup.WithContext(ctx)
	ch := make(chan string, concurrency)

	bar := newBar(size, fmt.Sprintf("[正在下载 %s] ", filenamifyTitle))
	bar.Start()

	downloadClient.SetOutputDirectory(tempVideoDir)

	for i := 0; i < concurrency; i++ {
		g.Go(func() error {
			return writeToTempVideoFile(ctx, ch, bar, tsURLPrefix)
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

	// Read temp ts files, decrypt and merge into the one final video file
	err = mergeTSFiles(tempVideoDir, filenamifyTitle, projectDir, decryptKey, videoEncryptType)

	return
}

func writeToTempVideoFile(ctx context.Context,
	tsFileNames chan string,
	bar *pb.ProgressBar,
	tsURLPrefix string) (err error) {
	for tsFileName := range tsFileNames {
		resp, err := downloadClient.R().
			SetContext(ctx).
			SetHeader(pgt.OriginHeaderName, pgt.GeekBang).
			SetOutput(tsFileName).
			Get(tsURLPrefix + tsFileName)
		if err != nil {
			for range tsFileNames {
			}
			return err
		}
		addBarValue(bar, resp.Size())
	}
	return nil
}

func mergeTSFiles(tempVideoDir, filenamifyTitle, projectDir string, key []byte, videoEncryptType EncryptType) error {
	tempTSFiles, err := ioutil.ReadDir(tempVideoDir)
	if err != nil {
		return err
	}
	fullPath := filepath.Join(projectDir, filenamifyTitle+TSExtension)
	finalVideoFile, err := os.OpenFile(fullPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
	defer func() {
		_ = finalVideoFile.Close()
	}()
	if err != nil {
		return err
	}
	for _, tempTSFile := range tempTSFiles {
		f, err := ioutil.ReadFile(filepath.Join(tempVideoDir, tempTSFile.Name()))
		if err != nil {
			return err
		}
		switch videoEncryptType {
		case HLSStandardEncrypt:
			aes128 := crypto.AESDecryptCBC(f, key, make([]byte, 16))
			// https://en.wikipedia.org/wiki/MPEG_transport_stream
			for j := 0; j < len(aes128); j++ {
				if aes128[j] == syncByte {
					aes128 = aes128[j:]
					break
				}
			}
			f = aes128
		case AliyunVodEncrypt:
			tsParser := m3u8.NewTSParser(f, string(key))
			f = tsParser.Decrypt()
		}
		if _, err := finalVideoFile.Write(f); err != nil {
			return err
		}
	}
	return nil
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

func getUniversityVideoTitle(articleID int, currentProduct geektime.Product) string {
	for _, v := range currentProduct.Articles {
		if v.AID == articleID {
			return v.Title
		}
	}
	return ""
}

func extractTSURLPrefix(m3u8url string) string {
	i := strings.LastIndex(m3u8url, "/")
	return m3u8url[:i+1]
}

func getPlayInfo(playInfoURL, quality string) (vod.PlayInfo, error) {
	var getPlayInfoResp GetPlayInfoResponse
	var playInfo vod.PlayInfo
	_, err := client.R().
		SetHeader(pgt.OriginHeaderName, pgt.GeekBangUniversity).
		SetResult(&getPlayInfoResp).
		Get(playInfoURL)

	if err != nil {
		return playInfo, err
	}

	playInfoList := getPlayInfoResp.PlayInfoList.PlayInfo
	for _, p := range playInfoList {
		if strings.EqualFold(p.Definition, quality) {
			playInfo = p
		}
	}
	return playInfo, nil
}
