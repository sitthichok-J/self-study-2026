# Angular วันที่ 2 (Day 17) — template, control flow และ signals

สถานะ: เขียนโค้ดและเทสต์เสร็จแล้ว (เตรียมล่วงหน้าตอน Day 16) ยังไม่ได้ commit/push

## สิ่งที่เพิ่มจากวันที่ 1

- `ClaimList` มีตัวกรองสถานะ (`statusFilter`) และช่องค้นหา (`searchText`) เก็บเป็น `signal`
- `filtered = computed(...)` สร้างรายการที่กรองแล้ว ไม่มีการเรียกฟังก์ชันกรองใน template
- `totalAmount`, `countByStatus`, `filteredAmount` เป็น `computed` แสดงด้านบนของหน้า
- แสดงรายการด้วย `@for ... track claim.id` พร้อม `@empty` เมื่อไม่พบผลลัพธ์
- `ClaimCard` ใช้ `@switch` เลือกป้ายสีตามสถานะ
- `effect()` ตัวเดียว log เมื่อตัวกรองเปลี่ยน

## ผลที่รันจริง (Linux ใน VM ของ Cowork, Node 22.23.4)

- `ng test --watch=false`: 19/19 ผ่าน (เดิม 9 ข้อยังผ่านทั้งหมด เพิ่มใหม่ 10 ข้อ รวม baseline snapshot ผลกรองแต่ละสถานะ 1 ข้อ)
- ทดลองแก้โค้ดให้พัง 4 แบบ เทสต์จับได้ทุกแบบ แล้วคืนโค้ดเดิม:
  - เปลี่ยนเงื่อนไขกรองเป็น `!==` (พัง 4 ข้อ)
  - ยอดรวมไม่นับรายการแรก (พัง 1 ข้อ)
  - ลบ `@empty` (พัง 1 ข้อ)
  - `@case ('APPROVED')` ใช้ class ของ rejected (พัง 1 ข้อ)
- ลบ `track` ออกจาก `@for` แล้ว `ng build` ไม่ผ่าน: `NG5002: @for loop must have a "track" expression`
- `ng build` ผ่าน initial total 215.43 kB (ประมาณ 59.33 kB หลังบีบอัด) วันที่ 1 คือ 208.99 kB

## ที่ยังไม่ได้ทำ

- ยังไม่ได้เปิดดูหน้าเว็บของวันที่ 2 ด้วย `ng serve` และยังไม่ได้แคปหน้าจอ (ตรวจผ่านเทสต์ jsdom เท่านั้น)
- ยังไม่ได้ลองบน Windows
- ข้อ OnPush กับการ mutate array ในรายการ "เสร็จเมื่อ" ของวันที่ 2 ยังไม่ได้ลองทำเป็นแบบฝึกจริง (ตอนนี้โค้ดใช้ signal ทุกจุด ยังไม่มีส่วนที่ mutate)
- ยังไม่ได้ commit/push

## เทียบกับ Vue

| Angular | Vue |
|---|---|
| `signal()` / `.set()` / `.update()` | `ref()` / `.value = ` |
| `computed()` (อ่านด้วย `x()`) | `computed()` (อ่านด้วย `.value`) |
| `effect()` | `watchEffect()` / `watch()` |
| `@switch` / `@case` | `v-if` / `v-else-if` ต่อกัน |
| `@for ... @empty` | `v-for` + `v-if` เช็กว่าอาร์เรย์ว่าง |
