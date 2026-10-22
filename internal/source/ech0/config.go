package ech0

import (
	"errors"
	"fmt"
)

const (
	AuthNone = "none"
	AuthHMAC = "hmac"

	// DefaultMaxSkew is the default replay window when auth_type is hmac (seconds).
	DefaultMaxSkew = 300
)

type Config struct {
	AuthType  string `yaml:"auth_type"`
	SecretKey string `yaml:"secret_key"`
	// MaxSkew is the allowed |now - X-Ech0-Timestamp| window in seconds.
	// When auth_type is hmac: 0 means DefaultMaxSkew; negative disables the check.
	// When auth_type is none: ignored.
	MaxSkew int `yaml:"max_skew"`
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

func (c *Config) skewSeconds() int {
	if c.MaxSkew < 0 {
		return -1 // disabled
	}
	if c.MaxSkew == 0 {
		return DefaultMaxSkew
	}
	return c.MaxSkew
}
