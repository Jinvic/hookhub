package ntfy

import (
	"strconv"
)

type Priority int

const (
	PriorityMin Priority = iota + 1
	PriorityLow
	PriorityDefault
	PriorityHigh
	PriorityMax
)

func ParsePriority(s string) Priority {
	if s == "" {
		return PriorityDefault
	}

	switch s {
	case "min":
		return PriorityMin
	case "low":
		return PriorityLow
	case "default":
		return PriorityDefault
	case "high":
		return PriorityHigh
	case "max":
		return PriorityMax
	default:
		p, err := strconv.Atoi(s)
		if err != nil {
			return PriorityDefault
		}
		if p < 1 || p > 5 {
			return PriorityDefault
		}
		return Priority(p)
	}

}

func (p Priority) String() string {
	return strconv.Itoa(int(p))
}
