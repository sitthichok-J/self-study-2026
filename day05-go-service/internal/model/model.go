// Package model: ชนิดข้อมูลกลางของ claim (ใช้ร่วมกันทุกชั้น)
package model

import (
	"errors"
	"time"
)

type Status string

const (
	StatusPending  Status = "PENDING"
	StatusApproved Status = "APPROVED"
	StatusRejected Status = "REJECTED"
)

func (s Status) Valid() bool {
	switch s {
	case StatusPending, StatusApproved, StatusRejected:
		return true
	}
	return false
}

// ErrNotFound: ทุก repository ต้องคืน error นี้ (หรือห่อด้วย %w) เมื่อไม่พบ claim
var ErrNotFound = errors.New("claim not found")

// Claim เก็บจำนวนเงินเป็น "สตางค์" (int64) ไม่ใช้ float เพื่อเลี่ยงการปัดเศษของเงิน
type Claim struct {
	ID           int64     `json:"id"`
	PatientRef   string    `json:"patient_ref"`
	Hospital     string    `json:"hospital"`
	AmountSatang int64     `json:"amount_satang"`
	Status       Status    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ListFilter: Limit == 0 หมายถึงไม่จำกัด (ชั้น service เป็นคนบังคับค่าเริ่มต้น/เพดาน)
type ListFilter struct {
	Status Status // ว่าง = ทุกสถานะ
	Limit  int
	Offset int
}
