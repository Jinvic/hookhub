package mox

import (
	"encoding/json"
	"fmt"
	"time"
)

// Outgoing event names from mox webhook.OutgoingEvent (forward-compatible: unknown accepted).
const (
	EventDelivered    = "delivered"
	EventSuppressed   = "suppressed"
	EventDelayed      = "delayed"
	EventFailed       = "failed"
	EventRelayed      = "relayed"
	EventExpanded     = "expanded"
	EventCanceled     = "canceled"
	EventUnrecognized = "unrecognized"
)

// DTOs mirror github.com/mjl-/mox/webhook JSON (PascalCase; no module dependency).

type Outgoing struct {
	Version          int
	Event            string
	DSN              bool
	Suppressing      bool
	QueueMsgID       int64
	FromID           string
	MessageID        string
	Subject          string
	WebhookQueued    time.Time
	SMTPCode         int
	SMTPEnhancedCode string
	Error            string
	Extra            map[string]string
}

type Incoming struct {
	Version    int
	From       []NameAddress
	To         []NameAddress
	CC         []NameAddress
	BCC        []NameAddress
	ReplyTo    []NameAddress
	Subject    string
	MessageID  string
	InReplyTo  string
	References []string
	Date       *time.Time
	Text       string
	HTML       string
	Structure  Structure
	Meta       IncomingMeta
}

type IncomingMeta struct {
	MsgID               int64
	MailFrom            string
	MailFromValidated   bool
	MsgFromValidated    bool
	RcptTo              string
	DKIMVerifiedDomains []string
	RemoteIP            string
	Received            time.Time
	MailboxName         string
	Automated           bool
}

type NameAddress struct {
	Name    string
	Address string
}

type Structure struct {
	ContentType        string
	ContentTypeParams  map[string]string
	ContentID          string
	ContentDisposition string
	Filename           string
	DecodedSize        int64
	Parts              []Structure
}

func toData(v any) (map[string]any, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}
	var data map[string]any
	if err := json.Unmarshal(b, &data); err != nil {
		return nil, fmt.Errorf("unmarshal payload map: %w", err)
	}
	if data == nil {
		data = map[string]any{}
	}
	return data, nil
}
