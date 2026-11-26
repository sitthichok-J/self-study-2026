# Day 24 — โน้ตวันที่ 9 (SSR + hydration)

## ทำอะไร
- เพิ่ม SSR ให้โปรเจกต์ Claim List: server เรนเดอร์ `/claims` กับ `/claims/:id` ตอนมี request แล้วฝังข้อมูลที่ดึงมาไว้ใน HTML (`ng-state`) ฝั่ง browser hydrate ต่อโดยใช้ DOM เดิมและไม่ยิง API ซ้ำ
- pager ของหน้า list ใช้ incremental hydration (`@defer (hydrate on interaction)`)

## เลือก render mode ราย route
| route | mode | เหตุผล |
|---|---|---|
| `''`, `claims`, `claims/:id`, `**` | Server | ข้อมูลเคลมเปลี่ยนตลอด prerender ตอน build จะได้ข้อมูลเก่า และ `:id` ไม่รู้ล่วงหน้า |
| `claims/new`, `claims/new-signal` | Client | อยู่หลัง `authGuard` ที่อ่านสถานะล็อกอินใน memory ของ browser server ไม่รู้ จึงส่งแค่ shell |

schematic ตั้งค่าเริ่มต้นเป็น `** → Prerender` ซึ่งไม่เหมาะกับแอปนี้ (build ได้ "Prerendered 0 static routes" เพราะ route มีพารามิเตอร์/redirect) จึงเขียนตารางใหม่และมีเทสต์ล็อกไว้

## สิ่งที่ต้องแก้เพิ่มนอกจาก `ng add`
- เวอร์ชัน: `ng add @angular/ssr` จะลง 22.3.0 ชนกับ core 22.2.6 ที่ล็อกไว้ ลง 22.2.8 เองแล้วค่อยรัน schematic
- `allowedHosts`: ค่าเริ่มต้น `[]` ทำให้ทุก request ได้ 400 ("Header host ... is not allowed") ต้องใส่ `localhost` (ตอนขึ้นเครื่องจริงใส่โฮสต์จริง)
- URL ของ API: โค้ดใช้ `/api/claims` (สัมพัทธ์) ตอน SSR HttpClient ยิงเข้า origin ของ request เอง จึงต้องมี proxy `/api` ใน `server.ts` (ที่ `ng serve` ใช้ `proxy.conf.json` ซึ่ง server ที่ build แล้วไม่ใช้) แยกเป็น `api-proxy.ts` เพื่อเทสต์
- select ขนาดหน้า: `[selected]` (property) ไม่ถูกเขียนลง HTML จาก server → ก่อน JS มาถึงเห็นเลือก 5 ทั้งที่รายการเป็น 10 เปลี่ยนเป็น `[attr.selected]`

## ที่ได้เรียนรู้ / ที่วัดได้
- HTTP transfer cache มากับ `provideClientHydration()` อัตโนมัติ: GET ที่ server ดึงถูกฝังใน HTML และ browser ไม่ยิงซ้ำ (เห็นว่าไม่มี request `/api` ตอนโหลดหน้า ด้วยสคริปต์ Playwright) `httpResource` ผ่าน HttpClient จึงได้ประโยชน์เหมือนกัน
- event replay: คลิก "ถัดไป" ก่อน hydrate (หน่วง main.js 2.5 วินาที) แล้วหน้าเปลี่ยนเป็น 2 หลังโหลดเสร็จ ยิง `offset=10` ครั้งเดียว
- hydration ใช้ DOM เดิม: node การ์ดใบแรกก่อนกับหลัง hydrate เป็นตัวเดียวกัน dev mode ขึ้นข้อความ "hydrated 12 component(s) ... 0 skipped" ไม่มี NG05xx
- ราคา: initial bundle 290.30 → 321.58 kB (เปิด hydration) → 331.83 kB (เพิ่ม `@defer` hydrate) incremental hydration ครอบแค่ pager ปุ่มสองปุ่มแต่ราคา ~10 kB ในแอปเล็กแบบนี้ไม่คุ้ม ถ้าจะใช้จริงควรครอบส่วนใหญ่และหนักกว่านี้ หรือไม่ใช้เลย
- เทสต์ใน jsdom: block ใน `@defer` เรนเดอร์ตอน idle จึงต้อง `detectChanges()` หลัง `whenStable()` อีกรอบก่อนอ่าน DOM (ไม่งั้น `disabled` ยังไม่ถูกอัปเดต) และเทสต์ว่ามี `@defer` ใช้ `deferBlockBehavior: Manual` + `getDeferBlocks()`
- เทสต์ไม่เจอบั๊ก select แต่ภาพตอนปิด JavaScript เจอ: ทั้งชุดผ่านเพราะ jsdom อ่าน property หลัง Angular ทำงานแล้ว ไม่ได้ดูสิ่งที่ server เขียนลง HTML
- เทสต์ "มี provider ของ hydration" อ่านชื่อ token ภายใน (`STATE_CACHE_ACTIVE`) เปราะ ถ้า Angular เปลี่ยนชื่อ เทสต์จะแดงทั้งที่ config ถูก (เขียนไว้ในคอมเมนต์ของเทสต์)

## ยังไม่รู้ / ไม่ได้ทดสอบ
- ไม่ได้ลองกับ backend จริง และไม่ได้ดู Network tab ด้วยตา (ใช้สคริปต์ Playwright ดักแทน)
- ยังไม่ได้ลองกรณี hydration mismatch (เช่นแก้ DOM ตรง ๆ) หรือ `ngSkipHydration`
- ไม่ได้วัดเวลา (TTFB / LCP) เปรียบเทียบ client-only กับ SSR จึงยังไม่รู้ว่าเร็วขึ้นแค่ไหน มีแต่หลักฐานว่าเนื้อหาอยู่ใน HTML แรก
- `ng serve` กับ SSR ในโหมด dev ไม่ได้ลองในรอบนี้

## เทียบกับที่รู้
- React: คล้าย `renderToString` + `hydrateRoot` หรือ Next.js ที่ทำให้ แต่ใน Angular เป็นส่วนของ framework เอง ฝั่ง React จำจากที่เคยใช้ ยังไม่ได้ลองซ้ำวันนี้
- Vue/Nuxt: `useFetch` ฝังข้อมูลใน payload คล้ายกับ transfer cache
