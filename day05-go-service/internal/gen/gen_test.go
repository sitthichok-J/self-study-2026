package gen

import (
	"reflect"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
)

func TestMap(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want []string
	}{
		{"หลายตัว", []int{1, 2, 3}, []string{"1", "2", "3"}},
		{"ว่าง", []int{}, []string{}},
		{"nil", nil, []string{}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := Map(tc.in, strconv.Itoa); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}

func TestFilter(t *testing.T) {
	even := func(n int) bool { return n%2 == 0 }
	if got := Filter([]int{1, 2, 3, 4, 5, 6}, even); !reflect.DeepEqual(got, []int{2, 4, 6}) {
		t.Fatalf("got %v", got)
	}
	if got := Filter([]int{1, 3}, even); len(got) != 0 {
		t.Fatalf("got %v", got)
	}
	if got := Filter[int](nil, even); got == nil || len(got) != 0 {
		t.Fatalf("nil input ต้องได้ slice ว่างที่ไม่ใช่ nil: %#v", got)
	}
}

func TestReduceChangesType(t *testing.T) {
	// T=string, A=int : รวมความยาว
	got := Reduce([]string{"ab", "cde", ""}, 0, func(a int, s string) int { return a + len(s) })
	if got != 5 {
		t.Fatalf("got %d", got)
	}
	if got := Reduce[int, int](nil, 7, func(a, v int) int { return a + v }); got != 7 {
		t.Fatalf("ว่างต้องได้ค่าเริ่มต้น got %d", got)
	}
}

type Satang int64

func TestSumNumberConstraint(t *testing.T) {
	if got := Sum([]int{1, 2, 3}); got != 6 {
		t.Fatalf("int %d", got)
	}
	if got := Sum([]float64{0.5, 0.25}); got != 0.75 {
		t.Fatalf("float %v", got)
	}
	if got := Sum([]Satang{100, 250}); got != 350 { // type ที่นิยามเองจาก int64 ผ่าน ~int64
		t.Fatalf("Satang %d", got)
	}
	if got := Sum[int](nil); got != 0 {
		t.Fatalf("nil %d", got)
	}
}

func TestGroupBy(t *testing.T) {
	words := []string{"apple", "avocado", "banana", "blueberry", "cherry"}
	got := GroupBy(words, func(s string) byte { return s[0] })
	want := map[byte][]string{'a': {"apple", "avocado"}, 'b': {"banana", "blueberry"}, 'c': {"cherry"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v", got)
	}
}

func TestCacheBasics(t *testing.T) {
	c := NewCache[string, int]()
	if _, ok := c.Get("a"); ok {
		t.Fatal("ต้องยังไม่มี")
	}
	c.Set("a", 1)
	if v, ok := c.Get("a"); !ok || v != 1 {
		t.Fatalf("got %d %v", v, ok)
	}
	calls := 0
	for i := 0; i < 3; i++ {
		if v := c.GetOrSet("b", func() int { calls++; return 42 }); v != 42 {
			t.Fatalf("got %d", v)
		}
	}
	if calls != 1 || c.Len() != 2 {
		t.Fatalf("calls=%d len=%d", calls, c.Len())
	}
}

func TestCacheConcurrent(t *testing.T) {
	c := NewCache[int, int]()
	var wg sync.WaitGroup
	var loads atomic.Int64
	for g := 0; g < 20; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for k := 0; k < 100; k++ {
				got := c.GetOrSet(k, func() int { loads.Add(1); return k * 2 })
				if got != k*2 {
					t.Errorf("key %d got %d", k, got)
				}
			}
		}()
	}
	wg.Wait()
	if c.Len() != 100 {
		t.Fatalf("len=%d", c.Len())
	}
}
