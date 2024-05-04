package config

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nicoxiang/geektime-downloader/internal/geektime"
)

const (
	// GeektimeDownloaderFolder app config and download root dolder name
	GeektimeDownloaderFolder = "geektime-downloader"
)

var userConfigDir string

func init() {
	userConfigDir, _ = os.UserConfigDir()
}

// ReadCookieFromConfigFile read cookies from app config file.
func ReadCookieFromConfigFile(phone string) ([]*http.Cookie, error) {
	dir := filepath.Join(userConfigDir, GeektimeDownloaderFolder)
	files, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	if len(files) == 0 {
		return nil, nil
	}
	now := time.Now()
	for _, fi := range files {
		if fi.IsDir() {
			continue
		}
		if strings.HasPrefix(fi.Name(), phone) {
			fullName := filepath.Join(userConfigDir, GeektimeDownloaderFolder, fi.Name())
			var cookies []*http.Cookie

			data, err := os.ReadFile(fullName)
			if err != nil {
				return nil, err
			}

			for _, line := range strings.Split(string(data), "\n") {
				s := strings.SplitN(line, " ", 3)
				if len(s) != 3 {
					continue
				}
				t, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", s[2])
				if err != nil || t.Before(now) {
					break
				}
				cookies = append(cookies, &http.Cookie{
					Name:     s[0],
					Value:    s[1],
					Domain:   geektime.GeekBangCookieDomain,
					HttpOnly: true,
					Expires:  t,
				})
			}
			return cookies, nil
		}
	}
	return nil, nil
}

// WriteCookieToConfigFile write cookies to config file with specified phone prefix file name.
func WriteCookieToConfigFile(phone string, cookies []*http.Cookie) error {
	dir := filepath.Join(userConfigDir, GeektimeDownloaderFolder)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	_ = removeConfig(dir, phone)

	file, err := os.CreateTemp(dir, phone)
	if err != nil {
		return err
	}
	defer file.Close()
	var sb strings.Builder
	for _, v := range cookies {
		sb = writeOnelineConfig(sb, v)
	}
	if _, err := file.Write([]byte(sb.String())); err != nil {
		return err
	}
	return nil
}

// RemoveConfig remove specified users' config
func RemoveConfig(phone string) error {
	dir := filepath.Join(userConfigDir, GeektimeDownloaderFolder)
	return removeConfig(dir, phone)
}

func removeConfig(dir, phone string) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return nil
	}
	for _, fi := range files {
		if fi.IsDir() {
			continue
		}
		if strings.HasPrefix(fi.Name(), phone) {
			fullName := filepath.Join(userConfigDir, GeektimeDownloaderFolder, fi.Name())
			if err := os.Remove(fullName); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeOnelineConfig(sb strings.Builder, cookie *http.Cookie) strings.Builder {
	sb.WriteString(cookie.Name)
	sb.WriteString(" ")
	sb.WriteString(cookie.Value)
	sb.WriteString(" ")
	sb.WriteString(cookie.Expires.String())
	sb.WriteString("\n")
	return sb
}
