# Runbook: Claim Tracker Lite บน kind

เขียนจากโค้ดและ manifest ของ Day 10, 13, 14 ขั้นตอนที่เป็นคำสั่ง kubectl ยังไม่ได้ลองทำตามทีละข้อบน cluster จริง
(เทสต์ที่มีตรวจแค่ว่าทุก alert มีหัวข้อในไฟล์นี้) ลองทำตามแล้วแก้ไฟล์นี้ตามที่เจอจริง

namespace ทั้งหมดคือ `claims` ดู Prometheus / Grafana ผ่าน port-forward:

```
kubectl -n claims port-forward svc/prometheus 19090:9090
kubectl -n claims port-forward svc/grafana 13000:3000
```

รหัส admin ของ Grafana อยู่ใน Secret `grafana` (`kubectl -n claims get secret grafana -o jsonpath={.data.admin-password}` แล้ว base64 decode)
dashboard ชื่อ "Claim Tracker – overview" (uid `claims-overview`)

## ApiDown

อาการ: Prometheus scrape claims-api ไม่ได้ (`up{job="claims-api"} == 0`) นาน 1 นาที severity page

เช็ก:

1. `kubectl -n claims get pods -l app=api` ดู STATUS และ RESTARTS
2. `kubectl -n claims describe pod -l app=api` ดู Last State ว่า OOMKilled ไหม (limit memory 128Mi) และ events ท้ายสุด
3. `kubectl -n claims logs deploy/api --tail=50` และ `--previous` ถ้า pod เพิ่ง restart
4. `kubectl -n claims get pods -l app=postgres` เพราะ api พึ่ง DB (ถ้า /readyz ไม่ผ่าน pod จะไม่รับ traffic)

สาเหตุที่น่าสงสัย: DB ไม่พร้อมหรือรหัสผ่านใน Secret `db` ไม่ตรงกับที่ postgres ถืออยู่, OOMKilled, image `claims-api:day14` ไม่ถูกโหลดเข้า kind

แก้: ถ้าเป็น OOMKilled เพิ่ม limit ใน k8s/api.yaml ของ Day 14 ถ้าเป็นที่ DB ดู postgres ก่อน ถ้าแค่ค้างลอง `kubectl -n claims rollout restart deploy/api`
ยืนยันว่าหาย: `up{job="claims-api"}` กลับเป็น 1 และ alert หายภายใน 1-2 นาที

## HighErrorRate

อาการ: สัดส่วน 5xx เกิน 5% ของ request ทั้งหมดใน 5 นาที นาน 5 นาที severity page

เช็ก:

1. ใน Prometheus: `sum by (route, code) (rate(http_requests_total{code=~"5.."}[5m]))` ดูว่า route ไหนเป็นตัวการ
2. `kubectl -n claims logs deploy/api --tail=100` หา error ก่อนหน้า 5xx และ trace_id ใน log
3. ดู `db_pool_acquired_conns / db_pool_max_conns` (มีใน dashboard)
4. ดู kafka/worker: `kubectl -n claims get pods -l app=kafka`

สาเหตุที่น่าสงสัย: DB ล่มหรือ pool เต็ม, kafka ไม่พร้อมทำให้ส่ง event ไม่ได้, deploy ใหม่ที่มี bug
แก้: แก้ที่ต้นเหตุตาม log ถ้าเป็นหลัง deploy ใหม่ ย้อนกลับ `kubectl -n claims rollout undo deploy/api`
ยืนยันว่าหาย: สัดส่วน 5xx (panel Error ratio ใน dashboard) ต่ำกว่า 5% และคงอยู่

## HighLatencyP95

อาการ: p95 ของ `http_request_duration_seconds` เกิน 1 วินาที นาน 5 นาที severity ticket

เช็ก:

1. ดู route ที่ช้า: `histogram_quantile(0.95, sum by (le, route) (rate(http_request_duration_seconds_bucket[5m])))`
2. `db_pool_acquired_conns` ใกล้ `db_pool_max_conns` ไหม (รอ connection)
3. `http_requests_in_flight` และ `go_goroutines` พุ่งไหม
4. `kubectl -n claims describe pod -l app=api` ดู request/limit ของ CPU และ memory เทียบกับโหลดที่เข้ามา

สาเหตุที่น่าสงสัย: pool เต็ม, query ช้า, โหลดสูงกว่าที่ pod เดียวรับไหว
แก้: ลดโหลดหรือเพิ่ม replicas ของ api ใน k8s/api.yaml (ยังไม่เคยลอง 2 replicas) ถ้าเป็นที่ query ดู log ช่วงนั้น
ยืนยันว่าหาย: p95 ต่ำกว่า 1 วินาทีต่อเนื่อง 5 นาที

## DeadLetterGrowing

