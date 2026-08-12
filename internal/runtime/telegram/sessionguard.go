package telegram

import (
	"context"
	"sync"
	"time"

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
}

func NewRedisSessionGuard(rdb *redis.Client, holder string) *RedisSessionGuard {
	if rdb == nil {
		return nil
	}
	return &RedisSessionGuard{rdb: rdb, holder: holder}
}

var _ SessionGuard = (*RedisSessionGuard)(nil)

func (g *RedisSessionGuard) Acquire(ctx context.Context, sessionID string) (func(), bool, error) {
	lease := telegramqueue.NewLease(g.rdb, telegramqueue.SessionLeaseKey(sessionID),
		g.holder+":"+sessionID, sessionLeaseTTL)
	ok, err := lease.Acquire(ctx)
	if err != nil || !ok {
		return func() {}, ok, err
	}
	return func() {
		// Release on a detached context: the turn's context may already be
		// cancelled, and holding the lease until it expires would stall the
		// next message in this conversation.
		releaseCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		_ = lease.Release(releaseCtx)
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

func (g *MemorySessionGuard) Acquire(_ context.Context, sessionID string) (func(), bool, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.active[sessionID] {
		return func() {}, false, nil
	}
	g.active[sessionID] = true
	return func() {
		g.mu.Lock()
		defer g.mu.Unlock()
		delete(g.active, sessionID)
	}, true, nil
}
