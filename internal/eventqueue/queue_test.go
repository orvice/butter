package eventqueue

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// redisForTest connects to the Redis named by REDIS_ADDR, skipping the test
// when none is configured. Every test namespaces its keys with a fresh ID.
func redisForTest(t *testing.T) *redis.Client {
	t.Helper()
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		t.Skip("REDIS_ADDR is required for Redis-backed queue tests")
	}
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	if err := rdb.Ping(t.Context()).Err(); err != nil {
		t.Fatalf("ping redis %s: %v", addr, err)
	}
	t.Cleanup(func() { _ = rdb.Close() })
	return rdb
}

func queueForTest(t *testing.T, rdb *redis.Client, name string) *Queue {
	t.Helper()
	ns := "butter:test:" + uuid.NewString() + ":" + name
	q := New(rdb, Config{
		Name:            name,
		StreamKey:       ns + ":stream",
		ConsumerGroup:   ns + ":group",
		DedupeKeyPrefix: ns + ":seen:",
		DedupeTTL:       time.Minute,
	})
	t.Cleanup(func() {
		ctx := t.Context()
		keys, _ := rdb.Keys(ctx, ns+"*").Result()
		if len(keys) > 0 {
			_ = rdb.Del(ctx, keys...).Err()
		}
	})
	if err := q.EnsureGroup(t.Context()); err != nil {
		t.Fatalf("ensure group: %v", err)
	}
	return q
}

func TestTwoQueuesKeepSeparateDedupeNamespacesAndStreams(t *testing.T) {
	rdb := redisForTest(t)
	telegram := queueForTest(t, rdb, "telegram")
	linear := queueForTest(t, rdb, "linear")
	ctx := t.Context()

	if _, err := telegram.Accept(ctx, "delivery-1", "tg payload"); err != nil {
		t.Fatalf("accept on first queue: %v", err)
	}
	// The same dedupe ID on another queue is a different event.
	if _, err := linear.Accept(ctx, "delivery-1", "linear payload"); err != nil {
		t.Fatalf("accept on second queue with the same dedupe ID: %v", err)
	}
	// Within one queue it is a duplicate.
	if _, err := telegram.Accept(ctx, "delivery-1", "tg payload again"); !errors.Is(err, ErrDuplicate) {
		t.Fatalf("second accept on first queue = %v, want ErrDuplicate", err)
	}

	got, err := linear.Read(ctx, "worker-a", 10, 0)
	if err != nil {
		t.Fatalf("read second queue: %v", err)
	}
	if len(got) != 1 || got[0].Payload != "linear payload" {
		t.Fatalf("second queue entries = %+v, want only its own payload", got)
	}
}

func TestAcceptedEventIsReadOnceAndReclaimedAfterIdle(t *testing.T) {
	rdb := redisForTest(t)
	q := queueForTest(t, rdb, "linear")
	ctx := t.Context()

	if _, err := q.Accept(ctx, "d-1", "payload"); err != nil {
		t.Fatalf("accept: %v", err)
	}
	first, err := q.Read(ctx, "worker-a", 10, 0)
	if err != nil || len(first) != 1 {
		t.Fatalf("first read = %+v, %v; want one entry", first, err)
	}
	// A second consumer sees nothing new while the entry is pending.
	if again, err := q.Read(ctx, "worker-b", 10, 10*time.Millisecond); err != nil || len(again) != 0 {
		t.Fatalf("second consumer read = %+v, %v; want nothing", again, err)
	}
	// The holder's own pending list still has it after a crash.
	if pending, err := q.ReadPending(ctx, "worker-a", 10); err != nil || len(pending) != 1 {
		t.Fatalf("pending = %+v, %v; want the held entry", pending, err)
	}
	// Once idle, another consumer takes it over, and the old owner's
	// heartbeat no longer holds it.
	time.Sleep(20 * time.Millisecond)
	claimed, err := q.Claim(ctx, "worker-b", 10*time.Millisecond, 10)
	if err != nil || len(claimed) != 1 || claimed[0].ID != first[0].ID {
		t.Fatalf("claim = %+v, %v; want the idle entry", claimed, err)
	}
	if err := q.Touch(ctx, "worker-a", first[0].ID); err == nil {
		t.Fatalf("old owner's touch succeeded after the entry was reclaimed")
	}
	if err := q.Touch(ctx, "worker-b", first[0].ID); err != nil {
		t.Fatalf("new owner's touch: %v", err)
	}
	if err := q.Ack(ctx, first[0].ID); err != nil {
		t.Fatalf("ack: %v", err)
	}
	if pending, err := q.ReadPending(ctx, "worker-b", 10); err != nil || len(pending) != 0 {
		t.Fatalf("pending after ack = %+v, %v; want nothing", pending, err)
	}
}
