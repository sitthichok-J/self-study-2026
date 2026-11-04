# Day 10 run results

Run at: 2026-11-04 00:08:38  Machine: LAPTOP-MICOTO

## Summary

57 / 57 steps passed

| Step | Result |
|---|---|
| Docker running | PASS |
| .env มี JWT_SECRET (>=32 ตัวอักษร) และรหัส Grafana, ไฟล์ถูก git-ignore | PASS |
| go test ./... (unit + baseline snapshot) | PASS |
| go vet ./... | PASS |
| เทสต์ static (compose, Prometheus, dashboard, migration, secrets) | PASS |
| docker compose up --build (api, worker, postgres, valkey, kafka, prometheus, grafana, jaeger) | PASS |
| api ready (GET /readyz 200) และ migration 000002 ผ่าน | PASS |
| ตาราง claims มีคอลัมน์ owner_id (NOT NULL) | PASS |
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

## Docker running  [PASS]

```
client 29.7.2 / server 29.7.2
```

## .env มี JWT_SECRET (>=32 ตัวอักษร) และรหัส Grafana, ไฟล์ถูก git-ignore  [PASS]

```
JWT_SECRET length=64 (ค่าไม่แสดง)
```

## go test ./... (unit + baseline snapshot)  [PASS]

```
?   	github.com/sitthichok-J/self-study-2026/day10-observability-security/cmd/api	[no test files]
?   	github.com/sitthichok-J/self-study-2026/day10-observability-security/cmd/mint	[no test files]
?   	github.com/sitthichok-J/self-study-2026/day10-observability-security/cmd/worker	[no test files]
ok  	github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/adapter/http	(cached)
?   	github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/adapter/kafka	[no test files]
?   	github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/adapter/postgres	[no test files]
?   	github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/adapter/redis	[no test files]
ok  	github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/auth	(cached)
ok  	github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/baseline	(cached)
ok  	github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/config	(cached)
ok  	github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/domain	(cached)
?   	github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/memory	[no test files]
ok  	github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/obs	(cached)
?   	github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/port	[no test files]
ok  	github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/service	(cached)
```

## go vet ./...  [PASS]

```

```

## เทสต์ static (compose, Prometheus, dashboard, migration, secrets)  [PASS]

```
... (30 earlier lines omitted)
test_no_hardcoded_secret_assignments_outside_tests (test_infra.TestNoSecretsInRepo.test_no_hardcoded_secret_assignments_outside_tests) ... ok
test_no_private_keys_cloud_keys_or_real_looking_tokens (test_infra.TestNoSecretsInRepo.test_no_private_keys_cloud_keys_or_real_looking_tokens) ... ok
test_alert_metrics_and_labels_exist_in_go (test_infra.TestPrometheus.test_alert_metrics_and_labels_exist_in_go) ... ok
test_alert_rules_are_well_formed (test_infra.TestPrometheus.test_alert_rules_are_well_formed) ... ok
test_rules_file_is_mounted_where_prometheus_expects_it (test_infra.TestPrometheus.test_rules_file_is_mounted_where_prometheus_expects_it) ... ok
test_scrape_targets_are_the_internal_metrics_ports (test_infra.TestPrometheus.test_scrape_targets_are_the_internal_metrics_ports) ... ok
test_the_required_alerts_exist (test_infra.TestPrometheus.test_the_required_alerts_exist) ... ok

----------------------------------------------------------------------
Ran 33 tests in 1.004s

OK
```

## docker compose up --build (api, worker, postgres, valkey, kafka, prometheus, grafana, jaeger)  [PASS]

```
... (147 earlier lines omitted)
 Container day10-migrate-1 Starting 
 Container day10-migrate-1 Started 
 Container day10-kafka-1 Healthy 
 Container day10-kafka-init-1 Starting 
 Container day10-kafka-init-1 Started 
 Container day10-migrate-1 Waiting 
 Container day10-valkey-1 Waiting 
 Container day10-kafka-init-1 Waiting 
 Container day10-valkey-1 Waiting 
 Container day10-migrate-1 Waiting 
 Container day10-kafka-init-1 Waiting 
 Container day10-migrate-1 Exited 
 Container day10-migrate-1 Exited 
 Container day10-valkey-1 Healthy 
 Container day10-valkey-1 Healthy 
 Container day10-kafka-init-1 Exited 
 Container day10-api-1 Starting 
 Container day10-kafka-init-1 Exited 
 Container day10-worker-1 Starting 
 Container day10-worker-1 Started 
 Container day10-api-1 Started 
 Container day10-prometheus-1 Starting 
 Container day10-prometheus-1 Started 
 Container day10-grafana-1 Starting 
 Container day10-grafana-1 Started
```

