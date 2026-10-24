package astrbot

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

// See https://docs.astrbot.app/dev/openapi.html
const imMessagesPath = "/api/v1/im/messages"

type AstrBot struct {
	Config *Config
	Client *http.Client // optional; nil uses http.DefaultClient
}

func New(config *Config) *AstrBot {
	return &AstrBot{Config: config}
}

func (a *AstrBot) Driver() string {
	return DriverAstrBot
}

type imMessageRequest struct {
	UMO     string `json:"umo"`
	Message string `json:"message"`
}

func (a *AstrBot) Send(ctx context.Context, msg *model.Message) error {
	if msg == nil {
		return fmt.Errorf("message is required")
	}
	text := composePlain(msg)
	if strings.TrimSpace(text) == "" {
		return fmt.Errorf("message body is required")
	}

	endpoint, err := a.messagesURL()
	if err != nil {
		return err
	}

	raw, err := json.Marshal(imMessageRequest{
		UMO:     a.Config.UMO,
		Message: text,
	})
	if err != nil {
		return fmt.Errorf("marshal body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.Config.Token)

	client := a.Client
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
		return fmt.Errorf("astrbot %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	return nil
}

func composePlain(msg *model.Message) string {
	title := strings.TrimSpace(msg.Title)
	body := strings.TrimSpace(msg.Body)
	switch {
	case title != "" && body != "":
		return title + "\n\n" + body
	case title != "":
		return title
	default:
		return body
	}
}

func (a *AstrBot) messagesURL() (string, error) {
	base := strings.TrimRight(a.Config.URL, "/")
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("parse url: %w", err)
	}
	// Absolute path: ignore any leftover path on base URL.
	u.Path = imMessagesPath
	u.RawQuery = ""
	u.Fragment = ""
	return u.String(), nil
}
