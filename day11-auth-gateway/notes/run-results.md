# Day 11 run results

Run at: 2026-11-05 17:05:20  Machine: LAPTOP-MICOTO

## Summary

86 / 86 steps passed

| Step | Result |
|---|---|
| Docker running | PASS |
| .env มี JWT_SECRET (>=32 ตัวอักษร), รหัส Grafana และรหัส demo user, ไฟล์ถูก git-ignore | PASS |
| go test ./... (unit + baseline snapshot) | PASS |
| go vet ./... | PASS |
| เทสต์ static (compose, Prometheus, dashboard, migration, secrets) | PASS |
| render config ของ Kong (ใส่ secret) และสร้าง cert สำหรับ localhost | PASS |
| docker compose up --build (api, worker, kong, web, postgres, valkey, kafka, prometheus, grafana, jaeger) | PASS |
| api ready (GET /readyz 200) และ migration 000002 ผ่าน | PASS |
| ตาราง claims มีคอลัมน์ owner_id (NOT NULL) | PASS |
| migration 000003: ตาราง users และ refresh_tokens มีอยู่ | PASS |
| Kong และ UI ขึ้น (UI /healthz 200, Kong ตอบ 401 เมื่อไม่มี token) | PASS |
| cmd/mint ออก token ได้ (alice, bob, ops=staff) | PASS |
| secret สั้นกว่า 32 byte ถูกปฏิเสธ | PASS |
| ไม่มี token -> 401 พร้อม WWW-Authenticate | PASS |
| payload ถูกแก้ -> 401 | PASS |
| alg=none -> 401 | PASS |
| scheme ไม่ใช่ Bearer -> 401 | PASS |
| /healthz ไม่ต้องมี token | PASS |
| alice POST /claims -> 201 PENDING, owner_id = alice | PASS |
| alice ส่ง key เดิมซ้ำ -> 200 id เดิม | PASS |
| bob ใช้ key เดียวกับ alice (body ต่าง) -> 201 เคลมใหม่ของ bob ไม่ใช่ 409 | PASS |
| IDOR: bob เปิดเคลมของ alice -> 404 เหมือนเคลมที่ไม่มีอยู่ และไม่มีข้อมูลหลุด | PASS |
| alice เปิดเคลมของตัวเอง -> 200 | PASS |
| ops (staff) เปิดเคลมของ alice -> 200 | PASS |
| alice list เห็นเฉพาะของตัวเอง | PASS |
| staff list เห็นของทุกคน | PASS |
| ส่ง owner_id มาใน body -> 400 (เจ้าของมาจาก token เท่านั้น) | PASS |
| security headers ครบ | PASS |
| /metrics ไม่มีบน port ของ API (404 แม้เป็น staff) | PASS |
| port metrics (9090) ไม่ถูก publish ออกเครื่อง | PASS |
| X-Request-Id ที่ถูกต้องถูกส่งกลับ | PASS |
| X-Request-Id แปลกๆ ถูกแทนด้วยค่าใหม่ 32 ตัว | PASS |
| port ที่ publish ผูก 127.0.0.1 ทั้งหมด | PASS |
| worker ตัดสินเป็น APPROVED (ผ่านเส้นทางเดิม outbox -> Kafka -> worker) | PASS |
| เกิน 50,000 บาท -> REJECTED over_limit (กฎของ Day 8 ยังอยู่) | PASS |
| ส่งเคลม poison (ต้องล้มทุกครั้ง) -> 201 | PASS |
| Prometheus scrape api + worker ได้ (2 targets up) | PASS |
| http_requests_total มี route template /claims/{id} (ไม่ใช่ id ดิบ) | PASS |
| label route มีแต่ template ไม่มี uuid | PASS |
| claims_submitted_total{result=created} >= 3 | PASS |
| claims_decided_total (worker) > 0 และ REJECTED/over_limit มีอยู่ | PASS |
| auth_failures_total นับครั้งที่ 401 (มี reason แยก) | PASS |
| authz_denied_total นับการเปิดเคลมคนอื่น | PASS |
| เคลม poison เข้า DLQ: consumer_messages_total{outcome=dlq} >= 1 และ retried >= 2 | PASS |
| db_pool_max_conns มีค่า (pool gauge ทำงาน) | PASS |
| Prometheus โหลด alert rules ครบ 7 ข้อ | PASS |
| alert ที่ควรดังตอนนี้: DeadLetterGrowing และ AccessDeniedSeen (จากที่ยิงทดสอบ) | PASS |
| Grafana ขึ้น (GET /api/health) | PASS |
| dashboard claims-overview ถูก provision (>= 15 panels) | PASS |
| Grafana ไม่ยอมให้เข้าโดยไม่ login | PASS |
| datasource Prometheus + Jaeger ถูก provision | PASS |
| ทุก query ใน dashboard รันบน Prometheus ได้ (status success) | PASS |
| Jaeger มี trace ของ claims-api | PASS |
| trace เดียวผ่านทั้ง claims-api และ claims-worker (trace context ข้าม Kafka ผ่าน payload) | PASS |
| trace นั้นมี span ของ repo.Create, outbox publish และ consume | PASS |
| log ของ request เป็น JSON มี trace_id, span_id, request_id | PASS |
| trace_id ในบรรทัด log เปิดหา trace เดียวกันใน Jaeger ได้ | PASS |
| log มี security_event ทั้ง auth_failed และ access_denied | PASS |
| log ของ api/worker ไม่มี token, JWT_SECRET หรือ patient_ref | PASS |
| api รันเป็น non-root | PASS |
| login ผิด: user มีอยู่กับไม่มีอยู่ ตอบ 401 ข้อความเดียวกันทุกตัวอักษร | PASS |
| login ผ่าน Kong -> 200 access token (sub=alice) + refresh cookie HttpOnly; Secure; SameSite=Strict; Path=/api/auth | PASS |
| refresh token ไม่อยู่ใน body ของ login (อยู่ใน cookie เท่านั้น) | PASS |
| DB เก็บเฉพาะ SHA-256 ของ refresh token (ค่าดิบค้นไม่เจอ, ค่า hash เจอ 1 แถว) | PASS |
| ผ่าน Kong ไม่มี token -> 401 (ตอบโดย Kong ไม่ถึง API) | PASS |
| ผ่าน Kong ด้วย access token ของ alice -> 200 เห็นเฉพาะของตัวเอง | PASS |
| ผ่าน Kong payload ถูกแก้ -> 401 จาก Kong | PASS |
| ผ่าน Kong alg=none -> 401 จาก Kong | PASS |
| Kong ไม่รับ token ใน query string (ใช้ token ที่ออกใหม่ ไม่ใช่ของ session) | PASS |
| route ที่ Kong ไม่ได้เปิด (/metrics) ได้ 404 ส่วน API ตรงตอบได้เฉพาะผ่าน 127.0.0.1 | PASS |
| refresh ไม่มี header X-Requested-With -> 403 (กัน CSRF) | PASS |
| refresh -> 200 access token ใหม่ (sub=alice) และ cookie ใหม่ไม่ซ้ำของเดิม (หมุนเวียน) | PASS |
| เทียบ: refresh ยิงตรงเข้า API (ไม่ผ่าน Kong) หมุน token ได้ -> 200 | PASS |
| ใช้ refresh token เดิมซ้ำ -> 401 (ตรวจพบการใช้ซ้ำ) | PASS |
| หลังตรวจพบการใช้ซ้ำ token ใหม่ของสายเดียวกันถูกเพิกถอนด้วย -> 401 | PASS |
| DB: แถว refresh ของ alice ที่ถูกเพิกถอนมีอย่างน้อย 1 แถว | PASS |
| refresh ไม่มี cookie -> 401 | PASS |
| refresh ของ bob ได้ token ในนาม bob เท่านั้น (sub มาจากแถวใน DB ไม่ใช่จาก client) | PASS |
| logout -> 204 ล้าง cookie และ refresh token นั้นใช้ต่อไม่ได้ (401) | PASS |
| CORS ของ /api/auth: origin ของ UI ได้ credentials=true, origin อื่นไม่ได้ header allow | PASS |
| TLS ที่ Kong: cert ตรวจผ่านเมื่อ pin cert ของเครื่อง และตอบ 401 เมื่อไม่มี token | PASS |
| TLS: ถ้าไม่ pin cert (self-signed) curl ปฏิเสธการเชื่อมต่อ (http_code 000) | PASS |
| UI (nginx) เสิร์ฟหน้า SPA และมี security header | PASS |
| rate limit ที่ /api/auth/login: ยิงผิด 30 ครั้งติดกัน -> เจอ 429 พร้อมโควต้าใน header | PASS |
| log ของ Kong และ API ไม่มี refresh token, access token หรือรหัสผ่าน demo | PASS |
| log API มี security_event: login_failed, refresh_reuse, csrf_header_missing | PASS |

