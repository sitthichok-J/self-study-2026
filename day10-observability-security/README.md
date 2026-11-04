# Day 10 — Observability + Security

ต่อจาก `day08-backend` (Claim Tracker Lite: Go + Postgres + Valkey + Kafka) ข้อมูลทั้งหมดเป็นข้อมูลสมมติ

โฟลเดอร์นี้เป็นสำเนาของ Day 8 ที่เติม 3 อย่าง: เห็นสถานะระบบ (metrics, trace, log), ล็อกการเข้าถึง (JWT + เจ้าของเคลม), และปิดช่องที่เจอตอนตรวจ

## ต้องอ่านก่อน: สถานะการตรวจ

| ส่วน | สถานะ |
|---|---|
| Go: obs, auth, service, handler, config, baseline | รันเทสต์แล้ว (107 เทสต์ รวม subtest 151, `-race`) |
| Python เทสต์ static (compose, Prometheus, dashboard, migration, secrets) | รันแล้ว 33 ข้อ |
| แก้โค้ดจงใจแล้วดูว่าเทสต์จับได้ไหม | Go 39 จุด จับได้ 39, static 23 จุด จับได้ 23 |
| `internal/adapter/postgres` + `cmd/api` + `cmd/worker` | คอมไพล์และรันกับ module จริง (`go test` ผ่านบนเครื่อง) ใช้งานกับ Postgres, Valkey, Kafka จริงใน compose แล้ว |
| `scripts/run-all.ps1` (ยก stack จริง) | **รันแล้ว** รอบที่ 3: 57/57 ขั้นผ่าน (รอบที่ 2: 55 ผ่าน 2 ล้ม แก้แล้ว ดูด้านล่าง) ผลอยู่ใน `notes/run-results.md` |
| compose + Prometheus + Grafana + Jaeger | **ขึ้นจริง** 8 service + migrate + kafka-init, scrape 2 target, Grafana provision dashboard ได้, Jaeger เห็น trace ข้าม api → Kafka → worker รูปอยู่ใน `notes/screenshots/` |
| alert rules, dashboard | Prometheus โหลด rule ครบ 7 ข้อ, alert ที่ควรดังดังจริง 2 ข้อ (DeadLetterGrowing, AccessDeniedSeen), ทุก query ใน dashboard รันได้ ยังไม่ผ่าน `promtool` ยังไม่ได้ดู dashboard ด้วยตา |
| govulncheck / gitleaks / Trivy | **รันแล้ว แก้แล้ว:** Trivy 17 → 0, image api 0, govulncheck "No vulnerabilities found" (ดู `docs/vulnerability-fixes.md`) gitleaks เหลือ 2 จุดที่คาดไว้ (`.env` ที่ git-ignore กับค่าปลอมในเทสต์) |
| Migration `000002_owner` | รันกับ Postgres จริงแล้ว (ตรวจคอลัมน์ `owner_id NOT NULL` ผ่าน) |

ผลรันจริงอยู่ใน `notes/run-results.md` (รอบล่าสุด) กับ `notes/security-scan.md` รอบก่อนแก้อยู่ใน `notes/run-results-run2-before-fix.md` รูปหน้าจออยู่ใน `notes/screenshots/`

สิ่งที่ผลรันจริงเจอ:

- รอบแรก 20 ล้ม เพราะ stack ของ Day 9 จองพอร์ต 9094 อยู่ ทำให้ stack Day 10 ไม่ขึ้นและสคริปต์ไปยิงใส่ stack เก่า ผลรอบนั้นใช้ไม่ได้ หยุด stack เก่าแล้วรันใหม่
- รอบที่ 2 ล้ม 2 ข้อ: (1) span `outbox publish` ไปอยู่คนละ trace กับ request เพราะ relay ไม่ได้อ่าน `traceparent` จาก payload แก้ที่ `internal/service/relay.go` พร้อมเทสต์ `TestOutboxPublishSpanJoinsRequestTrace` (ถอดการแก้แล้วเทสต์ล้มจริง) (2) สคริปต์ตรวจ alert เร็วเกินไป rule ประเมินทุก 5 วินาที แก้ให้รอสูงสุด 45 วินาที
- รอบที่ 3 ผ่านครบ 57/57
- สแกนรอบแรกเจอ dependency มีช่องโหว่ 17 รายการ แก้โดยขึ้น Go 1.26.10 และอัป pgx → 5.9.2, go-redis → 9.7.3, x/net → 0.60.0, x/text → 0.42.0, compress → 1.18.7 (x/crypto หลุดจาก go.mod) สแกนซ้ำสะอาด และ run-all ผ่าน 57/57 รายละเอียดใน `docs/vulnerability-fixes.md`

