# Day 13 run results

Run at: 2026-11-05 20:32:01  Machine: LAPTOP-MICOTO

## Summary

126 / 126 steps passed

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
| เทสต์ adapter กับ valkey จริง (บัญชีเพิกถอน token, cut ต่อ user, rate limit) ไม่ถูก skip | PASS |
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
| revocation 1: login ผ่าน Kong ได้ access token และใช้งานผ่าน Kong ได้ (200) | PASS |
| revocation 2: logout พร้อม Bearer ของ access token -> 204 และล้าง cookie | PASS |
| revocation 3: access token เดิม (ลายเซ็นและ exp ยังถูก) หลัง logout -> 401 ทันที ทั้งผ่าน Kong และยิงตรง API | PASS |
| revocation 4: login ใหม่ของ user เดียวกันได้ token ใหม่ที่ใช้งานได้ (เพิกถอนเฉพาะ token ที่ logout) | PASS |
| revocation 5: ใน valkey มีรายการ rev:jti:<jti> ที่มี TTL (หมดเองไม่เกินอายุ token + leeway) ไม่มีตัว token เก็บอยู่ | PASS |
| revocation 6: metric auth_failures_total{reason=revoked} นับ และ log API มี security_event token_revoked | PASS |
| revocation 7: ใช้ refresh token ซ้ำ -> 401 และ access token ทุกตัวของ bob ที่ออกก่อนหน้า (รวมตัวที่เพิ่ง refresh) ตายด้วย (401) | PASS |
| revocation 8: valkey มี rev:user:bob (เวลา cut เป็น unix) และ user อื่น (alice) ไม่ถูกกระทบ | PASS |
| revocation 9: bob login ใหม่หลังเหตุการณ์ได้ token ใหม่ที่ใช้งานได้ (cut ไม่ล็อก user ถาวร) | PASS |
| valkey ล่ม: คำขอที่มี token ถูกต้อง -> 503 + Retry-After (fail closed) ส่วน /healthz ยัง 200 และ token มั่ว ยัง 401 | PASS |
| valkey กลับมา: คำขอปกติ 200 อีกครั้ง และ token ที่ logout ไปก่อน restart ยัง 401 (blocklist อยู่รอดด้วย AOF) | PASS |
| ยก api ใหม่ด้วย USER_RATE_LIMIT=40 (เฉพาะรอบทดสอบ) | PASS |
| rate limit ต่อ user: เพดาน 40/นาที -> 40 ครั้งแรกผ่าน อีก 5 ครั้งโดน 429 พร้อม Retry-After | PASS |
| rate limit ต่อ user: user อื่นไม่โดนด้วย และ /healthz ไม่นับโควต้า | PASS |
| ตรวจเคลม 1: เคลม 30,000 บาท -> worker ตั้ง PENDING เหตุผล needs_review และไม่เปลี่ยนต่อ | PASS |
| ตรวจเคลม 2: เจ้าของเคลม (ไม่ใช่ staff) กดอนุมัติเอง -> 403 และสถานะไม่เปลี่ยน | PASS |
| ตรวจเคลม 3: staff อนุมัติ -> 200 APPROVED เหตุผล manual_approved | PASS |
| ตรวจเคลม 4: staff กดไม่อนุมัติซ้ำหลังอนุมัติแล้ว -> 409 และผลแรกยังอยู่ (เจ้าของเห็น APPROVED) | PASS |
| ตรวจเคลม 5: ยอด 20,000 บาทพอดียังอนุมัติอัตโนมัติ (ขอบเขตของกฎใหม่) | PASS |
| ตรวจเคลม 6: staff ไม่อนุมัติเคลมที่รอตรวจ -> 200 REJECTED เหตุผล manual_rejected | PASS |
| ตรวจเคลม 7: log บันทึกการตัดสินโดยคน (claim decided manually) ไม่มี patient_ref | PASS |
| Kong Manager (:8002) เปิดได้ และ Admin API (:8001) เห็น service auth-api, claims-api (อ่านอย่างเดียว ผูก 127.0.0.1) | PASS |
| Kafka UI (:8085) เห็น cluster claims-local ในโหมด read-only | PASS |
| Kafka UI เห็น topic claims.submitted และ claims.submitted.dlq | PASS |

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
?   	github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/cmd/api	[no test files]
?   	github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/cmd/keygen	[no test files]
?   	github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/cmd/mint	[no test files]
?   	github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/cmd/worker	[no test files]
ok  	github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/adapter/http	0.343s
?   	github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/adapter/kafka	[no test files]
?   	github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/adapter/postgres	[no test files]
ok  	github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/adapter/redis	(cached)
ok  	github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/auth	(cached)
?   	github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/auth/authtest	[no test files]
ok  	github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/baseline	(cached)
ok  	github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/config	(cached)
ok  	github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/domain	(cached)
?   	github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/memory	[no test files]
ok  	github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/obs	(cached)
?   	github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/port	[no test files]
ok  	github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/ratelimit	(cached)
ok  	github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/service	(cached)
```

## go vet ./...  [PASS]

```

