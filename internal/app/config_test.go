package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Jinvic/hookhub/internal/model"
)

func TestLoadConfigExample(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `
server:
  addr: ":0"
sources:
  - name: vikunja
    driver: vikunja
    config:
      auth_type: none
targets:
  - name: ntfy
    driver: ntfy
    config:
      url: https://ntfy.sh
      topic: test
      auth_type: none
routes:
  - name: r1
    match:
      source: vikunja
      type: "task.reminder.*"
    actions:
      - target: ntfy
        template:
          title: "{{ .Event.Subject }}"
          body: "x"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Sources) != 1 || cfg.Sources[0].Driver != "vikunja" {
		t.Fatalf("sources = %+v", cfg.Sources)
	}

	a, err := New(cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := a.sources["vikunja"]; !ok {
		t.Fatal("missing source instance")
	}
	if _, ok := a.targets["ntfy"]; !ok {
		t.Fatal("missing target instance")
	}
}

func TestConfigUnknownTarget(t *testing.T) {
	cfg := &Config{
		Sources: []SourceConfig{{Name: "s", Driver: "vikunja"}},
		Targets: []TargetConfig{{Name: "t", Driver: "ntfy"}},
		Routes: []model.Route{{
			Name: "r1",
			Actions: []model.Action{{
				Target: "missing",
			}},
		}},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error")
	}
}
