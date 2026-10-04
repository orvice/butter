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
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"go.orx.me/apps/butter/internal/redislease"
)

// ErrLeaseLost is the cause of a turn context the guard cancelled because the
// turn no longer holds its session: the lease lapsed before a renewal got
// through, or another holder has it.
var ErrLeaseLost = errors.New("session lease lost")

// Guard serializes turns within one logical session.
//
// Acquire returns (turnCtx, release, acquired, err). When acquired, the
// caller must run the turn on turnCtx — it is cancelled if the lease is lost,
// with ErrLeaseLost as its cause, so a Pod that was fenced out stops acting
// instead of racing the new holder — and must call release exactly once when
// the turn ends, whatever the outcome. acquired=false with a nil error means
// another turn currently holds the session. turnCtx carries the
// acquisition's token (Token).
type Guard interface {
	Acquire(ctx context.Context, sessionKey string) (context.Context, func(), bool, error)
}

type tokenKey struct{}

// Token returns the token of the acquisition that made ctx, the turn context
// Acquire returned, or a context derived from it. Every acquisition gets its
// own token, and a Redis lease is held under it, so a token names one turn's
// hold on its session: a Stop bound to it never reaches a later turn
// (ADR-0016 decision 4). ok is false for a context no acquisition made.
func Token(ctx context.Context) (token string, ok bool) {
	token, _ = ctx.Value(tokenKey{}).(string)
	return token, token != ""
}

// withToken is ctx carrying an acquisition's token.
func withToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, tokenKey{}, token)
}

// Lease is one acquisition's renewable lease. redislease.Lease implements it.
type Lease interface {
	Acquire(ctx context.Context) (bool, error)
	// Renew extends the lease; false with a nil error means another holder
	// has it, or nobody does.
	Renew(ctx context.Context) (bool, error)
	Release(ctx context.Context) error
}

// Leased is the cross-Pod Guard: a bounded, renewable lease per session key,
// fenced on a per-acquisition holder token. NewRedis builds it over Redis
// leases.
type Leased struct {
	holder string
	ttl    time.Duration
	lease  func(sessionKey, leaseHolder string) Lease
}

// NewRedis builds a Redis guard. `holder` must be unique per process (an
// instance ID); each acquisition additionally gets its own token so two turns
// on one Pod can never be mistaken for each other. `ttl` bounds how long a
// crashed worker blocks one session: the lease is renewed every ttl/3 while
// the turn runs, so the TTL only has to cover a few renewals, not a turn.
func NewRedis(rdb *redis.Client, holder, keyPrefix string, ttl time.Duration) *Leased {
	if rdb == nil {
		return nil
	}
	return NewLeased(holder, ttl, func(sessionKey, leaseHolder string) Lease {
		return redislease.New(rdb, keyPrefix+sessionKey, leaseHolder, ttl)
	})
}

// NewLeased builds the guard over the leases newLease makes, one per
// acquisition. NewRedis is NewLeased over Redis leases; tests pass leases
// whose renewals fail on demand.
func NewLeased(holder string, ttl time.Duration, newLease func(sessionKey, leaseHolder string) Lease) *Leased {
	return &Leased{holder: holder, ttl: ttl, lease: newLease}
}

var _ Guard = (*Leased)(nil)

func (g *Leased) Acquire(ctx context.Context, sessionKey string) (context.Context, func(), bool, error) {
	// The lease is held under the acquisition's token, which the turn context
	// carries.
	token := g.holder + ":" + sessionKey + ":" + uuid.NewString()
	lease := g.lease(sessionKey, token)
	acquiredAt := time.Now()
	ok, err := lease.Acquire(ctx)
	if err != nil || !ok {
		return ctx, func() {}, ok, err
	}
	leaseCtx, cancel := context.WithCancelCause(withToken(ctx, token))
	done := make(chan struct{})
	go func() {
		defer close(done)
		g.keep(leaseCtx, cancel, lease, acquiredAt)
	}()
	var once sync.Once
	return leaseCtx, func() {
		once.Do(func() {
			cancel(nil)
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

// keep renews the lease every ttl/3 until ctx ends, cancelling ctx with
// ErrLeaseLost once the turn no longer holds it.
//
// A renewal that errors (Redis unreachable, a timeout) is retried until the
// lease would really have lapsed: one TTL after the last renewal that got
// through, measured from when that renewal was sent. Until then nobody else
// can hold the session, so a transient error must not end the turn. Only that
// lapse, or a renewal answered "not the holder", does.
func (g *Leased) keep(ctx context.Context, lose context.CancelCauseFunc, lease Lease, renewedAt time.Time) {
	interval := g.ttl / 3
	retry := renewRetryInterval(g.ttl)
	timer := time.NewTimer(interval)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}
		lapse := renewedAt.Add(g.ttl)
		sent := time.Now()
		// An attempt never outlives the lease it tries to keep.
		attemptCtx, cancelAttempt := context.WithDeadline(ctx, lapse)
		renewed, err := lease.Renew(attemptCtx)
		cancelAttempt()
		switch {
		case ctx.Err() != nil:
			return
		case err == nil && renewed:
			renewedAt = sent
			timer.Reset(interval)
		case err == nil:
			lose(ErrLeaseLost)
			return
		default:
			wait := time.Until(lapse)
			if wait <= 0 {
				lose(ErrLeaseLost)
				return
			}
			timer.Reset(min(retry, wait))
		}
	}
}

// renewRetryInterval paces the retries of a renewal that errored: often
// enough to get several attempts in before the lease would lapse.
func renewRetryInterval(ttl time.Duration) time.Duration {
	return max(ttl/10, time.Millisecond)
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
	// Each acquisition gets its own token, as a Redis lease's does.
	return withToken(ctx, "memory:"+sessionKey+":"+uuid.NewString()), func() {
		once.Do(func() {
			g.mu.Lock()
			defer g.mu.Unlock()
			delete(g.active, sessionKey)
		})
	}, true, nil
}
