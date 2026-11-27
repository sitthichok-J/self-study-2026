# Day 25 — Angular วันที่ 10 (รอบสอง): e2e ด้วย Playwright ใน CI, accessibility, ng update, สรุปรอบ

โค้ดอยู่ใน `claim-list-ng/` โน้ตอยู่ใน `notes/day10.md` (ต่อจาก Day 24 คัดลอกโปรเจกต์มาแล้วเพิ่ม e2e) ไม่มีการแก้โค้ดแอป (`src/`) ในวันนี้

## สถานะ
- unit test 137/137 ผ่านเหมือน Day 24 (ไม่มีโค้ดใน `src/` เปลี่ยน) เพิ่ม e2e **19 ข้อ** ใน `e2e/` ผ่าน 19/19 รัน 3 รอบติดกันที่พอร์ตเดียวกันไม่มีเทสต์ที่ผ่าน ๆ ล้ม ๆ (ในรอบ `CI=true` ที่เปิด retry 1 ครั้ง ไม่มีข้อไหนถูกรายงานว่า flaky) รันบน Linux เท่านั้น
- e2e รันกับ production build แบบ SSR (`node dist/.../server.mjs`) + mock API จริง ๆ ไม่ใช่ `ng serve`: `playwright.config.ts` เปิดสองตัวให้เองที่พอร์ต 8090 (mock) และ 4310 (แอป) จึงไม่ชนกับที่เปิดเล่นอยู่ ใช้ `npm run ci:e2e` (build แล้วรัน)
- ข้อที่ตรวจ: HTML ดิบจาก server มีการ์ด 10 ใบ, เปลี่ยนหน้า/ย้อนกลับ/หน้าสุดท้าย, กรอง PENDING จากหน้า 2, ขนาดหน้า 5, 503 สองครั้ง (retry กู้ได้ ไม่เห็น error), 503 สามครั้ง (เห็น error แล้วกดลองใหม่), detail แบบ SSR และกดจากรายการ, id ที่ไม่มี, guard (ไม่ล็อกอินถูกส่งไป `/claims?login=required`), สร้างเคลมจนเห็นเคลมใหม่ใบแรก, patient_ref ผิดรูปแบบไม่ยิง POST, backend ตอบ 400
- accessibility: `@axe-core/playwright` สแกนกฎ WCAG 2.0/2.1 A และ AA 5 หน้า (list, detail, ฟอร์ม reactive, ฟอร์มที่มี error, ฟอร์ม Signal Forms) **ไม่เจอ violation ที่ axe ตรวจได้ทั้งหมด** ไม่ได้แปลว่าเข้าถึงได้ครบ axe จับได้เพียงส่วนหนึ่ง ยังไม่ได้ลอง screen reader และการใช้แป้นพิมพ์ล้วนด้วยตาเอง
- ภาพหน้าแอปจริง (production build SSR + mock) ถ่ายด้วย `notes/app-screenshots.mjs` อยู่ใน `notes/screenshots/5-` ถึง `11-`: รายการ, หน้าสุดท้าย, กรอง PENDING, detail, ฟอร์มที่มี error, สร้างเคลมแล้วอยู่ใบแรก (CLM-024), 503 ขึ้น error (ภาพถ่ายหลังเขียนเทสต์เสร็จ ไม่ได้เป็นตัวที่เจอบั๊ก mock ตัวนั้น บั๊กนั้นเจอจากเทสต์ e2e)
- ยืนยันว่า axe จับได้จริงด้วยการทำให้พัง (สีตัวหนังสือ badge PENDING จางลง → `color-contrast (serious)` 3 จุด ดู `notes/screenshots/4-axe-contrast-mutation.png`)
- ทำให้พัง 10 แบบ จับได้ 10 จาก 10 (ลบ label ของ patient_ref, สี badge ตัดกับพื้นไม่พอ, authGuard ปล่อยทุกคน, mock เคลมใหม่ต่อท้ายลิสต์, `/claims` เป็น Client mode, ไม่ retry GET 503, proxy ไม่ส่ง idempotency-key, ปุ่มถัดไปไม่ผูก hasNext, ไม่ส่ง status ไป backend, select สถานะไม่มีชื่อ) รอบแรกมี 2 แบบที่ "หลุด" แต่เป็นความผิดของวิธีทำพังเอง: แบบหนึ่ง build ไม่ผ่านแล้วสคริปต์รัน dist เก่า (แก้สคริปต์ให้หยุดเมื่อ build ล้ม) อีกแบบเปลี่ยน label เป็น span ซึ่งชื่อที่อ่านออกเสียงไม่เปลี่ยน จึงเขียนใหม่ทั้งสองแบบ
- แก้ mock: เคลมที่สร้างใหม่เดิมถูกต่อท้ายลิสต์ (ไปโผล่หน้าสุดท้าย) ทั้งที่ backend เรียงใหม่สุดก่อน เจอตอนเขียนเทสต์ "สร้างเคลมแล้วอยู่ใบแรก" แก้เป็น `unshift` และเพิ่ม `POST /__reset` (คืนข้อมูลตั้งต้น 23 รายการ) กับ env `PORT` (นี่เป็นความผิดพลาดของ mock ไม่ใช่ของ backend จริง)
- workflow อยู่ที่ `ci/day25.yml` (ให้วางที่ `.github/workflows/day25.yml`): unit test → ลง chromium → `npm run ci:e2e` → เก็บ `test-results` เมื่อล้ม ตรวจแค่ว่า YAML อ่านได้และรัน `CI=true npm run ci:e2e` บน Linux ผ่าน **ยังไม่เคยรันบน GitHub Actions จริง**
- `ng update`: ลองใน scratch แยก (ไม่แตะ day25-angular) Angular 22.2 → 22.3.0 ได้ unit 137/137, build ผ่าน (initial 333.44 kB เทียบ 331.83 kB), e2e 19/19 มี migration อัตโนมัติ 0 อัน รายละเอียดใน `notes/ng-update-22.3.txt` day25-angular ยังอยู่ที่ 22.2.x ตั้งใจ เพราะอยากให้การอัปเกรดเป็น commit แยกที่คุณตัดสินใจเอง
- เวอร์ชัน: `@playwright/test` 1.63.0 ตรงกับ chromium ที่ติดตั้งในเครื่อง cloud (อัปเป็น 1.65.1 ต้องลง browser ใหม่) `@axe-core/playwright` 4.14.1

