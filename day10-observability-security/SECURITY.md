# SECURITY.md — Day 10

ระบบนี้ใช้ข้อมูลสมมติเท่านั้น (`patient_ref` ต้องเป็น `FAKE-XXXX`) เอกสารนี้บอกว่าพิจารณา threat อะไร ปิดอะไรไปแล้ว และอะไรยังเปิดอยู่

## Threat ที่พิจารณา

อ้างอิง OWASP Top 10 ตามหัวข้อในแผน Day 10

| หัวข้อ | ดูอะไร |
|---|---|
| A01 Broken access control | ผู้ใช้เปิดเคลมของคนอื่นด้วย id (IDOR), อ่านรายการของคนอื่น, ใช้ idempotency key ซ้ำข้ามผู้ใช้, cache ข้ามผู้ใช้ |
| A02/A07 Authentication | JWT ปลอม, alg=none, แก้ payload, token หมดอายุ/อายุยาวเกิน, issuer/audience ผิด |
| A03 Injection | SQL ทุกคำสั่งเป็น parameterized, body ไม่รับ field แปลกปลอม |
| A05 Security misconfiguration | container root, port เปิดกว้าง, header, endpoint metrics เปิดสาธารณะ |
| A06 Vulnerable components | govulncheck, สแกน image |
| A09 Logging/monitoring failures | log เหตุการณ์ความปลอดภัย, ไม่ใส่ token/ข้อมูลส่วนบุคคลลง log |
| Secrets | ไม่มีใน repo |

## ปิดแล้ว

- **Authentication:** ทุก route ของ `/claims` ต้องมี `Authorization: Bearer <JWT>` ตรวจ HS256 อย่างเดียว (อัลกอริทึมกำหนดที่ server ไม่อ่านจาก token), เทียบ signature แบบ constant time, บังคับ `exp` `iss` `aud` `sub`, อายุ token เกิน 1 ชั่วโมงถูกปฏิเสธ, secret ต้องยาว 32 byte ขึ้นไป ไม่งั้น API ไม่ยอมสตาร์ท, ตอบ 401 ข้อความเดียวไม่บอกสาเหตุ (สาเหตุอยู่ใน log กับ metric)
- **Access control:** เคลมมีเจ้าของ (`owner_id` มาจาก `sub` ของ token ไม่ใช่จาก body) เปิดเคลมคนอื่นได้ 404 เหมือนไม่มีเคลมนั้น (แยกไม่ออกว่ามีหรือไม่), รายการคืนเฉพาะของตัวเอง, role `staff` เห็นทั้งหมด, เช็คสิทธิ์หลังอ่านจาก cache ด้วย, cache ของรายการแยกตามผู้ใช้, ผู้เรียกที่ไม่มีตัวตนถูกปฏิเสธ (fail closed)
- **Idempotency:** key ไม่ซ้ำกันเฉพาะภายในเจ้าของเดียวกัน (constraint `UNIQUE (owner_id, idempotency_key)`) ก่อนแก้ ใช้ key ของคนอื่นแล้วจะได้ 409 หรือได้เคลมของเขากลับมา
- **Injection:** query ใช้ parameter ทั้งหมด (`$1..$n`) body ปฏิเสธ field ที่ไม่รู้จัก ขนาดจำกัด 1 MB
- **Misconfiguration:** header `X-Content-Type-Options`, `Cache-Control: no-store`, `Content-Security-Policy: default-src 'none'`, `X-Frame-Options`, `Referrer-Policy` ทุก response; `/metrics` อยู่คนละ listener (`:9090`) ไม่เปิดที่ port ของ API และไม่ publish ออกเครื่อง; ทุก port ที่ publish ผูก `127.0.0.1`; container api/worker รันเป็น non-root (distroless `nonroot`), `read_only`, `cap_drop: ALL`, `no-new-privileges`; Grafana ปิด anonymous และ sign-up
- **Logging:** log เป็น JSON มี `trace_id` `span_id` `request_id`; key ที่ชื่อ authorization/token/password/secret/cookie/jwt/patient_ref ถูกแทนด้วย `[redacted]`; ไม่ log path ดิบ (log เป็น route template); event ความปลอดภัยมี `security_event` = `auth_failed` หรือ `access_denied`; มี metric `auth_failures_total` กับ `authz_denied_total` และ alert
- **Request id:** รับ `X-Request-Id` เฉพาะค่าที่ตรง `[A-Za-z0-9._-]{1,64}` นอกนั้นสร้างใหม่ (กัน header injection / log injection)
- **Secrets:** `JWT_SECRET` กับรหัส Grafana ต้องมาจาก environment (`${VAR:?}` ใน compose) สคริปต์รันสร้างค่าสุ่มลง `.env` ที่ถูก git-ignore; มีเทสต์ static สแกนหา private key / token ใน repo (ตัวเล็กแทน gitleaks)
- **Payload ของ Kafka:** มีแค่ id และ traceparent ไม่มีข้อมูลผู้ป่วย

## ยังไม่ปิด

- **ผลสแกน:** แก้ dependency และขึ้น Go 1.26.10 แล้ว Trivy 0 รายการ govulncheck ไม่พบช่องโหว่ รายละเอียดใน `docs/vulnerability-fixes.md` gitleaks เหลือ 2 จุดที่คาดไว้ (`.env` ที่ git-ignore, ค่าปลอมในเทสต์) ควรสแกนซ้ำเป็นระยะหรือใส่ใน CI เพราะช่องโหว่ใหม่เกิดได้เรื่อยๆ
- **Day 9 (UI + Kong) ยังไม่ส่ง token:** API ตัวนี้ต้องมี token แต่ UI ของ Day 9 ไม่มีหน้า login จึงยังใช้ต่อกันไม่ได้จนกว่าจะเพิ่ม และ Kong ยังไม่ตรวจ JWT เอง
- **JWT เป็น HS256 ที่ใช้ secret ร่วม:** ไม่มี IdP, ไม่มี refresh token, ไม่มี revocation, ไม่มี key rotation; `cmd/mint` เป็นเครื่องมือ dev เท่านั้น
- **ไม่มี TLS:** ทุกอย่างวิ่ง http ในเครือข่าย compose และ loopback
- **ไม่มี rate limit ที่ API:** ของ Day 9 อยู่ที่ Kong (120/นาที/IP) เรียก API ตรงที่ `127.0.0.1:8080` จึงไม่ถูกจำกัด
- **บริการรอบข้างยังเปิดแบบ dev:** Postgres ใช้รหัสผ่านทิ้งใน compose, Valkey ไม่มีรหัสผ่าน, Kafka เป็น PLAINTEXT ไม่มี auth
- **Redaction ตามชื่อ key เท่านั้น:** ถ้าเอา token ไปต่อใน string ของ message เอง จะไม่ถูกปิด
- **ข้อมูลใน Postgres ไม่ได้เข้ารหัส** (เป็นข้อมูลสมมติ)
- **Tracing/metrics ไม่ใช่ SDK มาตรฐาน:** เขียนเองด้วย stdlib (ดู README) ยังไม่ผ่านการใช้งานจริงนอกการทดสอบนี้
