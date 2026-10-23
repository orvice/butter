package pibox

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	piv1 "github.com/orvice/butter-box/pkg/proto/butterbox/pi/v1"
	"github.com/orvice/butter-box/pkg/proto/butterbox/pi/v1/piv1connect"
	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/workflowagents/parallelagent"
	adkrunner "google.golang.org/adk/v2/runner"
	"google.golang.org/adk/v2/session"
	"google.golang.org/protobuf/proto"

	"go.orx.me/apps/butter/internal/a2ui"
	"go.orx.me/apps/butter/internal/runtime/eventcommit"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func cardsDisabledForTest() *agentsv1.ResultCardConfig {
	return &agentsv1.ResultCardConfig{Generation: agentsv1.ResultCardGeneration_RESULT_CARD_GENERATION_DISABLED}
}

const createCardArgs = `{"messages":[{"updateComponents":{"components":[{"id":"root","component":"Text","text":"Healthy"}]}}],"fallback":"Healthy"}`

// hostPi is the external Box seam; prompts remain running until Butter has
// answered every tool. It does not simulate Butter's own implementation.
type hostPi struct {
	*fakePi
	caps         *piv1.GetCapabilitiesResponse
	capErr       error
	arguments    []string
	turnID       string
	index        int
	active       bool
	receipts     map[string]*piv1.RespondHostCallRequest
	deliveries   []*piv1.RespondHostCallRequest
	lostReply    bool
	checkReceipt func(*piv1.RespondHostCallRequest) error
	badIdentity  bool
	pollErr      error
	blockCalls   bool
}

func newHostPi() *hostPi {
	return &hostPi{fakePi: newFakePi(), caps: &piv1.GetCapabilitiesResponse{
		HostToolsProtocolVersion: 1, HostToolsContractSha256: hostContractSHA256,
		HostToolsExtensionReady: true, HostToolsMaxPayloadBytes: hostMaxPayloadBytes,
		HostToolsMaxCallsPerTurn: 64, HostToolsMaxRunSeconds: 86400,
	}, arguments: []string{createCardArgs}, receipts: make(map[string]*piv1.RespondHostCallRequest)}
}

func (f *hostPi) GetCapabilities(_ context.Context, _ *connect.Request[piv1.GetCapabilitiesRequest]) (*connect.Response[piv1.GetCapabilitiesResponse], error) {
	if f.capErr != nil {
		return nil, f.capErr
	}
	return connect.NewResponse(proto.Clone(f.caps).(*piv1.GetCapabilitiesResponse)), nil
}

func (f *hostPi) SubmitMessage(ctx context.Context, req *connect.Request[piv1.SubmitMessageRequest]) (*connect.Response[piv1.SubmitMessageResponse], error) {
	resp, err := f.fakePi.SubmitMessage(ctx, req)
	if err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.active = req.Msg.GetHostTools() != nil
	if f.active {
		f.turnID = fmt.Sprintf("turn-%d", len(f.submitReqs))
		f.index = 0
		resp.Msg.TurnId = f.turnID
	}
	return resp, nil
}

func (f *hostPi) GetTurn(ctx context.Context, req *connect.Request[piv1.GetTurnRequest]) (*connect.Response[piv1.GetTurnResponse], error) {
	f.mu.Lock()
	pending := f.active && f.index < len(f.arguments)
	f.mu.Unlock()
	if pending {
		return connect.NewResponse(&piv1.GetTurnResponse{Running: true}), nil
	}
	return f.fakePi.GetTurn(ctx, req)
}

