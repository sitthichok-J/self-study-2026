# Day 18 — Angular วันที่ 3: Service + HttpClient + httpResource

โค้ดอยู่ใน `claim-list-ng/` โน้ตอยู่ใน `notes/day3.md`

## สถานะ (เตรียมไว้ล่วงหน้า ยังไม่ได้ commit)
- เทสต์ 26/26 ผ่าน (รันซ้ำ 18/11 ใน Linux ยังไม่เคยรัน `ng test` บน Windows)
- `ng build` ผ่าน Initial total 236.18 kB
- mutation check 6 แบบ เทสต์จับได้ครบ: ไม่หาร 100, ไม่ล้าง error ตอนลองใหม่, ไม่ปิด loading ตอน error, เช็ค status 0 ผิด, ไม่ส่ง status, limit default ผิด
- รัน `ng serve` + proxy + mock API แล้วแคปหน้าจอ 5 รูปใน `notes/screenshots/` (loading, ข้อมูล, กรอง, error 503, ลองใหม่สำเร็จ) **รันบน Linux ไม่ใช่เครื่อง Windows**
- ส่วนที่ยังค้าง: httpResource ยังไม่ต่อเข้าหน้า, ตัวกรองสถานะยังกรองในหน้า (ไม่ส่ง status ไป API)

## ข้อควรรู้
- **Day 8 backend (Go + Docker) ยังไม่ได้รัน** ที่นี่ ใช้ `mock-api/server.mjs` จำลอง contract เดียวกัน (ข้อมูลเป็น FAKE-xxx) ไม่ใช่ข้อมูลจริง
- เปิดเทียบกับของจริงได้ถ้ารัน backend ที่ port 8080 แล้วไม่ต้องเปิด mock

## วิธีรัน
```
node mock-api/server.mjs        # port 8080 (หน่วงตอบ 1.5 วินาทีให้เห็นสถานะ loading)
curl -X POST localhost:8080/__fail   # สลับให้ตอบ 503 เพื่อดูหน้า error (ยิงซ้ำเพื่อปิด)
cd claim-list-ng && npm ci && npx ng serve   # proxy ตั้งไว้แล้ว เปิด http://localhost:4200
npx ng test --watch=false
```
