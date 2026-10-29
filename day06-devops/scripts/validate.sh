#!/usr/bin/env bash
# ตรวจ config ทั้งหมดโดยไม่ต้องมี docker/kubectl (ต้องมี python3 + pyyaml)
# ถ้ามี docker/kubectl จะตรวจเพิ่มด้วย
set -euo pipefail
cd "$(dirname "$0")/.."
python3 -m unittest discover -s tests -v
if command -v docker >/dev/null && docker info >/dev/null 2>&1; then
  docker compose config -q && echo "docker compose config: OK"
fi
if command -v kubectl >/dev/null; then
  kubectl kustomize k8s >/dev/null && echo "kustomize build: OK"
fi
