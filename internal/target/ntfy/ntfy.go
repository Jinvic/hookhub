package ntfy

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/Jinvic/hookhub/internal/model"
)

const NtfyTarget = "ntfy"

type Ntfy struct {
	Config *Config
	Client *http.Client // optional; nil uses http.DefaultClient
}

func New(config *Config) *Ntfy {
	return &Ntfy{Config: config}
}

func (n *Ntfy) Name() string {
	return NtfyTarget
}

func (n *Ntfy) Send(ctx context.Context, msg *model.Message) error {
	publishURL, err := n.publishURL()
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, publishURL, strings.NewReader(msg.Body))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}

	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	if msg.Title != "" {
		req.Header.Set("Title", msg.Title)
	}
	req.Header.Set("Priority", ParsePriority(msg.Priority).String())
	if msg.Format == "markdown" {
		req.Header.Set("Markdown", "yes")
	}
	setExtraHeaders(req, msg.Extra)

	switch n.Config.AuthType {
	case AuthBasic:
		req.SetBasicAuth(n.Config.Username, n.Config.Password)
	case AuthToken:
		req.Header.Set("Authorization", "Bearer "+n.Config.Token)
	}

	client := n.Client
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
		return fmt.Errorf("ntfy %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	return nil
}

func (n *Ntfy) publishURL() (string, error) {
	base := strings.TrimRight(n.Config.URL, "/")
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("parse url: %w", err)
	}
	u = u.JoinPath(n.Config.Topic)
	return u.String(), nil
}
