# Day 11 — Auth ครบวงจร: login ใน UI, refresh token, Kong ตรวจ JWT

ข้อมูลในตัวอย่างทั้งหมดเป็นข้อมูลสมมติ ตัวเลขผลรันจริงดูใน `notes/run-results.md` (ไม่ได้ก๊อปมาไว้ที่นี่)

## ทำอะไร

ต่อจาก Day 10: ผู้ใช้ login ใน UI แล้ว token ผ่าน Kong ไปถึง API ที่ตรวจเจ้าของข้อมูล

```
Vue UI (:3000) ──► Kong (:8000 http, :8443 TLS) ──► API (:8080, ผูก 127.0.0.1)
                    │ /api/auth  rate limit 20/นาที, ไม่ตรวจ JWT
                    └ /api       ตรวจ JWT (iss/exp/ลายเซ็น) + rate limit 120/นาที
```

- Go: `POST /auth/login`, `/auth/refresh`, `/auth/logout` ใช้ **stdlib** ล้วน (PBKDF2 จาก `crypto/pbkdf2`, HS256 JWT เดิมจาก Day 10) ไม่เพิ่ม dependency ใหม่ฝั่งนี้
- Postgres: ตาราง `users`, `refresh_tokens` (เก็บ SHA-256 ของ token) migration `000003`
- Vue: หน้า login, store เก็บ token ใน memory, interceptor แนบ `Authorization` และ refresh อัตโนมัติเมื่อเจอ 401 หนึ่งครั้ง (single-flight), route guard, ปุ่ม logout
- Kong: ปลั๊กอิน jwt, rate-limiting, CORS แบบ credentials, TLS 8443; config เป็นเทมเพลต `gateway/kong.yml.tpl`

## วิธีรัน

ต้องมี Docker Desktop และปิด stack ของ Day 9/10 ก่อน (`scripts\stack-down.bat`) เพราะพอร์ตชนกัน

```
scripts\run-all.bat          # เทสต์ทั้งหมด + ยก stack + ตรวจจริง + เขียน notes\run-results.md
scripts\run-all-keep.bat     # เหมือนกันแต่ปล่อย stack ไว้ให้เปิด UI ที่ http://localhost:3000
```

`run-all` สร้าง `.env` (JWT_SECRET, รหัส Grafana, `DEMO_PASSWORD`) และ cert ของเครื่องให้เอง ทั้งคู่ถูก git-ignore  user สมมติ: `alice`, `bob`, `ops` (staff) รหัสผ่านคือ `DEMO_PASSWORD` ใน `.env`

## เทสต์

- Go: `go test ./...` รวม baseline `testdata/baseline.json`
- Vue: `cd web && npm test` (vitest, มี snapshot ของสัญญา auth)
- static: `python -m unittest discover -s tests -v`
- mutation: แก้โค้ด auth จงใจแล้วดูว่าเทสต์ล้ม ผลจริงอยู่ใน `notes/exercises.md`

อ่านต่อ: `SECURITY.md` (อะไรปิดแล้ว/ยังไม่ปิด), `notes/exercises.md`
