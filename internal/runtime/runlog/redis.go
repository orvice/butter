package runlog

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"butterfly.orx.me/core/log"
	"github.com/redis/go-redis/v9"
)

// A log is a Redis Stream under the store's prefix plus its key. Each entry
// has two fields: its kind ('k') and its frame ('f'). Open leaves an empty
// stream, which exists, so a reader tells a log that has no entries yet from
// one that is gone. Appending checks, in one step, that the stream exists and
// holds exactly what its writer appended so far.

const (
	kindField  = "k"
	frameField = "f"
)

// openScript replaces whatever is at the key with an empty stream kept for
// ARGV[1] ms. Redis keeps a stream that was trimmed to nothing.
var openScript = redis.NewScript(`
redis.call('DEL', KEYS[1])
redis.call('XADD', KEYS[1], '*', 'k', 'open')
redis.call('XTRIM', KEYS[1], 'MAXLEN', 0)
redis.call('PEXPIRE', KEYS[1], ARGV[1])
return 1
`)

// appendScript appends entries (ARGV[3..], kind and frame pairs) to a stream
// that holds ARGV[2] entries, and keeps it for ARGV[1] ms. It answers -1 for
// a stream that is gone, -2 for one that holds something else, 0 when the
// entries are already there, and 1 when it appended them.
var appendScript = redis.NewScript(`
if redis.call('EXISTS', KEYS[1]) == 0 then
  return -1
end
local n = redis.call('XLEN', KEYS[1])
local at = tonumber(ARGV[2])
local count = (#ARGV - 2) / 2
if n == at + count then
  redis.call('PEXPIRE', KEYS[1], ARGV[1])
  return 0
end
if n ~= at then
  return -2
end
for i = 3, #ARGV, 2 do
  redis.call('XADD', KEYS[1], '*', 'k', ARGV[i], 'f', ARGV[i + 1])
end
redis.call('PEXPIRE', KEYS[1], ARGV[1])
return 1
`)

// Redis is the cross-Pod Store: one Stream per log. Its waits are
// multiplexed: one goroutine per store reads every log someone in this
// process waits on with a single blocking XREAD.
type Redis struct {
	rdb    *redis.Client
	prefix string
	tail   *tailer
}

// NewRedis builds the Redis store, its logs under keyPrefix. Returns nil when
// rdb is nil, so callers can fall back to the in-process store.
func NewRedis(rdb *redis.Client, keyPrefix string) *Redis {
	if rdb == nil {
		return nil
	}
	return &Redis{rdb: rdb, prefix: keyPrefix, tail: newTailer(rdb)}
}

var _ Store = (*Redis)(nil)

func (r *Redis) key(log string) string { return r.prefix + log }

// ttlMillis is ttl in whole milliseconds, at least one: a PEXPIRE of zero
// would delete the log.
func ttlMillis(ttl time.Duration) int64 { return max(ttl.Milliseconds(), 1) }

func (r *Redis) Open(ctx context.Context, key string, ttl time.Duration) error {
	if err := openScript.Run(ctx, r.rdb, []string{r.key(key)}, ttlMillis(ttl)).Err(); err != nil {
		return fmt.Errorf("open run log: %w", err)
	}
	return nil
}

func (r *Redis) Append(ctx context.Context, key string, at int, entries []Entry, ttl time.Duration) error {
	args := make([]any, 0, 2+2*len(entries))
	args = append(args, ttlMillis(ttl), at)
	for _, e := range entries {
		args = append(args, string(e.Kind), e.Frame)
	}
	result, err := appendScript.Run(ctx, r.rdb, []string{r.key(key)}, args...).Int64()
	if err != nil {
		return fmt.Errorf("append to run log: %w", err)
	}
	switch result {
	case -1:
		return ErrGone
	case -2:
		return ErrConflict
	default:
		return nil
	}
}

func (r *Redis) Expire(ctx context.Context, key string, ttl time.Duration) error {
	kept, err := r.rdb.PExpire(ctx, r.key(key), time.Duration(ttlMillis(ttl))*time.Millisecond).Result()
	if err != nil {
		return fmt.Errorf("keep run log: %w", err)
	}
	if !kept {
		return ErrGone
	}
	return nil
}

