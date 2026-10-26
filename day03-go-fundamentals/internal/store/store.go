// Package store: implementation ของ tasks.Storage (in-memory และไฟล์ JSON)
package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/sitthichok-J/self-study-2026/day03-go-fundamentals/internal/tasks"
)

// Memory เก็บใน RAM ใช้ใน test
type Memory struct{ data []tasks.Task }

func (m *Memory) Load() ([]tasks.Task, error) {
	out := make([]tasks.Task, len(m.data)) // คืนสำเนา กันผู้เรียกแก้ข้อมูลข้างใน
	copy(out, m.data)
	return out, nil
}

func (m *Memory) Save(t []tasks.Task) error {
	m.data = make([]tasks.Task, len(t))
	copy(m.data, t)
	return nil
}

// File เก็บเป็น JSON
type File struct{ Path string }

func (f File) Load() ([]tasks.Task, error) {
	b, err := os.ReadFile(f.Path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil // ยังไม่มีไฟล์ = รายการว่าง
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", f.Path, err)
	}
	var out []tasks.Task
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, fmt.Errorf("parse %s: %w", f.Path, err)
	}
	return out, nil
}

func (f File) Save(t []tasks.Task) error {
	if t == nil {
		t = []tasks.Task{}
	}
	b, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return fmt.Errorf("encode: %w", err)
	}
	if err := os.WriteFile(f.Path, b, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", f.Path, err)
	}
	return nil
}
