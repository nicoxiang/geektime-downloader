package cmd

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/briandowns/spinner"
	"github.com/go-resty/resty/v2"
	"github.com/manifoldco/promptui"
	"github.com/nicoxiang/geektime-downloader/internal/client"
	"github.com/nicoxiang/geektime-downloader/internal/geektime"

	"github.com/nicoxiang/geektime-downloader/internal/chromedp"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/file"
	"github.com/nicoxiang/geektime-downloader/internal/video"
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
	sp                  *spinner.Spinner
	products            []geektime.Product
	currentProductIndex int
	quality             string
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

	sp = spinner.New(spinner.CharSets[70], 100*time.Millisecond)
}

var rootCmd = &cobra.Command{
	Use:   "geektime-downloader",
	Short: "Geektime-downloader is used to download geek time lessons",
	Run: func(cmd *cobra.Command, args []string) {
		if quality != "ld" && quality != "sd" && quality != "hd" {
			exit("argument 'quality' is not valid")
		}
		readCookies := file.ReadCookieFromConfigFile(phone)
		if readCookies == nil {
			prompt := promptui.Prompt{
				Label: "请输入密码",
				Validate: func(s string) error {
					if strings.TrimSpace(s) == "" {
						return errors.New("密码不能为空")
					}
					return nil
				},
				Mask: '*',
			}
			pwd, err := prompt.Run()
			checkPromptError(err)
			sp.Prefix = "[ 正在登录... ]"
			sp.Start()
			readCookies, err = geektime.Login(phone, pwd)
			if err != nil {
				sp.Stop()
				checkGeekTimeError(err)
			}
			file.WriteCookieToConfigFile(phone, readCookies)
			sp.Stop()
			fmt.Println("登录成功")
		}
		client := client.NewTimeGeekRestyClient(readCookies)
		selectProduct(cmd.Context(), client)
	},
}

func selectProduct(ctx context.Context, client *resty.Client) {
	loadProducts(client)
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U00002714 {{if eq .Type `c1`}} {{ `专栏` | red }} {{else}} {{ `视频课` | red }} {{end}} {{ .Title | red }} {{ .AuthorName | red }}",
		Inactive: "{{if eq .Type `c1`}} {{ `专栏` }} {{else}} {{ `视频课` }} {{end}} {{ .Title }} {{ .AuthorName }}",
	}
	prompt := promptui.Select{
		Label:        "请选择课程: ",
		Items:        products,
		Templates:    templates,
		Size:         len(products),
		HideSelected: true,
	}
	index, _, err := prompt.Run()
	checkPromptError(err)
	currentProductIndex = index
	handleSelectProduct(ctx, client)
}

func handleSelectProduct(ctx context.Context, client *resty.Client) {
	currentProduct := products[currentProductIndex]
	type option struct {
		Text  string
		Value int
	}
	options := make([]option, 3)
	options[0] = option{"返回上一级", 0}
	if isColumn() {
		options[1] = option{"下载当前专栏所有文章", 1}
		options[2] = option{"选择文章", 2}
	} else if isVideo() {
		options[1] = option{"下载当前视频课所有视频", 1}
		options[2] = option{"选择视频", 2}
	}
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U00002714 {{ .Text | red }}",
		Inactive: "{{if eq .Value 0}} {{ .Text | green }} {{else}} {{ .Text }} {{end}}",
	}
	prompt := promptui.Select{
		Label:        fmt.Sprintf("当前选中的专栏为: %s, 请继续选择：", currentProduct.Title),
		Items:        options,
		Templates:    templates,
		Size:         len(options),
		HideSelected: true,
	}
	index, _, err := prompt.Run()
	checkPromptError(err)
	handleSelectProductOps(ctx, index, client)
}

func handleSelectProductOps(ctx context.Context, option int, client *resty.Client) {
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
	items := []geektime.ArticleSummary{
		{
			AID:   -1,
			Title: "返回上一级",
		},
	}
	items = append(items, articles...)
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U00002714 {{ .Title | red }}",
		Inactive: "{{if eq .AID -1}} {{ .Title | green }} {{else}} {{ .Title }} {{end}}",
	}
	prompt := promptui.Select{
		Label:        "请选择文章: ",
		Items:        items,
		Templates:    templates,
		Size:         len(items),
		HideSelected: true,
		CursorPos:    0,
	}
	index, _, err := prompt.Run()
	checkPromptError(err)
	handleSelectArticle(ctx, articles, index, client)
}

