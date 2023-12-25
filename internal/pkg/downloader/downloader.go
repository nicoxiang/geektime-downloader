package downloader

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/nicoxiang/geektime-downloader/internal/pkg/logger"
	"golang.org/x/sync/errgroup"
)

// Part use this struct as a channel type.
// The struct will have 2 fields. One for housing data and the other specifying its location within the final file.
type Part struct {
	Data  []byte
	Index int
}

// DownloadFileConcurrently download file in chunks, return total file size
func DownloadFileConcurrently(ctx context.Context, filepath string, url string, headers map[string]string, concurrency int) (int64, error) {
	resp, err := http.Head(url)
	if err != nil {
		return 0, err
	}

	fileSize, _ := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)

	g, ctx := errgroup.WithContext(ctx)

	results := make(chan Part, concurrency)

	chunkSize := fileSize / int64(concurrency)

	for i := 0; i < concurrency; i++ {
		i := i
		g.Go(func() error {
			return download(ctx, concurrency, i, chunkSize, url, results)
		})
	}

	go func() {
		err = g.Wait()
		close(results)
	}()

	out, err := os.OpenFile(filepath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return 0, err
	}
	defer out.Close()

	counter := 0
	var parts = make([][]byte, concurrency)
	for part := range results {
		counter++

		parts[part.Index] = part.Data
		if counter == concurrency {
			break
		}
	}

	if err != nil {
		return 0, err
	}

	for _, part := range parts {
		_, err = out.Write(part)
		if err != nil {
			return 0, err
		}
	}

	return fileSize, nil
}

func download(ctx context.Context, workers int, index int, chunkSize int64, url string, c chan Part) error {

	// calculate offset by multiplying
	// index with size
	start := int64(index) * chunkSize

	// Write data range in correct format
	// I'm reducing one from the end size to account for
	// the next chunk starting there
	dataRange := fmt.Sprintf("bytes=%d-%d", start, start+chunkSize-1)

	// if this is downloading the last chunk
	// rewrite the header. It's an easy way to specify
	// getting the rest of the file
	if index == workers-1 {
		dataRange = fmt.Sprintf("bytes=%d-", start)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)

	if err != nil {
		return err
	}

	req.Header.Add("Range", dataRange)

	// fix error: http2: server sent GOAWAY and closed the connection; LastStreamID=1999
	// error comes from io read, not request
	err = retry(3, 700*time.Millisecond, func() error {
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)

		if err != nil {
			return err
		}
		c <- Part{Index: index, Data: body}
		return nil
	})

	if err != nil {
		return err
	}

	return err
}

func retry(attempts int, sleep time.Duration, f func() error) (err error) {
	for i := 0; i < attempts; i++ {
		if i > 0 {
			time.Sleep(sleep)
			sleep *= 2

			logger.Infof("retry hanppen, times: %s", strconv.Itoa(i))
		}
		err = f()
		if err == nil || errors.Is(err, context.Canceled) {
			return err
		}
	}
	return fmt.Errorf("after %d attempts, last error: %s", attempts, err)
}
