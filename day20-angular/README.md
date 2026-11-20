# Day 20 — Angular วันที่ 5: เทสต์ build CI และปิดโปรเจกต์

โค้ดอยู่ใน `claim-list-ng/` โน้ตอยู่ใน `notes/day5.md` (ต่อจาก Day 19 โค้ดแอปไม่เปลี่ยน เพิ่มเฉพาะเทสต์ budget และ workflow)

## สถานะ
- เทสต์ 66/66 ผ่าน (12 ไฟล์) รันบน Linux (VM บนเครื่อง และ cloud) ยังไม่เคยรัน `ng test` บน Windows
- เพิ่ม `src/app/app.config.spec.ts` (2 เทสต์) อุดช่องโหว่ที่จดไว้ใน Day 19: เทสต์ routes ไม่จับกรณีลืม `withComponentInputBinding()` ใน `app.config.ts` ไฟล์ใหม่ใช้ providers ของ `appConfig` จริง
- ทำให้ app.config พัง 3 แบบ เทสต์จับได้ 3 จาก 3: ถอด `withComponentInputBinding()`, ถอด `provideHttpClient()`, เพิ่ม provider เกิน (แบบที่ 2 จับได้ด้วยเทสต์นับ provider เท่านั้น เพราะ `provideHttpClientTesting()` ให้ HttpClient อยู่แล้ว เทสต์ "inject ClaimService ได้" จึงอ่อนเกินไปและถูกลบทิ้ง)
- bundle budget ใน `angular.json`: initial เตือน 300kB / พัง 350kB (เดิม 500kB / 1MB) ตอนนี้ initial 279.85 kB ลองตั้ง error เหลือ 250kB แล้ว build ล้ม (`bundle initial exceeded maximum budget`) แล้วคืนค่าเดิม
- lazy chunk: claim-form 46.13 kB, claim-detail 2.91 kB (แยกจาก bundle หลัก)
- script ใหม่: `npm run test:ci` (ng test แบบไม่ watch) และ `npm run ci` (test:ci แล้ว build) ลองบน `npm ci` ใหม่ทั้งชุดใน VM Linux ผ่านทั้งสองขั้น
- รัน production build ผ่าน `ng serve --configuration production` กับ mock API แคปหน้าจอไว้ใน `notes/screenshots/` (รายการ, รายละเอียด, guard, mock ตอบ 503, สร้างเคลมสำเร็จ, ผลเทสต์, ผล build) ผลเหมือน Day 19 ทุกจุด: lazy chunk โหลดตอนเข้าหน้า, guard พากลับ `/claims?login=required`, ส่งซ้ำหลัง 503 ใช้ Idempotency-Key เดิมและ body เดิม
- มี deck สรุป (9 สไลด์) และร่างเมลแล้ว (เมลเป็นแบบร่าง ยังไม่ได้ส่ง)

## ยังไม่ได้ตรวจ
- **workflow `.github/workflows/day20.yml` ยังไม่เคยรันบน GitHub Actions จริง** ตอนเขียนไฟล์นี้ยังไม่ได้ push ที่ลองได้คือ YAML อ่านได้ และรันขั้นตอนชุดเดียวกันในเครื่อง หลัง push ให้ดูหน้า Actions แล้วจดผลตรงนี้ ข้อ "เทสผ่านทั้งในเครื่องและใน CI" ของแผนจึงยังไม่ถือว่าเสร็จจนกว่าจะเห็นผลบน Actions
- ผล: `[จดหลังดูหน้า Actions]`

## ข้อควรรู้
- **Day 8 backend (Go + Docker) ยังไม่ได้รัน** ใช้ `mock-api/server.mjs` จำลอง contract (Idempotency-Key, patient_ref FAKE-xxx, amount เป็นสตางค์, ห้ามมีฟิลด์อื่น) ข้อมูลเป็น FAKE-xxx อยู่ในหน่วยความจำ id เป็น CLM-xxx ไม่ใช่ uuid
- การตรวจ chunk และ header ตอนรันจริงทำด้วยสคริปต์ Playwright ที่ดัก request ไม่ได้ดู Network tab ด้วยตา
- ฟอร์มไม่มีช่อง "ประเภท" ตามแผน เพราะ backend ไม่รับฟิลด์นอกเหนือ patient_ref กับ amount
- ยังไม่ได้ทำ: Signal Forms, interceptor, SSR และ hydration, `ng update`

## วิธีรัน
```
node mock-api/server.mjs             # port 8080
curl -X POST localhost:8080/__fail   # สลับให้ตอบ 503 (ยิงซ้ำเพื่อปิด)
cd claim-list-ng && npm ci && npx ng serve   # เปิด http://localhost:4200 (กด "ล็อกอิน (จำลอง)" ก่อนเข้า /claims/new)
npm run test:ci                      # เทสต์ทั้งหมดแบบไม่ watch
npm run ci                           # เทสต์ แล้ว build production
```
