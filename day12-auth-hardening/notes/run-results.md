# Day 12 run results

Run at: 2026-11-05 19:04:08  Machine: LAPTOP-MICOTO

## Summary

101 / 101 steps passed

| Step | Result |
|---|---|
| Docker running | PASS |
| .env มีรหัส Grafana และรหัส demo user (ไม่มี JWT secret แล้ว), ไฟล์ถูก git-ignore | PASS |
| สร้างคู่คีย์ RSA k1 ด้วย cmd/keygen (k1.key + k1.pub) และโฟลเดอร์ keys ถูก git-ignore | PASS |
| keygen ไม่ยอมเขียนทับ kid เดิม (exit != 0) | PASS |
| go test ./... (unit + baseline snapshot) | PASS |
| go vet ./... | PASS |
| เทสต์ static (compose, Prometheus, dashboard, migration, secrets) | PASS |
| render config ของ Kong (ใส่ public key ทุก kid) และสร้าง cert สำหรับ localhost | PASS |
| docker compose up --build (api, worker, kong, web, postgres, valkey, kafka, prometheus, grafana, jaeger) | PASS |
| api ready (GET /readyz 200) และ migration 000002 ผ่าน | PASS |
| ตาราง claims มีคอลัมน์ owner_id (NOT NULL) | PASS |
| migration 000003: ตาราง users และ refresh_tokens มีอยู่ | PASS |
| Kong และ UI ขึ้น (UI /healthz 200, Kong ตอบ 401 เมื่อไม่มี token) | PASS |
| cmd/mint ออก token ได้ (alice, bob, ops=staff) | PASS |
| ไม่มีไดเรกทอรีคีย์ -> mint ปฏิเสธ (exit != 0) ไม่ออก token | PASS |
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
| log ของ api/worker ไม่มี token, private key หรือ patient_ref | PASS |
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
| access token เป็น RS256 มี kid=k1 (ทั้ง token จาก mint และจาก login ผ่าน Kong) | PASS |
| token RS256 ที่ถูกต้องผ่าน Kong -> 200 (ใช้เทียบกับ token ปลอมด้านล่าง) | PASS |
| โจมตี key confusion (HS256 เซ็นด้วย public key เป็น secret, อ้าง role=staff) -> 401 ทั้งที่ Kong และ API ตรง | PASS |
| token ที่อ้าง kid ที่ไม่รู้จัก (k9) -> 401 ทั้งที่ Kong และ API ตรง | PASS |
| Kong access log ใช้ format noquery: ยังมีบรรทัด request (path) แต่ไม่มี query string เลย (canary และ token ที่ส่งทาง ?jwt= ก่อนหน้าไม่อยู่ใน log) | PASS |
| 429 ที่ /api/auth ส่ง Retry-After และ Access-Control-Expose-Headers ให้ origin ของ UI อ่านได้ | PASS |
| หมุน key ขั้น 1: เพิ่ม k2, JWT_SIGNING_KID=k2, render Kong ใหม่ (รู้ทั้ง k1 และ k2) และยก api+gateway ใหม่ | PASS |
| ระหว่างหมุน key: token เดิมที่ออกด้วย k1 ยังใช้ผ่าน Kong ได้ (200) | PASS |
| login หลังหมุน key ได้ access token ที่ kid=k2 และใช้ผ่าน Kong ได้ (200) | PASS |
| log ของ API หลัง restart บอก signing_kid=k2 และ accepted_kids มี k1 กับ k2 (ไม่มีเนื้อคีย์) | PASS |
| หมุน key ขั้น 2: ถอด k1 -> token เดิมของ k1 ได้ 401 ทั้ง Kong และ API ตรง ส่วน token ของ k2 ยังได้ 200 | PASS |
| หลังถอด k1 login ยังใช้ได้ (refresh token ใน DB ไม่ผูกกับ key) และได้ token kid=k2 | PASS |
| log ทั้ง Kong และ API ไม่มีเนื้อ private key ของ k1 และ k2 | PASS |

## Docker running  [PASS]

```
client 29.7.2 / server 29.7.2
```

## .env มีรหัส Grafana และรหัส demo user (ไม่มี JWT secret แล้ว), ไฟล์ถูก git-ignore  [PASS]

