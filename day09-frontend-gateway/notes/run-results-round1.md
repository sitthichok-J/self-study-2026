# Day 9 run results

Run at: 2026-11-03 16:37:37  Machine: LAPTOP-MICOTO

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
| path นอก /api -> 404 จาก Kong (ไม่ถึง backend) | FAIL |
| method ที่ route ไม่อนุญาต (DELETE) -> 404 จาก Kong | FAIL |
| CORS preflight จาก web (localhost:3000) ผ่าน และอนุญาต Idempotency-Key | PASS |
| origin อื่นไม่ได้รับ Access-Control-Allow-Origin | FAIL |
| response จริงมี Allow-Origin และ expose Idempotent-Replayed ให้ JS อ่านได้ | PASS |
| web: GET / ส่ง index.html | PASS |
| web: deep link /claims/<id> ได้ index.html (SPA fallback) | PASS |
| web: asset js ได้ 200 พร้อม Cache-Control immutable | PASS |
| web: bundle ฝังที่อยู่ gateway (localhost:8000/api) | FAIL |
| web: security header (nosniff) และ /healthz | FAIL |
| api ดับ: gateway ตอบ 502/503/504 (UI แสดง error ได้ ไม่ค้าง) | PASS |
| api กลับมา: ผ่าน gateway ใช้ได้อีกครั้ง | PASS |
| rate limit 120/นาที: เกินแล้วได้ 429 + Retry-After | PASS |
| ตอน 429 ยังมี CORS header (browser อ่านข้อความ error ได้) | PASS |
| docker compose down -v (ไม่เหลือ container) | PASS |

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
... (10 earlier lines omitted)
304511616d6a: Pull complete
7246d4c6a55d: Pull complete
8776aa3d2ba7: Pull complete
Digest: sha256:fe4f6d74e20370acb04ad889077d14d0f64a9c8ea6a42bc60cddf1e844699914
Status: Downloaded newer image for node:22-alpine

added 145 packages in 4s
npm notice
npm notice New major version of npm available! 10.9.9 -> 12.2.1
npm notice Changelog: https://github.com/npm/cli/releases/tag/v12.2.1
npm notice To update run: npm install -g npm@12.2.1
npm notice
```

## vue-tsc typecheck  [PASS]

```

> claim-tracker-web@0.0.0 typecheck
> vue-tsc -b
```

## vitest (unit + baseline)  [PASS]

```
... (1 earlier lines omitted)
> claim-tracker-web@0.0.0 test
> vitest run


