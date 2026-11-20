package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"cel.dev/cel-go/cel"

	"github.com/Jinvic/hookhub/internal/model"
	"github.com/Jinvic/hookhub/internal/source"
	"github.com/Jinvic/hookhub/internal/target"
)

type routePrograms struct {
	match   cel.Program            // nil if match.when empty
	actions map[string]cel.Program // target name → program; missing/empty when = nil
}

type App struct {
	cfg          *Config
	sources      map[string]source.Source
	sourceVars   map[string]map[string]string // instance name → template Vars
	targets      map[string]target.Target
	whenPrograms map[string]routePrograms // route name → compiled when programs
	log          *slog.Logger
	server       *http.Server
}

func New(cfg *Config, log *slog.Logger) (*App, error) {
	if log == nil {
		log = slog.Default()
	}
	if cfg.Server.maxBodyBytes == 0 {
		n, err := parseMaxBody(cfg.Server.MaxBody)
		if err != nil {
			return nil, fmt.Errorf("server.max_body: %w", err)
		}
		cfg.Server.maxBodyBytes = n
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
	whenPrograms, err := compileRouteWhens(cfg.Routes)
	if err != nil {
		return nil, err
	}
	return &App{
		cfg:          cfg,
		sources:      sources,
		sourceVars:   sourceVars,
		targets:      targets,
		whenPrograms: whenPrograms,
		log:          log,
	}, nil
}

func compileRouteWhens(routes []model.Route) (map[string]routePrograms, error) {
	out := make(map[string]routePrograms, len(routes))
	for _, r := range routes {
		matchPrg, err := model.CompileWhen(r.Match.When)
		if err != nil {
			return nil, fmt.Errorf("route %q match.when: %w", r.Name, err)
		}
		actionPrgs := make(map[string]cel.Program, len(r.Actions))
		for _, a := range r.Actions {
			prg, err := model.CompileWhen(a.When)
			if err != nil {
				return nil, fmt.Errorf("route %q action target %q when: %w", r.Name, a.Target, err)
			}
			actionPrgs[a.Target] = prg
		}
		out[r.Name] = routePrograms{match: matchPrg, actions: actionPrgs}
	}
	return out, nil
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
