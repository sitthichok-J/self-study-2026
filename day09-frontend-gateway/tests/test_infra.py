"""ตรวจไฟล์ของ Day 9 แบบ static (ไม่ต้องมี Docker/Node)
รัน: python -m unittest discover -s tests   (ถ้ามี PyYAML จะตรวจโครงสร้าง YAML ละเอียดขึ้น)
"""
import re
import unittest
from pathlib import Path

try:
    import yaml
except ImportError:  # pragma: no cover
    yaml = None

ROOT = Path(__file__).resolve().parent.parent
D8 = ROOT.parent / "day08-backend"
WEB_ORIGIN = "http://localhost:3000"
GATEWAY_BASE = "http://localhost:8000/api"


def read(p: Path) -> str:
    return p.read_text(encoding="utf-8-sig")


def load(p: Path):
    return yaml.safe_load(read(p))


def load_all(p: Path):
    return [d for d in yaml.safe_load_all(read(p)) if d]


needs_yaml = unittest.skipIf(yaml is None, "ต้องติดตั้ง pyyaml")
needs_d8 = unittest.skipUnless(D8.exists(), "ไม่พบ ../day08-backend")


@needs_yaml
class TestKong(unittest.TestCase):
    def setUp(self):
        self.cfg = load(ROOT / "gateway" / "kong.yml")
        self.svc = self.cfg["services"][0]
        self.route = self.svc["routes"][0]
        self.plugins = {p["name"]: p["config"] for p in self.svc["plugins"]}

    def test_routes_api_prefix_to_backend(self):
        self.assertEqual(self.svc["url"], "http://api:8080")
        self.assertEqual(self.route["paths"], ["/api"])
        self.assertTrue(self.route["strip_path"], "backend ไม่มี prefix /api")

    def test_only_methods_the_ui_needs(self):
        self.assertEqual(sorted(self.route["methods"]), ["GET", "OPTIONS", "POST"])

    def test_no_automatic_retry_of_post(self):
        self.assertEqual(self.svc["retries"], 0)

    def test_cors_matches_web_origin_and_idempotency_header(self):
        c = self.plugins["cors"]
        self.assertEqual(c["origins"], [WEB_ORIGIN])
        self.assertNotIn("*", c["origins"])
        self.assertIn("Idempotency-Key", c["headers"])
        # ถ้าไม่ expose browser จะอ่าน header นี้จาก JS ไม่ได้
        self.assertIn("Idempotent-Replayed", c["exposed_headers"])
        self.assertFalse(c["credentials"])

    def test_rate_limit(self):
        r = self.plugins["rate-limiting"]
        self.assertEqual(r["minute"], 120)
        self.assertEqual(r["limit_by"], "ip")
        self.assertEqual(r["error_code"], 429)
        # หน้า detail poll ทุก 2 วินาที = 30 ครั้ง/นาที ต้องเหลือที่ว่างให้ผู้ใช้ทำอย่างอื่นด้วย
        self.assertGreaterEqual(r["minute"], 30 * 2)

    def test_body_limit_matches_backend(self):
        r = self.plugins["request-size-limiting"]
        self.assertEqual((r["allowed_payload_size"], r["size_unit"]), (1, "megabytes"))
        if D8.exists():
            self.assertIn("1<<20", read(D8 / "internal/adapter/http/handler.go"))

    def test_declarative_db_less(self):
        self.assertEqual(self.cfg["_format_version"], "3.0")


