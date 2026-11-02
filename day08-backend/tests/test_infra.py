"""Static checks for compose / migrations / Dockerfile / Go wiring (stdlib + PyYAML only).
They do NOT start anything: scripts/run-all.ps1 does that on a real machine.
Run: python3 -m unittest discover -s tests
"""
import json
import re
import unittest
from pathlib import Path

import yaml

ROOT = Path(__file__).resolve().parent.parent
COMPOSE = yaml.safe_load((ROOT / "docker-compose.yml").read_text(encoding="utf-8"))
SVC = COMPOSE["services"]
UP = (ROOT / "migrations" / "000001_init.up.sql").read_text(encoding="utf-8")
DOWN = (ROOT / "migrations" / "000001_init.down.sql").read_text(encoding="utf-8")
PORT = (ROOT / "internal" / "port" / "port.go").read_text(encoding="utf-8")
BASE = json.loads((ROOT / "testdata" / "baseline.json").read_text(encoding="utf-8"))
GO_FILES = [p for p in ROOT.rglob("*.go")]


class Compose(unittest.TestCase):
    def test_expected_services(self):
        self.assertEqual(set(SVC), {"postgres", "valkey", "kafka", "kafka-init", "migrate", "api", "worker"})

    def test_project_name(self):
        self.assertEqual(COMPOSE["name"], "day08")  # run-all.ps1 inspects day08-api-1 etc.

    def test_images_are_pinned(self):
        for name, s in SVC.items():
            if "image" in s:
                tag = s["image"].split(":")[-1]
                self.assertNotIn(tag, ("latest", s["image"]), f"{name} must pin a tag")

    def test_start_order(self):
        for name in ("api", "worker"):
            dep = SVC[name]["depends_on"]
            self.assertEqual(dep["migrate"]["condition"], "service_completed_successfully")
            self.assertEqual(dep["kafka-init"]["condition"], "service_completed_successfully")
        self.assertEqual(SVC["migrate"]["depends_on"]["postgres"]["condition"], "service_healthy")
        self.assertEqual(SVC["kafka-init"]["depends_on"]["kafka"]["condition"], "service_healthy")

    def test_healthchecks(self):
        for name in ("postgres", "valkey", "kafka"):
            self.assertIn("healthcheck", SVC[name], name)

    def test_topics_match_go_constants(self):
        cmd = " ".join(SVC["kafka-init"]["command"])
        for topic in re.findall(r'Topic\w+\s*=\s*"([^"]+)"', PORT):
            self.assertIn(f"--topic {topic} ", cmd + " ", f"topic {topic} not created by kafka-init")
        self.assertEqual(len(re.findall(r'Topic\w+\s*=\s*"', PORT)), 2)

    def test_auto_create_off(self):
        self.assertEqual(str(SVC["kafka"]["environment"]["KAFKA_AUTO_CREATE_TOPICS_ENABLE"]).lower(), "false")

    def test_services_use_internal_hostnames(self):
        for name in ("api", "worker"):
            env = SVC[name]["environment"]
            self.assertEqual(env["KAFKA_BROKERS"], "kafka:9092")
            self.assertEqual(env["REDIS_ADDR"], "valkey:6379")
        self.assertIn("@postgres:5432/", COMPOSE["x-db-url"])

    def test_only_api_publishes_http(self):
        self.assertIn("8080:8080", SVC["api"]["ports"])
        self.assertNotIn("ports", SVC["worker"])

    def test_build_args_pick_binary(self):
        self.assertEqual(SVC["api"]["build"]["args"]["CMD"], "api")
        self.assertEqual(SVC["worker"]["build"]["args"]["CMD"], "worker")
        for c in ("api", "worker"):
            self.assertTrue((ROOT / "cmd" / c / "main.go").exists())

    def test_worker_group_matches_default(self):
        cfg = (ROOT / "internal" / "config" / "config.go").read_text(encoding="utf-8")
        self.assertIn(SVC["worker"]["environment"]["KAFKA_GROUP"], cfg)


