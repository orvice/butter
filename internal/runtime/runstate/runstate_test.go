package runstate

// Run state contract (#401, ADR-0016 decision 6) and the Stop on it (#402,
// decisions 3 and 4): one suite, run against the in-process store and — when
// REDIS_ADDR is set — against Redis, so both are held to the same guarantees.

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
	store := NewRedis(rdb, prefix, ttl)
	t.Cleanup(func() {
		_ = store.Close()
		keys, _ := rdb.Keys(context.Background(), prefix+"*").Result()
		if len(keys) > 0 {
			_ = rdb.Del(context.Background(), keys...).Err()
		}
		_ = rdb.Close()
	})
	return store
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
	return State{RunID: "run-" + n, InvocationID: "inv-" + n, EventCount: 7, Detached: true, LeaseToken: "tok-" + n}
}

// nudgedWithin reports whether nudged is closed within d.
func nudgedWithin(nudged <-chan struct{}, d time.Duration) bool {
	select {
	case <-nudged:
		return true
	case <-time.After(d):
		return false
	}
}

func stop(t *testing.T, s Store, thread, invocationID string) (State, bool) {
	t.Helper()
	st, accepted, err := s.Stop(t.Context(), thread, invocationID)
	if err != nil {
		t.Fatalf("Stop(%s, %q): %v", thread, invocationID, err)
	}
	return st, accepted
}

func claim(t *testing.T, s Store, thread, invocationID string) Claim {
	t.Helper()
	c, err := s.Claim(t.Context(), thread, invocationID)
	if err != nil {
		t.Fatalf("Claim(%s, %s): %v", thread, invocationID, err)
	}
	return c
}

// lostNudges is a store whose nudges never arrive, as when the process's
// subscription is down: only the marker can tell a run it was stopped.
type lostNudges struct{ Store }

func (lostNudges) Watch(context.Context, string, string) (<-chan struct{}, func()) {
	return nil, func() {}
}

// stopOnBegin accepts a Stop the moment a run's state exists, before Keep
// returns: a Stop racing the run's start.
type stopOnBegin struct{ Store }

