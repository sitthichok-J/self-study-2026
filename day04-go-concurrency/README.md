# Day 4 — Go concurrency

worker pool ที่ cancel ได้และ shutdown แบบ graceful พร้อม pipeline, fan-in, errgroup (ฉบับเขียนเอง) และตัวอย่าง race/leak
กฎของฉันอยู่ที่ [`my-rules.md`](my-rules.md)

## รัน

```
go run ./cmd/pool -workers 4 -jobs 20 -delay 200ms     # กด Ctrl+C เพื่อดู graceful shutdown
```

ต้องใช้ Go 1.22 ขึ้นไป (โค้ดพึ่งพฤติกรรม loop variable แยกต่อรอบ)

## โครงสร้าง

| ที่ | หน้าที่ | แบบฝึกหัด / หัวข้อ |
|---|---|---|
| `internal/pool` | worker pool: รับงานจาก channel, N worker, cancel, graceful shutdown | งานส่งมอบหลัก |
| `internal/pipeline` | Generate → Map → Collect, Merge (fan-in), หยุดเมื่อ ctx cancel | ข้อ 4, fan-out/fan-in |
| `internal/race` | counter ด้วย mutex และ atomic, ตัวที่มี race (ใต้ build tag `racedemo`) | ข้อ 1 |
| `internal/leak` | goroutine leak และวิธีแก้ด้วย context | ข้อ 2 |
| `internal/egroup` | errgroup ฉบับเล็ก (Go, Wait, SetLimit, cancel เมื่อ error แรก) | errgroup |
| `internal/fetch` | ดึง URL หลายตัวพร้อมกัน จำกัด concurrency | ข้อ 3 |
| `cmd/pool` | CLI สาธิต worker pool + จัดการ Ctrl+C | ต่อทุกอย่างเข้าด้วยกัน |

หมายเหตุเรื่อง errgroup: ตอนสร้างในสภาพแวดล้อมที่ดึง `golang.org/x/sync` ไม่ได้ เลยเขียน `internal/egroup` ที่ API เหมือนกัน โปรเจกต์จริงให้ใช้ `golang.org/x/sync/errgroup` (เปลี่ยนแค่ import)
ไม่ได้ทำ rate limit (`time.Ticker` / `x/time/rate`) ในรอบนี้

## Test

```
go test ./... -race                                   # ทั้งหมด (ต้องผ่านด้วย -race)
go test ./... -race -count=30                         # รันซ้ำหาเทสต์ที่ไม่นิ่ง
go test -tags racedemo -race ./internal/race -run Unsafe   # ดูรายงาน DATA RACE (เจตนาให้แดง)
go test -run Baseline -update                         # อัปเดต baseline หลังเปลี่ยนพฤติกรรมโดยตั้งใจ
```

`TestBaseline` เทียบผลที่ต้องแน่นอน (ยอด counter, ผล pipeline, ผล pool) กับ `testdata/baseline.golden`
