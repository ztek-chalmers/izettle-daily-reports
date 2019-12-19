package loopback

import (
	"context"
	"net/http"
	"time"

	"github.com/pkg/browser"
	"golang.org/x/oauth2"
)

// Inspired by
// https://dev.to/douglasmakey/oauth2-example-with-go-3n8a

type Server struct {
	Config
}

func New(config Config) *Server {
	config.Auth.Oauth.RedirectURL = config.RedirectURL()
	return &Server{Config: config}
}

func (s *Server) LoginOrRefresh() (oauth2.TokenSource, error) {
	token, _ := s.Auth.Storage.Load()
	if token != nil {
		return s.Auth.Refresh(token)
	} else {
		// This nonce should be random and checked in the browser if this
		// was a web-application. Since the server only runs for a few
		// seconds we do not worry about CSRF attacks.
		url := s.Auth.Oauth.AuthCodeURL("abc123")
		source, err := s.LoopbackLogin(url)
		if err != nil {
			return nil, err
		}
		token, err = source.Token()
		if err != nil {
			return nil, err
		}
		_ = s.Auth.Storage.Persist(*token)
		return source, nil
	}
}

func (s *Server) LoopbackLogin(url string) (oauth2.TokenSource, error) {
	tokenCh := make(chan string)
	errCh := make(chan error)

	{
		server := &http.Server{Addr: s.Localhost()}
		handler := http.NewServeMux()
		handler.HandleFunc(loginPath, func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, url, http.StatusTemporaryRedirect)
		})
		handler.HandleFunc(callbackPath, func(w http.ResponseWriter, r *http.Request) {
			code := r.FormValue("code")
			_, _ = w.Write([]byte("You are now logged in! Please close this tab"))
			go func() {
				select {
				case <-time.After(6 * time.Second):
					_ = server.Shutdown(context.Background())
				}
			}()
			tokenCh <- code
		})
		server.Handler = handler
		go func() {
			var err error
			if s.TLSCert != "" {
				err = server.ListenAndServeTLS(s.TLSCert, s.TLSKey)
			} else {
				err = server.ListenAndServe()
			}
			if err != http.ErrServerClosed {
				errCh <- err
			}
		}()
	}

	{
		err := browser.OpenURL(s.LoginURL())
		if err != nil {
			return nil, err
		}
	}

	{
		select {
		case code := <-tokenCh:
			token, err := s.Auth.Oauth.Exchange(context.Background(), code)
			if err != nil {
				return nil, err
			}
			source := s.Auth.Oauth.TokenSource(context.Background(), token)
			return source, nil
		case err := <-errCh:
			return nil, err
		}
	}
}
