package linear

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// FollowUp is a message that arrived while its session's turn was running.
// It waits in the session's follow-up list until the lease holder drains it.
type FollowUp struct {
	RecordID        string `json:"record_id,omitempty"`
	DeliveryID      string `json:"delivery_id"`
	Text            string `json:"text"`
	PromptingUserID string `json:"prompting_user_id,omitempty"`
}

// ErrLeaseLost means another holder took the session.
var ErrLeaseLost = errors.New("linear session lease lost")

// SessionCoordinator serializes turns within one Linear Agent Session across
// Pods and queues follow-ups that arrive mid-turn (ADR-0015 §6).
//
// Enqueue-or-acquire and release-or-drain are each atomic with respect to
// the other, so a follow-up accepted during a turn is never lost and never
// runs concurrently with that turn.
type SessionCoordinator interface {
	// EnqueueOrAcquire takes the session's turn — returning a hold and any
	// backlog a previous holder left — or, when another turn holds it,
	// appends followUp to the session's list and returns a nil hold.
	EnqueueOrAcquire(ctx context.Context, sessionKey string, followUp FollowUp) (SessionHold, []FollowUp, error)
	// TryAcquire takes the session when it is free, with its backlog. A nil
	// hold means another turn holds it.
	TryAcquire(ctx context.Context, sessionKey string) (SessionHold, []FollowUp, error)
	// Discard removes and returns every queued follow-up.
	Discard(ctx context.Context, sessionKey string) ([]FollowUp, error)
	// RequestStop discards every queued follow-up and, when a turn holds
	// the session, tells that holder to stop. It never takes the lease, so
	// a stop never waits behind the turn it is meant to stop. held reports
	// whether a turn was running.
	RequestStop(ctx context.Context, sessionKey string) (held bool, discarded []FollowUp, err error)
}

// SessionHold is one holder's turn on a session.
type SessionHold interface {
	// Context is cancelled when the lease is lost to another holder.
	Context() context.Context
	// ReleaseOrDrain either releases the session, returning nothing, or —
	// when follow-ups are queued — takes them all and keeps the lease for
	// the turn that answers them.
	ReleaseOrDrain(ctx context.Context) ([]FollowUp, error)
	// Abandon gives the session up without draining: queued follow-ups
	// stay for the next holder.
	Abandon()
	// Stopped is closed when a stop is requested for this holder.
	Stopped() <-chan struct{}
	// AcknowledgeStop clears a handled stop so later turns in this hold run.
	AcknowledgeStop(ctx context.Context)
}

// --- In-process coordinator --------------------------------------------------

// MemoryCoordinator coordinates sessions within one process. It is the
// single-Pod fallback and what tests use.
type MemoryCoordinator struct {
	mu       sync.Mutex
	sessions map[string]*memorySession
}

type memorySession struct {
	holder string
	cancel context.CancelFunc
	items  []FollowUp
	hold   *memoryHold
}

func NewMemoryCoordinator() *MemoryCoordinator {
	return &MemoryCoordinator{sessions: map[string]*memorySession{}}
}

var _ SessionCoordinator = (*MemoryCoordinator)(nil)

func (c *MemoryCoordinator) session(key string) *memorySession {
	s, ok := c.sessions[key]
	if !ok {
		s = &memorySession{}
		c.sessions[key] = s
	}
	return s
}

// acquireLocked takes a free session. The caller holds c.mu.
func (c *MemoryCoordinator) acquireLocked(ctx context.Context, key string, s *memorySession) (SessionHold, []FollowUp) {
	holdCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))
	s.holder = uuid.NewString()
	s.cancel = cancel
	backlog := s.items
	s.items = nil
	hold := &memoryHold{c: c, key: key, token: s.holder, ctx: holdCtx, cancel: cancel, stopped: make(chan struct{})}
	s.hold = hold
	return hold, backlog
}

func (c *MemoryCoordinator) EnqueueOrAcquire(ctx context.Context, key string, followUp FollowUp) (SessionHold, []FollowUp, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	s := c.session(key)
	if s.holder != "" {
		s.items = append(s.items, followUp)
		return nil, nil, nil
	}
	hold, backlog := c.acquireLocked(ctx, key, s)
	return hold, backlog, nil
}

func (c *MemoryCoordinator) TryAcquire(ctx context.Context, key string) (SessionHold, []FollowUp, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	s := c.session(key)
	if s.holder != "" {
		return nil, nil, nil
	}
	hold, backlog := c.acquireLocked(ctx, key, s)
	return hold, backlog, nil
}

func (c *MemoryCoordinator) Discard(_ context.Context, key string) ([]FollowUp, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	s := c.session(key)
	items := s.items
	s.items = nil
	return items, nil
}

