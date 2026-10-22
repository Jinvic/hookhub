package gotify

import "testing"

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{name: "ok", cfg: Config{URL: "https://push.example", Token: "t"}, wantErr: false},
		{name: "missing url", cfg: Config{Token: "t"}, wantErr: true},
		{name: "missing token", cfg: Config{URL: "https://push.example"}, wantErr: true},
		{name: "empty", cfg: Config{}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatal(err)
			}
		})
	}
}
