package http

import (
	"context"
	"encoding/json/v2"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"connectrpc.com/connect"
	piv1 "github.com/orvice/butter-box/pkg/proto/butterbox/pi/v1"
	"github.com/orvice/butter-box/pkg/proto/butterbox/pi/v1/piv1connect"
	adksession "google.golang.org/adk/v2/session"

	"go.orx.me/apps/butter/internal/a2ui"
	internalagent "go.orx.me/apps/butter/internal/agent"
	"go.orx.me/apps/butter/internal/runtime/pibox"
	"go.orx.me/apps/butter/internal/testsupport/openaifake"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type piCardBox struct {
	piv1connect.UnimplementedPiServiceHandler
	store    adksession.Service
	answered atomic.Bool
}

func (*piCardBox) CreateSession(context.Context, *connect.Request[piv1.CreateSessionRequest]) (*connect.Response[piv1.CreateSessionResponse], error) {
	return connect.NewResponse(&piv1.CreateSessionResponse{Session: &piv1.Session{Id: "pi-session"}}), nil
}
func (*piCardBox) GetCapabilities(context.Context, *connect.Request[piv1.GetCapabilitiesRequest]) (*connect.Response[piv1.GetCapabilitiesResponse], error) {
	return connect.NewResponse(&piv1.GetCapabilitiesResponse{HostToolsProtocolVersion: 1, HostToolsContractSha256: "8adfef21935c063a895263b1beb91e67b74472ea0088a98377eacd858f9f491c", HostToolsExtensionReady: true, HostToolsMaxPayloadBytes: 256 << 10, HostToolsMaxCallsPerTurn: 64, HostToolsMaxRunSeconds: 86400}), nil
}
func (*piCardBox) SubmitMessage(_ context.Context, req *connect.Request[piv1.SubmitMessageRequest]) (*connect.Response[piv1.SubmitMessageResponse], error) {
	if req.Msg.GetHostTools() == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("host tools were not enabled"))
	}
	return connect.NewResponse(&piv1.SubmitMessageResponse{TurnCursor: "cursor", TurnId: "pi-turn"}), nil
}
func (f *piCardBox) GetTurn(context.Context, *connect.Request[piv1.GetTurnRequest]) (*connect.Response[piv1.GetTurnResponse], error) {
	if !f.answered.Load() {
		return connect.NewResponse(&piv1.GetTurnResponse{Running: true}), nil
	}
	return connect.NewResponse(&piv1.GetTurnResponse{Result: &piv1.TurnResult{Text: "Deployment healthy", StopReason: "stop"}}), nil
}
func (*piCardBox) ListHostCalls(context.Context, *connect.Request[piv1.ListHostCallsRequest]) (*connect.Response[piv1.ListHostCallsResponse], error) {
	return connect.NewResponse(&piv1.ListHostCallsResponse{Running: true, Calls: []*piv1.HostCall{{SessionId: "pi-session", TurnId: "pi-turn", CallId: "render-1", Tool: "render_ui", ArgumentsJson: `{"messages":[{"updateComponents":{"components":[{"id":"root","component":"Text","text":"Healthy"}]}}],"fallback":"Healthy"}`}}}), nil
}
func (f *piCardBox) RespondHostCall(ctx context.Context, req *connect.Request[piv1.RespondHostCallRequest]) (*connect.Response[piv1.RespondHostCallResponse], error) {
	sess, err := f.store.Get(ctx, &adksession.GetRequest{AppName: aguiAppName, UserID: "u1", SessionID: aguiSessionPrefix + "pi-cards"})
	if err != nil {
		return nil, err
	}
	if req.Msg.GetResultJson() == "" || len(a2ui.LiveCards(sess.Session.State())) != 1 {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("receipt before persistence"))
	}
	f.answered.Store(true)
	return connect.NewResponse(&piv1.RespondHostCallResponse{}), nil
}

type piCardFactory struct{ url string }

func (f piCardFactory) ClientFor(context.Context, string, string) (piv1connect.PiServiceClient, error) {
	return piv1connect.NewPiServiceClient(http.DefaultClient, f.url), nil
}

func TestAGUIPiCardsStreamAndRestoreThroughExistingA2UIContract(t *testing.T) {
	h := &a2uiHarness{t: t, backend: openaifake.New(t), sessions: newStoreLikeSessions(), guard: &fakeSessionGuard{}}
	box := &piCardBox{store: h.sessions}
	path, handler := piv1connect.NewPiServiceHandler(box)
	mux := http.NewServeMux()
	mux.Handle(path, handler)
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	h.boxBuilders = &internalagent.BoxAgentBuilders{Pi: pibox.AgentBuilder(piCardFactory{server.URL})}
	agents := []agentsv1.Agent{{AgentId: "pi", Name: "Pi", WorkspaceId: "ws-a", Type: agentsv1.AgentType_AGENT_TYPE_PI, Config: &agentsv1.AgentConfig{Pi: &agentsv1.PiAgentConfig{ButterboxId: "box"}}}}
	h.router = h.build(agents, nil)
	w := h.post("pi", a2uiBody("pi-cards", "show deployment"))
	if w.Code != http.StatusOK {
		t.Fatalf("POST=%d %s", w.Code, w.Body.String())
	}
	events := sseEvents(t, w.Body.String())
	values := a2uiValues(events)
	if len(values) == 0 || !box.answered.Load() {
		t.Fatalf("no durable Pi card: %s", w.Body.String())
	}
	customIndex := eventIndex(events, func(e map[string]any) bool { return e["type"] == "CUSTOM" && e["name"] == "butter.a2ui" })
	textIndex := eventIndex(events, func(e map[string]any) bool { return e["type"] == "TEXT_MESSAGE_CONTENT" })
	if customIndex < 0 || textIndex < customIndex {
		t.Fatal("card did not stream before the final reply")
	}
	surfaces := h.snapshotSurfaces("pi", "pi-cards")
	if len(surfaces) != 1 || surfaces[0]["surfaceId"] != values[0]["surfaceId"] {
		t.Fatalf("snapshot=%v stream=%v", surfaces, values)
	}
	// A new handler/runner reads the same persisted UI without invoking Pi.
	h.restart(agents)
	if restored := h.snapshotSurfaces("pi", "pi-cards"); len(restored) != 1 || restored[0]["surfaceId"] != surfaces[0]["surfaceId"] {
		t.Fatalf("restored=%v", restored)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/agui/pi/threads/pi-cards/messages", nil)
	history := httptest.NewRecorder()
	h.router.ServeHTTP(history, req)
	var body struct {
		Surfaces []struct {
			SurfaceID string `json:"surfaceId"`
			MessageID string `json:"messageId"`
		} `json:"surfaces"`
	}
	if err := json.Unmarshal(history.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if history.Code != http.StatusOK || len(body.Surfaces) != 1 || body.Surfaces[0].MessageID == "" {
		t.Fatalf("history=%s", history.Body.String())
	}
}
