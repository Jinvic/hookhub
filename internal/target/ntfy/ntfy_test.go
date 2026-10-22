package ntfy

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Jinvic/hookhub/internal/model"
)

func TestSendRequest(t *testing.T) {
	var got *http.Request
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"abc"}`))
	}))
	defer srv.Close()

	n := New(&Config{
		URL:      srv.URL + "/",
		Topic:    "work-tasks",
		AuthType: AuthToken,
		Token:    "tok",
	})
	n.Client = srv.Client()

	err := n.Send(context.Background(), &model.Message{
		Title:    "hello",
		Body:     "world",
		Priority: "high",
		Format:   "markdown",
		Extra: map[string]string{
			"tags":  "bell,warning",
			"click": "https://example.com",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if got.Method != http.MethodPost {
		t.Fatalf("method = %s", got.Method)
	}
	if got.URL.Path != "/work-tasks" {
		t.Fatalf("path = %s", got.URL.Path)
	}
	if gotBody != "world" {
		t.Fatalf("body = %q", gotBody)
	}
	if ct := got.Header.Get("Content-Type"); ct != "text/plain; charset=utf-8" {
		t.Fatalf("Content-Type = %q", ct)
	}
	if got.Header.Get("Title") != "hello" {
		t.Fatalf("Title = %q", got.Header.Get("Title"))
	}
	if got.Header.Get("Priority") != "4" {
		t.Fatalf("Priority = %q", got.Header.Get("Priority"))
	}
	if got.Header.Get("Tags") != "bell,warning" {
		t.Fatalf("Tags = %q", got.Header.Get("Tags"))
	}
	if got.Header.Get("Markdown") != "yes" {
		t.Fatalf("Markdown = %q", got.Header.Get("Markdown"))
	}
	if got.Header.Get("Click") != "https://example.com" {
		t.Fatalf("Click = %q", got.Header.Get("Click"))
	}
	if got.Header.Get("Authorization") != "Bearer tok" {
		t.Fatalf("Authorization = %q", got.Header.Get("Authorization"))
	}
}

func TestSendNilExtra(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	n := New(&Config{URL: srv.URL, Topic: "t", AuthType: AuthNone})
	n.Client = srv.Client()
	if err := n.Send(context.Background(), &model.Message{Body: "x"}); err != nil {
		t.Fatal(err)
	}
}

func TestSendErrorBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("denied"))
	}))
	defer srv.Close()

	n := New(&Config{URL: srv.URL, Topic: "t", AuthType: AuthNone})
	n.Client = srv.Client()
	err := n.Send(context.Background(), &model.Message{Body: "x"})
	if err == nil || err.Error() != "ntfy 401 Unauthorized: denied" {
		t.Fatalf("err = %v", err)
	}
}
