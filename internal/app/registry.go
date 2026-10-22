package app

import (
	"fmt"

	"github.com/Jinvic/hookhub/internal/source"
	"github.com/Jinvic/hookhub/internal/source/ech0"
	"github.com/Jinvic/hookhub/internal/source/vikunja"
	"github.com/Jinvic/hookhub/internal/target"
	"github.com/Jinvic/hookhub/internal/target/gotify"
	"github.com/Jinvic/hookhub/internal/target/httpout"
	"github.com/Jinvic/hookhub/internal/target/ntfy"
)

func buildSources(cfgs []SourceConfig) (map[string]source.Source, error) {
	out := make(map[string]source.Source, len(cfgs))
	for _, c := range cfgs {
		s, err := newSource(c)
		if err != nil {
			return nil, fmt.Errorf("source %q: %w", c.Name, err)
		}
		if s.Driver() != c.Driver {
			return nil, fmt.Errorf("source %q: driver mismatch: impl %q != config %q", c.Name, s.Driver(), c.Driver)
		}
		out[c.Name] = s
	}
	return out, nil
}

func buildTargets(cfgs []TargetConfig) (map[string]target.Target, error) {
	out := make(map[string]target.Target, len(cfgs))
	for _, c := range cfgs {
		t, err := newTarget(c)
		if err != nil {
			return nil, fmt.Errorf("target %q: %w", c.Name, err)
		}
		if t.Driver() != c.Driver {
			return nil, fmt.Errorf("target %q: driver mismatch: impl %q != config %q", c.Name, t.Driver(), c.Driver)
		}
		out[c.Name] = t
	}
	return out, nil
}

func newSource(c SourceConfig) (source.Source, error) {
	switch c.Driver {
	case vikunja.DriverVikunja:
		var cfg vikunja.Config
		if err := decodePluginConfig(c.Config, &cfg); err != nil {
			return nil, err
		}
		if err := cfg.Validate(); err != nil {
			return nil, err
		}
		return vikunja.New(&cfg), nil
	case ech0.DriverEch0:
		var cfg ech0.Config
		if err := decodePluginConfig(c.Config, &cfg); err != nil {
			return nil, err
		}
		if err := cfg.Validate(); err != nil {
			return nil, err
		}
		return ech0.New(&cfg), nil
	default:
		return nil, fmt.Errorf("unknown driver %q", c.Driver)
	}
}

func newTarget(c TargetConfig) (target.Target, error) {
	switch c.Driver {
	case ntfy.DriverNtfy:
		var cfg ntfy.Config
		if err := decodePluginConfig(c.Config, &cfg); err != nil {
			return nil, err
		}
		if err := cfg.Validate(); err != nil {
			return nil, err
		}
		return ntfy.New(&cfg), nil
	case gotify.DriverGotify:
		var cfg gotify.Config
		if err := decodePluginConfig(c.Config, &cfg); err != nil {
			return nil, err
		}
		if err := cfg.Validate(); err != nil {
			return nil, err
		}
		return gotify.New(&cfg), nil
	case httpout.DriverHTTP:
		var cfg httpout.Config
		if err := decodePluginConfig(c.Config, &cfg); err != nil {
			return nil, err
		}
		if err := cfg.Validate(); err != nil {
			return nil, err
		}
		return httpout.New(&cfg), nil
	default:
		return nil, fmt.Errorf("unknown driver %q", c.Driver)
	}
}
