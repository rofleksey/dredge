# Dredge — Project Requirements Document

| Field | Value |
| --- | --- |
| Product | Dredge |
| Module | `github.com/rofleksey/dredge` |
| API surface | OpenAPI 3.0 (`api/openapi.yaml`), version **0.1.0** |
| Document purpose | Capture **what** the system must do for operators and integrators; implementation detail lives in `AGENTS.md` and the codebase. |
| Last reviewed | 2026-04-21 |

---

## 1. Executive summary

**Dredge** is a self-hosted backend and web UI for **monitoring Twitch channels**, **recording chat and stream context**, **evaluating configurable automation rules**, and **delivering notifications** (for example Telegram or webhooks). It integrates Twitch **Helix**, **GraphQL**, **IRC**, and **OAuth**, persists state in **PostgreSQL**, exposes a **REST JSON API** (generated via **ogen**), and pushes **live updates over WebSockets** (`/ws`).

The primary operator is an **admin** authenticated to the system; almost every HTTP operation and the live socket require that role, with narrow exceptions (login, session bootstrap, liveness).

---

## 2. Goals and success criteria

### 2.1 Business / product goals

- **G1 — Centralized Twitch operations:** One place to configure monitored channels, linked Twitch accounts (main/bot), IRC monitor identity, and safety settings (blacklists, suspicion thresholds).
- **G2 — Observable chat and presence:** Persist and query chat history, stream sessions, chatter presence, and derived activity for moderation and review workflows.
- **G3 — Rule-driven automation:** Express reactions to chat, stream lifecycle, and time-based triggers through a structured rules engine with middleware (filters, regex, cooldowns) and actions (notify, send chat where applicable).
- **G4 — Actionable alerts:** Route events to external notification channels with operator-controlled templates and enable/disable toggles.
- **G5 — Assisted configuration (optional):** AI-assisted flows for settings and rule work where configured, without bypassing auth or admin boundaries.

### 2.2 Success criteria (measurable)

- **SC1:** An admin can add a Twitch channel, see live metadata where Helix allows, and receive IRC-sourced chat in the UI and database for monitored channels.
- **SC2:** Rules can be created, updated, deleted, counted, and validated (including regex testing with documented size limits for safety).
- **SC3:** The HTTP API matches the published OpenAPI contract; breaking changes require version or migration strategy.
- **SC4:** Health checks and observability hooks support deployment (liveness endpoint, optional metrics, structured logs, optional tracing and error reporting).

---

## 3. Stakeholders and users

| Role | Needs |
| --- | --- |
| **System owner / admin** | Secure login, full settings access, Twitch account linking, rule and notification management, operational visibility. |
| **Moderator / analyst (via UI)** | Read/search chat and activity, stream context, suspicion signals; depends on product UX built on the API. |
| **Platform / SRE** | Configurable server addresses, DB connectivity, rate limits, metrics endpoint separation, container-friendly deployment. |

There is **no multi-tenant public SaaS** requirement implied by the codebase: deployment is **single-tenant** with one configured admin identity.

---

## 4. Scope

### 4.1 In scope

- **Authentication and authorization:** Password-based admin login, JWT session, admin gate on API and `/ws` except documented exceptions.
- **Settings:** Twitch users (channels), per-channel IRC/notification flags, channel blacklist, suspicion thresholds, IRC monitor settings (anonymous vs linked OAuth).
- **Rules engine:** Rules CRUD, triggers, template variables, regex test endpoint, execution integrated with Twitch live pipeline and notifications.
- **Twitch integration:** Helix-backed browse and stream APIs, chat send where OAuth allows, chat history and message search APIs, chatter lists and watch hints, IRC monitor status and join history samples, stream-scoped messages/activity/leaderboard APIs, user activity and timelines.
- **Notifications:** Notification provider entries (e.g. Telegram, webhook), lifecycle management.
- **Twitch OAuth:** Start/callback flow for linking Twitch accounts to the app configuration.
- **Web UI:** Static SPA served by the Go process (`internal/webui`), CORS and WebSocket origin tied to configured public base URL.
- **Real-time:** WebSocket hub broadcasting live events to authenticated clients.
- **Persistence:** PostgreSQL schema with migrations applied on startup.
- **Observability:** Zap logging, Prometheus metrics on a dedicated listen address, optional OpenTelemetry trace/log export, optional Sentry.

