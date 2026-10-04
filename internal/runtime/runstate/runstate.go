// Package runstate records which run holds an AG-UI thread (ADR-0016
// decision 6).
//
// Every run, detached or not, writes its run state when it starts, next to
// its thread lease: its runId, its Invocation ID, and how many events the
// thread's session held before it. The run renews the state with the lease,
// at the same pace and for exactly as long as it holds the lease (Keep), so
// the state lapses with the lease when the process running the run dies or
// the run loses its thread. A thread read then never takes a dead run for a
// running one for longer than one TTL. Thread reads use it to answer
// "running" and to cut the session where the run started (#403).
package runstate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"butterfly.orx.me/core/log"
	"github.com/redis/go-redis/v9"
)

// State is what one run records about itself while it holds its thread.
type State struct {
	// RunID is the client's AG-UI runId.
	RunID string `json:"run_id"`
	// InvocationID is the ID of the run's Invocation record. It also fences
	// the run's later writes: a run renews and ends only its own state.
	InvocationID string `json:"invocation_id"`
	// EventCount is how many events the thread's session held before the
	// run: the run's own events start at this index.
	EventCount int `json:"event_count"`
	// Detached reports whether the client asked for a Detached Run, the
	// only kind a Stop or an attach reaches.
	Detached bool `json:"detached,omitempty"`
}

// Store keeps one run state per thread. A write lasts one TTL.
//
// thread is the thread's key, the same key the thread's lease is taken under.
type Store interface {
	// Begin records st as the thread's run state, replacing whatever the
	// thread held: only the lease holder begins a run, so a state already
	// there was left by a run that no longer holds the thread.
	Begin(ctx context.Context, thread string, st State) error
	// Renew keeps the thread's run state for one more TTL while it is still
	// the state of the run with this invocation ID. It never brings back a
	// state that lapsed or was ended, and never touches another run's:
	// renewed is false then, and nothing changes.
	Renew(ctx context.Context, thread, invocationID string) (renewed bool, err error)
	// End removes the thread's run state if it is still the one of the run
	// with this invocation ID. Another run's state is left alone.
	End(ctx context.Context, thread, invocationID string) error
	// Get returns the thread's current run state; ok is false when no run
	// holds the thread.
	Get(ctx context.Context, thread string) (st State, ok bool, err error)
	// TTL is how long one write lasts.
	TTL() time.Duration
}

var errNoInvocation = errors.New("run state needs an invocation ID")

// --- Keeping a run's state -----------------------------------------------------

// Kept is a run state its run keeps renewed. End it when the run ends.
type Kept struct {
	store  Store
	thread string
	st     State
	stop   context.CancelFunc
	done   chan struct{}
	once   sync.Once
}

// Keep records st as the thread's run state and renews it every TTL/3 until
// ctx ends or the state is ended. ctx is the run's lease context: the state
// is renewed only while the run holds its thread's lease, so it lapses with
// the lease when the run loses the thread or its process dies.
//
// A renewal that errors is retried sooner, as a lease renewal is. A renewal
// the store refuses — the state lapsed, was ended, or belongs to another run
// now — ends the renewals: a run never brings its state back.
func Keep(ctx context.Context, s Store, thread string, st State) (*Kept, error) {
	attempt, cancel := context.WithTimeout(ctx, s.TTL()/3)
	err := s.Begin(attempt, thread, st)
	cancel()
	if err != nil {
		return nil, err
	}
	renewCtx, stop := context.WithCancel(ctx)
	k := &Kept{store: s, thread: thread, st: st, stop: stop, done: make(chan struct{})}
	go k.renew(renewCtx)
	return k, nil
}

// State is the run state being kept.
func (k *Kept) State() State { return k.st }

// End stops the renewals and removes the run state, unless another run's
// state replaced it. Ending again does nothing.
func (k *Kept) End(ctx context.Context) error {
	var err error
	k.once.Do(func() {
		k.stop()
		<-k.done
		err = k.store.End(ctx, k.thread, k.st.InvocationID)
	})
	return err
}

func (k *Kept) renew(ctx context.Context) {
	defer close(k.done)
	ttl := k.store.TTL()
	interval := ttl / 3
	retry := max(ttl/10, time.Millisecond)
	timer := time.NewTimer(interval)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}
		attempt, cancel := context.WithTimeout(ctx, interval)
		renewed, err := k.store.Renew(attempt, k.thread, k.st.InvocationID)
		cancel()
		switch {
		case ctx.Err() != nil:
			return
		case err != nil:
			log.FromContext(ctx).Warn("run state renewal failed; retrying",
				"thread", k.thread, "invocation_id", k.st.InvocationID, "err", err)
			timer.Reset(retry)
		case !renewed:
			log.FromContext(ctx).Warn("run state lapsed or was replaced; no longer renewing it",
				"thread", k.thread, "invocation_id", k.st.InvocationID)
			return
		default:
			timer.Reset(interval)
		}
	}
}

// --- In-process store ----------------------------------------------------------

// Memory is the in-process Store. It serves a single process and the tests,
// and honors the same TTL as Redis.
type Memory struct {
	ttl    time.Duration
	mu     sync.Mutex
	states map[string]memoryState
}

