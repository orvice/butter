package plane

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func planeTestRedis(t *testing.T) *redis.Client {
	t.Helper()
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		t.Skip("REDIS_ADDR is required for the real Redis control contract")
	}
	// Keep the Plane stream fixtures separate from other package suites.
	client := redis.NewClient(&redis.Options{Addr: addr, DB: 12})
	if err := client.Ping(t.Context()).Err(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func controlFixture(t *testing.T, backend string) (RunControl, *Event, func()) {
	t.Helper()
	e := &Event{WorkspaceID: "test-" + uuid.NewString(), AppID: "a", InstallationID: "i",
		AgentRunID: "r", ActivityID: "prompt", Prompt: "work"}
	if backend == "memory" {
		c := NewMemoryRunControl()
		return c, e, func() {
			c.mu.Lock()
			defer c.mu.Unlock()
			delete(c.runs, controlKey(e))
		}
	}
	client := planeTestRedis(t)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		keys, _ := client.Keys(ctx, controlPrefix+e.WorkspaceID+":*").Result()
		if len(keys) > 0 {
			_ = client.Del(ctx, keys...).Err()
		}
	})
	return NewRedisRunControl(client), e, func() { _ = client.Del(t.Context(), controlKey(e)).Err() }
}

func TestRunControlContract(t *testing.T) {
	for _, backend := range []string{"memory", "redis"} {
		t.Run(backend, func(t *testing.T) {
			t.Run("StopBeforeFinishWinsAndWaitsForSettlement", func(t *testing.T) {
				c, e, _ := controlFixture(t, backend)
				if err := c.Begin(t.Context(), e, "holder"); err != nil {
					t.Fatal(err)
				}
				target, err := c.RequestStop(t.Context(), e)
				if err != nil || target != "holder" {
					t.Fatalf("Stop target=%q err=%v", target, err)
				}
				if stopped, err := c.Renew(t.Context(), e, "holder"); err != nil || !stopped {
					t.Fatalf("persistent marker: stopped=%v err=%v", stopped, err)
				}
				state, err := c.Finish(t.Context(), e, "holder", RunStateSucceeded)
				if err != nil || state != RunStateStopped {
					t.Fatalf("Finish=%s err=%v", state, err)
				}
				if _, done, err := c.Outcome(t.Context(), e, target); err != nil || done {
					t.Fatalf("confirmed before End: done=%v err=%v", done, err)
				}
				if err := c.End(t.Context(), e, "holder", state); err != nil {
					t.Fatal(err)
				}
				if got, done, err := c.Outcome(t.Context(), e, target); err != nil || !done || got != RunStateStopped {
					t.Fatalf("Outcome=%s done=%v err=%v", got, done, err)
				}
				// The cutoff also covers an older Prompt no Worker read.
				old := *e
				old.ActivityID = "unread-old"
				if err := c.Begin(t.Context(), &old, "old"); !errors.Is(err, ErrPromptDiscarded) {
					t.Fatalf("unread old prompt=%v", err)
				}
				next := *e
				next.ActivityID, next.AcceptedOrder = "new-after-stop", 1<<40
				if err := c.Begin(t.Context(), &next, "next"); err != nil {
					t.Fatal(err)
				}
				if stopped, err := c.Renew(t.Context(), &next, "next"); err != nil || stopped {
					t.Fatalf("old Stop reached new holder: stopped=%v err=%v", stopped, err)
				}
				// Settlement survives replacement of the current holder.
				if got, done, err := c.Outcome(t.Context(), e, target); err != nil || !done || got != RunStateStopped {
					t.Fatalf("old Outcome=%s done=%v err=%v", got, done, err)
				}
			})
			t.Run("FinishBeforeStopCannotBeRewritten", func(t *testing.T) {
				c, e, _ := controlFixture(t, backend)
				if err := c.Begin(t.Context(), e, "holder"); err != nil {
					t.Fatal(err)
				}
				if state, err := c.Finish(t.Context(), e, "holder", RunStateSucceeded); err != nil || state != RunStateSucceeded {
					t.Fatalf("Finish=%s err=%v", state, err)
				}
				if target, err := c.RequestStop(t.Context(), e); err != nil || target != "" {
					t.Fatalf("late Stop target=%s err=%v", target, err)
				}
				if state, err := c.Finish(t.Context(), e, "holder", RunStateFailed); err != nil || state != RunStateSucceeded {
					t.Fatalf("terminal result changed: %s %v", state, err)
				}
			})
			t.Run("LostHolderCannotWriteOrRerun", func(t *testing.T) {
				c, e, steal := controlFixture(t, backend)
				if err := c.Begin(t.Context(), e, "old"); err != nil {
					t.Fatal(err)
				}
				steal()
				if err := c.Begin(t.Context(), e, "retry"); !errors.Is(err, ErrTurnConsumed) {
					t.Fatalf("uncertain event ran again: %v", err)
				}
				next := *e
				next.ActivityID = "next"
				if err := c.Begin(t.Context(), &next, "new"); err != nil {
					t.Fatal(err)
				}
				if _, err := c.Renew(t.Context(), e, "old"); !errors.Is(err, ErrRunLeaseLost) {
					t.Fatalf("old Renew=%v", err)
				}
				if _, err := c.Finish(t.Context(), e, "old", RunStateSucceeded); !errors.Is(err, ErrRunLeaseLost) {
					t.Fatalf("old Finish=%v", err)
				}
				if err := c.End(t.Context(), e, "old", RunStateStopped); !errors.Is(err, ErrRunLeaseLost) {
					t.Fatalf("old End=%v", err)
				}
				if _, _, err := c.Outcome(t.Context(), e, "old"); !errors.Is(err, ErrRunLeaseLost) {
					t.Fatalf("lost holder falsely confirmed: %v", err)
				}
			})
			t.Run("ConcurrentStopAndFinishAgree", func(t *testing.T) {
				c, e, _ := controlFixture(t, backend)
				for i := 0; i < 30; i++ {
					e.AgentRunID = uuid.NewString()
					if err := c.Begin(t.Context(), e, "holder"); err != nil {
						t.Fatal(err)
					}
					var wg sync.WaitGroup
					wg.Add(2)
					var target string
					var state RunState
					var stopErr, finishErr error
					go func() { defer wg.Done(); target, stopErr = c.RequestStop(t.Context(), e) }()
					go func() { defer wg.Done(); state, finishErr = c.Finish(t.Context(), e, "holder", RunStateSucceeded) }()
					wg.Wait()
					if stopErr != nil || finishErr != nil || (target != "") != (state == RunStateStopped) {
						t.Fatalf("Stop=%q/%v Finish=%s/%v", target, stopErr, state, finishErr)
					}
				}
			})
		})
	}
}
