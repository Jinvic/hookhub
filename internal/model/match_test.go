package model

import "testing"

func TestMatchMatches(t *testing.T) {
	e := &Event{Source: "vikunja", Type: "task.reminder.fired"}

	tests := []struct {
		name string
		m    Match
		want bool
	}{
		{name: "empty match all", m: Match{}, want: true},
		{name: "source ok", m: Match{Source: "vikunja"}, want: true},
		{name: "source miss", m: Match{Source: "github"}, want: false},
		{name: "exact type", m: Match{Type: "task.reminder.fired"}, want: true},
		{name: "segment star", m: Match{Type: "task.reminder.*"}, want: true},
		{name: "star middle", m: Match{Type: "task.*.fired"}, want: true},
		{name: "too few segments", m: Match{Type: "task.*"}, want: false},
		{name: "too many segments", m: Match{Type: "task.reminder.fired.extra"}, want: false},
		{name: "no match bare type", m: Match{Type: "task.reminder"}, want: false},
		{name: "source and type", m: Match{Source: "vikunja", Type: "task.reminder.*"}, want: true},
		{name: "wrong source with type", m: Match{Source: "gitea", Type: "task.reminder.*"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.Matches(e); got != tt.want {
				t.Fatalf("Matches = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchNilEvent(t *testing.T) {
	if (Match{}).Matches(nil) {
		t.Fatal("nil event should not match")
	}
}
