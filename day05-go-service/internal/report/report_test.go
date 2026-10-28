package report

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/sitthichok-J/self-study-2026/day05-go-service/internal/model"
)

func sample(n, hospitals int, seed int64) []model.Claim {
	r := rand.New(rand.NewSource(seed))
	out := make([]model.Claim, n)
	for i := range out {
		out[i] = model.Claim{
			ID:           int64(i + 1),
			Hospital:     fmt.Sprintf("รพ.%04d", r.Intn(hospitals)),
			AmountSatang: int64(r.Intn(100000) + 1),
		}
	}
	return out
}

func TestSlowAndFastAgree(t *testing.T) {
	tests := []struct {
		name string
		in   []model.Claim
	}{
		{"ว่าง", nil},
		{"รายการเดียว", sample(1, 1, 1)},
		{"โรงพยาบาลเดียว", sample(50, 1, 2)},
		{"หลายแห่ง", sample(500, 40, 3)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if s, f := BuildSlow(tc.in), Build(tc.in); s != f {
				t.Fatalf("ไม่ตรงกัน\nslow=%q\nfast=%q", s, f)
			}
		})
	}
}

func TestBuildTotalsAndOrder(t *testing.T) {
	in := []model.Claim{
		{Hospital: "B", AmountSatang: 100},
		{Hospital: "A", AmountSatang: 50},
		{Hospital: "B", AmountSatang: 25},
	}
	want := "A\t1\t50\nB\t2\t125\n"
	if got := Build(in); got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

var sink string

func BenchmarkBuildSlow(b *testing.B) {
	in := sample(20000, 1000, 42)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = BuildSlow(in)
	}
}

func BenchmarkBuild(b *testing.B) {
	in := sample(20000, 1000, 42)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = Build(in)
	}
}
