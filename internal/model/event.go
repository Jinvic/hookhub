package model

import (
	"net/http"
	"time"
)

type Event struct {
	ID      string         // 幂等键：source + type + payload hash
	Source  string         // 配置实例名（由 HTTP handler 写入，非 driver）
	Type    string         // 事件种类，如 "task.reminder.fired"（非插件 driver）
	Time    time.Time      // 事件发生时间（优先取 payload，否则接收时间）
	Subject string         // 可选摘要，如 task.title
	Data    map[string]any // 解析后的结构化数据
	Raw     []byte         // 原始 body，供模板兜底与调试
	Headers http.Header
}
