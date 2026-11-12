"""Static checks for the Day 14 Kubernetes files (no Docker, no cluster needed).

Run:  python -m unittest discover -s tests -v
They check that the manifests still say what we think they say: probes, limits,
secrets kept out of ConfigMaps, ports lining up with kind-config and the Day 13
compose file, and that every probe path really exists in the Go code.

tests/baseline/k8s-summary.json is the snapshot of what the manifests do today.
If you change a manifest on purpose:  UPDATE_BASELINE=1 python -m unittest discover -s tests
and read the diff of that file before committing.
"""
import json
import os
import pathlib
import re
import sys
import unittest

import yaml

ROOT = pathlib.Path(__file__).resolve().parent.parent
APP13 = ROOT.parent / "day13-revocation-ui-cicd"
sys.path.insert(0, str(ROOT / "scripts"))
import render_kong  # noqa: E402


def docs(name):
    return [d for d in yaml.safe_load_all((ROOT / "k8s" / name).read_text(encoding="utf-8")) if d]


def all_docs():
    out = []
    for p in sorted((ROOT / "k8s").glob("*.yaml")):
        if p.name != "kustomization.yaml":
            out += docs(p.name)
    return out


def of_kind(kind):
    return [d for d in all_docs() if d["kind"] == kind]


def containers(dep):
    return dep["spec"]["template"]["spec"]["containers"]


def app_deployments():
    return [d for d in of_kind("Deployment") if d["metadata"]["name"] in ("api", "worker", "web")]


def env_keys(dep_name):
    dep = next(d for d in of_kind("Deployment") if d["metadata"]["name"] == dep_name)
    keys = {e["name"] for c in containers(dep) for e in c.get("env", [])}
    for c in containers(dep):
        for ef in c.get("envFrom", []):
            cm = next(x for x in of_kind("ConfigMap") if x["metadata"]["name"] == ef["configMapRef"]["name"])
            keys |= set(cm["data"])
    return keys


def summary():
    """What each workload does today, in a form that is stable to diff."""
    out = {}
    for d in of_kind("Deployment"):
        spec = d["spec"]["template"]["spec"]
        c = containers(d)[0]
        refs = sorted({e["valueFrom"]["secretKeyRef"]["name"] + "/" + e["valueFrom"]["secretKeyRef"]["key"]
                       for e in c.get("env", []) if "secretKeyRef" in e.get("valueFrom", {})})
        out[d["metadata"]["name"]] = {
            "image": c["image"],
            "replicas": d["spec"]["replicas"],
            "ports": [p["containerPort"] for p in c.get("ports", [])],
            "readiness": c.get("readinessProbe"),
            "liveness": c.get("livenessProbe"),
            "requests": c["resources"]["requests"],
            "limits": c["resources"]["limits"],
            "secretRefs": refs,
            "runAsNonRoot": spec.get("securityContext", {}).get("runAsNonRoot", False),
            "readOnlyRootFilesystem": c.get("securityContext", {}).get("readOnlyRootFilesystem", False),
            "strategy": d["spec"].get("strategy", {}).get("type", "RollingUpdate"),
        }
    out["_services"] = {s["metadata"]["name"]: {"type": s["spec"].get("type", "ClusterIP"),
                                                "ports": s["spec"]["ports"]} for s in of_kind("Service")}
    return out


