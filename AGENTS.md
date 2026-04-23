# AGENTS.md — Dredge

**`DREDGE.md`** is the PRD (what to build, API catalog, normative auth). Update it in the same change when you alter user-visible behavior, public vs admin routes, major HTTP/WebSocket contract, or operator-facing config semantics.

**Run / wiring:** `cmd/dredge/main.go` · FX `internal/app/fx.go` · module `github.com/rofleksey/dredge` (Go 1.25+).

**Admin gate:** Every OpenAPI operation and `GET /ws` needs an admin JWT except `POST /api/v1/auth/login`, `GET /api/v1/me` (`operationId: me`; Bearer, not admin-gated), `GET /health`. Do not drift that list without updating **OpenAPI**, **ogen admin middleware**, **`LiveWebsocketHandler`**, and **`DREDGE.md`** in one go.

**Never hand-edit:** `internal/http/gen/`, `internal/repository/mocks/`, or anything `make clean-generated` removes. Regen from `api/openapi.yaml`: `make gen` or `make generate` → `internal/http/gen` (`package gen`). After `internal/repository/store.go`: `make mocks`.

**Where things go:** Spec `api/openapi.yaml`. HTTP `internal/http/handler` (ogen interfaces, `gen.*` types); middleware `internal/http/middleware` (import `httpmw`); `internal/http/authctx`. Use cases `internal/usecase/...`: primary type **`Usecase`** (not `Service`); thin methods on `uc.go` / `helpers.go`; non-trivial or security-sensitive → dedicated file + tests. Allowed aggregate basenames there: `uc.go`, `types.go`, `deps.go`, `errors.go`, `const.go`, `helper.go`, `helpers.go`, `engine.go`, `engine_helpers.go` — no `misc.go`. Handlers: `handler_<operationId>.go` per ogen op (+ `_test` when worth it), `server.go` for wiring. Twitch I/O only `internal/service/twitch`. DB `internal/repository/postgres` + `migrations/`. Live `internal/ws`. Observability `internal/observability`.

**Makefile:** `make test` · `make lint` · `make build` · `make run` (`go run ./cmd/dredge`) · `make tidy` · `make clean-generated` · `make infra-up` / `make infra-down` / `make infra-rm` · `make frontend-build` · `make frontend-dev` · `make docker-build`. `make frontend-build` uses Unix `rm`/`cp` (Git Bash, WSL, or Linux/macOS). Example scoped test: `go test ./internal/http/handler -run TestHandler_Login_ok`.

OpenAPI → Go without `make`: `go run github.com/ogen-go/ogen/cmd/ogen@latest --target internal/http/gen --package gen --clean api/openapi.yaml`.

**Config:** `cp config.example.yaml config.yaml` — `config.yaml` is gitignored.

**Deployment:** In production, terminate **HTTPS** at a **reverse proxy** (nginx, Caddy, Traefik, etc.) in front of Dredge; the Go process serves **plain HTTP** only and does **not** embed TLS. PostgreSQL is accessed with a normal DSN; **TLS from the application to Postgres is not in scope**. Configure the proxy for **WebSocket** upgrades to `/ws`. Prometheus (if enabled) listens on `server.metrics_address` separately from `server.address` (`GET /health` is on the main listener). Prefer `observability.log_level` **`info`** or **`warn`** outside local dev.

**Lint:** `.golangci.yml` enables `wsl_v5` (blank lines between statement groups). Non-trivial edits: run `make lint`.

## Frontend development

### Documentation requirements

- Always read the **Frontend project requirements** chapter in `DREDGE.md` before frontend changes.
- After frontend changes that affect UI architecture, design tokens, routes, or coding standards:
  - Update the **Frontend project requirements** chapter in `DREDGE.md`.
  - Add a row to the `DREDGE.md` revision history table (`Version | Date | Notes`).
  - Keep details consistent with other functional and API requirements in `DREDGE.md`.

### Frontend change workflow

1. Review `DREDGE.md` frontend requirements and current route/component conventions.
2. Implement the frontend change.
3. Update `DREDGE.md` frontend requirements and revision history in the same change when applicable.
4. Verify references and conventions remain accurate after refactor/migration.
