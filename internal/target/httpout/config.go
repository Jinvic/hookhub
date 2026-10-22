package httpout

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

const (
	DriverHTTP = "http"

	AuthNone   = "none"
	AuthBearer = "bearer"
	AuthBasic  = "basic"
	AuthHeader = "header"

	BodyPlain = "plain"
	BodyJSON  = "json"
)

// Config is the generic outbound HTTP target.
// URL is taken only from config (never from Event/Message) to avoid SSRF.
type Config struct {
	URL         string `yaml:"url"`
	Method      string `yaml:"method"`       // default POST
	AuthType    string `yaml:"auth_type"`    // none | bearer | basic | header
	Token       string `yaml:"token"`        // bearer token, or value for auth_type=header
	Username    string `yaml:"username"`     // basic
	Password    string `yaml:"password"`     // basic
	HeaderName  string `yaml:"header_name"`  // auth_type=header
	HeaderValue string `yaml:"header_value"` // auth_type=header; if empty, uses token
	BodyFormat  string `yaml:"body_format"`  // plain | json；默认 plain
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

	method := c.method()
	switch method {
	case "POST", "PUT", "PATCH":
		// ok
	default:
		return fmt.Errorf("invalid method %q (want POST, PUT, or PATCH)", c.Method)
	}

	switch c.bodyFormat() {
	case BodyPlain, BodyJSON:
		// ok
	default:
		return fmt.Errorf("invalid body_format %q (want plain or json)", c.BodyFormat)
	}

	switch c.AuthType {
	case AuthNone, "":
		return nil
	case AuthBearer:
		if c.Token == "" {
			return errors.New("token is required when auth_type is bearer")
		}
		return nil
	case AuthBasic:
		if c.Username == "" {
			return errors.New("username is required when auth_type is basic")
		}
		if c.Password == "" {
			return errors.New("password is required when auth_type is basic")
		}
		return nil
	case AuthHeader:
		if c.HeaderName == "" {
			return errors.New("header_name is required when auth_type is header")
		}
		if c.headerValue() == "" {
			return errors.New("header_value or token is required when auth_type is header")
		}
		return nil
	default:
		return fmt.Errorf("invalid auth_type %q (want none, bearer, basic, or header)", c.AuthType)
	}
}

func (c *Config) method() string {
	if c.Method == "" {
		return "POST"
	}
	return strings.ToUpper(c.Method)
}

func (c *Config) bodyFormat() string {
	if c.BodyFormat == "" {
		return BodyPlain
	}
	return strings.ToLower(c.BodyFormat)
}

func (c *Config) headerValue() string {
	if c.HeaderValue != "" {
		return c.HeaderValue
	}
	return c.Token
}
