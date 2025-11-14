package course

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"golang.org/x/net/html"

	"github.com/nicoxiang/geektime-downloader/internal/audio"
	"github.com/nicoxiang/geektime-downloader/internal/config"
	"github.com/nicoxiang/geektime-downloader/internal/geektime"
	"github.com/nicoxiang/geektime-downloader/internal/markdown"
	"github.com/nicoxiang/geektime-downloader/internal/pdf"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/filenamify"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/files"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/logger"
	"github.com/nicoxiang/geektime-downloader/internal/ui"
	"github.com/nicoxiang/geektime-downloader/internal/video"
)

const (
	outputPDF   = 1 << 0 // 1
	outputMD    = 1 << 1 // 2
	outputAudio = 1 << 2 // 4
)

type CourseDownloader struct {
	ctx                context.Context
	cfg                *config.AppConfig
	geektimeClient     *geektime.Client
	concurrency        int
	waitRand           *rand.Rand
	downloadingSpinner *spinner.Spinner
}

func NewCourseDownloader(ctx context.Context, cfg *config.AppConfig, geektimeClient *geektime.Client, sp *spinner.Spinner) *CourseDownloader {
	concurrency := int(math.Ceil(float64(runtime.NumCPU()) / 2.0))
	if concurrency <= 0 {
		concurrency = 1
	}
	return &CourseDownloader{
		ctx:                ctx,
		cfg:                cfg,
		geektimeClient:     geektimeClient,
		concurrency:        concurrency,
		waitRand:           rand.New(rand.NewSource(time.Now().UnixNano())),
		downloadingSpinner: sp,
	}
}

// DownloadAll manages the bulk download process for all articles in a selected product (course).
// Returns an error if any step in the download process fails.
func (d *CourseDownloader) DownloadAll(course geektime.Course, productType ui.ProductTypeSelectOption) error {
	columnDir, err := d.mkDownloadColumnDir(course.Title)
	if err != nil {
		return err
	}

	if geektime.IsTextCourse(course) {
		fmt.Printf("正在下载专栏 《%s》 中的所有文章\n", course.Title)
		total := len(course.Articles)
		var downloaded int

		for _, article := range course.Articles {
			skip := d.skipDownloadTextArticle(article, columnDir, false)
			if !skip {
				logger.Infof("Begin download article, articleID: %d, articleTitle: %s", article.AID, article.Title)
				if err := d.downloadTextArticle(article, columnDir, false); err != nil {
					return err
				}
				d.waitRandomTime()
			}
			increaseDownloadedTextArticleCount(total, &downloaded)
		}
	} else {
		for _, article := range course.Articles {
			skip := d.skipDownloadVideoArticle(article, columnDir, false)
			if !skip {
				if err := d.downloadVideoArticle(course, productType, article, columnDir); err != nil {
					return err
				}
				d.waitRandomTime()
			}
		}
	}

	return nil
}

// DownloadArticle processes the download of a single article from Geektime.
// It handles both text-based courses and video content differently.
func (d *CourseDownloader) DownloadArticle(course geektime.Course, productType ui.ProductTypeSelectOption, article geektime.Article, overwrite bool) error {
	columnDir, err := d.mkDownloadColumnDir(course.Title)
	if err != nil {
		return err
	}

	if geektime.IsTextCourse(course) {
		d.downloadingSpinner.Prefix = fmt.Sprintf("[ 正在下载 《%s》... ]", article.Title)
		d.downloadingSpinner.Start()
		defer d.downloadingSpinner.Stop()
		skip := d.skipDownloadTextArticle(article, columnDir, overwrite)
		if skip {
			return nil
		}
		return d.downloadTextArticle(article, columnDir, overwrite)
	} else {
		skip := d.skipDownloadVideoArticle(article, columnDir, overwrite)
		if skip {
			return nil
		}
		return d.downloadVideoArticle(course, productType, article, columnDir)
	}
}

