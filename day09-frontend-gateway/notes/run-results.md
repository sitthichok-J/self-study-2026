# Day 9 run results

Run at: 2026-11-03 17:36:07  Machine: LAPTOP-MICOTO

## Summary

| Step | Result |
|---|---|
| Docker running | PASS |
| day08-backend อยู่ข้างๆ (compose include) | PASS |
| Docker Compose 2.20+ (รองรับ include) | PASS |
| npm ci | PASS |
| vue-tsc typecheck | PASS |
| vitest (unit + baseline) | PASS |
| vite build | PASS |
| docker compose up --build (backend + gateway + web) | PASS |
| api ready ผ่าน gateway (GET /api/readyz 200) | PASS |
| POST /api/claims ผ่าน gateway -> 201 PENDING | PASS |
| worker ตัดสินเป็น APPROVED (GET ผ่าน gateway) | PASS |
| Idempotency-Key ซ้ำ -> 200 id เดิม และ header Idempotent-Replayed ผ่าน gateway มาถึง | PASS |
| key เดิม body ต่าง -> 409 (ผ่าน gateway) | PASS |
| ไม่ส่ง Idempotency-Key -> 400 | PASS |
| body เกิน 1 MB ถูกตัดที่ gateway -> 413 | PASS |
| ยอดเกินวงเงิน -> REJECTED เหตุผล over_limit (แสดงใน UI ได้) | PASS |
| GET /api/claims?status=&limit=11 (ที่ UI ใช้ขอ limit+1) | PASS |
| path นอก /api -> 404 จาก Kong (ไม่ถึง backend) | PASS |
| method ที่ route ไม่อนุญาต (DELETE) -> 404 จาก Kong | PASS |
| CORS preflight จาก web (localhost:3000) ผ่าน และอนุญาต Idempotency-Key | PASS |
| origin อื่น (evil.example) ไม่ถูกอนุญาต: ไม่มี * และไม่สะท้อน origin กลับ | PASS |
| response จริงมี Allow-Origin และ expose Idempotent-Replayed ให้ JS อ่านได้ | PASS |
| web: GET / ส่ง index.html | PASS |
| web: deep link /claims/<id> ได้ index.html (SPA fallback) | PASS |
| web: asset js ได้ 200 พร้อม Cache-Control immutable | PASS |
| web: bundle ฝังที่อยู่ gateway (localhost:8000/api) | PASS |
| web: security header (nosniff) และ /healthz | PASS |
| api ดับ: gateway ตอบ 502/503/504 (UI แสดง error ได้ ไม่ค้าง) | PASS |
| api กลับมา: ผ่าน gateway ใช้ได้อีกครั้ง | PASS |
| rate limit 120/นาที: เกินแล้วได้ 429 + Retry-After | PASS |
| ตอน 429 ยังมี CORS header (browser อ่านข้อความ error ได้) | PASS |
| ปล่อย stack ไว้ (-Keep): เปิด http://localhost:3000 | PASS |

## Docker running  [PASS]

```
client 29.7.2 / server 29.7.2
```

## day08-backend อยู่ข้างๆ (compose include)  [PASS]

```
C:\Users\ASUS\Documents\Learning Self
```

## Docker Compose 2.20+ (รองรับ include)  [PASS]

```
5.4.0
```

## npm ci  [PASS]

```

added 145 packages in 3s
```

## vue-tsc typecheck  [PASS]

```

> claim-tracker-web@0.0.0 typecheck
> vue-tsc -b
```

## vitest (unit + baseline)  [PASS]

