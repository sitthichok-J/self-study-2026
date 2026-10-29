"""ตรวจความสอดคล้องของ Dockerfile / compose / k8s / CI กับโค้ด Day 5

รัน: python3 -m unittest discover -s tests -v   (ต้อง pip install pyyaml)
baseline: testdata/baseline.json อัปเดตด้วย UPDATE_BASELINE=1
"""
import json
import os
import re
import unittest
from pathlib import Path

import yaml

ROOT = Path(__file__).resolve().parent.parent
SVC = ROOT.parent / "day05-go-service"


def load_all(path):
    return [d for d in yaml.safe_load_all(path.read_text(encoding="utf-8")) if d]


def k8s_docs():
    docs = []
    for f in sorted((ROOT / "k8s").glob("*.yaml")):
        if f.name == "kustomization.yaml":
            continue
        docs += load_all(f)
    return docs


def find(docs, kind, name=None):
    for d in docs:
        if d["kind"] == kind and (name is None or d["metadata"]["name"] == name):
            return d
    raise AssertionError(f"ไม่พบ {kind} {name}")


def config_env_names():
    src = (SVC / "internal/config/config.go").read_text(encoding="utf-8")
    return set(re.findall(r'"([A-Z][A-Z_]{2,})"', src))


def parse_duration_s(s):
    m = re.fullmatch(r"(\d+)(s|m)", s)
    assert m, s
    return int(m.group(1)) * (60 if m.group(2) == "m" else 1)


class TestDockerfile(unittest.TestCase):
    def setUp(self):
        self.text = (ROOT / "Dockerfile").read_text(encoding="utf-8")
        self.lines = [l.strip() for l in self.text.splitlines() if l.strip() and not l.strip().startswith("#")]

    def test_multi_stage(self):
        self.assertEqual(len([l for l in self.lines if l.startswith("FROM ")]), 2)

    def test_no_latest_tag(self):
        for l in self.lines:
            if l.startswith("FROM "):
                self.assertNotRegex(l, r":latest\b")
                self.assertIn(":", l.split()[1], "ต้องระบุ tag")

    def test_runs_as_nonroot(self):
        self.assertIn("USER nonroot", self.lines)

    def test_cache_friendly_order(self):
        i_mod = self.lines.index("COPY go.mod go.sum ./")
        i_all = self.lines.index("COPY . .")
        self.assertLess(i_mod, i_all)
        self.assertIn("RUN go mod download", self.lines[i_mod:i_all])

    def test_builds_with_pgx_tag_and_static(self):
        build = [l for l in self.lines if l.startswith("RUN") and "go build" in l][0]
        self.assertIn("-tags pgx", build)
        self.assertIn("CGO_ENABLED=0", build)  # distroless/static ไม่มี libc

    def test_no_secret_in_image(self):
        self.assertNotRegex(self.text, r"(?i)\bENV\b.*(PASSWORD|SECRET|TOKEN)")

    def test_entrypoint_matches_binary(self):
        self.assertIn('ENTRYPOINT ["/api"]', self.lines)
        self.assertTrue((SVC / "cmd/api/main.go").exists())

    def test_exposed_port_matches_default_addr(self):
        cfg = (SVC / "internal/config/config.go").read_text(encoding="utf-8")
        port = re.search(r'or\(getenv\("ADDR"\), ":(\d+)"\)', cfg).group(1)
        self.assertIn(f"EXPOSE {port}", self.lines)

    def test_dockerignore_exists(self):
        self.assertTrue((ROOT / "Dockerfile.dockerignore").exists())


