# Day 8 - Capstone backend: Claim Tracker Lite

Go + Postgres + Valkey (Redis) + Kafka. Fake data only.

POST /claims saves a claim as PENDING and queues an event. A worker reads the event, decides APPROVED or REJECTED, and the result is read back through GET with a cache in front.

## Status of this code (read first)

| Part | State |
|---|---|
| domain / service / consumer / relay / HTTP handler | Unit tested, `go test ./... -race` passes |
| Postgres / Kafka / Redis adapters, cmd/api, cmd/worker | Compile (`go vet ./...` passes) and run in Docker |
| End to end (compose, Kafka, outbox, DLQ, shutdown) | `scripts\run-all.ps1` ran on a real machine: 25/25 steps PASS, see `notes/run-results.md` |

`go.sum` is included. Run `go mod tidy` again if you change dependencies.

Still not covered: no load test, no multi-instance run (several API relays or several workers), the Docker run used one Kafka broker only, and the Python tests are static checks.

## Run

```
cd day08-backend
scripts\run-all.bat          # builds, starts everything, checks it, writes notes\run-results.md, then tears down
scripts\run-all.ps1 -Keep    # same but leaves the stack running
```

Or by hand (needs `go mod tidy` once to create go.sum):

```
docker run --rm -v "%CD%:/src" -w /src golang:1.25 go mod tidy
docker compose up -d --build
curl -X POST localhost:8080/claims -H "Idempotency-Key: demo-1" -H "Content-Type: application/json" -d "{\"patient_ref\":\"FAKE-AB12\",\"amount\":150000}"
curl localhost:8080/claims/<id>
curl "localhost:8080/claims?status=APPROVED&limit=10"
docker compose down -v
```

Tests without Docker: `go test ./... -race` (Go 1.24+) and `python -m unittest discover -s tests`.

## API

| Call | Result |
|---|---|
| `POST /claims` body `{"patient_ref":"FAKE-AB12","amount":150000}` + header `Idempotency-Key` | 201 created PENDING. Same key + same body: 200 with `Idempotent-Replayed: true` and the same id. Same key + different body: 409. No key or bad body: 400 |
| `GET /claims/{id}` | 200, 404, 400 if id is not a uuid |
| `GET /claims?status=PENDING&limit=20&offset=0` | `{items, limit, offset}`. limit 1..100 (default 20), newest first |
| `GET /healthz`, `GET /readyz` | liveness (no dependencies) / readiness (pings Postgres) |

`amount` is an integer in satang (minor unit), never a float. `patient_ref` must match `FAKE-[A-Z0-9]{3,12}`, enforced in Go and by a CHECK in Postgres, so a real identifier cannot be stored by mistake.

Decision rule (simulated): amount above 5,000,000 satang (50,000 THB) is REJECTED (`over_limit`), otherwise APPROVED. `FAKE-POISON` always fails evaluation, to demo retry and DLQ.

## Layout

```
cmd/api        REST + outbox relay          cmd/worker   Kafka consumer
internal/domain    Claim, status rules, Evaluate (no I/O)
internal/service   Claims (submit/get/list/process), Consumer (retry+DLQ), Relay (outbox)
internal/port      interfaces: ClaimRepository, EventPublisher, Cache
internal/adapter   http (net/http) / postgres (pgx) / kafka (kafka-go) / redis (go-redis)
internal/memory    in-memory fakes of the ports, used by tests
migrations/        golang-migrate files    docs/  sequence diagram
```

service depends only on `port` interfaces, so every use case is tested with fakes. A Python test checks that service and domain never import an adapter.

![sequence](docs/sequence.png)

## Design decisions

- **Outbox instead of dual write.** Saving to Postgres and then calling Kafka can die in the middle. So the API writes the claim and an `outbox` row in one transaction, and a relay (goroutine inside the API) publishes unsent rows. Crash after publish but before marking sent means the event goes out twice, which is fine because the consumer is idempotent.
- **Idempotency key.** Unique index on `idempotency_key` plus a hash of the request. `INSERT ... ON CONFLICT DO NOTHING` makes concurrent duplicates safe (tested with 20 parallel submits in the fake).
- **Consumer is idempotent.** It re-reads the claim, skips anything that is no longer PENDING, and updates with `WHERE status = 'PENDING'`. The event holds only the claim id, never a copy of the data.
- **Retry then DLQ.** Up to 3 attempts with 0.5s / 1s backoff for transient errors. Permanent errors (bad payload, unknown claim) go straight to `claims.submitted.dlq` with reason, original topic, key and value. A message counts as done once it is processed or parked.
- **Offset committed after the work.** If the handler still fails (for example the DLQ publish fails) the adapter retries the same message instead of moving on, so a later commit cannot skip it. On SIGTERM the in-flight message finishes first.
- **Cache is best effort.** Valkey errors are logged and treated as a miss. Claim by id: 60s TTL, deleted when the worker changes the status. Lists: 10s TTL, not invalidated.
- **kafka-go** chosen for a small API. franz-go would be the choice for more control and speed. Kafka runs in KRaft mode (no ZooKeeper), topic auto-create is off and `kafka-init` creates 3 partitions + 1 for the DLQ.
- **pgx directly** (no sqlc) to keep the queries visible. About 6 queries.

## Known limits

- Relay runs inside every API instance. It is safe (SKIP LOCKED) but polling every 500 ms is not efficient. A separate relay process or Postgres LISTEN/NOTIFY, or CDC, would be next.
- Kafka delivery is at-least-once. There is no exactly-once. The idempotent consumer is what makes that acceptable.
- A poison message blocks its partition while it is retried. After the 3 attempts it moves to the DLQ. Nothing reads the DLQ yet, there is no replay tool.
- A poison claim stays PENDING forever. There is no `FAILED` status or manual review flow.
- List cache can be up to 10s stale after a new claim. By-id cache has a tiny race: a read that started before the worker update can write the old value after the worker deleted the key, stale up to 60s.
- Pagination is limit/offset. It gets slow on deep pages and can skip or repeat rows when data changes. Keyset pagination (created_at, id) would fix this.
- Idempotency keys are global, not per client, and never expire.
- No auth, no rate limit, no metrics or tracing (Day 10). Compose uses throw-away DB credentials for local use only.
- Not done from the optional list: gRPC `GetClaim`, testcontainers integration tests. The compose run script covers the integration path instead.

## Tests

| Where | What |
|---|---|
| `internal/domain` | validation table, transitions, decision rule, hash, uuid |
| `internal/service` | idempotency (same / different / concurrent), outbox relay (broker down, ordering), process idempotent on redelivery, retry then success, poison to DLQ, bad payload, permanent error not retried, DLQ failure not committed, shutdown stops retrying, cache hit/invalidate/outage, pagination |
| `internal/adapter/http` | status codes, body size limit, no internal error details leaked, readyz 503, liveness independent of DB |
| `internal/baseline` + `testdata/baseline.json` | snapshot of decisions, transitions, limits, topics, HTTP codes. If it fails on purpose: `go test ./internal/baseline -update` |
| `tests/test_infra.py` | compose, migration vs Go constants, Dockerfile, import rules, commit-after-handler, fake-data-only |
| `scripts/run-all.ps1` | the real thing: compile, up, API behaviour, async approve/reject, cache, DLQ, Kafka down then up, graceful exit |

Mutation check done in the sandbox: breaking the over-limit rule, the idempotency conflict check, the "already processed" skip, the permanent-error rule, or cache invalidation each makes tests fail.
