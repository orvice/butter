package telegram

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	"connectrpc.com/connect"
	piv1 "github.com/orvice/butter-box/pkg/proto/butterbox/pi/v1"
	"github.com/orvice/butter-box/pkg/proto/butterbox/pi/v1/piv1connect"
	adkrunner "google.golang.org/adk/v2/runner"
	adksession "google.golang.org/adk/v2/session"
	"google.golang.org/protobuf/proto"

	"go.orx.me/apps/butter/internal/runtime/pibox"
	"go.orx.me/apps/butter/internal/runtime/runner"
	"go.orx.me/apps/butter/internal/telegramapi"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type telegramPiClient struct {
	piv1connect.PiServiceClient

	mu      sync.Mutex
	creates []*piv1.CreateSessionRequest
	submits []*piv1.SubmitMessageRequest
}

func (c *telegramPiClient) CreateSession(_ context.Context, req *connect.Request[piv1.CreateSessionRequest]) (*connect.Response[piv1.CreateSessionResponse], error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.creates = append(c.creates, proto.Clone(req.Msg).(*piv1.CreateSessionRequest))
	return connect.NewResponse(&piv1.CreateSessionResponse{
		Session: &piv1.Session{Id: "pi-session-1", Cwd: req.Msg.GetCwd()},
	}), nil
}

func (c *telegramPiClient) SubmitMessage(_ context.Context, req *connect.Request[piv1.SubmitMessageRequest]) (*connect.Response[piv1.SubmitMessageResponse], error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.submits = append(c.submits, proto.Clone(req.Msg).(*piv1.SubmitMessageRequest))
	return connect.NewResponse(&piv1.SubmitMessageResponse{
		TurnCursor: fmt.Sprintf("cursor-%d", len(c.submits)),
	}), nil
}

func (c *telegramPiClient) GetTurn(_ context.Context, _ *connect.Request[piv1.GetTurnRequest]) (*connect.Response[piv1.GetTurnResponse], error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return connect.NewResponse(&piv1.GetTurnResponse{
		Result: &piv1.TurnResult{
			Text:       fmt.Sprintf("pi reply %d", len(c.submits)),
			StopReason: "stop",
		},
	}), nil
}

func (c *telegramPiClient) AbortSession(_ context.Context, _ *connect.Request[piv1.AbortSessionRequest]) (*connect.Response[piv1.AbortSessionResponse], error) {
	return connect.NewResponse(&piv1.AbortSessionResponse{}), nil
}

func (c *telegramPiClient) snapshot() ([]*piv1.CreateSessionRequest, []*piv1.SubmitMessageRequest) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]*piv1.CreateSessionRequest(nil), c.creates...),
		append([]*piv1.SubmitMessageRequest(nil), c.submits...)
}

type telegramPiFactory struct{ client *telegramPiClient }

func (f telegramPiFactory) ClientFor(context.Context, string, string) (piv1connect.PiServiceClient, error) {
	return f.client, nil
}

// This crosses the whole in-process exit path: a topic update is admitted,
// its photo is downloaded, the real runner and pibox bridge call PiService,
// and the answer edits the processing message in the originating topic.
func TestPiAgentAnswersInTelegramTopicWithContinuityAndPhoto(t *testing.T) {
	piClient := &telegramPiClient{}
	sessions := adksession.InMemoryService()
	agents, err := runner.NewServiceWithMCPHTTPClientFactory(
		t.Context(), []agentsv1.Agent{{
			Name:        "pi-coder",
			AgentId:     "pi-coder",
			WorkspaceId: "ws-a",
			Type:        agentsv1.AgentType_AGENT_TYPE_PI,
			Config: &agentsv1.AgentConfig{Pi: &agentsv1.PiAgentConfig{
				ButterboxId: "box-1",
				WorkingDir:  "projects/demo",
			}},
		}}, nil, nil, nil, nil,
		sessions, nil, nil, nil, 0, nil, adkrunner.PluginConfig{}, nil,
		pibox.AgentBuilder(telegramPiFactory{client: piClient}), nil,
	)
	if err != nil {
		t.Fatalf("build runner: %v", err)
	}

	fx := newOrchestratorFixture(t, func(config *agentsv1.TelegramDestinationConfig) {
		config.AgentId = "pi-coder"
		config.Model = "butter-override"
		config.SelectableModels = []string{"butter-override"}
	})
	fx.orchestrator.runner = agents
	fileClient := &fakeFileClient{
		file: telegramapi.File{
			FileID: "large", FilePath: "photos/file_1.png", FileSize: int64(len(pngBytes)),
		},
		data: pngBytes,
	}
	fx.orchestrator.SetFileClientFactory(
		func(context.Context, string, string) (telegramapi.FileClient, error) {
			return fileClient, nil
		},
	)

	if err := fx.orchestrator.Handle(t.Context(), fx.eventForStored(photoUpdate("inspect this"))); err != nil {
		t.Fatalf("photo turn: %v", err)
	}
	if err := fx.orchestrator.Handle(t.Context(), fx.eventForStored(message(realUser, "continue", ""))); err != nil {
		t.Fatalf("second turn: %v", err)
	}

	creates, submits := piClient.snapshot()
	if len(creates) != 1 {
		t.Fatalf("pi sessions created = %d, want one shared across topic turns", len(creates))
	}
	if len(submits) != 2 {
		t.Fatalf("pi submissions = %d, want two", len(submits))
	}
	if submits[0].GetSessionId() != submits[1].GetSessionId() {
		t.Fatalf("pi session changed between turns: %q != %q", submits[0].GetSessionId(), submits[1].GetSessionId())
	}
	if got := submits[0].GetImages(); len(got) != 1 || got[0].GetMimeType() != "image/png" || len(got[0].GetData()) == 0 {
		t.Fatalf("photo did not reach pi: %+v", got)
	}
	if got := submits[1].GetImages(); len(got) != 0 {
		t.Fatalf("second text turn unexpectedly carried images: %+v", got)
	}

	sent := fx.bots.Sent()
	if len(sent) != 4 {
		t.Fatalf("telegram deliveries = %d, want two placeholders and two edits", len(sent))
	}
	for _, index := range []int{0, 2} {
		if sent[index].Params.MessageThreadID != "42" {
			t.Fatalf("placeholder %d left the topic: thread %q", index/2+1, sent[index].Params.MessageThreadID)
		}
	}
	for index, want := range []string{"pi reply 1", "pi reply 2"} {
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
