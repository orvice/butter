// Package runstate records which run holds an AG-UI thread (ADR-0016
// decision 6), and carries a Stop to it (decision 4).
//
// Every run, detached or not, writes its run state when it starts, next to
// its thread lease: its runId, its Invocation ID, how many events the
// thread's session held before it, and the token of its lease acquisition.
// The run renews the state with the lease, at the same pace and for exactly
// as long as it holds the lease (Keep), so the state lapses with the lease
// when the process running the run dies or the run loses its thread. A thread
// read then never takes a dead run for a running one for longer than one TTL.
// Thread reads use it to answer "running" and to cut the session where the
// run started (#403).
//
// When a run ends, its state ends with it. A run that lived in its request
// removes it. A Detached Run's is kept, marked ended, as long as the run's Run
// Log (#404), so a late attach still finds the run and replays it; reads take
// an ended run for one that is not running, nothing renews it, and the next
// run on the thread replaces it.
//
// A Stop is accepted on that same state, never on the lease, so it never
// waits behind the run it stops. In one step it marks the state of a Detached
// Run that has not claimed its end with the run's lease token, and publishes
// a nudge to the run. The run claims its terminal state on the same state, in
// one step too (decision 3): a Stop accepted before the claim always ends the
// run CANCELLED, and one after it finds nothing running. A run learns of its
// Stop from the nudge, which it subscribes to before its state exists, or from
// the marker at its next renewal when the nudge never arrived (Keep). The
// marker lives in the run's own state, which the next run's replaces, so a
// Stop never reaches a later run on the thread.
package runstate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
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
	// LeaseToken is the token of the run's thread lease acquisition
	// (sessionguard.Token). A Stop's marker and nudge hold it, so a Stop
	// reaches the run that held the thread when the Stop was accepted.
	LeaseToken string `json:"lease_token,omitempty"`
	// Ended reports a run that ended, whose state End kept: the run is not
	// running. Get reports it; Begin never records it.
	Ended bool `json:"-"`
}

// Claim is what claiming a run's terminal state found.
type Claim struct {
	// Held is false when the state is no longer the run's: it lapsed, or a
	// later run's replaced it. Nothing was claimed then.
	Held bool
	// Stopped reports a Stop accepted for the run before the claim.
	Stopped bool
}

// Store keeps one run state per thread. A write lasts one TTL.
//
// thread is the thread's key, the same key the thread's lease is taken under.
type Store interface {
	// Begin records st as the thread's run state, replacing whatever the
	// thread held: only the lease holder begins a run, so a state already
	// there was left by a run that no longer holds the thread, or one that
	// ended.
	Begin(ctx context.Context, thread string, st State) error
	// Renew keeps the thread's run state for one more TTL while it is still
	// the state of the run with this invocation ID, running. It never brings
	// back a state that lapsed or was ended, never extends an ended one, and
	// never touches another run's: renewed is false then, and nothing
	// changes. stopped reports that a Stop was accepted for the run: the
	// marker a run checks at every renewal, in case its nudge never arrived.
	Renew(ctx context.Context, thread, invocationID string) (renewed, stopped bool, err error)
	// End ends the thread's run state if it is still the one of the run with
	// this invocation ID; another run's state is left alone. With keep zero
	// it removes the state. With keep positive it marks the state ended and
	// keeps it for keep from now, as long as the run's Run Log: Get then
	// reports it Ended.
	End(ctx context.Context, thread, invocationID string, keep time.Duration) error
	// Get returns the thread's current run state; ok is false when the
	// thread has none. A state End kept is reported Ended.
	Get(ctx context.Context, thread string) (st State, ok bool, err error)

	// Stop asks the thread's Detached Run to stop. In one step, while the
	// thread's state is a Detached Run's that has not claimed its end, it
	// marks the state with the run's lease token and publishes a nudge for
	// that token; accepted reports that it did, and st is the stopped run's
	// state. An idle thread, a run without the opt-in, a run that already
	// claimed its end and one that ended are left alone. A non-empty
	// invocationID stops only the run with that Invocation. Stopping a run
	// again before its claim is accepted again and changes nothing.
	Stop(ctx context.Context, thread, invocationID string) (st State, accepted bool, err error)
	// Claim claims the terminal state of the run with this invocation ID, in
	// one step on the state a Stop marks: it reports a Stop accepted before
	// it, and a Stop after it is not accepted.
	Claim(ctx context.Context, thread, invocationID string) (Claim, error)
	// Watch delivers a run's nudges: nudged is closed once a Stop is
	// accepted for the run whose lease token is token. A nudge can be lost;
	// the marker Renew reports is the safety net. unwatch ends the watch.
	Watch(ctx context.Context, thread, token string) (nudged <-chan struct{}, unwatch func())
	// Drop removes the thread's run state, whichever run's it is, ended or
	// not. Only the holder of the thread's lease may drop it: deleting a
	// thread does, so a reused threadId never reads as a run of the deleted
	// conversation.
	Drop(ctx context.Context, thread string) error

	// TTL is how long one write lasts.
	TTL() time.Duration
}