// DownloadSingleVideoProduct downloads a single video product.
// 每日一课，大厂案例等
func (d *CourseDownloader) DownloadSingleVideoProduct(title string, articleID int, sourceType int) error {
	columnDir, err := d.mkDownloadColumnDir(title)
	if err != nil {
		return err
	}
	return video.DownloadArticleVideo(d.ctx, d.geektimeClient, articleID, sourceType, columnDir, d.cfg.Quality, d.concurrency)
}

func increaseDownloadedTextArticleCount(total int, downloaded *int) {
	current := *downloaded + 1
	*downloaded = current
	if current > total {
		current = total
	}
	fmt.Printf("\r已完成下载%d/%d", current, total)
}

func (d *CourseDownloader) skipDownloadTextArticle(article geektime.Article, columnDir string, overwrite bool) bool {
	if overwrite {
		return false
	}

	needDownloadPDF := d.cfg.ColumnOutputType&outputPDF != 0
	needDownloadMD := d.cfg.ColumnOutputType&outputMD != 0
	needDownloadAudio := d.cfg.ColumnOutputType&outputAudio != 0

	if needDownloadPDF {
		pdfFileName := filepath.Join(columnDir, filenamify.Filenamify(article.Title)+pdf.PDFExtension)
		if !files.CheckFileExists(pdfFileName) {
			return false
		}
	}
	if needDownloadMD {
		markdownFileName := filepath.Join(columnDir, filenamify.Filenamify(article.Title)+markdown.MDExtension)
		if !files.CheckFileExists(markdownFileName) {
			return false
		}
	}
	if needDownloadAudio {
		audioFileName := filepath.Join(columnDir, filenamify.Filenamify(article.Title)+audio.MP3Extension)
		if !files.CheckFileExists(audioFileName) {
			return false
		}
	}

	return true
}

// downloadTextArticle downloads the content of a Geektime text article in various formats (PDF, Markdown, Audio, and Video).
// The function supports overwriting existing files if specified.
func (d *CourseDownloader) downloadTextArticle(article geektime.Article, columnDir string, overwrite bool) error {
	needDownloadPDF := d.cfg.ColumnOutputType&outputPDF != 0
	needDownloadMD := d.cfg.ColumnOutputType&outputMD != 0
	needDownloadAudio := d.cfg.ColumnOutputType&outputAudio != 0

	articleInfo, err := d.geektimeClient.V1ArticleInfo(article.AID)
	if err != nil {
		return err
	}

	hasVideo, videoURL := getVideoURLFromArticleContent(articleInfo.Data.ArticleContent)
	if hasVideo && videoURL != "" {
		if err := video.DownloadMP4(d.ctx, article.Title, columnDir, []string{videoURL}, overwrite); err != nil {
			return err
		}
	}

	if len(articleInfo.Data.InlineVideoSubtitles) > 0 {
		videoURLs := make([]string, len(articleInfo.Data.InlineVideoSubtitles))
		for i, v := range articleInfo.Data.InlineVideoSubtitles {
			videoURLs[i] = v.VideoURL
		}
		if err := video.DownloadMP4(d.ctx, article.Title, columnDir, videoURLs, overwrite); err != nil {
			return err
		}
	}

	if needDownloadPDF {
		if err := pdf.PrintArticlePageToPDF(d.ctx,
			article,
			columnDir,
			d.geektimeClient.Cookies,
			d.cfg,
		); err != nil {
			return err
		}
	}

	if needDownloadMD {
		if err := markdown.Download(d.ctx,
			articleInfo.Data.ArticleContent,
			article.Title,
			columnDir,
			article.AID,
		); err != nil {
			return err
		}
	}

	if needDownloadAudio {
		if err := audio.DownloadAudio(d.ctx, articleInfo.Data.AudioDownloadURL, columnDir, article.Title); err != nil {
			return err
		}
	}
	return nil
}

