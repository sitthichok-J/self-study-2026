# Day 10 security scan

Run at: 2026-11-04 00:02:33

## govulncheck ./... (exit code 3)

exit 0 = ไม่พบช่องโหว่ที่โค้ดเรียกใช้จริง, exit 3 = พบ (ดูรายการ)

```
go: downloading golang.org/x/telemetry v0.0.0-20261010193011-f71dc92da45e
go: downloading golang.org/x/mod v0.42.0
go: downloading golang.org/x/tools v0.52.0
go: downloading golang.org/x/sync v0.24.0
=== Symbol Results ===

Vulnerability #1: GO-2026-5004
    SQL Injection via placeholder confusion with dollar quoted string literals
    in github.com/jackc/pgx
  More info: https://pkg.go.dev/vuln/GO-2026-5004
  Module: github.com/jackc/pgx/v5
    Found in: github.com/jackc/pgx/v5@v5.9.0
    Fixed in: github.com/jackc/pgx/v5@v5.9.2
    Example traces found:
      #1: internal/adapter/postgres/repo.go:98:27: postgres.Repo.List calls pgxpool.Pool.Query, which eventually calls sanitize.SanitizeSQL

Vulnerability #2: GO-2025-3540
    Potential out of order responses when CLIENT SETINFO times out during
    connection establishment in github.com/redis/go-redis
  More info: https://pkg.go.dev/vuln/GO-2025-3540
  Module: github.com/redis/go-redis/v9
    Found in: github.com/redis/go-redis/v9@v9.7.0
    Fixed in: github.com/redis/go-redis/v9@v9.7.3
    Example traces found:
      #1: internal/adapter/redis/cache.go:20:66: redis.Cache.Ping calls redis.cmdable.Ping, which eventually calls redis.baseClient.initConn
      #2: internal/adapter/redis/cache.go:20:66: redis.Cache.Ping calls redis.cmdable.Ping, which eventually calls redis.baseClient.initConn
      #3: internal/adapter/redis/cache.go:20:66: redis.Cache.Ping calls redis.cmdable.Ping, which eventually calls redis.baseClient.initConn

Your code is affected by 2 vulnerabilities from 2 modules.
This scan found no other vulnerabilities in packages you import or modules you
require.
Use '-show verbose' for more details.
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

[90m5:02PM[0m [32mINF[0m [1mscanned ~363849 bytes (363.85 KB) in 276ms[0m
[90m5:02PM[0m [33mWRN[0m [1mleaks found: 2[0m
```

## Trivy: image api (HIGH,CRITICAL) (exit code 0)

exit 0 = สแกนจบ (ดูตารางว่ามีรายการหรือไม่) build: exit 0, save: exit 0

```
2026-11-03T17:02:52Z	INFO	[vuln] Vulnerability scanning is enabled
2026-11-03T17:02:54Z	INFO	Detected OS	family="debian" version="12.15"
2026-11-03T17:02:54Z	INFO	[debian] Detecting vulnerabilities...	os_version="12" pkg_num=5
2026-11-03T17:02:54Z	INFO	Number of language-specific files	num=1
2026-11-03T17:02:54Z	INFO	[gobinary] Detecting vulnerabilities...
2026-11-03T17:02:54Z	WARN	Using severities from other vendors for some vulnerabilities. Read https://trivy.dev/docs/v0.76/guide/scanner/vulnerability#severity-selection for details.

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
2026-11-03T17:02:55Z	INFO	[vuln] Vulnerability scanning is enabled
2026-11-03T17:02:55Z	INFO	[misconfig] Misconfiguration scanning is enabled
2026-11-03T17:02:55Z	INFO	[checks-client] Using existing checks from cache	path="/root/.cache/trivy/policy/content"
2026-11-03T17:02:56Z	INFO	[secret] Secret scanning is enabled
2026-11-03T17:02:56Z	INFO	[secret] If your scanning is slow, please try '--scanners vuln,misconfig' to disable secret scanning
2026-11-03T17:02:56Z	INFO	[secret] Please see https://trivy.dev/docs/v0.76/guide/scanner/secret#recommendation for faster secret detection
2026-11-03T17:02:56Z	INFO	Number of language-specific files	num=1
2026-11-03T17:02:56Z	INFO	[gomod] Detecting vulnerabilities...
2026-11-03T17:02:56Z	INFO	Detected config files	num=1
2026-11-03T17:02:56Z	WARN	Using severities from other vendors for some vulnerabilities. Read https://trivy.dev/docs/v0.76/guide/scanner/vulnerability#severity-selection for details.

Report Summary

┌────────────┬────────────┬─────────────────┬─────────┬───────────────────┐
│   Target   │    Type    │ Vulnerabilities │ Secrets │ Misconfigurations │
├────────────┼────────────┼─────────────────┼─────────┼───────────────────┤
│ go.mod     │   gomod    │        1        │    -    │         -         │
├────────────┼────────────┼─────────────────┼─────────┼───────────────────┤
│ Dockerfile │ dockerfile │        -        │    -    │         0         │
└────────────┴────────────┴─────────────────┴─────────┴───────────────────┘
Legend:
- '-': Not scanned
- '0': Clean (no security findings detected)


go.mod (gomod)
==============
Total: 1 (HIGH: 1, CRITICAL: 0)

┌──────────────────┬────────────────┬──────────┬────────┬───────────────────┬───────────────┬────────────────────────────────────────────────────────────┐
│     Library      │ Vulnerability  │ Severity │ Status │ Installed Version │ Fixed Version │                           Title                            │
├──────────────────┼────────────────┼──────────┼────────┼───────────────────┼───────────────┼────────────────────────────────────────────────────────────┤
│ golang.org/x/net │ CVE-2026-78669 │ HIGH     │ fixed  │ v0.57.0           │ 0.60.0        │ net/http: net/http/internal/http2: golang.org/x/net/http2: │
│                  │                │          │        │                   │               │ net/http: Denial of Service via excessive HTTP/2 SETTINGS  │
│                  │                │          │        │                   │               │ frames...                                                  │
│                  │                │          │        │                   │               │ https://avd.aquasec.com/nvd/cve-2026-78669                 │
└──────────────────┴────────────────┴──────────┴────────┴───────────────────┴───────────────┴────────────────────────────────────────────────────────────┘
```