class TestCompose(unittest.TestCase):
    def setUp(self):
        self.c = load_all(ROOT / "docker-compose.yml")[0]
        self.api = self.c["services"]["api"]
        self.db = self.c["services"]["db"]

    def test_services(self):
        self.assertEqual(set(self.c["services"]), {"db", "api"})

    def test_api_waits_for_healthy_db(self):
        self.assertEqual(self.api["depends_on"]["db"]["condition"], "service_healthy")
        self.assertIn("healthcheck", self.db)

    def test_database_url_uses_service_name(self):
        url = self.api["environment"]["DATABASE_URL"]
        self.assertRegex(url, r"@db:5432/claims")

    def test_env_names_are_known_to_config(self):
        known = config_env_names()
        for k in self.api["environment"]:
            self.assertIn(k, known, f"{k} ไม่มีใน config.go")

    def test_build_context_points_to_service(self):
        ctx = (ROOT / self.api["build"]["context"]).resolve()
        self.assertEqual(ctx, SVC.resolve())
        self.assertTrue((ctx / self.api["build"]["dockerfile"]).resolve().exists())

    def test_db_data_is_persistent(self):
        self.assertIn("pgdata", self.c["volumes"])
        self.assertTrue(any(v.startswith("pgdata:") for v in self.db["volumes"]))

    def test_init_sql_is_the_real_migration(self):
        mount = [v for v in self.db["volumes"] if "docker-entrypoint-initdb.d" in v][0]
        src = (ROOT / mount.split(":")[0]).resolve()
        self.assertEqual(src, (SVC / "migrations/000001_create_claims.up.sql").resolve())

    def test_stop_grace_period_exceeds_shutdown_timeout(self):
        timeout = parse_duration_s(self.api["environment"]["SHUTDOWN_TIMEOUT"])
        self.assertGreater(parse_duration_s(self.api["stop_grace_period"]), timeout)


class TestKubernetes(unittest.TestCase):
    def setUp(self):
        self.docs = k8s_docs()
        self.dep = find(self.docs, "Deployment", "claims-api")
        self.pod = self.dep["spec"]["template"]["spec"]
        self.ctr = self.pod["containers"][0]
        self.cm = find(self.docs, "ConfigMap", "claims-api-config")

    def test_replicas_and_zero_downtime(self):
        self.assertEqual(self.dep["spec"]["replicas"], 2)
        self.assertEqual(self.dep["spec"]["strategy"]["rollingUpdate"]["maxUnavailable"], 0)

    def test_selector_matches_labels(self):
        sel = self.dep["spec"]["selector"]["matchLabels"]
        labels = self.dep["spec"]["template"]["metadata"]["labels"]
        self.assertEqual(sel.items() <= labels.items(), True)
        svc = find(self.docs, "Service", "claims-api")
        self.assertEqual(svc["spec"]["selector"], sel)

    def test_probes_use_existing_route_and_port(self):
        handler = (SVC / "internal/handler/handler.go").read_text(encoding="utf-8")
        port = self.ctr["ports"][0]["containerPort"]
        for name in ("startupProbe", "readinessProbe", "livenessProbe"):
            p = self.ctr[name]["httpGet"]
            self.assertIn(f'"GET {p["path"]}"', handler, f"{name}: ไม่มี route {p['path']}")
            self.assertEqual(p["port"], port)

    def test_container_port_matches_addr(self):
        addr = self.cm["data"]["ADDR"]
        self.assertEqual(int(addr.split(":")[1]), self.ctr["ports"][0]["containerPort"])
        svc = find(self.docs, "Service", "claims-api")
        self.assertEqual(svc["spec"]["ports"][0]["targetPort"], self.ctr["ports"][0]["containerPort"])

    def test_resources_set(self):
        r = self.ctr["resources"]
        for k in ("cpu", "memory"):
            self.assertIn(k, r["requests"])
            self.assertIn(k, r["limits"])

    def test_security_context(self):
        s = self.ctr["securityContext"]
        self.assertTrue(s["runAsNonRoot"])
        self.assertFalse(s["allowPrivilegeEscalation"])
        self.assertTrue(s["readOnlyRootFilesystem"])
        self.assertEqual(s["capabilities"]["drop"], ["ALL"])

    def test_grace_period_exceeds_shutdown_timeout(self):
        timeout = parse_duration_s(self.cm["data"]["SHUTDOWN_TIMEOUT"])
        self.assertGreater(self.pod["terminationGracePeriodSeconds"], timeout)

    def test_env_names_known_to_config(self):
        known = config_env_names()
        for k in self.cm["data"]:
            self.assertIn(k, known, f"{k} ไม่มีใน config.go")
        for e in self.ctr["env"]:
            self.assertIn(e["name"], known)

    def test_database_url_comes_from_secret_with_service_host(self):
        ref = [e for e in self.ctr["env"] if e["name"] == "DATABASE_URL"][0]["valueFrom"]["secretKeyRef"]
        sec = find(self.docs, "Secret", ref["name"])
        self.assertIn(ref["key"], sec["stringData"])
        self.assertRegex(sec["stringData"][ref["key"]], r"@postgres:5432/claims")
        find(self.docs, "Service", "postgres")

    def test_image_is_not_latest(self):
        self.assertNotRegex(self.ctr["image"], r":latest$")
        self.assertIn(":", self.ctr["image"])

    def test_kustomization_lists_every_manifest(self):
        k = load_all(ROOT / "k8s/kustomization.yaml")[0]
        files = {f.name for f in (ROOT / "k8s").glob("*.yaml")} - {"kustomization.yaml"}
        self.assertEqual(set(k["resources"]), files)
        self.assertEqual(k["namespace"], "claims")

    def test_initdb_copy_equals_migration(self):
        a = (ROOT / "k8s/initdb/001_create_claims.sql").read_text(encoding="utf-8")
        b = (SVC / "migrations/000001_create_claims.up.sql").read_text(encoding="utf-8")
        self.assertEqual(a, b, "k8s/initdb ไม่ตรงกับ migration (copy ใหม่)")
        k = load_all(ROOT / "k8s/kustomization.yaml")[0]
        self.assertIn("initdb/001_create_claims.sql", k["configMapGenerator"][0]["files"])