## api ready (GET /readyz 200) และ migration 000002 ผ่าน  [PASS]

```
NAME                 IMAGE                             COMMAND                  SERVICE      CREATED          STATUS                     PORTS
day10-api-1          day10-api                         "/app"                   api          15 seconds ago   Up 3 seconds               127.0.0.1:8080->8080/tcp
day10-grafana-1      grafana/grafana:11.3.0            "/run.sh"                grafana      14 seconds ago   Up 2 seconds               127.0.0.1:3001->3000/tcp
day10-jaeger-1       jaegertracing/all-in-one:1.62.0   "/go/bin/all-in-one-…"   jaeger       15 seconds ago   Up 14 seconds              127.0.0.1:16686->16686/tcp
day10-kafka-1        apache/kafka:3.8.0                "/__cacert_entrypoin…"   kafka        15 seconds ago   Up 14 seconds (healthy)    127.0.0.1:9094->9094/tcp
day10-kafka-init-1   apache/kafka:3.8.0                "/bin/sh -c '/opt/ka…"   kafka-init   15 seconds ago   Exited (0) 3 seconds ago   
day10-migrate-1      migrate/migrate:v4.18.1           "migrate -path=/migr…"   migrate      15 seconds ago   Exited (0) 9 seconds ago   
day10-postgres-1     postgres:16-alpine                "docker-entrypoint.s…"   postgres     15 seconds ago   Up 14 seconds (healthy)    127.0.0.1:5432->5432/tcp
day10-prometheus-1   prom/prometheus:v2.55.1           "/bin/prometheus --c…"   prometheus   15 seconds ago   Up 2 seconds               127.0.0.1:9091->9090/tcp
day10-valkey-1       valkey/valkey:8-alpine            "tini -- docker-entr…"   valkey       15 seconds ago   Up 14 seconds (healthy)    127.0.0.1:6379->6379/tcp
day10-worker-1       day10-worker                      "/app"                   worker       15 seconds ago   Up 3 seconds
```

## ตาราง claims มีคอลัมน์ owner_id (NOT NULL)  [PASS]

```
owner_id|NO
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
{"id":"72632f93-1962-411f-b830-ad7e1762bf12","owner_id":"alice","patient_ref":"FAKE-AB12","amount":150000,"status":"PENDING","created_at":"2026-11-03T17:10:05.637968Z","updated_at":"2026-11-03T17:10:05.637968Z"}

```

## alice ส่ง key เดิมซ้ำ -> 200 id เดิม  [PASS]

```
{"id":"72632f93-1962-411f-b830-ad7e1762bf12","owner_id":"alice","patient_ref":"FAKE-AB12","amount":150000,"status":"PENDING","created_at":"2026-11-03T17:10:05.637968Z","updated_at":"2026-11-03T17:10:05.637968Z"}

```

## bob ใช้ key เดียวกับ alice (body ต่าง) -> 201 เคลมใหม่ของ bob ไม่ใช่ 409  [PASS]

```
{"id":"7a3af547-f95d-47ff-bbfb-88ceb0be015c","owner_id":"bob","patient_ref":"FAKE-ZZ99","amount":777,"status":"PENDING","created_at":"2026-11-03T17:10:05.714299Z","updated_at":"2026-11-03T17:10:05.714299Z"}

```

## IDOR: bob เปิดเคลมของ alice -> 404 เหมือนเคลมที่ไม่มีอยู่ และไม่มีข้อมูลหลุด  [PASS]

```
code=404 body={"error":"claim not found"}

```

## alice เปิดเคลมของตัวเอง -> 200  [PASS]

```
{"id":"72632f93-1962-411f-b830-ad7e1762bf12","owner_id":"alice","patient_ref":"FAKE-AB12","amount":150000,"status":"PENDING","created_at":"2026-11-03T17:10:05.637968Z","updated_at":"2026-11-03T17:10:05.637968Z"}

```

## ops (staff) เปิดเคลมของ alice -> 200  [PASS]

```
{"id":"72632f93-1962-411f-b830-ad7e1762bf12","owner_id":"alice","patient_ref":"FAKE-AB12","amount":150000,"status":"PENDING","created_at":"2026-11-03T17:10:05.637968Z","updated_at":"2026-11-03T17:10:05.637968Z"}

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
8721c7e1ead0048f6dd8969d32a07312
```

## port ที่ publish ผูก 127.0.0.1 ทั้งหมด  [PASS]

```
day10-api-1 127.0.0.1:8080->8080/tcp
day10-grafana-1 127.0.0.1:3001->3000/tcp
day10-jaeger-1 127.0.0.1:16686->16686/tcp
day10-kafka-1 127.0.0.1:9094->9094/tcp
day10-postgres-1 127.0.0.1:5432->5432/tcp
day10-prometheus-1 127.0.0.1:9091->9090/tcp
day10-valkey-1 127.0.0.1:6379->6379/tcp
day10-worker-1
```

