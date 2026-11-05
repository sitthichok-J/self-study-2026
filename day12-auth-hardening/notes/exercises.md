# แบบฝึกหัด Day 12

จดผลจริงใต้แต่ละข้อ (รันบนเครื่องตัวเองวันที่ 05/11/2026 ตัวเลขทั้งหมดมาจาก `run-results.md`, `security-scan.md` และเทสต์ที่รันจริง)

## 1. โจมตีตัวเอง: JWT หลังเปลี่ยนเป็น RS256

- ปลอม token HS256 โดยใช้ public key เป็น secret อ้าง role=staff (key confusion)
- ส่ง token ที่ `kid` ไม่รู้จัก
- หมุน key ระหว่างระบบรัน แล้วถอด key เก่า
- ยิง `?jwt=<token>` แล้วดู log ของ Kong

ผล: ทำเป็นขั้นใน `scripts/run-all.ps1` ผ่านทั้งหมด (101/101) ดูตารางใน `attack-evidence.md`

- key confusion: Kong 401 และยิงตรงเข้า API ก็ 401
- kid=k9: 401 ทั้งสองที่
- หมุน key: ระหว่างหมุน token เดิมของ k1 ยังได้ 200 login ใหม่ได้ kid=k2 หลังถอด k1 token k1 ได้ 401 ที่ Kong และที่ API ตรง ส่วน k2 ได้ 200 login ยังใช้ได้
- log ของ Kong: มีบรรทัด request แต่ไม่มี query string, token และค่า canary ไม่อยู่ใน log

## 2. Rate limit และข้อความ 429

ผล: login ผิดต่อเนื่อง 30 ครั้ง ได้ 401 อยู่ 8 ครั้งแล้ว 429 อีก 22 ครั้ง header โควต้า = 0, `Retry-After` = 52 วินาที Kong ส่ง `Access-Control-Expose-Headers: Retry-After,X-RateLimit-Remaining-Minute` และ `Allow-Origin` ของ UI จึงอ่านเวลารอได้ ข้อความ 429 ใน LoginView ทำแล้วและมีเทสต์ และลองบนหน้า UI จริงที่ localhost:3000 (stack ของ run-all-keep) แล้ว: ส่งฟอร์ม login ผิดซ้ำจนโดน 429 หน้าแสดง "ลองเข้าสู่ระบบถี่เกินไป รออีก 29 วินาทีแล้วลองใหม่" (ภาพ `screenshot-login-429.jpg`) ข้อจำกัด: ผมกรอกและส่งฟอร์มผ่านสคริปต์ของ extension ไม่ได้พิมพ์ด้วยมือ ส่วนการแยกเพดานต่อ user ยังไม่ได้ทำ การแยกเพดานต่อ user ยังไม่ได้ทำ

## 3. ทำให้เทสต์พิสูจน์ตัวเอง (mutation)

ผล: Go แก้โค้ดจงใจ 21 จุด (alg ไม่ล็อก, kid ไม่รู้จักใช้ key อื่นแทน, ไม่ตรวจลายเซ็น, header ไม่ถูกเซ็น, ไม่ตรวจ exp/aud/iss, ลดขนาด key ขั้นต่ำ, keygen เขียนทับ ฯลฯ) รอบแรกจับได้ 18 รอด 2 (extra public key ทับ key ตัวเองได้, kid ของ key เสริมไม่ถูกตรวจ) อีก 1 จุดเป็น pattern ของสคริปต์ผมเอง เพิ่มเทสต์แล้วได้ 21/21 ฝั่ง Vue ทำครั้งแรก 17 จุด รอบแรกจับได้ 11 รอด 6 (429 ใช้ข้อความทั่วไป, ปัดเศษ retry-after, retry-after ติดลบ/0/NaN, login ไม่ตั้ง restored, base64url `_`) เพิ่มเทสต์ 11 ข้อ (166 -> 177) ได้ 17/17 ส่วนข้อ "ให้ Kong ข้ามการตรวจ JWT" ไม่ได้ทดลอง มีแต่เทสต์ static ของ config

## 4. ทบทวนสิ้นวัน

ผล:

- เทสต์ผ่านแต่ใช้งานจริงพัง: `render-gateway.ps1` ไม่มี BOM PowerShell 5.1 อ่านภาษาไทยเพี้ยนจน parse ไม่ผ่าน Kong ไม่ขึ้น (Go test และ static test ไม่เห็นเพราะไม่ได้รันสคริปต์นี้จริง) แก้โดยใส่ BOM และเพิ่ม static test กัน
- ผลรันสุดท้าย: run-all ผ่าน 101/101, Go test + go vet ผ่าน, static test 51 ข้อ (รอบแรก 50 แล้วเพิ่มข้อ BOM รอบสองบนเครื่องจริงได้ 101/101 และ 51 ข้อ ภาพ `screenshot-run-all-101.jpg`), Vue 177 ข้อ
- สแกน: govulncheck ไม่พบ, Trivy image api 0 รายการ, Trivy fs และ gitleaks เจอ 2 ไฟล์ key ที่ run-all สร้างและ git-ignore (`certs/server.key`, `keys/k2.key`)
- ข้อจำกัดที่ปิดแล้ว: secret ร่วม HS256, ไม่มี key rotation/kid, token ใน query ไปอยู่ใน Kong log, UI ไม่บอกเวลารอตอน 429 ที่ยังเหลือ: private key เป็นไฟล์บนดิสก์และหมุนด้วยมือ, UI ยังเสิร์ฟผ่าน http, cert self-signed, rate limit นับในหน่วยความจำและต่อ IP อย่างเดียว, access token เพิกถอนก่อนหมดอายุไม่ได้ (ไม่เกิน 10 นาที), ยังไม่ได้ทำ mutation ฝั่ง Vue ทุกไฟล์ (ทำเฉพาะ auth) และลอง UI จริงแล้วเฉพาะหน้า login ตอนโดน 429 (ส่งฟอร์มผ่านสคริปต์ ไม่ได้พิมพ์ด้วยมือ)
- เวลาที่ใช้จริง: ไม่ได้จับเวลาไว้ รันจริงบนเครื่อง 2 รอบ (รอบแรกเจอบั๊ก BOM)
