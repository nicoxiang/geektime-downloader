package markdown

import (
	"context"
	"errors"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/JohannesKaufmann/dom"
	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/table"
	"golang.org/x/net/html"

	"github.com/nicoxiang/geektime-downloader/internal/geektime"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/downloader"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/filenamify"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/logger"
)

var (
	converterOnce sync.Once
	conv          *converter.Converter
	imgRegexp     = regexp.MustCompile(`!\[(.*?)]\((.*?)\)`)
)

// MDExtension ...
const MDExtension = ".md"

type markdownString struct {
	sync.Mutex
	s string
}

func (ms *markdownString) ReplaceAll(o, n string) {
	ms.Lock()
	defer ms.Unlock()
	ms.s = strings.ReplaceAll(ms.s, o, n)
}

// Download article as markdown
func Download(ctx context.Context, html, title, dir string, aid int) error {
	logger.Infof("Begin download article markdown, articleID: %d, title: %s", aid, title)

	select {
	case <-ctx.Done():
		return context.Canceled
	default:
	}

	markdwonFileName := path.Join(dir, filenamify.Filenamify(title)+MDExtension)

	// step1: convert to md string
	markdown, err := getDefaultConverter().ConvertString(html)
	if err != nil {
		logger.Errorf(err, "Failed to convert article html to markdown, articleID: %d, title: %s", aid, title)
		return err
	}
	// step2: download images
	ss := &markdownString{s: markdown}
	imageURLs := findAllImages(markdown)

	// images/aid/imageName.png
	imagesFolder := filepath.Join(dir, "images", strconv.Itoa(aid))

	if _, err := os.Stat(imagesFolder); errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(imagesFolder, os.ModePerm)
		if err != nil {
			return err
		}
	}

	err = writeImageFile(ctx, imageURLs, dir, imagesFolder, ss)
	if err != nil {
		logger.Errorf(err, "Failed to download article images, articleID: %d, title: %s, imagesURLs: %v", aid, title, imageURLs)
		return err
	}

	f, err := os.Create(markdwonFileName)
	defer func() {
		_ = f.Close()
	}()
	if err != nil {
		return err
	}
	// step3: write md file
	_, err = f.WriteString("# " + title + "\n" + ss.s)
	if err != nil {
		return err
	}
	logger.Infof("Finish download article markdown, articleID: %d, title: %s", aid, title)
	return nil
}

func findAllImages(md string) (images []string) {
	for _, matches := range imgRegexp.FindAllStringSubmatch(md, -1) {
		if len(matches) == 3 {
			s := matches[2]
			isImg, err := isImageURL(s)
			if err == nil && isImg {
				images = append(images, s)
			}
			// sometime exists broken image url, just ignore
		}
	}
	return
}

func getDefaultConverter() *converter.Converter {
	converterOnce.Do(func() {
		conv = converter.NewConverter(
			converter.WithPlugins(
				base.NewBasePlugin(),
				commonmark.NewCommonmarkPlugin(),
				table.NewTablePlugin(
					// 极客时间文章的表格没有 <th>，首行是列标题，把它提升为表头
					table.WithHeaderPromotion(true),
					// 单元格内允许换行，转换为 <br />，避免整张表被跳过
					table.WithNewlineBehavior(table.NewlineBehaviorPreserve),
				),
			),
		)
		// table 插件遇到 <ul>/<ol> 等节点会放弃整张表，
		// 所以在预渲染阶段把表格单元格内的列表解包成普通段落
		conv.Register.PreRenderer(unwrapListsInTables, converter.PriorityEarly)
	})
	return conv
}

// unwrapListsInTables 在预渲染阶段把 <table> 内部的 <ul>/<ol> 及其 <li>
// 解包，只保留子节点。极客时间文章的表格单元格经常包含列表，
// 而 html-to-markdown 的 table 插件遇到列表会跳过整张表。
func unwrapListsInTables(ctx converter.Context, doc *html.Node) {
	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		if dom.NodeName(n) == "table" {
			unwrapListNodes(n)
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
}

// unwrapListNodes 遍历 table 子树，将 ul/ol/li 节点替换为其子节点
func unwrapListNodes(tableNode *html.Node) {
	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		// 先收集 children，因为解包会修改链表
		var children []*html.Node
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			children = append(children, c)
		}
		for _, c := range children {
			walk(c)
		}
		name := dom.NodeName(n)
		if name == "ul" || name == "ol" || name == "li" {
			unwrapNode(n)
		}
	}
	for c := tableNode.FirstChild; c != nil; c = c.NextSibling {
		walk(c)
	}
}

// unwrapNode 把节点的所有子节点移到该节点的位置，然后移除该节点
func unwrapNode(n *html.Node) {
	parent := n.Parent
	if parent == nil {
		return
	}
	for c := n.FirstChild; c != nil; {
		next := c.NextSibling
		n.RemoveChild(c)
		parent.InsertBefore(c, n)
		c = next
	}
	parent.RemoveChild(n)
}

func writeImageFile(ctx context.Context,
	imageURLs []string,
	dir,
	imagesFolder string,
	ms *markdownString,
) (err error) {
	for _, imageURL := range imageURLs {
		segments := strings.Split(imageURL, "/")
		f := segments[len(segments)-1]
		if i := strings.Index(f, "?"); i > 0 {
			f = f[:i]
		}
		imageLocalFullPath := filepath.Join(imagesFolder, f)

		headers := make(map[string]string, 2)
		headers[geektime.Origin] = geektime.DefaultBaseURL
		headers[geektime.UserAgent] = geektime.DefaultUserAgent

		_, err := downloader.DownloadFileConcurrently(ctx, imageLocalFullPath, imageURL, headers, 1)
		if err != nil {
			return err
		}

		rel, _ := filepath.Rel(dir, imageLocalFullPath)
		ms.ReplaceAll(imageURL, filepath.ToSlash(rel))
	}
	return nil
}

func isImageURL(urlStr string) (bool, error) {
	// 解析 URL
	parsedURL, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return false, err
	}
	// 提取路径部分
	filePath := parsedURL.Path

	// 获取文件扩展名并转换为小写
	ext := strings.ToLower(path.Ext(filePath))

	// 检查扩展名是否属于图片类型
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".tiff":
		return true, nil
	default:
		return false, nil
	}
}
