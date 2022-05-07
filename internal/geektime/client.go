package geektime

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	pgt "github.com/nicoxiang/geektime-downloader/internal/pkg/geektime"
)

const (
	// UserAgentHeaderName ...
	UserAgentHeaderName = "User-Agent"
	// OriginHeaderName ...
	OriginHeaderName = "Origin"
	// UserAgent is Web browser User Agent
	UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.92 Safari/537.36"
	// ProductPath ...
	ProductPath = "/serv/v3/learn/product"
	// ArticlesPath ...
	ArticlesPath = "/serv/v1/column/articles"
	// ArticleInfoPath ...
	ArticleInfoPath = "/serv/v3/article/info"
)

var (
	geekTimeClient *resty.Client
	accountClient  *resty.Client
	// ErrAuthFailed ...
	ErrAuthFailed = errors.New("当前账户在其他设备登录或者登录已经过期, 请尝试重新登录")
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
	return fmt.Sprintf("make geektime api call %s failed, code %d, msg %s", e.Path, e.Code, e.Msg)
}

// Product ...
type Product struct {
	ID         int
	Title      string
	AuthorName string
	Type       string
	Articles   []Article
}

// ArticleSummary ...
type Article struct {
	AID   int
	Title string
}

// VideoInfo ...
type VideoInfo struct {
	M3U8URL string
	Size    int
}

func InitClient(cookies []*http.Cookie) {
	geekTimeClient = resty.New().
		SetBaseURL(pgt.GeekBang).
		SetCookies(cookies).
		SetRetryCount(1).
		SetTimeout(10*time.Second).
		SetHeader(UserAgentHeaderName, UserAgent).
		SetHeader(OriginHeaderName, pgt.GeekBang)

	accountClient = resty.New().
		SetBaseURL(pgt.GeekBangAccount).
		SetCookies(cookies).
		SetTimeout(10*time.Second).
		SetHeader(UserAgentHeaderName, UserAgent).
		SetHeader(OriginHeaderName, pgt.GeekBang)

	SiteCookies = cookies
}

// GetProductList call geektime api to get product list
func GetProductList() ([]Product, error) {
	ok, err := auth()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrAuthFailed
	}
	var products []Product
	products, err = appendProducts(0, products)
	if err != nil {
		return nil, err
	}
	return products, nil
}

// GetArticles call geektime api to get article list
func GetArticles(cid string) ([]Article, error) {
	ok, err := auth()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrAuthFailed
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
	_, err = geekTimeClient.R().
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

// GetVideoInfo call geektime api to get video info
func GetVideoInfo(articleID int, quality string) (VideoInfo, error) {
	var videoInfo VideoInfo
	ok, err := auth()
	if err != nil {
		return videoInfo, err
	}
	if !ok {
		return videoInfo, ErrAuthFailed
	}

	var result struct {
		Code int `json:"code"`
		Data struct {
			Info struct {
				ID    int    `json:"id"`
				Title string `json:"title"`
				Video struct {
					HLSVideos []struct {
						Size    int    `json:"size"`
						Quality string `json:"quality"`
						URL     string `json:"url"`
					} `json:"hls_medias"`
				} `json:"video"`
			} `json:"info"`
		} `json:"data"`
	}
	_, err = geekTimeClient.R().
		SetBody(
			map[string]interface{}{
				"id": articleID,
			}).
		SetResult(&result).
		Post(ArticleInfoPath)

	if err != nil {
		return videoInfo, err
	}

	if result.Code == 0 {
		for _, v := range result.Data.Info.Video.HLSVideos {
			if v.Quality == quality {
				return VideoInfo{
					v.URL,
					v.Size,
				}, nil
			}
		}
	}

	return videoInfo, ErrGeekTimeAPIBadCode{ArticleInfoPath, result.Code, ""}
}

// auth check if current user login is expired or login in another device
func auth() (bool, error) {
	var result struct {
		Code int `json:"code"`
	}
	t := fmt.Sprintf("%v", time.Now().Round(time.Millisecond).UnixNano()/(int64(time.Millisecond)/int64(time.Nanosecond)))
	resp, err := accountClient.R().
		SetPathParam("t", t).
		SetResult(&result).
		Get("/serv/v1/user/auth")

	if err != nil {
		return false, err
	}

	if resp.StatusCode() == 200 {
		if result.Code == 0 {
			return true, nil
		}
		// result Code -1
		// {\"error\":{\"msg\":\"未登录\",\"code\":-2000}
		return false, nil
	}
	// status code 452
	return false, nil
}

func appendProducts(prev int, products []Product) ([]Product, error) {
	var result struct {
		Code int `json:"code"`
		Data struct {
			List []struct {
				Score int `json:"score"`
			} `json:"list"`
			Products []struct {
				ID     int    `json:"id"`
				Title  string `json:"title"`
				Author struct {
					Name string `json:"name"`
				} `json:"author"`
				Type string `json:"type"`
			} `json:"products"`
			Page struct {
				More bool `json:"more"`
			} `json:"page"`
		} `json:"data"`
		Error struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
		} `json:"error"`
	}
	_, err := geekTimeClient.R().
		SetBody(
			map[string]interface{}{
				"desc":             false,
				"expire":           1,
				"last_learn":       0,
				"learn_status":     0,
				"prev":             prev,
				"size":             20,
				"sort":             1,
				"type":             "",
				"with_learn_count": 1,
			}).
		SetResult(&result).
		Post(ProductPath)

	if err != nil {
		return nil, err
	}

	if result.Code == 0 {
		for _, v := range result.Data.Products {
			// For now we can only download column and video
			if v.Type == "c1" || v.Type == "c3" {
				products = append(products, Product{
					ID:         v.ID,
					Title:      v.Title,
					AuthorName: v.Author.Name,
					Type:       v.Type,
				})
			}
		}
		if result.Data.Page.More {
			score := result.Data.List[0].Score
			products, err = appendProducts(score, products)
			if err != nil {
				return nil, err
			}
		}
		return products, nil
	}

	return nil, ErrGeekTimeAPIBadCode{ProductPath, result.Code, ""}
}
