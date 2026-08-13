package telegramqueue

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	// StreamKey holds every accepted Telegram update. One stream, not one
	// per Channel: a consumer group over a single stream is what lets any
	// Pod pick up any Channel's work, which is the whole point of the
	// multi-Pod design.
	StreamKey = "butter:telegram:updates"
	// ConsumerGroup is the shared consumer group all worker Pods join.
	ConsumerGroup = "butter-telegram-workers"

	// dedupeKeyPrefix marks an (channel, update) pair as already accepted.
	dedupeKeyPrefix = "butter:telegram:seen:"
	// dedupeTTL bounds how long a duplicate is recognized. Telegram retries
	// an undelivered update for far less than this.
	dedupeTTL = 24 * time.Hour
)

// ErrDuplicate reports that this (channel, update) pair was already accepted.
// It is not a failure: the caller acknowledges the delivery.
var ErrDuplicate = errors.New("telegram update already accepted")

// acceptScript deduplicates and appends in one round trip.
//
// Atomicity matters here, not speed: two Pods can receive the same Telegram
// retry concurrently. Doing SETNX and XADD as separate commands would let
// both pass the SETNX check window, or leave a dedupe marker set for an event
// that was never appended — which would silently drop the update forever.
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
}

func New(rdb *redis.Client) *Queue {
	if rdb == nil {
		return nil
	}
	return &Queue{rdb: rdb}
}

// Available reports whether a durable queue is wired at all. Callers use it
// to block enablement rather than discovering the gap at the first update.
func (q *Queue) Available() bool { return q != nil && q.rdb != nil }

// Ping verifies that the queue backend is reachable without imposing the
// persistence-policy checks required before accepting webhook deliveries.
func (q *Queue) Ping(ctx context.Context) error {
	if !q.Available() {
		return errors.New("telegram queue is not configured")
	}
	return q.rdb.Ping(ctx).Err()
}

// Accept durably records an event, returning ErrDuplicate when this
// (channel, update) pair was already taken. The returned ID is the Stream
// entry ID.
//
// The caller must treat a non-nil error other than ErrDuplicate as
// "not accepted" and answer Telegram with a retryable status: acknowledging
// an update we failed to enqueue loses it permanently.
func (q *Queue) Accept(ctx context.Context, event *Event) (string, error) {
	if !q.Available() {
		return "", errors.New("telegram queue is not configured")
	}
	payload, err := event.Encode()
	if err != nil {
		return "", err
	}
	dedupeKey := fmt.Sprintf("%s%s:%d", dedupeKeyPrefix, event.ChannelID, event.UpdateID)

	result, err := acceptScript.Run(ctx, q.rdb,
		[]string{dedupeKey, StreamKey},
		dedupeTTL.Milliseconds(), payload,
	).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrDuplicate
	}
	if err != nil {
		return "", fmt.Errorf("accept telegram update: %w", err)
	}
	id, _ := result.(string)
	return id, nil
}

// EnsureGroup creates the consumer group if it does not exist. It is safe to
// call from every Pod on every start.
func (q *Queue) EnsureGroup(ctx context.Context) error {
	if !q.Available() {
		return errors.New("telegram queue is not configured")
	}
	// MKSTREAM so the first Pod to start does not have to wait for an update
	// before the group can exist.
	err := q.rdb.XGroupCreateMkStream(ctx, StreamKey, ConsumerGroup, "0").Err()
	if err != nil && !isBusyGroup(err) {
		return fmt.Errorf("create telegram consumer group: %w", err)
	}
	return nil
}

// isBusyGroup recognizes "the group already exists", which every Pod after
// the first will see. Redis reports it only as a message prefix.
func isBusyGroup(err error) bool {
	return err != nil && strings.HasPrefix(err.Error(), "BUSYGROUP")
}

// Delivery is one claimed Stream entry.
type Delivery struct {
	// ID is the Stream entry ID, used to acknowledge.
	ID    string
	Event *Event
}

// Read claims up to count new entries for this consumer, blocking up to
// block for work to arrive.
func (q *Queue) Read(ctx context.Context, consumer string, count int64, block time.Duration) ([]Delivery, error) {
	return q.read(ctx, consumer, ">", count, block)
}

// ReadPending re-claims entries this consumer already holds but never
// acknowledged — the state a Pod is in after a crash mid-turn.
func (q *Queue) ReadPending(ctx context.Context, consumer string, count int64) ([]Delivery, error) {
	return q.read(ctx, consumer, "0", count, 0)
}

