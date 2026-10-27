//go:build racedemo

package race

import "testing"

func TestUnsafeCounter(t *testing.T) {
	// ใต้ -race เทสต์นี้จะ FAIL พร้อมรายงาน DATA RACE (เจตนา)
	// ถ้ารันโดยไม่มี -race จะได้ค่าน้อยกว่า 100000 เป็นบางครั้ง
	got := Hammer(&UnsafeCounter{}, 100, 1000)
	t.Logf("unsafe total = %d (ควรเป็น 100000)", got)
}
