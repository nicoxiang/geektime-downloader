package pdf

import (
	"bufio"
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/chromedp/cdproto/cdp"
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

// PrintArticlePageToPDF use chromedp to print article page and save
func PrintArticlePageToPDF(ctx context.Context,
	aid int,
	dir,
	title string,
	cookies []*http.Cookie,
	downloadComments bool,
	printPDFWaitSeconds int,
	printPDFTimeoutSeconds int,
) error {
	rateLimit := false

	pdfFileName := filepath.Join(dir, filenamify.Filenamify(title)+PDFExtension)

	// new tab
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, time.Duration(printPDFTimeoutSeconds)*time.Second)
	defer cancel()

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch responseReceivedEvent := ev.(type) {
		case *network.EventResponseReceived:
			response := responseReceivedEvent.Response
			if response.URL == geektime.DefaultBaseURL+"/serv/v1/article" && response.Status == 451 {
				logger.Warnf("Hit GeekTime rate limit when downloading article pdf, articleID: %d, pdfFileName: %s", aid, pdfFileName)
				rateLimit = true
				cancel()
			}
		}
	})

	logger.Infof("Begin download article pdf, articleID: %d, pdfFileName: %s", aid, pdfFileName)
	err := chromedp.Run(ctx,
		chromedp.Tasks{
			chromedp.Emulate(device.IPadPro11),
			setCookies(cookies),
			chromedp.Navigate(geektime.DefaultBaseURL + `/column/article/` + strconv.Itoa(aid)),
			chromedp.Sleep(time.Duration(printPDFWaitSeconds) * time.Second),
			hideRedundantElements(downloadComments),
			printToPDF(pdfFileName),
		},
	)
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

func hideRedundantElements(downloadComments bool) chromedp.ActionFunc {
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

		hideCommentsExpression := `
			var comments = document.querySelector('div[class^="Index_articleComments"]')
			if(comments){
				comments.style.display="none"
			}
		`
		if !downloadComments {
			s = s + hideCommentsExpression
		}

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

		file, _ := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, 0o666)

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
