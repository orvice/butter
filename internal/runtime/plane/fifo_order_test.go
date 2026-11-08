package plane

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func TestRedisFIFOChecksEveryUnfinishedPredecessor(t *testing.T) {
	for _, legacy := range []bool{false, true} {
		t.Run(map[bool]string{false: "new-orders", true: "legacy-completed-head"}[legacy], func(t *testing.T) {
			testRedisFIFOPredecessors(t, legacy)
		})
	}
}

func testRedisFIFOPredecessors(t *testing.T, legacy bool) {
	queue, base := planeQueueFixture(t)
	client := queue.client
	control := NewRedisRunControl(client)
	for _, id := range []string{"first", "second", "third"} {
		e := *base
		e.ActivityID = id
		if _, err := queue.Accept(t.Context(), &e); err != nil {
			t.Fatal(err)
		}
	}
	entries, err := queue.Read(t.Context(), "fifo-test", 3, time.Millisecond)
	if err != nil || len(entries) != 3 {
		t.Fatalf("read entries=%d err=%v", len(entries), err)
	}
	events := make([]*Event, 3)
	for i, entry := range entries {
		events[i], err = DecodeEvent(entry.Payload)
		if err != nil {
			t.Fatal(err)
		}
	}
	first, second, third := events[0], events[1], events[2]
	if err := control.Begin(t.Context(), first, "one"); err != nil {
		t.Fatal(err)
	}
	if _, err := control.Finish(t.Context(), first, "one", RunStateSucceeded); err != nil {
		t.Fatal(err)
	}
	if err := control.End(t.Context(), first, "one", RunStateSucceeded); err != nil {
		t.Fatal(err)
	}
	if size := client.ZCard(t.Context(), queueOrderKey(base)).Val(); size != 2 {
		t.Fatalf("completed order was not removed: pending=%d", size)
	}
	if legacy {
		// Existing #519 data retains completed orders alongside pending ones.
		if err := client.ZAdd(t.Context(), queueOrderKey(base), redis.Z{
			Score: float64(first.AcceptedOrder), Member: strconv.FormatInt(first.AcceptedOrder, 10),
		}).Err(); err != nil {
			t.Fatal(err)
		}
	}
	// The first order is complete. The second hasn't even reached a Worker.
	if err := control.Begin(t.Context(), third, "three"); !errors.Is(err, ErrRunBusy) {
		t.Fatalf("third overtook unread second: got %v, want ErrRunBusy", err)
	}
	if err := control.Begin(t.Context(), second, "two"); err != nil {
		t.Fatal(err)
	}
	if _, err := control.Finish(t.Context(), second, "two", RunStateSucceeded); err != nil {
		t.Fatal(err)
	}
	if err := control.End(t.Context(), second, "two", RunStateSucceeded); err != nil {
		t.Fatal(err)
	}
	if err := control.Begin(t.Context(), third, "three"); err != nil {
		t.Fatalf("third did not become eligible: %v", err)
	}
}

func TestDeniedFIFOAdmissionCannotChangeControl(t *testing.T) {
	for _, stop := range []bool{false, true} {
		t.Run(map[bool]string{false: "prompt-zadd", true: "stop-zremrangebyscore"}[stop], func(t *testing.T) {
			q, e := planeQueueFixture(t)
			c := NewRedisRunControl(q.client)
			if err := c.Begin(t.Context(), e, "holder"); err != nil {
				t.Fatal(err)
			}
			user, password := "plane-fifo-"+uuid.NewString(), uuid.NewString()
			command := "-zadd"
			if stop {
				command = "-zremrangebyscore"
			}
			if err := q.client.ACLSetUser(t.Context(), user, "on", ">"+password, "~*", "+@all", command).Err(); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = q.client.ACLDelUser(context.Background(), user).Err() })
			client := redis.NewClient(&redis.Options{Addr: q.client.Options().Addr, DB: 12, Username: user, Password: password})
			t.Cleanup(func() { _ = client.Close() })
			orderBefore := q.client.Get(t.Context(), orderKey).Val()
			e.Stop = stop
			if _, err := NewQueue(client).Accept(t.Context(), e); err == nil {
				t.Fatal("admitted event without required order-index permission")
			}
			if orderAfter := q.client.Get(t.Context(), orderKey).Val(); orderBefore != orderAfter {
				t.Fatalf("rejected admission advanced order: %q -> %q", orderBefore, orderAfter)
			}
			if stopped, err := c.Renew(t.Context(), e, "holder"); err != nil || stopped {
				t.Fatalf("rejected admission reached holder: stopped=%v err=%v", stopped, err)
			}
			if q.client.Exists(t.Context(), controlKey(e)+":cutoff").Val() != 0 {
				t.Fatal("rejected Stop left a cutoff")
			}
		})
	}
}

func TestFIFOEndFailureDoesNotPublishSettlement(t *testing.T) {
	for _, fault := range []string{"denied-zrem", "wrong-order-type", "wrong-done-type"} {
		t.Run(fault, func(t *testing.T) {
			q, e := planeQueueFixture(t)
			c := NewRedisRunControl(q.client)
			if err := c.Begin(t.Context(), e, "holder"); err != nil {
				t.Fatal(err)
			}
			if _, err := c.Finish(t.Context(), e, "holder", RunStateSucceeded); err != nil {
				t.Fatal(err)
			}
			writer := c
			switch fault {
			case "denied-zrem":
				user, password := "plane-end-"+uuid.NewString(), uuid.NewString()
				if err := q.client.ACLSetUser(t.Context(), user, "on", ">"+password, "~*", "+@all", "-zrem").Err(); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = q.client.ACLDelUser(context.Background(), user).Err() })
				client := redis.NewClient(&redis.Options{Addr: q.client.Options().Addr, DB: 12, Username: user, Password: password})
				t.Cleanup(func() { _ = client.Close() })
				writer = NewRedisRunControl(client)
			default:
				key := queueOrderKey(e)
				if fault == "wrong-done-type" {
					key = doneOrderKey(e)
				}
				if err := q.client.Set(t.Context(), key, "wrong-type", 0).Err(); err != nil {
					t.Fatal(err)
				}
			}
			if err := writer.End(t.Context(), e, "holder", RunStateSucceeded); err == nil {
				t.Fatal("End succeeded despite order-index failure")
			}
			if _, ended, err := c.Outcome(t.Context(), e, "holder"); err != nil || ended {
				t.Fatalf("failed End published settlement: ended=%v err=%v", ended, err)
			}
			if _, err := c.Renew(t.Context(), e, "holder"); err != nil {
				t.Fatalf("failed End released holder: %v", err)
			}
		})
	}
}
