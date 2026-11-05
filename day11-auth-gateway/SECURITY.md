# SECURITY.md — Day 11 (auth ครบวงจร)

ข้อมูลทั้งหมดเป็นข้อมูลสมมติ เอกสารนี้บอกว่าอะไรปิดแล้ว อะไรยังเปิดอยู่ ตามที่โค้ดและเทสต์ทำจริง

## Threat

- ผู้โจมตีที่ไม่มีบัญชี: ยิง API ตรง ปลอม/แก้ token เดา login
- ผู้ใช้ที่ login แล้วอยากเห็นข้อมูลคนอื่น (IDOR) หรือขอ token ในนาม user อื่น
- ผู้ที่ขโมย refresh token ได้ (เช่น จาก disk, log หรือเครือข่าย)
- เว็บอื่นที่หลอกให้ browser ของเหยื่อส่ง request มาที่ API (CSRF) หรือ XSS ที่พยายามอ่าน token
- ไม่อยู่ใน scope วันนี้: ผู้โจมตีที่คุมเครื่องหรือ container host, supply chain ของ base image

## ปิดแล้ว

- **ไม่มี token = ไม่ผ่านด่าน**: Kong ตรวจลายเซ็นและ exp ที่ทางเข้า API ตรวจซ้ำอีกชั้นและยังเป็นตัวตัดสินเจ้าของข้อมูล (Day 10)
- **access token อายุสั้น** (ค่าเริ่มต้น 10 นาที) เก็บใน memory ของ UI เท่านั้น ไม่เขียน localStorage/sessionStorage
- **refresh token** เป็นค่าสุ่ม 32 byte ฝั่ง server เก็บเฉพาะ SHA-256 ใช้ได้ครั้งเดียว (หมุนเวียน) ใช้ซ้ำ = เพิกถอนทั้งสาย
- **cookie** `HttpOnly; Secure; SameSite=Strict` จำกัด path ที่ `/api/auth`
- **CSRF**: `/auth/refresh` และ `/auth/logout` ต้องมี header `X-Requested-With: claims-web` และ CORS อนุญาตเฉพาะ origin ของ UI
- **รหัสผ่าน** PBKDF2-HMAC-SHA256 (stdlib) พร้อม salt ต่อ user ตอบ 401 ข้อความเดียวกันทั้งกรณีไม่มี user และรหัสผิด และคำนวณ hash หลอกเมื่อไม่มี user
- **rate limit** ที่ Kong: `/api/auth` 20 ครั้ง/นาที/IP, `/api` 120 ครั้ง/นาที/IP
- **TLS** ที่ Kong พอร์ต 8443 (cert สร้างเองสำหรับ localhost, key ไม่ commit)
- token ไม่รับจาก query string; ไม่ log token/รหัสผ่าน (มี log security_event แทน)
- port ทุกตัวผูก 127.0.0.1

## ยังไม่ปิด

- **JWT ใช้ secret ร่วม (HS256)** ระหว่างตัวออก token, Kong และ API: ใครได้ secret ก็ออก token ได้ ยังไม่มี key rotation และ `kid`
- **UI ยังเสิร์ฟผ่าน http://localhost:3000** (TLS มีที่ Kong แต่ UI ยังไม่ได้อยู่หลัง TLS) และ cookie `Secure` อาศัยที่ browser ยกเว้นให้ localhost
- **cert self-signed** ต้อง pin เองในเครื่อง ไม่มี CA จริง ไม่มีการต่ออายุอัตโนมัติ และ key อ่านได้ทุกคนในเครื่อง (ตั้งใจเพื่อให้ Kong ใน container อ่านได้)
- **rate limit นับในหน่วยความจำของ Kong** (policy local) ถ้ารันหลาย instance ต้องใช้ redis; ยังไม่มีการล็อกบัญชีแยกต่อ user (ตอนนี้จำกัดต่อ IP อย่างเดียว)
- **ไม่มีระบบจัดการ user** (สมัคร/รีเซ็ตรหัส/MFA) มีแค่ user สมมติที่ seed จาก `DEMO_PASSWORD`
- access token ที่ออกไปแล้วเพิกถอนไม่ได้จนกว่าจะหมดอายุ (≤10 นาที) เพราะ Kong ตรวจแบบ stateless
- ไม่มี OAuth/OIDC จริง, ไม่มี audit log แบบถาวร (มีแค่ log ที่ stdout)
- cookie `Max-Age` ตาม TTL ที่ตั้ง; ยังไม่ได้ทดสอบกับ browser หลายตัว (ทดสอบเฉพาะที่ระบุใน notes/run-results.md)
- **Kong access log บันทึก URL เต็ม** ถ้า client ส่ง token มาใน query string Kong จะปฏิเสธ (401) แต่ token ยังไปอยู่ใน log เจอตอนทดสอบจริงรอบแรก ยังไม่ได้ตั้งให้ตัด query ออกจาก log
- gitleaks ยังรายงาน 4 รายการ ทั้งหมดเป็นไฟล์ที่ run-all สร้างและ git-ignore (`.env`, `certs/server.key`, `gateway/.rendered/kong.yml`) ไม่ถูก commit