### 4.2 Out of scope (unless explicitly added later)

- Multi-organization tenancy, per-user RBAC beyond admin, or public signup flows.
- Non-Twitch chat platforms as first-class sources.
- Guaranteed delivery semantics for third-party webhooks (at-least-once / idempotency keys) beyond what the implementation provides today.
- Hosted managed service SLAs (self-hosted only from this document’s perspective).

---

## 5. Functional requirements

Requirements are grouped by capability. **FR IDs** are stable labels for traceability; **priority** is **Must** / **Should** / **Could**.

### 5.1 Authentication and session

| ID | Priority | Requirement |
| --- | --- | --- |
| **FR-AUTH-01** | Must | Provide `POST /auth/login` accepting admin credentials and returning a JWT suitable for `Authorization: Bearer` on subsequent calls. |
| **FR-AUTH-02** | Must | Provide `GET /me` for the authenticated principal (bootstrap / session validation); must **not** require admin beyond authentication. |
| **FR-AUTH-03** | Must | Enforce **admin role** on all other OpenAPI operations and on `/ws`, except `POST /auth/login`, `GET /me`, and `GET /health`. |
| **FR-AUTH-04** | Should | Support configurable **login rate limiting** per client IP per rolling window to reduce brute-force risk. |
| **FR-AUTH-05** | Could | Document WebSocket auth via `Authorization: Bearer` header **or** `?token=` query for browser clients that cannot set headers. |

### 5.2 Health and operations

| ID | Priority | Requirement |
| --- | --- | --- |
| **FR-OPS-01** | Must | Expose `GET /health` returning **200** plain-text liveness without authentication. |
| **FR-OPS-02** | Must | Apply database **migrations on application start** before components assume current schema. |
| **FR-OPS-03** | Should | Expose **Prometheus metrics** on a configurable address distinct from the main HTTP listener. |
| **FR-OPS-04** | Could | Support optional **OTLP log/trace** export and **Sentry** error reporting via configuration. |

### 5.3 Twitch channel configuration

| ID | Priority | Requirement |
| --- | --- | --- |
| **FR-TWCH-01** | Must | List Twitch users (channels) with optional **monitored-only** mode for lightweight lists. |
| **FR-TWCH-02** | Must | Create Twitch users by channel identity, rejecting unknown Twitch channels with a clear client error. |
| **FR-TWCH-03** | Must | Update per-channel settings including **IRC only when live**, **off-stream message notifications**, and **stream start notifications**, with validation when combinations are invalid (e.g. off-stream notifications vs live-only IRC). |
| **FR-TWCH-04** | Should | Reflect **live/offline** and enrichment fields where Helix and workers populate them (exact fields per API schema). |

### 5.4 IRC monitoring and presence

| ID | Priority | Requirement |
| --- | --- | --- |
| **FR-IRC-01** | Must | Allow configuration of IRC monitor identity (**anonymous vs linked OAuth** account). |
| **FR-IRC-02** | Should | Expose **IRC monitor status** and **joined-channel history samples** for troubleshooting. |
| **FR-IRC-03** | Must | Maintain **channel chatter** presence derived from IRC NAMES merge on a configurable interval. |

### 5.5 Chat ingestion, storage, and retrieval

| ID | Priority | Requirement |
| --- | --- | --- |
| **FR-CHAT-01** | Must | Persist chat messages with association to channel user, optional chatter user, optional stream session, timestamps, and metadata (e.g. badges, first-message flags) as defined by schema. |
| **FR-CHAT-02** | Must | Provide **paginated/filtered** chat message listing and **counts** for UI and API consumers. |
| **FR-CHAT-03** | Must | Provide **chat history** endpoints suitable for channel or context drill-down. |
| **FR-CHAT-04** | Should | Support sending chat messages via authenticated Helix path where linked accounts and Twitch policy allow (`/twitch/send`). |

### 5.6 Streams and analytics-style views