func (r *Redis) Read(ctx context.Context, key, after string, count int) ([]Record, error) {
	start := "-"
	if after != "" {
		id, ok := parseStreamID(after)
		if !ok {
			return nil, fmt.Errorf("read run log: malformed entry ID %q", after)
		}
		start = id.next().String()
	}
	pipe := r.rdb.TxPipeline()
	exists := pipe.Exists(ctx, r.key(key))
	entries := pipe.XRangeN(ctx, r.key(key), start, "+", int64(max(count, 1)))
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, fmt.Errorf("read run log: %w", err)
	}
	if exists.Val() == 0 {
		return nil, ErrGone
	}
	out := make([]Record, 0, len(entries.Val()))
	for _, msg := range entries.Val() {
		kind, _ := msg.Values[kindField].(string)
		frame, _ := msg.Values[frameField].(string)
		out = append(out, Record{ID: msg.ID, Entry: Entry{Kind: Kind(kind), Frame: []byte(frame)}})
	}
	return out, nil
}

func (r *Redis) Wait(ctx context.Context, key, after string, timeout time.Duration) {
	from := streamID{}
	if after != "" {
		id, ok := parseStreamID(after)
		if !ok {
			return
		}
		from = id
	}
	r.tail.wait(ctx, r.key(key), from, timeout)
}

func (r *Redis) Drop(ctx context.Context, key string) error {
	if err := r.rdb.Del(ctx, r.key(key)).Err(); err != nil {
		return fmt.Errorf("drop run log: %w", err)
	}
	return nil
}

// Close stops the store's reads. A process keeps its store for its lifetime;
// tests close it.
func (r *Redis) Close() { r.tail.close() }

// --- Stream IDs ------------------------------------------------------------------

// streamID is a Redis Stream entry ID, milliseconds and sequence.
type streamID struct{ ms, seq uint64 }

func parseStreamID(s string) (streamID, bool) {
	msPart, seqPart, ok := strings.Cut(s, "-")
	if !ok {
		return streamID{}, false
	}
	ms, err := strconv.ParseUint(msPart, 10, 64)
	if err != nil {
		return streamID{}, false
	}
	seq, err := strconv.ParseUint(seqPart, 10, 64)
	if err != nil {
		return streamID{}, false
	}
	return streamID{ms: ms, seq: seq}, true
}

func (id streamID) String() string {
	return strconv.FormatUint(id.ms, 10) + "-" + strconv.FormatUint(id.seq, 10)
}

func (id streamID) less(other streamID) bool {
	return id.ms < other.ms || (id.ms == other.ms && id.seq < other.seq)
}

// next is the smallest ID after id.
func (id streamID) next() streamID {
	if id.seq == ^uint64(0) {
		return streamID{ms: id.ms + 1}
	}
	return streamID{ms: id.ms, seq: id.seq + 1}
}

// --- Multiplexed waits -------------------------------------------------------------

const (
	// tailBlock bounds one XREAD. A log someone starts waiting on joins the
	// next one, so this is also how late its first wake-up can be.
	tailBlock = 100 * time.Millisecond
	// tailLinger keeps a log in the XREAD for a while after its last waiter
	// woke, for the waiter to come back with a later entry: a log that is
	// followed stays read without a gap.
	tailLinger = time.Second
	// tailMaxBackoff caps the pause after an XREAD that failed.
	tailMaxBackoff = time.Second
)

// tailer multiplexes the waits of a process's observers: one goroutine reads
// every log someone waits on with one blocking XREAD, on one connection, and
// wakes the waiters whose log grew. It runs while anyone waits.
type tailer struct {
	rdb *redis.Client

	mu      sync.Mutex
	logs    map[string]*tailedLog // by Redis key
	running bool
	closed  bool
	// stop ends the reads once the tailer is closed.
	stop     context.Context
	stopFunc context.CancelFunc
}

// tailedLog is one log the tailer reads.
type tailedLog struct {
	// last is the newest entry known to be in the log: the newest one a
	// waiter read, or the XREAD found. The XREAD reads from it.
	last    streamID
	waiters map[*tailWaiter]struct{}
	// idle is when the log's last waiter left; zero while someone waits.
	idle time.Time
}

