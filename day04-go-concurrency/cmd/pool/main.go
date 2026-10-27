// CLI สาธิต worker pool: pool -workers 4 -jobs 20 -delay 200ms
// กด Ctrl+C เพื่อดู graceful shutdown (งานที่กำลังทำเสร็จก่อน แล้วสรุปผล)
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"

	"github.com/sitthichok-J/self-study-2026/day04-go-concurrency/internal/pool"
)

func main() {
	workers := flag.Int("workers", 4, "จำนวน worker")
	n := flag.Int("jobs", 20, "จำนวนงาน")
	delay := flag.Duration("delay", 200*time.Millisecond, "เวลาที่แต่ละงานใช้")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	done, total := run(ctx, os.Stdout, *workers, *n, *delay)
	fmt.Printf("สรุป: ทำเสร็จ %d จาก %d งาน\n", done, total)
}

// run แยกออกมาเพื่อให้เทสต์ได้ คืน (จำนวนที่เสร็จ, จำนวนงานทั้งหมด)
func run(ctx context.Context, w io.Writer, workers, n int, delay time.Duration) (int, int) {
	jobs := make(chan pool.Job)
	go func() {
		defer close(jobs) // ผู้ส่งเป็นคนปิด
		for i := 1; i <= n; i++ {
			select {
			case jobs <- pool.Job{ID: i, Input: i}:
			case <-ctx.Done():
				return
			}
		}
	}()

	square := func(ctx context.Context, j pool.Job) (int, error) {
		select {
		case <-time.After(delay):
			return j.Input * j.Input, nil
		case <-ctx.Done():
			return 0, ctx.Err()
		}
	}

	var results []pool.Result
	for r := range pool.Run(ctx, jobs, workers, square) {
		results = append(results, r)
	}
	sort.Slice(results, func(i, j int) bool { return results[i].JobID < results[j].JobID })
	ok := 0
	for _, r := range results {
		if r.Err != nil {
			fmt.Fprintf(w, "job %d: %v\n", r.JobID, r.Err)
			continue
		}
		ok++
		fmt.Fprintf(w, "job %d: %d\n", r.JobID, r.Output)
	}
	return ok, n
}
