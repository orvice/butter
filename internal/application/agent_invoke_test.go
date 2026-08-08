package application

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/repo/config/memory"
	"go.orx.me/apps/butter/internal/runtime/runner"
	"go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// invokeTestRunner is a fake agentRunner whose resolution table maps
// agent_ids to registered runtime names, mirroring runner.Service semantics:
// a set-but-unknown agent_id never falls back to the legacy name.
type invokeTestRunner struct {
	idToName map[string]string
	gotAgent string
	response string
}

func (r *invokeTestRunner) IsReservedAgentName(string) bool { return false }

func (r *invokeTestRunner) Run(_ context.Context, agentName string, _ []*genai.Part, _ string, _ *agentsv1.ContextInfo, _ runner.EventCallback, _ runner.CompactionCallback) (string, error) {
	r.gotAgent = agentName
	return r.response, nil
}

func (r *invokeTestRunner) RunSSE(_ context.Context, agentName string, _ []*genai.Part, _ string, _ *agentsv1.ContextInfo, _ runner.EventCallback, _ runner.CompactionCallback) (string, error) {
	r.gotAgent = agentName
	return r.response, nil
}

func (r *invokeTestRunner) CancelInvocation(string, string) bool { return false }

func (r *invokeTestRunner) ResolveAgentRef(_, agentID, legacyName string) (string, bool) {
	if agentID != "" {
		name, ok := r.idToName[agentID]
		return name, ok
	}
	if legacyName == "" {
		return "", false
	}
	return legacyName, true
}

func (r *invokeTestRunner) GetAgentIdentity(name string) (string, string, bool) {
	for id, n := range r.idToName {
		if n == name {
			return id, name, true
		}
	}
	return "", name, true
}

func newInvokeTestService(fake *invokeTestRunner) (*AgentServiceServer, context.Context) {
	svc := NewAgentServiceServer(memory.New())
	svc.runnerSvc = fake
	return svc, workspace.WithID(context.Background(), wsTest)
}

func TestInvokeAgent_ByAgentID(t *testing.T) {
	fake := &invokeTestRunner{idToName: map[string]string{"helper-v2": "helper"}, response: "hi"}
	svc, ctx := newInvokeTestService(fake)

	resp, err := svc.InvokeAgent(ctx, connect.NewRequest(&agentsv1.InvokeAgentRequest{
		AgentId: "helper-v2",
		Input:   "hello",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if resp.Msg.GetResponse() != "hi" {
		t.Fatalf("response = %q, want hi", resp.Msg.GetResponse())
	}
	if fake.gotAgent != "helper" {
		t.Fatalf("runner invoked with %q, want helper (resolved from agent_id)", fake.gotAgent)
	}
}

func TestInvokeAgent_UnknownAgentIDIsNotFound(t *testing.T) {
	fake := &invokeTestRunner{idToName: map[string]string{}}
	svc, ctx := newInvokeTestService(fake)

	_, err := svc.InvokeAgent(ctx, connect.NewRequest(&agentsv1.InvokeAgentRequest{
		AgentId:   "ghost",
		AgentName: "helper", // must not fall back to the legacy name
		Input:     "hello",
	}))
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("err = %v, want CodeNotFound", err)
	}
	if fake.gotAgent != "" {
		t.Fatalf("runner was invoked with %q despite unknown agent_id", fake.gotAgent)
	}
}

func TestInvokeAgent_RequiresAgentRef(t *testing.T) {
	fake := &invokeTestRunner{}
	svc, ctx := newInvokeTestService(fake)

	_, err := svc.InvokeAgent(ctx, connect.NewRequest(&agentsv1.InvokeAgentRequest{Input: "hello"}))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("err = %v, want CodeInvalidArgument", err)
	}
}

func TestGetAgentRuntimeStatus_ByAgentID(t *testing.T) {
	repo := memory.New()
	svc := NewAgentServiceServer(repo)
	ctx := workspace.WithID(context.Background(), wsTest)

	if _, err := repo.CreateAgent(context.Background(), wsTest, &agentsv1.Agent{
		Name:    "helper",
		AgentId: "helper-v2",
	}); err != nil {
		t.Fatal(err)
	}

	resp, err := svc.GetAgentRuntimeStatus(ctx, connect.NewRequest(&agentsv1.GetAgentRuntimeStatusRequest{
		AgentId: "helper-v2",
	}))
	if err != nil {
		t.Fatal(err)
	}
	st := resp.Msg.GetStatus()
	if st.GetName() != "helper" || st.GetAgentId() != "helper-v2" {
		t.Fatalf("status = name %q agent_id %q, want helper / helper-v2", st.GetName(), st.GetAgentId())
	}
}

func TestGetAgentRuntimeStatus_UnknownAgentIDIsNotFound(t *testing.T) {
	svc := NewAgentServiceServer(memory.New())
	ctx := workspace.WithID(context.Background(), wsTest)

	_, err := svc.GetAgentRuntimeStatus(ctx, connect.NewRequest(&agentsv1.GetAgentRuntimeStatusRequest{
		AgentId: "ghost",
	}))
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("err = %v, want CodeNotFound", err)
	}
}

func TestListAgentRuntimeStatuses_ByAgentIDs(t *testing.T) {
	repo := memory.New()
	svc := NewAgentServiceServer(repo)
	ctx := workspace.WithID(context.Background(), wsTest)

	if _, err := repo.CreateAgent(context.Background(), wsTest, &agentsv1.Agent{Name: "helper", AgentId: "helper-v2"}); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.CreateAgent(context.Background(), wsTest, &agentsv1.Agent{Name: "writer", AgentId: "writer"}); err != nil {
		t.Fatal(err)
	}

	resp, err := svc.ListAgentRuntimeStatuses(ctx, connect.NewRequest(&agentsv1.ListAgentRuntimeStatusesRequest{
		AgentIds: []string{"writer"},
	}))
	if err != nil {
		t.Fatal(err)
	}
	statuses := resp.Msg.GetStatuses()
	if len(statuses) != 1 || statuses[0].GetName() != "writer" || statuses[0].GetAgentId() != "writer" {
		t.Fatalf("statuses = %+v, want single writer row with agent_id", statuses)
	}
}

func TestListAgentRuntimeStatuses_DefaultCarriesAgentIDs(t *testing.T) {
	repo := memory.New()
	svc := NewAgentServiceServer(repo)
	ctx := workspace.WithID(context.Background(), wsTest)

	if _, err := repo.CreateAgent(context.Background(), wsTest, &agentsv1.Agent{Name: "helper", AgentId: "helper-v2"}); err != nil {
		t.Fatal(err)
	}

	resp, err := svc.ListAgentRuntimeStatuses(ctx, connect.NewRequest(&agentsv1.ListAgentRuntimeStatusesRequest{}))
	if err != nil {
		t.Fatal(err)
	}
	statuses := resp.Msg.GetStatuses()
	if len(statuses) != 1 || statuses[0].GetAgentId() != "helper-v2" {
		t.Fatalf("statuses = %+v, want helper row carrying agent_id helper-v2", statuses)
	}
}
