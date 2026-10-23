package pibox

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"time"

	"connectrpc.com/connect"
	piv1 "github.com/orvice/butter-box/pkg/proto/butterbox/pi/v1"
	"github.com/orvice/butter-box/pkg/proto/butterbox/pi/v1/piv1connect"
	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/a2ui"
	"go.orx.me/apps/butter/internal/a2uitool"
	"go.orx.me/apps/butter/internal/runtime/eventcommit"
)

// Pinned to butter-box#30's bundled render_ui.json. Never echo an unknown
// server hash: a changed tool contract needs an explicit compatibility update.
const hostProtocolVersion = 1
const hostContractSHA256 = "8adfef21935c063a895263b1beb91e67b74472ea0088a98377eacd858f9f491c"
const hostMaxPayloadBytes = 256 << 10
const hostMaxCalls = 64

func (b *Bridge) hostToolsConfig(ctx context.Context, client piv1connect.PiServiceClient, sessionID string, ictx agent.InvocationContext) (*piv1.HostToolsConfig, error) {
	if _, ok := a2ui.RunFrom(ictx); !ok || !b.cardPolicy.Allowed() || !eventcommit.Available(ictx) {
		return nil, nil
	}
	// v1 grants cannot renew or be unlimited. Do not silently shorten a run
	// configured as unlimited (or longer than Box supports); keep it text-only.
	deadline, bounded := ctx.Deadline()
	if !bounded {
		return nil, nil
	}
	remaining := time.Until(deadline)
	if remaining <= 0 {
		return nil, ctx.Err()
	}
	seconds := uint32((remaining + time.Second - 1) / time.Second)
	capCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	resp, err := client.GetCapabilities(capCtx, connect.NewRequest(&piv1.GetCapabilitiesRequest{SessionId: sessionID}))
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if connect.CodeOf(err) == connect.CodeUnauthenticated || connect.CodeOf(err) == connect.CodePermissionDenied {
			return nil, b.actionable("check host tool capabilities", err)
		}
		return nil, nil // Legacy Box / unavailable handshake: ordinary text run.
	}
	caps := resp.Msg
	if !caps.GetHostToolsExtensionReady() || caps.GetHostToolsProtocolVersion() != hostProtocolVersion || caps.GetHostToolsContractSha256() != hostContractSHA256 || caps.GetHostToolsMaxRunSeconds() < seconds || caps.GetHostToolsMaxPayloadBytes() < hostMaxPayloadBytes || caps.GetHostToolsMaxCallsPerTurn() == 0 {
		return nil, nil
	}
	return &piv1.HostToolsConfig{
		ProtocolVersion: hostProtocolVersion, ContractSha256: hostContractSHA256,
		AllowedTools: []string{a2uitool.ToolName}, MaxRunSeconds: seconds,
	}, nil
}

type hostReceipt struct {
	digest  [32]byte
	request *piv1.RespondHostCallRequest
}

// awaitHost uses ListHostCalls for the wait, and GetTurn only as a pure poll.
// Both a pending tool and grant completion wake ListHostCalls; waiting only on
// GetTurn would deadlock a Pi model that is awaiting its render_ui result.
func (b *Bridge) awaitHost(ctx context.Context, client piv1connect.PiServiceClient, sessionID string, submitted submission, ictx agent.InvocationContext, yield func(*session.Event, error) bool) (*piv1.TurnResult, error) {
	receipts := make(map[string]hostReceipt)
	for {
		turn, err := b.pollTurn(ctx, client, sessionID, submitted.cursor, 0)
		if err != nil {
			return nil, err
		}
		if !turn.GetRunning() {
			return completedTurn(turn)
		}
		calls, err := b.listHostCalls(ctx, client, sessionID, submitted.turnID)
		if err != nil {
			return nil, err
		}
		for _, call := range calls {
			if call.GetSessionId() != sessionID || call.GetTurnId() != submitted.turnID || call.GetCallId() == "" || len(call.GetCallId()) > 256 || call.GetTool() != a2uitool.ToolName || len(call.GetArgumentsJson()) > hostMaxPayloadBytes {
				return nil, errors.New("pibox: invalid or misaddressed host tool request")
			}
			digest := sha256.Sum256([]byte(call.GetArgumentsJson()))
			receipt, duplicate := receipts[call.GetCallId()]
			if duplicate && receipt.digest != digest {
				return nil, errors.New("pibox: host tool call id reused with different arguments")
			}
			if !duplicate {
				if len(receipts) >= hostMaxCalls {
					return nil, errors.New("pibox: host tool call limit exceeded")
				}
				callEvent, event, request := b.prepareHostCall(ictx, call)
				if !yield(callEvent, nil) {
					return nil, errors.New("pibox: run stopped before host tool call persistence")
				}
				wait, discard := eventcommit.Expect(ictx, event.ID)
				if !yield(event, nil) {
					discard()
					return nil, errors.New("pibox: run stopped before host tool persistence")
				}
				if err := wait(ctx); err != nil {
					return nil, fmt.Errorf("pibox: persist host tool result: %w", err)
				}
				receipt = hostReceipt{digest: digest, request: request}
				receipts[call.GetCallId()] = receipt
			}
			// A failed/uncertain receipt retries these identical bytes only,
			// never validates again or creates another card/revision.
			if err := b.respondHostCall(ctx, client, receipt.request); err != nil {
				return nil, err
			}
		}
	}
}

