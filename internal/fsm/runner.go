package fsm

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/manifoldco/promptui"
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

type FSMRunner struct {
	ctx                 context.Context
	currentState        State
	config              *config.AppConfig
	selectedProductType ui.ProductTypeSelectOption
	selectedProduct     geektime.Course
	sp                  *spinner.Spinner
	geektimeClient      *geektime.Client
	waitRand            *rand.Rand
	concurrency         int
}

const (
	outputPDF   = 1 << 0 // 1
	outputMD    = 1 << 1 // 2
	outputAudio = 1 << 2 // 4
)

// NewFSMRunner creates and initializes a new FSMRunner instance
func NewFSMRunner(ctx context.Context, cfg *config.AppConfig, geektimeClient *geektime.Client) *FSMRunner {
	return &FSMRunner{
		ctx:            ctx,
		currentState:   StateSelectProductType,
		config:         cfg,
		sp:             spinner.New(spinner.CharSets[4], 100*time.Millisecond),
		geektimeClient: geektimeClient,
		concurrency:    int(math.Ceil(float64(runtime.NumCPU()) / 2.0)),
		waitRand:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Run executes the finite state machine loop, handling user input and state transitions.
func (r *FSMRunner) Run() error {
	for {
		select {
		case <-r.ctx.Done():
			fmt.Println("\n⏹ 检测到取消信号，准备退出...")
			return r.ctx.Err()
		default:
		}

		var err error
		switch r.currentState {
		case StateSelectProductType:
			var selectedOption ui.ProductTypeSelectOption
			selectedOption, err = ui.ProductTypeSelect(r.config.IsEnterprise)
			if err == nil {
				r.currentState = StateInputProductID
				r.selectedProductType = selectedOption
			}
		case StateInputProductID:
			var productID int
			productID, err = ui.ProductIDInput(r.selectedProductType)
			if err == nil {
				if r.selectedProductType.NeedSelectArticle {
					err = r.handleInputProductIDIfNeedSelectArticle(productID)
				} else {
					err = r.handleInputProductIDIfDownloadDirectly(productID)
				}
			}
		case StateProductAction:
			var index int
			index, err = ui.ProductAction(r.selectedProduct)
			if err == nil {
				switch index {
				case 0:
					r.currentState = StateSelectProductType
				case 1:
					err = r.handleDownloadAll()
				case 2:
					r.currentState = StateSelectArticle
				}
			}
		case StateSelectArticle:
			var index int
			index, err = ui.ArticleSelect(r.selectedProduct.Articles)
			if err == nil {
				err = r.handleSelectArticle(index)
			}
		case StateExit:
			return nil
		}

		if err != nil {
			switch {
			case errors.Is(err, context.Canceled), errors.Is(err, promptui.ErrInterrupt):
				fmt.Println("\n用户中断操作")
			case os.IsTimeout(err):
				logger.Errorf(err, "Request timed out")
				fmt.Fprintln(os.Stderr, "\n请求超时")
			default:
				logger.Errorf(err, "An error occurred")
				fmt.Fprintf(os.Stderr, "\nAn error occurred: %v\n", err)
			}
			return err
		}
	}
}

func (r *FSMRunner) handleInputProductIDIfNeedSelectArticle(productID int) error {
	// choose download all or download specified article
	r.sp.Prefix = "[ 正在加载课程信息... ]"
	r.sp.Start()
	var course geektime.Course
	var err error
	if r.isUniversity() {
		// university don't need check product type
		// if input invalid id, access mark is 0
		course, err = r.geektimeClient.UniversityCourseInfo(productID)
		if err != nil {
			r.sp.Stop()
			return err
		}
	} else if r.config.IsEnterprise {
		// TODO: check enterprise course type
		course, err = r.geektimeClient.EnterpriseCourseInfo(productID)
		if err != nil {
			r.sp.Stop()
			return err
		}
	} else {
		course, err = r.geektimeClient.CourseInfo(productID)
		if err == nil {
			valid := r.validateProductCode(course.Type)
			// if check product type fail, re-input product
			if !valid {
				r.sp.Stop()
				r.currentState = StateInputProductID
				return nil
			}
		} else {
			return err
		}
	}

	r.sp.Stop()
	if !course.Access {
		fmt.Fprint(os.Stderr, "尚未购买该课程\n")
		r.currentState = StateInputProductID
	}
	r.selectedProduct = course
	r.currentState = StateProductAction
	return nil
}

func (r *FSMRunner) handleInputProductIDIfDownloadDirectly(productID int) error {
	// when product type is daily lesson or qconplus,
	// input id means product id
	// download video directly
	productInfo, err := r.geektimeClient.ProductInfo(productID)
	if err != nil {
		return err
	}

	if productInfo.Data.Info.Extra.Sub.AccessMask == 0 {
		fmt.Fprint(os.Stderr, "尚未购买该课程\n")
		r.currentState = StateInputProductID
		return nil
	}

	if r.validateProductCode(productInfo.Data.Info.Type) {
		columnDir, err := r.mkDownloadColumnDir(productInfo.Data.Info.Title)
		if err != nil {
			return err
		}

		err = video.DownloadArticleVideo(r.ctx,
			r.geektimeClient,
			productInfo.Data.Info.Article.ID,
			r.selectedProductType.SourceType,
			columnDir,
			r.config.Quality,
			r.concurrency)
		if err != nil {
			return err
		}
	}
	r.currentState = StateInputProductID
	return nil
}

// validateProductCode checks if the product code field in the response body returned by the API
// exists in the selected product's accepted product types list.
func (r *FSMRunner) validateProductCode(productCode string) bool {
	for _, pt := range r.selectedProductType.AcceptProductTypes {
		if pt == productCode {
			return true
		}
	}
	fmt.Fprint(os.Stderr, "\r输入的课程 ID 有误\n")
	return false
}

func (r *FSMRunner) handleSelectArticle(index int) error {
	if index == 0 {
		r.currentState = StateProductAction
		return nil
	}
	a := r.selectedProduct.Articles[index-1]

	columnDir, err := r.mkDownloadColumnDir(r.selectedProduct.Title)
	if err != nil {
		return err
	}
	err = r.downloadArticle(a, columnDir)
	if err != nil {
		return err
	}
	fmt.Printf("\r%s 下载完成", a.Title)
	time.Sleep(time.Second)
	r.currentState = StateSelectArticle
	return nil
}

// handleDownloadAll manages the bulk download process for all articles in a selected product (course).
// Returns an error if any step in the download process fails.
func (r *FSMRunner) handleDownloadAll() error {
	columnDir, err := r.mkDownloadColumnDir(r.selectedProduct.Title)
	if err != nil {
		return err
	}
	if geektime.IsTextCourse(r.selectedProduct) {
		fmt.Printf("正在下载专栏 《%s》 中的所有文章\n", r.selectedProduct.Title)
		total := len(r.selectedProduct.Articles)
		var i int

		for _, article := range r.selectedProduct.Articles {
			skip := r.skipDownloadTextArticle(article, columnDir, false)
			if !skip {
				logger.Infof("开始下载文章：《%s》", article.Title)
				err = r.downloadTextArticle(article, columnDir, false)
				if err != nil {
					return err
				}
				r.waitRandomTime()
			}
			increaseDownloadedTextArticleCount(total, &i)
		}
	} else {
		for _, article := range r.selectedProduct.Articles {
			skip := r.skipDownloadVideoArticle(article, columnDir, false)
			if !skip {
				err = r.downloadVideoArticle(article, columnDir)
				if err != nil {
					return err
				}
				r.waitRandomTime()
			}
		}
	}
	r.currentState = StateSelectProductType
	return nil
}

func increaseDownloadedTextArticleCount(total int, i *int) {
	*i++
	fmt.Printf("\r已完成下载%d/%d", *i, total)
}

// downloadArticle processes the download of a single article from Geektime.
// It handles both text-based courses and video content differently.
func (r *FSMRunner) downloadArticle(article geektime.Article, columnDir string) error {
	var err error
	if geektime.IsTextCourse(r.selectedProduct) {
		r.sp.Prefix = fmt.Sprintf("[ 正在下载 《%s》... ]", article.Title)
		r.sp.Start()
		defer r.sp.Stop()
		skip := r.skipDownloadTextArticle(article, columnDir, true)
		if !skip {
			err = r.downloadTextArticle(article, columnDir, true)
		}
	} else {
		skip := r.skipDownloadVideoArticle(article, columnDir, true)
		if !skip {
			err = r.downloadVideoArticle(article, columnDir)
		}
	}
	if err != nil {
		return err
	}
	return nil
}

func (r *FSMRunner) skipDownloadTextArticle(article geektime.Article, columnDir string, overwrite bool) bool {
	if overwrite {
		return false
	}

	needDownloadPDF := r.config.ColumnOutputType&outputPDF != 0
	needDownloadMD := r.config.ColumnOutputType&outputMD != 0
	needDownloadAudio := r.config.ColumnOutputType&outputAudio != 0

	// Check only the files that are requested.
	// If any requested file does not exist, do not skip.
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

	return false
}

// downloadTextArticle downloads the content of a Geektime text article in various formats (PDF, Markdown, Audio, and Video).
// The function supports overwriting existing files if specified.
func (r *FSMRunner) downloadTextArticle(article geektime.Article, columnDir string, overwrite bool) error {
	needDownloadPDF := r.config.ColumnOutputType&outputPDF != 0
	needDownloadMD := r.config.ColumnOutputType&outputMD != 0
	needDownloadAudio := r.config.ColumnOutputType&outputAudio != 0

	var err error
	articleInfo, err := r.geektimeClient.V1ArticleInfo(article.AID)
	if err != nil {
		return err
	}

	hasVideo, videoURL := getVideoURLFromArticleContent(articleInfo.Data.ArticleContent)
	if hasVideo && videoURL != "" {
		err = video.DownloadMP4(r.ctx, article.Title, columnDir, []string{videoURL}, overwrite)
		if err != nil {
			return err
		}
	}

	if len(articleInfo.Data.InlineVideoSubtitles) > 0 {
		videoURLs := make([]string, len(articleInfo.Data.InlineVideoSubtitles))
		for i, v := range articleInfo.Data.InlineVideoSubtitles {
			videoURLs[i] = v.VideoURL
		}
		err = video.DownloadMP4(r.ctx, article.Title, columnDir, videoURLs, overwrite)
		if err != nil {
			return err
		}
	}

	if needDownloadPDF {
		err := pdf.PrintArticlePageToPDF(r.ctx,
			article.AID,
			columnDir,
			article.Title,
			r.geektimeClient.Cookies,
			r.config.DownloadComments,
			r.config.PrintPDFWaitSeconds,
			r.config.PrintPDFTimeoutSeconds,
		)
		if err != nil {
			return err
		}
	}

	if needDownloadMD {
		err := markdown.Download(r.ctx,
			articleInfo.Data.ArticleContent,
			article.Title,
			columnDir,
			article.AID,
		)
		if err != nil {
			return err
		}
	}

	if needDownloadAudio {
		err := audio.DownloadAudio(r.ctx, articleInfo.Data.AudioDownloadURL, columnDir, article.Title)
		if err != nil {
			return err
		}
	}
	return nil
}


func (r *FSMRunner) skipDownloadVideoArticle(article geektime.Article, columnDir string, overwrite bool) bool {
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
func (r *FSMRunner) downloadVideoArticle(article geektime.Article, columnDir string) error {
	dir := columnDir
	var err error
	// add sub dir
	if article.SectionTitle != "" {
		dir, err = r.mkDownloadProjectSectionDir(columnDir, article.SectionTitle)
		if err != nil {
			return err
		}
	}

	if r.isUniversity() {
		err = video.DownloadUniversityVideo(r.ctx, r.geektimeClient, article.AID, r.selectedProduct, dir, r.config.Quality, r.concurrency)
	} else if r.config.IsEnterprise {
		err = video.DownloadEnterpriseArticleVideo(r.ctx, r.geektimeClient, article.AID, dir, r.config.Quality, r.concurrency)
	} else {
		err = video.DownloadArticleVideo(r.ctx, r.geektimeClient, article.AID, r.selectedProductType.SourceType, dir, r.config.Quality, r.concurrency)
	}
	return err
}

// mkDownloadColumnDir creates a directory for downloading a column with the given columnName.
func (r *FSMRunner) mkDownloadColumnDir(columnName string) (string, error) {
	path := filepath.Join(r.config.DownloadFolder, r.config.Gcid, filenamify.Filenamify(columnName))
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return "", err
	}
	return path, nil
}

func (r *FSMRunner) mkDownloadProjectSectionDir(projectDir, sectionName string) (string, error) {
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
func (r *FSMRunner) waitRandomTime() {
	randomMillis := r.config.Interval*1000 + r.waitRand.Intn(2000)
	time.Sleep(time.Duration(randomMillis) * time.Millisecond)
}

func (r *FSMRunner) isUniversity() bool {
	return r.selectedProductType.Index == 4 && !r.config.IsEnterprise
}
