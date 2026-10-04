package runstate

// Run state contract (#401, ADR-0016 decision 6): one suite, run against the
// in-process store and — when REDIS_ADDR is set — against Redis, so both are
// held to the same guarantees.

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type storeFactory func(t *testing.T, ttl time.Duration) Store

func memoryStore(_ *testing.T, ttl time.Duration) Store { return NewMemory(ttl) }

func redisStore(t *testing.T, ttl time.Duration) Store {
	t.Helper()
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		t.Skip("REDIS_ADDR is required for the Redis run state contract")
	}
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	if err := rdb.Ping(t.Context()).Err(); err != nil {
		t.Fatalf("ping redis: %v", err)
	}
	prefix := "butter:test:" + uuid.NewString() + ":run:"
	t.Cleanup(func() {
		keys, _ := rdb.Keys(context.Background(), prefix+"*").Result()
		if len(keys) > 0 {
			_ = rdb.Del(context.Background(), keys...).Err()
		}
		_ = rdb.Close()
	})
	return NewRedis(rdb, prefix, ttl)
}

func TestStoreContract(t *testing.T) {
	for name, factory := range map[string]storeFactory{"memory": memoryStore, "redis": redisStore} {
		t.Run(name, func(t *testing.T) { runStoreContract(t, factory) })
	}
}

// waitUntil polls cond until it holds or the deadline passes.
func waitUntil(within time.Duration, cond func() bool) bool {
	deadline := time.Now().Add(within)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}
		time.Sleep(20 * time.Millisecond)
	}
	return cond()
}

func run(n string) State {
	return State{RunID: "run-" + n, InvocationID: "inv-" + n, EventCount: 7, Detached: true}
}

func get(t *testing.T, s Store, thread string) (State, bool) {
	t.Helper()
	st, ok, err := s.Get(context.Background(), thread)
	if err != nil {
		t.Fatalf("Get(%s): %v", thread, err)
	}
	return st, ok
}