func (q *Queue) read(ctx context.Context, consumer, start string, count int64, block time.Duration) ([]Delivery, error) {
	if !q.Available() {
		return nil, errors.New("telegram queue is not configured")
	}
	streams, err := q.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    ConsumerGroup,
		Consumer: consumer,
		Streams:  []string{StreamKey, start},
		Count:    count,
		Block:    block,
	}).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read telegram updates: %w", err)
	}

	var out []Delivery
	for _, stream := range streams {
		for _, message := range stream.Messages {
			payload, _ := message.Values["event"].(string)
			event, decodeErr := DecodeEvent(payload)
			if decodeErr != nil {
				// An undecodable entry can never succeed. Acknowledge it so
				// it stops being redelivered, and report it.
				_ = q.Ack(ctx, message.ID)
				return out, fmt.Errorf("drop undecodable telegram event %s: %w", message.ID, decodeErr)
			}
			out = append(out, Delivery{ID: message.ID, Event: event})
		}
	}
	return out, nil
}

// Ack marks entries as fully handled.
func (q *Queue) Ack(ctx context.Context, ids ...string) error {
	if !q.Available() || len(ids) == 0 {
		return nil
	}
	if err := q.rdb.XAck(ctx, StreamKey, ConsumerGroup, ids...).Err(); err != nil {
		return fmt.Errorf("ack telegram updates: %w", err)
	}
	return nil
}

// Touch resets one pending entry's idle time while its consumer is still
// handling it. Without this heartbeat, XAUTOCLAIM can steal a healthy long
// Agent turn merely because it exceeded reclaimIdle.
func (q *Queue) Touch(ctx context.Context, consumer, id string) error {
	if !q.Available() {
		return errors.New("telegram queue is not configured")
	}
	result, err := touchScript.Run(ctx, q.rdb, []string{StreamKey}, ConsumerGroup, consumer, id).Int64()
	if err != nil {
		return fmt.Errorf("touch telegram update %s: %w", id, err)
	}
	if result != 1 {
		return fmt.Errorf("touch telegram update %s: pending entry is no longer owned", id)
	}
	return nil
}

// Claim takes over entries idle longer than minIdle from whichever consumer
// holds them, so a crashed Pod's work does not stall.
func (q *Queue) Claim(ctx context.Context, consumer string, minIdle time.Duration, count int64) ([]Delivery, error) {
	if !q.Available() {
		return nil, errors.New("telegram queue is not configured")
	}
	messages, _, err := q.rdb.XAutoClaim(ctx, &redis.XAutoClaimArgs{
		Stream:   StreamKey,
		Group:    ConsumerGroup,
		Consumer: consumer,
		MinIdle:  minIdle,
		Start:    "0",
		Count:    count,
	}).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("claim telegram updates: %w", err)
	}

	var out []Delivery
	for _, message := range messages {
		payload, _ := message.Values["event"].(string)
		event, decodeErr := DecodeEvent(payload)
		if decodeErr != nil {
			_ = q.Ack(ctx, message.ID)
			continue
		}
		out = append(out, Delivery{ID: message.ID, Event: event})
	}
	return out, nil
}

// CheckReady verifies that Redis can be used as a durable queue. Reachability
// alone is insufficient: evicting or non-persistent Redis can acknowledge a
// Telegram update and then discard the only authoritative copy.
func (q *Queue) CheckReady(ctx context.Context) error {
	if !q.Available() {
		return errors.New("telegram queue is not configured")
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
	return validateDurabilityConfig(policy, persistence, snapshot)
}

// CheckLeaseReady verifies the Redis commands and Lua execution used by the
// renewable leadership lease. Long Polling must reject enablement if the queue
// is durable but the configured Redis ACL cannot acquire, renew, or release a
// consumer lease.
func (q *Queue) CheckLeaseReady(ctx context.Context) error {
	if !q.Available() {
		return errors.New("telegram queue is not configured")
	}
	probeID := uuid.NewString()
	lease := NewLease(q.rdb, "butter:telegram:lease:preflight:"+probeID, probeID, 10*time.Second)
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

func validateDurabilityConfig(policy, persistence, snapshot map[string]string) error {
	if policy["maxmemory-policy"] != "noeviction" {
		return errors.New("maxmemory-policy must be noeviction")
	}
	if persistence["appendonly"] == "yes" || strings.TrimSpace(snapshot["save"]) != "" {
		return nil
	}
	return errors.New("Redis persistence is disabled; enable AOF or RDB snapshots")
}
