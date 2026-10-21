package vikunja

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestParseReminderFired(t *testing.T) {
	body := []byte(`{
		"event_name": "task.reminder.fired",
		"time": "2024-01-02T03:04:05Z",
		"data": {
			"task": {"title": "买菜", "id": 1},
			"project": {"title": "日常", "id": 2},
			"user": {"username": "alice"},
			"reminder": {"reminder": "2024-01-02T03:00:00Z"}
		}
	}`)

	v := &Vikunja{}
	ev, err := v.Parse(nil, body)
	if err != nil {
		t.Fatal(err)
	}
	if ev.Type != "task.reminder.fired" {
		t.Fatalf("Type = %q", ev.Type)
	}
	if ev.Subject != "任务提醒: 买菜" {
		t.Fatalf("Subject = %q", ev.Subject)
	}
	task, ok := ev.Data["task"].(map[string]any)
	if !ok {
		t.Fatalf("Data.task type %T", ev.Data["task"])
	}
	if task["title"] != "买菜" {
		t.Fatalf("title = %v", task["title"])
	}
	if !strings.HasPrefix(ev.ID, "vikunja:task.reminder.fired:") {
		t.Fatalf("ID = %q", ev.ID)
	}
	if ev.Time.UTC() != time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC) {
		t.Fatalf("Time = %v", ev.Time)
	}
}

func TestParseTasksOverdueEmpty(t *testing.T) {
	body := []byte(`{
		"event_name": "tasks.overdue",
		"time": "2024-01-02T03:04:05Z",
		"data": {"tasks": [], "user": {}, "projects": {}}
	}`)
	v := &Vikunja{}
	ev, err := v.Parse(nil, body)
	if err != nil {
		t.Fatal(err)
	}
	if ev.Subject != "任务逾期: 0 个任务" {
		t.Fatalf("Subject = %q", ev.Subject)
	}
}

func TestParseUnknownEvent(t *testing.T) {
	body := []byte(`{"event_name": "task.created", "time": "2024-01-02T03:04:05Z", "data": {}}`)
	_, err := (&Vikunja{}).Parse(nil, body)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestVerifyHMAC(t *testing.T) {
	body := []byte(`{"event_name":"task.overdue","time":"2024-01-02T03:04:05Z","data":{"task":{"title":"x"}}}`)
	secret := "s3cret"
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	sig := hex.EncodeToString(mac.Sum(nil))

	req, _ := http.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("X-Vikunja-Signature", sig)

	v := &Vikunja{SecretKey: secret}
	if err := v.Verify(req, body); err != nil {
		t.Fatal(err)
	}

	req.Header.Set("X-Vikunja-Signature", "deadbeef")
	if err := v.Verify(req, body); err == nil {
		t.Fatal("expected invalid signature")
	}
}
