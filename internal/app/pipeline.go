package app

import (
	"context"

	"github.com/Jinvic/hookhub/internal/model"
)

// Process matches routes for e, renders templates, and sends to targets.
// Failures on a single action are logged and do not stop other actions.
func (a *App) Process(ctx context.Context, e *model.Event) {
	vars := a.sourceVars[e.Source]
	if vars == nil {
		vars = map[string]string{}
	}

	matched := 0
	for _, route := range a.cfg.Routes {
		if !route.Match.Matches(e) {
			continue
		}

		rp := a.whenPrograms[route.Name]
		ok, err := model.EvalWhen(rp.match, e, vars)
		if err != nil {
			a.log.Warn("match.when eval failed",
				"event_id", e.ID,
				"route", route.Name,
				"err", err,
			)
			continue
		}
		if !ok {
			a.log.Info("match.when skipped",
				"event_id", e.ID,
				"route", route.Name,
			)
			continue
		}
		matched++

		for _, action := range route.Actions {
			ok, err := model.EvalWhen(rp.actions[action.Target], e, vars)
			if err != nil {
				a.log.Warn("action.when eval failed",
					"event_id", e.ID,
					"route", route.Name,
					"target", action.Target,
					"err", err,
				)
				continue
			}
			if !ok {
				a.log.Info("action.when skipped",
					"event_id", e.ID,
					"route", route.Name,
					"target", action.Target,
				)
				continue
			}
			a.dispatchAction(ctx, e, route.Name, action)
		}
	}

	if matched == 0 {
		a.log.Info("no matching routes",
			"event_id", e.ID,
			"source", e.Source,
			"type", e.Type,
		)
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
