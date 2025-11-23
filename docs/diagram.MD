flowchart LR
  subgraph Client
    Web[Web App / Admin UI]
  end

  subgraph Monitoring
    HM[Health Monitor\n/polls /readyz\n/aggregates /status]
  end

  subgraph Core
    API[API Gateway (Go)\n/healthz /readyz /health/deps]
    PAY[Payments Service (Go)\n/healthz /readyz /health/provider]
    LED[Ledger Service (Go)\n/healthz /readyz /health/backlog]
    NOTIF[Notifications (opt)\n/healthz /readyz]
  end

  subgraph Infra
    PG[(Postgres)]
    BUS[(NATS/Kafka)]
  end

  PSP[[Stripe PSP]]

  %% Client paths
  Web -- HTTPS --> API

  %% Business flows (thin)
  API -- "create payment" --> PAY
  PAY -- "payment.* events" --> BUS
  BUS -- "consume" --> LED
  BUS -- "consume" --> NOTIF
  PSP <-- "webhooks" --> API
  PAY -- "read/write" --> PG
  LED -- "read/write" --> PG

  %% Health polling
  HM -- "GET /readyz" --> API
  HM -- "GET /readyz" --> PAY
  HM -- "GET /readyz" --> LED
  HM -- "GET /readyz" --> NOTIF

  %% Dependencies checked by /readyz
  API -. "DB ping, BUS ping, PSP check (synthetic via PAY)" .- PG
  ```mermaid
  flowchart LR
    %% Client
    Web["Web App / Admin UI"]

    %% Monitoring
    HM[["Health Monitor\n(polls /readyz, aggregates /status)"]]

    %% Core services
    subgraph Core
      API["API Gateway (Go)\n/healthz · /readyz · /health/deps"]
      PAY["Payments Service (Go)\n/healthz · /readyz · /health/provider"]
      LED["Ledger Service (Go)\n/healthz · /readyz · /health/backlog"]
      NOTIF["Notifications (opt)\n/healthz · /readyz"]
    end

    %% Infrastructure
    PG[(Postgres)]
    BUS[(NATS / Kafka)]
    PSP[["Stripe (PSP)"]]

    %% Client -> API
    Web -- "HTTPS" --> API

    %% Business flow
    API -- "create payment" --> PAY
    PAY -- "publish payment.* events" --> BUS
    BUS -- "consume" --> LED
    BUS -- "consume" --> NOTIF
    PSP <-- "webhooks" --> API
    PAY -- "read / write" --> PG
    LED -- "read / write" --> PG

    %% Health polling
    HM -- "GET /readyz" --> API
    HM -- "GET /readyz" --> PAY
    HM -- "GET /readyz" --> LED
    HM -- "GET /readyz" --> NOTIF

    %% Dependencies referenced by readiness checks (dashed)
    API -.-> PG
    API -.-> BUS
    PAY -.-> PG
    PAY -.-> PSP
    LED -.-> PG
    LED -.-> BUS

  ```

  ## Diagram legend

  - Web: frontend or admin UI used by operators or end-users.
  - API: gateway handling public requests, validation, auth and fan-out to services.
  - PAY (Payments): handles payment lifecycle and publishes domain events.
  - LED (Ledger): consumes events and writes double-entry transactions to Postgres.
  - NOTIF: optional notifications worker that sends emails/push/Slack.
  - HM (Health Monitor): polls each service `/readyz`, aggregates results and emits `/status`.
  - PG / BUS / PSP: infrastructure components (Postgres DB, message bus, payment service provider).

  ## How to view

  - GitHub now supports Mermaid diagrams in Markdown — open this file on GitHub to see the rendered diagram.
  - Locally, use the "Markdown Preview Mermaid Support" VS Code extension or visit https://mermaid.live to paste the code block and preview.

  ## Notes

  - Readiness checks should validate critical dependencies with tight timeouts (e.g., DB SELECT 1 with 100–300ms timeout).
  - The Health Monitor should treat PSP pings as circuit-broken and report degraded instead of hard-failing when the circuit is open.
