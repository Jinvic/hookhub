package ntfy

import "testing"

func TestConfigValidate(t *testing.T) {
	base := Config{URL: "https://ntfy.sh", Topic: "work"}

	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{name: "none ok", cfg: Config{URL: base.URL, Topic: base.Topic, AuthType: AuthNone}, wantErr: false},
		{name: "basic ok", cfg: Config{URL: base.URL, Topic: base.Topic, AuthType: AuthBasic, Username: "u", Password: "p"}, wantErr: false},
		{name: "token ok", cfg: Config{URL: base.URL, Topic: base.Topic, AuthType: AuthToken, Token: "t"}, wantErr: false},
		{name: "missing url", cfg: Config{Topic: base.Topic, AuthType: AuthNone}, wantErr: true},
		{name: "missing topic", cfg: Config{URL: base.URL, AuthType: AuthNone}, wantErr: true},
		{name: "empty auth", cfg: Config{URL: base.URL, Topic: base.Topic}, wantErr: true},
		{name: "unknown auth", cfg: Config{URL: base.URL, Topic: base.Topic, AuthType: "hmac"}, wantErr: true},
		{name: "basic missing user", cfg: Config{URL: base.URL, Topic: base.Topic, AuthType: AuthBasic, Password: "p"}, wantErr: true},
		{name: "basic missing pass", cfg: Config{URL: base.URL, Topic: base.Topic, AuthType: AuthBasic, Username: "u"}, wantErr: true},
		{name: "token missing", cfg: Config{URL: base.URL, Topic: base.Topic, AuthType: AuthToken}, wantErr: true},
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
