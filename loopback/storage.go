package loopback

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"golang.org/x/oauth2"
)

type Storage struct {
	Name string
}

func (s *Storage) Load() (*oauth2.Token, error) {
	bytes, err := ioutil.ReadFile(s.Filename())
	if err != nil {
		return nil, err
	}
	token := oauth2.Token{}
	err = json.Unmarshal(bytes, &token)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (s *Storage) Persist(token oauth2.Token) error {
	if token.RefreshToken == "" {
		return fmt.Errorf("can not perist oauth token without refreshToken")
	}
	bytes, err := json.Marshal(token)
	if err != nil {
		return err
	}
	err = os.MkdirAll("tokens", 0775)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(s.Filename(), bytes, 0755)
	return err
}

func (s *Storage) Filename() string {
	return "tokens/" + s.Name + ".token"
}
