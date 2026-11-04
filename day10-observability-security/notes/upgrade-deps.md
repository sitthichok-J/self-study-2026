# Day 10 dependency upgrade

Run at: 2026-11-04 00:07:18

exit code: 0 (0 = go get, tidy, vet and test all passed)

```
go version go1.26.10 linux/amd64
go: downloading golang.org/x/text v0.42.0
go: downloading golang.org/x/sync v0.23.0
go: upgraded go 1.25.0 => 1.26.0
go: upgraded github.com/jackc/pgx/v5 v5.9.0 => v5.9.2
go: upgraded github.com/redis/go-redis/v9 v9.7.0 => v9.7.3
go: upgraded golang.org/x/net v0.57.0 => v0.60.0
go: upgraded golang.org/x/sync v0.22.0 => v0.23.0
go: upgraded golang.org/x/text v0.41.0 => v0.42.0
?   	github.com/sitthichok-J/self-study-2026/day10-observability-security/cmd/api	[no test files]
?   	github.com/sitthichok-J/self-study-2026/day10-observability-security/cmd/mint	[no test files]
?   	github.com/sitthichok-J/self-study-2026/day10-observability-security/cmd/worker	[no test files]
ok  	github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/adapter/http	0.027s
?   	github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/adapter/kafka	[no test files]
?   	github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/adapter/postgres	[no test files]
?   	github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/adapter/redis	[no test files]
ok  	github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/auth	0.009s
ok  	github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/baseline	0.017s
ok  	github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/config	0.007s
ok  	github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/domain	0.007s
?   	github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/memory	[no test files]
ok  	github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/obs	0.023s
?   	github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/port	[no test files]
ok  	github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/service	0.071s
github.com/sitthichok-J/self-study-2026/day10-observability-security
github.com/bsm/ginkgo/v2 v2.12.0
github.com/bsm/gomega v1.27.10
github.com/cespare/xxhash/v2 v2.2.0
github.com/davecgh/go-spew v1.1.1
github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f
github.com/jackc/pgpassfile v1.0.0
github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761
github.com/jackc/pgx/v5 v5.9.2
github.com/jackc/puddle/v2 v2.2.2
github.com/klauspost/compress v1.18.7
github.com/kr/pretty v0.3.0
github.com/pierrec/lz4/v4 v4.1.15
github.com/pmezard/go-difflib v1.0.0
github.com/redis/go-redis/v9 v9.7.3
github.com/segmentio/kafka-go v0.4.47
github.com/stretchr/objx v0.4.0
github.com/stretchr/testify v1.11.1
github.com/xdg-go/pbkdf2 v1.0.0
github.com/xdg-go/scram v1.1.2
github.com/xdg-go/stringprep v1.0.4
github.com/yuin/goldmark v1.4.13
golang.org/x/crypto v0.57.0
golang.org/x/mod v0.41.0
golang.org/x/net v0.60.0
golang.org/x/sync v0.23.0
golang.org/x/sys v0.48.0
golang.org/x/term v0.46.0
golang.org/x/text v0.42.0
golang.org/x/tools v0.49.0
golang.org/x/xerrors v0.0.0-20190717185122-a985d3407aa7
gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c
gopkg.in/yaml.v3 v3.0.1
```
