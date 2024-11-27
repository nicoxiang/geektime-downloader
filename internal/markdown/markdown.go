package markdown

import (
	"context"
	"errors"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/nicoxiang/geektime-downloader/internal/geektime"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/downloader"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/filenamify"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/files"
)

var (
	converter *md.Converter
	imgRegexp = regexp.MustCompile(`!\[(.*?)]\((.*?)\)`)
)

// MDExtension ...
const MDExtension = ".md"

type markdownString struct {
	sync.Mutex
	s string
}

func (ms *markdownString) ReplaceAll(o, n string) {
	ms.Lock()
	defer ms.Unlock()
	ms.s = strings.ReplaceAll(ms.s, o, n)
}

// Download article as markdown
func Download(ctx context.Context, html, title, dir string, aid int, overwrite bool) (bool, error) {
	select {
	case <-ctx.Done():
		return false, context.Canceled
	default:
	}

	fullName := path.Join(dir, filenamify.Filenamify(title)+MDExtension)
	if files.CheckFileExists(fullName) && !overwrite {
		return true, nil
	}

	// step1: convert to md string
	markdown, err := getDefaultConverter().ConvertString(html)
	if err != nil {
		return false, err
	}
	// step2: download images
	var ss = &markdownString{s: markdown}
	imageURLs := findAllImages(markdown)

	// images/aid/imageName.png
	imagesFolder := filepath.Join(dir, "images", strconv.Itoa(aid))

	if _, err := os.Stat(imagesFolder); errors.Is(err, os.ErrNotExist) {
		os.MkdirAll(imagesFolder, os.ModePerm)
	}

	err = writeImageFile(ctx, imageURLs, dir, imagesFolder, ss)

	if err != nil {
		return false, err
	}

	f, err := os.Create(fullName)
	defer func() {
		_ = f.Close()
	}()
	if err != nil {
		return false, err
	}
	// step3: write md file
	_, err = f.WriteString("# " + title + "\n" + ss.s)
	if err != nil {
		return false, err
	}
	return false, nil
}

func findAllImages(md string) (images []string) {
	for _, matches := range imgRegexp.FindAllStringSubmatch(md, -1) {
		if len(matches) == 3 {
			s := matches[2]
			isImg, err := isImageURL(s)
			if err == nil && isImg {
				images = append(images, s)
			}
			// sometime exists broken image url, just ignore
		}
	}
	return
}

func getDefaultConverter() *md.Converter {
	if converter == nil {
		converter = md.NewConverter("", true, nil)
	}
	return converter
}

func writeImageFile(ctx context.Context,
	imageURLs []string,
	dir,
	imagesFolder string,
	ms *markdownString,
) (err error) {
	for _, imageURL := range imageURLs {
		segments := strings.Split(imageURL, "/")
		f := segments[len(segments)-1]
		if i := strings.Index(f, "?"); i > 0 {
			f = f[:i]
		}
		imageLocalFullPath := filepath.Join(imagesFolder, f)

		headers := make(map[string]string, 2)
		headers[geektime.Origin] = geektime.DefaultBaseURL
		headers[geektime.UserAgent] = geektime.DefaultUserAgent

		_, err := downloader.DownloadFileConcurrently(ctx, imageLocalFullPath, imageURL, headers, 1)

		if err != nil {
			return err
		}

		rel, _ := filepath.Rel(dir, imageLocalFullPath)
		ms.ReplaceAll(imageURL, filepath.ToSlash(rel))
	}
	return nil
}

func isImageURL(urlStr string) (bool, error) {
	// 解析 URL
	parsedURL, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return false, err
	}
	// 提取路径部分
	filePath := parsedURL.Path

	// 获取文件扩展名并转换为小写
	ext := strings.ToLower(path.Ext(filePath))

	// 检查扩展名是否属于图片类型
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".tiff":
		return true, nil
	default:
		return false, nil
	}
}