```
รหัส Grafana length=24 รหัส demo length=24 (ค่าไม่แสดง)
```

## สร้างคู่คีย์ RSA k1 ด้วย cmd/keygen (k1.key + k1.pub) และโฟลเดอร์ keys ถูก git-ignore  [PASS]

```
created /keys/k1.key and /keys/k1.pub (ไม่แสดงเนื้อคีย์)
```

## keygen ไม่ยอมเขียนทับ kid เดิม (exit != 0)  [PASS]

```
error: /keys/k1.key already exists; pick a new kid instead of overwriting
```

## go test ./... (unit + baseline snapshot)  [PASS]

```
?   	github.com/sitthichok-J/self-study-2026/day12-auth-hardening/cmd/api	[no test files]
?   	github.com/sitthichok-J/self-study-2026/day12-auth-hardening/cmd/keygen	[no test files]
?   	github.com/sitthichok-J/self-study-2026/day12-auth-hardening/cmd/mint	[no test files]
?   	github.com/sitthichok-J/self-study-2026/day12-auth-hardening/cmd/worker	[no test files]
ok  	github.com/sitthichok-J/self-study-2026/day12-auth-hardening/internal/adapter/http	(cached)
?   	github.com/sitthichok-J/self-study-2026/day12-auth-hardening/internal/adapter/kafka	[no test files]
?   	github.com/sitthichok-J/self-study-2026/day12-auth-hardening/internal/adapter/postgres	[no test files]
?   	github.com/sitthichok-J/self-study-2026/day12-auth-hardening/internal/adapter/redis	[no test files]
ok  	github.com/sitthichok-J/self-study-2026/day12-auth-hardening/internal/auth	(cached)
?   	github.com/sitthichok-J/self-study-2026/day12-auth-hardening/internal/auth/authtest	[no test files]
ok  	github.com/sitthichok-J/self-study-2026/day12-auth-hardening/internal/baseline	(cached)
ok  	github.com/sitthichok-J/self-study-2026/day12-auth-hardening/internal/config	(cached)
ok  	github.com/sitthichok-J/self-study-2026/day12-auth-hardening/internal/domain	(cached)
?   	github.com/sitthichok-J/self-study-2026/day12-auth-hardening/internal/memory	[no test files]
ok  	github.com/sitthichok-J/self-study-2026/day12-auth-hardening/internal/obs	(cached)
?   	github.com/sitthichok-J/self-study-2026/day12-auth-hardening/internal/port	[no test files]
ok  	github.com/sitthichok-J/self-study-2026/day12-auth-hardening/internal/service	(cached)
```

## go vet ./...  [PASS]

```

```

## เทสต์ static (compose, Prometheus, dashboard, migration, secrets)  [PASS]

```
... (48 earlier lines omitted)
test_no_hardcoded_secret_assignments_outside_tests (test_infra.TestNoSecretsInRepo.test_no_hardcoded_secret_assignments_outside_tests) ... ok
test_no_private_keys_cloud_keys_or_real_looking_tokens (test_infra.TestNoSecretsInRepo.test_no_private_keys_cloud_keys_or_real_looking_tokens) ... ok
test_alert_metrics_and_labels_exist_in_go (test_infra.TestPrometheus.test_alert_metrics_and_labels_exist_in_go) ... ok
test_alert_rules_are_well_formed (test_infra.TestPrometheus.test_alert_rules_are_well_formed) ... ok
test_rules_file_is_mounted_where_prometheus_expects_it (test_infra.TestPrometheus.test_rules_file_is_mounted_where_prometheus_expects_it) ... ok
test_scrape_targets_are_the_internal_metrics_ports (test_infra.TestPrometheus.test_scrape_targets_are_the_internal_metrics_ports) ... ok
test_the_required_alerts_exist (test_infra.TestPrometheus.test_the_required_alerts_exist) ... ok

----------------------------------------------------------------------
Ran 51 tests in 2.443s

OK
```

## render config ของ Kong (ใส่ public key ทุก kid) และสร้าง cert สำหรับ localhost  [PASS]

```
rendered gateway/.rendered/kong.yml with kids: k1
```

## docker compose up --build (api, worker, kong, web, postgres, valkey, kafka, prometheus, grafana, jaeger)  [PASS]

