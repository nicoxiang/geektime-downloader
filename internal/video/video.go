package video

import (
	"context"
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

type VideoEncryptType int

const (
	AliyunVodEncrypt VideoEncryptType = iota
	HLSStandardEncrypt
)

var (
	client *resty.Client
)

func init() {
	client = resty.New().
		SetHeader(pgt.UserAgentHeaderName, pgt.UserAgentHeaderValue).
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
	// TODO: error
	playAuthInfo, _ := geektime.GetPlayAuth(articleID, currentProduct.ID)
	clientRand := uuid.NewString()
	playInfoURL, _ := vod.BuildVodGetPlayInfoURL(playAuthInfo.PlayAuth, playAuthInfo.VideoID, clientRand)
	playInfo := getPlayInfo(playInfoURL, quality)
	tsURLPrefix := extractTSURLPrefix(playInfo.PlayURL)
	tsFileNames, _, _ := m3u8.Parse(playInfo.PlayURL)
	decryptKey := crypto.GetAESDecryptKey(clientRand, playInfo.Rand, playInfo.Plaintext)
	title := getUniversityVideoTitle(articleID, currentProduct)
	return download(ctx, tsURLPrefix, title, projectDir, tsFileNames, []byte(decryptKey), playInfo.Size, AliyunVodEncrypt, concurrency)
}

// DownloadHLSStandardEncryptVideo ...
func DownloadHLSStandardEncryptVideo(ctx context.Context, m3u8url, title, projectDir string, size int64, concurrency int) (err error) {
	tsURLPrefix := extractTSURLPrefix(m3u8url)
	tsFileNames, keyURI, err := m3u8.Parse(m3u8url)
	var decryptKey []byte
	if strings.HasPrefix(keyURI, "https://") || strings.HasPrefix(keyURI, "http://") {
		// TODO: err
		resp, _ := client.R().
			SetContext(ctx).
			SetHeader(pgt.OriginHeaderName, pgt.GeekBang).
			Get(keyURI)
		decryptKey = resp.Body()
	}
	if err != nil {
		return err
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
	videoEncryptType VideoEncryptType,
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

	setVideoDownloadClientOptions(tempVideoDir)

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
		resp, err := client.R().
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

func mergeTSFiles(tempVideoDir, filenamifyTitle, projectDir string, key []byte, videoEncryptType VideoEncryptType) error {
	tempTSFiles, err := ioutil.ReadDir(tempVideoDir)
	if err != nil {
		return err
	}
	// sort.Sort(ByNumericalFilename(tempTSFiles))
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
		// TODO: simplify
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
			if _, err := finalVideoFile.Write(aes128); err != nil {
				return err
			}
		case AliyunVodEncrypt:
			tsParser := m3u8.NewTSParser(f, string(key))
			f = tsParser.Decrypt()

			if _, err := finalVideoFile.Write(f); err != nil {
				return err
			}
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

// func resetRestyOptions() {
// 	geekbangClient.SetTimeout(10 * time.Second).
// 		SetRetryCount(1).
// 		SetOutputDirectory("")
// }

func setVideoDownloadClientOptions(tempVideoDir string) {
	client.SetTimeout(time.Minute).
		SetRetryCount(0).
		SetOutputDirectory(tempVideoDir)
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

func getPlayInfo(playInfoURL, quality string) vod.PlayInfo {
	var getPlayInfoResp GetPlayInfoResponse

	// err
	client.R().
		SetHeader(pgt.OriginHeaderName, pgt.GeekBangUniversity).
		SetResult(&getPlayInfoResp).
		Get(playInfoURL)

	playInfoList := getPlayInfoResp.PlayInfoList.PlayInfo
	var playInfo vod.PlayInfo
	for _, p := range playInfoList {
		if strings.EqualFold(p.Definition, quality) {
			playInfo = p
		}
	}
	return playInfo
}

// // ByNumericalFilename implement sort interface, order by file name suffix number
// type ByNumericalFilename []os.FileInfo

// func (nf ByNumericalFilename) Len() int      { return len(nf) }
// func (nf ByNumericalFilename) Swap(i, j int) { nf[i], nf[j] = nf[j], nf[i] }
// func (nf ByNumericalFilename) Less(i, j int) bool {
// 	// Use path names
// 	pathA := nf[i].Name()
// 	pathB := nf[j].Name()

// 	// Grab integer value of each filename by parsing the string and slicing off
// 	// the extension
// 	a, err1 := strconv.ParseInt(pathA[0:strings.LastIndex(pathA, ".")], 10, 64)
// 	b, err2 := strconv.ParseInt(pathB[0:strings.LastIndex(pathB, ".")], 10, 64)

// 	// If any were not numbers sort lexographically
// 	if err1 != nil || err2 != nil {
// 		return pathA < pathB
// 	}

// 	// Which integer is smaller?
// 	return a < b
// }