| ID | Priority | Requirement |
| --- | --- | --- |
| **FR-STR-01** | Must | Track **stream sessions** per monitored channel (Helix stream id, start/end, title/game snapshots). |
| **FR-STR-02** | Must | List streams and fetch a stream by id with related **messages**, **activity**, and **leaderboard** aggregates as per OpenAPI. |
| **FR-STR-03** | Should | Poll Helix for monitored sessions and metadata on a **configurable interval** to balance freshness and rate limits. |
| **FR-ACT-01** | Should | Record and expose **user activity events** and **timelines** for cross-channel behavior analysis. |

### 5.7 Suspicion and safety

| ID | Priority | Requirement |
| --- | --- | --- |
| **FR-SAFE-01** | Must | Maintain a **channel blacklist** (normalized logins) editable via API. |
| **FR-SAFE-02** | Must | Persist and expose **suspicion settings** (thresholds and related parameters per schema). |
| **FR-SAFE-03** | Should | Compute or flag **suspicious users/channels** consistent with configured thresholds and broadcast notable updates to live clients where implemented. |

### 5.8 Rules engine

| ID | Priority | Requirement |
| --- | --- | --- |
| **FR-RULE-01** | Must | Support rules composed of **event types** including at minimum: chat message, stream start, stream end, interval. |
| **FR-RULE-02** | Must | Support **middleware** concepts including channel filter, user filter, regex match, word contains, and cooldown, as persisted and evaluated by the engine. |
| **FR-RULE-03** | Must | Support **actions** including **notify** and **send chat** with structured `action_settings`. |
| **FR-RULE-04** | Must | Provide CRUD-style HTTP operations for rules (list/create/update/delete), counts, and **rule triggers** listing. |
| **FR-RULE-05** | Should | Expose **template variables** documentation endpoint for operator-authored templates. |
| **FR-RULE-06** | Must | Provide **regex test** endpoint with **bounded input size** to mitigate ReDoS (aligned with engine limits). |
| **FR-RULE-07** | Should | Record **rule trigger events** for auditing or debugging (per migration `0010_rule_trigger_events.sql`). |

### 5.9 Notifications

| ID | Priority | Requirement |
| --- | --- | --- |
| **FR-NOTIF-01** | Must | Store **notification entries** with provider type (e.g. Telegram, webhook), JSON settings, and enabled flag. |
| **FR-NOTIF-02** | Must | Support create/list/update/delete flows via HTTP API. |
| **FR-NOTIF-03** | Should | Allow rules and suspicion flows to target configured notifications with templated content. |

### 5.10 Linked Twitch accounts (OAuth)

| ID | Priority | Requirement |
| --- | --- | --- |
| **FR-OAUTH-01** | Must | Support OAuth **start** URLs and server-side **callback** handling using configured client id/secret and redirect URI. |
| **FR-OAUTH-02** | Must | Persist linked accounts with refresh tokens, support **main** vs **bot** typing, and soft-delete semantics as implemented. |
| **FR-OAUTH-03** | Should | Cache and refresh user OAuth tokens on a configurable TTL for Helix and IRC usage. |

### 5.11 Real-time WebSocket

| ID | Priority | Requirement |
| --- | --- | --- |
| **FR-WS-01** | Must | Provide `/ws` upgrade on the same origin policy as the web app, requiring authenticated admin (per **FR-AUTH-03**). |
| **FR-WS-02** | Should | Push **welcome payloads** and ongoing live events (chat, suspicion, monitor state, etc.) consistent with use case design. |

### 5.12 AI-assisted features

| ID | Priority | Requirement |
| --- | --- | --- |
| **FR-AI-01** | Could | When configured, expose **AI settings** and **conversation** APIs for guided changes (e.g. rule merges) with explicit **confirm** and **stop** operations. |
| **FR-AI-02** | Could | Ensure AI flows respect the same **auth/admin** model as the rest of the API surface. |

### 5.13 Web UI

| ID | Priority | Requirement |
| --- | --- | --- |
| **FR-UI-01** | Must | Serve the compiled SPA from the Go binary for production-style deployment. |
| **FR-UI-02** | Must | Honor **CORS** and **WebSocket Origin** checks against configured **public base URL** (no path or fragment). |

---

## 6. Non-functional requirements

