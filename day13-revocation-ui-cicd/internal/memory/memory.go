// Package memory has in-memory fakes of the ports, used by unit tests.
package memory

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/domain"
	"github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/port"
)

type Repo struct {
	mu     sync.Mutex
	claims map[string]domain.Claim
	byKey  map[string]string
	outbox []outboxRow
	nextID int64
	// FailGet makes Get return this error (to simulate DB trouble).
	FailGet error
}

type outboxRow struct {
	ev   port.OutboxEvent
	sent bool
}

func NewRepo() *Repo {
	return &Repo{claims: map[string]domain.Claim{}, byKey: map[string]string{}}
}

func (r *Repo) Create(_ context.Context, c domain.Claim, ev port.OutboxEvent) (domain.Claim, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	scoped := c.OwnerID + "\x00" + c.IdempotencyKey
	if id, ok := r.byKey[scoped]; ok {
		ex := r.claims[id]
		if ex.RequestHash != c.RequestHash {
			return domain.Claim{}, false, domain.ErrIdempotencyConflict
		}
		return ex, false, nil
	}
	r.claims[c.ID] = c
	r.byKey[scoped] = c.ID
	r.nextID++
	ev.ID = r.nextID
	r.outbox = append(r.outbox, outboxRow{ev: ev})
	return c, true, nil
}

func (r *Repo) Get(_ context.Context, id string) (domain.Claim, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.FailGet != nil {
		return domain.Claim{}, r.FailGet
	}
	c, ok := r.claims[id]
	if !ok {
		return domain.Claim{}, domain.ErrNotFound
	}
	return c, nil
}

func (r *Repo) List(_ context.Context, owner string, status *domain.Status, limit, offset int) ([]domain.Claim, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var all []domain.Claim
	for _, c := range r.claims {
		if (owner == "" || c.OwnerID == owner) && (status == nil || c.Status == *status) {
			all = append(all, c)
		}
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].CreatedAt.Equal(all[j].CreatedAt) {
			return all[i].ID > all[j].ID
		}
		return all[i].CreatedAt.After(all[j].CreatedAt)
	})
	if offset >= len(all) {
		return nil, nil
	}
	all = all[offset:]
	if len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}

func (r *Repo) UpdateStatus(_ context.Context, id string, from, to domain.Status, reason string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.claims[id]
	if !ok {
		return false, domain.ErrNotFound
	}
	if c.Status != from {
		return false, nil
	}
	c.Status, c.DecisionReason, c.UpdatedAt = to, reason, time.Now().UTC()
	r.claims[id] = c
	return true, nil
}

func (r *Repo) ProcessOutbox(_ context.Context, limit int, publish func(port.OutboxEvent) error) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	n := 0
	for i := range r.outbox {
		if n >= limit {
			break
		}
		if r.outbox[i].sent {
			continue
		}
		if err := publish(r.outbox[i].ev); err != nil {
			return n, err
		}
		r.outbox[i].sent = true
		n++
	}
	return n, nil
}

// OutboxPending counts unsent rows.
func (r *Repo) OutboxPending() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	n := 0
	for _, o := range r.outbox {
		if !o.sent {
			n++
		}
	}
	return n
}

func (r *Repo) Count() int { r.mu.Lock(); defer r.mu.Unlock(); return len(r.claims) }

// Publisher records messages. FailTimes makes the next N Publish calls fail;
// FailTopic restricts failures to one topic ("" = any).
type Publisher struct {
	mu        sync.Mutex
	Sent      []port.Message
	FailTimes int
	FailTopic string
}

func (p *Publisher) Publish(_ context.Context, topic, key string, value []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.FailTimes > 0 && (p.FailTopic == "" || p.FailTopic == topic) {
		p.FailTimes--
		return errors.New("broker unavailable")
	}
	p.Sent = append(p.Sent, port.Message{Topic: topic, Key: key, Value: value})
	return nil
}

func (p *Publisher) On(topic string) []port.Message {
	p.mu.Lock()
	defer p.mu.Unlock()
	var out []port.Message
	for _, m := range p.Sent {
		if m.Topic == topic {
			out = append(out, m)
		}
	}
	return out
}

type Cache struct {
	mu     sync.Mutex
	M      map[string][]byte
	Hits   int
	Misses int
	Fail   bool // every call errors, to prove cache is best effort
}

func NewCache() *Cache { return &Cache{M: map[string][]byte{}} }

func (c *Cache) Get(_ context.Context, k string) ([]byte, bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Fail {
		return nil, false, errors.New("cache down")
	}
	v, ok := c.M[k]
	if ok {
		c.Hits++
	} else {
		c.Misses++
	}
	return v, ok, nil
}

func (c *Cache) Set(_ context.Context, k string, v []byte, _ time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Fail {
		return errors.New("cache down")
	}
	c.M[k] = v
	return nil
}

func (c *Cache) Delete(_ context.Context, ks ...string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Fail {
		return errors.New("cache down")
	}
	for _, k := range ks {
		delete(c.M, k)
	}
	return nil
}
