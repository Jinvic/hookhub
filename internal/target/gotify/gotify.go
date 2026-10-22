package gotify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/Jinvic/hookhub/internal/model"
)

// DriverGotify is the plugin id used in config `driver:` and Driver().
const DriverGotify = "gotify"

const headerAppToken = "X-Gotify-Key"

type Gotify struct {
	Config *Config
	Client *http.Client // optional; nil uses http.DefaultClient
}

func New(config *Config) *Gotify {
	return &Gotify{Config: config}
}

func (g *Gotify) Driver() string {
	return DriverGotify
}

type pushBody struct {
	Title    string         `json:"title,omitempty"`
	Message  string         `json:"message"`
	Priority int            `json:"priority"`
	Extras   map[string]any `json:"extras,omitempty"`
}

func (g *Gotify) Send(ctx context.Context, msg *model.Message) error {
	if msg == nil || strings.TrimSpace(msg.Body) == "" {
		return fmt.Errorf("message body is required")
	}

	endpoint, err := g.messageURL()
	if err != nil {
		return err
	}

	payload := pushBody{
		Title:    msg.Title,
		Message:  msg.Body,
		Priority: ParsePriority(msg.Priority),
		Extras:   buildExtras(msg.Format, msg.Extra),
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(headerAppToken, g.Config.Token)

	client := g.Client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("gotify %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	return nil
}

func (g *Gotify) messageURL() (string, error) {
	base := strings.TrimRight(g.Config.URL, "/")
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("parse url: %w", err)
	}
	u = u.JoinPath("message")
	return u.String(), nil
}
