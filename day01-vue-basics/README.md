# Day 1 — Vue 3 Basics: Claim List

แอปฝึก Vue 3 พื้นฐาน (Vite + TypeScript + `<script setup>`) ข้อมูลทั้งหมดเป็นข้อมูลสมมติ

## วิธีรัน

```bash
npm install
npm run dev       # dev server
npm run build     # type-check (vue-tsc) + build
npm run preview   # เปิดดูผล build
```

ต้องใช้ Node.js LTS

## สิ่งที่แอปนี้สาธิต

| หัวข้อ | อยู่ที่ไหน |
|---|---|
| `ref`, `computed` (กรอง/สรุปยอด), `watch`, `onMounted` | `src/App.vue` |
| `reactive` + computed validation | `src/components/ClaimForm.vue` |
| props พร้อม type + `withDefaults` + default slot | `src/components/BaseButton.vue` |
| `defineModel` (v-model บน component) | `src/components/SearchBox.vue` |
| `defineEmits` พร้อม type | `ClaimForm.vue`, `ClaimItem.vue`, `ClaimList.vue` |
| named slot (`#empty`) | `ClaimList.vue` ← ใช้ใน `App.vue` |
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
