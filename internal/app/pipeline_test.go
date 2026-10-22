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

func (f *fakeTarget) Send(_ context.Context, msg *model.Message, _ map[string]any) error {
	cp := *msg
	f.msgs = append(f.msgs, &cp)
	return nil
}

var _ target.Target = (*fakeTarget)(nil)

func TestProcessFanOut(t *testing.T) {
	ft := &fakeTarget{name: "ntfy"}
	a := &App{
		cfg: &Config{
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
		},
		sourceVars: map[string]map[string]string{},
		targets:    map[string]target.Target{"ntfy": ft},
		log:        slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})),
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
