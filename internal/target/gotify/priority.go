package gotify

import "strconv"

// Gotify priorities are 0–10. See https://gotify.net/docs/priority
// Named values map from HookHub's channel-agnostic hints (same names as ntfy).
func ParsePriority(s string) int {
	if s == "" {
		return 5 // mid-range default when template omits priority
	}
	switch s {
	case "min":
		return 1
	case "low":
		return 3
	case "default":
		return 5
	case "high":
		return 8
	case "max":
		return 10
	default:
		p, err := strconv.Atoi(s)
		if err != nil || p < 0 || p > 10 {
			return 5
		}
		return p
	}
}
