package chromedp

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/device"
	pgt "github.com/nicoxiang/geektime-downloader/internal/pkg/geektime"
)

// PrintArticlePageToPDF use chromedp to print article page and save
func PrintArticlePageToPDF(ctx context.Context, aid int, filename string, cookies []*http.Cookie) {
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	var buf []byte
	err := chromedp.Run(ctx,
		chromedp.Tasks{
			chromedp.Emulate(device.IPadPro11),
			setCookies(cookies),
			navigateAndWaitFor(pgt.GeekBang+`/column/article/`+strconv.Itoa(aid), "networkIdle"),
			hideRedundantElements(),
			printToPDF(&buf),
		},
	)

	if err != nil {
		if !errors.Is(err, context.Canceled) {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}

	if err := ioutil.WriteFile(filename, buf, os.ModePerm); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func setCookies(cookies []*http.Cookie) chromedp.ActionFunc {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		expr := cdp.TimeSinceEpoch(time.Now().Add(180 * 24 * time.Hour))

		for _, c := range cookies {
			err := network.SetCookie(c.Name, c.Value).WithExpires(&expr).WithDomain(pgt.GeekBangCookieDomain).WithHTTPOnly(true).Do(ctx)
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
				var openAppdiv = document.getElementsByClassName('openApp')[0];
				if(openAppdiv){
					openAppdiv.parentNode.parentNode.style.display="none";
				}
				var audioBarDiv = document.getElementsByClassName('audio-float-bar')[0];
				if(audioBarDiv){
					audioBarDiv.style.display="none";
				}
				var leadsMobileDiv = document.getElementsByClassName('leads mobile')[0];
				if(leadsMobileDiv){
					leadsMobileDiv.style.display="none";
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

func printToPDF(res *[]byte) chromedp.ActionFunc {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		data, _, err := page.PrintToPDF().WithPrintBackground(false).Do(ctx)
				if err != nil {
					return err
				}
				*res = data
				return nil
	})
} 

func navigateAndWaitFor(url string, eventName string) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		_, _, _, err := page.Navigate(url).Do(ctx)
		if err != nil {
			return err
		}

		return waitFor(ctx, eventName)
	}
}

// waitFor blocks until eventName is received.
// Examples of events you can wait for:
//     init, DOMContentLoaded, firstPaint,
//     firstContentfulPaint, firstImagePaint,
//     firstMeaningfulPaintCandidate,
//     load, networkAlmostIdle, firstMeaningfulPaint, networkIdle
//
// This is not super reliable, I've already found incidental cases where
// networkIdle was sent before load. It's probably smart to see how
// puppeteer implements this exactly.
func waitFor(ctx context.Context, eventName string) error {
	ch := make(chan struct{})
	cctx, cancel := context.WithCancel(ctx)
	chromedp.ListenTarget(cctx, func(ev interface{}) {
		switch e := ev.(type) {
		case *page.EventLifecycleEvent:
			if e.Name == eventName {
				cancel()
				close(ch)
			}
		}
	})

	select {
	case <-ch:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
