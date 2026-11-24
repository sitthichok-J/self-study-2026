# Day 22 — Angular วันที่ 7 (รอบสอง): สร้างฟอร์มเคลมใหม่ด้วย Signal Forms แล้วเทียบกับ reactive forms

โค้ดอยู่ใน `claim-list-ng/` โน้ตอยู่ใน `notes/day7.md` (ต่อจาก Day 21 คัดลอกโปรเจกต์มาแล้วเพิ่มฟอร์มแบบ Signal Forms ไว้ข้างฟอร์มเดิม ไม่ได้แทนที่)

## สถานะ
- เทสต์ 106/106 ผ่าน (20 ไฟล์) รันบน Linux (cloud) ยังไม่เคยรัน `ng test` บน Windows
- เพิ่มหน้า `/claims/new-signal` (`claim-form-signal/`) ทำเรื่องเดียวกับ `/claims/new` ทุกอย่าง: patient_ref ต้องเป็น `FAKE-XXXX`, ยอดเงินบาทไม่เกิน 2 ทศนิยม, แปลงเป็นสตางค์ก่อนส่ง, key เดียวต่อการส่ง, error จาก backend แสดงใต้ช่องที่เกี่ยวข้อง
- ใช้ `form()` + validator `required/pattern/min/max/validate` จาก `@angular/forms/signals`, directive `[formField]` กับ `[formRoot]`, ส่งผ่าน `submission.action`
- ความเสถียรของ API: ใน d.ts ที่ติดตั้ง (`@angular/forms` 22.2.x) ตัว API หลักมี tag `@publicApi 22.0` (เฉพาะส่วน WebMCP เป็น `@experimental`) ส่วนหน้า angular.dev ที่เปิดอ่านไม่ได้ระบุป้ายความเสถียรชัดเจน จึงอ้างจาก d.ts เท่านั้น
- เทสต์ใหม่ 20 ข้อ: spec ของฟอร์มใหม่ (mirror ของ reactive), baseline snapshot (ตารางเหมือนของ reactive ทุกบรรทัด) และ `claim-form.parity.spec.ts` ป้อนค่า 7 คู่เข้าทั้งสองฟอร์ม ข้อความ error ต้องเท่ากัน และ POST body ต้องเท่ากัน (`{patient_ref:'FAKE-AB12', amount:1999}`)
- ทำให้พัง 9 แบบ จับได้ 9 จาก 9: ลบ pattern, ลบ min, ลบเช็คทศนิยม, ลบ max, ลบ required ของยอดเงิน, key ไม่ผูกกับ model, ไม่ map error backend ไปที่ช่อง, ไม่ navigate หลังสำเร็จ, ส่งบาทแทนสตางค์
- `describeSubmitError()` ย้ายจาก `claim-form.ts` ไป `api-error.ts` ให้สองฟอร์มใช้ข้อความเดียวกัน (เทสต์ของฟอร์มเดิมผ่านโดยไม่แก้)
- bundle: initial 288.48 kB (Day 21 = 284.12 kB เพิ่ม ~4.36 kB **ยังไม่ได้ไล่หาสาเหตุ**) ต่ำกว่า warning 300 kB; lazy chunk `claim-form-signal` 44.74 kB, `claim-form` 4.16 kB (โค้ด forms ถูกย้ายไปอยู่ chunk ร่วม 41.76 kB)
- รัน production build กับ mock API แล้วแคปใน `notes/screenshots/`: กดส่งตอนว่าง (ไม่ยิง POST), ค่าไม่ถูก, backend ตอบ 400 พูดถึง amount แล้ว error ไปอยู่ใต้ช่องยอดเงิน, POST ได้ 503 แล้วกดส่งซ้ำใช้ Idempotency-Key เดิม, สร้างสำเร็จ และภาพเทียบสองฟอร์มที่ข้อความ error เหมือนกัน
- mock เพิ่ม `POST /__reject?n=1` ให้ POST ถัดไปได้ 400 `invalid input: amount must be ...`

## ยังไม่ได้ทำ (ต้องทำบนเครื่องของคุณ)
- **ยังไม่ได้ต่อ Day 8 backend จริงกับ Angular** (เหมือน Day 21 เครื่อง cloud ไม่มี Docker และโหลด Go module ไม่ได้) ทุกอย่างผ่าน mock ซึ่งไม่ใช่ backend จริง
- Day 8 backend เคยรันครบ 25/25 เมื่อ 2026-11-02 แต่ยังไม่เคยต่อกับ Angular ขั้นตอนต่อจริง (ผลจริงจดตรงนี้): `[จดหลังรัน]`
  1. `cd day08-backend` แล้วรัน `scripts\run-all.ps1 -Keep`
  2. `cd day22-angular\claim-list-ng` แล้ว `npx ng serve` (proxy `/api` → :8080)
  3. กด "ล็อกอิน (จำลอง)" แล้วลองสร้างเคลมจากทั้ง `/claims/new` และ `/claims/new-signal` ดูว่า backend ตอบข้อความ 400 ตรงกับที่ map ไปช่อง (`patient_ref`/`amount`) หรือไม่ การ map ใช้ `includes()` กับข้อความ error จึงเปราะถ้าข้อความจริงต่างจาก mock
- ไม่มี workflow ของ day22 (ยังไม่เพิ่ม และผล Actions ของ `day20.yml` ก็ยังไม่ได้ดู)
- ข้อมูลทั้งหมดเป็นข้อมูลปลอม (`FAKE-`)

## วิธีรัน
```
node mock-api/server.mjs                       # port 8080
curl -X POST "localhost:8080/__reject?n=1"     # POST ถัดไปได้ 400 (ดู error ใต้ช่อง)
curl -X POST "localhost:8080/__fail?n=1"       # request ถัดไปได้ 503
cd claim-list-ng && npm ci && npx ng serve     # http://localhost:4200 (ล็อกอินจำลองก่อนเข้า /claims/new-signal)
npm run test:ci                                # เทสต์ทั้งหมดแบบไม่ watch
npm run ci                                     # เทสต์ แล้ว build production
```