## Docker running  [PASS]

```
client 29.7.2 / server 29.7.2
```

## .env มี JWT_SECRET (>=32 ตัวอักษร), รหัส Grafana และรหัส demo user, ไฟล์ถูก git-ignore  [PASS]

```
JWT_SECRET length=64 (ค่าไม่แสดง)
```

## go test ./... (unit + baseline snapshot)  [PASS]

```
?   	github.com/sitthichok-J/self-study-2026/day11-auth-gateway/cmd/api	[no test files]
?   	github.com/sitthichok-J/self-study-2026/day11-auth-gateway/cmd/mint	[no test files]
?   	github.com/sitthichok-J/self-study-2026/day11-auth-gateway/cmd/worker	[no test files]
ok  	github.com/sitthichok-J/self-study-2026/day11-auth-gateway/internal/adapter/http	(cached)
?   	github.com/sitthichok-J/self-study-2026/day11-auth-gateway/internal/adapter/kafka	[no test files]
?   	github.com/sitthichok-J/self-study-2026/day11-auth-gateway/internal/adapter/postgres	[no test files]
?   	github.com/sitthichok-J/self-study-2026/day11-auth-gateway/internal/adapter/redis	[no test files]
ok  	github.com/sitthichok-J/self-study-2026/day11-auth-gateway/internal/auth	(cached)
ok  	github.com/sitthichok-J/self-study-2026/day11-auth-gateway/internal/baseline	(cached)
ok  	github.com/sitthichok-J/self-study-2026/day11-auth-gateway/internal/config	(cached)
ok  	github.com/sitthichok-J/self-study-2026/day11-auth-gateway/internal/domain	(cached)
?   	github.com/sitthichok-J/self-study-2026/day11-auth-gateway/internal/memory	[no test files]
ok  	github.com/sitthichok-J/self-study-2026/day11-auth-gateway/internal/obs	(cached)
?   	github.com/sitthichok-J/self-study-2026/day11-auth-gateway/internal/port	[no test files]
ok  	github.com/sitthichok-J/self-study-2026/day11-auth-gateway/internal/service	(cached)
```

