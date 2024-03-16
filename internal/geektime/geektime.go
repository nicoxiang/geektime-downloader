package geektime

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/nicoxiang/geektime-downloader/internal/geektime/response"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/logger"
)

const (
	// DefaultBaseURL ...
	DefaultBaseURL = "https://time.geekbang.org"
	// DefaultUserAgent ...
	DefaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.92 Safari/537.36"
	// Origin ...
	Origin = "Origin"
	// UserAgent ...
	UserAgent = "User-Agent"
	// GeekBangUniversityBaseURL ...
	GeekBangUniversityBaseURL = "https://u.geekbang.org"
	// GeekBangAccountBaseURL ...
	GeekBangAccountBaseURL = "https://account.geekbang.org"
	// LoginPath ...
	LoginPath = "/account/ticket/login"
	// V1AuthPath ...
	V1AuthPath = "/serv/v1/user/auth"
	// V1ColumnArticlesPath get all articles summary info in one column
	V1ColumnArticlesPath = "/serv/v1/column/articles"
	// V1ArticlePath used in normal column
	V1ArticlePath = "/serv/v1/article"

	// V3ColumnInfoPath used in get normal column/video info
	V3ColumnInfoPath = "/serv/v3/column/info"
	// V3ProductInfoPath used in get daily lesson, qconplus product info
	V3ProductInfoPath = "/serv/v3/product/info"
	// V3ArticleInfoPath used in normal video, daily lesson, qconplus
	V3ArticleInfoPath = "/serv/v3/article/info"
	// V3VideoPlayAuthPath used in normal video, daily lesson, qconplus video play auth
	V3VideoPlayAuthPath = "/serv/v3/source_auth/video_play_auth"

	// UniversityV1VideoPlayAuthPath used in university video play auth
	UniversityV1VideoPlayAuthPath = "/serv/v1/video/play-auth"
	// UniversityV1MyClassInfoPath get university class info and all articles info in it
	UniversityV1MyClassInfoPath = "/serv/v1/myclass/info"

	// GeekBangCookieDomain ...
	GeekBangCookieDomain = ".geekbang.org"

	// GCID ...
	GCID = "GCID"
	// GCESS ...
	GCESS = "GCESS"
)

var (
	// ErrWrongPassword ...
	ErrWrongPassword = errors.New("密码错误, 请尝试重新登录")
	// ErrTooManyLoginAttemptTimes ...
	ErrTooManyLoginAttemptTimes = errors.New("密码输入错误次数过多，已触发验证码校验，请稍后再试")
	// ErrGeekTimeRateLimit ...
	ErrGeekTimeRateLimit = errors.New("已触发限流, 你可以选择重新登录/重新获取 cookie, 或者稍后再试, 然后生成剩余的文章")
	// ErrAuthFailed ...
	ErrAuthFailed = errors.New("当前账户在其他设备登录或者登录已经过期, 请尝试重新登录")
)

// Product ...
type Product struct {
	Access   bool
	ID       int
	Title    string
	Type     string
	IsVideo  bool
	Articles []Article
}

// Article ...
type Article struct {
	AID   int
	Title string
}

// ErrGeekTimeAPIBadCode ...
type ErrGeekTimeAPIBadCode struct {
	Path           string
	ResponseString string
}

// Error implements error interface
func (e ErrGeekTimeAPIBadCode) Error() string {
	return fmt.Sprintf("请求极客时间接口 %s 失败, ResponseBody: %s", e.Path, e.ResponseString)
}

// A Client manages communication with the Geektime API.
type Client struct {
	HTTPClient *resty.Client
	BaseURL    string
	Cookies    []*http.Cookie
}

// NewClient returns a new Geektime API client.
func NewClient(cs []*http.Cookie) *Client {
	httpClient := resty.New().
		SetCookies(cs).
		SetRetryCount(1).
		SetTimeout(10*time.Second).
		SetHeader("User-Agent", DefaultUserAgent).
		SetLogger(logger.DiscardLogger{})

	c := &Client{HTTPClient: httpClient, BaseURL: DefaultBaseURL, Cookies: cs}
	return c
}

// NewAccountClient ...
func NewAccountClient() *Client {
	httpClient := resty.New().
		SetRetryCount(1).
		SetTimeout(10*time.Second).
		SetHeader("User-Agent", DefaultUserAgent).
		SetLogger(logger.DiscardLogger{})

	c := &Client{HTTPClient: httpClient, BaseURL: GeekBangAccountBaseURL}
	return c
}

// NewUniversityClient ...
func NewUniversityClient(cs []*http.Cookie) *Client {
	httpClient := resty.New().
		SetCookies(cs).
		SetRetryCount(1).
		SetTimeout(10*time.Second).
		SetHeader("User-Agent", DefaultUserAgent).
		SetLogger(logger.DiscardLogger{})

	c := &Client{HTTPClient: httpClient, BaseURL: GeekBangUniversityBaseURL, Cookies: cs}
	return c
}