```
... (213 earlier lines omitted)
 Container day12-valkey-1 Waiting 
 Container day12-migrate-1 Waiting 
 Container day12-migrate-1 Waiting 
 Container day12-kafka-init-1 Waiting 
 Container day12-kafka-init-1 Waiting 
 Container day12-valkey-1 Healthy 
 Container day12-valkey-1 Healthy 
 Container day12-migrate-1 Exited 
 Container day12-migrate-1 Exited 
 Container day12-kafka-init-1 Exited 
 Container day12-api-1 Starting 
 Container day12-kafka-init-1 Exited 
 Container day12-worker-1 Starting 
 Container day12-worker-1 Started 
 Container day12-api-1 Started 
 Container day12-prometheus-1 Starting 
 Container day12-gateway-1 Starting 
 Container day12-prometheus-1 Started 
 Container day12-grafana-1 Starting 
 Container day12-gateway-1 Started 
 Container day12-gateway-1 Waiting 
 Container day12-grafana-1 Started 
 Container day12-gateway-1 Healthy 
 Container day12-web-1 Starting 
 Container day12-web-1 Started
```

## api ready (GET /readyz 200) และ migration 000002 ผ่าน  [PASS]

```
NAME                 IMAGE                             COMMAND                  SERVICE      CREATED          STATUS                            PORTS
day12-api-1          day12-api                         "/app"                   api          25 seconds ago   Up 14 seconds                     127.0.0.1:8080->8080/tcp
day12-gateway-1      kong:3.9.1                        "/docker-entrypoint.…"   gateway      25 seconds ago   Up 13 seconds (healthy)           127.0.0.1:8000->8000/tcp, 127.0.0.1:8443->8443/tcp
day12-grafana-1      grafana/grafana:11.3.0            "/run.sh"                grafana      25 seconds ago   Up 13 seconds                     127.0.0.1:3001->3000/tcp
day12-jaeger-1       jaegertracing/all-in-one:1.62.0   "/go/bin/all-in-one-…"   jaeger       25 seconds ago   Up 24 seconds                     127.0.0.1:16686->16686/tcp
day12-kafka-1        apache/kafka:3.8.0                "/__cacert_entrypoin…"   kafka        25 seconds ago   Up 24 seconds (healthy)           127.0.0.1:9094->9094/tcp
day12-kafka-init-1   apache/kafka:3.8.0                "/bin/sh -c '/opt/ka…"   kafka-init   25 seconds ago   Exited (0) 14 seconds ago         
day12-migrate-1      migrate/migrate:v4.18.1           "migrate -path=/migr…"   migrate      25 seconds ago   Exited (0) 20 seconds ago         
day12-postgres-1     postgres:16-alpine                "docker-entrypoint.s…"   postgres     25 seconds ago   Up 24 seconds (healthy)           127.0.0.1:5432->5432/tcp
day12-prometheus-1   prom/prometheus:v2.55.1           "/bin/prometheus --c…"   prometheus   25 seconds ago   Up 13 seconds                     127.0.0.1:9091->9090/tcp
day12-valkey-1       valkey/valkey:8-alpine            "tini -- docker-entr…"   valkey       25 seconds ago   Up 24 seconds (healthy)           127.0.0.1:6379->6379/tcp
day12-web-1          day12-web                         "/docker-entrypoint.…"   web          25 seconds ago   Up 2 seconds (health: starting)   127.0.0.1:3000->8080/tcp
day12-worker-1       day12-worker                      "/app"                   worker       25 seconds ago   Up 14 seconds
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

## ไม่มีไดเรกทอรีคีย์ -> mint ปฏิเสธ (exit != 0) ไม่ออก token  [PASS]

```
error: signing key: open /nonexistent/k1.key: no such file or directory
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
{"id":"0c1383db-d175-4e6a-ac99-129afa2a044e","owner_id":"alice","patient_ref":"FAKE-AB12","amount":150000,"status":"PENDING","created_at":"2026-11-05T12:05:12.99053Z","updated_at":"2026-11-05T12:05:12.99053Z"}

```

## alice ส่ง key เดิมซ้ำ -> 200 id เดิม  [PASS]

```
{"id":"0c1383db-d175-4e6a-ac99-129afa2a044e","owner_id":"alice","patient_ref":"FAKE-AB12","amount":150000,"status":"PENDING","created_at":"2026-11-05T12:05:12.99053Z","updated_at":"2026-11-05T12:05:12.99053Z"}