## ยังไม่ได้ทำ (ต้องทำบนเครื่องของคุณ)
- **ยังไม่ได้ต่อ Day 8 backend จริง** e2e ทั้งหมดวิ่งกับ mock API ผลจริงจดตรงนี้: `[จดหลังรัน]` ทำตามขั้นตอนเดิมของ Day 24 (`run-all.ps1 -Keep` แล้วรัน server ตั้ง `BACKEND_URL`) จะลองรัน e2e กับ backend จริงต้องแก้ `webServer`/`API_PORT` และชุดข้อมูล เพราะเทสต์พึ่งข้อมูลตั้งต้น 23 รายการของ mock และ endpoint `/__reset` `/__fail` ที่ backend จริงไม่มี
- ไม่ได้รัน e2e ใน VM บนเครื่องคุณ: `npx playwright install chromium` ถูกบล็อกโดย allowlist ของเครือข่าย (403 จาก cdn.playwright.dev) ใน VM ตรวจได้แค่ `npm ci` + unit 137/137 + build (331.83 kB) ส่วน e2e รันได้เฉพาะบนเครื่อง cloud
- ยังไม่ได้รัน `npx playwright test` บน Windows (`npx playwright install` ครั้งแรกบน Windows ยังไม่เคยลอง)
- ต้องวาง `ci/day25.yml` ที่ `.github/workflows/` เอง แล้วดูผลหน้า Actions ของ `day25.yml` (และ `day20.yml` ที่ยังไม่ได้ดู) จดผลตรงนี้: `[จดหลังดูหน้า Actions]`
- e2e ทดสอบเฉพาะ chromium ไม่ได้ลอง firefox/webkit และหน้าจอมือถือ
- a11y ตรวจแค่ที่ axe ตรวจอัตโนมัติได้ ไม่ได้ลองใช้แป้นพิมพ์/screen reader จริง
- ข้อจำกัดของหน้า list ที่ค้างจาก Day 23 และ allowedHosts ที่ใส่แค่ localhost ของ Day 24 ยังเหมือนเดิม
- ข้อมูลทั้งหมดเป็นข้อมูลปลอม (`FAKE-`)

## วิธีรัน
```
cd claim-list-ng && npm ci
npx playwright install chromium          # ครั้งแรกเท่านั้น (ต้องตรงกับเวอร์ชันใน package-lock)
npm run ci:e2e                           # build production (SSR) แล้วรัน e2e + axe
npx playwright test e2e/a11y.e2e.ts      # เฉพาะ accessibility (ต้อง build ไว้แล้ว)
npm run test:ci                          # unit test (vitest) 137 ข้อ
```
