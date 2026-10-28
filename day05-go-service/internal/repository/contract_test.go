package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sitthichok-J/self-study-2026/day05-go-service/internal/model"
	"github.com/sitthichok-J/self-study-2026/day05-go-service/internal/service"
)

// runContract คือชุดเทสต์ที่ "ทุก implementation ต้องผ่านเหมือนกัน"
// Memory รันเสมอ  Postgres รันใน postgres_integration_test.go (build tag integration)
func runContract(t *testing.T, newRepo func(t *testing.T) service.Repository) {
	ctx := context.Background()
	at := time.Date(2026, 10, 28, 9, 0, 0, 0, time.UTC)
	mk := func(ref, hosp string, amt int64, st model.Status) model.Claim {
		return model.Claim{PatientRef: ref, Hospital: hosp, AmountSatang: amt, Status: st, CreatedAt: at, UpdatedAt: at}
	}

	t.Run("create then get", func(t *testing.T) {
		r := newRepo(t)
		c, err := r.Create(ctx, mk("P1", "H1", 1000, model.StatusPending))
		if err != nil || c.ID == 0 {
			t.Fatalf("create: %+v %v", c, err)
		}
		got, err := r.Get(ctx, c.ID)
		if err != nil {
			t.Fatal(err)
		}
		if got.PatientRef != "P1" || got.Hospital != "H1" || got.AmountSatang != 1000 || got.Status != model.StatusPending || !got.CreatedAt.Equal(at) {
			t.Fatalf("got %+v", got)
		}
	})

	t.Run("ids increase", func(t *testing.T) {
		r := newRepo(t)
		a, _ := r.Create(ctx, mk("A", "H", 1, model.StatusPending))
		b, _ := r.Create(ctx, mk("B", "H", 1, model.StatusPending))
		if b.ID <= a.ID {
			t.Fatalf("ids %d %d", a.ID, b.ID)
		}
	})

	t.Run("get missing", func(t *testing.T) {
		r := newRepo(t)
		if _, err := r.Get(ctx, 999); !errors.Is(err, model.ErrNotFound) {
			t.Fatalf("err=%v", err)
		}
	})

	t.Run("list filter paging order", func(t *testing.T) {
		r := newRepo(t)
		for i, st := range []model.Status{model.StatusPending, model.StatusApproved, model.StatusPending, model.StatusRejected, model.StatusPending} {
			r.Create(ctx, mk("P", "H", int64(i+1), st))
		}
		all, _ := r.List(ctx, model.ListFilter{})
		if len(all) != 5 {
			t.Fatalf("all=%d", len(all))
		}
		for i := 1; i < len(all); i++ {
			if all[i-1].ID >= all[i].ID {
				t.Fatal("ต้องเรียงตาม id")
			}
		}
		pend, _ := r.List(ctx, model.ListFilter{Status: model.StatusPending})
		if len(pend) != 3 {
			t.Fatalf("pending=%d", len(pend))
		}
		page, _ := r.List(ctx, model.ListFilter{Limit: 2, Offset: 1})
		if len(page) != 2 || page[0].ID != all[1].ID {
			t.Fatalf("page=%+v", page)
		}
		beyond, err := r.List(ctx, model.ListFilter{Offset: 100})
		if err != nil || beyond == nil || len(beyond) != 0 {
			t.Fatalf("offset เกิน: %#v %v", beyond, err)
		}
	})

	t.Run("update", func(t *testing.T) {
		r := newRepo(t)
		c, _ := r.Create(ctx, mk("P", "H", 10, model.StatusPending))
		c.Hospital, c.AmountSatang, c.Status = "H2", 99, model.StatusApproved
		if err := r.Update(ctx, c); err != nil {
			t.Fatal(err)
		}
		got, _ := r.Get(ctx, c.ID)
		if got.Hospital != "H2" || got.AmountSatang != 99 || got.Status != model.StatusApproved {
			t.Fatalf("got %+v", got)
		}
		if err := r.Update(ctx, model.Claim{ID: 999}); !errors.Is(err, model.ErrNotFound) {
			t.Fatalf("err=%v", err)
		}
	})

	t.Run("delete", func(t *testing.T) {
		r := newRepo(t)
		c, _ := r.Create(ctx, mk("P", "H", 10, model.StatusPending))
		if err := r.Delete(ctx, c.ID); err != nil {
			t.Fatal(err)
		}
		if _, err := r.Get(ctx, c.ID); !errors.Is(err, model.ErrNotFound) {
			t.Fatalf("err=%v", err)
		}
		if err := r.Delete(ctx, c.ID); !errors.Is(err, model.ErrNotFound) {
			t.Fatalf("ลบซ้ำ err=%v", err)
		}
	})

	t.Run("update status many is all-or-nothing", func(t *testing.T) {
		r := newRepo(t)
		a, _ := r.Create(ctx, mk("A", "H", 1, model.StatusPending))
		b, _ := r.Create(ctx, mk("B", "H", 1, model.StatusPending))
		later := at.Add(time.Hour)

		if err := r.UpdateStatusMany(ctx, []int64{a.ID, b.ID}, model.StatusApproved, later); err != nil {
			t.Fatal(err)
		}
		ga, _ := r.Get(ctx, a.ID)
		if ga.Status != model.StatusApproved || !ga.UpdatedAt.Equal(later) {
			t.Fatalf("got %+v", ga)
		}

		c, _ := r.Create(ctx, mk("C", "H", 1, model.StatusPending))
		err := r.UpdateStatusMany(ctx, []int64{c.ID, 999}, model.StatusRejected, later)
		if !errors.Is(err, model.ErrNotFound) {
			t.Fatalf("err=%v", err)
		}
		gc, _ := r.Get(ctx, c.ID)
		if gc.Status != model.StatusPending {
			t.Fatalf("ต้อง rollback: c.Status=%s", gc.Status)
		}
	})
}