// Login call geektime login api and return auth cookies
func (c *Client) Login(phone, password string) ([]*http.Cookie, error) {
	var res struct {
		Code int `json:"code"`
		Data struct {
			UID  int    `json:"uid"`
			Name string `json:"nickname"`
		} `json:"data"`
		Error struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
		} `json:"error"`
	}

	r := c.newRequest(resty.MethodPost,
		LoginPath,
		nil,
		map[string]interface{}{
			"country":   86,
			"appid":     1,
			"platform":  3,
			"cellphone": phone,
			"password":  password,
		},
		&res,
	)

	logger.Infof("Login request start")
	resp, err := r.Execute(r.Method, r.URL)

	if err != nil {
		return nil, err
	}

	if resp.RawResponse.StatusCode != 200 || res.Code != 0 {
		logger.Warnf("Login request end, status code: %d, response body: %s",
			resp.RawResponse.StatusCode,
			resp.String(),
		)
	}

	if res.Code == 0 {
		var cookies []*http.Cookie
		for _, c := range resp.Cookies() {
			if c.Name == GCID || c.Name == GCESS {
				cookies = append(cookies, c)
			}
		}
		return cookies, nil
	} else if res.Error.Code == -3031 {
		return nil, ErrWrongPassword
	} else if res.Error.Code == -3005 {
		return nil, ErrTooManyLoginAttemptTimes
	}
	return nil, ErrGeekTimeAPIBadCode{LoginPath, resp.String()}
}

// Auth check if current user login is expired or login in another device
func (c *Client) Auth(cs []*http.Cookie) error {
	var res struct {
		Code int `json:"code"`
	}
	t := fmt.Sprintf("%v", time.Now().Round(time.Millisecond).UnixNano()/(int64(time.Millisecond)/int64(time.Nanosecond)))
	c.HTTPClient.SetCookies(cs)
	params := make(map[string]string, 2)
	params["t"] = t
	params["v_t"] = t
	r := c.newRequest(resty.MethodGet,
		V1AuthPath,
		params,
		nil,
		res,
	)
	r.SetHeader(Origin, DefaultBaseURL)

	logger.Infof("Auth request start")
	resp, err := r.Execute(r.Method, r.URL)

	if err != nil {
		return err
	}

	if resp.RawResponse.StatusCode != 200 || res.Code != 0 {
		logger.Warnf("Auth request end, status code: %d, response body: %s",
			resp.RawResponse.StatusCode,
			resp.String(),
		)

		// result Code -1
		// {\"error\":{\"msg\":\"未登录\",\"code\":-2000}
		return ErrAuthFailed
	}

	return nil
}

// ColumnInfo get normal column info, like v3 product info
func (c *Client) ColumnInfo(productID int) (Product, error) {
	var res response.V3ColumnInfoResponse
	r := c.newRequest(resty.MethodPost,
		V3ColumnInfoPath,
		nil,
		map[string]interface{}{
			"product_id":             productID,
			"with_recommend_article": true,
		},
		&res,
	)
	if _, err := do(r); err != nil {
		return Product{}, err
	}

	return Product{
		Access:  res.Data.Extra.Sub.AccessMask > 0,
		ID:      res.Data.ID,
		Type:    res.Data.Type,
		Title:   res.Data.Title,
		IsVideo: res.Data.IsVideo,
	}, nil
}

// ColumnArticles call geektime api to get article list
func (c *Client) ColumnArticles(cid string) ([]Article, error) {
	res := &response.V1ColumnArticlesResponse{}
	r := c.newRequest(resty.MethodPost,
		V1ColumnArticlesPath,
		nil,
		map[string]interface{}{
			"cid":    cid,
			"order":  "earliest",
			"prev":   0,
			"sample": false,
			"size":   500, //get all articles
		},
		res,
	)
	if _, err := do(r); err != nil {
		return nil, err
	}

	var articles []Article
	for _, v := range res.Data.List {
		articles = append(articles, Article{
			AID:   v.ID,
			Title: v.ArticleTitle,
		})
	}
	return articles, nil
}

// V1ArticleInfo ...
func (c *Client) V1ArticleInfo(articleID int) (response.V1ArticleResponse, error) {
	var res response.V1ArticleResponse
	r := c.newRequest(resty.MethodPost,
		V1ArticlePath,
		nil,
		map[string]interface{}{
			"id":                strconv.Itoa(articleID),
			"include_neighbors": true,
			"is_freelyread":     true,
			"reverse":           false,
		},
		&res,
	)
	if _, err := do(r); err != nil {
		return response.V1ArticleResponse{}, err
	}
	return res, nil
}