class TestManifestHygiene(unittest.TestCase):
    def test_every_file_parses_and_has_the_claims_namespace(self):
        for d in all_docs():
            if d["kind"] != "Namespace":
                self.assertEqual(d["metadata"].get("namespace"), "claims", d["metadata"]["name"])

    def test_kustomization_lists_every_manifest_file(self):
        k = yaml.safe_load((ROOT / "k8s" / "kustomization.yaml").read_text(encoding="utf-8"))
        files = {p.name for p in (ROOT / "k8s").glob("*.yaml")} - {"kustomization.yaml"}
        self.assertEqual(set(k["resources"]), files)

    def test_no_latest_or_untagged_images(self):
        for d in of_kind("Deployment") + of_kind("Job"):
            for c in d["spec"]["template"]["spec"]["containers"]:
                self.assertRegex(c["image"], r":[A-Za-z0-9._-]+$", c["image"])
                self.assertFalse(c["image"].endswith(":latest"), c["image"])

    def test_every_deployment_has_readiness_probe_and_resources(self):
        for d in of_kind("Deployment"):
            for c in containers(d):
                self.assertIn("readinessProbe", c, d["metadata"]["name"])
                self.assertIn("memory", c["resources"]["limits"], d["metadata"]["name"])
                self.assertIn("cpu", c["resources"]["requests"], d["metadata"]["name"])
                self.assertIn("memory", c["resources"]["requests"], d["metadata"]["name"])

    def test_app_deployments_also_have_liveness_probe(self):
        for d in app_deployments() + [x for x in of_kind("Deployment") if x["metadata"]["name"] == "gateway"]:
            self.assertIn("livenessProbe", containers(d)[0], d["metadata"]["name"])

    def test_app_containers_are_locked_down(self):
        for d in app_deployments():
            spec = d["spec"]["template"]["spec"]
            c = containers(d)[0]
            self.assertTrue(spec["securityContext"]["runAsNonRoot"], d["metadata"]["name"])
            self.assertFalse(c["securityContext"]["allowPrivilegeEscalation"])
            self.assertEqual(c["securityContext"]["capabilities"]["drop"], ["ALL"])
        for name in ("api", "worker", "web"):  # root filesystem read-only (web gets an emptyDir at /tmp)
            d = next(x for x in app_deployments() if x["metadata"]["name"] == name)
            self.assertTrue(containers(d)[0]["securityContext"]["readOnlyRootFilesystem"])

    def test_service_links_are_off(self):
        # Service "kafka" would inject KAFKA_PORT and stop apache/kafka from starting
        for d in of_kind("Deployment") + of_kind("Job"):
            self.assertFalse(d["spec"]["template"]["spec"]["enableServiceLinks"], d["metadata"]["name"])

    def test_app_images_use_the_day14_tag_that_kindctl_loads(self):
        import kindctl
        want = {f"{n}:{kindctl.TAG}" for n in kindctl.IMAGES}
        got = {containers(d)[0]["image"] for d in app_deployments()}
        self.assertEqual(want, got)

    def test_kindctl_waits_for_exactly_the_deployments_we_ship(self):
        import kindctl
        self.assertEqual(set(kindctl.DEPLOYS), {d["metadata"]["name"] for d in of_kind("Deployment")})


class TestSecretsStayOutOfRepo(unittest.TestCase):
    def test_configmaps_hold_no_secret_looking_keys(self):
        for cm in of_kind("ConfigMap"):
            for k in cm["data"]:
                self.assertNotRegex(k, r"(?i)PASSWORD|SECRET|TOKEN|DATABASE_URL|PRIVATE", k)

    def test_no_kind_secret_is_committed(self):
        self.assertEqual(of_kind("Secret"), [])

    def test_secrets_used_are_the_three_deploy_creates(self):
        names = set()
        for d in of_kind("Deployment") + of_kind("Job"):
            spec = d["spec"]["template"]["spec"]
            for c in spec["containers"]:
                for e in c.get("env", []):
                    if "secretKeyRef" in e.get("valueFrom", {}):
                        names.add(e["valueFrom"]["secretKeyRef"]["name"])
            for v in spec.get("volumes", []):
                if "secret" in v:
                    names.add(v["secret"]["secretName"])
        self.assertEqual(names, {"db", "app", "jwt-keys"})
        src = (ROOT / "scripts" / "kindctl.py").read_text(encoding="utf-8")
        for n in names:
            self.assertIn(f'make_secret("{n}"', src)

    def test_work_dir_is_git_ignored(self):
        self.assertIn(".work/", (ROOT / ".gitignore").read_text(encoding="utf-8").split())

    def test_no_private_key_or_hardcoded_password_in_files(self):
        for p in ROOT.rglob("*"):
            if not p.is_file() or ".work" in p.parts or "__pycache__" in p.parts or p.suffix not in {".yaml", ".yml", ".py", ".md", ".json"}:
                continue
            if p.name == "test_k8s.py":  # this file spells the patterns it looks for
                continue
            t = p.read_text(encoding="utf-8")
            self.assertNotIn("BEGIN PRIVATE KEY", t, str(p))
            self.assertNotRegex(t, r"(?i)password\s*[:=]\s*['\"][^'\"]{4,}['\"]", str(p))


