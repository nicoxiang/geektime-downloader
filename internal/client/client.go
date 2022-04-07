package client

import (
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	pgt "github.com/nicoxiang/geektime-downloader/internal/pkg/geektime"
)

// UserAgent is Web browser User Agent
const UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.92 Safari/537.36"

var simpleClient *resty.Client
var noParseResponseClient *resty.Client
var accountClient *resty.Client

// NewTimeGeekRestyClient new Http Client with user auth cookies
func NewTimeGeekRestyClient(cookies []*http.Cookie) *resty.Client {
	return resty.New().
		SetTimeout(10*time.Second).
		SetHeader("User-Agent", UserAgent).
		SetHeader("Origin", pgt.GeekBang).
		SetBaseURL(pgt.GeekBang).
		SetCookies(cookies).
		SetRetryCount(1).
		SetRetryWaitTime(500 * time.Millisecond).
		SetRetryMaxWaitTime(2 * time.Second)
}

// NewTimeGeekAccountRestyClient new Http Client with geekbang account domain
func NewTimeGeekAccountRestyClient() *resty.Client {
	if accountClient == nil {
		accountClient = resty.New().
			SetTimeout(10*time.Second).
			SetHeader("User-Agent", UserAgent).
			SetHeader("Origin", pgt.GeekBang).
			SetBaseURL(pgt.GeekBangAccount)
	}
	return accountClient
}

// NewNoParseResponseRestyClient new Http Client with SetDoNotParseResponse option, used for download video
func NewNoParseResponseRestyClient() *resty.Client {
	if noParseResponseClient == nil {
		noParseResponseClient = resty.New().
			SetTimeout(30*time.Second).
			SetHeader("User-Agent", UserAgent).
			SetHeader("Origin", pgt.GeekBang).
			SetDoNotParseResponse(true)
	}
	return noParseResponseClient
}

// New ...
func New() *resty.Client {
	if simpleClient == nil {
		simpleClient = resty.New().
			SetTimeout(10*time.Second).
			SetHeader("User-Agent", UserAgent).
			SetHeader("Origin", pgt.GeekBang).
			// Set retry count to non zero to enable retries
			SetRetryCount(1).
			// You can override initial retry wait time.
			// Default is 100 milliseconds.
			SetRetryWaitTime(500 * time.Millisecond).
			// MaxWaitTime can be overridden as well.
			// Default is 2 seconds.
			SetRetryMaxWaitTime(2 * time.Second)
	}
	return simpleClient
}
