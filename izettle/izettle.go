package izettle

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Client struct {
	httpClient *http.Client
}

func Login(user, password string) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Jar: jar,
	}
	loginURL := "https://login.izettle.com/login?username=" + url.QueryEscape(user)
	resp, err := client.Get(loginURL)
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return nil, err
	}
	// Find the review items
	n := doc.Find("[name=_csrf]").First()
	token, _ := n.Attr("value")

	form := url.Values{}
	form.Add("_csrf", token)
	form.Add("username", user)
	form.Add("password", password)
	form.Add("button", "")

	resp, err = client.PostForm(loginURL, form)
	if err != nil {
		return nil, err
	}
	_ = resp.Body.Close()
	host, _ := url.Parse("https://izettle.com")
	if len(jar.Cookies(host)) == 0 {
		return nil, fmt.Errorf("failed to login")
	}
	return &Client{
		httpClient: client,
	}, nil
}

type User struct {
	ID   string
	Name string
}

func (i *Client) ListUsers() (map[string]User, error) {
	resp, err := i.httpClient.Get("https://my.izettle.com/settings/staff")
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return nil, err
	}
	n := doc.Find("[name=csrf-param]").First()
	param, _ := n.Attr("content")
	n = doc.Find("[name=csrf-token]").First()
	token, _ := n.Attr("content")

	staffURL := fmt.Sprintf("https://my.izettle.com/settings/staff?filter=accepted&%s=%s", param, url.QueryEscape(token))
	req, _ := http.NewRequest("GET", staffURL, nil)
	req.Header.Add("X-CSRF-Token", token)
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	resp, err = i.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	doc, err = goquery.NewDocumentFromReader(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return nil, err
	}
	users := make(map[string]User)
	nodes := doc.Find(".user")
	if len(nodes.Nodes) == 0 {
		return nil, fmt.Errorf("could not find any staff")
	}
	for _, n := range nodes.Nodes {
		nameValue := doc.FindNodes(n).Find(".name").Nodes[0].FirstChild.Data
		idValue, _ := doc.FindNodes(n).Find(".button").Attr("data-overlay-from")
		name := strings.ToLower(strings.Split(nameValue, " ")[0])
		id := strings.Split(idValue, "-")[1]
		users[name] = User{
			ID:   id,
			Name: name,
		}
	}

	return users, nil
}

type Report struct {
	User User
	From string
	To   string
}

func (i *Client) ListReports(user User) ([]Report, error) {
	resp, err := i.httpClient.Get("https://my.izettle.com/reports/summary?user=" + user.ID)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return nil, err
	}
	var data map[string]map[string][]map[string]*json.RawMessage
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}
	var reports []Report
	months := data["daily"]
	for _, month := range months {
		for _, day := range month {
			startTime, ok := day["aggregateStart"]
			if !ok {
				return nil, fmt.Errorf("report summary does not have a timestamp")
			}
			start := strings.Split(string((*startTime)[1:]), "T")[0]
			endTime, ok := day["aggregateEnd"]
			if !ok {
				return nil, fmt.Errorf("report summary does not have a timestamp")
			}
			end := strings.Split(string((*endTime)[1:]), "T")[0]
			reports = append(reports, Report{
				User: user,
				From: start,
				To:   end,
			})
		}
	}

	return reports, nil
}

func (i *Client) getAuthToken() (string, error) {
	host, _ := url.Parse("https://izettle.com")
	cookies := i.httpClient.Jar.Cookies(host)
	for _, c := range cookies {
		if c.Name == "_izsessionat" {
			return c.Value, nil
		}
	}
	return "", fmt.Errorf("could not find auth token in cookies")
}

func (i *Client) GenerateReport(report Report) (io.Reader, error) {
	token, err := i.getAuthToken()
	if err != nil {
		return nil, err
	}

	generateURL := fmt.Sprintf("https://reports.izettle.com/report/purchases/generate?fromDate=%s&toDate=%s&reportType=PDF&subAccountUserId=%s", report.From, report.To, report.User.ID)
	req, _ := http.NewRequest("POST", generateURL, nil)
	req.Header.Add("authorization", "Bearer "+token)
	resp, err := i.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return nil, err
	}
	var generateResult map[string]string
	err = json.Unmarshal(body, &generateResult)
	if err != nil {
		return nil, err
	}

	maxTries := 10
	tries := 0
	for tries < maxTries {
		tries++
		statusURL := fmt.Sprintf("https://reports.izettle.com/report/purchases/%s/status", generateResult["uuid"])
		req, _ = http.NewRequest("GET", statusURL, nil)
		req.Header.Add("authorization", "Bearer "+token)
		resp, err = i.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		body, err = ioutil.ReadAll(resp.Body)
		_ = resp.Body.Close()
		var status map[string]string
		err = json.Unmarshal(body, &status)
		if err != nil {
			return nil, err
		}
		if status["status"] == "PROCESSED" {
			break
		}
		time.Sleep(3)
	}
	if tries == maxTries {
		return nil, fmt.Errorf("")
	}

	reportURL := fmt.Sprintf("https://reports.izettle.com/report/purchases/%s", generateResult["uuid"])
	req, _ = http.NewRequest("GET", reportURL, nil)
	req.Header.Add("authorization", "Bearer "+token)
	resp, err = i.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	body, err = ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()

	return bytes.NewReader(body), nil
}