func (c *MemoryCoordinator) RequestStop(_ context.Context, key string) (bool, []FollowUp, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	s := c.session(key)
	discarded := s.items
	s.items = nil
	if s.holder == "" || s.hold == nil || s.hold.token != s.holder {
		return false, discarded, nil
	}
	s.hold.stop()
	return true, discarded, nil
}

// Steal simulates another holder taking the session: the current holder's
// context is cancelled and its lease is gone. Used by tests.
func (c *MemoryCoordinator) Steal(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	s := c.session(key)
	if s.cancel != nil {
		s.cancel()
	}
	s.holder = "stolen-" + uuid.NewString()
	s.cancel = nil
}

type memoryHold struct {
	c       *MemoryCoordinator
	key     string
	token   string
	ctx     context.Context
	cancel  context.CancelFunc
	stopMu  sync.Mutex
	stopped chan struct{}
	isStop  bool
}

func (h *memoryHold) Context() context.Context { return h.ctx }

func (h *memoryHold) stop() {
	h.stopMu.Lock()
	defer h.stopMu.Unlock()
	if !h.isStop {
		h.isStop = true
		close(h.stopped)
	}
}

func (h *memoryHold) Stopped() <-chan struct{} {
	h.stopMu.Lock()
	defer h.stopMu.Unlock()
	return h.stopped
}

func (h *memoryHold) AcknowledgeStop(context.Context) {
	h.stopMu.Lock()
	defer h.stopMu.Unlock()
	if h.isStop {
		h.isStop = false
		h.stopped = make(chan struct{})
	}
}

func (h *memoryHold) ReleaseOrDrain(context.Context) ([]FollowUp, error) {
	h.c.mu.Lock()
	defer h.c.mu.Unlock()
	s := h.c.session(h.key)
	if s.holder != h.token {
		return nil, ErrLeaseLost
	}
	if len(s.items) == 0 {
		s.holder = ""
		s.cancel = nil
		s.hold = nil
		h.cancel()
		return nil, nil
	}
	items := s.items
	s.items = nil
	return items, nil
}

func (h *memoryHold) Abandon() {
	h.c.mu.Lock()
	defer h.c.mu.Unlock()
	s := h.c.session(h.key)
	if s.holder == h.token {
		s.holder = ""
		s.cancel = nil
		s.hold = nil
	}
	h.cancel()
}

// --- Redis coordinator ---------------------------------------------------------

const (
	sessionLeasePrefix = "butter:linear:lease:session:"
	followUpListPrefix = "butter:linear:followups:"
	stopMarkerPrefix   = "butter:linear:stop:"
	stopChannelPrefix  = "butter:linear:stop-nudge:"
	// followUpListTTL bounds how long an abandoned list survives.
	followUpListTTL = 24 * time.Hour
	// stopMarkerTTL bounds how long a stop waits for its holder to see it.
	stopMarkerTTL = 10 * time.Minute
)

// enqueueOrAcquireScript takes the free lease (draining any backlog a dead
// holder left) or queues the follow-up, in one step.
var enqueueOrAcquireScript = redis.NewScript(`
if redis.call('SET', KEYS[1], ARGV[1], 'NX', 'PX', ARGV[2]) then
  local items = redis.call('LRANGE', KEYS[2], 0, -1)
  redis.call('DEL', KEYS[2])
  table.insert(items, 1, 'acquired')
  return items
end
redis.call('RPUSH', KEYS[2], ARGV[3])
redis.call('PEXPIRE', KEYS[2], ARGV[4])
return {'queued'}
`)

var tryAcquireScript = redis.NewScript(`
if redis.call('SET', KEYS[1], ARGV[1], 'NX', 'PX', ARGV[2]) then
  local items = redis.call('LRANGE', KEYS[2], 0, -1)
  redis.call('DEL', KEYS[2])
  table.insert(items, 1, 'acquired')
  return items
end
return {'busy'}
`)

// releaseOrDrainScript releases when nothing is queued, or drains the list
// and keeps the lease — never both, never neither.
var releaseOrDrainScript = redis.NewScript(`
if redis.call('GET', KEYS[1]) ~= ARGV[1] then
  return {'lost'}
end
local items = redis.call('LRANGE', KEYS[2], 0, -1)
if #items == 0 then
  redis.call('DEL', KEYS[1])
  return {'released'}
end
redis.call('DEL', KEYS[2])
redis.call('PEXPIRE', KEYS[1], ARGV[2])
table.insert(items, 1, 'drained')
return items
`)

var discardScript = redis.NewScript(`
local items = redis.call('LRANGE', KEYS[1], 0, -1)
redis.call('DEL', KEYS[1])
return items
`)

