package visma

import (
	"fmt"
	"izettle-daily-reports/loopback"

	"golang.org/x/oauth2"
)

type Environment struct {
	Name         string
	ClientID     string
	ClientSecret string
	ApiURL       string
	AuthURL      string
	TokenURL     string
}

func Login(environment Environment) (*Client, error) {
	server := loopback.New(loopback.Config{
		Port:    44300,
		TLSCert: "server.crt",
		TLSKey:  "server.key",
		Auth: &loopback.Auth{
			Storage: &loopback.Storage{Name: fmt.Sprintf("visma-%s", environment.Name)},
			Oauth: &oauth2.Config{
				ClientID:     environment.ClientID,
				ClientSecret: environment.ClientSecret,
				Scopes:       []string{"ea:api", "ea:accounting", "ea:sales", "offline_access"},
				Endpoint: oauth2.Endpoint{
					AuthURL:  environment.AuthURL,
					TokenURL: environment.TokenURL,
				},
			},
		},
	})
	token, err := server.LoginOrRefresh()
	if err != nil {
		return nil, err
	}
	return &Client{
		token: token,
		url:   environment.ApiURL,
	}, nil
}
