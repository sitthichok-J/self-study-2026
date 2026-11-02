package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sitthichok-J/self-study-2026/day08-backend/internal/domain"
	"github.com/sitthichok-J/self-study-2026/day08-backend/internal/memory"
	"github.com/sitthichok-J/self-study-2026/day08-backend/internal/port"
)

var quiet = slog.New(slog.NewTextHandler(io.Discard, nil))

type env struct {
	repo  *memory.Repo
	cache *memory.Cache
	pub   *memory.Publisher
	svc   *Claims
	cons  *Consumer
	relay *Relay
}

func newEnv() *env {
	e := &env{repo: memory.NewRepo(), cache: memory.NewCache(), pub: &memory.Publisher{}}
	var n atomic.Int64
	base := time.Date(2026, 11, 3, 9, 0, 0, 0, time.UTC)
	e.svc = NewClaims(e.repo, e.cache, quiet).
		WithClock(func() time.Time { return base.Add(time.Duration(n.Add(1)) * time.Second) }).
		WithIDs(func() string { return fmt.Sprintf("00000000-0000-4000-8000-%012d", n.Add(1)) })
	e.cons = NewConsumer(e.svc, e.pub, quiet)
	e.cons.Sleep = func(context.Context, time.Duration) {}
	e.relay = NewRelay(e.repo, e.pub, quiet)
	return e
}

var ctx = context.Background()

func submit(t *testing.T, e *env, key, ref string, amt int64) domain.Claim {
	t.Helper()
	c, _, err := e.svc.Submit(ctx, SubmitInput{key, ref, amt})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func msgFor(id string) port.Message {
	b, _ := json.Marshal(SubmitEvent{Event: port.TopicSubmitted, Version: 1, ClaimID: id})
	return port.Message{Topic: port.TopicSubmitted, Key: id, Value: b}
}

// ---- submit / idempotency ----

func TestSubmitCreatesPendingAndOutbox(t *testing.T) {
	e := newEnv()
	c, created, err := e.svc.Submit(ctx, SubmitInput{"k1", "FAKE-AB12", 1000})
	if err != nil || !created {
		t.Fatalf("created=%v err=%v", created, err)
	}
	if c.Status != domain.StatusPending {
		t.Fatalf("status %s", c.Status)
	}
	if e.repo.OutboxPending() != 1 {
		t.Fatalf("expected 1 outbox event, got %d", e.repo.OutboxPending())
	}
}

func TestSubmitInvalidWritesNothing(t *testing.T) {
	e := newEnv()
	_, _, err := e.svc.Submit(ctx, SubmitInput{"k1", "real-name", 1000})
	if !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("want ErrInvalid got %v", err)
	}
	if e.repo.Count() != 0 || e.repo.OutboxPending() != 0 {
		t.Fatal("invalid input must not write claim or outbox")
	}
}

func TestIdempotencySameKeySamePayload(t *testing.T) {
	e := newEnv()
	a := submit(t, e, "k1", "FAKE-AB12", 1000)
	b, created, err := e.svc.Submit(ctx, SubmitInput{"k1", "FAKE-AB12", 1000})
	if err != nil || created {
		t.Fatalf("replay should return existing: created=%v err=%v", created, err)
	}
	if a.ID != b.ID {
		t.Fatal("replay must return the same claim id")
	}
	if e.repo.Count() != 1 || e.repo.OutboxPending() != 1 {
		t.Fatalf("replay must not add rows: claims=%d outbox=%d", e.repo.Count(), e.repo.OutboxPending())
	}
}

func TestIdempotencySameKeyDifferentPayload(t *testing.T) {
	e := newEnv()
	submit(t, e, "k1", "FAKE-AB12", 1000)
	_, _, err := e.svc.Submit(ctx, SubmitInput{"k1", "FAKE-AB12", 2000})
	if !errors.Is(err, domain.ErrIdempotencyConflict) {
		t.Fatalf("want conflict got %v", err)
	}
}

func TestIdempotencyConcurrent(t *testing.T) {
	e := newEnv()
	var wg sync.WaitGroup
	ids := make(chan string, 20)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c, _, err := e.svc.Submit(ctx, SubmitInput{"same", "FAKE-AB12", 1000})
			if err == nil {
				ids <- c.ID
			}
		}()
	}
	wg.Wait()
	close(ids)
	seen := map[string]bool{}
	for id := range ids {
		seen[id] = true
	}
	if len(seen) != 1 || e.repo.Count() != 1 {
		t.Fatalf("20 concurrent submits with one key must yield 1 claim, got %d ids / %d rows", len(seen), e.repo.Count())
	}
}

