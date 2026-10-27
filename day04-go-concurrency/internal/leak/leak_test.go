package leak

import (
	"context"
	"runtime"
	"testing"
	"time"
)

// รอจน goroutine ทั้งหมดนิ่ง แล้วคืนจำนวน (วัดแบบ delta ไม่พึ่งตัวเลขสัมบูรณ์)
func settled() int {
	time.Sleep(50 * time.Millisecond)
	return runtime.NumGoroutine()
}

func TestLeakyKeepsGoroutines(t *testing.T) {
	before := settled()
	for i := 0; i < 5; i++ {
		Leaky()
	}
	after := settled()
	if after-before != 5 {
		t.Fatalf("ต้องค้าง 5 goroutine, before=%d after=%d", before, after)
	}
}

func TestFixedStopsOnCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	done := Fixed(ctx)

	select {
	case <-done:
		t.Fatal("ต้องยังไม่จบก่อน cancel")
	case <-time.After(30 * time.Millisecond):
	}

	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("cancel แล้ว goroutine ต้องจบ")
	}
}

func TestFixedWithCancelledCtxEndsImmediately(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	select {
	case <-Fixed(ctx):
	case <-time.After(time.Second):
		t.Fatal("ต้องจบทันที")
	}
}
