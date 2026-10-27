// Package pipeline: แบบฝึกหัดข้อ 4 — producer -> transformer -> consumer ที่หยุดเมื่อ ctx ถูก cancel
// กติกา: stage ที่สร้าง channel เป็นคนปิด channel นั้น (defer close) และทุกการส่งต้อง select กับ ctx.Done()
package pipeline

import (
	"context"
	"sync"
)

// Generate ส่งค่า nums ออกทีละตัว
func Generate(ctx context.Context, nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			select {
			case out <- n:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// Map รับค่าจาก in ทำ fn แล้วส่งต่อ
func Map(ctx context.Context, in <-chan int, fn func(int) int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for v := range in {
			select {
			case out <- fn(v):
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// Merge (fan-in) รวมหลาย channel เป็นช่องเดียว ปิดช่องรวมเมื่อทุกช่องต้นทางจบ
func Merge(ctx context.Context, ins ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup
	wg.Add(len(ins))
	for _, in := range ins {
		go func() {
			defer wg.Done()
			for v := range in {
				select {
				case out <- v:
				case <-ctx.Done():
					return
				}
			}
		}()
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

// Collect อ่านจนช่องปิด (consumer)
func Collect(in <-chan int) []int {
	var out []int
	for v := range in {
		out = append(out, v)
	}
	return out
}
