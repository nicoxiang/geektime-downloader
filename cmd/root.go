package cmd

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/chromedp/chromedp"
	"github.com/manifoldco/promptui"
	"github.com/nicoxiang/geektime-downloader/internal/geektime"
	"github.com/nicoxiang/geektime-downloader/internal/pdf"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/file"
	pgt "github.com/nicoxiang/geektime-downloader/internal/pkg/geektime"
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
	gcid                string
	gcess               string
	concurrency         int
	downloadFolder      string
	sp                  *spinner.Spinner
	products            []geektime.Product
	currentProductIndex int
	quality             string
	downloadComments    bool
)

func init() {
	rand.Seed(time.Now().UnixNano())

	userHomeDir, _ := os.UserHomeDir()
	concurrency = int(math.Ceil(float64(runtime.NumCPU()) / 2.0))
	defaultDownloadFolder := filepath.Join(userHomeDir, file.GeektimeDownloaderFolder)

	rootCmd.Flags().StringVarP(&phone, "phone", "u", "", "你的极客时间账号(手机号)")
	rootCmd.Flags().StringVar(&gcid, "gcid", "", "极客时间 cookie 值 gcid")
	rootCmd.Flags().StringVar(&gcess, "gcess", "", "极客时间 cookie 值 gcess")
	rootCmd.Flags().StringVarP(&downloadFolder, "folder", "f", defaultDownloadFolder, "专栏和视频课的下载目标位置")
	rootCmd.Flags().StringVarP(&quality, "quality", "q", "sd", "下载视频清晰度(ld标清,sd高清,hd超清)")
	rootCmd.Flags().BoolVar(&downloadComments, "comments", true, "是否需要专栏的第一页评论")

	sp = spinner.New(spinner.CharSets[4], 100*time.Millisecond)
}

var rootCmd = &cobra.Command{
	Use:   "geektime-downloader",
	Short: "Geektime-downloader is used to download geek time lessons",
	Run: func(cmd *cobra.Command, args []string) {
		if quality != "ld" && quality != "sd" && quality != "hd" {
			exitWithMsg("argument 'quality' is not valid")
		}
		var readCookies []*http.Cookie
		if phone != "" {
			rc, err := file.ReadCookieFromConfigFile(phone)
			if err != nil {
				exitWithError(err)
			}
			readCookies = rc
		} else if gcid != "" && gcess != "" {
			readCookies = readCookiesFromInput()
		} else {
			exitWithMsg("argument 'phone' or cookie value is not valid")
		}
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
			if err := file.WriteCookieToConfigFile(phone, readCookies); err != nil {
				exitWithError(err)
			}
			sp.Stop()
			fmt.Println("登录成功")
		}
		geektime.InitClient(readCookies)
		selectProduct(cmd.Context())
	},
}

func selectProduct(ctx context.Context) {
	loadProducts()
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "{{ `>` | red }} {{if eq .Type `c1`}} {{ `专栏` | red }} {{else}} {{ `视频课` | red }} {{end}} {{ .Title | red }} {{ .AuthorName | red }}",
		Inactive: "{{if eq .Type `c1`}} {{ `专栏` }} {{else}} {{ `视频课` }} {{end}} {{ .Title }} {{ .AuthorName }}",
	}
	prompt := promptui.Select{
		Label:        "我的课程列表, 请选择: ",
		Items:        products,
		Templates:    templates,
		Size:         20,
		HideSelected: true,
	}
	index, _, err := prompt.Run()
	checkPromptError(err)
	currentProductIndex = index
	handleSelectProduct(ctx)
}

func handleSelectProduct(ctx context.Context) {
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
		Active:   "{{ `>` | red }} {{ .Text | red }}",
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
	handleSelectProductOps(ctx, index)
}

func handleSelectProductOps(ctx context.Context, option int) {
	switch option {
	case 0:
		selectProduct(ctx)
	case 1:
		handleDownloadAll(ctx)
	case 2:
		selectArticle(ctx)
	}
}

