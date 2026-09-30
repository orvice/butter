// Package eventqueue is the durable hand-off between a public receive path
// (a webhook handler, a long-polling leader) and the workers that act on what
// it accepted. It is protocol-neutral: callers choose the Stream, consumer
// group and dedupe namespace, and carry their own payload encoding.
//
// Redis Streams is the queue, not a cache. Acknowledging a delivery upstream
// is defined as "the payload is durably in the Stream": answering before that
// would tell the sender the event is handled while it exists only in one
// Pod's memory. This is why every caller must require Redis to be configured
// for persistence and no eviction (CheckReady).
package eventqueue

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"go.orx.me/apps/butter/internal/redislease"
)

// payloadField is the Stream entry field that carries the payload.
const payloadField = "event"

// ErrDuplicate reports that this dedupe ID was already accepted. It is not a
// failure: the caller acknowledges the delivery.
var ErrDuplicate = errors.New("event already accepted")

// Config names one queue's Redis namespace.
type Config struct {
	// Name labels errors, e.g. "telegram".
	Name string
	// StreamKey holds every accepted event. One stream per protocol, not per
	// tenant: a consumer group over a single stream is what lets any Pod pick
	// up any tenant's work.
	StreamKey string
	// ConsumerGroup is the shared consumer group all worker Pods join.
	ConsumerGroup string
	// DedupeKeyPrefix marks a dedupe ID as already accepted.
	DedupeKeyPrefix string
	// DedupeTTL bounds how long a duplicate is recognized. It must exceed the
	// sender's redelivery window.
	DedupeTTL time.Duration
	// LeasePreflightKeyPrefix namespaces the probe lease CheckLeaseReady
	// takes.
	LeasePreflightKeyPrefix string
}

// acceptScript deduplicates and appends in one round trip.
//
// Atomicity matters here, not speed: two Pods can receive the same redelivery
// concurrently. Doing SETNX and XADD as separate commands would let both pass
// the SETNX check window, or leave a dedupe marker set for an event that was
// never appended — which would silently drop the event forever.
var acceptScript = redis.NewScript(`
local seen = redis.call('SET', KEYS[1], '1', 'NX', 'PX', ARGV[1])
if not seen then
  return nil
end
return redis.call('XADD', KEYS[2], '*', 'event', ARGV[2])
`)

// touchScript refreshes a pending entry only while it still belongs to the
// expected consumer. XPENDING and XCLAIM run atomically inside the script so
// a stale heartbeat cannot steal work back from a new owner.
var touchScript = redis.NewScript(`
local pending = redis.call('XPENDING', KEYS[1], ARGV[1], ARGV[3], ARGV[3], 1)
if #pending == 0 or pending[1][2] ~= ARGV[2] then
  return 0
end
redis.call('XCLAIM', KEYS[1], ARGV[1], ARGV[2], 0, ARGV[3], 'JUSTID')
return 1
`)

// Queue is the Redis Streams implementation of the receive hand-off.
type Queue struct {
	rdb *redis.Client
	cfg Config
}

// New builds a queue over rdb. It returns nil when rdb is nil, and a nil
// Queue reports Available() == false.
func New(rdb *redis.Client, cfg Config) *Queue {
	if rdb == nil {
		return nil
	}
	return &Queue{rdb: rdb, cfg: cfg}
}

// Entry is one claimed Stream entry.
type Entry struct {
	// ID is the Stream entry ID, used to acknowledge.
	ID      string
	Payload string
}

// Available reports whether a durable queue is wired at all. Callers use it
// to block enablement rather than discovering the gap at the first event.
func (q *Queue) Available() bool { return q != nil && q.rdb != nil }

func (q *Queue) notConfigured() error {
	name := "event"
	if q != nil && q.cfg.Name != "" {
		name = q.cfg.Name
	}
	return fmt.Errorf("%s queue is not configured", name)
}

