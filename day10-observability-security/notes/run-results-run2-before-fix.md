# Day 10 run results

Run at: 2026-11-03 23:32:35  Machine: LAPTOP-MICOTO

## Summary

55 / 57 steps passed

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
| alert ที่ควรดังตอนนี้: DeadLetterGrowing และ AccessDeniedSeen (จากที่ยิงทดสอบ) | FAIL |
| Grafana ขึ้น (GET /api/health) | PASS |
| dashboard claims-overview ถูก provision (>= 15 panels) | PASS |
| Grafana ไม่ยอมให้เข้าโดยไม่ login | PASS |
| datasource Prometheus + Jaeger ถูก provision | PASS |
| ทุก query ใน dashboard รันบน Prometheus ได้ (status success) | PASS |
| Jaeger มี trace ของ claims-api | PASS |
| trace เดียวผ่านทั้ง claims-api และ claims-worker (trace context ข้าม Kafka ผ่าน payload) | PASS |
| trace นั้นมี span ของ repo.Create, outbox publish และ consume | FAIL |
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
Ran 33 tests in 1.009s

OK
```

## docker compose up --build (api, worker, postgres, valkey, kafka, prometheus, grafana, jaeger)  [PASS]

```
... (131 earlier lines omitted)
 Container day10-migrate-1 Starting 
 Container day10-migrate-1 Started 
 Container day10-kafka-1 Healthy 
 Container day10-kafka-init-1 Starting 
 Container day10-kafka-init-1 Started 
 Container day10-migrate-1 Waiting 
 Container day10-kafka-init-1 Waiting 
 Container day10-valkey-1 Waiting 
 Container day10-migrate-1 Waiting 
 Container day10-kafka-init-1 Waiting 
 Container day10-valkey-1 Waiting 
 Container day10-migrate-1 Exited 
 Container day10-valkey-1 Healthy 
 Container day10-valkey-1 Healthy 
 Container day10-migrate-1 Exited 
 Container day10-kafka-init-1 Exited 
 Container day10-worker-1 Starting 
 Container day10-kafka-init-1 Exited 
 Container day10-api-1 Starting 
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
day10-api-1          day10-api                         "/app"                   api          14 seconds ago   Up 3 seconds               127.0.0.1:8080->8080/tcp
day10-grafana-1      grafana/grafana:11.3.0            "/run.sh"                grafana      14 seconds ago   Up 2 seconds               127.0.0.1:3001->3000/tcp
day10-jaeger-1       jaegertracing/all-in-one:1.62.0   "/go/bin/all-in-one-…"   jaeger       15 seconds ago   Up 13 seconds              127.0.0.1:16686->16686/tcp
day10-kafka-1        apache/kafka:3.8.0                "/__cacert_entrypoin…"   kafka        15 seconds ago   Up 13 seconds (healthy)    127.0.0.1:9094->9094/tcp
day10-kafka-init-1   apache/kafka:3.8.0                "/bin/sh -c '/opt/ka…"   kafka-init   14 seconds ago   Exited (0) 3 seconds ago   
day10-migrate-1      migrate/migrate:v4.18.1           "migrate -path=/migr…"   migrate      14 seconds ago   Exited (0) 9 seconds ago   
day10-postgres-1     postgres:16-alpine                "docker-entrypoint.s…"   postgres     15 seconds ago   Up 13 seconds (healthy)    127.0.0.1:5432->5432/tcp
day10-prometheus-1   prom/prometheus:v2.55.1           "/bin/prometheus --c…"   prometheus   14 seconds ago   Up 2 seconds               127.0.0.1:9091->9090/tcp
day10-valkey-1       valkey/valkey:8-alpine            "tini -- docker-entr…"   valkey       15 seconds ago   Up 13 seconds (healthy)    127.0.0.1:6379->6379/tcp
day10-worker-1       day10-worker                      "/app"                   worker       14 seconds ago   Up 3 seconds
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
{"id":"78bb47ef-e9b9-41db-a26d-229b095dd056","owner_id":"alice","patient_ref":"FAKE-AB12","amount":150000,"status":"PENDING","created_at":"2026-11-03T16:33:09.203534Z","updated_at":"2026-11-03T16:33:09.203534Z"}

```

## alice ส่ง key เดิมซ้ำ -> 200 id เดิม  [PASS]

```
{"id":"78bb47ef-e9b9-41db-a26d-229b095dd056","owner_id":"alice","patient_ref":"FAKE-AB12","amount":150000,"status":"PENDING","created_at":"2026-11-03T16:33:09.203534Z","updated_at":"2026-11-03T16:33:09.203534Z"}

```

## bob ใช้ key เดียวกับ alice (body ต่าง) -> 201 เคลมใหม่ของ bob ไม่ใช่ 409  [PASS]

```
{"id":"3833027a-4a0b-4512-8170-a66645a6c773","owner_id":"bob","patient_ref":"FAKE-ZZ99","amount":777,"status":"PENDING","created_at":"2026-11-03T16:33:09.259857Z","updated_at":"2026-11-03T16:33:09.259857Z"}

```

## IDOR: bob เปิดเคลมของ alice -> 404 เหมือนเคลมที่ไม่มีอยู่ และไม่มีข้อมูลหลุด  [PASS]

```
code=404 body={"error":"claim not found"}

