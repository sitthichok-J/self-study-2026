# Day 19 — Angular วันที่ 4: Routing + ฟอร์ม

โค้ดอยู่ใน `claim-list-ng/` โน้ตอยู่ใน `notes/day4.md` (ต่อจาก Day 18)

## สถานะ
- เทสต์ 64/64 ผ่าน (11 ไฟล์) รันใน Linux ฝั่ง cloud ยังไม่เคยรัน `ng test` บน Windows
- `ng build` ผ่าน Initial total 279.85 kB, lazy chunk: claim-form 46.13 kB, claim-detail 2.91 kB
- mutation check 10 แบบ เทสต์จับได้ครบ (รายการอยู่ใน notes/day4.md)
- รัน `ng serve` + mock API จริงแล้ว (Linux ฝั่ง cloud ไม่ใช่ Windows) ถ่ายหน้าจอ 8 ภาพไว้ใน `notes/screenshots/` ผลที่ดูได้: `/` ไป `/claims`, กดรายละเอียดไป `/claims/CLM-002` และ chunk ของหน้ารายละเอียดโหลดตอนกด, `/claims/nope` ขึ้นไม่พบเคลม, ยังไม่ล็อกอินกดสร้างเคลมถูกพากลับ `/claims?login=required`, ฟอร์มว่างส่งไม่ได้ (POST 0 ครั้ง), mock ตอบ 503 แล้วกดส่งซ้ำสำเร็จด้วย Idempotency-Key เดิม ได้ CLM-006
- การตรวจ chunk และ header ทำด้วยสคริปต์ Playwright ที่ดัก request ไม่ได้ดู Network tab ด้วยตา และหลักฐาน chunk มาจาก dev server ของ `ng serve` (เสริมกับขนาด chunk จาก `ng build`)
- มี deck สรุป (8 สไลด์) และร่างเมลแล้ว (เมลเป็นแบบร่าง ยังไม่ได้ส่ง)

## ข้อควรรู้
- **Day 8 backend (Go + Docker) ยังไม่ได้รัน** ใช้ `mock-api/server.mjs` (เพิ่ม GET /claims/{id} และ POST /claims ที่ validate ตามกฎของ backend: Idempotency-Key, patient_ref FAKE-xxx, amount สตางค์, ห้ามมีฟิลด์อื่น) ข้อมูลเป็น FAKE-xxx อยู่ในหน่วยความจำ และ id เป็น CLM-xxx ไม่ใช่ uuid เหมือนของจริง
- ฟอร์มไม่มีช่อง "ประเภท" ที่แผนเขียนไว้ เพราะ backend ไม่รับฟิลด์นอกเหนือ patient_ref กับ amount

## วิธีรัน
```
node mock-api/server.mjs             # port 8080
curl -X POST localhost:8080/__fail   # สลับให้ตอบ 503 (ยิงซ้ำเพื่อปิด)
cd claim-list-ng && npm ci && npx ng serve   # เปิด http://localhost:4200
#  กด "ล็อกอิน (จำลอง)" ก่อน ถึงเข้า /claims/new ได้
npx ng test --watch=false
```
