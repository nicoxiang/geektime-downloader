package geektime

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	pgt "github.com/nicoxiang/geektime-downloader/internal/pkg/geektime"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/logger"
)

const (
	// ProductPath ...
	ProductPath = "/serv/v3/learn/product"
	// ArticlesPath ...
	ArticlesPath = "/serv/v1/column/articles"
	// ArticleV1Path ...
	ArticleV1Path = "/serv/v1/article"
	// ColumnInfoV3Path ...
	ColumnInfoV3Path = "/serv/v3/column/info"
)

var (
	geekTimeClient *resty.Client
	accountClient  *resty.Client
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

// VideoInfo ...
type VideoInfo struct {
	M3U8URL string
	Size    int
}

// ArticleInfo ...
type ArticleInfo struct {
	ArticleContent   string
	AudioDownloadURL string
}

// ColumnResponse ...
type ColumnResponse struct {
	Code int `json:"code"`
	Data struct {
		ArticleTitle     string `json:"article_title"`
		ArticleContent   string `json:"article_content"`
		AudioDownloadURL string `json:"audio_download_url"`
	} `json:"data"`
}

// VideoResponse ...
type VideoResponse struct {
	Code int `json:"code"`
	Data struct {
		ArticleTitle string `json:"article_title"`
		HLSVideos    struct {
			SD struct {
				Size int    `json:"size"`
				URL  string `json:"url"`
			} `json:"sd"`
			HD struct {
				Size int    `json:"size"`
				URL  string `json:"url"`
			} `json:"hd"`
			LD struct {
				Size int    `json:"size"`
				URL  string `json:"url"`
			} `json:"ld"`
		} `json:"hls_videos"`
	} `json:"data"`
}

// ArticleResponse type constraint, column and video response are different,
// hls_videos field in video response is struct, but in column response its slice
type ArticleResponse interface {
	ColumnResponse | VideoResponse
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

	SiteCookies = cookies
}

// GetColumnInfo  ..
func GetColumnInfo(productID int) (Product, error) {
	var p Product
	if err := Auth(); err != nil {
		return p, err
	}
	var result struct {
		Code int `json:"code"`
		Data struct {
			ID    int    `json:"id"`
			Type  string `json:"type"`
			Title string `json:"title"`
			Extra struct {
				Sub struct {
					AccessMask int `json:"access_mask"`
				} `json:"sub"`
			} `json:"extra"`
		} `json:"data"`
	}
	resp, err := geekTimeClient.R().
		SetBody(
			map[string]interface{}{
				"product_id":             productID,
				"with_recommend_article": true,
			}).
		SetResult(&result).
		Post(ColumnInfoV3Path)

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

	return p, ErrGeekTimeAPIBadCode{ColumnInfoV3Path, result.Code, ""}
}

// GetArticles call geektime api to get article list
func GetArticles(cid string) ([]Article, error) {
	if err := Auth(); err != nil {
		return nil, err
	}

	var result struct {
		Code int `json:"code"`
		Data struct {
			List []struct {
				ID           int    `json:"id"`
				ArticleTitle string `json:"article_title"`
			} `json:"list"`
		} `json:"data"`
	}
	resp, err := geekTimeClient.R().
		SetBody(
			map[string]interface{}{
				"cid":    cid,
				"order":  "earliest",
				"prev":   0,
				"sample": false,
				"size":   500,
			}).
		SetResult(&result).
		Post(ArticlesPath)

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

	return nil, ErrGeekTimeAPIBadCode{ArticlesPath, result.Code, ""}
}

// GetColumnArticleInfo ...
func GetColumnArticleInfo(articleID int) (ArticleInfo, error) {
	var a ArticleInfo
	ar, err := GetArticleResponse[ColumnResponse](articleID)
	if err != nil {
		return a, err
	}
	if ar.Code != 0 {
		return a, ErrGeekTimeAPIBadCode{ArticleV1Path, ar.Code, ""}
	}

	return ArticleInfo{
		ar.Data.ArticleContent,
		ar.Data.AudioDownloadURL,
	}, err
}

// GetVideoInfo ...
func GetVideoInfo(articleID int, quality string) (VideoInfo, error) {
	var v VideoInfo
	a, err := GetArticleResponse[VideoResponse](articleID)
	if err != nil {
		return v, err
	}
	if a.Code != 0 {
		return v, ErrGeekTimeAPIBadCode{ArticleV1Path, a.Code, ""}
	}
	if quality == "sd" {
		v = VideoInfo{
			M3U8URL: a.Data.HLSVideos.SD.URL,
			Size:    a.Data.HLSVideos.SD.Size,
		}
	} else if quality == "hd" {
		v = VideoInfo{
			M3U8URL: a.Data.HLSVideos.HD.URL,
			Size:    a.Data.HLSVideos.HD.Size,
		}
	} else if quality == "ld" {
		v = VideoInfo{
			M3U8URL: a.Data.HLSVideos.LD.URL,
			Size:    a.Data.HLSVideos.LD.Size,
		}
	}
	return v, nil
}

// GetArticleResponse get column or video response
func GetArticleResponse[R ArticleResponse](articleID int) (R, error) {
	var result R
	if err := Auth(); err != nil {
		return result, err
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
		Post(ArticleV1Path)

	if err != nil {
		return result, err
	}

	if resp.RawResponse.StatusCode == 451 {
		return result, pgt.ErrGeekTimeRateLimit
	}

	return result, nil
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
