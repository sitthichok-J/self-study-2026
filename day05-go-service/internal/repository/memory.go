// Package repository: implementation ของ service.Repository (in-memory และ Postgres)
package repository

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/sitthichok-J/self-study-2026/day05-go-service/internal/model"
)

// Memory เก็บใน RAM ใช้ตอนเทสต์และตอนรันโดยไม่มี DB
type Memory struct {
	mu     sync.RWMutex
	nextID int64
	data   map[int64]model.Claim
}

func NewMemory() *Memory { return &Memory{nextID: 1, data: make(map[int64]model.Claim)} }

func (m *Memory) Create(_ context.Context, c model.Claim) (model.Claim, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c.ID = m.nextID
	m.nextID++
	m.data[c.ID] = c
	return c, nil
}

func (m *Memory) Get(_ context.Context, id int64) (model.Claim, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c, ok := m.data[id]
	if !ok {
		return model.Claim{}, model.ErrNotFound
	}
	return c, nil
}

func (m *Memory) List(_ context.Context, f model.ListFilter) ([]model.Claim, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	all := make([]model.Claim, 0, len(m.data))
	for _, c := range m.data {
		if f.Status == "" || c.Status == f.Status {
			all = append(all, c)
		}
	}
	sort.Slice(all, func(i, j int) bool { return all[i].ID < all[j].ID })
	if f.Offset >= len(all) {
		return []model.Claim{}, nil
	}
	all = all[f.Offset:]
	if f.Limit > 0 && f.Limit < len(all) {
		all = all[:f.Limit]
	}
	return all, nil
}

func (m *Memory) Update(_ context.Context, c model.Claim) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.data[c.ID]; !ok {
		return model.ErrNotFound
	}
	m.data[c.ID] = c
	return nil
}

func (m *Memory) Delete(_ context.Context, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.data[id]; !ok {
		return model.ErrNotFound
	}
	delete(m.data, id)
	return nil
}

// UpdateStatusMany: ตรวจว่ามีครบทุก id ก่อน แล้วค่อยแก้ทั้งหมด (all-or-nothing)
func (m *Memory) UpdateStatusMany(_ context.Context, ids []int64, s model.Status, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, id := range ids {
		if _, ok := m.data[id]; !ok {
			return fmt.Errorf("claim %d: %w", id, model.ErrNotFound)
		}
	}
	for _, id := range ids {
		c := m.data[id]
		c.Status, c.UpdatedAt = s, at
		m.data[id] = c
	}
	return nil
}
