package liveness

// Liveness registry contract (#390): one suite, run against the in-process
// registry and — when REDIS_ADDR is set — against Redis, so both are held to
// the same guarantees.

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type registryFactory func(t *testing.T, ttl time.Duration) Registry

func memoryRegistry(_ *testing.T, ttl time.Duration) Registry { return NewMemory(ttl) }

func redisRegistry(t *testing.T, ttl time.Duration) Registry {
	t.Helper()
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		t.Skip("REDIS_ADDR is required for the Redis liveness contract")
	}
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	if err := rdb.Ping(t.Context()).Err(); err != nil {
		t.Fatalf("ping redis: %v", err)
	}
	r := NewRedis(rdb, ttl)
	r.prefix = "butter:test:" + uuid.NewString() + ":instance:"
	t.Cleanup(func() {
		keys, _ := rdb.Keys(context.Background(), r.prefix+"*").Result()
		if len(keys) > 0 {
			_ = rdb.Del(context.Background(), keys...).Err()
		}
		_ = rdb.Close()
	})
	return r
}

func TestRegistryContract(t *testing.T) {
	for name, factory := range map[string]registryFactory{"memory": memoryRegistry, "redis": redisRegistry} {
		t.Run(name, func(t *testing.T) { runRegistryContract(t, factory) })
	}
}

// waitUntil polls cond until it holds or the deadline passes.
func waitUntil(t *testing.T, within time.Duration, cond func() bool) bool {
	t.Helper()
	deadline := time.Now().Add(within)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}
		time.Sleep(20 * time.Millisecond)
	}
	return cond()
}

func alive(t *testing.T, r Registry, ids ...string) map[string]bool {
	t.Helper()
	got, err := r.Alive(context.Background(), ids)
	if err != nil {
		t.Fatalf("Alive(%v): %v", ids, err)
	}
	return got
}

func runRegistryContract(t *testing.T, factory registryFactory) {
	t.Run("APublishedInstanceIsAliveAndAnUnknownOneIsNot", func(t *testing.T) {
		r := factory(t, time.Minute)
		if err := r.Publish(t.Context(), "a"); err != nil {
			t.Fatalf("Publish: %v", err)
		}
		got := alive(t, r, "a", "never-published")
		if !got["a"] || got["never-published"] {
			t.Fatalf("Alive = %v, want a alive and never-published gone", got)
		}
		if len(got) != 2 {
			t.Fatalf("Alive answered %d IDs, want one entry per ID asked", len(got))
		}
	})

	t.Run("AskingAboutNoInstancesAnswersNothing", func(t *testing.T) {
		r := factory(t, time.Minute)
		if got := alive(t, r); len(got) != 0 {
			t.Fatalf("Alive() = %v, want empty", got)
		}
	})

	t.Run("APublicationLapsesOneTTLAfterItWasMade", func(t *testing.T) {
		const ttl = 300 * time.Millisecond
		r := factory(t, ttl)
		if r.TTL() != ttl {
			t.Fatalf("TTL() = %v, want %v", r.TTL(), ttl)
		}
		if err := r.Publish(t.Context(), "a"); err != nil {
			t.Fatalf("Publish: %v", err)
		}
		if !waitUntil(t, 3*time.Second, func() bool { return !alive(t, r, "a")["a"] }) {
			t.Fatal("a publication outlived its TTL")
		}
	})

	t.Run("RepublishingExtendsThePublication", func(t *testing.T) {
		const ttl = 400 * time.Millisecond
		r := factory(t, ttl)
		if err := r.Publish(t.Context(), "a"); err != nil {
			t.Fatalf("Publish: %v", err)
		}
		time.Sleep(ttl * 3 / 4)
		if err := r.Publish(t.Context(), "a"); err != nil {
			t.Fatalf("second Publish: %v", err)
		}
		// Past the first publication's TTL, inside the second's.
		time.Sleep(ttl / 2)
		if !alive(t, r, "a")["a"] {
			t.Fatal("a republished instance lapsed on its first publication's TTL")
		}
	})

	// The two-process case behind #390: B starting, or sweeping, never makes
	// A read as gone; only A ceasing to renew does.
	t.Run("TwoInstancesStayAliveUntilOneStopsRenewing", func(t *testing.T) {
		const ttl = 300 * time.Millisecond
		r := factory(t, ttl)
		ctxA, crashA := context.WithCancel(t.Context())
		defer crashA()
		if err := Keep(ctxA, r, "a"); err != nil {
			t.Fatalf("Keep(a): %v", err)
		}
		if err := Keep(t.Context(), r, "b"); err != nil {
			t.Fatalf("Keep(b): %v", err)
		}
		// Several TTLs later both are still alive: Keep renews them.
		time.Sleep(3 * ttl)
		if got := alive(t, r, "a", "b"); !got["a"] || !got["b"] {
			t.Fatalf("Alive = %v, want both alive while they renew", got)
		}

		crashA() // A stops renewing, as a crashed process does.
		if !waitUntil(t, 3*time.Second, func() bool { return !alive(t, r, "a")["a"] }) {
			t.Fatal("A still reads as alive long after it stopped renewing")
		}
		if !alive(t, r, "b")["b"] {
			t.Fatal("B lapsed although it kept renewing")
		}
	})
}

func TestNewRedisWithoutAClientIsNil(t *testing.T) {
	if NewRedis(nil, time.Minute) != nil {
		t.Fatal("NewRedis(nil) built a registry")
	}
}

func TestRedisKeysAreNamespacedPerInstance(t *testing.T) {
	r := NewRedis(redis.NewClient(&redis.Options{Addr: "127.0.0.1:0"}), time.Minute)
	t.Cleanup(func() { _ = r.rdb.Close() })
	if got := r.key("abc"); got != "butter:instance:abc" {
		t.Fatalf("key = %q, want butter:instance:abc", got)
	}
}