class TestPortsLineUp(unittest.TestCase):
    def setUp(self):
        self.kind = yaml.safe_load((ROOT / "kind-config.yaml").read_text(encoding="utf-8"))
        self.maps = self.kind["nodes"][0]["extraPortMappings"]

    def test_kind_ports_are_loopback_only(self):
        for m in self.maps:
            self.assertEqual(m["listenAddress"], "127.0.0.1")

    def test_nodeports_match_kind_port_mappings(self):
        node_ports = {s["spec"]["ports"][0]["nodePort"] for s in of_kind("Service") if s["spec"].get("type") == "NodePort"}
        self.assertEqual(node_ports, {m["containerPort"] for m in self.maps})

    def test_host_ports_match_what_day13_web_and_kong_expect(self):
        host = {m["containerPort"]: m["hostPort"] for m in self.maps}
        self.assertEqual(host[30080], 8000)  # VITE_API_BASE=http://localhost:8000/api is baked into the web image
        self.assertEqual(host[30300], 3000)  # kong CORS origin is http://localhost:3000
        tpl = (APP13 / "gateway" / "kong.yml.tpl").read_text(encoding="utf-8")
        self.assertIn("http://localhost:3000", tpl)

    def test_services_point_at_real_container_ports(self):
        deps = {d["metadata"]["name"]: d for d in of_kind("Deployment")}
        for s in of_kind("Service"):
            dep = deps[s["spec"]["selector"]["app"]]
            self.assertEqual(dep["spec"]["template"]["metadata"]["labels"], s["spec"]["selector"])
            cports = {p["containerPort"] for p in containers(dep)[0]["ports"]}
            for p in s["spec"]["ports"]:
                self.assertIn(p.get("targetPort", p["port"]), cports, s["metadata"]["name"])

    def test_kong_listens_where_the_service_sends_traffic(self):
        gw = next(d for d in of_kind("Deployment") if d["metadata"]["name"] == "gateway")
        env = {e["name"]: e["value"] for e in containers(gw)[0]["env"]}
        self.assertEqual(env["KONG_PROXY_LISTEN"], "0.0.0.0:8000")
        self.assertEqual(env["KONG_ADMIN_LISTEN"], "off")  # DB-less admin API can rewrite config: never exposed


class TestMatchesDay13(unittest.TestCase):
    """Day 6 and Day 9 manifests drifted from the real app; this keeps Day 14 from doing the same."""

    def test_api_gets_every_env_var_the_day13_compose_gives_it(self):
        compose = yaml.safe_load((APP13 / "docker-compose.yml").read_text(encoding="utf-8"))
        want = set(compose["services"]["api"]["environment"])
        intentionally_missing = {"OTLP_ENDPOINT"}  # no Jaeger in this cluster; empty = no trace export
        self.assertEqual(want - env_keys("api"), intentionally_missing)

    def test_worker_gets_every_env_var_the_day13_compose_gives_it(self):
        compose = yaml.safe_load((APP13 / "docker-compose.yml").read_text(encoding="utf-8"))
        want = set(compose["services"]["worker"]["environment"])
        self.assertEqual(want - env_keys("worker"), {"OTLP_ENDPOINT"})

    def test_worker_gets_no_jwt_material(self):
        worker = next(d for d in of_kind("Deployment") if d["metadata"]["name"] == "worker")
        self.assertFalse([k for k in env_keys("worker") if k.startswith("JWT")])
        self.assertNotIn("volumes", worker["spec"]["template"]["spec"])

    def test_refresh_cookie_path_matches_the_kong_auth_route(self):
        tpl = (APP13 / "gateway" / "kong.yml.tpl").read_text(encoding="utf-8")
        self.assertIn('paths: ["/api/auth"]', tpl)
        cm = of_kind("ConfigMap")[0]["data"]
        self.assertEqual(cm["REFRESH_COOKIE_PATH"], "/api/auth")

    def test_valkey_keeps_the_revocation_list_across_pod_restarts(self):
        v = next(d for d in of_kind("Deployment") if d["metadata"]["name"] == "valkey")
        self.assertIn("--appendonly", containers(v)[0]["args"])
        self.assertTrue(any("persistentVolumeClaim" in x for x in v["spec"]["template"]["spec"]["volumes"]))

    def test_probe_paths_exist_in_the_go_code(self):
        handler = (APP13 / "internal" / "adapter" / "http" / "handler.go").read_text(encoding="utf-8")
        obs = (APP13 / "internal" / "obs" / "setup.go").read_text(encoding="utf-8")
        for d in of_kind("Deployment"):
            for c in containers(d):
                for probe in ("readinessProbe", "livenessProbe"):
                    http = c.get(probe, {}).get("httpGet")
                    if not http:
                        continue
                    src = {"api": handler, "worker": obs}.get(d["metadata"]["name"])
                    if src is not None:
                        self.assertIn(f'"GET {http["path"]}"', src, f'{d["metadata"]["name"]} {http["path"]}')

    def test_kafka_controller_quorum_points_at_localhost_not_its_own_service(self):
        # CI run #1: with 1@kafka:9093 the pod could not reach itself through the Service before it was Ready
        k = next(d for d in of_kind("Deployment") if d["metadata"]["name"] == "kafka")
        env = {e["name"]: e.get("value") for e in containers(k)[0]["env"]}
        self.assertEqual(env["KAFKA_CONTROLLER_QUORUM_VOTERS"], "1@localhost:9093")

    def test_web_probe_path_exists_in_nginx_conf(self):
        conf = (APP13 / "web" / "nginx.conf").read_text(encoding="utf-8")
        self.assertIn("/healthz", conf)


