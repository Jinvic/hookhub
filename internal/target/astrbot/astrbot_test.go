package astrbot

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Jinvic/hookhub/internal/model"
)

func TestSendPlain(t *testing.T) {
	var got *http.Request
	var payload imMessageRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r
		b, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(b, &payload); err != nil {
			t.Errorf("unmarshal: %v", err)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	a := New(&Config{
		URL:   srv.URL + "/ignored/prefix",
		Token: "abk_test",
		UMO:   "webchat:FriendMessage:probe",
	})
	a.Client = srv.Client()

	err := a.Send(context.Background(), &model.Message{
		Title: "标题",
		Body:  "正文",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Method != http.MethodPost {
		t.Fatalf("method = %s", got.Method)
	}
	if got.URL.Path != "/api/v1/im/messages" {
		t.Fatalf("path = %s", got.URL.Path)
	}
	if got.Header.Get("Authorization") != "Bearer abk_test" {
		t.Fatalf("Authorization = %q", got.Header.Get("Authorization"))
	}
	if ct := got.Header.Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type = %q", ct)
	}
	if payload.UMO != "webchat:FriendMessage:probe" {
		t.Fatalf("umo = %q", payload.UMO)
	}
	if payload.Message != "标题\n\n正文" {
		t.Fatalf("message = %q", payload.Message)
	}
}

func TestComposePlain(t *testing.T) {
	if got := composePlain(&model.Message{Body: "x"}); got != "x" {
		t.Fatalf("body only = %q", got)
	}
	if got := composePlain(&model.Message{Title: "t"}); got != "t" {
		t.Fatalf("title only = %q", got)
	}
}

func TestSendRequiresBody(t *testing.T) {
	a := New(&Config{URL: "https://example.com", Token: "t", UMO: "u"})
	if err := a.Send(context.Background(), &model.Message{}); err == nil {
		t.Fatal("expected error")
	}
}

func TestSendErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`Insufficient API key scope`))
	}))
	defer srv.Close()

	a := New(&Config{URL: srv.URL, Token: "bad", UMO: "u"})
	a.Client = srv.Client()
	err := a.Send(context.Background(), &model.Message{Body: "x"})
	if err == nil || err.Error() != "astrbot 403 Forbidden: Insufficient API key scope" {
		t.Fatalf("err = %v", err)
	}
}
