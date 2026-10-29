"""Static checks for day07-terraform (no terraform binary needed).
These do NOT prove that terraform accepts the code: scripts/run-all.ps1 does that on a real machine.
Run: python3 -m unittest discover -s tests
"""
import json
import re
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
INFRA = ROOT / "infra"
VPC = ROOT / "infra-vpc"
BASELINE = json.loads((ROOT / "testdata" / "baseline.json").read_text(encoding="utf-8"))


def tf_text(folder: Path) -> str:
    return "\n".join(p.read_text(encoding="utf-8") for p in sorted(folder.rglob("*.tf")))


def blocks(text: str, kind: str):
    """yield (labels, body) for top-level `kind "a" "b" {` blocks, matching braces."""
    for m in re.finditer(rf'^{kind}\s+((?:"[^"]+"\s*)+)\{{', text, re.M):
        labels = re.findall(r'"([^"]+)"', m.group(1))
        depth, i = 1, m.end()
        while depth and i < len(text):
            depth += {"{": 1, "}": -1}.get(text[i], 0)
            i += 1
        yield labels, text[m.end():i - 1]


class Braces(unittest.TestCase):
    def test_balanced_braces_and_brackets(self):
        for f in list(ROOT.rglob("*.tf")):
            t = re.sub(r'"(?:\\.|[^"\\])*"', '""', f.read_text(encoding="utf-8"))  # drop strings
            t = re.sub(r"#.*", "", t)
            for a, b in ("{}", "[]", "()"):
                self.assertEqual(t.count(a), t.count(b), f"{f.name}: unbalanced {a}{b}")


class Layout(unittest.TestCase):
    def test_expected_files(self):
        for p in BASELINE["files"]:
            self.assertTrue((ROOT / p).exists(), p)

    def test_provider_pinned(self):
        for d in (INFRA, VPC):
            t = (d / "versions.tf").read_text(encoding="utf-8")
            self.assertRegex(t, r'version\s*=\s*"~>\s*5\.\d+"')
            self.assertRegex(t, r'required_version\s*=\s*">=\s*1\.\d+"')

    def test_gitignore_blocks_state(self):
        g = (ROOT / ".gitignore").read_text(encoding="utf-8")
        for pat in (".terraform/", "*.tfstate", "*.tfplan"):
            self.assertIn(pat, g)

    def test_variables_and_outputs_documented(self):
        for d in (INFRA, VPC, INFRA / "modules" / "queue"):
            t = tf_text(d) if d != INFRA else (d / "variables.tf").read_text(encoding="utf-8") + (d / "outputs.tf").read_text(encoding="utf-8")
            for kind in ("variable", "output"):
                for labels, body in blocks(t, kind):
                    self.assertIn("description", body, f"{d.name}: {kind} {labels[0]} has no description")
                    if kind == "variable":
                        self.assertRegex(body, r"\btype\s*=", f"{labels[0]} has no type")


class Secrets(unittest.TestCase):
    def test_no_credentials_in_files(self):
        for f in ROOT.rglob("*"):
            if not f.is_file() or f.suffix not in {".tf", ".tfvars", ".yml", ".ps1", ".bat", ".md"}:
                continue
            t = f.read_text(encoding="utf-8", errors="ignore")
            self.assertIsNone(re.search(r"AKIA[0-9A-Z]{16}", t), f"{f}: looks like an AWS access key")
            self.assertIsNone(re.search(r'(access_key|secret_key|password)\s*=\s*"[^"]+"', t), f"{f}: hard-coded credential")


class S3Checks(unittest.TestCase):
    def setUp(self):
        self.t = tf_text(INFRA)

    def test_public_access_fully_blocked(self):
        b = dict(((l[0], body) for l, body in blocks(self.t, "resource") if l[0] == "aws_s3_bucket_public_access_block"))
        self.assertIn("aws_s3_bucket_public_access_block", b)
        for k in ("block_public_acls", "block_public_policy", "ignore_public_acls", "restrict_public_buckets"):
            self.assertRegex(b["aws_s3_bucket_public_access_block"], rf"{k}\s*=\s*true")

    def test_versioning_and_encryption(self):
        self.assertIn('resource "aws_s3_bucket_versioning"', self.t)
        self.assertRegex(self.t, r'status\s*=\s*"Enabled"')
        self.assertRegex(self.t, r'sse_algorithm\s*=\s*"AES256"')


class SqsChecks(unittest.TestCase):
    def setUp(self):
        self.t = tf_text(INFRA / "modules" / "queue")

    def test_dlq_and_redrive(self):
        self.assertIn('resource "aws_sqs_queue" "dlq"', self.t)
        self.assertIn("deadLetterTargetArn", self.t)
        self.assertIn("aws_sqs_queue.dlq.arn", self.t)
        self.assertRegex(self.t, r"maxReceiveCount\s*=\s*var\.max_receive_count")

    def test_module_used_by_root(self):
        self.assertRegex(tf_text(INFRA), r'module "claims_queue"\s*\{[^}]*source\s*=\s*"\./modules/queue"')

    def test_max_receive_default_is_positive(self):
        m = re.search(r'variable "max_receive_count"\s*\{.*?default\s*=\s*(\d+)', tf_text(INFRA / "modules" / "queue"), re.S)
        self.assertGreaterEqual(int(m.group(1)), 1)