// Ping verifies that the queue backend is reachable without imposing the
// persistence-policy checks required before accepting deliveries.
func (q *Queue) Ping(ctx context.Context) error {
	if !q.Available() {
		return q.notConfigured()
	}
	return q.rdb.Ping(ctx).Err()
}

// Accept durably records a payload, returning ErrDuplicate when dedupeID was
// already taken. The returned ID is the Stream entry ID.
//
// The caller must treat a non-nil error other than ErrDuplicate as "not
// accepted" and answer the sender with a retryable status: acknowledging an
// event we failed to enqueue loses it permanently.
func (q *Queue) Accept(ctx context.Context, dedupeID, payload string) (string, error) {
	if !q.Available() {
		return "", q.notConfigured()
	}
	result, err := acceptScript.Run(ctx, q.rdb,
		[]string{q.cfg.DedupeKeyPrefix + dedupeID, q.cfg.StreamKey},
		q.cfg.DedupeTTL.Milliseconds(), payload,
	).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrDuplicate
	}
	if err != nil {
		return "", fmt.Errorf("accept %s event: %w", q.cfg.Name, err)
	}
	id, _ := result.(string)
	return id, nil
}

// EnsureGroup creates the consumer group if it does not exist. It is safe to
// call from every Pod on every start.
func (q *Queue) EnsureGroup(ctx context.Context) error {
	if !q.Available() {
		return q.notConfigured()
	}
	// MKSTREAM so the first Pod to start does not have to wait for an event
	// before the group can exist.
	err := q.rdb.XGroupCreateMkStream(ctx, q.cfg.StreamKey, q.cfg.ConsumerGroup, "0").Err()
	if err != nil && !isBusyGroup(err) {
		return fmt.Errorf("create %s consumer group: %w", q.cfg.Name, err)
	}
	return nil
}

// isBusyGroup recognizes "the group already exists", which every Pod after
// the first will see. Redis reports it only as a message prefix.
func isBusyGroup(err error) bool {
	return err != nil && strings.HasPrefix(err.Error(), "BUSYGROUP")
}

// Read claims up to count new entries for this consumer, blocking up to
// block for work to arrive.
func (q *Queue) Read(ctx context.Context, consumer string, count int64, block time.Duration) ([]Entry, error) {
	return q.read(ctx, consumer, ">", count, block)
}

// ReadPending re-claims entries this consumer already holds but never
// acknowledged — the state a Pod is in after a crash mid-turn.
func (q *Queue) ReadPending(ctx context.Context, consumer string, count int64) ([]Entry, error) {
	return q.read(ctx, consumer, "0", count, 0)
}

func (q *Queue) read(ctx context.Context, consumer, start string, count int64, block time.Duration) ([]Entry, error) {
	if !q.Available() {
		return nil, q.notConfigured()
	}
	streams, err := q.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    q.cfg.ConsumerGroup,
		Consumer: consumer,
		Streams:  []string{q.cfg.StreamKey, start},
		Count:    count,
		Block:    block,
	}).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s events: %w", q.cfg.Name, err)
	}
	var out []Entry
	for _, stream := range streams {
		for _, message := range stream.Messages {
			payload, _ := message.Values[payloadField].(string)
			out = append(out, Entry{ID: message.ID, Payload: payload})
		}
	}
	return out, nil
}

// Ack marks entries as fully handled.
func (q *Queue) Ack(ctx context.Context, ids ...string) error {
	if !q.Available() || len(ids) == 0 {
		return nil
	}
	if err := q.rdb.XAck(ctx, q.cfg.StreamKey, q.cfg.ConsumerGroup, ids...).Err(); err != nil {
		return fmt.Errorf("ack %s events: %w", q.cfg.Name, err)
	}
	return nil
}

// Touch resets one pending entry's idle time while its consumer is still
// handling it. Without this heartbeat, XAUTOCLAIM can steal a healthy long
// turn merely because it exceeded the reclaim idle time.
func (q *Queue) Touch(ctx context.Context, consumer, id string) error {
	if !q.Available() {
		return q.notConfigured()
	}
	result, err := touchScript.Run(ctx, q.rdb, []string{q.cfg.StreamKey}, q.cfg.ConsumerGroup, consumer, id).Int64()
	if err != nil {
		return fmt.Errorf("touch %s event %s: %w", q.cfg.Name, id, err)
	}
	if result != 1 {
		return fmt.Errorf("touch %s event %s: pending entry is no longer owned", q.cfg.Name, id)
	}
	return nil
}

