package telegram

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	"connectrpc.com/connect"
	cursorv1 "github.com/orvice/butter-box/pkg/proto/butterbox/cursor/v1"
	"github.com/orvice/butter-box/pkg/proto/butterbox/cursor/v1/cursorv1connect"
	adkrunner "google.golang.org/adk/v2/runner"
	adksession "google.golang.org/adk/v2/session"
	"google.golang.org/protobuf/proto"

	internalagent "go.orx.me/apps/butter/internal/agent"
	"go.orx.me/apps/butter/internal/runtime/cursorbox"
	"go.orx.me/apps/butter/internal/runtime/runner"
	"go.orx.me/apps/butter/internal/telegramapi"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type telegramCursorClient struct {
	cursorv1connect.CursorServiceClient

	mu      sync.Mutex
	creates []*cursorv1.CreateSessionRequest
	sends   []*cursorv1.SendMessageRequest
}

func (c *telegramCursorClient) CreateSession(_ context.Context, req *connect.Request[cursorv1.CreateSessionRequest]) (*connect.Response[cursorv1.CreateSessionResponse], error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.creates = append(c.creates, proto.Clone(req.Msg).(*cursorv1.CreateSessionRequest))
	return connect.NewResponse(&cursorv1.CreateSessionResponse{SessionId: "cursor-agent-1"}), nil
}

func (c *telegramCursorClient) SendMessage(_ context.Context, req *connect.Request[cursorv1.SendMessageRequest]) (*connect.Response[cursorv1.SendMessageResponse], error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sends = append(c.sends, proto.Clone(req.Msg).(*cursorv1.SendMessageRequest))
	return connect.NewResponse(&cursorv1.SendMessageResponse{Text: fmt.Sprintf("cursor reply %d", len(c.sends))}), nil
}

func (c *telegramCursorClient) snapshot() ([]*cursorv1.CreateSessionRequest, []*cursorv1.SendMessageRequest) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]*cursorv1.CreateSessionRequest(nil), c.creates...),
		append([]*cursorv1.SendMessageRequest(nil), c.sends...)
}

type telegramCursorFactory struct{ client *telegramCursorClient }

func (f telegramCursorFactory) ClientFor(context.Context, string, string) (cursorv1connect.CursorServiceClient, error) {
	return f.client, nil
}

// The Cursor twin of the pi exit-path test: a topic update runs through the
// real runner and cursorbox bridge, keeps one Cursor session across turns,
// forwards the photo, and answers in the originating topic. The Destination's
// Butter model override is ignored because a Cursor agent owns its model.
func TestCursorAgentAnswersInTelegramTopicWithContinuityAndPhoto(t *testing.T) {
	cursorClient := &telegramCursorClient{}
	agents, err := runner.NewServiceWithMCPHTTPClientFactory(
		t.Context(), []agentsv1.Agent{{
			Name:        "cursor-coder",
			AgentId:     "cursor-coder",
			WorkspaceId: "ws-a",
			Type:        agentsv1.AgentType_AGENT_TYPE_CURSOR,
			Config: &agentsv1.AgentConfig{Cursor: &agentsv1.CursorAgentConfig{
				ButterboxId: "box-1",
				WorkingDir:  "projects/demo",
			}},
		}}, nil, nil, nil, nil,
		adksession.InMemoryService(), nil, nil, nil, 0, nil, adkrunner.PluginConfig{}, nil,
		&internalagent.BoxAgentBuilders{Cursor: cursorbox.AgentBuilder(telegramCursorFactory{client: cursorClient})},
	)
	if err != nil {
		t.Fatalf("build runner: %v", err)
	}

	fx := newOrchestratorFixture(t, func(config *agentsv1.TelegramDestinationConfig) {
		config.AgentId = "cursor-coder"
		config.Model = "butter-override"
		config.SelectableModels = []string{"butter-override"}
	})
	fx.orchestrator.runner = agents
	fileClient := &fakeFileClient{
		file: telegramapi.File{FileID: "large", FilePath: "photos/file_1.png", FileSize: int64(len(pngBytes))},
		data: pngBytes,
	}
	fx.orchestrator.SetFileClientFactory(func(context.Context, string, string) (telegramapi.FileClient, error) {
		return fileClient, nil
	})

	if err := fx.orchestrator.Handle(t.Context(), fx.eventForStored(photoUpdate("inspect this"))); err != nil {
		t.Fatalf("photo turn: %v", err)
	}
	if err := fx.orchestrator.Handle(t.Context(), fx.eventForStored(message(realUser, "continue", ""))); err != nil {
		t.Fatalf("second turn: %v", err)
	}

	creates, sends := cursorClient.snapshot()
	if len(creates) != 1 || creates[0].GetCwd() != "projects/demo" {
		t.Fatalf("cursor sessions created = %+v, want one in projects/demo", creates)
	}
	if len(sends) != 2 || sends[0].GetSessionId() != sends[1].GetSessionId() {
		t.Fatalf("cursor sends = %+v, want two in one session", sends)
	}
	if got := sends[0].GetImages(); len(got) != 1 || got[0].GetMimeType() != "image/png" {
		t.Fatalf("photo did not reach cursor: %+v", got)
	}

	sent := fx.bots.Sent()
	if len(sent) != 4 {
		t.Fatalf("telegram deliveries = %d, want two placeholders and two edits", len(sent))
	}
	for index, want := range []string{"cursor reply 1", "cursor reply 2"} {
		delivery := sent[index*2+1]
		if delivery.Edit == nil || !strings.Contains(delivery.Edit.Text, want) {
			t.Fatalf("reply %d = %+v, want it to contain %q", index+1, delivery.Edit, want)
		}
		if sent[index*2].Params.MessageThreadID != "42" {
			t.Fatalf("placeholder %d left the topic", index+1)
		}
	}
}
