package chromedp

import (
	"context"
	"io/ioutil"
	"log"
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
func PrintArticlePageToPDF(aid int, filename string, cookies []*http.Cookie) error {
	var buf []byte
	ctx, cancel := chromedp.NewContext(
		context.Background(),
		chromedp.WithLogf(log.Printf),
	)
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	err := chromedp.Run(ctx,
		chromedp.Tasks{
			chromedp.Emulate(device.IPadPro11),
			setCookies(cookiesToMap(cookies)),
			navigateAndWaitFor(pgt.GeekBang+`/column/article/`+strconv.Itoa(aid), "networkIdle"),
			// chromedp.Navigate(pgt.GeekBang + `/column/article/` + strconv.Itoa(aid)),
			// chromedp.ActionFunc(func(ctx context.Context) error {
			// 	time.Sleep(time.Second * 5)
			// 	return nil
			// }),
			chromedp.ActionFunc(func(ctx context.Context) error {
				s := `
					var divs = document.getElementsByClassName('openApp');
					for (var i = 0; i < divs.length; ++i){
						if(divs[i].innerText === "打开APP"){
							divs[i].parentNode.parentNode.style.display="none";
							break;
						}
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
			}),
			chromedp.ActionFunc(func(ctx context.Context) error {
				var err error
				buf, _, err = page.PrintToPDF().WithPrintBackground(true).Do(ctx)
				return err
			}),
		},
	)

	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, buf, os.ModePerm)
}

func setCookies(cookies map[string]string) chromedp.ActionFunc {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		expr := cdp.TimeSinceEpoch(time.Now().Add(180 * 24 * time.Hour))

		for key, value := range cookies {
			err := network.SetCookie(key, value).WithExpires(&expr).WithDomain(pgt.GeekBangCookieDomain).WithHTTPOnly(true).Do(ctx)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func cookiesToMap(cookies []*http.Cookie) map[string]string {
	cookieMap := make(map[string]string, len(cookies))
	for _, c := range cookies {
		cookieMap[c.Name] = c.Value
	}
	return cookieMap
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
