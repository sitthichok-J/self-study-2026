package day04

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/sitthichok-J/self-study-2026/day04-go-concurrency/internal/pipeline"
	"github.com/sitthichok-J/self-study-2026/day04-go-concurrency/internal/pool"
	"github.com/sitthichok-J/self-study-2026/day04-go-concurrency/internal/race"
)

// Baseline: เก็บผลที่ "ต้องแน่นอน" ของแต่ละชิ้นไว้เป็นไฟล์ golden (ไม่เก็บอะไรที่ขึ้นกับจังหวะเวลา)
// เปลี่ยนโดยตั้งใจ: go test -run Baseline -update แล้ว review diff ของ testdata/baseline.golden
var update = flag.Bool("update", false, "เขียน golden ใหม่")

func snapshot() string {
	ctx := context.Background()
	var b strings.Builder
	w := func(k string, v any) { fmt.Fprintf(&b, "%s = %v\n", k, v) }

	w("mutex counter 50x200", race.Hammer(&race.MutexCounter{}, 50, 200))
	w("atomic counter 50x200", race.Hammer(&race.AtomicCounter{}, 50, 200))
	w("pipeline double", pipeline.Collect(pipeline.Map(ctx, pipeline.Generate(ctx, 1, 2, 3, 4), func(n int) int { return n * 2 })))

	merged := pipeline.Collect(pipeline.Merge(ctx, pipeline.Generate(ctx, 1, 2), pipeline.Generate(ctx, 3)))
	sort.Ints(merged)
	w("fan-in sorted", merged)

	jobs := make(chan pool.Job)
	go func() {
		defer close(jobs)
		for i := 1; i <= 5; i++ {
			jobs <- pool.Job{ID: i, Input: i}
		}
	}()
	var rs []pool.Result
	for r := range pool.Run(ctx, jobs, 3, func(_ context.Context, j pool.Job) (int, error) { return j.Input * 10, nil }) {
		rs = append(rs, r)
	}
	sort.Slice(rs, func(i, j int) bool { return rs[i].JobID < rs[j].JobID })
	for _, r := range rs {
		w(fmt.Sprintf("pool job %d", r.JobID), r.Output)
	}
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
