# Day 16 — Angular วันที่ 1 (แผน 5 วัน เริ่มจาก 0)

ข้อมูลในตัวอย่างทั้งหมดเป็นข้อมูลสมมติ แผนเต็ม 5 วันอยู่ในเอกสารแผนเรียน Angular 5 วัน (ไม่ได้เก็บใน repo)

## สถานะวันที่ 1

- สร้างโปรเจกต์ `claim-list-ng` ด้วย Angular CLI 22.2.7 (`@angular/core` 22.2.6, TypeScript 6.0.3, Vitest 5.1.0)
- มี component 2 ตัว: `ClaimCard` (รับ `claim` ผ่าน `input.required`, `selected` ผ่าน `input`, ส่ง `select` ผ่าน `output`) กับ `ClaimList` (เก็บรายการเคลมสมมติ 5 รายการใน `signal`, วนด้วย `@for` + `track`, แสดงรายการที่เลือกด้วย `@if`)
- เทสต์ 9 ข้อผ่านทั้งหมด (`ng test --watch=false`) รวม baseline snapshot 1 ข้อ (`src/app/claim-list/__snapshots__`)
- ลองแก้โค้ดให้พังตั้งใจ 3 แบบ (ส่ง id ผิด, แก้ยอดเงินในข้อมูลสมมติ, ถอด binding `selected`) เทสต์จับได้ทุกแบบ แล้วคืนโค้ดเดิม เทสต์ผ่าน 9/9 อีกครั้ง
- `ng build` ผ่าน: initial bundle 208.99 kB (ประมาณ 57.68 kB หลังบีบอัด)

## ที่ยังไม่ได้ทำ

- ยังไม่ได้เปิดหน้าเว็บดูด้วยตาตัวเอง (ดูจากภาพที่ Playwright แคปเท่านั้น ดูหัวข้อผลรัน ng serve ด้านล่าง)
- รันและเทสต์ใน Linux ฝั่ง cloud (Node 22.23.4) ยังไม่ได้ลองบนเครื่อง Windows ของตัวเอง ใน repo มีแค่ซอร์สกับ `package-lock.json` ต้อง `npm install` เองก่อนรัน
- ยังไม่ได้ต่อ backend Day 8 (เป็นงานวันที่ 3) ข้อมูลตอนนี้เป็น mock ทั้งหมด
- ยังไม่มี CI สำหรับโฟลเดอร์นี้ (ไว้วันที่ 5)

## ผลรัน ng serve จริง (16/11/2026)

รัน `ng serve` แล้วเปิดด้วย Playwright/Chromium ใน Linux ฝั่ง cloud และรัน `ng test` ซ้ำจากซอร์สที่ commit แล้ว ได้ 9/9 ผ่าน

- เปิดมาเจอการ์ด 5 ใบ และข้อความ "ยังไม่ได้เลือกรายการ" (`notes/screenshots/1-initial.png`)
- กดปุ่ม "เลือก" แล้วการ์ดถูกเน้นและข้อความเปลี่ยนเป็น "เลือกอยู่: id" (`notes/screenshots/2-selected.png`)
- ไม่เจอ error ใน console

## ตอนรัน `ng new` เจออะไร (ใช้ `--defaults`)

ค่าเริ่มต้นของ CLI 22.2.7 ที่เห็นจากไฟล์ที่ถูกสร้าง:

- ไม่มี `zone.js` ใน `package.json` และ `app.config.ts` ไม่ใส่ provider ของ zone แปลว่าโปรเจกต์ใหม่ไม่ใช้ zone.js (ดูจากไฟล์ ไม่ได้เปิดเอกสารทางการยืนยัน)
- test runner คือ Vitest กับ jsdom
- `app.config.ts` มีแค่ `provideBrowserGlobalErrorListeners()` กับ `provideRouter(routes)`
- ไม่มี SSR
- ชื่อไฟล์ component ไม่มีคำต่อท้าย `.component` (เช่น `claim-card.ts`, `claim-card.html`)
- `@Component` ที่ CLI สร้างไม่มี `standalone: true` และไม่มี `changeDetection` เพราะเป็นค่าเริ่มต้นแล้ว

## เทียบกับ Vue

| Angular | Vue |
|---|---|
| `input()` / `input.required()` | props |
| `output()` | emits |
| `signal()` อ่านด้วย `count()` | `ref()` อ่านด้วย `.value` |
| `@for (x of xs; track x.id)` | `v-for` + `:key` |
| `@if` / `@else` | `v-if` / `v-else` |
| `imports: [ClaimCard]` ใน component | import แล้วใช้ใน `<script setup>` |

## วิธีรัน

```
cd day16-angular\claim-list-ng
npm install
npx ng test --watch=false
npx ng serve
```
