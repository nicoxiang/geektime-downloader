package audio

import (
	"context"
	"os"
	"path/filepath"

	"github.com/nicoxiang/geektime-downloader/internal/geektime"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/downloader"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/filenamify"
)

const (
	// MP3Extension ...
	MP3Extension = ".mp3"
)

// DownloadAudio ...
func DownloadAudio(ctx context.Context, downloadAudioURL, dir, title string) error {
	if downloadAudioURL == "" {
		return nil
	}
	filenamifyTitle := filenamify.Filenamify(title)

	dst := filepath.Join(dir, filenamifyTitle+MP3Extension)

	headers := make(map[string]string, 2)
	headers[geektime.Origin] = geektime.DefaultBaseURL
	headers[geektime.UserAgent] = geektime.DefaultUserAgent

	_, err := downloader.DownloadFileConcurrently(ctx, dst, downloadAudioURL, headers, 1)

	if err != nil {
		_ = os.Remove(dst)
	}

	return err
}
