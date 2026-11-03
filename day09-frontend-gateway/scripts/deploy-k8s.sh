#!/usr/bin/env bash
# deploy ขึ้น cluster ที่ kubectl ชี้อยู่ (kind/minikube) ใช้ทั้งในเครื่องและใน CI
# ต้องมี image claims-api:local, claims-worker:local, claims-web:local อยู่ใน cluster แล้ว (kind load docker-image)
set -euo pipefail
cd "$(dirname "$0")/.."
NS=claims

kubectl apply -f k8s/namespace.yaml

# รหัสผ่าน DB: ใช้ของเดิมถ้ามี Secret อยู่แล้ว (postgres ยังถือรหัสเดิมอยู่) ไม่งั้นใช้ DB_PASSWORD หรือสุ่มใหม่
PASS="$(kubectl -n "$NS" get secret db -o jsonpath='{.data.password}' 2>/dev/null | base64 -d 2>/dev/null || true)"
if [ -z "$PASS" ]; then
  PASS="${DB_PASSWORD:-$(head -c 24 /dev/urandom | od -An -tx1 | tr -d ' \n')}"
fi
if ! [[ "$PASS" =~ ^[A-Za-z0-9]+$ ]]; then
  echo "DB_PASSWORD ต้องเป็นตัวอักษรหรือตัวเลขเท่านั้น (ไม่ต้อง URL-encode)" >&2
  exit 2
fi
kubectl -n "$NS" create secret generic db \
  --from-literal=password="$PASS" \
  --from-literal=url="postgres://claims:${PASS}@postgres:5432/claims?sslmode=disable" \
  --dry-run=client -o yaml | kubectl apply -f -

# ไฟล์ต้นฉบับมีที่เดียว: migration จาก Day 8 และ kong.yml ของ gateway
kubectl -n "$NS" create configmap migrations --from-file=../day08-backend/migrations \
  --dry-run=client -o yaml | kubectl apply -f -
kubectl -n "$NS" create configmap kong-config --from-file=kong.yml=gateway/kong.yml \
  --dry-run=client -o yaml | kubectl apply -f -

# Job แก้ spec ไม่ได้ ลบของเก่าก่อน apply ใหม่
kubectl -n "$NS" delete job kafka-init migrate --ignore-not-found --wait=true
kubectl apply -k k8s

kubectl -n "$NS" wait --for=condition=complete job/kafka-init job/migrate --timeout=300s
for d in postgres valkey kafka api worker gateway web; do
  kubectl -n "$NS" rollout status "deploy/$d" --timeout=300s
done
echo "deploy ok"