func (d *CourseDownloader) skipDownloadVideoArticle(article geektime.Article, columnDir string, overwrite bool) bool {
	dir := columnDir
	fileName := filenamify.Filenamify(article.Title) + video.TSExtension
	fullPath := filepath.Join(dir, fileName)
	if files.CheckFileExists(fullPath) && !overwrite {
		return true
	}
	return false
}

// downloadVideoArticle downloads a video article to the specified column directory.
// It handles different types of video content including university courses, enterprise content,
// and regular article videos.
func (d *CourseDownloader) downloadVideoArticle(course geektime.Course, productType ui.ProductTypeSelectOption, article geektime.Article, columnDir string) error {
	dir := columnDir
	var err error
	if article.SectionTitle != "" {
		dir, err = d.mkDownloadProjectSectionDir(columnDir, article.SectionTitle)
		if err != nil {
			return err
		}
	}

	if productType.IsUniversity() {
		err = video.DownloadUniversityVideo(d.ctx, d.geektimeClient, article.AID, course, dir, d.cfg.Quality, d.concurrency)
	} else if d.cfg.IsEnterprise {
		err = video.DownloadEnterpriseArticleVideo(d.ctx, d.geektimeClient, article.AID, dir, d.cfg.Quality, d.concurrency)
	} else {
		err = video.DownloadArticleVideo(d.ctx, d.geektimeClient, article.AID, productType.SourceType, dir, d.cfg.Quality, d.concurrency)
	}
	return err
}

// mkDownloadColumnDir creates a directory for downloading a column with the given columnName.
func (d *CourseDownloader) mkDownloadColumnDir(columnName string) (string, error) {
	path := filepath.Join(d.cfg.DownloadFolder, filenamify.Filenamify(columnName))
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return "", err
	}
	return path, nil
}

// mkDownloadProjectSectionDir creates a sub directory for project sections.
func (d *CourseDownloader) mkDownloadProjectSectionDir(projectDir, sectionName string) (string, error) {
	path := filepath.Join(projectDir, filenamify.Filenamify(sectionName))
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return "", err
	}
	return path, nil
}

// Sometime video exist in article content, see issue #104
// <p>
// <video poster="https://static001.geekbang.org/resource/image/6a/f7/6ada085b44eddf37506b25ad188541f7.jpg" preload="none" controls="">
// <source src="https://media001.geekbang.org/customerTrans/fe4a99b62946f2c31c2095c167b26f9c/30d99c0d-16d14089303-0000-0000-01d-dbacd.mp4" type="video/mp4">
// <source src="https://media001.geekbang.org/2ce11b32e3e740ff9580185d8c972303/a01ad13390fe4afe8856df5fb5d284a2-f2f547049c69fa0d4502ab36d42ea2fa-sd.m3u8" type="application/x-mpegURL">
// <source src="https://media001.geekbang.org/2ce11b32e3e740ff9580185d8c972303/a01ad13390fe4afe8856df5fb5d284a2-2528b0077e78173fd8892de4d7b8c96d-hd.m3u8" type="application/x-mpegURL"></video>
// </p>
func getVideoURLFromArticleContent(content string) (hasVideo bool, videoURL string) {
	if !strings.Contains(content, "<video") || !strings.Contains(content, "<source") {
		return false, ""
	}
	doc, err := html.Parse(strings.NewReader(content))
	if err != nil {
		return false, ""
	}
	hasVideo, videoURL = false, ""
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "video" {
			hasVideo = true
		}
		if n.Type == html.ElementNode && n.Data == "source" {
			for _, a := range n.Attr {
				if a.Key == "src" && hasVideo && strings.HasSuffix(a.Val, ".mp4") {
					videoURL = a.Val
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return hasVideo, videoURL
}

// waitRandomTime wait interval seconds of time plus a 2000ms max jitter
func (d *CourseDownloader) waitRandomTime() {
	randomMillis := d.cfg.Interval*1000 + d.waitRand.Intn(2000)
	time.Sleep(time.Duration(randomMillis) * time.Millisecond)
}