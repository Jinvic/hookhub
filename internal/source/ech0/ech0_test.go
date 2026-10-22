package ech0

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestParseEchoCreated(t *testing.T) {
	body := []byte(`{
		"topic": "echo.created",
		"event_name": "EchoCreated",
		"payload_raw": {"id": 42, "content": "hello"},
		"metadata": {"via": "api"},
		"occurred_at": 1710000000
	}`)

	req, _ := http.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set(headerEvent, "echo.created")
	req.Header.Set(headerEventID, "evt-123")

	ev, err := (&Ech0{}).Parse(req, body)
	if err != nil {
		t.Fatal(err)
	}
	if ev.Type != "echo.created" {
		t.Fatalf("Type = %q", ev.Type)
	}
	if ev.ID != "evt-123" {
		t.Fatalf("ID = %q", ev.ID)
	}
	if ev.Subject != "动态发布" {
		t.Fatalf("Subject = %q", ev.Subject)
	}
	if ev.Time.UTC() != time.Unix(1710000000, 0).UTC() {
		t.Fatalf("Time = %v", ev.Time)
	}
	payload, ok := ev.Data["payload_raw"].(map[string]any)
	if !ok {
		t.Fatalf("payload_raw type %T", ev.Data["payload_raw"])
	}
	if payload["content"] != "hello" {
		t.Fatalf("content = %v", payload["content"])
	}
	if ev.Data["event_name"] != "EchoCreated" {
		t.Fatalf("event_name = %v", ev.Data["event_name"])
	}
}

func TestParseWebhookTest(t *testing.T) {
	body := []byte(`{
		"topic": "webhook.test",
		"event_name": "WebhookTest",
		"payload_raw": {},
		"metadata": null,
		"occurred_at": 1710000000
	}`)
	ev, err := (&Ech0{}).Parse(nil, body)
	if err != nil {
		t.Fatal(err)
	}
	if ev.Subject != "Webhook 测试" {
		t.Fatalf("Subject = %q", ev.Subject)
	}
	if ev.ID == "" || ev.ID[:4] != "ech0" {
		t.Fatalf("ID = %q", ev.ID)
	}
}

func TestParseUnknownTopic(t *testing.T) {
	body := []byte(`{"topic":"foo.bar","event_name":"X","payload_raw":{},"occurred_at":1}`)
	_, err := (&Ech0{}).Parse(nil, body)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseTopicMismatch(t *testing.T) {
	body := []byte(`{"topic":"echo.created","event_name":"EchoCreated","payload_raw":{},"occurred_at":1}`)
	req, _ := http.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set(headerEvent, "echo.updated")
	_, err := (&Ech0{}).Parse(req, body)
	if err == nil {
		t.Fatal("expected mismatch error")
	}
}

func TestVerifyHMAC(t *testing.T) {
	body := []byte(`{"topic":"echo.created","event_name":"EchoCreated","payload_raw":{},"occurred_at":1710000000}`)
	secret := "s3cret"
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	sig := sigPrefix + hex.EncodeToString(mac.Sum(nil))

	fixed := time.Unix(1710000000, 0).UTC()
	e := New(&Config{AuthType: AuthHMAC, SecretKey: secret})
	e.now = func() time.Time { return fixed }

	req, _ := http.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set(headerSignature, sig)
	req.Header.Set(headerTimestamp, "1710000000")

	if err := e.Verify(req, body); err != nil {
		t.Fatal(err)
	}

	req.Header.Set(headerSignature, sigPrefix+"deadbeef")
	if err := e.Verify(req, body); err == nil {
		t.Fatal("expected invalid signature")
	}
}

func TestVerifyTimestampSkew(t *testing.T) {
	body := []byte(`{}`)
	secret := "s3cret"
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	sig := sigPrefix + hex.EncodeToString(mac.Sum(nil))

	fixed := time.Unix(1710000000, 0).UTC()
	e := New(&Config{AuthType: AuthHMAC, SecretKey: secret, MaxSkew: 60})
	e.now = func() time.Time { return fixed }

	req, _ := http.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set(headerSignature, sig)
	req.Header.Set(headerTimestamp, fmt.Sprintf("%d", fixed.Unix()-120))

	if err := e.Verify(req, body); err == nil {
		t.Fatal("expected skew error")
	}
}

func TestVerifyNoneSkipsChecks(t *testing.T) {
	e := New(&Config{AuthType: AuthNone})
	req, _ := http.NewRequest(http.MethodPost, "/", nil)
	if err := e.Verify(req, []byte(`{}`)); err != nil {
		t.Fatal(err)
	}
}