```

## เทสต์ static (compose, Prometheus, dashboard, migration, secrets)  [PASS]

```
... (60 earlier lines omitted)
test_backend_runs_the_valkey_tests_for_real (test_infra.TestWorkflow.test_backend_runs_the_valkey_tests_for_real) ... ok
test_every_job_has_a_timeout_and_expected_jobs_exist (test_infra.TestWorkflow.test_every_job_has_a_timeout_and_expected_jobs_exist) ... ok
test_images_are_built_not_pushed_and_scanned (test_infra.TestWorkflow.test_images_are_built_not_pushed_and_scanned) ... ok
test_least_privilege_and_no_dangerous_triggers (test_infra.TestWorkflow.test_least_privilege_and_no_dangerous_triggers) ... ok
test_scans_cover_deps_secrets_and_files (test_infra.TestWorkflow.test_scans_cover_deps_secrets_and_files) ... ok
test_triggers_are_limited_to_this_folder (test_infra.TestWorkflow.test_triggers_are_limited_to_this_folder) ... ok
test_workflow_paths_match_real_folders (test_infra.TestWorkflow.test_workflow_paths_match_real_folders) ... ok

----------------------------------------------------------------------
Ran 63 tests in 3.382s

OK
```

## render config ของ Kong (ใส่ public key ทุก kid) และสร้าง cert สำหรับ localhost  [PASS]

```
rendered gateway/.rendered/kong.yml with kids: k1
```

## docker compose up --build (api, worker, kong, web, postgres, valkey, kafka, prometheus, grafana, jaeger)  [PASS]

```
... (227 earlier lines omitted)
 Container day13-migrate-1 Waiting 
 Container day13-kafka-init-1 Waiting 
 Container day13-kafka-init-1 Waiting 
 Container day13-valkey-1 Waiting 
 Container day13-kafka-ui-1 Started 
 Container day13-migrate-1 Exited 
 Container day13-migrate-1 Exited 
 Container day13-valkey-1 Healthy 
 Container day13-valkey-1 Healthy 
 Container day13-kafka-init-1 Exited 
 Container day13-api-1 Starting 
 Container day13-kafka-init-1 Exited 
 Container day13-worker-1 Starting 
 Container day13-worker-1 Started 
 Container day13-api-1 Started 
 Container day13-prometheus-1 Starting 
 Container day13-gateway-1 Starting 
 Container day13-prometheus-1 Started 
 Container day13-grafana-1 Starting 
 Container day13-gateway-1 Started 
 Container day13-gateway-1 Waiting 
 Container day13-grafana-1 Started 
 Container day13-gateway-1 Healthy 
 Container day13-web-1 Starting 
 Container day13-web-1 Started
