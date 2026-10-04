// Package liveness tells whether the Butter process that owns a piece of work
// is still running (#390, ADR-0016 decision 3).
//
// Every process gets one instance ID per start, publishes a short-lived
// liveness record under it, and keeps renewing that record for as long as it
// runs. A process that crashes, is killed, or is cut off from the registry
// stops renewing, and its record lapses one TTL later. Work stamped with an
// instance ID whose record has lapsed was left behind by a process that is
// gone; work stamped with a live one is still in hand, whichever process is
// asking. Another process starting says nothing about either.
package liveness

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"butterfly.orx.me/core/log"
	"github.com/redis/go-redis/v9"
)

const (
	// KeyPrefix namespaces the liveness keys in Redis: one key per process,
	// butter:instance:{id}.
	KeyPrefix = "butter:instance:"
	// DefaultTTL is how long one publication lasts. Keep renews at a third of
	// it, so a process has to miss two renewals in a row before it reads as
	// gone.
	DefaultTTL = 30 * time.Second
)

// Registry is where processes publish that they are alive and look up
// whether others still are.
type Registry interface {
	// Publish marks the instance alive for one TTL from now.
	Publish(ctx context.Context, id string) error
	// Alive reports, for each of ids, whether its last publication is still
	// current.
	Alive(ctx context.Context, ids []string) (map[string]bool, error)
	// TTL is how long one publication lasts.
	TTL() time.Duration
}

// Keep publishes id on r now and renews it every TTL/3 until ctx ends.
//
// It returns the first publication's error, but keeps renewing either way, so
// a registry that is unreachable at startup is caught up with once it is
// back. Only a process that stops renewing lets its publication lapse; the
// application renews for as long as it runs, and tests cancel ctx to stand in
// for a crash.
func Keep(ctx context.Context, r Registry, id string) error {
	var published time.Time
	err := publish(ctx, r, id)
	if err == nil {
		published = time.Now()
	}
	go renew(ctx, r, id, published)
	return err
}

// publish gives one publication as long as the renewal interval: a slower
// one has failed anyway, and must not hold up startup or the next renewal.
func publish(ctx context.Context, r Registry, id string) error {
	ctx, cancel := context.WithTimeout(ctx, r.TTL()/3)
	defer cancel()
	return r.Publish(ctx, id)
}

func renew(ctx context.Context, r Registry, id string, published time.Time) {
	ticker := time.NewTicker(r.TTL() / 3)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		if err := publish(ctx, r, id); err != nil {
			if ctx.Err() == nil {
				log.FromContext(ctx).Warn("instance liveness renewal failed", "instance_id", id, "err", err)
			}
			continue
		}
		now := time.Now()
		switch {
		case published.IsZero():
			log.FromContext(ctx).Info("instance liveness published after a failed start", "instance_id", id)
		case now.Sub(published) > r.TTL():
			log.FromContext(ctx).Warn("instance liveness lapsed before it was renewed; "+
				"other processes may have failed this process's QUEUED/RUNNING invocations as stale meanwhile",
				"instance_id", id, "gap", now.Sub(published))
		}
		published = now
	}
}

// --- In-process registry ------------------------------------------------------

// Memory is an in-process Registry. In a deployment without Redis it holds
// this process alone: there is one process, so every other instance reads as
// gone. Tests use it in place of Redis, and it honors the same TTL.
type Memory struct {
	ttl     time.Duration
	mu      sync.Mutex
	expires map[string]time.Time
}

// NewMemory builds an in-process registry whose publications last ttl.
func NewMemory(ttl time.Duration) *Memory {
	return &Memory{ttl: ttl, expires: make(map[string]time.Time)}
}

var _ Registry = (*Memory)(nil)

func (m *Memory) Publish(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	for other, until := range m.expires {
		if !now.Before(until) {
			delete(m.expires, other)
		}
	}
	m.expires[id] = now.Add(m.ttl)
	return nil
}

func (m *Memory) Alive(_ context.Context, ids []string) (map[string]bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	out := make(map[string]bool, len(ids))
	for _, id := range ids {
		until, ok := m.expires[id]
		out[id] = ok && now.Before(until)
	}
	return out, nil
}

func (m *Memory) TTL() time.Duration { return m.ttl }

// --- Redis registry ---------------------------------------------------------------

// Redis is the cross-Pod Registry: one key per instance, written with the TTL
// on every publication. The key holds the host name, so an operator can tell
// which Pod an instance ID belongs to while it is alive.
type Redis struct {
	rdb    *redis.Client
	ttl    time.Duration
	prefix string
	host   string
}

// NewRedis builds the Redis registry. Returns nil when rdb is nil, so callers
// can fall back to the in-process registry.
func NewRedis(rdb *redis.Client, ttl time.Duration) *Redis {
	if rdb == nil {
		return nil
	}
	host, err := os.Hostname()
	if err != nil || host == "" {
		host = "unknown"
	}
	return &Redis{rdb: rdb, ttl: ttl, prefix: KeyPrefix, host: host}
}

var _ Registry = (*Redis)(nil)

func (r *Redis) key(id string) string { return r.prefix + id }

func (r *Redis) Publish(ctx context.Context, id string) error {
	if err := r.rdb.Set(ctx, r.key(id), r.host, r.ttl).Err(); err != nil {
		return fmt.Errorf("publish liveness of instance %q: %w", id, err)
	}
	return nil
}

func (r *Redis) Alive(ctx context.Context, ids []string) (map[string]bool, error) {
	out := make(map[string]bool, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	pipe := r.rdb.Pipeline()
	exists := make([]*redis.IntCmd, len(ids))
	for i, id := range ids {
		exists[i] = pipe.Exists(ctx, r.key(id))
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, fmt.Errorf("read instance liveness: %w", err)
	}
	for i, id := range ids {
		out[id] = exists[i].Val() > 0
	}
	return out, nil
}

func (r *Redis) TTL() time.Duration { return r.ttl }
