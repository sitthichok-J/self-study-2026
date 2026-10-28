// Package service: business logic ของ claims ไม่รู้จัก HTTP และไม่รู้ว่า DB คืออะไร
// interface Repository ถูกประกาศที่นี่ (ฝั่งผู้ใช้) และเล็กเท่าที่ใช้
package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/sitthichok-J/self-study-2026/day05-go-service/internal/model"
	"github.com/sitthichok-J/self-study-2026/day05-go-service/internal/report"
)

type Repository interface {
	Create(ctx context.Context, c model.Claim) (model.Claim, error)
	Get(ctx context.Context, id int64) (model.Claim, error)
	List(ctx context.Context, f model.ListFilter) ([]model.Claim, error)
	Update(ctx context.Context, c model.Claim) error
	Delete(ctx context.Context, id int64) error
	// UpdateStatusMany เปลี่ยนสถานะหลายรายการแบบ all-or-nothing (ถ้ามี id ที่ไม่พบ ต้องไม่มีอะไรเปลี่ยน)
	UpdateStatusMany(ctx context.Context, ids []int64, s model.Status, at time.Time) error
}

const (
	DefaultLimit = 20
	MaxLimit     = 100
	MaxBatch     = 100
)

// ValidationError ใช้บอก field ที่ผิด handler แปลงเป็น 400
type ValidationError struct{ Field, Reason string }

func (e *ValidationError) Error() string { return fmt.Sprintf("invalid %s: %s", e.Field, e.Reason) }

func invalid(field, reason string) error { return &ValidationError{Field: field, Reason: reason} }

// Input คือข้อมูลที่ผู้ใช้ส่งมาตอนสร้าง/แก้ไข
type Input struct {
	PatientRef   string
	Hospital     string
	AmountSatang int64
	Status       model.Status
}

type Service struct {
	repo Repository
	now  func() time.Time
}

// New: now เป็น nil = ใช้ time.Now (ส่ง fake เข้ามาตอนเทสต์เพื่อให้เวลานิ่ง)
func New(repo Repository, now func() time.Time) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{repo: repo, now: now}
}

func normalize(in Input) (Input, error) {
	in.PatientRef = strings.TrimSpace(in.PatientRef)
	in.Hospital = strings.TrimSpace(in.Hospital)
	switch {
	case in.PatientRef == "":
		return in, invalid("patient_ref", "required")
	case utf8.RuneCountInString(in.PatientRef) > 32:
		return in, invalid("patient_ref", "max 32 characters")
	case in.Hospital == "":
		return in, invalid("hospital", "required")
	case utf8.RuneCountInString(in.Hospital) > 100:
		return in, invalid("hospital", "max 100 characters")
	case in.AmountSatang <= 0:
		return in, invalid("amount_satang", "must be greater than 0")
	}
	return in, nil
}

func (s *Service) Create(ctx context.Context, in Input) (model.Claim, error) {
	if in.Status == "" {
		in.Status = model.StatusPending
	}
	if !in.Status.Valid() {
		return model.Claim{}, invalid("status", "must be PENDING, APPROVED or REJECTED")
	}
	in, err := normalize(in)
	if err != nil {
		return model.Claim{}, err
	}
	now := s.now()
	c, err := s.repo.Create(ctx, model.Claim{
		PatientRef: in.PatientRef, Hospital: in.Hospital, AmountSatang: in.AmountSatang,
		Status: in.Status, CreatedAt: now, UpdatedAt: now,
	})
	if err != nil {
		return model.Claim{}, fmt.Errorf("create claim: %w", err)
	}
	return c, nil
}

func (s *Service) Get(ctx context.Context, id int64) (model.Claim, error) {
	c, err := s.repo.Get(ctx, id)
	if err != nil {
		return model.Claim{}, fmt.Errorf("get claim %d: %w", id, err)
	}
	return c, nil
}

func (s *Service) List(ctx context.Context, f model.ListFilter) ([]model.Claim, error) {
	if f.Status != "" && !f.Status.Valid() {
		return nil, invalid("status", "must be PENDING, APPROVED or REJECTED")
	}
	if f.Limit < 0 || f.Offset < 0 {
		return nil, invalid("limit/offset", "must not be negative")
	}
	switch {
	case f.Limit == 0:
		f.Limit = DefaultLimit
	case f.Limit > MaxLimit:
		f.Limit = MaxLimit
	}
	list, err := s.repo.List(ctx, f)
	if err != nil {
		return nil, fmt.Errorf("list claims: %w", err)
	}
	return list, nil
}

// Update แทนที่ฟิลด์ทั้งหมด (PUT) แต่คง ID และ CreatedAt เดิม
func (s *Service) Update(ctx context.Context, id int64, in Input) (model.Claim, error) {
	if !in.Status.Valid() {
		return model.Claim{}, invalid("status", "must be PENDING, APPROVED or REJECTED")
	}
	in, err := normalize(in)
	if err != nil {
		return model.Claim{}, err
	}
	cur, err := s.repo.Get(ctx, id)
	if err != nil {
		return model.Claim{}, fmt.Errorf("update claim %d: %w", id, err)
	}
	cur.PatientRef, cur.Hospital, cur.AmountSatang, cur.Status = in.PatientRef, in.Hospital, in.AmountSatang, in.Status
	cur.UpdatedAt = s.now()
	if err := s.repo.Update(ctx, cur); err != nil {
		return model.Claim{}, fmt.Errorf("update claim %d: %w", id, err)
	}
	return cur, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete claim %d: %w", id, err)
	}
	return nil
}

// ApproveMany อนุมัติหลายรายการใน transaction เดียว (all-or-nothing)
func (s *Service) ApproveMany(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return invalid("ids", "required")
	}
	if len(ids) > MaxBatch {
		return invalid("ids", fmt.Sprintf("max %d per request", MaxBatch))
	}
	seen := make(map[int64]struct{}, len(ids))
	uniq := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			return invalid("ids", "must be positive")
		}
		if _, dup := seen[id]; !dup {
			seen[id] = struct{}{}
			uniq = append(uniq, id)
		}
	}
	if err := s.repo.UpdateStatusMany(ctx, uniq, model.StatusApproved, s.now()); err != nil {
		return fmt.Errorf("approve claims: %w", err)
	}
	return nil
}

// All ดึงทุกรายการ (ใช้กับรายงาน)
func (s *Service) All(ctx context.Context) ([]model.Claim, error) {
	list, err := s.repo.List(ctx, model.ListFilter{})
	if err != nil {
		return nil, fmt.Errorf("list all claims: %w", err)
	}
	return list, nil
}

// Report สรุปยอดรวมต่อโรงพยาบาล
func (s *Service) Report(ctx context.Context) (string, error) {
	list, err := s.All(ctx)
	if err != nil {
		return "", err
	}
	return report.Build(list), nil
}

// IsValidation ช่วยให้ชั้นบนเช็กโดยไม่ต้องรู้ชนิดจริง
func IsValidation(err error) bool {
	var ve *ValidationError
	return errors.As(err, &ve)
}
