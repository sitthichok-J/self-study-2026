# Angular วันที่ 3 (Day 18): Service, DI, HttpClient, loading/error, httpResource

## สิ่งที่ทำ
- `ClaimService` (`providedIn: 'root'`) ใช้ `inject(HttpClient)` เรียก `GET /api/claims` ส่ง `limit/offset/status` ผ่าน `HttpParams`
- `provideHttpClient()` ใส่ใน `app.config.ts`
- interface `Claim` ตาม contract ของ Day 8 backend (amount เป็นสตางค์ → แปลงเป็นบาทตอนแสดงผลด้วย `satangToBaht`)
- `ClaimList` มี 3 สถานะ: กำลังโหลด / error (มีปุ่มลองใหม่) / มีข้อมูล ใช้ `takeUntilDestroyed` กัน subscribe ค้าง
- `claim-resource.ts` ทำแบบเดียวกันด้วย `httpResource()` ไว้เทียบ (ยังไม่ได้ต่อเข้า component)
- CORS: ใช้ dev proxy (`proxy.conf.json`: `/api` → `localhost:8080` ตัด prefix `/api`)

## เทียบ Observable กับ httpResource
- Observable: เขียน subscribe/next/error + state (loading/error/claims) เองเป็น signal ควบคุมได้ละเอียด แต่โค้ดยาวกว่า
- httpResource: ได้ `value()/isLoading()/error()` ให้เลย เปลี่ยน signal ตัวกรองแล้วยิง request ใหม่เอง
- ข้อที่เจอตอนเขียนเทสต์: อ่าน `value()` ตอน resource อยู่ในสถานะ error จะ **โยน error** ต้องเช็ค `hasValue()` หรือ `error()` ก่อน

## switchMap (ที่ต้องอธิบายได้)
เมื่อมี event ใหม่ (เช่นเปลี่ยนตัวกรอง) มันยกเลิก request เก่าที่ยังค้างอยู่ แล้วสลับไปฟังตัวใหม่ ผลคือไม่มีกรณี response เก่าตอบช้ากว่าแล้วมาทับของใหม่ (httpResource ทำเรื่องนี้ให้เองข้างใน)

## เทสต์
- 26 เทสต์ (service 3, card 5, list 13, app 2, httpResource 3) + baseline snapshot ของ ClaimList
- เจอบั๊กในเทสต์ตัวเอง: `await whenStable()` ตอน request ยังค้างอยู่จะรอไม่จบ (HTTP ที่รออยู่นับเป็น pending task) → ใช้ `TestBed.tick()` แล้ว flush ก่อนค่อย settle