[1m[30m[46m RUN [49m[39m[22m [36mv5.0.5 [39m[90m/app[39m

 [32m✓[39m tests/store.test.ts [2m([22m[2m4 tests[22m[2m)[22m[32m 19[2mms[22m[39m
 [32m✓[39m tests/lib.test.ts [2m([22m[2m40 tests[22m[2m)[22m[32m 22[2mms[22m[39m
 [32m✓[39m tests/api.test.ts [2m([22m[2m12 tests[22m[2m)[22m[32m 35[2mms[22m[39m
 [32m✓[39m tests/composables.test.ts [2m([22m[2m10 tests[22m[2m)[22m[32m 40[2mms[22m[39m
 [32m✓[39m tests/baseline.test.ts [2m([22m[2m3 tests[22m[2m)[22m[32m 9[2mms[22m[39m
 [32m✓[39m tests/router.test.ts [2m([22m[2m3 tests[22m[2m)[22m[33m 359[2mms[22m[39m
   [32m✓[39m router [2m(3)[22m
     [33m[2m✓[22m[39m เส้นทางหลัก[33m 329[2mms[22m[39m
 [32m✓[39m tests/views.test.ts [2m([22m[2m15 tests[22m[2m)[22m[32m 233[2mms[22m[39m

[2m Test Files [22m [1m[32m7 passed[39m[22m[90m (7)[39m
[2m      Tests [22m [1m[32m87 passed[39m[22m[90m (87)[39m
[2m   Start at [22m 09:37:59
[2m   Duration [22m 3.12s[2m (environment 73%, transform 16%, import 6%, tests 4%)[22m
```

## vite build  [PASS]

```
... (5 earlier lines omitted)
transforming...
✓ 51 modules transformed.
rendering chunks...
computing gzip size...
dist/index.html                            0.40 kB │ gzip:  0.27 kB
dist/assets/index-CtzyqVOP.css             1.20 kB │ gzip:  0.60 kB
dist/assets/claim-Dum9UFE-.js              0.26 kB │ gzip:  0.21 kB
dist/assets/NotFoundView-iyApb9u_.js       0.43 kB │ gzip:  0.34 kB
dist/assets/ClaimFormView-CuKNvPS4.js      1.98 kB │ gzip:  1.00 kB
dist/assets/ClaimDetailView-BzmSk5tC.js    2.26 kB │ gzip:  1.22 kB
dist/assets/ClaimListView-BHSkG2Mb.js      2.97 kB │ gzip:  1.48 kB
dist/assets/lib-DnJJE1sN.js                3.34 kB │ gzip:  1.66 kB
dist/assets/index-CJ-QP-6Z.js            100.14 kB │ gzip: 38.75 kB

✓ built in 464ms
```

## docker compose up --build (backend + gateway + web)  [PASS]

```
... (393 earlier lines omitted)
 Container day09-kafka-1 Healthy 
 Container day09-kafka-init-1 Starting 
 Container day09-kafka-init-1 Started 
 Container day09-migrate-1 Waiting 
 Container day09-kafka-init-1 Waiting 
 Container day09-migrate-1 Waiting 
 Container day09-valkey-1 Waiting 
 Container day09-kafka-init-1 Waiting 
 Container day09-valkey-1 Waiting 
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
day09-api-1          day09-api                 "/app"                   api          28 seconds ago   Up 16 seconds                     0.0.0.0:8080->8080/tcp, [::]:8080->8080/tcp
day09-gateway-1      kong:3.9.1                "/docker-entrypoint.…"   gateway      28 seconds ago   Up 15 seconds (healthy)           0.0.0.0:8000->8000/tcp, [::]:8000->8000/tcp
day09-kafka-1        apache/kafka:3.8.0        "/__cacert_entrypoin…"   kafka        28 seconds ago   Up 27 seconds (healthy)           0.0.0.0:9094->9094/tcp, [::]:9094->9094/tcp
day09-kafka-init-1   apache/kafka:3.8.0        "/bin/sh -c '/opt/ka…"   kafka-init   28 seconds ago   Exited (0) 16 seconds ago         
day09-migrate-1      migrate/migrate:v4.18.1   "migrate -path=/migr…"   migrate      28 seconds ago   Exited (0) 23 seconds ago         
day09-postgres-1     postgres:16-alpine        "docker-entrypoint.s…"   postgres     28 seconds ago   Up 27 seconds (healthy)           0.0.0.0:5432->5432/tcp, [::]:5432->5432/tcp
day09-valkey-1       valkey/valkey:8-alpine    "tini -- docker-entr…"   valkey       28 seconds ago   Up 27 seconds (healthy)           0.0.0.0:6379->6379/tcp, [::]:6379->6379/tcp
day09-web-1          day09-web                 "/docker-entrypoint.…"   web          28 seconds ago   Up 4 seconds (health: starting)   0.0.0.0:3000->8080/tcp, [::]:3000->8080/tcp
day09-worker-1       day09-worker              "/app"                   worker       28 seconds ago   Up 16 seconds
```

## POST /api/claims ผ่าน gateway -> 201 PENDING  [PASS]

```
{"id":"d0f434a9-9264-48ac-bfce-d070c594fab2","patient_ref":"FAKE-AB12","amount":150000,"status":"PENDING","created_at":"2026-11-03T09:39:06.600695Z","updated_at":"2026-11-03T09:39:06.600695Z"}

```

## worker ตัดสินเป็น APPROVED (GET ผ่าน gateway)  [PASS]

```
{"id":"d0f434a9-9264-48ac-bfce-d070c594fab2","patient_ref":"FAKE-AB12","amount":150000,"status":"APPROVED","decision_reason":"ok","created_at":"2026-11-03T09:39:06.600695Z","updated_at":"2026-11-03T09:39:07.885267Z"}

```

## Idempotency-Key ซ้ำ -> 200 id เดิม และ header Idempotent-Replayed ผ่าน gateway มาถึง  [PASS]

```
code=200 replayed=true
```

## key เดิม body ต่าง -> 409 (ผ่าน gateway)  [PASS]

```

```

## ไม่ส่ง Idempotency-Key -> 400  [PASS]

```

```

## body เกิน 1 MB ถูกตัดที่ gateway -> 413  [PASS]

```
code=413 
```

## ยอดเกินวงเงิน -> REJECTED เหตุผล over_limit (แสดงใน UI ได้)  [PASS]

```
{"id":"a0a208ff-785d-43af-aa19-c262e0e84b02","patient_ref":"FAKE-OVER1","amount":600000000,"status":"REJECTED","decision_reason":"over_limit","created_at":"2026-11-03T09:39:08.989623Z","updated_at":"2026-11-03T09:39:10.357904Z"}

```

## GET /api/claims?status=&limit=11 (ที่ UI ใช้ขอ limit+1)  [PASS]

```
{"items":[{"id":"d0f434a9-9264-48ac-bfce-d070c594fab2","patient_ref":"FAKE-AB12","amount":150000,"status":"APPROVED","decision_reason":"ok","created_at":"2026-11-03T09:39:06.600695Z","updated_at":"2026-11-03T09:39:07.885267Z"}],"limit":11,"offset":0}

```

## path นอก /api -> 404 จาก Kong (ไม่ถึง backend)  [FAIL]

```

```

## method ที่ route ไม่อนุญาต (DELETE) -> 404 จาก Kong  [FAIL]

```

```

## CORS preflight จาก web (localhost:3000) ผ่าน และอนุญาต Idempotency-Key  [PASS]

```
code=200 origin=http://localhost:3000 headers=accept,content-type,idempotency-key
```

## origin อื่นไม่ได้รับ Access-Control-Allow-Origin  [FAIL]

```
code=200 origin='http://localhost:3000'
```

## response จริงมี Allow-Origin และ expose Idempotent-Replayed ให้ JS อ่านได้  [PASS]

```
origin=http://localhost:3000 expose=idempotent-replayed,retry-after
```

## web: GET / ส่ง index.html  [PASS]

```
... (6 earlier lines omitted)
    <script type="module" crossorigin src="/assets/index-DqG577bK.js"></script>
    <link rel="stylesheet" crossorigin href="/assets/index-CtzyqVOP.css">
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
file=/assets/index-DqG577bK.js cache=public, max-age=31536000, immutable
```

## web: bundle ฝังที่อยู่ gateway (localhost:8000/api)  [FAIL]

```
found=False
```

## web: security header (nosniff) และ /healthz  [FAIL]

```
nosniff=
```

## api ดับ: gateway ตอบ 502/503/504 (UI แสดง error ได้ ไม่ค้าง)  [PASS]

```
code=504 
```

## api กลับมา: ผ่าน gateway ใช้ได้อีกครั้ง  [PASS]

```

```

## rate limit 120/นาที: เกินแล้วได้ 429 + Retry-After  [PASS]

```
ครั้งแรกที่ 429 คือคำขอที่ 107 (รวม 429 ทั้งหมด 34 จาก 140) Retry-After=36 body=
```

## ตอน 429 ยังมี CORS header (browser อ่านข้อความ error ได้)  [PASS]

```
code=429 origin=http://localhost:3000
```

## docker compose down -v (ไม่เหลือ container)  [PASS]

```
... (30 earlier lines omitted)
 Container day09-postgres-1 Stopped 
 Container day09-postgres-1 Removing 
 Container day09-postgres-1 Removed 
 Container day09-kafka-1 Stopped 
 Container day09-kafka-1 Removing 
 Container day09-kafka-1 Removed 
 Network day09_default Removing 
 Network day09_default Removed
```


