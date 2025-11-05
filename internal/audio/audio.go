package audio

import (
	"context"
	"os"
	"path/filepath"

	"github.com/nicoxiang/geektime-downloader/internal/geektime"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/downloader"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/filenamify"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/logger"
)

const (
	// MP3Extension ...
	MP3Extension = ".mp3"
)

// DownloadAudio ...
func DownloadAudio(ctx context.Context, downloadAudioURL, dir, title string) error {
	logger.Infof("Begin download article audio, title: %s", title)
	if downloadAudioURL == "" {
		return nil
	}
	audioFileName := filepath.Join(dir, filenamify.Filenamify(title)+MP3Extension)

	headers := make(map[string]string, 2)
	headers[geektime.Origin] = geektime.DefaultBaseURL
	headers[geektime.UserAgent] = geektime.DefaultUserAgent

	_, err := downloader.DownloadFileConcurrently(ctx, audioFileName, downloadAudioURL, headers, 1)
	if err != nil {
		logger.Errorf(err, "Failed to download article audio, title: %s", title)
		_ = os.Remove(audioFileName)
	}
	logger.Infof("Finish download article audio, title: %s", title)
	return err
}
