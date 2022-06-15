package cmd

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
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
	"github.com/nicoxiang/geektime-downloader/internal/config"
	"github.com/nicoxiang/geektime-downloader/internal/geektime"
	"github.com/nicoxiang/geektime-downloader/internal/markdown"
	"github.com/nicoxiang/geektime-downloader/internal/pdf"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/filenamify"
	pgt "github.com/nicoxiang/geektime-downloader/internal/pkg/geektime"
	"github.com/nicoxiang/geektime-downloader/internal/video"
	"github.com/spf13/cobra"
)

var (
	phone            string
	gcid             string
	gcess            string
	concurrency      int
	downloadFolder   string
	sp               *spinner.Spinner
	currentProduct   geektime.Product
	quality          string
	downloadComments bool
	columnOutputType int8
)

func init() {
	userHomeDir, _ := os.UserHomeDir()
	concurrency = int(math.Ceil(float64(runtime.NumCPU()) / 2.0))
	defaultDownloadFolder := filepath.Join(userHomeDir, config.GeektimeDownloaderFolder)

	rootCmd.Flags().StringVarP(&phone, "phone", "u", "", "你的极客时间账号(手机号)")
	rootCmd.Flags().StringVar(&gcid, "gcid", "", "极客时间 cookie 值 gcid")
	rootCmd.Flags().StringVar(&gcess, "gcess", "", "极客时间 cookie 值 gcess")
	rootCmd.Flags().StringVarP(&downloadFolder, "folder", "f", defaultDownloadFolder, "专栏和视频课的下载目标位置")
	rootCmd.Flags().StringVarP(&quality, "quality", "q", "sd", "下载视频清晰度(ld标清,sd高清,hd超清)")
	rootCmd.Flags().BoolVar(&downloadComments, "comments", true, "是否需要专栏的第一页评论")
	rootCmd.Flags().Int8Var(&columnOutputType, "output", 1, "下载专栏的输出格式(1pdf,2markdown,3all)")

	sp = spinner.New(spinner.CharSets[4], 100*time.Millisecond)
}

var rootCmd = &cobra.Command{
	Use:   "geektime-downloader",
	Short: "Geektime-downloader is used to download geek time lessons",
	Run: func(cmd *cobra.Command, args []string) {
		if quality != "ld" && quality != "sd" && quality != "hd" {
			exitWithMsg("argument 'quality' is not valid")
		}
		if columnOutputType <= 0 || columnOutputType >= 4 {
			exitWithMsg("argument 'columnOutputType' is not valid")
		}
		var readCookies []*http.Cookie
		if phone != "" {
			rc, err := config.ReadCookieFromConfigFile(phone)
			checkError(err)
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
				Mask:        '*',
				HideEntered: true,
			}
			pwd, err := prompt.Run()
			checkError(err)
			sp.Prefix = "[ 正在登录... ]"
			sp.Start()
			readCookies, err = geektime.Login(phone, pwd)
			if err != nil {
				sp.Stop()
				checkError(err)
			}
			err = config.WriteCookieToConfigFile(phone, readCookies)
			checkError(err)
			sp.Stop()
			fmt.Fprintln(os.Stderr, "登录成功")
		}
		geektime.InitClient(readCookies)

		//first time auth check
		if err := geektime.Auth(); err != nil {
			checkError(geektime.ErrAuthFailed)
		}

		selectProduct(cmd.Context())
	},
}

func selectProduct(ctx context.Context) {
	prompt := promptui.Prompt{
		Label: "请输入课程 ID",
		Validate: func(s string) error {
			if strings.TrimSpace(s) == "" {
				return errors.New("课程 ID 不能为空")
			}
			if _, err := strconv.Atoi(s); err != nil {
				return errors.New("课程 ID 格式不合法")
			}
			return nil
		},
		HideEntered: true,
	}
	s, err := prompt.Run()
	checkError(err)

	// ignore, because checked before
	id, _ := strconv.Atoi(s)
	loadProduct(ctx, id)

	productOps(ctx)
}

func productOps(ctx context.Context) {
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
		Stdout:       NoBellStdout,
	}
	index, _, err := prompt.Run()
	checkError(err)

	switch index {
	case 0:
		selectProduct(ctx)
	case 1:
		handleDownloadAll(ctx)
	case 2:
		selectArticle(ctx)
	}
}

func selectArticle(ctx context.Context) {
	loadArticles()
	items := []geektime.Article{
		{
			AID:   -1,
			Title: "返回上一级",
		},
	}
	items = append(items, currentProduct.Articles...)
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
		Stdout:       NoBellStdout,
	}
	index, _, err := prompt.Run()
	checkError(err)
	handleSelectArticle(ctx, index)
}

func handleSelectArticle(ctx context.Context, index int) {
	if index == 0 {
		productOps(ctx)
	}
	a := currentProduct.Articles[index-1]

	projectDir, err := mkDownloadProjectDir(downloadFolder, phone, gcid, currentProduct.Title)
	checkError(err)
	downloadArticle(ctx, a, projectDir)
	fmt.Printf("\r%s 下载完成", a.Title)
	time.Sleep(time.Second)
	selectArticle(ctx)
}

