package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/sitthichok-J/self-study-2026/day12-auth-hardening/internal/obs"
	"github.com/sitthichok-J/self-study-2026/day12-auth-hardening/internal/port"
)

// Consumer turns a Kafka message into a Process call with bounded retry and a
// dead-letter topic. Handle returns nil once the message is "done" (processed
// or parked in the DLQ), which is the signal for the adapter to commit.
type Consumer struct {
	claims      *Claims
	pub         port.EventPublisher
	log         *slog.Logger
	MaxAttempts int
	Backoff     func(attempt int) time.Duration
	Sleep       func(context.Context, time.Duration)
}

func NewConsumer(c *Claims, pub port.EventPublisher, log *slog.Logger) *Consumer {
	if log == nil {
		log = slog.Default()
	}
	return &Consumer{
		claims: c, pub: pub, log: log, MaxAttempts: 3,
		Backoff: func(a int) time.Duration { return time.Duration(a) * 500 * time.Millisecond },
		Sleep: func(ctx context.Context, d time.Duration) {
			t := time.NewTimer(d)
			defer t.Stop()
			select {
			case <-ctx.Done():
			case <-t.C:
			}
		},
	}
}

func (c *Consumer) Handle(ctx context.Context, m port.Message) error {
	var ev SubmitEvent
	bad := json.Unmarshal(m.Value, &ev) != nil || ev.ClaimID == ""
	// Continue the trace of the request that created the claim (if the
	// producer put a valid traceparent in the payload).
	if sc, ok := obs.ParseTraceparent(ev.Traceparent); ok && !bad {
		ctx = obs.ContextWithRemote(ctx, sc)
	}
	ctx, sp := obs.StartKind(ctx, "consume "+m.Topic, obs.KindConsumer, obs.Attr{Key: "messaging.destination.name", Val: m.Topic})
	defer sp.End()

	if bad {
		sp.MarkError("bad_payload")
		return c.deadLetter(ctx, m, "bad_payload")
	}
	var lastErr error
	for attempt := 1; attempt <= c.MaxAttempts; attempt++ {
		lastErr = c.claims.Process(ctx, ev.ClaimID)
		if lastErr == nil {
			obs.M().ConsumerOutcome.Inc("processed")
			return nil
		}
		if IsPermanent(lastErr) {
			sp.SetError(lastErr)
			return c.deadLetter(ctx, m, lastErr.Error())
		}
		c.log.WarnContext(ctx, "process failed", "claim_id", ev.ClaimID, "attempt", attempt, "err", lastErr)
		if ctx.Err() != nil {
			return ctx.Err() // shutting down: do not commit, message is redelivered
		}
		if attempt < c.MaxAttempts {
			obs.M().ConsumerOutcome.Inc("retried")
			c.Sleep(ctx, c.Backoff(attempt))
		}
	}
	sp.SetError(lastErr)
	return c.deadLetter(ctx, m, fmt.Sprintf("max_attempts: %v", lastErr))
}

type deadLetterEnvelope struct {
	Reason        string          `json:"reason"`
	OriginalTopic string          `json:"original_topic"`
	OriginalKey   string          `json:"original_key"`
	Value         json.RawMessage `json:"value"`
	FailedAt      time.Time       `json:"failed_at"`
}

func (c *Consumer) deadLetter(ctx context.Context, m port.Message, reason string) error {
	val := json.RawMessage(m.Value)
	if !json.Valid(m.Value) {
		val, _ = json.Marshal(string(m.Value))
	}
	b, _ := json.Marshal(deadLetterEnvelope{
		Reason: reason, OriginalTopic: m.Topic, OriginalKey: m.Key, Value: val, FailedAt: c.claims.now().UTC(),
	})
	if err := c.pub.Publish(ctx, port.TopicDLQ, m.Key, b); err != nil {
		return errors.Join(errors.New("publish to dlq failed"), err) // not done: keep message
	}
	obs.M().ConsumerOutcome.Inc("dlq")
	c.log.ErrorContext(ctx, "message sent to dlq", "key", m.Key, "reason", reason)
	return nil
}