FAKE_PUB = "-----BEGIN PUBLIC KEY-----\nMFwwDQYJKoZIhvcNAQEBBQADSwAwSAJBAFAKEFAKEFAKEFAKEFAKEFAKEFAKEFAKEFAKE\n-----END PUBLIC KEY-----\n"


class TestRenderKong(unittest.TestCase):
    TPL = (APP13 / "gateway" / "kong.yml.tpl").read_text(encoding="utf-8")

    def test_one_entry_per_kid_and_result_is_valid_yaml(self):
        out = render_kong.render(self.TPL, {"k1": FAKE_PUB, "k2": FAKE_PUB})
        secrets_ = yaml.safe_load(out)["consumers"][0]["jwt_secrets"]
        self.assertEqual([s["key"] for s in secrets_], ["k1", "k2"])
        self.assertTrue(all(s["algorithm"] == "RS256" for s in secrets_))
        self.assertNotIn("example-kid", out)

    def test_services_are_untouched_by_rendering(self):
        out = yaml.safe_load(render_kong.render(self.TPL, {"k1": FAKE_PUB}))
        self.assertEqual([s["name"] for s in out["services"]], ["auth-api", "claims-api"])

    def test_bad_kid_and_non_public_key_are_refused(self):
        with self.assertRaises(ValueError):
            render_kong.render(self.TPL, {"../k1": FAKE_PUB})
        with self.assertRaises(ValueError):
            render_kong.render(self.TPL, {"k1": "-----BEGIN PRIVATE KEY-----\nx\n-----END PRIVATE KEY-----"})
        with self.assertRaises(ValueError):
            render_kong.render(self.TPL, {})

    def test_template_without_markers_is_refused(self):
        with self.assertRaises(ValueError):
            render_kong.render("consumers: []\n", {"k1": FAKE_PUB})


class TestWorkflow(unittest.TestCase):
    def wf(self):
        p = ROOT / "ci" / "day14.yml"
        self.assertTrue(p.exists(), "ci/day14.yml missing")
        return p.read_text(encoding="utf-8"), yaml.safe_load(p.read_text(encoding="utf-8"))

    def test_has_a_kind_job_that_runs_the_smoke_test_and_dumps_state_on_failure(self):
        text, y = self.wf()
        self.assertIn("kind", y["jobs"])
        self.assertIn("kindctl.py all", text)
        self.assertIn("if: failure()", text)
        self.assertIn("kubectl", text)

    def test_actions_are_pinned_to_a_version_and_no_secrets_are_used(self):
        text, _ = self.wf()
        uses = re.findall(r"^\s*-?\s*uses:\s*([^\s]+)", text, re.M)
        self.assertTrue(uses)
        for u in uses:
            self.assertRegex(u, r"@v\d+", u)
            self.assertNotIn("trivy-action", u)  # Day 13: call Trivy from a pinned image instead
        self.assertNotIn("secrets.", text)

    def test_permissions_are_read_only(self):
        _, y = self.wf()
        self.assertEqual(y["permissions"], {"contents": "read"})


class TestBaseline(unittest.TestCase):
    PATH = ROOT / "tests" / "baseline" / "k8s-summary.json"

    def test_manifests_still_do_what_the_baseline_says(self):
        now = json.loads(json.dumps(summary()))  # normalise to JSON types
        if os.environ.get("UPDATE_BASELINE") == "1":
            self.PATH.write_text(json.dumps(now, indent=2, sort_keys=True, ensure_ascii=False) + "\n", encoding="utf-8")
        old = json.loads(self.PATH.read_text(encoding="utf-8"))
        self.assertEqual(old, now, "manifest changed: if intended, UPDATE_BASELINE=1 and review the diff")


if __name__ == "__main__":
    unittest.main()