func handleDownloadAll(ctx context.Context) {
	loadArticles()
	projectDir, err := mkDownloadProjectDir(downloadFolder, phone, gcid, currentProduct.Title)
	checkError(err)
	downloaded, err := findDownloadedArticleFileNames(projectDir)
	checkError(err)
	if isColumn() {
		rand.Seed(time.Now().UnixNano())
		fmt.Printf("正在下载专栏 《%s》 中的所有文章\n", currentProduct.Title)
		total := len(currentProduct.Articles)
		var i int

		var chromedpCtx context.Context
		var cancel context.CancelFunc

		if columnOutputType == 3 || columnOutputType == 1 {
			chromedpCtx, cancel = chromedp.NewContext(ctx)
			// start the browser
			err := chromedp.Run(chromedpCtx)
			checkError(err)
			defer cancel()
		}

		for _, a := range currentProduct.Articles {
			fileName := filenamify.Filenamify(a.Title)
			var b int8
			_, pdfExists := downloaded[fileName+pdf.PDFExtension]
			if pdfExists {
				b = 1
			}
			_, mdExists := downloaded[fileName+markdown.MDExtension]
			if mdExists {
				b |= (1 << 1)
			}

			if b == columnOutputType {
				increasePDFCount(total, &i)
				continue
			}

			if (columnOutputType&1 == 1) && !pdfExists {
				if err := pdf.PrintArticlePageToPDF(chromedpCtx,
					a.AID,
					projectDir,
					a.Title,
					geektime.SiteCookies,
					downloadComments,
				); err != nil {
					// ensure chrome killed before os exit
					cancel()
					checkError(err)
				}
			}
			if ((columnOutputType>>1)&1 == 1) && !mdExists {
				html, err := geektime.GetColumnContent(a.AID)
				checkError(err)
				err = markdown.Download(ctx, html, a.Title, projectDir, a.AID, concurrency)
				checkError(err)
			}

			increasePDFCount(total, &i)
			r := rand.Intn(2000)
			time.Sleep(time.Duration(r) * time.Millisecond)
		}
	} else if isVideo() {
		for _, a := range currentProduct.Articles {
			fileName := filenamify.Filenamify(a.Title) + video.TSExtension
			if _, ok := downloaded[fileName]; ok {
				continue
			}
			videoInfo, err := geektime.GetVideoInfo(a.AID, quality)
			checkError(err)
			err = video.DownloadVideo(ctx, videoInfo.M3U8URL, a.Title, projectDir, int64(videoInfo.Size), concurrency)
			checkError(err)
		}
	}
	selectProduct(ctx)
}

func increasePDFCount(total int, i *int) {
	(*i)++
	fmt.Printf("\r已完成下载%d/%d", *i, total)
}

func loadArticles() {
	if len(currentProduct.Articles) <= 0 {
		sp.Prefix = "[ 正在加载文章列表... ]"
		sp.Start()
		articles, err := geektime.GetArticles(strconv.Itoa(currentProduct.ID))
		checkError(err)
		currentProduct.Articles = articles
		sp.Stop()
	}
}

func loadProduct(ctx context.Context, productID int) {
	sp.Prefix = "[ 正在加载课程信息... ]"
	sp.Start()
	p, err := geektime.GetColumnInfo(productID)
	if err != nil {
		sp.Stop()
		checkError(err)
	}
	sp.Stop()
	if !p.Access {
		fmt.Fprint(os.Stderr, "尚未购买该课程\n")
		selectProduct(ctx)
	}
	currentProduct = p
}

func downloadArticle(ctx context.Context, article geektime.Article, projectDir string) {
	if isColumn() {
		sp.Prefix = fmt.Sprintf("[ 正在下载 《%s》... ]", article.Title)
		sp.Start()

		if columnOutputType&1 == 1 {
			chromedpCtx, cancel := chromedp.NewContext(ctx)
			// start the browser
			err := chromedp.Run(chromedpCtx)
			checkError(err)
			defer cancel()
			err = pdf.PrintArticlePageToPDF(chromedpCtx,
				article.AID,
				projectDir,
				article.Title,
				geektime.SiteCookies,
				downloadComments,
			)
			if err != nil {
				sp.Stop()
				// ensure chrome killed before os exit
				cancel()
				checkError(err)
			}
		}

		if (columnOutputType>>1)&1 == 1 {
			html, err := geektime.GetColumnContent(article.AID)
			checkError(err)
			err = markdown.Download(ctx, html, article.Title, projectDir, article.AID, concurrency)
			checkError(err)
		}
		sp.Stop()
	} else if isVideo() {
		videoInfo, err := geektime.GetVideoInfo(article.AID, quality)
		checkError(err)
		err = video.DownloadVideo(ctx, videoInfo.M3U8URL, article.Title, projectDir, int64(videoInfo.Size), concurrency)
		checkError(err)
	}
}

func isColumn() bool {
	return currentProduct.Type == "c1"
}

func isVideo() bool {
	return currentProduct.Type == "c3"
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

func findDownloadedArticleFileNames(projectDir string) (map[string]struct{}, error) {
	files, err := ioutil.ReadDir(projectDir)
	res := make(map[string]struct{}, len(files))
	if err != nil {
		return res, err
	}
	if len(files) == 0 {
		return res, nil
	}
	for _, f := range files {
		res[f.Name()] = struct{}{}
	}
	return res, nil
}

func mkDownloadProjectDir(downloadFolder, phone, gcid, projectName string) (string, error) {
	userName := phone
	if gcid != "" {
		userName = gcid
	}
	path := filepath.Join(downloadFolder, userName, filenamify.Filenamify(projectName))
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return "", err
	}
	return path, nil
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
		checkError(err)
	}
}
