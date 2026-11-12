# Day 14 — รัน Claim Tracker Lite บน Kubernetes จริงด้วย kind

ข้อมูลในตัวอย่างทั้งหมดเป็นข้อมูลสมมติ ผลที่รันจริงจะอยู่ใน `notes/run-results.md` (ไฟล์นี้ยังไม่มีจนกว่าจะรันบน cluster จริงครั้งแรก)

## สถานะตอนนี้

เขียน manifest, สคริปต์ deploy/smoke, เทสต์ static และ workflow เสร็จแล้ว เทสต์ static 34 ข้อผ่านในเครื่อง (รันแค่ Python ไม่ต้องมี Docker) แต่ **ยังไม่เคยสร้าง cluster จริงเลย** ทุกอย่างที่เกี่ยวกับ pod, probe, rollout ยังไม่ได้พิสูจน์ จนกว่าจะรัน `kindctl.py all` สำเร็จและมีผลใน `notes/`

## ทำอะไร

Day 6 และ Day 9 เขียน manifest ไว้แต่ไม่เคย apply และแอปเปลี่ยนไปมากตั้งแต่ Day 13 (JWT RS256 จากไฟล์คีย์, blocklist ใน valkey, rate limit ต่อ user, Kong ตรวจ JWT) วันนี้เขียน manifest ใหม่ให้ตรงกับแอปของ Day 13 แล้วรันจริงใน kind

```
browser ─► localhost:3000 ─► [kind node] NodePort 30300 ─► web (nginx :8080)
browser ─► localhost:8000 ─► [kind node] NodePort 30080 ─► gateway (Kong :8000) ─► api (:8080) ─► postgres, valkey, kafka
                                                                                  worker ◄── kafka   (worker ─► postgres, valkey)
```

- image ของแอปใช้ source ของ Day 13 (ไม่ copy โค้ด) build เป็น `claims-api:day14`, `claims-worker:day14`, `claims-web:day14` แล้ว `kind load`
- ความลับสร้างตอน deploy ไม่อยู่ใน repo: Secret `db` (รหัสสุ่ม), `app` (รหัส demo user สุ่ม), `jwt-keys` (คู่คีย์ RSA ที่ `cmd/keygen` ของ Day 13 สร้าง)
- Kong ได้แค่ public key (render จากเทมเพลตของ Day 13 ด้วย `scripts/render_kong.py`) ฟัง HTTP อย่างเดียว ไม่เปิด Admin API
- valkey เปิด AOF และเก็บลง PVC ตามที่ Day 13 ต้องการ เพื่อให้บัญชี token ที่ถูกเพิกถอนรอด pod ถูกสร้างใหม่
- ไม่มี Prometheus/Grafana/Jaeger ใน cluster นี้ (`OTLP_ENDPOINT` เว้นว่าง = ไม่ส่ง trace) อันนั้นเป็นของ Day 15

## ไฟล์

| ไฟล์ | หน้าที่ |
|---|---|
| `kind-config.yaml` | cluster node เดียว เปิดพอร์ต 8000 และ 3000 ที่ 127.0.0.1 เท่านั้น |
| `k8s/` | manifest ทั้งหมด (kustomize) |
| `scripts/kindctl.py` | `build`, `up`, `load`, `deploy`, `smoke`, `down`, `all` (ไฟล์เดียวใช้ทั้งในเครื่องและใน CI) |
| `scripts/render_kong.py` | ทำ kong.yml จากเทมเพลต + public key |
| `tests/test_k8s.py` | เทสต์ static + baseline (`tests/baseline/k8s-summary.json`) |
| `ci/day14.yml` | workflow (สำเนา: ตัวจริงอยู่ `.github/workflows/day14.yml`) |
| `notes/exercises.md` | แบบฝึกหัด rollout/rollback/kill pod ที่ต้องจดผลเอง |

## วิธีรัน

ต้องมี Docker Desktop (เปิดอยู่), kind, kubectl และ Python 3 + `pip install pyyaml`

```
python -m unittest discover -s tests -v       # เทสต์ static ไม่ต้องมี cluster
python scripts/kindctl.py all                 # build + สร้าง cluster + load + deploy + smoke
python scripts/kindctl.py down                # ลบ cluster
```

ทีละขั้น: `build` → `up` → `load` → `deploy` → `smoke` ถ้า pod ไม่ขึ้นดูตารางอาการใน doc ของวัน (`kubectl describe pod`, `kubectl logs <pod> --previous`)

user สมมติคือ `alice`, `bob`, `ops` (staff) รหัสผ่านสุ่มอยู่ใน `.work/demo-password` (git-ignore) เปิด http://localhost:3000 เพื่อลอง login

## เปลี่ยน manifest แล้วเทสต์ล้ม

baseline ใน `tests/baseline/k8s-summary.json` คือสิ่งที่ manifest ทำอยู่ตอนนี้ ถ้าแก้ตั้งใจ รัน `UPDATE_BASELINE=1 python -m unittest discover -s tests` แล้วอ่าน diff ของไฟล์นั้นก่อน commit

## ที่ยังไม่ได้ทำ / ข้อจำกัด

- ยังไม่ได้รันบน cluster จริง (ดู "สถานะตอนนี้")
- postgres และ kafka ตัวเดียวไม่ทำ HA, postgres ใช้ emptyDir (ข้อมูลหายเมื่อ pod ถูกสร้างใหม่)
- ไม่มี TLS ที่ Kong ใน cluster นี้ (cookie refresh ตั้ง `Secure` เหมือน compose: browser ยอมให้ใช้บน localhost)
- ไม่มี Ingress controller ใช้ NodePort + extraPortMappings ของ kind แทน
- ขั้น kind ใน CI ยังไม่เคยรัน
