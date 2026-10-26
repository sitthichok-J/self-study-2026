package tasks_test

import (
	"errors"
	"testing"

	"github.com/sitthichok-J/self-study-2026/day03-go-fundamentals/internal/store"
	"github.com/sitthichok-J/self-study-2026/day03-go-fundamentals/internal/tasks"
)

type failStore struct{ err error }

func (f failStore) Load() ([]tasks.Task, error) { return nil, f.err }
func (f failStore) Save([]tasks.Task) error     { return f.err }

func TestAdd(t *testing.T) {
	tests := []struct {
		name    string
		title   string
		wantErr error
	}{
		{"ปกติ", "เขียนเทสต์", nil},
		{"trim ช่องว่าง", "  งาน  ", nil},
		{"ว่าง", "", tasks.ErrEmptyTitle},
		{"มีแต่ช่องว่าง", "   ", tasks.ErrEmptyTitle},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := tasks.New(&store.Memory{})
			got, err := s.Add(tc.title)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("err=%v want %v", err, tc.wantErr)
			}
			if err == nil && got.ID != 1 {
				t.Fatalf("id=%d", got.ID)
			}
		})
	}
}

func TestIDsDoNotReuseAfterRemove(t *testing.T) {
	s := tasks.New(&store.Memory{})
	s.Add("a")
	s.Add("b")
	if err := s.Remove(1); err != nil {
		t.Fatal(err)
	}
	got, _ := s.Add("c")
	if got.ID != 3 {
		t.Fatalf("id=%d want 3", got.ID)
	}
}

func TestDoneAndNotFound(t *testing.T) {
	s := tasks.New(&store.Memory{})
	s.Add("a")
	if err := s.Done(1); err != nil {
		t.Fatal(err)
	}
	list, _ := s.List()
	if !list[0].Done {
		t.Fatal("ต้อง done")
	}
	err := s.Done(99)
	var nf *tasks.NotFoundError
	if !errors.As(err, &nf) || nf.ID != 99 {
		t.Fatalf("errors.As ล้มเหลว: %v", err)
	}
	if err.Error() != "done: task 99 not found" {
		t.Fatalf("ข้อความ: %q", err.Error())
	}
	if err := s.Remove(99); !errors.As(err, &nf) {
		t.Fatalf("remove ต้องเป็น NotFoundError: %v", err)
	}
}

func TestStorageErrorsAreWrapped(t *testing.T) {
	boom := errors.New("disk down")
	s := tasks.New(failStore{err: boom})
	calls := map[string]func() error{
		"add":    func() error { _, e := s.Add("x"); return e },
		"list":   func() error { _, e := s.List(); return e },
		"done":   func() error { return s.Done(1) },
		"remove": func() error { return s.Remove(1) },
	}
	for name, fn := range calls {
		t.Run(name, func(t *testing.T) {
			if err := fn(); !errors.Is(err, boom) {
				t.Fatalf("errors.Is ต้องเจอ error ต้นทาง: %v", err)
			}
		})
	}
}
