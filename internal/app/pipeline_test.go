package app

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/Jinvic/hookhub/internal/model"
	"github.com/Jinvic/hookhub/internal/target"
)

type fakeTarget struct {
	name string
	msgs []*model.Message
}

func (f *fakeTarget) Driver() string { return f.name }

func (f *fakeTarget) Send(_ context.Context, msg *model.Message) error {
	cp := *msg
	f.msgs = append(f.msgs, &cp)
	return nil
}

var _ target.Target = (*fakeTarget)(nil)

func TestProcessFanOut(t *testing.T) {
	ft := &fakeTarget{name: "ntfy"}
	cfg := &Config{
		Routes: []model.Route{{
			Name:  "r1",
			Match: model.Match{Source: "vikunja", Type: "task.reminder.*"},
			Actions: []model.Action{{
				Target: "ntfy",
				Template: model.Template{
					Title: "{{ .Event.Subject }}",
					Body:  "hi {{ .Event.Data.task.title }}",
				},
			}},
		}},
	}
	whens, err := compileRouteWhens(cfg.Routes)
	if err != nil {
		t.Fatal(err)
	}
	a := &App{
		cfg:          cfg,
		sourceVars:   map[string]map[string]string{},
		targets:      map[string]target.Target{"ntfy": ft},
		whenPrograms: whens,
		log:          slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})),
	}

	a.Process(context.Background(), &model.Event{
		ID:      "1",
		Source:  "vikunja",
		Type:    "task.reminder.fired",
		Subject: "提醒",
		Data:    map[string]any{"task": map[string]any{"title": "买菜"}},
	})

	if len(ft.msgs) != 1 {
		t.Fatalf("sent %d messages", len(ft.msgs))
	}
	if ft.msgs[0].Title != "提醒" || ft.msgs[0].Body != "hi 买菜" {
		t.Fatalf("msg = %+v", ft.msgs[0])
	}
}

func TestProcessMatchWhenFilters(t *testing.T) {
	ft := &fakeTarget{name: "ntfy"}
	cfg := &Config{
		Routes: []model.Route{{
			Name: "r1",
			Match: model.Match{
				Source: "vikunja",
				Type:   "task.reminder.*",
				When:   `data.project.id == 2`,
			},
			Actions: []model.Action{{
				Target:   "ntfy",
				Template: model.Template{Body: "x"},
			}},
		}},
	}
	whens, err := compileRouteWhens(cfg.Routes)
	if err != nil {
		t.Fatal(err)
	}
	a := &App{
		cfg:          cfg,
		sourceVars:   map[string]map[string]string{},
		targets:      map[string]target.Target{"ntfy": ft},
		whenPrograms: whens,
		log:          slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})),
	}

	a.Process(context.Background(), &model.Event{
		ID:     "1",
		Source: "vikunja",
		Type:   "task.reminder.fired",
		Data:   map[string]any{"project": map[string]any{"id": float64(3)}},
	})
	if len(ft.msgs) != 0 {
		t.Fatalf("expected filtered, got %d", len(ft.msgs))
	}

	a.Process(context.Background(), &model.Event{
		ID:     "2",
		Source: "vikunja",
		Type:   "task.reminder.fired",
		Data:   map[string]any{"project": map[string]any{"id": float64(2)}},
	})
	if len(ft.msgs) != 1 {
		t.Fatalf("expected 1 send, got %d", len(ft.msgs))
	}
}

func TestProcessActionWhenFanOut(t *testing.T) {
	ntfy := &fakeTarget{name: "ntfy"}
	astrbot := &fakeTarget{name: "astrbot"}
	cfg := &Config{
		Routes: []model.Route{{
			Name:  "r1",
			Match: model.Match{Source: "ech0", Type: "comment.created"},
			Actions: []model.Action{
				{
					Target:   "ntfy",
					When:     `data.payload_raw.Comment.nickname != "bot"`,
					Template: model.Template{Body: "ntfy"},
				},
				{
					Target:   "astrbot",
					Template: model.Template{Body: "astrbot"},
				},
			},
		}},
	}
	whens, err := compileRouteWhens(cfg.Routes)
	if err != nil {
		t.Fatal(err)
	}
	a := &App{
		cfg:          cfg,
		sourceVars:   map[string]map[string]string{},
		targets:      map[string]target.Target{"ntfy": ntfy, "astrbot": astrbot},
		whenPrograms: whens,
		log:          slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})),
	}

	a.Process(context.Background(), &model.Event{
		ID:     "1",
		Source: "ech0",
		Type:   "comment.created",
		Data: map[string]any{
			"payload_raw": map[string]any{
				"Comment": map[string]any{"nickname": "bot"},
			},
		},
	})
	if len(ntfy.msgs) != 0 {
		t.Fatalf("ntfy should skip bot, got %d", len(ntfy.msgs))
	}
	if len(astrbot.msgs) != 1 {
		t.Fatalf("astrbot should still send, got %d", len(astrbot.msgs))
	}
}

func TestNewRejectsBadWhen(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{Addr: ":0"},
		Sources: []SourceConfig{{
			Name: "vikunja", Driver: "vikunja",
			Config: map[string]any{"auth_type": "none"},
		}},
		Targets: []TargetConfig{{
			Name: "ntfy", Driver: "ntfy",
			Config: map[string]any{"url": "https://ntfy.sh", "topic": "t", "auth_type": "none"},
		}},
		Routes: []model.Route{{
			Name:  "bad",
			Match: model.Match{Source: "vikunja", When: `"not-bool"`},
			Actions: []model.Action{{
				Target:   "ntfy",
				Template: model.Template{Body: "x"},
			}},
		}},
	}
	_, err := New(cfg, nil)
	if err == nil {
		t.Fatal("expected compile error")
	}
}
