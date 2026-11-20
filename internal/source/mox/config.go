package mox

import (
	"errors"
	"fmt"
)

const (
	KindIncoming = "incoming"
	KindOutgoing = "outgoing"

	AuthNone          = "none"
	AuthAuthorization = "authorization"
)

type Config struct {
	Kind          string `yaml:"kind"`
	AuthType      string `yaml:"auth_type"`
	Authorization string `yaml:"authorization"`
}

func (c *Config) Validate() error {
	switch c.Kind {
	case KindIncoming, KindOutgoing:
		// ok
	case "":
		return errors.New("kind is required")
	default:
		return fmt.Errorf("invalid kind: %q (want incoming or outgoing)", c.Kind)
	}

	switch c.AuthType {
	case AuthNone:
		return nil
	case AuthAuthorization:
		if c.Authorization == "" {
			return errors.New("authorization is required when auth_type is authorization")
		}
		return nil
	case "":
		return errors.New("auth_type is required")
	default:
		return fmt.Errorf("invalid auth_type: %q (want none or authorization)", c.AuthType)
	}
}
