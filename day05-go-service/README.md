# day05-go-service

REST API จัดการ claims (CRUD) เขียนด้วย Go มาตรฐาน (net/http) โครงสร้าง cmd/internal
มี slog JSON log + request_id, graceful shutdown, Postgres + migration, pprof

## รัน

```
# in-memory (ไม่ต้องมี DB)
go run ./cmd/api

# Postgres
docker compose up -d
go get github.com/jackc/pgx/v5
export DATABASE_URL='postgres://claims:claims@localhost:5433/claims?sslmode=disable'
# apply migrations/000001_create_claims.up.sql ด้วย psql หรือ migrate
go run -tags pgx ./cmd/api
```

ตัวแปร: ADDR (:8080), DATABASE_URL, LOG_LEVEL, SHUTDOWN_TIMEOUT, PPROF_ADDR, DB_MAX_OPEN_CONNS, DB_MAX_IDLE_CONNS, DB_CONN_LIFETIME

## Endpoint

| method | path | หมายเหตุ |
|---|---|---|
| GET | /healthz | |
| POST | /claims | สร้าง (status เริ่มต้น PENDING) |
| GET | /claims?status=&limit=&offset= | limit เริ่มต้น 20 สูงสุด 100 |
| GET/PUT/DELETE | /claims/{id} | |
| POST | /claims/approve | `{"ids":[...]}` all-or-nothing ใน transaction |
| GET | /claims/report | ยอดรวมต่อโรงพยาบาล |

จำนวนเงินเก็บเป็นสตางค์ (int64)

## เทสต์

```
go test ./... -race -cover
go test ./internal/report -bench . -benchmem
bash scripts/check_sql.sh          # ต้องมี Postgres: PGURL=... (ตรวจ migration + query จริงด้วย psql)
TEST_DATABASE_URL=... go test -tags 'integration pgx' ./internal/repository
```

baseline: `go test .` เทียบกับ `testdata/baseline.golden` (อัปเดตด้วย `-update`)

## สิ่งที่ยังไม่ได้ทดสอบ

ตอนทำใน sandbox ดึง pgx / testcontainers / golang-migrate ไม่ได้ จึง:
- ตรวจ migration และ query ทั้งหมดกับ Postgres 16 จริงด้วย `scripts/check_sql.sh`
- แต่ integration test ฝั่ง Go (`-tags 'integration pgx'`) และ build `-tags pgx` ยังไม่เคยรัน ต้อง `go get` pgx ก่อน

ดู notes/profile.md สำหรับผล profile ก่อน/หลัง