## go vet ./...  [PASS]

```

```

## เทสต์ static (compose, Prometheus, dashboard, migration, secrets)  [PASS]

```
... (41 earlier lines omitted)
test_no_hardcoded_secret_assignments_outside_tests (test_infra.TestNoSecretsInRepo.test_no_hardcoded_secret_assignments_outside_tests) ... ok
test_no_private_keys_cloud_keys_or_real_looking_tokens (test_infra.TestNoSecretsInRepo.test_no_private_keys_cloud_keys_or_real_looking_tokens) ... ok
test_alert_metrics_and_labels_exist_in_go (test_infra.TestPrometheus.test_alert_metrics_and_labels_exist_in_go) ... ok
test_alert_rules_are_well_formed (test_infra.TestPrometheus.test_alert_rules_are_well_formed) ... ok
test_rules_file_is_mounted_where_prometheus_expects_it (test_infra.TestPrometheus.test_rules_file_is_mounted_where_prometheus_expects_it) ... ok
test_scrape_targets_are_the_internal_metrics_ports (test_infra.TestPrometheus.test_scrape_targets_are_the_internal_metrics_ports) ... ok
test_the_required_alerts_exist (test_infra.TestPrometheus.test_the_required_alerts_exist) ... ok

----------------------------------------------------------------------
Ran 44 tests in 2.209s

OK
```

## render config ของ Kong (ใส่ secret) และสร้าง cert สำหรับ localhost  [PASS]

```
rendered gateway/.rendered/kong.yml and certs/
```

## docker compose up --build (api, worker, kong, web, postgres, valkey, kafka, prometheus, grafana, jaeger)  [PASS]

