# Day 12 — ปิดช่องโหว่ Day 11: JWT RS256 + kid + หมุน key, log ไม่มี query, ข้อความ 429

ข้อมูลในตัวอย่างทั้งหมดเป็นข้อมูลสมมติ ตัวเลขผลรันจริงดูใน `notes/run-results.md` (ไม่ได้ก๊อปมาไว้ที่นี่)

## ทำอะไร

ต่อจาก Day 11 ปิดข้อที่ SECURITY.md ของวันนั้นบอกว่ายังเปิดอยู่ 3 ข้อ

```
Vue UI (:3000) ──► Kong (:8000 http, :8443 TLS) ──► API (:8080, ผูก 127.0.0.1)
                    │ /api/auth  rate limit 20/นาที, ไม่ตรวจ JWT, เปิด Retry-After ให้ UI อ่าน
                    └ /api       ตรวจ JWT RS256 ด้วย public key ตาม kid + rate limit 120/นาที
```

- **JWT: HS256 secret ร่วม -> RS256 + `kid` + หมุน key** ตัวออก token (API) ถือ private key คนเดียว Kong ถือแค่ public key ทุก kid ที่ยังรับ alg ถูกล็อกฝั่ง server (HS256/none ถูกปฏิเสธ กัน key confusion) ใช้ stdlib `crypto/rsa` ไม่เพิ่ม dependency
- **Kong access log ไม่มี query string**: ใช้ log format `noquery` (บันทึก path ไม่บันทึก `?...`)
- **UI แสดง 429 พร้อมเวลารอ**: Kong เปิด `Retry-After` ผ่าน CORS แล้ว LoginView บอกเป็นวินาที
- mutation test ทั้ง Go และ Vue (ครั้งแรกที่ทำฝั่ง Vue)

ขั้นตอนหมุน key ดู `SECURITY.md` หัวข้อ "หมุน key"

## วิธีรัน

ต้องมี Docker Desktop และปิด stack ของ Day 9/10 ก่อน (`scripts\stack-down.bat`) เพราะพอร์ตชนกัน

```
scripts\run-all.bat          # เทสต์ทั้งหมด + ยก stack + ตรวจจริง + เขียน notes\run-results.md
scripts\run-all-keep.bat     # เหมือนกันแต่ปล่อย stack ไว้ให้เปิด UI ที่ http://localhost:3000
```

`run-all` สร้าง `.env` (รหัส Grafana, `DEMO_PASSWORD` ไม่มี JWT secret แล้ว), คู่คีย์ RSA ใน `keys/` และ cert ของเครื่องให้เอง ทั้งหมดถูก git-ignore  user สมมติ: `alice`, `bob`, `ops` (staff) รหัสผ่านคือ `DEMO_PASSWORD` ใน `.env`

## เทสต์

- Go: `go test ./...` รวม baseline `testdata/baseline.json`
- Vue: `cd web && npm test` (vitest, มี snapshot ของสัญญา auth)
- static: `python -m unittest discover -s tests -v`
- mutation: แก้โค้ด auth จงใจแล้วดูว่าเทสต์ล้ม ผลจริงอยู่ใน `notes/exercises.md`
- หลักฐานการโจมตีและวิธีแก้: `notes/attack-evidence.md` (ภาพสรุป `notes/attack-evidence.png`)

อ่านต่อ: `SECURITY.md` (อะไรปิดแล้ว/ยังไม่ปิด), `notes/attack-evidence.md`, `notes/exercises.md`
