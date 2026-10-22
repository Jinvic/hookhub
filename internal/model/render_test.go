package model

import (
	"strings"
	"testing"
)

func TestTemplateRender(t *testing.T) {
	e := &Event{
		Source:  "vikunja",
		Type:    "task.reminder.fired",
		Subject: "买菜",
		Data: map[string]any{
			"task": map[string]any{
				"title": "买菜",
				"id":    1,
			},
			"project": map[string]any{
				"title": "日常",
			},
			"reminder": map[string]any{
				"reminder": "2024-01-02T03:00:00Z",
			},
		},
	}

	tmpl := Template{
		Title:    "{{ .Event.Data.task.title }}",
		Body:     "项目：{{ .Event.Data.project.title }}\n提醒：{{ .Event.Data.reminder.reminder }}",
		Format:   "markdown",
		Priority: "high",
		Extra: map[string]string{
			"tags":  "bell",
			"click": "{{ .Vars.base_url }}/tasks/{{ .Event.Data.task.id }}",
		},
	}

	msg, err := tmpl.Render(RenderContext{
		Event: e,
		Vars:  map[string]string{"base_url": "https://vikunja.example.com"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if msg.Title != "买菜" {
		t.Fatalf("Title = %q", msg.Title)
	}
	if !strings.Contains(msg.Body, "项目：日常") || !strings.Contains(msg.Body, "提醒：2024-01-02T03:00:00Z") {
		t.Fatalf("Body = %q", msg.Body)
	}
	if msg.Format != "markdown" {
		t.Fatalf("Format = %q", msg.Format)
	}
	if msg.Priority != "high" {
		t.Fatalf("Priority = %q", msg.Priority)
	}
	if msg.Extra["tags"] != "bell" {
		t.Fatalf("tags = %q", msg.Extra["tags"])
	}
	if msg.Extra["click"] != "https://vikunja.example.com/tasks/1" {
		t.Fatalf("click = %q", msg.Extra["click"])
	}
}

func TestTemplateRenderMissingKey(t *testing.T) {
	e := &Event{Data: map[string]any{}}
	_, err := (Template{Title: "{{ .Event.Data.task.title }}"}).Render(RenderContext{Event: e})
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestTemplateRenderNilEvent(t *testing.T) {
	_, err := (Template{Title: "x"}).Render(RenderContext{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestTemplateRenderSubject(t *testing.T) {
	e := &Event{Subject: "摘要"}
	msg, err := (Template{Title: "{{ .Event.Subject }}"}).Render(RenderContext{Event: e})
	if err != nil {
		t.Fatal(err)
	}
	if msg.Title != "摘要" {
		t.Fatalf("Title = %q", msg.Title)
	}
}