func selectArticle(ctx context.Context) {
	articles := loadArticles()
	items := []geektime.Article{
		{
			AID:   -1,
			Title: "返回上一级",
		},
	}
	items = append(items, articles...)
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "{{ `>` | red }} {{ .Title | red }}",
		Inactive: "{{if eq .AID -1}} {{ .Title | green }} {{else}} {{ .Title }} {{end}}",
	}
	prompt := promptui.Select{
		Label:        "请选择文章: ",
		Items:        items,
		Templates:    templates,
		Size:         20,
		HideSelected: true,
		CursorPos:    0,
	}
	index, _, err := prompt.Run()
	checkPromptError(err)
	handleSelectArticle(ctx, articles, index)
}

func handleSelectArticle(ctx context.Context, articles []geektime.Article, index int) {
	if index == 0 {
		handleSelectProduct(ctx)
	}
	a := articles[index-1]

	projectDir, err := file.MkDownloadProjectFolder(downloadFolder, phone, gcid, products[currentProductIndex].Title)
	if err != nil {
		exitWithError(err)
	}
	downloadArticle(ctx, a, projectDir)
	fmt.Printf("\r%s 下载完成", a.Title)
	time.Sleep(time.Second)
	selectArticle(ctx)
}

func handleDownloadAll(ctx context.Context) {
	cTitle := products[currentProductIndex].Title
	articles := loadArticles()

	folder, err := file.MkDownloadProjectFolder(downloadFolder, phone, gcid, cTitle)
	if err != nil {
		exitWithError(err)
	}
	downloaded, err := file.FindDownloadedArticleFileNames(folder)
	if err != nil {
		exitWithError(err)
	}
	if isColumn() {
		fmt.Printf("正在下载专栏 《%s》 中的所有文章\n", cTitle)
		total := len(articles)
		var i int

		chromedpCtx, cancel := chromedp.NewContext(ctx)
		// start the browser
		err := chromedp.Run(chromedpCtx)
		if err != nil {
			exitWithError(err)
		}
		defer cancel()

		for _, a := range articles {
			fileName := getDownloadFileName(a)
			if _, ok := downloaded[fileName]; ok {
				increasePDFCount(total, &i)
				continue
			}
			fileFullPath := filepath.Join(folder, fileName)
			if err := pdf.PrintArticlePageToPDF(chromedpCtx, a.AID, fileFullPath, geektime.SiteCookies, downloadComments); err != nil {
				// ensure chrome killed before os exit
				cancel()
				checkGeekTimeError(err)
			}
			increasePDFCount(total, &i)
			r := rand.Intn(2000)
			time.Sleep(time.Duration(r) * time.Millisecond)
		}
	} else if isVideo() {
		for _, a := range articles {
			fileName := getDownloadFileName(a)
			if _, ok := downloaded[fileName]; ok {
				continue
			}
			videoInfo, err := geektime.GetVideoInfo(a.AID, quality)
			checkGeekTimeError(err)
			err = video.DownloadVideo(ctx, videoInfo.M3U8URL, fileName, folder, int64(videoInfo.Size), concurrency)
			checkGeekTimeError(err)
		}
	}
	selectProduct(ctx)
}

func increasePDFCount(total int, i *int) {
	(*i)++
	fmt.Printf("\r已完成下载%d/%d", *i, total)
}

