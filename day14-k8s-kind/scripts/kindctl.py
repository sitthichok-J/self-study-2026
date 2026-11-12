"""ตัวคุมงาน Day 14: build image, สร้าง cluster kind, โหลด image, deploy, smoke test

รันได้เหมือนกันทั้งบนเครื่อง (Windows/Linux) และใน GitHub Actions เพราะเป็น Python ไฟล์เดียว
ต้องมี docker, kind, kubectl ใน PATH   ใช้:  python scripts/kindctl.py <build|up|load|deploy|smoke|down|all>
ไม่มีความลับในไฟล์นี้: รหัสผ่านสุ่มตอนรันแล้วเก็บใน .work/ (git-ignore) และใน Secret ของ cluster
"""
import base64
import json
import os
import pathlib
import secrets
import subprocess
import sys
import time
import urllib.error
import urllib.request

HERE = pathlib.Path(__file__).resolve().parent
ROOT = HERE.parent
APP13 = ROOT.parent / "day13-revocation-ui-cicd"
WORK = ROOT / ".work"
CLUSTER = "claim"
NS = "claims"
TAG = "day14"
IMAGES = {
    "claims-api": (APP13, ["CMD=api"]),
    "claims-worker": (APP13, ["CMD=worker"]),
    "claims-web": (APP13 / "web", ["VITE_API_BASE=http://localhost:8000/api"]),
}
DEPLOYS = ["postgres", "valkey", "kafka", "api", "worker", "gateway", "web"]
GW = "http://localhost:8000"
WEB = "http://localhost:3000"


def run(cmd, check=True, capture=False, input_text=None):
    print("+", " ".join(str(c) for c in cmd), flush=True)
    r = subprocess.run(cmd, check=False, text=True, input=input_text,
                       stdout=subprocess.PIPE if capture else None,
                       stderr=subprocess.STDOUT if capture else None)
    if check and r.returncode != 0:
        if capture and r.stdout:
            print(r.stdout)
        sys.exit(f"คำสั่งล้มเหลว (exit {r.returncode}): {' '.join(str(c) for c in cmd)}")
    return r


def kubectl(*args, **kw):
    return run(["kubectl", "--context", f"kind-{CLUSTER}", *args], **kw)


def apply_yaml(text):
    kubectl("apply", "-f", "-", input_text=text)


def make_secret(name, *opts):
    out = kubectl("-n", NS, "create", "secret", "generic", name, *opts,
                  "--dry-run=client", "-o", "yaml", capture=True).stdout
    apply_yaml(out)


def make_configmap(name, *opts):
    out = kubectl("-n", NS, "create", "configmap", name, *opts,
                  "--dry-run=client", "-o", "yaml", capture=True).stdout
    apply_yaml(out)


def cmd_build():
    for name, (ctx, args) in IMAGES.items():
        cmd = ["docker", "build", "-t", f"{name}:{TAG}"]
        for a in args:
            cmd += ["--build-arg", a]
        run(cmd + [str(ctx)])


def cmd_up():
    have = run(["kind", "get", "clusters"], capture=True).stdout.split()
    if CLUSTER in have:
        print("cluster", CLUSTER, "มีอยู่แล้ว ข้ามขั้นสร้าง")
        return
    run(["kind", "create", "cluster", "--config", str(ROOT / "kind-config.yaml")])


def cmd_load():
    run(["kind", "load", "docker-image", *[f"{n}:{TAG}" for n in IMAGES], "--name", CLUSTER])


def ensure_keys():
    keys = WORK / "keys"
    if (keys / "k1.key").exists() and (keys / "k1.pub").exists():
        return keys
    keys.mkdir(parents=True, exist_ok=True)
    run(["docker", "build", "--build-arg", "CMD=keygen", "-t", f"claims-keygen:{TAG}", str(APP13)])
    cmd = ["docker", "run", "--rm"]
    if hasattr(os, "getuid"):  # Linux/CI: เขียนไฟล์เป็น user ของเรา ไม่ใช่ uid ของ image
        os.chmod(keys, 0o777)
        cmd += ["--user", f"{os.getuid()}:{os.getgid()}"]
    run(cmd + ["-v", f"{keys}:/keys", f"claims-keygen:{TAG}", "-dir", "/keys", "-kid", "k1"])
    return keys


def stored(path, make):
    p = WORK / path
    if not p.exists():
        WORK.mkdir(parents=True, exist_ok=True)
        p.write_text(make(), encoding="ascii")
    return p.read_text(encoding="ascii").strip()


