package m3u8

import (
	"bufio"
	"regexp"
	"strings"

	"github.com/nicoxiang/geektime-downloader/internal/geektime"
)

var (
	// regex pattern for extracting `key=value` parameters from a line
	linePattern = regexp.MustCompile(`([a-zA-Z-]+)=("[^"]+"|[^",]+)`)
)

// Parse do m3u8 url GET request, and extract ts file names and check if it's encrypt video
func Parse(client *geektime.Client, m3u8url string) (tsFileNames []string, isVodEncryptVideo bool, err error) {
	m3u8Resp, err := client.RestyClient.R().SetDoNotParseResponse(true).Get(m3u8url)
	if err != nil {
		return nil, false, err
	}
	defer m3u8Resp.RawBody().Close()
	s := bufio.NewScanner(m3u8Resp.RawBody())
	var lines []string
	for s.Scan() {
		lines = append(lines, s.Text())
	}

	gotKeyURI := false

	for _, line := range lines {
		// geektime video ONLY has one EXT-X-KEY
		if strings.HasPrefix(line, "#EXT-X-KEY") && !gotKeyURI {
			// ONLY Method and URI, IV not present
			params := parseLineParameters(line)
			isVodEncryptVideo, gotKeyURI = params["MEATHOD"] == "AES-128", true
		}
		if !strings.HasPrefix(line, "#") && strings.HasSuffix(line, ".ts") {
			tsFileNames = append(tsFileNames, line)
		}
	}

	return
}

// parseLineParameters extra parameters in string `line`
func parseLineParameters(line string) map[string]string {
	r := linePattern.FindAllStringSubmatch(line, -1)
	params := make(map[string]string)
	for _, arr := range r {
		params[arr[1]] = strings.Trim(arr[2], "\"")
	}
	return params
}