func runStoreContract(t *testing.T, factory storeFactory) {
	t.Run("ARecordedRunIsReadBackWholeAndOthersAreIdle", func(t *testing.T) {
		s := factory(t, time.Minute)
		if err := s.Begin(t.Context(), "u1:agui-t-1", run("a")); err != nil {
			t.Fatalf("Begin: %v", err)
		}
		if st, ok := get(t, s, "u1:agui-t-1"); !ok || st != run("a") {
			t.Fatalf("Get = %+v, %v; want %+v", st, ok, run("a"))
		}
		if _, ok := get(t, s, "u1:agui-t-2"); ok {
			t.Fatal("a thread no run started reads as running")
		}
	})

	t.Run("AStateNeedsAnInvocationID", func(t *testing.T) {
		s := factory(t, time.Minute)
		st := run("a")
		st.InvocationID = ""
		if err := s.Begin(t.Context(), "t", st); err == nil {
			t.Fatal("Begin accepted a state without an invocation ID")
		}
	})

	t.Run("AStateLapsesOneTTLAfterItWasWritten", func(t *testing.T) {
		const ttl = 300 * time.Millisecond
		s := factory(t, ttl)
		if s.TTL() != ttl {
			t.Fatalf("TTL() = %v, want %v", s.TTL(), ttl)
		}
		if err := s.Begin(t.Context(), "t", run("a")); err != nil {
			t.Fatalf("Begin: %v", err)
		}
		if !waitUntil(3*time.Second, func() bool { _, ok := get(t, s, "t"); return !ok }) {
			t.Fatal("a run state outlived its TTL without renewal")
		}
	})

	t.Run("RenewingKeepsTheStatePastItsFirstTTL", func(t *testing.T) {
		const ttl = 400 * time.Millisecond
		s := factory(t, ttl)
		if err := s.Begin(t.Context(), "t", run("a")); err != nil {
			t.Fatalf("Begin: %v", err)
		}
		time.Sleep(ttl * 3 / 4)
		if renewed, err := s.Renew(t.Context(), "t", "inv-a"); err != nil || !renewed {
			t.Fatalf("Renew = %v, %v; want renewed", renewed, err)
		}
		// Past the first write's TTL, inside the renewal's.
		time.Sleep(ttl / 2)
		if st, ok := get(t, s, "t"); !ok || st != run("a") {
			t.Fatalf("Get = %+v, %v; want the renewed state", st, ok)
		}
	})

	t.Run("ALapsedOrEndedStateIsNeverRenewedBack", func(t *testing.T) {
		const ttl = 200 * time.Millisecond
		s := factory(t, ttl)
		if err := s.Begin(t.Context(), "t", run("a")); err != nil {
			t.Fatalf("Begin: %v", err)
		}
		if !waitUntil(3*time.Second, func() bool { _, ok := get(t, s, "t"); return !ok }) {
			t.Fatal("the state did not lapse")
		}
		if renewed, err := s.Renew(t.Context(), "t", "inv-a"); err != nil || renewed {
			t.Fatalf("Renew of a lapsed state = %v, %v; want refused", renewed, err)
		}
		if err := s.Begin(t.Context(), "t", run("b")); err != nil {
			t.Fatalf("Begin b: %v", err)
		}
		if err := s.End(t.Context(), "t", "inv-b"); err != nil {
			t.Fatalf("End b: %v", err)
		}
		if renewed, err := s.Renew(t.Context(), "t", "inv-b"); err != nil || renewed {
			t.Fatalf("Renew of an ended state = %v, %v; want refused", renewed, err)
		}
		if st, ok := get(t, s, "t"); ok {
			t.Fatalf("Get = %+v; a lapsed or ended state came back", st)
		}
	})

	t.Run("ARunNeverRenewsOrEndsAnotherRunsState", func(t *testing.T) {
		s := factory(t, time.Minute)
		// Run b took the thread after run a lost it; a still tries to act.
		if err := s.Begin(t.Context(), "t", run("b")); err != nil {
			t.Fatalf("Begin: %v", err)
		}
		if renewed, err := s.Renew(t.Context(), "t", "inv-a"); err != nil || renewed {
			t.Fatalf("Renew by a = %v, %v; want refused", renewed, err)
		}
		if err := s.End(t.Context(), "t", "inv-a"); err != nil {
			t.Fatalf("End by a: %v", err)
		}
		if st, ok := get(t, s, "t"); !ok || st != run("b") {
			t.Fatalf("Get = %+v, %v; want run b's state untouched", st, ok)
		}
	})

	t.Run("BeginReplacesAStateLeftByAnEarlierRun", func(t *testing.T) {
		s := factory(t, time.Minute)
		if err := s.Begin(t.Context(), "t", run("a")); err != nil {
			t.Fatalf("Begin a: %v", err)
		}
		b := State{RunID: "run-b", InvocationID: "inv-b", EventCount: 12}
		if err := s.Begin(t.Context(), "t", b); err != nil {
			t.Fatalf("Begin b: %v", err)
		}
		if st, ok := get(t, s, "t"); !ok || st != b {
			t.Fatalf("Get = %+v, %v; want run b's state alone", st, ok)
		}
	})

	t.Run("EndRemovesTheRunsOwnState", func(t *testing.T) {
		s := factory(t, time.Minute)
		if err := s.Begin(t.Context(), "t", run("a")); err != nil {
			t.Fatalf("Begin: %v", err)
		}
		if err := s.End(t.Context(), "t", "inv-a"); err != nil {
			t.Fatalf("End: %v", err)
		}
		if _, ok := get(t, s, "t"); ok {
			t.Fatal("an ended run still reads as running")
		}
		// Ending again, or ending a thread nobody holds, is harmless.
		if err := s.End(t.Context(), "t", "inv-a"); err != nil {
			t.Fatalf("second End: %v", err)
		}
	})

	t.Run("ThreadsAreIndependent", func(t *testing.T) {
		s := factory(t, time.Minute)
		if err := s.Begin(t.Context(), "u1:agui-t-1", run("a")); err != nil {
			t.Fatalf("Begin t-1: %v", err)
		}
		if err := s.Begin(t.Context(), "u2:agui-t-1", run("b")); err != nil {
			t.Fatalf("Begin t-1 of u2: %v", err)
		}
		if err := s.End(t.Context(), "u2:agui-t-1", "inv-b"); err != nil {
			t.Fatalf("End: %v", err)
		}
		if st, ok := get(t, s, "u1:agui-t-1"); !ok || st.InvocationID != "inv-a" {
			t.Fatalf("Get = %+v, %v; another caller's thread was touched", st, ok)
		}
	})

	// Keep is what a run uses: the state lives as long as the run holds its
	// lease and keeps it, and not a moment longer.

	t.Run("AKeptStateOutlivesItsTTLUntilItEnds", func(t *testing.T) {
		const ttl = 300 * time.Millisecond
		s := factory(t, ttl)
		kept, err := Keep(t.Context(), s, "t", run("a"))
		if err != nil {
			t.Fatalf("Keep: %v", err)
		}
		time.Sleep(3 * ttl)
		if st, ok := get(t, s, "t"); !ok || st != kept.State() {
			t.Fatalf("Get after 3 TTLs = %+v, %v; want the kept state", st, ok)
		}
		if err := kept.End(t.Context()); err != nil {
			t.Fatalf("End: %v", err)
		}
		if _, ok := get(t, s, "t"); ok {
			t.Fatal("an ended kept state still reads as running")
		}
		if err := kept.End(t.Context()); err != nil {
			t.Fatalf("second End: %v", err)
		}
	})

	t.Run("AKeptStateLapsesWithTheLease", func(t *testing.T) {
		const ttl = 300 * time.Millisecond
		s := factory(t, ttl)
		lease, loseLease := context.WithCancel(t.Context())
		kept, err := Keep(lease, s, "t", run("a"))
		if err != nil {
			t.Fatalf("Keep: %v", err)
		}
		defer func() { _ = kept.End(context.Background()) }()
		loseLease()
		if !waitUntil(3*time.Second, func() bool { _, ok := get(t, s, "t"); return !ok }) {
			t.Fatal("the state outlived the lease it was kept with")
		}
	})

	t.Run("AKeptStateNeverComesBackOnceAnotherRunEndedIt", func(t *testing.T) {
		const ttl = 300 * time.Millisecond
		s := factory(t, ttl)
		// Run a lost its thread but has not noticed yet; run b took the
		// thread, ran and ended.
		kept, err := Keep(t.Context(), s, "t", run("a"))
		if err != nil {
			t.Fatalf("Keep: %v", err)
		}
		defer func() { _ = kept.End(context.Background()) }()
		if err := s.Begin(t.Context(), "t", run("b")); err != nil {
			t.Fatalf("Begin b: %v", err)
		}
		if err := s.End(t.Context(), "t", "inv-b"); err != nil {
			t.Fatalf("End b: %v", err)
		}
		// Several of a's renewal intervals later, a has not brought its
		// state back over the thread b used.
		time.Sleep(2 * ttl)
		if st, ok := get(t, s, "t"); ok {
			t.Fatalf("Get = %+v; a run that lost its thread brought its state back", st)
		}
	})

	t.Run("KeepingNeedsAnInvocationID", func(t *testing.T) {
		s := factory(t, time.Minute)
		st := run("a")
		st.InvocationID = ""
		if _, err := Keep(t.Context(), s, "t", st); err == nil {
			t.Fatal("Keep accepted a state without an invocation ID")
		}
	})
}

func TestNewRedisWithoutAClientIsNil(t *testing.T) {
	if NewRedis(nil, "p:", time.Minute) != nil {
		t.Fatal("NewRedis(nil) built a store")
	}
}

func TestRedisKeysSitUnderThePrefix(t *testing.T) {
	r := NewRedis(redis.NewClient(&redis.Options{Addr: "127.0.0.1:0"}), "butter:agui:run:session:", time.Minute)
	t.Cleanup(func() { _ = r.rdb.Close() })
	if got := r.key("u1:agui-t-1"); got != "butter:agui:run:session:u1:agui-t-1" {
		t.Fatalf("key = %q", got)
	}
}
