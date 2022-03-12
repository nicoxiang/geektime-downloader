package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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

var Phone string
var Concurrency int
var DownloadFolder string
var userHomeDir string
var l *spinner.Spinner

func init() {
	userHomeDir, _ = os.UserHomeDir()
	downloadFolder := filepath.Join(userHomeDir, util.GeektimeDownloaderFolder)
	rootCmd.Flags().StringVarP(&Phone, "phone", "u", "", "你的极客时间账号(手机号)")
	rootCmd.Flags().StringVarP(&DownloadFolder, "folder", "f", downloadFolder, "PDF 文件下载目标位置")
	rootCmd.Flags().IntVarP(&Concurrency, "concurrency", "c", 5, "下载文章的并发数")
	l = loader.NewSpinner()
}

type RootContextKey string

const ColumnsKey RootContextKey = "columns"
const SelectedColumnPtrKey RootContextKey = "selectedColumnPtr"

var rootCmd = &cobra.Command{
	Use:   "geektime-downloader",
	Short: "Geektime-downloader is used to download geek time lessons",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		readCookies, err := util.ReadCookieFromConfigFile(Phone)
		if err != nil {
			printErrAndExit(err)
		}
		if readCookies == nil {
			pwd := prompt.PromptGetPwd()
			loader.Run(l, "[ 正在登录... ]", func() {
				errMsg, cookies := geektime.Login(Phone, pwd)
				if errMsg != "" {
					fmt.Fprintln(os.Stderr, errMsg)
					os.Exit(1)
				}
				readCookies = cookies
				err := util.WriteCookieToTempFile(Phone, cookies)
				if err != nil {
					printErrAndExit(err)
				}
			})
			fmt.Println("登录成功")
		}
		client := geektime.NewTimeGeekRestyClient(readCookies)

		loader.Run(l, "[ 正在加载已购买专栏列表... ]", func() {
			columns, err := geektime.GetColumnList(client)
			ctx = context.WithValue(ctx, ColumnsKey, columns)
			if err != nil {
				printErrAndExit(err)
			}
		})

		selectColumn(ctx, client)
	},
}

func selectColumn(ctx context.Context, client *resty.Client) {
	columns := ctx.Value(ColumnsKey).([]geektime.ColumnSummary)
	if len(columns) == 0 {
		if err := util.RemoveConfig(Phone); err != nil {
			printErrAndExit(err)
		} else {
			fmt.Println("当前账户在其他设备登录, 请尝试重新登录")
			os.Exit(1)
		}
	}
	selectedColumnIndex := prompt.PromptSelectColumn(columns)
	ctx = context.WithValue(ctx, SelectedColumnPtrKey, &columns[selectedColumnIndex])
	handleSelectColumn(ctx, client)
}

func handleSelectColumn(ctx context.Context, client *resty.Client) {
	c := ctx.Value(SelectedColumnPtrKey).(*geektime.ColumnSummary)
	option := prompt.PromptSelectDownLoadAllOrSelectArticles(c.Title)
	handleSelectDownloadAll(option, ctx, client)
}

func handleSelectDownloadAll(option int, ctx context.Context, client *resty.Client) {
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
	c := loadArticles(ctx, client)
	index := prompt.PromptSelectArticles(c.Articles)
	handleSelectArticle(c.Articles, index, ctx, client)
}

func handleSelectArticle(articles []geektime.ArticleSummary, index int, ctx context.Context, client *resty.Client) {
	if index == 0 {
		handleSelectColumn(ctx, client)
	}
	c := ctx.Value(SelectedColumnPtrKey).(*geektime.ColumnSummary)
	a := articles[index-1]
	folder, err := mkColumnDownloadFolder(c.Title)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	loader.Run(l, fmt.Sprintf("[ 正在下载文章 《%s》... ]", a.Title), func() {
		err := chromedp.PrintArticlePageToPDF(a.AID, filepath.Join(folder, util.FileName(a.Title, "pdf")), client.Cookies)
		if err != nil {
			printErrAndExit(err)
		}
	})
	selectArticle(ctx, client)
}

func handleDownloadAll(ctx context.Context, client *resty.Client) {
	c := loadArticles(ctx, client)
	wp := workerpool.New(Concurrency)
	var counter uint64
	folder, err := mkColumnDownloadFolder(c.Title)
	if err != nil {
		printErrAndExit(err)
	}
	for _, a := range c.Articles {
		aid := a.AID
		title := a.Title
		wp.Submit(func() {
			prefix := fmt.Sprintf("[ 正在下载专栏 《%s》 中的所有文章, 已完成下载%d/%d ... ]", c.Title, counter, len(c.Articles))
			loader.Run(l, prefix, func() {
				err := chromedp.PrintArticlePageToPDF(aid, filepath.Join(folder, util.FileName(title, "pdf")), client.Cookies)
				if err != nil {
					printErrAndExit(err)
				} else {
					atomic.AddUint64(&counter, 1)
				}
			})
		})
	}
	wp.StopWait()
	selectColumn(ctx, client)
}

func loadArticles(ctx context.Context, client *resty.Client) *geektime.ColumnSummary {
	c := ctx.Value(SelectedColumnPtrKey).(*geektime.ColumnSummary)
	if len(c.Articles) <= 0 {
		loader.Run(l, "[ 正在加载文章列表...]", func() {
			articles, err := geektime.GetArticles(strconv.Itoa(c.CID), client)
			if err != nil {
				printErrAndExit(err)
			}
			c.Articles = articles
		})
	}
	return c
}

func mkColumnDownloadFolder(columnName string) (string, error) {
	path := filepath.Join(DownloadFolder, columnName)
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return "", err
		}
	}
	return path, nil
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		printErrAndExit(err)
	}
}

func printErrAndExit(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
