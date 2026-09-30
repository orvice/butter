package linear

// Session coordination contract (ADR-0015 §6, #366): one suite, run against
// the in-memory coordinator and — when REDIS_ADDR is set — against Redis, so
// the Lua scripts are held to the same guarantees.

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type coordinatorFactory func(t *testing.T) (SessionCoordinator, func(key string))

func memoryCoordinator(*testing.T) (SessionCoordinator, func(string)) {
	c := NewMemoryCoordinator()
	return c, c.Steal
}

func redisCoordinator(t *testing.T) (SessionCoordinator, func(string)) {
	t.Helper()
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		t.Skip("REDIS_ADDR is required for the Redis coordination contract")
	}
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	if err := rdb.Ping(t.Context()).Err(); err != nil {
		t.Fatalf("ping redis: %v", err)
	}
	c := NewRedisCoordinator(rdb, "pod-"+uuid.NewString(), 300*time.Millisecond)
	c.prefix = "butter:test:" + uuid.NewString() + ":"
	t.Cleanup(func() {
		keys, _ := rdb.Keys(context.Background(), c.prefix+"*").Result()
		if len(keys) > 0 {
			_ = rdb.Del(context.Background(), keys...).Err()
		}
		_ = rdb.Close()
	})
	steal := func(key string) {
		_ = rdb.Set(context.Background(), c.keys(key)[0], "someone-else", time.Minute).Err()
	}
	return c, steal
}

func TestSessionCoordinationContract(t *testing.T) {
	for name, factory := range map[string]coordinatorFactory{"memory": memoryCoordinator, "redis": redisCoordinator} {
		t.Run(name, func(t *testing.T) { runCoordinationContract(t, factory) })
	}
}

func followUp(n int) FollowUp {
	return FollowUp{DeliveryID: fmt.Sprintf("d-%d", n), Text: fmt.Sprintf("message %d", n)}
}

