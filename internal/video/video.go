package video

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/google/uuid"
	"github.com/nicoxiang/geektime-downloader/internal/geektime"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/crypto"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/downloader"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/filenamify"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/files"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/m3u8"
	"github.com/nicoxiang/geektime-downloader/internal/video/vod"
)

const (
	// syncByte = uint8(71) //0x47

	// TSExtension ...
	TSExtension = ".ts"
)

// EncryptType enum
type EncryptType int

// const (
// 	// AliyunVodEncrypt ...
// 	AliyunVodEncrypt EncryptType = iota
// 	// HLSStandardEncrypt ...
// 	HLSStandardEncrypt
// )

// GetPlayInfoResponse is the response struct for api GetPlayInfo
type GetPlayInfoResponse struct {
	RequestID    string                        `json:"RequestId" xml:"RequestId"`
	VideoBase    vod.VideoBase                 `json:"VideoBase" xml:"VideoBase"`
	PlayInfoList vod.PlayInfoListInGetPlayInfo `json:"PlayInfoList" xml:"PlayInfoList"`
}

// DownloadArticleVideo download normal video cource ...
// sourceType: normal video cource 1
func DownloadArticleVideo(ctx context.Context,
	client *geektime.Client,
	articleID int,
	sourceType int,
	projectDir string,
	quality string,
	concurrency int,
) error {

	articleInfo, err := client.V3ArticleInfo(articleID)
	if err != nil {
		return err
	}
	if articleInfo.Data.Info.Video.ID == "" {
		return nil
	}
	playAuth, err := client.VideoPlayAuth(articleInfo.Data.Info.ID, sourceType, articleInfo.Data.Info.Video.ID)
	if err != nil {
		return err
	}
	return downloadAliyunVodEncryptVideo(ctx,
		client,
		playAuth,
		articleInfo.Data.Info.Title,
		projectDir,
		quality,
		articleInfo.Data.Info.Video.ID,
		concurrency)
}

// DownloadEnterpriseArticleVideo download enterprise video
func DownloadEnterpriseArticleVideo(ctx context.Context,
	client *geektime.Client,
	articleID int,
	projectDir string,
	quality string,
	concurrency int,
) error {
	articleInfo, err := client.V1EnterpriseArticleDetail(strconv.Itoa(articleID))
	if err != nil {
		return err
	}
	if articleInfo.Data.Video.ID == "" {
		return nil
	}
	playAuth, err := client.EnterpriseVideoPlayAuth(strconv.Itoa(articleID), articleInfo.Data.Video.ID)
	if err != nil {
		return err
	}
	return downloadAliyunVodEncryptVideo(ctx,
		client,
		playAuth,
		articleInfo.Data.Article.Title,
		projectDir,
		quality,
		articleInfo.Data.Video.ID,
		concurrency)
}

// DownloadUniversityVideo ...
func DownloadUniversityVideo(ctx context.Context,
	client *geektime.Client,
	articleID int,
	currentProduct geektime.Course,
	projectDir string,
	quality string,
	concurrency int) error {
	playAuthInfo, err := client.UniversityVideoPlayAuth(articleID, currentProduct.ID)
	if err != nil {
		return err
	}

	videoTitle := getUniversityVideoTitle(articleID, currentProduct)
	return downloadAliyunVodEncryptVideo(ctx,
		client,
		playAuthInfo.Data.PlayAuth,
		videoTitle,
		projectDir,
		quality,
		playAuthInfo.Data.VID,
		concurrency)
}

func downloadAliyunVodEncryptVideo(ctx context.Context,
	client *geektime.Client,
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
	playInfo, err := getPlayInfo(client, playInfoURL, quality)
	if err != nil {
		return err
	}
	tsURLPrefix := extractTSURLPrefix(playInfo.PlayURL)

	tsFileNames, isVodEncryptVideo, err := m3u8.Parse(client, playInfo.PlayURL)
	if err != nil {
		return err
	}

	decryptKey := ""
	if isVodEncryptVideo {
		decryptKey = crypto.GetAESDecryptKey(clientRand, playInfo.Rand, playInfo.Plaintext)
	}
	return download(ctx, tsURLPrefix, videoTitle, projectDir, tsFileNames, []byte(decryptKey), playInfo.Size, isVodEncryptVideo, concurrency)
}

// DownloadMP4 download MP4 resources in article
func DownloadMP4(ctx context.Context, title, projectDir string, mp4URLs []string, overwrite bool) (err error) {
	filenamifyTitle := filenamify.Filenamify(title)
	videoDir := filepath.Join(projectDir, "videos", filenamifyTitle)
	if err = os.MkdirAll(videoDir, os.ModePerm); err != nil {
		return
	}

	for _, mp4URL := range mp4URLs {
		u, _ := url.Parse(mp4URL)
		dst := filepath.Join(videoDir, path.Base(u.Path))

		if files.CheckFileExists(dst) && !overwrite {
			continue
		}

		headers := make(map[string]string, 2)
		headers[geektime.Origin] = geektime.DefaultBaseURL
		headers[geektime.UserAgent] = geektime.DefaultUserAgent

		_, err := downloader.DownloadFileConcurrently(ctx, dst, mp4URL, headers, 5)
		if err != nil {
			return nil
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
	isVodEncryptVideo bool,
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

	for _, tsFileName := range tsFileNames {
		u := tsURLPrefix + tsFileName
		dst := filepath.Join(tempVideoDir, tsFileName)

		headers := make(map[string]string, 2)
		headers[geektime.Origin] = geektime.DefaultBaseURL
		headers[geektime.UserAgent] = geektime.DefaultUserAgent

		fileSize, err := downloader.DownloadFileConcurrently(ctx, dst, u, headers, concurrency)
		if err != nil {
			return err
		}

		addBarValue(bar, fileSize)
	}

	bar.Finish()

	// Read temp ts files, decrypt and merge into the one final video file
	err = mergeTSFiles(tempVideoDir, filenamifyTitle, projectDir, decryptKey, isVodEncryptVideo)

	return
}

func mergeTSFiles(tempVideoDir, filenamifyTitle, projectDir string, key []byte, isVodEncryptVideo bool) error {
	tempTSFiles, err := os.ReadDir(tempVideoDir)
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
		f, err := os.ReadFile(filepath.Join(tempVideoDir, tempTSFile.Name()))
		if err != nil {
			return err
		}

		if isVodEncryptVideo {
			tsParser := m3u8.NewTSParser(f, string(key))
			f = tsParser.Decrypt()
		}

		// case HLSStandardEncrypt:
		// 	aes128 := crypto.AESDecryptCBC(f, key, make([]byte, 16))
		// 	// https://en.wikipedia.org/wiki/MPEG_transport_stream
		// 	for j := 0; j < len(aes128); j++ {
		// 	if aes128[j] == syncByte {
		// 		aes128 = aes128[j:]
		// 		break
		// 	}
		// }
		// f = aes128

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

func getUniversityVideoTitle(articleID int, currentProduct geektime.Course) string {
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

func getPlayInfo(client *geektime.Client, playInfoURL, quality string) (vod.PlayInfo, error) {
	var getPlayInfoResp GetPlayInfoResponse
	var playInfo vod.PlayInfo
	_, err := client.RestyClient.R().
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
