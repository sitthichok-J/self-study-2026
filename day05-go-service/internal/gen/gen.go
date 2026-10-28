// Package gen: generics (แบบฝึกหัดข้อ 1) ใช้เมื่อโค้ดซ้ำจริงๆ ไม่ใช่เพื่อความเท่
package gen

import "sync"

func Map[T, U any](in []T, fn func(T) U) []U {
	out := make([]U, 0, len(in))
	for _, v := range in {
		out = append(out, fn(v))
	}
	return out
}

func Filter[T any](in []T, keep func(T) bool) []T {
	out := make([]T, 0)
	for _, v := range in {
		if keep(v) {
			out = append(out, v)
		}
	}
	return out
}

func Reduce[T, A any](in []T, init A, fn func(A, T) A) A {
	acc := init
	for _, v := range in {
		acc = fn(acc, v)
	}
	return acc
}

// Number ตัวเลขทุกชนิดที่บวกได้ (ใช้ ~ เพื่อรวม type ที่นิยามเองจาก int ด้วย)
type Number interface {
	~int | ~int32 | ~int64 | ~float32 | ~float64
}

func Sum[T Number](in []T) T {
	return Reduce(in, T(0), func(a, v T) T { return a + v })
}

// GroupBy จัดกลุ่มตาม key (key ต้อง comparable เพราะใช้เป็น key ของ map)
func GroupBy[T any, K comparable](in []T, key func(T) K) map[K][]T {
	out := make(map[K][]T)
	for _, v := range in {
		k := key(v)
		out[k] = append(out[k], v)
	}
	return out
}

// Cache ปลอดภัยต่อการใช้หลาย goroutine
type Cache[K comparable, V any] struct {
	mu sync.RWMutex
	m  map[K]V
}

func NewCache[K comparable, V any]() *Cache[K, V] { return &Cache[K, V]{m: make(map[K]V)} }

func (c *Cache[K, V]) Get(k K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.m[k]
	return v, ok
}

func (c *Cache[K, V]) Set(k K, v V) {
	c.mu.Lock()
	c.m[k] = v
	c.mu.Unlock()
}

// GetOrSet คืนค่าที่มีอยู่ ถ้าไม่มีเรียก load แล้วเก็บไว้ (load อาจถูกเรียกซ้ำถ้าชนกัน แต่ค่าที่เก็บจะเป็นของคนที่ชนะ)
func (c *Cache[K, V]) GetOrSet(k K, load func() V) V {
	if v, ok := c.Get(k); ok {
		return v
	}
	v := load()
	c.mu.Lock()
	defer c.mu.Unlock()
	if existing, ok := c.m[k]; ok {
		return existing
	}
	c.m[k] = v
	return v
}

func (c *Cache[K, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.m)
}
