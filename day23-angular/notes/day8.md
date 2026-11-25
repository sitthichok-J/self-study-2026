# Day 23 — โน้ตวันที่ 8 (store + httpResource + แบ่งหน้า)

## ทำอะไร
- ย้าย state ของหน้า list (status, ขนาดหน้า, หน้า) ไปไว้ใน `ClaimListStore` แล้วให้ `httpResource` ยิง request ตาม query ที่ computed ออกมา
- ส่งตัวกรองสถานะและการแบ่งหน้าไปให้ backend แทนการโหลด 100 รายการมากรองเอง

## ก่อน / หลัง (หน้า list)
| เรื่อง | Day 22 (service + Observable) | Day 23 (store + httpResource) |
|---|---|---|
| โหลดข้อมูล | `service.list().subscribe` + `takeUntilDestroyed` เขียน next/error เอง | `httpResource` ยิงเองเมื่อ signal ใน callback เปลี่ยน |
| loading / error | signal `loading`, `error` ตั้งมือในทุกทาง | `isLoading()`, `error()` มาจาก resource |
| กรองสถานะ | ใน `computed` ฝั่ง client | ส่ง `status` ไป backend |
| จำนวนข้อมูลที่โหลด | 100 รายการเสมอ | ครั้งละ pageSize+1 |
| request เก่าที่ค้าง | ไม่ได้ดูแล | ถูกยกเลิกเมื่อ query เปลี่ยน (มีเทสต์ `cancelled`) |
| บรรทัด claim-list.ts | 96 | 80 (+ store 55 + resource 30) |

บรรทัดรวมเพิ่มขึ้น (ts 96 → 165) แต่ส่วนที่ซับซ้อน (query, หน้า, ขอบเขต) แยกออกมาเทสต์ได้ตรง ๆ ที่ store โดยไม่ต้องเรนเดอร์ component

## ที่ได้เรียนรู้
- อ่านจาก d.ts ที่ติดตั้ง: `httpResource` ติด `@publicApi 22.0` (Angular 22.2.6)
- เทสต์: ห้าม `await whenStable()` ตอนที่ resource ยังมี request ค้างอยู่ (นับเป็น pending task รอไม่จบ) ให้ `TestBed.tick()` แล้ว `expectOne` และ flush ก่อนค่อยรอ ตอนที่ resource error อ่าน `value()` ไม่ได้ (โยน error) ต้องเช็ค `hasValue()` ก่อน
- ตั้ง signal สองตัวพร้อมกัน (status กับ page) แล้วได้ request เดียว เพราะ resource อ่านค่าตอน tick ไม่ใช่ทันทีที่ set (ผ่านเทสต์ แต่ยังไม่ได้อ่านเอกสารยืนยันว่าเป็นพฤติกรรมที่รับประกัน)
- ตั้ง status เป็นค่าเดิมซ้ำ ไม่ยิง request ใหม่ (ทดลองแล้ว)
- ตัวกรองที่อยู่ในกิ่ง `@else` ของ `@if (loading())` จะถูกสร้างใหม่ทุกครั้งที่โหลด ค่าใน select เด้งกลับเป็นค่าเริ่มต้น ย้ายออกมานอกกิ่งแล้ว (เทสต์ตรวจว่าเป็น element เดิม)
- เทสต์ผ่านหมดแต่ UI ยังผิดได้: select ขนาดหน้าแสดง 5 ทั้งที่ใช้ 10 เทสต์ไม่เคยดูค่าเริ่มต้น เจอเพราะเปิดดูภาพหน้าจอ
- ทำให้พังแล้วเทสต์ไม่จับ = binding ที่ไม่จำเป็น (`[selected]` ของ select สถานะ) จึงลบทิ้ง มากกว่าเพิ่มเทสต์ให้มัน

## เทียบกับที่รู้
- React: คล้าย TanStack Query / SWR ที่ query key เปลี่ยนแล้ว fetch ใหม่และยกเลิกของเก่า ส่วน `httpResource` ผูกกับ signal และไม่มี cache ข้าม query ในที่นี้ (ฝั่ง React จำจากที่เคยใช้ ยังไม่ได้ลองซ้ำวันนี้)
- Vue: `watch` + `fetch` หรือ `useFetch` ที่ ref เปลี่ยนแล้วยิงใหม่ ใกล้เคียงกับแนวคิดนี้

## ที่ยังค้าง
- ต่อ backend จริง ดูว่า limit pageSize+1 และ ORDER BY ทำงานตามคาด (ดู README)
- query ยังไม่ผูก URL (รีเฟรชแล้วกลับหน้า 1), ไม่มี total, ระหว่างโหลดรายการหาย
- ยังไม่ได้ลองเทียบกับ `rxResource` หรือ store library (NgRx Signals) เป็นแค่ service ธรรมดา
