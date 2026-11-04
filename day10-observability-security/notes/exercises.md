# แบบฝึกหัด Day 10

## 1. โจมตีตัวเอง (IDOR)

ก่อนแก้ (Day 8): ไม่มี auth เลย ใครรู้ id ก็เปิดเคลมได้ และ idempotency key ใช้ร่วมกันทั้งระบบ

ลองแล้วเขียนเป็นเทสต์ใน `internal/adapter/http/security_test.go` และ `internal/service/obs_test.go`:

| ลองทำ | ผลหลังแก้ |
|---|---|
| เปิดเคลมของ alice ด้วย token ของ mallory | 404 body เหมือนเคลมที่ไม่มีอยู่ทุกตัวอักษร ไม่มีข้อมูลหลุด |
| เปิดซ้ำตอนเคลมอยู่ใน cache แล้ว | 404 (เช็คเจ้าของหลังอ่าน cache ด้วย) |
| list ของ alice แล้ว bob ขอ list ต่อ | bob ไม่เห็นของ alice (cache แยกตามผู้ใช้ ทดสอบทั้งกรณี staff เรียกก่อน) |
| ส่ง `owner_id` ใน body | 400 (field ที่ไม่รู้จัก) เจ้าของมาจาก token เท่านั้น |
| bob ใช้ Idempotency-Key เดียวกับ alice | 201 เคลมใหม่ของ bob (ก่อนแก้: 409 หรือได้เคลมของ alice) |
| ไม่มี token / token แก้ payload / alg=none / ลายเซ็นว่าง / หมดอายุ / อายุ 1 วัน / issuer, audience ผิด | 401 ข้อความเดียวกันหมด |

ทุกครั้งที่ถูกปฏิเสธมี log `security_event` และ metric นับ

ผลบน Postgres จริง: ทดสอบแล้วใน `run-all` (รอบที่ 3 ผ่าน 57/57) ข้อ 401/404, idempotency ข้ามผู้ใช้ และ list แยกตามผู้ใช้ ผ่านบน stack จริง ผลอยู่ใน `notes/run-results.md`

## 2. สแกนช่องโหว่

รันแล้ว: Trivy เจอ dependency 17 รายการ govulncheck เจอ pgx กับ go-redis ที่โค้ดเรียกใช้จริง 2 ตัว แก้โดยขึ้น Go 1.26.10 และอัป dependency แล้วสแกนซ้ำสะอาด (gitleaks เหลือ 2 จุดที่คาดไว้) ตารางช่องโหว่และวิธีแก้อยู่ใน `docs/vulnerability-fixes.md` ผลดิบใน `notes/security-scan.md`

ส่วนเทสต์ static (สแกน private key, token แบบ JWT, key ของ AWS/GitHub/Slack, hardcode secret) ยังอยู่ใน `tests/test_infra.py` ใช้ควบคู่กับตัวสแกนจริง

## 3. ทำให้ช้าแล้วหาจุดช้าด้วย trace

`TestSlowRepoShowsUpAsTheLongestSpan` ครอบ repo ด้วยตัวหน่วง 60 ms ที่ `Get` แล้วเรียก `Get` ผ่าน service ภายใต้ span ของ request เรียง span ลูกตามเวลา ตัวที่นานสุดคือ `repo.Get` (60 ms ขึ้นไป) ตรงกับที่ใส่ไว้

ดูใน Jaeger จริงแล้ว: trace ของ `POST /claims` มี span `repo.Create`, `outbox publish` และ `consume` ใน trace เดียวข้าม claims-api กับ claims-worker (รูปใน `notes/screenshots/05-jaeger-trace-detail.jpg`) เคสช้าโดยจงใจยังทำเป็นเทสต์เท่านั้น ยังไม่ได้ใส่ตัวหน่วงในระบบจริง

## สิ่งที่เรียนรู้ระหว่างทำ

- เทสต์ที่ผ่านไม่ได้บอกว่าเทสต์ดี ต้องแก้โค้ดจงใจ (ตัดเช็ค JWT, ตัดเช็คเจ้าของ, ถอด header) แล้วดูว่าเทสต์แดงจริง รอบแรกมี 5 จุดที่คอมไพล์ไม่ผ่านเพราะเขียนการแก้ผิด ไม่ได้ถูกจับ เลยแก้ให้คอมไพล์แล้วรันใหม่
- Prometheus `increase()` มองไม่เห็น series ที่เกิดมาที่ 1
- ใส่ metric label เป็น path ดิบ = series ไม่จำกัด ใช้ pattern ของ `ServeMux` (`r.Pattern`) แต่ต้องให้ middleware ครอบ mux โดยตรง ไม่งั้นอ่านไม่ได้
- 404 กับ "ไม่ใช่ของคุณ" ต้องหน้าตาเหมือนกันทุกบิต รวมถึง body