var errNoInvocation = errors.New("run state needs an invocation ID")

// stoppable reports whether a Stop for invocationID ("" for any run) reaches
// the run whose state this is.
func stoppable(st State, claimed bool, invocationID string) bool {
	return st.Detached && !st.Ended && !claimed && (invocationID == "" || st.InvocationID == invocationID)
}

// --- Keeping a run's state -----------------------------------------------------

// Kept is a run state its run keeps renewed. End it when the run ends.
type Kept struct {
	store   Store
	thread  string
	st      State
	stop    context.CancelFunc
	done    chan struct{}
	unwatch func()
	once    sync.Once

	stopped  chan struct{}
	stopOnce sync.Once
}

// Keep records st as the thread's run state and renews it every TTL/3 until
// ctx ends or the state is ended. ctx is the run's lease context: the state
// is renewed only while the run holds its thread's lease, so it lapses with
// the lease when the run loses the thread or its process dies.
//
// A renewal that errors is retried sooner, as a lease renewal is. A renewal
// the store refuses — the state lapsed, was ended, or belongs to another run
// now — ends the renewals: a run never brings its state back.
//
// A Detached Run subscribes to its Stop nudges before its state exists, so a
// Stop accepted from then on either nudges it or leaves the marker its next
// renewal finds: a Stop that races the run's start is not lost. Stopped
// reports either.
func Keep(ctx context.Context, s Store, thread string, st State) (*Kept, error) {
	var nudged <-chan struct{}
	unwatch := func() {}
	if st.Detached {
		nudged, unwatch = s.Watch(ctx, thread, st.LeaseToken)
	}
	attempt, cancel := context.WithTimeout(ctx, s.TTL()/3)
	err := s.Begin(attempt, thread, st)
	cancel()
	if err != nil {
		unwatch()
		return nil, err
	}
	renewCtx, stop := context.WithCancel(ctx)
	k := &Kept{
		store: s, thread: thread, st: st, stop: stop, done: make(chan struct{}),
		unwatch: unwatch, stopped: make(chan struct{}),
	}
	go k.renew(renewCtx, nudged)
	return k, nil
}

// State is the run state being kept.
func (k *Kept) State() State { return k.st }

// Stopped is closed once the run learns that a Stop was accepted for it.
func (k *Kept) Stopped() <-chan struct{} { return k.stopped }

// Claim claims the run's terminal state (Store.Claim).
func (k *Kept) Claim(ctx context.Context) (Claim, error) {
	return k.store.Claim(ctx, k.thread, k.st.InvocationID)
}

// End stops the renewals and the run's watch, and ends the run state unless
// another run's state replaced it: it removes it, or with keep positive keeps
// it that long, marked ended (Store.End). Ending again does nothing.
func (k *Kept) End(ctx context.Context, keep time.Duration) error {
	var err error
	k.once.Do(func() {
		k.stop()
		<-k.done
		k.unwatch()
		err = k.store.End(ctx, k.thread, k.st.InvocationID, keep)
	})
	return err
}

