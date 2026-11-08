package plane

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrRunBusy         = errors.New("plane run is held by another turn")
	ErrRunLeaseLost    = errors.New("plane run control lease lost")
	ErrPromptDiscarded = errors.New("plane prompt was discarded by Stop")
	ErrTurnConsumed    = errors.New("plane turn may already have executed")
	ErrStopPending     = errors.New("plane Stop is waiting for the holder to settle")
)

const (
	controlPrefix = "butter:plane:run:"
	orderKey      = "butter:plane:accepted-order"
	controlTTL    = 30 * time.Second
	outcomeTTL    = 24 * time.Hour
)

// RunControl is the shared fact on which a turn's completion and Stop race.
// It holds no conversation or pending prompts: those stay in the session
// service and the existing event Stream. The cutoff only filters Stream work
// accepted before a Stop, including work no Worker has read yet.
type RunControl interface {
	Begin(context.Context, *Event, string) error
	Renew(context.Context, *Event, string) (stopped bool, err error)
	Finish(context.Context, *Event, string, RunState) (RunState, error)
	End(context.Context, *Event, string, RunState) error
	RequestStop(context.Context, *Event) (token string, err error)
	Outcome(context.Context, *Event, string) (RunState, bool, error)
}

func controlKey(e *Event) string { return controlPrefix + planeRunKey(e) }
func startedKey(e *Event) string {
	return controlKey(e) + ":started:" + e.ActivityID + ":" + e.Action
}
func outcomeKey(e *Event, token string) string { return controlKey(e) + ":end:" + token }
func queueOrderKey(e *Event) string            { return controlKey(e) + ":orders" }
func doneOrderKey(e *Event) string             { return controlKey(e) + ":done" }

// Admission and direct Stop use exactly the same holder-bound marker rule.
const markStopLua = `
local function markStop(key)
  if redis.call('HGET', key, 'phase') == 'running' then
    local token = redis.call('HGET', key, 'token')
    redis.call('HSET', key, 'stop', token)
    return token
  end
  return ''
end
`

var beginControlScript = redis.NewScript(`
local cutoff = redis.call('GET', KEYS[2])
if cutoff and tonumber(ARGV[2]) <= tonumber(cutoff) then return 'discarded' end
if redis.call('EXISTS', KEYS[3]) == 1 then return 'consumed' end
if tonumber(ARGV[2]) > 0 then
  local previous = redis.call('ZRANGEBYSCORE', KEYS[4], '-inf', tonumber(ARGV[2]) - 1, 'LIMIT', 0, 1)
  if #previous > 0 and redis.call('SISMEMBER', KEYS[5], previous[1]) == 0 then return 'busy' end
end
local phase = redis.call('HGET', KEYS[1], 'phase')
if phase and phase ~= 'ended' then return 'busy' end
redis.call('DEL', KEYS[1])
redis.call('HSET', KEYS[1], 'token', ARGV[1], 'phase', 'running')
redis.call('PEXPIRE', KEYS[1], ARGV[3])
redis.call('SET', KEYS[3], ARGV[1])
return 'started'
`)
var renewControlScript = redis.NewScript(`
if redis.call('HGET', KEYS[1], 'token') ~= ARGV[1] or
   redis.call('HGET', KEYS[1], 'phase') == 'ended' then return -1 end
redis.call('PEXPIRE', KEYS[1], ARGV[2])
if redis.call('HGET', KEYS[1], 'stop') == ARGV[1] then return 1 end
return 0
`)
var finishControlScript = redis.NewScript(`
if redis.call('HGET', KEYS[1], 'token') ~= ARGV[1] then return 'lost' end
local phase = redis.call('HGET', KEYS[1], 'phase')
if phase == 'finishing' then return redis.call('HGET', KEYS[1], 'result') end
if phase ~= 'running' then return 'lost' end
local result = ARGV[2]
if redis.call('HGET', KEYS[1], 'stop') == ARGV[1] then result = 'stopped' end
redis.call('HSET', KEYS[1], 'phase', 'finishing', 'result', result)
return result
`)
var endControlScript = redis.NewScript(`
if redis.call('HGET', KEYS[1], 'token') ~= ARGV[1] or
   redis.call('HGET', KEYS[1], 'phase') ~= 'finishing' or
   redis.call('HGET', KEYS[1], 'result') ~= ARGV[2] then return 0 end
if not redis.acl_check_cmd('HSET', KEYS[1], 'phase', 'ended') or
   not redis.acl_check_cmd('PEXPIRE', KEYS[1], ARGV[3]) or
   not redis.acl_check_cmd('SET', KEYS[2], ARGV[2], 'PX', ARGV[3]) or
   not redis.acl_check_cmd('SADD', KEYS[4], ARGV[4]) then
  return redis.error_reply('plane terminal control permission is missing')
end
redis.call('HSET', KEYS[1], 'phase', 'ended')
redis.call('PEXPIRE', KEYS[1], ARGV[3])
redis.call('SET', KEYS[2], ARGV[2], 'PX', ARGV[3])
if ARGV[4] ~= '0' then redis.call('SADD', KEYS[4], ARGV[4]) end
return 1
`)
var stopControlScript = redis.NewScript(markStopLua + `
local order = redis.call('INCR', KEYS[3])
redis.call('SET', KEYS[2], order)
redis.call('ZREMRANGEBYSCORE', KEYS[4], '-inf', order)
return markStop(KEYS[1])
`)
var outcomeControlScript = redis.NewScript(`
local result = redis.call('GET', KEYS[2])
if result then return result end
if redis.call('HGET', KEYS[1], 'token') ~= ARGV[1] then return 'lost' end
return 'pending'
`)

