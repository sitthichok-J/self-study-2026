package service

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/domain"
	"github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/memory"
	"github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/obs"
	"github.com/sitthichok-J/self-study-2026/day10-observability-security/internal/port"
)

// instrument swaps in fresh metrics and an in-memory tracer for one test.
func instrument(t *testing.T) (*obs.Metrics, *obs.MemExporter) {
	t.Helper()
	m := obs.NewMetrics()
	mem := &obs.MemExporter{}
	oldM := obs.SetMetrics(m)
	oldT := obs.SetTracer(obs.NewTracer("test", mem))
	t.Cleanup(func() { obs.SetMetrics(oldM); obs.SetTracer(oldT) })
	return m, mem
}

// ---- authorization rules in the service ----

func TestGetEnforcesOwnerAndStaff(t *testing.T) {
	e := newEnv()
	c := submit(t, e, "k", "FAKE-AB12", 100)
	if _, err := e.svc.Get(ctx, domain.Viewer{ID: "mallory"}, c.ID); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("other user: %v", err)
	}
	if _, err := e.svc.Get(ctx, domain.Viewer{}, c.ID); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("anonymous viewer must be refused: %v", err)
	}
	if _, err := e.svc.Get(ctx, domain.Viewer{ID: "ops", Staff: true}, c.ID); err != nil {
		t.Fatalf("staff: %v", err)
	}
	// second read is served from cache; the check must still apply
	_, _ = e.svc.Get(ctx, alice, c.ID)
	if _, err := e.svc.Get(ctx, domain.Viewer{ID: "mallory"}, c.ID); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("cached read bypassed the check: %v", err)
	}
}

func TestListWithoutIdentityFailsClosed(t *testing.T) {
	e := newEnv()
	submit(t, e, "k", "FAKE-AB12", 100)
	if _, err := e.svc.List(ctx, domain.Viewer{}, nil, 10, 0); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("got %v", err)
	}
}

func TestSubmitNeedsAnOwner(t *testing.T) {
	e := newEnv()
	_, _, err := e.svc.Submit(ctx, SubmitInput{IdempotencyKey: "k", PatientRef: "FAKE-AB12", Amount: 1})
	if !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("got %v", err)
	}
}

func TestIdempotencyIsPerOwnerInRepo(t *testing.T) {
	e := newEnv()
	a, _, _ := e.svc.Submit(ctx, SubmitInput{Owner: "alice", IdempotencyKey: "k", PatientRef: "FAKE-AB12", Amount: 1})
	b, created, err := e.svc.Submit(ctx, SubmitInput{Owner: "bob", IdempotencyKey: "k", PatientRef: "FAKE-AB12", Amount: 1})
	if err != nil || !created || a.ID == b.ID {
		t.Fatalf("bob must get his own claim: %v %v", err, created)
	}
	// alice replaying her own key still works
	again, created, _ := e.svc.Submit(ctx, SubmitInput{Owner: "alice", IdempotencyKey: "k", PatientRef: "FAKE-AB12", Amount: 1})
	if created || again.ID != a.ID {
		t.Fatal("own replay must return the original")
	}
}

func TestMemoryListOwnerFilter(t *testing.T) {
	e := newEnv()
	e.svc.Submit(ctx, SubmitInput{Owner: "alice", IdempotencyKey: "1", PatientRef: "FAKE-AB12", Amount: 1})
	e.svc.Submit(ctx, SubmitInput{Owner: "bob", IdempotencyKey: "2", PatientRef: "FAKE-AB12", Amount: 1})
	for owner, want := range map[string]int{"alice": 1, "bob": 1, "nobody": 0, "": 2} {
		got, _ := e.repo.List(ctx, owner, nil, 10, 0)
		if len(got) != want {
			t.Errorf("owner %q: %d claims, want %d", owner, len(got), want)
		}
	}
}

// ---- metrics ----