func (k *Kept) markStopped() {
	k.stopOnce.Do(func() { close(k.stopped) })
}

func (k *Kept) renew(ctx context.Context, nudged <-chan struct{}) {
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
		case <-nudged:
			// The fast path: the Stop's nudge.
			nudged = nil
			k.markStopped()
			continue
		case <-timer.C:
		}
		attempt, cancel := context.WithTimeout(ctx, interval)
		renewed, stopped, err := k.store.Renew(attempt, k.thread, k.st.InvocationID)
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
			if stopped {
				// The safety net: the marker, for a nudge that never
				// arrived.
				k.markStopped()
			}
			timer.Reset(interval)
		}
	}
}

// --- Nudges --------------------------------------------------------------------

// watchers are the runs of one process waiting for a Stop nudge, by thread.
type watchers struct {
	mu       sync.Mutex
	byThread map[string]map[*watch]struct{}
}

// watch is one run waiting for the nudge of a Stop bound to its token.
type watch struct {
	token  string
	nudged chan struct{}
	once   sync.Once
}

func (w *watch) nudge() { w.once.Do(func() { close(w.nudged) }) }

func (ws *watchers) add(thread, token string) (<-chan struct{}, func()) {
	w := &watch{token: token, nudged: make(chan struct{})}
	ws.mu.Lock()
	defer ws.mu.Unlock()
	if ws.byThread == nil {
		ws.byThread = map[string]map[*watch]struct{}{}
	}
	if ws.byThread[thread] == nil {
		ws.byThread[thread] = map[*watch]struct{}{}
	}
	ws.byThread[thread][w] = struct{}{}
	return w.nudged, func() {
		ws.mu.Lock()
		defer ws.mu.Unlock()
		delete(ws.byThread[thread], w)
		if len(ws.byThread[thread]) == 0 {
			delete(ws.byThread, thread)
		}
	}
}

// nudge wakes the thread's runs waiting with this token: the nudge of a Stop
// bound to it reaches no other run.
func (ws *watchers) nudge(thread, token string) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	for w := range ws.byThread[thread] {
		if w.token == token {
			w.nudge()
		}
	}
}

// --- In-process store ----------------------------------------------------------

// Memory is the in-process Store. It serves a single process and the tests,
// and honors the same TTL as Redis. A Stop nudges the runs of the process
// directly.
type Memory struct {
	ttl     time.Duration
	mu      sync.Mutex
	states  map[string]memoryState
	waiting watchers
}

type memoryState struct {
	st      State
	expires time.Time
	// stopped is the Stop's marker, bound to st.LeaseToken.
	stopped bool
	claimed bool
}

// NewMemory builds an in-process store whose writes last ttl.
func NewMemory(ttl time.Duration) *Memory {
	return &Memory{ttl: ttl, states: make(map[string]memoryState)}
}

var _ Store = (*Memory)(nil)

// current returns the thread's unexpired state. The caller holds m.mu.
func (m *Memory) current(thread string, now time.Time) (memoryState, bool) {
	held, ok := m.states[thread]
	if !ok {
		return memoryState{}, false
	}
	if !now.Before(held.expires) {
		delete(m.states, thread)
		return memoryState{}, false
	}
	return held, true
}

func (m *Memory) Begin(_ context.Context, thread string, st State) error {
	if st.InvocationID == "" {
		return errNoInvocation
	}
	st.Ended = false
	m.mu.Lock()
	defer m.mu.Unlock()
	m.states[thread] = memoryState{st: st, expires: time.Now().Add(m.ttl)}
	return nil
}

func (m *Memory) Renew(_ context.Context, thread, invocationID string) (bool, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	held, ok := m.current(thread, now)
	if !ok || held.st.InvocationID != invocationID || held.st.Ended {
		return false, false, nil
	}
	held.expires = now.Add(m.ttl)
	m.states[thread] = held
	return true, held.stopped, nil
}

func (m *Memory) End(_ context.Context, thread, invocationID string, keep time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	held, ok := m.current(thread, now)
	if !ok || held.st.InvocationID != invocationID {
		return nil
	}
	if keep <= 0 {
		delete(m.states, thread)
		return nil
	}
	held.st.Ended = true
	held.expires = now.Add(keep)
	m.states[thread] = held
	return nil
}

