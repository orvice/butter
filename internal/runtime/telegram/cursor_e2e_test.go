package telegram

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	"connectrpc.com/connect"
	adkrunner "google.golang.org/adk/v2/runner"
	adksession "google.golang.org/adk/v2/session"
	"google.golang.org/protobuf/proto"

	"go.orx.me/apps/butter/internal/runtime/cursorbox"
	"go.orx.me/apps/butter/internal/runtime/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/butterbox/cursor/v1"
	"go.orx.me/apps/butter/pkg/proto/butterbox/cursor/v1/cursorv1connect"
)

// telegramCursorClient is a typed fake of the box's CursorService, recording
// the requests so the test can assert the Cursor session is created once and
// reused across topic turns.
type telegramCursorClient struct {
	cursorv1connect.UnimplementedCursorServiceHandler

	mu      sync.Mutex
	creates []*cursorv1.CreateSessionRequest
	sends   []*cursorv1.SendMessageRequest
}

func (c *telegramCursorClient) CreateSession(_ context.Context, req *connect.Request[cursorv1.CreateSessionRequest]) (*connect.Response[cursorv1.CreateSessionResponse], error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.creates = append(c.creates, proto.Clone(req.Msg).(*cursorv1.CreateSessionRequest))
	return connect.NewResponse(&cursorv1.CreateSessionResponse{SessionId: "cursor-session-1"}), nil
}

func (c *telegramCursorClient) SendMessage(_ context.Context, req *connect.Request[cursorv1.SendMessageRequest]) (*connect.Response[cursorv1.SendMessageResponse], error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sends = append(c.sends, proto.Clone(req.Msg).(*cursorv1.SendMessageRequest))
	return connect.NewResponse(&cursorv1.SendMessageResponse{
		Text: fmt.Sprintf("cursor reply %d", len(c.sends)),
	}), nil
}

func (c *telegramCursorClient) snapshot() ([]*cursorv1.CreateSessionRequest, []*cursorv1.SendMessageRequest) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]*cursorv1.CreateSessionRequest(nil), c.creates...),
		append([]*cursorv1.SendMessageRequest(nil), c.sends...)
}

type telegramCursorFactory struct{ client *telegramCursorClient }

// ClientFor mirrors the production factory: an unknown box fails resolution
// before any RPC is attempted.
func (f telegramCursorFactory) ClientFor(_ context.Context, _ string, butterboxID string) (cursorv1connect.CursorServiceClient, error) {
	if butterboxID != "box-1" {
		return nil, fmt.Errorf("cursorbox: butterbox %q no longer exists in this workspace; point the agent at a registered ButterBox", butterboxID)
	}
	return f.client, nil
}

