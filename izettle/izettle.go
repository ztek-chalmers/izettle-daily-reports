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
	// Create cookie jar to store cookies which are set by later requests
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Jar: jar}

	// Get csrf token from login page
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
	n := doc.Find("[name=_csrf]").First()
	token, _ := n.Attr("value")

	// Create the login form and simulate a login
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

	// Make sue that the login added new cookies to the root domain
	host, _ := url.Parse("https://izettle.com")
	if len(jar.Cookies(host)) == 0 {
		return nil, fmt.Errorf("failed to login")
	}

	return &Client{httpClient: client}, nil
}

type User struct {
	ID   string
	Name string
}

func (i *Client) ListUsers() (map[string]User, error) {
	// Get csrf tokens from the staff page to load users
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

	// Get all active users
	staffURL := fmt.Sprintf("https://my.izettle.com/settings/staff?filter=accepted&%s=%s", param, url.QueryEscape(token))
	req, _ := http.NewRequest("GET", staffURL, nil)
	// These headers are required, don't ask me why...
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
	userNodes := doc.Find(".user")
	if len(userNodes.Nodes) == 0 {
		return nil, fmt.Errorf("failed to list users")
	}
	for _, n := range userNodes.Nodes {
		// The name is formatted "KOMITEE .", we ae only interested in the "first name"
		nameValue := doc.FindNodes(n).Find(".name").Nodes[0].FirstChild.Data
		name := strings.ToLower(strings.Split(nameValue, " ")[0])
		// The id is formatted "#user-123", we are only interested in the numbers
		idValue, _ := doc.FindNodes(n).Find(".button").Attr("data-overlay-from")
		id := strings.Split(idValue, "-")[1]
		users[name] = User{
			ID:   id,
			Name: name,
		}
	}
	return users, nil
}

type DayReport struct {
	User User
	Date string
}

func (i *Client) ListReports(user User) ([]DayReport, error) {
	// Get the report summary for a user, this includes all reports which are available
	// grouped by month.
	resp, err := i.httpClient.Get("https://my.izettle.com/reports/summary?user=" + user.ID)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return nil, err
	}
	// Based on the JSON structure we are planning on serializing
	//          "daily"."2019-01".[*].{aggregateStart,...}
	// We leave the leaves as raw messages since they can both contain
	// strings and nested objects.
	var data map[string]map[string][]map[string]*json.RawMessage
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}
	// Time to flatten the data
	var reports []DayReport
	months := data["daily"]
	for _, month := range months {
		for _, day := range month {
			startTime, ok := day["aggregateStart"]
			if !ok {
				return nil, fmt.Errorf("report summary does not have a timestamp")
			}
			// Since we do not run json.UnMarshal(...) the string will begin with '"',
			// create a new slice ignoring the initial character.
			// This timestamp is in the form 2019-09-21T01:56:09.462+0000.
			// and we only care about the year month and day. Therefor we also
			// cut the string at T. There are nicer ways of doing this but this works
			start := strings.Split(string((*startTime)[1:]), "T")[0]
			reports = append(reports, DayReport{
				User: user,
				Date: start,
			})
		}
	}
	return reports, nil
}

func (i *Client) DayReportToPDF(report DayReport) (io.Reader, error) {
	pdfURL := fmt.Sprintf("https://my.izettle.com/reports.pdf?user=%s&aggregation=day&date=%s&type=pdf", report.User.ID, report.Date)
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

type Report struct {
	User User
	From string
	To   string
}

func (r *Report) Day() DayReport {
	return DayReport{
		User: r.User,
		Date: r.From,
	}
}

// getAuthorization returns an authorization token from the cookie jar
func (i *Client) getAuthorization() (string, error) {
	host, _ := url.Parse("https://izettle.com")
	cookies := i.httpClient.Jar.Cookies(host)
	for _, c := range cookies {
		if c.Name == "_izsessionat" {
			return "Bearer " + c.Value, nil
		}
	}
	return "", fmt.Errorf("could not find authentication token in the cookie jar")
}

func (i *Client) authorizedGet(method, url string) ([]byte, error) {
	authorization, err := i.getAuthorization()
	if err != nil {
		return nil, err
	}
	req, _ := http.NewRequest(method, url, nil)
	req.Header.Add("authorization", authorization)
	resp, err := i.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (i *Client) authorizedGetJSON(method, url string, out interface{}) error {
	body, err := i.authorizedGet(method, url)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, out)
}

// ReportToPDF generates reports for arbitrary dates
func (i *Client) ReportToPDF(report Report) (io.Reader, error) {
	// Trigger report generation
	reportType := "PDF"
	generateURL := fmt.Sprintf("https://reports.izettle.com/report/purchases/generate?fromDate=%s&toDate=%s&reportType=%s&subAccountUserId=%s", report.From, report.To, reportType, report.User.ID)
	var generateResult map[string]string
	err := i.authorizedGetJSON("POST", generateURL, &generateURL)
	if err != nil {
		return nil, err
	}

	// Wait for the report to be generated.
	maxTries := 10
	tries := 0
	for tries < maxTries {
		tries++
		statusURL := fmt.Sprintf("https://reports.izettle.com/report/purchases/%s/status", generateResult["uuid"])
		var status map[string]string
		err := i.authorizedGetJSON("GET", statusURL, &status)
		if err != nil {
			return nil, err
		}
		if status["status"] == "PROCESSED" {
			break
		}
		time.Sleep(time.Duration(tries) * time.Second)
	}
	if tries == maxTries {
		return nil, fmt.Errorf("the report ")
	}

	// Fetch the report
	reportURL := fmt.Sprintf("https://reports.izettle.com/report/purchases/%s", generateResult["uuid"])
	body, err := i.authorizedGet("GET", reportURL)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(body), nil
}
