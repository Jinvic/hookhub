package model

import "strings"

// Matches reports whether e satisfies this rule.
// Source: empty means any; otherwise exact instance name.
// Type: empty means any; otherwise '.'-segment match where '*' is exactly one segment.
func (m Match) Matches(e *Event) bool {
	if e == nil {
		return false
	}
	if m.Source != "" && e.Source != m.Source {
		return false
	}
	if m.Type != "" && !matchTypePattern(m.Type, e.Type) {
		return false
	}
	return true
}

func matchTypePattern(pattern, value string) bool {
	pp := strings.Split(pattern, ".")
	vv := strings.Split(value, ".")
	if len(pp) != len(vv) {
		return false
	}
	for i := range pp {
		if pp[i] == "*" {
			continue
		}
		if pp[i] != vv[i] {
			return false
		}
	}
	return true
}
