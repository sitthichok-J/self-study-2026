"""Static checks for the Day 12 deploy files (no Docker needed).

Run:  python -m unittest discover -s tests -v
They are the 'does the config still say what we think it says' net: ports,
hardening flags, and that every metric / label used by alerts and the
dashboard really exists in the Go code.
"""
import json
import pathlib
import re
import unittest

import yaml

ROOT = pathlib.Path(__file__).resolve().parent.parent


def read(p):
    return (ROOT / p).read_text(encoding="utf-8")


def load_yaml(p):
    return yaml.safe_load(read(p))


def declared_metrics():
    """name -> (type, [labels]) parsed from the Go sources."""
    out = {}
    src = read("internal/obs/app.go") + read("internal/obs/metrics.go") + read("cmd/api/main.go")
    for m in re.finditer(r'New(Counter|Histogram)Vec\(\s*"([a-z_:0-9]+)"\s*,\s*"[^"]*"\s*,?\s*(?:nil\s*,?\s*)?([^)]*)\)', src):
        kind, name, rest = m.groups()
        out[name] = (kind.lower(), re.findall(r'"([a-z_]+)"', rest))
    for m in re.finditer(r'NewGaugeFunc\(\s*"([a-z_:0-9]+)"', src):
        out[m.group(1)] = ("gauge", [])
    return out


METRICS = declared_metrics()


def metric_uses(expr):
    """(metric, [labels used in its selector or by-clauses]) found in a PromQL string."""
    found = []
    for m in re.finditer(r'\b([a-z_][a-z0-9_]*)\s*(\{([^}]*)\})?', expr):
        name = m.group(1)
        base = re.sub(r'_(bucket|sum|count)$', '', name)
        if name in METRICS:
            real = name
        elif base in METRICS and METRICS[base][0] == "histogram":
            real = base
        else:
            continue
        labels = re.findall(r'([a-z_]+)\s*(?:=~|!=|=|!~)', m.group(3) or "")
        found.append((real, labels))
    return found


class TestComposeHardening(unittest.TestCase):
    def setUp(self):
        self.c = load_yaml("docker-compose.yml")
        self.svc = self.c["services"]

    def test_every_published_port_is_loopback_only(self):
        for name, s in self.svc.items():
            for p in s.get("ports", []):
                self.assertTrue(str(p).startswith("127.0.0.1:"), f"{name} publishes {p} on all interfaces")

    def test_metrics_port_is_not_published(self):
        for name in ("api", "worker"):
            for p in self.svc[name].get("ports", []):
                self.assertNotIn("9090", str(p).split(":")[-1], f"{name} publishes its metrics port")

    def test_otlp_port_is_not_published(self):
        for p in self.svc["jaeger"].get("ports", []):
            self.assertNotIn("4318", str(p))
            self.assertNotIn("4317", str(p))

    def test_app_containers_are_locked_down(self):
        for name in ("api", "worker"):
            s = self.svc[name]
            self.assertTrue(s.get("read_only"), name)
            self.assertIn("ALL", s.get("cap_drop", []), name)
            self.assertIn("no-new-privileges:true", s.get("security_opt", []), name)

    def test_nothing_is_privileged_or_host_networked(self):
        for name, s in self.svc.items():
            self.assertFalse(s.get("privileged"), name)
            self.assertNotEqual(s.get("network_mode"), "host", name)
            self.assertNotIn("/var/run/docker.sock", " ".join(map(str, s.get("volumes", []))), name)

    def test_secrets_are_required_from_the_environment(self):
        self.assertRegex(self.svc["grafana"]["environment"]["GF_SECURITY_ADMIN_PASSWORD"], r"^\$\{GRAFANA_ADMIN_PASSWORD:\?")

    def test_api_uses_rsa_keys_from_a_read_only_mount_and_no_shared_secret(self):
        api = self.svc["api"]
        self.assertNotIn("JWT_SECRET", api["environment"])
        self.assertEqual(api["environment"]["JWT_KEYS_DIR"], "/keys")
        self.assertRegex(api["environment"]["JWT_SIGNING_KID"], r"^\$\{JWT_SIGNING_KID:-k1\}$")
        self.assertIn("./keys:/keys:ro", api["volumes"])

    def test_worker_gets_no_jwt_material(self):
        env = self.svc["worker"]["environment"]
        for k in env:
            self.assertFalse(k.startswith("JWT_"), k)
        self.assertNotIn("/keys", " ".join(map(str, self.svc["worker"].get("volumes", []))))

    def test_grafana_has_no_anonymous_access(self):
        env = self.svc["grafana"]["environment"]
        self.assertEqual(str(env["GF_AUTH_ANONYMOUS_ENABLED"]).lower(), "false")
        self.assertEqual(str(env["GF_USERS_ALLOW_SIGN_UP"]).lower(), "false")

    def test_jaeger_accepts_otlp_and_apps_point_at_it(self):
        self.assertEqual(str(self.svc["jaeger"]["environment"]["COLLECTOR_OTLP_ENABLED"]).lower(), "true")
        for name in ("api", "worker"):
            self.assertEqual(self.svc[name]["environment"]["OTLP_ENDPOINT"], "http://jaeger:4318")

    def test_images_are_pinned(self):
        for name, s in self.svc.items():
            img = s.get("image")
            if img:
                self.assertIn(":", img, f"{name} has no tag")
                self.assertFalse(img.endswith(":latest"), f"{name} uses :latest")

    def test_dockerfile_runs_nonroot(self):
        d = read("Dockerfile")
        self.assertIn("distroless/static-debian12:nonroot", d)
        self.assertRegex(d, r"USER 65532:65532")
        self.assertIn("CGO_ENABLED=0", d)