func handleSelectArticle(ctx context.Context, articles []geektime.ArticleSummary, index int, client *resty.Client) {
	if index == 0 {
		handleSelectProduct(ctx, client)
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
		// classic bounded work pooling pattern
		var counter uint64
		var wg sync.WaitGroup
		wg.Add(concurrency)
		ch := make(chan geektime.ArticleSummary, concurrency)
		fmt.Printf("正在下载专栏 《%s》 中的所有文章\n", cTitle)

		p := downloadSuccessPrinter{i: 0}
		for i := 0; i < concurrency; i++ {
			go func() {
				printPDF(ctx, ch, &wg, folder, client.Cookies, len(articles), &p)
			}()
		}
		
		for _, a := range articles {
			fileName := file.Filenamify(a.Title) + PDFExtension
			if _, ok := downloaded[fileName]; ok {
				atomic.AddUint64(&counter, 1)
				p.print(len(articles))
				continue
			}
			ch <- a
		}
		close(ch)
		wg.Wait()
	} else if isVideo() {
		for _, a := range articles {
			fileName := file.Filenamify(a.Title) + TSExtension
			if _, ok := downloaded[fileName]; ok {
				continue
			}
			videoInfo, err := geektime.GetVideoInfo(a.AID, "ld", client)
			checkGeekTimeError(err)
			video.DownloadVideo(ctx, videoInfo.M3U8URL, a.Title, folder, int64(videoInfo.Size), concurrency)
		}
	}
	selectProduct(ctx, client)
}

type downloadSuccessPrinter struct {
	mu sync.Mutex
	i  int
}

func (p *downloadSuccessPrinter) print(total int) {
	p.mu.Lock()
	p.i++
	fmt.Printf("\r已完成下载%d/%d", p.i, total)
	p.mu.Unlock()
}

func printPDF(ctx context.Context, ch chan geektime.ArticleSummary, wg *sync.WaitGroup, folder string, cookies []*http.Cookie, total int, printer *downloadSuccessPrinter) {
	defer wg.Done()

	for a := range ch {
		aid := a.AID
		fileName := file.Filenamify(a.Title) + PDFExtension
		fileFullPath := filepath.Join(folder, fileName)
		chromedp.PrintArticlePageToPDF(ctx, aid, fileFullPath, cookies)
		printer.print(total)
	}	
}

func loadProducts(client *resty.Client) {
	if len(products) > 0 {
		return
	}
	sp.Prefix = "[ 正在加载已购买课程列表... ]"
	sp.Start()
	p, err := geektime.GetProductList(client)
	if err != nil {
		sp.Stop()
		checkGeekTimeError(err)
	}
	if len(p) <= 0 {
		sp.Stop()
		fmt.Print("当前账户没有已购买课程")
		os.Exit(1)
	}
	products = p
	sp.Stop()
}

func loadArticles(client *resty.Client) []geektime.ArticleSummary {
	p := products[currentProductIndex]
	if len(p.Articles) <= 0 {
		sp.Prefix = "[ 正在加载文章列表... ]"
		sp.Start()
		articles, err := geektime.GetArticles(strconv.Itoa(p.ID), client)
		checkGeekTimeError(err)
		products[currentProductIndex].Articles = articles
		sp.Stop()
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

	if isColumn() {
		sp.Prefix = fmt.Sprintf("[ 正在下载 《%s》... ]", article.Title)
		sp.Start()
		chromedp.PrintArticlePageToPDF(ctx,
			article.AID,
			fileFullPath,
			client.Cookies,
		)
		sp.Stop()
	} else if isVideo() {
		videoInfo, err := geektime.GetVideoInfo(article.AID, "ld", client)
		checkGeekTimeError(err)
		video.DownloadVideo(ctx, videoInfo.M3U8URL, article.Title, projectDir, int64(videoInfo.Size), concurrency)
	}
}

func isColumn() bool {
	return products[currentProductIndex].Type == "c1"
}

func isVideo() bool {
	return products[currentProductIndex].Type == "c3"
}

func checkGeekTimeError(err error) {
	if err != nil {
		if errors.Is(err, geektime.ErrAuthFailed) {
			file.RemoveConfig(phone)
			fmt.Print(err.Error())
		} else if errors.Is(err, geektime.ErrWrongPassword) ||
			errors.Is(err, geektime.ErrTooManyLoginAttemptTimes) {
			fmt.Print(err.Error())
		} else {
			fmt.Fprintf(os.Stderr, "An error occurred: %v\n", err)
		}
		os.Exit(1)
	}
}

func checkPromptError(err error) {
	if err != nil {
		if !errors.Is(err, promptui.ErrInterrupt) {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}

func exit(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
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
		exit(err.Error())
	}
}
