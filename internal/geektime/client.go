package geektime

import (
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	pgt "github.com/nicoxiang/geektime-downloader/internal/pkg/geektime"
)

// UserAgent is Web browser User Agent
const UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.92 Safari/537.36"

// NewTimeGeekRestyClient new Http Client with user auth cookies
func NewTimeGeekRestyClient(cookies []*http.Cookie) *resty.Client {
	return resty.New().
		SetTimeout(30*time.Second).
		SetHeader("User-Agent", UserAgent).
		SetHeader("Origin", pgt.GeekBang).
		SetBaseURL(pgt.GeekBang).
		SetCookies(cookies)
}
