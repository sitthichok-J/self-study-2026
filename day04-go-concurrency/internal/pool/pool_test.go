package pool

import (
	"context"
	"errors"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func feed(ctx context.Context, n int) <-chan Job {
	jobs := make(chan Job)
	go func() {
		defer close(jobs)
		for i := 1; i <= n; i++ {
			select {
			case jobs <- Job{ID: i, Input: i}:
			case <-ctx.Done():
				return
			}
		}
	}()
	return jobs
}

func collect(ch <-chan Result) []Result {
	var out []Result
	for r := range ch {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].JobID < out[j].JobID })
	return out
}

func square(_ context.Context, j Job) (int, error) { return j.Input * j.Input, nil }

func TestAllJobsProcessed(t *testing.T) {
	ctx := context.Background()
	for _, workers := range []int{1, 3, 10} {
		got := collect(Run(ctx, feed(ctx, 20), workers, square))
		if len(got) != 20 {
			t.Fatalf("workers=%d got %d results", workers, len(got))
		}
		for i, r := range got {
			if r.JobID != i+1 || r.Output != (i+1)*(i+1) || r.Err != nil {
				t.Fatalf("workers=%d result %d = %+v", workers, i, r)
			}
		}
	}
}

func TestZeroWorkersTreatedAsOne(t *testing.T) {
	ctx := context.Background()
	if got := collect(Run(ctx, feed(ctx, 3), 0, square)); len(got) != 3 {
		t.Fatalf("got %d", len(got))
	}
}

func TestErrorsAreReportedPerJob(t *testing.T) {
	ctx := context.Background()
	boom := errors.New("odd")
	fn := func(_ context.Context, j Job) (int, error) {
		if j.Input%2 == 1 {
			return 0, boom
		}
		return j.Input, nil
	}
	got := collect(Run(ctx, feed(ctx, 6), 3, fn))
	if len(got) != 6 {
		t.Fatalf("got %d", len(got))
	}
	for _, r := range got {
		odd := r.JobID%2 == 1
		if odd && !errors.Is(r.Err, boom) || !odd && r.Err != nil {
			t.Fatalf("job %d err=%v", r.JobID, r.Err)
		}
	}
}

func TestConcurrencyNeverExceedsWorkers(t *testing.T) {
	ctx := context.Background()
	var cur, peak atomic.Int64
	fn := func(_ context.Context, j Job) (int, error) {
		n := cur.Add(1)
		for {
			p := peak.Load()
			if n <= p || peak.CompareAndSwap(p, n) {
				break
			}
		}
		time.Sleep(5 * time.Millisecond)
		cur.Add(-1)
		return 0, nil
	}
	collect(Run(ctx, feed(ctx, 30), 4, fn))
	if p := peak.Load(); p > 4 || p < 2 {
		t.Fatalf("peak=%d ต้องอยู่ระหว่าง 2..4", p)
	}
}

// graceful shutdown: งานที่เริ่มแล้วต้องเสร็จ งานที่ยังไม่เริ่มต้องไม่ถูกหยิบ
func TestGracefulShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	const workers = 3
	gate := make(chan struct{})
	var started sync.WaitGroup
	started.Add(workers)
	var once [workers]sync.Once
	var idx atomic.Int64
	fn := func(_ context.Context, j Job) (int, error) { // จงใจไม่ดู ctx = งานที่กำลังทำจนเสร็จ
		i := idx.Add(1) - 1
		if int(i) < workers {
			once[i].Do(started.Done)
		}
		<-gate
		return j.Input, nil
	}
	out := Run(ctx, feed(ctx, 10), workers, fn)

	started.Wait() // ทั้ง 3 worker เริ่มงานแล้ว
	cancel()
	close(gate) // ปล่อยให้งานที่ค้างอยู่ทำต่อจนเสร็จ

	got := collect(out)
	if len(got) != workers {
		t.Fatalf("ต้องได้ผล %d ชิ้น (เฉพาะที่เริ่มแล้ว) ได้ %d", workers, len(got))
	}
}

func TestCancelBeforeStartProcessesNothing(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if got := collect(Run(ctx, feed(ctx, 10), 3, square)); len(got) != 0 {
		t.Fatalf("got %d", len(got))
	}
}

func TestResultsClosedWhenJobsClosed(t *testing.T) {
	jobs := make(chan Job)
	close(jobs)
	out := Run(context.Background(), jobs, 2, square)
	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("ไม่ควรมีผล")
		}
	case <-time.After(time.Second):
		t.Fatal("ช่องผลต้องปิด")
	}
}
