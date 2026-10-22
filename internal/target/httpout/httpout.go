package httpout

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Jinvic/hookhub/internal/model"
)

// HTTP delivers a rendered Message to a fixed URL from config.
type HTTP struct {
	Config *Config
	Client *http.Client // optional; nil uses a client that does not follow redirects
}

func New(config *Config) *HTTP {
	return &HTTP{Config: config}
}

func (h *HTTP) Driver() string {
	return DriverHTTP
}

type jsonPayload struct {
	Title    string            `json:"title,omitempty"`
	Message  string            `json:"message"`
	Priority string            `json:"priority,omitempty"`
	Format   string            `json:"format,omitempty"`
	Extra    map[string]string `json:"extra,omitempty"`
}

func (h *HTTP) Send(ctx context.Context, msg *model.Message) error {
	if msg == nil || strings.TrimSpace(msg.Body) == "" {
		return fmt.Errorf("message body is required")
	}

	body, contentType, err := h.encodeBody(msg)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, h.Config.method(), h.Config.URL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", contentType)

	switch h.Config.AuthType {
	case AuthBearer:
		req.Header.Set("Authorization", "Bearer "+h.Config.Token)
	case AuthBasic:
		req.SetBasicAuth(h.Config.Username, h.Config.Password)
	case AuthHeader:
		req.Header.Set(h.Config.HeaderName, h.Config.headerValue())
	case AuthNone, "":
		// no auth
	}

	client := h.Client
	if client == nil {
		client = noRedirectClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("http %s: %s", resp.Status, strings.TrimSpace(string(respBody)))
	}
	return nil
}

func (h *HTTP) encodeBody(msg *model.Message) ([]byte, string, error) {
	switch h.Config.bodyFormat() {
	case BodyJSON:
		raw, err := json.Marshal(jsonPayload{
			Title:    msg.Title,
			Message:  msg.Body,
			Priority: msg.Priority,
			Format:   msg.Format,
			Extra:    msg.Extra,
		})
		if err != nil {
			return nil, "", fmt.Errorf("marshal body: %w", err)
		}
		return raw, "application/json", nil
	default:
		return []byte(msg.Body), "text/plain; charset=utf-8", nil
	}
}

// noRedirectClient refuses redirects so a misconfigured upstream cannot bounce to an internal URL.
var noRedirectClient = &http.Client{
	CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	},
}
