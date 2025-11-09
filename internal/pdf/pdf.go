package pdf

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/device"

	"github.com/nicoxiang/geektime-downloader/internal/geektime"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/filenamify"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/logger"
)

// PDFExtension ...
const PDFExtension = ".pdf"

// DownloadCommentsMode values for PrintArticlePageToPDF
const (
	DownloadCommentsNone = iota
	DownloadCommentsFirstPage
	DownloadCommentsAll
)

type V4CommentListResponse struct {
	Code int `json:"code"`
	Data struct {
		Page struct {
			More  bool `json:"more"`
			Count int  `json:"count"`
		} `json:"page"`
	} `json:"data"`
	Error struct{} `json:"error"`
	Extra struct {
		Cost      float64 `json:"cost"`
		RequestID string  `json:"request-id"`
	} `json:"extra"`
}

// PrintArticlePageToPDF use chromedp to print article page and save
func PrintArticlePageToPDF(parentCtx context.Context,
	aid int,
	dir,
	title string,
	cookies []*http.Cookie,
	downloadComments int,
	printPDFWaitSeconds int,
	printPDFTimeoutSeconds int,
) error {
	rateLimit := false

	pdfFileName := filepath.Join(dir, filenamify.Filenamify(title)+PDFExtension)

	ctx, cancel := chromedp.NewContext(parentCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, time.Duration(printPDFTimeoutSeconds)*time.Second)
	defer cancel()

	var commentsDone uint32 = 0

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch responseReceivedEvent := ev.(type) {
		case *network.EventResponseReceived:
			response := responseReceivedEvent.Response
			// rate limit detection
			if response.URL == geektime.DefaultBaseURL+"/serv/v1/article" && response.Status == 451 {
				logger.Warnf("Hit GeekTime rate limit when downloading article pdf, articleID: %d, pdfFileName: %s", aid, pdfFileName)
				rateLimit = true
				cancel()
				return
			}

			// if downloadComments is DownloadCommentsAll, monitor comment list responses
			if downloadComments == DownloadCommentsAll {
				if strings.Contains(strings.ToLower(response.URL), "comment/list") {
					reqID := responseReceivedEvent.RequestID
					url := response.URL

					fetchAndHandleCommentList(ctx, reqID, url, &commentsDone)
				}
			}
		}
	})

	tasks := chromedp.Tasks{
		network.Enable(),
		chromedp.Emulate(device.IPadPro11),
		setCookies(cookies),
		chromedp.Navigate(geektime.DefaultBaseURL + `/column/article/` + strconv.Itoa(aid)),
		chromedp.Sleep(time.Duration(printPDFWaitSeconds) * time.Second),
	}

	switch downloadComments {
	case DownloadCommentsAll:
		tasks = append(tasks, touchScrollAction(&commentsDone))
	case DownloadCommentsNone:
		tasks = append(tasks, hideCommentsBlock())
	}

	tasks = append(tasks, hideRedundantElements(), printToPDF(pdfFileName))

	logger.Infof("Begin download article pdf, articleID: %d, pdfFileName: %s", aid, pdfFileName)

	err := chromedp.Run(ctx, tasks)
	if err != nil {
		if rateLimit {
			logger.Warnf("Hit GeekTime rate limit when downloading article pdf, articleID: %d, pdfFileName: %s", aid, pdfFileName)
			return geektime.ErrGeekTimeRateLimit
		}
		logger.Errorf(err, "Failed to download article pdf")
		return err
	}

	return nil
}

func setCookies(cookies []*http.Cookie) chromedp.ActionFunc {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		expr := cdp.TimeSinceEpoch(time.Now().Add(180 * 24 * time.Hour))

		for _, c := range cookies {
			err := network.SetCookie(c.Name, c.Value).WithExpires(&expr).WithDomain(geektime.GeekBangCookieDomain).WithHTTPOnly(true).Do(ctx)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func hideRedundantElements() chromedp.ActionFunc {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		s := `
			var headMain = document.getElementsByClassName('main')[0];
   			if(headMain){
      				headMain.style.display="none";
			}
   			var bottomWrapper = document.getElementsByClassName('sub-bottom-wrapper')[0];
   			if(bottomWrapper){
      				bottomWrapper.style.display="none";
			}
			var openAppdiv = document.getElementsByClassName('openApp')[0];
			if(openAppdiv){
				openAppdiv.parentNode.parentNode.parentNode.style.display="none";
			}
			var audioPlayer = document.querySelector('div[class^="ColumnArticleMiniAudioPlayer"]');
			if(audioPlayer){
				audioPlayer.style.display="none"
			}
			var audioFloatBar = document.querySelector('div[class*="audio-float-bar"]');
			if(audioFloatBar){
				audioFloatBar.style.display="none"
			}
			var leadsWrapper = document.querySelector('div[class^="leads-wrapper"]');
			if(leadsWrapper){
				leadsWrapper.style.display="none";
			}
			var unPreviewImage = document.querySelector('img[alt="unpreview"]');
			if(unPreviewImage){
				unPreviewImage.style.display="none"
			}
			var gotoColumn = document.querySelector('div[class^="Index_articleColumn"]');
			if(gotoColumn){
				gotoColumn.style.display="none"
			}
			var favBtn = document.querySelector('div[class*="Index_favBtn"]');
			if(favBtn){
				favBtn.style.display="none"
			}
			var likeModule = document.querySelector('div[class^="ArticleLikeModuleMobile"]');
			if(likeModule){
				likeModule.style.display="none"
			}
			var switchBtns = document.querySelector('div[class^="Index_switchBtns"]');
			if(switchBtns){
				switchBtns.style.display="none"
			}
			var writeComment = document.querySelector('div[class*="Index_writeComment"]');
			if(writeComment){
				writeComment.style.display="none"
			}
			var moreBtns = document.querySelectorAll('div[class^=CommentItem_more]');
			for (let btn of moreBtns) {
				btn.click();
			}
		`

		_, exp, err := runtime.Evaluate(s).Do(ctx)
		if err != nil {
			return err
		}

		if exp != nil {
			return exp
		}

		return nil
	})
}

