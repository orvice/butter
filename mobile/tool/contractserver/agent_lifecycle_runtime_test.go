package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

func TestAgentLifecycleTombstoneRestoreContract(t *testing.T) {
	ctx := context.Background()
	handler, _, err := newFixture(ctx)
	if err != nil {
		t.Fatalf("fixture: %v", err)
	}
	server := httptest.NewServer(handler)
	defer server.Close()
	login, err := agentsv1connect.NewAuthServiceClient(server.Client(), server.URL+"/api").Login(ctx,
		connect.NewRequest(&agentsv1.LoginRequest{Username: "mobile-test", Password: "contract-test-password"}))
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	token := login.Msg.GetToken()
	wsHeader := func() connect.Option {
		return connect.WithInterceptors(connect.UnaryInterceptorFunc(
			func(next connect.UnaryFunc) connect.UnaryFunc {
				return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
					req.Header().Set("Authorization", "Bearer "+token)
					req.Header().Set("X-Workspace-ID", "alpha")
					return next(ctx, req)
				}
			}))
	}
	client := agentsv1connect.NewAgentServiceClient(http.DefaultClient, server.URL+"/api", wsHeader())

	id := "lifecycle-contract-agent"
	_, err = client.CreateAgent(ctx, connect.NewRequest(&agentsv1.CreateAgentRequest{
		Agent: &agentsv1.Agent{
			Name: id, AgentId: id,
			Type: agentsv1.AgentType_AGENT_TYPE_LLM,
			Config: &agentsv1.AgentConfig{
				Model: "fast", Instruction: "hi",
			},
		},
	}))
	if err != nil {
		t.Fatalf("CreateAgent: %v", err)
	}

	del, err := client.DeleteAgent(ctx, connect.NewRequest(&agentsv1.DeleteAgentRequest{AgentId: id}))
	if err != nil {
		t.Fatalf("DeleteAgent: %v", err)
	}
	if del.Msg.GetOperation().GetStatus() != agentsv1.AgentOperationStatus_AGENT_OPERATION_STATUS_SUCCEEDED {
		t.Fatalf("delete operation: %v", del.Msg.GetOperation())
	}

	got, err := client.GetAgent(ctx, connect.NewRequest(&agentsv1.GetAgentRequest{AgentId: id}))
	if err != nil {
		t.Fatalf("GetAgent after delete: %v", err)
	}
	if got.Msg.GetAgent().GetLifecycleStatus() != agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_DELETED {
		t.Fatalf("want DELETED, got %v", got.Msg.GetAgent().GetLifecycleStatus())
	}

	list, err := client.ListAgentOperations(ctx, connect.NewRequest(&agentsv1.ListAgentOperationsRequest{}))
	if err != nil {
		t.Fatalf("ListAgentOperations: %v", err)
	}
	if len(list.Msg.GetOperations()) == 0 {
		t.Fatal("expected operations")
	}

}
