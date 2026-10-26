package store

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/sitthichok-J/self-study-2026/day03-go-fundamentals/internal/tasks"
)

// ทั้งสอง implementation ต้องผ่านชุดเทสต์เดียวกัน (พิสูจน์ว่าสลับกันได้)
func implementations(t *testing.T) map[string]tasks.Storage {
	return map[string]tasks.Storage{
		"memory": &Memory{},
		"file":   File{Path: filepath.Join(t.TempDir(), "t.json")},
	}
}

func TestRoundTrip(t *testing.T) {
	want := []tasks.Task{{ID: 1, Title: "a"}, {ID: 2, Title: "ไทย", Done: true}}
	for name, st := range implementations(t) {
		t.Run(name, func(t *testing.T) {
			if got, err := st.Load(); err != nil || len(got) != 0 {
				t.Fatalf("เริ่มต้นต้องว่าง: %v %v", got, err)
			}
			if err := st.Save(want); err != nil {
				t.Fatal(err)
			}
			got, err := st.Load()
			if err != nil || !reflect.DeepEqual(got, want) {
				t.Fatalf("got %v err %v", got, err)
			}
		})
	}
}

func TestMemoryReturnsCopy(t *testing.T) {
	m := &Memory{}
	m.Save([]tasks.Task{{ID: 1, Title: "a"}})
	got, _ := m.Load()
	got[0].Title = "changed"
	again, _ := m.Load()
	if again[0].Title != "a" {
		t.Fatal("Load ต้องคืนสำเนา")
	}
}

func TestFileCorruptAndMissing(t *testing.T) {
	p := filepath.Join(t.TempDir(), "bad.json")
	os.WriteFile(p, []byte("{not json"), 0o644)
	if _, err := (File{Path: p}).Load(); err == nil {
		t.Fatal("ต้อง error")
	}
	if _, err := (File{Path: filepath.Join(t.TempDir(), "none.json")}).Load(); err != nil {
		t.Fatalf("ไฟล์ไม่มี = ว่าง: %v", err)
	}
	if err := (File{Path: filepath.Join(t.TempDir(), "no/dir/x.json")}).Save(nil); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("เขียนลง dir ที่ไม่มี: %v", err)
	}
}