@needs_yaml
class TestCompose(unittest.TestCase):
    def setUp(self):
        self.c = load(ROOT / "docker-compose.yml")
        self.s = self.c["services"]

    def test_includes_day08(self):
        inc = self.c["include"][0]
        self.assertEqual(inc["path"], "../day08-backend/docker-compose.yml")
        self.assertEqual(inc["project_directory"], "../day08-backend")

    def test_gateway_pinned_and_admin_api_off(self):
        g = self.s["gateway"]
        self.assertEqual(g["image"], "kong:3.9.1")
        self.assertEqual(g["environment"]["KONG_ADMIN_LISTEN"], "off")
        self.assertEqual(g["environment"]["KONG_DATABASE"], "off")
        self.assertEqual(g["ports"], ["8000:8000"])
        self.assertIn("./gateway/kong.yml:/kong/kong.yml:ro", g["volumes"])

    def test_web_points_at_gateway_and_waits_for_it(self):
        w = self.s["web"]
        self.assertEqual(w["build"]["args"]["VITE_API_BASE"], GATEWAY_BASE)
        self.assertEqual(w["ports"], ["3000:8080"])
        self.assertEqual(w["depends_on"]["gateway"]["condition"], "service_healthy")

    def test_upstream_name_exists_in_day08_compose(self):
        if not D8.exists():
            self.skipTest("ไม่พบ ../day08-backend")
        d8 = load(D8 / "docker-compose.yml")
        self.assertIn("api", d8["services"])
        self.assertIn("8080:8080", d8["services"]["api"]["ports"])

    def test_web_origin_is_the_published_port(self):
        self.assertEqual(WEB_ORIGIN, "http://localhost:" + self.s["web"]["ports"][0].split(":")[0])


class TestWebImage(unittest.TestCase):
    def test_dockerfile(self):
        d = read(ROOT / "web" / "Dockerfile")
        self.assertEqual(d.count("FROM "), 2, "ต้อง multi-stage")
        self.assertIn("npm ci", d)
        self.assertIn("nginx-unprivileged", d)
        self.assertIn(f"ARG VITE_API_BASE={GATEWAY_BASE}", d)
        self.assertNotRegex(d, r"FROM \S+:latest")

    def test_nginx_spa_and_cache_rules(self):
        n = read(ROOT / "web" / "nginx.conf")
        self.assertIn("listen 8080;", n)
        self.assertIn("try_files $uri /index.html;", n)
        self.assertIn("immutable", n)
        self.assertIn("location = /healthz", n)
        self.assertIn("nosniff", n)

    def test_security_headers_present_in_every_location_that_sets_headers(self):
        # nginx: location ที่มี add_header ของตัวเองจะไม่สืบทอด add_header ระดับ server
        n = read(ROOT / "web" / "nginx.conf")
        blocks = re.findall(r"location [^{]+\{(.*?)\n  \}", n, re.S)
        self.assertGreaterEqual(len(blocks), 3)
        for b in blocks:
            if "add_header" in b:
                for h in ("X-Content-Type-Options nosniff", "X-Frame-Options DENY", "Referrer-Policy no-referrer"):
                    self.assertIn(h, b)

    def test_dockerignore_keeps_node_modules_out(self):
        self.assertIn("node_modules", read(ROOT / "web" / ".dockerignore"))
        self.assertIn("node_modules", read(ROOT / "web" / ".gitignore"))


