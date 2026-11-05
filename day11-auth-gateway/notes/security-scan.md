# Day 11 security scan

Run at: 2026-11-05 17:09:50

## govulncheck ./... (exit code 0)

exit 0 = ไม่พบช่องโหว่ที่โค้ดเรียกใช้จริง, exit 3 = พบ (ดูรายการ)

```
No vulnerabilities found.
```

## gitleaks (สแกนไฟล์ในโฟลเดอร์ ซ่อนค่า) (exit code 1)

exit 0 = ไม่พบ secret, exit 1 = พบ

```

    ○
    │╲
    │ ○
    ○ ░
    ░    gitleaks

Finding:     JWT_SECRET=[1;3;mREDACTED[0m
Secret:      [1;3;mREDACTED[0m
RuleID:      generic-api-key
Entropy:     3.870657
File:        /repo/.env
Line:        1
Fingerprint: /repo/.env:generic-api-key:1

Finding:     GRAFANA_ADMIN_PASSWORD=[1;3;mREDACTED[0m
Secret:      [1;3;mREDACTED[0m
RuleID:      generic-api-key
Entropy:     3.522055
File:        /repo/.env
Line:        2
Fingerprint: /repo/.env:generic-api-key:2

Finding:     [1;3;mREDACTED[0m
Secret:      [1;3;mREDACTED[0m
RuleID:      private-key
Entropy:     6.009150
File:        /repo/certs/server.key
Line:        1
Fingerprint: /repo/certs/server.key:private-key:1

Finding:     secret: "[1;3;mREDACTED[0m"
Secret:      [1;3;mREDACTED[0m
RuleID:      generic-api-key
Entropy:     3.870657
File:        /repo/gateway/.rendered/kong.yml
Line:        10
Fingerprint: /repo/gateway/.rendered/kong.yml:generic-api-key:10

[90m10:09AM[0m [32mINF[0m [1mscanned ~532283 bytes (532.28 KB) in 366ms[0m
[90m10:09AM[0m [33mWRN[0m [1mleaks found: 4[0m
```

## Trivy: image api (HIGH,CRITICAL) (exit code 0)

exit 0 = สแกนจบ (ดูตารางว่ามีรายการหรือไม่) build: exit 0, save: exit 0

```
2026-11-05T10:09:59Z	INFO	[vuln] Vulnerability scanning is enabled
2026-11-05T10:10:00Z	INFO	Detected OS	family="debian" version="12.15"
2026-11-05T10:10:00Z	INFO	[debian] Detecting vulnerabilities...	os_version="12" pkg_num=5
2026-11-05T10:10:00Z	INFO	Number of language-specific files	num=1
2026-11-05T10:10:00Z	INFO	[gobinary] Detecting vulnerabilities...
2026-11-05T10:10:00Z	WARN	Using severities from other vendors for some vulnerabilities. Read https://trivy.dev/docs/v0.76/guide/scanner/vulnerability#severity-selection for details.

Report Summary

┌───────────────────────────────────────────┬──────────┬─────────────────┐
│                  Target                   │   Type   │ Vulnerabilities │
├───────────────────────────────────────────┼──────────┼─────────────────┤
│ /tmp/in/day11-api-scan.tar (debian 12.15) │  debian  │        0        │
├───────────────────────────────────────────┼──────────┼─────────────────┤
│ app                                       │ gobinary │        0        │
└───────────────────────────────────────────┴──────────┴─────────────────┘
Legend:
- '-': Not scanned
- '0': Clean (no security findings detected)
```

## Trivy: fs (vuln + secret + misconfig, รวม Dockerfile และ compose) (exit code 0)

ดูตารางรายการท้ายผล

```
2026-11-05T10:10:01Z	INFO	[vuln] Vulnerability scanning is enabled
2026-11-05T10:10:01Z	INFO	[misconfig] Misconfiguration scanning is enabled
2026-11-05T10:10:01Z	INFO	[checks-client] Using existing checks from cache	path="/root/.cache/trivy/policy/content"
2026-11-05T10:10:01Z	INFO	[secret] Secret scanning is enabled
2026-11-05T10:10:01Z	INFO	[secret] If your scanning is slow, please try '--scanners vuln,misconfig' to disable secret scanning
2026-11-05T10:10:01Z	INFO	[secret] Please see https://trivy.dev/docs/v0.76/guide/scanner/secret#recommendation for faster secret detection
2026-11-05T10:10:02Z	INFO	[npm] Run "npm install" to collect the license information of packages	dir="web/node_modules"
2026-11-05T10:10:02Z	INFO	Suppressing dependencies for development and testing. To display them, try the '--include-dev-deps' flag.
2026-11-05T10:10:02Z	INFO	Number of language-specific files	num=2
2026-11-05T10:10:02Z	INFO	[gomod] Detecting vulnerabilities...
2026-11-05T10:10:02Z	INFO	[npm] Detecting vulnerabilities...
2026-11-05T10:10:02Z	INFO	Detected config files	num=2
2026-11-05T10:10:02Z	WARN	Using severities from other vendors for some vulnerabilities. Read https://trivy.dev/docs/v0.76/guide/scanner/vulnerability#severity-selection for details.

Report Summary

┌───────────────────────┬────────────┬─────────────────┬─────────┬───────────────────┐
│        Target         │    Type    │ Vulnerabilities │ Secrets │ Misconfigurations │
├───────────────────────┼────────────┼─────────────────┼─────────┼───────────────────┤
│ go.mod                │   gomod    │        0        │    -    │         -         │
├───────────────────────┼────────────┼─────────────────┼─────────┼───────────────────┤
│ web/package-lock.json │    npm     │        0        │    -    │         -         │
├───────────────────────┼────────────┼─────────────────┼─────────┼───────────────────┤
│ Dockerfile            │ dockerfile │        -        │    -    │         0         │
├───────────────────────┼────────────┼─────────────────┼─────────┼───────────────────┤
│ web/Dockerfile        │ dockerfile │        -        │    -    │         0         │
├───────────────────────┼────────────┼─────────────────┼─────────┼───────────────────┤
│ certs/server.key      │    text    │        -        │    1    │         -         │
└───────────────────────┴────────────┴─────────────────┴─────────┴───────────────────┘
Legend:
- '-': Not scanned
- '0': Clean (no security findings detected)


certs/server.key (secrets)
==========================
Total: 1 (HIGH: 1, CRITICAL: 0)

HIGH: AsymmetricPrivateKey (private-key)
════════════════════════════════════════
Asymmetric Private Key
────────────────────────────────────────
 certs/server.key:2-27 (offset: 28 bytes)
────────────────────────────────────────
   1   [private-key-marker]
   2 ┌ ****************************************************************
   3 │ ****************************************************************
   4 │ ****************************************************************
   5 │ ****************************************************************
   6 │ ****************************************************************
   7 │ ****************************************************************
   8 │ ****************************************************************
   9 │ ****************************************************************
  10 │ ****************************************************************
  11 │ ****************************************************************
  12 │ ****************************************************************
  13 │ ****************************************************************
  14 │ ****************************************************************
  15 │ ****************************************************************
  16 │ ****************************************************************
  17 │ ****************************************************************
  18 │ ****************************************************************
  19 │ ****************************************************************
  20 │ ****************************************************************
  21 │ ****************************************************************
  22 │ ****************************************************************
  23 │ ****************************************************************
  24 │ ****************************************************************
  25 │ ****************************************************************
  26 │ ****************************************************************
  27 └ ************************
  28   [private-key-marker]
────────────────────────────────────────
```


