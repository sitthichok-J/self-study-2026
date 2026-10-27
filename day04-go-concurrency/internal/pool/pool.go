// Package pool: worker pool ที่รับงานจาก channel, รัน N worker, cancel ได้ด้วย context
// และ graceful shutdown (งานที่กำลังทำอยู่ทำให้เสร็จ งานที่ยังไม่เริ่มไม่ถูกหยิบ)
package pool

import (
	"context"
	"sync"
)

// Job คืองานหนึ่งชิ้น ID ใช้จับคู่กับผล
type Job struct {
	ID    int
	Input int
}

// Result คือผลของงานหนึ่งชิ้น
type Result struct {
	JobID  int
	Output int
	Err    error
}

// Func คือสิ่งที่ worker ทำกับงาน ควรเคารพ ctx ถ้าใช้เวลานาน
type Func func(ctx context.Context, j Job) (int, error)

// Run เริ่ม workers ตัว อ่านจาก jobs และส่งผลออกทาง channel ที่คืนมา
//
// กติกา:
//   - ช่องผลลัพธ์ถูกปิดเมื่อ worker ทุกตัวจบ (ผู้เรียกวน range จนปิดได้เลย)
//   - ผู้เรียก "ต้องอ่านผลให้หมด" ไม่งั้น worker จะค้างตอนส่งผล
//   - เมื่อ ctx ถูก cancel: worker จะไม่หยิบงานใหม่ แต่งานที่กำลังทำจะส่งผลให้ครบ (graceful)
//   - ผู้ส่ง jobs เป็นคนปิด jobs เมื่อไม่มีงานแล้ว
func Run(ctx context.Context, jobs <-chan Job, workers int, fn Func) <-chan Result {
	if workers < 1 {
		workers = 1
	}
	out := make(chan Result)
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			worker(ctx, jobs, out, fn)
		}()
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func worker(ctx context.Context, jobs <-chan Job, out chan<- Result, fn Func) {
	for {
		// เช็ก cancel ก่อนหยิบงานใหม่ (select ที่มีหลายกิ่งพร้อมจะสุ่ม จึงเช็กแยกก่อน)
		if ctx.Err() != nil {
			return
		}
		select {
		case <-ctx.Done():
			return
		case j, ok := <-jobs:
			if !ok {
				return
			}
			v, err := fn(ctx, j)
			out <- Result{JobID: j.ID, Output: v, Err: err} // ส่งผลให้ครบแม้ ctx ถูก cancel ไปแล้ว
		}
	}
}
