# HookHub

> Receive webhooks from anywhere, route them anywhere.

可插件化的 Webhook 中间件：按 Source 验签/解析为统一 `Event`，用 YAML 路由匹配并渲染为 `Message`，再投递到 Target（如 ntfy）。

## 快速开始

```bash
cp config.example.yaml config.yaml
# 编辑 config.yaml，密钥用环境变量：
#   export VIKUNJA_WEBHOOK_SECRET=...
#   export NTFY_TOKEN=...   # 若使用 token 鉴权

go run ./cmd/hookhub -config config.yaml
```

- Webhook：`POST /hook/{name}`（`name` = `sources[].name`）
- 健康检查：`GET /healthz`、`GET /readyz`
- 调试（只记日志、不转发）：`POST /hook/{name}?debug=1`

## 配置要点

| 字段 | 含义 |
| ------ | ------ |
| `sources[].name` | 实例名 → URL 路径与 `Event.Source` |
| `sources[].driver` | 插件实现（如 `vikunja`） |
| `sources[].vars` | 模板变量（如 `base_url`），渲染时为 `.Vars` |
| `sources[].config` | 传给 driver 的配置 |
| `targets[].name` | 供 `routes[].actions[].target` 引用 |
| `match.type` | 事件种类分段通配（`*` = 恰好一段） |

模板上下文为 `{ Event, Vars }`，例如 `{{ .Event.Data.task.title }}`、`{{ .Vars.base_url }}`。

## 投递语义

当前实现为 **at-most-once**：校验/解析成功后立即 HTTP 200，再异步转发。进程崩溃或发送失败可能丢消息；不做持久化队列。

## 开发

```bash
go test ./...
```

设计文档见 [docs/draft/v0.3.md](docs/draft/v0.3.md)。
