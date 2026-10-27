//go:build racedemo

package race

// UnsafeCounter มี data race ตั้งใจให้ดูรายงานของ race detector
// รัน: go test -tags racedemo -race ./internal/race -run Unsafe
type UnsafeCounter struct{ n int64 }

func (c *UnsafeCounter) Inc()         { c.n++ }
func (c *UnsafeCounter) Value() int64 { return c.n }
