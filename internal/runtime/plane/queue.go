package plane

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"go.orx.me/apps/butter/internal/eventqueue"
	"time"
)

const (
	StreamKey        = "butter:plane:events"
	ConsumerGroup    = "butter-plane-workers"
	ControlStreamKey = "butter:plane:controls"
)

type StreamQueue struct {
	inner    *eventqueue.Queue
	controls *eventqueue.Queue
	client   *redis.Client
}

// The webhook's successful append is also the Stop linearization point.
// No Worker scheduling or model capacity participates in this decision.
var acceptPlaneScript = redis.NewScript(markStopLua + `
-- Lua is atomic but does not roll back commands after a runtime error.
-- Validate destination types before setting a cutoff or a holder marker.
local types = {'string', 'stream', 'hash', 'string', 'string', 'zset'}
for i = 1, #KEYS do
  local kind = redis.call('TYPE', KEYS[i]).ok
  if kind ~= 'none' and kind ~= types[i] then
    return redis.error_reply('plane admission key has an incompatible type')
  end
end
if redis.call('EXISTS', KEYS[1]) == 1 then return nil end
local event = cjson.decode(ARGV[1])
-- Check every potentially denied mutation before changing the Stop fact.
if not redis.acl_check_cmd('INCR', KEYS[5]) or
   not redis.acl_check_cmd('XADD', KEYS[2], '*', 'event', ARGV[1]) or
   not redis.acl_check_cmd('SET', KEYS[1], '1', 'PX', ARGV[2]) then
  return redis.error_reply('plane admission write permission is missing')
end
if event.stop and (not redis.acl_check_cmd('SET', KEYS[4], '0') or
   not redis.acl_check_cmd('HGET', KEYS[3], 'phase') or
   not redis.acl_check_cmd('HSET', KEYS[3], 'stop', 'token')) then
  return redis.error_reply('plane Stop control permission is missing')
end
local order = redis.call('INCR', KEYS[5])
event.accepted_order = order
if event.stop then
  redis.call('SET', KEYS[4], order)
  event.stop_target = markStop(KEYS[3])
  redis.call('ZREMRANGEBYSCORE', KEYS[6], '-inf', order)
else
  redis.call('ZADD', KEYS[6], order, tostring(order))
end
local id = redis.call('XADD', KEYS[2], '*', 'event', cjson.encode(event))
redis.call('SET', KEYS[1], '1', 'PX', ARGV[2])
return id
`)

func NewQueue(rdb *redis.Client) *StreamQueue {
	if rdb == nil {
		return nil
	}
	return &StreamQueue{client: rdb,
		inner:    eventqueue.New(rdb, eventqueue.Config{Name: "plane", StreamKey: StreamKey, ConsumerGroup: ConsumerGroup, DedupeKeyPrefix: "butter:plane:seen:", DedupeTTL: 24 * time.Hour, LeasePreflightKeyPrefix: "butter:plane:lease:preflight:"}),
		controls: eventqueue.New(rdb, eventqueue.Config{Name: "plane control", StreamKey: ControlStreamKey, ConsumerGroup: ConsumerGroup}),
	}
}
func (q *StreamQueue) Accept(ctx context.Context, e *Event) (string, error) {
	if q == nil || q.inner == nil {
		return "", context.Canceled
	}
	p, err := e.Encode()
	if err != nil {
		return "", err
	}
	id := e.InstallationID + ":" + e.AgentRunID + ":" + e.ActivityID + ":" + e.Action
	stream := StreamKey
	if e.Stop {
		stream = ControlStreamKey
	}
	if _, err := acceptPlaneScript.Run(ctx, q.client,
		[]string{"butter:plane:seen:" + id, stream, controlKey(e), controlKey(e) + ":cutoff", orderKey, queueOrderKey(e)},
		p, (24 * time.Hour).Milliseconds()).Result(); err != nil {
		if errors.Is(err, redis.Nil) {
			return "", ErrDuplicate
		}
		return "", err
	}
	return id, nil
}
func (q *StreamQueue) Available() bool                { return q != nil && q.inner != nil && q.inner.Available() }
func (q *StreamQueue) Ping(ctx context.Context) error { return q.inner.Ping(ctx) }
func (q *StreamQueue) EnsureGroup(ctx context.Context) error {
	if err := q.inner.EnsureGroup(ctx); err != nil {
		return err
	}
	return q.controls.EnsureGroup(ctx)
}
func (q *StreamQueue) Read(ctx context.Context, c string, n int64, b time.Duration) ([]eventqueue.Entry, error) {
	return q.inner.Read(ctx, c, n, b)
}
func (q *StreamQueue) Ack(ctx context.Context, ids ...string) error { return q.inner.Ack(ctx, ids...) }
