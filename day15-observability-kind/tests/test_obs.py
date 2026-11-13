"""เทสต์ static ของ Day 15 (ไม่ต้องมี docker หรือ cluster): ตรวจว่าไฟล์ observability ยังเชื่อมกันครบ

รัน:  python -m unittest discover -s tests -v
ตรวจ: manifest ของ Prometheus/Grafana, Service metrics ชี้ตรงกับ pod ของ Day 14, ชื่อ job/พอร์ต/ConfigMap/Secret ตรงกัน,
ทุก alert ของ Day 13 มี runbook และมีเทสต์ promtool ทั้งเคสลั่นกับไม่ลั่น, metric ที่ alert/dashboard ใช้มีอยู่จริงในโค้ด Go
ส่วน promtool เองรันใน CI (ที่นี่ไม่มี promtool)

tests/baseline/obs-summary.json คือ snapshot ของสิ่งที่ไฟล์พวกนี้ทำตอนนี้
ถ้าแก้ตั้งใจ:  UPDATE_BASELINE=1 python -m unittest discover -s tests  แล้วอ่าน diff ของไฟล์นั้นก่อน commit
"""
import json
import os
import pathlib
import re
import unittest

import yaml

ROOT = pathlib.Path(__file__).resolve().parent.parent
APP13 = ROOT.parent / "day13-revocation-ui-cicd"
DAY14 = ROOT.parent / "day14-k8s-kind"
PROMQL_WORDS = {"histogram_quantile", "clamp_min"}


def load_all(path):
    return [d for d in yaml.safe_load_all(path.read_text(encoding="utf-8")) if d]


OBS = load_all(ROOT / "k8s" / "observability.yaml")
ALERTS = yaml.safe_load((APP13 / "deploy" / "prometheus" / "alerts.yml").read_text(encoding="utf-8"))
PROM_CFG = yaml.safe_load((ROOT / "k8s" / "prometheus" / "prometheus.yml").read_text(encoding="utf-8"))
PROM13 = yaml.safe_load((APP13 / "deploy" / "prometheus" / "prometheus.yml").read_text(encoding="utf-8"))
PROMTOOL = yaml.safe_load((ROOT / "promtool" / "tests.yml").read_text(encoding="utf-8"))
DASH = json.loads((APP13 / "deploy" / "grafana" / "dashboards" / "claims-overview.json").read_text(encoding="utf-8"))
OBSCTL = (ROOT / "scripts" / "obsctl.py").read_text(encoding="utf-8")
RUNBOOK = (ROOT / "RUNBOOK.md").read_text(encoding="utf-8")
COMPOSE = (APP13 / "docker-compose.yml").read_text(encoding="utf-8")


def of_kind(kind):
    return [d for d in OBS if d["kind"] == kind]


def by_name(kind, name):
    return next(d for d in of_kind(kind) if d["metadata"]["name"] == name)


def alert_rules():
    return [r for g in ALERTS["groups"] for r in g["rules"]]


def dash_exprs():
    out = []
    for p in DASH["panels"]:
        out += [t["expr"] for t in p.get("targets", [])]
        for sub in p.get("panels", []):
            out += [t["expr"] for t in sub.get("targets", [])]
    return out


def metric_names(exprs):
    used = set()
    for e in exprs:
        used |= set(re.findall(r"\b([a-z][a-z0-9]*(?:_[a-z0-9]+)+)\b", e)) - PROMQL_WORDS
    return {n[:-7] if n.endswith("_bucket") else n for n in used}


def go_source():
    parts = []
    for p in APP13.rglob("*.go"):
        if not p.name.endswith("_test.go"):
            parts.append(p.read_text(encoding="utf-8"))
    return "\n".join(parts)


