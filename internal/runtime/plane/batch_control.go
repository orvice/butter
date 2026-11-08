package plane

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/redis/go-redis/v9"
)

// Reservations do not mark a Prompt started. Losing one before StartBatch is
// safe to retry. All selected members cross the execution boundary atomically.
var reserveBatchScript = redis.NewScript(`
local types = {'hash', 'string', 'zset', 'set', 'hash'}
for i = 1, #KEYS do
  local kind = redis.call('TYPE', KEYS[i]).ok
  local want = types[i] or 'string'
  if kind ~= 'none' and kind ~= want then return redis.error_reply('invalid Plane reservation index') end
end
local orders = cjson.decode(ARGV[3])
local cutoff = tonumber(redis.call('GET', KEYS[2]) or '-1')
if tonumber(orders[1]) <= cutoff then return 'discarded' end
local selected = {}
for i, order in ipairs(orders) do
  selected[order] = true
  if redis.call('EXISTS', KEYS[5+i]) == 1 then return 'consumed' end
end
local earlier = redis.call('ZRANGEBYSCORE', KEYS[3], '-inf', tonumber(orders[#orders]))
for _, order in ipairs(earlier) do
  if not selected[order] and redis.call('SISMEMBER', KEYS[4], order) == 0 then return 'busy' end
end
local phase = redis.call('HGET', KEYS[1], 'phase')
if phase and phase ~= 'ended' then return 'busy' end
if not redis.acl_check_cmd('DEL', KEYS[1]) or
   not redis.acl_check_cmd('HSET', KEYS[1], 'token', ARGV[1], 'phase', 'reserved', 'orders', ARGV[3], 'agent', ARGV[4]) or
   not redis.acl_check_cmd('PEXPIRE', KEYS[1], ARGV[2]) then return redis.error_reply('Plane reservation permission missing') end
redis.call('DEL', KEYS[1])
redis.call('HSET', KEYS[1], 'token', ARGV[1], 'phase', 'reserved', 'orders', ARGV[3], 'agent', ARGV[4])
redis.call('PEXPIRE', KEYS[1], ARGV[2])
return 'reserved'
`)

var startBatchScript = redis.NewScript(`
if redis.call('HGET', KEYS[1], 'token') ~= ARGV[1] or
   redis.call('HGET', KEYS[1], 'phase') ~= 'reserved' or
   redis.call('HGET', KEYS[1], 'orders') ~= ARGV[2] then return 'lost' end
local orders = cjson.decode(ARGV[2])
local cutoff = tonumber(redis.call('GET', KEYS[2]) or '-1')
if tonumber(orders[1]) <= cutoff or redis.call('HGET', KEYS[1], 'stop') == ARGV[1] then return 'discarded' end
for i = 3, #KEYS do
  local kind = redis.call('TYPE', KEYS[i]).ok
  if kind ~= 'none' then return redis.error_reply('Plane started marker already exists') end
  if not redis.acl_check_cmd('SET', KEYS[i], ARGV[1]) then return redis.error_reply('Plane start permission missing') end
end
if not redis.acl_check_cmd('HSET', KEYS[1], 'phase', 'running') then return redis.error_reply('Plane start permission missing') end
for i = 3, #KEYS do redis.call('SET', KEYS[i], ARGV[1]) end
redis.call('HSET', KEYS[1], 'phase', 'running')
return 'started'
`)

var releaseBatchScript = redis.NewScript(`
if redis.call('HGET', KEYS[1], 'token') == ARGV[1] and
   redis.call('HGET', KEYS[1], 'phase') == 'reserved' then redis.call('DEL', KEYS[1]) end
return 1
`)

