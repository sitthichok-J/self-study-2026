// Package demo รวมตัวอย่างพิสูจน์พฤติกรรมของ Go ที่คนมักพลาด (แบบฝึกหัดข้อ 1, 2, 5, 6)
package demo

// AppendSharing สร้าง slice a ที่ cap เหลือ แล้ว append สองครั้งจากตัวเดียวกัน
// คืนค่าว่า b กับ c ใช้ array เดียวกันหรือไม่ (ถ้าใช่ c จะทับค่าของ b)
func AppendSharing() (shared bool, bFirst, cFirst int) {
	a := make([]int, 3, 10) // len 3, cap 10 -> append ไม่ reallocate
	b := append(a, 100)
	c := append(a, 200)
	return &b[0] == &c[0], b[3], c[3]
}

// AppendRealloc เหมือนด้านบนแต่ cap เต็ม -> append แล้ว reallocate แยก array
func AppendRealloc() (shared bool, bLast, cLast int) {
	a := make([]int, 3, 3)
	b := append(a, 100)
	c := append(a, 200)
	return &b[0] == &c[0], b[3], c[3]
}

// SubSliceShares: sub-slice แก้ค่าแล้วกระทบ slice ต้นทาง
func SubSliceShares() (orig []int) {
	orig = []int{1, 2, 3, 4, 5}
	sub := orig[1:3]
	sub[0] = 99
	return orig
}

// SmallCopy เก็บแค่ 2 ตัวแรกโดยไม่ถือ array ใหญ่ไว้ (กัน memory leak)
func SmallCopy(big []int) []int {
	out := make([]int, 2)
	copy(out, big)
	return out
}

// ClosureVars ใช้ดูพฤติกรรม loop variable (Go 1.22+ แยกตัวแปรต่อรอบ)
func ClosureVars() []int {
	var fs []func() int
	for i := 0; i < 3; i++ {
		fs = append(fs, func() int { return i })
	}
	out := make([]int, 0, 3)
	for _, f := range fs {
		out = append(out, f())
	}
	return out
}
