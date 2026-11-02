# Day 8 run results

Run at: 2026-11-02 16:43:18  Machine: LAPTOP-MICOTO

## Summary

| Step | Result |
|---|---|
| Docker running | PASS |
| go mod tidy (creates go.sum) | PASS |
| gofmt clean (no files listed) | PASS |
| go vet ./... (compiles every package incl. adapters) | PASS |
| go test ./... -race | PASS |
| docker compose up --build | PASS |
| api ready (/readyz 200) | PASS |
| migration created claims + outbox | PASS |
| topics exist | PASS |
| POST /claims -> 201 PENDING | PASS |
| same Idempotency-Key + same body -> 200, same id | PASS |
| same key + different body -> 409 | PASS |
| no Idempotency-Key -> 400 | PASS |
| real-looking patient_ref refused -> 400 | PASS |
| worker approves claim (async) | PASS |
| worker rejects over-limit claim | PASS |
| GET /claims?status=APPROVED lists it | PASS |
| GET by id is cached in Valkey | PASS |
| outbox drained (0 unsent) | PASS |
| poison claim: retried then parked in DLQ, stays PENDING | PASS |
| Kafka down: API still accepts (201), event waits in outbox | PASS |
| Kafka back: relay sends, claim gets APPROVED, nothing lost | PASS |
| worker and api exit 0 on SIGTERM | PASS |
| worker log sample (info only, for the record) | PASS |
| docker compose down -v | PASS |

## Docker running  [PASS]

```
client 29.7.2 / server 29.7.2
```

## go mod tidy (creates go.sum)  [PASS]

```
... (8 earlier lines omitted)
go: downloading github.com/jackc/puddle/v2 v2.2.2
go: downloading github.com/jackc/pgpassfile v1.0.0
go: downloading golang.org/x/text v0.21.0
go: downloading github.com/bsm/ginkgo/v2 v2.12.0
go: downloading github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761
go: downloading github.com/bsm/gomega v1.27.10
go: downloading golang.org/x/crypto v0.31.0
go: downloading github.com/cespare/xxhash/v2 v2.2.0
go: downloading github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f
go: downloading golang.org/x/net v0.17.0
go: downloading github.com/klauspost/compress v1.15.9
go: downloading github.com/pierrec/lz4/v4 v4.1.15
go: downloading github.com/xdg-go/scram v1.1.2
go: downloading golang.org/x/sync v0.10.0
go: downloading github.com/xdg-go/stringprep v1.0.4
go: downloading github.com/xdg-go/pbkdf2 v1.0.0
go: downloading github.com/davecgh/go-spew v1.1.1
go: downloading github.com/pmezard/go-difflib v1.0.0
go: downloading gopkg.in/yaml.v3 v3.0.1
go: downloading golang.org/x/net v0.21.0
```

## gofmt clean (no files listed)  [PASS]

```
listed: 
```

## go vet ./... (compiles every package incl. adapters)  [PASS]

```

```

## go test ./... -race  [PASS]

```
?   	github.com/sitthichok-J/self-study-2026/day08-backend/cmd/api	[no test files]
?   	github.com/sitthichok-J/self-study-2026/day08-backend/cmd/worker	[no test files]
ok  	github.com/sitthichok-J/self-study-2026/day08-backend/internal/adapter/http	1.138s
?   	github.com/sitthichok-J/self-study-2026/day08-backend/internal/adapter/kafka	[no test files]
?   	github.com/sitthichok-J/self-study-2026/day08-backend/internal/adapter/postgres	[no test files]
?   	github.com/sitthichok-J/self-study-2026/day08-backend/internal/adapter/redis	[no test files]
ok  	github.com/sitthichok-J/self-study-2026/day08-backend/internal/baseline	1.029s
ok  	github.com/sitthichok-J/self-study-2026/day08-backend/internal/config	1.021s
ok  	github.com/sitthichok-J/self-study-2026/day08-backend/internal/domain	1.026s
?   	github.com/sitthichok-J/self-study-2026/day08-backend/internal/memory	[no test files]
?   	github.com/sitthichok-J/self-study-2026/day08-backend/internal/port	[no test files]
ok  	github.com/sitthichok-J/self-study-2026/day08-backend/internal/service	1.032s
```

