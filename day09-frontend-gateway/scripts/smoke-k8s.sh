#!/usr/bin/env bash
# ทดสอบผ่าน gateway ใน cluster: ส่งเคลม -> รอ worker ตัดสิน -> ตรวจ web และ CORS
set -euo pipefail
NS=claims
kubectl -n "$NS" port-forward svc/gateway 8000:8000 >/tmp/pf-gateway.log 2>&1 &
PF1=$!
kubectl -n "$NS" port-forward svc/web 3000:3000 >/tmp/pf-web.log 2>&1 &
PF2=$!
trap 'kill $PF1 $PF2 2>/dev/null || true' EXIT

for i in $(seq 1 30); do
  curl -fsS http://localhost:8000/api/healthz >/dev/null 2>&1 && break
  sleep 1
  [ "$i" = 30 ] && { echo "gateway ไม่ตอบ"; cat /tmp/pf-gateway.log; exit 1; }
done

KEY="smoke-$(date +%s)"
BODY='{"patient_ref":"FAKE-SMOKE1","amount":150000}'
RES=$(curl -fsS -X POST http://localhost:8000/api/claims -H "Idempotency-Key: $KEY" -H "Content-Type: application/json" -d "$BODY")
ID=$(echo "$RES" | sed -n 's/.*"id":"\([^"]*\)".*/\1/p')
[ -n "$ID" ] || { echo "ไม่ได้ id: $RES"; exit 1; }

STATUS=""
for i in $(seq 1 30); do
  STATUS=$(curl -fsS "http://localhost:8000/api/claims/$ID" | sed -n 's/.*"status":"\([^"]*\)".*/\1/p')
  [ "$STATUS" = "APPROVED" ] && break
  sleep 1
done
[ "$STATUS" = "APPROVED" ] || { echo "สถานะสุดท้าย: $STATUS"; exit 1; }

curl -fsS http://localhost:3000/ | grep -q '<div id="app">' || { echo "web ไม่ส่ง index.html"; exit 1; }
HDR=$(curl -sS -i -X OPTIONS http://localhost:8000/api/claims -H "Origin: http://localhost:3000" -H "Access-Control-Request-Method: POST" -H "Access-Control-Request-Headers: idempotency-key,content-type")
echo "$HDR" | grep -qi '^access-control-allow-origin: http://localhost:3000' || { echo "CORS preflight ไม่ผ่าน"; echo "$HDR"; exit 1; }
echo "smoke ok: claim $ID APPROVED ผ่าน gateway"