def cmd_deploy():
    keys = ensure_keys()
    sys.path.insert(0, str(HERE))
    import render_kong
    kong = WORK / "kong.yml"
    render_kong.main(["render_kong.py", str(APP13 / "gateway" / "kong.yml.tpl"), str(keys), str(kong)])

    kubectl("apply", "-f", str(ROOT / "k8s" / "namespace.yaml"))
    # รหัสผ่าน DB: ใช้ของเดิมถ้ามี Secret อยู่แล้ว (postgres ยังถือรหัสเดิม) ไม่งั้นสุ่มใหม่
    r = kubectl("-n", NS, "get", "secret", "db", "-o", "jsonpath={.data.password}", check=False, capture=True)
    db_pw = base64.b64decode(r.stdout).decode() if r.returncode == 0 and r.stdout.strip() else secrets.token_hex(12)
    demo_pw = stored("demo-password", lambda: secrets.token_hex(12))
    make_secret("db", f"--from-literal=password={db_pw}",
                f"--from-literal=url=postgres://claims:{db_pw}@postgres:5432/claims?sslmode=disable")
    make_secret("app", f"--from-literal=demo-password={demo_pw}")
    make_secret("jwt-keys", f"--from-file=k1.key={keys / 'k1.key'}", f"--from-file=k1.pub={keys / 'k1.pub'}")
    make_configmap("migrations", f"--from-file={APP13 / 'migrations'}")
    make_configmap("kong-config", f"--from-file=kong.yml={kong}")

    # Job แก้ spec ไม่ได้ ลบของเก่าก่อน apply ใหม่
    kubectl("-n", NS, "delete", "job", "kafka-init", "migrate", "--ignore-not-found", "--wait=true")
    kubectl("apply", "-k", str(ROOT / "k8s"))
    kubectl("-n", NS, "wait", "--for=condition=complete", "job/kafka-init", "job/migrate", "--timeout=300s")
    for d in DEPLOYS:
        kubectl("-n", NS, "rollout", "status", f"deploy/{d}", "--timeout=300s")
    print("deploy ok")


def http(method, url, body=None, headers=None):
    req = urllib.request.Request(url, method=method, headers=headers or {},
                                 data=None if body is None else json.dumps(body).encode())
    if body is not None:
        req.add_header("Content-Type", "application/json")
    try:
        with urllib.request.urlopen(req, timeout=15) as r:
            return r.status, r.read().decode(), dict(r.headers)
    except urllib.error.HTTPError as e:
        return e.code, e.read().decode(), dict(e.headers)
    except (urllib.error.URLError, OSError) as e:
        return 0, str(e), {}


def cmd_smoke():
    """ทดสอบผ่าน Kong ใน cluster จริง: เหมือนที่ browser เห็น"""
    demo_pw = (WORK / "demo-password").read_text(encoding="ascii").strip()
    results = []

    def check(name, ok, detail=""):
        results.append(ok)
        print(("PASS" if ok else "FAIL"), name, ("- " + detail) if detail and not ok else "")

    code = 0
    for _ in range(60):  # Kong อาจยังโหลด config ไม่เสร็จหลัง rollout
        code, _, _ = http("GET", f"{GW}/api/claims")
        if code == 401:
            break
        time.sleep(1)
    check("gateway ตอบ และไม่มี token ถูกปฏิเสธ 401", code == 401, f"code={code}")

    code, body, _ = http("POST", f"{GW}/api/auth/login", {"username": "alice", "password": demo_pw})
    token = json.loads(body).get("access_token", "") if code == 200 else ""
    check("login alice ผ่าน Kong -> 200 + access token", code == 200 and token.count(".") == 2, f"code={code}")
    auth = {"Authorization": f"Bearer {token}"}

    code, _, _ = http("GET", f"{GW}/api/claims", headers=auth)
    check("GET /api/claims พร้อม token -> 200", code == 200, f"code={code}")

    key = f"smoke-{int(time.time())}"
    code, body, _ = http("POST", f"{GW}/api/claims", {"patient_ref": "FAKE-SMOKE1", "amount": 150000},
                         headers={**auth, "Idempotency-Key": key})
    cid = json.loads(body).get("id", "") if code in (200, 201) else ""
    check("POST /api/claims -> สร้างเคลม (201)", code == 201 and bool(cid), f"code={code} body={body[:120]}")

    status = ""
    for _ in range(40):
        c, b, _ = http("GET", f"{GW}/api/claims/{cid}", headers=auth)
        status = json.loads(b).get("status", "") if c == 200 else ""
        if status == "APPROVED":
            break
        time.sleep(1)
    check("worker ตัดสินเคลมเป็น APPROVED (api -> kafka -> worker -> postgres)", status == "APPROVED", f"status={status}")

    code, body, _ = http("GET", f"{WEB}/")
    check("web ส่ง index.html", code == 200 and '<div id="app">' in body, f"code={code}")

    code, _, hdr = http("OPTIONS", f"{GW}/api/claims", headers={
        "Origin": WEB, "Access-Control-Request-Method": "POST",
        "Access-Control-Request-Headers": "authorization,idempotency-key,content-type"})
    allow = {k.lower(): v for k, v in hdr.items()}.get("access-control-allow-origin", "")
    check("CORS preflight จาก web ผ่าน", allow == WEB, f"code={code} allow-origin={allow}")

    ok = all(results)
    print(f"smoke {'ok' if ok else 'FAILED'}: {sum(results)}/{len(results)}")
    return 0 if ok else 1


def cmd_down():
    run(["kind", "delete", "cluster", "--name", CLUSTER])


def main(argv):
    cmds = {"build": cmd_build, "up": cmd_up, "load": cmd_load, "deploy": cmd_deploy, "smoke": cmd_smoke, "down": cmd_down}
    if len(argv) != 2 or (argv[1] != "all" and argv[1] not in cmds):
        print(__doc__)
        return 2
    if argv[1] == "all":
        for n in ("build", "up", "load", "deploy"):
            cmds[n]()
        return cmd_smoke()
    return cmds[argv[1]]() or 0


if __name__ == "__main__":
    sys.exit(main(sys.argv))