// ProductInfo used to get daily lesson or qconplus product info
func (c *Client) ProductInfo(productID int) (response.V3ProductInfoResponse, error) {
	var res response.V3ProductInfoResponse
	r := c.newRequest(resty.MethodPost,
		V3ProductInfoPath,
		nil,
		map[string]interface{}{
			"id": productID,
		},
		&res,
	)
	if _, err := do(r); err != nil {
		return response.V3ProductInfoResponse{}, err
	}
	return res, nil
}

// V3ArticleInfo used to get daily lesson or qconplus article info
func (c *Client) V3ArticleInfo(articleID int) (response.V3ArticleInfoResponse, error) {
	var res response.V3ArticleInfoResponse
	r := c.newRequest(resty.MethodPost,
		V3ArticleInfoPath,
		nil,
		map[string]interface{}{
			"id": articleID,
		},
		&res,
	)
	if _, err := do(r); err != nil {
		return response.V3ArticleInfoResponse{}, err
	}
	return res, nil
}

// VideoPlayAuth get play auth string
func (c *Client) VideoPlayAuth(articleID, sourceType int, videoID string) (string, error) {
	var res response.V3VideoPlayAuthResponse
	r := c.newRequest(resty.MethodPost,
		V3VideoPlayAuthPath,
		nil,
		map[string]interface{}{
			"aid":         articleID,
			"source_type": sourceType,
			"video_id":    videoID,
		},
		&res,
	)
	if _, err := do(r); err != nil {
		return "", err
	}
	return res.Data.PlayAuth, nil
}

// UniversityVideoPlayAuth ...
func (c *Client) UniversityVideoPlayAuth(articleID, classID int) (response.V1VideoPlayAuthResponse, error) {
	var res response.V1VideoPlayAuthResponse
	r := c.newRequest(resty.MethodPost,
		UniversityV1VideoPlayAuthPath,
		nil,
		map[string]interface{}{
			"article_id": articleID,
			"class_id":   classID,
		},
		&res,
	)
	if _, err := do(r); err != nil {
		return response.V1VideoPlayAuthResponse{}, err
	}
	return res, nil
}

// MyClassProduct ...
func (c *Client) MyClassProduct(classID int) (Product, error) {
	var p Product

	var res response.V1MyClassInfoResponse
	r := c.newRequest(resty.MethodPost,
		UniversityV1MyClassInfoPath,
		nil,
		map[string]interface{}{
			"class_id": classID,
		},
		&res,
	)

	resp, err := do(r)
	if err != nil {
		return p, err
	}

	if res.Code != 0 {
		if res.Error.Code == -5001 {
			p.Access = false
			return p, nil
		}
		return p, ErrGeekTimeAPIBadCode{UniversityV1MyClassInfoPath, resp.String()}
	}

	p = Product{
		Access:  true,
		ID:      classID,
		Title:   res.Data.Title,
		Type:    "",
		IsVideo: true,
	}
	var articles []Article
	for _, lesson := range res.Data.Lessons {
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

func (c *Client) newRequest(method, url string, params map[string]string, body interface{}, res interface{}) *resty.Request {
	r := c.HTTPClient.R()
	r.Method = method
	r.URL = c.BaseURL + url
	r.SetHeader(Origin, c.BaseURL)
	if len(params) > 0 {
		r.SetQueryParams(params)
	}
	if body != nil {
		r.SetBody(body)
	}
	r.SetResult(res)
	return r
}

func do(r *resty.Request) (*resty.Response, error) {
	logger.Infof("Http request start, method: %s, url: %s",
		r.Method,
		r.URL,
	)
	resp, err := r.Execute(r.Method, r.URL)

	if err != nil {
		return nil, err
	}

	statusCode := resp.RawResponse.StatusCode
	if statusCode != 200 {
		logNotOkResponse(resp)
		if statusCode == 451 {
			return nil, ErrGeekTimeRateLimit
		} else if statusCode == 452 {
			return nil, ErrAuthFailed
		}
	}

	rv := reflect.ValueOf(r.Result)
	f := reflect.Indirect(rv).FieldByName("Code")
	code := int(f.Int())

	if code == 0 {
		return resp, nil
	}

	logNotOkResponse(resp)
	//未登录或者已失效
	if code == -3050 || code == -2000 {
		return nil, ErrAuthFailed
	}

	return nil, ErrGeekTimeAPIBadCode{r.URL, resp.String()}
}

func logNotOkResponse(resp *resty.Response) {
	logger.Warnf("Http request end, method: %s, url: %s, status code: %d, response body: %s",
		resp.RawResponse.Request.Method,
		resp.RawResponse.Request.URL,
		resp.RawResponse.StatusCode,
		resp.String(),
	)
}
