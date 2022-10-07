package geektime

import (
	"fmt"
	"net/http"
	"reflect"
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

	// V3ColumnInfoPath used in get normal column/video info
	V3ColumnInfoPath = "/serv/v3/column/info"
	// V3ProductInfoPath used in get daily lesson, qconplus product info
	V3ProductInfoPath = "/serv/v3/product/info"
	// V3ArticleInfoPath used in normal video, daily lesson, qconplus
	V3ArticleInfoPath = "serv/v3/article/info"
	// V3VideoPlayAuthPath used in normal video, daily lesson, qconplus video play auth
	V3VideoPlayAuthPath = "serv/v3/source_auth/video_play_auth"

	// UniversityV1VideoPlayAuthPath used in university video play auth
	UniversityV1VideoPlayAuthPath = "/serv/v1/video/play-auth"
	// UniversityV1MyClassInfoPath get university class info and all articles info in it
	UniversityV1MyClassInfoPath = "/serv/v1/myclass/info"
)

var (
	geekTimeClient  *resty.Client
	ugeekTimeClient *resty.Client
	// SiteCookies ...
	SiteCookies []*http.Cookie
)

// ErrGeekTimeAPIBadCode ...
type ErrGeekTimeAPIBadCode struct {
	Path string
	ResponseString string
}

// Error implements error interface
func (e ErrGeekTimeAPIBadCode) Error() string {
	return fmt.Sprintf("请求极客时间接口 %s 失败", e.Path)
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

	ugeekTimeClient = resty.New().
		SetBaseURL(pgt.GeekBangUniversity).
		SetCookies(cookies).
		SetRetryCount(1).
		SetTimeout(10*time.Second).
		SetHeader(pgt.UserAgentHeaderName, pgt.UserAgentHeaderValue).
		SetHeader(pgt.OriginHeaderName, pgt.GeekBangUniversity).
		SetLogger(logger.DiscardLogger{})

	SiteCookies = cookies
}

// PostV3ColumnInfo get normal column info, like v3 product info
func PostV3ColumnInfo(productID int) (Product, error) {
	var p Product
	result, err := makeAPICall[response.V3ColumnInfoResponse](geekTimeClient, V3ColumnInfoPath,
		map[string]interface{}{
			"product_id":             productID,
			"with_recommend_article": true,
		})

	if err != nil {
		return p, err
	}
	return Product{
		Access: result.Data.Extra.Sub.AccessMask > 0,
		ID:     result.Data.ID,
		Type:   result.Data.Type,
		Title:  result.Data.Title,
	}, nil
}

// PostV1ColumnArticles call geektime api to get article list
func PostV1ColumnArticles(cid string) ([]Article, error) {
	result, err := makeAPICall[response.V1ColumnArticlesResponse](geekTimeClient, V1ColumnArticlesPath,
		map[string]interface{}{
			"cid":    cid,
			"order":  "earliest",
			"prev":   0,
			"sample": false,
			"size":   500, //get all articles
		})

	if err != nil {
		return nil, err
	}

	var articles []Article
	for _, v := range result.Data.List {
		articles = append(articles, Article{
			AID:   v.ID,
			Title: v.ArticleTitle,
		})
	}
	return articles, nil
}

// GetArticleInfo ...
func GetArticleInfo(articleID int) (response.V1ArticleResponse, error) {
	return makeAPICall[response.V1ArticleResponse](geekTimeClient, V1ArticlePath,
		map[string]interface{}{
			"id":                strconv.Itoa(articleID),
			"include_neighbors": true,
			"is_freelyread":     true,
			"reverse":           false,
		})
}

// GetMyClassProduct ...
func GetMyClassProduct(classID int) (Product, error) {
	var p Product
	var result response.V1MyClassInfoResponse
	resp, err := ugeekTimeClient.R().SetBody(
		map[string]interface{}{
			"class_id": classID,
		}).
		SetResult(&result).
		Post(UniversityV1MyClassInfoPath)

	if err != nil {
		return p, err
	}

	if resp.RawResponse.StatusCode == 451 {
		return p, pgt.ErrGeekTimeRateLimit
	} else if resp.RawResponse.StatusCode == 452 {
		return p, pgt.ErrAuthFailed
	}

	if result.Code != 0 {
		if result.Error.Code == -5001 {
			p.Access = false
			return p, nil
		}
		return p, ErrGeekTimeAPIBadCode{UniversityV1MyClassInfoPath, resp.String()}
	}

	p = Product{
		Access: true,
		ID:     classID,
		Title:  result.Data.Title,
		Type:   string(ProductTypeUniversityVideo),
	}
	var articles []Article
	for _, lesson := range result.Data.Lessons {
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

// PostUniversityV1VideoPlayAuth ...
func PostUniversityV1VideoPlayAuth(articleID, classID int) (response.V1VideoPlayAuthResponse, error) {
	return makeAPICall[response.V1VideoPlayAuthResponse](ugeekTimeClient, UniversityV1VideoPlayAuthPath,
		map[string]interface{}{
			"article_id": articleID,
			"class_id":   classID,
		})
}

// PostV3ProductInfo used to get daily lesson or qconplus product info
func PostV3ProductInfo(productID int) (response.V3ProductInfoResponse, error) {
	return makeAPICall[response.V3ProductInfoResponse](geekTimeClient, V3ProductInfoPath,
		map[string]interface{}{
			"id": productID,
		})
}

// PostV3ArticleInfo used to get daily lesson or qconplus article info
func PostV3ArticleInfo(articleID int) (response.V3ArticleInfoResponse, error) {
	return makeAPICall[response.V3ArticleInfoResponse](geekTimeClient, V3ArticleInfoPath,
		map[string]interface{}{
			"id": articleID,
		})
}

// PostV3VideoPlayAuth get play auth string
func PostV3VideoPlayAuth(articleID, sourceType int, videoID string) (string, error) {
	result, err := makeAPICall[response.V3VideoPlayAuthResponse](geekTimeClient, V3VideoPlayAuthPath,
		map[string]interface{}{
			"aid":         articleID,
			"source_type": sourceType,
			"video_id":    videoID,
		})

	if err != nil {
		return "", err
	}

	return result.Data.PlayAuth, nil
}

func makeAPICall[T any](client *resty.Client, url string, body interface{}) (T, error) {
	var result T

	resp, err := client.R().
		SetBody(body).
		SetResult(&result).
		Post(url)

	if err != nil {
		return result, err
	}

	if resp.RawResponse.StatusCode == 451 {
		return result, pgt.ErrGeekTimeRateLimit
	} else if resp.RawResponse.StatusCode == 452 {
		return result, pgt.ErrAuthFailed
	}

	r := reflect.ValueOf(result)
	f := reflect.Indirect(r).FieldByName("Code")
	code := int(f.Int())

	if code == 0 {
		return result, nil
	}

	return result, ErrGeekTimeAPIBadCode{url, resp.String()}
}

// Auth check if current user login is expired or login in another device
func Auth() error {
	var result struct {
		Code int `json:"code"`
	}
	t := fmt.Sprintf("%v", time.Now().Round(time.Millisecond).UnixNano()/(int64(time.Millisecond)/int64(time.Nanosecond)))
	resp, err := resty.New().
		SetBaseURL(pgt.GeekBangAccount).
		SetCookies(SiteCookies).
		SetTimeout(10*time.Second).
		SetHeader(pgt.UserAgentHeaderName, pgt.UserAgentHeaderValue).
		SetHeader(pgt.OriginHeaderName, pgt.GeekBang).
		SetLogger(logger.DiscardLogger{}).
		R().
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