// ---- outbox relay ----

func TestRelayPublishesAndMarksSent(t *testing.T) {
	e := newEnv()
	c := submit(t, e, "k1", "FAKE-AB12", 1000)
	n, err := e.relay.RunOnce(ctx)
	if err != nil || n != 1 {
		t.Fatalf("n=%d err=%v", n, err)
	}
	got := e.pub.On(port.TopicSubmitted)
	if len(got) != 1 || got[0].Key != c.ID {
		t.Fatalf("unexpected messages %+v", got)
	}
	if e.repo.OutboxPending() != 0 {
		t.Fatal("event should be marked sent")
	}
	if n, _ := e.relay.RunOnce(ctx); n != 0 {
		t.Fatal("second run must send nothing")
	}
}

func TestRelayBrokerDownKeepsEventAndRetries(t *testing.T) {
	e := newEnv()
	submit(t, e, "k1", "FAKE-AB12", 1000)
	e.pub.FailTimes = 1
	if _, err := e.relay.RunOnce(ctx); err == nil {
		t.Fatal("expected publish error")
	}
	if e.repo.OutboxPending() != 1 {
		t.Fatal("failed publish must stay in outbox (no lost event)")
	}
	if n, err := e.relay.RunOnce(ctx); err != nil || n != 1 {
		t.Fatalf("retry should succeed: n=%d err=%v", n, err)
	}
}

func TestRelayStopsAtFirstFailurePreservingOrder(t *testing.T) {
	e := newEnv()
	submit(t, e, "k1", "FAKE-AB12", 1000)
	submit(t, e, "k2", "FAKE-AB13", 1000)
	e.pub.FailTimes = 1
	n, err := e.relay.RunOnce(ctx)
	if err == nil || n != 0 {
		t.Fatalf("n=%d err=%v", n, err)
	}
	if got := len(e.pub.On(port.TopicSubmitted)); got != 0 {
		t.Fatalf("later event must not jump ahead, sent %d", got)
	}
}

// ---- process (worker) ----

func TestProcessApprovesAndRejects(t *testing.T) {
	e := newEnv()
	ok := submit(t, e, "k1", "FAKE-AB12", 1000)
	big := submit(t, e, "k2", "FAKE-AB13", domain.OverLimit+1)
	for _, c := range []domain.Claim{ok, big} {
		if err := e.svc.Process(ctx, c.ID); err != nil {
			t.Fatal(err)
		}
	}
	a, _ := e.repo.Get(ctx, ok.ID)
	b, _ := e.repo.Get(ctx, big.ID)
	if a.Status != domain.StatusApproved || b.Status != domain.StatusRejected || b.DecisionReason != "over_limit" {
		t.Fatalf("a=%s b=%s/%s", a.Status, b.Status, b.DecisionReason)
	}
}

func TestProcessIsIdempotentOnRedelivery(t *testing.T) {
	e := newEnv()
	c := submit(t, e, "k1", "FAKE-AB12", 1000)
	for i := 0; i < 3; i++ {
		if err := e.svc.Process(ctx, c.ID); err != nil {
			t.Fatalf("delivery %d: %v", i, err)
		}
	}
	got, _ := e.repo.Get(ctx, c.ID)
	if got.Status != domain.StatusApproved {
		t.Fatalf("status %s", got.Status)
	}
}

func TestProcessDoesNotOverwriteFinalState(t *testing.T) {
	e := newEnv()
	c := submit(t, e, "k1", "FAKE-AB12", domain.OverLimit+1)
	_ = e.svc.Process(ctx, c.ID) // REJECTED
	// simulate a bug/replay: manually try to flip it
	if ok, _ := e.repo.UpdateStatus(ctx, c.ID, domain.StatusPending, domain.StatusApproved, "x"); ok {
		t.Fatal("conditional update must not touch non-PENDING rows")
	}
	got, _ := e.repo.Get(ctx, c.ID)
	if got.Status != domain.StatusRejected {
		t.Fatalf("final state changed to %s", got.Status)
	}
}

func TestProcessUnknownClaim(t *testing.T) {
	e := newEnv()
	err := e.svc.Process(ctx, "00000000-0000-4000-8000-0000000000ff")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("got %v", err)
	}
}

