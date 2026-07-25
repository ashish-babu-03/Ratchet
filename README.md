# Ratchet

A concurrent, per-client token bucket rate limiter built in Go — created to validate hands-on Go concurrency proficiency (goroutines, channels, context cancellation) alongside a Java/Kotlin backend background.

## What it does

Each client gets an independent token bucket (default: 10 tokens, refilling at 2/sec). Requests are allowed or rejected (`429`) based on token availability. Idle client buckets are automatically evicted via a background goroutine to prevent unbounded memory growth.

## How to run

```bash
go run main.go
```

Server starts on `:8080`. Hit it with:
```bash
curl "http://localhost:8080/request?client=yourname"
```

## Run with Docker

```bash
docker build -t ratchet .
docker run -p 8080:8080 ratchet
```

## Design decisions

- **Per-client isolation via a map** (`map[string]*TokenBucket`), guarded by a `sync.Mutex` — chosen over channels here specifically because protecting shared map access is a textbook mutex use case, not everything in Go needs to be channel-based.
- **Idle eviction via a background goroutine** — a ticker checks every 10s (configurable) and removes buckets idle beyond a threshold, using `context.Context` for clean shutdown signaling.
- **Explicit error handling** throughout — no exceptions, matching idiomatic Go.

## Verified, live

- Correct per-client throttling under sequential load (single client, burst of 15+ requests).
- No crash or data corruption under 75 concurrent requests across 5 simultaneous new clients (10 allowed / 5 rejected per client, consistent across all 5).
- Idle eviction confirmed firing automatically in real runtime output.

## Known limitations

- Single-process, in-memory only — no persistence or multi-instance coordination (unlike Quorum, this isn't distributed).
- Bucket refill rate/capacity are hardcoded, not yet configurable via flags or env vars.

## Why this exists

Built to demonstrate practical Go proficiency — goroutines, channel-free concurrency safety via mutexes, and context-based lifecycle management — as a fast, hands-on stack transfer from a Java/Kotlin backend background.
