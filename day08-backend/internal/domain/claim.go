// Package domain holds the Claim entity and its rules. No I/O here.
package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"time"
)

type Status string

const (
	StatusPending  Status = "PENDING"
	StatusApproved Status = "APPROVED"
	StatusRejected Status = "REJECTED"
)

var (
	ErrNotFound            = errors.New("claim not found")
	ErrInvalid             = errors.New("invalid input")
	ErrIdempotencyRequired = errors.New("idempotency key required")
	ErrIdempotencyConflict = errors.New("idempotency key reused with a different request")
	ErrInvalidTransition   = errors.New("invalid status transition")
)

// patient_ref must look like FAKE-XXXX so real identifiers can never be
// stored by accident (the exercise is fake-data only).
var patientRefRe = regexp.MustCompile(`^FAKE-[A-Z0-9]{3,12}$`)

const MaxAmount = 1_000_000_000 // satang, sanity cap

type Claim struct {
	ID             string    `json:"id"`
	PatientRef     string    `json:"patient_ref"`
	Amount         int64     `json:"amount"` // satang (minor unit), never float
	Status         Status    `json:"status"`
	DecisionReason string    `json:"decision_reason,omitempty"`
	IdempotencyKey string    `json:"-"`
	RequestHash    string    `json:"-"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// NewClaim validates input and returns a PENDING claim.
func NewClaim(id, idemKey, patientRef string, amount int64, now time.Time) (Claim, error) {
	if idemKey == "" {
		return Claim{}, ErrIdempotencyRequired
	}
	if len(idemKey) > 100 {
		return Claim{}, fmt.Errorf("%w: idempotency key longer than 100 chars", ErrInvalid)
	}
	if !patientRefRe.MatchString(patientRef) {
		return Claim{}, fmt.Errorf("%w: patient_ref must match FAKE-[A-Z0-9]{3,12}", ErrInvalid)
	}
	if amount <= 0 || amount > MaxAmount {
		return Claim{}, fmt.Errorf("%w: amount must be between 1 and %d", ErrInvalid, MaxAmount)
	}
	return Claim{
		ID: id, PatientRef: patientRef, Amount: amount, Status: StatusPending,
		IdempotencyKey: idemKey, RequestHash: RequestHash(patientRef, amount),
		CreatedAt: now.UTC(), UpdatedAt: now.UTC(),
	}, nil
}

// RequestHash identifies "the same request" for idempotency checks.
func RequestHash(patientRef string, amount int64) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s|%d", patientRef, amount)))
	return hex.EncodeToString(sum[:])
}

// CanTransition: only PENDING can move, and only to a final state.
func CanTransition(from, to Status) error {
	if from == StatusPending && (to == StatusApproved || to == StatusRejected) {
		return nil
	}
	return fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, from, to)
}

func ParseStatus(s string) (Status, error) {
	switch Status(s) {
	case StatusPending, StatusApproved, StatusRejected:
		return Status(s), nil
	}
	return "", fmt.Errorf("%w: unknown status %q", ErrInvalid, s)
}

var uuidRe = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func IsUUID(s string) bool { return uuidRe.MatchString(s) }

// NewUUID returns a random v4 UUID (stdlib only).
func NewUUID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// Decision is the outcome of evaluating a claim.
type Decision struct {
	Status Status
	Reason string
}

// ErrTransient marks a failure worth retrying (e.g. downstream check down).
var ErrTransient = errors.New("transient evaluation failure")

// OverLimit: amounts above this (satang) are rejected. 50,000 THB.
const OverLimit = 5_000_000

// PoisonRef is a demo hook: this patient_ref always fails evaluation so the
// retry -> DLQ path can be shown. It stands in for a real downstream outage.
const PoisonRef = "FAKE-POISON"

// Evaluate is the simulated "verification" step.
func Evaluate(c Claim) (Decision, error) {
	if c.PatientRef == PoisonRef {
		return Decision{}, fmt.Errorf("%w: simulated downstream failure", ErrTransient)
	}
	if c.Amount > OverLimit {
		return Decision{StatusRejected, "over_limit"}, nil
	}
	return Decision{StatusApproved, "ok"}, nil
}
