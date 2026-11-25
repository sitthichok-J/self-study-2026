# Day 23 — Angular วันที่ 8 (รอบสอง): state store, filter/แบ่งหน้าส่งไป API, httpResource ในหน้า list

โค้ดอยู่ใน `claim-list-ng/` โน้ตอยู่ใน `notes/day8.md` (ต่อจาก Day 22 คัดลอกโปรเจกต์มาแล้วแก้หน้า list)

## สถานะ
- เทสต์ 126/126 ผ่าน (21 ไฟล์) รันบน Linux (cloud และ VM บนเครื่อง ยังไม่เคยรัน `ng test` บน Windows)
- หน้า list เดิมโหลด 100 รายการครั้งเดียวแล้วกรองในเครื่อง ตอนนี้ส่ง `status`, `limit`, `offset` ไปให้ backend กรองและแบ่งหน้า
- `claim-list.store.ts` (`ClaimListStore`) เก็บ state ที่เดียว: `status`, `pageSize`, `page` → `query` (computed) → `claimsResource` (`httpResource`) แล้วเปิดค่า `items`, `loading`, `error`, `hasNext`, `hasPrev` กับ action `setStatus/setPageSize/next/prev/reload` ผูกที่ระดับ component (`providers: [ClaimListStore]`) จึงเกิดใหม่ทุกครั้งที่เปิดหน้า
- `claimsResource` (เดิมมีไว้เป็นแบบฝึกแต่ไม่ได้ใช้ในหน้า) รับ query แล้วส่ง `limit = pageSize + 1`, `offset = page * pageSize`, และ `status` เมื่อไม่ใช่ ALL: backend ไม่ส่ง total มาให้ จึงขอเกินมา 1 แถว ถ้าได้แถวเกินแปลว่ามีหน้าถัดไป (ไม่มีหน้าว่างต่อท้ายเมื่อจำนวนหารลงตัวพอดี)
- เปลี่ยน status หรือขนาดหน้าแล้วกลับหน้าแรกทันที และยิง request เดียว (มีเทสต์)
- ที่พฤติกรรมเปลี่ยนจากเดิม (ตั้งใจ): ยอดรวมกับจำนวนต่อสถานะนับเฉพาะ**ในหน้านี้** (ข้อความบนหน้าเขียนบอกไว้), ช่องค้นหาทำงานเฉพาะรายการในหน้าที่โหลดมา (backend ไม่มี search), ตัวกรองย้ายออกนอกกิ่ง `@if (loading())` ไม่งั้น select ถูกสร้างใหม่ทุกครั้งที่โหลดแล้วเด้งกลับเป็น "ทั้งหมด"
- เทสต์: store spec ใหม่ 11 ข้อ (รวม baseline ลำดับ query ที่ออกไป backend), spec ของ `claimsResource` เขียนใหม่ (offset/limit, เปลี่ยนหน้าระหว่างที่ request เก่าค้างแล้วเก่าถูกยกเลิก), spec ของหน้า list ปรับให้ตัวกรองสถานะต้องยิง request + เพิ่ม 7 ข้อของแบ่งหน้า baseline snapshot ของหน้า list เดิมผ่านโดยไม่ต้องแก้ตาราง
- ทำให้พัง 16 แบบ จับได้ 16 จาก 16 (ไม่ขอ +1, offset ไม่คูณ, ไม่กลับหน้าแรกเมื่อเปลี่ยน status/ขนาดหน้า, ไม่ตัดแถวเกิน, hasNext ใช้ >=, next/prev ไม่เช็คขอบ, ส่ง status=ALL, ไม่ส่ง status, error เป็น null เสมอ, reload ไม่ทำงาน, ปุ่มถัดไปไม่ผูก hasNext, ค้นหาไม่ดู id, select ขนาดหน้าไม่ผูกค่า, ตัวกรองอยู่ในกิ่ง loading) รอบแรกมีหนึ่งแบบที่เทสต์ไม่จับ (ลบ `[selected]` ของ select สถานะ) เพราะ binding นั้นไม่จำเป็นเมื่อ select อยู่นอกกิ่ง loading จึงลบ binding ทิ้ง
- เจอบั๊กจากภาพหน้าจอ ไม่ใช่จากเทสต์: ตอนแรก select "ต่อหน้า" แสดง 5 ทั้งที่ store ขอ 10 (ผมลบ `[selected]` ออกเกินไป) แก้แล้วและเพิ่มเทสต์ค่าเริ่มต้น
- bundle: initial 290.30 kB (Day 22 = 288.48 เพิ่ม ~1.82 kB น่าจะมาจาก store/pager แต่ไม่ได้วัดแยก; ส่วน +4.36 kB ของ Day 22 ก็ยังไม่ได้ไล่สาเหตุ) ต่ำกว่า warning 300 kB
- ความเสถียร: `httpResource` ใน d.ts ที่ติดตั้ง (Angular 22.2.6) ติด `@publicApi 22.0`
- รัน production build กับ mock API แล้วแคปใน `notes/screenshots/` และเก็บรายการ request จริงไว้ที่ `notes/request-log.txt`: หน้า 1/2/3, กรอง PENDING จากหน้า 3 (กลับหน้า 1, select ค้างที่ PENDING), ขนาดหน้า 5, หน้าถัดไปได้ 503 สามครั้ง (interceptor Day 21 retry 2 ครั้ง รวม 3 request) แล้วขึ้น error กดลองใหม่หาย
- mock เพิ่มข้อมูลปลอมเป็น 23 รายการ เรียงใหม่สุดก่อน และตอบ 400 เมื่อ limit/offset เกินช่วง