## ข้อแตกต่างจากแผน (สำคัญ)

แผนบอกให้ใช้ `prometheus/client_golang` และ OpenTelemetry SDK แต่ตอนเขียนโค้ดนี้ดึง Go module ภายนอกไม่ได้ (proxy ปฏิเสธ) ถ้าใช้ SDK จะคอมไพล์และเทสต์ไม่ได้เลย จึงเขียนด้วย stdlib ล้วนใน `internal/obs`:

- metrics: counter, histogram, gauge แล้วพิมพ์เป็น Prometheus text format (0.0.4) Prometheus อ่านได้ตามปกติ
- trace: W3C `traceparent`, span พ่อ-ลูก, ส่งออกด้วย OTLP/HTTP JSON (`POST /v1/traces`) Jaeger รับได้ที่พอร์ต 4318
- ยังไม่มี: sampling ที่ปรับได้, exemplar, metric ของ Kafka consumer lag, auto-instrument ของ DB/Kafka (ทำ span เองที่ชั้น service)
- ข้อดีคือเห็นว่าข้างในทำงานยังไง ข้อเสียคือไม่ใช่โค้ดที่ผ่านการใช้งานจริง ใช้งานจริงควรเปลี่ยนไปใช้ SDK มาตรฐาน (ส่วนที่ต้องแก้อยู่ที่ `internal/obs` กับ `cmd/*/main.go`)

## ทำอะไรไป

### Observability

| เรื่อง | ทำที่ไหน | หมายเหตุ |
|---|---|---|
| `/metrics` | listener แยก `:9090` (`obs.ServeMetrics`) | ไม่อยู่ที่พอร์ต API (8080) และ compose ไม่ publish ออกเครื่อง |
| RED | `http_requests_total{method,route,code}`, `http_request_duration_seconds{method,route}` | label `route` เป็น template เช่น `/claims/{id}` ถ้าเป็น 404 ใช้ `unmatched` กัน series บวม |
| ธุรกิจ | `claims_submitted_total`, `claims_decided_total`, `outbox_publish_total`, `consumer_messages_total{outcome}` | outcome = processed / retried / dlq |
| USE | `db_pool_*`, `go_goroutines`, `go_heap_alloc_bytes`, `http_requests_in_flight` | pool ของ pgx เท่านั้น ยังไม่มี Kafka lag |
| Security | `auth_failures_total{reason}`, `authz_denied_total` | reason มีชุดตายตัว 10 ค่า |
| Trace | span ของ HTTP, `repo.*`, `cache.Get`, `outbox publish`, `consume ...` | trace ข้าม Kafka ด้วยการใส่ `traceparent` ใน payload ของ event (ไม่ต้องแก้ adapter ของ Kafka) |
| Log | `slog` JSON + `trace_id` `span_id` `request_id` | key อ่อนไหวถูกแทนด้วย `[redacted]` |
| Dashboard | `deploy/grafana/dashboards/claims-overview.json` (20 รายการ: 4 แถว + 16 panel) | provision ให้เองตอนเปิด Grafana |
| Alert | `deploy/prometheus/alerts.yml` 7 ข้อ | ApiDown, HighErrorRate, HighLatencyP95, DeadLetterGrowing, AuthFailureSpike, AccessDeniedSeen, DbPoolSaturated |

เรื่องที่เจอระหว่างทำ: counter ที่เพิ่งเกิดมาที่ค่า 1 `increase()` ของ Prometheus จะมองไม่เห็นการเพิ่มครั้งแรก alert "DLQ > 0" จะพลาดข้อความแรกที่เข้า DLQ จึงสร้างทุก series ที่รู้จักไว้ที่ 0 ตั้งแต่สตาร์ท (มีเทสต์)

