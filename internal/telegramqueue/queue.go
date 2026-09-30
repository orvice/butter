package telegramqueue

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"go.orx.me/apps/butter/internal/eventqueue"
)

const (
	// StreamKey holds every accepted Telegram update. One stream, not one
	// per Channel: a consumer group over a single stream is what lets any
	// Pod pick up any Channel's work, which is the whole point of the
	// multi-Pod design.
	StreamKey = "butter:telegram:updates"
	// ConsumerGroup is the shared consumer group all worker Pods join.
	ConsumerGroup = "butter-telegram-workers"

	// dedupeKeyPrefix marks an (channel, update) pair as already accepted.
	dedupeKeyPrefix = "butter:telegram:seen:"
	// dedupeTTL bounds how long a duplicate is recognized. Telegram retries
	// an undelivered update for far less than this.
	dedupeTTL = 24 * time.Hour
	// leasePreflightKeyPrefix namespaces CheckLeaseReady's probe lease.
	leasePreflightKeyPrefix = "butter:telegram:lease:preflight:"
)

// errNotConfigured is what every operation reports on a nil queue.
var errNotConfigured = errors.New("telegram queue is not configured")

// ErrDuplicate reports that this (channel, update) pair was already accepted.
// It is not a failure: the caller acknowledges the delivery.
var ErrDuplicate = eventqueue.ErrDuplicate

// Queue is the Telegram receive hand-off: the shared durable event queue
// (internal/eventqueue) under Telegram's keys, carrying encoded Events.
type Queue struct {
	inner *eventqueue.Queue
}

func New(rdb *redis.Client) *Queue {
	if rdb == nil {
		return nil
	}
	return &Queue{inner: eventqueue.New(rdb, eventqueue.Config{
		Name:                    "telegram",
		StreamKey:               StreamKey,
		ConsumerGroup:           ConsumerGroup,
		DedupeKeyPrefix:         dedupeKeyPrefix,
		DedupeTTL:               dedupeTTL,
		LeasePreflightKeyPrefix: leasePreflightKeyPrefix,
	})}
}

func (q *Queue) queue() *eventqueue.Queue {
	if q == nil {
		return nil
	}
	return q.inner
}

// Available reports whether a durable queue is wired at all. Callers use it
// to block enablement rather than discovering the gap at the first update.
func (q *Queue) Available() bool { return q.queue().Available() }

// Ping verifies that the queue backend is reachable without imposing the
// persistence-policy checks required before accepting webhook deliveries.
func (q *Queue) Ping(ctx context.Context) error {
	if !q.Available() {
		return errNotConfigured
	}
	return q.inner.Ping(ctx)
}

// Accept durably records an event, returning ErrDuplicate when this
// (channel, update) pair was already taken. The returned ID is the Stream
// entry ID.
//
// The caller must treat a non-nil error other than ErrDuplicate as
// "not accepted" and answer Telegram with a retryable status: acknowledging
// an update we failed to enqueue loses it permanently.
func (q *Queue) Accept(ctx context.Context, event *Event) (string, error) {
	if !q.Available() {
		return "", errNotConfigured
	}
	payload, err := event.Encode()
	if err != nil {
		return "", err
	}
	return q.inner.Accept(ctx, fmt.Sprintf("%s:%d", event.ChannelID, event.UpdateID), payload)
}

// EnsureGroup creates the consumer group if it does not exist. It is safe to
// call from every Pod on every start.
func (q *Queue) EnsureGroup(ctx context.Context) error {
	if !q.Available() {
		return errNotConfigured
	}
	return q.inner.EnsureGroup(ctx)
}

// Delivery is one claimed Stream entry.
type Delivery struct {
	// ID is the Stream entry ID, used to acknowledge.
	ID    string
	Event *Event
}

// Read claims up to count new entries for this consumer, blocking up to
// block for work to arrive.
func (q *Queue) Read(ctx context.Context, consumer string, count int64, block time.Duration) ([]Delivery, error) {
	if !q.Available() {
		return nil, errNotConfigured
	}
	entries, err := q.inner.Read(ctx, consumer, count, block)
	if err != nil {
		return nil, err
	}
	return q.decodeAll(ctx, entries)
}

// ReadPending re-claims entries this consumer already holds but never
// acknowledged — the state a Pod is in after a crash mid-turn.
func (q *Queue) ReadPending(ctx context.Context, consumer string, count int64) ([]Delivery, error) {
	if !q.Available() {
		return nil, errNotConfigured
	}
	entries, err := q.inner.ReadPending(ctx, consumer, count)
	if err != nil {
		return nil, err
	}
	return q.decodeAll(ctx, entries)
}

// decodeAll decodes read entries in order. An undecodable entry can never
// succeed: it is acknowledged so it stops being redelivered, and reported.
func (q *Queue) decodeAll(ctx context.Context, entries []eventqueue.Entry) ([]Delivery, error) {
	var out []Delivery
	for _, entry := range entries {
		event, decodeErr := DecodeEvent(entry.Payload)
		if decodeErr != nil {
			_ = q.Ack(ctx, entry.ID)
			return out, fmt.Errorf("drop undecodable telegram event %s: %w", entry.ID, decodeErr)
		}
		out = append(out, Delivery{ID: entry.ID, Event: event})
	}
	return out, nil
}

// Ack marks entries as fully handled.
func (q *Queue) Ack(ctx context.Context, ids ...string) error { return q.queue().Ack(ctx, ids...) }

// Touch resets one pending entry's idle time while its consumer is still
// handling it. Without this heartbeat, XAUTOCLAIM can steal a healthy long
// Agent turn merely because it exceeded reclaimIdle.
func (q *Queue) Touch(ctx context.Context, consumer, id string) error {
	if !q.Available() {
		return errNotConfigured
	}
	return q.inner.Touch(ctx, consumer, id)
}

// Claim takes over entries idle longer than minIdle from whichever consumer
// holds them, so a crashed Pod's work does not stall.
func (q *Queue) Claim(ctx context.Context, consumer string, minIdle time.Duration, count int64) ([]Delivery, error) {
	if !q.Available() {
		return nil, errNotConfigured
	}
	entries, err := q.inner.Claim(ctx, consumer, minIdle, count)
	if err != nil {
		return nil, err
	}
	var out []Delivery
	for _, entry := range entries {
		event, decodeErr := DecodeEvent(entry.Payload)
		if decodeErr != nil {
			_ = q.Ack(ctx, entry.ID)
			continue
		}
		out = append(out, Delivery{ID: entry.ID, Event: event})
	}
	return out, nil
}

// CheckReady verifies that Redis can be used as a durable queue. Reachability
// alone is insufficient: evicting or non-persistent Redis can acknowledge a
// Telegram update and then discard the only authoritative copy.
func (q *Queue) CheckReady(ctx context.Context) error {
	if !q.Available() {
		return errNotConfigured
	}
	return q.inner.CheckReady(ctx)
}

// CheckLeaseReady verifies the Redis commands and Lua execution used by the
// renewable leadership lease. Long Polling must reject enablement if the queue
// is durable but the configured Redis ACL cannot acquire, renew, or release a
// consumer lease.
func (q *Queue) CheckLeaseReady(ctx context.Context) error {
	if !q.Available() {
		return errNotConfigured
	}
	return q.inner.CheckLeaseReady(ctx)
}