// ---- consumer: retry + DLQ ----

func TestConsumerHappyPath(t *testing.T) {
	e := newEnv()
	c := submit(t, e, "k1", "FAKE-AB12", 1000)
	if err := e.cons.Handle(ctx, msgFor(c.ID)); err != nil {
		t.Fatal(err)
	}
	if len(e.pub.On(port.TopicDLQ)) != 0 {
		t.Fatal("nothing should reach the DLQ")
	}
	got, _ := e.repo.Get(ctx, c.ID)
	if got.Status != domain.StatusApproved {
		t.Fatal("not approved")
	}
}

func TestConsumerRetriesTransientThenSucceeds(t *testing.T) {
	e := newEnv()
	c := submit(t, e, "k1", "FAKE-AB12", 1000)
	e.repo.FailGet = errors.New("db blip")
	calls := 0
	e.cons.Sleep = func(context.Context, time.Duration) {
		calls++
		if calls == 2 {
			e.repo.FailGet = nil // recovers before the 3rd attempt
		}
	}
	if err := e.cons.Handle(ctx, msgFor(c.ID)); err != nil {
		t.Fatal(err)
	}
	if calls != 2 || len(e.pub.On(port.TopicDLQ)) != 0 {
		t.Fatalf("sleeps=%d dlq=%d", calls, len(e.pub.On(port.TopicDLQ)))
	}
}

func TestConsumerSendsPoisonToDLQAfterMaxAttempts(t *testing.T) {
	e := newEnv()
	c := submit(t, e, "k1", domain.PoisonRef, 1000)
	if err := e.cons.Handle(ctx, msgFor(c.ID)); err != nil {
		t.Fatalf("parked message counts as done: %v", err)
	}
	dlq := e.pub.On(port.TopicDLQ)
	if len(dlq) != 1 {
		t.Fatalf("dlq=%d", len(dlq))
	}
	var env deadLetterEnvelope
	if err := json.Unmarshal(dlq[0].Value, &env); err != nil {
		t.Fatal(err)
	}
	if env.OriginalTopic != port.TopicSubmitted || env.Reason == "" || dlq[0].Key != c.ID {
		t.Fatalf("bad envelope %+v", env)
	}
	got, _ := e.repo.Get(ctx, c.ID)
	if got.Status != domain.StatusPending {
		t.Fatal("poison claim should stay PENDING for manual follow-up")
	}
}

func TestConsumerAttemptCountIsBounded(t *testing.T) {
	e := newEnv()
	c := submit(t, e, "k1", domain.PoisonRef, 1000)
	sleeps := 0
	e.cons.Sleep = func(context.Context, time.Duration) { sleeps++ }
	e.cons.MaxAttempts = 4
	_ = e.cons.Handle(ctx, msgFor(c.ID))
	if sleeps != 3 { // sleeps between attempts only
		t.Fatalf("sleeps=%d want 3", sleeps)
	}
}

func TestConsumerBadPayloadGoesStraightToDLQ(t *testing.T) {
	e := newEnv()
	m := port.Message{Topic: port.TopicSubmitted, Key: "x", Value: []byte("not json")}
	if err := e.cons.Handle(ctx, m); err != nil {
		t.Fatal(err)
	}
	if len(e.pub.On(port.TopicDLQ)) != 1 {
		t.Fatal("bad payload should be parked")
	}
}

func TestConsumerUnknownClaimIsPermanent(t *testing.T) {
	e := newEnv()
	sleeps := 0
	e.cons.Sleep = func(context.Context, time.Duration) { sleeps++ }
	if err := e.cons.Handle(ctx, msgFor("00000000-0000-4000-8000-0000000000ff")); err != nil {
		t.Fatal(err)
	}
	if sleeps != 0 || len(e.pub.On(port.TopicDLQ)) != 1 {
		t.Fatalf("permanent errors must not be retried: sleeps=%d", sleeps)
	}
}

func TestConsumerDLQPublishFailureDoesNotCommit(t *testing.T) {
	e := newEnv()
	c := submit(t, e, "k1", domain.PoisonRef, 1000)
	e.pub.FailTimes, e.pub.FailTopic = 1, port.TopicDLQ
	if err := e.cons.Handle(ctx, msgFor(c.ID)); err == nil {
		t.Fatal("must return error so the offset is not committed")
	}
}

