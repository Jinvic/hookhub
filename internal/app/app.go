package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/Jinvic/hookhub/internal/source"
	"github.com/Jinvic/hookhub/internal/target"
)

type App struct {
	cfg        *Config
	sources    map[string]source.Source
	sourceVars map[string]map[string]string // instance name → template Vars
	targets    map[string]target.Target
	log        *slog.Logger
	server     *http.Server
}

func New(cfg *Config, log *slog.Logger) (*App, error) {
	if log == nil {
		log = slog.Default()
	}
	sources, err := buildSources(cfg.Sources)
	if err != nil {
		return nil, err
	}
	targets, err := buildTargets(cfg.Targets)
	if err != nil {
		return nil, err
	}
	sourceVars := make(map[string]map[string]string, len(cfg.Sources))
	for _, s := range cfg.Sources {
		vars := s.Vars
		if vars == nil {
			vars = map[string]string{}
		}
		sourceVars[s.Name] = vars
	}
	return &App{
		cfg:        cfg,
		sources:    sources,
		sourceVars: sourceVars,
		targets:    targets,
		log:        log,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	a.server = &http.Server{
		Addr:              a.cfg.Server.Addr,
		Handler:           a.routes(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		a.log.Info("listening", "addr", a.cfg.Server.Addr, "routes", a.httpRoutes())
		err := a.server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := a.server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown: %w", err)
		}
		return <-errCh
	case err := <-errCh:
		return err
	}
}
