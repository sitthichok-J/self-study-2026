# Day 24 — Angular วันที่ 9 (รอบสอง): SSR + hydration

โค้ดอยู่ใน `claim-list-ng/` โน้ตอยู่ใน `notes/day9.md` (ต่อจาก Day 23 คัดลอกโปรเจกต์มาแล้วเพิ่ม SSR)

## สถานะ
- เทสต์ 137/137 ผ่าน (23 ไฟล์) รันบน Linux (cloud และ VM บนเครื่อง ยังไม่เคยรัน `ng test` บน Windows)
- เพิ่ม `@angular/ssr` กับ `@angular/platform-server` + express: server render หน้า `/claims` และ `/claims/:id` ตอนมี request, `claims/new` กับ `claims/new-signal` เป็น client-only (guard ล็อกอิน server ไม่รู้ว่าใครล็อกอิน) ไม่ใช้ prerender เพราะข้อมูลมาจาก API สด
- `ng add @angular/ssr` ใช้ตรง ๆ ไม่ได้: มันจะลง ssr 22.3.0 ซึ่งต้องการ core 22.3 แต่โปรเจกต์ล็อกที่ 22.2.6 จึงลง `@angular/ssr@22.2.8` + `platform-server@22.2.6` เองด้วย `--legacy-peer-deps` แล้วรัน schematic ทับ (เปลี่ยน specifier ใน package.json เป็น ^22.2.x ตามที่ schematic เขียน) ถ้าอัปเกรดทั้งชุดเป็น 22.3 ภายหลัง ต้องไล่ดู lock อีกที
- `provideClientHydration()` ใน `app.config.ts` + `@defer (hydrate on interaction)` ครอบ pager ของหน้า list (incremental hydration): HTML จาก server มี pager ครบ แต่ยังไม่ผูก event จนกว่าจะมีคนแตะ
- ฝั่ง server เรียก `/api/claims` (URL สัมพัทธ์) เข้า origin ของ request เอง จึงเพิ่ม `src/api-proxy.ts` ใน `server.ts` ให้ส่ง `/api/*` ต่อไป `BACKEND_URL` (ค่าเริ่มต้น `http://localhost:8080`) ส่ง header แค่ content-type / accept / idempotency-key (ไม่ส่ง cookie)
- `angular.json` ตั้ง `security.allowedHosts: ["localhost"]` ถ้าไม่ตั้ง (ค่าเริ่มต้นของ schematic คือ `[]`) server ตอบ 400 ทุก request (ป้องกัน SSRF) ตอนขึ้นเครื่องจริงต้องใส่ชื่อโฮสต์จริง
- เทสต์ใหม่: `api-proxy.spec.ts` 5 ข้อ (ต่อ URL, header ที่ส่งต่อ, status ผ่านตรง, 502 เมื่อ backend ล่ม, baseline header), `app.routes.server.spec.ts` 4 ข้อ (baseline render mode, route ที่มี guard ต้องเป็น Client, ห้าม Prerender, ทุก route มี entry), เพิ่มใน `app.config.spec.ts` (provider ของ hydration, จำนวน provider 3→4), `claim-list.spec.ts` (pager อยู่ใน `@defer` + select ขนาดหน้ามี attribute `selected`)
- ทำให้พัง 15 แบบ จับได้ 15 จาก 15 (ลบ hydration, route ฟอร์มเป็น Server x2, `**` เป็น Prerender, ลบ entry ของ `claims/:id`, proxy ส่ง cookie / ไม่ส่ง idempotency-key / status เป็น 200 / error เป็น 500 / ต่อ URL ผิด / ไม่ส่ง content-type / ส่ง body ตอน GET / ไม่ตอบ 502, ลบ `@defer`, select ใช้ property แทน attribute) ทุกแบบถูกจับตั้งแต่รอบแรก ไม่มีแบบที่หลุด
- เจอบั๊กจากภาพหน้าจอที่ปิด JavaScript ไม่ใช่จากเทสต์: select "ต่อหน้า" ใน HTML จาก server แสดง 5 ทั้งที่รายการเป็นขนาด 10 เพราะ `[selected]` เป็น property ที่ server เขียนลง HTML ไม่ได้ (hydrate แล้วค่อยถูก) เปลี่ยนเป็น `[attr.selected]` และเพิ่มเทสต์
- bundle (production, ฝั่ง browser): initial 290.30 kB (Day 23) → 321.58 kB หลังเปิด hydration (+31.28 kB) → 331.83 kB หลังเพิ่ม `@defer (hydrate on interaction)` (+10.25 kB) ตัวเลขสองขั้นวัดจากการ build จริงทีละขั้น ส่วนสาเหตุภายในไม่ได้ไล่ต่อ งบ warning เดิม 300 kB เตือนขึ้นมา จึงขยับ warning เป็น 340 kB (error ยังที่ 350 kB เหลือที่ว่างแค่ ~18 kB)
- รันจริง (production build + mock API) ด้วย `notes/ssr-check.mjs` ผลเต็มอยู่ใน `notes/ssr-check-log.txt` (สคริปต์ชี้ไป server พอร์ต 4300 ตอนรัน `PORT=4300`):
  - HTML ดิบของ `/claims` มีการ์ด 10 ใบ + pager + ข้อมูลฝังใน `ng-state` ส่วน `/claims/new` ได้แค่ shell (541 bytes)
  - ปิด JavaScript ก็เห็นรายการ
  - เปิด JavaScript: element การ์ดใบแรกเป็น node เดิมหลัง hydrate (ไม่ถูกสร้างใหม่) ไม่มี request `/api` ตอนโหลดหน้า (ใช้ข้อมูลที่ฝังมา) ไม่มี console error/warning
  - หน่วง `main.js` 2.5 วินาทีแล้วคลิก "ถัดไป" ก่อน hydrate: หลังโหลดเสร็จหน้าเปลี่ยนเป็นหน้า 2 และยิง `offset=10` หนึ่งครั้ง (event replay ทำงาน)
  - chunk lazy: เปิด `/claims` ได้ `main-*.js` ไฟล์เดียว, กด "สร้างเคลม" หลังล็อกอินโหลดเพิ่ม `chunk-COUH_0Mw.js`, `chunk-Gjte3idg.js`, เปิด `/claims/CLM-003` ตรง ๆ โหลด `chunk-Bavni-ed.js` ของหน้า detail
  - dev build: console ขึ้น "Angular hydrated 12 component(s) and 286 node(s), 0 component(s) were skipped. 1 defer block(s) ... incremental hydration" ไม่มี NG05xx
