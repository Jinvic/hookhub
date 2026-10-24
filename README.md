# HookHub

[English](README.en.md)

从任意服务接收 Webhook，按规则转发到任意目标。

用 Go 写的小型插件式 Webhook 中枢：YAML 配置路由，Source 验签并解析为统一 `Event`，模板渲染为 `Message`，再投递到 Target（如 ntfy）。

**已实现驱动**：

| 角色 | Driver | 说明 |
| ---- | ------ | ---- |
| Source | `vikunja` | HMAC / none；提醒与逾期等事件 |
| Source | `ech0` | HMAC / none；`X-Ech0-*` 验签与时间窗；动态 / 评论 / 用户等 topic |
| Target | `ntfy` | Header 发布；none / basic / token |
| Target | `gotify` | `POST /message` + `X-Gotify-Key`；JSON extras（markdown / click 等） |
| Target | `http` | 通用出站：配置固定 URL；`plain` / `json`；none / bearer / basic / header |
| Target | `astrbot` | `POST /api/v1/im/messages`；Bearer API Key；纯文本 `message` + 配置 `umo` |

---

## 快速开始

```bash
cp config.example.yaml config.yaml
# 编辑配置，密钥放环境变量：
export VIKUNJA_WEBHOOK_SECRET=...
# export NTFY_TOKEN=...   # 若使用 token 鉴权

go run ./cmd/hookhub -config config.yaml
# hookhub -version
```

| 端点 | 说明 |
| ---- | ---- |
| `POST /hook/{name}` | Webhook 入口；`{name}` = `sources[].name`（实例名，不是 driver） |
| `GET /healthz` / `GET /readyz` | 存活 / 就绪探针 |
| `POST /hook/{name}?debug=1` | 仅记录 payload 预览，不转发 |

默认监听地址：`server.addr`（示例为 `:8080`）。日志级别：`server.log_level`（`debug` / `info` / `warn` / `error`，默认 `info`）。启动时会打印具体路由列表（如 `POST /hook/vikunja`）。

---

## 配置

见 [`config.example.yaml`](config.example.yaml)。密钥请用 `${ENV}` 占位。

| 字段 | 含义 |
| ---- | ---- |
| `sources[].name` | 实例 ID → URL 路径与 `Event.Source` |
| `sources[].driver` | 插件 ID（如 `vikunja`） |
| `sources[].vars` | 模板 `.Vars`（如 `base_url`） |
| `sources[].config` | 驱动专属配置 |
| `targets[].name` | 供 `actions[].target` 引用 |
| `targets[].driver` | 插件 ID（如 `ntfy`） |
| `match.source` | 精确匹配实例名 |
| `match.type` | 按 `.` 分段的事件类型通配（`*` = 一段） |

**模板上下文**：`{ Event, Vars }`

```text
{{ .Event.Subject }}
{{ .Event.Data.task.title }}
{{ .Vars.base_url }}/tasks/{{ .Event.Data.task.id }}
{{ formatTime .Event.Data.reminder.reminder .Vars.timezone }}
{{ formatTime .Event.Data.reminder.reminder "Asia/Shanghai" "2006-01-02 15:04:05" }}
```

`formatTime` 将 RFC3339 / `time.Time` 转到指定 IANA 时区；默认布局 `2006-01-02 15:04`。时区建议放在 `sources[].vars.timezone`。

模板缺键只导致该 Action 失败；扇出时其它 Action 继续执行。

---

## 投递语义

**At-most-once。** 校验 / 解析成功后立即返回 HTTP `200`，再异步处理。进程崩溃或发送失败可能丢消息；当前无持久化队列、无自动重试。

---

## Docker

镜像：[`jinvic/hookhub`](https://hub.docker.com/r/jinvic/hookhub)

```bash
cp config.example.yaml config.yaml
# 编辑 config.yaml，并导出密钥

docker compose up -d

# 或
docker run --rm -p 8080:8080 \
  -v "$PWD/config.yaml:/app/config.yaml:ro" \
  -e VIKUNJA_WEBHOOK_SECRET \
  jinvic/hookhub:latest
```

---

## 发布

推送 git 标签 `v*`（如 `v0.3.0`）会触发 GitHub Actions：

1. 构建并推送多架构镜像（`linux/amd64`、`linux/arm64`）到 Docker Hub  
   （需仓库 Secrets：`DOCKERHUB_USERNAME`、`DOCKERHUB_TOKEN`）
2. 交叉编译二进制并创建 GitHub Release  
   Release 说明取自 [`CHANGELOG.md`](CHANGELOG.md) 中对应版本小节（如 `## [0.3.0]`）

```bash
# 1. 更新本版本的 CHANGELOG.md
# 2. 提交后：
git tag v0.3.0
git push origin v0.3.0
```

二进制命名形如 `hookhub_<version>_<os>_<arch>.tar.gz`（Windows 为 `.zip`）。

---

## 开发

```bash
go test ./...
go run ./cmd/hookhub -config config.yaml
```

目录结构：`cmd/hookhub`、`internal/{app,model,source,target}`。

设计说明：[docs/draft/v0.4.md](docs/draft/v0.4.md)（历史草案：[v0.3](docs/draft/v0.3.md)、[v0.2](docs/draft/v0.2.md)）。