func TestBusinessMetrics(t *testing.T) {
	m, _ := instrument(t)
	e := newEnv()
	submit(t, e, "k1", "FAKE-AB12", 100)
	e.svc.Submit(ctx, SubmitInput{Owner: "alice", IdempotencyKey: "k1", PatientRef: "FAKE-AB12", Amount: 100}) // replay
	big := submit(t, e, "k2", "FAKE-AB12", domain.OverLimit+1)
	if m.ClaimsSubmitted.Value("created") != 2 || m.ClaimsSubmitted.Value("replayed") != 1 {
		t.Fatalf("submit counters wrong")
	}
	e.relay.RunOnce(ctx)
	if m.OutboxPublished.Value("ok") != 2 {
		t.Fatal("outbox ok counter")
	}
	e.pub.FailTimes = 1
	submit(t, e, "k3", "FAKE-AB12", 1)
	e.relay.RunOnce(ctx)
	if m.OutboxPublished.Value("error") != 1 {
		t.Fatal("outbox error counter")
	}
	for _, msg := range e.pub.On(port.TopicSubmitted) {
		if err := e.cons.Handle(ctx, msg); err != nil {
			t.Fatal(err)
		}
	}
	if m.ClaimsDecided.Value("APPROVED", "ok") != 1 || m.ClaimsDecided.Value("REJECTED", "over_limit") != 1 {
		t.Fatalf("decision counters wrong (big=%s)", big.ID)
	}
	if m.ConsumerOutcome.Value("processed") != 2 {
		t.Fatal("processed counter")
	}
}

func TestConsumerRetryAndDLQCounters(t *testing.T) {
	m, _ := instrument(t)
	e := newEnv()
	c := submit(t, e, "k", domain.PoisonRef, 1)
	if err := e.cons.Handle(ctx, msgFor(c.ID)); err != nil {
		t.Fatal(err)
	}
	// MaxAttempts=3: two retries, then the DLQ
	if m.ConsumerOutcome.Value("retried") != 2 || m.ConsumerOutcome.Value("dlq") != 1 || m.ConsumerOutcome.Value("processed") != 0 {
		t.Fatalf("retried=%v dlq=%v", m.ConsumerOutcome.Value("retried"), m.ConsumerOutcome.Value("dlq"))
	}
	if err := e.cons.Handle(ctx, port.Message{Topic: port.TopicSubmitted, Key: "x", Value: []byte("not json")}); err != nil {
		t.Fatal(err)
	}
	if m.ConsumerOutcome.Value("dlq") != 2 {
		t.Fatal("bad payload must be counted as dlq")
	}
}

// ---- tracing across the Kafka hop ----

func TestTraceContinuesFromRequestThroughOutboxToWorker(t *testing.T) {
	_, mem := instrument(t)
	e := newEnv()
	reqCtx, root := obs.StartKind(ctx, "POST /claims", obs.KindServer)
	c, _, err := e.svc.Submit(reqCtx, SubmitInput{Owner: "alice", IdempotencyKey: "k", PatientRef: "FAKE-AB12", Amount: 100})
	if err != nil {
		t.Fatal(err)
	}
	root.End()
	e.relay.RunOnce(ctx) // the relay runs on its own goroutine in prod: no request context
	msgs := e.pub.On(port.TopicSubmitted)
	if len(msgs) != 1 {
		t.Fatal("no message")
	}
	var ev SubmitEvent
	_ = json.Unmarshal(msgs[0].Value, &ev)
	sc, ok := obs.ParseTraceparent(ev.Traceparent)
	if !ok || sc.TraceID != root.Context().TraceID {
		t.Fatalf("payload traceparent %q does not belong to the request trace", ev.Traceparent)
	}
	if err := e.cons.Handle(context.Background(), msgs[0]); err != nil {
		t.Fatal(err)
	}
	var consume *obs.SpanData
	for _, s := range mem.Spans() {
		s := s
		if strings.HasPrefix(s.Name, "consume ") {
			consume = &s
		}
	}
	if consume == nil || consume.Context.TraceID != root.Context().TraceID || consume.Kind != obs.KindConsumer {
		t.Fatalf("worker span is not part of the request trace: %+v", consume)
	}
	_ = c
}

