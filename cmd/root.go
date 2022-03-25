package cmd

import (
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

		selectColumn(client)
	},
}

func selectColumn(client *resty.Client) {
	if len(columns) == 0 {
		if err := util.RemoveConfig(phone); err != nil {
			printErrAndExit(err)
		} else {
			fmt.Println("当前账户在其他设备登录, 请尝试重新登录")
			os.Exit(1)
		}
	}
	currentColumnIndex = prompt.SelectColumn(columns)
	handleSelectColumn(client)
}

func handleSelectColumn(client *resty.Client) {
	option := prompt.SelectDownLoadAllOrSelectArticles(columns[currentColumnIndex].Title)
	handleSelectDownloadAll(option, client)
}

func handleSelectDownloadAll(option int, client *resty.Client) {
	switch option {
	case 0:
		selectColumn(client)
	case 1:
		handleDownloadAll(client)
	case 2:
		selectArticle(client)
	}
}

func selectArticle(client *resty.Client) {
	articles := loadArticles(client)
	index := prompt.SelectArticles(articles)
	handleSelectArticle(articles, index, client)
}

func handleSelectArticle(articles []geektime.ArticleSummary, index int, client *resty.Client) {
	if index == 0 {
		handleSelectColumn(client)
	}
	a := articles[index-1]
	folder, err := mkColumnDownloadFolder(phone, columns[currentColumnIndex].Title)
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
	selectArticle(client)
}

func handleDownloadAll(client *resty.Client) {
	cTitle := columns[currentColumnIndex].Title
	articles := loadArticles(client)
	wp := workerpool.New(concurrency)
	var counter uint64
	folder, err := mkColumnDownloadFolder(phone, cTitle)
	if err != nil {
		printErrAndExit(err)
	}
	for _, a := range articles {
		aid := a.AID
		title := a.Title
		wp.Submit(func() {
			prefix := fmt.Sprintf("[ 正在下载专栏 《%s》 中的所有文章, 已完成下载%d/%d ... ]", cTitle, counter, len(articles))
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
	selectColumn(client)
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
