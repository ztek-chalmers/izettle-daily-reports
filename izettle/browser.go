package izettle

import (
	"bytes"
	"context"
	"fmt"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"

	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

type BrowersClient struct {
	httpClient *http.Client
}

func BrowserLoginEmail(email, password string) (*BrowersClient, string, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	taskCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	session := ""
	loginURL := fmt.Sprintf("https://login.izettle.com/login?username=%s", email)
	tasks := chromedp.Tasks{
		chromedp.Navigate(loginURL),
		chromedp.WaitVisible(`#password`, chromedp.ByID),
		chromedp.SendKeys(`#password`, password, chromedp.ByID),
		chromedp.Click("#submitBtn"),
		chromedp.WaitVisible(`.dashboard`, chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			cookies, err := network.GetAllCookies().Do(ctx)
			if err != nil {
				return err
			}
			for i, cookie := range cookies {
				log.Printf("chrome cookie %d: %+v", i, cookie)
				if cookie.Name == "_izsessionat" {
					session = cookie.Value
					return nil
				}
			}
			return fmt.Errorf("Cookie _izsessionat not found")
		}),
	}

	// ensure that the browser process is started
	if err := chromedp.Run(taskCtx, tasks); err != nil {
		return nil, "", err
	}

	return BrowserLoginCookie(session), session, nil
}

func BrowserLoginCookie(cookie string) *BrowersClient {
	// Create cookie jar to store cookies which are set by later requests
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}
	cookieURL, err := url.Parse("https://my.izettle.com")
	if err != nil {
		panic(err)
	}
	jar.SetCookies(cookieURL, []*http.Cookie{{
		Name: "_izsessionat", Value: cookie,
	}})

	client := &http.Client{Jar: jar}
	return &BrowersClient{httpClient: client}
}

func (i *BrowersClient) IsLoggedIn() bool {
	url := fmt.Sprintf("https://my.izettle.com/dashboard")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false
	}
	client := http.Client{
		Jar: i.httpClient.Jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	if strings.Contains(resp.Header.Get("Location"), "login.izettle.com") {
		return false
	}
	return true
}

func (i *BrowersClient) DayReportToPDF(report Report) (io.Reader, error) {
	date := report.Date.String()
	pdfURL := fmt.Sprintf("https://my.izettle.com/reports.pdf?user=%d&aggregation=day&date=%s&type=pdf", report.UserID, date)
	resp, err := i.httpClient.Get(pdfURL)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(body), nil
}
