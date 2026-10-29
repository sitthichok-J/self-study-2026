# day06-devops

Docker / Compose / Kubernetes / CI สำหรับ API จาก `../day05-go-service`

```
Dockerfile                multi-stage (golang:1.25 -> distroless nonroot)
Dockerfile.dockerignore   ไฟล์ที่ไม่ copy เข้า build context
docker-compose.yml        api + postgres (healthcheck, volume, init SQL จาก migration จริง)
.env.example              ตัวอย่างตัวแปรของ compose
k8s/                      namespace, configmap, secret, postgres, api (2 replicas + probes), kustomization
ci/                       github-actions-ci.yml, gitlab-ci.yml (เลือกใช้ตามที่ทำงาน)
notes/                    compose-vs-k8s.md, exercises.md (คำสั่งแบบฝึกหัด 1-5)
tests/ testdata/          เทสต์ตรวจความสอดคล้อง + baseline.json
scripts/validate.sh       รันเทสต์ (+ docker compose config / kustomize ถ้ามีเครื่องมือ)
```

## ลำดับที่ควรทำ

1. เตรียม dependency ของ Day 5 (ครั้งเดียว): `cd ../day05-go-service && go get github.com/jackc/pgx/v5 && go mod tidy` (ให้ได้ go.sum ที่ Dockerfile ใช้)
2. ตรวจ config: `pip install pyyaml && bash scripts/validate.sh`
3. Compose: `docker compose up --build -d` แล้วทดสอบตาม `notes/exercises.md` ข้อ 2
4. Kubernetes: ตาม `notes/exercises.md` ข้อ 3-5 (ต้องมี kind + kubectl)
5. CI: คัดลอก `ci/github-actions-ci.yml` ไป `<repo>/.github/workflows/ci.yml` หรือ `ci/gitlab-ci.yml` ไป `<repo>/.gitlab-ci.yml`

## สถานะการทดสอบ

- ตรวจแล้ว: เทสต์ 33 ข้อใน `tests/` (ไฟล์ YAML อ่านได้, env var ตรงกับ `config.go`, probe ชี้ route ที่มีจริง, พอร์ตตรงกัน, grace period > SHUTDOWN_TIMEOUT, ไม่มี tag latest, รันเป็น non-root, copy ของ migration ใน k8s ตรงกับต้นฉบับ) และลองทำให้ config พังแล้วเทสต์จับได้
- ยังไม่ได้รันจริง: `docker build`, `docker compose up`, kind/kubectl และ pipeline ทั้งสองแบบ ใน sandbox ไม่มี Docker daemon กับ cluster และดึง pgx ไม่ได้ ต้องรันบนเครื่อง ถ้าเจอ error ให้ส่งมาแก้
- ข้อควรรู้: `/healthz` ยังไม่เช็ก DB; ถ้า API ต่อ DB ไม่ได้ตอนเริ่ม process จะจบแล้ว K8s restart ให้ ดู `notes/compose-vs-k8s.md`

## Ran on a real machine

`scripts/run-all.bat` (calls `run-all.ps1`) runs Go build/test, docker build, Compose, a kind cluster and actionlint, then writes `notes/run-results.md`. Last run: 22 of 22 steps passed (Docker Desktop, Go 1.26, kind with node image v1.34, kubectl 1.36). The CI pipelines themselves were not run, only actionlint on the GitHub Actions file.

Two things the run found and fixed: the Dockerfile needed `golang:1.25` because `go.mod` moved to Go 1.25 with pgx, and the api pod needed a numeric `runAsUser: 65532` for `runAsNonRoot` to work with the distroless `nonroot` image. If Docker Desktop uses cgroup v1, kind's newest node image fails to start, so the script pins `kindest/node:v1.34.0`. `scripts/cleanup-run.bat` stops a stuck run and removes the leftover containers and cluster.
