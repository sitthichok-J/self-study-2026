package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/sitthichok-J/self-study-2026/day05-go-service/internal/model"
	"github.com/sitthichok-J/self-study-2026/day05-go-service/internal/repository"
)

var t0 = time.Date(2026, 10, 28, 9, 0, 0, 0, time.UTC)

func newSvc() (*Service, *time.Time) {
	now := t0
	return New(repository.NewMemory(), func() time.Time { return now }), &now
}

func good() Input {
	return Input{PatientRef: "P-001", Hospital: "รพ.ตัวอย่าง", AmountSatang: 125000}
}

func TestCreateDefaultsAndTrim(t *testing.T) {
	s, _ := newSvc()
	c, err := s.Create(context.Background(), Input{PatientRef: "  P-001 ", Hospital: " H ", AmountSatang: 1})
	if err != nil {
		t.Fatal(err)
	}
	if c.ID != 1 || c.PatientRef != "P-001" || c.Hospital != "H" || c.Status != model.StatusPending ||
		!c.CreatedAt.Equal(t0) || !c.UpdatedAt.Equal(t0) {
		t.Fatalf("%+v", c)
	}
}

func TestCreateValidation(t *testing.T) {
	long := strings.Repeat("ก", 33)
	tests := []struct {
		name  string
		mod   func(*Input)
		field string
	}{
		{"patient_ref ว่าง", func(i *Input) { i.PatientRef = "  " }, "patient_ref"},
		{"patient_ref ยาวเกิน (นับเป็นตัวอักษร ไม่ใช่ byte)", func(i *Input) { i.PatientRef = long }, "patient_ref"},
		{"hospital ว่าง", func(i *Input) { i.Hospital = "" }, "hospital"},
		{"hospital ยาวเกิน", func(i *Input) { i.Hospital = strings.Repeat("x", 101) }, "hospital"},
		{"amount เป็น 0", func(i *Input) { i.AmountSatang = 0 }, "amount_satang"},
		{"amount ติดลบ", func(i *Input) { i.AmountSatang = -5 }, "amount_satang"},
		{"status ไม่รู้จัก", func(i *Input) { i.Status = "DONE" }, "status"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s, _ := newSvc()
			in := good()
			tc.mod(&in)
			_, err := s.Create(context.Background(), in)
			var ve *ValidationError
			if !errors.As(err, &ve) || ve.Field != tc.field {
				t.Fatalf("err=%v want field %s", err, tc.field)
			}
			if !IsValidation(err) {
				t.Fatal("IsValidation ต้องเป็นจริง")
			}
		})
	}
	if _, err := New(repository.NewMemory(), nil).Create(context.Background(), good()); err != nil { // now=nil ใช้ time.Now
		t.Fatal(err)
	}
}

func TestPatientRef32RuneBoundaryOK(t *testing.T) {
	s, _ := newSvc()
	in := good()
	in.PatientRef = strings.Repeat("ก", 32)
	if _, err := s.Create(context.Background(), in); err != nil {
		t.Fatalf("32 ตัวอักษรต้องผ่าน: %v", err)
	}
}

func TestGetNotFoundWrapped(t *testing.T) {
	s, _ := newSvc()
	_, err := s.Get(context.Background(), 42)
	if !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("err=%v", err)
	}
	if !strings.Contains(err.Error(), "42") {
		t.Fatalf("ข้อความควรมี id: %v", err)
	}
}

func TestListDefaultsClampAndValidation(t *testing.T) {
	s, _ := newSvc()
	ctx := context.Background()
	for i := 0; i < 130; i++ {
		if _, err := s.Create(ctx, good()); err != nil {
			t.Fatal(err)
		}
	}
	tests := []struct {
		name    string
		f       model.ListFilter
		wantLen int
		wantErr bool
	}{
		{"ค่าเริ่มต้น = 20", model.ListFilter{}, DefaultLimit, false},
		{"จำกัดเพดาน 100", model.ListFilter{Limit: 5000}, MaxLimit, false},
		{"ระบุเอง", model.ListFilter{Limit: 7}, 7, false},
		{"offset", model.ListFilter{Limit: 100, Offset: 100}, 30, false},
		{"status ผิด", model.ListFilter{Status: "X"}, 0, true},
		{"limit ติดลบ", model.ListFilter{Limit: -1}, 0, true},
		{"offset ติดลบ", model.ListFilter{Offset: -1}, 0, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := s.List(ctx, tc.f)
			if (err != nil) != tc.wantErr || len(got) != tc.wantLen {
				t.Fatalf("len=%d err=%v", len(got), err)
			}
			if tc.wantErr && !IsValidation(err) {
				t.Fatalf("ต้องเป็น validation: %v", err)
			}
		})
	}
}

func TestUpdateKeepsCreatedAtAndBumpsUpdatedAt(t *testing.T) {
	s, now := newSvc()
	ctx := context.Background()
	c, _ := s.Create(ctx, good())
	*now = t0.Add(2 * time.Hour)
	in := Input{PatientRef: "P-002", Hospital: "H2", AmountSatang: 999, Status: model.StatusApproved}
	u, err := s.Update(ctx, c.ID, in)
	if err != nil {
		t.Fatal(err)
	}
	if u.ID != c.ID || !u.CreatedAt.Equal(t0) || !u.UpdatedAt.Equal(*now) || u.Status != model.StatusApproved || u.Hospital != "H2" {
		t.Fatalf("%+v", u)
	}
	got, _ := s.Get(ctx, c.ID)
	if got != u {
		t.Fatalf("ที่เก็บไม่ตรง %+v vs %+v", got, u)
	}
}