func runCoordinationContract(t *testing.T, factory coordinatorFactory) {
	t.Run("FollowUpsQueueBehindTheHolderAndDrainInOrder", func(t *testing.T) {
		c, _ := factory(t)
		ctx := t.Context()
		hold, backlog, err := c.EnqueueOrAcquire(ctx, "s", followUp(0))
		if err != nil || hold == nil || len(backlog) != 0 {
			t.Fatalf("first EnqueueOrAcquire = %v, %v, %v; want the lease and no backlog", hold, backlog, err)
		}
		for i := 1; i <= 3; i++ {
			queued, _, err := c.EnqueueOrAcquire(ctx, "s", followUp(i))
			if err != nil || queued != nil {
				t.Fatalf("EnqueueOrAcquire(%d) = %v, %v; want it queued", i, queued, err)
			}
		}
		drained, err := hold.ReleaseOrDrain(ctx)
		if err != nil || len(drained) != 3 || drained[0].Text != "message 1" || drained[2].Text != "message 3" {
			t.Fatalf("ReleaseOrDrain = %+v, %v; want messages 1-3 in order", drained, err)
		}
		// The holder kept the lease: a new message still queues.
		if queued, _, _ := c.EnqueueOrAcquire(ctx, "s", followUp(4)); queued != nil {
			t.Fatal("a message acquired the session while the holder drained it")
		}
		if drained, _ := hold.ReleaseOrDrain(ctx); len(drained) != 1 {
			t.Fatalf("second drain = %+v, want message 4", drained)
		}
		if drained, err := hold.ReleaseOrDrain(ctx); err != nil || len(drained) != 0 {
			t.Fatalf("final ReleaseOrDrain = %+v, %v; want a release", drained, err)
		}
		next, _, err := c.EnqueueOrAcquire(ctx, "s", followUp(5))
		if err != nil || next == nil {
			t.Fatalf("after release EnqueueOrAcquire = %v, %v; want the lease", next, err)
		}
		next.Abandon()
	})

	t.Run("NoFollowUpIsLostOrRunConcurrently", func(t *testing.T) {
		c, _ := factory(t)
		ctx := t.Context()
		const senders = 40
		var active, maxActive atomic.Int32
		var mu sync.Mutex
		ran := map[string]int{}
		run := func(items []FollowUp) {
			n := active.Add(1)
			for {
				m := maxActive.Load()
				if n <= m || maxActive.CompareAndSwap(m, n) {
					break
				}
			}
			mu.Lock()
			for _, item := range items {
				ran[item.DeliveryID]++
			}
			mu.Unlock()
			time.Sleep(time.Millisecond)
			active.Add(-1)
		}
		var wg sync.WaitGroup
		for i := 0; i < senders; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				item := followUp(i)
				hold, backlog, err := c.EnqueueOrAcquire(ctx, "s", item)
				if err != nil {
					t.Errorf("EnqueueOrAcquire: %v", err)
					return
				}
				if hold == nil {
					return
				}
				items := append(backlog, item)
				for len(items) > 0 {
					run(items)
					if items, err = hold.ReleaseOrDrain(ctx); err != nil {
						t.Errorf("ReleaseOrDrain: %v", err)
						return
					}
				}
			}()
		}
		wg.Wait()
		if maxActive.Load() != 1 {
			t.Fatalf("turns ran concurrently: max %d at once", maxActive.Load())
		}
		for i := 0; i < senders; i++ {
			if n := ran[fmt.Sprintf("d-%d", i)]; n != 1 {
				t.Fatalf("follow-up %d ran %d times, want exactly once", i, n)
			}
		}
	})

	t.Run("AnAbandonedSessionLeavesItsBacklogToTheNextHolder", func(t *testing.T) {
		c, _ := factory(t)
		ctx := t.Context()
		hold, _, _ := c.EnqueueOrAcquire(ctx, "s", followUp(0))
		c.EnqueueOrAcquire(ctx, "s", followUp(1))
		hold.Abandon()
		next, backlog, err := c.TryAcquire(ctx, "s")
		if err != nil || next == nil || len(backlog) != 1 || backlog[0].Text != "message 1" {
			t.Fatalf("TryAcquire = %v, %+v, %v; want the lease and the backlog", next, backlog, err)
		}
		if busy, _, _ := c.TryAcquire(ctx, "s"); busy != nil {
			t.Fatal("TryAcquire took a held session")
		}
		next.Abandon()
	})

	t.Run("LosingTheLeaseCancelsTheHolder", func(t *testing.T) {
		c, steal := factory(t)
		ctx := t.Context()
		hold, _, _ := c.EnqueueOrAcquire(ctx, "s", followUp(0))
		steal("s")
		select {
		case <-hold.Context().Done():
		case <-time.After(2 * time.Second):
			t.Fatal("the holder's context was not cancelled after losing the lease")
		}
		if _, err := hold.ReleaseOrDrain(ctx); err == nil {
			t.Fatal("ReleaseOrDrain succeeded after the lease was lost")
		}
	})

	t.Run("AStopReachesTheHolderAndDiscardsTheQueue", func(t *testing.T) {
		c, _ := factory(t)
		ctx := t.Context()
		hold, _, _ := c.EnqueueOrAcquire(ctx, "s", followUp(0))
		c.EnqueueOrAcquire(ctx, "s", followUp(1))
		held, discarded, err := c.RequestStop(ctx, "s")
		if err != nil || !held || len(discarded) != 1 || discarded[0].Text != "message 1" {
			t.Fatalf("RequestStop = %v, %+v, %v; want held with the queued follow-up discarded", held, discarded, err)
		}
		select {
		case <-hold.Stopped():
		case <-time.After(2 * time.Second):
			t.Fatal("the holder was not told to stop")
		}
		if drained, _ := hold.ReleaseOrDrain(ctx); len(drained) != 0 {
			t.Fatalf("drained after stop = %+v, want nothing", drained)
		}
	})

	t.Run("AStopWithNothingRunningReportsIdle", func(t *testing.T) {
		c, _ := factory(t)
		held, discarded, err := c.RequestStop(t.Context(), "s")
		if err != nil || held || len(discarded) != 0 {
			t.Fatalf("RequestStop = %v, %+v, %v; want idle", held, discarded, err)
		}
	})

	t.Run("AStopRightAfterTheTurnStartsIsHonored", func(t *testing.T) {
		c, _ := factory(t)
		ctx := t.Context()
		hold, _, _ := c.EnqueueOrAcquire(ctx, "s", followUp(0))
		if held, _, _ := c.RequestStop(ctx, "s"); !held {
			t.Fatal("RequestStop did not see the holder")
		}
		select {
		case <-hold.Stopped():
		case <-time.After(2 * time.Second):
			t.Fatal("a stop right after the turn started was lost")
		}
		hold.Abandon()
	})

	t.Run("AStopDoesNotCarryOverToTheNextHolderOrTurn", func(t *testing.T) {
		c, _ := factory(t)
		ctx := t.Context()
		hold, _, _ := c.EnqueueOrAcquire(ctx, "s", followUp(0))
		c.RequestStop(ctx, "s")
		<-hold.Stopped()
		hold.AcknowledgeStop(ctx)
		select {
		case <-hold.Stopped():
			t.Fatal("an acknowledged stop still reads as stopped")
		case <-time.After(50 * time.Millisecond):
		}
		hold.Abandon()
		next, _, _ := c.TryAcquire(ctx, "s")
		if next == nil {
			t.Fatal("TryAcquire after abandon failed")
		}
		select {
		case <-next.Stopped():
			t.Fatal("the next holder inherited a stop")
		case <-time.After(50 * time.Millisecond):
		}
		next.Abandon()
	})

	t.Run("SessionsAreIndependent", func(t *testing.T) {
		c, _ := factory(t)
		ctx := t.Context()
		a, _, _ := c.EnqueueOrAcquire(ctx, "a", followUp(0))
		b, _, _ := c.EnqueueOrAcquire(ctx, "b", followUp(0))
		if a == nil || b == nil {
			t.Fatal("one session's turn blocked another's")
		}
		a.Abandon()
		b.Abandon()
	})
}

func TestAStopBetweenAcquisitionAndSubscriptionIsCaughtByTheMarker(t *testing.T) {
	sc, _ := redisCoordinator(t)
	c := sc.(*RedisCoordinator)
	ctx := t.Context()
	c.beforeSubscribe = func() {
		// The lease exists but the holder is not listening yet: the nudge
		// goes nowhere and only the marker can carry the stop.
		if held, _, err := c.RequestStop(ctx, "s"); err != nil || !held {
			t.Errorf("RequestStop = %v, %v", held, err)
		}
	}
	hold, _, err := c.EnqueueOrAcquire(ctx, "s", followUp(0))
	if err != nil || hold == nil {
		t.Fatalf("EnqueueOrAcquire = %v, %v", hold, err)
	}
	c.beforeSubscribe = nil
	select {
	case <-hold.Stopped():
	case <-time.After(2 * time.Second):
		t.Fatal("the stop marker was not honored")
	}
	ttl, err := c.rdb.PTTL(ctx, c.stopKey("s")).Result()
	if err != nil || ttl <= 0 || ttl > stopMarkerTTL {
		t.Fatalf("stop marker TTL = %v, %v; want a bounded TTL", ttl, err)
	}
	hold.Abandon()
}
