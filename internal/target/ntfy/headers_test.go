package ntfy

import (
	"net/http"
	"testing"
)

func TestSetExtraHeadersAliases(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "http://example/", nil)
	setExtraHeaders(req, map[string]string{
		"tag":      "warning",
		"in":       "30m",
		"a":        "https://example.com/a.png",
		"file":     "shot.png",
		"action":   "view, Open, https://example.com",
		"mail":     "me@example.com",
		"cache":    "no",
		"call":     "yes",
		"timezone": "Asia/Shanghai",
		"unknown":  "ignored",
	})

	checks := map[string]string{
		"Tags":     "warning",
		"Delay":    "30m",
		"Attach":   "https://example.com/a.png",
		"Filename": "shot.png",
		"Actions":  "view, Open, https://example.com",
		"Email":    "me@example.com",
		"Cache":    "no",
		"Call":     "yes",
		"Timezone": "Asia/Shanghai",
	}
	for h, want := range checks {
		if got := req.Header.Get(h); got != want {
			t.Errorf("%s = %q, want %q", h, got, want)
		}
	}
	if req.Header.Get("Unknown") != "" {
		t.Fatal("unknown Extra key must not become a header")
	}
}

func TestSetExtraHeadersPrefersCanonicalAlias(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "http://example/", nil)
	setExtraHeaders(req, map[string]string{
		"tags":  "bell",
		"tag":   "warning",
		"delay": "1h",
		"at":    "tomorrow 10am",
	})
	if got := req.Header.Get("Tags"); got != "bell" {
		t.Fatalf("Tags = %q, want bell (tags before tag)", got)
	}
	if got := req.Header.Get("Delay"); got != "1h" {
		t.Fatalf("Delay = %q, want 1h (delay before at)", got)
	}
}
