package strcat

import "testing"

func TestAllThreeAgree(t *testing.T) {
	for _, n := range []int{0, 1, 2, 10, 1000} {
		p, b, f := Plus(n), Builder(n), Sprintf(n)
		if p != b || b != f {
			t.Fatalf("n=%d ไม่ตรงกัน len=%d/%d/%d", n, len(p), len(b), len(f))
		}
	}
	if got := Builder(3); got != "0,1,2," {
		t.Fatalf("got %q", got)
	}
}

var sink string

func benchN(b *testing.B, f func(int) string, n int) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sink = f(n)
	}
}

func BenchmarkPlus_1000(b *testing.B)    { benchN(b, Plus, 1000) }
func BenchmarkBuilder_1000(b *testing.B) { benchN(b, Builder, 1000) }
func BenchmarkSprintf_1000(b *testing.B) { benchN(b, Sprintf, 1000) }