func (f *hostPi) ListHostCalls(ctx context.Context, req *connect.Request[piv1.ListHostCallsRequest]) (*connect.Response[piv1.ListHostCallsResponse], error) {
	if f.blockCalls {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.pollErr != nil {
		return nil, f.pollErr
	}
	if f.index >= len(f.arguments) {
		return connect.NewResponse(&piv1.ListHostCallsResponse{}), nil
	}
	id := req.Msg.GetSessionId()
	if f.badIdentity {
		id = "another-session"
	}
	arg := f.arguments[f.index]
	// A later call can use the real surface handle returned by the first.
	for _, receipt := range f.receipts {
		var result map[string]any
		if json.Unmarshal([]byte(receipt.GetResultJson()), &result) == nil {
			if surface, ok := result["surface_id"].(string); ok {
				arg = strings.ReplaceAll(arg, "$SURFACE", surface)
				break
			}
		}
	}
	return connect.NewResponse(&piv1.ListHostCallsResponse{Running: true, Calls: []*piv1.HostCall{{
		SessionId: id, TurnId: f.turnID, CallId: fmt.Sprintf("call-%d", f.index), Tool: "render_ui", ArgumentsJson: arg,
	}}}), nil
}

func (f *hostPi) RespondHostCall(_ context.Context, req *connect.Request[piv1.RespondHostCallRequest]) (*connect.Response[piv1.RespondHostCallResponse], error) {
	if f.checkReceipt != nil {
		if err := f.checkReceipt(req.Msg); err != nil {
			return nil, connect.NewError(connect.CodeFailedPrecondition, err)
		}
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	copy := proto.Clone(req.Msg).(*piv1.RespondHostCallRequest)
	f.deliveries = append(f.deliveries, copy)
	key := copy.GetTurnId() + ":" + copy.GetCallId()
	if prior := f.receipts[key]; prior != nil {
		if !proto.Equal(prior, copy) {
			return nil, connect.NewError(connect.CodeAlreadyExists, errors.New("conflict"))
		}
	} else {
		f.receipts[key] = copy
		f.index++
	}
	if f.lostReply {
		f.lostReply = false
		return nil, connect.NewError(connect.CodeUnavailable, errors.New("reply lost after acceptance"))
	}
	return connect.NewResponse(&piv1.RespondHostCallResponse{}), nil
}

func hostHarness(t *testing.T, f *hostPi) (*Bridge, *harness) {
	t.Helper()
	path, handler := piv1connect.NewPiServiceHandler(f)
	mux := http.NewServeMux()
	mux.Handle(path, handler)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	b := NewBridge(piAgentProto("box-1", ""), staticFactory{srv.URL})
	b.pollWaitSeconds = 0
	b.pollRetryDelay = time.Millisecond
	return b, newHarness(t, b)
}

func uiContext(t *testing.T) context.Context {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	t.Cleanup(cancel)
	return a2ui.WithRun(ctx, &a2ui.Run{ThreadID: "thread-1", RunID: "run-1", MessageID: "reply-1"})
}

func storedSession(t *testing.T, h *harness) session.Session {
	t.Helper()
	r, err := h.sessions.Get(t.Context(), &session.GetRequest{AppName: "test-app", UserID: "u1", SessionID: "s1"})
	if err != nil {
		t.Fatal(err)
	}
	return r.Session
}

func TestHostCardsPersistBeforeReceiptAndCanUpdate(t *testing.T) {
	f := newHostPi()
	f.arguments = append(f.arguments, `{"surface_id":"$SURFACE","messages":[{"updateComponents":{"components":[{"id":"root","component":"Text","text":"Updated"}]}}]}`)
	_, h := hostHarness(t, f)
	f.checkReceipt = func(req *piv1.RespondHostCallRequest) error {
		res, err := h.sessions.Get(t.Context(), &session.GetRequest{AppName: "test-app", UserID: "u1", SessionID: "s1"})
		if err != nil {
			return err
		}
		cards := a2ui.LiveCards(res.Session.State())
		if len(cards) != 1 {
			return errors.New("card was acknowledged before storage")
		}
		var result struct {
			Revision int `json:"revision"`
		}
		if err := json.Unmarshal([]byte(req.GetResultJson()), &result); err != nil {
			return err
		}
		if cards[0].Revision != result.Revision {
			return errors.New("stored revision differs from receipt")
		}
		return nil
	}
	if out, err := h.turn(uiContext(t), textContent("show status")); err != nil || out != "hello from pi" {
		t.Fatalf("turn=(%q,%v)", out, err)
	}
	cards := a2ui.LiveCards(storedSession(t, h).State())
	if len(cards) != 1 || cards[0].Revision != 2 || cards[0].Created.MessageID != "reply-1" {
		t.Fatalf("cards=%+v", cards)
	}
	if f.submitReqs[0].GetHostTools() == nil {
		t.Fatal("host tool was not enabled")
	}
}

func TestHostReceiptLostReplyRetriesWithoutAnotherCard(t *testing.T) {
	f := newHostPi()
	f.lostReply = true
	_, h := hostHarness(t, f)
	if _, err := h.turn(uiContext(t), textContent("show status")); err != nil {
		t.Fatal(err)
	}
	cards := a2ui.LiveCards(storedSession(t, h).State())
	if len(cards) != 1 || cards[0].Revision != 1 || len(f.deliveries) != 2 || !proto.Equal(f.deliveries[0], f.deliveries[1]) {
		t.Fatalf("cards=%d deliveries=%d", len(cards), len(f.deliveries))
	}
}

func TestHostToolErrorReturnsToSameTurnAndCanCorrect(t *testing.T) {
	f := newHostPi()
	f.arguments = []string{`{"messages":[]}`, createCardArgs}
	_, h := hostHarness(t, f)
	if _, err := h.turn(uiContext(t), textContent("show status")); err != nil {
		t.Fatal(err)
	}
	if f.deliveries[0].GetError() == "" || f.deliveries[1].GetResultJson() == "" {
		t.Fatal("expected failed tool result followed by a corrected card")
	}
	if len(a2ui.LiveCards(storedSession(t, h).State())) != 1 {
		t.Fatal("corrected call did not create a card")
	}
}

func TestHostCardsFallbackWithoutCapability(t *testing.T) {
	for _, mode := range []string{"legacy", "not-ready", "hash", "version", "unbounded", "policy", "no-ui"} {
		t.Run(mode, func(t *testing.T) {
			f := newHostPi()
			b, h := hostHarness(t, f)
			ctx := uiContext(t)
			switch mode {
			case "legacy":
				f.capErr = connect.NewError(connect.CodeUnimplemented, errors.New("old Box"))
			case "not-ready":
				f.caps.HostToolsExtensionReady = false
			case "hash":
				f.caps.HostToolsContractSha256 = "unknown"
			case "version":
				f.caps.HostToolsProtocolVersion = 2
			case "unbounded":
				b.maxRun = 0
				ctx = a2ui.WithRun(t.Context(), &a2ui.Run{})
			case "policy":
				b.cardPolicy = a2ui.CardPolicy{}.Narrow(cardsDisabledForTest())
			case "no-ui":
				ctx = t.Context()
			}
			if _, err := h.turn(ctx, textContent("hi")); err != nil {
				t.Fatal(err)
			}
			if f.submitReqs[0].GetHostTools() != nil {
				t.Fatal("unsupported run enabled host tools")
			}
		})
	}
}

func TestHostRejectsMisaddressedRequestAndAborts(t *testing.T) {
	f := newHostPi()
	f.badIdentity = true
	_, h := hostHarness(t, f)
	if _, err := h.turn(uiContext(t), textContent("hi")); err == nil {
		t.Fatal("accepted foreign request")
	}
	if len(f.deliveries) != 0 || f.abortCount() != 1 || len(a2ui.LiveCards(storedSession(t, h).State())) != 0 {
		t.Fatal("foreign request changed state or was not aborted")
	}
}

type failCardStore struct{ session.Service }

func (s failCardStore) AppendEvent(ctx context.Context, sess session.Session, evt *session.Event) error {
	for k := range evt.Actions.StateDelta {
		if _, ok := a2ui.CardIDFromKey(k); ok {
			return errors.New("storage unavailable")
		}
	}
	return s.Service.AppendEvent(ctx, sess, evt)
}

func TestHostDoesNotAcknowledgeFailedPersistenceEvenUnderParallel(t *testing.T) {
	f := newHostPi()
	b, h := hostHarness(t, f)
	leaf, err := b.BuildAgent("pi-coder", "test")
	if err != nil {
		t.Fatal(err)
	}
	root, err := parallelagent.New(parallelagent.Config{AgentConfig: agent.Config{Name: "parallel", SubAgents: []agent.Agent{leaf}}})
	if err != nil {
		t.Fatal(err)
	}
	r, err := adkrunner.New(adkrunner.Config{AppName: "test-app", Agent: root, SessionService: eventcommit.Service(failCardStore{h.sessions})})
	if err != nil {
		t.Fatal(err)
	}
	h.runner = r
	if _, err := h.turn(uiContext(t), textContent("show status")); err == nil {
		t.Fatal("expected persistence failure")
	}
	// The leaf may be unwinding asynchronously under the parallel adapter.
	deadline := time.Now().Add(time.Second)
	for f.abortCount() == 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.deliveries) != 0 || len(f.abortedIDs) != 1 {
		t.Fatalf("receipts=%d aborts=%d", len(f.deliveries), len(f.abortedIDs))
	}
}
