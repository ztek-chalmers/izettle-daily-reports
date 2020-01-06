package visma

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/appscode/go-querystring/query"
	"github.com/ddliu/go-httpclient"
	"golang.org/x/oauth2"
)

type Client struct {
	token oauth2.TokenSource
	url   string
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