// requestStopScript discards the queue and, while a turn holds the
// session, marks that holder stopped and nudges it — one step, so a stop can
// neither miss the holder it was meant for nor reach the next one.
var requestStopScript = redis.NewScript(`
local items = redis.call('LRANGE', KEYS[2], 0, -1)
redis.call('DEL', KEYS[2])
local holder = redis.call('GET', KEYS[1])
if not holder then
  table.insert(items, 1, 'idle')
  return items
end
redis.call('SET', KEYS[3], holder, 'PX', ARGV[1])
redis.call('PUBLISH', ARGV[2], holder)
table.insert(items, 1, 'held')
return items
`)

var clearStopScript = redis.NewScript(`
if redis.call('GET', KEYS[1]) == ARGV[1] then
  return redis.call('DEL', KEYS[1])
end
return 0
`)

var renewScript = redis.NewScript(`
if redis.call('GET', KEYS[1]) == ARGV[1] then
  return redis.call('PEXPIRE', KEYS[1], ARGV[2])
end
return 0
`)

var abandonScript = redis.NewScript(`
if redis.call('GET', KEYS[1]) == ARGV[1] then
  return redis.call('DEL', KEYS[1])
end
return 0
`)

// RedisCoordinator is the cross-Pod coordinator: a renewable lease per
// session, fenced on a per-acquisition token, plus a follow-up list.
type RedisCoordinator struct {
	rdb    *redis.Client
	holder string
	ttl    time.Duration
	prefix string
	// beforeSubscribe runs between acquiring a session and subscribing to
	// its stop nudges. Tests use it to race a stop against a turn's start.
	beforeSubscribe func()
}

// NewRedisCoordinator builds the coordinator. ttl bounds how long a crashed
// holder blocks a session; the lease is renewed while a turn runs.
func NewRedisCoordinator(rdb *redis.Client, holder string, ttl time.Duration) *RedisCoordinator {
	if rdb == nil {
		return nil
	}
	return &RedisCoordinator{rdb: rdb, holder: holder, ttl: ttl}
}

var _ SessionCoordinator = (*RedisCoordinator)(nil)

func (c *RedisCoordinator) keys(key string) []string {
	return []string{c.prefix + sessionLeasePrefix + key, c.prefix + followUpListPrefix + key}
}

func (c *RedisCoordinator) stopKey(key string) string     { return c.prefix + stopMarkerPrefix + key }
func (c *RedisCoordinator) stopChannel(key string) string { return c.prefix + stopChannelPrefix + key }

func (c *RedisCoordinator) RequestStop(ctx context.Context, key string) (bool, []FollowUp, error) {
	keys := append(c.keys(key), c.stopKey(key))
	result, err := requestStopScript.Run(ctx, c.rdb, []string{keys[0], keys[1], keys[2]},
		stopMarkerTTL.Milliseconds(), c.stopChannel(key)).Result()
	if err != nil {
		return false, nil, fmt.Errorf("request linear stop: %w", err)
	}
	status, raw := scriptResult(result)
	discarded, err := decodeFollowUps(raw)
	if err != nil {
		return false, nil, err
	}
	return status == "held", discarded, nil
}

func decodeFollowUps(raw []any) ([]FollowUp, error) {
	out := make([]FollowUp, 0, len(raw))
	for _, item := range raw {
		s, _ := item.(string)
		var f FollowUp
		if err := json.Unmarshal([]byte(s), &f); err != nil {
			return nil, fmt.Errorf("decode linear follow-up: %w", err)
		}
		out = append(out, f)
	}
	return out, nil
}

func scriptResult(result any) (string, []any) {
	items, _ := result.([]any)
	if len(items) == 0 {
		return "", nil
	}
	status, _ := items[0].(string)
	return status, items[1:]
}

func (c *RedisCoordinator) acquired(ctx context.Context, key, token string, raw []any) (SessionHold, []FollowUp, error) {
	backlog, err := decodeFollowUps(raw)
	if err != nil {
		return nil, nil, err
	}
	return c.startHold(ctx, key, token), backlog, nil
}

func (c *RedisCoordinator) token() string { return c.holder + ":" + uuid.NewString() }

func (c *RedisCoordinator) EnqueueOrAcquire(ctx context.Context, key string, followUp FollowUp) (SessionHold, []FollowUp, error) {
	item, err := json.Marshal(followUp)
	if err != nil {
		return nil, nil, err
	}
	token := c.token()
	result, err := enqueueOrAcquireScript.Run(ctx, c.rdb, c.keys(key),
		token, c.ttl.Milliseconds(), string(item), followUpListTTL.Milliseconds()).Result()
	if err != nil {
		return nil, nil, fmt.Errorf("enqueue or acquire linear session: %w", err)
	}
	status, raw := scriptResult(result)
	if status != "acquired" {
		return nil, nil, nil
	}
	return c.acquired(ctx, key, token, raw)
}

