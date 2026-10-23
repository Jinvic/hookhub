package model

import (
	"fmt"
	"text/template"
	"time"
)

const defaultTimeLayout = "2006-01-02 15:04"

func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"formatTime": formatTime,
	}
}

// formatTime formats a time value in the given IANA timezone.
//
//	{{ formatTime .Event.Data.reminder.reminder "Asia/Shanghai" }}
//	{{ formatTime .Event.Data.reminder.reminder .Vars.timezone }}
//	{{ formatTime .Event.Time "Asia/Shanghai" "2006-01-02 15:04:05" }}
//
// v may be time.Time or an RFC3339 / RFC3339Nano string (as in Vikunja JSON).
// layout defaults to "2006-01-02 15:04". Empty timezone means UTC.
func formatTime(v any, timezone string, layout ...string) (string, error) {
	t, err := parseTimeValue(v)
	if err != nil {
		return "", err
	}

	loc := time.UTC
	if timezone != "" {
		loc, err = time.LoadLocation(timezone)
		if err != nil {
			return "", fmt.Errorf("timezone %q: %w", timezone, err)
		}
	}

	l := defaultTimeLayout
	if len(layout) > 0 && layout[0] != "" {
		l = layout[0]
	}
	return t.In(loc).Format(l), nil
}

func parseTimeValue(v any) (time.Time, error) {
	switch x := v.(type) {
	case time.Time:
		return x, nil
	case string:
		if x == "" {
			return time.Time{}, fmt.Errorf("empty time string")
		}
		if t, err := time.Parse(time.RFC3339Nano, x); err == nil {
			return t, nil
		}
		if t, err := time.Parse(time.RFC3339, x); err == nil {
			return t, nil
		}
		return time.Time{}, fmt.Errorf("parse time %q: want RFC3339", x)
	case nil:
		return time.Time{}, fmt.Errorf("nil time")
	default:
		return time.Time{}, fmt.Errorf("unsupported time type %T", v)
	}
}