var endBatchScript = redis.NewScript(`
if redis.call('HGET', KEYS[1], 'token') ~= ARGV[1] or
   redis.call('HGET', KEYS[1], 'phase') ~= 'finishing' or
   redis.call('HGET', KEYS[1], 'result') ~= ARGV[2] or
   redis.call('HGET', KEYS[1], 'orders') ~= ARGV[4] then return 0 end
local orders = cjson.decode(ARGV[4])
local types = {'hash', 'string', 'zset', 'set', 'hash'}
for i = 1, #KEYS do
  local want = types[i] or 'string'
  local kind = redis.call('TYPE', KEYS[i]).ok
  if kind ~= 'none' and kind ~= want then return redis.error_reply('invalid Plane batch terminal index') end
end
if not redis.acl_check_cmd('HSET', KEYS[1], 'phase', 'ended') or
   not redis.acl_check_cmd('PEXPIRE', KEYS[1], ARGV[3]) or
   not redis.acl_check_cmd('SET', KEYS[2], ARGV[2], 'PX', ARGV[3]) then return redis.error_reply('Plane end permission missing') end
for i, order in ipairs(orders) do
  if not redis.acl_check_cmd('SET', KEYS[5+i], ARGV[2]) or
     not redis.acl_check_cmd('SADD', KEYS[4], order) or
     not redis.acl_check_cmd('ZREM', KEYS[3], order) or
     not redis.acl_check_cmd('HDEL', KEYS[5], order) then return redis.error_reply('Plane end permission missing') end
end
for i, order in ipairs(orders) do
  redis.call('SET', KEYS[5+i], ARGV[2])
  redis.call('SADD', KEYS[4], order)
  redis.call('ZREM', KEYS[3], order)
  redis.call('HDEL', KEYS[5], order)
end
redis.call('HSET', KEYS[1], 'phase', 'ended')
redis.call('PEXPIRE', KEYS[1], ARGV[3])
redis.call('SET', KEYS[2], ARGV[2], 'PX', ARGV[3])
return 1
`)

func batchOrders(e *Event) string {
	orders := []string{}
	for _, member := range membersOf(e) {
		orders = append(orders, formatOrder(member.AcceptedOrder))
	}
	raw, _ := json.Marshal(orders)
	return string(raw)
}

func (c *RedisRunControl) reserveBatch(ctx context.Context, e *Event, token string) error {
	keys := []string{controlKey(e), controlKey(e) + ":cutoff", queueOrderKey(e), doneOrderKey(e), streamIndexKey(e)}
	for _, member := range e.Members {
		keys = append(keys, startedKey(member))
	}
	result, err := reserveBatchScript.Run(ctx, c.client, keys, token, controlTTL.Milliseconds(), batchOrders(e), e.AgentID).Text()
	if err != nil {
		return err
	}
	switch result {
	case "reserved":
		return nil
	case "discarded":
		return ErrPromptDiscarded
	case "consumed":
		return ErrRunBusy // replay through prepare to recover the stored outcome
	default:
		return ErrRunBusy
	}
}

func (c *RedisRunControl) StartBatch(ctx context.Context, e *Event, token string) error {
	keys := []string{controlKey(e), controlKey(e) + ":cutoff"}
	for _, member := range e.Members {
		keys = append(keys, startedKey(member))
	}
	result, err := startBatchScript.Run(ctx, c.client, keys, token, batchOrders(e)).Text()
	if err != nil {
		return err
	}
	if result == "discarded" {
		return ErrPromptDiscarded
	}
	if result != "started" {
		return ErrRunLeaseLost
	}
	return nil
}

func (c *RedisRunControl) ReleaseBatch(ctx context.Context, e *Event, token string) error {
	return releaseBatchScript.Run(ctx, c.client, []string{controlKey(e)}, token).Err()
}

func (c *RedisRunControl) endBatch(ctx context.Context, e *Event, token string, state RunState) error {
	keys := []string{controlKey(e), outcomeKey(e, token), queueOrderKey(e), doneOrderKey(e), streamIndexKey(e)}
	for _, member := range e.Members {
		keys = append(keys, eventResultKey(member))
	}
	n, err := endBatchScript.Run(ctx, c.client, keys, token, string(state), outcomeTTL.Milliseconds(), batchOrders(e)).Int()
	if err != nil {
		return err
	}
	if n != 1 {
		return errors.Join(ErrRunLeaseLost, errors.New("batch End fence rejected"))
	}
	return nil
}
