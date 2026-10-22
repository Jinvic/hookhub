package gotify

import "errors"

type Config struct {
	URL   string `yaml:"url"`   // Gotify server base, e.g. https://push.example.de
	Token string `yaml:"token"` // application token (X-Gotify-Key)
}

func (c *Config) Validate() error {
	if c.URL == "" {
		return errors.New("url is required")
	}
	if c.Token == "" {
		return errors.New("token is required")
	}
	return nil
}
