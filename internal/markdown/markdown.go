package markdown

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/cavaliergopher/grab/v3"
	"github.com/nicoxiang/geektime-downloader/internal/geektime"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/filenamify"
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

// Download ...
func Download(ctx context.Context, grabClient *grab.Client, html, title, dir string, aid, concurrency int) error {
	select {
	case <-ctx.Done():
		return context.Canceled
	default:
	}
	// step1: convert to md string
	markdown, err := getDefaultConverter().ConvertString(html)
	if err != nil {
		return err
	}
	// step2: download images
	var ss = &markdownString{s: markdown}
	imageURLs := findAllImages(markdown)

	// images/aid/imageName.png
	imagesFolder := filepath.Join(dir, "images", strconv.Itoa(aid))

	err = writeImageFile(ctx, grabClient, imageURLs, dir, imagesFolder, ss, concurrency)

	if err != nil {
		return err
	}

	fullName := path.Join(dir, filenamify.Filenamify(title)+MDExtension)
	f, err := os.Create(fullName)
	defer func() {
		_ = f.Close()
	}()
	if err != nil {
		return err
	}
	// step3: write md file
	_, err = f.WriteString("# " + title + "\n" + ss.s)
	if err != nil {
		return err
	}
	return nil
}

func findAllImages(md string) (images []string) {
	for _, matches := range imgRegexp.FindAllStringSubmatch(md, -1) {
		if len(matches) == 3 {
			images = append(images, matches[2])
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
	grabClient *grab.Client,
	imageURLs []string,
	dir,
	imagesFolder string,
	ms *markdownString,
	concurrency int,
) (err error) {
	reqs := make([]*grab.Request, len(imageURLs))
	for i, imageURL := range imageURLs {
		segments := strings.Split(imageURL, "/")
		f := segments[len(segments)-1]
		if i := strings.Index(f, "?"); i > 0 {
			f = f[:i]
		}
		imageLocalFullPath := filepath.Join(imagesFolder, f)
		request, _ := grab.NewRequest(imageLocalFullPath, imageURL)
		request.HTTPRequest.Header.Set(geektime.Origin, geektime.DefaultBaseURL)
		request = request.WithContext(ctx)
		reqs[i] = request
		rel, _ := filepath.Rel(dir, imageLocalFullPath)
		ms.ReplaceAll(imageURL, filepath.ToSlash(rel))
	}
	respch := grabClient.DoBatch(concurrency, reqs...)

	// check each response
	for resp := range respch {
		if err := resp.Err(); err != nil {
			return err
		}
	}
	return nil
}
