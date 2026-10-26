package day03

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/sitthichok-J/self-study-2026/day03-go-fundamentals/internal/demo"
)

// Baseline: เก็บ "พฤติกรรมปัจจุบัน" ของตัวอย่างทั้งหมดเป็นไฟล์ golden
// ถ้าพฤติกรรมเปลี่ยนโดยไม่ตั้งใจ เทสต์นี้จะแดง
// เปลี่ยนโดยตั้งใจ: go test -run Baseline -update แล้ว review diff ของ testdata/baseline.golden
var update = flag.Bool("update", false, "เขียน golden ใหม่")

func snapshot() string {
	var b strings.Builder
	w := func(k string, v any) { fmt.Fprintf(&b, "%s = %v\n", k, v) }
	s, x, y := demo.AppendSharing()
	w("AppendSharing", []any{s, x, y})
	s, x, y = demo.AppendRealloc()
	w("AppendRealloc", []any{s, x, y})
	w("SubSliceShares", demo.SubSliceShares())
	w("ClosureVars", demo.ClosureVars())
	w("WriteNilMap", demo.WriteNilMap())
	w("BuggyFail is nil", demo.IsNilInterface(demo.BuggyFail()))
	w("FixedFail is nil", demo.IsNilInterface(demo.FixedFail()))
	w("DeferOrder", demo.DeferOrder())
	w("DeferInLoop", demo.DeferInLoop())
	w("DeferPerIteration", demo.DeferPerIteration())
	return b.String()
}

func TestBaseline(t *testing.T) {
	const path = "testdata/baseline.golden"
	got := snapshot()
	if *update {
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ไม่มี golden (รัน -update ก่อน): %v", err)
	}
	if got != string(want) {
		t.Fatalf("baseline เปลี่ยน\n--- want\n%s--- got\n%s", want, got)
	}
}