class IamChecks(unittest.TestCase):
    ALLOWED = set(BASELINE["iam_allowed_actions"])

    def setUp(self):
        self.t = (INFRA / "iam.tf").read_text(encoding="utf-8")

    def test_no_wildcard_actions_or_resources(self):
        for labels, body in blocks(self.t, "data"):
            if labels[0] != "aws_iam_policy_document":
                continue
            for acts in re.findall(r"actions\s*=\s*\[(.*?)\]", body, re.S):
                for a in re.findall(r'"([^"]+)"', acts):
                    self.assertNotIn("*", a, f"wildcard action {a} in {labels[1]}")
            for res in re.findall(r"resources\s*=\s*\[(.*?)\]", body, re.S):
                self.assertNotRegex(res, r'"\*"', f"wildcard resource in {labels[1]}")

    def test_actions_are_only_the_expected_ones(self):
        used = set(re.findall(r'"((?:sqs|s3):[A-Za-z]+)"', self.t))
        self.assertTrue(used <= self.ALLOWED, f"unexpected actions: {used - self.ALLOWED}")

    def test_consumer_cannot_send_or_purge(self):
        body = next(b for l, b in blocks(self.t, "data") if l[1] == "consumer")
        for bad in ("SendMessage", "PurgeQueue", "DeleteQueue", "s3:"):
            self.assertNotIn(bad, body)

    def test_consumer_can_receive_and_delete(self):
        body = next(b for l, b in blocks(self.t, "data") if l[1] == "consumer")
        for need in ("sqs:ReceiveMessage", "sqs:DeleteMessage"):
            self.assertIn(need, body)

    def test_trust_policy_is_ecs_tasks(self):
        self.assertIn("ecs-tasks.amazonaws.com", self.t)


class VpcChecks(unittest.TestCase):
    def setUp(self):
        self.t = tf_text(VPC)

    def test_private_route_table_has_no_default_route(self):
        body = next(b for l, b in blocks(self.t, "resource") if l == ["aws_route_table", "private"])
        self.assertNotIn("0.0.0.0/0", body)

    def test_public_route_goes_to_igw(self):
        body = next(b for l, b in blocks(self.t, "resource") if l == ["aws_route_table", "public"])
        self.assertIn("aws_internet_gateway.main.id", body)

    def test_no_nat_gateway(self):
        self.assertNotIn("aws_nat_gateway", self.t)

    def test_db_only_from_app_and_app_only_from_alb(self):
        db = next(b for l, b in blocks(self.t, "resource") if l == ["aws_security_group", "db"])
        app = next(b for l, b in blocks(self.t, "resource") if l == ["aws_security_group", "app"])
        self.assertIn("aws_security_group.app.id", db)
        self.assertNotIn("cidr_blocks", db)
        self.assertIn("aws_security_group.alb.id", app)
        self.assertIn("8080", app)
        self.assertIn("5432", db)

    def test_two_azs_by_default(self):
        m = re.search(r'variable "azs".*?default\s*=\s*\[(.*?)\]', self.t, re.S)
        self.assertEqual(len(re.findall(r'"[^"]+"', m.group(1))), 2)

    def test_plan_works_offline(self):
        p = (VPC / "provider.tf").read_text(encoding="utf-8")
        for k in ("skip_credentials_validation", "skip_metadata_api_check", "skip_requesting_account_id"):
            self.assertRegex(p, rf"{k}\s*=\s*true")


class LocalStack(unittest.TestCase):
    def test_compose(self):
        c = (ROOT / "docker-compose.yml").read_text(encoding="utf-8")
        self.assertIn("4566:4566", c)
        self.assertRegex(c, r"image:\s*localstack/localstack:\d+\.\d+")
        self.assertIn("healthcheck", c)
        for s in ("s3", "sqs", "iam", "sts", "ec2"):
            self.assertIn(s, re.search(r"SERVICES:\s*(\S+)", c).group(1).split(","))

    def test_provider_endpoints_cover_services(self):
        p = (INFRA / "provider.tf").read_text(encoding="utf-8")
        for s in ("s3", "sqs", "iam", "sts", "ec2"):
            self.assertRegex(p, rf"\b{s}\s*=\s*var\.localstack_endpoint")
        self.assertIn("var.use_localstack", p)

    def test_default_endpoint_matches_compose_service_name(self):
        v = (INFRA / "variables.tf").read_text(encoding="utf-8")
        self.assertIn("http://localstack:4566", v)
        self.assertIn("name: day07", (ROOT / "docker-compose.yml").read_text(encoding="utf-8"))


class RunScript(unittest.TestCase):
    def test_script_cleans_up_and_writes_results(self):
        s = (ROOT / "scripts" / "run-all.ps1").read_text(encoding="utf-8")
        self.assertIn("destroy", s)
        self.assertIn("compose down", s)
        self.assertIn("run-results.md", s)
        self.assertIn("hashicorp/terraform:1.9", s)

    def test_bat_launcher(self):
        self.assertIn("run-all.ps1", (ROOT / "scripts" / "run-all.bat").read_text(encoding="utf-8"))


class Baseline(unittest.TestCase):
    def test_resource_inventory_matches_baseline(self):
        inv = {}
        for d in (INFRA, VPC):
            for labels, _ in blocks(tf_text(d), "resource"):
                inv.setdefault(d.name, []).append(labels[0])
        got = {k: sorted(v) for k, v in inv.items()}
        self.assertEqual(got, BASELINE["resources"])


if __name__ == "__main__":
    unittest.main()