```

## alice เปิดเคลมของตัวเอง -> 200  [PASS]

```
{"id":"78bb47ef-e9b9-41db-a26d-229b095dd056","owner_id":"alice","patient_ref":"FAKE-AB12","amount":150000,"status":"PENDING","created_at":"2026-11-03T16:33:09.203534Z","updated_at":"2026-11-03T16:33:09.203534Z"}

```

## ops (staff) เปิดเคลมของ alice -> 200  [PASS]

```
{"id":"78bb47ef-e9b9-41db-a26d-229b095dd056","owner_id":"alice","patient_ref":"FAKE-AB12","amount":150000,"status":"PENDING","created_at":"2026-11-03T16:33:09.203534Z","updated_at":"2026-11-03T16:33:09.203534Z"}

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
d8e035c3a8abd7d0e3e6e3c80c78a464
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
{"id":"78bb47ef-e9b9-41db-a26d-229b095dd056","owner_id":"alice","patient_ref":"FAKE-AB12","amount":150000,"status":"APPROVED","decision_reason":"ok","created_at":"2026-11-03T16:33:09.203534Z","updated_at":"2026-11-03T16:33:10.360129Z"}

```

## เกิน 50,000 บาท -> REJECTED over_limit (กฎของ Day 8 ยังอยู่)  [PASS]

```
{"id":"a226e3eb-596a-4d3b-8d15-d22caff4e4e4","owner_id":"alice","patient_ref":"FAKE-OVER1","amount":600000000,"status":"REJECTED","decision_reason":"over_limit","created_at":"2026-11-03T16:33:13.878887Z","updated_at":"2026-11-03T16:33:15.316Z"}

```

## ส่งเคลม poison (ต้องล้มทุกครั้ง) -> 201  [PASS]

```
{"id":"c48cca58-a4a9-4eeb-9e3d-bc24a3fb404c","owner_id":"alice","patient_ref":"FAKE-POISON","amount":100,"status":"PENDING","created_at":"2026-11-03T16:33:15.952951Z","updated_at":"2026-11-03T16:33:15.952951Z"}

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

## alert ที่ควรดังตอนนี้: DeadLetterGrowing และ AccessDeniedSeen (จากที่ยิงทดสอบ)  [FAIL]

```
firing: 
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
traces=13
```

## trace เดียวผ่านทั้ง claims-api และ claims-worker (trace context ข้าม Kafka ผ่าน payload)  [PASS]

```
repo.UpdateStatus | consume claims.submitted | repo.Get | repo.Create | POST /claims
```

## trace นั้นมี span ของ repo.Create, outbox publish และ consume  [FAIL]

```
repo.UpdateStatus | consume claims.submitted | repo.Get | repo.Create | POST /claims
```

## log ของ request เป็น JSON มี trace_id, span_id, request_id  [PASS]

```
api-1     | {"time":"2026-11-03T16:33:09.209688045Z","level":"INFO","msg":"request","method":"POST","route":"/claims","status":201,"duration_ms":6.341,"bytes":212,"trace_id":"eea89db14b4c98baf0ea2a1fdc5d627b","span_id":"8de10fe1996714ce","request_id":"d160ea18150bd7abb88eca7021a9c379"}
```

## trace_id ในบรรทัด log เปิดหา trace เดียวกันใน Jaeger ได้  [PASS]

```
trace_id=eea89db14b4c98baf0ea2a1fdc5d627b
```

## log มี security_event ทั้ง auth_failed และ access_denied  [PASS]

```
... (1 earlier lines omitted)
api-1     | {"time":"2026-11-03T16:33:09.122190924Z","level":"WARN","msg":"authentication failed","security_event":"auth_failed","reason":"signature","trace_id":"ab449a18a0e85b9c887c0b05df709a74","span_id":"1bf13d5cf3ef4149","request_id":"13ce8f06e6cf14e3e16ebd5951c4b506"}
api-1     | {"time":"2026-11-03T16:33:09.137603818Z","level":"WARN","msg":"authentication failed","security_event":"auth_failed","reason":"alg","trace_id":"b18026c1b03261e590e4def8fa56a346","span_id":"314e0bf712e45c78","request_id":"7ae2a1ece9e8cd5bd3772de75d10f535"}
api-1     | {"time":"2026-11-03T16:33:09.155297301Z","level":"WARN","msg":"authentication failed","security_event":"auth_failed","reason":"missing","trace_id":"8bcc9dd269955f8621e555d63f432ac8","span_id":"5aa45039de4c301d","request_id":"e46abbb525f93c9b4bfc1a880cb53807"}
api-1     | {"time":"2026-11-03T16:33:09.28702238Z","level":"WARN","msg":"access denied to claim","security_event":"access_denied","sub":"bob","claim_id":"78bb47ef-e9b9-41db-a26d-229b095dd056","trace_id":"5fe92ca21e3988611b5d37ec5bfa0d4b","span_id":"ad6c56cbc7fc6ead","request_id":"9e1d862a861a89e7bf15edafe8eb3de7"}
```

## log ของ api/worker ไม่มี token, JWT_SECRET หรือ patient_ref  [PASS]

```
สะอาด
```

## api รันเป็น non-root  [PASS]

```
SERVICE  #   UID    PID    PPID   C   STIME  TTY  TIME      CMD
api      1   65532  30599  30575  0   16:32  ?    00:00:00  /app
```


