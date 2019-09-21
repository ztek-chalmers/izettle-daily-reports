package google

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/pkg/browser"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Inspired by
// https://dev.to/douglasmakey/oauth2-example-with-go-3n8a

func loopbackLogin(config *oauth2.Config) (string, error) {
	tokenCh := make(chan string)
	errCh := make(chan error)

	var server *http.Server
	server = &http.Server{
		Addr: ":8000",
		Handler: newHttpHandler(config, func(token string) {
			go func() {
				_ = server.Shutdown(context.Background())
			}()
			tokenCh <- token
		}),
	}
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	err := browser.OpenURL("http://localhost:8000/auth/google/login")
	if err != nil {
		return "", err
	}

	select {
	case token := <-tokenCh:
		return token, nil
	case err := <-errCh:
		return "", err
	}
}

func readTokenFile() (*oauth2.Token, error) {
	codeBytes, err := ioutil.ReadFile("./token")
	if err != nil {
		return nil, err
	}

	token := oauth2.Token{}
	err = json.Unmarshal(codeBytes, &token)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func writeTokenFile(token oauth2.Token) error {
	bytes, err := json.Marshal(token)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("./token", bytes, 0755)
	if err != nil {
		return err
	}
	return nil
}

func Login(id, secret string, scope []string) (*oauth2.TokenSource, error) {
	config := &oauth2.Config{
		RedirectURL:  "http://localhost:8000/auth/google/callback",
		ClientID:     id,
		ClientSecret: secret,
		Scopes:       scope,
		Endpoint:     google.Endpoint,
	}

	token, _ := readTokenFile()
	if token != nil {
		tokenSource := config.TokenSource(context.Background(), token)
		newToken, err := tokenSource.Token()
		if err != nil {
			log.Fatalln(err)
		}
		if newToken.AccessToken != token.AccessToken {
			_ = writeTokenFile(*newToken)
		}
		return &tokenSource, nil
	} else {
		var err error
		code, err := loopbackLogin(config)
		if err != nil {
			return nil, err
		}
		token, err := config.Exchange(context.Background(), code)
		if err != nil {
			log.Fatal(err)
		}
		_ = writeTokenFile(*token)
		tokenSource := config.TokenSource(context.Background(), token)
		return &tokenSource, nil
	}
}

func newHttpHandler(config *oauth2.Config, callback func(token string)) http.Handler {
	mux := http.NewServeMux()
	// OauthGoogle
	mux.HandleFunc("/auth/google/login", oauthGoogleLogin(config))
	mux.HandleFunc("/auth/google/callback", oauthGoogleCallback(callback))
	return mux
}

func oauthGoogleLogin(config *oauth2.Config) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Create oauthState cookie
		oauthState := generateStateOauthCookie(w)
		// AuthCodeURL receive state that is a token to protect the user from CSRF attacks. You must always provide a non-empty string and
		// validate that it matches the the state query parameter on your redirect callback.
		u := config.AuthCodeURL(oauthState)
		http.Redirect(w, r, u, http.StatusTemporaryRedirect)
	}
}

func oauthGoogleCallback(cb func(token string)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.FormValue("code")
		_, _ = w.Write([]byte("<script>window.close();</script>"))
		cb(code)
	}
}

func generateStateOauthCookie(w http.ResponseWriter) string {
	var expiration = time.Now().Add(365 * 24 * time.Hour)

	b := make([]byte, 16)
	_, _ = rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{Name: "oauthstate", Value: state, Expires: expiration}
	http.SetCookie(w, &cookie)

	return state
}
