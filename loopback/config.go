package loopback

import "strconv"

type Config struct {
	Port    int
	TLSCert string
	TLSKey  string
	Auth    *Auth
}

const callbackPath = "/callback"
const loginPath = "/login"

func (c *Config) RedirectURL() string {
	return c.LocalURL() + callbackPath
}

func (c *Config) LoginURL() string {
	return c.LocalURL() + loginPath
}

func (c *Config) LocalURL() string {
	if c.TLSCert != "" {
		return "https://" + c.Localhost()
	} else {
		return "http://" + c.Localhost()
	}
}

func (c *Config) Localhost() string {
	localAddr := "localhost:" + strconv.Itoa(c.Port)
	if c.Port == 0 {
		localAddr = "localhost:4959"
	}
	return localAddr
}
