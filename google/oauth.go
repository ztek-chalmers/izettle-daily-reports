package google

import (
	"izettle-daily-reports/loopback"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func Login(id, secret string, scope []string) (oauth2.TokenSource, error) {
	server := loopback.New(loopback.Config{
		Port: 8000,
		Auth: &loopback.Auth{
			Storage: &loopback.Storage{Name: "google"},
			Oauth: &oauth2.Config{
				ClientID:     id,
				ClientSecret: secret,
				Scopes:       scope,
				Endpoint:     google.Endpoint,
			},
		},
	})
	return server.LoginOrRefresh()
}