```

## bob ใช้ key เดียวกับ alice (body ต่าง) -> 201 เคลมใหม่ของ bob ไม่ใช่ 409  [PASS]

```
{"id":"d8311156-cc2f-4867-bfab-6bfa5895ede0","owner_id":"bob","patient_ref":"FAKE-ZZ99","amount":777,"status":"PENDING","created_at":"2026-11-05T12:05:13.03925Z","updated_at":"2026-11-05T12:05:13.03925Z"}

```

## IDOR: bob เปิดเคลมของ alice -> 404 เหมือนเคลมที่ไม่มีอยู่ และไม่มีข้อมูลหลุด  [PASS]

```
code=404 body={"error":"claim not found"}

```

## alice เปิดเคลมของตัวเอง -> 200  [PASS]

```
{"id":"0c1383db-d175-4e6a-ac99-129afa2a044e","owner_id":"alice","patient_ref":"FAKE-AB12","amount":150000,"status":"PENDING","created_at":"2026-11-05T12:05:12.99053Z","updated_at":"2026-11-05T12:05:12.99053Z"}

```

## ops (staff) เปิดเคลมของ alice -> 200  [PASS]

```
{"id":"0c1383db-d175-4e6a-ac99-129afa2a044e","owner_id":"alice","patient_ref":"FAKE-AB12","amount":150000,"status":"PENDING","created_at":"2026-11-05T12:05:12.99053Z","updated_at":"2026-11-05T12:05:12.99053Z"}

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
50f63e4d6655091819544d1d4a74fffb
```

## port ที่ publish ผูก 127.0.0.1 ทั้งหมด  [PASS]

```
day12-api-1 127.0.0.1:8080->8080/tcp
day12-gateway-1 127.0.0.1:8000->8000/tcp, 127.0.0.1:8443->8443/tcp
day12-grafana-1 127.0.0.1:3001->3000/tcp
day12-jaeger-1 127.0.0.1:16686->16686/tcp
day12-kafka-1 127.0.0.1:9094->9094/tcp
day12-postgres-1 127.0.0.1:5432->5432/tcp
day12-prometheus-1 127.0.0.1:9091->9090/tcp
day12-valkey-1 127.0.0.1:6379->6379/tcp
day12-web-1 127.0.0.1:3000->8080/tcp
day12-worker-1
```

## worker ตัดสินเป็น APPROVED (ผ่านเส้นทางเดิม outbox -> Kafka -> worker)  [PASS]

```
{"id":"0c1383db-d175-4e6a-ac99-129afa2a044e","owner_id":"alice","patient_ref":"FAKE-AB12","amount":150000,"status":"APPROVED","decision_reason":"ok","created_at":"2026-11-05T12:05:12.99053Z","updated_at":"2026-11-05T12:05:14.030988Z"}

```

## เกิน 50,000 บาท -> REJECTED over_limit (กฎของ Day 8 ยังอยู่)  [PASS]

```
{"id":"ded36520-7984-4dc9-9d37-5262ae0a4f92","owner_id":"alice","patient_ref":"FAKE-OVER1","amount":600000000,"status":"REJECTED","decision_reason":"over_limit","created_at":"2026-11-05T12:05:17.663872Z","updated_at":"2026-11-05T12:05:19.005271Z"}

