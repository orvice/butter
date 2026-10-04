package runlog

// The Run Log contract (#404, ADR-0016 decision 5): one suite, run against
// the in-process store and — when REDIS_ADDR is set — against Redis, so both
// are held to the same guarantees.

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type storeFactory func(t *testing.T) Store

func memoryStore(*testing.T) Store { return NewMemory() }

// redisClient connects to the Redis that REDIS_ADDR names, skipping the test
// without one, and removes every key under prefix when the test ends.
func redisClient(t *testing.T, prefix string, opts func(*redis.Options)) *redis.Client {
	t.Helper()
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		t.Skip("REDIS_ADDR is required for the Redis run log contract")
	}
	options := &redis.Options{Addr: addr}
	if opts != nil {
		opts(options)
	}
	rdb := redis.NewClient(options)
	if err := rdb.Ping(t.Context()).Err(); err != nil {
		t.Fatalf("ping redis: %v", err)
	}
	t.Cleanup(func() {
		keys, _ := rdb.Keys(context.Background(), prefix+"*").Result()
		if len(keys) > 0 {
			_ = rdb.Del(context.Background(), keys...).Err()
		}
		_ = rdb.Close()
	})
	return rdb
}

func redisStore(t *testing.T) Store {
	t.Helper()
	prefix := "butter:test:" + uuid.NewString() + ":log:"
	store := NewRedis(redisClient(t, prefix, nil), prefix)
	t.Cleanup(store.Close)
	return store
}

func TestStoreContract(t *testing.T) {
	for name, factory := range map[string]storeFactory{"memory": memoryStore, "redis": redisStore} {
		t.Run(name, func(t *testing.T) { runStoreContract(t, factory) })
	}
}

func event(s string) Entry { return Entry{Kind: KindEvent, Frame: []byte("data: " + s + "\n\n")} }

func end(s string) Entry { return Entry{Kind: KindEnd, Frame: []byte("data: " + s + "\n\n")} }

func truncated() Entry { return Entry{Kind: KindTruncated} }

// entriesOf drops the IDs of what a read returned.
func entriesOf(recs []Record) []Entry {
	out := make([]Entry, 0, len(recs))
	for _, r := range recs {
		out = append(out, r.Entry)
	}
	return out
}

func requireEntries(t *testing.T, label string, got []Record, want ...Entry) {
	t.Helper()
	entries := entriesOf(got)
	if len(entries) != len(want) {
		t.Fatalf("%s: read %d entries %q, want %d %q", label, len(entries), entries, len(want), want)
	}
	for i := range want {
		if entries[i].Kind != want[i].Kind || string(entries[i].Frame) != string(want[i].Frame) {
			t.Fatalf("%s: entry %d = %q, want %q", label, i, entries[i], want[i])
		}
	}
}

func open(t *testing.T, s Store, key string, ttl time.Duration) {
	t.Helper()
	if err := s.Open(t.Context(), key, ttl); err != nil {
		t.Fatalf("Open(%s): %v", key, err)
	}
}

func appendAt(t *testing.T, s Store, key string, at int, entries ...Entry) {
	t.Helper()
	if err := s.Append(t.Context(), key, at, entries, time.Minute); err != nil {
		t.Fatalf("Append(%s, %d): %v", key, at, err)
	}
}

func read(t *testing.T, s Store, key, after string, count int) []Record {
	t.Helper()
	recs, err := s.Read(t.Context(), key, after, count)
	if err != nil {
		t.Fatalf("Read(%s, %q): %v", key, after, err)
	}
	return recs
}

