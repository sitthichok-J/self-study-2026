package demo

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestAppendSharing(t *testing.T) {
	shared, b, c := AppendSharing()
	if !shared || b != 200 || c != 200 {
		t.Fatalf("cap เหลือ: ต้องแชร์และถูกทับ, got shared=%v b=%d c=%d", shared, b, c)
	}
	shared, b, c = AppendRealloc()
	if shared || b != 100 || c != 200 {
		t.Fatalf("cap เต็ม: ต้องแยกกัน, got shared=%v b=%d c=%d", shared, b, c)
	}
}

func TestSubSliceShares(t *testing.T) {
	if got := SubSliceShares(); !reflect.DeepEqual(got, []int{1, 99, 3, 4, 5}) {
		t.Fatalf("got %v", got)
	}
}

func TestSmallCopyIndependent(t *testing.T) {
	big := []int{1, 2, 3, 4}
	s := SmallCopy(big)
	s[0] = 42
	if big[0] != 1 || cap(s) != 2 {
		t.Fatalf("ต้องไม่แชร์ array: big=%v cap=%d", big, cap(s))
	}
}

func TestClosureVarsGo122(t *testing.T) {
	if got := ClosureVars(); !reflect.DeepEqual(got, []int{0, 1, 2}) {
		t.Fatalf("go.mod เป็น 1.22 ต้องได้ 0 1 2 แต่ได้ %v", got)
	}
}

func TestWriteNilMapPanics(t *testing.T) {
	err := WriteNilMap()
	if err == nil || !strings.Contains(err.Error(), "nil map") {
		t.Fatalf("ต้อง panic เรื่อง nil map, got %v", err)
	}
}

func TestCounter(t *testing.T) {
	var c Counter // zero value ใช้ได้
	if c.Get("a") != 0 {
		t.Fatal("อ่าน nil map ต้องได้ 0")
	}
	c.Inc("a")
	c.Inc("a")
	if c.Get("a") != 2 {
		t.Fatalf("got %d", c.Get("a"))
	}
}

func TestNilInterface(t *testing.T) {
	tests := []struct {
		name    string
		fn      func() error
		wantNil bool
	}{
		{"buggy: pointer nil ห่อใน interface", BuggyFail, false},
		{"fixed: nil ตรงๆ", FixedFail, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsNilInterface(tc.fn()); got != tc.wantNil {
				t.Fatalf("got %v want %v", got, tc.wantNil)
			}
		})
	}
}

func TestDescribe(t *testing.T) {
	tests := []struct {
		in   any
		want string
	}{
		{nil, "nil"}, {1, "int"}, {"x", "string:x"},
		{errors.New("boom"), "error:boom"}, {1.5, "other"},
	}
	for _, tc := range tests {
		if got := Describe(tc.in); got != tc.want {
			t.Errorf("Describe(%v)=%q want %q", tc.in, got, tc.want)
		}
	}
}

func TestDeferOrder(t *testing.T) {
	want := []string{"closure 2", "closure 1", "closure 0", "arg 2", "arg 1", "arg 0"}
	if got := DeferOrder(); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v", got)
	}
}

func TestDeferInLoopVsPerIteration(t *testing.T) {
	inLoop := []string{"open 0", "open 1", "open 2", "end of func", "close 2", "close 1", "close 0"}
	if got := DeferInLoop(); !reflect.DeepEqual(got, inLoop) {
		t.Fatalf("in loop: %v", got)
	}
	per := []string{"open 0", "close 0", "open 1", "close 1", "open 2", "close 2"}
	if got := DeferPerIteration(); !reflect.DeepEqual(got, per) {
		t.Fatalf("per iteration: %v", got)
	}
}
