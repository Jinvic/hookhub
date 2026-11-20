package app

import "testing"

func TestParseMaxBody(t *testing.T) {
	tests := []struct {
		in      string
		want    int
		wantErr bool
	}{
		{in: "", want: defaultMaxBody},
		{in: "  ", want: defaultMaxBody},
		{in: "512", want: 512},
		{in: "512b", want: 512},
		{in: "512B", want: 512},
		{in: "1k", want: 1 << 10},
		{in: "512k", want: 512 << 10},
		{in: "1m", want: 1 << 20},
		{in: "2M", want: 2 << 20},
		{in: "1g", want: 1 << 30},
		{in: "0", wantErr: true},
		{in: "0m", wantErr: true},
		{in: "m", wantErr: true},
		{in: "1x", wantErr: true},
		{in: "-1k", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := parseMaxBody(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("got %d want %d", got, tt.want)
			}
		})
	}
}