| ID | Category | Requirement |
| --- | --- | --- |
| **NFR-01** | Security | Store secrets (JWT signing key, Twitch client secret, DB password) via configuration, not in source control; use `config.example.yaml` as template only. |
| **NFR-02** | Security | Use constant-time or standard library comparisons where applicable for auth; never log raw tokens or refresh tokens at info level. |
| **NFR-03** | Reliability | Tolerate Twitch/API transient failures with retries or backoff where already implemented; surface operator-visible errors on the API. |
| **NFR-04** | Performance | DB queries for chat and timelines must remain paginated; support pool tuning via config (`max_conns` / `min_conns`). |
| **NFR-05** | Maintainability | OpenAPI is the contract; server types and routing are **generated** (`make gen`); handlers map to `gen.*` responses without forking the contract in ad-hoc handlers. |
| **NFR-06** | Quality | Automated tests (`make test`) and lint (`make lint`) gate changes; new behavior should extend tests following existing patterns. |

---

## 7. Technical constraints and dependencies

- **Language / runtime:** Go **1.25+** (see `go.mod`).
- **Database:** PostgreSQL reachable via **pgx** pool DSN in config.
- **External services:** Twitch developer application (client id/secret, registered redirect URL), optional Telegram/webhook endpoints, optional Sentry and OTLP collectors.
- **Frontend:** Built with project Makefile (`make frontend-build`), output embedded or copied under `internal/webui/static/`.
- **Migrations:** SQL under `internal/repository/postgres/migrations/`, executed during application lifecycle startup.
- **HTTPS and TLS:** The application listens for **HTTP** only. Production assumes a **reverse proxy** terminates **HTTPS** for clients; **TLS is not implemented inside the Go server**. **TLS from the application to PostgreSQL is not required** for the documented deployment model.
- **WebSockets:** The operator’s reverse proxy must support upgrading and proxying **`/ws`** to the backend.

---

## 8. Configuration (summary)

Configuration is file-based (`config.yaml`, gitignored). Notable groups:

- **Server:** listen address, public **base_url** (CORS + WS origin), **metrics** address, login rate limit.
- **Database:** DSN and optional pool limits.
- **JWT:** signing secret and TTL.
- **Admin:** bootstrap email/password.
- **Twitch:** API credentials, OAuth redirect and SPA return URL, polling intervals for viewers, chatters, stream sessions, token cache TTL.
- **Observability:** service name, log level, Sentry DSN, log/trace exporter selection.

See `config.example.yaml` for authoritative keys and defaults.

---

## 9. API catalog (reference)

All paths below are **admin-gated** unless noted. Full request/response schemas live in `api/openapi.yaml`.

| Area | Paths (summary) |
| --- | --- |
| Auth | `POST /auth/login` (public), `GET /me` (auth only) |
| Settings | `/settings/twitch-users`, `…/update`, `…/channel-blacklist`, `…/suspicion-settings`, `…/irc-monitor-settings`, `…/rules*`, `…/rule-triggers`, `…/notifications*`, `…/twitch-accounts*` |
| Twitch data | `/twitch/send`, `…/chat/history`, `…/messages`, `…/users`, `…/channels/live`, `…/channels/chatters`, `…/watch/hints`, `…/irc-monitor/status`, `…/irc-monitor/joined-history`, `…/streams`, `…/streams/{streamId}`, `…/streams/{streamId}/messages|activity|leaderboard`, `…/users/activity`, `…/users/activity/timeline` |
| AI (optional) | `/ai/settings`, `/ai/conversations`, `/ai/conversations/{id}`, `…/messages`, `…/confirm`, `…/stop` |
| Non-OpenAPI | `GET /health` (public), `GET /ws` (admin), `GET/POST` Twitch OAuth callback route (see handler constants) |

---

## 10. Related documents

- **`AGENTS.md`** — Contributor and coding-agent guide: layout, codegen commands, auth exception list (must stay in sync with code), layering conventions, and style gates.
- **`api/openapi.yaml`** — Normative HTTP contract for version 0.1.0.

---

## 11. Revision history

| Version | Date | Notes |
| --- | --- | --- |
| 0.1 | 2026-04-21 | Initial PRD from repository survey (OpenAPI 0.1.0, `fx` wiring, migrations, use case modules). |
