# PocketPay

PocketPay is a small payment management platform (MVP) that lets users create and track payments, manage subscriptions and invoices, view wallet/ledger balances, and receive real-time updates. The codebase is organized as a set of small Go services (API Gateway, Payments, Ledger, optional Notification worker) and supporting infrastructure (Postgres, message bus).

## Architecture overview

High level:

- API Gateway (Go / Gin) — public HTTP surface, authentication, validation and fan-out to internal services.
- Payments Service (Go) — payment state machine and provider integration (Stripe adapter).
- Ledger Service (Go) — consumes events and posts double-entry transactions to Postgres.
- Health Monitor — polls services' readiness endpoints and exposes aggregated status.
- Postgres — primary data store for payments, ledger and idempotency.
- Message Bus (NATS or Kafka) — event delivery for domain events (payment.*).

See `docs/ARCHITECTURE.MD` and `docs/diagram.MD` for a full architecture description and diagrams.

## Prerequisites

- Docker (20.10+)
- Docker Compose (v2+)
- Go (1.20+ or matching the module go version) — only required if you want to build/run services locally without Docker
- (Optional) `swag` CLI if you want to generate Swagger docs locally: `go install github.com/swaggo/swag/cmd/swag@latest`

## Installation & setup (local development)

- Clone the repository:

```bash
git clone <your-repo-url> pocketpay
cd pocketpay
```

- Build and run with Docker Compose (recommended):

```bash
# build images and start services
docker compose up --build

# or run in the background
docker compose up --build -d

# check running containers
docker compose ps

# view logs (follow):
docker compose logs -f
```

Notes:
- Compose creates services defined in `docker-compose.yml`. By default the API listens on container port 8000 which is mapped to host port 8011 in the provided compose file.
- If you change DB init scripts or want the initialization SQL to run again, remove the Postgres volume before `up` so the container re-initializes:

```bash
docker compose down -v
docker compose up --build
```

## Usage / Health checks

You can check service health and readiness with the built-in health endpoints. Example host mapping (default in this repo): `http://localhost:8011` → API Gateway.

Basic endpoints (examples):

- Liveness: `GET /healthz`
- Readiness: `GET /readyz`
- Extended diagnostics: `GET /health/deps`

Example curl commands:

```bash
# Liveness
curl -sS http://localhost:8011/healthz | jq .

# Readiness
curl -sS http://localhost:8011/readyz | jq .

# Health Monitor aggregated status (if configured)
curl -sS http://localhost:8011/status | jq .
```

If you enabled Swagger UI for the API, open:

```
http://localhost:8011/swagger/index.html
```

## API documentation (health endpoints)

This README documents the health APIs used by the services. Each service exposes the following endpoints (replace host/port per service if running independently):

1. GET /healthz

- Purpose: liveness probe — indicates the process is running.
- Response example:

```json
{
  "status": "ok",
  "uptime": "1m23s"
}
```

2. GET /readyz

- Purpose: readiness probe — validates critical dependencies (DB, bus, provider) are reachable.
- Typical response (success):

```json
{
  "status": "ready",
  "checks": {
    "db": "ok",
    "bus": "ok",
    "provider": "ok"
  }
}
```

- Typical response (degraded / not ready):

```json
{
  "status": "not_ready",
  "checks": {
    "db": "ok",
    "bus": "unreachable",
    "provider": "degraded"
  }
}
```

3. GET /health/deps

- Purpose: extended diagnostics for manual triage — includes additional details such as DB latency, last consumer offset, PSP ping results.

Note: the exact response shape differs per service. Check the API's Swagger or the service code for precise fields.

## Testing

Manual testing

- Start the system with Docker Compose as described above.
- Use `curl` (or Postman) to exercise endpoints. Example flow:

```bash
# create a payment (API path depends on the service routing)
curl -X POST http://localhost:8011/api/v1/createPayment -H 'Content-Type: application/json' \
  -d '{"amount":1000, "currency":"USD", "customer_id":"abc123"}'

# check payments service readiness
curl -sS http://localhost:8011/readyz | jq .
```

Automated tests

- If services include Go unit tests you can run them locally within each service folder:

```bash
cd user_service
go test ./... 

cd payment_service
go test ./...
```

Integration testing

- For integration tests that depend on Postgres and messaging, run `docker compose up --build` and then execute your test suite which targets service endpoints.

## Project structure

Top-level layout (important folders/files):

```
/
├─ docs/                  # Architecture docs, mermaid diagrams
├─ dev/                   # Primary API/service used for quick local dev (contains cmd/api)
├─ user_service/          # User service code and Dockerfile
├─ payment_service/       # Payment service code and Dockerfile
├─ docker-compose.yml     # Compose file for local development
├─ dev/Dockerfile         # Dockerfile used by the dev service
└─ README.md              # <-- this file
```

Adjust paths if you split services into separate repos; this mono-repo keeps everything together for local development convenience.

## Migrations and schema

- Use a migration tool (e.g., `golang-migrate/migrate`) to run DB schema migrations. Each service should own and run migrations for its schema in CI or at startup (recommended to run them as a separate job).
- Example (golang-migrate):

```bash
migrate -path ./migrations -database "${DB_URL}" up
```

## Further notes

- Secrets: for development we use environment variables in `docker-compose.yml`. For production, use Docker secrets, Vault or other secret backends.
- Observability: the repo includes hooks for Prometheus metrics and OpenTelemetry tracing — enable and configure exporters in production.
- Swagger: to populate swagger UI with API docs run `swag init -g ./cmd/api/main.go` in the appropriate service folder and rebuild the binary so the generated `docs` package is embedded in the binary.

---

If you want, I can also:

- Add a short quickstart script (`scripts/quickstart.sh`) to automate the common commands above
- Generate a snapshot SVG of the architecture diagram and commit it to `docs/`
- Add a small health-check script that polls all services and exits non-zero when any are unhealthy

Tell me which follow-up you want and I'll implement it.