package plane

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
	"go.orx.me/apps/butter/internal/runtime/memoryhook"
	"go.orx.me/apps/butter/internal/userinput"
	"google.golang.org/genai"
)

var errTurnInProgress = fmt.Errorf("%w: this prompt is already in the active batch", ErrRunBusy)

var ErrPromptRejected = errors.New("Plane prompt is no longer authorized")

var ErrInvalidInput = errors.New("invalid Plane prompt input")

const maxBatchPrompts = 32
const batchTextBudget = userinput.MaxTextBytes

func streamIndexKey(e *Event) string { return controlKey(e) + ":stream-index" }
func eventResultKey(e *Event) string { return startedKey(e) + ":result" }
func formatOrder(n int64) string     { return strconv.FormatInt(n, 10) }

type consumedTurn struct {
	invocation string
	state      RunState
}

func (e *consumedTurn) Error() string { return "plane prompt already started: " + string(e.state) }
func (e *consumedTurn) Unwrap() error { return ErrTurnConsumed }

type redisBatchSource struct{ client *redis.Client }

// Recovery never re-executes a started turn. A reserved turn has no started
// marker and can be reclaimed after its lease expires. Per-event results outlive
// the lease and the 24-hour Stop observation window.
var recoverBatchEventScript = redis.NewScript(`
local types = {'hash', 'string', 'string', 'zset', 'hash', 'string'}
for i, want in ipairs(types) do
  local kind = redis.call('TYPE', KEYS[i]).ok
  if kind ~= 'none' and kind ~= want then return redis.error_reply('invalid Plane recovery index') end
end
local token = redis.call('GET', KEYS[2])
if not token then
  local cutoff = redis.call('GET', KEYS[6])
  if cutoff and tonumber(ARGV[1]) <= tonumber(cutoff) then return {'discarded'} end
  return {'ready'}
end
local current = redis.call('HGET', KEYS[1], 'token')
local phase = redis.call('HGET', KEYS[1], 'phase')
if token == current and phase ~= 'ended' then return {'busy'} end
if not redis.acl_check_cmd('SET', KEYS[3], 'uncertain') or
   not redis.acl_check_cmd('ZREM', KEYS[4], ARGV[1]) or
   not redis.acl_check_cmd('HDEL', KEYS[5], ARGV[1]) then return redis.error_reply('Plane recovery permission missing') end
local result = redis.call('GET', KEYS[3])
if not result then
  if token == current and phase == 'ended' then
    result = redis.call('HGET', KEYS[1], 'result')
  end
  result = result or 'uncertain'
  redis.call('SET', KEYS[3], result)
end
redis.call('ZREM', KEYS[4], ARGV[1])
redis.call('HDEL', KEYS[5], ARGV[1])
return {'consumed', token, result}
`)

func (s *redisBatchSource) prepare(ctx context.Context, e *Event) error {
	result, err := recoverBatchEventScript.Run(ctx, s.client,
		[]string{controlKey(e), startedKey(e), eventResultKey(e), queueOrderKey(e), streamIndexKey(e), controlKey(e) + ":cutoff"},
		e.AcceptedOrder).StringSlice()
	if err != nil {
		return err
	}
	switch result[0] {
	case "discarded":
		return ErrPromptDiscarded
	case "busy":
		return errTurnInProgress
	case "consumed":
		return &consumedTurn{invocation: result[1], state: RunState(result[2])}
	}
	selected := []*Event{}
	// The index contains only references into the existing Stream, no payload.
	orders, err := s.client.ZRangeByScore(ctx, queueOrderKey(e), &redis.ZRangeBy{
		Min: strconv.FormatInt(e.AcceptedOrder, 10), Max: "+inf", Count: maxBatchPrompts,
	}).Result()
	if err != nil {
		return err
	}
	if len(orders) == 0 || orders[0] != strconv.FormatInt(e.AcceptedOrder, 10) {
		// Legacy entries have no admission index. They remain single turns.
		copy := *e
		copy.Members = nil
		e.Members = []*Event{&copy}
		return nil
	}
	for i, order := range orders {
		var candidate *Event
		if i == 0 {
			copy := *e
			copy.Members = nil
			candidate = &copy
		} else {
			id, err := s.client.HGet(ctx, streamIndexKey(e), order).Result()
			if errors.Is(err, redis.Nil) {
				break
			} // pre-upgrade singleton
			if err != nil {
				return err
			}
			rows, err := s.client.XRangeN(ctx, StreamKey, id, id, 1).Result()
			if err != nil {
				return err
			}
			if len(rows) != 1 {
				return errors.New("plane pending Stream payload is missing")
			}
			raw, ok := rows[0].Values["event"].(string)
			if !ok {
				return ErrMalformed
			}
			candidate, err = DecodeEvent(raw)
			if err != nil {
				return err
			}
		}
		if candidate.Stop || planeRunKey(candidate) != planeRunKey(e) || candidate.AgentID != e.AgentID {
			break
		}
		trial := append(append([]*Event{}, selected...), candidate)
		_, size, _, err := batchInput(trial)
		if err != nil {
			if i == 0 {
				return err
			}
			break
		}
		if size > batchTextBudget && len(selected) > 0 {
			break
		}
		selected = trial
	}
	e.Members = selected
	return nil
}

