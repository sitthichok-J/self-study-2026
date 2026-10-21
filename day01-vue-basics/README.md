# Day 1 — Vue 3 Basics: Claim List

แอปฝึก Vue 3 พื้นฐาน (Vite + TypeScript + `<script setup>`) ข้อมูลทั้งหมดเป็นข้อมูลสมมติ

## วิธีรัน

```bash
npm install
npm run dev       # dev server
npm run build     # type-check (vue-tsc) + build
npm run preview   # เปิดดูผล build
npm test          # unit test (Vitest + Vue Test Utils)
```

ต้องใช้ Node.js 20.19 ขึ้นไป (แนะนำ 22 LTS) เพราะ Vite 8 ใช้ `styleText` จาก `node:util` ที่ Node 18 ไม่มี

## สิ่งที่แอปนี้สาธิต

| หัวข้อ | อยู่ที่ไหน |
|---|---|
| `ref`, `computed` (กรอง/สรุปยอด), `watch`, `onMounted` | `src/App.vue` |
| `reactive` + computed validation | `src/components/ClaimForm.vue` |
| props พร้อม type + `withDefaults` + default slot | `src/components/BaseButton.vue` |
| `defineModel` (v-model บน component) | `src/components/SearchBox.vue` |
| `defineEmits` พร้อม type | `ClaimForm.vue`, `ClaimItem.vue`, `ClaimList.vue` |
| named slot (`#empty`) | `ClaimList.vue` ← ใช้ใน `App.vue` |
| `ref` + `watchEffect` + localStorage (ปุ่มสลับธีม ตามระบบ/สว่าง/มืด) | `src/components/ThemeToggle.vue` |
| `v-for` + `:key`, `v-if` / `v-else`, `v-model.trim/.number` | ทั่วทั้งแอป |

## โครงสร้าง

```
src/
  App.vue                 # state หลัก, filter, summary
  types.ts                # Claim, ClaimStatus, STATUS_LABEL
  styles.css              # CSS variables (light/dark)
  components/
    BaseButton.vue
    SearchBox.vue
    ClaimForm.vue
    ClaimItem.vue
    ClaimList.vue
screenshots/              # ภาพหน้าจอแต่ละสถานะ
```

## ภาพหน้าจอ

| ไฟล์ | สถานะ |
|---|---|
| `01-initial.png` | หน้าแรก |
| `02-search.png` | ค้นหา "รพ.ตัวอย่าง A" |
| `03-filter-pending.png` | กรองสถานะรอตรวจสอบ |
| `04-validation.png` | validation หลังกดส่งโดยยังไม่กรอก |
| `05-added-approved.png` | เพิ่มเคลมใหม่แล้วอนุมัติ |
| `06-empty-slot.png` | ไม่พบรายการ (named slot) |
| `07-dark.png` | dark mode |

## ข้อสังเกตจากการทำ (Day 1)

- ฟอร์มเดิมโชว์ error ตั้งแต่เปิดหน้า เพราะ `computed` ตรวจจากค่าว่างตั้งแต่แรก แก้โดยเพิ่ม `submitted` ref ให้แสดง error หลังกดส่งครั้งแรก
- `computed` เหมาะกับค่าที่ derive จาก state (filter, ผลรวม) ส่วน `watch` ใช้กับ side effect (ตัวอย่างนี้แค่ log)

## Unit test

Vitest + @vue/test-utils + jsdom อยู่ในโฟลเดอร์ `tests/` (26 เทสต์)

| ไฟล์ | ทดสอบอะไร |
|---|---|
| `ClaimForm.test.ts` | ไม่โชว์ error ตอนเปิดหน้า, validation 3 ข้อ, emit พร้อม trim, ล้างฟอร์มหลังส่ง |
| `ClaimItem.test.ts` | ปุ่มตามสถานะ, emit `change-status` / `remove` |
| `ThemeToggle.test.ts` | `data-theme` บน `<html>`, จำค่าใน localStorage, กลับเป็น system |
| `baseline.test.ts` | baseline: snapshot ของทุกคอมโพเนนต์ + หน้า App, ค่าเริ่มต้น (4 รายการ, ยอดรวม 56,000.50), `STATUS_LABEL` |
| `App.test.ts` | ค้นหา, กรองสถานะ, อนุมัติแล้วตัวเลขสรุปเปลี่ยน, empty slot, เพิ่มเคลม |

### กติกา

- ทุกครั้งที่แก้โค้ด ให้เพิ่ม/แก้ unit test ของส่วนที่เปลี่ยนด้วย
- `baseline.test.ts` เป็นจุดอ้างอิงพฤติกรรมปัจจุบัน ถ้าแดงแปลว่าพฤติกรรมหรือหน้าตาเปลี่ยน ถ้าตั้งใจให้เปลี่ยนค่อยรัน `npm run test:update` แล้วดู diff ของ snapshot ก่อน commit