// tailWaiter is one observer waiting for entries after its own.
type tailWaiter struct {
	after streamID
	woke  chan struct{}
}

func newTailer(rdb *redis.Client) *tailer {
	stop, stopFunc := context.WithCancel(context.Background())
	return &tailer{rdb: rdb, logs: map[string]*tailedLog{}, stop: stop, stopFunc: stopFunc}
}

// wait returns once the log at key may hold an entry after after, after
// timeout, or when ctx ends.
func (t *tailer) wait(ctx context.Context, key string, after streamID, timeout time.Duration) {
	w := &tailWaiter{after: after, woke: make(chan struct{})}
	t.mu.Lock()
	l := t.logs[key]
	if l == nil {
		l = &tailedLog{last: after, waiters: map[*tailWaiter]struct{}{}}
		t.logs[key] = l
	}
	if l.last.less(after) {
		l.last = after
	}
	if after.less(l.last) {
		// An entry after the waiter's is already known.
		t.mu.Unlock()
		return
	}
	l.waiters[w] = struct{}{}
	l.idle = time.Time{}
	if !t.running && !t.closed {
		t.running = true
		go t.loop()
	}
	t.mu.Unlock()

	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-w.woke:
	case <-timer.C:
	case <-ctx.Done():
	}
	t.mu.Lock()
	delete(l.waiters, w)
	if len(l.waiters) == 0 && l.idle.IsZero() {
		l.idle = time.Now()
	}
	t.mu.Unlock()
}

// watched lists the logs to read and where to read each from, forgetting
// those nobody waited on for a while. ok is false when there is none left,
// and the loop ends. The caller holds t.mu.
func (t *tailer) watched(now time.Time) (streams []string, ok bool) {
	keys := make([]string, 0, len(t.logs))
	ids := make([]string, 0, len(t.logs))
	for key, l := range t.logs {
		if len(l.waiters) == 0 && !l.idle.IsZero() && now.Sub(l.idle) > tailLinger {
			delete(t.logs, key)
			continue
		}
		keys = append(keys, key)
		ids = append(ids, l.last.String())
	}
	return append(keys, ids...), len(keys) > 0
}

func (t *tailer) loop() {
	backoff := time.Duration(0)
	for {
		t.mu.Lock()
		streams, ok := t.watched(time.Now())
		if !ok || t.closed {
			t.running = false
			t.mu.Unlock()
			return
		}
		t.mu.Unlock()

		res, err := t.rdb.XRead(t.stop, &redis.XReadArgs{Streams: streams, Count: 1, Block: tailBlock}).Result()
		switch {
		case errors.Is(err, redis.Nil):
			// Nothing new within the block.
			backoff = 0
			continue
		case err != nil:
			if t.stop.Err() != nil {
				continue
			}
			// The waiters time out and read for themselves meanwhile.
			backoff = min(max(2*backoff, 50*time.Millisecond), tailMaxBackoff)
			log.FromContext(t.stop).Warn("run log reads failed; retrying", "err", err, "retry_in", backoff.String())
			select {
			case <-t.stop.Done():
			case <-time.After(backoff):
			}
			continue
		}
		backoff = 0
		t.wake(res)
	}
}

// wake records what the XREAD found and wakes the waiters it concerns.
func (t *tailer) wake(streams []redis.XStream) {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now()
	for _, s := range streams {
		l := t.logs[s.Stream]
		if l == nil || len(s.Messages) == 0 {
			continue
		}
		id, ok := parseStreamID(s.Messages[len(s.Messages)-1].ID)
		if !ok {
			continue
		}
		if l.last.less(id) {
			l.last = id
		}
		for w := range l.waiters {
			if w.after.less(l.last) {
				close(w.woke)
				delete(l.waiters, w)
			}
		}
		if len(l.waiters) == 0 && l.idle.IsZero() {
			l.idle = now
		}
	}
}

func (t *tailer) close() {
	t.mu.Lock()
	t.closed = true
	t.mu.Unlock()
	t.stopFunc()
}
