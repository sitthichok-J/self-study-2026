# Compose vs Kubernetes

| เรื่อง | Docker Compose | Kubernetes |
|---|---|---|
| ใช้ทำอะไร | รันหลาย container บนเครื่องเดียว | รันบนหลายเครื่อง (cluster) |
| ไฟล์ | `docker-compose.yml` ไฟล์เดียว | หลาย manifest (Deployment, Service, ConfigMap, Secret ...) |
| ถ้า container ตาย | `restart: unless-stopped` ช่วยเริ่มใหม่ | Deployment สร้าง pod ใหม่ให้ครบจำนวน replicas เอง |
| scale | `--scale api=3` (บนเครื่องเดียว) | `replicas: N` + HPA ปรับตามโหลดได้ |
| health check | `healthcheck` + `depends_on` ตอนเริ่ม | startup / readiness / liveness probe ตลอดเวลา readiness fail = ไม่ส่ง traffic เข้า |
| อัปเดต | down/up (มี downtime สั้นๆ) | rolling update ไม่ downtime, `rollout undo` ย้อนได้ |
| config / secret | `environment`, `.env` | ConfigMap, Secret แยกจากตัว image |
| network | ชื่อ service เป็น hostname (`db`) | Service (ClusterIP) + DNS (`postgres`) |
| ความยาก | ต่ำ ตั้งเร็ว | สูง ต้องรู้หลาย object |

## ใช้อะไรเมื่อไหร่

- Compose: dev บนเครื่อง, เทสต์ integration, demo, ระบบเล็กที่มีเครื่องเดียว
- Kubernetes: production ที่ต้องการ high availability, scale ตามโหลด, deploy ไม่ downtime, หลายทีมใช้ cluster ร่วมกัน
- ไม่จำเป็นต้องใช้ K8s ถ้าบริการเล็กและมีเครื่องเดียว ความซับซ้อนที่เพิ่มต้องคุ้มกับสิ่งที่ได้

## ในงานนี้

- ไฟล์เดียวกัน (`/healthz`, env var เดียวกัน, image เดียวกัน) ใช้ได้ทั้งสองแบบ
- Compose ใช้ `depends_on: service_healthy` รอ DB ก่อน API เริ่ม ส่วน K8s ไม่มีลำดับแบบนี้ API ต่อ DB ไม่ติดตอนแรกจะ ping ไม่ผ่านแล้วจบ pod จะถูก restart จนกว่า DB พร้อม (จุดที่ควรปรับต่อ: retry ตอน start)
- `/healthz` ตอนนี้ตอบ ok ถ้า process ทำงาน ไม่ได้เช็ก DB ถ้าอยากให้ readiness สะท้อนว่า DB ใช้ได้ต้องเพิ่ม endpoint แยก (เช่น `/readyz`) แล้วเพิ่มเทสต์
