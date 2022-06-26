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
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/go-resty/resty/v2"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/filenamify"
	pgt "github.com/nicoxiang/geektime-downloader/internal/pkg/geektime"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/logger"
	"golang.org/x/sync/errgroup"
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
func Download(ctx context.Context, html, title, dir string, aid, concurrency int) error {
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

	c := resty.New()
	c.SetOutputDirectory(imagesFolder).
		SetRetryCount(1).
		SetTimeout(5*time.Second).
		SetHeader(pgt.UserAgentHeaderName, pgt.UserAgentHeaderValue).
		SetHeader(pgt.OriginHeaderName, pgt.GeekBang).
		SetLogger(logger.DiscardLogger{})

	g := new(errgroup.Group)
	ch := make(chan string, concurrency)

	for i := 0; i < concurrency; i++ {
		g.Go(func() error {
			return writeImageFile(ctx, ch, dir, imagesFolder, c, ss)
		})
	}

	for _, imageURL := range imageURLs {
		ch <- imageURL
	}
	close(ch)
	err = g.Wait()
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

func writeImageFile(ctx context.Context, imageURLs chan string, dir, imagesFolder string, c *resty.Client, ms *markdownString) (err error) {
	var es []error
loop:
	for {
		select {
		case <-ctx.Done():
			for range imageURLs {
			}
		case imageURL, ok := <-imageURLs:
			if !ok {
				break loop
			}
			if imageURL == "" {
				return
			}
			segments := strings.Split(imageURL, "/")
			f := segments[len(segments)-1]
			if i := strings.Index(f, "?"); i > 0 {
				f = f[:i]
			}
			imageLocalFullPath := filepath.Join(imagesFolder, f)
			rel, err := filepath.Rel(dir, imageLocalFullPath)
			if err != nil {
				es = append(es, err)
				break loop
			}

			_, err = c.R().
				SetContext(ctx).
				SetOutput(f).
				Get(imageURL)
			if err != nil {
				es = append(es, err)
				continue
			}

			ms.ReplaceAll(imageURL, filepath.ToSlash(rel))
		}
	}
	if len(es) > 0 {
		return es[0]
	}
	return nil
}
