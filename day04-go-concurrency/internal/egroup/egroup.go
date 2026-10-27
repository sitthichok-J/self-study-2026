// Package egroup: errgroup ฉบับเล็กที่เขียนเองเพื่อเรียนรู้พฤติกรรม
// (ใน sandbox ที่ทำงานดึง golang.org/x/sync ไม่ได้) API ตั้งใจให้เหมือน golang.org/x/sync/errgroup
// โปรเจกต์จริงให้ใช้ของ x/sync แทน เปลี่ยนแค่ import
//
// พฤติกรรม: Go รันฟังก์ชันใน goroutine, error แรกที่เกิดถูกเก็บและ cancel ctx ที่เหลือ,
// SetLimit จำกัดจำนวนที่รันพร้อมกัน (Go จะ block จนมีที่ว่าง), Wait รอทุกตัวแล้วคืน error แรก
package egroup

import (
	"context"
	"sync"
)

type Group struct {
	cancel context.CancelCauseFunc
	wg     sync.WaitGroup
	sem    chan struct{}
	once   sync.Once
	err    error
}

// WithContext คืน Group และ ctx ที่จะถูก cancel เมื่อมีงานใด error หรือเมื่อ Wait จบ
func WithContext(ctx context.Context) (*Group, context.Context) {
	ctx, cancel := context.WithCancelCause(ctx)
	return &Group{cancel: cancel}, ctx
}

// SetLimit จำกัดจำนวน goroutine ที่ทำงานพร้อมกัน ต้องเรียกก่อน Go (n<0 = ไม่จำกัด)
func (g *Group) SetLimit(n int) {
	if n < 0 {
		g.sem = nil
		return
	}
	g.sem = make(chan struct{}, n)
}

// Go รัน f ใน goroutine ใหม่ ถ้าตั้ง limit และเต็ม จะ block จนมีที่ว่าง
func (g *Group) Go(f func() error) {
	if g.sem != nil {
		g.sem <- struct{}{}
	}
	g.wg.Add(1)
	go func() {
		defer func() {
			if g.sem != nil {
				<-g.sem
			}
			g.wg.Done()
		}()
		if err := f(); err != nil {
			g.once.Do(func() {
				g.err = err
				if g.cancel != nil {
					g.cancel(err)
				}
			})
		}
	}()
}

// Wait รอทุกงาน แล้วคืน error แรก (ถ้ามี)
func (g *Group) Wait() error {
	g.wg.Wait()
	if g.cancel != nil {
		g.cancel(g.err)
	}
	return g.err
}