```

## api ready (GET /readyz 200) และ migration 000002 ผ่าน  [PASS]

```
NAME                 IMAGE                             COMMAND                  SERVICE      CREATED          STATUS                            PORTS
day13-api-1          day13-api                         "/app"                   api          28 seconds ago   Up 14 seconds                     127.0.0.1:8080->8080/tcp
day13-gateway-1      kong:3.9.1                        "/docker-entrypoint.…"   gateway      28 seconds ago   Up 13 seconds (healthy)           127.0.0.1:8000-8002->8000-8002/tcp, 127.0.0.1:8443->8443/tcp
day13-grafana-1      grafana/grafana:11.3.0            "/run.sh"                grafana      28 seconds ago   Up 13 seconds                     127.0.0.1:3001->3000/tcp
day13-jaeger-1       jaegertracing/all-in-one:1.62.0   "/go/bin/all-in-one-…"   jaeger       29 seconds ago   Up 27 seconds                     127.0.0.1:16686->16686/tcp
day13-kafka-1        apache/kafka:3.8.0                "/__cacert_entrypoin…"   kafka        29 seconds ago   Up 27 seconds (healthy)           127.0.0.1:9094->9094/tcp
day13-kafka-init-1   apache/kafka:3.8.0                "/bin/sh -c '/opt/ka…"   kafka-init   28 seconds ago   Exited (0) 15 seconds ago         
day13-kafka-ui-1     provectuslabs/kafka-ui:v0.7.2     "/bin/sh -c 'java --…"   kafka-ui     29 seconds ago   Up 20 seconds                     127.0.0.1:8085->8080/tcp
day13-migrate-1      migrate/migrate:v4.18.1           "migrate -path=/migr…"   migrate      29 seconds ago   Exited (0) 23 seconds ago         
day13-postgres-1     postgres:16-alpine                "docker-entrypoint.s…"   postgres     29 seconds ago   Up 27 seconds (healthy)           127.0.0.1:5432->5432/tcp
day13-prometheus-1   prom/prometheus:v2.55.1           "/bin/prometheus --c…"   prometheus   28 seconds ago   Up 14 seconds                     127.0.0.1:9091->9090/tcp
day13-valkey-1       valkey/valkey:8-alpine            "tini -- docker-entr…"   valkey       29 seconds ago   Up 27 seconds (healthy)           127.0.0.1:6379->6379/tcp
day13-web-1          day13-web                         "/docker-entrypoint.…"   web          28 seconds ago   Up 2 seconds (health: starting)   127.0.0.1:3000->8080/tcp
day13-worker-1       day13-worker                      "/app"                   worker       28 seconds ago   Up 14 seconds
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

## เทสต์ adapter กับ valkey จริง (บัญชีเพิกถอน token, cut ต่อ user, rate limit) ไม่ถูก skip  [PASS]

```
=== RUN   TestRevokerTokenAndExpiry
--- PASS: TestRevokerTokenAndExpiry (2.00s)
=== RUN   TestRevokerUserCut
--- PASS: TestRevokerUserCut (0.00s)
=== RUN   TestLimiterWindow
--- PASS: TestLimiterWindow (2.21s)
PASS
ok  	github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/adapter/redis	4.221s
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
{"id":"97619478-4e88-445e-9b64-eef15bc16ff7","owner_id":"alice","patient_ref":"FAKE-AB12","amount":150000,"status":"PENDING","created_at":"2026-11-05T13:33:54.39876Z","updated_at":"2026-11-05T13:33:54.39876Z"}

```

## alice ส่ง key เดิมซ้ำ -> 200 id เดิม  [PASS]

```
{"id":"97619478-4e88-445e-9b64-eef15bc16ff7","owner_id":"alice","patient_ref":"FAKE-AB12","amount":150000,"status":"PENDING","created_at":"2026-11-05T13:33:54.39876Z","updated_at":"2026-11-05T13:33:54.39876Z"}

```

## bob ใช้ key เดียวกับ alice (body ต่าง) -> 201 เคลมใหม่ของ bob ไม่ใช่ 409  [PASS]

```
{"id":"b3c936ba-eebc-4d85-b687-d9a0e6175742","owner_id":"bob","patient_ref":"FAKE-ZZ99","amount":777,"status":"PENDING","created_at":"2026-11-05T13:33:54.477741Z","updated_at":"2026-11-05T13:33:54.477741Z"}

```

## IDOR: bob เปิดเคลมของ alice -> 404 เหมือนเคลมที่ไม่มีอยู่ และไม่มีข้อมูลหลุด  [PASS]

```
code=404 body={"error":"claim not found"}

```

## alice เปิดเคลมของตัวเอง -> 200  [PASS]

