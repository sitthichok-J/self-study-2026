# Day 10 security scan

Run at: 2026-11-04 00:10:46

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
Entropy:     3.898133
File:        /repo/.env
Line:        1
Fingerprint: /repo/.env:generic-api-key:1

Finding:     t.Setenv("JWT_SECRET", "[1;3;mREDACTED[0m"
Secret:      [1;3;mREDACTED[0m
RuleID:      generic-api-key
Entropy:     4.000000
File:        /repo/internal/config/config_test.go
Line:        42
Fingerprint: /repo/internal/config/config_test.go:generic-api-key:42

[90m5:10PM[0m [32mINF[0m [1mscanned ~370207 bytes (370.21 KB) in 269ms[0m
[90m5:10PM[0m [33mWRN[0m [1mleaks found: 2[0m
```

## Trivy: image api (HIGH,CRITICAL) (exit code 0)

exit 0 = สแกนจบ (ดูตารางว่ามีรายการหรือไม่) build: exit 0, save: exit 0

```
2026-11-03T17:10:55Z	INFO	[vuln] Vulnerability scanning is enabled
2026-11-03T17:10:56Z	INFO	Detected OS	family="debian" version="12.15"
2026-11-03T17:10:56Z	INFO	[debian] Detecting vulnerabilities...	os_version="12" pkg_num=5
2026-11-03T17:10:56Z	INFO	Number of language-specific files	num=1
2026-11-03T17:10:56Z	INFO	[gobinary] Detecting vulnerabilities...
2026-11-03T17:10:56Z	WARN	Using severities from other vendors for some vulnerabilities. Read https://trivy.dev/docs/v0.76/guide/scanner/vulnerability#severity-selection for details.

Report Summary

┌───────────────────────────────────────────┬──────────┬─────────────────┐
│                  Target                   │   Type   │ Vulnerabilities │
├───────────────────────────────────────────┼──────────┼─────────────────┤
│ /tmp/in/day10-api-scan.tar (debian 12.15) │  debian  │        0        │
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
2026-11-03T17:10:57Z	INFO	[vuln] Vulnerability scanning is enabled
2026-11-03T17:10:57Z	INFO	[misconfig] Misconfiguration scanning is enabled
2026-11-03T17:10:57Z	INFO	[checks-client] Using existing checks from cache	path="/root/.cache/trivy/policy/content"
2026-11-03T17:10:58Z	INFO	[secret] Secret scanning is enabled
2026-11-03T17:10:58Z	INFO	[secret] If your scanning is slow, please try '--scanners vuln,misconfig' to disable secret scanning
2026-11-03T17:10:58Z	INFO	[secret] Please see https://trivy.dev/docs/v0.76/guide/scanner/secret#recommendation for faster secret detection
2026-11-03T17:10:59Z	INFO	Number of language-specific files	num=1
2026-11-03T17:10:59Z	INFO	[gomod] Detecting vulnerabilities...
2026-11-03T17:10:59Z	INFO	Detected config files	num=1
2026-11-03T17:10:59Z	WARN	Using severities from other vendors for some vulnerabilities. Read https://trivy.dev/docs/v0.76/guide/scanner/vulnerability#severity-selection for details.

Report Summary

┌────────────┬────────────┬─────────────────┬─────────┬───────────────────┐
│   Target   │    Type    │ Vulnerabilities │ Secrets │ Misconfigurations │
├────────────┼────────────┼─────────────────┼─────────┼───────────────────┤
│ go.mod     │   gomod    │        0        │    -    │         -         │
├────────────┼────────────┼─────────────────┼─────────┼───────────────────┤
│ Dockerfile │ dockerfile │        -        │    -    │         0         │
└────────────┴────────────┴─────────────────┴─────────┴───────────────────┘
Legend:
- '-': Not scanned
- '0': Clean (no security findings detected)
```