```
... (3 earlier lines omitted)


[1m[30m[46m RUN [49m[39m[22m [36mv5.0.5 [39m[90m/app[39m

 [32m✓[39m tests/api.test.ts [2m([22m[2m12 tests[22m[2m)[22m[32m 27[2mms[22m[39m
 [32m✓[39m tests/lib.test.ts [2m([22m[2m40 tests[22m[2m)[22m[32m 32[2mms[22m[39m
 [32m✓[39m tests/store.test.ts [2m([22m[2m4 tests[22m[2m)[22m[32m 14[2mms[22m[39m
 [32m✓[39m tests/baseline.test.ts [2m([22m[2m3 tests[22m[2m)[22m[32m 10[2mms[22m[39m
 [32m✓[39m tests/ThemeToggle.test.ts [2m([22m[2m4 tests[22m[2m)[22m[32m 52[2mms[22m[39m
 [32m✓[39m tests/composables.test.ts [2m([22m[2m10 tests[22m[2m)[22m[32m 26[2mms[22m[39m
 [32m✓[39m tests/router.test.ts [2m([22m[2m3 tests[22m[2m)[22m[33m 591[2mms[22m[39m
   [32m✓[39m router [2m(3)[22m
     [33m[2m✓[22m[39m เส้นทางหลัก[33m 567[2mms[22m[39m
 [32m✓[39m tests/components.test.ts [2m([22m[2m11 tests[22m[2m)[22m[32m 93[2mms[22m[39m
 [32m✓[39m tests/views.test.ts [2m([22m[2m17 tests[22m[2m)[22m[32m 261[2mms[22m[39m

[2m Test Files [22m [1m[32m9 passed[39m[22m[90m (9)[39m
[2m      Tests [22m [1m[32m104 passed[39m[22m[90m (104)[39m
[2m   Start at [22m 10:36:17
[2m   Duration [22m 3.20s[2m (environment 66%, transform 21%, import 7%, tests 5%)[22m
```

## vite build  [PASS]

```
... (9 earlier lines omitted)
dist/index.html                             0.40 kB │ gzip:  0.27 kB
dist/assets/lib-Dn0tixp6.css                0.58 kB │ gzip:  0.25 kB
dist/assets/ClaimDetailView-CNuFQSOr.css    0.78 kB │ gzip:  0.37 kB
dist/assets/ClaimFormView-CqdCBrb4.css      0.87 kB │ gzip:  0.37 kB
dist/assets/index-BV-rlZDr.css              1.94 kB │ gzip:  0.80 kB
dist/assets/ClaimListView-C7oS--bB.css      1.94 kB │ gzip:  0.68 kB
dist/assets/claim-DtjVd_HF.js               0.29 kB │ gzip:  0.23 kB
dist/assets/NotFoundView-BHUtsuzG.js        0.38 kB │ gzip:  0.31 kB
dist/assets/ClaimFormView-BX3Whx5r.js       2.24 kB │ gzip:  1.13 kB
dist/assets/ClaimDetailView-D4_FzVDf.js     2.37 kB │ gzip:  1.28 kB
dist/assets/ClaimListView-T5VaAaCL.js       3.61 kB │ gzip:  1.72 kB
dist/assets/lib-BCDNixhF.js                 3.70 kB │ gzip:  1.87 kB
dist/assets/index-BpHREYZm.js             102.19 kB │ gzip: 39.71 kB

✓ built in 1.00s
```

## docker compose up --build (backend + gateway + web)  [PASS]

```
... (195 earlier lines omitted)
 Container day09-kafka-1 Healthy 
 Container day09-kafka-init-1 Starting 
 Container day09-kafka-init-1 Started 
 Container day09-kafka-init-1 Waiting 
 Container day09-valkey-1 Waiting 
 Container day09-valkey-1 Waiting 
 Container day09-migrate-1 Waiting 
 Container day09-migrate-1 Waiting 
 Container day09-kafka-init-1 Waiting 
 Container day09-migrate-1 Exited 
 Container day09-valkey-1 Healthy 
 Container day09-valkey-1 Healthy 
 Container day09-migrate-1 Exited 
 Container day09-kafka-init-1 Exited 
 Container day09-api-1 Starting 
 Container day09-kafka-init-1 Exited 
 Container day09-worker-1 Starting 
 Container day09-worker-1 Started 
 Container day09-api-1 Started 
 Container day09-gateway-1 Starting 
 Container day09-gateway-1 Started 
 Container day09-gateway-1 Waiting 
 Container day09-gateway-1 Healthy 
 Container day09-web-1 Starting 
 Container day09-web-1 Started
```

