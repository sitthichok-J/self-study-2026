package pipeline

import (
	"context"
	"reflect"
	"sort"
	"testing"
	"time"
)

func double(n int) int { return n * 2 }

func TestGenerateMapCollect(t *testing.T) {
	ctx := context.Background()
	got := Collect(Map(ctx, Generate(ctx, 1, 2, 3), double))
	if want := []int{2, 4, 6}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestEmptyInput(t *testing.T) {
	ctx := context.Background()
	if got := Collect(Map(ctx, Generate(ctx), double)); len(got) != 0 {
		t.Fatalf("got %v", got)
	}
}

func TestMergeFanIn(t *testing.T) {
	ctx := context.Background()
	got := Collect(Merge(ctx, Generate(ctx, 1, 2), Generate(ctx, 3, 4), Generate(ctx, 5)))
	sort.Ints(got) // ลำดับของ fan-in ไม่แน่นอน
	if want := []int{1, 2, 3, 4, 5}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

// drainClosed คืน true ถ้า channel ถูกปิดภายใน timeout (หมายถึงทุก stage จบแล้ว)
func drainClosed(ch <-chan int, timeout time.Duration) bool {
	deadline := time.After(timeout)
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				return true
			}
		case <-deadline:
			return false
		}
	}
}

func TestCancelStopsPipeline(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	nums := make([]int, 100000)
	out := Map(ctx, Generate(ctx, nums...), double)

	<-out // รับแค่ตัวเดียว
	cancel()
	// หลัง cancel ต้องปิดเร็วๆ ไม่ใช่รอให้ 100000 ตัวหมด
	if !drainClosed(out, time.Second) {
		t.Fatal("pipeline ไม่หยุดหลัง cancel")
	}
}

func TestCancelledBeforeStartYieldsLittle(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	nums := make([]int, 1000)
	got := Collect(Merge(ctx, Generate(ctx, nums...)))
	if len(got) >= len(nums) {
		t.Fatalf("ควรได้น้อยกว่า %d, ได้ %d", len(nums), len(got))
	}
}
