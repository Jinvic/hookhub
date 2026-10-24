package astrbot

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// DriverAstrBot is the plugin id used in config `driver:` and Driver().
const DriverAstrBot = "astrbot"

type Config struct {
	URL   string `yaml:"url"`   // AstrBot base, e.g. https://astrbot.example.com
	Token string `yaml:"token"` // API Key (Bearer / abk_…)
	UMO   string `yaml:"umo"`   // destination session, e.g. webchat:FriendMessage:xxx
}

func (c *Config) Validate() error {
	if c.URL == "" {
		return errors.New("url is required")
	}
	u, err := url.Parse(c.URL)
	if err != nil {
		return fmt.Errorf("parse url: %w", err)
	}
	switch strings.ToLower(u.Scheme) {
	case "http", "https":
		// ok
	default:
		return fmt.Errorf("url scheme must be http or https, got %q", u.Scheme)
	}
	if u.Host == "" {
		return errors.New("url host is required")
	}
	if c.Token == "" {
		return errors.New("token is required")
	}
	if c.UMO == "" {
		return errors.New("umo is required")
	}
	return nil
}
