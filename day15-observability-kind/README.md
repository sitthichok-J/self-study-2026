# Day 15 — Observability บน kind (Prometheus + Grafana + runbook)

เอาของที่ Day 10 เขียนไว้ (metrics, alert, dashboard) มาต่อกับ cluster kind ของ Day 14 แล้วตรวจว่า Prometheus เห็น api/worker จริง

## สถานะ (อ่านก่อน)

- เขียนแล้ว: manifest Prometheus + Grafana, สคริปต์ `obsctl.py`, เทสต์ promtool สำหรับ alert ทั้ง 7 ตัว, runbook, เทสต์ static 30 ข้อ + baseline, workflow `day15`
- รันจริงแล้ว: เทสต์ static 30 ข้อผ่านบนเครื่อง (ลองแก้ไฟล์ให้พังตั้งใจ 9 แบบ เทสต์จับได้ครบ)
- **ยังไม่เคยรัน:** promtool (ที่นี่ไม่มี), cluster, Prometheus/Grafana จริง ผลทั้งหมดของส่วนนั้นต้องรอ CI รอบแรก ถ้าแดงให้แก้ตามที่เห็นแล้วจดลง `notes/run-results.md`
- ยังไม่ได้ทำ: ย้ายจาก `internal/obs` (เขียนเองใน Day 10) ไปใช้ client_golang และ OTel SDK, Jaeger/collector ใน cluster (trace ยังไม่ถูกส่งไปไหนใน k8s เพราะ `OTLP_ENDPOINT` ว่างไว้), เปิด Grafana ดูด้วยตา
- เหตุที่ยังไม่ย้าย lib: เครื่องที่ใช้เขียนไม่มี Go และโหลด module ไม่ได้ จึงสร้าง `go.sum` และคอมไพล์ไม่ได้

## โครง

```
k8s/observability.yaml     Service api-metrics / worker-metrics (พอร์ต 9090) + Prometheus + Grafana
k8s/prometheus/            prometheus.yml (job เดิมของ Day 13 แต่ target เป็น Service ใน cluster)
k8s/grafana/               datasource (เหลือแค่ Prometheus)
promtool/tests.yml         เทสต์ alert: ทุกตัวมีเคสลั่นและเคสไม่ลั่น
scripts/obsctl.py          deploy ConfigMap/Secret/manifest แล้ว smoke
RUNBOOK.md                 ขั้นตอนตอน alert ลั่น (ทีละ alert)
tests/                     เทสต์ static + baseline
ci/day15.yml               สำเนา workflow (ตัวจริงอยู่ .github/workflows/day15.yml)
```

alerts.yml, dashboard และ provisioning ใช้ไฟล์ของ Day 13 ตรง ๆ ไม่ copy (ส่งเข้า cluster เป็น ConfigMap ตอน deploy)

## รัน

```
python day14-k8s-kind/scripts/kindctl.py all      # ระบบของ Day 14
python day15-observability-kind/scripts/obsctl.py all
python -m unittest discover -s day15-observability-kind/tests -v
```

smoke ตรวจ: Prometheus พร้อม, สร้าง traffic ผ่าน Kong, scrape api/worker ได้, เห็น http_requests_total / claims_submitted_total / claims_decided_total, โหลด alert ครบ 7 ตัว, ApiDown ไม่ลั่น, Grafana เห็น dashboard claims-overview และต่อ Prometheus ได้
(รายการ metric ใน dashboard ที่ยังไม่มี series พิมพ์เป็น INFO ไม่นับผ่าน/ไม่ผ่าน)

## ต้องดูด้วยตาเอง (เทสต์ตรวจแทนไม่ได้)

ตอนมี cluster: `kubectl -n claims port-forward svc/grafana 13000:3000` เปิด http://127.0.0.1:13000 (admin + รหัสใน `.work/grafana-password`) ดู dashboard "Claim Tracker – overview" แล้วเช็ก

- กราฟ request rate ขยับเมื่อยิง request
- เคลมที่ worker ตัดสินเพิ่มตามที่สร้าง
- latency ไม่เป็นศูนย์ และหน่วยถูก
- ไม่มี panel ที่ขึ้น No data ในช่วงที่ยิง (บางตัวอาจว่างถ้ายังไม่เกิดเหตุ เช่น panic)

เจออะไรจดลง `notes/run-results.md` ตามที่เห็นจริง