// This crosses the whole in-process exit path: a topic update is admitted,
// the real runner and cursorbox bridge call CursorService, and the answer
// edits the processing message in the originating topic. Two turns share one
// Cursor session, proving multi-turn context continuity survives the full
// Telegram surface.
func TestCursorAgentAnswersInTelegramTopicWithContinuity(t *testing.T) {
	cursorClient := &telegramCursorClient{}
	sessions := adksession.InMemoryService()
	agents, err := runner.NewServiceWithMCPHTTPClientFactory(
		t.Context(), []agentsv1.Agent{{
			Name:        "cursor-coder",
			AgentId:     "cursor-coder",
			WorkspaceId: "ws-a",
			Type:        agentsv1.AgentType_AGENT_TYPE_CURSOR,
			Config: &agentsv1.AgentConfig{Cursor: &agentsv1.CursorAgentConfig{
				ButterboxId: "box-1",
				WorkingDir:  "projects/demo",
				Model:       "composer-2.5",
				Mode:        "agent",
			}},
		}}, nil, nil, nil, nil,
		sessions, nil, nil, nil, 0, nil, adkrunner.PluginConfig{}, nil,
		nil, cursorbox.AgentBuilder(telegramCursorFactory{client: cursorClient}),
	)
	if err != nil {
		t.Fatalf("build runner: %v", err)
	}

	fx := newOrchestratorFixture(t, func(config *agentsv1.TelegramDestinationConfig) {
		config.AgentId = "cursor-coder"
	})
	fx.orchestrator.runner = agents

	if err := fx.orchestrator.Handle(t.Context(), fx.eventForStored(message(realUser, "first turn", ""))); err != nil {
		t.Fatalf("first turn: %v", err)
	}
	if err := fx.orchestrator.Handle(t.Context(), fx.eventForStored(message(realUser, "second turn", ""))); err != nil {
		t.Fatalf("second turn: %v", err)
	}

	creates, sends := cursorClient.snapshot()
	if len(creates) != 1 {
		t.Fatalf("cursor sessions created = %d, want one shared across topic turns", len(creates))
	}
	if len(sends) != 2 {
		t.Fatalf("cursor sends = %d, want two", len(sends))
	}
	if sends[0].GetSessionId() != sends[1].GetSessionId() {
		t.Fatalf("cursor session changed between turns: %q != %q", sends[0].GetSessionId(), sends[1].GetSessionId())
	}
	if got := sends[0].GetMessage(); got != "first turn" {
		t.Fatalf("first message: got %q", got)
	}
	if creates[0].GetCwd() != "projects/demo" || creates[0].GetModel() != "composer-2.5" || creates[0].GetMode() != "agent" {
		t.Fatalf("create request lost config: %+v", creates[0])
	}

	sent := fx.bots.Sent()
	if len(sent) != 4 {
		t.Fatalf("telegram deliveries = %d, want two placeholders and two edits", len(sent))
	}
	for index, want := range []string{"cursor reply 1", "cursor reply 2"} {
		delivery := sent[index*2+1]
		if delivery.Edit == nil || !strings.Contains(delivery.Edit.Text, want) {
			got := ""
			if delivery.Edit != nil {
				got = delivery.Edit.Text
			}
			t.Fatalf("reply %d text = %q, want it to contain %q", index+1, got, want)
		}
	}
}

// A Cursor agent bound to an unknown box must fail honestly on the Telegram
// surface: the orchestrator returns the detailed resolution error (so the
// queue records it) and the topic gets the failure notice instead of silence.
func TestCursorAgentTelegramErrorSurfacesInTopic(t *testing.T) {
	failing := &telegramCursorClient{}
	sessions := adksession.InMemoryService()
	agents, err := runner.NewServiceWithMCPHTTPClientFactory(
		t.Context(), []agentsv1.Agent{{
			Name:        "cursor-coder",
			AgentId:     "cursor-coder",
			WorkspaceId: "ws-a",
			Type:        agentsv1.AgentType_AGENT_TYPE_CURSOR,
			Config: &agentsv1.AgentConfig{Cursor: &agentsv1.CursorAgentConfig{
				ButterboxId: "missing-box",
				WorkingDir:  "projects/demo",
			}},
		}}, nil, nil, nil, nil,
		sessions, nil, nil, nil, 0, nil, adkrunner.PluginConfig{}, nil,
		nil, cursorbox.AgentBuilder(telegramCursorFactory{client: failing}),
	)
	if err != nil {
		t.Fatalf("build runner: %v", err)
	}

	fx := newOrchestratorFixture(t, func(config *agentsv1.TelegramDestinationConfig) {
		config.AgentId = "cursor-coder"
	})
	fx.orchestrator.runner = agents

	err = fx.orchestrator.Handle(t.Context(), fx.eventForStored(message(realUser, "hi", "")))
	if err == nil || !strings.Contains(err.Error(), "ButterBox") {
		t.Fatalf("expected a ButterBox resolution error, got %v", err)
	}
	sent := fx.bots.Sent()
	if len(sent) == 0 || sent[len(sent)-1].Edit == nil {
		t.Fatal("expected the failure notice to be delivered in the topic")
	}
	if !strings.Contains(sent[len(sent)-1].Edit.Text, "could not complete") {
		t.Fatalf("failure notice text = %q", sent[len(sent)-1].Edit.Text)
	}
}