func (m *Memory) Get(_ context.Context, thread string) (State, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	held, ok := m.current(thread, time.Now())
	return held.st, ok, nil
}

func (m *Memory) Stop(_ context.Context, thread, invocationID string) (State, bool, error) {
	m.mu.Lock()
	held, ok := m.current(thread, time.Now())
	if !ok || !stoppable(held.st, held.claimed, invocationID) {
		m.mu.Unlock()
		return State{}, false, nil
	}
	held.stopped = true
	m.states[thread] = held
	m.mu.Unlock()
	m.waiting.nudge(thread, held.st.LeaseToken)
	return held.st, true, nil
}

func (m *Memory) Claim(_ context.Context, thread, invocationID string) (Claim, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	held, ok := m.current(thread, time.Now())
	if !ok || held.st.InvocationID != invocationID {
		return Claim{}, nil
	}
	held.claimed = true
	m.states[thread] = held
	return Claim{Held: true, Stopped: held.stopped}, nil
}

func (m *Memory) Watch(_ context.Context, thread, token string) (<-chan struct{}, func()) {
	return m.waiting.add(thread, token)
}

func (m *Memory) Drop(_ context.Context, thread string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.states, thread)
	return nil
}

func (m *Memory) TTL() time.Duration { return m.ttl }

// --- Redis store -----------------------------------------------------------------

// The run state is a hash: the invocation ID that fences writes, the state
// itself as JSON, and what the scripts read without decoding it — whether the
// run detached and its lease token. A Stop adds its marker ('stop', holding
// the lease token), the claim its own field ('claimed') and an End that keeps
// the state its mark ('ended'); Begin starts from an empty hash, so none
// outlives the run it was for. Renewing, ending, stopping and claiming are
// each one atomic step that checks the invocation ID, so a run that lost its
// thread can neither renew nor end the state of the run that took it, and a
// Stop and a claim are ordered.

var beginScript = redis.NewScript(`
redis.call('DEL', KEYS[1])
redis.call('HSET', KEYS[1], 'invocation_id', ARGV[1], 'state', ARGV[2], 'detached', ARGV[4], 'lease_token', ARGV[5])
redis.call('PEXPIRE', KEYS[1], ARGV[3])
return 1
`)

// renewScript answers 0 for a state that is not the run's, 1 when it renewed
// it, and 2 when it renewed a state a Stop marked.
var renewScript = redis.NewScript(`
if redis.call('HGET', KEYS[1], 'invocation_id') ~= ARGV[1] or redis.call('HGET', KEYS[1], 'ended') == '1' then
  return 0
end
redis.call('PEXPIRE', KEYS[1], ARGV[2])
if redis.call('HEXISTS', KEYS[1], 'stop') == 1 then
  return 2
end
return 1
`)

// endScript ends the run's state: it removes it, or with a keep (ARGV[2] ms)
// marks it ended and keeps it that long.
var endScript = redis.NewScript(`
if redis.call('HGET', KEYS[1], 'invocation_id') ~= ARGV[1] then
  return 0
end
if tonumber(ARGV[2]) > 0 then
  redis.call('HSET', KEYS[1], 'ended', '1')
  redis.call('PEXPIRE', KEYS[1], ARGV[2])
  return 1
end
return redis.call('DEL', KEYS[1])
`)

// stopScript marks the state of a Detached Run that has not claimed its end,
// and nudges the run, in one step: a Stop can neither miss the run it was
// meant for nor reach the next one. ARGV[1] is the invocation ID the Stop is
// for, empty for any run, and ARGV[2] the thread's nudge channel.
var stopScript = redis.NewScript(`
local inv = redis.call('HGET', KEYS[1], 'invocation_id')
if not inv then
  return {0}
end
if ARGV[1] ~= '' and inv ~= ARGV[1] then
  return {0}
end
if redis.call('HGET', KEYS[1], 'detached') ~= '1' or redis.call('HEXISTS', KEYS[1], 'claimed') == 1
  or redis.call('HGET', KEYS[1], 'ended') == '1' then
  return {0}
end
local token = redis.call('HGET', KEYS[1], 'lease_token') or ''
redis.call('HSET', KEYS[1], 'stop', token)
redis.call('PUBLISH', ARGV[2], token)
return {1, redis.call('HGET', KEYS[1], 'state')}
`)

