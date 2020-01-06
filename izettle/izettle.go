package izettle

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

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

func (c *Client) GetRequest(url, resource string, respType interface{}) error {
	http, err := c.Http()
	if err != nil {
		return err
	}
	resp, err := http.Get(url + resource)
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
