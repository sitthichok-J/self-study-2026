package service

import (
	"errors"
	"testing"

	"github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/domain"
)

var staff = domain.Viewer{ID: "ops", Staff: true}

func TestProcessLeavesReviewBandPending(t *testing.T) {
	e := newEnv()
	c := submit(t, e, "k1", "FAKE-AB12", domain.ReviewLimit+1)
	for i := 0; i < 3; i++ { // redelivery must not change anything
		if err := e.svc.Process(ctx, c.ID); err != nil {
			t.Fatal(err)
		}
	}
	got, _ := e.repo.Get(ctx, c.ID)
	if got.Status != domain.StatusPending || got.DecisionReason != "needs_review" {
		t.Fatalf("%s/%s", got.Status, got.DecisionReason)
	}
}

func TestDecideAfterReviewBand(t *testing.T) {
	e := newEnv()
	c := submit(t, e, "k1", "FAKE-AB12", domain.OverLimit)
	_ = e.svc.Process(ctx, c.ID)
	got, err := e.svc.Decide(ctx, staff, c.ID, domain.StatusApproved, "manual_approved")
	if err != nil || got.Status != domain.StatusApproved || got.DecisionReason != "manual_approved" {
		t.Fatalf("%+v %v", got, err)
	}
	// the worker seeing the same message again must not undo the decision
	_ = e.svc.Process(ctx, c.ID)
	now, _ := e.repo.Get(ctx, c.ID)
	if now.Status != domain.StatusApproved {
		t.Fatalf("worker overwrote: %s", now.Status)
	}
}

func TestDecideBeforeWorkerWins(t *testing.T) {
	e := newEnv()
	c := submit(t, e, "k1", "FAKE-AB12", 1000) // worker would approve it
	if _, err := e.svc.Decide(ctx, staff, c.ID, domain.StatusRejected, "manual_rejected"); err != nil {
		t.Fatal(err)
	}
	_ = e.svc.Process(ctx, c.ID)
	got, _ := e.repo.Get(ctx, c.ID)
	if got.Status != domain.StatusRejected {
		t.Fatalf("person's decision lost: %s", got.Status)
	}
}

func TestDecideRules(t *testing.T) {
	e := newEnv()
	c := submit(t, e, "k1", "FAKE-AB12", domain.OverLimit)
	if _, err := e.svc.Decide(ctx, domain.Viewer{ID: "alice"}, c.ID, domain.StatusApproved, "x"); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("non-staff: %v", err)
	}
	if _, err := e.svc.Decide(ctx, domain.Viewer{}, c.ID, domain.StatusApproved, "x"); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("anonymous: %v", err)
	}
	if _, err := e.svc.Decide(ctx, staff, "nope", domain.StatusApproved, "x"); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("bad id: %v", err)
	}
	if _, err := e.svc.Decide(ctx, staff, c.ID, domain.StatusPending, "x"); !errors.Is(err, domain.ErrInvalidTransition) {
		t.Fatalf("to PENDING: %v", err)
	}
	if _, err := e.svc.Decide(ctx, staff, "00000000-0000-4000-8000-0000000000ff", domain.StatusApproved, "x"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("unknown: %v", err)
	}
	if _, err := e.svc.Decide(ctx, staff, c.ID, domain.StatusApproved, "manual_approved"); err != nil {
		t.Fatal(err)
	}
	if _, err := e.svc.Decide(ctx, staff, c.ID, domain.StatusRejected, "manual_rejected"); !errors.Is(err, domain.ErrInvalidTransition) {
		t.Fatalf("second decision: %v", err)
	}
}

func TestDecideDropsCachedCopy(t *testing.T) {
	e := newEnv()
	c := submit(t, e, "k1", "FAKE-AB12", domain.OverLimit)
	_, _ = e.svc.Get(ctx, staff, c.ID) // warms the cache with PENDING
	_, _ = e.svc.Decide(ctx, staff, c.ID, domain.StatusApproved, "manual_approved")
	got, _ := e.svc.Get(ctx, staff, c.ID)
	if got.Status != domain.StatusApproved {
		t.Fatalf("stale cache: %s", got.Status)
	}
}