class TestPrometheus(unittest.TestCase):
    def test_scrape_targets_are_the_internal_metrics_ports(self):
        p = load_yaml("deploy/prometheus/prometheus.yml")
        t = {j["job_name"]: j["static_configs"][0]["targets"] for j in p["scrape_configs"]}
        self.assertEqual(t, {"claims-api": ["api:9090"], "claims-worker": ["worker:9090"]})

    def test_rules_file_is_mounted_where_prometheus_expects_it(self):
        p = load_yaml("deploy/prometheus/prometheus.yml")
        self.assertEqual(p["rule_files"], ["/etc/prometheus/alerts.yml"])
        vols = " ".join(load_yaml("docker-compose.yml")["services"]["prometheus"]["volumes"])
        self.assertIn(":/etc/prometheus/alerts.yml:ro", vols)

    def test_alert_rules_are_well_formed(self):
        rules = load_yaml("deploy/prometheus/alerts.yml")["groups"][0]["rules"]
        names = [r["alert"] for r in rules]
        self.assertEqual(len(names), len(set(names)))
        for r in rules:
            self.assertIn("severity", r["labels"], r["alert"])
            self.assertIn("summary", r["annotations"], r["alert"])
            e = r["expr"]
            for a, b in ("()", "[]", "{}"):
                self.assertEqual(e.count(a), e.count(b), f"{r['alert']} unbalanced {a}{b}")

    def test_the_required_alerts_exist(self):
        names = {r["alert"] for r in load_yaml("deploy/prometheus/alerts.yml")["groups"][0]["rules"]}
        self.assertTrue({"HighErrorRate", "ApiDown", "DeadLetterGrowing", "AuthFailureSpike", "DbPoolSaturated"} <= names)

    def test_alert_metrics_and_labels_exist_in_go(self):
        for r in load_yaml("deploy/prometheus/alerts.yml")["groups"][0]["rules"]:
            uses = metric_uses(r["expr"])
            if r["alert"] == "ApiDown":
                continue  # uses the built-in 'up'
            self.assertTrue(uses, f"{r['alert']} references no known metric")
            for name, labels in uses:
                for lb in labels:
                    if lb in ("job", "le"):
                        continue
                    self.assertIn(lb, METRICS[name][1], f"{r['alert']}: {name} has no label {lb}")


