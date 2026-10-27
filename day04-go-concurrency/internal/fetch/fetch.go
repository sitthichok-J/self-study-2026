// Package fetch: แบบฝึกหัดข้อ 3 — ดึง URL หลายตัวพร้อมกัน จำกัด concurrency
package fetch

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/sitthichok-J/self-study-2026/day04-go-concurrency/internal/egroup"
)

// Result คือผลของ URL หนึ่งตัว (index ตรงกับตำแหน่งใน urls)
type Result struct {
	URL    string
	Status int
	Bytes  int
}

// All ดึง urls พร้อมกันไม่เกิน limit ตัว คืนผลตามลำดับเดิม
// ถ้าตัวใด error: ctx ของที่เหลือถูก cancel และคืน error แรก
func All(ctx context.Context, client *http.Client, urls []string, limit int) ([]Result, error) {
	g, ctx := egroup.WithContext(ctx)
	g.SetLimit(limit)
	results := make([]Result, len(urls)) // แต่ละ goroutine เขียน index ของตัวเอง ไม่ชนกัน ไม่ต้อง lock
	for i, u := range urls {
		g.Go(func() error {
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
			if err != nil {
				return fmt.Errorf("fetch %s: %w", u, err)
			}
			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("fetch %s: %w", u, err)
			}
			defer resp.Body.Close()
			n, err := io.Copy(io.Discard, resp.Body)
			if err != nil {
				return fmt.Errorf("read %s: %w", u, err)
			}
			results[i] = Result{URL: u, Status: resp.StatusCode, Bytes: int(n)}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return results, nil
}
