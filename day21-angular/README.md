# Day 21 — Angular วันที่ 6 (รอบสอง): interceptor และเตรียมต่อ backend จริง

โค้ดอยู่ใน `claim-list-ng/` โน้ตอยู่ใน `notes/day6.md` (ต่อจาก Day 20 คัดลอกโปรเจกต์มาแล้วเพิ่ม interceptor)

## สถานะ
- เทสต์ 86/86 ผ่าน (18 ไฟล์) รันบน Linux (cloud) ยังไม่เคยรัน `ng test` บน Windows
- interceptor แบบฟังก์ชัน 3 ตัว ลงใน `app.config.ts` ด้วย `withInterceptors([idempotency, error, retry])`:
  - `idempotency.interceptor.ts`: POST ที่ไม่มี `Idempotency-Key` ถูกเติมให้ ถ้าฟอร์มใส่มาแล้ว (key เดียวต่อการส่ง) จะไม่เขียนทับ
  - `error.interceptor.ts`: แปลง `HttpErrorResponse` เป็น `ApiError` (status + ข้อความ `{"error": ...}` จาก backend) ที่เดียว
  - `retry.interceptor.ts`: retry เฉพาะ GET ที่ได้ 503 สูงสุด 2 ครั้ง (รอ 300ms แล้ว 600ms) POST ไม่ retry
- ลำดับใน array ทดลองเองใน `interceptor-order.spec.ts`: ตัวแรกเห็น request ก่อนและเห็น response หลังสุด จึงวาง retry ไว้ท้ายสุด (เห็น error ดิบก่อนถูกแปลง) ผลนี้มาจากการทดสอบ ไม่ได้อ่านจากเอกสาร
- component (`claim-list`, `claim-form`) อ่าน error ผ่าน `readApiError()` ซึ่งรับได้ทั้ง `ApiError` และ `HttpErrorResponse` ข้อความที่แสดงเหมือนเดิมทุกข้อ (เทสต์เดิม 66 ข้อผ่านโดยไม่ต้องแก้)
- เทสต์ใหม่ 20 ข้อ: api-error, แต่ละ interceptor, ลำดับ, พฤติกรรมผ่าน `appConfig` จริง และ baseline snapshot ของ request ที่ออกไป (`interceptors.baseline.spec.ts`)
- ทำให้พัง 7 แบบ จับได้ 7 จาก 7: ลบเช็ค key เดิม, retry 2→5, retry ทุก method, retry ทุก status, สลับลำดับ retry/error, ทิ้งข้อความ backend, ถอด interceptor ออกจาก config
- bundle: initial 284.12 kB (Day 20 = 279.85 kB เพิ่ม ~4.3 kB) ยังต่ำกว่า warning 300 kB
- รัน production build กับ mock API แล้วแคปใน `notes/screenshots/`: GET ได้ 503 สองครั้งแล้วสำเร็จ (3 request ใช้ ~3.4 วิ), 503 สามครั้งแล้วหน้าแสดง error พร้อมปุ่มลองใหม่, POST ได้ 503 ไม่ retry และกดส่งซ้ำใช้ key เดิม

## ยังไม่ได้ทำ (ต้องทำบนเครื่องของคุณ)
- **ยังไม่ได้ต่อ Day 8 backend จริงกับ Angular** เครื่อง cloud ที่ใช้ทำวันนี้ไม่มี Docker และโหลด Go module ไม่ได้ จึงรัน API ไม่ได้ (ลง Postgres/Redis ได้และ apply migration ได้ แต่ build API ไม่ได้) ทุกอย่างข้างบนจึงผ่าน mock เท่านั้น
- สิ่งที่รู้อยู่แล้ว: Day 8 backend เคยรันครบ 25/25 ขั้นบนเครื่องจริงเมื่อ 2026-11-02 (`day08-backend/notes/run-results.md`) แต่ยังไม่เคยต่อกับ Angular
- ขั้นตอนต่อจริง (ผลจริงจดตรงนี้): `[จดหลังรัน]`
  1. `cd day08-backend` แล้วรัน `scripts\run-all.ps1 -Keep` (ยกทั้ง stack ปล่อยไว้)
  2. `curl "localhost:8080/claims?limit=5"` ดูว่า response เป็น `{items, limit, offset}` และ id เป็น uuid
  3. `cd day21-angular\claim-list-ng` แล้ว `npx ng serve` (proxy `/api` → :8080) เปิดหน้า list, detail, สร้างเคลม
  4. ถ้าพังให้จดว่า backend ตอบอะไร และ mock ต่างจากของจริงตรงไหน
- backend Day 8 ไม่ตั้ง CORS ต้องเรียกผ่าน dev proxy เสมอ
- ฟอร์มยังไม่มีช่อง "ประเภท" เพราะ backend ไม่รับฟิลด์นอก patient_ref กับ amount (ตัดสินใจ: ไม่เพิ่ม)
- ผล GitHub Actions ของ `day20.yml` ยังไม่ได้ดู (วันนี้ไม่ได้เพิ่ม workflow ใหม่ `day20.yml` ยังตรวจแค่ path `day20-angular/**`)

## วิธีรัน
```
node mock-api/server.mjs                       # port 8080
curl -X POST "localhost:8080/__fail?n=2"       # request ถัดไป 2 ตัวได้ 503 แล้วกลับปกติ (ดู retry)
curl -X POST localhost:8080/__fail             # สลับตอบ 503 ตลอด (ยิงซ้ำเพื่อปิด)
cd claim-list-ng && npm ci && npx ng serve     # http://localhost:4200 (กด "ล็อกอิน (จำลอง)" ก่อนเข้า /claims/new)
npm run test:ci                                # เทสต์ทั้งหมดแบบไม่ watch
npm run ci                                     # เทสต์ แล้ว build production
```
