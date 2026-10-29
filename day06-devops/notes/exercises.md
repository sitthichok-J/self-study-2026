# แบบฝึกหัดวันที่ 6 (คำสั่งสำหรับลองบนเครื่อง)

## 1. เปรียบเทียบขนาด image (golang vs distroless)
```
docker build -f day06-devops/Dockerfile --target build -t claims-api:build day05-go-service
docker build -f day06-devops/Dockerfile -t claims-api:dev day05-go-service
docker images | grep claims-api
```
บันทึกขนาดทั้งสองไว้ในตารางสรุป (stage build มีทั้ง Go toolchain จึงใหญ่กว่ามาก)

## 2. Compose
```
cd day06-devops
docker compose up --build -d
docker compose ps          # db ต้อง healthy
curl localhost:8080/healthz
curl -XPOST localhost:8080/claims -d '{"patient_ref":"P001","hospital":"รพ.ทดสอบ","amount_satang":150000}'
docker compose down        # เก็บข้อมูลไว้ (volume pgdata)
docker compose down -v     # ลบข้อมูลด้วย
```

## 3. kind + self-healing
```
kind create cluster --name claims
kind load docker-image claims-api:dev --name claims
kubectl apply -k day06-devops/k8s
kubectl -n claims get pods -w
kubectl -n claims delete pod -l app=claims-api --field-selector=status.phase=Running   # ลบ pod แล้วดูว่าถูกสร้างใหม่
kubectl -n claims port-forward svc/claims-api 8080:80
```

## 4. ทำให้ readiness fail
```
kubectl -n claims patch deployment claims-api --type=json -p='[{"op":"replace","path":"/spec/template/spec/containers/0/readinessProbe/httpGet/path","value":"/nope"}]'
kubectl -n claims get pods              # READY 0/1
kubectl -n claims get endpoints claims-api   # รายการ endpoint ว่าง = ไม่มี traffic เข้า
kubectl -n claims rollout undo deployment/claims-api
```
ข้อสังเกต: เพราะ `maxUnavailable: 0` pod เก่ายังรับ traffic ระหว่าง rollout ที่ติด ไม่มี downtime

## 5. ลอง graceful shutdown ใน K8s
```
kubectl -n claims logs -f deploy/claims-api &
kubectl -n claims rollout restart deployment/claims-api
```
ดูว่า log มี `shutting down` และ `shutdown complete` ตอน pod เก่าถูกปิด
