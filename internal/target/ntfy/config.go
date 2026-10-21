package ntfy

import (
	"errors"
	"fmt"
)

const (
	AuthNone  = "none"
	AuthBasic = "basic"
	AuthToken = "token"
)

type Config struct {
	URL      string `yaml:"url"`
	Topic    string `yaml:"topic"`
	AuthType string `yaml:"auth_type"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Token    string `yaml:"token"`
}

func (c *Config) Validate() error {
	if c.URL == "" {
		return errors.New("url is required")
	}
	if c.Topic == "" {
		return errors.New("topic is required")
	}

	switch c.AuthType {
	case AuthNone:
		return nil
	case AuthBasic:
		if c.Username == "" {
			return errors.New("username is required when auth_type is basic")
		}
		if c.Password == "" {
			return errors.New("password is required when auth_type is basic")
		}
		return nil
	case AuthToken:
		if c.Token == "" {
			return errors.New("token is required when auth_type is token")
		}
		return nil
	case "":
		return errors.New("auth_type is required")
	default:
		return fmt.Errorf("invalid auth_type: %q (want none, basic, or token)", c.AuthType)
	}
}
