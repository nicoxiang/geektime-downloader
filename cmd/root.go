package cmd

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/briandowns/spinner"
	"github.com/go-resty/resty/v2"
	"github.com/nicoxiang/geektime-downloader/cmd/prompt"
	"github.com/nicoxiang/geektime-downloader/internal/client"
	"github.com/nicoxiang/geektime-downloader/internal/geektime"
	"github.com/nicoxiang/geektime-downloader/internal/loader"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/chromedp"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/file"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/video"
	"github.com/spf13/cobra"
)

// File extension
const (
	PDFExtension = ".pdf"
	TSExtension  = ".ts"
)

var (
	phone               string
	concurrency         int
	downloadFolder      string
	l                   *spinner.Spinner
	products            []geektime.Product
	currentProductIndex int
	quality				string
)

func init() {
	userHomeDir, _ := os.UserHomeDir()
	defaultConcurency := int(math.Ceil(float64(runtime.NumCPU()) / 2.0))
	defaultDownloadFolder := filepath.Join(userHomeDir, file.GeektimeDownloaderFolder)
	rootCmd.Flags().StringVarP(&phone, "phone", "u", "", "你的极客时间账号(手机号)(required)")
	_ = rootCmd.MarkFlagRequired("phone")
	rootCmd.Flags().StringVarP(&downloadFolder, "folder", "f", defaultDownloadFolder, "专栏和视频课的下载目标位置")
	rootCmd.Flags().IntVarP(&concurrency, "concurrency", "c", defaultConcurency, "下载并发数")
	rootCmd.Flags().StringVarP(&quality, "quality", "q", "sd", "下载视频清晰度(ld标清,sd高清,hd超清)")
	l = loader.NewSpinner()
}

var rootCmd = &cobra.Command{
	Use:   "geektime-downloader",
	Short: "Geektime-downloader is used to download geek time lessons",
	Run: func(cmd *cobra.Command, args []string) {
		if quality != "ld" && quality != "sd" && quality != "hd" {
			fmt.Fprintln(os.Stderr, "argument 'quality' is not valid")
			os.Exit(1)
		}
		readCookies := file.ReadCookieFromConfigFile(phone)
		if readCookies == nil {
			pwd := prompt.GetPwd()
			loader.Run(l, "[ 正在登录... ]", func() {
				readCookies = geektime.Login(phone, pwd)
				file.WriteCookieToConfigFile(phone, readCookies)
			})
			fmt.Println("登录成功")
		}
		client := client.NewTimeGeekRestyClient(readCookies)

		loader.Run(l, "[ 正在加载已购买课程列表... ]", func() {
			p, err := geektime.GetProductList(client)
			if err != nil {
				printErrAndExit(err)
			}
			products = p
		})

		selectProduct(cmd.Context(), client)
	},
}

func selectProduct(ctx context.Context, client *resty.Client) {
	currentProductIndex = prompt.SelectProduct(products)
	handleSelectColumn(ctx, client)
}

func handleSelectColumn(ctx context.Context, client *resty.Client) {
	option := prompt.SelectDownLoadAllOrSelectArticles(
		products[currentProductIndex].Title,
		products[currentProductIndex].Type,
	)
	handleSelectDownloadAll(ctx, option, client)
}

func handleSelectDownloadAll(ctx context.Context, option int, client *resty.Client) {
	switch option {
	case 0:
		selectProduct(ctx, client)
	case 1:
		handleDownloadAll(ctx, client)
	case 2:
		selectArticle(ctx, client)
	}
}

func selectArticle(ctx context.Context, client *resty.Client) {
	articles := loadArticles(client)
	index := prompt.SelectArticles(articles)
	handleSelectArticle(ctx, articles, index, client)
}

func handleSelectArticle(ctx context.Context, articles []geektime.ArticleSummary, index int, client *resty.Client) {
	if index == 0 {
		handleSelectColumn(ctx, client)
	}
	a := articles[index-1]
	projectDir := file.MkDownloadProjectFolder(downloadFolder, phone, products[currentProductIndex].Title)
	downloadArticle(ctx, a, projectDir, client)
	fmt.Printf("\r%s 下载完成", a.Title)
	time.Sleep(time.Second)
	selectArticle(ctx, client)
}

