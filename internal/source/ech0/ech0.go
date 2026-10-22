package ech0

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Jinvic/hookhub/internal/model"
)

// DriverEch0 is the plugin id used in config `driver:` and Driver().
const DriverEch0 = "ech0"

const (
	headerEvent     = "X-Ech0-Event"
	headerEventID   = "X-Ech0-Event-ID"
	headerTimestamp = "X-Ech0-Timestamp"
	headerSignature = "X-Ech0-Signature"
	sigPrefix       = "sha256="
)

type Ech0 struct {
	Config *Config
	// now is used by tests; nil means time.Now.
	now func() time.Time
}

func New(config *Config) *Ech0 {
	return &Ech0{Config: config}
}

func (e *Ech0) Driver() string {
	return DriverEch0
}

func (e *Ech0) Verify(req *http.Request, body []byte) error {
	if e.Config.AuthType == AuthNone {
		return nil
	}

	if err := e.verifyTimestamp(req); err != nil {
		return err
	}

	sigHeader := req.Header.Get(headerSignature)
	if !strings.HasPrefix(sigHeader, sigPrefix) {
		return fmt.Errorf("invalid signature header")
	}
	received := sigHeader[len(sigPrefix):]

	mac := hmac.New(sha256.New, []byte(e.Config.SecretKey))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(received), []byte(expected)) {
		return fmt.Errorf("invalid signature")
	}
	return nil
}

func (e *Ech0) verifyTimestamp(req *http.Request) error {
	skew := e.Config.skewSeconds()
	if skew < 0 {
		return nil
	}

	raw := req.Header.Get(headerTimestamp)
	if raw == "" {
		return fmt.Errorf("missing %s", headerTimestamp)
	}
	sec, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid %s: %w", headerTimestamp, err)
	}

	ts := time.Unix(sec, 0).UTC()
	now := e.clock()
	diff := now.Sub(ts)
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Duration(skew)*time.Second {
		return fmt.Errorf("timestamp outside allowed skew (%ds)", skew)
	}
	return nil
}

func (e *Ech0) Parse(req *http.Request, body []byte) (*model.Event, error) {
	var p Payload
	if err := json.Unmarshal(body, &p); err != nil {
		return nil, fmt.Errorf("unmarshal body: %w", err)
	}

	if req != nil {
		if hdr := req.Header.Get(headerEvent); hdr != "" {
			if p.Topic == "" {
				p.Topic = hdr
			} else if p.Topic != hdr {
				return nil, fmt.Errorf("topic mismatch: body %q vs header %q", p.Topic, hdr)
			}
		}
	}

	data, err := p.ParseData()
	if err != nil {
		return nil, err
	}

	id := ""
	if req != nil {
		id = req.Header.Get(headerEventID)
	}
	if id == "" {
		sum := sha256.Sum256(body)
		id = fmt.Sprintf("%s:%s:%x", DriverEch0, p.Topic, sum[:8])
	}

	var t time.Time
	if p.OccurredAt > 0 {
		t = time.Unix(p.OccurredAt, 0).UTC()
	}

	return &model.Event{
		ID: id,
		// Source is filled by the HTTP handler with the config instance name.
		Type:    p.Topic,
		Time:    t,
		Subject: subjectFor(p.Topic),
		Data:    data,
		Raw:     body,
	}, nil
}

func (e *Ech0) clock() time.Time {
	if e.now != nil {
		return e.now().UTC()
	}
	return time.Now().UTC()
}
