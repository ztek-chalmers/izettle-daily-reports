package visma

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/appscode/go-querystring/query"
	"github.com/ddliu/go-httpclient"
	"golang.org/x/oauth2"
)

type Client struct {
	token oauth2.TokenSource
	url   string
}

func DateFromString(t string) (Date, error) {
	var d Date
	err := d.UnmarshalJSON([]byte("\"" + t + "\""))
	return d, err
}

func (d *Date) Time() time.Time {
	return d.t
}

func (d *Date) String() string {
	return d.t.Format("2006-01-02")
}

type Date struct {
	t time.Time
}

func (d *Date) UnmarshalJSON(data []byte) error {
	// Ignore null, like in the main JSON package.
	if string(data) == "null" {
		return nil
	}
	// Fractional seconds are handled implicitly by Parse.
	var err error
	noQuote := data[1 : len(data)-2]
	date := strings.Split(string(noQuote), "T")
	part := strings.Split(date[0], "-")
	var intPart []int
	for _, p := range part {
		i, err := strconv.Atoi(p)
		if err != nil {
			return err
		}
		intPart = append(intPart, i)
	}

	d.t = time.Date(intPart[0], time.Month(intPart[1]), intPart[2], 0, 0, 0, 0, time.UTC)
	return err
}

func (d Date) EncodeValues(key string, v *url.Values) error {
	b, err := d.MarshalJSON()
	if err != nil {
		return err
	}
	if len(b) != 0 {
		v.Add(key, string(b[1:len(b)-1]))
	}
	return nil
}

func (d Date) MarshalJSON() ([]byte, error) {
	if y := d.t.Year(); y < 0 || y >= 10000 {
		// RFC 3339 is clear that years are 4 digits exactly.
		// See golang.org/issue/4556#c15 for more discussion.
		return nil, errors.New("Time.MarshalJSON: year outside of range [0,9999]")
	}

	if d.t.Year() < 1000 {
		return []byte{}, nil
	}

	b := make([]byte, 0, len(`"2019-12-12"`))
	b = append(b, '"')
	b = append(b, []byte(d.String())...)
	b = append(b, '"')
	return b, nil
}

type Meta struct {
	CurrentPage          int       `url:"CurrentPage,omitempty"`
	PageSize             int       `url:"PageSize,omitempty"`
	TotalNumberOfPages   int       `url:"TotalNumberOfPages,omitempty"`
	TotalNumberOfResults int       `url:"TotalNumberOfResults,omitempty"`
	ServerTimeUtc        time.Time `url:"ServerTimeUtc,omitempty"`
}

func (c *Client) URL(resource string) string {
	return c.url + resource
}

func (c *Client) Http() (*httpclient.HttpClient, error) {
	token, err := c.token.Token()
	if err != nil {
		return nil, err
	}
	return httpclient.Defaults(httpclient.Map{
		"Authorization": "Bearer " + token.AccessToken,
		"Content-Type":  "application/x-www-form-urlencoded",
	}), nil
}

func (c *Client) GetRequest(resource string, respType interface{}) error {
	http, err := c.Http()
	if err != nil {
		return err
	}

	//client := &http.Client{}
	resp, err := http.Get(c.URL(resource))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("GET request failed: %s. Got '%s' for %s", string(respData), resp.Status, resp.Request.URL)
	}
	return json.Unmarshal(respData, respType)
}

func (c *Client) PostRequest(resource string, reqType interface{}, respType interface{}) error {
	http, err := c.Http()
	if err != nil {
		return err
	}
	form, err := query.Values(reqType)
	if err != nil {
		return err
	}
	resp, err := http.Post(c.URL(resource), form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("POST request failed: %s. Got '%s' for %s", string(respData), resp.Status, resp.Request.URL)
	}
	return json.Unmarshal(respData, respType)
}