@needs_yaml
class TestK8s(unittest.TestCase):
    def setUp(self):
        self.docs = []
        for f in sorted((ROOT / "k8s").glob("*.yaml")):
            if f.name == "kustomization.yaml":
                continue
            self.docs += [(f.name, d) for d in load_all(f)]
        self.dep = [d for _, d in self.docs if d["kind"] == "Deployment"]
        self.svc = [d for _, d in self.docs if d["kind"] == "Service"]
        self.jobs = [d for _, d in self.docs if d["kind"] == "Job"]

    def pod(self, d):
        return d["spec"]["template"]["spec"]

    def test_kustomization_lists_every_file(self):
        k = load(ROOT / "k8s" / "kustomization.yaml")
        files = {f.name for f in (ROOT / "k8s").glob("*.yaml")} - {"kustomization.yaml"}
        self.assertEqual(set(k["resources"]), files)

    def test_service_links_off_everywhere(self):
        # Service ชื่อ kafka สร้าง env KAFKA_PORT ซึ่งทำให้ apache/kafka เริ่มไม่ขึ้น
        for d in self.dep + self.jobs:
            spec = self.pod(d) if d["kind"] == "Deployment" else d["spec"]["template"]["spec"]
            self.assertIs(spec.get("enableServiceLinks"), False, d["metadata"]["name"])

    def test_no_secret_or_password_literals_in_repo(self):
        self.assertFalse([d for _, d in self.docs if d["kind"] == "Secret"])
        for f in (ROOT / "k8s").glob("*.yaml"):
            self.assertNotRegex(read(f), r"(?i)password:\s*\S|postgres://\w+:\w+@", f.name)

    def test_images_pinned_or_local(self):
        for d in self.dep + self.jobs:
            spec = self.pod(d) if d["kind"] == "Deployment" else d["spec"]["template"]["spec"]
            for c in spec["containers"]:
                img = c["image"]
                self.assertNotIn(":latest", img)
                self.assertIn(":", img, f"{img} ไม่ได้ระบุ tag")

    def test_app_pods_hardened(self):
        for name in ("api", "worker", "web"):
            d = next(x for x in self.dep if x["metadata"]["name"] == name)
            sc = self.pod(d)["securityContext"]
            self.assertTrue(sc["runAsNonRoot"], name)
            c = self.pod(d)["containers"][0]["securityContext"]
            self.assertIs(c["allowPrivilegeEscalation"], False, name)
            self.assertEqual(c["capabilities"]["drop"], ["ALL"], name)

    def test_every_deployment_has_memory_limit(self):
        for d in self.dep:
            for c in self.pod(d)["containers"]:
                self.assertIn("memory", c["resources"]["limits"], d["metadata"]["name"])

    def test_service_selectors_match_deployments(self):
        labels = {d["metadata"]["name"]: d["spec"]["selector"]["matchLabels"] for d in self.dep}
        for s in self.svc:
            self.assertIn(s["metadata"]["name"], labels)
            self.assertEqual(s["spec"]["selector"], labels[s["metadata"]["name"]])

    def test_api_has_probes_and_graceful_period(self):
        api = next(x for x in self.dep if x["metadata"]["name"] == "api")
        c = self.pod(api)["containers"][0]
        self.assertEqual(c["readinessProbe"]["httpGet"]["path"], "/readyz")
        self.assertEqual(c["livenessProbe"]["httpGet"]["path"], "/healthz")
        self.assertGreaterEqual(self.pod(api)["terminationGracePeriodSeconds"], 30)

    def test_names_that_deploy_script_creates_are_the_ones_manifests_use(self):
        script = read(ROOT / "scripts" / "deploy-k8s.sh")
        text = "\n".join(read(f) for f in (ROOT / "k8s").glob("*.yaml"))
        for name in set(re.findall(r"configMap: \{ name: ([\w-]+) \}", text)):
            self.assertIn(f"create configmap {name}", script)
        keys = set(re.findall(r"secretKeyRef: \{ name: db, key: (\w+) \}", text))
        self.assertEqual(keys, {"password", "url"})
        for k in keys:
            self.assertIn(f'--from-literal={k}=', script)

    def test_gateway_config_has_single_source(self):
        script = read(ROOT / "scripts" / "deploy-k8s.sh")
        self.assertIn("kong.yml=gateway/kong.yml", script)
        self.assertIn("../day08-backend/migrations", script)
        self.assertFalse((ROOT / "k8s" / "kong.yml").exists())

    @needs_d8
    def test_env_names_match_day08_compose(self):
        d8 = load(D8 / "docker-compose.yml")["services"]
        for name in ("api", "worker"):
            want = set(d8[name]["environment"])
            d = next(x for x in self.dep if x["metadata"]["name"] == name)
            got = {e["name"] for e in self.pod(d)["containers"][0]["env"]}
            self.assertEqual(got, want, name)

    @needs_d8
    def test_topics_match_backend_constants(self):
        init = next(j for j in self.jobs if j["metadata"]["name"] == "kafka-init")
        script = init["spec"]["template"]["spec"]["containers"][0]["args"][0]
        port = read(D8 / "internal/port/port.go")
        for topic in re.findall(r'"(claims\.submitted[\w.]*)"', port):
            self.assertIn(f"--topic {topic} ", script)

    @needs_d8
    def test_migrate_image_matches_day08(self):
        want = load(D8 / "docker-compose.yml")["services"]["migrate"]["image"]
        job = next(j for j in self.jobs if j["metadata"]["name"] == "migrate")
        self.assertEqual(job["spec"]["template"]["spec"]["containers"][0]["image"], want)


