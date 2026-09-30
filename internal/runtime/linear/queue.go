package linear

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	"butterfly.orx.me/core/log"
	"go.orx.me/apps/butter/internal/eventqueue"
)

const (
	// StreamKey holds every accepted Linear delivery. It is separate from
	// Telegram's so a Telegram backlog can never push a created session
	// past Linear's 10-second deadline.
	StreamKey = "butter:linear:events"
	// ConsumerGroup is the consumer group every worker Pod joins.
	ConsumerGroup = "butter-linear-workers"

	dedupeKeyPrefix         = "butter:linear:seen:"
	dedupeTTL               = 24 * time.Hour
	leasePreflightKeyPrefix = "butter:linear:lease:preflight:"
)

// ErrDuplicate reports that the delivery was already accepted.
var ErrDuplicate = eventqueue.ErrDuplicate

// Delivery is one claimed Stream entry.
type Delivery struct {
	ID    string
	Event *Event
}

// Queue is the Linear receive hand-off over the shared durable event queue.
type Queue struct {
	inner *eventqueue.Queue
}

// NewQueue builds the queue, or returns nil without Redis.
func NewQueue(rdb *redis.Client) *Queue {
	if rdb == nil {
		return nil
	}
	return &Queue{inner: eventqueue.New(rdb, eventqueue.Config{
		Name:                    "linear",
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

// Available reports whether a durable queue is wired at all.
func (q *Queue) Available() bool { return q.queue().Available() }

// Ping, CheckReady and CheckLeaseReady back the enablement probe.
func (q *Queue) Ping(ctx context.Context) error            { return q.queue().Ping(ctx) }
func (q *Queue) CheckReady(ctx context.Context) error      { return q.queue().CheckReady(ctx) }
func (q *Queue) CheckLeaseReady(ctx context.Context) error { return q.queue().CheckLeaseReady(ctx) }

// Accept durably records an event under its delivery ID.
func (q *Queue) Accept(ctx context.Context, event *Event) (string, error) {
	payload, err := event.Encode()
	if err != nil {
		return "", err
	}
	return q.queue().Accept(ctx, event.AppID+":"+event.DeliveryID, payload)
}

func (q *Queue) EnsureGroup(ctx context.Context) error { return q.queue().EnsureGroup(ctx) }

func (q *Queue) Read(ctx context.Context, consumer string, count int64, block time.Duration) ([]Delivery, error) {
	entries, err := q.queue().Read(ctx, consumer, count, block)
	if err != nil {
		return nil, err
	}
	return q.decode(ctx, entries), nil
}

func (q *Queue) ReadPending(ctx context.Context, consumer string, count int64) ([]Delivery, error) {
	entries, err := q.queue().ReadPending(ctx, consumer, count)
	if err != nil {
		return nil, err
	}
	return q.decode(ctx, entries), nil
}

func (q *Queue) Claim(ctx context.Context, consumer string, minIdle time.Duration, count int64) ([]Delivery, error) {
	entries, err := q.queue().Claim(ctx, consumer, minIdle, count)
	if err != nil {
		return nil, err
	}
	return q.decode(ctx, entries), nil
}

func (q *Queue) Touch(ctx context.Context, consumer, id string) error {
	return q.queue().Touch(ctx, consumer, id)
}

func (q *Queue) Ack(ctx context.Context, ids ...string) error { return q.queue().Ack(ctx, ids...) }

// decode drops (and acknowledges) entries that can never be processed,
// reporting each.
func (q *Queue) decode(ctx context.Context, entries []eventqueue.Entry) []Delivery {
	out := make([]Delivery, 0, len(entries))
	for _, entry := range entries {
		event, err := DecodeEvent(entry.Payload)
		if err != nil {
			log.FromContext(ctx).Error("dropping undecodable linear event", "stream_id", entry.ID, "err", err)
			_ = q.Ack(ctx, entry.ID)
			continue
		}
		out = append(out, Delivery{ID: entry.ID, Event: event})
	}
	return out
}