func loadProducts() {
	if len(products) > 0 {
		return
	}
	sp.Prefix = "[ 正在加载已购买课程列表... ]"
	sp.Start()
	p, err := geektime.GetProductList()
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

func loadArticles() []geektime.Article {
	p := products[currentProductIndex]
	if len(p.Articles) <= 0 {
		sp.Prefix = "[ 正在加载文章列表... ]"
		sp.Start()
		articles, err := geektime.GetArticles(strconv.Itoa(p.ID))
		checkGeekTimeError(err)
		products[currentProductIndex].Articles = articles
		sp.Stop()
	}
	return products[currentProductIndex].Articles
}

func downloadArticle(ctx context.Context, article geektime.Article, projectDir string) {
	fileName := getDownloadFileName(article)
	fileFullPath := filepath.Join(projectDir, fileName)

	if isColumn() {
		sp.Prefix = fmt.Sprintf("[ 正在下载 《%s》... ]", article.Title)
		sp.Start()
		chromedpCtx, cancel := chromedp.NewContext(ctx)
		// start the browser
		err := chromedp.Run(chromedpCtx)
		if err != nil {
			exitWithError(err)
		}
		defer cancel()
		err = pdf.PrintArticlePageToPDF(chromedpCtx,
			article.AID,
			fileFullPath,
			geektime.SiteCookies,
			downloadComments,
		)
		sp.Stop()
		if err != nil {
			// ensure chrome killed before os exit
			cancel()
			checkGeekTimeError(err)
		}
	} else if isVideo() {
		videoInfo, err := geektime.GetVideoInfo(article.AID, quality)
		checkGeekTimeError(err)
		err = video.DownloadVideo(ctx, videoInfo.M3U8URL, fileName, projectDir, int64(videoInfo.Size), concurrency)
		checkGeekTimeError(err)
	}
}

func getDownloadFileName(article geektime.Article) string {
	var ext string
	if isColumn() {
		ext = PDFExtension
	} else if isVideo() {
		ext = TSExtension
	}
	return file.Filenamify(article.Title) + ext
}

func isColumn() bool {
	return products[currentProductIndex].Type == "c1"
}

func isVideo() bool {
	return products[currentProductIndex].Type == "c3"
}

func readCookiesFromInput() []*http.Cookie {
	oneyear := time.Now().Add(180 * 24 * time.Hour)
	cookies := make([]*http.Cookie, 2)
	m := make(map[string]string, 2)
	m[pgt.GCID] = gcid
	m[pgt.GCESS] = gcess
	c := 0
	for k, v := range m {
		cookies[c] = &http.Cookie{
			Name:     k,
			Value:    v,
			Domain:   pgt.GeekBangCookieDomain,
			HttpOnly: true,
			Expires:  oneyear,
		}
		c++
	}
	return cookies
}

func checkGeekTimeError(err error) {
	if err != nil {
		if errors.Is(err, context.Canceled) {
			os.Exit(1)
		} else if errors.Is(err, geektime.ErrWrongPassword) ||
			errors.Is(err, geektime.ErrTooManyLoginAttemptTimes) {
			exitWithMsg(err.Error())
		} else if errors.Is(err, pdf.ErrGeekTimeRateLimit) ||
			errors.Is(err, geektime.ErrAuthFailed) {

			// New line after print pdf success msg
			if errors.Is(err, pdf.ErrGeekTimeRateLimit) {
				fmt.Println()
			}

			fmt.Fprintln(os.Stderr, err.Error())
			if err := file.RemoveConfig(phone); err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
			}
			os.Exit(1)
		} else if os.IsTimeout(err) {
			exitWhenClientTimeout()
		} else if _, ok := err.(*geektime.ErrGeekTimeAPIBadCode); ok {
			exitWithMsg(err.Error())
		} else {
			// Client error, others
			exitWithError(err)
		}
	}
}

func checkPromptError(err error) {
	if err != nil {
		if !errors.Is(err, promptui.ErrInterrupt) {
			fmt.Fprint(os.Stderr, err)
		}
		os.Exit(1)
	}
}

func exitWhenClientTimeout() {
	exitWithMsg("Request Timeout")
}

// Unexpected error
func exitWithError(err error) {
	fmt.Fprintf(os.Stderr, "An error occurred: %v\n", err.Error())
	os.Exit(1)
}

func exitWithMsg(msg string, args ...interface{}) {
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
		exitWithError(err)
	}
}
