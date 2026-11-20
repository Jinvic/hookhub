package app

import (
	"fmt"
	"strconv"
	"strings"
)

const defaultMaxBody = 1 << 20 // 1m

// parseMaxBody parses a human body-size string (1024-based).
// Empty means default 1m. Accepts bare bytes or a suffix b/k/m/g (case-insensitive).
func parseMaxBody(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return defaultMaxBody, nil
	}

	s = strings.ToLower(s)
	mult := int64(1)
	switch {
	case strings.HasSuffix(s, "b"):
		s = strings.TrimSuffix(s, "b")
	case strings.HasSuffix(s, "k"):
		mult = 1 << 10
		s = strings.TrimSuffix(s, "k")
	case strings.HasSuffix(s, "m"):
		mult = 1 << 20
		s = strings.TrimSuffix(s, "m")
	case strings.HasSuffix(s, "g"):
		mult = 1 << 30
		s = strings.TrimSuffix(s, "g")
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("invalid max_body: missing number")
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid max_body %q: %w", s, err)
	}
	if n <= 0 {
		return 0, fmt.Errorf("max_body must be > 0")
	}
	out := n * mult
	if out > int64(^uint(0)>>1) {
		return 0, fmt.Errorf("max_body too large")
	}
	return int(out), nil
}
