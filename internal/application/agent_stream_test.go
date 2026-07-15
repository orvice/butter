package application

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/repo/config/memory"
	"go.orx.me/apps/butter/internal/runtime/runner"
	"go.orx.me/apps/butter/internal/transport/connectx"
	"go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

// streamTestRunner records the parts forwarded by the service so tests can
// assert on what the runner would execute.
type streamTestRunner struct {
	gotAgent string
	gotParts []*genai.Part
	response string
}

func (r *streamTestRunner) IsReservedAgentName(string) bool { return false }

func (r *streamTestRunner) Run(_ context.Context, agentName string, parts []*genai.Part, _ string, _ *agentsv1.ContextInfo, _ runner.EventCallback, _ runner.CompactionCallback) (string, error) {
	r.gotAgent, r.gotParts = agentName, parts
	return r.response, nil
}

func (r *streamTestRunner) RunSSE(_ context.Context, agentName string, parts []*genai.Part, _ string, _ *agentsv1.ContextInfo, _ runner.EventCallback, _ runner.CompactionCallback) (string, error) {
	r.gotAgent, r.gotParts = agentName, parts
	return r.response, nil
}

func (r *streamTestRunner) CancelInvocation(string, string) bool { return false }

// newStreamAgentTestClient serves the real Connect handler over httptest so
// StreamAgent is exercised end-to-end through the RPC seam. The middleware
// injects the workspace id the same way the production auth middleware does.
func newStreamAgentTestClient(t *testing.T, fake *streamTestRunner) agentsv1connect.AgentServiceClient {
	t.Helper()
	svc := NewAgentServiceServer(memory.New())
	svc.runnerSvc = fake
	path, handler := agentsv1connect.NewAgentServiceHandler(svc, connectx.HandlerOptions()...)
	mux := http.NewServeMux()
	mux.Handle(path, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		handler.ServeHTTP(w, req.WithContext(workspace.WithID(req.Context(), wsTest)))
	}))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return agentsv1connect.NewAgentServiceClient(srv.Client(), srv.URL)
}

// runStreamAgent drives the stream to completion and returns the final
// response text, or the stream error.
func runStreamAgent(t *testing.T, client agentsv1connect.AgentServiceClient, req *agentsv1.StreamAgentRequest) (string, error) {
	t.Helper()
	stream, err := client.StreamAgent(context.Background(), connect.NewRequest(req))
	if err != nil {
		return "", err
	}
	defer stream.Close()
	var final string
	for stream.Receive() {
		if f := stream.Msg().GetFinal(); f != nil {
			final = f.GetResponse()
		}
	}
	if err := stream.Err(); err != nil {
		return "", err
	}
	return final, nil
}

func textInput(text string) *agentsv1.InputPart {
	return &agentsv1.InputPart{Part: &agentsv1.InputPart_Text{Text: text}}
}

func imageInput(mimeType string, data []byte) *agentsv1.InputPart {
	return &agentsv1.InputPart{Part: &agentsv1.InputPart_InlineData{
		InlineData: &agentsv1.InlineData{MimeType: mimeType, Data: data},
	}}
}