class TestDashboard(unittest.TestCase):
    def setUp(self):
        self.d = json.loads(read("deploy/grafana/dashboards/claims-overview.json"))
        self.panels = [p for p in self.d["panels"] if p["type"] != "row"]

    def test_dashboard_is_valid_and_stable(self):
        self.assertEqual(self.d["uid"], "claims-overview")
        ids = [p["id"] for p in self.d["panels"]]
        self.assertEqual(len(ids), len(set(ids)))
        self.assertFalse(self.d["editable"])

    def test_panels_do_not_overlap_and_fit_the_grid(self):
        boxes = [(p["gridPos"]["x"], p["gridPos"]["y"], p["gridPos"]["w"], p["gridPos"]["h"]) for p in self.d["panels"]]
        for x, y, w, h in boxes:
            self.assertLessEqual(x + w, 24)
        for i, a in enumerate(boxes):
            for b in boxes[i + 1:]:
                overlap = a[0] < b[0] + b[2] and b[0] < a[0] + a[2] and a[1] < b[1] + b[3] and b[1] < a[1] + a[3]
                self.assertFalse(overlap, f"{a} overlaps {b}")

    def test_red_and_use_sections_exist(self):
        titles = " ".join(p["title"] for p in self.d["panels"] if p["type"] == "row")
        self.assertIn("RED", titles)
        self.assertIn("USE", titles)
        self.assertIn("Security", titles)

    def test_every_panel_queries_a_known_metric_with_known_labels(self):
        for p in self.panels:
            self.assertTrue(p["targets"], p["title"])
            for t in p["targets"]:
                self.assertEqual(t["datasource"]["uid"], "prometheus")
                uses = metric_uses(t["expr"])
                self.assertTrue(uses, f"{p['title']}: no known metric in {t['expr']}")
                for name, labels in uses:
                    for lb in labels:
                        if lb in ("job", "le"):
                            continue
                        self.assertIn(lb, METRICS[name][1], f"{p['title']}: {name} has no label {lb}")
                for by in re.findall(r"by \(([^)]*)\)", t["expr"]):
                    for lb in [x.strip() for x in by.split(",")]:
                        if lb in ("le", "job"):
                            continue
                        self.assertTrue(any(lb in METRICS[n][1] for n, _ in uses), f"{p['title']}: by-label {lb} unknown")

    def test_datasource_uids_match_provisioning(self):
        ds = {d["uid"] for d in load_yaml("deploy/grafana/provisioning/datasources/datasources.yml")["datasources"]}
        self.assertEqual(ds, {"prometheus", "jaeger"})

    def test_dashboard_provider_path_matches_the_mount(self):
        path = load_yaml("deploy/grafana/provisioning/dashboards/dashboards.yml")["providers"][0]["options"]["path"]
        vols = " ".join(load_yaml("docker-compose.yml")["services"]["grafana"]["volumes"])
        self.assertIn(f":{path}:ro", vols)


class TestMetricContract(unittest.TestCase):
    def test_parser_finds_the_expected_families(self):
        for n in ("http_requests_total", "http_request_duration_seconds", "consumer_messages_total",
                  "auth_failures_total", "authz_denied_total", "db_pool_acquired_conns", "go_goroutines"):
            self.assertIn(n, METRICS)

    def test_counters_end_in_total_and_histograms_in_seconds(self):
        for n, (kind, _) in METRICS.items():
            if kind == "counter":
                self.assertTrue(n.endswith("_total"), n)
            if kind == "histogram":
                self.assertTrue(n.endswith("_seconds"), n)

    def test_no_high_cardinality_label_names(self):
        for n, (_, labels) in METRICS.items():
            for lb in labels:
                self.assertNotIn(lb, {"id", "claim_id", "user", "sub", "owner", "path", "url", "trace_id", "request_id"}, n)


