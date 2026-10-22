package gotify

import "testing"

func TestParsePriority(t *testing.T) {
	tests := []struct {
		in   string
		want int
	}{
		{in: "", want: 5},
		{in: "min", want: 1},
		{in: "low", want: 3},
		{in: "default", want: 5},
		{in: "high", want: 8},
		{in: "max", want: 10},
		{in: "0", want: 0},
		{in: "10", want: 10},
		{in: "11", want: 5},
		{in: "nope", want: 5},
	}
	for _, tt := range tests {
		if got := ParsePriority(tt.in); got != tt.want {
			t.Fatalf("%q: got %d want %d", tt.in, got, tt.want)
		}
	}
}
