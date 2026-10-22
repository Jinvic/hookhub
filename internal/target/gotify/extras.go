package gotify

import "strings"

// buildExtras maps Message.Format / Extra into Gotify extras namespaces.
// See https://gotify.net/docs/msgextras
func buildExtras(format string, extra map[string]string) map[string]any {
	out := map[string]any{}

	contentType := "text/plain"
	if format == "markdown" {
		contentType = "text/markdown"
	}
	out["client::display"] = map[string]any{
		"contentType": contentType,
	}

	notif := map[string]any{}
	if v := firstExtra(extra, "click"); v != "" {
		notif["click"] = map[string]any{"url": v}
	}
	if v := firstExtra(extra, "big_image", "bigImageUrl", "bigimage"); v != "" {
		notif["bigImageUrl"] = v
	}
	if len(notif) > 0 {
		out["client::notification"] = notif
	}

	if v := firstExtra(extra, "intent", "intent_url", "intentUrl"); v != "" {
		out["android::action"] = map[string]any{
			"onReceive": map[string]any{"intentUrl": v},
		}
	}

	return out
}

func firstExtra(extra map[string]string, keys ...string) string {
	if extra == nil {
		return ""
	}
	for _, k := range keys {
		if v := strings.TrimSpace(extra[k]); v != "" {
			return v
		}
	}
	return ""
}
