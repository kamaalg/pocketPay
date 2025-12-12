# Observability (Logging & Tracing) — Dev QuickStart

This project now includes a lightweight observability integration for HTTP services (auth_service, user_service, payment_service, transaction_service).

What was added
- `observability` Go package with:
  - `Init(serviceName)` — creates a JSON `zap` logger and an OpenTelemetry TracerProvider that sends OTLP/HTTP to the collector (default `otel-collector:4318`).
  - `GinMiddleware(logger)` — attaches a `X-Request-Id` header and a request-scoped zap logger to Gin context.
- Services instrumented: `auth_service`, `user_service`, `payment_service`, `transaction_service` (ledger_service/gRPC intentionally left alone per request).
- Docker Compose: `otel-collector` and `jaeger` services (dev-only) and a sample collector config in `docs/otel-collector-config.yaml`.

How to run locally (simple):

1. Start the infrastructure and collector:

```bash
# from repo root
docker compose up -d postgres otel-collector jaeger
# then start services you want to test
docker compose up -d auth_service user_service payment_service transaction_service nginx
```

2. Make a request to a service (example):

```bash
curl -v "http://localhost:1234/api/v1/healthz"
# or directly to service port if you mapped differently
```

3. Open Jaeger UI at http://localhost:16686 and search for traces by service name (e.g., `auth_service`).

Notes
- For local dev the collector is configured to export traces to Jaeger and to log them.
- For production, configure secure endpoints, proper sampling, and a durable logs backend.
- If you want me to also add correlation to logs (trace_id/span_id fields in every log entry) I can add small helpers and replace some prints with structured logs — but I kept the change minimal to be low-risk.

