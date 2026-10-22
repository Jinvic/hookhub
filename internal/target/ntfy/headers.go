package ntfy

import (
	"net/http"
	"strings"
)

// extraHeaderPrefs maps Extra keys (aliases, preferred first) to ntfy HTTP headers.
// Only whitelisted keys are forwarded; unknown Extra entries are ignored.
// See https://docs.ntfy.sh/publish/#list-of-all-parameters
var extraHeaderPrefs = []struct {
	keys   []string
	header string
}{
	{[]string{"tags", "tag"}, "Tags"},
	{[]string{"delay", "at", "in"}, "Delay"},
	{[]string{"timezone"}, "Timezone"},
	{[]string{"click"}, "Click"},
	{[]string{"icon"}, "Icon"},
	{[]string{"attach", "a"}, "Attach"},
	{[]string{"filename", "file"}, "Filename"},
	{[]string{"actions", "action"}, "Actions"},
	{[]string{"email", "mail"}, "Email"},
	{[]string{"call"}, "Call"},
	{[]string{"cache"}, "Cache"},
}

func setExtraHeaders(req *http.Request, extra map[string]string) {
	if extra == nil {
		return
	}
	for _, p := range extraHeaderPrefs {
		for _, k := range p.keys {
			if v := strings.TrimSpace(extra[k]); v != "" {
				req.Header.Set(p.header, v)
				break
			}
		}
	}
}