class TestMigrations(unittest.TestCase):
    def test_every_up_has_a_down(self):
        ups = {p.name.replace(".up.sql", "") for p in (ROOT / "migrations").glob("*.up.sql")}
        downs = {p.name.replace(".down.sql", "") for p in (ROOT / "migrations").glob("*.down.sql")}
        self.assertEqual(ups, downs)

    def test_owner_migration_scopes_idempotency_per_owner(self):
        up = read("migrations/000002_owner.up.sql")
        self.assertIn("owner_id", up)
        self.assertIn("UNIQUE (owner_id, idempotency_key)", up)
        self.assertIn("DROP CONSTRAINT claims_idempotency_key_uq", up)
        down = read("migrations/000002_owner.down.sql")
        self.assertIn("UNIQUE (idempotency_key)", down)

    def test_repo_sql_matches_the_scoped_constraint(self):
        sql = read("internal/adapter/postgres/repo.go")
        self.assertIn("ON CONFLICT (owner_id, idempotency_key)", sql)
        self.assertIn("WHERE owner_id = $1 AND idempotency_key = $2", sql)
        self.assertIn("owner_id = $4::text", sql)


class TestNoSecretsInRepo(unittest.TestCase):
    """A small stand-in for gitleaks (the real tool is in scripts/security-scan.ps1)."""

    SKIP_DIRS = {".git", "node_modules", "certs", ".rendered", "keys"}  # สร้างตอนรันและ git-ignore (มี key/secret ของเครื่องนี้)

    def files(self):
        for p in ROOT.rglob("*"):
            if p.is_file() and not (set(p.parts) & self.SKIP_DIRS) and p.suffix not in {".png", ".mp4", ".gif"}:
                yield p

    def test_no_private_keys_cloud_keys_or_real_looking_tokens(self):
        bad = []
        pats = [r"-----BEGIN [A-Z ]*PRIVATE KEY-----", r"AKIA[0-9A-Z]{16}", r"eyJ[A-Za-z0-9_-]{10,}\.eyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}",
                r"ghp_[A-Za-z0-9]{30,}", r"xox[baprs]-[A-Za-z0-9-]{10,}"]
        for p in self.files():
            if p.name == "test_infra.py":
                continue
            try:
                t = p.read_text(encoding="utf-8")
            except UnicodeDecodeError:
                continue
            for pat in pats:
                if re.search(pat, t):
                    bad.append(f"{p.relative_to(ROOT)}: {pat}")
        self.assertEqual(bad, [])

    def test_no_hardcoded_secret_assignments_outside_tests(self):
        bad = []
        for p in self.files():
            if p.suffix not in {".go", ".yml", ".yaml", ".ps1", ".bat", ".sh"} or p.name.endswith("_test.go"):
                continue
            for n, line in enumerate(p.read_text(encoding="utf-8").splitlines(), 1):
                if re.search(r'(?i)(jwt_?secret|admin_?password|api_?key)\s*[:=]\s*["\']?[A-Za-z0-9/+_-]{12,}', line) and "${" not in line and "$env:" not in line and "os.Getenv" not in line and "Env" not in line:
                    bad.append(f"{p.relative_to(ROOT)}:{n}")
        self.assertEqual(bad, [])

    def test_env_file_is_git_ignored(self):
        self.assertIn(".env", read(".gitignore").split())


class TestDocsExist(unittest.TestCase):
    def test_security_md_lists_open_items(self):
        t = read("SECURITY.md")
        for h in ("## ปิดแล้ว", "## ยังไม่ปิด", "## Threat"):
            self.assertIn(h, t)

    def test_readme_names_the_stdlib_deviation(self):
        self.assertIn("stdlib", read("README.md"))


