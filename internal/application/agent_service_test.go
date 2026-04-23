package application

import (
	"context"
	"errors"
	"testing"

	"github.com/twitchtv/twirp"

	configrepo "go.orx.me/apps/butter/internal/repo/config"
	"go.orx.me/apps/butter/internal/repo/config/memory"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type reloadTracker struct {
	calls int
	err   error
}

func (r *reloadTracker) ReloadRunner(context.Context) error {
	r.calls++
	return r.err
}

func (r *reloadTracker) ReloadChannels(context.Context) error {
	r.calls++
	return r.err
}

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

func TestChannelServiceServer_ReloadsRuntime(t *testing.T) {
	store := memory.New()
	svc := NewChannelServiceServer(store)
	runtime := &reloadTracker{}
	svc.SetRuntime(runtime)
	ctx := context.Background()

	_, err := svc.CreateChannel(ctx, &agentsv1.CreateChannelRequest{
		Channel: &agentsv1.AgentChannel{Name: "ch1", AgentName: "agent1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if runtime.calls != 1 {
		t.Fatalf("expected 1 reload call, got %d", runtime.calls)
	}

	_, err = svc.UpdateChannel(ctx, &agentsv1.UpdateChannelRequest{
		Channel: &agentsv1.AgentChannel{Name: "ch1", AgentName: "agent2"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if runtime.calls != 2 {
		t.Fatalf("expected 2 reload calls, got %d", runtime.calls)
	}

	_, err = svc.DeleteChannel(ctx, &agentsv1.DeleteChannelRequest{Name: "ch1"})
	if err != nil {
		t.Fatal(err)
	}
	if runtime.calls != 3 {
		t.Fatalf("expected 3 reload calls, got %d", runtime.calls)
	}
}

func TestChannelServiceServer_ReloadError(t *testing.T) {
	store := memory.New()
	svc := NewChannelServiceServer(store)
	svc.SetRuntime(&reloadTracker{err: errors.New("boom")})

	_, err := svc.CreateChannel(context.Background(), &agentsv1.CreateChannelRequest{
		Channel: &agentsv1.AgentChannel{Name: "ch1", AgentName: "agent1"},
	})
	if twerr, ok := err.(twirp.Error); !ok || twerr.Code() != twirp.Internal {
		t.Fatalf("expected Internal, got %v", err)
	}
	if _, err := store.GetChannel(context.Background(), "ch1"); !errors.Is(err, configrepo.ErrNotFound) {
		t.Fatalf("expected rollback to remove channel, got %v", err)
	}
}

func TestAgentServiceServer_ReloadsRuntime(t *testing.T) {
	store := memory.New()
	svc := NewAgentServiceServer(store)
	runtime := &reloadTracker{}
	svc.SetRuntime(runtime)

	_, err := svc.CreateAgent(context.Background(), &agentsv1.CreateAgentRequest{
		Agent: &agentsv1.Agent{Name: "a1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if runtime.calls != 1 {
		t.Fatalf("expected 1 reload call, got %d", runtime.calls)
	}
}

func TestAgentServiceServer_ReloadErrorRollsBackCreateUpdateDelete(t *testing.T) {
	ctx := context.Background()

	t.Run("create", func(t *testing.T) {
		store := memory.New()
		svc := NewAgentServiceServer(store)
		svc.SetRuntime(&reloadTracker{err: errors.New("boom")})

		_, err := svc.CreateAgent(ctx, &agentsv1.CreateAgentRequest{
			Agent: &agentsv1.Agent{Name: "a1"},
		})
		if twerr, ok := err.(twirp.Error); !ok || twerr.Code() != twirp.Internal {
			t.Fatalf("expected Internal, got %v", err)
		}
		if _, err := store.GetAgent(ctx, "a1"); !errors.Is(err, configrepo.ErrNotFound) {
			t.Fatalf("expected rollback to remove agent, got %v", err)
		}
	})

	t.Run("update", func(t *testing.T) {
		store := memory.New()
		if _, err := store.CreateAgent(ctx, &agentsv1.Agent{Name: "a1", Description: "before"}); err != nil {
			t.Fatalf("seed agent: %v", err)
		}

		svc := NewAgentServiceServer(store)
		svc.SetRuntime(&reloadTracker{err: errors.New("boom")})

		_, err := svc.UpdateAgent(ctx, &agentsv1.UpdateAgentRequest{
			Agent: &agentsv1.Agent{Name: "a1", Description: "after"},
		})
		if twerr, ok := err.(twirp.Error); !ok || twerr.Code() != twirp.Internal {
			t.Fatalf("expected Internal, got %v", err)
		}

		agent, err := store.GetAgent(ctx, "a1")
		if err != nil {
			t.Fatalf("get rolled back agent: %v", err)
		}
		if agent.GetDescription() != "before" {
			t.Fatalf("expected rollback to restore previous description, got %q", agent.GetDescription())
		}
	})

	t.Run("delete", func(t *testing.T) {
		store := memory.New()
		if _, err := store.CreateAgent(ctx, &agentsv1.Agent{Name: "a1", Description: "before"}); err != nil {
			t.Fatalf("seed agent: %v", err)
		}

		svc := NewAgentServiceServer(store)
		svc.SetRuntime(&reloadTracker{err: errors.New("boom")})

		_, err := svc.DeleteAgent(ctx, &agentsv1.DeleteAgentRequest{Name: "a1"})
		if twerr, ok := err.(twirp.Error); !ok || twerr.Code() != twirp.Internal {
			t.Fatalf("expected Internal, got %v", err)
		}

		agent, err := store.GetAgent(ctx, "a1")
		if err != nil {
			t.Fatalf("get rolled back agent: %v", err)
		}
		if agent.GetDescription() != "before" {
			t.Fatalf("expected rollback to restore deleted agent, got %q", agent.GetDescription())
		}
	})
}

func TestMCPServerServiceServer_ReloadsRuntime(t *testing.T) {
	store := memory.New()
	svc := NewMCPServerServiceServer(store)
	runtime := &reloadTracker{}
	svc.SetRuntime(runtime)

	_, err := svc.CreateMCPServer(context.Background(), &agentsv1.CreateMCPServerRequest{
		McpServer: &agentsv1.MCPServer{Id: "m1", Name: "mcp1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if runtime.calls != 1 {
		t.Fatalf("expected 1 reload call, got %d", runtime.calls)
	}
}

func TestMCPServerServiceServer_ReloadErrorRollsBackCreate(t *testing.T) {
	store := memory.New()
	svc := NewMCPServerServiceServer(store)
	svc.SetRuntime(&reloadTracker{err: errors.New("boom")})

	_, err := svc.CreateMCPServer(context.Background(), &agentsv1.CreateMCPServerRequest{
		McpServer: &agentsv1.MCPServer{Id: "m1", Name: "mcp1"},
	})
	if twerr, ok := err.(twirp.Error); !ok || twerr.Code() != twirp.Internal {
		t.Fatalf("expected Internal, got %v", err)
	}
	if _, err := store.GetMCPServer(context.Background(), "m1"); !errors.Is(err, configrepo.ErrNotFound) {
		t.Fatalf("expected rollback to remove mcp server, got %v", err)
	}
}

func TestRemoteAgentServiceServer_ReloadsRuntime(t *testing.T) {
	store := memory.New()
	svc := NewRemoteAgentServiceServer(store)
	runtime := &reloadTracker{}
	svc.SetRuntime(runtime)

	_, err := svc.CreateRemoteAgent(context.Background(), &agentsv1.CreateRemoteAgentRequest{
		RemoteAgent: &agentsv1.RemoteAgent{Id: "r1", Name: "ra1", Url: "http://example.com"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if runtime.calls != 1 {
		t.Fatalf("expected 1 reload call, got %d", runtime.calls)
	}
}

func TestRemoteAgentServiceServer_ReloadErrorRollsBackCreate(t *testing.T) {
	store := memory.New()
	svc := NewRemoteAgentServiceServer(store)
	svc.SetRuntime(&reloadTracker{err: errors.New("boom")})

	_, err := svc.CreateRemoteAgent(context.Background(), &agentsv1.CreateRemoteAgentRequest{
		RemoteAgent: &agentsv1.RemoteAgent{Id: "r1", Name: "ra1", Url: "http://example.com"},
	})
	if twerr, ok := err.(twirp.Error); !ok || twerr.Code() != twirp.Internal {
		t.Fatalf("expected Internal, got %v", err)
	}
	if _, err := store.GetRemoteAgent(context.Background(), "r1"); !errors.Is(err, configrepo.ErrNotFound) {
		t.Fatalf("expected rollback to remove remote agent, got %v", err)
	}
}
