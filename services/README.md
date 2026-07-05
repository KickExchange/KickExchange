# services

- [matching-engine/](matching-engine/README.md) — C++20 matching engine, binary TCP protocol. Order books, matching, trade execution. No SQL, no HTTP, no OAuth.
- Trading Service (Go) — not yet implemented (Sprint 2). Owns auth, order validation, PostgreSQL, talks to matching-engine over the binary protocol.
- GraphQL Gateway — not yet implemented (Sprint 3). Talks to Trading Service, never talks to matching-engine directly.

See [docs/ARCHITECTURE.md](../docs/ARCHITECTURE.md) for the full communication flow.