// Claim takes over entries idle longer than minIdle from whichever consumer
// holds them, so a crashed Pod's work does not stall.
func (q *Queue) Claim(ctx context.Context, consumer string, minIdle time.Duration, count int64) ([]Entry, error) {
	if !q.Available() {
		return nil, q.notConfigured()
	}
	messages, _, err := q.rdb.XAutoClaim(ctx, &redis.XAutoClaimArgs{
		Stream:   q.cfg.StreamKey,
		Group:    q.cfg.ConsumerGroup,
		Consumer: consumer,
		MinIdle:  minIdle,
		Start:    "0",
		Count:    count,
	}).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("claim %s events: %w", q.cfg.Name, err)
	}
	out := make([]Entry, 0, len(messages))
	for _, message := range messages {
		payload, _ := message.Values[payloadField].(string)
		out = append(out, Entry{ID: message.ID, Payload: payload})
	}
	return out, nil
}

// CheckReady verifies that Redis can be used as a durable queue. Reachability
// alone is insufficient: evicting or non-persistent Redis can acknowledge a
// delivery and then discard the only authoritative copy.
func (q *Queue) CheckReady(ctx context.Context) error {
	if !q.Available() {
		return q.notConfigured()
	}
	if err := q.Ping(ctx); err != nil {
		return fmt.Errorf("redis is unreachable: %w", err)
	}
	policy, err := q.rdb.ConfigGet(ctx, "maxmemory-policy").Result()
	if err != nil {
		return fmt.Errorf("read maxmemory-policy: %w", err)
	}
	persistence, err := q.rdb.ConfigGet(ctx, "appendonly").Result()
	if err != nil {
		return fmt.Errorf("read appendonly: %w", err)
	}
	snapshot, err := q.rdb.ConfigGet(ctx, "save").Result()
	if err != nil {
		return fmt.Errorf("read save schedule: %w", err)
	}
	return ValidateDurabilityConfig(policy, persistence, snapshot)
}

// CheckLeaseReady verifies the Redis commands and Lua execution used by the
// renewable leases that coordinate queue consumers. A durable queue is not
// enough if the configured Redis ACL cannot acquire, renew, or release a
// lease.
func (q *Queue) CheckLeaseReady(ctx context.Context) error {
	if !q.Available() {
		return q.notConfigured()
	}
	probeID := uuid.NewString()
	lease := redislease.New(q.rdb, q.cfg.LeasePreflightKeyPrefix+probeID, probeID, 10*time.Second)
	acquired, err := lease.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire probe lease: %w", err)
	}
	if !acquired {
		return errors.New("acquire probe lease: lease was not acquired")
	}
	defer func() {
		releaseCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		_ = lease.Release(releaseCtx)
	}()

	renewed, err := lease.Renew(ctx)
	if err != nil {
		return fmt.Errorf("renew probe lease: %w", err)
	}
	if !renewed {
		return errors.New("renew probe lease: lease ownership was lost")
	}
	if err := lease.Release(ctx); err != nil {
		return fmt.Errorf("release probe lease: %w", err)
	}
	return nil
}

// ValidateDurabilityConfig accepts the CONFIG GET answers of a Redis that
// neither evicts nor loses its data on restart.
func ValidateDurabilityConfig(policy, persistence, snapshot map[string]string) error {
	if policy["maxmemory-policy"] != "noeviction" {
		return errors.New("maxmemory-policy must be noeviction")
	}
	if persistence["appendonly"] == "yes" || strings.TrimSpace(snapshot["save"]) != "" {
		return nil
	}
	return errors.New("Redis persistence is disabled; enable AOF or RDB snapshots")
}
