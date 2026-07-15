# trading-service

Go service holding the persistent TCP connection to the Matching Engine. No database, no HTTP, no OAuth, no GraphQL yet — this PR only proves the engine connection stays up.

## Run it

```
cd infra/docker
docker compose up --build
```

Expected logs:

```
connecting addr=matching-engine:9000
connected addr=matching-engine:9000
HELLO OK accepted_version=1 engine_version=...
trading service ready
```

No further reconnect/disconnect lines should appear while both containers keep running.

## Manual acceptance check (30-minute soak)

`docker compose up --build`, leave it running, come back in 30 minutes. If no reconnect/disconnect log lines appeared in between, the connection held. This isn't run in CI (see `tests/`) — it's a manual sanity check before merging behavior changes to `engineclient`.

## Configuration (env vars)

| Var | Default |
|---|---|
| `ENGINE_HOST` | `matching-engine` |
| `ENGINE_PORT` | `9000` |
| `HEARTBEAT_INTERVAL` | `5s` |
| `PROTOCOL_VERSION` | `1` |

## Architecture

Everything protocol-related lives in `internal/engineclient` — nothing else in this service knows TCP exists. See `internal/engineclient/client.go` for the `EngineClient` interface.

- One writer goroutine, one reader goroutine per connection — never more than one goroutine touches the socket for writes.
- Requests (`Submit*`/`Modify*`/`Cancel`) are correlated to responses via `request_id`, tracked in a pending-request table with a 5s timeout.
- `Submit*`/`Modify*` complete on `Accepted`; `Cancel` completes on `OrderDone` (its only response type). Fills and later terminal events arrive afterwards via `OnAsyncEvent` — not called by anything this PR.
- On disconnect: in-flight requests are failed immediately, then reconnect retries with exponential backoff (1s→2s→4s...capped at 30s). A protocol version mismatch on the *first* connect is fatal (exits non-zero) — retrying can't fix a version the engine will never accept.

`Submit*`/`ModifyLimit`/`Cancel` are implemented and wired end-to-end but unused by `main.go` — Sprint 2's goal is proving the connection, not order flow. `internal/service` picks that up next.

## Known follow-ups

- Protocol codec is hand-ported from the engine's C++ source (`services/matching-engine/src/protocol/{message_types.hpp,codec.cpp}`) — no shared schema/codegen. Candidate for an ADR if manual sync becomes painful.
- Matching engine has no dedicated health port yet, so neither service has a Docker healthcheck.
- `order_id` must fit the engine's 32-bit internal representation even though the wire format is 64-bit.