type RedisRunControl struct{ client *redis.Client }

func NewRedisRunControl(client *redis.Client) *RedisRunControl {
	return &RedisRunControl{client: client}
}
func (c *RedisRunControl) Begin(ctx context.Context, e *Event, token string) error {
	result, err := beginControlScript.Run(ctx, c.client,
		[]string{controlKey(e), controlKey(e) + ":cutoff", startedKey(e), queueOrderKey(e), doneOrderKey(e)},
		token, e.AcceptedOrder, controlTTL.Milliseconds()).Text()
	if err != nil {
		return err
	}
	switch result {
	case "started":
		return nil
	case "discarded":
		return ErrPromptDiscarded
	case "consumed":
		return ErrTurnConsumed
	default:
		return ErrRunBusy
	}
}
func (c *RedisRunControl) Renew(ctx context.Context, e *Event, token string) (bool, error) {
	n, err := renewControlScript.Run(ctx, c.client, []string{controlKey(e)}, token, controlTTL.Milliseconds()).Int()
	if err != nil {
		return false, err
	}
	if n < 0 {
		return false, ErrRunLeaseLost
	}
	return n == 1, nil
}
func (c *RedisRunControl) Finish(ctx context.Context, e *Event, token string, state RunState) (RunState, error) {
	result, err := finishControlScript.Run(ctx, c.client, []string{controlKey(e)}, token, string(state)).Text()
	if err != nil {
		return "", err
	}
	if result == "lost" {
		return "", ErrRunLeaseLost
	}
	return RunState(result), nil
}
func (c *RedisRunControl) End(ctx context.Context, e *Event, token string, state RunState) error {
	n, err := endControlScript.Run(ctx, c.client, []string{controlKey(e), outcomeKey(e, token), queueOrderKey(e), doneOrderKey(e)},
		token, string(state), outcomeTTL.Milliseconds(), max(e.AcceptedOrder, 0)).Int()
	if err != nil {
		return err
	}
	if n != 1 {
		return ErrRunLeaseLost
	}
	return nil
}
func (c *RedisRunControl) RequestStop(ctx context.Context, e *Event) (string, error) {
	return stopControlScript.Run(ctx, c.client, []string{controlKey(e), controlKey(e) + ":cutoff", orderKey, queueOrderKey(e)}).Text()
}
func (c *RedisRunControl) Outcome(ctx context.Context, e *Event, token string) (RunState, bool, error) {
	result, err := outcomeControlScript.Run(ctx, c.client,
		[]string{controlKey(e), outcomeKey(e, token)}, token).Text()
	if err != nil {
		return "", false, err
	}
	if result == "lost" {
		return "", false, ErrRunLeaseLost
	}
	return RunState(result), result != "pending", nil
}

