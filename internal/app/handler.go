package app

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/Jinvic/hookhub/internal/model"
)

const (
	maxBodyBytes    = 1 << 20 // 1 MiB
	debugBodyLogMax = 512
	processTimeout  = 30 * time.Second
)

func (a *App) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", a.handleHealthz)
	mux.HandleFunc("GET /readyz", a.handleHealthz)
	mux.HandleFunc("POST /hook/{name}", a.handleHook)
	return mux
}

func (a *App) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (a *App) handleHook(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	src, ok := a.sources[name]
	if !ok {
		http.Error(w, "unknown source", http.StatusNotFound)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, maxBodyBytes+1))
	if err != nil {
		http.Error(w, "read body", http.StatusBadRequest)
		return
	}
	if len(body) > maxBodyBytes {
		http.Error(w, "body too large", http.StatusRequestEntityTooLarge)
		return
	}

	if err := src.Verify(r, body); err != nil {
		a.log.Warn("verify failed", "source", name, "err", err)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	event, err := src.Parse(r, body)
	if err != nil {
		a.log.Warn("parse failed", "source", name, "err", err)
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