@needs_yaml
class TestCI(unittest.TestCase):
    def setUp(self):
        self.text = read(ROOT / "ci" / "github-actions.yml")
        self.wf = load(ROOT / "ci" / "github-actions.yml")

    def test_jobs(self):
        self.assertEqual(list(self.wf["jobs"]), ["backend", "web", "static", "e2e-kind", "publish"])

    def test_steps_run_the_checks_the_plan_asks_for(self):
        for cmd in ("go vet ./...", "go test ./... -race", "npm run typecheck", "npm test", "npm run build"):
            self.assertIn(cmd, self.text)

    def test_actions_pinned_to_major_version(self):
        for u in re.findall(r"uses: (\S+)", self.text):
            self.assertRegex(u, r"@v\d+$", u)

    def test_publish_only_on_main_push_and_least_privilege(self):
        pub = self.wf["jobs"]["publish"]
        self.assertIn("refs/heads/main", pub["if"])
        self.assertEqual(self.wf["permissions"], {"contents": "read"})
        self.assertEqual(pub["permissions"]["packages"], "write")
        for name, job in self.wf["jobs"].items():
            if name != "publish":
                self.assertNotIn("permissions", job, name)

    def test_only_expected_secrets(self):
        self.assertEqual(set(re.findall(r"secrets\.(\w+)", self.text)), {"GITHUB_TOKEN", "K8S_DB_PASSWORD"})

    def test_ghcr_owner_lowercased(self):
        self.assertIn("${GITHUB_REPOSITORY_OWNER,,}", self.text)

    def test_deploy_waits_for_all_checks(self):
        self.assertEqual(sorted(self.wf["jobs"]["e2e-kind"]["needs"]), ["backend", "static", "web"])
        self.assertEqual(self.wf["jobs"]["publish"]["needs"], ["e2e-kind"])


class TestScripts(unittest.TestCase):
    def test_shell_scripts_strict(self):
        for n in ("deploy-k8s.sh", "smoke-k8s.sh"):
            s = read(ROOT / "scripts" / n)
            self.assertIn("set -euo pipefail", s, n)
            self.assertTrue(s.startswith("#!/usr/bin/env bash"), n)

    def test_deploy_never_logs_or_hardcodes_password(self):
        s = read(ROOT / "scripts" / "deploy-k8s.sh")
        # แค่พูดถึงชื่อตัวแปรในข้อความ error ได้ แต่ต้องไม่พิมพ์ "ค่า"
        self.assertNotRegex(s, r"echo[^\n]*\$\{?(PASS|DB_PASSWORD)\b")
        self.assertIn("/dev/urandom", s)

    def test_deploy_reuses_existing_password(self):
        # postgres ใช้ emptyDir ที่ยังถือรหัสเดิมอยู่ ถ้าสุ่มใหม่ทุกครั้งที่ deploy ซ้ำ API จะต่อ DB ไม่ได้
        s = read(ROOT / "scripts" / "deploy-k8s.sh")
        self.assertLess(s.index("get secret db"), s.index("DB_PASSWORD:-"))

    def test_powershell_has_bom_because_it_contains_thai(self):
        raw = (ROOT / "scripts" / "run-all.ps1").read_bytes()
        self.assertTrue(raw.startswith(b"\xef\xbb\xbf"), "PowerShell 5.1 อ่านไฟล์ไม่มี BOM เป็น ANSI แล้วภาษาไทยเพี้ยน")

    def test_run_all_covers_the_gateway_features(self):
        s = read(ROOT / "scripts" / "run-all.ps1")
        for must in ("413", "429", "Retry-After", "Access-Control-Allow-Origin", "Idempotent-Replayed",
                     "docker compose stop api", "down -v", "-Keep", "ErrorDetails"):
            self.assertIn(must, s)

    def test_evil_origin_check_does_not_expect_empty_header(self):
        # Kong ที่ตั้ง origin เดียวจะตอบ origin ที่ตั้งไว้เสมอ (พบตอนรันจริง) ห้ามเขียน test ให้คาดว่า header ว่าง
        s = read(ROOT / "scripts" / "run-all.ps1")
        self.assertIn("http://evil.example", s)
        self.assertNotIn("(Hd $pe 'Access-Control-Allow-Origin') -eq ''", s)

    def test_rate_limit_step_is_last_check(self):
        s = read(ROOT / "scripts" / "run-all.ps1")
        self.assertGreater(s.index("rate limit 120"), s.index("docker compose stop api"))
        self.assertGreater(s.index("rate limit 120"), s.index("web: GET / ส่ง index.html"))


