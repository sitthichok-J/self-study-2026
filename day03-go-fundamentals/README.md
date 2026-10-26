# Day 3 — Go fundamentals

CLI เก็บรายการงานลงไฟล์ JSON ใช้เป็นโจทย์ให้ครบหัวข้อ: interface, custom error, wrap error, table-driven test
กับชุดตัวอย่างพิสูจน์พฤติกรรมที่คนมักพลาด (slice/map/nil interface/defer/recover)

## รัน

```
go run ./cmd/tasks -file tasks.json add เขียนเทสต์
go run ./cmd/tasks -file tasks.json list
go run ./cmd/tasks -file tasks.json done 1
go run ./cmd/tasks -file tasks.json rm 1
```

ต้องใช้ Go 1.22 ขึ้นไป (`go.mod` ตั้ง 1.22 เพราะพึ่งพฤติกรรม loop variable ใหม่)

## โครงสร้าง

| ที่ | หน้าที่ | หัวข้อ |
|---|---|---|
| `internal/tasks` | logic + `Storage` interface (ประกาศฝั่งผู้ใช้), `NotFoundError`, `ErrEmptyTitle` | struct/interface, error |
| `internal/store` | `Memory` และ `File` (JSON) ใช้ชุดเทสต์เดียวกัน | interface หลาย implementation |
| `internal/demo` | ตัวอย่างพิสูจน์: append แชร์ array, nil map, nil interface, defer | slice/map, defer |
| `internal/httpmw` | `Recover` middleware คืน 500 | panic/recover |
| `cmd/tasks` | CLI | ต่อทุกอย่างเข้าด้วยกัน, log error ที่ชั้นบนสุดครั้งเดียว |
| `mistakes.md` | 5 ข้อผิดพลาดพร้อมวิธีหลีกเลี่ยง | งานส่งมอบ |

## Test

```
go test ./...                       # ทั้งหมด
go test ./... -race
go test -run Baseline -update       # อัปเดต baseline หลังเปลี่ยนพฤติกรรมโดยตั้งใจ (ดู diff ของ testdata/baseline.golden)
```

`TestBaseline` เทียบผลของตัวอย่างทั้งหมดกับ `testdata/baseline.golden`
