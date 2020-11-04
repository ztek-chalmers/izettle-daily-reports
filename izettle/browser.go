package izettle

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

type BrowersClient struct {
	httpClient *http.Client
}

func BrowserLogin(session string) *BrowersClient {
	// Create cookie jar to store cookies which are set by later requests
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}
	cookieURL, err := url.Parse("https://my.izettle.com")
	jar.SetCookies(cookieURL, []*http.Cookie{{
		Name: "_izsessionat", Value: session,
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