```
... (239 earlier lines omitted)
 Container day11-valkey-1 Waiting 
 Container day11-migrate-1 Waiting 
 Container day11-migrate-1 Waiting 
 Container day11-kafka-init-1 Waiting 
 Container day11-valkey-1 Waiting 
 Container day11-migrate-1 Exited 
 Container day11-migrate-1 Exited 
 Container day11-valkey-1 Healthy 
 Container day11-valkey-1 Healthy 
 Container day11-kafka-init-1 Exited 
 Container day11-api-1 Starting 
 Container day11-kafka-init-1 Exited 
 Container day11-worker-1 Starting 
 Container day11-worker-1 Started 
 Container day11-api-1 Started 
 Container day11-gateway-1 Starting 
 Container day11-prometheus-1 Starting 
 Container day11-prometheus-1 Started 
 Container day11-grafana-1 Starting 
 Container day11-gateway-1 Started 
 Container day11-gateway-1 Waiting 
 Container day11-grafana-1 Started 
 Container day11-gateway-1 Healthy 
 Container day11-web-1 Starting 
 Container day11-web-1 Started
```

## api ready (GET /readyz 200) และ migration 000002 ผ่าน  [PASS]

```
NAME                 IMAGE                             COMMAND                  SERVICE      CREATED          STATUS                            PORTS
day11-api-1          day11-api                         "/app"                   api          26 seconds ago   Up 14 seconds                     127.0.0.1:8080->8080/tcp
day11-gateway-1      kong:3.9.1                        "/docker-entrypoint.…"   gateway      26 seconds ago   Up 13 seconds (healthy)           127.0.0.1:8000->8000/tcp, 127.0.0.1:8443->8443/tcp
day11-grafana-1      grafana/grafana:11.3.0            "/run.sh"                grafana      26 seconds ago   Up 13 seconds                     127.0.0.1:3001->3000/tcp
day11-jaeger-1       jaegertracing/all-in-one:1.62.0   "/go/bin/all-in-one-…"   jaeger       27 seconds ago   Up 25 seconds                     127.0.0.1:16686->16686/tcp
day11-kafka-1        apache/kafka:3.8.0                "/__cacert_entrypoin…"   kafka        27 seconds ago   Up 25 seconds (healthy)           127.0.0.1:9094->9094/tcp
day11-kafka-init-1   apache/kafka:3.8.0                "/bin/sh -c '/opt/ka…"   kafka-init   27 seconds ago   Exited (0) 14 seconds ago         
day11-migrate-1      migrate/migrate:v4.18.1           "migrate -path=/migr…"   migrate      27 seconds ago   Exited (0) 21 seconds ago         
day11-postgres-1     postgres:16-alpine                "docker-entrypoint.s…"   postgres     27 seconds ago   Up 25 seconds (healthy)           127.0.0.1:5432->5432/tcp
day11-prometheus-1   prom/prometheus:v2.55.1           "/bin/prometheus --c…"   prometheus   26 seconds ago   Up 13 seconds                     127.0.0.1:9091->9090/tcp
day11-valkey-1       valkey/valkey:8-alpine            "tini -- docker-entr…"   valkey       27 seconds ago   Up 25 seconds (healthy)           127.0.0.1:6379->6379/tcp
day11-web-1          day11-web                         "/docker-entrypoint.…"   web          26 seconds ago   Up 2 seconds (health: starting)   127.0.0.1:3000->8080/tcp
day11-worker-1       day11-worker                      "/app"                   worker       26 seconds ago   Up 14 seconds
```

## ตาราง claims มีคอลัมน์ owner_id (NOT NULL)  [PASS]

```
owner_id|NO
```

## migration 000003: ตาราง users และ refresh_tokens มีอยู่  [PASS]

```
2
```

## Kong และ UI ขึ้น (UI /healthz 200, Kong ตอบ 401 เมื่อไม่มี token)  [PASS]

```
web=200 gateway=401 body={"message":"Unauthorized"}
```

## cmd/mint ออก token ได้ (alice, bob, ops=staff)  [PASS]

```
token ไม่แสดงในไฟล์ผล
```

## secret สั้นกว่า 32 byte ถูกปฏิเสธ  [PASS]