### Security

รายละเอียดอยู่ใน [SECURITY.md](SECURITY.md) สรุป: ทุก route ของ `/claims` ต้องมี JWT (HS256, ตรวจ exp/iss/aud/sub), เคลมมีเจ้าของ, เปิดเคลมคนอื่นได้ 404, รายการและ cache แยกตามผู้ใช้, idempotency key แยกตามเจ้าของ, header ความปลอดภัย, port ผูก 127.0.0.1, container read-only non-root

บั๊กจริงที่เจอตอนตรวจ (อยู่ใน Day 8): idempotency key ไม่ซ้ำกันทั้งระบบ ผู้ใช้คนที่สองส่ง key เดียวกันจะได้ 409 หรือได้เคลมของคนแรกกลับมา แก้ด้วย migration `000002_owner` (`UNIQUE (owner_id, idempotency_key)`)

พฤติกรรมของ Day 8 ที่ไม่ควรเปลี่ยนยังเป็นเหมือนเดิม: ส่วน `decisions`, `http`, `limits`, `topics`, `transitions` ใน `testdata/baseline.json` เทียบกับของ Day 8 แล้วไม่เปลี่ยน (ผู้ใช้ที่ล็อกอินแล้วเห็นผลเหมือนเดิม) ที่เพิ่มคือ `security_http`, `security_headers`, `jwt`, `metric_families`

## วิธีรัน

ต้องมี Docker Desktop

```
scripts\run-all.bat          # รันทุกอย่าง เขียนผลที่ notes\run-results.md แล้วปิด stack
scripts\run-all-keep.bat     # เหมือนกัน แต่ปล่อย stack ไว้ให้เปิดดู
scripts\security-scan.bat    # govulncheck + gitleaks + Trivy -> notes\security-scan.md
scripts\stack-down.bat
```

สคริปต์สร้าง `JWT_SECRET` กับรหัส Grafana แบบสุ่มลง `.env` (ถูก git-ignore) ครั้งแรกที่รัน ตอน `-Keep` เปิดดูได้ที่:

| อะไร | ที่อยู่ |
|---|---|
| API | http://localhost:8080 (ต้องมี token) |
| Grafana | http://localhost:3001 ผู้ใช้ `admin` รหัสอยู่ใน `.env` |
| Jaeger | http://localhost:16686 |
| Prometheus | http://localhost:9091 |

ออก token ทดลองเอง:

```
docker build --build-arg CMD=mint -t day10-mint .
docker run --rm -e JWT_SECRET=<ค่าจาก .env> day10-mint -sub alice
```

รันเทสต์เองโดยไม่ใช้สคริปต์:

```
docker run --rm -v "%CD%:/src" -w /src golang:1.25 go test ./...
docker run --rm -v "%CD%:/src" -w /src python:3.12-alpine sh -c "pip install -q pyyaml && python -m unittest discover -s tests -v"
```

## โครงสร้างที่เพิ่ม

```
internal/obs/        metrics, trace, OTLP, logging, middleware (stdlib)
internal/auth/       JWT HS256 + Principal
cmd/mint/            ออก token ทดลอง
migrations/000002_owner.*.sql
deploy/              prometheus, grafana (provisioning + dashboard)
scripts/             run-all, security-scan
tests/test_infra.py  เทสต์ static
SECURITY.md  notes/exercises.md  docs/two-week-summary.md
```

## ข้อจำกัดที่รู้

- ยังไม่มี UI/Kong ของ Day 9 ที่ส่ง token (UI ไม่มีหน้า login, Kong ไม่ตรวจ JWT) ใช้ Day 9 ต่อกับ API ตัวนี้ยังไม่ได้
- ไม่มี refresh token, revocation, key rotation, TLS, rate limit ที่ API
- ยังไม่ได้วัด Kafka consumer lag
- เปลี่ยน signature ของ `Get`/`List`/`Submit` ใน service เพื่อรับผู้เรียก เทสต์ของ Day 8 ถูกแก้ตามนั้น (ตรรกะเดิมยังผ่านทั้งหมด)
