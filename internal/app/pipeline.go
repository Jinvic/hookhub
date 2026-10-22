package app

import (
	"context"

	"github.com/Jinvic/hookhub/internal/model"
)

// Process matches routes for e, renders templates, and sends to targets.
// Failures on a single action are logged and do not stop other actions.
func (a *App) Process(ctx context.Context, e *model.Event) {
	routes := MatchingRoutes(a.cfg.Routes, e)
	if len(routes) == 0 {
		a.log.Info("no matching routes",
			"event_id", e.ID,
			"source", e.Source,
			"type", e.Type,
		)
		return
	}

	for _, route := range routes {
		for _, action := range route.Actions {
			a.dispatchAction(ctx, e, route.Name, action)
		}
	}
}

func (a *App) dispatchAction(ctx context.Context, e *model.Event, routeName string, action model.Action) {
	msg, err := action.Template.Render(model.RenderContext{
		Event: e,
		Vars:  a.sourceVars[e.Source],
	})
	if err != nil {
		a.log.Error("render failed",
			"event_id", e.ID,
			"route", routeName,
			"target", action.Target,
			"err", err,
		)
		return
	}

	t, ok := a.targets[action.Target]
	if !ok {
		a.log.Error("target not found",
			"event_id", e.ID,
			"route", routeName,
			"target", action.Target,
		)
		return
	}

	if err := t.Send(ctx, &msg); err != nil {
		a.log.Error("send failed",
			"event_id", e.ID,
			"route", routeName,
			"target", action.Target,
			"err", err,
		)
		return
	}

	a.log.Info("sent",
		"event_id", e.ID,
		"route", routeName,
		"target", action.Target,
		"source", e.Source,
		"type", e.Type,
	)
}
