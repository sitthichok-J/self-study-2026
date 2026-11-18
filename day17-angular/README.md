# Day 17 — Angular วันที่ 2 (แผน 5 วัน)

ต่อจาก `day16-angular` (Angular วันที่ 1) ข้อมูลในตัวอย่างทั้งหมดเป็นข้อมูลสมมติ

## สถานะ

- โค้ดอยู่ใน `claim-list-ng` เป็นของวันที่ 1 บวกส่วนที่เพิ่มในวันที่ 2: ตัวกรองสถานะ, ช่องค้นหา, `computed` (รายการที่กรองแล้ว ยอดรวม จำนวนต่อสถานะ), `@empty`, `@switch` ป้ายสี, `effect` หนึ่งตัว
- เทสต์ 19 ข้อผ่าน (วันที่ 1 เดิม 9 ข้อ + เพิ่ม 10 ข้อ รวม baseline snapshot ผลกรองแต่ละสถานะ) รันใน Linux VM ด้วย `ng test --watch=false`
- ทดลองแก้โค้ดให้พัง 4 แบบ เทสต์จับได้ทุกแบบ
- `ng build` ผ่าน 215.43 kB
- เปิด `ng serve` ดูบนเครื่อง Windows แล้ว (ตัวกรอง ค้นหา @empty เลือกการ์ดทำงาน) เก็บภาพ 4 ใบไว้ที่ `notes/screenshots/`

## ที่ยังไม่ได้ทำ

- ยังไม่ได้ลอง mutate array เพื่อดูพฤติกรรม OnPush ตามข้อสุดท้ายของแผน
- ยังไม่ได้รัน `ng test` บน Windows
- ยังไม่มี backend (ไว้วันที่ 3 = Day 18) และยังไม่มี CI

รายละเอียดอยู่ที่ `notes/day2.md`

## วิธีรัน

```
cd day17-angular\claim-list-ng
npm install
npx ng test --watch=false
npx ng serve
```