```
error: JWT secret must be at least 32 bytes
```

## ไม่มี token -> 401 พร้อม WWW-Authenticate  [PASS]

```
code=401 body={"error":"unauthorized"}

```

## payload ถูกแก้ -> 401  [PASS]

```
{"error":"unauthorized"}

```

## alg=none -> 401  [PASS]

```
{"error":"unauthorized"}

```

## scheme ไม่ใช่ Bearer -> 401  [PASS]

```
{"error":"unauthorized"}

```

## /healthz ไม่ต้องมี token  [PASS]

```
{"status":"ok"}

```

## alice POST /claims -> 201 PENDING, owner_id = alice  [PASS]

```
{"id":"4736a24e-093a-4c27-b855-d639baa58cd8","owner_id":"alice","patient_ref":"FAKE-AB12","amount":150000,"status":"PENDING","created_at":"2026-11-05T10:06:18.083744Z","updated_at":"2026-11-05T10:06:18.083744Z"}

```

## alice ส่ง key เดิมซ้ำ -> 200 id เดิม  [PASS]

```
{"id":"4736a24e-093a-4c27-b855-d639baa58cd8","owner_id":"alice","patient_ref":"FAKE-AB12","amount":150000,"status":"PENDING","created_at":"2026-11-05T10:06:18.083744Z","updated_at":"2026-11-05T10:06:18.083744Z"}

```

## bob ใช้ key เดียวกับ alice (body ต่าง) -> 201 เคลมใหม่ของ bob ไม่ใช่ 409  [PASS]

```
{"id":"886f9d29-ec1b-4c7b-b7c5-ae820260c380","owner_id":"bob","patient_ref":"FAKE-ZZ99","amount":777,"status":"PENDING","created_at":"2026-11-05T10:06:18.154487Z","updated_at":"2026-11-05T10:06:18.154487Z"}

```

## IDOR: bob เปิดเคลมของ alice -> 404 เหมือนเคลมที่ไม่มีอยู่ และไม่มีข้อมูลหลุด  [PASS]

```
code=404 body={"error":"claim not found"}

```

## alice เปิดเคลมของตัวเอง -> 200  [PASS]

```
{"id":"4736a24e-093a-4c27-b855-d639baa58cd8","owner_id":"alice","patient_ref":"FAKE-AB12","amount":150000,"status":"PENDING","created_at":"2026-11-05T10:06:18.083744Z","updated_at":"2026-11-05T10:06:18.083744Z"}

```

## ops (staff) เปิดเคลมของ alice -> 200  [PASS]

```
{"id":"4736a24e-093a-4c27-b855-d639baa58cd8","owner_id":"alice","patient_ref":"FAKE-AB12","amount":150000,"status":"PENDING","created_at":"2026-11-05T10:06:18.083744Z","updated_at":"2026-11-05T10:06:18.083744Z"}

```

## alice list เห็นเฉพาะของตัวเอง  [PASS]

```
items=1 ของคนอื่น=0
```

## staff list เห็นของทุกคน  [PASS]

```
owners: alice,bob
```

## ส่ง owner_id มาใน body -> 400 (เจ้าของมาจาก token เท่านั้น)  [PASS]

```
{"error":"request body must be valid JSON with patient_ref and amount"}

```

## security headers ครบ  [PASS]

```
nosniff=nosniff cache=no-store csp=default-src 'none'; frame-ancestors 'none'
```

## /metrics ไม่มีบน port ของ API (404 แม้เป็น staff)  [PASS]

```
code=404
```

## port metrics (9090) ไม่ถูก publish ออกเครื่อง  [PASS]

```
Unable to connect to the remote server
```

## X-Request-Id ที่ถูกต้องถูกส่งกลับ  [PASS]

```
run-req-123
```

## X-Request-Id แปลกๆ ถูกแทนด้วยค่าใหม่ 32 ตัว  [PASS]

```
086c204b5ec11d8879b56f5cdda19311
```

## port ที่ publish ผูก 127.0.0.1 ทั้งหมด  [PASS]

