package vikunja

import "testing"

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{name: "none ok", cfg: Config{AuthType: AuthNone}, wantErr: false},
		{name: "none ignores secret", cfg: Config{AuthType: AuthNone, SecretKey: "x"}, wantErr: false},
		{name: "hmac ok", cfg: Config{AuthType: AuthHMAC, SecretKey: "s"}, wantErr: false},
		{name: "hmac missing secret", cfg: Config{AuthType: AuthHMAC}, wantErr: true},
		{name: "empty auth", cfg: Config{}, wantErr: true},
		{name: "unknown auth", cfg: Config{AuthType: "basic"}, wantErr: true},
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