// gone reports whether the log at key reads as gone.
func gone(t *testing.T, s Store, key string) bool {
	t.Helper()
	_, err := s.Read(t.Context(), key, "", 1)
	if err != nil && !errors.Is(err, ErrGone) {
		t.Fatalf("Read(%s): %v", key, err)
	}
	return errors.Is(err, ErrGone)
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

// waitReturns runs Wait and reports how long it took.
func waitReturns(s Store, key, after string, timeout time.Duration) time.Duration {
	start := time.Now()
	s.Wait(context.Background(), key, after, timeout)
	return time.Since(start)
}

func runStoreContract(t *testing.T, factory storeFactory) {
	t.Run("AnOpenedLogIsEmptyAndAnUnopenedOneIsGone", func(t *testing.T) {
		s := factory(t)
		open(t, s, "u1:agui-t-1:inv-a", time.Minute)
		if recs := read(t, s, "u1:agui-t-1:inv-a", "", 10); len(recs) != 0 {
			t.Fatalf("an opened log holds %q", entriesOf(recs))
		}
		if !gone(t, s, "u1:agui-t-1:inv-b") {
			t.Fatal("a log nobody opened reads as there")
		}
	})

	t.Run("EntriesAreReadBackInOrderFromAnyPoint", func(t *testing.T) {
		s := factory(t)
		open(t, s, "l", time.Minute)
		appendAt(t, s, "l", 0, event("a"), event("b"))
		appendAt(t, s, "l", 2, event("c"), end("d"))
		all := read(t, s, "l", "", 100)
		requireEntries(t, "from the start", all, event("a"), event("b"), event("c"), end("d"))
		requireEntries(t, "a page", read(t, s, "l", "", 2), event("a"), event("b"))
		requireEntries(t, "after the second", read(t, s, "l", all[1].ID, 100), event("c"), end("d"))
		requireEntries(t, "after the last", read(t, s, "l", all[3].ID, 100))
		if !all[3].Kind.Final() || all[2].Kind.Final() || !KindTruncated.Final() {
			t.Fatal("Final misreports a kind")
		}
	})

	t.Run("ARetriedAppendAddsNothingTwice", func(t *testing.T) {
		s := factory(t)
		open(t, s, "l", time.Minute)
		appendAt(t, s, "l", 0, event("a"), event("b"))
		// The reply was lost; the writer sends the same batch again.
		appendAt(t, s, "l", 0, event("a"), event("b"))
		requireEntries(t, "after the retry", read(t, s, "l", "", 100), event("a"), event("b"))
		// An append at a length the log does not have is refused, and
		// changes nothing.
		if err := s.Append(t.Context(), "l", 5, []Entry{event("x")}, time.Minute); !errors.Is(err, ErrConflict) {
			t.Fatalf("Append at the wrong length = %v, want ErrConflict", err)
		}
		requireEntries(t, "after the conflict", read(t, s, "l", "", 100), event("a"), event("b"))
	})

	t.Run("ATruncationMarkerKeepsItsKind", func(t *testing.T) {
		s := factory(t)
		open(t, s, "l", time.Minute)
		appendAt(t, s, "l", 0, event("a"), truncated())
		requireEntries(t, "read", read(t, s, "l", "", 100), event("a"), Entry{Kind: KindTruncated, Frame: []byte{}})
	})

	t.Run("OpeningReplacesAnEarlierLog", func(t *testing.T) {
		s := factory(t)
		open(t, s, "l", time.Minute)
		appendAt(t, s, "l", 0, event("old"))
		open(t, s, "l", time.Minute)
		if recs := read(t, s, "l", "", 100); len(recs) != 0 {
			t.Fatalf("a reopened log holds %q", entriesOf(recs))
		}
		appendAt(t, s, "l", 0, event("new"))
		requireEntries(t, "after reopening", read(t, s, "l", "", 100), event("new"))
	})

	t.Run("ALogLapsesOneTTLAfterItsLastWrite", func(t *testing.T) {
		s := factory(t)
		open(t, s, "l", 300*time.Millisecond)
		if !waitUntil(3*time.Second, func() bool { return gone(t, s, "l") }) {
			t.Fatal("a log outlived its TTL")
		}
	})

	t.Run("AppendingAndExpiringKeepTheLog", func(t *testing.T) {
		const ttl = 400 * time.Millisecond
		s := factory(t)
		open(t, s, "l", ttl)
		time.Sleep(ttl * 3 / 4)
		if err := s.Append(t.Context(), "l", 0, []Entry{event("a")}, ttl); err != nil {
			t.Fatalf("Append: %v", err)
		}
		time.Sleep(ttl * 3 / 4)
		if err := s.Expire(t.Context(), "l", ttl); err != nil {
			t.Fatalf("Expire: %v", err)
		}
		// Past the first two writes' TTLs, inside the last one's.
		time.Sleep(ttl / 2)
		requireEntries(t, "kept", read(t, s, "l", "", 10), event("a"))
	})

	t.Run("AGoneLogNeverComesBack", func(t *testing.T) {
		const ttl = 200 * time.Millisecond
		s := factory(t)
		open(t, s, "l", ttl)
		if err := s.Append(t.Context(), "l", 0, []Entry{event("a")}, ttl); err != nil {
			t.Fatalf("Append: %v", err)
		}
		if !waitUntil(3*time.Second, func() bool { return gone(t, s, "l") }) {
			t.Fatal("the log did not lapse")
		}
		if err := s.Append(t.Context(), "l", 1, []Entry{event("b")}, time.Minute); !errors.Is(err, ErrGone) {
			t.Fatalf("Append to a lapsed log = %v, want ErrGone", err)
		}
		if err := s.Expire(t.Context(), "l", time.Minute); !errors.Is(err, ErrGone) {
			t.Fatalf("Expire of a lapsed log = %v, want ErrGone", err)
		}
		if !gone(t, s, "l") {
			t.Fatal("a lapsed log came back")
		}
	})

	t.Run("DropRemovesTheLog", func(t *testing.T) {
		s := factory(t)
		open(t, s, "l", time.Minute)
		appendAt(t, s, "l", 0, event("a"))
		if err := s.Drop(t.Context(), "l"); err != nil {
			t.Fatalf("Drop: %v", err)
		}
		if !gone(t, s, "l") {
			t.Fatal("a dropped log reads as there")
		}
		if err := s.Append(t.Context(), "l", 1, []Entry{event("b")}, time.Minute); !errors.Is(err, ErrGone) {
			t.Fatalf("Append to a dropped log = %v, want ErrGone", err)
		}
		if err := s.Drop(t.Context(), "l"); err != nil {
			t.Fatalf("Drop of a dropped log: %v", err)
		}
	})

	t.Run("LogsAreIndependent", func(t *testing.T) {
		s := factory(t)
		open(t, s, "u1:agui-t-1:inv-a", time.Minute)
		open(t, s, "u2:agui-t-1:inv-b", time.Minute)
		appendAt(t, s, "u1:agui-t-1:inv-a", 0, event("a"))
		appendAt(t, s, "u2:agui-t-1:inv-b", 0, event("b"))
		if err := s.Drop(t.Context(), "u2:agui-t-1:inv-b"); err != nil {
			t.Fatalf("Drop: %v", err)
		}
		requireEntries(t, "the other log", read(t, s, "u1:agui-t-1:inv-a", "", 10), event("a"))
	})

	// Wait is how observers follow a log without polling it.

	t.Run("AWaitReturnsOnceTheLogGrows", func(t *testing.T) {
		s := factory(t)
		open(t, s, "l", time.Minute)
		appendAt(t, s, "l", 0, event("a"))
		first := read(t, s, "l", "", 10)[0].ID
		woke := make(chan time.Duration, 1)
		go func() { woke <- waitReturns(s, "l", first, 10*time.Second) }()
		time.Sleep(150 * time.Millisecond)
		appendAt(t, s, "l", 1, event("b"))
		select {
		case <-woke:
		case <-time.After(5 * time.Second):
			t.Fatal("the wait did not return once the log grew")
		}
		requireEntries(t, "after the wait", read(t, s, "l", first, 10), event("b"))
	})

	t.Run("AWaitReturnsAtOnceForEntriesAlreadyThere", func(t *testing.T) {
		s := factory(t)
		open(t, s, "l", time.Minute)
		appendAt(t, s, "l", 0, event("a"), event("b"))
		first := read(t, s, "l", "", 10)[0].ID
		if took := waitReturns(s, "l", first, 10*time.Second); took > 2*time.Second {
			t.Fatalf("waiting with an entry already after the reader's took %v", took)
		}
		if took := waitReturns(s, "l", "", 10*time.Second); took > 2*time.Second {
			t.Fatalf("waiting from the start of a log with entries took %v", took)
		}
	})

	t.Run("AWaitOnAQuietLogTimesOut", func(t *testing.T) {
		s := factory(t)
		open(t, s, "l", time.Minute)
		appendAt(t, s, "l", 0, event("a"))
		last := read(t, s, "l", "", 10)[0].ID
		if took := waitReturns(s, "l", last, 300*time.Millisecond); took < 250*time.Millisecond || took > 3*time.Second {
			t.Fatalf("a wait on a quiet log returned after %v, want about its timeout", took)
		}
		// A wait whose context ends returns then.
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		start := time.Now()
		s.Wait(ctx, "l", last, 10*time.Second)
		if took := time.Since(start); took > 3*time.Second {
			t.Fatalf("a cancelled wait returned after %v", took)
		}
	})

	t.Run("ManyWaitersOnManyLogsAllWake", func(t *testing.T) {
		s := factory(t)
		const logs = 12
		var wg sync.WaitGroup
		woke := make(chan string, logs*2)
		for i := range logs {
			key := fmt.Sprintf("l-%d", i)
			open(t, s, key, time.Minute)
			appendAt(t, s, key, 0, event("start"))
			first := read(t, s, key, "", 1)[0].ID
			for range 2 {
				wg.Go(func() {
					if waitReturns(s, key, first, 10*time.Second) < 9*time.Second {
						woke <- key
					}
				})
			}
		}
		time.Sleep(200 * time.Millisecond)
		for i := range logs {
			appendAt(t, s, fmt.Sprintf("l-%d", i), 1, event("next"))
		}
		wg.Wait()
		if len(woke) != logs*2 {
			t.Fatalf("%d of %d waiters woke when their log grew", len(woke), logs*2)
		}
	})
}

// The Redis store's waits share one blocking read: observers waiting on many
// logs at once hold no connection each, so a client with a two-connection
// pool still serves them all, and the writers beside them.
func TestRedisWaitsShareOneConnection(t *testing.T) {
	prefix := "butter:test:" + uuid.NewString() + ":log:"
	rdb := redisClient(t, prefix, func(o *redis.Options) {
		o.PoolSize = 2
		o.PoolTimeout = time.Second
	})
	s := NewRedis(rdb, prefix)
	t.Cleanup(s.Close)
	const logs = 20
	var wg sync.WaitGroup
	woke := make(chan struct{}, logs)
	for i := range logs {
		key := fmt.Sprintf("l-%d", i)
		open(t, s, key, time.Minute)
		wg.Go(func() {
			if waitReturns(s, key, "", 20*time.Second) < 15*time.Second {
				woke <- struct{}{}
			}
		})
	}
	time.Sleep(300 * time.Millisecond)
	for i := range logs {
		appendAt(t, s, fmt.Sprintf("l-%d", i), 0, event("first"))
	}
	wg.Wait()
	if len(woke) != logs {
		t.Fatalf("%d of %d waiters woke", len(woke), logs)
	}
	if stats := rdb.PoolStats(); stats.Timeouts > 0 {
		t.Fatalf("pool timeouts = %d: an observer held a connection of its own", stats.Timeouts)
	}
}

// The in-process store lets go of the logs that lapsed whenever a log opens,
// so the logs of ended runs do not pile up in a process nobody reads them in.
func TestMemoryLetsGoOfLapsedLogsAsLogsOpen(t *testing.T) {
	m := NewMemory()
	open(t, m, "old", 50*time.Millisecond)
	if err := m.Append(t.Context(), "old", 0, []Entry{event("a")}, 50*time.Millisecond); err != nil {
		t.Fatalf("Append: %v", err)
	}
	time.Sleep(100 * time.Millisecond)
	open(t, m, "new", time.Minute)
	m.mu.Lock()
	_, kept := m.logs["old"]
	m.mu.Unlock()
	if kept {
		t.Fatal("a lapsed log is still held after another log opened")
	}
}

func TestNewRedisWithoutAClientIsNil(t *testing.T) {
	if NewRedis(nil, "p:") != nil {
		t.Fatal("NewRedis(nil) built a store")
	}
}

func TestRedisKeysSitUnderThePrefix(t *testing.T) {
	r := NewRedis(redis.NewClient(&redis.Options{Addr: "127.0.0.1:0"}), "butter:agui:log:session:")
	t.Cleanup(func() {
		r.Close()
		_ = r.rdb.Close()
	})
	if got := r.key("u1:agui-t-1:inv-1"); got != "butter:agui:log:session:u1:agui-t-1:inv-1" {
		t.Fatalf("key = %q", got)
	}
}

func TestStreamIDs(t *testing.T) {
	id, ok := parseStreamID("1700000000000-7")
	if !ok || id.String() != "1700000000000-7" || id.next().String() != "1700000000000-8" {
		t.Fatalf("parsed %+v, %v", id, ok)
	}
	if !id.less(id.next()) || id.next().less(id) || id.less(id) {
		t.Fatal("less misorders IDs")
	}
	if wrap := (streamID{ms: 5, seq: ^uint64(0)}).next(); wrap != (streamID{ms: 6}) {
		t.Fatalf("next of the last sequence = %+v", wrap)
	}
	for _, bad := range []string{"", "12", "a-1", "1-b", "-"} {
		if _, ok := parseStreamID(bad); ok {
			t.Fatalf("parsed %q", bad)
		}
	}
}
