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

	"golang.org/x/sync/errgroup"

	"github.com/nicoxiang/geektime-downloader/internal/pkg/logger"
)

// Part use this struct as a channel type.
// The struct will have 2 fields. One for housing data and the other specifying its location within the final file.
type Part struct {
	Data  []byte
	Index int
	Offset int64
}

// DownloadFileConcurrently download file in chunks, return total file size
func DownloadFileConcurrently(ctx context.Context, filepath string, url string, headers map[string]string, concurrency int) (int64, error) {
	// Use HEAD with context so it can be cancelled by parent ctx (Ctrl+C)
	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return 0, err
	}
	// propagate headers (if any)
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	if resp.Body != nil {
		_ = resp.Body.Close()
	}

	fileSize, _ := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	if fileSize <= 0 {
		out, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o666)
		if err != nil {
			return 0, err
		}
		defer func() {
			_ = out.Close()
		}()
		return fileSize, nil
	}

	if concurrency <= 0 {
		concurrency = 1
	}
	if int64(concurrency) > fileSize {
		concurrency = int(fileSize)
	}

	g, ctx := errgroup.WithContext(ctx)

	results := make(chan Part, concurrency)

	chunkSize := fileSize / int64(concurrency)

	for i := 0; i < concurrency; i++ {
		i := i
		g.Go(func() error {
			return download(ctx, concurrency, i, chunkSize, url, results)
		})
	}

	errCh := make(chan error, 1)
	go func() {
		err := g.Wait()
		close(results)
		errCh <- err
		close(errCh)
	}()

	out, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o666)
	if err != nil {
		return 0, err
	}
	removeOnError := false
	defer func() {
		_ = out.Close()
		if removeOnError {
			_ = os.Remove(filepath)
		}
	}()

	nextIndex := 0
	pending := make(map[int]Part, concurrency)
	var writeErr error

	for part := range results {
		if writeErr != nil {
			continue
		}
		pending[part.Index] = part
		for {
			nextPart, ok := pending[nextIndex]
			if !ok {
				break
			}
			_, err := out.WriteAt(nextPart.Data, nextPart.Offset)
			if err != nil {
				writeErr = err
				break
			}
			delete(pending, nextIndex)
			nextIndex++
		}
	}

	if writeErr != nil {
		removeOnError = true
		return 0, writeErr
	}

	if err := <-errCh; err != nil {
		removeOnError = true
		return 0, err
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
	err = retry(ctx, 3, 700*time.Millisecond, func() error {
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer func() {
			_ = resp.Body.Close()
		}()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		c <- Part{Index: index, Offset: start, Data: body}
		return nil
	})
	if err != nil {
		return err
	}

	return err
}

func retry(ctx context.Context, attempts int, sleep time.Duration, f func() error) (err error) {
	for i := 0; i < attempts; i++ {
		if i > 0 {
			// backoff but allow cancellation
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(sleep):
			}
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
