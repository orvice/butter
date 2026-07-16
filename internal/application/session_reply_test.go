package application

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/runtime/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// replyTestRunner records the parts forwarded by ReplySession so tests can
// assert on what the runner would execute.
type replyTestRunner struct {
	gotAgent string
	gotParts []*genai.Part
	response string
}

func (r *replyTestRunner) Run(_ context.Context, agentName string, parts []*genai.Part, _ string, _ *agentsv1.ContextInfo, _ runner.EventCallback, _ runner.CompactionCallback) (string, error) {
	r.gotAgent, r.gotParts = agentName, parts
	return r.response, nil
}

func newReplySessionTestService(fake *replyTestRunner) *SessionServiceServer {
	svc := NewSessionServiceServer()
	svc.runnerSvc = fake
	return svc
}

func TestReplySession_PartsTextAndImageReachRunner(t *testing.T) {
	fake := &replyTestRunner{response: "the chart shows growth"}
	svc := newReplySessionTestService(fake)

	imgData := []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a, 9, 8, 7}
	resp, err := svc.ReplySession(context.Background(), connect.NewRequest(&agentsv1.ReplySessionRequest{
		AgentName: "vision-agent",
		AppName:   "telegram",
		UserId:    "u1",
		SessionId: "s1",
		Parts: []*agentsv1.InputPart{
			textInput("what does this chart show?"),
			imageInput("image/png", imgData),
		},
	}))
	if err != nil {
		t.Fatal(err)
	}
	if resp.Msg.GetResponse() != "the chart shows growth" {
		t.Fatalf("expected runner response, got %q", resp.Msg.GetResponse())
	}
	if fake.gotAgent != "vision-agent" {
		t.Fatalf("expected runner invoked for vision-agent, got %q", fake.gotAgent)
	}
	if len(fake.gotParts) != 2 {
		t.Fatalf("expected 2 genai parts forwarded to runner, got %d", len(fake.gotParts))
	}
	if fake.gotParts[0].Text != "what does this chart show?" {
		t.Fatalf("expected first part text, got %+v", fake.gotParts[0])
	}
	img := fake.gotParts[1].InlineData
	if img == nil {
		t.Fatalf("expected second part to be inline data, got %+v", fake.gotParts[1])
	}
	if img.MIMEType != "image/png" || !bytes.Equal(img.Data, imgData) {
		t.Fatalf("expected png bytes forwarded unchanged, got mime=%q len=%d", img.MIMEType, len(img.Data))
	}
}

func TestReplySession_MessageOnlyBackwardCompat(t *testing.T) {
	fake := &replyTestRunner{response: "done"}
	svc := newReplySessionTestService(fake)

	resp, err := svc.ReplySession(context.Background(), connect.NewRequest(&agentsv1.ReplySessionRequest{
		AgentName: "chat-agent",
		AppName:   "api",
		UserId:    "u1",
		SessionId: "s1",
		Message:   "continue please",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if resp.Msg.GetResponse() != "done" {
		t.Fatalf("expected runner response, got %q", resp.Msg.GetResponse())
	}
	if len(fake.gotParts) != 1 || fake.gotParts[0].Text != "continue please" {
		t.Fatalf("expected single text part from message, got %+v", fake.gotParts)
	}
}

func TestReplySession_PartsTakePriorityOverMessage(t *testing.T) {
	fake := &replyTestRunner{response: "ok"}
	svc := newReplySessionTestService(fake)

	_, err := svc.ReplySession(context.Background(), connect.NewRequest(&agentsv1.ReplySessionRequest{
		AgentName: "chat-agent",
		AppName:   "api",
		UserId:    "u1",
		SessionId: "s1",
		Message:   "ignored legacy text",
		Parts:     []*agentsv1.InputPart{textInput("from parts")},
	}))
	if err != nil {
		t.Fatal(err)
	}
	if len(fake.gotParts) != 1 || fake.gotParts[0].Text != "from parts" {
		t.Fatalf("expected parts to win over message, runner got %+v", fake.gotParts)
	}
}

func TestReplySession_UnsupportedMimeTypeRejected(t *testing.T) {
	fake := &replyTestRunner{}
	svc := newReplySessionTestService(fake)

	_, err := svc.ReplySession(context.Background(), connect.NewRequest(&agentsv1.ReplySessionRequest{
		AgentName: "vision-agent",
		AppName:   "api",
		UserId:    "u1",
		SessionId: "s1",
		Parts:     []*agentsv1.InputPart{imageInput("application/zip", []byte{0x50, 0x4b})},
	}))
	assertInvalidArgument(t, err)
	if fake.gotParts != nil {
		t.Fatalf("runner must not be invoked for rejected input, got %+v", fake.gotParts)
	}
}

func TestReplySession_OversizedMessageRejected(t *testing.T) {
	fake := &replyTestRunner{}
	svc := newReplySessionTestService(fake)

	// The legacy message field carries the same 1 MiB cap as StreamAgent's
	// message and as a text part, so no input path is unbounded.
	_, err := svc.ReplySession(context.Background(), connect.NewRequest(&agentsv1.ReplySessionRequest{
		AgentName: "chat-agent",
		AppName:   "api",
		UserId:    "u1",
		SessionId: "s1",
		Message:   strings.Repeat("a", 1<<20+1),
	}))
	assertInvalidArgument(t, err)
	if fake.gotParts != nil {
		t.Fatalf("runner must not be invoked for rejected input, got %d parts", len(fake.gotParts))
	}
}
