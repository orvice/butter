// Package redislease provides a Redis-backed lease: a renewable, self-expiring
// token that elects exactly one holder for a logical role. It is the shared
// primitive behind the Telegram runtime's leadership and session leases and
// the AG-UI cross-Pod session guard.
package redislease

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Lease is a Redis-backed leadership token.
//
// A lease is renewed rather than held, so a Pod that crashes or stalls loses
// it after the TTL instead of blocking the fleet forever. All mutations are
// fenced on the holder value: a Pod that lost the lease during a pause cannot
// release or renew the new holder's token.
type Lease struct {
	rdb    *redis.Client
	key    string
	holder string
	ttl    time.Duration
}

// releaseScript deletes the key only if we still hold it, so a Pod that lost
// the lease during a pause cannot delete the new leader's token.
var releaseScript = redis.NewScript(`
if redis.call('GET', KEYS[1]) == ARGV[1] then
  return redis.call('DEL', KEYS[1])
end
return 0
`)

// renewScript extends the TTL only while we still hold it. A Pod that was
// preempted learns it from the return value rather than silently continuing
// to act as leader.
var renewScript = redis.NewScript(`
if redis.call('GET', KEYS[1]) == ARGV[1] then
  return redis.call('PEXPIRE', KEYS[1], ARGV[2])
end
return 0
`)

// New builds a lease for the given logical role. `holder` must be unique per
// process (an instance ID).
func New(rdb *redis.Client, key, holder string, ttl time.Duration) *Lease {
	if rdb == nil {
		return nil
	}
	return &Lease{rdb: rdb, key: key, holder: holder, ttl: ttl}
}

// Acquire tries to take or keep the lease. It returns true while this holder
// owns it.
func (l *Lease) Acquire(ctx context.Context) (bool, error) {
	if l == nil {
		return false, errors.New("lease is not configured")
	}
	ok, err := l.rdb.SetNX(ctx, l.key, l.holder, l.ttl).Result()
	if err != nil {
		return false, fmt.Errorf("acquire lease %q: %w", l.key, err)
	}
	if ok {
		return true, nil
	}
	// Not a new acquisition — but we may already be the holder, in which case
	// this call doubles as the renewal.
	return l.Renew(ctx)
}

// Renew extends the lease. It returns false once another holder has taken it.
func (l *Lease) Renew(ctx context.Context) (bool, error) {
	if l == nil {
		return false, errors.New("lease is not configured")
	}
	result, err := renewScript.Run(ctx, l.rdb, []string{l.key}, l.holder, l.ttl.Milliseconds()).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return false, fmt.Errorf("renew lease %q: %w", l.key, err)
	}
	renewed, _ := result.(int64)
	return renewed == 1, nil
}

// Release gives up the lease immediately, so a graceful shutdown does not
// make the next claimant wait out the TTL.
func (l *Lease) Release(ctx context.Context) error {
	if l == nil {
		return nil
	}
	if err := releaseScript.Run(ctx, l.rdb, []string{l.key}, l.holder).Err(); err != nil && !errors.Is(err, redis.Nil) {
		return fmt.Errorf("release lease %q: %w", l.key, err)
	}
	return nil
}