type memoryState struct {
	st      State
	expires time.Time
}

// NewMemory builds an in-process store whose writes last ttl.
func NewMemory(ttl time.Duration) *Memory {
	return &Memory{ttl: ttl, states: make(map[string]memoryState)}
}

var _ Store = (*Memory)(nil)

// current returns the thread's unexpired state. The caller holds m.mu.
func (m *Memory) current(thread string, now time.Time) (State, bool) {
	held, ok := m.states[thread]
	if !ok {
		return State{}, false
	}
	if !now.Before(held.expires) {
		delete(m.states, thread)
		return State{}, false
	}
	return held.st, true
}

func (m *Memory) Begin(_ context.Context, thread string, st State) error {
	if st.InvocationID == "" {
		return errNoInvocation
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.states[thread] = memoryState{st: st, expires: time.Now().Add(m.ttl)}
	return nil
}

func (m *Memory) Renew(_ context.Context, thread, invocationID string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	held, ok := m.current(thread, now)
	if !ok || held.InvocationID != invocationID {
		return false, nil
	}
	m.states[thread] = memoryState{st: held, expires: now.Add(m.ttl)}
	return true, nil
}

func (m *Memory) End(_ context.Context, thread, invocationID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if held, ok := m.current(thread, time.Now()); ok && held.InvocationID == invocationID {
		delete(m.states, thread)
	}
	return nil
}

func (m *Memory) Get(_ context.Context, thread string) (State, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	st, ok := m.current(thread, time.Now())
	return st, ok, nil
}

func (m *Memory) TTL() time.Duration { return m.ttl }

// --- Redis store -----------------------------------------------------------------

// The run state is a hash: the invocation ID that fences writes, and the
// state itself as JSON. Renewing and ending are each one atomic step that
// checks the invocation ID, so a run that lost its thread can neither
// renew nor end the state of the run that took it.

var beginScript = redis.NewScript(`
redis.call('DEL', KEYS[1])
redis.call('HSET', KEYS[1], 'invocation_id', ARGV[1], 'state', ARGV[2])
redis.call('PEXPIRE', KEYS[1], ARGV[3])
return 1
`)

var renewScript = redis.NewScript(`
if redis.call('HGET', KEYS[1], 'invocation_id') == ARGV[1] then
  return redis.call('PEXPIRE', KEYS[1], ARGV[2])
end
return 0
`)

var endScript = redis.NewScript(`
if redis.call('HGET', KEYS[1], 'invocation_id') == ARGV[1] then
  return redis.call('DEL', KEYS[1])
end
return 0
`)

// Redis is the cross-Pod Store: one hash per thread, under keyPrefix plus
// the thread's key, written with the TTL.
type Redis struct {
	rdb    *redis.Client
	prefix string
	ttl    time.Duration
}

// NewRedis builds the Redis store. Returns nil when rdb is nil, so callers
// can fall back to the in-process store.
func NewRedis(rdb *redis.Client, keyPrefix string, ttl time.Duration) *Redis {
	if rdb == nil {
		return nil
	}
	return &Redis{rdb: rdb, prefix: keyPrefix, ttl: ttl}
}

var _ Store = (*Redis)(nil)

func (r *Redis) key(thread string) string { return r.prefix + thread }

func (r *Redis) Begin(ctx context.Context, thread string, st State) error {
	if st.InvocationID == "" {
		return errNoInvocation
	}
	raw, err := json.Marshal(st)
	if err != nil {
		return fmt.Errorf("encode run state: %w", err)
	}
	if err := beginScript.Run(ctx, r.rdb, []string{r.key(thread)},
		st.InvocationID, string(raw), r.ttl.Milliseconds()).Err(); err != nil {
		return fmt.Errorf("record run state: %w", err)
	}
	return nil
}

func (r *Redis) Renew(ctx context.Context, thread, invocationID string) (bool, error) {
	renewed, err := renewScript.Run(ctx, r.rdb, []string{r.key(thread)},
		invocationID, r.ttl.Milliseconds()).Int64()
	if err != nil {
		return false, fmt.Errorf("renew run state: %w", err)
	}
	return renewed == 1, nil
}

func (r *Redis) End(ctx context.Context, thread, invocationID string) error {
	if err := endScript.Run(ctx, r.rdb, []string{r.key(thread)}, invocationID).Err(); err != nil {
		return fmt.Errorf("end run state: %w", err)
	}
	return nil
}

func (r *Redis) Get(ctx context.Context, thread string) (State, bool, error) {
	raw, err := r.rdb.HGet(ctx, r.key(thread), "state").Result()
	if errors.Is(err, redis.Nil) {
		return State{}, false, nil
	}
	if err != nil {
		return State{}, false, fmt.Errorf("read run state: %w", err)
	}
	var st State
	if err := json.Unmarshal([]byte(raw), &st); err != nil {
		return State{}, false, fmt.Errorf("decode run state: %w", err)
	}
	return st, true, nil
}

func (r *Redis) TTL() time.Duration { return r.ttl }