```
{"id":"97619478-4e88-445e-9b64-eef15bc16ff7","owner_id":"alice","patient_ref":"FAKE-AB12","amount":150000,"status":"PENDING","created_at":"2026-11-05T13:33:54.39876Z","updated_at":"2026-11-05T13:33:54.39876Z"}

```

## ops (staff) เปิดเคลมของ alice -> 200  [PASS]

```
{"id":"97619478-4e88-445e-9b64-eef15bc16ff7","owner_id":"alice","patient_ref":"FAKE-AB12","amount":150000,"status":"PENDING","created_at":"2026-11-05T13:33:54.39876Z","updated_at":"2026-11-05T13:33:54.39876Z"}

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
19f67c57c79f14190d4f7a541a7ec6c2
```

## port ที่ publish ผูก 127.0.0.1 ทั้งหมด  [PASS]

```
day13-api-1 127.0.0.1:8080->8080/tcp
day13-gateway-1 127.0.0.1:8000-8002->8000-8002/tcp, 127.0.0.1:8443->8443/tcp
day13-grafana-1 127.0.0.1:3001->3000/tcp
day13-jaeger-1 127.0.0.1:16686->16686/tcp
day13-kafka-1 127.0.0.1:9094->9094/tcp
day13-kafka-ui-1 127.0.0.1:8085->8080/tcp
day13-postgres-1 127.0.0.1:5432->5432/tcp
day13-prometheus-1 127.0.0.1:9091->9090/tcp
day13-valkey-1 127.0.0.1:6379->6379/tcp
day13-web-1 127.0.0.1:3000->8080/tcp
day13-worker-1
```

## worker ตัดสินเป็น APPROVED (ผ่านเส้นทางเดิม outbox -> Kafka -> worker)  [PASS]

```
{"id":"97619478-4e88-445e-9b64-eef15bc16ff7","owner_id":"alice","patient_ref":"FAKE-AB12","amount":150000,"status":"APPROVED","decision_reason":"ok","created_at":"2026-11-05T13:33:54.39876Z","updated_at":"2026-11-05T13:33:55.504273Z"}

```

## เกิน 50,000 บาท -> REJECTED over_limit (กฎของ Day 8 ยังอยู่)  [PASS]

```
{"id":"b7bdaaab-4923-4557-b402-f661484839d6","owner_id":"alice","patient_ref":"FAKE-OVER1","amount":600000000,"status":"REJECTED","decision_reason":"over_limit","created_at":"2026-11-05T13:33:59.346246Z","updated_at":"2026-11-05T13:34:00.456024Z"}

```

## ส่งเคลม poison (ต้องล้มทุกครั้ง) -> 201  [PASS]

```
{"id":"d63affb4-9348-487c-a2bc-14d64e02799f","owner_id":"alice","patient_ref":"FAKE-POISON","amount":100,"status":"PENDING","created_at":"2026-11-05T13:34:01.414621Z","updated_at":"2026-11-05T13:34:01.414621Z"}

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
traces=1
```

## trace เดียวผ่านทั้ง claims-api และ claims-worker (trace context ข้าม Kafka ผ่าน payload)  [PASS]

```
consume claims.submitted | outbox publish claims.submitted | repo.Create | POST /claims | repo.Get | repo.UpdateStatus
```

## trace นั้นมี span ของ repo.Create, outbox publish และ consume  [PASS]

```
consume claims.submitted | outbox publish claims.submitted | repo.Create | POST /claims | repo.Get | repo.UpdateStatus
```

## log ของ request เป็น JSON มี trace_id, span_id, request_id  [PASS]

```
api-1     | {"time":"2026-11-05T13:33:54.405754109Z","level":"INFO","msg":"request","method":"POST","route":"/claims","status":201,"duration_ms":10.924,"bytes":210,"trace_id":"ffb81b21691c00553b8d5dbe55f07521","span_id":"ac204cee302a35ed","request_id":"526cd17cd5a537d910470682d32fcd28"}
```

## trace_id ในบรรทัด log เปิดหา trace เดียวกันใน Jaeger ได้  [PASS]

```
trace_id=ffb81b21691c00553b8d5dbe55f07521
```

## log มี security_event ทั้ง auth_failed และ access_denied  [PASS]