อาการ: มี message ลง dead-letter topic `claims.submitted.dlq` (`consumer_messages_total{outcome="dlq"}` เพิ่มใน 10 นาที) severity ticket

เช็ก:

1. `kubectl -n claims logs deploy/worker --tail=100` หาบรรทัดที่บอกสาเหตุที่ message ถูกส่งเข้า DLQ
2. ดูว่าเป็นเคลมเดียวหรือหลายอัน: `sum by (outcome) (rate(consumer_messages_total[5m]))`
3. ถ้าสงสัย DB: `kubectl -n claims get pods -l app=postgres`

สาเหตุที่น่าสงสัย: payload ที่ worker อ่านไม่ได้, DB หรือ valkey ใช้ไม่ได้ชั่วคราวจนลองซ้ำครบ
แก้: ถ้าเป็นข้อมูลเสียเฉพาะอัน ไม่ต้องแก้ระบบ บันทึกไว้ ถ้าเป็นเพราะ dependency ล่ม แก้ dependency แล้วค่อยหาวิธีส่ง message ใน DLQ กลับ (ยังไม่มีขั้นตอนนี้ใน repo)
ยืนยันว่าหาย: ตัวเลข dlq ไม่เพิ่มอีก

## AuthFailureSpike

อาการ: api ปฏิเสธ Bearer token เกิน 1 ครั้งต่อวินาที นาน 5 นาที severity ticket

ข้อควรรู้ (เจอตอนรัน smoke ใน CI): `auth_failures_total` นับเฉพาะ token ที่ถึง api แล้วถูกปฏิเสธ (หมดอายุ/ถูก revoke) login รหัสผ่านผิดแค่เขียน log `login failed` ไม่เพิ่มตัวนับนี้ และ request ที่ไม่มี token หรือ token เสียถูก Kong ตัดก่อนถึง api ดังนั้น alert นี้ไม่เห็นการเดารหัสผ่าน

เช็ก:

1. `sum by (reason) (rate(auth_failures_total[5m]))` ดู reason ว่า token ถูกปฏิเสธเพราะอะไร
2. `kubectl -n claims logs deploy/api --tail=100` หาบรรทัด `authentication failed` และ `login failed` (login ผิดดูได้จาก log เท่านั้น)
3. เช็กว่า Secret `jwt-keys` กับ Kong config ใช้คีย์ชุดเดียวกัน (key ไม่ตรงทำให้ token ทุกตัวถูกปฏิเสธ)

สาเหตุที่น่าสงสัย: client ที่ผิดตัวหนึ่งวนยิงด้วย token เก่า, token ถูก revoke, เปลี่ยนคีย์แล้วลืมเปลี่ยนอีกฝั่ง
แก้: ถ้ามี client วนยิง จำกัดที่ gateway ถ้าเป็นคีย์ไม่ตรง deploy ใหม่ให้ Kong กับ api ใช้คีย์เดียวกัน
ยืนยันว่าหาย: อัตรา auth_failures_total ลดต่ำกว่า 1 ต่อวินาที

## AccessDeniedSeen

อาการ: owner check ปฏิเสธ request อย่างน้อย 1 ครั้งใน 10 นาที (`authz_denied_total` เพิ่ม) severity ticket

เช็ก:

1. `sum by (resource) (increase(authz_denied_total[10m]))`
2. `kubectl -n claims logs deploy/api --tail=200` ดู request ช่วงที่ตัวเลขเพิ่ม (ยังไม่ได้ตรวจว่า log บอกผู้ใช้ครบแค่ไหน)

สาเหตุที่น่าสงสัย: มีคนลองเปิดเคลมของคนอื่น (IDOR) หรือ UI ส่ง id ผิด
แก้: ถ้าเป็น UI bug แก้ที่ web ถ้าเป็นความพยายามเจาะ บันทึกและตามผู้ใช้นั้น
ยืนยันว่าหาย: ตัวเลขไม่เพิ่มต่อใน 10 นาที

## DbPoolSaturated

อาการ: connection pool ใช้เกิน 90% (`db_pool_acquired_conns / db_pool_max_conns`) นาน 2 นาที severity ticket

เช็ก:

1. ดู `db_pool_acquired_conns`, `db_pool_idle_conns` และ `db_pool_max_conns` (มีใน dashboard) idle ควรเหลือบ้าง
2. `http_requests_in_flight` สูงผิดปกติไหม ถ้าสูงคือ request ค้างถือ connection
3. `kubectl -n claims logs deploy/postgres --tail=50`

สาเหตุที่น่าสงสัย: โหลดสูง, query ช้าถือ connection นาน, request ค้างเพราะ dependency ช้า
แก้: ลดโหลด หรือหา query ที่ช้า การเพิ่มขนาด pool ต้องดูว่า postgres รับ connection ไหวไหมก่อน
ยืนยันว่าหาย: สัดส่วนต่ำกว่า 90% และ latency p95 ปกติ
