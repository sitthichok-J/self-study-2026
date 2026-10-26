package demo

import "fmt"

// WriteNilMap เขียนลง nil map -> panic แล้วจับมาเป็น error เพื่อทดสอบ
func WriteNilMap() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered: %v", r)
		}
	}()
	var m map[string]int
	m["a"] = 1
	return nil
}

// Counter แก้แล้ว: สร้าง map ก่อนเขียนเสมอ
type Counter struct{ m map[string]int }

func (c *Counter) Inc(k string) {
	if c.m == nil {
		c.m = make(map[string]int)
	}
	c.m[k]++
}

func (c *Counter) Get(k string) int { return c.m[k] } // อ่าน nil map ได้ ได้ 0