- ความเสถียร: ตรวจ d.ts ที่ติดตั้งของ API ที่ใช้ใน Day 24 ไม่ครบ (ไม่ได้ไล่ `@publicApi` ของ incremental hydration / `RenderMode` ทีละตัว) จึงยังไม่ยืนยันว่าเสถียรหรือยังเป็น developer preview

## ยังไม่ได้ทำ (ต้องทำบนเครื่องของคุณ)
- **ยังไม่ได้ต่อ Day 8 backend จริง** (เครื่อง cloud ไม่มี Docker และโหลด Go module ไม่ได้) ทุกอย่างผ่าน mock ตัว proxy ทดสอบกับ fetch ปลอมและ mock เท่านั้น ผลจริงจดตรงนี้: `[จดหลังรัน]`
  1. `cd day08-backend` แล้วรัน `scripts\run-all.ps1 -Keep`
  2. `cd day24-angular\claim-list-ng` → `npm ci` → `npm run build` → `node dist/claim-list-ng/server/server.mjs` (พอร์ต 4000 ตั้ง `PORT=` ได้) แล้วเปิด `http://localhost:4000/claims`
  3. ดู view-source ว่ามีการ์ดอยู่ใน HTML และเปิดแท็บ Network ดูด้วยตาว่าไม่มี request `/api/claims` ตอนโหลดหน้า (ในรอบนี้ตรวจด้วยสคริปต์ Playwright เท่านั้น ไม่ได้ดู Network tab ด้วยตา)
- ไม่ได้ทดลองบน Windows (`npm ci`, build, รัน server)
- `allowedHosts` ใส่แค่ `localhost` ยังไม่ได้ตั้งสำหรับโฮสต์จริง
- ไม่มี workflow ของ day24 และผล Actions ของ `day20.yml` ก็ยังไม่ได้ดู
- ข้อจำกัดของหน้า list ที่ค้างจาก Day 23 ยังเหมือนเดิม (ไม่มี total, query ไม่ผูก URL, รายการหายตอนโหลด, offset เลื่อน)
- สถานะล็อกอินยังเป็นของ browser (signal ใน memory) ฝั่ง server ไม่รู้จึงปล่อยหน้าฟอร์มเป็น client-only ถ้าทำ auth จริง (cookie/token) ต้องคิดเรื่อง SSR ใหม่
- ข้อมูลทั้งหมดเป็นข้อมูลปลอม (`FAKE-`)

## วิธีรัน
```
node mock-api/server.mjs                       # port 8080 (23 รายการ)
cd claim-list-ng && npm ci
npm run build && node dist/claim-list-ng/server/server.mjs   # SSR ที่ http://localhost:4000/claims (BACKEND_URL ปรับปลายทางได้)
npx ng serve                                   # dev server เดิม (ใช้ proxy.conf.json) ยังไม่ได้ลองว่า ng serve render ฝั่ง server ได้ถูกต้องในรอบนี้
npm run test:ci                                # เทสต์ทั้งหมดแบบไม่ watch
npm run ci                                     # เทสต์ แล้ว build production
```
