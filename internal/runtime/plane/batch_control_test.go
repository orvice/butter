package plane

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func restrictedBatchClient(t *testing.T, q *StreamQueue, denied string) *redis.Client {
	t.Helper()
	user, password := "plane-batch-"+uuid.NewString(), uuid.NewString()
	if err := q.client.ACLSetUser(t.Context(), user, "on", ">"+password, "~*", "+@all", "-"+denied).Err(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = q.client.ACLDelUser(context.Background(), user).Err() })
	client := redis.NewClient(&redis.Options{Addr: q.client.Options().Addr, DB: q.client.Options().DB, Username: user, Password: password})
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func TestRedisBatchScriptFailuresDoNotPartiallyCommit(t *testing.T) {
	for _, phase := range []string{"reserve", "start", "end", "recover"} {
		t.Run(phase, func(t *testing.T) {
			q, base := planeQueueFixture(t)
			events := acceptBatch(t, q, base, "first", "second")
			e := events[0]
			source := &redisBatchSource{client: q.client}
			if err := source.prepare(t.Context(), e); err != nil {
				t.Fatal(err)
			}
			control := NewRedisRunControl(q.client)
			if phase == "reserve" {
				denied := NewRedisRunControl(restrictedBatchClient(t, q, "hset"))
				if err := denied.Begin(t.Context(), e, "holder"); err == nil {
					t.Fatal("reservation succeeded without HSET permission")
				}
				if q.client.Exists(t.Context(), controlKey(e)).Val() != 0 {
					t.Fatal("rejected reservation left a holder")
				}
				return
			}
			if err := control.Begin(t.Context(), e, "holder"); err != nil {
				t.Fatal(err)
			}
			if phase == "start" {
				denied := NewRedisRunControl(restrictedBatchClient(t, q, "set"))
				if err := denied.StartBatch(t.Context(), e, "holder"); err == nil {
					t.Fatal("execution started without marker permission")
				}
				for _, member := range e.Members {
					if q.client.Exists(t.Context(), startedKey(member)).Val() != 0 {
						t.Fatal("partial execution boundary")
					}
				}
				if q.client.HGet(t.Context(), controlKey(e), "phase").Val() != "reserved" {
					t.Fatal("rejected execution changed the reservation")
				}
				return
			}
			if err := control.StartBatch(t.Context(), e, "holder"); err != nil {
				t.Fatal(err)
			}
			if phase == "recover" {
				q.client.Del(t.Context(), controlKey(e))
				denied := &redisBatchSource{client: restrictedBatchClient(t, q, "zrem")}
				if err := denied.prepare(t.Context(), e); err == nil {
					t.Fatal("recovery succeeded without cleanup permission")
				}
				if q.client.Exists(t.Context(), eventResultKey(e)).Val() != 0 {
					t.Fatal("rejected recovery partially wrote an outcome")
				}
				return
			}
			if _, err := control.Finish(t.Context(), e, "holder", RunStateSucceeded); err != nil {
				t.Fatal(err)
			}
			denied := NewRedisRunControl(restrictedBatchClient(t, q, "hdel"))
			if err := denied.End(t.Context(), e, "holder", RunStateSucceeded); err == nil {
				t.Fatal("settled without terminal cleanup permission")
			}
			for _, member := range e.Members {
				if q.client.Exists(t.Context(), eventResultKey(member)).Val() != 0 {
					t.Fatal("partial batch settlement")
				}
			}
			if _, done, err := control.Outcome(t.Context(), e, "holder"); err != nil || done {
				t.Fatalf("failed End published completion: done=%v err=%v", done, err)
			}
		})
	}
}

func TestReservedBatchStopAndMemberFencing(t *testing.T) {
	for _, backend := range []string{"memory", "redis"} {
		t.Run(backend, func(t *testing.T) {
			q, base := planeQueueFixture(t)
			events := acceptBatch(t, q, base, "first", "second")
			e := events[0]
			e.Members = []*Event{events[0], events[1]}
			if backend == "memory" {
				// Accepted orders may have gaps due to other Runs.
				events[1].AcceptedOrder += 100
			}
			var control interface {
				RunControl
				StartBatch(context.Context, *Event, string) error
				ReleaseBatch(context.Context, *Event, string) error
			} = NewMemoryRunControl()
			if backend == "redis" {
				control = NewRedisRunControl(q.client)
			}
			if err := control.Begin(t.Context(), e, "holder"); err != nil {
				t.Fatal(err)
			}
			changed := *e
			changed.Members = e.Members[:1]
			if err := control.StartBatch(t.Context(), &changed, "holder"); !errors.Is(err, ErrRunLeaseLost) {
				t.Fatalf("changed batch crossed the reserved boundary: %v", err)
			}
			if _, err := control.Finish(t.Context(), e, "holder", RunStateSucceeded); !errors.Is(err, ErrRunLeaseLost) {
				t.Fatalf("unstarted reservation finished an invocation: %v", err)
			}
			target, err := control.RequestStop(t.Context(), e)
			if err != nil || target != "holder" {
				t.Fatalf("reserved Stop target=%s err=%v", target, err)
			}
			if err := control.StartBatch(t.Context(), e, "holder"); !errors.Is(err, ErrPromptDiscarded) {
				t.Fatalf("reserved batch started after Stop: %v", err)
			}
			if err := control.ReleaseBatch(t.Context(), e, "holder"); err != nil {
				t.Fatal(err)
			}
			if err := control.Begin(t.Context(), events[1], "after-stop"); !errors.Is(err, ErrPromptDiscarded) {
				t.Fatalf("Stop failed to discard a reserved later member: %v", err)
			}
		})
	}
}
