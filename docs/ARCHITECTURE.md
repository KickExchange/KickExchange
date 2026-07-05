# KickExchange Architecture v1.0

Status: **Frozen baseline.** This document captures the architectural decisions locked at the end of Sprint 0. Any future change to a decision recorded here (e.g. replacing Go, changing GraphQL, redesigning the protocol) must go through an Architecture Decision Record (ADR) in `docs/adr/`, not an ad hoc implementation choice.

## Vision

KickExchange is a low-latency exchange platform. Football players are the first asset class, not the point of the system — the exchange itself is reusable across asset classes.

## Repository Strategy

- Monorepo.
- Workflow: feature branch -> PR -> squash merge -> `main`.
- No direct commits to `main`.
- One feature = one PR.

## Frontend

- Next.js
- React
- TypeScript
- Tailwind CSS

## API Layer — GraphQL Gateway

Responsibilities:

- Authentication
- Query aggregation
- GraphQL subscriptions (later)
- Calls Trading Service
- Calls Market Data Service (later)

The gateway never talks directly to the matching engine.

## Trading Service

Language: **Go**

Responsibilities:

- Authenticate requests
- Validate orders
- Talk to PostgreSQL
- Maintain TCP connection(s) to the matching engine
- Return execution results
- Publish events (later)

This is the "brain" of the backend.

## Matching Engine

Language: **C++20**

Responsibilities:

- Maintain order books
- Match orders
- Execute trades
- Binary TCP server

Explicitly out of scope for the matching engine: SQL, HTTP, OAuth, football-specific logic.

## Communication Flow

```
Browser
   |
HTTPS / GraphQL
   |
GraphQL Gateway
   |
REST (initially)
   |
Trading Service
   |
Custom Binary TCP
   |
Matching Engine
```

## Database

PostgreSQL.

Initial tables: `users`, `assets`, `orders`, `trades`, `portfolio`.

No football statistics, clubs, or matches stored here.

## Authentication

OAuth only (Google, GitHub initially). No password auth.

## Asset Model

Every tradable thing is an `asset`. Football players are one instance:

```
AssetType = FOOTBALL_PLAYER
```

## Football Data Ownership

Football data is owned by external APIs. We only store `external_id`, `symbol`, `display_name`. Everything else is fetched from the external source.

## Infrastructure

Locked:

- Docker
- Docker Compose
- GitHub Actions

Later:

- Redis
- NATS
- Prometheus
- Grafana
- Kubernetes

## Development Philosophy

Design first. Implementation second. Documentation alongside code. Testing with every feature.

## Intentionally Postponed

Not because they're bad — because they don't help execute the first trade:

- Analytics
- Notifications
- Leaderboards
- Redis
- NATS
- Kubernetes
- Observability stack

These return once the end-to-end trading path works.

## Roadmap

```
Sprint 0 (done)
  Architecture, tech stack, repo, workflow

Sprint 1
  Refactor matching engine, design binary protocol, TCP server

Sprint 2
  Trading service, PostgreSQL, OAuth, assets, orders

Sprint 3
  GraphQL gateway, Next.js, place first trade

Sprint 4
  Market data integration, live prices, WebSockets

Sprint 5+
  Redis, NATS, observability, Kubernetes, CI/CD enhancements
```

## Open Design Decisions (pre-Sprint 1)

Three foundational designs remain before implementation starts, each tracked as its own spec:

1. **Binary TCP protocol** — contract between Trading Service and Matching Engine (packet header, message types, serialization, error codes, versioning, heartbeats).
2. **Matching engine internals** — refactor existing LOB codebase (`C:\Users\vedvi\Desktop\LOB`) into `TCP Server -> Packet Decoder -> Matching Engine -> Order Book` layers.
3. **Trading Service API** — REST endpoints, request/response contracts, validation, error handling (consumed internally by the GraphQL gateway).

## Change Process

Any change to a decision in this document requires an ADR in `docs/adr/` describing context, decision, and consequences. This document is updated to reflect the new baseline only after the ADR is accepted.