func TestUpdateErrors(t *testing.T) {
	s, _ := newSvc()
	ctx := context.Background()
	in := good()
	in.Status = model.StatusPending
	if _, err := s.Update(ctx, 99, in); !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("ไม่พบ: %v", err)
	}
	in.Status = ""
	if _, err := s.Update(ctx, 1, in); !IsValidation(err) {
		t.Fatalf("PUT ต้องระบุ status: %v", err)
	}
	in.Status = model.StatusPending
	in.Hospital = ""
	if _, err := s.Update(ctx, 1, in); !IsValidation(err) {
		t.Fatalf("hospital ว่าง: %v", err)
	}
}

func TestDelete(t *testing.T) {
	s, _ := newSvc()
	ctx := context.Background()
	c, _ := s.Create(ctx, good())
	if err := s.Delete(ctx, c.ID); err != nil {
		t.Fatal(err)
	}
	if err := s.Delete(ctx, c.ID); !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("err=%v", err)
	}
}

func TestApproveMany(t *testing.T) {
	ctx := context.Background()
	t.Run("สำเร็จและตัด id ซ้ำ", func(t *testing.T) {
		s, now := newSvc()
		a, _ := s.Create(ctx, good())
		b, _ := s.Create(ctx, good())
		*now = t0.Add(time.Minute)
		if err := s.ApproveMany(ctx, []int64{a.ID, b.ID, a.ID}); err != nil {
			t.Fatal(err)
		}
		ga, _ := s.Get(ctx, a.ID)
		if ga.Status != model.StatusApproved || !ga.UpdatedAt.Equal(*now) {
			t.Fatalf("%+v", ga)
		}
	})
	t.Run("มี id ที่ไม่พบ = ไม่มีอะไรเปลี่ยน", func(t *testing.T) {
		s, _ := newSvc()
		a, _ := s.Create(ctx, good())
		err := s.ApproveMany(ctx, []int64{a.ID, 999})
		if !errors.Is(err, model.ErrNotFound) {
			t.Fatalf("err=%v", err)
		}
		if ga, _ := s.Get(ctx, a.ID); ga.Status != model.StatusPending {
			t.Fatalf("ต้องไม่เปลี่ยน: %+v", ga)
		}
	})
	t.Run("input ผิด", func(t *testing.T) {
		s, _ := newSvc()
		big := make([]int64, MaxBatch+1)
		for i := range big {
			big[i] = int64(i + 1)
		}
		for name, ids := range map[string][]int64{"ว่าง": nil, "id ศูนย์": {0}, "id ติดลบ": {-1, 2}, "เกินเพดาน": big} {
			if err := s.ApproveMany(ctx, ids); !IsValidation(err) {
				t.Fatalf("%s: err=%v", name, err)
			}
		}
	})
}

func TestReport(t *testing.T) {
	s, _ := newSvc()
	ctx := context.Background()
	for _, h := range []string{"B", "A", "B"} {
		in := good()
		in.Hospital, in.AmountSatang = h, 100
		s.Create(ctx, in)
	}
	got, err := s.Report(ctx)
	if err != nil || got != "A\t1\t100\nB\t2\t200\n" {
		t.Fatalf("got %q err %v", got, err)
	}
	if all, _ := s.All(ctx); len(all) != 3 {
		t.Fatalf("all=%d", len(all))
	}
}

// failRepo จำลอง DB พัง ทุกเมธอดคืน error เดียวกัน
type failRepo struct{ err error }

func (f failRepo) Create(context.Context, model.Claim) (model.Claim, error) {
	return model.Claim{}, f.err
}
func (f failRepo) Get(context.Context, int64) (model.Claim, error) { return model.Claim{}, f.err }
func (f failRepo) List(context.Context, model.ListFilter) ([]model.Claim, error) {
	return nil, f.err
}
func (f failRepo) Update(context.Context, model.Claim) error { return f.err }
func (f failRepo) Delete(context.Context, int64) error       { return f.err }
func (f failRepo) UpdateStatusMany(context.Context, []int64, model.Status, time.Time) error {
	return f.err
}

func TestRepoErrorsAreWrappedNotLost(t *testing.T) {
	boom := errors.New("db down")
	s := New(failRepo{err: boom}, nil)
	ctx := context.Background()
	calls := map[string]func() error{
		"create": func() error { _, e := s.Create(ctx, good()); return e },
		"get":    func() error { _, e := s.Get(ctx, 1); return e },
		"list":   func() error { _, e := s.List(ctx, model.ListFilter{}); return e },
		"update": func() error {
			in := good()
			in.Status = model.StatusPending
			_, e := s.Update(ctx, 1, in)
			return e
		},
		"delete":  func() error { return s.Delete(ctx, 1) },
		"approve": func() error { return s.ApproveMany(ctx, []int64{1}) },
		"report":  func() error { _, e := s.Report(ctx); return e },
	}
	for name, fn := range calls {
		t.Run(name, func(t *testing.T) {
			err := fn()
			if !errors.Is(err, boom) {
				t.Fatalf("errors.Is ต้องเจอต้นเหตุ: %v", err)
			}
			if IsValidation(err) {
				t.Fatal("error ของ DB ต้องไม่ถูกมองเป็น validation")
			}
		})
	}
}
