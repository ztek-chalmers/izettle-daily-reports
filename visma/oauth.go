package visma

import (
	"izettle-daily-reports/loopback"

	"golang.org/x/oauth2"
)

var SandboxEndpoint = oauth2.Endpoint{
	AuthURL:  "https://identity-sandbox.test.vismaonline.com/connect/authorize",
	TokenURL: "https://identity-sandbox.test.vismaonline.com/connect/token",
}

var ProductionEndpoint = oauth2.Endpoint{
	AuthURL:  "https://identity.vismaonline.com/connect/authorize",
	TokenURL: "https://identity.vismaonline.com/connect/token",
}

func Login(id, secret string) (oauth2.TokenSource, error) {
	server := loopback.New(loopback.Config{
		Port:    44300,
		TLSCert: "server.crt",
		TLSKey:  "server.key",
		Auth: &loopback.Auth{
			Storage: &loopback.Storage{Name: "visma"},
			Oauth: &oauth2.Config{
				ClientID:     id,
				ClientSecret: secret,
				Scopes:       []string{"ea:api", "ea:accounting", "offline_access"},
				Endpoint:     SandboxEndpoint,
			},
		},
	})
	return server.LoginOrRefresh()
}
