package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/Jinvic/hookhub/internal/model"
)

const (
	maxBodyBytes    = 1 << 20 // 1 MiB
	debugBodyLogMax = 512
	processTimeout  = 30 * time.Second
	testTimeout     = 10 * time.Second
)

func (a *App) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", a.handleHealthz)
	mux.HandleFunc("GET /readyz", a.handleHealthz)
	mux.HandleFunc("POST /hook/{name}", a.handleHook)
	mux.HandleFunc("POST /test/target/{name}", a.handleTestTarget)
	return mux
}

// httpRoutes returns the concrete HTTP paths this process serves.
func (a *App) httpRoutes() []string {
	sourceNames := make([]string, 0, len(a.sources))
	for name := range a.sources {
		sourceNames = append(sourceNames, name)
	}
	sort.Strings(sourceNames)

	targetNames := make([]string, 0, len(a.targets))
	for name := range a.targets {
		targetNames = append(targetNames, name)
	}
	sort.Strings(targetNames)

	out := make([]string, 0, 2+len(sourceNames)+len(targetNames))
	out = append(out, "GET /healthz", "GET /readyz")
	for _, name := range sourceNames {
		out = append(out, "POST /hook/"+name)
	}
	for _, name := range targetNames {
		out = append(out, "POST /test/target/"+name)
	}
	return out
}

func (a *App) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

type testTargetRequest struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

func (a *App) handleTestTarget(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	t, ok := a.targets[name]
	if !ok {
		a.log.Warn("unknown target", "target", name, "status", http.StatusNotFound)
		http.Error(w, "unknown target", http.StatusNotFound)
		return
	}

	msg := model.Message{
		Title: "HookHub test",
		Body:  fmt.Sprintf("connectivity check at %s", time.Now().UTC().Format(time.RFC3339)),
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, maxBodyBytes+1))
	if err != nil {
		a.log.Warn("read test body failed", "target", name, "err", err, "status", http.StatusBadRequest)
		http.Error(w, "read body", http.StatusBadRequest)
		return
	}
	if len(body) > maxBodyBytes {
		a.log.Warn("test body too large", "target", name, "size", len(body), "status", http.StatusRequestEntityTooLarge)
		http.Error(w, "body too large", http.StatusRequestEntityTooLarge)
		return
	}
	if len(strings.TrimSpace(string(body))) > 0 {
		var req testTargetRequest
		if err := json.Unmarshal(body, &req); err != nil {
			a.log.Warn("invalid test body", "target", name, "err", err, "status", http.StatusBadRequest)
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if title := strings.TrimSpace(req.Title); title != "" {
			msg.Title = title
		}
		if b := strings.TrimSpace(req.Body); b != "" {
			msg.Body = b
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), testTimeout)
	defer cancel()

	if err := t.Send(ctx, &msg); err != nil {
		a.log.Warn("target test failed", "target", name, "err", err, "status", http.StatusBadGateway)
		http.Error(w, "send failed: "+err.Error(), http.StatusBadGateway)
		return
	}

	a.log.Info("target test ok", "target", name, "driver", t.Driver())
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (a *App) handleHook(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	src, ok := a.sources[name]
	if !ok {
		a.log.Warn("unknown source", "source", name, "status", http.StatusNotFound)
		http.Error(w, "unknown source", http.StatusNotFound)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, maxBodyBytes+1))
	if err != nil {
		a.log.Warn("read body failed", "source", name, "err", err, "status", http.StatusBadRequest)
		http.Error(w, "read body", http.StatusBadRequest)
		return
	}
	if len(body) > maxBodyBytes {
		a.log.Warn("body too large", "source", name, "size", len(body), "limit", maxBodyBytes, "status", http.StatusRequestEntityTooLarge)
		http.Error(w, "body too large", http.StatusRequestEntityTooLarge)
		return
	}

	if err := src.Verify(r, body); err != nil {
		a.log.Warn("verify failed", "source", name, "err", err, "status", http.StatusUnauthorized)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	event, err := src.Parse(r, body)
	if err != nil {
		a.log.Warn("parse failed", "source", name, "err", err, "status", http.StatusBadRequest)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	event.Source = name
	if event.Time.IsZero() {
		event.Time = time.Now().UTC()
	}
	if event.Headers == nil {
		event.Headers = r.Header.Clone()
	}

	if r.URL.Query().Get("debug") == "1" {
		a.log.Info("debug hook",
			"source", name,
			"type", event.Type,
			"event_id", event.ID,
			"subject", event.Subject,
			"body_len", len(body),
			"body_preview", truncateForLog(body, debugBodyLogMax),
		)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("debug"))
		return
	}

	// ACK first (at-most-once): avoid upstream retry storms. Process asynchronously.
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	go a.processAsync(event)
}

func (a *App) processAsync(event *model.Event) {
	ctx, cancel := context.WithTimeout(context.Background(), processTimeout)
	defer cancel()
	a.Process(ctx, event)
}

func truncateForLog(b []byte, max int) string {
	if len(b) <= max {
		return string(b)
	}
	return string(b[:max]) + "…"
}
