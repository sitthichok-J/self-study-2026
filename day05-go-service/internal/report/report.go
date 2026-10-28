// Package report: รายงานยอดรวมต่อโรงพยาบาล มี 2 เวอร์ชัน
//   - BuildSlow: ช้าโดยจงใจ (แบบฝึกหัด pprof ข้อ 4) O(n*h) + ต่อ string ด้วย += + bubble sort
//   - Build: เร็ว ใช้ map + strings.Builder + sort ของ stdlib
//
// ทั้งสองต้องให้ผลลัพธ์เหมือนกันทุกตัวอักษร (มีเทสต์เทียบ)
package report

import (
	"fmt"
	"sort"
	"strings"

	"github.com/sitthichok-J/self-study-2026/day05-go-service/internal/model"
)

func line(h string, count int, total int64) string {
	return fmt.Sprintf("%s\t%d\t%d\n", h, count, total)
}

func BuildSlow(claims []model.Claim) string {
	var hospitals []string
	for _, c := range claims {
		found := false
		for _, h := range hospitals { // ค้นเชิงเส้นทุกครั้ง
			if h == c.Hospital {
				found = true
				break
			}
		}
		if !found {
			hospitals = append(hospitals, c.Hospital)
		}
	}
	for i := 0; i < len(hospitals); i++ { // bubble sort
		for j := 0; j < len(hospitals)-1-i; j++ {
			if hospitals[j] > hospitals[j+1] {
				hospitals[j], hospitals[j+1] = hospitals[j+1], hospitals[j]
			}
		}
	}
	out := ""
	for _, h := range hospitals {
		count, total := 0, int64(0)
		for _, c := range claims { // วนทุก claim ต่อโรงพยาบาลหนึ่งแห่ง
			if c.Hospital == h {
				count++
				total += c.AmountSatang
			}
		}
		out += line(h, count, total) // ต่อ string ทีละบรรทัด (copy ทั้งก้อนทุกครั้ง)
	}
	return out
}

func Build(claims []model.Claim) string {
	type agg struct {
		count int
		total int64
	}
	byHosp := make(map[string]*agg)
	for _, c := range claims {
		a := byHosp[c.Hospital]
		if a == nil {
			a = &agg{}
			byHosp[c.Hospital] = a
		}
		a.count++
		a.total += c.AmountSatang
	}
	names := make([]string, 0, len(byHosp))
	for h := range byHosp {
		names = append(names, h)
	}
	sort.Strings(names)
	var b strings.Builder
	for _, h := range names {
		a := byHosp[h]
		b.WriteString(line(h, a.count, a.total))
	}
	return b.String()
}
