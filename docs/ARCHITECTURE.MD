## System Purpose

The system is a personal payment management platform. It lets a signed-in user:

- Create and track payments
- Manage subscriptions and invoices
- View wallet / ledger balances
- Receive real-time status updates (SSE / WebSocket)

Non-functional must-haves:

- Observability & health: reliable health endpoints, dependency checks, and alerting
- Correctness: idempotent payment operations and a double-entry ledger
- Resilience: retries, circuit breakers, and backpressure
- Security: least-privilege service credentials, TLS, and JWT for public endpoints

## Service Boundaries

### API Gateway (Go / Gin)

- Public HTTP surface: authentication, request validation, idempotency keys
- Fan-out to internal services; streams server-sent events (SSE) or WebSocket to the UI
- Health endpoints:
	- `GET /healthz` — liveness (process up)
	- `GET /readyz` — readiness (DB, message bus, etc.)
	- `GET /health/deps` — extended diagnostics (PSP ping, cache, object store)

### Payments Service (Go)

- Payment state machine: `CREATED → AUTHORIZED → CAPTURED → SETTLED → REFUNDED|FAILED`
- Integrates with provider adapters (Stripe now; PayPal/Braintree later)
- Emits domain events to the bus (`payment.*`)

Health endpoints:

- `GET /healthz` — process liveness
- `GET /readyz` — DB pool, bus connection
- `GET /health/provider` — PSP sandbox ping (with timeout & circuit breaker)

### Ledger Service (Go)

- Consumes payment events and posts double-entry transactions to Postgres
- Exposes querying endpoints such as:
	- `GET /accounts/:id/balance`
	- `GET /ledger/tx/:id`

Health endpoints:

- `GET /healthz`, `GET /readyz` (DB + bus)
- `GET /health/backlog` — event lag metric ("oldest unprocessed offset age")

### Health Monitor / Aggregator (Go)

- Polls each service's readiness on a short interval (e.g., every 15s)
- Consolidates status into `/status` and pushes alerts to Slack / Email
- Provides a JSON UI for admin/front-end dashboards
- Health endpoints for the monitor itself: `GET /healthz`, `GET /readyz`

### Notifications Service (optional)

- Subscribes to domain events and sends emails, push notifications, or Slack messages
- Health: `GET /healthz`, `GET /readyz` (checks SMTP provider or external API)

### Postgres

- Primary data store for payments, ledger, idempotency keys, and audit logs
- Health check: TCP + `SELECT 1` probe

### Message Bus (NATS or Kafka)

- Reliable async event delivery for domain events (`payment.created`, `payment.captured`, ...)
- Health check: broker connectivity and consumer lag

### External PSP (Stripe first)

- Drives real payments and webhooks which transition payment state
- Health: synthetic, short test call with strict timeout and circuit breaker

## Data Flow (Health-centric)

- The Health Monitor polls `GET /readyz` for API, Payments, Ledger and Notifications every 15s
- Each service's `readyz` performs fast checks on critical dependencies:
	- DB: `SELECT 1` with a 100–300ms timeout
	- Bus: ping / metadata fetch
	- PSP: short capabilities call (circuit-broken; run less frequently)
- The Health Monitor aggregates results and exposes `GET /status`. It emits alerts when:
	- a service reports not ready
	- consumer lag (Ledger) exceeds a threshold (e.g., >30s)
	- PSP failures exceed an error rate threshold in the last 5m
- Admin UI or `/admin/health` can fetch the Health Monitor JSON to render a traffic-light dashboard

(Business flow in brief: `API → Payments → events → Ledger`. PSP webhooks follow: `PSP → API → Payments → events → Ledger`. Events are also observed by Notifications.)

## Communication Patterns (Health Monitoring)

- Synchronous polling: Health Monitor calls `GET /readyz` on each service
- Push alerts via email or Slack webhooks when thresholds breach
- Consumer lag metrics: Ledger emits Prometheus metrics; Health Monitor either scrapes metrics or calls a JSON metrics endpoint
- Circuit breaker around PSP pings to avoid cascading failures; when the breaker is open, `/readyz` should report degraded rather than down

## Health endpoint semantics

- `/healthz` — liveness: the process is running (no dependency checks)
- `/readyz` — readiness: critical dependencies are reachable and within SLO
- `/health/deps` — extended diagnostics for manual triage (not intended for probe checks)

## Technology stack (and why)

- **Go (Gin / chi)** — small runtime, low latency I/O, great for lightweight services and health probes
- **Postgres** — strong transactional guarantees required for ledger correctness (use serializable where needed)
- **NATS or Kafka** — simple, fast eventing; NATS is easier to operate for an MVP
- **Stripe SDK** — good developer ergonomics; webhooks drive state transitions safely
- **Prometheus + Grafana** — metrics, alerting, and dashboards (ready/healthy gauges, consumer lag, webhook success rate)
- **OpenTelemetry (Go SDK)** — distributed traces across API → Payments → Ledger
- **Docker / Docker Compose** — reproducible local dev environment; Compose health dependencies in CI/dev
- **NGINX (optional)** — TLS termination, rate limiting, and static admin pages