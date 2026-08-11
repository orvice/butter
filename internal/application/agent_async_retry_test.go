package application

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	adksession "google.golang.org/adk/v2/session"
	"google.golang.org/protobuf/types/known/timestamppb"

	inputpartmemory "go.orx.me/apps/butter/internal/repo/inputpart/memory"
	invocationmemory "go.orx.me/apps/butter/internal/repo/invocation/memory"
	"go.orx.me/apps/butter/internal/runtime/asyncrun"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

var errFailingRunner = errors.New("model backend unavailable")

func seedTerminalInvocation(t *testing.T, repo *invocationmemory.Store, id, sessionID, userID string, status agentsv1.InvocationStatus) *agentsv1.Invocation {
	t.Helper()
	inv := &agentsv1.Invocation{
		Id:          id,
		AgentName:   "test-agent-name",
		AgentId:     "test-agent",
		AppName:     "web-chat",
		UserId:      userID,
		SessionId:   sessionID,
		Status:      status,
		Input:       "original message",
		Error:       "boom",
		Source:      sourceDashboardAsync,
		WorkspaceId: wsTest,
		StartedAt:   timestamppb.Now(),
		FinishedAt:  timestamppb.Now(),
	}
	if err := repo.Save(context.Background(), inv); err != nil {
		t.Fatal(err)
	}
	return inv
}

func TestGetAgentInvocation_LatestBySessionReturnsTerminal(t *testing.T) {
	invRepo := invocationmemory.New()
	ipRepo := inputpartmemory.New()
	svc := &AgentServiceServer{invRepo: invRepo, inputPartRepo: ipRepo}
	seedTerminalInvocation(t, invRepo, "inv-failed", "sess-a", "user-1", agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED)

	ctx := testContextWithUser(wsTest, "user-1")

	// The active-only lookup must keep reporting NotFound for an idle session.
	_, err := svc.GetAgentInvocation(ctx, connect.NewRequest(&agentsv1.GetAgentInvocationRequest{
		SessionId: "sess-a",
	}))
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("active lookup: expected NotFound, got %v", err)
	}

	resp, err := svc.GetAgentInvocation(ctx, connect.NewRequest(&agentsv1.GetAgentInvocationRequest{
		SessionId: "sess-a",
		Latest:    true,
	}))
	if err != nil {
		t.Fatal(err)
	}
	if resp.Msg.GetInvocation().GetId() != "inv-failed" {
		t.Fatalf("latest = %q, want inv-failed", resp.Msg.GetInvocation().GetId())
	}
	if resp.Msg.GetInvocation().GetStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED {
		t.Fatalf("status = %v, want FAILED", resp.Msg.GetInvocation().GetStatus())
	}
}

func TestGetAgentInvocation_LatestBySessionPicksMostRecent(t *testing.T) {
	invRepo := invocationmemory.New()
	svc := &AgentServiceServer{invRepo: invRepo}

	older := seedTerminalInvocation(t, invRepo, "inv-old", "sess-b", "user-1", agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED)
	older.StartedAt = timestamppb.New(time.Now().Add(-time.Hour))
	if err := invRepo.Save(context.Background(), older); err != nil {
		t.Fatal(err)
	}
	seedTerminalInvocation(t, invRepo, "inv-new", "sess-b", "user-1", agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED)

	ctx := testContextWithUser(wsTest, "user-1")
	resp, err := svc.GetAgentInvocation(ctx, connect.NewRequest(&agentsv1.GetAgentInvocationRequest{
		SessionId: "sess-b",
		Latest:    true,
	}))
	if err != nil {
		t.Fatal(err)
	}
	if resp.Msg.GetInvocation().GetId() != "inv-new" {
		t.Fatalf("latest = %q, want inv-new", resp.Msg.GetInvocation().GetId())
	}
}

func TestGetAgentInvocation_LatestBySessionPrivateOwnership(t *testing.T) {
	invRepo := invocationmemory.New()
	svc := &AgentServiceServer{invRepo: invRepo}
	seedTerminalInvocation(t, invRepo, "inv-private", "sess-c", "user-1", agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED)

	otherCtx := testContextWithUser(wsTest, "user-2")
	_, err := svc.GetAgentInvocation(otherCtx, connect.NewRequest(&agentsv1.GetAgentInvocationRequest{
		SessionId: "sess-c",
		Latest:    true,
	}))
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("expected NotFound for another user's private chat, got %v", err)
	}
}

