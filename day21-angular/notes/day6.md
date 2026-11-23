# Day 21 — โน้ตวันที่ 6 (interceptor)

## ทำอะไร
- เขียน interceptor 3 ตัว (เติม Idempotency-Key, แปลง error, retry GET เมื่อ 503) แล้วลงทะเบียนใน `provideHttpClient(withInterceptors([...]))`
- แยก `ApiError` + `readApiError()` ให้ component ไม่ต้องเช็ค `instanceof HttpErrorResponse` เอง
- เพิ่ม `POST /__fail?n=2` ใน mock เพื่อให้ตอบ 503 เป็นจำนวนครั้งที่กำหนดแล้วกลับปกติ (ไว้ดู retry)

## ที่ได้เรียนรู้
- อ่านจาก angular.dev (หน้า `withInterceptors` และ type `HttpInterceptorFn`): interceptor คือฟังก์ชัน `(req, next) => Observable<HttpEvent>` ใช้ `inject()` ข้างในได้ แก้ request ต้อง `req.clone()` หน้านั้นไม่ได้บอกลำดับการทำงานของหลายตัว
- ลำดับ: ทดลองเอง (`interceptor-order.spec.ts`) ได้ `A:req, B:req, B:res, A:res` สำหรับ `[A, B]` ตัวแรกใน array อยู่นอกสุด ดังนั้น retry ต้องอยู่หลัง error ใน array จึงจะเห็น `HttpErrorResponse` ดิบ ทดลองสลับแล้วเทสต์ล้ม
- `provideHttpClientTesting()` ไม่ได้ปิด interceptor ของ `provideHttpClient(withInterceptors(...))` เทสต์จึงใช้ providers ของ appConfig จริงได้ (ใส่ `RETRY_DELAY_MS = 0` เพื่อไม่ต้องรอ)
- จำนวน provider ใน appConfig ไม่เปลี่ยนเมื่อเพิ่ม `withInterceptors` (ยังเป็น 3) เทสต์นับ provider จึงไม่จับการลืมลงทะเบียน interceptor ต้องมีเทสต์ที่ดูพฤติกรรม
- retry ของ POST อันตราย: ถ้า key ไม่คงที่จะสร้างเคลมซ้ำ จึง retry เฉพาะ GET ส่วน POST ให้ผู้ใช้กดส่งซ้ำเอง (ฟอร์มถือ key เดิมไว้ backend จึงตอบ replay)

## เทียบกับที่รู้
- Vue/React ที่ใช้ axios: `interceptors.request.use` / `response.use` คล้ายกัน แต่ axios เรียก request interceptor ย้อนลำดับ ส่วน Angular ตามลำดับใน array ทั้งสองทาง (ผ่านการทดลองข้างบน ส่วนฝั่ง axios จำจากที่เคยใช้ ยังไม่ได้ทดลองซ้ำวันนี้)
- Angular ผูกกับ DI: interceptor `inject()` service อื่นได้เลย ไม่ต้องส่ง instance เข้ามาเอง

## ที่ยังค้าง
- ต่อ backend จริง (ดูขั้นตอนใน README) เพราะเครื่อง cloud รันไม่ได้
- ผลบน GitHub Actions ของ `day20.yml` (workflow นี้ยังไม่ครอบคลุม `day21-angular`)