func TestOutboxPublishSpanJoinsRequestTrace(t *testing.T) {
	_, mem := instrument(t)
	e := newEnv()
	reqCtx, root := obs.StartKind(ctx, "POST /claims", obs.KindServer)
	if _, _, err := e.svc.Submit(reqCtx, SubmitInput{Owner: "alice", IdempotencyKey: "k", PatientRef: "FAKE-AB12", Amount: 100}); err != nil {
		t.Fatal(err)
	}
	root.End()
	e.relay.RunOnce(ctx)
	var pub *obs.SpanData
	for _, s := range mem.Spans() {
		s := s
		if strings.HasPrefix(s.Name, "outbox publish ") {
			pub = &s
		}
	}
	if pub == nil || pub.Context.TraceID != root.Context().TraceID || pub.Kind != obs.KindProducer {
		t.Fatalf("outbox publish span is not in the request trace: %+v", pub)
	}
}

func TestConsumerIgnoresForgedTraceparent(t *testing.T) {
	_, mem := instrument(t)
	e := newEnv()
	c := submit(t, e, "k", "FAKE-AB12", 1)
	b, _ := json.Marshal(SubmitEvent{Event: port.TopicSubmitted, ClaimID: c.ID, Traceparent: "00-ZZZ-garbage"})
	if err := e.cons.Handle(ctx, port.Message{Topic: port.TopicSubmitted, Key: c.ID, Value: b}); err != nil {
		t.Fatal(err)
	}
	if len(mem.Spans()) == 0 {
		t.Fatal("a span is still created with a fresh trace")
	}
}

// ---- exercise 3: slow it down on purpose, find it with the trace ----

type slowRepo struct {
	port.ClaimRepository
	delay time.Duration
}

func (s slowRepo) Get(ctx context.Context, id string) (domain.Claim, error) {
	time.Sleep(s.delay)
	return s.ClaimRepository.Get(ctx, id)
}

func TestSlowRepoShowsUpAsTheLongestSpan(t *testing.T) {
	_, mem := instrument(t)
	base := memory.NewRepo()
	svc := NewClaims(slowRepo{base, 60 * time.Millisecond}, nil, quiet)
	c, _, _ := svc.Submit(ctx, SubmitInput{Owner: "alice", IdempotencyKey: "k", PatientRef: "FAKE-AB12", Amount: 1})

	rctx, root := obs.StartKind(ctx, "GET /claims/{id}", obs.KindServer)
	if _, err := svc.Get(rctx, alice, c.ID); err != nil {
		t.Fatal(err)
	}
	root.End()

	var inTrace []obs.SpanData
	for _, s := range mem.Spans() {
		if s.Context.TraceID == root.Context().TraceID {
			inTrace = append(inTrace, s)
		}
	}
	// the slowest *child* is where the time went
	var kids []obs.SpanData
	for _, s := range inTrace {
		if s.Name != "GET /claims/{id}" {
			kids = append(kids, s)
		}
	}
	sort.Slice(kids, func(i, j int) bool { return kids[i].Duration() > kids[j].Duration() })
	if len(kids) == 0 || kids[0].Name != "repo.Get" || kids[0].Duration() < 50*time.Millisecond {
		t.Fatalf("trace did not point at repo.Get: %+v", kids)
	}
}

func TestRepoErrorMarksSpan(t *testing.T) {
	_, mem := instrument(t)
	e := newEnv()
	e.repo.FailGet = errors.New("db down")
	_, _ = e.svc.Get(ctx, alice, "00000000-0000-4000-8000-0000000000ff")
	var found bool
	for _, s := range mem.Spans() {
		if s.Name == "repo.Get" && s.ErrMsg == "db down" {
			found = true
		}
	}
	if !found {
		t.Fatal("failed repo call must produce an error span")
	}
}

func TestEventPayloadStaysSmallAndHasNoPersonalData(t *testing.T) {
	instrument(t)
	e := newEnv()
	submit(t, e, "k", "FAKE-AB12", 100)
	e.relay.RunOnce(ctx)
	raw := string(e.pub.On(port.TopicSubmitted)[0].Value)
	if strings.Contains(raw, "FAKE-AB12") || strings.Contains(raw, "alice") {
		t.Fatalf("event must carry ids only: %s", raw)
	}
}