func handleDownloadAll(ctx context.Context, client *resty.Client) {
	cTitle := products[currentProductIndex].Title
	articles := loadArticles(client)

	folder := file.MkDownloadProjectFolder(downloadFolder, phone, cTitle)
	downloaded := file.FindDownloadedArticleFileNames(folder)
	if isColumn() {
		c := len(articles)
		var counter uint64
		var wg sync.WaitGroup
		wg.Add(c)
		fn := func(a geektime.ArticleSummary) {
			defer wg.Done()
			aid := a.AID
			fileName := file.Filenamify(a.Title) + PDFExtension
			fileFullPath := filepath.Join(folder, fileName)
			chromedp.PrintArticlePageToPDF(ctx, aid, fileFullPath, client.Cookies)
			atomic.AddUint64(&counter, 1)
			fmt.Printf("\r已完成下载%d/%d", counter, c)
		}
		ch := printPDFworker(concurrency, fn)
		fmt.Printf("正在下载专栏 《%s》 中的所有文章\n", cTitle)
		for _, a := range articles {
			fileName := file.Filenamify(a.Title) + PDFExtension
			if _, ok := downloaded[fileName]; ok {
				atomic.AddUint64(&counter, 1)
				continue
			}
			ch <- a
		}
	} else if isVideo() {
		for _, a := range articles {
			fileName := file.Filenamify(a.Title) + TSExtension
			if _, ok := downloaded[fileName]; ok {
				continue
			}
			videoInfo, err := geektime.GetVideoInfo(a.AID, "ld", client)
			if err != nil {
				printErrAndExit(err)
			}
			video.DownloadVideo(ctx, videoInfo.M3U8URL, a.Title, folder, int64(videoInfo.Size), concurrency)
		}
	}
	selectProduct(ctx, client)
}

type printPDFTask func(a geektime.ArticleSummary)

func printPDFworker(limit int, t printPDFTask) chan<- geektime.ArticleSummary {
	ch := make(chan geektime.ArticleSummary)
	for i := 0; i < limit; i++ {
		go func() {
			for {
				a, ok := <-ch
				if !ok {
					return
				}
				t(a)
			}
		}()
	}
	return ch
}

func loadArticles(client *resty.Client) []geektime.ArticleSummary {
	p := products[currentProductIndex]
	if len(p.Articles) <= 0 {
		loader.Run(l, "[ 正在加载文章列表...]", func() {
			articles, err := geektime.GetArticles(strconv.Itoa(p.ID), client)
			if err != nil {
				printErrAndExit(err)
			}
			products[currentProductIndex].Articles = articles
		})
	}
	return products[currentProductIndex].Articles
}

func downloadArticle(ctx context.Context, article geektime.ArticleSummary, projectDir string, client *resty.Client) {
	var ext string
	if isColumn() {
		ext = PDFExtension
	} else if isVideo() {
		ext = TSExtension
	}
	fileName := file.Filenamify(article.Title) + ext
	fileFullPath := filepath.Join(projectDir, fileName)

	if products[currentProductIndex].Type == "c1" {
		loader.Run(l, fmt.Sprintf("[ 正在下载 《%s》... ]", article.Title), func() {
			chromedp.PrintArticlePageToPDF(ctx,
				article.AID,
				fileFullPath,
				client.Cookies,
			)
		})
	} else if products[currentProductIndex].Type == "c3" {
		videoInfo, err := geektime.GetVideoInfo(article.AID, "ld", client)
		if err != nil {
			printErrAndExit(err)
		}
		video.DownloadVideo(ctx, videoInfo.M3U8URL, article.Title, projectDir, int64(videoInfo.Size), concurrency)
	}
}

func isColumn() bool {
	return products[currentProductIndex].Type == "c1"
}

func isVideo() bool {
	return products[currentProductIndex].Type == "c3"
}

// Execute ...
func Execute() {
	ctx := context.Background()

	ctx, cancel := context.WithCancel(ctx)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer func() {
		signal.Stop(c)
		cancel()
	}()
	go func() {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
		}
	}()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		printErrAndExit(err)
	}
}

func printErrAndExit(err error) {
	if errors.Is(err, geektime.ErrAuthFailed) {
		file.RemoveConfig(phone)
	}
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
