// Package sessionguard serializes agent turns within one logical session
// across Pods.
//
// Two turns for the same conversation running at once would interleave their
// history writes, producing a session that reads as two people talking over
// each other. Unrelated sessions are deliberately unaffected: the lease is
// per session key, which is what keeps the fleet parallel. The Telegram
// runtime carries its own equivalent (internal/runtime/telegram.SessionGuard);
// this package is the protocol-neutral variant used by the AG-UI endpoint.
package sessionguard

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"go.orx.me/apps/butter/internal/redislease"
)

// Guard serializes turns within one logical session.
//
// Acquire returns (turnCtx, release, acquired, err). When acquired, the
// caller must run the turn on turnCtx — it is cancelled if the lease is lost,
// so a Pod that was fenced out stops acting instead of racing the new holder —
// and must call release exactly once when the turn ends, whatever the
// outcome. acquired=false with a nil error means another turn currently holds
// the session.
type Guard interface {
	Acquire(ctx context.Context, sessionKey string) (context.Context, func(), bool, error)
}

// Redis is the cross-Pod Guard: a bounded, renewable Redis lease per session
// key, fenced on a per-acquisition holder token.
type Redis struct {
	holder    string
	keyPrefix string
	ttl       time.Duration
	lease     func(sessionKey, leaseHolder string) renewableLease
}

type renewableLease interface {
	Acquire(ctx context.Context) (bool, error)
	Renew(ctx context.Context) (bool, error)
	Release(ctx context.Context) error
}

// NewRedis builds a Redis guard. `holder` must be unique per process (an
// instance ID); each acquisition additionally gets its own token so two turns
// on one Pod can never be mistaken for each other. `ttl` bounds how long a
// crashed worker blocks one session — it has to exceed a normal turn, or a
// slow agent would lose its own lease mid-run.
func NewRedis(rdb *redis.Client, holder, keyPrefix string, ttl time.Duration) *Redis {
	if rdb == nil {
		return nil
	}
	g := &Redis{holder: holder, keyPrefix: keyPrefix, ttl: ttl}
	g.lease = func(sessionKey, leaseHolder string) renewableLease {
		return redislease.New(rdb, g.keyPrefix+sessionKey, leaseHolder, g.ttl)
	}
	return g
}

var _ Guard = (*Redis)(nil)

func (g *Redis) Acquire(ctx context.Context, sessionKey string) (context.Context, func(), bool, error) {
	lease := g.lease(sessionKey, g.holder+":"+sessionKey+":"+uuid.NewString())
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
			// next turn in this conversation.
			releaseCtx, releaseCancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
			defer releaseCancel()
			_ = lease.Release(releaseCtx)
		})
	}, true, nil
}

// Memory serializes sessions within one process. It is the single-Pod
// fallback and what tests use.
type Memory struct {
	mu     sync.Mutex
	active map[string]bool
}

func NewMemory() *Memory {
	return &Memory{active: make(map[string]bool)}
}

var _ Guard = (*Memory)(nil)

func (g *Memory) Acquire(ctx context.Context, sessionKey string) (context.Context, func(), bool, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.active[sessionKey] {
		return ctx, func() {}, false, nil
	}
	g.active[sessionKey] = true
	var once sync.Once
	return ctx, func() {
		once.Do(func() {
			g.mu.Lock()
			defer g.mu.Unlock()
			delete(g.active, sessionKey)
		})
	}, true, nil
}
