// Package tasks: logic ของรายการงาน ผูกกับ interface Storage ที่ประกาศฝั่งผู้ใช้ (consumer)
package tasks

import (
	"errors"
	"fmt"
	"strings"
)

type Task struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

// Storage เล็กๆ เท่าที่ package นี้ใช้ ("accept interfaces")
type Storage interface {
	Load() ([]Task, error)
	Save([]Task) error
}

var ErrEmptyTitle = errors.New("title is empty")

// NotFoundError คือ custom error ที่มีข้อมูลประกอบ
type NotFoundError struct{ ID int }

func (e *NotFoundError) Error() string { return fmt.Sprintf("task %d not found", e.ID) }

type Service struct{ st Storage }

// New คืน struct (ไม่ใช่ interface) ตามหลัก "return structs"
func New(st Storage) *Service { return &Service{st: st} }

func (s *Service) Add(title string) (Task, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return Task{}, ErrEmptyTitle
	}
	list, err := s.st.Load()
	if err != nil {
		return Task{}, fmt.Errorf("add: %w", err)
	}
	id := 1
	for _, t := range list {
		if t.ID >= id {
			id = t.ID + 1
		}
	}
	t := Task{ID: id, Title: title}
	if err := s.st.Save(append(list, t)); err != nil {
		return Task{}, fmt.Errorf("add: %w", err)
	}
	return t, nil
}

func (s *Service) List() ([]Task, error) {
	list, err := s.st.Load()
	if err != nil {
		return nil, fmt.Errorf("list: %w", err)
	}
	return list, nil
}

func (s *Service) find(list []Task, id int) (int, error) {
	for i, t := range list {
		if t.ID == id {
			return i, nil
		}
	}
	return -1, &NotFoundError{ID: id}
}

func (s *Service) Done(id int) error {
	list, err := s.st.Load()
	if err != nil {
		return fmt.Errorf("done: %w", err)
	}
	i, err := s.find(list, id)
	if err != nil {
		return fmt.Errorf("done: %w", err)
	}
	list[i].Done = true
	if err := s.st.Save(list); err != nil {
		return fmt.Errorf("done: %w", err)
	}
	return nil
}

func (s *Service) Remove(id int) error {
	list, err := s.st.Load()
	if err != nil {
		return fmt.Errorf("remove: %w", err)
	}
	i, err := s.find(list, id)
	if err != nil {
		return fmt.Errorf("remove: %w", err)
	}
	list = append(list[:i], list[i+1:]...)
	if err := s.st.Save(list); err != nil {
		return fmt.Errorf("remove: %w", err)
	}
	return nil
}
