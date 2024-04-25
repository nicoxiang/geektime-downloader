package geektime

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/logger"
)

const (
	// GeekBangAccountBaseURL ...
	GeekBangAccountBaseURL = "https://account.geekbang.org"
	// LoginPath ...
	LoginPath = "/account/ticket/login"
	// V1AuthPath ...
	V1AuthPath = "/serv/v1/user/auth"
)

// Login call geektime login api and return auth cookies
func Login(phone, password string) ([]*http.Cookie, error) {
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

	client := resty.New().
		SetTimeout(DefaultTimeout).
		SetHeader(UserAgent, DefaultUserAgent).
		SetLogger(logger.DiscardLogger{})

	logger.Infof("Login request start")

	resp, err := client.R().
		SetHeader(Origin, DefaultBaseURL).
		SetBody(map[string]interface{}{
			"country":   86,
			"appid":     1,
			"platform":  3,
			"cellphone": phone,
			"password":  password,
		}).
		SetResult(&res).
		Post(GeekBangAccountBaseURL + LoginPath)

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
func Auth(cs []*http.Cookie) error {
	var res struct {
		Code int `json:"code"`
	}
	t := fmt.Sprintf("%v", time.Now().Round(time.Millisecond).UnixNano()/(int64(time.Millisecond)/int64(time.Nanosecond)))
	params := make(map[string]string, 2)
	params["t"] = t
	params["v_t"] = t

	client := resty.New().
		SetTimeout(DefaultTimeout).
		SetHeader(UserAgent, DefaultUserAgent).
		SetLogger(logger.DiscardLogger{})

	logger.Infof("Auth request start")

	resp, err := client.R().
		SetQueryParams(params).
		SetCookies(cs).
		SetHeader(Origin, DefaultBaseURL).
		SetResult(&res).
		Get(GeekBangAccountBaseURL + V1AuthPath)

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
