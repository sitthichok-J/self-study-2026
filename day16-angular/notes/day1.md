# บันทึกวันที่ 1

- ตอบคำถามเช็กลิสต์ในแผน: standalone component ไม่ต้องมี NgModule เพราะ component ประกาศสิ่งที่ใช้ใน template ผ่านอาเรย์ `imports` ของตัวเอง (`ClaimList` import `ClaimCard`)
- `input.required<Claim>()` ทำให้ compiler บังคับว่าคนใช้ต้องส่ง `claim` มาเสมอ ถ้าลืมจะ error ตอน build
- `output<string>()` เรียก `.emit()` เหมือน emits ใน Vue แต่ฝั่งแม่ผูกด้วย `(select)="..."` และรับค่าด้วย `$event`
