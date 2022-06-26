package audio

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/filenamify"
	pgt "github.com/nicoxiang/geektime-downloader/internal/pkg/geektime"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/logger"
)

const (
	// MP3Extension ...
	MP3Extension = ".mp3"
)

// DownloadAudio ...
func DownloadAudio(ctx context.Context, downloadAudioURL, dir, title string) error {
	filenamifyTitle := filenamify.Filenamify(title)
	c := resty.New()
	c.SetOutputDirectory(dir).
		SetRetryCount(1).
		SetTimeout(time.Minute).
		SetHeader(pgt.UserAgentHeaderName, pgt.UserAgentHeaderValue).
		SetHeader(pgt.OriginHeaderName, pgt.GeekBang).
		SetLogger(logger.DiscardLogger{})

	_, err := c.R().
		SetContext(ctx).
		SetOutput(filenamifyTitle + MP3Extension).
		Get(downloadAudioURL)

	if errors.Is(err, context.Canceled) {
		fullName := filepath.Join(dir, filenamifyTitle + MP3Extension)
		_ = os.Remove(fullName)
	}

	return err
}
