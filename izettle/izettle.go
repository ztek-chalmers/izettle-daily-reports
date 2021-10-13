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
	"strconv"
	"strings"
	"time"
)

type Client struct {
	httpClient *http.Client
}

func Login(session string) (*Client, error) {
	// Create cookie jar to store cookies which are set by later requests
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	cookieURL, err := url.Parse("https://my.izettle.com")
	jar.SetCookies(cookieURL, []*http.Cookie{{
		Name: "_izsessionat", Value: session,
	}})

	client := &http.Client{Jar: jar}
	return &Client{httpClient: client}, nil
}

type User struct {
	ID   string
	UUID string
	Name string
}

type IZettleAccount struct {
	ID        int
	UUID      string
	Status    string
	FirstName string
}

type IZettleSelfAccount struct {
	Data struct {
		ID   int    `json:"userId"`
		UUID string `json:"userUUID"`
	}
}

func (i *Client) ListUsers() (map[string]User, error) {
	// Get all accounts except for ztyret
	subAccounts := []IZettleAccount{}
	err := i.authorizedGetJSON("GET", "https://secure.izettle.com/api/resources/subaccounts", &subAccounts)
	if err != nil {
		return nil, err
	}

	// Get ztyrets account number
	selfAccount := IZettleSelfAccount{}
	err = i.authorizedGetJSON("GET", "https://secure.izettle.com/api/resources/user-profiles/self", &selfAccount)
	if err != nil {
		return nil, err
	}
	subAccounts = append(subAccounts, IZettleAccount{
		ID:        selfAccount.Data.ID,
		UUID:      selfAccount.Data.UUID,
		Status:    "ACCEPTED",
		FirstName: "ztyret",
	})

	users := make(map[string]User)
	for _, account := range subAccounts {
		if account.Status == "ACCEPTED" {
			name := strings.ToLower(account.FirstName)
			// TODO: This fixes the mappings between the systems. Make configurable outside of the source code.
			if name == "nydaltonz" {
				name = "daltonz"
			}
			if name == "zåg" {
				name = "zag"
			}
			users[name] = User{
				ID:   strconv.Itoa(account.ID),
				UUID: account.UUID,
				Name: name,
			}
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
	summaryURL := fmt.Sprintf("https://reports.izettle.com/report/purchases/summary/day?startDateTime=%s&endDateTime=%s&subAccountUserId=%s", "2020-01-01T00:00:00.000Z", "2030-01-01T00:00:00.000Z", user.UUID)
	var data []map[string]*json.RawMessage
	err := i.authorizedGetJSON("GET", summaryURL, &data)
	if err != nil {
		return nil, err
	}
	// Time to flatten the data
	var reports []DayReport
	for _, day := range data {
		// Use aggregateEnd since due to the timezone, the start of the day is xxxx-xx-01Z22:00:00, instead of xxxx-xx-02Z00:00:00
		startTime, ok := day["aggregateEnd"]
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
	return reports, nil
}

func (i *Client) DayReportToPDF(report DayReport) (io.Reader, error) {
	generateResponse, err := i.authorizedGet("POST", fmt.Sprintf("https://reports.izettle.com/report/purchases/generate/v2?fromDate=%s&toDate=%s&subAccountUserId=%s&reportType=PDF", report.Date, report.Date, report.User.UUID))
	if err != nil {
		return nil, err
	}
	var request struct {
		UUID string
	}
	err = json.Unmarshal(generateResponse, &request)
	if err != nil {
		return nil, err
	}

	// The generation request is added to a queue. Loop a few times using an exponential backoff strategy.
	for k := 1; k <= 4; k++ {
		delay := time.Duration(1 << k)
		time.Sleep(time.Second * delay)
		body, err := i.authorizedGet("GET", fmt.Sprintf("https://reports.izettle.com/report/purchases/%s/status", request.UUID))
		if err != nil {
			return nil, err
		}
		var status struct {
			Status string
		}
		err = json.Unmarshal(body, &status)
		if err != nil {
			return nil, err
		}
		if status.Status == "PROCESSED" {
			pdf, err := i.authorizedGet("GET", fmt.Sprintf("https://reports.izettle.com/report/purchases/%s", request.UUID))
			if err != nil {
				return nil, err
			}
			return bytes.NewReader(pdf), nil
		}
	}
	return nil, fmt.Errorf("report generator timed out.")
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
	host, _ := url.Parse("https://my.izettle.com")
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

func (i *Client) authorizedGetJSON(method, resourceURL string, out interface{}) error {
	body, err := i.authorizedGet(method, resourceURL)
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