```
... (1 earlier lines omitted)
api-1     | {"time":"2026-11-05T13:33:54.242148437Z","level":"WARN","msg":"authentication failed","security_event":"auth_failed","reason":"signature","trace_id":"410c290c6404c70826034e91f9760097","span_id":"959f1c7936fd7155","request_id":"a3e832c201de88e22b6f27fc9bb22abd"}
api-1     | {"time":"2026-11-05T13:33:54.272352013Z","level":"WARN","msg":"authentication failed","security_event":"auth_failed","reason":"alg","trace_id":"14abfa6051896bfbba711959c6302144","span_id":"69058dfc629fb87f","request_id":"44471522d4e1147efac9d0f21f06daea"}
api-1     | {"time":"2026-11-05T13:33:54.318612105Z","level":"WARN","msg":"authentication failed","security_event":"auth_failed","reason":"missing","trace_id":"1ace0560e055cc1dfa0ad81662c136e5","span_id":"a13f8fa1c4a20fe2","request_id":"e2ee205f78201a7f651204c984baf8c2"}
api-1     | {"time":"2026-11-05T13:33:54.519657174Z","level":"WARN","msg":"access denied to claim","security_event":"access_denied","sub":"bob","claim_id":"97619478-4e88-445e-9b64-eef15bc16ff7","trace_id":"486d65a999814676dc99c9d90d24d2d8","span_id":"8c9227b9944e0a91","request_id":"60de8c435c1002163b5ccf6bc8078540"}
```

## log ของ api/worker ไม่มี token, private key หรือ patient_ref  [PASS]

```
สะอาด
```

## api รันเป็น non-root  [PASS]

```
SERVICE  #   UID    PID    PPID   C   STIME  TTY  TIME      CMD
api      1   65532  81569  81544  1   13:33  ?    00:00:01  /app
```

## login ผิด: user มีอยู่กับไม่มีอยู่ ตอบ 401 ข้อความเดียวกันทุกตัวอักษร  [PASS]

```
code=401/401 body={"error":"invalid username or password"}
 เวลา 4177ms vs 121ms (ตัวอย่างเดียว ไม่ใช่การพิสูจน์เรื่อง timing)
```

## login ผ่าน Kong -> 200 access token (sub=alice) + refresh cookie HttpOnly; Secure; SameSite=Strict; Path=/api/auth  [PASS]

```
code=200 sub=alice expires_in=600 cookie(ไม่แสดงค่า)=refresh_token=[value]; Path=/api/auth; Expires=Fri, 06 Nov 2026 01:34:35 GMT; Max-Age=43200; HttpOnly; Secure; SameSite=Strict
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
401=8 429=22 header-remaining=0 retry-after=21
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
มีบรรทัด request=True พบ query/canary/token ใน log=False | ตัวอย่าง: gateway-1  | 172.18.0.1 [05/Nov/2026:13:33:34 +0000] "GET /api/claims HTTP/1.1" 401 26 "Mozilla/5.0 (Windows NT; Windows NT 10.0; en-US) WindowsPowerShell/5.1.26100.9711" || gateway-1  | 172.18.0.1 [05/Nov/2026:13:34:36 +0000] "GET /api/claims HTTP/1.1" 401 26 "Mozilla/5.0 (Windows NT; Windows NT 10.0; en-US) WindowsPowerShell/5.1.26100.9711"
```

## 429 ที่ /api/auth ส่ง Retry-After และ Access-Control-Expose-Headers ให้ origin ของ UI อ่านได้  [PASS]

```
code=429 Retry-After=17 Expose=Retry-After,X-RateLimit-Remaining-Minute Allow-Origin=http://localhost:3000 body={
  "message":"Too many authentication attempts",
  "request_id":"d8fdb438fc92fe7faa8b02423f7bd243"
}
```

## หมุน key ขั้น 1: เพิ่ม k2, JWT_SIGNING_KID=k2, render Kong ใหม่ (รู้ทั้ง k1 และ k2) และยก api+gateway ใหม่  [PASS]

