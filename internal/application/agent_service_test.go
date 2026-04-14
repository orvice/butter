package application

import (
	"context"
	"testing"

	"github.com/twitchtv/twirp"

	"go.orx.me/apps/butter/internal/repo/config/memory"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestAgentServiceServer_CRUD(t *testing.T) {
	store := memory.New()
	svc := NewAgentServiceServer(store)
	ctx := context.Background()

	// List empty
	resp, err := svc.ListAgents(ctx, &agentsv1.ListAgentsRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.GetAgents()) != 0 {
		t.Fatalf("expected 0 agents, got %d", len(resp.GetAgents()))
	}

	// Create
	createResp, err := svc.CreateAgent(ctx, &agentsv1.CreateAgentRequest{
		Agent: &agentsv1.Agent{Name: "a1", Description: "test"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if createResp.GetAgent().GetName() != "a1" {
		t.Fatalf("expected a1, got %s", createResp.GetAgent().GetName())
	}

	// Create duplicate
	_, err = svc.CreateAgent(ctx, &agentsv1.CreateAgentRequest{
		Agent: &agentsv1.Agent{Name: "a1"},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if twerr, ok := err.(twirp.Error); !ok || twerr.Code() != twirp.AlreadyExists {
		t.Fatalf("expected AlreadyExists, got %v", err)
	}

	// Get
	getResp, err := svc.GetAgent(ctx, &agentsv1.GetAgentRequest{Name: "a1"})
	if err != nil {
		t.Fatal(err)
	}
	if getResp.GetAgent().GetDescription() != "test" {
		t.Fatalf("expected test, got %s", getResp.GetAgent().GetDescription())
	}

	// Get not found
	_, err = svc.GetAgent(ctx, &agentsv1.GetAgentRequest{Name: "nope"})
	if twerr, ok := err.(twirp.Error); !ok || twerr.Code() != twirp.NotFound {
		t.Fatalf("expected NotFound, got %v", err)
	}

	// Update
	updateResp, err := svc.UpdateAgent(ctx, &agentsv1.UpdateAgentRequest{
		Agent: &agentsv1.Agent{Name: "a1", Description: "updated"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if updateResp.GetAgent().GetDescription() != "updated" {
		t.Fatalf("expected updated, got %s", updateResp.GetAgent().GetDescription())
	}

	// Delete
	_, err = svc.DeleteAgent(ctx, &agentsv1.DeleteAgentRequest{Name: "a1"})
	if err != nil {
		t.Fatal(err)
	}

	// Delete not found
	_, err = svc.DeleteAgent(ctx, &agentsv1.DeleteAgentRequest{Name: "a1"})
	if twerr, ok := err.(twirp.Error); !ok || twerr.Code() != twirp.NotFound {
		t.Fatalf("expected NotFound, got %v", err)
	}
}

func TestMCPServerServiceServer_CRUD(t *testing.T) {
	store := memory.New()
	svc := NewMCPServerServiceServer(store)
	ctx := context.Background()

	created, err := svc.CreateMCPServer(ctx, &agentsv1.CreateMCPServerRequest{
		McpServer: &agentsv1.MCPServer{Id: "m1", Name: "mcp1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.GetMcpServer().GetId() != "m1" {
		t.Fatalf("expected m1, got %s", created.GetMcpServer().GetId())
	}

	_, err = svc.CreateMCPServer(ctx, &agentsv1.CreateMCPServerRequest{
		McpServer: &agentsv1.MCPServer{Id: "m1"},
	})
	if twerr, ok := err.(twirp.Error); !ok || twerr.Code() != twirp.AlreadyExists {
		t.Fatalf("expected AlreadyExists, got %v", err)
	}

	got, err := svc.GetMCPServer(ctx, &agentsv1.GetMCPServerRequest{Id: "m1"})
	if err != nil {
		t.Fatal(err)
	}
	if got.GetMcpServer().GetName() != "mcp1" {
		t.Fatalf("expected mcp1, got %s", got.GetMcpServer().GetName())
	}

	_, err = svc.DeleteMCPServer(ctx, &agentsv1.DeleteMCPServerRequest{Id: "m1"})
	if err != nil {
		t.Fatal(err)
	}

	_, err = svc.GetMCPServer(ctx, &agentsv1.GetMCPServerRequest{Id: "m1"})
	if twerr, ok := err.(twirp.Error); !ok || twerr.Code() != twirp.NotFound {
		t.Fatalf("expected NotFound, got %v", err)
	}
}

func TestRemoteAgentServiceServer_CRUD(t *testing.T) {
	store := memory.New()
	svc := NewRemoteAgentServiceServer(store)
	ctx := context.Background()

	created, err := svc.CreateRemoteAgent(ctx, &agentsv1.CreateRemoteAgentRequest{
		RemoteAgent: &agentsv1.RemoteAgent{Id: "r1", Name: "ra1", Url: "http://example.com"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.GetRemoteAgent().GetId() != "r1" {
		t.Fatalf("expected r1, got %s", created.GetRemoteAgent().GetId())
	}

	_, err = svc.CreateRemoteAgent(ctx, &agentsv1.CreateRemoteAgentRequest{
		RemoteAgent: &agentsv1.RemoteAgent{Id: "r1"},
	})
	if twerr, ok := err.(twirp.Error); !ok || twerr.Code() != twirp.AlreadyExists {
		t.Fatalf("expected AlreadyExists, got %v", err)
	}

	got, err := svc.GetRemoteAgent(ctx, &agentsv1.GetRemoteAgentRequest{Id: "r1"})
	if err != nil {
		t.Fatal(err)
	}
	if got.GetRemoteAgent().GetName() != "ra1" {
		t.Fatalf("expected ra1, got %s", got.GetRemoteAgent().GetName())
	}

	_, err = svc.DeleteRemoteAgent(ctx, &agentsv1.DeleteRemoteAgentRequest{Id: "r1"})
	if err != nil {
		t.Fatal(err)
	}

	_, err = svc.GetRemoteAgent(ctx, &agentsv1.GetRemoteAgentRequest{Id: "r1"})
	if twerr, ok := err.(twirp.Error); !ok || twerr.Code() != twirp.NotFound {
		t.Fatalf("expected NotFound, got %v", err)
	}
}