class TestManifests(unittest.TestCase):
    def test_everything_in_claims_namespace(self):
        for d in OBS:
            self.assertEqual(d["metadata"].get("namespace"), "claims", d["metadata"]["name"])

    def test_containers_are_hardened(self):
        for d in of_kind("Deployment"):
            spec = d["spec"]["template"]["spec"]
            self.assertTrue(spec["securityContext"]["runAsNonRoot"], d["metadata"]["name"])
            self.assertEqual(spec["securityContext"]["seccompProfile"]["type"], "RuntimeDefault")
            for c in spec["containers"]:
                self.assertFalse(c["securityContext"]["allowPrivilegeEscalation"], c["name"])
                self.assertEqual(c["securityContext"]["capabilities"]["drop"], ["ALL"], c["name"])
                self.assertIn("memory", c["resources"]["limits"], c["name"])
                self.assertIn("requests", c["resources"], c["name"])

    def test_images_are_pinned_and_match_day13_compose(self):
        for d in of_kind("Deployment"):
            for c in d["spec"]["template"]["spec"]["containers"]:
                img = c["image"]
                self.assertNotIn(":latest", img)
                self.assertIn(":", img, f"{img} ไม่มี tag")
                self.assertIn(img, COMPOSE, f"{img} ไม่ตรงกับที่ Day 13 compose ใช้")

    def test_probes_exist(self):
        for d in of_kind("Deployment"):
            for c in d["spec"]["template"]["spec"]["containers"]:
                self.assertIn("readinessProbe", c, c["name"])
                self.assertIn("livenessProbe", c, c["name"])

    def test_no_secret_values_in_repo_files(self):
        text = (ROOT / "k8s" / "observability.yaml").read_text(encoding="utf-8")
        self.assertNotIn("kind: Secret", text)
        pw = by_name("Deployment", "grafana")["spec"]["template"]["spec"]["containers"][0]["env"]
        admin = next(e for e in pw if e["name"] == "GF_SECURITY_ADMIN_PASSWORD")
        self.assertIn("secretKeyRef", admin["valueFrom"])
        self.assertNotIn("value", admin)

    def test_grafana_anonymous_access_off(self):
        env = {e["name"]: e.get("value") for e in by_name("Deployment", "grafana")["spec"]["template"]["spec"]["containers"][0]["env"]}
        self.assertEqual(env["GF_AUTH_ANONYMOUS_ENABLED"], "false")
        self.assertEqual(env["GF_USERS_ALLOW_SIGN_UP"], "false")


