# Self Study 2026

บันทึกและโค้ดจากการเรียนรู้ด้วยตัวเอง แยกโฟลเดอร์ตามวัน ข้อมูลในตัวอย่างทั้งหมดเป็นข้อมูลสมมติ

| วัน | โฟลเดอร์ | หัวข้อ |
|---|---|---|
| 1 | [day01-vue-basics](day01-vue-basics) | Vue 3 พื้นฐาน: แอป Claim List (ref, computed, watch, props/emits, slots) |
| 2 | [day02-vue-architecture](day02-vue-architecture) | Vue 3 ขั้นกลาง: Claim Console (Pinia, router, ชั้น API, composable) |
| 3 | [day03-go-fundamentals](day03-go-fundamentals) | Go พื้นฐาน: CLI เก็บรายการงานลง JSON (interface, custom error, table-driven test) |
| 4 | [day04-go-concurrency](day04-go-concurrency) | Go concurrency: worker pool ที่ cancel ได้, pipeline, fan-in, errgroup, graceful shutdown |
| 5 | [day05-go-service](day05-go-service) | REST API จัดการ claims ด้วย net/http + Postgres + migration, slog, pprof |
| 6 | [day06-devops](day06-devops) | Docker, Compose, Kubernetes, CI สำหรับ API ของ Day 5 (pipeline ตรวจแบบ static ด้วย actionlint ยังไม่เคยรันจริง) |
| 7 | [day07-terraform](day07-terraform) | Terraform + AWS: S3, SQS + DLQ, IAM, VPC รันบน LocalStack ยังไม่ได้ apply กับ AWS จริง |
| 8 | [day08-backend](day08-backend) | Capstone backend: Claim Tracker Lite (Go, Postgres, Valkey, Kafka, outbox, DLQ) รันจริงผ่าน 25/25 ขั้น |
| 9 | [day09-frontend-gateway](day09-frontend-gateway) | Capstone frontend + gateway + CI/CD: Vue 3 + Kong + pipeline ต่อกับ backend ของ Day 8 รันจริงผ่าน 32/32 ขั้น (k8s และ pipeline บน GitHub ยังไม่เคยรัน) |
| 10 | [day10-observability-security](day10-observability-security) | Observability + Security: Prometheus, Grafana, Jaeger, JWT, ตรวจเจ้าของข้อมูล, สแกนช่องโหว่และแก้แล้ว รันจริงผ่าน 57/57 ขั้น (UI/Kong ของ Day 9 ยังไม่ส่ง token) |
| 11 | [day11-auth-gateway](day11-auth-gateway) | Auth ครบวงจร: login ใน Vue, refresh token หมุนเวียน, Kong ตรวจ JWT + rate limit + TLS รันจริงผ่าน 86/86 ขั้น (JWT ยังเป็น secret ร่วม HS256, ยังไม่ได้ลอง UI ด้วยมือ) |
| 12 | [day12-auth-hardening](day12-auth-hardening) | ปิดช่องโหว่ Day 11: JWT RS256 + kid + หมุน key, Kong log ไม่มี query string, UI แสดงเวลารอตอนโดน 429 รันจริงผ่าน 101/101 ขั้น (ฟอร์ม 429 ส่งผ่านสคริปต์ ไม่ได้พิมพ์ด้วยมือ, access token ที่ออกไปแล้วยังเพิกถอนไม่ได้) |
| 13 | [day13-revocation-ui-cicd](day13-revocation-ui-cicd) | เพิกถอน access token ทันที (jti + blocklist ใน Valkey), rate limit ต่อ user, เจ้าหน้าที่ตรวจเคลมยอดสูงบนหน้า UI, workflow CI บน GitHub Actions (ยังไม่เคยรันบน GitHub), หน้าจอ Kong Manager / Kafka UI / Grafana รันจริงผ่าน 126/126 ขั้น (ลองหน้า UI ด้วยมือบางส่วน ดู SECURITY.md) |
| 14 | [day14-k8s-kind](day14-k8s-kind) | รัน Claim Tracker Lite บน Kubernetes จริงด้วย kind ใน GitHub Actions รอบที่ 2 เขียว smoke 7/7 ผ่าน Kong จนถึง worker (ยังไม่ได้รันบนเครื่อง Windows, ยังไม่ได้ลองแบบฝึกหัด rollout/OOMKilled) |
| 15 | [day15-observability-kind](day15-observability-kind) | Prometheus + Grafana บน cluster ของ Day 14, alert 7 ตัวจาก Day 13, promtool test 14 เคส, runbook CI รอบที่ 2 เขียว smoke 11/11 (ยังไม่ได้ย้ายไป client_golang / OTel SDK เพราะเครื่องที่เขียนไม่มี Go, ยังไม่ได้เปิด Grafana ดูด้วยตา, AuthFailureSpike ไม่เห็น login ผิด) |
| 16 | [day16-angular](day16-angular) | Angular วันที่ 1 (แผน 5 วัน) claim-list-ng: ClaimCard กับ ClaimList ใช้ input/output/signal/@for, เทสต์ 9 ข้อ + baseline snapshot, ลองแก้โค้ดให้พัง 3 แบบเทสต์จับได้ครบ, รัน ng serve แคปหน้าจอแล้ว (ใน Linux ฝั่ง cloud ยังไม่ได้ลองบน Windows, ยังเป็น mock ไม่มี backend, ยังไม่มี CI) |
| 17 | [day17-angular](day17-angular) | Angular วันที่ 2 ตัวกรองสถานะ ช่องค้นหา และยอดรวมด้วย signal/computed, @for/@empty/@switch, effect log, เทสต์ 19 ข้อ + baseline snapshot, ทำให้พัง 4 แบบเทสจับได้ครบ, ng serve ดูบน Windows แคปหน้าจอแล้ว (ยังรัน ng test บน Windows ไม่ได้, ยังเป็น mock ไม่มี backend) |
| 18 | [day18-angular](day18-angular) | Angular วันที่ 3 ClaimService + HttpClient (inject), สถานะ loading/error/ลองใหม่, สตางค์→บาท, httpResource เทียบกับ Observable, dev proxy, เทสต์ 26 ข้อ + baseline snapshot, ทำให้พัง 6 แบบเทสจับได้ครบ (รันบน Linux กับ mock API ที่จำลอง contract ของ Day 8 ไม่ได้รัน backend จริง ไม่ได้รันบน Windows, httpResource ยังไม่ต่อเข้าหน้า) |
| 19 | [day19-angular](day19-angular) | Angular วันที่ 4 routing 3 หน้า (/claims, /claims/new, /claims/:id) + lazy loading, authGuard (CanActivateFn คืน UrlTree), ฟอร์มสร้างเคลม reactive + validator เอง + Idempotency-Key, เทสต์ 64 ข้อ + baseline snapshot 2 ชุด, ทำให้พัง 10 แบบเทสจับได้ครบ (รันบน Linux กับ mock API ที่จำลอง contract ของ Day 8 ไม่ได้รัน backend จริง ไม่ได้รันบน Windows, ยังไม่ได้ทำ Signal Forms, ไม่มีช่อง ประเภท เพราะ backend ไม่รับ) |
