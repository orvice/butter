package redislease

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Guard serializes work on one logical key across Pods. It is the reusable
// form of the Telegram session guard: Acquire takes a per-key lease, keeps it
// renewed in the background, and cancels the returned context if the lease is
// lost mid-work so the holder stops instead of running unfenced.
type Guard struct {
	rdb    *redis.Client
	holder string
	ttl    time.Duration
	lease  func(key, leaseHolder string) renewableLease
}

type renewableLease interface {
	Acquire(ctx context.Context) (bool, error)
	Renew(ctx context.Context) (bool, error)
	Release(ctx context.Context) error
}

// NewGuard builds a guard whose leases live for ttl and are renewed at ttl/3.
// `holder` must be unique per process (an instance ID). Returns nil when rdb
// is nil so single-Pod deployments can treat the guard as absent.
func NewGuard(rdb *redis.Client, holder string, ttl time.Duration) *Guard {
	if rdb == nil {
		return nil
	}
	g := &Guard{rdb: rdb, holder: holder, ttl: ttl}
	g.lease = func(key, leaseHolder string) renewableLease {
		return New(g.rdb, key, leaseHolder, g.ttl)
	}
	return g
}

// Acquire tries to take the key's lease. ok=false means another holder has it.
// On success the returned context is cancelled if the lease is lost, and the
// release func must be called once the work is done.
func (g *Guard) Acquire(ctx context.Context, key string) (context.Context, func(), bool, error) {
	lease := g.lease(key, g.holder+":"+key+":"+uuid.NewString())
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
			// Release on a detached context: the work's context may already be
			// cancelled, and holding the lease until it expires would stall the
			// next claimant.
			releaseCtx, releaseCancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
			defer releaseCancel()
			_ = lease.Release(releaseCtx)
		})
	}, true, nil
}
