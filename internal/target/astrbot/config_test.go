package astrbot

import "testing"

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{name: "ok", cfg: Config{URL: "https://astrbot.example", Token: "abk_x", UMO: "webchat:FriendMessage:x"}, wantErr: false},
		{name: "missing url", cfg: Config{Token: "t", UMO: "u"}, wantErr: true},
		{name: "bad scheme", cfg: Config{URL: "ftp://x", Token: "t", UMO: "u"}, wantErr: true},
		{name: "missing token", cfg: Config{URL: "https://x", UMO: "u"}, wantErr: true},
		{name: "missing umo", cfg: Config{URL: "https://x", Token: "t"}, wantErr: true},
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
