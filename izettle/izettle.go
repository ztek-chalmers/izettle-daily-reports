package izettle

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/ddliu/go-httpclient"
	"golang.org/x/oauth2"
)

const productURL = "https://products.izettle.com"
const purchaseURL = "https://purchase.izettle.com"
const oauthURL = "https://oauth.izettle.com"

type Client struct {
	token oauth2.TokenSource
}

func (c *Client) Http() (*httpclient.HttpClient, error) {
	token, err := c.token.Token()
	if err != nil {
		return nil, err
	}
	return httpclient.Defaults(httpclient.Map{
		"Authorization": "Bearer " + token.AccessToken,
		"Content-Type":  "application/json",
	}), nil
}

func (c *Client) GetRequest(url string) ([]byte, error) {
	http, err := c.Http()
	if err != nil {
		return nil, err
	}
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("GET request failed: %s. Got '%s' for %s", string(respData), resp.Status, resp.Request.URL)
	}
	return respData, nil
}

func (c *Client) GetAllRequest(url string, add func(data []byte) error) error {
	resp := &struct {
		LinkURLS []string
	}{}

	data, err := c.GetRequest(url)
	if err != nil {
		return err
	}
	err = add(data)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, resp)
	if err != nil {
		return err
	}

	for _, url := range resp.LinkURLS {
		parts := strings.Split(url, ";")
		url := strings.TrimRight(strings.TrimLeft(parts[0], "<"), ">")
		rel := strings.TrimSpace(parts[1])
		if rel == "rel=\"next\"" {
			time.Sleep(1 * time.Second)
			err := c.GetAllRequest(url, add)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
