# Dredge

Self-hosted Twitch monitoring backend and web UI (Go + PostgreSQL + Vue). See **[DREDGE.md](DREDGE.md)** for product requirements and **[AGENTS.md](AGENTS.md)** for layout, codegen, and contribution conventions.

## Quick start (development)

1. `cp config.example.yaml config.yaml` and edit secrets and Twitch settings.
2. Start Postgres (e.g. `make infra-up` when Docker is available).
3. `make run` — API and embedded UI on `server.address` (default `:8080`).

## Production deployment

- **HTTPS:** Terminate TLS at a **reverse proxy** in front of Dredge. The Go binary serves **plain HTTP** only; it does not load TLS certificates.
- **Postgres:** Use a normal connection string from the app to Postgres; **in-app TLS to the database is not part of the documented model** (private network or proxy-side encryption is an operator choice).
- **WebSockets:** Configure the proxy to pass through **`/ws`** with WebSocket upgrade headers.
- **Ports:** Main HTTP/API/UI on `server.address`; optional Prometheus on `server.metrics_address` (no `/metrics` on the main listener).
- **OAuth:** Register Twitch redirect and SPA return URLs for your public hostnames (see `config.example.yaml` and Twitch developer console).
- **Container:** `make docker-build` or use the published image workflow; health check hits `GET /health` on the main port.

## CI

Pull requests and pushes run **Go tests**, **golangci-lint**, and **govulncheck** (`.github/workflows/ci.yml`). Docker image build/push remains in `.github/workflows/build-push.yml`.