class TestWiring(unittest.TestCase):
    def day14(self, name):
        return [d for d in yaml.safe_load_all((DAY14 / "k8s" / name).read_text(encoding="utf-8")) if d]

    def test_metrics_services_select_day14_pods(self):
        for svc, app in (("api-metrics", "api"), ("worker-metrics", "worker")):
            sel = by_name("Service", svc)["spec"]["selector"]
            self.assertEqual(sel, {"app": app})
            dep = next(d for d in self.day14(f"{app}.yaml") if d["kind"] == "Deployment")
            self.assertEqual(dep["spec"]["template"]["metadata"]["labels"]["app"], app)

    def test_metrics_port_matches_day14_config(self):
        port = by_name("Service", "api-metrics")["spec"]["ports"][0]["port"]
        cfg = self.day14("config.yaml")
        addr = next(d for d in cfg if d["metadata"]["name"] == "api-config")["data"]["METRICS_ADDR"]
        self.assertEqual(addr, f":{port}")
        self.assertEqual(by_name("Service", "worker-metrics")["spec"]["ports"][0]["port"], port)

    def test_scrape_targets_are_the_metrics_services(self):
        services = {d["metadata"]["name"]: d["spec"]["ports"][0]["port"] for d in of_kind("Service")}
        for sc in PROM_CFG["scrape_configs"]:
            for t in sc["static_configs"][0]["targets"]:
                host, port = t.split(":")
                self.assertEqual(services.get(host), int(port), t)

    def test_job_names_same_as_day13(self):
        self.assertEqual([s["job_name"] for s in PROM_CFG["scrape_configs"]], [s["job_name"] for s in PROM13["scrape_configs"]])

    def test_alerts_only_use_jobs_we_scrape(self):
        jobs = {s["job_name"] for s in PROM_CFG["scrape_configs"]}
        for r in alert_rules():
            for j in re.findall(r'job="([^"]+)"', r["expr"]):
                self.assertIn(j, jobs, r["alert"])

    def test_rule_file_path_matches_mount(self):
        prom = by_name("Deployment", "prometheus")["spec"]["template"]["spec"]["containers"][0]
        mount = next(m for m in prom["volumeMounts"] if m["name"] == "config")["mountPath"]
        for rf in PROM_CFG["rule_files"]:
            self.assertEqual(str(pathlib.PurePosixPath(rf).parent), mount)
        self.assertIn(f"--config.file={mount}/prometheus.yml", prom["args"])

    def test_grafana_datasource_matches_prometheus_service_and_dashboard(self):
        ds = yaml.safe_load((ROOT / "k8s" / "grafana" / "datasources.yml").read_text(encoding="utf-8"))["datasources"]
        self.assertEqual(len(ds), 1)
        self.assertEqual(ds[0]["url"], "http://prometheus:9090")
        self.assertEqual(by_name("Service", "prometheus")["spec"]["ports"][0]["port"], 9090)
        uids = {p["datasource"]["uid"] for p in DASH["panels"] if "datasource" in p}
        self.assertEqual(uids, {ds[0]["uid"]})

    def test_dashboard_provider_path_matches_mount(self):
        prov = yaml.safe_load((APP13 / "deploy" / "grafana" / "provisioning" / "dashboards" / "dashboards.yml").read_text(encoding="utf-8"))
        path = prov["providers"][0]["options"]["path"]
        graf = by_name("Deployment", "grafana")["spec"]["template"]["spec"]["containers"][0]
        self.assertEqual(next(m for m in graf["volumeMounts"] if m["name"] == "dashboards")["mountPath"], path)

    def test_configmaps_and_secret_are_created_by_obsctl(self):
        names = set()
        for d in of_kind("Deployment"):
            for v in d["spec"]["template"]["spec"]["volumes"]:
                if "configMap" in v:
                    names.add(v["configMap"]["name"])
        self.assertEqual(names, {"prometheus-config", "grafana-datasources", "grafana-dashboard-provider", "grafana-dashboards"})
        for n in names:
            self.assertIn(f'make_configmap("{n}"', OBSCTL)
        self.assertIn('"secret", "generic", "grafana"', OBSCTL)

    def test_obsctl_uses_the_same_ports_as_service(self):
        self.assertIn("19090:9090", OBSCTL)
        self.assertIn("13000:3000", OBSCTL)


class TestRules(unittest.TestCase):
    def test_every_alert_has_runbook_section(self):
        for r in alert_rules():
            self.assertRegex(RUNBOOK, rf"(?m)^## {r['alert']}$", f"runbook ไม่มีหัวข้อ {r['alert']}")

    def test_runbook_has_no_heading_for_unknown_alert(self):
        names = {r["alert"] for r in alert_rules()}
        for h in re.findall(r"(?m)^## (\w+)$", RUNBOOK):
            self.assertIn(h, names, f"runbook มีหัวข้อ {h} ที่ไม่มี alert")

    def test_every_alert_has_firing_and_quiet_promtool_case(self):
        fired, quiet = set(), set()
        for t in PROMTOOL["tests"]:
            for a in t["alert_rule_test"]:
                (fired if a["exp_alerts"] else quiet).add(a["alertname"])
        for r in alert_rules():
            self.assertIn(r["alert"], fired, f"ไม่มีเคสที่ {r['alert']} ต้องลั่น")
            self.assertIn(r["alert"], quiet, f"ไม่มีเคสที่ {r['alert']} ต้องไม่ลั่น")

    def test_promtool_annotations_and_severity_match_alerts_yml(self):
        rules = {r["alert"]: r for r in alert_rules()}
        for t in PROMTOOL["tests"]:
            for a in t["alert_rule_test"]:
                for e in a["exp_alerts"]:
                    r = rules[a["alertname"]]
                    self.assertEqual(e["exp_annotations"]["summary"], r["annotations"]["summary"], a["alertname"])
                    self.assertEqual(e["exp_labels"]["severity"], r["labels"]["severity"], a["alertname"])

    def test_promtool_tests_reference_alerts_yml(self):
        self.assertEqual(PROMTOOL["rule_files"], ["alerts.yml"])
        self.assertEqual(PROMTOOL["evaluation_interval"], "1m")

    def test_metrics_used_by_alerts_and_dashboard_exist_in_go_code(self):
        src = go_source()
        names = metric_names([r["expr"] for r in alert_rules()] + dash_exprs())
        self.assertGreater(len(names), 15)
        for n in sorted(names):
            self.assertIn(f'"{n}"', src, f"metric {n} ไม่มีในโค้ด Go")

    def test_dashboard_panels_have_titles(self):
        for p in DASH["panels"]:
            self.assertTrue(p.get("title"), p)


