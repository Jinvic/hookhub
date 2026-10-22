package app

import (
	"testing"

	"github.com/Jinvic/hookhub/internal/model"
)

func TestMatchingRoutes(t *testing.T) {
	routes := []model.Route{
		{Name: "a", Match: model.Match{Source: "vikunja", Type: "task.reminder.*"}},
		{Name: "b", Match: model.Match{Source: "vikunja", Type: "task.overdue"}},
		{Name: "c", Match: model.Match{Source: "github"}},
	}
	e := &model.Event{Source: "vikunja", Type: "task.reminder.fired"}
	got := MatchingRoutes(routes, e)
	if len(got) != 1 || got[0].Name != "a" {
		t.Fatalf("got %#v", got)
	}
}
