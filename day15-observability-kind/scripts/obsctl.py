"""ตัวคุมงาน Day 15: ติดตั้ง Prometheus + Grafana เข้า cluster ของ Day 14 แล้วตรวจว่าเห็น metrics จริง

ต้องมี cluster ของ Day 14 รันอยู่ (python day14-k8s-kind/scripts/kindctl.py all) และ kubectl ใน PATH
ใช้:  python scripts/obsctl.py <deploy|smoke|all>
ไฟล์ตั้งค่า (prometheus.yml ใน k8s/prometheus, alerts.yml และ dashboard ของ Day 13) ถูกส่งเข้า cluster เป็น ConfigMap
รหัส admin ของ Grafana สุ่มตอนรัน เก็บใน .work/ (git-ignore) และ Secret ของ cluster
"""
import base64
import json
import pathlib
import secrets
import subprocess
import sys
import re
import time
import urllib.error
import urllib.parse
import urllib.request

HERE = pathlib.Path(__file__).resolve().parent
ROOT = HERE.parent
APP13 = ROOT.parent / "day13-revocation-ui-cicd"
DAY14 = ROOT.parent / "day14-k8s-kind"
WORK = ROOT / ".work"
CLUSTER = "claim"
NS = "claims"
PROM = "http://127.0.0.1:19090"
GRAF = "http://127.0.0.1:13000"
GW = "http://localhost:8000"
ALERTS = APP13 / "deploy" / "prometheus" / "alerts.yml"
PROMQL_WORDS = {"histogram_quantile", "clamp_min"}
DASHBOARD = APP13 / "deploy" / "grafana" / "dashboards" / "claims-overview.json"
PROVIDER = APP13 / "deploy" / "grafana" / "provisioning" / "dashboards" / "dashboards.yml"


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


def make_configmap(name, *opts):
    out = kubectl("-n", NS, "create", "configmap", name, *opts, "--dry-run=client", "-o", "yaml", capture=True).stdout
    apply_yaml(out)


def cmd_deploy():
    make_configmap("prometheus-config",
                   f"--from-file=prometheus.yml={ROOT / 'k8s' / 'prometheus' / 'prometheus.yml'}",
                   f"--from-file=alerts.yml={ALERTS}")
    make_configmap("grafana-datasources", f"--from-file=datasources.yml={ROOT / 'k8s' / 'grafana' / 'datasources.yml'}")
    make_configmap("grafana-dashboard-provider", f"--from-file=dashboards.yml={PROVIDER}")
    make_configmap("grafana-dashboards", f"--from-file=claims-overview.json={DASHBOARD}")

    WORK.mkdir(exist_ok=True)
    pw_file = WORK / "grafana-password"
    if not pw_file.exists():
        pw_file.write_text(secrets.token_hex(12), encoding="ascii")
    pw = pw_file.read_text(encoding="ascii").strip()
    out = kubectl("-n", NS, "create", "secret", "generic", "grafana", f"--from-literal=admin-password={pw}",
                  "--dry-run=client", "-o", "yaml", capture=True).stdout
    apply_yaml(out)

    kubectl("apply", "-k", str(ROOT / "k8s"))
    # ConfigMap เปลี่ยนแล้ว pod เดิมไม่อ่านใหม่เอง รีสตาร์ตให้เห็นไฟล์ล่าสุด
    kubectl("-n", NS, "rollout", "restart", "deploy/prometheus", "deploy/grafana")
    for d in ("prometheus", "grafana"):
        kubectl("-n", NS, "rollout", "status", f"deploy/{d}", "--timeout=300s")
    print("deploy ok")


def http(method, url, body=None, headers=None, basic=None):
    req = urllib.request.Request(url, method=method, headers=headers or {},
                                 data=None if body is None else json.dumps(body).encode())
    if body is not None:
        req.add_header("Content-Type", "application/json")
    if basic:
        req.add_header("Authorization", "Basic " + base64.b64encode(basic.encode()).decode())
    try:
        with urllib.request.urlopen(req, timeout=15) as r:
            return r.status, r.read().decode()
    except urllib.error.HTTPError as e:
        return e.code, e.read().decode()
    except (urllib.error.URLError, OSError) as e:
        return 0, str(e)


def prom_query(expr):
    code, body = http("GET", PROM + "/api/v1/query?query=" + urllib.parse.quote(expr))
    if code != 200:
        return None
    return json.loads(body)["data"]["result"]


def wait_for(fn, seconds):
    for _ in range(seconds):
        v = fn()
        if v:
            return v
        time.sleep(1)
    return None


def exprs_in(dashboard):
    out = []
    for p in dashboard.get("panels", []):
        for t in p.get("targets", []):
            out.append(t["expr"])
        for sub in p.get("panels", []):
            for t in sub.get("targets", []):
                out.append(t["expr"])
    return out


def cmd_smoke():
    import yaml
    forwards = [
        subprocess.Popen(["kubectl", "--context", f"kind-{CLUSTER}", "-n", NS, "port-forward", "svc/prometheus", "19090:9090"],
                         stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL),
        subprocess.Popen(["kubectl", "--context", f"kind-{CLUSTER}", "-n", NS, "port-forward", "svc/grafana", "13000:3000"],
                         stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL),
    ]
    try:
        return smoke_checks(yaml)
    finally:
        for p in forwards:
            p.terminate()