class TestWorkflow(unittest.TestCase):
    def setUp(self):
        self.text = (ROOT / "ci" / "day15.yml").read_text(encoding="utf-8")
        self.wf = yaml.safe_load(self.text)

    def test_images_pinned(self):
        self.assertNotIn(":latest", self.text)
        self.assertIn("prom/prometheus:v2.55.1", self.text)

    def test_only_expected_actions(self):
        uses = set(re.findall(r"(?m)^\s*-?\s*uses:\s*(\S+)", self.text))
        self.assertEqual(uses, {"actions/checkout@v4", "helm/kind-action@v1"})

    def test_jobs_and_paths(self):
        self.assertEqual(set(self.wf["jobs"]), {"static", "kind"})
        self.assertEqual(self.wf["jobs"]["kind"]["needs"], ["static"])
        self.assertEqual(self.wf["permissions"], {"contents": "read"})
        self.assertIn("day15-observability-kind/**", self.wf[True]["push"]["paths"])

    def test_promtool_commands_present(self):
        for cmd in ("check config", "check rules", "test rules"):
            self.assertIn(cmd, self.text)

    def test_kind_job_brings_up_day14_first(self):
        self.assertLess(self.text.index("kindctl.py all"), self.text.index("obsctl.py all"))

    def test_repo_copy_matches_if_present(self):
        repo_copy = ROOT.parent / ".github" / "workflows" / "day15.yml"
        if not repo_copy.exists():
            self.skipTest("ยังไม่ได้วาง .github/workflows/day15.yml")
        self.assertEqual(repo_copy.read_text(encoding="utf-8").split("\n", 2)[2], self.text.split("\n", 2)[2])


def summary():
    deps = {}
    for d in of_kind("Deployment"):
        c = d["spec"]["template"]["spec"]["containers"][0]
        deps[d["metadata"]["name"]] = {"image": c["image"], "ports": [p["containerPort"] for p in c["ports"]],
                                       "memory_limit": c["resources"]["limits"]["memory"]}
    return {
        "deployments": deps,
        "services": {d["metadata"]["name"]: [p["port"] for p in d["spec"]["ports"]] for d in of_kind("Service")},
        "scrape_targets": {s["job_name"]: s["static_configs"][0]["targets"] for s in PROM_CFG["scrape_configs"]},
        "alerts": {r["alert"]: r["labels"]["severity"] for r in alert_rules()},
        "dashboard_metrics": sorted(metric_names(dash_exprs())),
        "alert_metrics": sorted(metric_names([r["expr"] for r in alert_rules()])),
        "promtool_cases": len(PROMTOOL["tests"]),
    }


class TestBaseline(unittest.TestCase):
    def test_summary_matches_snapshot(self):
        path = ROOT / "tests" / "baseline" / "obs-summary.json"
        now = json.dumps(summary(), indent=2, sort_keys=True, ensure_ascii=False) + "\n"
        if os.environ.get("UPDATE_BASELINE"):
            path.write_text(now, encoding="utf-8")
        self.assertEqual(path.read_text(encoding="utf-8"), now,
                         "ไฟล์ observability เปลี่ยนจาก baseline: ถ้าตั้งใจ ให้รันใหม่ด้วย UPDATE_BASELINE=1 แล้วอ่าน diff")


if __name__ == "__main__":
    unittest.main()