func retryableHostError(err error) bool {
	switch connect.CodeOf(err) {
	case connect.CodeUnavailable, connect.CodeDeadlineExceeded, connect.CodeResourceExhausted, connect.CodeUnknown:
		return true
	default:
		return false
	}
}

func (b *Bridge) listHostCalls(ctx context.Context, client piv1connect.PiServiceClient, sessionID, turnID string) ([]*piv1.HostCall, error) {
	var err error
	for attempt := range maxConsecutivePollFailures {
		callCtx, cancel := context.WithTimeout(ctx, time.Duration(b.pollWaitSeconds)*time.Second+15*time.Second)
		var resp *connect.Response[piv1.ListHostCallsResponse]
		resp, err = client.ListHostCalls(callCtx, connect.NewRequest(&piv1.ListHostCallsRequest{SessionId: sessionID, TurnId: turnID, WaitSeconds: b.pollWaitSeconds}))
		cancel()
		if err == nil {
			return resp.Msg.GetCalls(), nil
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if !retryableHostError(err) {
			break
		}
		if attempt+1 < maxConsecutivePollFailures && !b.retry(ctx) {
			return nil, ctx.Err()
		}
	}
	return nil, fmt.Errorf("pibox: read host tool requests: %w", err)
}

func (b *Bridge) respondHostCall(ctx context.Context, client piv1connect.PiServiceClient, request *piv1.RespondHostCallRequest) error {
	var err error
	for attempt := range maxConsecutivePollFailures {
		callCtx, cancel := context.WithTimeout(ctx, controlCallTimeout)
		_, err = client.RespondHostCall(callCtx, connect.NewRequest(request))
		cancel()
		if err == nil {
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if !retryableHostError(err) {
			break
		}
		if attempt+1 < maxConsecutivePollFailures && !b.retry(ctx) {
			return ctx.Err()
		}
	}
	return fmt.Errorf("pibox: deliver host tool receipt: %w", err)
}

// Both successful and failed calls are recorded with scoped IDs and their
// exact receipt. The non-partial function response reaches the existing sink,
// avoiding its final-text filter. No raw Pi transcript is exposed to clients.
func (b *Bridge) prepareHostCall(ictx agent.InvocationContext, call *piv1.HostCall) (*session.Event, *session.Event, *piv1.RespondHostCallRequest) {
	event := session.NewEvent(ictx, ictx.InvocationID())
	event.Author = ictx.Agent().Name()
	digest := sha256.Sum256([]byte(call.GetSessionId() + "\x00" + call.GetTurnId() + "\x00" + call.GetCallId()))
	id := "pibox-" + hex.EncodeToString(digest[:])
	request := &piv1.RespondHostCallRequest{SessionId: call.GetSessionId(), TurnId: call.GetTurnId(), CallId: call.GetCallId()}
	var args a2uitool.Args
	var result a2uitool.Result
	value := jsontext.Value(call.GetArgumentsJson())
	err := json.Unmarshal(value, &args, json.RejectUnknownMembers(true))
	if err == nil && value.Kind() != '{' {
		err = errors.New("render_ui arguments must be a JSON object")
	}
	if err != nil {
		err = a2uitool.RecordFailure(ictx, fmt.Errorf("invalid render_ui arguments: %w", err))
	}
	if err == nil {
		result, err = a2uitool.Prepare(ictx, b.cardPolicy, ictx.Session().State(), ictx.InvocationID(), args, func(card *a2ui.Card) error {
			event.Actions.StateDelta[a2ui.CardKey(card.ID)] = card.StateValue()
			return nil
		})
	}
	response := make(map[string]any)
	if err != nil {
		message := err.Error()
		// Box v1 caps an ordinary tool error at 8 KiB. Keep generated errors
		// bounded even when a malicious property name was included in one.
		if len(message) > 4000 {
			message = "render_ui arguments are invalid; answer in text or use the supported catalog"
		}
		response["error"] = message
		request.Outcome = &piv1.RespondHostCallRequest_Error{Error: message}
	} else {
		response = map[string]any{"surface_id": result.SurfaceID, "revision": result.Revision, "status": result.Status}
		encoded, _ := json.Marshal(result) // Fixed result shape: strings and an integer.
		request.Outcome = &piv1.RespondHostCallRequest_ResultJson{ResultJson: string(encoded)}
	}
	callEvent := session.NewEvent(ictx, ictx.InvocationID())
	callEvent.Author = ictx.Agent().Name()
	callEvent.Content = &genai.Content{Role: genai.RoleModel, Parts: []*genai.Part{
		{FunctionCall: &genai.FunctionCall{ID: id, Name: a2uitool.ToolName, Args: map[string]any{
			"surface_id": args.SurfaceID, "messages": args.Messages, "fallback": args.Fallback,
		}}},
	}}
	// Keep model calls and user-role tool results in separate events, as ADK
	// does, so a delegating LLM can later consume this history correctly.
	event.Content = &genai.Content{Role: genai.RoleUser, Parts: []*genai.Part{
		{FunctionResponse: &genai.FunctionResponse{ID: id, Name: a2uitool.ToolName, Response: response}},
	}}
	return callEvent, event, request
}
