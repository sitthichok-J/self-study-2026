# SECURITY.md — Day 12 (ปิดช่องโหว่ของ Day 11)

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
- **JWT เป็น RS256 + `kid`** (Day 12): private key อยู่ที่ API ตัวเดียว (mount อ่านอย่างเดียว) Kong และ API ตรวจด้วย public key ตาม `kid` ส่วน worker ไม่มี key เลย alg ถูกล็อกที่ server: token HS256 ที่เซ็นด้วย public key เป็น secret (key confusion) ได้ 401 ทั้งที่ Kong และ API, `kid` ที่ไม่รู้จักได้ 401
- **หมุน key ได้โดยไม่ทำให้ user หลุด** (ทดสอบจริงใน run-all): เพิ่ม k2 -> token เดิมของ k1 ยังผ่าน, token ใหม่เป็น k2 -> ถอด k1 -> token ของ k1 ได้ 401, k2 ยังผ่าน, login ยังใช้ได้
- **Kong access log ไม่มี query string** (format `noquery`): ทดสอบด้วยการยิง `?jwt=<token>` แล้วเช็กว่าไม่มี token/ค่า canary ใน log แต่ยังเห็น path ของ request
- **UI บอกเวลารอเมื่อโดน 429**: Kong ส่ง `Retry-After` และ `Access-Control-Expose-Headers` ให้ origin ของ UI อ่านได้

## ยังไม่ปิด

- **private key เป็นไฟล์ในโฟลเดอร์ `keys/`** (0644 เพราะ container ที่ไม่ใช่ root ต้องอ่านได้) ใครอ่านเครื่องนี้ได้ก็ออก token ได้ ยังไม่ได้ใช้ KMS/HSM และการหมุน key เป็นงานมือ ไม่มี schedule อัตโนมัติ
- **UI ยังเสิร์ฟผ่าน http://localhost:3000** (TLS มีที่ Kong แต่ UI ยังไม่ได้อยู่หลัง TLS) และ cookie `Secure` อาศัยที่ browser ยกเว้นให้ localhost
- **cert self-signed** ต้อง pin เองในเครื่อง ไม่มี CA จริง ไม่มีการต่ออายุอัตโนมัติ และ key อ่านได้ทุกคนในเครื่อง (ตั้งใจเพื่อให้ Kong ใน container อ่านได้)
- **rate limit นับในหน่วยความจำของ Kong** (policy local) ถ้ารันหลาย instance ต้องใช้ redis; ยังไม่มีการล็อกบัญชีแยกต่อ user (ตอนนี้จำกัดต่อ IP อย่างเดียว)
- **ไม่มีระบบจัดการ user** (สมัคร/รีเซ็ตรหัส/MFA) มีแค่ user สมมติที่ seed จาก `DEMO_PASSWORD`
- access token ที่ออกไปแล้วเพิกถอนไม่ได้จนกว่าจะหมดอายุ (≤10 นาที) เพราะ Kong ตรวจแบบ stateless
- ไม่มี OAuth/OIDC จริง, ไม่มี audit log แบบถาวร (มีแค่ log ที่ stdout)
- cookie `Max-Age` ตาม TTL ที่ตั้ง; ยังไม่ได้ทดสอบกับ browser หลายตัว (ทดสอบเฉพาะที่ระบุใน notes/run-results.md)
- gitleaks รายงาน 2 รายการ ทั้งหมดเป็นไฟล์ที่ run-all สร้างและ git-ignore (`certs/server.key`, `keys/k2.key`) ไม่ถูก commit (Trivy fs เจอ 2 ไฟล์เดียวกัน)
- UI ฝั่งหน้า login ยังไม่ได้ปิดปุ่มระหว่างรอ 429 (บอกเวลาอย่างเดียว) ลองบนเบราว์เซอร์จริงแล้วเฉพาะหน้า login ตอนโดน 429 (ส่งฟอร์มผ่านสคริปต์) หน้าอื่นยังไม่ได้ลองด้วยมือ

## หมุน key (ขั้นตอน)

1. สร้างคู่ใหม่ที่ kid ใหม่ (ไม่เขียนทับ kid เดิม): `go run ./cmd/keygen -dir keys -kid k2`
2. render Kong ใหม่ (`scripts\render-gateway.ps1`) ให้รู้ public key ทั้ง k1 และ k2
3. ตั้ง `JWT_SIGNING_KID=k2` แล้วยก api + gateway ใหม่ (`docker compose up -d --no-deps --force-recreate api gateway`) token ใหม่เป็น k2 ส่วน token k1 ที่ยังไม่หมดอายุยังใช้ได้
4. รออย่างน้อยเท่าอายุ access token (10 นาที) แล้วลบ `k1.key` และ `k1.pub` render Kong ใหม่และยกใหม่อีกรอบ token k1 จะได้ 401

refresh token ไม่ผูกกับ key (เก็บ hash ใน DB) จึง login/refresh ต่อได้หลังถอด k1