```

## ส่งเคลม poison (ต้องล้มทุกครั้ง) -> 201  [PASS]

```
{"id":"27a66b9e-7b81-413f-8cb0-af275a61d787","owner_id":"alice","patient_ref":"FAKE-POISON","amount":100,"status":"PENDING","created_at":"2026-11-05T12:05:19.71762Z","updated_at":"2026-11-05T12:05:19.71762Z"}

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
traces=7
```

## trace เดียวผ่านทั้ง claims-api และ claims-worker (trace context ข้าม Kafka ผ่าน payload)  [PASS]

```
repo.Get | consume claims.submitted | repo.UpdateStatus | outbox publish claims.submitted | POST /claims | repo.Create
```

## trace นั้นมี span ของ repo.Create, outbox publish และ consume  [PASS]

```
repo.Get | consume claims.submitted | repo.UpdateStatus | outbox publish claims.submitted | POST /claims | repo.Create
```

## log ของ request เป็น JSON มี trace_id, span_id, request_id  [PASS]

```
api-1     | {"time":"2026-11-05T12:05:12.994214434Z","level":"INFO","msg":"request","method":"POST","route":"/claims","status":201,"duration_ms":3.922,"bytes":210,"trace_id":"82b3a039454bdb58cd16856f6d2c6079","span_id":"0d48e6ea975eb09b","request_id":"6f3b5a6e87261f55f7d4046b3f0a63dd"}
```

## trace_id ในบรรทัด log เปิดหา trace เดียวกันใน Jaeger ได้  [PASS]

```
trace_id=82b3a039454bdb58cd16856f6d2c6079
```

## log มี security_event ทั้ง auth_failed และ access_denied  [PASS]

```
... (1 earlier lines omitted)
api-1     | {"time":"2026-11-05T12:05:12.915478046Z","level":"WARN","msg":"authentication failed","security_event":"auth_failed","reason":"signature","trace_id":"301b62d951309ef7dc9ba4a5d4c90fc6","span_id":"5ae3a8f8958181ae","request_id":"6988757292c8d751259f43708f937b71"}
api-1     | {"time":"2026-11-05T12:05:12.924850614Z","level":"WARN","msg":"authentication failed","security_event":"auth_failed","reason":"alg","trace_id":"da2b24a3bbfab1bc65a37db71894bdb4","span_id":"cb55b29b24a22aaf","request_id":"bd9d2be808da7387f83b9ab6258517ca"}
api-1     | {"time":"2026-11-05T12:05:12.943955715Z","level":"WARN","msg":"authentication failed","security_event":"auth_failed","reason":"missing","trace_id":"20b0f076ceadf9be02cf1fd5de7781b3","span_id":"32593563cd50948e","request_id":"c2f7027151e0a9d5b2c53d386c0d1acd"}
api-1     | {"time":"2026-11-05T12:05:13.064008852Z","level":"WARN","msg":"access denied to claim","security_event":"access_denied","sub":"bob","claim_id":"0c1383db-d175-4e6a-ac99-129afa2a044e","trace_id":"70b42ff07fe11045b1f6294af024f10d","span_id":"e2d25ee6d5031cc6","request_id":"1493d63aa0bacc7fb3f94b435aa552c2"}
```

## log ของ api/worker ไม่มี token, private key หรือ patient_ref  [PASS]

```
สะอาด
```

## api รันเป็น non-root  [PASS]

```
SERVICE  #   UID    PID   PPID  C   STIME  TTY  TIME      CMD
api      1   65532  3810  3785  1   12:04  ?    00:00:00  /app
```

## login ผิด: user มีอยู่กับไม่มีอยู่ ตอบ 401 ข้อความเดียวกันทุกตัวอักษร  [PASS]

```
code=401/401 body={"error":"invalid username or password"}
 เวลา 4087ms vs 108ms (ตัวอย่างเดียว ไม่ใช่การพิสูจน์เรื่อง timing)
