# Day 14 แบบฝึกหัด: ลงมือทำแล้วจดผลจริง

ช่อง "ผลที่เห็น" ว่างไว้ ใส่เฉพาะที่เห็นจริงจากเครื่องตัวเอง ข้อไหนไม่ได้ลองให้เขียนว่า "ไม่ได้ลอง"

| ข้อ | ทำอะไร | คำสั่ง | ผลที่เห็น |
|---|---|---|---|
| 1 | pod ทั้งหมด Ready | `kubectl -n claims get pods` | |
| 2 | smoke test ผ่านครบ | `python scripts/kindctl.py smoke` | |
| 3 | เปิดหน้าเว็บ login ด้วย alice | เปิด http://localhost:3000 | |
| 4 | เห็น Secret ใน cluster แต่ไม่มีใน repo | `kubectl -n claims get secret` | |
| 5 | memory limit ต่ำมากจนเจอ OOMKilled | `kubectl -n claims set resources deploy/worker --limits=memory=8Mi` แล้ว `describe pod` ดู Last State จากนั้นตั้งคืน 128Mi | |
| 6 | rollout: เปลี่ยนแท็กเป็นตัวที่ build ใหม่ | `docker tag claims-api:day14 claims-api:day14b` + `kind load` + `kubectl -n claims set image deploy/api api=claims-api:day14b` + `rollout status` | |
| 7 | rollback: ใส่ image ที่ไม่มีจริง | `set image ... api=claims-api:nope` แล้ว `kubectl -n claims rollout undo deploy/api` | |
| 8 | ฆ่า pod api (replica เดียว) จดเวลา | `kubectl -n claims delete pod -l app=api` แล้ววัดเวลา Ready ใหม่ | |
| 9 | api 2 replica แล้วฆ่าอีกครั้ง เทียบช่วงที่เว็บล่ม | `kubectl -n claims scale deploy/api --replicas=2` | |
| 10 | ฆ่า pod valkey: token ที่ถูกเพิกถอนต้องยังถูกปฏิเสธ (AOF + PVC) | logout alice แล้วใช้ token เดิม, `delete pod -l app=valkey`, ใช้ token เดิมอีก | |
| 11 | trivy config ของ manifest | ดูผลใน CI รอบแรก | |
