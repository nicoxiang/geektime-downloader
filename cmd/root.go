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
	"github.com/manifoldco/promptui"
	"github.com/nicoxiang/geektime-downloader/internal/audio"
	"github.com/nicoxiang/geektime-downloader/internal/config"
	"github.com/nicoxiang/geektime-downloader/internal/geektime"
	"github.com/nicoxiang/geektime-downloader/internal/markdown"
	"github.com/nicoxiang/geektime-downloader/internal/pdf"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/filenamify"
	"github.com/nicoxiang/geektime-downloader/internal/video"
	"github.com/spf13/cobra"
	"golang.org/x/net/html"
)

var (
	phone                string
	gcid                 string
	gcess                string
	concurrency          int
	downloadFolder       string
	sp                   *spinner.Spinner
	selectedProduct      geektime.Product
	quality              string
	downloadComments     bool
	selectedProductType  productTypeSelectOption
	columnOutputType     int
	waitSeconds          int
	productTypeOptions   = make([]productTypeSelectOption, 7)
	geektimeClient       *geektime.Client
	geekEnterpriseClient *geektime.Client
	accountClient        *geektime.Client
	universityClient     *geektime.Client
)

type productTypeSelectOption struct {
	Index              int
	Text               string
	SourceType         int
	AcceptProductTypes []string
	needSelectArticle  bool
}

type articleOpsOption struct {
	Text  string
	Value int
}

func init() {
	userHomeDir, _ := os.UserHomeDir()
	concurrency = int(math.Ceil(float64(runtime.NumCPU()) / 2.0))
	defaultDownloadFolder := filepath.Join(userHomeDir, config.GeektimeDownloaderFolder)
	setProductTypeOptions()
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
}

func setProductTypeOptions() {
	productTypeOptions[0] = productTypeSelectOption{0, "普通课程", 1, []string{"c1", "c3"}, true}
	productTypeOptions[1] = productTypeSelectOption{1, "每日一课", 2, []string{"d"}, false}
	productTypeOptions[2] = productTypeSelectOption{2, "公开课", 1, []string{"p35", "p29", "p30"}, true}
	productTypeOptions[3] = productTypeSelectOption{3, "大厂案例", 4, []string{"q"}, false}
	productTypeOptions[4] = productTypeSelectOption{4, "训练营", 5, []string{""}, true} //custom source type, not use
	productTypeOptions[5] = productTypeSelectOption{5, "其他", 1, []string{"x", "c6"}, true}
	productTypeOptions[6] = productTypeSelectOption{6, "企业版训练营", 6, []string{"c44"}, true}
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
		geekEnterpriseClient = geektime.NewEnterpriseClient(readCookies)
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
	selectedProductType = productTypeOptions[index]
	letInputProductID(ctx)
}