class TestGatewayAuth(unittest.TestCase):
    def setUp(self):
        self.tpl = yaml.safe_load(read("gateway/kong.yml.tpl"))
        self.svc = {s["name"]: s for s in self.tpl["services"]}
        self.compose = load_yaml("docker-compose.yml")["services"]

    def plugins(self, name):
        return {p["name"]: p["config"] for p in self.svc[name].get("plugins", [])}

    def test_claims_route_requires_jwt_and_auth_route_does_not(self):
        jwt = self.plugins("claims-api")["jwt"]
        self.assertEqual(jwt["key_claim_name"], "kid")  # Day 12: key picked by kid (rotation)
        self.assertIn("exp", jwt["claims_to_verify"])
        self.assertEqual(jwt["uri_param_names"], [])
        self.assertNotIn("jwt", self.plugins("auth-api"))

    def test_auth_route_is_rate_limited_tighter_than_api(self):
        a = self.plugins("auth-api")["rate-limiting"]["minute"]
        b = self.plugins("claims-api")["rate-limiting"]["minute"]
        self.assertLess(a, b)

    def test_auth_cors_allows_credentials_only_for_the_ui_origin(self):
        c = self.plugins("auth-api")["cors"]
        self.assertTrue(c["credentials"])
        self.assertEqual(c["origins"], ["http://localhost:3000"])
        self.assertIn("X-Requested-With", c["headers"])

    def test_template_has_rsa_key_markers_and_no_shared_secret(self):
        t = read("gateway/kong.yml.tpl")
        self.assertEqual(t.count("__JWT_KEYS_BEGIN__"), 1)
        self.assertEqual(t.count("__JWT_KEYS_END__"), 1)
        self.assertNotIn("HS256", t)
        self.assertNotRegex(t, r"(?m)^\s*secret:")
        for c in self.tpl["consumers"][0]["jwt_secrets"]:
            self.assertEqual(c["algorithm"], "RS256")
            self.assertIn("rsa_public_key", c)

    def test_kong_picks_the_key_by_kid_so_keys_can_rotate(self):
        self.assertEqual(self.plugins("claims-api")["jwt"]["key_claim_name"], "kid")

    def test_kong_access_log_drops_the_query_string(self):
        env = self.compose["gateway"]["environment"]
        self.assertEqual(env["KONG_PROXY_ACCESS_LOG"], "/dev/stdout noquery")
        fmt = env["KONG_NGINX_HTTP_LOG_FORMAT"]
        self.assertTrue(fmt.startswith("noquery "))
        self.assertIn("$$uri", fmt)            # path only
        self.assertNotIn("$$request ", fmt)    # $request would carry the query string
        self.assertNotIn("$$request_uri", fmt)
        self.assertNotIn("$$args", fmt)
        self.assertNotIn("$$query_string", fmt)

    def test_429_headers_are_readable_by_the_ui(self):
        exposed = self.plugins("auth-api")["cors"]["exposed_headers"]
        self.assertIn("Retry-After", exposed)

    def test_key_files_are_git_and_docker_ignored(self):
        self.assertIn("keys/", read(".gitignore").split())
        self.assertIn("keys", read(".dockerignore").split())
        self.assertIn("certs", read(".dockerignore").split())

    def test_render_script_reads_public_keys_only(self):
        r = read("scripts/render-gateway.ps1")
        self.assertIn("*.pub", r)
        self.assertNotRegex(r, r"\*\.key|\$kid\.key|\.BaseName\s*\+\s*'\.key")  # never touches <kid>.key
        self.assertIn("BEGIN PUBLIC KEY", r)

    def test_rendered_config_and_certs_are_git_ignored(self):
        ig = read(".gitignore").split()
        self.assertIn("certs/", ig)
        self.assertIn("gateway/.rendered/", ig)
        self.assertIn("keys/", ig)

    def test_gateway_listens_on_tls_and_ports_are_loopback(self):
        g = self.compose["gateway"]
        self.assertIn("8443 ssl", g["environment"]["KONG_PROXY_LISTEN"])
        self.assertTrue(all(str(p).startswith("127.0.0.1:") for p in g["ports"]))
        # Day 13: Admin API / Kong Manager are on for viewing, so they must be published on loopback only
        # (in DB-less mode POST /config on the Admin API replaces the whole gateway config)
        pubs = [str(p) for p in g["ports"]]
        self.assertIn("127.0.0.1:8001:8001", pubs)
        self.assertIn("127.0.0.1:8002:8002", pubs)
        self.assertEqual(g["environment"]["KONG_DATABASE"], "off")
        self.assertRegex(g["image"], r":\d+\.\d+\.\d+$")

    def test_cookie_path_matches_the_kong_route(self):
        env = self.compose["api"]["environment"]
        self.assertEqual(env["REFRESH_COOKIE_PATH"], self.svc["auth-api"]["routes"][0]["paths"][0])
        self.assertEqual(str(env["REFRESH_COOKIE_SECURE"]).lower(), "true")

    def test_web_image_runs_as_non_root(self):
        self.assertRegex(read("web/Dockerfile"), r"(?m)^USER 101$")

    def test_demo_password_comes_from_the_environment(self):
        self.assertRegex(self.compose["api"]["environment"]["DEMO_PASSWORD"], r"^\$\{DEMO_PASSWORD:\?")
        self.assertNotIn("DEMO_PASSWORD", self.compose["worker"]["environment"])

    def test_web_starts_after_gateway_and_api_has_short_access_ttl(self):
        self.assertIn("gateway", self.compose["web"]["depends_on"])
        self.assertEqual(self.compose["api"]["environment"]["ACCESS_TTL"], "10m")

    def test_ui_never_uses_browser_storage_for_tokens(self):
        for p in (ROOT / "web" / "src").rglob("*"):
            if p.is_file() and p.suffix in {".ts", ".vue"} and p.name != "ThemeToggle.vue":  # ธีมเก็บ localStorage ได้ ไม่ใช่ token
                t = p.read_text(encoding="utf-8")
                self.assertNotRegex(t, r"(localStorage|sessionStorage)\.(set|get)Item", str(p))