func hideCommentsBlock() chromedp.ActionFunc {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		hideCommentsExpression := `
			var comments = document.querySelector('div[class^="Index_articleComments"]')
			if(comments){
				comments.style.display="none"
			}
		`
		_, exp, err := runtime.Evaluate(hideCommentsExpression).Do(ctx)
		if err != nil {
			return err
		}

		if exp != nil {
			return exp
		}

		return nil
	})
}

func printToPDF(fileName string) chromedp.ActionFunc {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		_, stream, err := page.PrintToPDF().
			WithMarginTop(0.4).
			WithMarginBottom(0.4).
			WithMarginLeft(0.4).
			WithMarginRight(0.4).
			WithTransferMode(page.PrintToPDFTransferModeReturnAsStream).
			Do(ctx)
		if err != nil {
			return err
		}

		reader := &streamReader{
			ctx:    ctx,
			handle: stream,
			r:      nil,
			pos:    0,
			eof:    false,
		}

		defer func() {
			_ = reader.Close()
		}()

		// open file truncating any existing content so we always write from start
		file, err := os.Create(fileName)
		if err != nil {
			return err
		}

		defer func() {
			_ = file.Close()
		}()

		buffer := bufio.NewReader(reader)

		_, err = buffer.WriteTo(file)
		if err != nil {
			return err
		}

		logger.Infof("Finish download article pdf, pdfFileName: %s", fileName)

		return nil
	})
}

// touchScrollAction returns a chromedp.Action that performs repeated touch
// scroll gestures until commentsDone is set or the action context is canceled.
func touchScrollAction(commentsDone *uint32) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		for {
			if atomic.LoadUint32(commentsDone) == 1 {
				return nil
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			logger.Infof("Performing touch scroll to load more comments")

			c := chromedp.FromContext(ctx)

			ec := cdp.WithExecutor(ctx, c.Target)

			_ = input.DispatchTouchEvent(input.TouchStart, []*input.TouchPoint{{X: 300, Y: 800}}).Do(ec)

			_ = input.DispatchTouchEvent(input.TouchMove, []*input.TouchPoint{{X: 300, Y: 400}}).Do(ec)

			_ = input.DispatchTouchEvent(input.TouchEnd, []*input.TouchPoint{}).Do(ec)

			time.Sleep(500 * time.Millisecond)
		}
	})
}

// fetchAndHandleCommentList fetches the response body for a comment list request
// and updates commentsDone when there are no more pages. It runs its work in a
// separate goroutine so callers can call it without blocking the ListenTarget
// callback.
func fetchAndHandleCommentList(ctx context.Context, reqID network.RequestID, url string, commentsDone *uint32) {
	go func() {
		logger.Infof("Begin fetching comment list: %s", url)

		c := chromedp.FromContext(ctx)

		body, err := network.GetResponseBody(reqID).Do(cdp.WithExecutor(ctx, c.Target))
		if err != nil {
			logger.Errorf(err, "Failed to get comment response body")
			return
		}

		var commentResp V4CommentListResponse
		if err := json.Unmarshal(body, &commentResp); err != nil {
			logger.Errorf(err, "Failed to unmarshal comment response body")
			return
		}

		if commentResp.Code == 0 {
			logger.Infof("Article hasMoreComments=%v", commentResp.Data.Page.More)
			if !commentResp.Data.Page.More {
				logger.Infof("All comments have been loaded, preparing to generate PDF")
				atomic.StoreUint32(commentsDone, 1)
			}
		} else {
			logger.Errorf(nil, "Failed to fetch comment list, response code: %d", commentResp.Code)
		}
	}()
}
