# HookHub

[中文](README.md)

Receive webhooks from anywhere, route them anywhere.

A small, plugin-style webhook hub in Go. Configure routes in YAML: a Source verifies and parses into a unified `Event`, templates render a `Message`, then a Target delivers it (e.g. ntfy).

**Current drivers**：

| Role | Driver | Notes |
| ---- | ------ | ----- |
| Source | `vikunja` | HMAC / none; reminder and overdue events |
| Target | `ntfy` | Header publish; none / basic / token |

---

## Quick start

```bash
cp config.example.yaml config.yaml
# Edit config.yaml. Put secrets in the environment:
export VIKUNJA_WEBHOOK_SECRET=...
# export NTFY_TOKEN=...   # if using token auth

go run ./cmd/hookhub -config config.yaml
# hookhub -version
```

| Endpoint | Description |
| -------- | ----------- |
| `POST /hook/{name}` | Webhook ingress; `{name}` = `sources[].name` (instance id, not driver) |
| `GET /healthz` / `GET /readyz` | Liveness / readiness |
| `POST /hook/{name}?debug=1` | Log payload preview only; do not forward |

Default listen address: `server.addr` (example: `:8080`).

---

## Configuration

See [`config.example.yaml`](config.example.yaml). Prefer `${ENV}` placeholders for secrets.

| Field | Meaning |
| ----- | ------- |
| `sources[].name` | Instance id → URL path and `Event.Source` |
| `sources[].driver` | Plugin id (e.g. `vikunja`) |
| `sources[].vars` | Template `.Vars` (e.g. `base_url`) |
| `sources[].config` | Driver-specific settings |
| `targets[].name` | Referenced by `actions[].target` |
| `targets[].driver` | Plugin id (e.g. `ntfy`) |
| `match.source` | Exact instance name |
| `match.type` | Event type glob by `.` segments (`*` = one segment) |

**Template context**: `{ Event, Vars }`

```text
{{ .Event.Subject }}
{{ .Event.Data.task.title }}
{{ .Vars.base_url }}/tasks/{{ .Event.Data.task.id }}
```

A missing template key fails that action only; other actions in a fan-out still run.

---

## Delivery semantics

**At-most-once.** After verify/parse succeeds, HookHub responds `200` immediately, then processes asynchronously. A crash or send failure may drop messages. There is no durable queue and no automatic retry yet.

---

## Docker

Image: [`jinvic/hookhub`](https://hub.docker.com/r/jinvic/hookhub)

```bash
cp config.example.yaml config.yaml
# edit config.yaml and export secrets

docker compose up -d

# or
docker run --rm -p 8080:8080 \
  -v "$PWD/config.yaml:/app/config.yaml:ro" \
  -e VIKUNJA_WEBHOOK_SECRET \
  jinvic/hookhub:latest
```

---

## Releases

Pushing a git tag `v*` (e.g. `v0.3.0`) triggers GitHub Actions to:

1. Build and push multi-arch (`linux/amd64`, `linux/arm64`) images to Docker Hub  
   (requires repo Secrets: `DOCKERHUB_USERNAME`, `DOCKERHUB_TOKEN`)
2. Cross-compile binaries and create a GitHub Release  
   Release notes come from the matching section in [`CHANGELOG.md`](CHANGELOG.md) (e.g. `## [0.3.0]`)

```bash
# 1. Update CHANGELOG.md for this version
# 2. Commit, then:
git tag v0.3.0
git push origin v0.3.0
```

Binaries are named `hookhub_<version>_<os>_<arch>.tar.gz` (Windows: `.zip`).

---

## Development

```bash
go test ./...
go run ./cmd/hookhub -config config.yaml
```

Layout: `cmd/hookhub`, `internal/{app,model,source,target}`.

Design notes: [docs/draft/v0.3.md](docs/draft/v0.3.md) (earlier draft: [v0.2](docs/draft/v0.2.md)).
