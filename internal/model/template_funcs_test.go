package model

import (
	"testing"
	"time"
)

func TestFormatTimeRFC3339ToShanghai(t *testing.T) {
	// 2024-01-02T03:00:00Z == 2024-01-02 11:00 in Asia/Shanghai
	got, err := formatTime("2024-01-02T03:00:00Z", "Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}
	if got != "2024-01-02 11:00" {
		t.Fatalf("got %q", got)
	}
}

func TestFormatTimeCustomLayout(t *testing.T) {
	got, err := formatTime("2024-01-02T03:00:00Z", "Asia/Shanghai", "2006-01-02 15:04:05 MST")
	if err != nil {
		t.Fatal(err)
	}
	if got != "2024-01-02 11:00:00 CST" {
		t.Fatalf("got %q", got)
	}
}

func TestFormatTimeTimeValue(t *testing.T) {
	ts := time.Date(2024, 1, 2, 3, 0, 0, 0, time.UTC)
	got, err := formatTime(ts, "UTC")
	if err != nil {
		t.Fatal(err)
	}
	if got != "2024-01-02 03:00" {
		t.Fatalf("got %q", got)
	}
}

func TestFormatTimeEmptyTZUsesUTC(t *testing.T) {
	got, err := formatTime("2024-01-02T03:00:00Z", "")
	if err != nil {
		t.Fatal(err)
	}
	if got != "2024-01-02 03:00" {
		t.Fatalf("got %q", got)
	}
}

func TestFormatTimeBadTZ(t *testing.T) {
	_, err := formatTime("2024-01-02T03:00:00Z", "Not/AZone")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestTemplateRenderFormatTime(t *testing.T) {
	e := &Event{
		Data: map[string]any{
			"reminder": map[string]any{
				"reminder": "2024-01-02T03:00:00Z",
			},
		},
	}
	msg, err := (Template{
		Body: `提醒：{{ formatTime .Event.Data.reminder.reminder .Vars.timezone }}`,
	}).Render(RenderContext{
		Event: e,
		Vars:  map[string]string{"timezone": "Asia/Shanghai"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if msg.Body != "提醒：2024-01-02 11:00" {
		t.Fatalf("Body = %q", msg.Body)
	}
}
