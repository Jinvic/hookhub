package vikunja

import "fmt"

func subjectFor(name EventName, data map[string]any) string {
	switch name {
	case EventTaskReminderFired:
		if title := nestedString(data, "task", "title"); title != "" {
			return fmt.Sprintf("任务提醒: %s", title)
		}
		return "任务提醒"
	case EventTaskOverdue:
		if title := nestedString(data, "task", "title"); title != "" {
			return fmt.Sprintf("任务逾期: %s", title)
		}
		return "任务逾期"
	case EventTasksOverdue:
		tasks, _ := data["tasks"].([]any)
		n := len(tasks)
		if n == 0 {
			return "任务逾期: 0 个任务"
		}
		title := ""
		if first, ok := tasks[0].(map[string]any); ok {
			title, _ = first["title"].(string)
		}
		if title == "" {
			return fmt.Sprintf("任务逾期: %d 个任务", n)
		}
		if n == 1 {
			return fmt.Sprintf("任务逾期: %s", title)
		}
		return fmt.Sprintf("任务逾期: %s 等 %d 个任务", title, n)
	default:
		return ""
	}
}

func nestedString(m map[string]any, keys ...string) string {
	cur := any(m)
	for _, k := range keys {
		obj, ok := cur.(map[string]any)
		if !ok {
			return ""
		}
		cur = obj[k]
	}
	s, _ := cur.(string)
	return s
}