```
day11-api-1 127.0.0.1:8080->8080/tcp
day11-gateway-1 127.0.0.1:8000->8000/tcp, 127.0.0.1:8443->8443/tcp
day11-grafana-1 127.0.0.1:3001->3000/tcp
day11-jaeger-1 127.0.0.1:16686->16686/tcp
day11-kafka-1 127.0.0.1:9094->9094/tcp
day11-postgres-1 127.0.0.1:5432->5432/tcp
day11-prometheus-1 127.0.0.1:9091->9090/tcp
day11-valkey-1 127.0.0.1:6379->6379/tcp
day11-web-1 127.0.0.1:3000->8080/tcp
day11-worker-1
```

## worker ตัดสินเป็น APPROVED (ผ่านเส้นทางเดิม outbox -> Kafka -> worker)  [PASS]

```
{"id":"4736a24e-093a-4c27-b855-d639baa58cd8","owner_id":"alice","patient_ref":"FAKE-AB12","amount":150000,"status":"APPROVED","decision_reason":"ok","created_at":"2026-11-05T10:06:18.083744Z","updated_at":"2026-11-05T10:06:19.21746Z"}

```

## เกิน 50,000 บาท -> REJECTED over_limit (กฎของ Day 8 ยังอยู่)  [PASS]

```
{"id":"d2f0aebf-6026-43b4-ad19-6e2ada65ffe0","owner_id":"alice","patient_ref":"FAKE-OVER1","amount":600000000,"status":"REJECTED","decision_reason":"over_limit","created_at":"2026-11-05T10:06:23.035566Z","updated_at":"2026-11-05T10:06:24.179265Z"}

```

## ส่งเคลม poison (ต้องล้มทุกครั้ง) -> 201  [PASS]

```
{"id":"3d438aa8-0946-4fa9-a4dd-93738bc892f1","owner_id":"alice","patient_ref":"FAKE-POISON","amount":100,"status":"PENDING","created_at":"2026-11-05T10:06:25.116137Z","updated_at":"2026-11-05T10:06:25.116137Z"}

```

## Prometheus scrape api + worker ได้ (2 targets up)  [PASS]

```
claims-api=up, claims-worker=up
```

## http_requests_total มี route template /claims/{id} (ไม่ใช่ id ดิบ)  [PASS]

```
sum=7
```

## label route มีแต่ template ไม่มี uuid  [PASS]

```
/claims, /claims/{id}, /healthz, /readyz, unmatched
```

## claims_submitted_total{result=created} >= 3  [PASS]

```
value=4
```

## claims_decided_total (worker) > 0 และ REJECTED/over_limit มีอยู่  [PASS]

```
decided=3
```

## auth_failures_total นับครั้งที่ 401 (มี reason แยก)  [PASS]

```
total=4
```

## authz_denied_total นับการเปิดเคลมคนอื่น  [PASS]

```
value=1
```

## เคลม poison เข้า DLQ: consumer_messages_total{outcome=dlq} >= 1 และ retried >= 2  [PASS]

```
dlq=1 retried=2
```

## db_pool_max_conns มีค่า (pool gauge ทำงาน)  [PASS]

```
max=16
```

## Prometheus โหลด alert rules ครบ 7 ข้อ  [PASS]

```
ApiDown, HighErrorRate, HighLatencyP95, DeadLetterGrowing, AuthFailureSpike, AccessDeniedSeen, DbPoolSaturated
```

## alert ที่ควรดังตอนนี้: DeadLetterGrowing และ AccessDeniedSeen (จากที่ยิงทดสอบ)  [PASS]

```
firing: DeadLetterGrowing, AccessDeniedSeen
```

## Grafana ขึ้น (GET /api/health)  [PASS]

```
{
  "database": "ok",
  "version": "11.3.0",
  "commit": "d9455ff7db73b694db7d412e49a68bec767f2b5a"
}
```

## dashboard claims-overview ถูก provision (>= 15 panels)  [PASS]

```
code=200 panels=20
```

## Grafana ไม่ยอมให้เข้าโดยไม่ login  [PASS]

```
code=401
```

## datasource Prometheus + Jaeger ถูก provision  [PASS]

```
Jaeger,Prometheus
```

## ทุก query ใน dashboard รันบน Prometheus ได้ (status success)  [PASS]

```
ทุก panel ผ่าน
```

## Jaeger มี trace ของ claims-api  [PASS]

