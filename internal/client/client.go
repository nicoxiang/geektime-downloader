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
var accountClient *resty.Client

// NewTimeGeekRestyClient new Http Client with user auth cookies
func NewTimeGeekRestyClient(cookies []*http.Cookie) *resty.Client {
	c := resty.New().
		SetBaseURL(pgt.GeekBang).
		SetCookies(cookies).
		SetRetryCount(1)

	SetGeekBangHeaders(c)
	return c
}

// NewTimeGeekAccountRestyClient new Http Client with geekbang account domain
func NewTimeGeekAccountRestyClient() *resty.Client {
	if accountClient == nil {
		accountClient = resty.New().
			SetBaseURL(pgt.GeekBangAccount)

		SetGeekBangHeaders(accountClient)
	}
	return accountClient
}

// New ...
func New() *resty.Client {
	if simpleClient == nil {
		simpleClient = resty.New().		
			SetRetryCount(1)

		SetGeekBangHeaders(simpleClient)
	}
	return simpleClient
}

func SetGeekBangHeaders(client *resty.Client) {
	client.
		SetHeader("User-Agent", UserAgent).
		SetHeader("Origin", pgt.GeekBang).
		SetTimeout(10 * time.Second)
}
