# matching-engine

C++20 matching engine. Order books, matching, trade execution over a custom binary TCP protocol. No SQL, no HTTP, no OAuth, no football logic — that's Trading Service's job.

## Run it

```
cd infra/docker
docker compose up --build
```

Listens on TCP port 9000 (mapped to host `9000` by compose; other containers on `exchange-network` reach it at `matching-engine:9000`).

## Build/test natively (Linux)

```
cmake -S . -B build -DCMAKE_BUILD_TYPE=Release
cmake --build build -j
ctest --test-dir build --output-on-failure
```

## How another service calls this

One persistent TCP connection per caller (this engine only accepts one connection at a time — no reconnect-per-order, no HTTP, no gRPC). Protocol v1, little-endian, x86_64 only.

1. **Handshake first.** Send `Hello{min_version, max_version}`, engine replies `HelloAck{accepted_version, ok, engine_version}`. If `ok == 0`, engine closes the connection.
2. **Then send order requests**, one of: `NewLimit`, `NewMarket`, `NewStop`, `NewStopLimit`, `ModifyLimit`, `ModifyStop`, `ModifyStopLimit`, `CancelOrder`. Every message carries a 28-byte header (`version, msg_type, flags, sequence_number, request_id, asset_id, payload_size`) followed by a fixed-size payload for that `msg_type` (see `src/protocol/codec.hpp` `wire_size()`).
3. **Assign your own `order_id`** (caller-assigned, e.g. a DB sequence — the engine doesn't generate ids). Note: narrowed to 32-bit `int` internally, so ids must fit in that range.
4. **Read responses**: `Accepted`/`Rejected` correlate to the request via `request_id`. `Executed`/`OrderDone` correlate via `order_id` instead — a single incoming order can trigger fills against orders from earlier, unrelated requests, so those carry `request_id = 0` (not a direct response to anything just sent).
5. **Send `Heartbeat`** every 5s while idle; engine disconnects after ~15s of silence.

Typical flow for one order that crosses the book:

```
-> NewLimit (request_id=1)
<- Accepted
<- Executed   (this order's fill,   request_id=1)
<- Executed   (resting order's fill, request_id=0)
<- OrderDone  (resting order,        request_id=0)
<- OrderDone  (this order,           request_id=1)
```

See `scripts/smoke_test.py` for a minimal working client (independent of this repo's own codec, drives the raw bytes) — good reference for implementing a client in another language.

## Known limitations

- One connection at a time (no healthcheck port yet — see `docs/ARCHITECTURE.md` follow-ups).
- `order_id`/`shares`/`price` are 64-bit on the wire but narrowed to 32-bit `int` internally.
- googletest is fetched via CMake `FetchContent` at build time (network dependency).
