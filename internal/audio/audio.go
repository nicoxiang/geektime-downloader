package audio

import (
	"context"
	"os"
	"path/filepath"

	"github.com/cavaliergopher/grab/v3"
	"github.com/nicoxiang/geektime-downloader/internal/geektime"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/filenamify"
)

const (
	// MP3Extension ...
	MP3Extension = ".mp3"
)

// DownloadAudio ...
func DownloadAudio(ctx context.Context, grabClient *grab.Client, downloadAudioURL, dir, title string) error {
	if downloadAudioURL == "" {
		return nil
	}
	filenamifyTitle := filenamify.Filenamify(title)

	dst := filepath.Join(dir, filenamifyTitle + MP3Extension)
	request, _ := grab.NewRequest(dst, downloadAudioURL)
	request = request.WithContext(ctx)
	request.HTTPRequest.Header.Set(geektime.Origin, geektime.DefaultBaseURL)

	resp := grabClient.Do(request)
	
	err := resp.Err()
	if err != nil {
		fullName := filepath.Join(dir, filenamifyTitle + MP3Extension)
		_ = os.Remove(fullName)
	}

	return err
}
