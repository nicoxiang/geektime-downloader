package geektime

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/nicoxiang/geektime-downloader/internal/geektime/response"
	pgt "github.com/nicoxiang/geektime-downloader/internal/pkg/geektime"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/logger"
)

const (

	// V1ColumnArticlesPath get all articles summary info in one column
	V1ColumnArticlesPath = "/serv/v1/column/articles"
	// V1ArticlePath used in normal column
	V1ArticlePath = "/serv/v1/article"
	// V1VideoPlayAuthPath used in university video play auth
	V1VideoPlayAuthPath = "/serv/v1/video/play-auth"
	// V1MyClassInfoPath get university class info and all articles info in it
	V1MyClassInfoPath = "/serv/v1/myclass/info"

	// V3ColumnInfoPath used in get normal column/video info
	V3ColumnInfoPath = "/serv/v3/column/info"
	// V3ProductInfoPath used in get daily lesson, qconplus product info
	V3ProductInfoPath = "/serv/v3/product/info"
	// V3ArticleInfoPath used in normal video, daily lesson, qconplus 
	V3ArticleInfoPath = "serv/v3/article/info"
	// V3VideoPlayAuthPath used in normal video, daily lesson, qconplus video play auth
	V3VideoPlayAuthPath = "serv/v3/source_auth/video_play_auth"

	// ProductTypeColumn c1 column
	ProductTypeColumn = "c1"
	// ProductTypeNormalVideo c3 normal video
	ProductTypeNormalVideo = "c3"
	// ProductTypeUniversityVideo u university video
	ProductTypeUniversityVideo = "u"
)

var (
	geekTimeClient  *resty.Client
	accountClient   *resty.Client
	ugeekTimeClient *resty.Client
	// SiteCookies ...
	SiteCookies []*http.Cookie
)

// ErrGeekTimeAPIBadCode ...
type ErrGeekTimeAPIBadCode struct {
	Path string
	Code int
	Msg  string
}

// Error implements error interface
func (e ErrGeekTimeAPIBadCode) Error() string {
	return fmt.Sprintf("请求极客时间接口 %s 失败, code %d, msg %s", e.Path, e.Code, e.Msg)
}

// Product ...
type Product struct {
	Access   bool
	ID       int
	Title    string
	Type     string
	Articles []Article
}

// Article ...
type Article struct {
	AID   int
	Title string
}

// ArticleInfo ...
type ArticleInfo struct {
	ArticleContent   string
	AudioDownloadURL string
}

// InitClient init golbal clients with cookies
func InitClient(cookies []*http.Cookie) {
	geekTimeClient = resty.New().
		SetBaseURL(pgt.GeekBang).
		SetCookies(cookies).
		SetRetryCount(1).
		SetTimeout(10*time.Second).
		SetHeader(pgt.UserAgentHeaderName, pgt.UserAgentHeaderValue).
		SetHeader(pgt.OriginHeaderName, pgt.GeekBang).
		SetLogger(logger.DiscardLogger{})

	accountClient = resty.New().
		SetBaseURL(pgt.GeekBangAccount).
		SetCookies(cookies).
		SetTimeout(10*time.Second).
		SetHeader(pgt.UserAgentHeaderName, pgt.UserAgentHeaderValue).
		SetHeader(pgt.OriginHeaderName, pgt.GeekBang).
		SetLogger(logger.DiscardLogger{})

	ugeekTimeClient = resty.New().
		SetBaseURL(pgt.GeekBangUniversity).
		SetCookies(cookies).
		SetTimeout(10*time.Second).
		SetHeader(pgt.UserAgentHeaderName, pgt.UserAgentHeaderValue).
		SetHeader(pgt.OriginHeaderName, pgt.GeekBangUniversity).
		SetLogger(logger.DiscardLogger{})

	SiteCookies = cookies
}

