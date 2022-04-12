package file

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	pgt "github.com/nicoxiang/geektime-downloader/internal/pkg/geektime"
)

const (
	// MAXLENGTH Max file name length
	MAXLENGTH = 80
	// GeektimeDownloaderFolder app config and download root dolder name
	GeektimeDownloaderFolder = "geektime-downloader"
)

var userConfigDir string

func init() {
	userConfigDir, _ = os.UserConfigDir()
}

// ByNumericalFilename implement sort interface, order by file name suffix number
type ByNumericalFilename []os.FileInfo

func (nf ByNumericalFilename) Len() int      { return len(nf) }
func (nf ByNumericalFilename) Swap(i, j int) { nf[i], nf[j] = nf[j], nf[i] }
func (nf ByNumericalFilename) Less(i, j int) bool {
	// Use path names
	pathA := nf[i].Name()
	pathB := nf[j].Name()

	// Grab integer value of each filename by parsing the string and slicing off
	// the extension
	a, err1 := strconv.ParseInt(pathA[0:strings.LastIndex(pathA, ".")], 10, 64)
	b, err2 := strconv.ParseInt(pathB[0:strings.LastIndex(pathB, ".")], 10, 64)

	// If any were not numbers sort lexographically
	if err1 != nil || err2 != nil {
		return pathA < pathB
	}

	// Which integer is smaller?
	return a < b
}

// ReadCookieFromConfigFile read cookies from app config file.
func ReadCookieFromConfigFile(phone string) []*http.Cookie {
	dir := filepath.Join(userConfigDir, GeektimeDownloaderFolder)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		panic(err)
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
			var cookies []*http.Cookie
			oneyear := time.Now().Add(180 * 24 * time.Hour)

			data, err := ioutil.ReadFile(fullName)
			if err != nil {
				panic(err)
			}

			for _, line := range strings.Split(string(data), "\n") {
				s := strings.SplitN(line, " ", 2)
				if len(s) != 2 {
					continue
				}
				cookies = append(cookies, &http.Cookie{
					Name:     s[0],
					Value:    s[1],
					Domain:   pgt.GeekBangCookieDomain,
					HttpOnly: true,
					Expires:  oneyear,
				})
			}
			return cookies
		}
	}
	return nil
}

// WriteCookieToConfigFile write cookies to config file with specified phone prefix file name.
func WriteCookieToConfigFile(phone string, cookies []*http.Cookie) {
	dir := filepath.Join(userConfigDir, GeektimeDownloaderFolder)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		panic(err)
	}
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	for _, fi := range files {
		// config file already exists
		if strings.HasPrefix(fi.Name(), phone) {
			return
		}
	}
	file, err := ioutil.TempFile(dir, phone)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	var sb strings.Builder
	for _, v := range cookies {
		sb = writeOnelineConfig(sb, v.Name, v.Value)
	}
	if _, err := file.Write([]byte(sb.String())); err != nil {
		panic(err)
	}
}

// RemoveConfig remove specified users' config
func RemoveConfig(phone string) {
	dir := filepath.Join(userConfigDir, GeektimeDownloaderFolder)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	if len(files) == 0 {
		return
	}
	for _, fi := range files {
		if fi.IsDir() {
			continue
		}
		if strings.HasPrefix(fi.Name(), phone) {
			fullName := filepath.Join(userConfigDir, GeektimeDownloaderFolder, fi.Name())
			if err := os.Remove(fullName); err != nil {
				panic(err)
			}
		}
	}
}

// MkDownloadProjectFolder creates download project directory if not exist
func MkDownloadProjectFolder(downloadFolder, phone, projectName string) string {
	path := filepath.Join(downloadFolder, phone, Filenamify(projectName))
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		panic(err)
	}
	return path
}

// FindDownloadedArticleFileNames find all downloaded articles file name in specified account and column
func FindDownloadedArticleFileNames(projectDir string) map[string]struct{} {
	res := make(map[string]struct{})
	files, err := ioutil.ReadDir(projectDir)
	if err != nil {
		panic(err)
	}
	if len(files) == 0 {
		return res
	}
	for _, f := range files {
		res[f.Name()] = struct{}{}
	}
	return res
}

func writeOnelineConfig(sb strings.Builder, key string, value string) strings.Builder {
	sb.WriteString(key)
	sb.WriteString(" ")
	sb.WriteString(value)
	sb.WriteString("\n")
	return sb
}
