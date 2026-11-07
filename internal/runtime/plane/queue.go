package plane

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"go.orx.me/apps/butter/internal/eventqueue"
	"time"
)

const (
	StreamKey     = "butter:plane:events"
	ConsumerGroup = "butter-plane-workers"
)

type StreamQueue struct{ inner *eventqueue.Queue }

func NewQueue(rdb *redis.Client) *StreamQueue {
	if rdb == nil {
		return nil
	}
	return &StreamQueue{inner: eventqueue.New(rdb, eventqueue.Config{Name: "plane", StreamKey: StreamKey, ConsumerGroup: ConsumerGroup, DedupeKeyPrefix: "butter:plane:seen:", DedupeTTL: 24 * time.Hour, LeasePreflightKeyPrefix: "butter:plane:lease:preflight:"})}
}
func (q *StreamQueue) Accept(ctx context.Context, e *Event) (string, error) {
	if q == nil || q.inner == nil {
		return "", context.Canceled
	}
	p, err := e.Encode()
	if err != nil {
		return "", err
	}
	id := e.InstallationID + ":" + e.AgentRunID + ":" + e.ActivityID + ":" + e.Action
	if _, err := q.inner.Accept(ctx, id, p); err != nil {
		if errors.Is(err, eventqueue.ErrDuplicate) {
			return "", ErrDuplicate
		}
		return "", err
	}
	return id, nil
}
func (q *StreamQueue) Available() bool                       { return q != nil && q.inner != nil && q.inner.Available() }
func (q *StreamQueue) Ping(ctx context.Context) error        { return q.inner.Ping(ctx) }
func (q *StreamQueue) EnsureGroup(ctx context.Context) error { return q.inner.EnsureGroup(ctx) }
func (q *StreamQueue) Read(ctx context.Context, c string, n int64, b time.Duration) ([]eventqueue.Entry, error) {
	return q.inner.Read(ctx, c, n, b)
}
func (q *StreamQueue) Ack(ctx context.Context, ids ...string) error { return q.inner.Ack(ctx, ids...) }
