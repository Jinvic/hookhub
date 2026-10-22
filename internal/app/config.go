package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Jinvic/hookhub/internal/model"
	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

type Config struct {
	Server  ServerConfig   `yaml:"server"`
	Sources []SourceConfig `yaml:"sources"`
	Targets []TargetConfig `yaml:"targets"`
	Routes  []model.Route  `yaml:"routes"`
}

type ServerConfig struct {
	Addr     string `yaml:"addr"`
	LogLevel string `yaml:"log_level"` // debug | info | warn | error；默认 info
}

type SourceConfig struct {
	Name   string            `yaml:"name"`
	Driver string            `yaml:"driver"`
	Vars   map[string]string `yaml:"vars"` // template Vars (e.g. base_url); not passed to the driver
	Config map[string]any    `yaml:"config"`
}

type TargetConfig struct {
	Name   string         `yaml:"name"`
	Driver string         `yaml:"driver"`
	Config map[string]any `yaml:"config"`
}

func LoadConfig(path string) (*Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	expanded := os.ExpandEnv(string(raw))

	v := viper.New()
	ext := strings.TrimPrefix(filepath.Ext(path), ".")
	if ext == "" {
		ext = "yaml"
	}
	v.SetConfigType(ext)

	if err := v.ReadConfig(strings.NewReader(expanded)); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg, viperDecoderYAML()); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	if cfg.Server.Addr == "" {
		cfg.Server.Addr = ":8080"
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func viperDecoderYAML() viper.DecoderConfigOption {
	return func(c *mapstructure.DecoderConfig) {
		c.TagName = "yaml"
		c.WeaklyTypedInput = true
	}
}

func (c *Config) Validate() error {
	if len(c.Sources) == 0 {
		return fmt.Errorf("sources: at least one is required")
	}
	if len(c.Targets) == 0 {
		return fmt.Errorf("targets: at least one is required")
	}

	sources := make(map[string]struct{}, len(c.Sources))
	for _, s := range c.Sources {
		if s.Name == "" {
			return fmt.Errorf("sources: name is required")
		}
		if s.Driver == "" {
			return fmt.Errorf("sources %q: driver is required", s.Name)
		}
		if _, ok := sources[s.Name]; ok {
			return fmt.Errorf("sources: duplicate name %q", s.Name)
		}
		sources[s.Name] = struct{}{}
	}

	targets := make(map[string]struct{}, len(c.Targets))
	for _, t := range c.Targets {
		if t.Name == "" {
			return fmt.Errorf("targets: name is required")
		}
		if t.Driver == "" {
			return fmt.Errorf("targets %q: driver is required", t.Name)
		}
		if _, ok := targets[t.Name]; ok {
			return fmt.Errorf("targets: duplicate name %q", t.Name)
		}
		targets[t.Name] = struct{}{}
	}

	for _, r := range c.Routes {
		if r.Name == "" {
			return fmt.Errorf("routes: name is required")
		}
		for _, a := range r.Actions {
			if a.Target == "" {
				return fmt.Errorf("routes %q: action target is required", r.Name)
			}
			if _, ok := targets[a.Target]; !ok {
				return fmt.Errorf("routes %q: unknown target %q", r.Name, a.Target)
			}
		}
	}
	return nil
}

// decodePluginConfig maps a loose config object onto a typed plugin config.
func decodePluginConfig[T any](raw map[string]any, dst *T) error {
	if raw == nil {
		raw = map[string]any{}
	}
	dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:          "yaml",
		Result:           dst,
		WeaklyTypedInput: true,
	})
	if err != nil {
		return err
	}
	return dec.Decode(raw)
}
