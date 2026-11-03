package workstation

import (
	"context"
	"sync"
	"time"
)

// Budget coordinates provider requests that share one external identity. The
// production implementation will need a durable/Redis-backed implementation;
// this memory implementation is useful for a single process and contract tests.
type Budget interface {
	Reserve(context.Context, string, time.Time) (time.Duration, error)
}

type budgetWindow struct {
	started time.Time
	used    int
}

// MemoryBudget limits each key to Limit reservations per Window. A successful
// reservation returns zero; a rejected reservation returns the time remaining
// in the current window. It is safe for concurrent scopes and connections.
type MemoryBudget struct {
	mu      sync.Mutex
	Limit   int
	Window  time.Duration
	buckets map[string]budgetWindow
}

func NewMemoryBudget(limit int, window time.Duration) *MemoryBudget {
	return &MemoryBudget{Limit: limit, Window: window, buckets: map[string]budgetWindow{}}
}

func (b *MemoryBudget) Reserve(ctx context.Context, key string, now time.Time) (time.Duration, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	if b == nil || b.Limit <= 0 || b.Window <= 0 || key == "" {
		return 0, nil
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	bucket := b.buckets[key]
	if bucket.started.IsZero() || !now.Before(bucket.started.Add(b.Window)) {
		bucket = budgetWindow{started: now}
	}
	if bucket.used >= b.Limit {
		return bucket.started.Add(b.Window).Sub(now), nil
	}
	bucket.used++
	b.buckets[key] = bucket
	return 0, nil
}

func BudgetKey(source SourceIdentity) string {
	return source.Provider + "\x00" + source.Origin + "\x00" + source.TenantID
}
