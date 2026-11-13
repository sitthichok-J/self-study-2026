# ผลรัน CI ของ Day 15 (จริง)

workflow `day15` มี 2 job: `static` (unittest + promtool 3 คำสั่ง) และ `kind` (ระบบ Day 14 ทั้งชุด แล้วติดตั้ง Prometheus + Grafana และ smoke)

## รอบที่ 1: แดง (commit 83dc23a)

- job kind ล้มที่ขั้นติดตั้ง Prometheus + Grafana แล้วตรวจ metric
- deploy และ rollout ผ่าน smoke ผ่าน 10 จาก 11 ข้อ ข้อที่ล้มคือเช็กว่า `auth_failures_total` เพิ่มหลัง login ผิด
- สาเหตุ: เช็กของผมเองผิด ตัวนับนี้เพิ่มเฉพาะตอน Bearer token ถึง api แล้วถูกปฏิเสธ (middleware `authed` ใน handler.go)
  login รหัสผ่านผิดแค่เขียน log `login failed` ไม่เพิ่มตัวนับ และ request ที่ไม่มี token ถูก Kong ตัดก่อนถึง api
  ผมเขียนเช็กโดยไม่ได้อ่านโค้ดส่วนนี้ก่อน

## รอบที่ 2: เขียว (commit 81c47a9)

- success รวม 3m08s (static 11s, kind 2m52s) smoke ok: 11/11
- แก้: เอาเช็ก auth_failures_total และ request login ผิดออกจาก smoke ใช้ `claims_submitted_total` (เพิ่มตอน api สร้างเคลม) แทน และเขียนข้อจำกัดลง RUNBOOK

บรรทัดผลจาก log จริงของรอบที่ 2:

```
PASS Prometheus พร้อม (/-/ready)
PASS สร้าง traffic ผ่าน gateway (login + GET + POST เคลม)
PASS Prometheus scrape claims-api และ claims-worker ได้ (health=up)
PASS เห็น http_requests_total จาก api (> 0)
PASS เห็น claims_decided_total จาก worker (> 0)
PASS เห็น claims_submitted_total จาก api หลังสร้างเคลม (> 0)
PASS Prometheus โหลด alert ครบ 7 ตัวจาก alerts.yml
PASS ApiDown ไม่ลั่นตอนระบบปกติ
PASS Grafana พร้อม (/api/health)
PASS Grafana มี dashboard claims-overview จาก provisioning
PASS Grafana ต่อ Prometheus ได้ (datasource health)
INFO metric ใน dashboard ที่ Prometheus ยังไม่มี series: ไม่มี
smoke ok: 11/11
```

## ยังไม่ได้ลอง

- เปิด Grafana ดูกราฟด้วยตา (ตรวจแค่ว่า dashboard ถูกโหลดและ datasource ต่อได้)
- alert ที่ลั่นจริงใน cluster (เทสต์ alert ใช้ promtool กับข้อมูลจำลอง)
- ทำตามขั้นตอนใน RUNBOOK.md บน cluster
- รันบนเครื่อง Windows ของตัวเอง และรันบน pull request
- ย้ายไป client_golang / OTel SDK (ต้องรัน go get บนเครื่องที่มี Go)
