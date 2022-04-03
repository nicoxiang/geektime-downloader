package geektime

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/nicoxiang/geektime-downloader/internal/client"
)

var ErrAuthFailed = errors.New("当前账户在其他设备登录, 请尝试重新登录")

func Auth(cookies []*http.Cookie) bool {
	t := fmt.Sprintf("%v", time.Now().Round(time.Millisecond).UnixNano()/(int64(time.Millisecond)/int64(time.Nanosecond)))
	resp, err := client.NewTimeGeekAccountRestyClient().R().SetCookies(cookies).SetPathParam("t", t).Get("/serv/v1/user/auth")

	if err != nil {
		panic(err)
	}

	if resp.StatusCode() == 200 {
		return true
	} else {
		// 452
		return false
	}
}
