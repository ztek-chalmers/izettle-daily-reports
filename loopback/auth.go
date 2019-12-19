package loopback

import (
	"context"

	"golang.org/x/oauth2"
)

type Auth struct {
	Storage *Storage
	Oauth   *oauth2.Config
}

func (a *Auth) Refresh(token *oauth2.Token) (oauth2.TokenSource, error) {
	tokenSource := a.Oauth.TokenSource(context.Background(), token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, err
	}
	if newToken.AccessToken != token.AccessToken {
		_ = a.Storage.Persist(*newToken)
	}
	return tokenSource, nil
}