func (c *RedisCoordinator) TryAcquire(ctx context.Context, key string) (SessionHold, []FollowUp, error) {
	token := c.token()
	result, err := tryAcquireScript.Run(ctx, c.rdb, c.keys(key), token, c.ttl.Milliseconds()).Result()
	if err != nil {
		return nil, nil, fmt.Errorf("acquire linear session: %w", err)
	}
	status, raw := scriptResult(result)
	if status != "acquired" {
		return nil, nil, nil
	}
	return c.acquired(ctx, key, token, raw)
}

func (c *RedisCoordinator) Discard(ctx context.Context, key string) ([]FollowUp, error) {
	result, err := discardScript.Run(ctx, c.rdb, c.keys(key)[1:]).Result()
	if err != nil {
		return nil, fmt.Errorf("discard linear follow-ups: %w", err)
	}
	raw, _ := result.([]any)
	return decodeFollowUps(raw)
}

func (c *RedisCoordinator) startHold(ctx context.Context, key, token string) *redisHold {
	holdCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))
	h := &redisHold{c: c, key: key, token: token, ctx: holdCtx, cancel: cancel,
		done: make(chan struct{}), stopped: make(chan struct{})}
	if c.beforeSubscribe != nil {
		c.beforeSubscribe()
	}
	// Subscribe first, then read the marker: a stop between acquisition
	// and subscription is caught by the marker, one after by the nudge.
	h.sub = c.rdb.Subscribe(holdCtx, c.stopChannel(key))
	if _, err := h.sub.Receive(holdCtx); err != nil {
		_ = h.sub.Close()
		h.sub = nil
	}
	h.checkMarker(holdCtx)
	go h.renew()
	return h
}

type redisHold struct {
	c      *RedisCoordinator
	key    string
	token  string
	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}
	once   sync.Once
	sub    *redis.PubSub

	stopMu  sync.Mutex
	stopped chan struct{}
	isStop  bool
}

func (h *redisHold) Context() context.Context { return h.ctx }

func (h *redisHold) markStopped() {
	h.stopMu.Lock()
	defer h.stopMu.Unlock()
	if !h.isStop {
		h.isStop = true
		close(h.stopped)
	}
}

func (h *redisHold) Stopped() <-chan struct{} {
	h.stopMu.Lock()
	defer h.stopMu.Unlock()
	return h.stopped
}

func (h *redisHold) AcknowledgeStop(ctx context.Context) {
	clearCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	_ = clearStopScript.Run(clearCtx, h.c.rdb, []string{h.c.stopKey(h.key)}, h.token).Err()
	h.stopMu.Lock()
	defer h.stopMu.Unlock()
	if h.isStop {
		h.isStop = false
		h.stopped = make(chan struct{})
	}
}

// checkMarker honors a stop marked for this holder.
func (h *redisHold) checkMarker(ctx context.Context) {
	if marker, err := h.c.rdb.Get(ctx, h.c.stopKey(h.key)).Result(); err == nil && marker == h.token {
		h.markStopped()
	}
}

func (h *redisHold) renew() {
	defer close(h.done)
	ticker := time.NewTicker(h.c.ttl / 3)
	defer ticker.Stop()
	var nudges <-chan *redis.Message
	if h.sub != nil {
		nudges = h.sub.Channel()
	}
	for {
		select {
		case <-h.ctx.Done():
			return
		case msg, ok := <-nudges:
			if !ok {
				nudges = nil
				continue
			}
			if msg.Payload == h.token {
				h.markStopped()
			}
		case <-ticker.C:
			renewed, err := renewScript.Run(h.ctx, h.c.rdb, h.c.keys(h.key)[:1], h.token, h.c.ttl.Milliseconds()).Int64()
			if err != nil || renewed == 0 {
				h.cancel()
				return
			}
			// The nudge is the fast path; the marker is the safety net.
			h.checkMarker(h.ctx)
		}
	}
}

func (h *redisHold) stop() {
	h.once.Do(func() {
		h.cancel()
		<-h.done
		if h.sub != nil {
			_ = h.sub.Close()
		}
	})
}

func (h *redisHold) ReleaseOrDrain(ctx context.Context) ([]FollowUp, error) {
	if h.ctx.Err() != nil {
		return nil, ErrLeaseLost
	}
	runCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	result, err := releaseOrDrainScript.Run(runCtx, h.c.rdb, h.c.keys(h.key), h.token, h.c.ttl.Milliseconds()).Result()
	if err != nil {
		return nil, fmt.Errorf("release or drain linear session: %w", err)
	}
	status, raw := scriptResult(result)
	switch status {
	case "released":
		h.stop()
		return nil, nil
	case "drained":
		return decodeFollowUps(raw)
	default:
		h.stop()
		return nil, ErrLeaseLost
	}
}

func (h *redisHold) Abandon() {
	h.stop()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = abandonScript.Run(ctx, h.c.rdb, h.c.keys(h.key)[:1], h.token).Err()
}
