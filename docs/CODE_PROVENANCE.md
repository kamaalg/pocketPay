# Code Provenance

This document records external assistance and third-party code used while developing PocketPay. It lists AI tools, libraries and frameworks, and other sources consulted. For each entry I include a brief description, the URL (where applicable), and the access date.

---

## AI Tools Used

- ChatGPT (OpenAI)
  - URL: https://chat.openai.com/
  - Dates used: 2025-11-10 → 2025-11-11
  - What it helped with: interactive code edits and reviews, explanations (e.g., why to pass `&in` to `ShouldBindJSON`), implementing graceful HTTP server startup and shutdown, integrating gin-swagger, adding pgx/pgxpool DB connection patterns, composing docker-compose changes, generating README and docs, formatting architecture and diagrams, and general pair-programming guidance.
  - Notes: code snippets and edits produced by the assistant were reviewed and adapted for this repository.

No other AI assistants (Copilot, Bard, etc.) were intentionally used to generate code in this repository unless explicitly noted above.

---

## Code Sources / External References

The development primarily used official project documentation and package homepages for APIs and examples. No single Stack Overflow thread or tutorial was copied verbatim into the repository; examples are adaptations of documented usage patterns.

- Gin (web framework)
  - URL: https://github.com/gin-gonic/gin
  - Accessed: 2025-11-10
  - Usage: HTTP routing, middleware, binding helpers (ShouldBindJSON) and logging.

- swaggo / gin-swagger (Swagger UI middleware)
  - URL: https://github.com/swaggo/gin-swagger
  - Accessed: 2025-11-11
  - Usage: embedding Swagger UI and wiring the swagger docs route.

- swag (swag CLI for OpenAPI annotations)
  - URL: https://github.com/swaggo/swag
  - Accessed: 2025-11-11
  - Usage: recommended tool to generate the `docs` package for swagger (developer instructions provided in README).

- pgx / pgxpool (Postgres driver and connection pool)
  - URL: https://github.com/jackc/pgx
  - Accessed: 2025-11-11
  - Usage: recommended Postgres connection pool usage example added to `user_service` main.

- Postgres (Docker image)
  - URL: https://hub.docker.com/_/postgres
  - Accessed: 2025-11-10
  - Usage: base database service in `docker-compose.yml` and initialization scripts guidelines.

- swaggo/files (swagger UI files)
  - URL: https://github.com/swaggo/files
  - Accessed: 2025-11-11
  - Usage: static assets for the swagger UI.

- NATS / Kafka (message bus choices)
  - URLs: https://nats.io/ and https://kafka.apache.org/
  - Accessed: 2025-11-10
  - Usage: architectural choice for message bus (no vendor-specific code added by the assistant; referenced in architecture docs).

If/when specific code examples or snippets were directly adapted from a tutorial or Stack Overflow post, they would be recorded here with URL and date. At the time of writing, no such external snippets were inserted verbatim — all code was composed or adapted from official package usage patterns.

---

## Libraries & Frameworks included in the repository

Below are the major runtime dependencies used in the project (pulled via `go.mod` or referenced in Dockerfiles). Versions recorded where available in the repository at time of writing.

- github.com/gin-gonic/gin — HTTP framework for Go (v1.10.0 in `dev/go.mod`)
- github.com/swaggo/gin-swagger — Swagger UI wrapper for Gin (v1.6.1)
- github.com/swaggo/files — Swagger UI files (v1.0.1)
- github.com/swaggo/swag — CLI for generating swagger docs (v1.8.12)
- github.com/jackc/pgx/v5/pgxpool — Postgres driver & pool (used in examples)
- postgres Docker image (official) — used for the database service in Docker Compose

Accessed: 2025-11-10 → 2025-11-11 (versions reflect module files present in the repository at that time).

---

## Collaboration

- No classmates or external collaborators contributed code to this repository according to project logs. If collaborators contributed, list their names, affiliation, and what they contributed here.

---

## Verification and Attribution Notes

- All code generated or suggested by AI tools was reviewed and executed (where applicable) locally by the developer. The developer is responsible for verifying correctness, security, and license compatibility before deploying or publishing.
- Third-party libraries are used under their respective licenses — please review each dependency's license if you plan to redistribute or publish the project.

---

If you want, I can extend this file to include commit-level provenance (which commits were produced with AI assistance) or attach the exact assistant transcripts (redacted for secrets) for auditing. Tell me which level of provenance detail you need.