class TestDay13Tools(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.compose = load_yaml("docker-compose.yml")["services"]

    def test_kafka_ui_is_pinned_loopback_and_read_only(self):
        k = self.compose["kafka-ui"]
        self.assertRegex(k["image"], r":v\d+\.\d+\.\d+$")
        self.assertEqual(k["ports"], ["127.0.0.1:8085:8080"])
        self.assertEqual(str(k["environment"]["KAFKA_CLUSTERS_0_READONLY"]).lower(), "true")
        self.assertEqual(str(k["environment"]["DYNAMIC_CONFIG_ENABLED"]).lower(), "false")
        self.assertIn("kafka", k["depends_on"])

    def test_every_published_port_is_loopback(self):
        for name, svc in self.compose.items():
            for p in svc.get("ports", []):
                self.assertTrue(str(p).startswith("127.0.0.1:"), f"{name}: {p}")

    def test_api_gets_the_valkey_it_needs_for_revocation(self):
        api = self.compose["api"]
        self.assertEqual(api["environment"]["REDIS_ADDR"], "valkey:6379")
        self.assertIn("valkey", api["depends_on"])

    def test_auth_route_cors_allows_authorization_for_logout(self):
        tpl = load_yaml("gateway/kong.yml.tpl")
        auth = next(s for s in tpl["services"] if s["name"] == "auth-api")
        cors = next(p for p in auth["plugins"] if p["name"] == "cors")
        self.assertIn("Authorization", cors["config"]["headers"])
        self.assertIn("Retry-After", cors["config"]["exposed_headers"])
        # the claims route must let the UI send POST (approve / reject)
        api = next(s for s in tpl["services"] if s["name"] == "claims-api")
        self.assertIn("POST", api["routes"][0]["methods"])


class TestWorkflow(unittest.TestCase):
    """ci/day13.yml is the copy of .github/workflows/day13.yml (GitHub only reads the repo root)."""

    @classmethod
    def setUpClass(cls):
        cls.text = read("ci/day13.yml")
        cls.wf = yaml.safe_load(cls.text)
        cls.on = cls.wf.get("on", cls.wf.get(True))  # PyYAML reads the key `on` as True

    def test_triggers_are_limited_to_this_folder(self):
        for ev in ("push", "pull_request"):
            self.assertIn("day13-revocation-ui-cicd/**", self.on[ev]["paths"])
        self.assertEqual(self.on["push"]["branches"], ["main"])

    def test_least_privilege_and_no_dangerous_triggers(self):
        self.assertEqual(self.wf["permissions"], {"contents": "read"})
        self.assertNotIn("pull_request_target", self.on)
        self.assertNotIn("workflow_run", self.on)
        self.assertNotRegex(self.text, r"packages:\s*write|id-token|secrets\.")

    def test_every_job_has_a_timeout_and_expected_jobs_exist(self):
        self.assertEqual(set(self.wf["jobs"]), {"backend", "web", "static", "images", "scans"})
        for name, j in self.wf["jobs"].items():
            self.assertIn("timeout-minutes", j, name)

    def test_actions_are_pinned_to_a_version(self):
        uses = re.findall(r"uses:\s*(\S+)", self.text)
        self.assertGreaterEqual(len(uses), 8)
        for u in uses:
            self.assertRegex(u, r"@(v\d+|\d+\.\d+\.\d+)$", u)

    def test_backend_runs_the_valkey_tests_for_real(self):
        b = self.wf["jobs"]["backend"]
        self.assertIn("valkey", b["services"])
        self.assertRegex(b["services"]["valkey"]["image"], r":\d")
        self.assertTrue(b["env"]["REDIS_TEST_ADDR"].startswith("localhost:"))
        runs = " ".join(str(s.get("run", "")) for s in b["steps"])
        self.assertIn("-run 'Revoker|Limiter' ./internal/adapter/redis", runs)
        self.assertIn("go vet", runs)
        self.assertIn("-race", runs)

    def test_images_are_built_not_pushed_and_scanned(self):
        j = self.wf["jobs"]["images"]
        self.assertEqual(j["needs"], ["backend", "web", "static"])
        build = next(s for s in j["steps"] if s.get("uses", "").startswith("docker/build-push-action"))
        self.assertIs(build["with"]["push"], False)
        scan = next(s for s in j["steps"] if "aquasec/trivy:" in str(s.get("run", "")))
        self.assertIn("--exit-code 1", scan["run"])
        self.assertIn("--severity HIGH,CRITICAL", scan["run"])
        names = {m["name"] for m in j["strategy"]["matrix"]["include"]}
        self.assertEqual(names, {"api", "worker", "web"})

    def test_scans_cover_deps_secrets_and_files(self):
        t = self.text
        for needle in ("govulncheck", "gitleaks", "trivy fs", "--scanners vuln,secret,misconfig"):
            self.assertIn(needle, t)

    def test_trivy_runs_from_a_pinned_image_not_the_action(self):
        # trivy-action tags were force-pushed to malicious commits in 2026: do not depend on a movable tag
        self.assertNotIn("aquasecurity/trivy-action", self.text)
        for img in re.findall(r"aquasec/trivy:(\S+)", self.text):
            self.assertRegex(img, r"^\d+\.\d+\.\d+$")
        self.assertNotRegex(self.text, r"aquasec/trivy:(latest|0\.69\.[456])")

    def test_workflow_paths_match_real_folders(self):
        self.assertTrue((ROOT / "web" / "package-lock.json").exists())
        self.assertTrue((ROOT / "go.sum").exists())
        for c in ("api", "worker"):
            self.assertIn(f"CMD={c}", self.text)
        self.assertIn("ARG VITE_API_BASE", read("web/Dockerfile"))
        self.assertIn("ARG CMD", read("Dockerfile"))


class PowerShellEncodingTest(unittest.TestCase):
    """Windows PowerShell 5.1 reads a script without a BOM as ANSI, which breaks Thai text."""

    def test_non_ascii_ps1_has_bom(self):
        for f in (ROOT / "scripts").glob("*.ps1"):
            raw = f.read_bytes()
            if any(c > 127 for c in raw):
                self.assertTrue(raw.startswith(b"\xef\xbb\xbf"), f"{f.name} needs a UTF-8 BOM")


if __name__ == "__main__":
    unittest.main()