## docker compose up --build  [PASS]

```
... (801 earlier lines omitted)
 Container day08-valkey-1 Started 
 Container day08-postgres-1 Started 
 Container day08-postgres-1 Waiting 
 Container day08-postgres-1 Healthy 
 Container day08-migrate-1 Starting 
 Container day08-migrate-1 Started 
 Container day08-kafka-1 Healthy 
 Container day08-kafka-init-1 Starting 
 Container day08-kafka-init-1 Started 
 Container day08-valkey-1 Waiting 
 Container day08-kafka-init-1 Waiting 
 Container day08-migrate-1 Waiting 
 Container day08-valkey-1 Waiting 
 Container day08-kafka-init-1 Waiting 
 Container day08-migrate-1 Waiting 
 Container day08-valkey-1 Healthy 
 Container day08-valkey-1 Healthy 
 Container day08-migrate-1 Exited 
 Container day08-migrate-1 Exited 
 Container day08-kafka-init-1 Exited 
 Container day08-kafka-init-1 Exited 
 Container day08-api-1 Starting 
 Container day08-worker-1 Starting 
 Container day08-worker-1 Started 
 Container day08-api-1 Started
```

## api ready (/readyz 200)  [PASS]

```
NAME                 IMAGE                     COMMAND                  SERVICE      CREATED          STATUS                     PORTS
day08-api-1          day08-api                 "/app"                   api          13 seconds ago   Up Less than a second      0.0.0.0:8080->8080/tcp, [::]:8080->8080/tcp
day08-kafka-1        apache/kafka:3.8.0        "/__cacert_entrypoin…"   kafka        13 seconds ago   Up 12 seconds (healthy)    0.0.0.0:9094->9094/tcp, [::]:9094->9094/tcp
day08-kafka-init-1   apache/kafka:3.8.0        "/bin/sh -c '/opt/ka…"   kafka-init   13 seconds ago   Exited (0) 1 second ago    
day08-migrate-1      migrate/migrate:v4.18.1   "migrate -path=/migr…"   migrate      13 seconds ago   Exited (0) 7 seconds ago   
day08-postgres-1     postgres:16-alpine        "docker-entrypoint.s…"   postgres     13 seconds ago   Up 12 seconds (healthy)    0.0.0.0:5432->5432/tcp, [::]:5432->5432/tcp
day08-valkey-1       valkey/valkey:8-alpine    "tini -- docker-entr…"   valkey       13 seconds ago   Up 12 seconds (healthy)    0.0.0.0:6379->6379/tcp, [::]:6379->6379/tcp
day08-worker-1       day08-worker              "/app"                   worker       13 seconds ago   Up Less than a second
```

## migration created claims + outbox  [PASS]

```
2
```

## topics exist  [PASS]

```
__consumer_offsets
claims.submitted
claims.submitted.dlq
```

## POST /claims -> 201 PENDING  [PASS]

```
{"id":"b08d8a7d-4185-442a-b75b-b35f8f707da7","patient_ref":"FAKE-AB12","amount":150000,"status":"PENDING","created_at":"2026-11-02T09:46:29.723147Z","updated_at":"2026-11-02T09:46:29.723147Z"}

```

## same Idempotency-Key + same body -> 200, same id  [PASS]

```
{"id":"b08d8a7d-4185-442a-b75b-b35f8f707da7","patient_ref":"FAKE-AB12","amount":150000,"status":"PENDING","created_at":"2026-11-02T09:46:29.723147Z","updated_at":"2026-11-02T09:46:29.723147Z"}

```

## same key + different body -> 409  [PASS]

```

```

## no Idempotency-Key -> 400  [PASS]

```

```

## real-looking patient_ref refused -> 400  [PASS]

```

```

## worker approves claim (async)  [PASS]

