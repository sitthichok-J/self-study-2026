package repository

import (
	"context"
	"sync"
	"testing"

	"github.com/sitthichok-J/self-study-2026/day05-go-service/internal/model"
	"github.com/sitthichok-J/self-study-2026/day05-go-service/internal/service"
)

func TestMemoryContract(t *testing.T) {
	runContract(t, func(t *testing.T) service.Repository { return NewMemory() })
}

func TestMemoryConcurrentCreate(t *testing.T) {
	m := NewMemory()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				m.Create(context.Background(), model.Claim{PatientRef: "P"})
			}
		}()
	}
	wg.Wait()
	all, _ := m.List(context.Background(), model.ListFilter{})
	if len(all) != 1000 {
		t.Fatalf("got %d", len(all))
	}
	seen := map[int64]bool{}
	for _, c := range all {
		if seen[c.ID] {
			t.Fatalf("id ซ้ำ %d", c.ID)
		}
		seen[c.ID] = true
	}
}

func TestQueriesAreParameterized(t *testing.T) {
	for name, q := range Queries() {
		found := false
		for i := 0; i+1 < len(q); i++ {
			if q[i] == '$' && q[i+1] >= '1' && q[i+1] <= '9' {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("%s ไม่มี placeholder $n", name)
		}
	}
}
