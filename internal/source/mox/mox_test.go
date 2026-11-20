package mox

import (
	"net/http"
	"testing"
	"time"
)

func TestVerifyAuthorization(t *testing.T) {
	m := New(&Config{Kind: KindIncoming, AuthType: AuthAuthorization, Authorization: "Basic dGVzdA=="})
	req, _ := http.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Authorization", "Basic dGVzdA==")
	if err := m.Verify(req, nil); err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Authorization", "Basic wrong")
	if err := m.Verify(req, nil); err == nil {
		t.Fatal("expected error")
	}
}

func TestVerifyNone(t *testing.T) {
	m := New(&Config{Kind: KindIncoming, AuthType: AuthNone})
	req, _ := http.NewRequest(http.MethodPost, "/", nil)
	if err := m.Verify(req, nil); err != nil {
		t.Fatal(err)
	}
}

func TestParseIncoming(t *testing.T) {
	body := []byte(`{
		"Version": 0,
		"From": [{"Name": "", "Address": "mox@localhost"}],
		"To": [{"Name": "", "Address": "mjl@localhost"}],
		"CC": [],
		"BCC": [],
		"ReplyTo": [],
		"Subject": "hi",
		"MessageID": "<QnxzgulZK51utga6agH_rg@mox.example>",
		"InReplyTo": "",
		"References": [],
		"Date": "2024-03-27T00:00:00Z",
		"Text": "hello world\n",
		"HTML": "",
		"Structure": {
			"ContentType": "text/plain",
			"ContentTypeParams": {"charset": "utf-8"},
			"ContentID": "",
			"ContentDisposition": "",
			"Filename": "",
			"DecodedSize": 12,
			"Parts": []
		},
		"Meta": {
			"MsgID": 201,
			"MailFrom": "mox@localhost",
			"MailFromValidated": false,
			"MsgFromValidated": true,
			"RcptTo": "mjl@localhost",
			"DKIMVerifiedDomains": ["localhost"],
			"RemoteIP": "127.0.0.1",
			"Received": "2024-03-27T00:00:03Z",
			"MailboxName": "Inbox",
			"Automated": false
		}
	}`)

	req, _ := http.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set(headerWebhookID, "123")

	ev, err := New(&Config{Kind: KindIncoming, AuthType: AuthNone}).Parse(req, body)
	if err != nil {
		t.Fatal(err)
	}
	if ev.Type != "incoming" {
		t.Fatalf("Type = %q", ev.Type)
	}
	if ev.ID != "123" {
		t.Fatalf("ID = %q", ev.ID)
	}
	if ev.Subject != "hi" {
		t.Fatalf("Subject = %q", ev.Subject)
	}
	wantTime := time.Date(2024, 3, 27, 0, 0, 3, 0, time.UTC)
	if !ev.Time.Equal(wantTime) {
		t.Fatalf("Time = %v", ev.Time)
	}
	meta, ok := ev.Data["Meta"].(map[string]any)
	if !ok {
		t.Fatalf("Meta type %T", ev.Data["Meta"])
	}
	if meta["MsgID"] != float64(201) {
		t.Fatalf("MsgID = %v", meta["MsgID"])
	}
}

func TestParseOutgoingFailed(t *testing.T) {
	body := []byte(`{
		"Version": 0,
		"Event": "failed",
		"DSN": true,
		"Suppressing": true,
		"QueueMsgID": 102,
		"FromID": "MDEyMzQ1Njc4OWFiY2RlZg",
		"MessageID": "<QnxzgulZK51utga6agH_rg@mox.example>",
		"Subject": "subject of original message",
		"WebhookQueued": "2024-03-27T00:00:00Z",
		"SMTPCode": 554,
		"SMTPEnhancedCode": "5.4.0",
		"Error": "timeout connecting to host",
		"Extra": {"userid": "456"}
	}`)

	ev, err := New(&Config{Kind: KindOutgoing, AuthType: AuthNone}).Parse(nil, body)
	if err != nil {
		t.Fatal(err)
	}
	if ev.Type != "outgoing.failed" {
		t.Fatalf("Type = %q", ev.Type)
	}
	if ev.Subject != "subject of original message" {
		t.Fatalf("Subject = %q", ev.Subject)
	}
	if ev.Data["Suppressing"] != true {
		t.Fatalf("Suppressing = %v", ev.Data["Suppressing"])
	}
	if ev.Data["Error"] != "timeout connecting to host" {
		t.Fatalf("Error = %v", ev.Data["Error"])
	}
	if ev.ID == "" || ev.ID[:3] != "mox" {
		t.Fatalf("ID = %q", ev.ID)
	}
}

func TestParseKindMismatch(t *testing.T) {
	incomingBody := []byte(`{
		"Version": 0,
		"From": [{"Address": "a@b.c"}],
		"Subject": "x",
		"Meta": {"MsgID": 1, "Received": "2024-03-27T00:00:00Z"}
	}`)
	_, err := New(&Config{Kind: KindOutgoing, AuthType: AuthNone}).Parse(nil, incomingBody)
	if err == nil {
		t.Fatal("expected outgoing/incoming mismatch error")
	}

	outgoingBody := []byte(`{"Version":0,"Event":"delivered","Subject":"ok","WebhookQueued":"2024-03-27T00:00:00Z"}`)
	_, err = New(&Config{Kind: KindIncoming, AuthType: AuthNone}).Parse(nil, outgoingBody)
	if err == nil {
		t.Fatal("expected incoming/outgoing mismatch error")
	}
}

func TestParseIncomingSubjectFallback(t *testing.T) {
	body := []byte(`{
		"Version": 0,
		"From": [{"Address": "from@example.com"}],
		"Subject": "",
		"Meta": {"MsgID": 1, "Received": "2024-03-27T00:00:00Z"}
	}`)
	ev, err := New(&Config{Kind: KindIncoming, AuthType: AuthNone}).Parse(nil, body)
	if err != nil {
		t.Fatal(err)
	}
	if ev.Subject != "from@example.com" {
		t.Fatalf("Subject = %q", ev.Subject)
	}
}