```
traces=3
```

## trace เดียวผ่านทั้ง claims-api และ claims-worker (trace context ข้าม Kafka ผ่าน payload)  [PASS]

```
repo.Get | repo.UpdateStatus | consume claims.submitted | outbox publish claims.submitted | repo.Create | POST /claims
```

## trace นั้นมี span ของ repo.Create, outbox publish และ consume  [PASS]

```
repo.Get | repo.UpdateStatus | consume claims.submitted | outbox publish claims.submitted | repo.Create | POST /claims
```

## log ของ request เป็น JSON มี trace_id, span_id, request_id  [PASS]

```
api-1     | {"time":"2026-11-05T10:06:18.090947277Z","level":"INFO","msg":"request","method":"POST","route":"/claims","status":201,"duration_ms":7.346,"bytes":212,"trace_id":"cbe990c5e6177fbc8eba26d9d914c264","span_id":"836a78b2e38b9d43","request_id":"b8b6ac1a9ce7a79c37d763c6411c67db"}
```

## trace_id ในบรรทัด log เปิดหา trace เดียวกันใน Jaeger ได้  [PASS]

```
trace_id=cbe990c5e6177fbc8eba26d9d914c264
```

## log มี security_event ทั้ง auth_failed และ access_denied  [PASS]

```
... (1 earlier lines omitted)
api-1     | {"time":"2026-11-05T10:06:17.996825657Z","level":"WARN","msg":"authentication failed","security_event":"auth_failed","reason":"signature","trace_id":"4902d1ae571b6f5449b2e2300eb8bbd4","span_id":"1e0a4a1bc20e66f9","request_id":"3d193600fed592ff050f408981c3b38d"}
api-1     | {"time":"2026-11-05T10:06:18.008918016Z","level":"WARN","msg":"authentication failed","security_event":"auth_failed","reason":"alg","trace_id":"6bbdd2ec1754e8814a96dfa4c070c47a","span_id":"21dca7a807a5dbe2","request_id":"5b4807a23f3b6e9971c436d5064f11b3"}
api-1     | {"time":"2026-11-05T10:06:18.033032933Z","level":"WARN","msg":"authentication failed","security_event":"auth_failed","reason":"missing","trace_id":"35c4c4fe12fce3956e780f1083678ace","span_id":"d8b3f241f926ca5f","request_id":"e85b47dae394d566599a2cd62216648d"}
api-1     | {"time":"2026-11-05T10:06:18.208329066Z","level":"WARN","msg":"access denied to claim","security_event":"access_denied","sub":"bob","claim_id":"4736a24e-093a-4c27-b855-d639baa58cd8","trace_id":"eafca7ce4d362369474aebf283b212f5","span_id":"f20d2551ddabd09a","request_id":"1facd60abb5efee0c07bfd4c129cd1b1"}
```

## log ของ api/worker ไม่มี token, JWT_SECRET หรือ patient_ref  [PASS]

```
สะอาด
```

## api รันเป็น non-root  [PASS]

```
SERVICE  #   UID    PID    PPID   C   STIME  TTY  TIME      CMD
api      1   65532  97872  97846  1   10:05  ?    00:00:00  /app
```

## login ผิด: user มีอยู่กับไม่มีอยู่ ตอบ 401 ข้อความเดียวกันทุกตัวอักษร  [PASS]

```
code=401/401 body={"error":"invalid username or password"}
 เวลา 4115ms vs 108ms (ตัวอย่างเดียว ไม่ใช่การพิสูจน์เรื่อง timing)
```

## login ผ่าน Kong -> 200 access token (sub=alice) + refresh cookie HttpOnly; Secure; SameSite=Strict; Path=/api/auth  [PASS]

```
code=200 sub=alice expires_in=600 cookie(ไม่แสดงค่า)=refresh_token=[value]; Path=/api/auth; Expires=Thu, 05 Nov 2026 22:06:55 GMT; Max-Age=43200; HttpOnly; Secure; SameSite=Strict
```

## refresh token ไม่อยู่ใน body ของ login (อยู่ใน cookie เท่านั้น)  [PASS]

```
body มีแค่ access_token, token_type, expires_in
```