```
created /keys/k2.key and /keys/k2.pub
... (1 earlier lines omitted)
 Container day13-api-1 Recreate 
 Container day13-api-1 Recreated 
 Container day13-gateway-1 Recreate 
 Container day13-gateway-1 Recreated 
 Container day13-api-1 Starting 
 Container day13-api-1 Started 
 Container day13-gateway-1 Starting 
 Container day13-gateway-1 Started
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
api-1  | {"time":"2026-11-05T13:34:49.928382188Z","level":"INFO","msg":"jwt keys loaded","signing_kid":"k2","accepted_kids":["k1","k2"]}
```

## หมุน key ขั้น 2: ถอด k1 -> token เดิมของ k1 ได้ 401 ทั้ง Kong และ API ตรง ส่วน token ของ k2 ยังได้ 200  [PASS]

```
k1 ผ่าน Kong=401 k1 ตรง API=401 k2 ผ่าน Kong=200
... (1 earlier lines omitted)
 Container day13-api-1 Recreate 
 Container day13-api-1 Recreated 
 Container day13-gateway-1 Recreate 
 Container day13-gateway-1 Recreated 
 Container day13-api-1 Starting 
 Container day13-api-1 Started 
 Container day13-gateway-1 Starting 
 Container day13-gateway-1 Started
```

## หลังถอด k1 login ยังใช้ได้ (refresh token ใน DB ไม่ผูกกับ key) และได้ token kid=k2  [PASS]

```
login=200 kid=k2
```

## log ทั้ง Kong และ API ไม่มีเนื้อ private key ของ k1 และ k2  [PASS]

```
สะอาด
```

## revocation 1: login ผ่าน Kong ได้ access token และใช้งานผ่าน Kong ได้ (200)  [PASS]

```
login=200 claims=200
```

## revocation 2: logout พร้อม Bearer ของ access token -> 204 และล้าง cookie  [PASS]

```
code=204
```

## revocation 3: access token เดิม (ลายเซ็นและ exp ยังถูก) หลัง logout -> 401 ทันที ทั้งผ่าน Kong และยิงตรง API  [PASS]

```
kong=401 body={"error":"unauthorized"}
 api=401
```

## revocation 4: login ใหม่ของ user เดียวกันได้ token ใหม่ที่ใช้งานได้ (เพิกถอนเฉพาะ token ที่ logout)  [PASS]

```
login=200 claims=200
```

## revocation 5: ใน valkey มีรายการ rev:jti:<jti> ที่มี TTL (หมดเองไม่เกินอายุ token + leeway) ไม่มีตัว token เก็บอยู่  [PASS]

```
key(jti เท่านั้น)=rev:jti:XMwp85a_qt7_nv_fUGXuPg ttl=629s
```

## revocation 6: metric auth_failures_total{reason=revoked} นับ และ log API มี security_event token_revoked  [PASS]

```
revoked=2
```

## revocation 7: ใช้ refresh token ซ้ำ -> 401 และ access token ทุกตัวของ bob ที่ออกก่อนหน้า (รวมตัวที่เพิ่ง refresh) ตายด้วย (401)  [PASS]

```
refresh=200 ก่อนซ้ำ=200 ซ้ำ=401 token เก่า=401 token จาก refresh=401
```

## revocation 8: valkey มี rev:user:bob (เวลา cut เป็น unix) และ user อื่น (alice) ไม่ถูกกระทบ  [PASS]

```
rev:user:bob=1793885718
```

## revocation 9: bob login ใหม่หลังเหตุการณ์ได้ token ใหม่ที่ใช้งานได้ (cut ไม่ล็อก user ถาวร)  [PASS]

```
login=200 claims=200
```

## valkey ล่ม: คำขอที่มี token ถูกต้อง -> 503 + Retry-After (fail closed) ส่วน /healthz ยัง 200 และ token มั่ว ยัง 401  [PASS]

```
claims=503 retry-after=5 healthz=200 garbage=401
```

## valkey กลับมา: คำขอปกติ 200 อีกครั้ง และ token ที่ logout ไปก่อน restart ยัง 401 (blocklist อยู่รอดด้วย AOF)  [PASS]

```
normal=True revoked-token-after-restart=401
```

## ยก api ใหม่ด้วย USER_RATE_LIMIT=40 (เฉพาะรอบทดสอบ)  [PASS]