```
{"id":"b08d8a7d-4185-442a-b75b-b35f8f707da7","patient_ref":"FAKE-AB12","amount":150000,"status":"APPROVED","decision_reason":"ok","created_at":"2026-11-02T09:46:29.723147Z","updated_at":"2026-11-02T09:46:31.548186Z"}

```

## worker rejects over-limit claim  [PASS]

```
{"id":"c7f3fd6f-e1bc-4f59-9bf7-4235c94b9bee","patient_ref":"FAKE-BIG1","amount":9000000,"status":"REJECTED","decision_reason":"over_limit","created_at":"2026-11-02T09:46:31.953634Z","updated_at":"2026-11-02T09:46:33.070809Z"}

```

## GET /claims?status=APPROVED lists it  [PASS]

```
{"items":[{"id":"b08d8a7d-4185-442a-b75b-b35f8f707da7","patient_ref":"FAKE-AB12","amount":150000,"status":"APPROVED","decision_reason":"ok","created_at":"2026-11-02T09:46:29.723147Z","updated_at":"2026-11-02T09:46:31.548186Z"}],"limit":5,"offset":0}

```

## GET by id is cached in Valkey  [PASS]

```
exists claim:b08d8a7d-4185-442a-b75b-b35f8f707da7 -> 1
```

## outbox drained (0 unsent)  [PASS]

```
unsent=0
```

## poison claim: retried then parked in DLQ, stays PENDING  [PASS]

```
{"reason":"max_attempts: transient evaluation failure: simulated downstream failure","original_topic":"claims.submitted","original_key":"a47c2830-c3a6-4d24-bd82-070d34860f77","value":{"event":"claims.submitted","version":1,"claim_id":"a47c2830-c3a6-4d24-bd82-070d34860f77","occurred_at":"2026-11-02T09:46:35.212923636Z"},"failed_at":"2026-11-02T09:46:38.074429132Z"}
Processed a total of 1 messages
status=PENDING
```

## Kafka down: API still accepts (201), event waits in outbox  [PASS]

```
code=201 unsent=1
```

## Kafka back: relay sends, claim gets APPROVED, nothing lost  [PASS]

```
status ok=True unsent=0
```

## worker and api exit 0 on SIGTERM  [PASS]

```
worker=0 api=0
```

## worker log sample (info only, for the record)  [PASS]

```
worker-1  | {"time":"2026-11-02T09:46:26.465661751Z","level":"INFO","msg":"worker started","group":"claims-worker","topic":"claims.submitted"}
worker-1  | {"time":"2026-11-02T09:46:36.57419921Z","level":"WARN","msg":"process failed","claim_id":"a47c2830-c3a6-4d24-bd82-070d34860f77","attempt":1,"err":"transient evaluation failure: simulated downstream failure"}
worker-1  | {"time":"2026-11-02T09:46:37.073213109Z","level":"WARN","msg":"process failed","claim_id":"a47c2830-c3a6-4d24-bd82-070d34860f77","attempt":2,"err":"transient evaluation failure: simulated downstream failure"}
worker-1  | {"time":"2026-11-02T09:46:38.07434995Z","level":"WARN","msg":"process failed","claim_id":"a47c2830-c3a6-4d24-bd82-070d34860f77","attempt":3,"err":"transient evaluation failure: simulated downstream failure"}
worker-1  | {"time":"2026-11-02T09:46:39.085657726Z","level":"ERROR","msg":"message sent to dlq","key":"a47c2830-c3a6-4d24-bd82-070d34860f77","reason":"max_attempts: transient evaluation failure: simulated downstream failure"}
worker-1  | {"time":"2026-11-02T09:47:21.143185638Z","level":"INFO","msg":"worker stopped"}
```

## docker compose down -v  [PASS]

```
... (24 earlier lines omitted)
 Container day08-postgres-1 Removed 
 Container day08-kafka-1 Stopped 
 Container day08-kafka-1 Removing 
 Container day08-kafka-1 Removed 
 Network day08_default Removing 
 Network day08_default Removed
```


