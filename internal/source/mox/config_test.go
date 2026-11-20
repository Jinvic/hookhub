package mox

import "testing"

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{name: "incoming auth ok", cfg: Config{Kind: KindIncoming, AuthType: AuthAuthorization, Authorization: "Basic x"}, wantErr: false},
		{name: "outgoing none ok", cfg: Config{Kind: KindOutgoing, AuthType: AuthNone}, wantErr: false},
		{name: "none ignores authorization", cfg: Config{Kind: KindIncoming, AuthType: AuthNone, Authorization: "x"}, wantErr: false},
		{name: "missing kind", cfg: Config{AuthType: AuthNone}, wantErr: true},
		{name: "bad kind", cfg: Config{Kind: "both", AuthType: AuthNone}, wantErr: true},
		{name: "missing auth_type", cfg: Config{Kind: KindIncoming}, wantErr: true},
		{name: "auth missing value", cfg: Config{Kind: KindIncoming, AuthType: AuthAuthorization}, wantErr: true},
		{name: "unknown auth", cfg: Config{Kind: KindIncoming, AuthType: "hmac"}, wantErr: true},
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