// claimScript answers 0 for a state that is not the run's, 1 when it claimed
// the run's end, and 2 when a Stop was accepted before the claim.
var claimScript = redis.NewScript(`
if redis.call('HGET', KEYS[1], 'invocation_id') ~= ARGV[1] then
  return 0
end
redis.call('HSET', KEYS[1], 'claimed', '1')
if redis.call('HEXISTS', KEYS[1], 'stop') == 1 then
  return 2
end
return 1
`)

// nudgeSubscribeTimeout bounds subscribing to the Stop nudges. Without the
// subscription a run still learns of a Stop, from the marker at its next
// renewal.
const nudgeSubscribeTimeout = 5 * time.Second

// Redis is the cross-Pod Store: one hash per thread, under keyPrefix plus
// the thread's key, written with the TTL. A Stop's nudge is published on the
// thread's channel, keyPrefix + "stop:" + the thread's key; each process
// receives every nudge through one pattern subscription, made on the first
// watch, and hands it to its runs.
type Redis struct {
	rdb    *redis.Client
	prefix string
	ttl    time.Duration

	waiting watchers
	// subscribing serializes making the subscription.
	subscribing sync.Mutex
	subMu       sync.Mutex
	sub         *redis.PubSub
	closed      bool
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

func (r *Redis) nudgePrefix() string { return r.prefix + "stop:" }

func (r *Redis) nudgeChannel(thread string) string { return r.nudgePrefix() + thread }

func (r *Redis) Begin(ctx context.Context, thread string, st State) error {
	if st.InvocationID == "" {
		return errNoInvocation
	}
	raw, err := json.Marshal(st)
	if err != nil {
		return fmt.Errorf("encode run state: %w", err)
	}
	detached := "0"
	if st.Detached {
		detached = "1"
	}
	if err := beginScript.Run(ctx, r.rdb, []string{r.key(thread)},
		st.InvocationID, string(raw), r.ttl.Milliseconds(), detached, st.LeaseToken).Err(); err != nil {
		return fmt.Errorf("record run state: %w", err)
	}
	return nil
}

func (r *Redis) Renew(ctx context.Context, thread, invocationID string) (bool, bool, error) {
	renewed, err := renewScript.Run(ctx, r.rdb, []string{r.key(thread)},
		invocationID, r.ttl.Milliseconds()).Int64()
	if err != nil {
		return false, false, fmt.Errorf("renew run state: %w", err)
	}
	return renewed > 0, renewed == 2, nil
}

func (r *Redis) End(ctx context.Context, thread, invocationID string, keep time.Duration) error {
	keepMs := int64(0)
	if keep > 0 {
		keepMs = max(keep.Milliseconds(), 1)
	}
	if err := endScript.Run(ctx, r.rdb, []string{r.key(thread)}, invocationID, keepMs).Err(); err != nil {
		return fmt.Errorf("end run state: %w", err)
	}
	return nil
}

func (r *Redis) Get(ctx context.Context, thread string) (State, bool, error) {
	fields, err := r.rdb.HMGet(ctx, r.key(thread), "state", "ended").Result()
	if err != nil {
		return State{}, false, fmt.Errorf("read run state: %w", err)
	}
	raw, ok := fields[0].(string)
	if !ok {
		return State{}, false, nil
	}
	st, _, err := decodeState(raw)
	if err != nil {
		return State{}, false, err
	}
	st.Ended = fields[1] == "1"
	return st, true, nil
}

func decodeState(raw string) (State, bool, error) {
	var st State
	if err := json.Unmarshal([]byte(raw), &st); err != nil {
		return State{}, false, fmt.Errorf("decode run state: %w", err)
	}
	return st, true, nil
}

func (r *Redis) Stop(ctx context.Context, thread, invocationID string) (State, bool, error) {
	result, err := stopScript.Run(ctx, r.rdb, []string{r.key(thread)},
		invocationID, r.nudgeChannel(thread)).Slice()
	if err != nil {
		return State{}, false, fmt.Errorf("stop run: %w", err)
	}
	if len(result) < 2 {
		return State{}, false, nil
	}
	raw, _ := result[1].(string)
	st, _, err := decodeState(raw)
	if err != nil {
		// The Stop was accepted all the same; only the report of which run
		// it reached is unreadable.
		log.FromContext(ctx).Warn("stopped run's state is unreadable", "thread", thread, "err", err)
	}
	return st, true, nil
}

func (r *Redis) Claim(ctx context.Context, thread, invocationID string) (Claim, error) {
	claimed, err := claimScript.Run(ctx, r.rdb, []string{r.key(thread)}, invocationID).Int64()
	if err != nil {
		return Claim{}, fmt.Errorf("claim run end: %w", err)
	}
	return Claim{Held: claimed > 0, Stopped: claimed == 2}, nil
}

func (r *Redis) Drop(ctx context.Context, thread string) error {
	if err := r.rdb.Del(ctx, r.key(thread)).Err(); err != nil {
		return fmt.Errorf("drop run state: %w", err)
	}
	return nil
}

func (r *Redis) TTL() time.Duration { return r.ttl }

// Watch registers the run's watch, then makes sure this process receives the
// nudges. A subscription that cannot be made is logged and retried by the
// next watch; meanwhile the run learns of a Stop at its next renewal.
func (r *Redis) Watch(ctx context.Context, thread, token string) (<-chan struct{}, func()) {
	nudged, unwatch := r.waiting.add(thread, token)
	subCtx, cancel := context.WithTimeout(ctx, nudgeSubscribeTimeout)
	defer cancel()
	if err := r.subscribe(subCtx); err != nil {
		log.FromContext(ctx).Warn("stop nudges unavailable; a Stop reaches the run at its next renewal",
			"thread", thread, "err", err)
	}
	return nudged, unwatch
}

// subscribe makes the process's pattern subscription to the nudges, once,
// and waits for Redis to confirm it, so a watch that returns is listening.
// go-redis reconnects and resubscribes it on its own.
func (r *Redis) subscribe(ctx context.Context) error {
	r.subscribing.Lock()
	defer r.subscribing.Unlock()
	r.subMu.Lock()
	ready := r.sub != nil || r.closed
	r.subMu.Unlock()
	if ready {
		return nil
	}
	sub := r.rdb.PSubscribe(ctx, escapeGlob(r.nudgePrefix())+"*")
	if _, err := sub.Receive(ctx); err != nil {
		_ = sub.Close()
		return fmt.Errorf("subscribe to stop nudges: %w", err)
	}
	r.subMu.Lock()
	defer r.subMu.Unlock()
	if r.closed {
		return sub.Close()
	}
	r.sub = sub
	go r.dispatch(sub.Channel())
	return nil
}

// dispatch hands every nudge to the runs of this process waiting for it.
func (r *Redis) dispatch(msgs <-chan *redis.Message) {
	prefix := r.nudgePrefix()
	for msg := range msgs {
		if thread, ok := strings.CutPrefix(msg.Channel, prefix); ok {
			r.waiting.nudge(thread, msg.Payload)
		}
	}
}

// Close ends the process's subscription to the nudges. A process keeps it for
// its lifetime; tests close it.
func (r *Redis) Close() error {
	r.subMu.Lock()
	defer r.subMu.Unlock()
	r.closed = true
	sub := r.sub
	r.sub = nil
	if sub == nil {
		return nil
	}
	return sub.Close()
}

// escapeGlob escapes the glob characters of a Redis pattern.
func escapeGlob(s string) string {
	var b strings.Builder
	for _, c := range s {
		switch c {
		case '*', '?', '[', ']', '\\':
			b.WriteByte('\\')
		}
		b.WriteRune(c)
	}
	return b.String()
}
