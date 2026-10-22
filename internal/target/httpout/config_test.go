package httpout

import "testing"

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{name: "plain ok", cfg: Config{URL: "https://hooks.example/x", AuthType: AuthNone}, wantErr: false},
		{name: "empty auth ok", cfg: Config{URL: "https://hooks.example/x"}, wantErr: false},
		{name: "json ok", cfg: Config{URL: "https://hooks.example/x", BodyFormat: BodyJSON}, wantErr: false},
		{name: "bearer ok", cfg: Config{URL: "https://hooks.example/x", AuthType: AuthBearer, Token: "t"}, wantErr: false},
		{name: "bearer missing token", cfg: Config{URL: "https://hooks.example/x", AuthType: AuthBearer}, wantErr: true},
		{name: "basic ok", cfg: Config{URL: "https://hooks.example/x", AuthType: AuthBasic, Username: "u", Password: "p"}, wantErr: false},
		{name: "header ok", cfg: Config{URL: "https://hooks.example/x", AuthType: AuthHeader, HeaderName: "X-Token", Token: "t"}, wantErr: false},
		{name: "header missing name", cfg: Config{URL: "https://hooks.example/x", AuthType: AuthHeader, Token: "t"}, wantErr: true},
		{name: "missing url", cfg: Config{}, wantErr: true},
		{name: "bad scheme", cfg: Config{URL: "ftp://x"}, wantErr: true},
		{name: "bad method", cfg: Config{URL: "https://x", Method: "GET"}, wantErr: true},
		{name: "bad body", cfg: Config{URL: "https://x", BodyFormat: "xml"}, wantErr: true},
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
