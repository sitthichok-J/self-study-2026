# Day 12 หลักฐาน: โจมตี/ทดสอบอะไร ผลเป็นอย่างไร แก้แบบไหน

ทุกแถวมาจากการรันจริงบนเครื่องวันที่ 05/11/2026 (`run-results.md` ผ่าน 101/101, `security-scan.md`) และ mutation test ที่รันจริง ภาพสรุปอยู่ที่ `attack-evidence.png` ภาพหน้าจอจริง: `screenshot-run-all-101.jpg` (run-all ผ่าน 101/101) และ `screenshot-login-429.jpg` (หน้า login แสดงข้อความ 429)

## A. ช่องโหว่ที่ Day 11 ทิ้งไว้ และวิธีแก้

| ช่องโหว่เดิม | ทำอะไรเพื่อพิสูจน์ | ผลจริง | แก้แบบไหน |
|---|---|---|---|
| JWT ใช้ secret ร่วม HS256 ใครได้ secret ก็ออก token ได้ | ปลอม token HS256 ที่เซ็นด้วย public key เป็น secret อ้าง role=staff (key confusion) ยิงทั้ง Kong และ API ตรง | Kong 401, API 401 | เปลี่ยนเป็น RS256 ล็อก alg ที่ server (รับ RS256 เท่านั้น) ตรวจลายเซ็นด้วย public key ตาม `kid` Kong ถือแค่ public key |
| ไม่มี key rotation | หมุน k1 -> k2 ระหว่างระบบรัน แล้วถอด k1 | ระหว่างหมุน token k1 ยังได้ 200, login ใหม่ได้ kid=k2 หลังถอด k1: token k1 ได้ 401 ทั้ง Kong และ API, token k2 ได้ 200, login ยังใช้ได้ | token มี `kid` ใน header, รับหลาย public key พร้อมกัน, ขั้นตอนหมุนใน SECURITY.md |
| `kid` ที่ไม่รู้จัก | ส่ง token ที่อ้าง kid=k9 | Kong 401, API 401 | kid ไม่อยู่ในชุดที่รับ = ปฏิเสธเหมือนลายเซ็นผิด ไม่ fallback ไปใช้ key อื่น |
| Kong access log เก็บ URL เต็ม token ใน query ไปอยู่ใน log | ยิง `?jwt=<token>` และค่า canary แล้วค้นใน log ของ Kong | ไม่พบ token/canary ใน log แต่ยังเห็นบรรทัด `GET /api/claims ... 401` | ตั้ง log format ชื่อ `noquery` บันทึก path ไม่บันทึก query |
| UI ไม่บอกเวลารอเมื่อโดน 429 | login ผิด 30 ครั้งติดกันผ่าน Kong และส่งฟอร์มผิดซ้ำบนหน้า UI จริง | 401 x 8 แล้ว 429 x 22, `Retry-After` 52 วินาที, หน้า UI แสดง "รออีก 29 วินาที", `Access-Control-Expose-Headers: Retry-After,X-RateLimit-Remaining-Minute` UI อ่านได้ | Kong เปิด header ผ่าน CORS, LoginView แสดง "รออีก N วินาที" (เทสต์ Vue ครอบคลุม 429 มี/ไม่มี Retry-After, เศษวินาทีปัดขึ้น) |

## B. การโจมตีที่ยังต้องผ่านเหมือนเดิม (กันถดถอยจาก Day 11 หลังเปลี่ยน JWT)

| โจมตี | ผลจริง |
|---|---|
| ไม่มี token เข้า Kong | 401 (Kong ตอบ ไม่ถึง API) |
| แก้ payload | 401 `Bad token; invalid JSON` |
| alg=none | 401 `Bad token; invalid alg` |
| ใช้ refresh token ซ้ำ | 401 และสายเดียวกันถูกเพิกถอน |
| IDOR: bob เปิดเคลมของ alice | 404 เหมือนเคลมที่ไม่มีอยู่ |
| ขอ refresh โดยไม่มี header `X-Requested-With` | 403 |
| log ของ Kong/API/worker | ไม่มี token, รหัสผ่าน, private key ของ k1 และ k2 |

## C. เทสต์พิสูจน์ตัวเอง (mutation)

| ฝั่ง | แก้โค้ดจงใจ | รอบแรก | หลังเพิ่มเทสต์ |
|---|---|---|---|
| Go (jwt.go, keys.go, config.go) | 21 จุด | จับได้ 18 (รอด 2 จุด + 1 จุดสคริปต์ผมเขียน pattern ผิด) | 21/21 |
| Vue (LoginView, lib, http, auth store) | 17 จุด | จับได้ 11 รอด 6 | 17/17 |

จุดที่รอด Go: ให้ extra public key ทับ key ของตัวเอง, ไม่ตรวจ kid ของ key เสริม -> เพิ่ม `TestNewKeySetExtraRules` จุดที่รอด Vue: 429 ใช้ข้อความทั่วไปแทนของหน้า login, ปัดเศษ retry-after ลง, retry-after ติดลบ/0/NaN, login ไม่ตั้ง restored, base64url `_` -> เพิ่มเทสต์ 11 ข้อ (Vue 166 -> 177)

## D. สแกน

| เครื่องมือ | ผล |
|---|---|
| govulncheck | No vulnerabilities found |
| Trivy image api (HIGH,CRITICAL) | 0 รายการ |
| Trivy fs | 2 รายการ คือ `certs/server.key` และ `keys/k2.key` (สร้างตอนรัน git-ignore ไม่ถูก commit) |
| gitleaks | 2 รายการ ไฟล์เดียวกัน |

## E. สิ่งที่เจอตอนรันจริงรอบแรก

`scripts/render-gateway.ps1` ไม่มี BOM PowerShell 5.1 อ่านภาษาไทยเพี้ยนแล้ว parse ไม่ผ่าน ทำให้ Kong ไม่ขึ้น และ run-all ค้างรอ gateway รอบแรก (Go test ทั้งชุดผ่านปกติ) แก้โดยใส่ BOM และเพิ่ม static test `test_non_ascii_ps1_has_bom` รอบสองผ่านทั้งหมด (101/101, static test 51 ข้อ)
