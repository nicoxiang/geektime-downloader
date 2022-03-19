package util

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	pgt "github.com/nicoxiang/geektime-downloader/internal/pkg/geektime"
)

const (
	// Max file name length
	MAXLENGTH                = 80
	// App config and download root dolder name
	GeektimeDownloaderFolder = "geektime-downloader"
	// Expire key in config file 
	ExpireConfigLineKey      = "EXPIRE"
	// Expire value layout in config file
	ExpireLayout             = "Mon, 02 Jan 2006 15:04:05 -0700"
)

var userConfigDir string

func init() {
	userConfigDir, _ = os.UserConfigDir()
}

// Convert a string to a valid safe filename
func FileName(name string, ext string) string {
	rep := strings.NewReplacer("\n", " ", "/", " ", "|", "-", ": ", "：", ":", "：", "'", "’", "\t", " ")
	name = rep.Replace(name)

	if runtime.GOOS == "windows" {
		rep := strings.NewReplacer("\"", " ", "?", " ", "*", " ", "\\", " ", "<", " ", ">", " ", ":", " ", "：", " ")
		name = rep.Replace(name)
	}

	name = strings.TrimSpace(name)

	limitedName := limitLength(name, MAXLENGTH)
	if ext != "" {
		return fmt.Sprintf("%s.%s", limitedName, ext)
	}
	return limitedName
}

func limitLength(s string, length int) string {
	ellipses := "..."

	if str := []rune(s); len(str) > length {
		s = string(str[:length-len(ellipses)]) + ellipses
	}

	return s
}

// Read cookies from app config file, if cookie has expired, delete old config file.
func ReadCookieFromConfigFile(phone string) ([]*http.Cookie, error) {
	dir := filepath.Join(userConfigDir, GeektimeDownloaderFolder)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	if len(files) == 0 {
		return nil, nil
	}
	for _, fi := range files {
		if fi.IsDir() {
			continue
		}
		if strings.HasPrefix(fi.Name(), phone) {
			fullName := filepath.Join(userConfigDir, GeektimeDownloaderFolder, fi.Name())
			var cookies []*http.Cookie
			oneyear := time.Now().Add(180 * 24 * time.Hour)

			data, err := ioutil.ReadFile(fullName)
			if err != nil {
				return nil, err
			}

			for _, line := range strings.Split(string(data), "\n") {
				s := strings.SplitN(line, " ", 2)
				if len(s) != 2 {
					continue
				}
				if s[0] == ExpireConfigLineKey && !checkExpire(s[1]) {
					err := os.Remove(fullName)
					return nil, err
				}
				cookies = append(cookies, &http.Cookie{
					Name:     s[0],
					Value:    s[1],
					Domain:   pgt.GeekBangCookieDomain,
					HttpOnly: true,
					Expires:  oneyear,
				})
			}
			return cookies, nil
		}
	}
	return nil, nil
}

// Write cookies to config file with specified phone prefix file name, 
// and write cookie 'GCESS' expire date into config too. 
func WriteCookieToConfigFile(phone string, cookies []*http.Cookie) error {
	dir := filepath.Join(userConfigDir, GeektimeDownloaderFolder)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, fi := range files {
		// config file already exists
		if strings.HasPrefix(fi.Name(), phone) {
			return nil
		}
	}
	file, err := ioutil.TempFile(dir, phone)
	if err != nil {
		return err
	}
	defer file.Close()
	var sb strings.Builder
	for _, v := range cookies {
		if v.Name == "GCESS" {
			sb = writeOnelineConfig(sb, ExpireConfigLineKey, v.Expires.Format(ExpireLayout))
		}
		sb = writeOnelineConfig(sb, v.Name, v.Value)
	}
	if _, err := file.Write([]byte(sb.String())); err != nil {
		return err
	}
	return nil
}

// Remove specified users' config
func RemoveConfig(phone string) error {
	dir := filepath.Join(userConfigDir, GeektimeDownloaderFolder)
	files, err := ioutil.ReadDir(dir)
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

func checkExpire(value string) bool {
	expire, err := time.Parse(ExpireLayout, value)
	if err != nil {
		return false
	}
	if time.Now().After(expire) {
		return false
	}
	return true
}

func writeOnelineConfig(sb strings.Builder, key string, value string) strings.Builder {
	sb.WriteString(key)
	sb.WriteString(" ")
	sb.WriteString(value)
	sb.WriteString("\n")
	return sb
}
