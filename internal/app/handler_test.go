package app

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/Jinvic/hookhub/internal/model"
	"github.com/Jinvic/hookhub/internal/target"
)

type failTarget struct {
	name string
	err  error
	msgs []*model.Message
}

func (f *failTarget) Driver() string { return f.name }

func (f *failTarget) Send(_ context.Context, msg *model.Message) error {
	if f.err != nil {
		return f.err
	}
	cp := *msg
	f.msgs = append(f.msgs, &cp)
	return nil
}

var _ target.Target = (*failTarget)(nil)

func TestHandleTestTargetOK(t *testing.T) {
	ft := &failTarget{name: "ntfy"}
	a := &App{
		targets: map[string]target.Target{"ntfy": ft},
		log:     slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})),
	}

	req := httptest.NewRequest(http.MethodPost, "/test/target/ntfy", strings.NewReader(`{"title":"t","body":"b"}`))
	req.SetPathValue("name", "ntfy")
	rr := httptest.NewRecorder()
	a.handleTestTarget(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d body = %q", rr.Code, rr.Body.String())
	}
	if rr.Body.String() != "ok" {
		t.Fatalf("body = %q", rr.Body.String())
	}
	if len(ft.msgs) != 1 || ft.msgs[0].Title != "t" || ft.msgs[0].Body != "b" {
		t.Fatalf("msgs = %+v", ft.msgs)
	}
}

func TestHandleTestTargetDefaultMessage(t *testing.T) {
	ft := &failTarget{name: "ntfy"}
	a := &App{
		targets: map[string]target.Target{"ntfy": ft},
		log:     slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})),
	}

	req := httptest.NewRequest(http.MethodPost, "/test/target/ntfy", nil)
	req.SetPathValue("name", "ntfy")
	rr := httptest.NewRecorder()
	a.handleTestTarget(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	if len(ft.msgs) != 1 {
		t.Fatalf("msgs = %d", len(ft.msgs))
	}
	if ft.msgs[0].Title != "HookHub test" {
		t.Fatalf("title = %q", ft.msgs[0].Title)
	}
	if !strings.Contains(ft.msgs[0].Body, "connectivity check") {
		t.Fatalf("body = %q", ft.msgs[0].Body)
	}
}

func TestHandleTestTargetUnknown(t *testing.T) {
	a := &App{
		targets: map[string]target.Target{},
		log:     slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})),
	}
	req := httptest.NewRequest(http.MethodPost, "/test/target/missing", nil)
	req.SetPathValue("name", "missing")
	rr := httptest.NewRecorder()
	a.handleTestTarget(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rr.Code)
	}
}

func TestHandleTestTargetSendFailed(t *testing.T) {
	ft := &failTarget{name: "ntfy", err: errors.New("boom")}
	a := &App{
		targets: map[string]target.Target{"ntfy": ft},
		log:     slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})),
	}
	req := httptest.NewRequest(http.MethodPost, "/test/target/ntfy", bytes.NewReader(nil))
	req.SetPathValue("name", "ntfy")
	rr := httptest.NewRecorder()
	a.handleTestTarget(rr, req)
	if rr.Code != http.StatusBadGateway {
		t.Fatalf("status = %d body = %q", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "boom") {
		t.Fatalf("body = %q", rr.Body.String())
	}
}

func TestHandleTestTargetInvalidJSON(t *testing.T) {
	ft := &failTarget{name: "ntfy"}
	a := &App{
		targets: map[string]target.Target{"ntfy": ft},
		log:     slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})),
	}
	req := httptest.NewRequest(http.MethodPost, "/test/target/ntfy", strings.NewReader(`{`))
	req.SetPathValue("name", "ntfy")
	rr := httptest.NewRecorder()
	a.handleTestTarget(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rr.Code)
	}
}
