package cmd

import (
	"context"
	"fmt"
	"math"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"sync/atomic"

	"github.com/briandowns/spinner"
	"github.com/gammazero/workerpool"
	"github.com/go-resty/resty/v2"
	"github.com/nicoxiang/geektime-downloader/cmd/prompt"
	"github.com/nicoxiang/geektime-downloader/internal/geektime"
	"github.com/nicoxiang/geektime-downloader/internal/loader"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/chromedp"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/util"
	"github.com/spf13/cobra"
)

const Pdf = "pdf"

var (
	phone              string
	concurrency        int
	downloadFolder     string
	l                  *spinner.Spinner
	columns            []geektime.ColumnSummary
	currentColumnIndex int
)

func init() {
	userHomeDir, _ := os.UserHomeDir()
	defaultConcurency := int(math.Ceil(float64(runtime.NumCPU()) / 2.0))
	defaultDownloadFolder := filepath.Join(userHomeDir, util.GeektimeDownloaderFolder)
	rootCmd.Flags().StringVarP(&phone, "phone", "u", "", "你的极客时间账号(手机号)(required)")
	_ = rootCmd.MarkFlagRequired("phone")
	rootCmd.Flags().StringVarP(&downloadFolder, "folder", "f", defaultDownloadFolder, "PDF 文件下载目标位置")
	rootCmd.Flags().IntVarP(&concurrency, "concurrency", "c", defaultConcurency, "下载文章的并发数")
	l = loader.NewSpinner()
}

var rootCmd = &cobra.Command{
	Use:   "geektime-downloader",
	Short: "Geektime-downloader is used to download geek time lessons",
	Run: func(cmd *cobra.Command, args []string) {
		readCookies, err := util.ReadCookieFromConfigFile(phone)
		if err != nil {
			printErrAndExit(err)
		}
		if readCookies == nil {
			pwd := prompt.GetPwd()
			loader.Run(l, "[ 正在登录... ]", func() {
				errMsg, cookies := geektime.Login(phone, pwd)
				if errMsg != "" {
					fmt.Fprintln(os.Stderr, errMsg)
					os.Exit(1)
				}
				readCookies = cookies
				err := util.WriteCookieToConfigFile(phone, cookies)
				if err != nil {
					printErrAndExit(err)
				}
			})
			fmt.Println("登录成功")
		}
		client := geektime.NewTimeGeekRestyClient(readCookies)

		loader.Run(l, "[ 正在加载已购买专栏列表... ]", func() {
			c, err := geektime.GetColumnList(client)
			if err != nil {
				printErrAndExit(err)
			}
			columns = c
		})

		selectColumn(cmd.Context(), client)
	},
}

func selectColumn(ctx context.Context, client *resty.Client) {
	if len(columns) == 0 {
		if err := util.RemoveConfig(phone); err != nil {
			printErrAndExit(err)
		} else {
			fmt.Println("当前账户在其他设备登录, 请尝试重新登录")
			os.Exit(1)
		}
	}
	currentColumnIndex = prompt.SelectColumn(columns)
	handleSelectColumn(ctx, client)
}

func handleSelectColumn(ctx context.Context, client *resty.Client) {
	option := prompt.SelectDownLoadAllOrSelectArticles(columns[currentColumnIndex].Title)
	handleSelectDownloadAll(ctx, option, client)
}

func handleSelectDownloadAll(ctx context.Context, option int, client *resty.Client) {
	switch option {
	case 0:
		selectColumn(ctx, client)
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
	folder, err := util.MkDownloadColumnFolder(downloadFolder, phone, columns[currentColumnIndex].Title)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	loader.Run(l, fmt.Sprintf("[ 正在下载文章 《%s》... ]", a.Title), func() {
		chromedp.PrintArticlePageToPDF(ctx, a.AID, filepath.Join(folder, util.FileName(a.Title, Pdf)), client.Cookies)
	})
	selectArticle(ctx, client)
}

func handleDownloadAll(ctx context.Context, client *resty.Client) {
	cTitle := columns[currentColumnIndex].Title
	articles := loadArticles(client)
	wp := workerpool.New(concurrency)
	var counter uint64
	folder, err := util.MkDownloadColumnFolder(downloadFolder, phone, cTitle)
	if err != nil {
		printErrAndExit(err)
	}
	downloaded, err := util.FindDownloadedArticleFileNames(downloadFolder, phone, cTitle)
	if err != nil {
		printErrAndExit(err)
	}
	for _, a := range articles {
		aid := a.AID
		fileName := util.FileName(a.Title, Pdf)
		wp.Submit(func() {
			prefix := fmt.Sprintf("[ 正在下载专栏 《%s》 中的所有文章, 已完成下载%d/%d ... ]", cTitle, counter, len(articles))
			loader.Run(l, prefix, func() {
				if _, ok := downloaded[fileName]; !ok {
					chromedp.PrintArticlePageToPDF(ctx, aid, filepath.Join(folder, fileName), client.Cookies)
				}
				atomic.AddUint64(&counter, 1)
			})
		})
	}
	wp.StopWait()
	selectColumn(ctx, client)
}

func loadArticles(client *resty.Client) []geektime.ArticleSummary {
	c := columns[currentColumnIndex]
	if len(c.Articles) <= 0 {
		loader.Run(l, "[ 正在加载文章列表...]", func() {
			articles, err := geektime.GetArticles(strconv.Itoa(c.CID), client)
			if err != nil {
				printErrAndExit(err)
			}
			columns[currentColumnIndex].Articles = articles
		})
	}
	return columns[currentColumnIndex].Articles
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
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