func membersOf(e *Event) []*Event {
	if len(e.Members) != 0 {
		return e.Members
	}
	return []*Event{e}
}

func batchInput(events []*Event) ([]*genai.Part, int, map[string]string, error) {
	var turn userinput.Turn
	total := 0
	actors := map[string]bool{}
	provenance := []map[string]any{}
	for _, e := range events {
		if strings.TrimSpace(e.Prompt) == "" {
			return nil, 0, nil, ErrPromptRequired
		}
		// Identity is server-supplied metadata, separate from the user's text.
		header, _ := json.Marshal(map[string]string{"plane_actor": e.ActorUserID, "activity_id": e.ActivityID})
		if len(events) > 1 {
			if err := turn.AddText(string(header)); err != nil {
				return nil, 0, nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
			}
			total += len(header)
		}
		if err := turn.AddText(e.Prompt); err != nil {
			return nil, 0, nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
		}
		total += len(e.Prompt)
		actors[e.ActorUserID] = true
		provenance = append(provenance, map[string]any{"actor": e.ActorUserID, "activity": e.ActivityID, "order": e.AcceptedOrder})
	}
	raw, _ := json.Marshal(provenance)
	metadata := map[string]string{"plane_prompts": string(raw)}
	if len(actors) == 1 && !actors[""] {
		for actor := range actors {
			metadata[memoryhook.PrincipalMetadataKey] = "plane:" + actor
		}
	} else {
		metadata[memoryhook.DisableCaptureMetadataKey] = "true"
		metadata["memory_capture_disabled_reason"] = "mixed or unknown Plane actors; single-principal capture cannot represent this batch"
	}
	return turn.Parts(), total, metadata, nil
}

func planeSession(e *Event, agentID string) string {
	// Length-delimited identity avoids collisions between external identifiers.
	return fmt.Sprintf("plane:%d:%s:%d:%s:%d:%s:%d:%s:%d:%s",
		len(e.WorkspaceID), e.WorkspaceID, len(e.AppID), e.AppID,
		len(e.InstallationID), e.InstallationID, len(e.AgentRunID), e.AgentRunID, len(agentID), agentID)
}

// Retire only a never-started rejected event, so it cannot block FIFO forever.
var retireBatchEventScript = redis.NewScript(`
if redis.call('EXISTS', KEYS[1]) == 1 then return 0 end
local phase = redis.call('HGET', KEYS[4], 'phase')
if phase == 'reserved' then return 0 end
local types = {'string', 'zset', 'hash', 'hash'}
for i = 2, #KEYS do
  local kind = redis.call('TYPE', KEYS[i]).ok
  if kind ~= 'none' and kind ~= types[i] then return redis.error_reply('invalid Plane rejection index') end
end
if not redis.acl_check_cmd('ZREM', KEYS[2], ARGV[1]) or
   not redis.acl_check_cmd('HDEL', KEYS[3], ARGV[1]) then return redis.error_reply('Plane rejection permission missing') end
redis.call('ZREM', KEYS[2], ARGV[1])
redis.call('HDEL', KEYS[3], ARGV[1])
return 1
`)

func (s *redisBatchSource) retire(ctx context.Context, e *Event) error {
	n, err := retireBatchEventScript.Run(ctx, s.client, []string{startedKey(e), queueOrderKey(e), streamIndexKey(e), controlKey(e)}, e.AcceptedOrder).Int()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrRunBusy
	}
	return nil
}

func planeServiceUser(e *Event) string {
	return fmt.Sprintf("plane:%d:%s:%d:%s:%d:%s", len(e.WorkspaceID), e.WorkspaceID, len(e.AppID), e.AppID, len(e.InstallationID), e.InstallationID)
}
