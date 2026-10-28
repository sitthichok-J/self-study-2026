// Package strcat: แบบฝึกหัดข้อ 2 — ต่อ string 3 แบบ (ผลต้องเหมือนกัน ต่างกันที่ประสิทธิภาพ)
package strcat

import (
	"fmt"
	"strconv"
	"strings"
)

// Plus ใช้ += ทุกรอบ: ได้ string ใหม่ + copy ของเดิมทั้งก้อนทุกรอบ = O(n²)
func Plus(n int) string {
	s := ""
	for i := 0; i < n; i++ {
		s += strconv.Itoa(i) + ","
	}
	return s
}

// Builder ใช้ strings.Builder: เขียนต่อท้าย buffer เดียว (Grow ล่วงหน้าได้)
func Builder(n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteString(strconv.Itoa(i))
		b.WriteByte(',')
	}
	return b.String()
}

// Sprintf สะดวกแต่แพง: parse format string + boxing อาร์กิวเมนต์เป็น interface ทุกรอบ
func Sprintf(n int) string {
	s := ""
	for i := 0; i < n; i++ {
		s = fmt.Sprintf("%s%d,", s, i)
	}
	return s
}