## DB เก็บเฉพาะ SHA-256 ของ refresh token (ค่าดิบค้นไม่เจอ, ค่า hash เจอ 1 แถว)  [PASS]

```
rows=1 raw=0 hash=1
```

## ผ่าน Kong ไม่มี token -> 401 (ตอบโดย Kong ไม่ถึง API)  [PASS]

```
{"message":"Unauthorized"}
```

## ผ่าน Kong ด้วย access token ของ alice -> 200 เห็นเฉพาะของตัวเอง  [PASS]

```
code=200 items=3 ของคนอื่น=0
```

## ผ่าน Kong payload ถูกแก้ -> 401 จาก Kong  [PASS]

```
{"message":"Bad token; invalid JSON"}
```

## ผ่าน Kong alg=none -> 401 จาก Kong  [PASS]

```
{"message":"Bad token; invalid alg"}
```

## Kong ไม่รับ token ใน query string (ใช้ token ที่ออกใหม่ ไม่ใช่ของ session)  [PASS]

```
code=401
```

## route ที่ Kong ไม่ได้เปิด (/metrics) ได้ 404 ส่วน API ตรงตอบได้เฉพาะผ่าน 127.0.0.1  [PASS]

```
kong /metrics=404 api direct /healthz=200
```

## refresh ไม่มี header X-Requested-With -> 403 (กัน CSRF)  [PASS]

```
{"error":"forbidden"}

```

## refresh -> 200 access token ใหม่ (sub=alice) และ cookie ใหม่ไม่ซ้ำของเดิม (หมุนเวียน)  [PASS]

```
code=200 sub=alice cookie ใหม่ต่างจากเดิม=True body={"access_token":"[jwt]","token_type":"Bearer","expires_in":600}

```

## เทียบ: refresh ยิงตรงเข้า API (ไม่ผ่าน Kong) หมุน token ได้ -> 200  [PASS]

```
login=200 refresh=200 body={"access_token":"[jwt]","token_type":"Bearer","expires_in":600}

```

## ใช้ refresh token เดิมซ้ำ -> 401 (ตรวจพบการใช้ซ้ำ)  [PASS]

```
code=401 body={"error":"unauthorized"}

```

## หลังตรวจพบการใช้ซ้ำ token ใหม่ของสายเดียวกันถูกเพิกถอนด้วย -> 401  [PASS]

```
code=401
```

## DB: แถว refresh ของ alice ที่ถูกเพิกถอนมีอย่างน้อย 1 แถว  [PASS]

```
revoked|total|used = 2|4|2
```

## refresh ไม่มี cookie -> 401  [PASS]

```
code=401
```

## refresh ของ bob ได้ token ในนาม bob เท่านั้น (sub มาจากแถวใน DB ไม่ใช่จาก client)  [PASS]

```
sub=bob
```

## logout -> 204 ล้าง cookie และ refresh token นั้นใช้ต่อไม่ได้ (401)  [PASS]

```
logout=204 refresh_after=401
```

## CORS ของ /api/auth: origin ของ UI ได้ credentials=true, origin อื่นไม่ได้ header allow  [PASS]

```
ui: origin=http://localhost:3000 cred=true  evil: origin=http://localhost:3000
```

## TLS ที่ Kong: cert ตรวจผ่านเมื่อ pin cert ของเครื่อง และตอบ 401 เมื่อไม่มี token  [PASS]

```
401 verify=0
```

## TLS: ถ้าไม่ pin cert (self-signed) curl ปฏิเสธการเชื่อมต่อ (http_code 000)  [PASS]

```
000 verify=0 (exit 60)
```

## UI (nginx) เสิร์ฟหน้า SPA และมี security header  [PASS]

```
code=200 x-frame=DENY
```

## rate limit ที่ /api/auth/login: ยิงผิด 30 ครั้งติดกัน -> เจอ 429 พร้อมโควต้าใน header  [PASS]

```
401=27 429=3 header-remaining=0 retry-after=58
```

## log ของ Kong และ API ไม่มี refresh token, access token หรือรหัสผ่าน demo  [PASS]

```
สะอาด
```

## log API มี security_event: login_failed, refresh_reuse, csrf_header_missing  [PASS]

```
login_failed=29 refresh_reuse=1 csrf_header_missing=1 refresh_failed=3
```