@needs_d8
class TestFrontendMatchesBackend(unittest.TestCase):
    """กฎฝั่ง client ต้องตรงกับ Day 8 ถ้า backend เปลี่ยนแล้วลืมแก้ UI เทสต์นี้ต้องล้ม"""

    def setUp(self):
        self.go = read(D8 / "internal/domain/claim.go")
        self.lib = read(ROOT / "web/src/lib.ts")
        self.types = read(ROOT / "web/src/types/claim.ts")

    def test_patient_ref_pattern(self):
        self.assertIn("FAKE-[A-Z0-9]{3,12}", self.go)
        self.assertIn("FAKE-[A-Z0-9]{3,12}", self.lib)

    def test_max_amount(self):
        self.assertIn("MaxAmount = 1_000_000_000", self.go)
        self.assertIn("MAX_SATANG = 1_000_000_000", self.lib)

    def test_statuses(self):
        go_status = set(re.findall(r'Status\w+\s+Status = "(\w+)"', self.go))
        ts_status = set(re.findall(r"'(PENDING|APPROVED|REJECTED|FAILED)'", self.types.split("STATUS_LABEL")[0]))
        self.assertEqual(go_status, ts_status)

    def test_decision_reasons_have_labels(self):
        # พบตอนรันจริง: reason "ok" ไม่มีป้ายภาษาไทย หน้าจอเลยโชว์คำดิบ
        reasons = set(re.findall(r'Decision\{Status\w+,\s*"(\w+)"\}', self.go))
        self.assertTrue({"ok", "over_limit"} <= reasons, reasons)
        label_block = self.types.split("REASON_LABEL")[1].split("}")[0]
        for r in reasons:
            self.assertRegex(label_block, rf"\b{r}:", r)

    def test_claim_json_fields_exist_in_ui_type(self):
        for f in re.findall(r'json:"(\w+)(?:,omitempty)?"', self.go):
            if f != "-":
                self.assertIn(f, self.types, f)

    def test_ui_sends_idempotency_key_header(self):
        self.assertIn("'Idempotency-Key'", read(ROOT / "web/src/api/claims.ts"))

    def test_ui_reads_replay_header_backend_sets(self):
        self.assertIn("Idempotent-Replayed", read(D8 / "internal/adapter/http/handler.go"))
        self.assertIn("Idempotent-Replayed", read(ROOT / "web/src/api/claims.ts"))


class TestFakeDataOnly(unittest.TestCase):
    def test_ui_has_no_real_looking_identifiers(self):
        for f in (ROOT / "web" / "src").rglob("*"):
            if f.suffix in {".ts", ".vue"}:
                t = read(f)
                self.assertNotRegex(t, r"\b\d{13}\b", f"{f.name}: เลข 13 หลักคล้ายเลขบัตรประชาชน")
                for ref in re.findall(r"patient_ref[^\n]*['\"]([A-Z]+-[A-Z0-9]+)['\"]", t):
                    self.assertTrue(ref.startswith("FAKE-"), f"{f.name}: {ref}")

    def test_form_hint_says_fake(self):
        self.assertIn("FAKE-", read(ROOT / "web/src/views/ClaimFormView.vue"))


if __name__ == "__main__":
    unittest.main()
