package config

import (
	"net/http"
	"time"

	"github.com/nicoxiang/geektime-downloader/internal/geektime"
)

const (
	// GeektimeDownloaderFolder app config and download root dolder name
	GeektimeDownloaderFolder = "geektime-downloader"
)

type AppConfig struct {
	Gcid                   string
	Gcess                  string
	DownloadFolder         string
	Quality                string
	DownloadComments       int
	ColumnOutputType       int
	PrintPDFWaitSeconds    int
	PrintPDFTimeoutSeconds int
	Interval               int
	IsEnterprise           bool
}

func ReadCookiesFromInput(cfg *AppConfig) []*http.Cookie {
	oneyear := time.Now().Add(180 * 24 * time.Hour)
	cookies := make([]*http.Cookie, 2)
	m := make(map[string]string, 2)
	m[geektime.GCID] = cfg.Gcid
	m[geektime.GCESS] = cfg.Gcess
	c := 0
	for k, v := range m {
		cookies[c] = &http.Cookie{
			Name:     k,
			Value:    v,
			Domain:   geektime.GeekBangCookieDomain,
			HttpOnly: true,
			Expires:  oneyear,
		}
		c++
	}
	return cookies
}
