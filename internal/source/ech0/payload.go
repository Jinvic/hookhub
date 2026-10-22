package ech0

import (
	"encoding/json"
	"fmt"
)

// Known topics from https://ech0.app/docs/guide/webhook
const (
	TopicUserCreated               = "user.created"
	TopicUserUpdated               = "user.updated"
	TopicUserDeleted               = "user.deleted"
	TopicEchoCreated               = "echo.created"
	TopicEchoUpdated               = "echo.updated"
	TopicEchoDeleted               = "echo.deleted"
	TopicCommentCreated            = "comment.created"
	TopicCommentStatusUpdated      = "comment.status.updated"
	TopicCommentDeleted            = "comment.deleted"
	TopicResourceUploaded          = "resource.uploaded"
	TopicSystemSnapshot            = "system.snapshot"
	TopicSystemExport              = "system.export"
	TopicSystemSnapshotSchedUpdate = "system.snapshot_schedule.updated"
	TopicWebhookTest               = "webhook.test"
)

var knownTopics = map[string]struct{}{
	TopicUserCreated:               {},
	TopicUserUpdated:               {},
	TopicUserDeleted:               {},
	TopicEchoCreated:               {},
	TopicEchoUpdated:               {},
	TopicEchoDeleted:               {},
	TopicCommentCreated:            {},
	TopicCommentStatusUpdated:      {},
	TopicCommentDeleted:            {},
	TopicResourceUploaded:          {},
	TopicSystemSnapshot:            {},
	TopicSystemExport:              {},
	TopicSystemSnapshotSchedUpdate: {},
	TopicWebhookTest:               {},
}

type Payload struct {
	Topic      string          `json:"topic"`
	EventName  string          `json:"event_name"`
	PayloadRaw json.RawMessage `json:"payload_raw"`
	Metadata   json.RawMessage `json:"metadata"`
	OccurredAt int64           `json:"occurred_at"`
}

func (p *Payload) ParseData() (map[string]any, error) {
	if p.Topic == "" {
		return nil, fmt.Errorf("missing topic")
	}
	if _, ok := knownTopics[p.Topic]; !ok {
		return nil, fmt.Errorf("unknown topic: %q", p.Topic)
	}

	data := map[string]any{
		"event_name": p.EventName,
	}

	payload, err := unmarshalObjectOrEmpty(p.PayloadRaw)
	if err != nil {
		return nil, fmt.Errorf("unmarshal payload_raw: %w", err)
	}
	data["payload_raw"] = payload

	meta, err := unmarshalObjectOrEmpty(p.Metadata)
	if err != nil {
		return nil, fmt.Errorf("unmarshal metadata: %w", err)
	}
	data["metadata"] = meta

	return data, nil
}

func unmarshalObjectOrEmpty(raw json.RawMessage) (map[string]any, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return map[string]any{}, nil
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	if m == nil {
		m = map[string]any{}
	}
	return m, nil
}
