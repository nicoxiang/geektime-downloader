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
	"github.com/cavaliergopher/grab/v3"
	"github.com/manifoldco/promptui"
	"github.com/nicoxiang/geektime-downloader/internal/audio"
	"github.com/nicoxiang/geektime-downloader/internal/config"
	"github.com/nicoxiang/geektime-downloader/internal/geektime"
	"github.com/nicoxiang/geektime-downloader/internal/markdown"
	"github.com/nicoxiang/geektime-downloader/internal/pdf"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/filenamify"
	"github.com/nicoxiang/geektime-downloader/internal/video"
	"github.com/spf13/cobra"
)

var (
	phone              string
	gcid               string
	gcess              string
	concurrency        int
	downloadFolder     string
	sp                 *spinner.Spinner
	currentProduct     geektime.Product
	quality            string
	downloadComments   bool
	sourceType         int //video source type
	columnOutputType   int
	waitSeconds        int
	productTypeOptions = make([]productTypeSelectOption, 4)
	geektimeClient     *geektime.Client
	accountClient      *geektime.Client
	universityClient   *geektime.Client
	grabClient         *grab.Client
)

type productTypeSelectOption struct {
	Text  string
	Value int
}

func init() {
	userHomeDir, _ := os.UserHomeDir()
	concurrency = int(math.Ceil(float64(runtime.NumCPU()) / 2.0))
	defaultDownloadFolder := filepath.Join(userHomeDir, config.GeektimeDownloaderFolder)
	productTypeOptions[0] = productTypeSelectOption{"普通课程", 1}
	productTypeOptions[1] = productTypeSelectOption{"每日一课", 2}
	productTypeOptions[2] = productTypeSelectOption{"大厂案例", 4}
	productTypeOptions[3] = productTypeSelectOption{"训练营", 5} //custom source type, not use

	rootCmd.Flags().StringVarP(&phone, "phone", "u", "", "你的极客时间账号(手机号)")
	rootCmd.Flags().StringVar(&gcid, "gcid", "", "极客时间 cookie 值 gcid")
	rootCmd.Flags().StringVar(&gcess, "gcess", "", "极客时间 cookie 值 gcess")
	rootCmd.Flags().StringVarP(&downloadFolder, "folder", "f", defaultDownloadFolder, "专栏和视频课的下载目标位置")
	rootCmd.Flags().StringVarP(&quality, "quality", "q", "sd", "下载视频清晰度(ld标清,sd高清,hd超清)")
	rootCmd.Flags().BoolVar(&downloadComments, "comments", true, "是否需要专栏的第一页评论")
	rootCmd.Flags().IntVar(&columnOutputType, "output", 1, "专栏的输出内容(1pdf,2markdown,4audio)可自由组合")
	rootCmd.Flags().IntVar(&waitSeconds, "wait-seconds", 8, "Chrome生成PDF前的等待页面加载时间, 单位为秒, 默认8秒")

	rootCmd.MarkFlagsMutuallyExclusive("phone", "gcid")
	rootCmd.MarkFlagsMutuallyExclusive("phone", "gcess")
	rootCmd.MarkFlagsRequiredTogether("gcid", "gcess")

	sp = spinner.New(spinner.CharSets[4], 100*time.Millisecond)
	accountClient = geektime.NewAccountClient()
	grabClient = grab.NewClient()
	grabClient.UserAgent = geektime.DefaultUserAgent
}

var rootCmd = &cobra.Command{
	Use:   "geektime-downloader",
	Short: "Geektime-downloader is used to download geek time lessons",
	Run: func(cmd *cobra.Command, args []string) {
		if quality != "ld" && quality != "sd" && quality != "hd" {
			exitWithMsg("argument 'quality' is not valid")
		}
		if columnOutputType <= 0 || columnOutputType >= 8 {
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
			readCookies, err = accountClient.Login(phone, pwd)
			if err != nil {
				sp.Stop()
				checkError(err)
			}
			err = config.WriteCookieToConfigFile(phone, readCookies)
			checkError(err)
			sp.Stop()
			fmt.Fprintln(os.Stderr, "登录成功")
		}

		// first time auth check
		if err := accountClient.Auth(readCookies); err != nil {
			checkError(err)
		}
		geektimeClient = geektime.NewClient(readCookies)
		universityClient = geektime.NewUniversityClient(readCookies)
		selectProductType(cmd.Context())
	},
}

func selectProductType(ctx context.Context) {
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "{{ `>` | red }} {{ .Text | red }}",
		Inactive: "{{ .Text }}",
	}
	prompt := promptui.Select{
		Label:        "请选择想要下载的产品类型",
		Items:        productTypeOptions,
		Templates:    templates,
		Size:         len(productTypeOptions),
		HideSelected: true,
		Stdout:       NoBellStdout,
	}
	index, _, err := prompt.Run()
	checkError(err)
	sourceType = productTypeOptions[index].Value
	letInputProductID(ctx)
}

