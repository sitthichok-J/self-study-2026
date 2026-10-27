// Package race: แบบฝึกหัดข้อ 1 — counter ที่ปลอดภัยด้วย mutex และ atomic
// (แบบที่ไม่ปลอดภัยอยู่ใน unsafe_demo.go ใต้ build tag racedemo เพื่อไม่ให้ go test -race ปกติแดง)
package race

import (
	"sync"
	"sync/atomic"
)

// Counter คือสิ่งที่ทุก implementation ต้องทำได้
type Counter interface {
	Inc()
	Value() int64
}

// MutexCounter ปกป้อง state ด้วย mutex (เหมาะเมื่อมีหลายฟิลด์ที่ต้องเปลี่ยนพร้อมกัน)
type MutexCounter struct {
	mu sync.Mutex
	n  int64
}

func (c *MutexCounter) Inc() {
	c.mu.Lock()
	c.n++
	c.mu.Unlock()
}

func (c *MutexCounter) Value() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.n
}

// AtomicCounter ใช้ atomic (เร็วกว่า เหมาะกับตัวเลขตัวเดียว)
type AtomicCounter struct{ n atomic.Int64 }

func (c *AtomicCounter) Inc()         { c.n.Add(1) }
func (c *AtomicCounter) Value() int64 { return c.n.Load() }

// Hammer ให้ workers goroutine เรียก Inc คนละ perWorker ครั้ง แล้วรอจนเสร็จ
func Hammer(c Counter, workers, perWorker int) int64 {
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < perWorker; j++ {
				c.Inc()
			}
		}()
	}
	wg.Wait()
	return c.Value()
}