## ยังไม่ได้ทำ (ต้องทำบนเครื่องของคุณ)
- **ยังไม่ได้ต่อ Day 8 backend จริงกับ Angular** (เครื่อง cloud ไม่มี Docker และโหลด Go module ไม่ได้) ทุกอย่างผ่าน mock ซึ่งไม่ใช่ backend จริง ผลจริงจดตรงนี้: `[จดหลังรัน]`
  1. `cd day08-backend` แล้วรัน `scripts\run-all.ps1 -Keep`
  2. ใส่เคลมให้เกิน 20 รายการ (สร้างผ่านหน้า /claims/new) แล้วดูว่าหน้า 2 ได้แถวถัดไป เรียงตาม `created_at DESC, id DESC` (อ่านจาก `repo.go` ของ Day 8 ยังไม่เคยรันกับ Angular)
  3. `cd day23-angular\claim-list-ng` แล้ว `npx ng serve` ลองกรอง/เปลี่ยนหน้า/ขนาดหน้า ดูว่า backend รับ `limit` ที่ pageSize+1 (MaxLimit ของ backend คือ 100 จึงห้ามตั้งขนาดหน้าเกิน 99)
- offset-pagination มีข้อจำกัด: ถ้ามีเคลมใหม่เข้ามาระหว่างที่ผู้ใช้เปลี่ยนหน้า แถวจะเลื่อนและอาจเห็นซ้ำ/ตกหล่น ยังไม่ได้ลองกับ backend จริง
- ยังไม่มี "หน้า X จาก Y" เพราะ backend ไม่ส่ง total และ query ยังไม่ผูกกับ URL (รีเฟรชแล้วกลับหน้า 1 ทั้งหมด)
- ระหว่างโหลดหน้าถัดไป รายการหายไปแล้วขึ้น "กำลังโหลด" (ไม่ได้เก็บรายการเก่าไว้ให้ดูระหว่างรอ)
- ไม่มี workflow ของ day23 และผล Actions ของ `day20.yml` ก็ยังไม่ได้ดู
- ข้อมูลทั้งหมดเป็นข้อมูลปลอม (`FAKE-`)

## วิธีรัน
```
node mock-api/server.mjs                       # port 8080 (23 รายการ)
curl -X POST "localhost:8080/__fail?n=3"       # request ถัดไป 3 ตัวได้ 503 (ดู retry + error ของหน้า list)
cd claim-list-ng && npm ci && npx ng serve     # http://localhost:4200/claims
npm run test:ci                                # เทสต์ทั้งหมดแบบไม่ watch
npm run ci                                     # เทสต์ แล้ว build production
```
