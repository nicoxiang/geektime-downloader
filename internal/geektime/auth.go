package geektime

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/nicoxiang/geektime-downloader/internal/client"
)

// ErrAuthFailed ...
var ErrAuthFailed = errors.New("当前账户在其他设备登录或者登录已经过期, 请尝试重新登录")

// Auth check if current user login is expired or login in another device
func Auth(cookies []*http.Cookie) bool {
	t := fmt.Sprintf("%v", time.Now().Round(time.Millisecond).UnixNano()/(int64(time.Millisecond)/int64(time.Nanosecond)))
	resp, err := client.NewTimeGeekAccountRestyClient().R().SetCookies(cookies).SetPathParam("t", t).Get("/serv/v1/user/auth")

	if err != nil {
		panic(err)
	}

	if resp.StatusCode() == 200 {
		return true
	}
	// 452
	return false
}