func TestStreamAgent_PartsTextAndImageReachRunner(t *testing.T) {
	fake := &streamTestRunner{response: "the image shows a cat"}
	client := newStreamAgentTestClient(t, fake)

	imgData := []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a, 1, 2, 3}
	final, err := runStreamAgent(t, client, &agentsv1.StreamAgentRequest{
		AgentName: "vision-agent",
		Parts: []*agentsv1.InputPart{
			textInput("what is in this picture?"),
			imageInput("image/png", imgData),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if final != "the image shows a cat" {
		t.Fatalf("expected final response from runner, got %q", final)
	}
	if fake.gotAgent != "vision-agent" {
		t.Fatalf("expected runner invoked for vision-agent, got %q", fake.gotAgent)
	}
	if len(fake.gotParts) != 2 {
		t.Fatalf("expected 2 genai parts forwarded to runner, got %d", len(fake.gotParts))
	}
	if fake.gotParts[0].Text != "what is in this picture?" {
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

func TestStreamAgent_MessageOnlyBackwardCompat(t *testing.T) {
	fake := &streamTestRunner{response: "hello"}
	client := newStreamAgentTestClient(t, fake)

	final, err := runStreamAgent(t, client, &agentsv1.StreamAgentRequest{
		AgentName: "chat-agent",
		Message:   "hi there",
	})
	if err != nil {
		t.Fatal(err)
	}
	if final != "hello" {
		t.Fatalf("expected final response, got %q", final)
	}
	if len(fake.gotParts) != 1 || fake.gotParts[0].Text != "hi there" {
		t.Fatalf("expected single text part from message, got %+v", fake.gotParts)
	}
}

func TestStreamAgent_PartsTakePriorityOverMessage(t *testing.T) {
	fake := &streamTestRunner{response: "ok"}
	client := newStreamAgentTestClient(t, fake)

	_, err := runStreamAgent(t, client, &agentsv1.StreamAgentRequest{
		AgentName: "chat-agent",
		Message:   "ignored legacy text",
		Parts:     []*agentsv1.InputPart{textInput("from parts")},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(fake.gotParts) != 1 || fake.gotParts[0].Text != "from parts" {
		t.Fatalf("expected parts to win over message, runner got %+v", fake.gotParts)
	}
}

func TestStreamAgent_ImageOnlyPartsAccepted(t *testing.T) {
	fake := &streamTestRunner{response: "a screenshot"}
	client := newStreamAgentTestClient(t, fake)

	final, err := runStreamAgent(t, client, &agentsv1.StreamAgentRequest{
		AgentName: "vision-agent",
		Parts:     []*agentsv1.InputPart{imageInput("image/jpeg", []byte{0xff, 0xd8, 0xff})},
	})
	if err != nil {
		t.Fatal(err)
	}
	if final != "a screenshot" {
		t.Fatalf("expected image-only request to succeed, got %q", final)
	}
	if len(fake.gotParts) != 1 || fake.gotParts[0].InlineData == nil {
		t.Fatalf("expected single inline data part, got %+v", fake.gotParts)
	}
}

func assertInvalidArgument(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected CodeInvalidArgument error, got nil")
	}
	var connectErr *connect.Error
	if !errors.As(err, &connectErr) || connectErr.Code() != connect.CodeInvalidArgument {
		t.Fatalf("expected CodeInvalidArgument, got %v", err)
	}
}

func TestStreamAgent_UnsupportedMimeTypeRejected(t *testing.T) {
	fake := &streamTestRunner{}
	client := newStreamAgentTestClient(t, fake)

	_, err := runStreamAgent(t, client, &agentsv1.StreamAgentRequest{
		AgentName: "vision-agent",
		Parts:     []*agentsv1.InputPart{imageInput("application/pdf", []byte("%PDF-1.7"))},
	})
	assertInvalidArgument(t, err)
	if fake.gotParts != nil {
		t.Fatalf("runner must not be invoked for rejected input, got %+v", fake.gotParts)
	}
}

func TestStreamAgent_OversizedImageRejected(t *testing.T) {
	fake := &streamTestRunner{}
	client := newStreamAgentTestClient(t, fake)

	oversized := make([]byte, 10<<20+1)
	_, err := runStreamAgent(t, client, &agentsv1.StreamAgentRequest{
		AgentName: "vision-agent",
		Parts:     []*agentsv1.InputPart{imageInput("image/png", oversized)},
	})
	assertInvalidArgument(t, err)
	if fake.gotParts != nil {
		t.Fatalf("runner must not be invoked for rejected input, got %d parts", len(fake.gotParts))
	}
}

func TestStreamAgent_TotalPayloadOverLimitRejected(t *testing.T) {
	fake := &streamTestRunner{}
	client := newStreamAgentTestClient(t, fake)

	// Three 7 MiB images: each under the 10 MiB per-image cap, 21 MiB total.
	img := make([]byte, 7<<20)
	_, err := runStreamAgent(t, client, &agentsv1.StreamAgentRequest{
		AgentName: "vision-agent",
		Parts: []*agentsv1.InputPart{
			imageInput("image/png", img),
			imageInput("image/png", img),
			imageInput("image/png", img),
		},
	})
	assertInvalidArgument(t, err)
	if fake.gotParts != nil {
		t.Fatalf("runner must not be invoked for rejected input, got %d parts", len(fake.gotParts))
	}
}

func TestStreamAgent_TooManyImagesRejected(t *testing.T) {
	fake := &streamTestRunner{}
	client := newStreamAgentTestClient(t, fake)

	parts := make([]*agentsv1.InputPart, 0, 11)
	for range 11 {
		parts = append(parts, imageInput("image/png", []byte{1}))
	}
	_, err := runStreamAgent(t, client, &agentsv1.StreamAgentRequest{
		AgentName: "vision-agent",
		Parts:     parts,
	})
	assertInvalidArgument(t, err)
	if fake.gotParts != nil {
		t.Fatalf("runner must not be invoked for rejected input, got %d parts", len(fake.gotParts))
	}
}

func TestStreamAgent_EmptyPartsAndMessageRejected(t *testing.T) {
	fake := &streamTestRunner{}
	client := newStreamAgentTestClient(t, fake)

	_, err := runStreamAgent(t, client, &agentsv1.StreamAgentRequest{
		AgentName: "chat-agent",
	})
	assertInvalidArgument(t, err)
}