// PostV3ColumnInfo  ..
func PostV3ColumnInfo(productID int) (Product, error) {
	var p Product
	if err := Auth(); err != nil {
		return p, err
	}
	var result response.V3ColumnInfoResponse
	resp, err := geekTimeClient.R().
		SetBody(
			map[string]interface{}{
				"product_id":             productID,
				"with_recommend_article": true,
			}).
		SetResult(&result).
		Post(V3ColumnInfoPath)

	if err != nil {
		return p, err
	}

	if resp.RawResponse.StatusCode == 451 {
		return p, pgt.ErrGeekTimeRateLimit
	}

	if result.Code == 0 {
		p = Product{
			Access: result.Data.Extra.Sub.AccessMask > 0,
			ID:     result.Data.ID,
			Type:   result.Data.Type,
			Title:  result.Data.Title,
		}
		return p, nil
	}

	return p, ErrGeekTimeAPIBadCode{V3ColumnInfoPath, result.Code, ""}
}

// PostV1ColumnArticles call geektime api to get article list
func PostV1ColumnArticles(cid string) ([]Article, error) {
	if err := Auth(); err != nil {
		return nil, err
	}

	var result response.V1ColumnArticlesResponse
	resp, err := geekTimeClient.R().
		SetBody(
			map[string]interface{}{
				"cid":    cid,
				"order":  "earliest",
				"prev":   0,
				"sample": false,
				"size":   500, //get all articles
			}).
		SetResult(&result).
		Post(V1ColumnArticlesPath)

	if err != nil {
		return nil, err
	}

	if resp.RawResponse.StatusCode == 451 {
		return nil, pgt.ErrGeekTimeRateLimit
	}

	if result.Code == 0 {
		var articles []Article
		for _, v := range result.Data.List {
			articles = append(articles, Article{
				AID:   v.ID,
				Title: v.ArticleTitle,
			})
		}
		return articles, nil
	}

	return nil, ErrGeekTimeAPIBadCode{V1ColumnArticlesPath, result.Code, ""}
}

// GetArticleInfo ...
func GetArticleInfo(articleID int) (ArticleInfo, error) {
	var result response.V1ArticleResponse
	var articleInfo ArticleInfo
	if err := Auth(); err != nil {
		return articleInfo, err
	}

	resp, err := geekTimeClient.R().
		SetBody(
			map[string]interface{}{
				"id":                strconv.Itoa(articleID),
				"include_neighbors": true,
				"is_freelyread":     true,
				"reverse":           false,
			}).
		SetResult(&result).
		Post(V1ArticlePath)

	if err != nil {
		return articleInfo, err
	}

	if resp.RawResponse.StatusCode == 451 {
		return articleInfo, pgt.ErrGeekTimeRateLimit
	}

	if result.Code == 0 {
		return ArticleInfo{
			result.Data.ArticleContent,
			result.Data.AudioDownloadURL,
		}, nil
	}

	return articleInfo, ErrGeekTimeAPIBadCode{V3VideoPlayAuthPath, result.Code, ""}
}

// Auth check if current user login is expired or login in another device
func Auth() error {
	var result struct {
		Code int `json:"code"`
	}
	t := fmt.Sprintf("%v", time.Now().Round(time.Millisecond).UnixNano()/(int64(time.Millisecond)/int64(time.Nanosecond)))
	resp, err := accountClient.R().
		SetPathParam("t", t).
		SetResult(&result).
		Get("/serv/v1/user/auth")

	if err != nil {
		return err
	}

	if resp.StatusCode() == 200 {
		if result.Code == 0 {
			return nil
		}
		// result Code -1
		// {\"error\":{\"msg\":\"未登录\",\"code\":-2000}
		return pgt.ErrAuthFailed
	}
	// status code 452
	return pgt.ErrAuthFailed
}

