# สรุป 2 สัปดาห์ (21 ต.ค. – 4 พ.ย. 2026)

Sitthichok · แผนเรียนด้วยตัวเองระหว่างรอเข้าไซต์ลูกค้า · โค้ด: https://github.com/sitthichok-J/self-study-2026 · ข้อมูลทั้งหมดเป็นข้อมูลสมมติ

## ทำอะไรไปบ้าง

| วัน | เรื่อง | ผลงาน |
|---|---|---|
| 1–2 | Vue 3 | แอป Claim List แล้วต่อเป็น Claim Console (Router, Pinia, composable, เรียก API, เทสต์) |
| 3–5 | Go | พื้นฐานที่คนมักพลาด, concurrency (worker pool, pipeline, shutdown), REST API ต่อ Postgres พร้อมวัดและแก้จุดช้าด้วย pprof |
| 6–7 | DevOps | Docker, compose, Kubernetes, pipeline CI; Terraform บน LocalStack (S3, SQS, IAM) |
| 8 | Capstone backend | Claim Tracker Lite: Go + Postgres + Valkey + Kafka, outbox, idempotency, retry + DLQ |
| 9 | Capstone frontend | Vue + Kong gateway + nginx + manifest k8s + pipeline 5 job, รันจริงผ่าน 32 ขั้น |
| 10 | Observability + Security | metrics, trace, log, JWT, เช็คเจ้าของเคลม, dashboard, alert |

## ได้เรียนอะไร

- ของที่ทดสอบผ่านยังพังได้ตอนใช้จริง: ตอน Day 9 เปิด UI จริงถึงเห็นคำว่า `ok` โผล่ในหน้าจอ และ nginx ไม่ส่ง security header ที่ asset
- ระบบกระจาย: outbox, at-least-once, idempotency ต้องคิดคู่กัน และ idempotency ต้องผูกกับเจ้าของ (บั๊กที่เจอตอนตรวจ security ของ Day 10)
- เห็นระบบได้ต้องวางตั้งแต่แรก: label ของ metric, trace ข้าม Kafka, log ที่ไม่มีข้อมูลส่วนบุคคล
- ความปลอดภัยเริ่มจากค่าเริ่มต้นที่ปิด: ไม่มี token = ปฏิเสธ, ไม่มีเจ้าของ = ปฏิเสธ, port ผูก localhost
- ทุกครั้งที่แก้โค้ดมีเทสต์ + baseline snapshot และใช้การแก้โค้ดจงใจเช็คว่าเทสต์ไม่ได้ผ่านเฉยๆ

## ปัญหาที่เจอ

- ในที่เขียนโค้ดดึง Go module ภายนอกไม่ได้หลายรอบ (Day 4, 5, 10) บางส่วนเลยคอมไพล์กับ module จริงและรันกับ Postgres จริงไม่ได้ในรอบเขียน
- ของหลายอย่างเขียนแล้วแต่ยังไม่เคยรันจริง: manifest k8s และสคริปต์ deploy, pipeline บน GitHub, Terraform บน AWS จริง, compose ของ Day 10
- PowerShell 5.1: อ่านไฟล์ไม่มี BOM เป็น ANSI, อ่าน body ของ error ต้องผ่าน `ErrorDetails`
- ภาพหน้า UI แคปจากเว็บใน Docker เป็นไฟล์ตรงๆ ไม่ได้ ต้อง render จาก build จริง

## อยากต่อยอด

- รัน `run-all` ของ Day 10 บนเครื่องจริง แล้วแก้สิ่งที่พัง
- ย้าย metrics/trace ไปใช้ SDK มาตรฐาน (client_golang, OpenTelemetry) และวัด Kafka consumer lag
- ให้ Day 9 UI มี login และส่ง token ให้ Kong ตรวจ JWT
- รัน k8s กับ pipeline บน GitHub จริง
- สถานะ FAILED กับเครื่องมือ replay จาก DLQ, keyset pagination, load test
