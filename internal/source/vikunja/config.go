package vikunja

import (
	"errors"
	"fmt"
)

const (
	AuthNone = "none"
	AuthHMAC = "hmac"
)

type Config struct {
	AuthType  string `yaml:"auth_type"`
	SecretKey string `yaml:"secret_key"`
}

func (c *Config) Validate() error {
	switch c.AuthType {
	case AuthNone:
		return nil
	case AuthHMAC:
		if c.SecretKey == "" {
			return errors.New("secret_key is required when auth_type is hmac")
		}
		return nil
	case "":
		return errors.New("auth_type is required")
	default:
		return fmt.Errorf("invalid auth_type: %q (want none or hmac)", c.AuthType)
	}
}