func TestGetAgentInvocation_IncludeInputPartsRestoresOriginalInput(t *testing.T) {
	invRepo := invocationmemory.New()
	ipRepo := inputpartmemory.New()
	svc := &AgentServiceServer{invRepo: invRepo, inputPartRepo: ipRepo}
	seedTerminalInvocation(t, invRepo, "inv-parts", "sess-d", "user-1", agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED)

	stored := []*agentsv1.InputPart{
		{Part: &agentsv1.InputPart_Text{Text: "describe this image"}},
		{Part: &agentsv1.InputPart_InlineData{InlineData: &agentsv1.InlineData{
			MimeType: "image/png",
			Data:     []byte{0x89, 0x50, 0x4e, 0x47},
		}}},
	}
	if err := ipRepo.SaveAll(context.Background(), "inv-parts", stored); err != nil {
		t.Fatal(err)
	}

	ctx := testContextWithUser(wsTest, "user-1")
	resp, err := svc.GetAgentInvocation(ctx, connect.NewRequest(&agentsv1.GetAgentInvocationRequest{
		InvocationId:      "inv-parts",
		IncludeInputParts: true,
	}))
	if err != nil {
		t.Fatal(err)
	}
	parts := resp.Msg.GetInputParts()
	if len(parts) != 2 {
		t.Fatalf("parts = %d, want 2", len(parts))
	}
	if parts[0].GetText() != "describe this image" {
		t.Fatalf("part[0] = %q, want original text", parts[0].GetText())
	}
	if parts[1].GetInlineData().GetMimeType() != "image/png" || len(parts[1].GetInlineData().GetData()) != 4 {
		t.Fatalf("part[1] inline data not preserved: %+v", parts[1])
	}

	// Without the flag the payload stays lean.
	resp, err = svc.GetAgentInvocation(ctx, connect.NewRequest(&agentsv1.GetAgentInvocationRequest{
		InvocationId: "inv-parts",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Msg.GetInputParts()) != 0 {
		t.Fatal("input parts returned without include_input_parts")
	}
}

func TestGetAgentInvocation_IncludeInputPartsEmptyAfterCleanup(t *testing.T) {
	invRepo := invocationmemory.New()
	ipRepo := inputpartmemory.New()
	svc := &AgentServiceServer{invRepo: invRepo, inputPartRepo: ipRepo}
	seedTerminalInvocation(t, invRepo, "inv-cleaned", "sess-e", "user-1", agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED)

	ctx := testContextWithUser(wsTest, "user-1")
	resp, err := svc.GetAgentInvocation(ctx, connect.NewRequest(&agentsv1.GetAgentInvocationRequest{
		InvocationId:      "inv-cleaned",
		IncludeInputParts: true,
	}))
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Msg.GetInputParts()) != 0 {
		t.Fatal("expected no parts for a cleaned-up successful invocation")
	}
}

// TestSubmitAgentInvocation_OperationalErrorStaysOutOfSessionEvents drives a
// real coordinator with a failing runner and asserts the failure is recorded
// only on the Invocation — never appended as an Agent-authored Session event
// that would poison the conversation context.
func TestSubmitAgentInvocation_OperationalErrorStaysOutOfSessionEvents(t *testing.T) {
	invRepo := invocationmemory.New()
	ipRepo := inputpartmemory.New()
	fake := &asyncTestRunner{idToName: map[string]string{"test-agent": "test-agent-name"}, err: errFailingRunner}
	coord := asyncrun.New(invRepo, ipRepo, fake, asyncrun.Config{MaxRunDuration: 5 * time.Second})
	sessionSvc := adksession.InMemoryService()

	svc := &AgentServiceServer{
		runnerSvc:     fake,
		invRepo:       invRepo,
		inputPartRepo: ipRepo,
		asyncCoord:    coord,
		sessionSvc:    sessionSvc,
	}

	ctx := testContextWithUser(wsTest, "user-1")
	resp, err := svc.SubmitAgentInvocation(ctx, connect.NewRequest(&agentsv1.SubmitAgentInvocationRequest{
		RequestId: "req-op-err",
		AgentId:   "test-agent",
		Message:   "hello",
	}))
	if err != nil {
		t.Fatal(err)
	}

	deadline := time.After(3 * time.Second)
	for {
		got, _ := invRepo.Get(context.Background(), wsTest, resp.Msg.GetInvocationId())
		if got != nil && got.GetStatus() == agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED {
			if got.GetError() != errFailingRunner.Error() {
				t.Fatalf("invocation error = %q, want %q", got.GetError(), errFailingRunner.Error())
			}
			break
		}
		select {
		case <-deadline:
			t.Fatal("timed out waiting for FAILED")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}

	sessResp, err := sessionSvc.Get(ctx, &adksession.GetRequest{
		AppName:   "web-chat",
		UserID:    "user-1",
		SessionID: resp.Msg.GetSessionId(),
	})
	if err != nil {
		t.Fatal(err)
	}
	for evt := range sessResp.Session.Events().All() {
		content, jsonErr := json.Marshal(evt.Content)
		if jsonErr == nil && strings.Contains(string(content), errFailingRunner.Error()) {
			t.Fatalf("operational error leaked into session event: %s", content)
		}
	}
}
