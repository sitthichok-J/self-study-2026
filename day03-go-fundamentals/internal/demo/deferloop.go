package demo

import "fmt"

// DeferOrder: defer ทำงาน LIFO และประเมินอาร์กิวเมนต์ตอนประกาศ
func DeferOrder() []string {
	var log []string
	func() {
		for i := 0; i < 3; i++ {
			defer func() { log = append(log, fmt.Sprintf("closure %d", i)) }()
		}
	}()
	func() {
		for i := 0; i < 3; i++ {
			defer func(n int) { log = append(log, fmt.Sprintf("arg %d", n)) }(i)
		}
	}()
	return log
}

// DeferInLoop: defer ในลูปทำงานตอนจบฟังก์ชัน ไม่ใช่จบรอบ
func DeferInLoop() []string {
	var log []string
	func() {
		for i := 0; i < 3; i++ {
			log = append(log, fmt.Sprintf("open %d", i))
			defer func() { log = append(log, fmt.Sprintf("close %d", i)) }()
		}
		log = append(log, "end of func")
	}()
	return log
}

// DeferPerIteration: แยกเป็นฟังก์ชันย่อย -> ปิดทุกรอบ
func DeferPerIteration() []string {
	var log []string
	for i := 0; i < 3; i++ {
		func() {
			log = append(log, fmt.Sprintf("open %d", i))
			defer func() { log = append(log, fmt.Sprintf("close %d", i)) }()
		}()
	}
	return log
}
