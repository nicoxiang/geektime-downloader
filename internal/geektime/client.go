package geektime

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/logger"
)

const (
	DefaultTimeout = 10 * time.Second
	// Origin ...
	Origin = "Origin"
	// UserAgent ...
	UserAgent = "User-Agent"
	// DefaultUserAgent ...
	DefaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.92 Safari/537.36"
)

// A Client manages communication with the Geektime API.
type Client struct {
	RestyClient *resty.Client
	Cookies     []*http.Cookie
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

// NewClient returns a new Geektime API client.
func NewClient(cs []*http.Cookie) *Client {
	restyClient := resty.New().
		SetCookies(cs).
		SetRetryCount(1).
		SetTimeout(DefaultTimeout).
		SetHeader(UserAgent, DefaultUserAgent).
		SetLogger(logger.DiscardLogger{})

	c := &Client{RestyClient: restyClient, Cookies: cs}
	return c
}

// newRequest new http request
func (c *Client) newRequest(
	method string,
	baseURL string,
	path string,
	params map[string]string,
	body interface{},
	result interface{}) *resty.Request {
	r := c.RestyClient.R()
	r.Method = method
	r.URL = baseURL + path
	r.SetHeader(Origin, baseURL)
	if len(params) > 0 {
		r.SetQueryParams(params)
	}
	if body != nil {
		r.SetBody(body)
	}
	r.SetResult(result)
	return r
}

// do perform http request
func do(request *resty.Request) (*resty.Response, error) {
	logger.Infof("Http request start, method: %s, url: %s, request body: %v",
		request.Method,
		request.URL,
		request.Body,
	)
	resp, err := request.Execute(request.Method, request.URL)

	if err != nil {
		return nil, err
	}

	statusCode := resp.RawResponse.StatusCode

	logger.Infof("Http request end, method: %s, url: %s, status code: %d",
		resp.RawResponse.Request.Method,
		resp.RawResponse.Request.URL,
		resp.RawResponse.StatusCode,
	)

	if statusCode != 200 {
		logNotOkResponse(resp)
		if statusCode == 451 {
			return nil, ErrGeekTimeRateLimit
		} else if statusCode == 452 {
			return nil, ErrAuthFailed
		}
	}

	rv := reflect.ValueOf(request.Result)
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

	return nil, ErrGeekTimeAPIBadCode{request.URL, resp.String()}
}

func logNotOkResponse(resp *resty.Response) {
	logger.Warnf("Http request not ok, response body: %s", resp.String())
}
