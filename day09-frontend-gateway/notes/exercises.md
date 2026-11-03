# Day 9 - แบบฝึกหัดและคำถาม

## ทำแล้ว

- หน้ารายการ (filter, pagination), ฟอร์ม (validation), หน้ารายละเอียด (polling) พร้อมเทสต์
- gateway: route, rate limit, CORS, จำกัดขนาด body
- pipeline: lint/test, build image, push ghcr, deploy kind (เขียนแล้ว ยังไม่เคยรันจริง)

## ข้อเสริมที่ยังไม่ได้ทำ

- Playwright e2e หนึ่ง flow (ส่งเคลม แล้วเห็นสถานะเปลี่ยน)
- Compose profile dev กับ full
- ตรวจ JWT หรือ API key ที่ gateway

## คำถามค้างคา และคำตอบจากการรันจริง

1. strip_path ของ Kong กับ path `/api` ทำงานตามที่คิดไหม: ทำงานถูก (`/api/claims` ถึง backend เป็น `/claims`)
2. response 429 ที่ Kong สร้างเอง มี header CORS ติดมาไหม: มี (ขั้น "ตอน 429 ยังมี CORS header" ผ่าน) browser อ่านข้อความ error ได้
3. api ดับแล้ว Kong ตอบอะไร: 504
4. rate limit 120/นาที พอสำหรับ UI ไหม: ตัว limit ทำงานจริง (429 ครั้งแรกที่คำขอที่ 107 ของรอบทดสอบ เพราะมีคำขออื่นใช้โควตาไปก่อน) แต่ยังไม่ได้ลองใช้ UI ต่อเนื่องนานๆ ว่า poll กับ refresh จะชนเพดานไหม
5. e2e-kind ใน CI ใช้เวลาเท่าไร: ยังไม่ทราบ ยังไม่เคยรัน
