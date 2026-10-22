# Changelog

本项目的重要变更记录于此。格式参考 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)，版本号遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

All notable changes to this project are documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.5.0] - 2026-10-22

### 新增

- Target 驱动：Gotify（`POST /message`、`X-Gotify-Key`；priority 0–10；extras 映射 format / click / big_image / intent）

### 变更

- `Target.Send` 移除未使用的 `cfg` 参数，签名改为 `Send(ctx, msg)`

---

### Added

- Target driver: Gotify (`POST /message`, `X-Gotify-Key`; priority 0–10; extras for format / click / big_image / intent)

### Changed

- `Target.Send` drops unused `cfg`; signature is now `Send(ctx, msg)`

## [0.4.1] - 2026-10-22

### 新增

- 可配置日志级别：`server.log_level`（`debug` / `info` / `warn` / `error`）
- 启动时打印具体 HTTP 路由列表（如 `POST /hook/vikunja`）
- HTTP 错误路径补充 Warn 日志，避免 `http.Error` 静默返回

---

### Added

- Configurable log level via `server.log_level` (`debug` / `info` / `warn` / `error`)
- Startup log of concrete HTTP routes (e.g. `POST /hook/vikunja`)
- Warn logs on HTTP error paths so `http.Error` responses are not silent

## [0.4.0] - 2026-10-22

### 新增

- Source 驱动：Ech0（对照 [Webhook 文档](https://ech0.app/docs/guide/webhook)：HMAC 验签、`X-Ech0-Timestamp` 防重放、topic 白名单）

---

### Added

- Source driver: Ech0 (per [Webhook docs](https://ech0.app/docs/guide/webhook): HMAC verify, `X-Ech0-Timestamp` skew check, topic allowlist)

## [0.3.0] - 2026-10-22

### 新增

- Webhook 核心：YAML 驱动路由，`text/template` 以 `{ Event, Vars }` 为上下文渲染
- Source 驱动：Vikunja（HMAC / none 验签，提醒与逾期事件）
- Target 驱动：ntfy（Header 发布，none / basic / token 鉴权，Extra 头白名单）
- HTTP 接口：`POST /hook/{name}`、`GET /healthz`、`GET /readyz`、`?debug=1`
- 配置：Viper + `${ENV}` 插值；`sources[].vars` 作为模板变量（如 `base_url`）
- 投递语义：at-most-once——先 HTTP 200 ACK，再异步处理
- Docker 镜像 `jinvic/hookhub`，含 Dockerfile 与 docker-compose.yml
- CI：推送 `v*` 标签时构建多架构镜像至 Docker Hub，并发布二进制到 GitHub Release

### 变更

- 插件接口以 `Driver()` 替代 `Name()`；路径 `{name}` 为配置实例名
- 配置用 `driver` 字段选择插件（不再使用 `type`）

---

### Added

- Webhook hub core: YAML-driven routes, `text/template` rendering with `{ Event, Vars }` context
- Source driver: Vikunja (HMAC / none verify, reminder and overdue events)
- Target driver: ntfy (Header publish, auth none/basic/token, Extra header whitelist)
- HTTP API: `POST /hook/{name}`, `GET /healthz`, `GET /readyz`, `?debug=1`
- Config via Viper + `${ENV}` expansion; `sources[].vars` for template variables (e.g. `base_url`)
- At-most-once delivery: ACK HTTP 200 then process asynchronously
- Docker image `jinvic/hookhub`, with `Dockerfile` and `docker-compose.yml`
- CI: push `v*` tags builds multi-arch image to Docker Hub and GitHub Release binaries

### Changed

- Plugin interface uses `Driver()` instead of `Name()`; path `{name}` is config instance id
- Config selects plugins with `driver` field (not `type`)

## [0.2.0] - 2026-10-21

### 新增

- 设计草案 v0.2：明确 at-most-once、扁平 `internal/app` 布局，v0.x 不做 CEL

---

### Added

- Design draft v0.2: at-most-once semantics, flat `internal/app` layout, no CEL for v0.x

## [0.1.0] - 2026-10-21

### 新增

- 初始设计草案与项目骨架

---

### Added

- Initial design draft and project scaffold