func letInputProductID(ctx context.Context) {
	prompt := promptui.Prompt{
		Label: fmt.Sprintf("请输入%s的课程 ID", findProductTypeText(sourceType)),
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

	if sourceType == 1 || sourceType == 5 {
		// when source type is normal cource or university cource
		// choose download all or download specified article
		loadProduct(ctx, id)
		productOps(ctx)
	} else {
		// when source type is daily lesson or qconplus,
		// input id means product id
		// download video directly
		productInfo, err := geektimeClient.ProductInfo(id)
		checkError(err)

		if checkProductType(productInfo.Data.Info.Type, sourceType) {
			projectDir, err := mkDownloadProjectDir(downloadFolder, phone, gcid, productInfo.Data.Info.Title)
			checkError(err)

			err = video.DownloadArticleVideo(ctx,
				geektimeClient,
				grabClient,
				productInfo.Data.Info.Article.ID,
				sourceType,
				projectDir,
				quality,
				concurrency)

			checkError(err)
		}
		letInputProductID(ctx)
	}
}

func loadProduct(ctx context.Context, productID int) {
	sp.Prefix = "[ 正在加载课程信息... ]"
	sp.Start()
	var p geektime.Product
	var err error
	if sourceType == 5 {
		p, err = universityClient.MyClassProduct(productID)
		// university don't need check product type
		// if input invalid id, access mark is 0
	} else if sourceType == 1 {
		p, err = geektimeClient.ColumnInfo(productID)
		if err == nil {
			c := checkProductType(p.Type, sourceType)
			// if check product type fail, re-input product
			if !c {
				sp.Stop()
				letInputProductID(ctx)
			}
		}
	}

	if err != nil {
		sp.Stop()
		checkError(err)
	}
	sp.Stop()
	if !p.Access {
		fmt.Fprint(os.Stderr, "尚未购买该课程\n")
		letInputProductID(ctx)
	}
	currentProduct = p
}

func productOps(ctx context.Context) {
	options := make([]productTypeSelectOption, 3)
	options[0] = productTypeSelectOption{"重新选择课程", 0}
	if isText() {
		options[1] = productTypeSelectOption{"下载当前专栏所有文章", 1}
		options[2] = productTypeSelectOption{"选择文章", 2}
	} else if isVideo() {
		options[1] = productTypeSelectOption{"下载所有视频", 1}
		options[2] = productTypeSelectOption{"选择视频", 2}
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
		selectProductType(ctx)
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
	if isText() {
		rand.Seed(time.Now().UnixNano())
		fmt.Printf("正在下载专栏 《%s》 中的所有文章\n", currentProduct.Title)
		total := len(currentProduct.Articles)
		var i int

		needDownloadPDF := columnOutputType&1 == 1
		needDownloadMD := (columnOutputType>>1)&1 == 1
		needDownloadAudio := (columnOutputType>>2)&1 == 1

		for _, a := range currentProduct.Articles {
			fileName := filenamify.Filenamify(a.Title)
			var b int
			if _, exists := downloaded[fileName+pdf.PDFExtension]; exists {
				b = setBit(b, 0)
			}
			if _, exists := downloaded[fileName+markdown.MDExtension]; exists {
				b = setBit(b, 1)
			}
			if _, exists := downloaded[fileName+audio.MP3Extension]; exists {
				b = setBit(b, 2)
			}

			if b == columnOutputType {
				increasePDFCount(total, &i)
				continue
			}

			articleInfo, err := geektimeClient.V1ArticleInfo(a.AID)
			checkError(err)

			if needDownloadPDF {
				err = pdf.PrintArticlePageToPDF(ctx,
					a.AID,
					projectDir,
					a.Title,
					geektimeClient.Cookies,
					downloadComments,
					waitSeconds,
				)
				if err != nil {
					checkError(err)
				}

				if len(articleInfo.Data.InlineVideoSubtitles) > 0 {
					videoURLs := make([]string, len(articleInfo.Data.InlineVideoSubtitles))
					for i, v := range articleInfo.Data.InlineVideoSubtitles {
						videoURLs[i] = v.VideoURL
					}
					err = video.DownloadMP4(ctx, grabClient, a.Title, projectDir, videoURLs)
				}
			}

			if needDownloadMD {
				err = markdown.Download(ctx,
					grabClient,
					articleInfo.Data.ArticleContent,
					a.Title,
					projectDir,
					a.AID,
					concurrency)
			}

			if needDownloadAudio {
				err = audio.DownloadAudio(ctx, grabClient, articleInfo.Data.AudioDownloadURL, projectDir, a.Title)
			}

			checkError(err)

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
			if sourceType == 1 {
				err := video.DownloadArticleVideo(ctx, geektimeClient, grabClient, a.AID, sourceType, projectDir, quality, concurrency)
				checkError(err)
			} else if sourceType == 5 {
				err := video.DownloadUniversityVideo(ctx, universityClient, grabClient, a.AID, currentProduct, projectDir, quality, concurrency)
				checkError(err)
			}
		}
	}
	selectProductType(ctx)
}

func increasePDFCount(total int, i *int) {
	(*i)++
	fmt.Printf("\r已完成下载%d/%d", *i, total)
}

func loadArticles() {
	if len(currentProduct.Articles) <= 0 {
		sp.Prefix = "[ 正在加载文章列表... ]"
		sp.Start()
		articles, err := geektimeClient.ColumnArticles(strconv.Itoa(currentProduct.ID))
		checkError(err)
		currentProduct.Articles = articles
		sp.Stop()
	}
}

func downloadArticle(ctx context.Context, article geektime.Article, projectDir string) {
	if isText() {
		needDownloadPDF := columnOutputType&1 == 1
		needDownloadMD := (columnOutputType>>1)&1 == 1
		needDownloadAudio := (columnOutputType>>2)&1 == 1

		articleInfo, err := geektimeClient.V1ArticleInfo(article.AID)
		checkError(err)

		sp.Prefix = fmt.Sprintf("[ 正在下载 《%s》... ]", article.Title)
		if len(articleInfo.Data.InlineVideoSubtitles) > 0 {
			sp.Prefix = fmt.Sprintf("[ 正在下载 《%s》, 该文章中包含视频, 请耐心等待... ]", article.Title)
		}
		sp.Start()

		if needDownloadPDF {
			checkError(err)
			err = pdf.PrintArticlePageToPDF(ctx,
				article.AID,
				projectDir,
				article.Title,
				geektimeClient.Cookies,
				downloadComments,
				waitSeconds,
			)
			if err != nil {
				sp.Stop()
				checkError(err)
			}

			if len(articleInfo.Data.InlineVideoSubtitles) > 0 {
				videoURLs := make([]string, len(articleInfo.Data.InlineVideoSubtitles))
				for i, v := range articleInfo.Data.InlineVideoSubtitles {
					videoURLs[i] = v.VideoURL
				}
				err = video.DownloadMP4(ctx, grabClient, article.Title, projectDir, videoURLs)
			}
		}

		if needDownloadMD {
			err = markdown.Download(ctx,
				grabClient,
				articleInfo.Data.ArticleContent,
				article.Title,
				projectDir,
				article.AID,
				concurrency)
		}

		if needDownloadAudio {
			err = audio.DownloadAudio(ctx, grabClient, articleInfo.Data.AudioDownloadURL, projectDir, article.Title)
		}

		sp.Stop()
		checkError(err)
	} else if isVideo() {
		if sourceType == 1 {
			err := video.DownloadArticleVideo(ctx, geektimeClient, grabClient, article.AID, sourceType, projectDir, quality, concurrency)
			checkError(err)
		} else if sourceType == 5 {
			err := video.DownloadUniversityVideo(ctx, universityClient, grabClient, article.AID, currentProduct, projectDir, quality, concurrency)
			checkError(err)
		}
	}
}

func isText() bool {
	return currentProduct.Type == string(geektime.ProductTypeColumn)
}

func isVideo() bool {
	return currentProduct.Type == string(geektime.ProductTypeNormalVideo) ||
		currentProduct.Type == string(geektime.ProductTypeUniversityVideo)
}

// Sets the bit at pos in the integer n.
func setBit(n int, pos uint) int {
	n |= (1 << pos)
	return n
}

func readCookiesFromInput() []*http.Cookie {
	oneyear := time.Now().Add(180 * 24 * time.Hour)
	cookies := make([]*http.Cookie, 2)
	m := make(map[string]string, 2)
	m[geektime.GCID] = gcid
	m[geektime.GCESS] = gcess
	c := 0
	for k, v := range m {
		cookies[c] = &http.Cookie{
			Name:     k,
			Value:    v,
			Domain:   geektime.GeekBangCookieDomain,
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

func checkProductType(productType string, sourceType int) bool {
	if (productType == string(geektime.ProductTypeDailyLesson) && sourceType == 2) ||
		(productType == string(geektime.ProductTypeQCONPlus) && sourceType == 4) ||
		(productType == string(geektime.ProductTypeColumn) && sourceType == 1) ||
		(productType == string(geektime.ProductTypeNormalVideo) && sourceType == 1) {
		return true
	}
	fmt.Fprint(os.Stderr, "\r输入的课程 ID 有误\n")
	return false
}

func findProductTypeText(sourceType int) string {
	for _, option := range productTypeOptions {
		if option.Value == sourceType {
			return option.Text
		}
	}
	return ""
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
