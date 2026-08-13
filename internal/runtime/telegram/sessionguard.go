package telegram

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"go.orx.me/apps/butter/internal/telegramqueue"
)

// sessionLeaseTTL bounds how long a crashed worker blocks one session. It has
// to exceed a normal turn, or a slow agent would lose its own lease mid-run.
const sessionLeaseTTL = 5 * time.Minute

// RedisSessionGuard serializes turns within one derived session across Pods.
//
// Two updates for the same conversation running at once would interleave
// their history writes, producing a session that reads as two people talking
// over each other. Unrelated sessions are deliberately unaffected: the lease
// is per session, which is what keeps the fleet parallel.
type RedisSessionGuard struct {
	rdb    *redis.Client
	holder string
	ttl    time.Duration
	lease  func(sessionID, leaseHolder string) renewableLease
}

type renewableLease interface {
	Acquire(ctx context.Context) (bool, error)
	Renew(ctx context.Context) (bool, error)
	Release(ctx context.Context) error
}

func NewRedisSessionGuard(rdb *redis.Client, holder string) *RedisSessionGuard {
	if rdb == nil {
		return nil
	}
	g := &RedisSessionGuard{rdb: rdb, holder: holder, ttl: sessionLeaseTTL}
	g.lease = func(sessionID, leaseHolder string) renewableLease {
		return telegramqueue.NewLease(g.rdb, telegramqueue.SessionLeaseKey(sessionID),
			leaseHolder, g.ttl)
	}
	return g
}

var _ SessionGuard = (*RedisSessionGuard)(nil)

func (g *RedisSessionGuard) Acquire(ctx context.Context, sessionID string) (context.Context, func(), bool, error) {
	lease := g.lease(sessionID, g.holder+":"+sessionID+":"+uuid.NewString())
	ok, err := lease.Acquire(ctx)
	if err != nil || !ok {
		return ctx, func() {}, ok, err
	}
	leaseCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(g.ttl / 3)
		defer ticker.Stop()
		for {
			select {
			case <-leaseCtx.Done():
				return
			case <-ticker.C:
				renewed, renewErr := lease.Renew(leaseCtx)
				if renewErr != nil || !renewed {
					cancel()
					return
				}
			}
		}
	}()
	var once sync.Once
	return leaseCtx, func() {
		once.Do(func() {
			cancel()
			<-done
			// Release on a detached context: the turn's context may already be
			// cancelled, and holding the lease until it expires would stall the
			// next message in this conversation.
			releaseCtx, releaseCancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
			defer releaseCancel()
			_ = lease.Release(releaseCtx)
		})
	}, true, nil
}

// MemorySessionGuard serializes sessions within one process. It is the
// single-Pod fallback and what tests use.
type MemorySessionGuard struct {
	mu     sync.Mutex
	active map[string]bool
}

func NewMemorySessionGuard() *MemorySessionGuard {
	return &MemorySessionGuard{active: make(map[string]bool)}
}

var _ SessionGuard = (*MemorySessionGuard)(nil)

func (g *MemorySessionGuard) Acquire(ctx context.Context, sessionID string) (context.Context, func(), bool, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.active[sessionID] {
		return ctx, func() {}, false, nil
	}
	g.active[sessionID] = true
	return ctx, func() {
		g.mu.Lock()
		defer g.mu.Unlock()
		delete(g.active, sessionID)
	}, true, nil
}
