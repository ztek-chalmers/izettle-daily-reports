package izettle

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"izettle-daily-reports/loopback"
	"net/http"
	"net/url"

	"golang.org/x/oauth2"
)

var Endpoint = oauth2.Endpoint{
	AuthURL:  "https://oauth.izettle.com/authorize",
	TokenURL: "https://oauth.izettle.com/token",
}

func Login(user, password, id, secret string) (*Client, error) {
	storage := &loopback.Storage{Name: "izettle"}
	oauth := &oauth2.Config{
		ClientID:     id,
		ClientSecret: secret,
		Endpoint:     Endpoint,
	}
	auth := &loopback.Auth{
		Storage: storage,
		Oauth:   oauth,
	}

	token, err := storage.Load()
	if err == nil {
		token, err := auth.Refresh(token)
		if err != nil {
			return nil, err
		}
		return &Client{token: token}, nil
	}
	token, err = fetchToken(user, password, id, secret)
	if err != nil {
		return nil, err
	}
	_ = storage.Persist(*token)
	return &Client{token: oauth.TokenSource(context.Background(), token)}, nil
}

// fetchToken uses a password grant instead of an ordinary oauth
// login since this is a private integration
// https://github.com/iZettle/api-documentation/blob/master/authorization.adoc
func fetchToken(user, password, id, secret string) (*oauth2.Token, error) {
	bodyStr := fmt.Sprintf("grant_type=password&client_id=%s&client_secret=%s&username=%s&password=%s", id, secret, user, password)
	body := bytes.NewReader([]byte(url.PathEscape(bodyStr)))
	resp, err := http.Post(oauthURL+"/token", "application/x-www-form-urlencoded", body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	token := &oauth2.Token{}
	err = json.Unmarshal(bytes, token)
	if err != nil {
		return nil, err
	}
	return token, nil
}