## worker ตัดสินเป็น APPROVED (ผ่านเส้นทางเดิม outbox -> Kafka -> worker)  [PASS]

```
{"id":"72632f93-1962-411f-b830-ad7e1762bf12","owner_id":"alice","patient_ref":"FAKE-AB12","amount":150000,"status":"APPROVED","decision_reason":"ok","created_at":"2026-11-03T17:10:05.637968Z","updated_at":"2026-11-03T17:10:06.942818Z"}

```

## เกิน 50,000 บาท -> REJECTED over_limit (กฎของ Day 8 ยังอยู่)  [PASS]

```
{"id":"f9920674-8ff9-485e-bda0-353f3a21cb24","owner_id":"alice","patient_ref":"FAKE-OVER1","amount":600000000,"status":"REJECTED","decision_reason":"over_limit","created_at":"2026-11-03T17:10:10.392899Z","updated_at":"2026-11-03T17:10:11.418363Z"}

```

## ส่งเคลม poison (ต้องล้มทุกครั้ง) -> 201  [PASS]

```
{"id":"d53f4434-a72d-4b0d-b8dd-655f0ab6210c","owner_id":"alice","patient_ref":"FAKE-POISON","amount":100,"status":"PENDING","created_at":"2026-11-03T17:10:12.444156Z","updated_at":"2026-11-03T17:10:12.444156Z"}

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
traces=14
```

## trace เดียวผ่านทั้ง claims-api และ claims-worker (trace context ข้าม Kafka ผ่าน payload)  [PASS]

```
repo.Get | consume claims.submitted | repo.UpdateStatus | repo.Create | POST /claims | outbox publish claims.submitted
```

## trace นั้นมี span ของ repo.Create, outbox publish และ consume  [PASS]

```
repo.Get | consume claims.submitted | repo.UpdateStatus | repo.Create | POST /claims | outbox publish claims.submitted
```

## log ของ request เป็น JSON มี trace_id, span_id, request_id  [PASS]

```
api-1     | {"time":"2026-11-03T17:10:05.644479399Z","level":"INFO","msg":"request","method":"POST","route":"/claims","status":201,"duration_ms":6.674,"bytes":212,"trace_id":"1e3876ecdc748be6aeaa6f504bc44dc7","span_id":"d4981edd8d2f7651","request_id":"3b15aff91eddd199dbd089a31705acd0"}
```

## trace_id ในบรรทัด log เปิดหา trace เดียวกันใน Jaeger ได้  [PASS]

```
trace_id=1e3876ecdc748be6aeaa6f504bc44dc7
```

## log มี security_event ทั้ง auth_failed และ access_denied  [PASS]

```
... (1 earlier lines omitted)
api-1     | {"time":"2026-11-03T17:10:05.559308012Z","level":"WARN","msg":"authentication failed","security_event":"auth_failed","reason":"signature","trace_id":"10bce466b501776b1b5ce0eafbc3832f","span_id":"4f5654b2613fc7ea","request_id":"065dfdb0a03bf2ff4b81dd292fef534b"}
api-1     | {"time":"2026-11-03T17:10:05.573825227Z","level":"WARN","msg":"authentication failed","security_event":"auth_failed","reason":"alg","trace_id":"a15c63c2bd815b50f5285b4c198d00cf","span_id":"675dae206743ae75","request_id":"4f48f88134053e5d3b94b763a86f2b7f"}
api-1     | {"time":"2026-11-03T17:10:05.594653802Z","level":"WARN","msg":"authentication failed","security_event":"auth_failed","reason":"missing","trace_id":"0ba0c132fda1e3fbb5cca9fff3bf021e","span_id":"4e6ffcc13f9812e3","request_id":"50cf3a9a69a46570c1ffef40dfb5eba6"}
api-1     | {"time":"2026-11-03T17:10:05.757181379Z","level":"WARN","msg":"access denied to claim","security_event":"access_denied","sub":"bob","claim_id":"72632f93-1962-411f-b830-ad7e1762bf12","trace_id":"d4d78c6abcb98db78ac44cf4f4082abf","span_id":"1975be34d6bbea23","request_id":"e82e2e97e3fb7f92c08c3f8e8d42cf13"}
```

## log ของ api/worker ไม่มี token, JWT_SECRET หรือ patient_ref  [PASS]

```
สะอาด
```

## api รันเป็น non-root  [PASS]

```
SERVICE  #   UID    PID    PPID   C   STIME  TTY  TIME      CMD
api      1   65532  21478  21453  0   17:09  ?    00:00:00  /app
```


