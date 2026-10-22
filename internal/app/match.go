package app

import "github.com/Jinvic/hookhub/internal/model"

// MatchingRoutes returns every route whose Match accepts e (fan-out: all hits).
// Single-rule logic lives on model.Match.Matches; this file only walks the table.
func MatchingRoutes(routes []model.Route, e *model.Event) []model.Route {
	var out []model.Route
	for i := range routes {
		if routes[i].Match.Matches(e) {
			out = append(out, routes[i])
		}
	}
	return out
}
