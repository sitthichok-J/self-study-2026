package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/sitthichok-J/self-study-2026/day08-backend/internal/port"
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
	if err := json.Unmarshal(m.Value, &ev); err != nil || ev.ClaimID == "" {
		return c.deadLetter(ctx, m, "bad_payload")
	}
	var lastErr error
	for attempt := 1; attempt <= c.MaxAttempts; attempt++ {
		lastErr = c.claims.Process(ctx, ev.ClaimID)
		if lastErr == nil {
			return nil
		}
		if IsPermanent(lastErr) {
			return c.deadLetter(ctx, m, lastErr.Error())
		}
		c.log.Warn("process failed", "claim_id", ev.ClaimID, "attempt", attempt, "err", lastErr)
		if ctx.Err() != nil {
			return ctx.Err() // shutting down: do not commit, message is redelivered
		}
		if attempt < c.MaxAttempts {
			c.Sleep(ctx, c.Backoff(attempt))
		}
	}
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
	c.log.Error("message sent to dlq", "key", m.Key, "reason", reason)
	return nil
}
