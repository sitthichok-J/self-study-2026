// Package leak: แบบฝึกหัดข้อ 2 — goroutine leak และวิธีแก้ด้วย context
package leak

import "context"

// Leaky เริ่ม goroutine ที่รอรับค่าจาก channel ที่ไม่มีใครส่ง/ปิด
// ผู้เรียกไม่มีทางหยุดมันได้ = leak (goroutine ค้างตลอดอายุโปรเซส)
func Leaky() {
	ch := make(chan int)
	go func() {
		<-ch // block ตลอดกาล
	}()
}

// Fixed เหมือน Leaky แต่ฟัง ctx.Done() ด้วย cancel แล้ว goroutine จบ
// คืน channel ที่ปิดเมื่อ goroutine จบ ไว้ให้เทสต์รอ
func Fixed(ctx context.Context) <-chan struct{} {
	ch := make(chan int)
	done := make(chan struct{})
	go func() {
		defer close(done)
		select {
		case <-ch:
		case <-ctx.Done():
		}
	}()
	return done
}