## api ready ผ่าน gateway (GET /api/readyz 200)  [PASS]

```
NAME                 IMAGE                     COMMAND                  SERVICE      CREATED          STATUS                            PORTS
day09-api-1          day09-api                 "/app"                   api          27 seconds ago   Up 15 seconds                     0.0.0.0:8080->8080/tcp, [::]:8080->8080/tcp
day09-gateway-1      kong:3.9.1                "/docker-entrypoint.…"   gateway      27 seconds ago   Up 15 seconds (healthy)           0.0.0.0:8000->8000/tcp, [::]:8000->8000/tcp
day09-kafka-1        apache/kafka:3.8.0        "/__cacert_entrypoin…"   kafka        27 seconds ago   Up 26 seconds (healthy)           0.0.0.0:9094->9094/tcp, [::]:9094->9094/tcp
day09-kafka-init-1   apache/kafka:3.8.0        "/bin/sh -c '/opt/ka…"   kafka-init   27 seconds ago   Exited (0) 16 seconds ago         
day09-migrate-1      migrate/migrate:v4.18.1   "migrate -path=/migr…"   migrate      27 seconds ago   Exited (0) 22 seconds ago         
day09-postgres-1     postgres:16-alpine        "docker-entrypoint.s…"   postgres     27 seconds ago   Up 26 seconds (healthy)           0.0.0.0:5432->5432/tcp, [::]:5432->5432/tcp
day09-valkey-1       valkey/valkey:8-alpine    "tini -- docker-entr…"   valkey       27 seconds ago   Up 26 seconds (healthy)           0.0.0.0:6379->6379/tcp, [::]:6379->6379/tcp
day09-web-1          day09-web                 "/docker-entrypoint.…"   web          27 seconds ago   Up 4 seconds (health: starting)   0.0.0.0:3000->8080/tcp, [::]:3000->8080/tcp
day09-worker-1       day09-worker              "/app"                   worker       27 seconds ago   Up 15 seconds
```

## POST /api/claims ผ่าน gateway -> 201 PENDING  [PASS]

```
{"id":"bc04adb5-b77f-483f-8128-9e447c66f6ab","patient_ref":"FAKE-AB12","amount":150000,"status":"PENDING","created_at":"2026-11-03T10:36:57.762362Z","updated_at":"2026-11-03T10:36:57.762362Z"}

```

## worker ตัดสินเป็น APPROVED (GET ผ่าน gateway)  [PASS]

```
{"id":"bc04adb5-b77f-483f-8128-9e447c66f6ab","patient_ref":"FAKE-AB12","amount":150000,"status":"APPROVED","decision_reason":"ok","created_at":"2026-11-03T10:36:57.762362Z","updated_at":"2026-11-03T10:36:59.229326Z"}

```

## Idempotency-Key ซ้ำ -> 200 id เดิม และ header Idempotent-Replayed ผ่าน gateway มาถึง  [PASS]

```
code=200 replayed=true
```

## key เดิม body ต่าง -> 409 (ผ่าน gateway)  [PASS]

```
{"error":"Idempotency-Key was already used with a different request"}

```

## ไม่ส่ง Idempotency-Key -> 400  [PASS]

```
{"error":"Idempotency-Key header is required"}

```

## body เกิน 1 MB ถูกตัดที่ gateway -> 413  [PASS]

```
code=413 {
  "message":"Request size limit exceeded",
  "request_id":"d29ddb01eff6a51094a7211bf2555d2d"
}
```

## ยอดเกินวงเงิน -> REJECTED เหตุผล over_limit (แสดงใน UI ได้)  [PASS]

