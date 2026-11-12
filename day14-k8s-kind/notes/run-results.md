# Day 14 ผลที่รันจริง

ผลทั้งหมดมาจาก GitHub Actions (workflow `day14`, runner Ubuntu, kind ที่ `helm/kind-action` ติดตั้ง) ยังไม่มีผลจากเครื่อง Windows ของตัวเอง ตัวเลขด้านล่างคือที่เห็นใน log ไม่ได้เขียนจากความคาดหมาย

## รอบที่ 1 (commit eb7ae7a): แดง ที่ job kind (7 นาที 40 วินาที)

- job static ผ่าน (เทสต์ static + `trivy config` ของ manifest แอป)
- build image 3 ตัว, `kind load`, สร้าง Secret/ConfigMap, `kubectl apply -k` ผ่านหมด
- `kubectl wait` ของ job `kafka-init` และ `migrate` หมดเวลา 300 วินาที
- สถานะ pod ตอนนั้น: kafka `Error` (restart 3 ครั้ง), kafka-init `Error`, worker `CrashLoopBackOff`, api `Running` แต่ restart 3 ครั้ง, migrate `Completed` หลัง restart 2 ครั้ง
- log ตอนล้มยาวเกินจน GitHub ตัดทิ้ง จึงยังไม่ได้อ่านข้อความ error ของ kafka โดยตรง สาเหตุที่ตั้งสมมติฐานไว้คือ `KAFKA_CONTROLLER_QUORUM_VOTERS=1@kafka:9093` ให้ pod ต่อหาตัวเองผ่านชื่อ Service ซึ่งใช้ไม่ได้จน pod Ready (pod จะ Ready ไม่ได้ถ้า controller ยังไม่เจอตัวเอง)

## รอบที่ 2 (commit 629614b): เขียว (3 นาที 18 วินาที)

แก้เป็น `1@localhost:9093` อย่างเดียว (กับทำให้ขั้นแสดงสถานะตอนล้มพิมพ์สั้นลง) ผลในขั้น deploy และ smoke:

```
job.batch/kafka-init condition met
job.batch/migrate condition met
deployment "postgres" / "valkey" / "kafka" / "api" / "worker" / "gateway" / "web" successfully rolled out
deploy ok
PASS gateway ตอบ และไม่มี token ถูกปฏิเสธ 401
PASS login alice ผ่าน Kong -> 200 + access token
PASS GET /api/claims พร้อม token -> 200
PASS POST /api/claims -> สร้างเคลม (201)
PASS worker ตัดสินเคลมเป็น APPROVED (api -> kafka -> worker -> postgres)
PASS web ส่ง index.html
PASS CORS preflight จาก web ผ่าน
smoke ok: 7/7
```

หมายเหตุ: ยืนยันได้แค่ว่าการแก้ quorum ทำให้ผ่าน ยังไม่ได้ยืนยันด้วย log ว่าสาเหตุของรอบแรกคือข้อนี้จริง (เพราะ log ถูกตัด)

## สิ่งที่รอบนี้ยืนยันแล้วเพิ่มจาก static test

- image web ที่ตั้ง `readOnlyRootFilesystem: true` พร้อม emptyDir ที่ `/tmp` ขึ้นและผ่าน readiness ได้จริง
- api และ worker รันเป็น non-root พร้อม root filesystem อ่านอย่างเดียวได้จริง
- Kong ที่ได้แค่ public key (render จากเทมเพลต Day 13) ตรวจ JWT ของ api ได้จริง (login แล้วเรียก `/api/claims` ผ่าน)
- `trivy config` ของ manifest แอปผ่านใน job static

## ยังไม่ได้ลอง

- แบบฝึกหัดใน `notes/exercises.md` ทั้งหมด (OOMKilled, rollout, rollback, ฆ่า pod, valkey restart กับ token ที่ถูกเพิกถอน)
- รันบนเครื่อง Windows ผ่าน Docker Desktop
- เปิดหน้าเว็บด้วยเบราว์เซอร์และ login จริง (smoke ตรวจแค่ index.html กับ API)
- รันบน pull request