// MemoryRunControl holds the same contract for a single process and tests.
type MemoryRunControl struct {
	mu       sync.Mutex
	order    int64
	runs     map[string]*memoryControl
	started  map[string]bool
	cutoffs  map[string]int64
	outcomes map[string]RunState
	accepted map[string]map[int64]bool
	done     map[string]map[int64]bool
}
type memoryControl struct {
	token, phase string
	stopped      bool
	result       RunState
	expiry       time.Time
}

func NewMemoryRunControl() *MemoryRunControl {
	return &MemoryRunControl{runs: map[string]*memoryControl{}, started: map[string]bool{},
		cutoffs: map[string]int64{}, outcomes: map[string]RunState{}, accepted: map[string]map[int64]bool{}, done: map[string]map[int64]bool{}}
}
func (c *MemoryRunControl) current(e *Event) *memoryControl {
	r := c.runs[controlKey(e)]
	if r != nil && time.Now().Before(r.expiry) {
		return r
	}
	return nil
}
func (c *MemoryRunControl) Begin(_ context.Context, e *Event, token string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.order = max(c.order, e.AcceptedOrder)
	if cutoff, ok := c.cutoffs[controlKey(e)]; ok && e.AcceptedOrder <= cutoff {
		return ErrPromptDiscarded
	}
	if c.started[startedKey(e)] {
		return ErrTurnConsumed
	}
	if e.AcceptedOrder > 0 {
		c.accepted[controlKey(e)] = ensureOrderSet(c.accepted[controlKey(e)])
		c.accepted[controlKey(e)][e.AcceptedOrder] = true
		cutoff := c.cutoffs[controlKey(e)]
		for order := range c.accepted[controlKey(e)] {
			if order < e.AcceptedOrder && order > cutoff && !c.done[controlKey(e)][order] {
				return ErrRunBusy
			}
		}
	}
	if r := c.current(e); r != nil && r.phase != "ended" {
		return ErrRunBusy
	}
	c.runs[controlKey(e)] = &memoryControl{token: token, phase: "running", expiry: time.Now().Add(controlTTL)}
	c.started[startedKey(e)] = true
	return nil
}
func (c *MemoryRunControl) Renew(_ context.Context, e *Event, token string) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	r := c.current(e)
	if r == nil || r.token != token || r.phase == "ended" {
		return false, ErrRunLeaseLost
	}
	r.expiry = time.Now().Add(controlTTL)
	return r.stopped, nil
}
func (c *MemoryRunControl) Finish(_ context.Context, e *Event, token string, state RunState) (RunState, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	r := c.current(e)
	if r == nil || r.token != token || r.phase == "ended" {
		return "", ErrRunLeaseLost
	}
	if r.phase == "finishing" {
		return r.result, nil
	}
	if r.stopped {
		state = RunStateStopped
	}
	r.phase, r.result = "finishing", state
	return state, nil
}
func (c *MemoryRunControl) End(_ context.Context, e *Event, token string, state RunState) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	r := c.current(e)
	if r == nil || r.token != token || r.phase != "finishing" || r.result != state {
		return ErrRunLeaseLost
	}
	r.phase, r.expiry = "ended", time.Now().Add(outcomeTTL)
	c.outcomes[outcomeKey(e, token)] = state
	if e.AcceptedOrder > 0 {
		c.done[controlKey(e)] = ensureOrderSet(c.done[controlKey(e)])
		c.done[controlKey(e)][e.AcceptedOrder] = true
	}
	return nil
}
func (c *MemoryRunControl) RequestStop(_ context.Context, e *Event) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.order++
	c.cutoffs[controlKey(e)] = c.order
	for order := range c.accepted[controlKey(e)] {
		if order <= c.order {
			delete(c.accepted[controlKey(e)], order)
		}
	}
	if r := c.current(e); r != nil && r.phase == "running" {
		r.stopped = true
		return r.token, nil
	}
	return "", nil
}

func ensureOrderSet(in map[int64]bool) map[int64]bool {
	if in == nil {
		return map[int64]bool{}
	}
	return in
}
func (c *MemoryRunControl) Outcome(_ context.Context, e *Event, token string) (RunState, bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if state, ok := c.outcomes[outcomeKey(e, token)]; ok {
		return state, true, nil
	}
	if r := c.current(e); r == nil || r.token != token {
		return "", false, ErrRunLeaseLost
	}
	return "", false, nil
}
