package video

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/cavaliergopher/grab/v3"
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
	client *resty.Client
	grabClient *grab.Client
)

func init() {
	client = resty.New().
	SetHeader(pgt.UserAgentHeaderName, pgt.UserAgentHeaderValue).
	SetRetryCount(1).
	SetTimeout(10 * time.Second).
	SetLogger(logger.DiscardLogger{})

	grabClient = grab.NewClient()
	grabClient.UserAgent = pgt.UserAgentHeaderValue
}

// GetPlayInfoResponse is the response struct for api GetPlayInfo
type GetPlayInfoResponse struct {
	RequestID    string                        `json:"RequestId" xml:"RequestId"`
	VideoBase    vod.VideoBase                 `json:"VideoBase" xml:"VideoBase"`
	PlayInfoList vod.PlayInfoListInGetPlayInfo `json:"PlayInfoList" xml:"PlayInfoList"`
}

// DownloadArticleVideo download normal video cource ...
// sourceType: normal video cource 1
func DownloadArticleVideo(ctx context.Context,
	articleID int,
	sourceType int,
	projectDir string,
	quality string,
	concurrency int,
) error {

	articleInfo, err := geektime.PostV3ArticleInfo(articleID)
	if err != nil {
		return err
	}
	playAuth, err := geektime.PostV3VideoPlayAuth(articleInfo.Data.Info.ID, sourceType, articleInfo.Data.Info.Video.ID)
	if err != nil {
		return err
	}
	return downloadAliyunVodEncryptVideo(ctx,
		playAuth,
		articleInfo.Data.Info.Title,
		projectDir,
		quality,
		articleInfo.Data.Info.Video.ID,
		concurrency)
}

// DownloadUniversityVideo ...
func DownloadUniversityVideo(ctx context.Context,
	articleID int,
	currentProduct geektime.Product,
	projectDir string,
	quality string,
	concurrency int) error {

	playAuthInfo, err := geektime.PostUniversityV1VideoPlayAuth(articleID, currentProduct.ID)
	if err != nil {
		return err
	}

	videoTitle := getUniversityVideoTitle(articleID, currentProduct)
	return downloadAliyunVodEncryptVideo(ctx,
		playAuthInfo.Data.PlayAuth,
		videoTitle,
		projectDir,
		quality,
		playAuthInfo.Data.VID,
		concurrency)
}

func downloadAliyunVodEncryptVideo(ctx context.Context,
	playAuth,
	videoTitle,
	projectDir,
	quality,
	videoID string,
	concurrency int) error {

	clientRand := uuid.NewString()
	playInfoURL, err := vod.BuildVodGetPlayInfoURL(playAuth, videoID, clientRand)
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
	return download(ctx, tsURLPrefix, videoTitle, projectDir, tsFileNames, []byte(decryptKey), playInfo.Size, AliyunVodEncrypt, concurrency)
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

// DownloadMP4 ...
func DownloadMP4(ctx context.Context, title, projectDir string, mp4URLs []string) (err error) {
	filenamifyTitle := filenamify.Filenamify(title)
	videoDir := filepath.Join(projectDir, "videos", filenamifyTitle)
	if err = os.MkdirAll(videoDir, os.ModePerm); err != nil {
		return
	}

	reqs := make([]*grab.Request, len(mp4URLs))
	for _, mp4URL := range mp4URLs {
		u, _ := url.Parse(mp4URL)
		dst := filepath.Join(videoDir, path.Base(u.Path))
		request, err := grab.NewRequest(dst, mp4URL)
		if err != nil {
			//TODO: error handling
			panic(err)
		}
		request.HTTPRequest.Header.Set(pgt.OriginHeaderName, pgt.GeekBang)
		reqs = append(reqs, request)
	}
	
	respch := grabClient.DoBatch(0, reqs...)

	// check each response
	for resp := range respch {
		if err := resp.Err(); err != nil {
			return err
		}
	}
	return
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

	bar := newBar(size, fmt.Sprintf("[正在下载 %s] ", filenamifyTitle))
	bar.Start()

	reqs := make([]*grab.Request, len(tsFileNames))
	for _, tsFileName := range tsFileNames {
		u := tsURLPrefix + tsFileName
		dst := filepath.Join(tempVideoDir, tsFileName)
		request, err := grab.NewRequest(dst, u)
		if err != nil {
			//TODO: error handling
			panic(err)
		}
		request.HTTPRequest.Header.Set(pgt.OriginHeaderName, pgt.GeekBang)
		reqs = append(reqs, request)
	}
	
	respch := grabClient.DoBatch(concurrency, reqs...)

	// check each response
	for resp := range respch {
		if err := resp.Err(); err != nil {
			return err
		}
		addBarValue(bar, resp.BytesComplete())
	}
	
	bar.Finish()

	// Read temp ts files, decrypt and merge into the one final video file
	err = mergeTSFiles(tempVideoDir, filenamifyTitle, projectDir, decryptKey, videoEncryptType)

	return
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