func letInputProductID(ctx context.Context) {
	prompt := promptui.Prompt{
		Label: fmt.Sprintf("请输入%s的课程 ID", selectedProductType.Text),
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

	if selectedProductType.needSelectArticle {
		// choose download all or download specified article
		loadProduct(ctx, id)
		productOps(ctx)
	} else {
		// when product type is daily lesson or qconplus,
		// input id means product id
		// download video directly
		productInfo, err := geektimeClient.ProductInfo(id)
		checkError(err)

		if productInfo.Data.Info.Extra.Sub.AccessMask == 0 {
			fmt.Fprint(os.Stderr, "尚未购买该课程\n")
			letInputProductID(ctx)
		}

		if checkProductType(productInfo.Data.Info.Type) {
			projectDir, err := mkDownloadProjectDir(downloadFolder, phone, gcid, productInfo.Data.Info.Title)
			checkError(err)

			err = video.DownloadArticleVideo(ctx,
				geektimeClient,
				productInfo.Data.Info.Article.ID,
				selectedProductType.SourceType,
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
	if isUniversity() {
		p, err = universityClient.GetUniversityProductInfo(productID)
		// university don't need check product type
		// if input invalid id, access mark is 0
	} else if isEnterprise() {
		p, err = geekEnterpriseClient.GetEnterpriseProductInfo(productID)
	} else {
		p, err = geektimeClient.GetNormalProductInfo(productID)
		if err == nil {
			c := checkProductType(p.Type)
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
	selectedProduct = p
}

func productOps(ctx context.Context) {
	options := make([]articleOpsOption, 3)
	options[0] = articleOpsOption{"重新选择课程", 0}
	if isText() {
		options[1] = articleOpsOption{"下载当前专栏所有文章", 1}
		options[2] = articleOpsOption{"选择文章", 2}
	} else {
		options[1] = articleOpsOption{"下载所有视频", 1}
		options[2] = articleOpsOption{"选择视频", 2}
	}
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "{{ `>` | red }} {{ .Text | red }}",
		Inactive: "{{if eq .Value 0}} {{ .Text | green }} {{else}} {{ .Text }} {{end}}",
	}
	prompt := promptui.Select{
		Label:        fmt.Sprintf("当前选中的专栏为: %s, 请继续选择：", selectedProduct.Title),
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
	items := []geektime.Article{
		{
			AID:   -1,
			Title: "返回上一级",
		},
	}
	items = append(items, selectedProduct.Articles...)
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
	a := selectedProduct.Articles[index-1]

	projectDir, err := mkDownloadProjectDir(downloadFolder, phone, gcid, selectedProduct.Title)
	checkError(err)
	downloadArticle(ctx, a, projectDir)
	fmt.Printf("\r%s 下载完成", a.Title)
	time.Sleep(time.Second)
	selectArticle(ctx)
}

func handleDownloadAll(ctx context.Context) {
	projectDir, err := mkDownloadProjectDir(downloadFolder, phone, gcid, selectedProduct.Title)
	checkError(err)
	downloaded, err := findDownloadedArticleFileNames(projectDir)
	checkError(err)
	if isText() {
		rand.Seed(time.Now().UnixNano())
		fmt.Printf("正在下载专栏 《%s》 中的所有文章\n", selectedProduct.Title)
		total := len(selectedProduct.Articles)
		var i int

		needDownloadPDF := columnOutputType&1 == 1
		needDownloadMD := (columnOutputType>>1)&1 == 1
		needDownloadAudio := (columnOutputType>>2)&1 == 1

		for _, a := range selectedProduct.Articles {
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

			hasVideo, videoURL := getVideoURLFromArticleContent(articleInfo.Data.ArticleContent)

			if hasVideo && videoURL != "" {
				err = video.DownloadMP4(ctx, a.Title, projectDir, []string{videoURL})
			}

			if len(articleInfo.Data.InlineVideoSubtitles) > 0 {
				videoURLs := make([]string, len(articleInfo.Data.InlineVideoSubtitles))
				for i, v := range articleInfo.Data.InlineVideoSubtitles {
					videoURLs[i] = v.VideoURL
				}
				err = video.DownloadMP4(ctx, a.Title, projectDir, videoURLs)
			}

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
			}

			if needDownloadMD {
				err = markdown.Download(ctx,
					articleInfo.Data.ArticleContent,
					a.Title,
					projectDir,
					a.AID)
			}

			if needDownloadAudio {
				err = audio.DownloadAudio(ctx, articleInfo.Data.AudioDownloadURL, projectDir, a.Title)
			}

			checkError(err)

			increasePDFCount(total, &i)
			r := rand.Intn(2000)
			time.Sleep(time.Duration(r) * time.Millisecond)
		}
	} else {
		for _, a := range selectedProduct.Articles {
			sectionDir := projectDir
			fileName := filenamify.Filenamify(a.Title) + video.TSExtension
			if _, ok := downloaded[fileName]; ok {
				continue
			}
			// add sub dir
			if a.SectionTitle != "" {
				sectionDir, err = mkDownloadProjectSectionDir(projectDir, a.SectionTitle)
				checkError(err)
			}
			if isUniversity() {
				err := video.DownloadUniversityVideo(ctx, universityClient, a.AID, selectedProduct, sectionDir, quality, concurrency)
				checkError(err)
			} else if isEnterprise() {
				err := video.DownloadEnterpriseArticleVideo(ctx, geekEnterpriseClient, a.AID, selectedProductType.SourceType, sectionDir, quality, concurrency)
				checkError(err)
			} else {
				err := video.DownloadArticleVideo(ctx, geektimeClient, a.AID, selectedProductType.SourceType, sectionDir, quality, concurrency)
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

func downloadArticle(ctx context.Context, article geektime.Article, projectDir string) {
	if isText() {
		needDownloadPDF := columnOutputType&1 == 1
		needDownloadMD := (columnOutputType>>1)&1 == 1
		needDownloadAudio := (columnOutputType>>2)&1 == 1

		articleInfo, err := geektimeClient.V1ArticleInfo(article.AID)
		checkError(err)

		sp.Prefix = fmt.Sprintf("[ 正在下载 《%s》... ]", article.Title)
		hasVideo, videoURL := getVideoURLFromArticleContent(articleInfo.Data.ArticleContent)
		if len(articleInfo.Data.InlineVideoSubtitles) > 0 || hasVideo && videoURL != "" {
			sp.Prefix = fmt.Sprintf("[ 正在下载 《%s》, 该文章中包含视频, 请耐心等待... ]", article.Title)
		}
		sp.Start()

		if hasVideo && videoURL != "" {
			err = video.DownloadMP4(ctx, article.Title, projectDir, []string{videoURL})
		}

		if len(articleInfo.Data.InlineVideoSubtitles) > 0 {
			videoURLs := make([]string, len(articleInfo.Data.InlineVideoSubtitles))
			for i, v := range articleInfo.Data.InlineVideoSubtitles {
				videoURLs[i] = v.VideoURL
			}
			err = video.DownloadMP4(ctx, article.Title, projectDir, videoURLs)
		}

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
		}

		if needDownloadMD {
			err = markdown.Download(ctx,
				articleInfo.Data.ArticleContent,
				article.Title,
				projectDir,
				article.AID)
		}

		if needDownloadAudio {
			err = audio.DownloadAudio(ctx, articleInfo.Data.AudioDownloadURL, projectDir, article.Title)
		}

		sp.Stop()
		checkError(err)
	} else {
		if isUniversity() {
			err := video.DownloadUniversityVideo(ctx, universityClient, article.AID, selectedProduct, projectDir, quality, concurrency)
			checkError(err)
		} else if isEnterprise() {
			err := video.DownloadEnterpriseArticleVideo(ctx, geekEnterpriseClient, article.AID, selectedProductType.SourceType, projectDir, quality, concurrency)
			checkError(err)
		} else {
			err := video.DownloadArticleVideo(ctx, geektimeClient, article.AID, selectedProductType.SourceType, projectDir, quality, concurrency)
			checkError(err)
		}
	}
}

func isText() bool {
	return !selectedProduct.IsVideo
}

func isUniversity() bool {
	return selectedProductType.Index == 4
}

func isEnterprise() bool {
	return selectedProductType.Index == 6
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
	res := make(map[string]struct{})
	limit := 2
	err := filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("访问路径时出错：%v\n", err)
			return err
		}
		// 计算当前路径的深度
		depth := len(filepath.SplitList(path)) - len(filepath.SplitList(projectDir))
		if depth >= limit {
			return filepath.SkipDir // 如果达到限制深度，则跳过该文件夹及其子文件夹
		}
		if !info.IsDir() {
			res[info.Name()] = struct{}{}
		}
		return nil
	})
	checkError(err)
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

func mkDownloadProjectSectionDir(downloadFolder, sectionName string) (string, error) {
	path := filepath.Join(downloadFolder, filenamify.Filenamify(sectionName))
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return "", err
	}
	return path, nil
}

func checkProductType(productType string) bool {
	for _, pt := range selectedProductType.AcceptProductTypes {
		if pt == productType {
			return true
		}
	}
	fmt.Fprint(os.Stderr, "\r输入的课程 ID 有误\n")
	return false
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