```
{"id":"d5ee2a4c-1fb4-4636-b8a6-ef918a5a1e3d","patient_ref":"FAKE-OVER1","amount":600000000,"status":"REJECTED","decision_reason":"over_limit","created_at":"2026-11-03T10:37:00.100101Z","updated_at":"2026-11-03T10:37:01.209218Z"}

```

## GET /api/claims?status=&limit=11 (ที่ UI ใช้ขอ limit+1)  [PASS]

```
{"items":[{"id":"bc04adb5-b77f-483f-8128-9e447c66f6ab","patient_ref":"FAKE-AB12","amount":150000,"status":"APPROVED","decision_reason":"ok","created_at":"2026-11-03T10:36:57.762362Z","updated_at":"2026-11-03T10:36:59.229326Z"}],"limit":11,"offset":0}

```

## path นอก /api -> 404 จาก Kong (ไม่ถึง backend)  [PASS]

```
{
  "message":"no Route matched with those values",
  "request_id":"be1e1557cca74328ab34477ad3d3103d"
}
```

## method ที่ route ไม่อนุญาต (DELETE) -> 404 จาก Kong  [PASS]

```
{
  "message":"no Route matched with those values",
  "request_id":"b4b9244f1f020f5b4e8cae00bf9f9167"
}
```

## CORS preflight จาก web (localhost:3000) ผ่าน และอนุญาต Idempotency-Key  [PASS]

```
code=200 origin=http://localhost:3000 headers=accept,content-type,idempotency-key
```

## origin อื่น (evil.example) ไม่ถูกอนุญาต: ไม่มี * และไม่สะท้อน origin กลับ  [PASS]

```
code=200 allow-origin='http://localhost:3000'
```

## response จริงมี Allow-Origin และ expose Idempotent-Replayed ให้ JS อ่านได้  [PASS]

```
origin=http://localhost:3000 expose=idempotent-replayed,retry-after
```

## web: GET / ส่ง index.html  [PASS]

```
... (6 earlier lines omitted)
    <script type="module" crossorigin src="/assets/index-Bq6lpzE8.js"></script>
    <link rel="stylesheet" crossorigin href="/assets/index-BV-rlZDr.css">
  </head>
  <body>
    <div id="app"></div>
  </body>
</html>

```

## web: deep link /claims/<id> ได้ index.html (SPA fallback)  [PASS]

```
code=200
```

## web: asset js ได้ 200 พร้อม Cache-Control immutable  [PASS]

```
file=/assets/index-Bq6lpzE8.js cache=public, max-age=31536000, immutable
```

## web: bundle ฝังที่อยู่ gateway (localhost:8000/api)  [PASS]

```
found=True in assets/lib-CymUMawN.js
```

## web: security header (nosniff) และ /healthz  [PASS]

```
nosniff=nosniff
```

## api ดับ: gateway ตอบ 502/503/504 (UI แสดง error ได้ ไม่ค้าง)  [PASS]

```
code=504 The upstream server is timing out
request_id: fd0a1c08218e8e79ba03183b66da4a34

```

## api กลับมา: ผ่าน gateway ใช้ได้อีกครั้ง  [PASS]

```

```

## rate limit 120/นาที: เกินแล้วได้ 429 + Retry-After  [PASS]

```
ครั้งแรกที่ 429 คือคำขอที่ 114 (รวม 429 ทั้งหมด 27 จาก 140) Retry-After=46 body={
  "message":"API rate limit exceeded",
  "request_id":"d8bbf50c06f9aa26291f3a4ff9e2c44e"
}
```

## ตอน 429 ยังมี CORS header (browser อ่านข้อความ error ได้)  [PASS]

```
code=429 origin=http://localhost:3000
```

## ปล่อย stack ไว้ (-Keep): เปิด http://localhost:3000  [PASS]

```
ปิดด้วย docker compose down -v
```


