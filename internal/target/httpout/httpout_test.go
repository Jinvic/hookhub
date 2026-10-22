package httpout

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
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	h := New(&Config{URL: srv.URL + "/hook", AuthType: AuthNone})
	h.Client = srv.Client()

	if err := h.Send(context.Background(), &model.Message{Body: "hello"}); err != nil {
		t.Fatal(err)
	}
	if got.Method != http.MethodPost {
		t.Fatalf("method = %s", got.Method)
	}
	if got.URL.Path != "/hook" {
		t.Fatalf("path = %s", got.URL.Path)
	}
	if gotBody != "hello" {
		t.Fatalf("body = %q", gotBody)
	}
	if ct := got.Header.Get("Content-Type"); ct != "text/plain; charset=utf-8" {
		t.Fatalf("Content-Type = %q", ct)
	}
}

func TestSendJSONBearer(t *testing.T) {
	var got *http.Request
	var payload jsonPayload
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &payload)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	h := New(&Config{
		URL:        srv.URL,
		AuthType:   AuthBearer,
		Token:      "secret",
		BodyFormat: BodyJSON,
	})
	h.Client = srv.Client()

	err := h.Send(context.Background(), &model.Message{
		Title:    "t",
		Body:     "m",
		Priority: "high",
		Format:   "markdown",
		Extra:    map[string]string{"click": "https://example.com"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Header.Get("Authorization") != "Bearer secret" {
		t.Fatalf("Authorization = %q", got.Header.Get("Authorization"))
	}
	if payload.Title != "t" || payload.Message != "m" || payload.Priority != "high" {
		t.Fatalf("payload = %+v", payload)
	}
	if payload.Extra["click"] != "https://example.com" {
		t.Fatalf("extra = %v", payload.Extra)
	}
}

func TestSendCustomHeader(t *testing.T) {
	var got *http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	h := New(&Config{
		URL:        srv.URL,
		AuthType:   AuthHeader,
		HeaderName: "X-Custom-Token",
		Token:      "abc",
	})
	h.Client = srv.Client()
	if err := h.Send(context.Background(), &model.Message{Body: "x"}); err != nil {
		t.Fatal(err)
	}
	if got.Header.Get("X-Custom-Token") != "abc" {
		t.Fatalf("header = %q", got.Header.Get("X-Custom-Token"))
	}
}

func TestSendErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("nope"))
	}))
	defer srv.Close()

	h := New(&Config{URL: srv.URL})
	h.Client = srv.Client()
	err := h.Send(context.Background(), &model.Message{Body: "x"})
	if err == nil || err.Error() != "http 502 Bad Gateway: nope" {
		t.Fatalf("err = %v", err)
	}
}

func TestSendRequiresBody(t *testing.T) {
	h := New(&Config{URL: "https://example.com"})
	if err := h.Send(context.Background(), &model.Message{}); err == nil {
		t.Fatal("expected error")
	}
}

func TestNoFollowRedirect(t *testing.T) {
	final := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer final.Close()

	redir := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, final.URL, http.StatusFound)
	}))
	defer redir.Close()

	h := New(&Config{URL: redir.URL})
	// use default noRedirectClient
	err := h.Send(context.Background(), &model.Message{Body: "x"})
	if err == nil {
		t.Fatal("expected error on redirect response")
	}
}
