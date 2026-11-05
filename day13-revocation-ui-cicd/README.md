# Day 13 — เพิกถอน access token, rate limit ต่อ user, ตรวจเคลมในหน้า UI, CI/CD และหน้าจอดูระบบ

ข้อมูลในตัวอย่างทั้งหมดเป็นข้อมูลสมมติ ตัวเลขผลรันจริงดูใน `notes/run-results.md` (ไม่ได้ก๊อปมาไว้ที่นี่)

## ทำอะไร

ต่อจาก Day 12 ปิดข้อที่ SECURITY.md ของวันนั้นบอกว่ายังเปิดอยู่ แล้วเพิ่มงานที่ค้างมาตลอด

```
Vue UI (:3000) ──► Kong (:8000 / :8443 TLS) ──► API (:8080)
                    │ ตรวจลายเซ็น+exp, rate limit ต่อ IP      │ ตรวจลายเซ็นซ้ำ + เช็ก blocklist (valkey)
                    │ Admin API :8001 / Kong Manager :8002    │ rate limit ต่อ user (valkey)
                    └ (ผูก 127.0.0.1 ทั้งคู่, อ่านอย่างเดียว)  └ approve / reject (เฉพาะ staff)
Kafka UI :8085 ── ดู topic claims.submitted / DLQ (read-only)
```

1. **เพิกถอน access token ทันที** (ข้อที่ Day 12 บอกว่าทำไม่ได้)
   - ทุก token มี `jti` (สุ่ม 16 byte จาก `crypto/rand`) และ API ต้องเห็น `jti` ถึงรับ token ใช้ stdlib ล้วน ไม่เพิ่ม dependency ฝั่ง auth (valkey ใช้ go-redis ตัวเดิม)
   - logout ที่ส่ง Bearer ไปด้วย: ขึ้นบัญชี `rev:jti:<jti>` ใน valkey จนถึงเวลาที่ token หมดอายุเอง (exp + leeway) จากนั้นรายการหายเอง
   - refresh token ถูกใช้ซ้ำ (ถูกขโมย): ตั้ง `rev:user:<user>` = เวลาปัจจุบัน token ของ user นั้นที่ออกก่อนเวลานี้ตายทั้งหมด (เทียบด้วย `iat` หน่วยวินาที ออกวินาทีเดียวกันก็ถูกบล็อก ฝั่งปลอดภัยไว้ก่อน)
   - valkey ตอบไม่ได้ = ตรวจ blocklist ไม่ได้ = **ปิดไว้ก่อน** (503 พร้อม `Retry-After`) ไม่ปล่อยผ่าน
   - valkey เปิด AOF + volume เพื่อให้บัญชีรอด restart
2. **rate limit ต่อ user** ที่ API (fixed window ใน valkey, ค่าเริ่มต้น 100 ครั้ง/นาที ตั้งด้วย `USER_RATE_LIMIT`, 0 = ปิด) ต่างจาก Kong ที่นับต่อ IP: บัญชีเดียวใช้หลาย IP ก็ชน และคนหลายคนหลัง IP เดียวไม่ชนกันเอง ตอบ 429 พร้อม `Retry-After` ถ้า valkey ล่มจะ **ปล่อยผ่าน** (เลือก availability เพราะ Kong ยังจำกัดต่อ IP อยู่) และ log ไว้
3. **ตรวจเคลมโดยเจ้าหน้าที่**: ยอดเกิน 20,000 ถึง 50,000 บาท worker ไม่ตัดสินเอง (คง PENDING เหตุผล `needs_review`) staff กด **อนุมัติ / ไม่อนุมัติ** ในหน้ารายละเอียด (กดสองครั้งเพื่อยืนยัน) -> `POST /claims/{id}/approve|reject` เฉพาะ staff, ตัดสินซ้ำได้ 409, user ธรรมดาได้ 403
4. **CI/CD จริงบน GitHub Actions**: ไฟล์ `ci/day13.yml` (วางที่ `.github/workflows/day13.yml` ของ repo): test + vet + race, valkey จริงสำหรับเทสต์ adapter, Vue typecheck/test/build, เทสต์ static, build image 3 ตัว + Trivy, govulncheck, gitleaks, Trivy fs (เรียกจาก image ระบุเวอร์ชัน ไม่ใช้ trivy-action) ไม่ push image และไม่มี secret
5. **หน้าจอดูระบบ**: Kong Manager, Kafka UI, Grafana + Prometheus + Jaeger, Valkey ผ่าน Redis Insight (ภาพแคปอยู่ใน `notes/`)

## วิธีรัน

ต้องมี Docker Desktop และปิด stack เก่าก่อน (`scripts\stack-down.bat`)

```
scripts\run-all.bat          # เทสต์ทั้งหมด + ยก stack + ตรวจจริง + เขียน notes\run-results.md
scripts\run-all-keep.bat     # เหมือนกันแต่ปล่อย stack ไว้ให้เปิดหน้าจอต่างๆ
```

| หน้าจอ | ที่อยู่ |
|---|---|
| UI | http://localhost:3000 (user สมมติ `alice`, `bob`, `ops` = staff, รหัสผ่านคือ `DEMO_PASSWORD` ใน `.env`) |
| Kong Manager | http://localhost:8002 (Admin API :8001) |
| Kafka UI | http://localhost:8085 |
| Grafana / Prometheus / Jaeger | :3001 / :9091 / :16686 |

ทุกพอร์ตผูก 127.0.0.1 (static test ตรวจ) Admin API ของ Kong ใน DB-less รับ `POST /config` ได้ จึงห้ามเปิดออกนอกเครื่อง

## เทสต์

- Go: `go test ./...` รวม baseline `testdata/baseline.json` (มีส่วน `day13_http`)
- Valkey จริง: `REDIS_TEST_ADDR=host:6379 go test ./internal/adapter/redis` (ไม่ตั้งตัวแปร = skip; run-all รันบน network ของ compose)
- Vue: `cd web && npm test` และ `npm run typecheck`
- static: `python -m unittest discover -s tests -v` (รวมตรวจ `ci/day13.yml`)

อ่านต่อ: `SECURITY.md` (อะไรปิดแล้ว/ยังไม่ปิด), `notes/`
