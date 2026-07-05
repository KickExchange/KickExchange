# KickExchange

Low-latency exchange platform. Football players are the first asset class; the exchange itself is reusable across asset classes.

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for the full architecture (frozen baseline, changes go through ADRs in `docs/adr/`).

## Repo layout

- `frontend/` — Next.js/React/TypeScript
- `services/` — backend services (GraphQL gateway, Trading Service, matching engine)
- `infra/` — Docker/Docker Compose/CI config
- `docs/` — architecture and design docs
- `shared/` — code shared across services
- `scripts/` — dev/ops scripts

## Quick start

Bring up the backend services:

```
cd infra/docker
docker compose up --build
```

See each service's own README for details (e.g. [services/matching-engine/README.md](services/matching-engine/README.md)).

## Workflow

Monorepo. Feature branch -> PR -> squash merge -> `main`. No direct commits to `main`. One feature = one PR.