def smoke_checks(yaml):
    results = []

    def check(name, ok, detail=""):
        results.append(ok)
        print(("PASS" if ok else "FAIL"), name, ("- " + detail) if detail and not ok else "")

    ready = wait_for(lambda: http("GET", PROM + "/-/ready")[0] == 200, 60)
    check("Prometheus พร้อม (/-/ready)", bool(ready))

    # สร้าง traffic ผ่าน gateway ให้มี metrics ให้อ่าน (ใช้ฟังก์ชัน http และรหัส demo ของ Day 14)
    sys.path.insert(0, str(DAY14 / "scripts"))
    import kindctl
    demo_pw = (DAY14 / ".work" / "demo-password").read_text(encoding="ascii").strip()
    code, body, _ = kindctl.http("POST", f"{GW}/api/auth/login", {"username": "alice", "password": demo_pw})
    token = json.loads(body).get("access_token", "") if code == 200 else ""
    auth = {"Authorization": f"Bearer {token}"}
    for _ in range(5):
        kindctl.http("GET", f"{GW}/api/claims", headers=auth)
    kindctl.http("POST", f"{GW}/api/auth/login", {"username": "alice", "password": "wrong-on-purpose"})
    code, body, _ = kindctl.http("POST", f"{GW}/api/claims", {"patient_ref": "FAKE-OBS1", "amount": 150000},
                                 headers={**auth, "Idempotency-Key": f"obs-{int(time.time())}"})
    check("สร้าง traffic ผ่าน gateway (login + GET + POST เคลม)", code == 201 and bool(token), f"login/post code={code}")

    def targets_up():
        c, b = http("GET", PROM + "/api/v1/targets")
        if c != 200:
            return None
        act = json.loads(b)["data"]["activeTargets"]
        jobs = {t["labels"]["job"]: t["health"] for t in act}
        return jobs if jobs.get("claims-api") == "up" and jobs.get("claims-worker") == "up" else None

    check("Prometheus scrape claims-api และ claims-worker ได้ (health=up)", bool(wait_for(targets_up, 90)))

    r = wait_for(lambda: prom_query("sum(http_requests_total)") or None, 60)
    total = float(r[0]["value"][1]) if r else 0
    check("เห็น http_requests_total จาก api (> 0)", total > 0, f"total={total}")

    r = wait_for(lambda: [x for x in (prom_query("sum(claims_decided_total)") or []) if float(x["value"][1]) > 0], 60)
    check("เห็น claims_decided_total จาก worker (> 0)", bool(r))

    r = prom_query('sum(auth_failures_total)') or []
    check("เห็น auth_failures_total หลัง login ผิดตั้งใจ (> 0)", bool(r) and float(r[0]["value"][1]) > 0)

    c, b = http("GET", PROM + "/api/v1/rules")
    loaded = set()
    if c == 200:
        for g in json.loads(b)["data"]["groups"]:
            loaded |= {x["name"] for x in g["rules"] if x["type"] == "alerting"}
    want = {r["alert"] for g in yaml.safe_load(ALERTS.read_text(encoding="utf-8"))["groups"] for r in g["rules"]}
    check(f"Prometheus โหลด alert ครบ {len(want)} ตัวจาก alerts.yml", loaded == want, f"ขาด={sorted(want - loaded)} เกิน={sorted(loaded - want)}")

    r = prom_query('ALERTS{alertstate="firing",alertname="ApiDown"}')
    check("ApiDown ไม่ลั่นตอนระบบปกติ", r == [])

    pw = (WORK / "grafana-password").read_text(encoding="ascii").strip()
    basic = f"admin:{pw}"
    ok = wait_for(lambda: http("GET", GRAF + "/api/health")[0] == 200, 60)
    check("Grafana พร้อม (/api/health)", bool(ok))
    c, b = http("GET", GRAF + "/api/search?query=Claim", basic=basic)
    uids = [d.get("uid") for d in json.loads(b)] if c == 200 else []
    check("Grafana มี dashboard claims-overview จาก provisioning", "claims-overview" in uids, f"code={c} uids={uids}")
    c, b = http("GET", GRAF + "/api/datasources/uid/prometheus/health", basic=basic)
    check("Grafana ต่อ Prometheus ได้ (datasource health)", c == 200 and '"OK"' in b, f"code={c} body={b[:100]}")

    # แค่รายงาน ไม่นับผ่าน/ไม่ผ่าน: metric ที่ dashboard ใช้ แต่ยังไม่มี series เลย (บางตัวยังไม่เคยเกิดเหตุ)
    names = set(json.loads(http("GET", PROM + "/api/v1/label/__name__/values")[1]).get("data", []))
    used = set()
    for e in exprs_in(json.loads(DASHBOARD.read_text(encoding="utf-8"))):
        used |= set(re.findall(r"\b([a-z][a-z0-9]*(?:_[a-z0-9]+)+)\b", e)) - PROMQL_WORDS
    used = {n[:-7] if n.endswith("_bucket") else n for n in used}
    gap = sorted(n for n in used if n not in names and n + "_bucket" not in names)
    print("INFO metric ใน dashboard ที่ Prometheus ยังไม่มี series:", gap or "ไม่มี")

    ok_all = all(results)
    print(f"smoke {'ok' if ok_all else 'FAILED'}: {sum(results)}/{len(results)}")
    return 0 if ok_all else 1


def main(argv):
    cmds = {"deploy": cmd_deploy, "smoke": cmd_smoke}
    if len(argv) != 2 or (argv[1] != "all" and argv[1] not in cmds):
        print(__doc__)
        return 2
    if argv[1] == "all":
        cmd_deploy()
        return cmd_smoke()
    return cmds[argv[1]]() or 0


if __name__ == "__main__":
    sys.exit(main(sys.argv))