class TestCI(unittest.TestCase):
    def test_github_workflow(self):
        w = load_all(ROOT / "ci/github-actions-ci.yml")[0]
        steps = json.dumps(w["jobs"]["test"]["steps"])
        for needle in ("gofmt -l", "go vet", "go test -race", "integration pgx"):
            self.assertIn(needle, steps)
        self.assertEqual(w["jobs"]["image"]["needs"], "test")
        self.assertIn("postgres:16", w["jobs"]["test"]["services"]["postgres"]["image"])

    def test_gitlab_ci(self):
        g = load_all(ROOT / "ci/gitlab-ci.yml")[0]
        self.assertEqual(g["stages"], ["lint", "test", "build"])
        self.assertIn("go test -race", json.dumps(g["test"]["script"]))
        self.assertIn("go vet", json.dumps(g["lint"]["script"]))

    def test_no_hardcoded_registry_secret(self):
        text = (ROOT / "ci/gitlab-ci.yml").read_text(encoding="utf-8")
        self.assertIn("$CI_REGISTRY_PASSWORD", text)
        self.assertNotRegex(text, r"(?i)password:\s*\S+\n.*registry")


class TestBaseline(unittest.TestCase):
    """snapshot ของโครงสร้างทั้งหมด: ถ้าเปลี่ยนโดยตั้งใจให้รันด้วย UPDATE_BASELINE=1"""

    def snapshot(self):
        docs = k8s_docs()
        dep = find(docs, "Deployment", "claims-api")
        ctr = dep["spec"]["template"]["spec"]["containers"][0]
        comp = load_all(ROOT / "docker-compose.yml")[0]
        return {
            "k8s_kinds": sorted(f'{d["kind"]}/{d["metadata"]["name"]}' for d in docs),
            "api_replicas": dep["spec"]["replicas"],
            "api_probes": {n: ctr[n]["httpGet"]["path"] for n in ("startupProbe", "readinessProbe", "livenessProbe")},
            "api_resources": ctr["resources"],
            "compose_services": sorted(comp["services"]),
            "compose_ports": {n: s.get("ports") for n, s in comp["services"].items()},
            "dockerfile_from": [l.split()[1] for l in (ROOT / "Dockerfile").read_text().splitlines() if l.startswith("FROM ")],
        }

    def test_baseline(self):
        path = ROOT / "testdata/baseline.json"
        got = self.snapshot()
        if os.environ.get("UPDATE_BASELINE"):
            path.write_text(json.dumps(got, indent=2, ensure_ascii=False, sort_keys=True) + "\n", encoding="utf-8")
        want = json.loads(path.read_text(encoding="utf-8"))
        self.assertEqual(got, want)


if __name__ == "__main__":
    unittest.main()
