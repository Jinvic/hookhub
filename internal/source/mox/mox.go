package mox

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Jinvic/hookhub/internal/model"
)

// DriverMox is the plugin id used in config `driver:` and Driver().
const DriverMox = "mox"

const headerWebhookID = "X-Mox-Webhook-ID"

type Mox struct {
	Config *Config
}

func New(config *Config) *Mox {
	return &Mox{Config: config}
}

func (m *Mox) Driver() string {
	return DriverMox
}

func (m *Mox) Verify(req *http.Request, _ []byte) error {
	if m.Config.AuthType == AuthNone {
		return nil
	}
	got := req.Header.Get("Authorization")
	if !hmac.Equal([]byte(got), []byte(m.Config.Authorization)) {
		return fmt.Errorf("invalid authorization")
	}
	return nil
}

func (m *Mox) Parse(req *http.Request, body []byte) (*model.Event, error) {
	switch m.Config.Kind {
	case KindIncoming:
		return m.parseIncoming(req, body)
	case KindOutgoing:
		return m.parseOutgoing(req, body)
	default:
		return nil, fmt.Errorf("invalid kind %q", m.Config.Kind)
	}
}

func (m *Mox) parseIncoming(req *http.Request, body []byte) (*model.Event, error) {
	var probe struct {
		Event string `json:"Event"`
	}
	if err := json.Unmarshal(body, &probe); err != nil {
		return nil, fmt.Errorf("unmarshal body: %w", err)
	}
	if probe.Event != "" {
		return nil, fmt.Errorf("kind mismatch: incoming source got outgoing event %q", probe.Event)
	}

	var p Incoming
	if err := json.Unmarshal(body, &p); err != nil {
		return nil, fmt.Errorf("unmarshal incoming: %w", err)
	}

	const eventType = "incoming"
	data, err := toData(&p)
	if err != nil {
		return nil, err
	}

	var t time.Time
	if !p.Meta.Received.IsZero() {
		t = p.Meta.Received.UTC()
	} else if p.Date != nil {
		t = p.Date.UTC()
	}

	return &model.Event{
		ID:      eventID(req, eventType, body),
		Type:    eventType,
		Time:    t,
		Subject: subjectIncoming(&p),
		Data:    data,
		Raw:     body,
	}, nil
}

func (m *Mox) parseOutgoing(req *http.Request, body []byte) (*model.Event, error) {
	var p Outgoing
	if err := json.Unmarshal(body, &p); err != nil {
		return nil, fmt.Errorf("unmarshal outgoing: %w", err)
	}
	if p.Event == "" {
		return nil, fmt.Errorf("kind mismatch: outgoing source got body without Event")
	}

	eventType := "outgoing." + p.Event
	data, err := toData(&p)
	if err != nil {
		return nil, err
	}

	var t time.Time
	if !p.WebhookQueued.IsZero() {
		t = p.WebhookQueued.UTC()
	}

	return &model.Event{
		ID:      eventID(req, eventType, body),
		Type:    eventType,
		Time:    t,
		Subject: subjectOutgoing(&p),
		Data:    data,
		Raw:     body,
	}, nil
}

func eventID(req *http.Request, eventType string, body []byte) string {
	if req != nil {
		if id := req.Header.Get(headerWebhookID); id != "" {
			return id
		}
	}
	sum := sha256.Sum256(body)
	return fmt.Sprintf("%s:%s:%x", DriverMox, eventType, sum[:8])
}