func TestConsumerStopsRetryingWhenShuttingDown(t *testing.T) {
	e := newEnv()
	c := submit(t, e, "k1", domain.PoisonRef, 1000)
	cctx, cancel := context.WithCancel(ctx)
	cancel()
	err := e.cons.Handle(cctx, msgFor(c.ID))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
	if len(e.pub.On(port.TopicDLQ)) != 0 {
		t.Fatal("shutdown must not dead-letter the message")
	}
}

// ---- read side + cache ----

func TestGetUsesCacheAndInvalidatesOnProcess(t *testing.T) {
	e := newEnv()
	c := submit(t, e, "k1", "FAKE-AB12", 1000)
	got, _ := e.svc.Get(ctx, c.ID) // miss -> fill
	if got.Status != domain.StatusPending || e.cache.Misses != 1 {
		t.Fatalf("misses=%d", e.cache.Misses)
	}
	_, _ = e.svc.Get(ctx, c.ID) // hit
	if e.cache.Hits != 1 {
		t.Fatalf("hits=%d", e.cache.Hits)
	}
	_ = e.svc.Process(ctx, c.ID)
	got, _ = e.svc.Get(ctx, c.ID)
	if got.Status != domain.StatusApproved {
		t.Fatalf("stale cache after process: %s", got.Status)
	}
}

func TestCacheFailureFallsBackToDB(t *testing.T) {
	e := newEnv()
	c := submit(t, e, "k1", "FAKE-AB12", 1000)
	e.cache.Fail = true
	got, err := e.svc.Get(ctx, c.ID)
	if err != nil || got.ID != c.ID {
		t.Fatalf("cache outage must not break reads: %v", err)
	}
	if err := e.svc.Process(ctx, c.ID); err != nil {
		t.Fatalf("cache outage must not break processing: %v", err)
	}
}

func TestNilCacheIsAllowed(t *testing.T) {
	repo := memory.NewRepo()
	svc := NewClaims(repo, nil, quiet)
	c, _, err := svc.Submit(ctx, SubmitInput{"k", "FAKE-AB12", 5})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Get(ctx, c.ID); err != nil {
		t.Fatal(err)
	}
}

func TestGetValidation(t *testing.T) {
	e := newEnv()
	if _, err := e.svc.Get(ctx, "nope"); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("got %v", err)
	}
	if _, err := e.svc.Get(ctx, "00000000-0000-4000-8000-0000000000ff"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("got %v", err)
	}
}

func TestListFilterAndPagination(t *testing.T) {
	e := newEnv()
	var ids []string
	for i := 0; i < 5; i++ {
		c := submit(t, e, fmt.Sprintf("k%d", i), fmt.Sprintf("FAKE-AB%d", i), 1000)
		ids = append(ids, c.ID)
	}
	_ = e.svc.Process(ctx, ids[0])
	_ = e.svc.Process(ctx, ids[1])

	approved := domain.StatusApproved
	got, err := e.svc.List(ctx, &approved, 10, 0)
	if err != nil || len(got) != 2 {
		t.Fatalf("approved=%d err=%v", len(got), err)
	}
	pending := domain.StatusPending
	p1, _ := e.svc.List(ctx, &pending, 2, 0)
	p2, _ := e.svc.List(ctx, &pending, 2, 2)
	if len(p1) != 2 || len(p2) != 1 || p1[0].ID == p2[0].ID {
		t.Fatalf("pagination wrong: %d %d", len(p1), len(p2))
	}
	all, _ := e.svc.List(ctx, nil, 0, 0)
	if len(all) != 5 {
		t.Fatalf("default list=%d", len(all))
	}
}

func TestListEmptyIsNotNil(t *testing.T) {
	e := newEnv()
	got, err := e.svc.List(ctx, nil, 10, 0)
	if err != nil || got == nil || len(got) != 0 {
		t.Fatalf("want empty non-nil slice, got %#v %v", got, err)
	}
}

func TestListBounds(t *testing.T) {
	e := newEnv()
	for _, c := range [][2]int{{-1, 0}, {MaxLimit + 1, 0}, {10, -1}} {
		if _, err := e.svc.List(ctx, nil, c[0], c[1]); !errors.Is(err, domain.ErrInvalid) {
			t.Errorf("limit=%d offset=%d should be invalid, got %v", c[0], c[1], err)
		}
	}
}