class Migration(unittest.TestCase):
    def test_tables(self):
        self.assertIn("CREATE TABLE claims", UP)
        self.assertIn("CREATE TABLE outbox", UP)
        self.assertIn("DROP TABLE IF EXISTS claims", DOWN)
        self.assertIn("DROP TABLE IF EXISTS outbox", DOWN)

    def test_idempotency_key_unique(self):
        self.assertRegex(UP, r"UNIQUE \(idempotency_key\)")

    def test_status_values_match_go(self):
        dom = (ROOT / "internal" / "domain" / "claim.go").read_text(encoding="utf-8")
        for s in re.findall(r'Status\w+\s+Status = "(\w+)"', dom):
            self.assertIn(f"'{s}'", UP)
        self.assertEqual(len(re.findall(r'Status\w+\s+Status = "', dom)), 3)

    def test_fake_ref_check_matches_go_regex(self):
        dom = (ROOT / "internal" / "domain" / "claim.go").read_text(encoding="utf-8")
        go_re = re.search(r"regexp.MustCompile\(`(\^FAKE[^`]+)`\)", dom).group(1)
        self.assertIn(go_re, UP)

    def test_amount_is_integer_not_float(self):
        self.assertRegex(UP, r"amount\s+bigint")
        self.assertNotRegex(UP, r"(?i)\b(float|double|real|numeric)\b")

    def test_partial_index_for_unsent_outbox(self):
        self.assertRegex(UP, r"outbox_unsent_idx ON outbox \(id\) WHERE sent_at IS NULL")

    def test_up_down_are_a_pair(self):
        names = sorted(p.name for p in (ROOT / "migrations").iterdir())
        self.assertEqual(names, ["000001_init.down.sql", "000001_init.up.sql"])


class Dockerfile(unittest.TestCase):
    D = (ROOT / "Dockerfile").read_text(encoding="utf-8")

    def test_multistage_nonroot(self):
        self.assertIn("AS build", self.D)
        self.assertIn("distroless", self.D)
        self.assertIn("USER 65532:65532", self.D)

    def test_static_binary(self):
        self.assertIn("CGO_ENABLED=0", self.D)


class GoWiring(unittest.TestCase):
    def test_third_party_imports_are_in_go_mod(self):
        mod = (ROOT / "go.mod").read_text(encoding="utf-8")
        own = "github.com/sitthichok-J/self-study-2026/day08-backend"
        seen = set()
        for f in GO_FILES:
            for imp in re.findall(r'^\s*(?:\w+\s+)?"([\w.\-]+\.[\w.\-]+/[^"]+)"', f.read_text(encoding="utf-8"), re.M):
                if not imp.startswith(own):
                    seen.add(imp)
        self.assertTrue(seen)
        for imp in seen:
            self.assertTrue(any(imp.startswith(m) for m in re.findall(r"^\s+(\S+)\s+v", mod, re.M)), imp)

    def test_service_depends_only_on_ports(self):
        for f in (ROOT / "internal" / "service").glob("*.go"):
            if f.name.endswith("_test.go"):
                continue
            t = f.read_text(encoding="utf-8")
            self.assertNotIn("/adapter/", t, f.name)
            self.assertNotRegex(t, r"jackc/pgx|segmentio|go-redis", f.name)

    def test_domain_has_no_io_imports(self):
        for f in (ROOT / "internal" / "domain").glob("*.go"):
            if f.name.endswith("_test.go"):
                continue
            t = f.read_text(encoding="utf-8")
            for bad in ("net/http", "database/sql", "/adapter/", "os\""):
                self.assertNotIn(bad, t, f.name)

    def test_no_secrets_in_go(self):
        for f in GO_FILES:
            self.assertNotRegex(f.read_text(encoding="utf-8"), r"(?i)password\s*[:=]\s*\"[^\"]+\"", f.name)

    def test_commit_only_after_handler_in_kafka_adapter(self):
        t = (ROOT / "internal" / "adapter" / "kafka" / "kafka.go").read_text(encoding="utf-8")
        self.assertLess(t.index("h(hctx, pm)"), t.index("CommitMessages"))
        self.assertNotIn("CommitInterval", t)  # no auto-commit

    def test_all_adapters_use_ports(self):
        for name in ("postgres", "kafka", "redis"):
            self.assertTrue(list((ROOT / "internal" / "adapter" / name).glob("*.go")))


class Baseline(unittest.TestCase):
    def test_baseline_has_expected_sections(self):
        self.assertEqual(set(BASE), {"decisions", "http", "limits", "topics", "transitions"})

    def test_http_baseline_codes(self):
        self.assertEqual(BASE["http"]["POST create"], 201)
        self.assertEqual(BASE["http"]["POST replay"], 200)
        self.assertEqual(BASE["http"]["POST conflict"], 409)


class FakeDataOnly(unittest.TestCase):
    def test_no_13_digit_ids_in_repo(self):
        for p in ROOT.rglob("*"):
            if p.is_file() and p.suffix in {".go", ".md", ".sql", ".yml", ".ps1", ".json", ".mmd"} and "run-results" not in p.name:
                m = re.search(r"(?<!\d)\d{13}(?!\d)", p.read_text(encoding="utf-8"))
                # the only allowed hit is the deliberate 'real-looking' negative example
                if m:
                    self.assertIn(m.group(0), {"1100100123456", "1234567890123"}, p.name)


if __name__ == "__main__":
    unittest.main()
