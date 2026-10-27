package egroup

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestWaitReturnsNilWhenAllOK(t *testing.T) {
	g, _ := WithContext(context.Background())
	var n atomic.Int64
	for i := 0; i < 10; i++ {
		g.Go(func() error { n.Add(1); return nil })
	}
	if err := g.Wait(); err != nil || n.Load() != 10 {
		t.Fatalf("err=%v n=%d", err, n.Load())
	}
}

func TestFirstErrorWinsAndCancelsContext(t *testing.T) {
	g, ctx := WithContext(context.Background())
	first := errors.New("first")
	g.Go(func() error { return first })
	g.Go(func() error {
		<-ctx.Done() // ต้องถูก cancel จาก error ของอีกตัว
		return errors.New("second (หลัง cancel)")
	})
	if err := g.Wait(); !errors.Is(err, first) {
		t.Fatalf("err=%v", err)
	}
	if !errors.Is(context.Cause(ctx), first) {
		t.Fatalf("cause=%v", context.Cause(ctx))
	}
}

func TestWaitCancelsContextEvenOnSuccess(t *testing.T) {
	g, ctx := WithContext(context.Background())
	g.Go(func() error { return nil })
	g.Wait()
	if ctx.Err() == nil {
		t.Fatal("ctx ต้องถูก cancel หลัง Wait")
	}
}

func TestSetLimit(t *testing.T) {
	for _, limit := range []int{1, 2, 5} {
		g, _ := WithContext(context.Background())
		g.SetLimit(limit)
		var cur, peak atomic.Int64
		for i := 0; i < 20; i++ {
			g.Go(func() error {
				n := cur.Add(1)
				for {
					p := peak.Load()
					if n <= p || peak.CompareAndSwap(p, n) {
						break
					}
				}
				time.Sleep(3 * time.Millisecond)
				cur.Add(-1)
				return nil
			})
		}
		g.Wait()
		if p := peak.Load(); p > int64(limit) {
			t.Fatalf("limit=%d peak=%d", limit, p)
		}
	}
}

func TestNegativeLimitMeansUnlimited(t *testing.T) {
	g, _ := WithContext(context.Background())
	g.SetLimit(-1)
	for i := 0; i < 5; i++ {
		g.Go(func() error { return nil })
	}
	if err := g.Wait(); err != nil {
		t.Fatal(err)
	}
}