```

## login ผ่าน Kong -> 200 access token (sub=alice) + refresh cookie HttpOnly; Secure; SameSite=Strict; Path=/api/auth  [PASS]

```
code=200 sub=alice expires_in=600 cookie(ไม่แสดงค่า)=refresh_token=[value]; Path=/api/auth; Expires=Fri, 06 Nov 2026 00:05:49 GMT; Max-Age=43200; HttpOnly; Secure; SameSite=Strict
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
401=8 429=22 header-remaining=0 retry-after=7
```

## log ของ Kong และ API ไม่มี refresh token, access token หรือรหัสผ่าน demo  [PASS]

```
สะอาด
```

## log API มี security_event: login_failed, refresh_reuse, csrf_header_missing  [PASS]

```
login_failed=10 refresh_reuse=1 csrf_header_missing=1 refresh_failed=3
```

## access token เป็น RS256 มี kid=k1 (ทั้ง token จาก mint และจาก login ผ่าน Kong)  [PASS]

```
mint: alg=RS256 kid=k1 | login: alg=RS256 kid=k1
```

## token RS256 ที่ถูกต้องผ่าน Kong -> 200 (ใช้เทียบกับ token ปลอมด้านล่าง)  [PASS]

```
code=200
```

## โจมตี key confusion (HS256 เซ็นด้วย public key เป็น secret, อ้าง role=staff) -> 401 ทั้งที่ Kong และ API ตรง  [PASS]

```
kong=401 api=401
```

## token ที่อ้าง kid ที่ไม่รู้จัก (k9) -> 401 ทั้งที่ Kong และ API ตรง  [PASS]

```
kong=401 api=401
```

## Kong access log ใช้ format noquery: ยังมีบรรทัด request (path) แต่ไม่มี query string เลย (canary และ token ที่ส่งทาง ?jwt= ก่อนหน้าไม่อยู่ใน log)  [PASS]

```
มีบรรทัด request=True พบ query/canary/token ใน log=False | ตัวอย่าง: gateway-1  | 172.18.0.1 [05/Nov/2026:12:05:07 +0000] "GET /api/claims HTTP/1.1" 401 26 "Mozilla/5.0 (Windows NT; Windows NT 10.0; en-US) WindowsPowerShell/5.1.26100.9711" || gateway-1  | 172.18.0.1 [05/Nov/2026:12:05:50 +0000] "GET /api/claims HTTP/1.1" 401 26 "Mozilla/5.0 (Windows NT; Windows NT 10.0; en-US) WindowsPowerShell/5.1.26100.9711"
```

## 429 ที่ /api/auth ส่ง Retry-After และ Access-Control-Expose-Headers ให้ origin ของ UI อ่านได้  [PASS]

```
code=429 Retry-After=3 Expose=Retry-After,X-RateLimit-Remaining-Minute Allow-Origin=http://localhost:3000 body={
  "message":"Too many authentication attempts",
  "request_id":"3e73d73acd55ac1b5670654b4dea4660"
}
```

## หมุน key ขั้น 1: เพิ่ม k2, JWT_SIGNING_KID=k2, render Kong ใหม่ (รู้ทั้ง k1 และ k2) และยก api+gateway ใหม่  [PASS]

```
created /keys/k2.key and /keys/k2.pub
... (1 earlier lines omitted)
 Container day12-api-1 Recreate 
 Container day12-api-1 Recreated 
 Container day12-gateway-1 Recreate 
 Container day12-gateway-1 Recreated 
 Container day12-api-1 Starting 
 Container day12-api-1 Started 
 Container day12-gateway-1 Starting 
 Container day12-gateway-1 Started
```

## ระหว่างหมุน key: token เดิมที่ออกด้วย k1 ยังใช้ผ่าน Kong ได้ (200)  [PASS]

```
code=200
```

## login หลังหมุน key ได้ access token ที่ kid=k2 และใช้ผ่าน Kong ได้ (200)  [PASS]

```
login=200 kid=k2 claims=200
```

## log ของ API หลัง restart บอก signing_kid=k2 และ accepted_kids มี k1 กับ k2 (ไม่มีเนื้อคีย์)  [PASS]

```
api-1  | {"time":"2026-11-05T12:06:02.379648923Z","level":"INFO","msg":"jwt keys loaded","signing_kid":"k2","accepted_kids":["k1","k2"]}
```

## หมุน key ขั้น 2: ถอด k1 -> token เดิมของ k1 ได้ 401 ทั้ง Kong และ API ตรง ส่วน token ของ k2 ยังได้ 200  [PASS]

```
k1 ผ่าน Kong=401 k1 ตรง API=401 k2 ผ่าน Kong=200
... (1 earlier lines omitted)
 Container day12-api-1 Recreate 
 Container day12-api-1 Recreated 
 Container day12-gateway-1 Recreate 
 Container day12-gateway-1 Recreated 
 Container day12-api-1 Starting 
 Container day12-api-1 Started 
 Container day12-gateway-1 Starting 
 Container day12-gateway-1 Started
```

## หลังถอด k1 login ยังใช้ได้ (refresh token ใน DB ไม่ผูกกับ key) และได้ token kid=k2  [PASS]

```
login=200 kid=k2
```

## log ทั้ง Kong และ API ไม่มีเนื้อ private key ของ k1 และ k2  [PASS]

```
สะอาด
```


