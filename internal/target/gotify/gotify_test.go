package gotify

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Jinvic/hookhub/internal/model"
)

func TestSendRequest(t *testing.T) {
	var got *http.Request
	var gotBody pushBody
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r
		b, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(b, &gotBody); err != nil {
			t.Errorf("unmarshal: %v", err)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":1}`))
	}))
	defer srv.Close()

	g := New(&Config{URL: srv.URL + "/", Token: "apptok"})
	g.Client = srv.Client()

	err := g.Send(context.Background(), &model.Message{
		Title:    "hello",
		Body:     "world",
		Priority: "high",
		Format:   "markdown",
		Extra: map[string]string{
			"click":     "https://example.com",
			"big_image": "https://example.com/img.png",
			"intent":    "https://example.com/intent",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if got.Method != http.MethodPost {
		t.Fatalf("method = %s", got.Method)
	}
	if got.URL.Path != "/message" {
		t.Fatalf("path = %s", got.URL.Path)
	}
	if ct := got.Header.Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type = %q", ct)
	}
	if got.Header.Get(headerAppToken) != "apptok" {
		t.Fatalf("token header = %q", got.Header.Get(headerAppToken))
	}
	if gotBody.Title != "hello" || gotBody.Message != "world" || gotBody.Priority != 8 {
		t.Fatalf("body = %+v", gotBody)
	}

	display, _ := gotBody.Extras["client::display"].(map[string]any)
	if display["contentType"] != "text/markdown" {
		t.Fatalf("contentType = %v", display["contentType"])
	}
	notif, _ := gotBody.Extras["client::notification"].(map[string]any)
	click, _ := notif["click"].(map[string]any)
	if click["url"] != "https://example.com" {
		t.Fatalf("click = %v", click)
	}
	if notif["bigImageUrl"] != "https://example.com/img.png" {
		t.Fatalf("bigImageUrl = %v", notif["bigImageUrl"])
	}
	android, _ := gotBody.Extras["android::action"].(map[string]any)
	onRecv, _ := android["onReceive"].(map[string]any)
	if onRecv["intentUrl"] != "https://example.com/intent" {
		t.Fatalf("intentUrl = %v", onRecv)
	}
}

func TestSendRequiresBody(t *testing.T) {
	g := New(&Config{URL: "https://example", Token: "t"})
	if err := g.Send(context.Background(), &model.Message{}); err == nil {
		t.Fatal("expected error")
	}
}

func TestSendErrorBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("denied"))
	}))
	defer srv.Close()

	g := New(&Config{URL: srv.URL, Token: "bad"})
	g.Client = srv.Client()
	err := g.Send(context.Background(), &model.Message{Body: "x"})
	if err == nil || err.Error() != "gotify 401 Unauthorized: denied" {
		t.Fatalf("err = %v", err)
	}
}

func TestBuildExtrasPlain(t *testing.T) {
	ex := buildExtras("text", nil)
	display, _ := ex["client::display"].(map[string]any)
	if display["contentType"] != "text/plain" {
		t.Fatalf("%v", display)
	}
	if _, ok := ex["client::notification"]; ok {
		t.Fatal("unexpected notification extras")
	}
}
