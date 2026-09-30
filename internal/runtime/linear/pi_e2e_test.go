package linear

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"connectrpc.com/connect"
	piv1 "github.com/orvice/butter-box/pkg/proto/butterbox/pi/v1"
	"github.com/orvice/butter-box/pkg/proto/butterbox/pi/v1/piv1connect"
	adkrunner "google.golang.org/adk/v2/runner"
	adksession "google.golang.org/adk/v2/session"
	"google.golang.org/protobuf/proto"

	internalagent "go.orx.me/apps/butter/internal/agent"
	"go.orx.me/apps/butter/internal/runtime/pibox"
	"go.orx.me/apps/butter/internal/runtime/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// linearPiClient is a typed fake of the ButterBox PiService.
type linearPiClient struct {
	piv1connect.PiServiceClient

	mu      sync.Mutex
	creates []*piv1.CreateSessionRequest
	submits []*piv1.SubmitMessageRequest
	aborts  []*piv1.AbortSessionRequest
	// hold, when set, keeps GetTurn from settling until it is closed or
	// the call is cancelled.
	hold chan struct{}
	// submitted is signalled on every SubmitMessage.
	submitted chan struct{}
}

func (c *linearPiClient) CreateSession(_ context.Context, req *connect.Request[piv1.CreateSessionRequest]) (*connect.Response[piv1.CreateSessionResponse], error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.creates = append(c.creates, proto.Clone(req.Msg).(*piv1.CreateSessionRequest))
	return connect.NewResponse(&piv1.CreateSessionResponse{
		Session: &piv1.Session{Id: fmt.Sprintf("pi-session-%d", len(c.creates)), Cwd: req.Msg.GetCwd()},
	}), nil
}

func (c *linearPiClient) SubmitMessage(_ context.Context, req *connect.Request[piv1.SubmitMessageRequest]) (*connect.Response[piv1.SubmitMessageResponse], error) {
	c.mu.Lock()
	c.submits = append(c.submits, proto.Clone(req.Msg).(*piv1.SubmitMessageRequest))
	n, submitted := len(c.submits), c.submitted
	c.mu.Unlock()
	if submitted != nil {
		submitted <- struct{}{}
	}
	return connect.NewResponse(&piv1.SubmitMessageResponse{TurnCursor: fmt.Sprintf("cursor-%d", n)}), nil
}

func (c *linearPiClient) GetTurn(ctx context.Context, _ *connect.Request[piv1.GetTurnRequest]) (*connect.Response[piv1.GetTurnResponse], error) {
	c.mu.Lock()
	hold, n := c.hold, len(c.submits)
	c.mu.Unlock()
	if hold != nil {
		select {
		case <-hold:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return connect.NewResponse(&piv1.GetTurnResponse{
		Result: &piv1.TurnResult{Text: fmt.Sprintf("pi reply %d", n), StopReason: "stop"},
	}), nil
}

func (c *linearPiClient) AbortSession(_ context.Context, req *connect.Request[piv1.AbortSessionRequest]) (*connect.Response[piv1.AbortSessionResponse], error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.aborts = append(c.aborts, proto.Clone(req.Msg).(*piv1.AbortSessionRequest))
	return connect.NewResponse(&piv1.AbortSessionResponse{}), nil
}

func (c *linearPiClient) snapshot() (creates []*piv1.CreateSessionRequest, submits []*piv1.SubmitMessageRequest, aborts []*piv1.AbortSessionRequest) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append(creates, c.creates...), append(submits, c.submits...), append(aborts, c.aborts...)
}

type linearPiFactory struct{ client *linearPiClient }

func (f linearPiFactory) ClientFor(context.Context, string, string) (piv1connect.PiServiceClient, error) {
	return f.client, nil
}

// piRunner builds the real runner with one Pi Agent backed by client.
func piRunner(t *testing.T, client *linearPiClient) *runner.Service {
	t.Helper()
	agents, err := runner.NewServiceWithMCPHTTPClientFactory(
		t.Context(), []agentsv1.Agent{{
			Name:        "pi-coder",
			AgentId:     "support",
			WorkspaceId: "ws-a",
			Type:        agentsv1.AgentType_AGENT_TYPE_PI,
			Config: &agentsv1.AgentConfig{Pi: &agentsv1.PiAgentConfig{
				ButterboxId: "box-1",
				WorkingDir:  "projects/demo",
			}},
		}}, nil, nil, nil, nil,
		adksession.InMemoryService(), nil, nil, nil, 0, nil, adkrunner.PluginConfig{}, nil,
		&internalagent.BoxAgentBuilders{Pi: pibox.AgentBuilder(linearPiFactory{client: client})},
	)
	if err != nil {
		t.Fatalf("build runner: %v", err)
	}
	return agents
}

// This crosses the whole in-process path for a Pi Agent: two events on one
// Linear Agent Session reach the real runner and pibox bridge, and both
// turns run in the same box session.
func TestAPiAgentKeepsOneBoxSessionAcrossLinearTurns(t *testing.T) {
	fx := newOrchestratorFixture(t, nil)
	piClient := &linearPiClient{}
	fx.orch.runner = piRunner(t, piClient)

	fx.handle(t, fx.event(ActionCreated))
	fx.handle(t, fx.prompted("continue"))

	creates, submits, _ := piClient.snapshot()
	if len(creates) != 1 {
		t.Fatalf("pi sessions created = %d, want one shared across the Linear turns", len(creates))
	}
	if len(submits) != 2 || submits[0].GetSessionId() != submits[1].GetSessionId() {
		t.Fatalf("pi submissions = %+v; want two in one box session", submits)
	}
	acts := fx.linear.Activities()
	if last := acts[len(acts)-1]; last.Type != "response" || last.Body != "pi reply 2" {
		t.Fatalf("last activity = %+v; want the second pi reply", last)
	}
}
