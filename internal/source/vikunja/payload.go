package vikunja

import (
	"encoding/json"
	"fmt"
	"time"
)

type EventName string

const (
	EventTaskOverdue       EventName = "task.overdue"
	EventTaskReminderFired EventName = "task.reminder.fired"
	EventTasksOverdue      EventName = "tasks.overdue"
)

type Payload struct {
	EventName EventName       `json:"event_name"`
	Time      time.Time       `json:"time"`
	Data      json.RawMessage `json:"data"`
}

func (p *Payload) ParseData() (map[string]any, error) {
	switch p.EventName {
	case EventTaskOverdue, EventTaskReminderFired, EventTasksOverdue:
		// ok
	default:
		return nil, fmt.Errorf("unknown event: %q", p.EventName)
	}

	var data map[string]any
	if err := json.Unmarshal(p.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal data: %w", err)
	}
	if data == nil {
		data = map[string]any{}
	}
	return data, nil
}