func (s stopOnBegin) Begin(ctx context.Context, thread string, st State) error {
	if err := s.Store.Begin(ctx, thread, st); err != nil {
		return err
	}
	_, _, err := s.Stop(ctx, thread, "")
	return err
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
		if renewed, _, err := s.Renew(t.Context(), "t", "inv-a"); err != nil || !renewed {
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
		if renewed, _, err := s.Renew(t.Context(), "t", "inv-a"); err != nil || renewed {
			t.Fatalf("Renew of a lapsed state = %v, %v; want refused", renewed, err)
		}
		if err := s.Begin(t.Context(), "t", run("b")); err != nil {
			t.Fatalf("Begin b: %v", err)
		}
		if err := s.End(t.Context(), "t", "inv-b"); err != nil {
			t.Fatalf("End b: %v", err)
		}
		if renewed, _, err := s.Renew(t.Context(), "t", "inv-b"); err != nil || renewed {
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
		if renewed, _, err := s.Renew(t.Context(), "t", "inv-a"); err != nil || renewed {
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

	// A Stop (decision 4) is accepted on the run state, in one step with its
	// nudge, and ordered with the run's claim of its end (decision 3).

	t.Run("AStopReachesTheDetachedRunHoldingTheThread", func(t *testing.T) {
		s := factory(t, time.Minute)
		nudged, unwatch := s.Watch(t.Context(), "t", "tok-a")
		defer unwatch()
		if err := s.Begin(t.Context(), "t", run("a")); err != nil {
			t.Fatalf("Begin: %v", err)
		}
		if st, accepted := stop(t, s, "t", ""); !accepted || st != run("a") {
			t.Fatalf("Stop = %+v, %v; want run a stopped", st, accepted)
		}
		if !nudgedWithin(nudged, 5*time.Second) {
			t.Fatal("the run was not nudged")
		}
		// The marker a renewal finds, for a nudge that never arrived.
		if renewed, stopped, err := s.Renew(t.Context(), "t", "inv-a"); err != nil || !renewed || !stopped {
			t.Fatalf("Renew = %v, %v, %v; want renewed and stopped", renewed, stopped, err)
		}
		// Stopping again before the claim is accepted again.
		if _, accepted := stop(t, s, "t", ""); !accepted {
			t.Fatal("a second Stop before the claim was refused")
		}
		if c := claim(t, s, "t", "inv-a"); !c.Held || !c.Stopped {
			t.Fatalf("Claim = %+v; an accepted Stop must end the run stopped", c)
		}
		// Still the run's state, for reads, until the run ends it.
		if st, ok := get(t, s, "t"); !ok || st != run("a") {
			t.Fatalf("Get = %+v, %v; want run a's state", st, ok)
		}
	})

	t.Run("AStopFindsNothingOnAnIdleThreadOrARunWithoutTheOptIn", func(t *testing.T) {
		s := factory(t, time.Minute)
		if _, accepted := stop(t, s, "t", ""); accepted {
			t.Fatal("a Stop on an idle thread was accepted")
		}
		inRequest := run("a")
		inRequest.Detached = false
		nudged, unwatch := s.Watch(t.Context(), "t", inRequest.LeaseToken)
		defer unwatch()
		if err := s.Begin(t.Context(), "t", inRequest); err != nil {
			t.Fatalf("Begin: %v", err)
		}
		if _, accepted := stop(t, s, "t", ""); accepted {
			t.Fatal("a Stop reached a run without the opt-in")
		}
		if _, stopped, err := s.Renew(t.Context(), "t", "inv-a"); err != nil || stopped {
			t.Fatalf("Renew = stopped %v, %v; want no marker", stopped, err)
		}
		if nudgedWithin(nudged, 200*time.Millisecond) {
			t.Fatal("a run without the opt-in was nudged")
		}
		if c := claim(t, s, "t", "inv-a"); !c.Held || c.Stopped {
			t.Fatalf("Claim = %+v; want held, not stopped", c)
		}
	})

	t.Run("AStopAfterTheClaimFindsNothingRunning", func(t *testing.T) {
		s := factory(t, time.Minute)
		nudged, unwatch := s.Watch(t.Context(), "t", "tok-a")
		defer unwatch()
		if err := s.Begin(t.Context(), "t", run("a")); err != nil {
			t.Fatalf("Begin: %v", err)
		}
		if c := claim(t, s, "t", "inv-a"); !c.Held || c.Stopped {
			t.Fatalf("Claim = %+v; want held, not stopped", c)
		}
		if _, accepted := stop(t, s, "t", ""); accepted {
			t.Fatal("a Stop after the claim was accepted")
		}
		if _, stopped, _ := s.Renew(t.Context(), "t", "inv-a"); stopped {
			t.Fatal("a Stop refused after the claim left a marker")
		}
		if nudgedWithin(nudged, 200*time.Millisecond) {
			t.Fatal("a Stop refused after the claim nudged the run")
		}
	})

	t.Run("AStopForAnInvocationReachesOnlyThatRun", func(t *testing.T) {
		s := factory(t, time.Minute)
		if err := s.Begin(t.Context(), "t", run("b")); err != nil {
			t.Fatalf("Begin: %v", err)
		}
		if _, accepted := stop(t, s, "t", "inv-a"); accepted {
			t.Fatal("a Stop for run a reached run b")
		}
		if _, stopped, _ := s.Renew(t.Context(), "t", "inv-b"); stopped {
			t.Fatal("a refused Stop marked run b")
		}
		if st, accepted := stop(t, s, "t", "inv-b"); !accepted || st.InvocationID != "inv-b" {
			t.Fatalf("Stop for run b = %+v, %v", st, accepted)
		}
	})

	t.Run("AStopNeverReachesALaterRun", func(t *testing.T) {
		s := factory(t, time.Minute)
		if err := s.Begin(t.Context(), "t", run("a")); err != nil {
			t.Fatalf("Begin a: %v", err)
		}
		if _, accepted := stop(t, s, "t", ""); !accepted {
			t.Fatal("the Stop for run a was refused")
		}
		// Run b takes the thread, over the state run a left behind.
		nudgedB, unwatchB := s.Watch(t.Context(), "t", "tok-b")
		defer unwatchB()
		if err := s.Begin(t.Context(), "t", run("b")); err != nil {
			t.Fatalf("Begin b: %v", err)
		}
		if renewed, stopped, err := s.Renew(t.Context(), "t", "inv-b"); err != nil || !renewed || stopped {
			t.Fatalf("Renew b = %v, %v, %v; run a's Stop reached run b", renewed, stopped, err)
		}
		if nudgedWithin(nudgedB, 200*time.Millisecond) {
			t.Fatal("run a's Stop nudged run b")
		}
		// A Stop for run b nudges b, and nobody waiting with a's token.
		nudgedA, unwatchA := s.Watch(t.Context(), "t", "tok-a")
		defer unwatchA()
		if _, accepted := stop(t, s, "t", ""); !accepted {
			t.Fatal("the Stop for run b was refused")
		}
		if !nudgedWithin(nudgedB, 5*time.Second) {
			t.Fatal("run b was not nudged")
		}
		if nudgedWithin(nudgedA, 200*time.Millisecond) {
			t.Fatal("run b's Stop nudged a watch with run a's token")
		}
		if c := claim(t, s, "t", "inv-b"); !c.Held || !c.Stopped {
			t.Fatalf("Claim b = %+v; want stopped", c)
		}
	})

	t.Run("AClaimOfAStateNoLongerTheRunsHoldsNothing", func(t *testing.T) {
		s := factory(t, time.Minute)
		if c := claim(t, s, "t", "inv-a"); c.Held || c.Stopped {
			t.Fatalf("Claim on an idle thread = %+v", c)
		}
		if err := s.Begin(t.Context(), "t", run("b")); err != nil {
			t.Fatalf("Begin: %v", err)
		}
		if c := claim(t, s, "t", "inv-a"); c.Held {
			t.Fatalf("run a claimed run b's state: %+v", c)
		}
		// Run b's state is untouched: a Stop still reaches b.
		if _, accepted := stop(t, s, "t", ""); !accepted {
			t.Fatal("run a's claim attempt closed run b's window")
		}
	})

	t.Run("DropRemovesWhicheverRunsStateItIs", func(t *testing.T) {
		s := factory(t, time.Minute)
		if err := s.Begin(t.Context(), "t", run("a")); err != nil {
			t.Fatalf("Begin: %v", err)
		}
		if err := s.Drop(t.Context(), "t"); err != nil {
			t.Fatalf("Drop: %v", err)
		}
		if st, ok := get(t, s, "t"); ok {
			t.Fatalf("Get after Drop = %+v", st)
		}
		if _, accepted := stop(t, s, "t", ""); accepted {
			t.Fatal("a Stop found a dropped run")
		}
		if err := s.Drop(t.Context(), "t"); err != nil {
			t.Fatalf("Drop of an idle thread: %v", err)
		}
	})

	// Keep tells a Detached Run of its Stop, by nudge or by marker.

	t.Run("AKeptRunLearnsOfItsStopFromTheNudge", func(t *testing.T) {
		// Renewals are a minute apart: only the nudge can tell the run.
		s := factory(t, time.Minute)
		kept, err := Keep(t.Context(), s, "t", run("a"))
		if err != nil {
			t.Fatalf("Keep: %v", err)
		}
		defer func() { _ = kept.End(context.Background()) }()
		if _, accepted := stop(t, s, "t", ""); !accepted {
			t.Fatal("the Stop was refused")
		}
		if !nudgedWithin(kept.Stopped(), 5*time.Second) {
			t.Fatal("the kept run never learned of its Stop")
		}
		if c, err := kept.Claim(t.Context()); err != nil || !c.Held || !c.Stopped {
			t.Fatalf("Claim = %+v, %v; want stopped", c, err)
		}
	})

	t.Run("AKeptRunLearnsOfItsStopAtARenewalWhenTheNudgeIsLost", func(t *testing.T) {
		const ttl = 300 * time.Millisecond
		s := lostNudges{factory(t, ttl)}
		kept, err := Keep(t.Context(), s, "t", run("a"))
		if err != nil {
			t.Fatalf("Keep: %v", err)
		}
		defer func() { _ = kept.End(context.Background()) }()
		if _, accepted := stop(t, s, "t", ""); !accepted {
			t.Fatal("the Stop was refused")
		}
		if !nudgedWithin(kept.Stopped(), 3*time.Second) {
			t.Fatal("the run never found the marker at its renewals")
		}
	})

	t.Run("AStopThatRacesTheRunsStartIsNotLost", func(t *testing.T) {
		// The run subscribed before its state existed, so the nudge of a Stop
		// accepted the moment it did reaches it, a minute before any renewal
		// would.
		s := stopOnBegin{factory(t, time.Minute)}
		kept, err := Keep(t.Context(), s, "t", run("a"))
		if err != nil {
			t.Fatalf("Keep: %v", err)
		}
		defer func() { _ = kept.End(context.Background()) }()
		if !nudgedWithin(kept.Stopped(), 5*time.Second) {
			t.Fatal("a Stop accepted as the run started was lost")
		}
	})

	t.Run("AKeptRunWithoutTheOptInIsNeverStopped", func(t *testing.T) {
		const ttl = 300 * time.Millisecond
		s := factory(t, ttl)
		inRequest := run("a")
		inRequest.Detached = false
		kept, err := Keep(t.Context(), s, "t", inRequest)
		if err != nil {
			t.Fatalf("Keep: %v", err)
		}
		defer func() { _ = kept.End(context.Background()) }()
		if _, accepted := stop(t, s, "t", ""); accepted {
			t.Fatal("a Stop reached a run without the opt-in")
		}
		if nudgedWithin(kept.Stopped(), ttl) {
			t.Fatal("a run without the opt-in learned of a Stop")
		}
	})

	t.Run("EndingAKeptRunEndsItsWatch", func(t *testing.T) {
		s := factory(t, time.Minute)
		kept, err := Keep(t.Context(), s, "t", run("a"))
		if err != nil {
			t.Fatalf("Keep: %v", err)
		}
		if err := kept.End(t.Context()); err != nil {
			t.Fatalf("End: %v", err)
		}
		// The next run on the thread is stopped; the ended one hears nothing.
		if err := s.Begin(t.Context(), "t", run("a")); err != nil {
			t.Fatalf("Begin: %v", err)
		}
		if _, accepted := stop(t, s, "t", ""); !accepted {
			t.Fatal("the Stop was refused")
		}
		if nudgedWithin(kept.Stopped(), 200*time.Millisecond) {
			t.Fatal("an ended run still received nudges")
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
	if got := r.nudgeChannel("u1:agui-t-1"); got != "butter:agui:run:session:stop:u1:agui-t-1" {
		t.Fatalf("nudge channel = %q", got)
	}
}

func TestEscapeGlobQuotesPatternCharacters(t *testing.T) {
	if got := escapeGlob(`a*b?c[d]e\f:`); got != `a\*b\?c\[d\]e\\f:` {
		t.Fatalf("escapeGlob = %q", got)
	}
}
