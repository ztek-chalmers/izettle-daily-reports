package visma

import (
	"izettle-daily-reports/loopback"

	"golang.org/x/oauth2"
)

var SandboxEndpoint = oauth2.Endpoint{
	AuthURL:  "https://identity-sandbox.test.vismaonline.com/connect/authorize",
	TokenURL: "https://identity-sandbox.test.vismaonline.com/connect/token",
}
var SandboxURL = "https://eaccountingapi-sandbox.test.vismaonline.com/v2/"

var ProductionEndpoint = oauth2.Endpoint{
	AuthURL:  "https://identity.vismaonline.com/connect/authorize",
	TokenURL: "https://identity.vismaonline.com/connect/token",
}
var ProductionURL = "https://eaccountingapi.vismaonline.com/v2/"

const Production = true
const Sandbox = false

func Login(id, secret string, production bool) (*Client, error) {
	var endpoint oauth2.Endpoint
	if production {
		endpoint = ProductionEndpoint
	} else {
		endpoint = SandboxEndpoint

	}
	server := loopback.New(loopback.Config{
		Port:    44300,
		TLSCert: "server.crt",
		TLSKey:  "server.key",
		Auth: &loopback.Auth{
			Storage: &loopback.Storage{Name: "visma"},
			Oauth: &oauth2.Config{
				ClientID:     id,
				ClientSecret: secret,
				Scopes:       []string{"ea:api", "ea:accounting", "ea:sales", "offline_access"},
				Endpoint:     endpoint,
			},
		},
	})
	token, err := server.LoginOrRefresh()
	if err != nil {
		return nil, err
	}
	var url string
	if production {
		url = ProductionURL
	} else {
		url = SandboxURL
	}
	return &Client{
		token: token,
		url:   url,
	}, nil
}
