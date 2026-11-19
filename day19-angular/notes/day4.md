# Angular วันที่ 4 (Day 19): routing และฟอร์ม

## สิ่งที่ทำ
- Route: `/` → `/claims`, `/claims` (ClaimList), `/claims/new` (ฟอร์ม, มี guard), `/claims/:id` (รายละเอียด) ตั้งใน `app.routes.ts`
  - `claims/new` ต้องอยู่ก่อน `claims/:id` ไม่งั้นคำว่า "new" ถูกอ่านเป็น id (มีเทสต์กันไว้)
  - หน้าฟอร์มกับหน้ารายละเอียดเป็น lazy (`loadComponent`) `ng build` แยก chunk ให้จริง: claim-form 46.13 kB, claim-detail 2.91 kB, ส่วน initial 279.85 kB
- `provideRouter(routes, withComponentInputBinding())` ทำให้ `:id` ใน URL เข้า `input.required<string>()` ชื่อ `id` ของ ClaimDetail ได้เลย ไม่ต้องใช้ ActivatedRoute
- ClaimDetail ใช้ `httpResource` (id เป็น signal เปลี่ยนแล้วยิงใหม่เอง) มี loading / 404 "ไม่พบเคลมนี้" / error อื่น
- ฟอร์มสร้างเคลม (reactive forms, `NonNullableFormBuilder`): patient_ref (required + pattern `FAKE-[A-Z0-9]{3,12}`), ยอดเงินเป็นบาท (required, min 0.01, ทศนิยมไม่เกิน 2 ตำแหน่ง, ไม่เกิน 10,000,000 บาท) แล้วแปลงเป็นสตางค์ก่อนส่ง
- Validator ที่เขียนเอง 2 ตัวใน `validators.ts`: `maxBaht(limit)` และ `maxTwoDecimals` (คืน `null` ถ้าผ่าน คืน object ถ้าไม่ผ่าน)
- `POST /claims` ต้องมี header `Idempotency-Key` (backend ตอบ 400 ถ้าไม่มี): สร้าง key ครั้งเดียวต่อฟอร์ม กดซ้ำหรือ retry ข้อมูลเดิมใช้ key เดิม ถ้าแก้ข้อมูลให้ key ใหม่ ไม่งั้น backend ตอบ 409
- Guard `authGuard` (functional `CanActivateFn`) อ่านสถานะจาก `AuthService` (signal จำลอง) ถ้ายังไม่ล็อกอินคืน `UrlTree` ไป `/claims?login=required` ปุ่ม "ล็อกอิน (จำลอง)" บน navbar สลับสถานะ

## สิ่งที่แตกต่างจากแผน (ตั้งใจ)
- แผนให้ฟอร์มมี "ประเภท" แต่ backend ของ Day 8 ใช้ `DisallowUnknownFields` และรับแค่ `patient_ref` กับ `amount` ถ้าส่งฟิลด์เพิ่มจะได้ 400 จึงไม่ใส่ช่องประเภท
- ยังไม่ได้ทำ Signal Forms (อ่านแนวคิดอย่างเดียวตามแผน)

## เทียบกับที่รู้
- router เหมือน vue-router แต่ guard/lazy loading เป็นฟังก์ชันใน route config และ `withComponentInputBinding` คล้ายการส่ง route param เป็น props
- ฟอร์มเป็นตัวที่หนักที่สุด: โครงฟอร์มอยู่ใน class (FormGroup/FormControl) ไม่ใช่ใน template, validator เป็นฟังก์ชันที่คืน error object, state (touched/dirty/invalid) อยู่ที่ control ไม่ใช่ signal จึงให้ template เรียก `fieldError(name)`

## เทสต์
- 64 เทสต์ใน 11 ไฟล์ (เพิ่มจาก 26): validator, guard, service (get/create), detail, form (10 ข้อ), routes, app/nav
- baseline snapshot เพิ่ม 2 ชุด: ตาราง route (path/redirect/lazy/guard) และข้อความ error ของฟอร์มต่อค่าที่กรอก
- mutation 10 แบบ เทสต์จับได้ครบ: สลับลำดับ route, guard ผ่านเสมอ, ปัดสตางค์ด้วย floor, ไม่รีเซ็ต key, `>=` แทน `>` ใน maxBaht, ส่งบาทไม่แปลงสตางค์, ไม่กันกดซ้ำ, ไม่มี min(0.01), ไม่ส่ง header, route detail เป็น eager
- ช่องโหว่ที่รู้: เทสต์ routes สร้าง config ของตัวเอง จึงไม่จับกรณีลืมใส่ `withComponentInputBinding()` ใน `app.config.ts` (ตอนรันหน้าจริงหน้ารายละเอียดรับ :id ได้ จึงมั่นใจว่าตั้งค่าไว้ แต่ยังไม่มีเทสต์กันเรื่องนี้)
- เจอตอนเขียนเทสต์: `RouterTestingHarness.routeNativeElement` คือ element ของ component ที่ route อยู่เอง (ไม่ใช่ตัวห่อ) ต้องเช็ค `tagName` ไม่ใช่ `querySelector` หา component เอง

## ผลรันจริง (ng serve + mock API บน Linux cloud ไม่ใช่ Windows ไม่ใช่ backend จริง)
- `/` redirect ไป `/claims` เห็น 5 การ์ด กดรายละเอียด CLM-002 ไป `/claims/CLM-002` ยอด 32000 บาท จำนวนไฟล์ JS ที่โหลดเพิ่มจาก 1 เป็น 2 รีเฟรชแล้วยังอยู่หน้าเดิม
- `/claims/nope` ขึ้น "โหลดเคลมไม่สำเร็จ: ไม่พบเคลมนี้"
- ยังไม่ล็อกอินกด "สร้างเคลม" ได้ URL `/claims?login=required` ไม่เห็นฟอร์ม กดล็อกอิน (จำลอง) แล้วเข้า `/claims/new` ได้ และ chunk ของฟอร์มโหลดตอนนั้น (unique JS chunks 3)
- ส่งฟอร์มว่างได้ข้อความ กรอก patient_ref / กรอกยอดเงิน และไม่มี POST; patient_ref `abc` กับยอด 10000000.5 ขึ้นข้อความผิดรูปแบบกับเกิน 10,000,000 บาท
- ให้ mock ตอบ 503 กดส่งขึ้น "backend ตอบ HTTP 503" ยังอยู่หน้าเดิม เปิดคืนแล้วกดส่งซ้ำ สำเร็จ ได้ CLM-006 1234.56 บาท; สอง POST ใช้ Idempotency-Key เดียวกัน body `{"patient_ref":"FAKE-XY99","amount":123456}`
- console error มีแค่ 404 กับ 503 ที่ตั้งใจให้เกิด
- ภาพอยู่ใน `notes/screenshots/` (หน้าจอเป็น Linux ไม่ใช่ Windows)
