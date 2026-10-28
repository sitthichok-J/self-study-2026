# Profile note: /debug/report-slow (before / after)

เครื่องวัด: sandbox 2 vCPU, Go 1.22+, ข้อมูล 20,000 claims / 1,000 โรงพยาบาล
วิธีวัด: `go test -run xxx -bench . -benchmem ./internal/report` แล้ว `-cpuprofile` + `go tool pprof -top`

## ก่อนแก้ (BuildSlow)

- เวลา 266.6 ms/op, 11.9 MB/op, 4,024 allocs/op
- pprof (flat): runtime.duffcopy 38.9%, BuildSlow เอง 22.9%, memeqbody 16.8%, memequal 9.2%
- อ่านผล: เวลาหมดไปกับการ copy struct ทีละตัวและเทียบ string ซ้ำๆ ไม่ใช่การจัดสรร memory
  สาเหตุคือค้นหาโรงพยาบาลด้วยการวนลิสต์ทุกครั้ง (O(n·h)) และเรียงด้วย bubble sort ที่ copy ค่าสลับกัน

## แก้อะไร (Build)

1. รวมยอดด้วย map[hospital] เดียว ผ่านข้อมูลรอบเดียว (O(n))
2. เรียงด้วย sort.Slice (O(h log h))
3. ต่อ string ด้วย strings.Builder แทน `+=`

เทสต์ `internal/report` ยืนยันว่าผลลัพธ์ของสองฟังก์ชันเหมือนกันทุกกรณี

## หลังแก้

- เวลา 1.21 ms/op, 301 KB/op, 4,038 allocs/op
- เร็วขึ้นประมาณ 220 เท่า, memory ต่อครั้งลดประมาณ 40 เท่า
- pprof: ตอนนี้เวลากระจายอยู่ที่ map access (mapaccess1_faststr ~29%) ซึ่งเป็นงานที่ต้องทำจริง ไม่มีจุดร้อนเดียวแล้ว
- allocs/op เกือบเท่าเดิม (4,024 → 4,038) เพราะยังสร้าง string/struct ต่อโรงพยาบาล ไม่ใช่คอขวดแล้ว

## ตัวอย่างต่อ string 1,000 ชิ้น (internal/strcat)

| วิธี | ns/op | B/op | allocs/op |
|---|---|---|---|
| `+=` (Plus) | 966,800 | 2,033,559 | 1,900 |
| strings.Builder | 27,115 | 15,424 | 912 |
| fmt.Sprintf | 1,050,338 | 2,053,751 | 2,744 |

Builder เร็วกว่า `+=` ประมาณ 36 เท่า เพราะ `+=` copy สายอักขระทั้งก้อนทุกรอบ (O(n²) ในจำนวนไบต์)

## ข้อจำกัด

- ตัวเลขมาจาก sandbox 2 vCPU ค่าสัมบูรณ์จะต่างบนเครื่องอื่น แต่สัดส่วนควรใกล้เคียง
- วัดผ่าน benchmark ไม่ใช่ผ่าน HTTP จริง; `/debug/report-slow` เปิดบน pprof listener (PPROF_ADDR) เท่านั้น
  ถ้าจะ profile ผ่าน server ใช้ `go tool pprof http://127.0.0.1:6060/debug/pprof/profile?seconds=10`