// GetMyClassProduct ...
func GetMyClassProduct(classID int) (Product, error) {
	var p Product
	var resp response.V1MyClassInfoResponse
	_, err := ugeekTimeClient.R().SetBody(
		map[string]interface{}{
			"class_id": classID,
		}).
		SetResult(&resp).
		Post(V1MyClassInfoPath)

	if err != nil {
		return p, err
	}

	if resp.Code != 0 {
		if resp.Error.Code == -5001 {
			p.Access = false
			return p, nil
		}
		return p, ErrGeekTimeAPIBadCode{V1VideoPlayAuthPath, resp.Code, ""}
	}

	p = Product{
		Access: true,
		ID:     classID,
		Title:  resp.Data.Title,
		Type:   ProductTypeUniversityVideo,
	}
	var articles []Article
	for _, lesson := range resp.Data.Lessons {
		for _, article := range lesson.Articles {
			// ONLY download university video lessons
			if article.VideoTime > 0 {
				articles = append(articles, Article{
					AID:   article.ArticleID,
					Title: article.ArticleTitle,
				})
			}
		}
	}
	p.Articles = articles

	return p, nil
}

// PostV1VideoPlayAuth ...
func PostV1VideoPlayAuth(articleID, classID int) (response.V1VideoPlayAuthResponse, error) {
	var result response.V1VideoPlayAuthResponse
	_, err := ugeekTimeClient.R().SetBody(
		map[string]interface{}{
			"article_id": articleID,
			"class_id":   classID,
		}).
		SetResult(&result).
		Post(V1VideoPlayAuthPath)

	if err != nil {
		return result, err
	}

	if result.Code == 0 {
		return result, nil
	}
	
	return result, ErrGeekTimeAPIBadCode{V1VideoPlayAuthPath, result.Code, ""}
}

// PostV3ProductInfo only used to get daily lesson product info temporarily
func PostV3ProductInfo(productID int) (response.V3ProductInfoResponse, error) {
	var result response.V3ProductInfoResponse
	if err := Auth(); err != nil {
		return result, err
	}
	resp, err := geekTimeClient.R().
		SetBody(
			map[string]interface{}{
				"id": productID,
			}).
		SetResult(&result).
		Post(V3ProductInfoPath)

	if err != nil {
		return result, err
	}

	if resp.RawResponse.StatusCode == 451 {
		return result, pgt.ErrGeekTimeRateLimit
	}

	if result.Code == 0 {
		return result, nil
	}

	return result, ErrGeekTimeAPIBadCode{V3ProductInfoPath, result.Code, ""}
}

// PostV3ArticleInfo only used to get daily lesson article info temporarily
func PostV3ArticleInfo(articleID int) (response.V3ArticleInfoResponse, error) {
	var result response.V3ArticleInfoResponse
	if err := Auth(); err != nil {
		return result, err
	}
	resp, err := geekTimeClient.R().
		SetBody(
			map[string]interface{}{
				"id": articleID,
			}).
		SetResult(&result).
		Post(V3ArticleInfoPath)

	if err != nil {
		return result, err
	}

	if resp.RawResponse.StatusCode == 451 {
		return result, pgt.ErrGeekTimeRateLimit
	}

	if result.Code == 0 {
		return result, nil
	}

	return result, ErrGeekTimeAPIBadCode{V3ArticleInfoPath, result.Code, ""}
}

// PostV3VideoPlayAuth get play auth string
func PostV3VideoPlayAuth(articleID, sourceType int, videoID string) (string, error) {
	var result struct {
		Code int `json:"code"`
		Data struct {
			PlayAuth string `json:"play_auth"`
		} `json:"data"`
	}
	if err := Auth(); err != nil {
		return "", err
	}
	resp, err := geekTimeClient.R().
		SetBody(
			map[string]interface{}{
				"aid":         articleID,
				"source_type": sourceType,
				"video_id":    videoID,
			}).
		SetResult(&result).
		Post(V3VideoPlayAuthPath)

	if err != nil {
		return "", err
	}

	if resp.RawResponse.StatusCode == 451 {
		return "", pgt.ErrGeekTimeRateLimit
	}

	if result.Code == 0 {
		return result.Data.PlayAuth, nil
	}

	return "", ErrGeekTimeAPIBadCode{V3VideoPlayAuthPath, result.Code, ""}
}