```
... (1 earlier lines omitted)
 Container day13-api-1 Recreate 
 Container day13-api-1 Recreated 
 Container day13-gateway-1 Recreate 
 Container day13-gateway-1 Recreated 
 Container day13-api-1 Starting 
 Container day13-api-1 Started 
 Container day13-gateway-1 Starting 
 Container day13-gateway-1 Started
```

## rate limit ต่อ user: เพดาน 40/นาที -> 40 ครั้งแรกผ่าน อีก 5 ครั้งโดน 429 พร้อม Retry-After  [PASS]

```
200=40 429=5 retry-after=60 body={"error":"too many requests"}

```

## rate limit ต่อ user: user อื่นไม่โดนด้วย และ /healthz ไม่นับโควต้า  [PASS]

```
other=200 healthz=200
```

## ตรวจเคลม 1: เคลม 30,000 บาท -> worker ตั้ง PENDING เหตุผล needs_review และไม่เปลี่ยนต่อ  [PASS]

```
{"id":"1698b076-7e7b-4bfb-811c-25c5a9a5c609","owner_id":"alice","patient_ref":"FAKE-REV01","amount":3000000,"status":"PENDING","decision_reason":"needs_review","created_at":"2026-11-05T13:35:40.347558Z","updated_at":"2026-11-05T13:35:41.405372Z"}

```

## ตรวจเคลม 2: เจ้าของเคลม (ไม่ใช่ staff) กดอนุมัติเอง -> 403 และสถานะไม่เปลี่ยน  [PASS]

```
code=403 body={"error":"forbidden"}

```

## ตรวจเคลม 3: staff อนุมัติ -> 200 APPROVED เหตุผล manual_approved  [PASS]

```
{"id":"1698b076-7e7b-4bfb-811c-25c5a9a5c609","owner_id":"alice","patient_ref":"FAKE-REV01","amount":3000000,"status":"APPROVED","decision_reason":"manual_approved","created_at":"2026-11-05T13:35:40.347558Z","updated_at":"2026-11-05T13:35:42.435821Z"}

```

## ตรวจเคลม 4: staff กดไม่อนุมัติซ้ำหลังอนุมัติแล้ว -> 409 และผลแรกยังอยู่ (เจ้าของเห็น APPROVED)  [PASS]

```
reject=409 status=APPROVED
```

## ตรวจเคลม 5: ยอด 20,000 บาทพอดียังอนุมัติอัตโนมัติ (ขอบเขตของกฎใหม่)  [PASS]

```
{"id":"2d48e344-f0d1-419f-a094-fbf67603bc9b","owner_id":"alice","patient_ref":"FAKE-EDGE1","amount":2000000,"status":"APPROVED","decision_reason":"ok","created_at":"2026-11-05T13:35:42.486205Z","updated_at":"2026-11-05T13:35:43.899842Z"}

```

## ตรวจเคลม 6: staff ไม่อนุมัติเคลมที่รอตรวจ -> 200 REJECTED เหตุผล manual_rejected  [PASS]

```
{"id":"420c7361-c8be-42ff-9f7f-def3ba5e6ff7","owner_id":"alice","patient_ref":"FAKE-REV02","amount":4000000,"status":"REJECTED","decision_reason":"manual_rejected","created_at":"2026-11-05T13:35:44.55436Z","updated_at":"2026-11-05T13:35:46.623964Z"}

```

## ตรวจเคลม 7: log บันทึกการตัดสินโดยคน (claim decided manually) ไม่มี patient_ref  [PASS]

```
พบบรรทัด claim decided manually; ไม่พบ patient_ref ใน log
```

## Kong Manager (:8002) เปิดได้ และ Admin API (:8001) เห็น service auth-api, claims-api (อ่านอย่างเดียว ผูก 127.0.0.1)  [PASS]

```
manager=200 admin=200 services=claims-api, auth-api
```

## Kafka UI (:8085) เห็น cluster claims-local ในโหมด read-only  [PASS]

```
code=200 body=[{"name":"claims-local","defaultCluster":null,"status":"online","lastError":null,"brokerCount":1,"onlinePartitionCount":54,"topicCount":3,"bytesInPerSec":null,"bytesOutPerSec":null,"readOnly":true,"ve
```

## Kafka UI เห็น topic claims.submitted และ claims.submitted.dlq  [PASS]

```
code=200
```


