package cmd

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
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

var (
	phone          string
	concurrency    int
	downloadFolder string
	l              *spinner.Spinner
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

type rootContextKey string

const columnsKey rootContextKey = "columns"
const selectedColumnPtrKey rootContextKey = "selectedColumnPtr"

var rootCmd = &cobra.Command{
	Use:   "geektime-downloader",
	Short: "Geektime-downloader is used to download geek time lessons",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

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
			columns, err := geektime.GetColumnList(client)
			ctx = context.WithValue(ctx, columnsKey, columns)
			if err != nil {
				printErrAndExit(err)
			}
		})

		selectColumn(ctx, client)
	},
}

func selectColumn(ctx context.Context, client *resty.Client) {
	columns := ctx.Value(columnsKey).([]geektime.ColumnSummary)
	if len(columns) == 0 {
		if err := util.RemoveConfig(phone); err != nil {
			printErrAndExit(err)
		} else {
			fmt.Println("当前账户在其他设备登录, 请尝试重新登录")
			os.Exit(1)
		}
	}
	selectedColumnIndex := prompt.SelectColumn(columns)
	ctx = context.WithValue(ctx, selectedColumnPtrKey, &columns[selectedColumnIndex])
	handleSelectColumn(ctx, client)
}

func handleSelectColumn(ctx context.Context, client *resty.Client) {
	c := ctx.Value(selectedColumnPtrKey).(*geektime.ColumnSummary)
	option := prompt.SelectDownLoadAllOrSelectArticles(c.Title)
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
	c := loadArticles(ctx, client)
	index := prompt.SelectArticles(c.Articles)
	handleSelectArticle(ctx, c.Articles, index, client)
}

func handleSelectArticle(ctx context.Context, articles []geektime.ArticleSummary, index int, client *resty.Client) {
	if index == 0 {
		handleSelectColumn(ctx, client)
	}
	c := ctx.Value(selectedColumnPtrKey).(*geektime.ColumnSummary)
	a := articles[index-1]
	folder, err := mkColumnDownloadFolder(phone, c.Title)
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
	wp := workerpool.New(concurrency)
	var counter uint64
	folder, err := mkColumnDownloadFolder(phone, c.Title)
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
	c := ctx.Value(selectedColumnPtrKey).(*geektime.ColumnSummary)
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

func mkColumnDownloadFolder(phone, columnName string) (string, error) {
	path := filepath.Join(downloadFolder, phone, columnName)
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return "", err
		}
	}
	return path, nil
}

// Execute func
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		printErrAndExit(err)
	}
}

func printErrAndExit(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
