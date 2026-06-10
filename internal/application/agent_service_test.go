package application

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	"go.orx.me/apps/butter/internal/repo/config/memory"
	"go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const wsTest = "ws-test"

func testCtx() context.Context {
	return workspace.WithID(context.Background(), wsTest)
}

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

func testMCPServer(id, name string) *agentsv1.MCPServer {
	return &agentsv1.MCPServer{
		Id:        id,
		Name:      name,
		Transport: agentsv1.MCPServerTransport_MCP_SERVER_TRANSPORT_STREAMABLE_HTTP,
		Url:       "https://mcp.example.com/mcp",
	}
}

func TestAgentServiceServer_CRUD(t *testing.T) {
	store := memory.New()
	svc := NewAgentServiceServer(store)
	ctx := testCtx()

	// List empty
	resp, err := svc.ListAgents(ctx, connect.NewRequest(&agentsv1.ListAgentsRequest{}))
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Msg.GetAgents()) != 0 {
		t.Fatalf("expected 0 agents, got %d", len(resp.Msg.GetAgents()))
	}

	// Create
	createResp, err := svc.CreateAgent(ctx, connect.NewRequest(&agentsv1.CreateAgentRequest{
		Agent: &agentsv1.Agent{Name: "a1", Description: "test"},
	}))
	if err != nil {
		t.Fatal(err)
	}
	if createResp.Msg.GetAgent().GetName() != "a1" {
		t.Fatalf("expected a1, got %s", createResp.Msg.GetAgent().GetName())
	}

	// Create duplicate
	_, err = svc.CreateAgent(ctx, connect.NewRequest(&agentsv1.CreateAgentRequest{
		Agent: &agentsv1.Agent{Name: "a1"},
	}))
	if err == nil {
		t.Fatal("expected error")
	}
	if twerr, ok := err.(*connect.Error); !ok || twerr.Code() != connect.CodeAlreadyExists {
		t.Fatalf("expected AlreadyExists, got %v", err)
	}

	// Get
	getResp, err := svc.GetAgent(ctx, connect.NewRequest(&agentsv1.GetAgentRequest{Name: "a1"}))
	if err != nil {
		t.Fatal(err)
	}
	if getResp.Msg.GetAgent().GetDescription() != "test" {
		t.Fatalf("expected test, got %s", getResp.Msg.GetAgent().GetDescription())
	}

	// Get not found
	_, err = svc.GetAgent(ctx, connect.NewRequest(&agentsv1.GetAgentRequest{Name: "nope"}))
	if twerr, ok := err.(*connect.Error); !ok || twerr.Code() != connect.CodeNotFound {
		t.Fatalf("expected NotFound, got %v", err)
	}

	// Update
	updateResp, err := svc.UpdateAgent(ctx, connect.NewRequest(&agentsv1.UpdateAgentRequest{
		Agent: &agentsv1.Agent{Name: "a1", Description: "updated"},
	}))
	if err != nil {
		t.Fatal(err)
	}
	if updateResp.Msg.GetAgent().GetDescription() != "updated" {
		t.Fatalf("expected updated, got %s", updateResp.Msg.GetAgent().GetDescription())
	}

	// Delete
	_, err = svc.DeleteAgent(ctx, connect.NewRequest(&agentsv1.DeleteAgentRequest{Name: "a1"}))
	if err != nil {
		t.Fatal(err)
	}

	// Delete not found
	_, err = svc.DeleteAgent(ctx, connect.NewRequest(&agentsv1.DeleteAgentRequest{Name: "a1"}))
	if twerr, ok := err.(*connect.Error); !ok || twerr.Code() != connect.CodeNotFound {
		t.Fatalf("expected NotFound, got %v", err)
	}
}

func TestMCPServerServiceServer_CRUD(t *testing.T) {
	store := memory.New()
	svc := NewMCPServerServiceServer(store)
	ctx := testCtx()

	created, err := svc.CreateMCPServer(ctx, connect.NewRequest(&agentsv1.CreateMCPServerRequest{
		McpServer: testMCPServer("m1", "mcp1"),
	}))
	if err != nil {
		t.Fatal(err)
	}
	if created.Msg.GetMcpServer().GetId() != "m1" {
		t.Fatalf("expected m1, got %s", created.Msg.GetMcpServer().GetId())
	}

	_, err = svc.CreateMCPServer(ctx, connect.NewRequest(&agentsv1.CreateMCPServerRequest{
		McpServer: testMCPServer("m1", "mcp1 duplicate"),
	}))
	if twerr, ok := err.(*connect.Error); !ok || twerr.Code() != connect.CodeAlreadyExists {
		t.Fatalf("expected AlreadyExists, got %v", err)
	}

	got, err := svc.GetMCPServer(ctx, connect.NewRequest(&agentsv1.GetMCPServerRequest{Id: "m1"}))
	if err != nil {
		t.Fatal(err)
	}
	if got.Msg.GetMcpServer().GetName() != "mcp1" {
		t.Fatalf("expected mcp1, got %s", got.Msg.GetMcpServer().GetName())
	}

	_, err = svc.DeleteMCPServer(ctx, connect.NewRequest(&agentsv1.DeleteMCPServerRequest{Id: "m1"}))
	if err != nil {
		t.Fatal(err)
	}

	_, err = svc.GetMCPServer(ctx, connect.NewRequest(&agentsv1.GetMCPServerRequest{Id: "m1"}))
	if twerr, ok := err.(*connect.Error); !ok || twerr.Code() != connect.CodeNotFound {
		t.Fatalf("expected NotFound, got %v", err)
	}
}

func TestMCPServerServiceServer_ValidationRejectsUnsupportedTransportAndMissingURL(t *testing.T) {
	store := memory.New()
	svc := NewMCPServerServiceServer(store)

	_, err := svc.CreateMCPServer(testCtx(), connect.NewRequest(&agentsv1.CreateMCPServerRequest{
		McpServer: &agentsv1.MCPServer{
			Id:        "m1",
			Name:      "mcp1",
			Transport: agentsv1.MCPServerTransport_MCP_SERVER_TRANSPORT_UNSPECIFIED,
			Url:       "https://mcp.example.com/mcp",
		},
	}))
	if twerr, ok := err.(*connect.Error); !ok || twerr.Code() != connect.CodeInvalidArgument {
		t.Fatalf("expected InvalidArgument for unsupported transport, got %v", err)
	}

	_, err = svc.CreateMCPServer(testCtx(), connect.NewRequest(&agentsv1.CreateMCPServerRequest{
		McpServer: &agentsv1.MCPServer{
			Id:        "m2",
			Name:      "mcp2",
			Transport: agentsv1.MCPServerTransport_MCP_SERVER_TRANSPORT_STREAMABLE_HTTP,
		},
	}))
	if twerr, ok := err.(*connect.Error); !ok || twerr.Code() != connect.CodeInvalidArgument {
		t.Fatalf("expected InvalidArgument for missing URL, got %v", err)
	}
}

func TestRemoteAgentServiceServer_CRUD(t *testing.T) {
	store := memory.New()
	svc := NewRemoteAgentServiceServer(store)
	ctx := testCtx()

	created, err := svc.CreateRemoteAgent(ctx, connect.NewRequest(&agentsv1.CreateRemoteAgentRequest{
		RemoteAgent: &agentsv1.RemoteAgent{Id: "r1", Name: "ra1", Url: "http://example.com"},
	}))
	if err != nil {
		t.Fatal(err)
	}
	if created.Msg.GetRemoteAgent().GetId() != "r1" {
		t.Fatalf("expected r1, got %s", created.Msg.GetRemoteAgent().GetId())
	}

	_, err = svc.CreateRemoteAgent(ctx, connect.NewRequest(&agentsv1.CreateRemoteAgentRequest{
		RemoteAgent: &agentsv1.RemoteAgent{Id: "r1"},
	}))
	if twerr, ok := err.(*connect.Error); !ok || twerr.Code() != connect.CodeAlreadyExists {
		t.Fatalf("expected AlreadyExists, got %v", err)
	}

	got, err := svc.GetRemoteAgent(ctx, connect.NewRequest(&agentsv1.GetRemoteAgentRequest{Id: "r1"}))
	if err != nil {
		t.Fatal(err)
	}
	if got.Msg.GetRemoteAgent().GetName() != "ra1" {
		t.Fatalf("expected ra1, got %s", got.Msg.GetRemoteAgent().GetName())
	}

	_, err = svc.DeleteRemoteAgent(ctx, connect.NewRequest(&agentsv1.DeleteRemoteAgentRequest{Id: "r1"}))
	if err != nil {
		t.Fatal(err)
	}

	_, err = svc.GetRemoteAgent(ctx, connect.NewRequest(&agentsv1.GetRemoteAgentRequest{Id: "r1"}))
	if twerr, ok := err.(*connect.Error); !ok || twerr.Code() != connect.CodeNotFound {
		t.Fatalf("expected NotFound, got %v", err)
	}
}

func TestRemoteAgentServiceServer_DaemonValidation(t *testing.T) {
	store := memory.New()
	svc := NewRemoteAgentServiceServer(store)
	ctx := testCtx()

	_, err := svc.CreateRemoteAgent(ctx, connect.NewRequest(&agentsv1.CreateRemoteAgentRequest{
		RemoteAgent: &agentsv1.RemoteAgent{
			Id:       "daemon-agent",
			Name:     "Daemon Agent",
			Protocol: agentsv1.RemoteAgentProtocol_REMOTE_AGENT_PROTOCOL_DAEMON,
		},
	}))
	if twerr, ok := err.(*connect.Error); !ok || twerr.Code() != connect.CodeInvalidArgument {
		t.Fatalf("expected InvalidArgument for missing daemon_runtime_id, got %v", err)
	}

	_, err = svc.CreateRemoteAgent(ctx, connect.NewRequest(&agentsv1.CreateRemoteAgentRequest{
		RemoteAgent: &agentsv1.RemoteAgent{
			Id:              "daemon-agent",
			Name:            "Daemon Agent",
			Protocol:        agentsv1.RemoteAgentProtocol_REMOTE_AGENT_PROTOCOL_DAEMON,
			DaemonRuntimeId: "runtime-1",
			AcpRuntime:      "shell",
		},
	}))
	if twerr, ok := err.(*connect.Error); !ok || twerr.Code() != connect.CodeInvalidArgument {
		t.Fatalf("expected InvalidArgument for unsupported acp_runtime, got %v", err)
	}

	_, err = svc.CreateRemoteAgent(ctx, connect.NewRequest(&agentsv1.CreateRemoteAgentRequest{
		RemoteAgent: &agentsv1.RemoteAgent{
			Id:              "daemon-agent",
			Name:            "Daemon Agent",
			Protocol:        agentsv1.RemoteAgentProtocol_REMOTE_AGENT_PROTOCOL_DAEMON,
			DaemonRuntimeId: "runtime-1",
			AcpRuntime:      "opencode",
		},
	}))
	if twerr, ok := err.(*connect.Error); !ok || twerr.Code() != connect.CodeNotFound {
		t.Fatalf("expected NotFound for missing daemon runtime, got %v", err)
	}

	if _, err := store.CreateDaemonRuntime(ctx, wsTest, &agentsv1.DaemonRuntime{
		Id:   "runtime-1",
		Name: "Runtime 1",
	}); err != nil {
		t.Fatal(err)
	}
	created, err := svc.CreateRemoteAgent(ctx, connect.NewRequest(&agentsv1.CreateRemoteAgentRequest{
		RemoteAgent: &agentsv1.RemoteAgent{
			Id:              "daemon-agent",
			Name:            "Daemon Agent",
			Protocol:        agentsv1.RemoteAgentProtocol_REMOTE_AGENT_PROTOCOL_DAEMON,
			DaemonRuntimeId: "runtime-1",
			AcpRuntime:      "opencode",
		},
	}))
	if err != nil {
		t.Fatal(err)
	}
	if created.Msg.GetRemoteAgent().GetDaemonRuntimeId() != "runtime-1" {
		t.Fatalf("expected runtime-1, got %q", created.Msg.GetRemoteAgent().GetDaemonRuntimeId())
	}
}

func TestChannelServiceServer_ReloadsRuntime(t *testing.T) {
	store := memory.New()
	svc := NewChannelServiceServer(store)
	runtime := &reloadTracker{}
	svc.SetRuntime(runtime)
	ctx := testCtx()

	_, err := svc.CreateChannel(ctx, connect.NewRequest(&agentsv1.CreateChannelRequest{
		Channel: &agentsv1.AgentChannel{Name: "ch1", AgentName: "agent1"},
	}))
	if err != nil {
		t.Fatal(err)
	}
	if runtime.calls != 1 {
		t.Fatalf("expected 1 reload call, got %d", runtime.calls)
	}

	_, err = svc.UpdateChannel(ctx, connect.NewRequest(&agentsv1.UpdateChannelRequest{
		Channel: &agentsv1.AgentChannel{Name: "ch1", AgentName: "agent2"},
	}))
	if err != nil {
		t.Fatal(err)
	}
	if runtime.calls != 2 {
		t.Fatalf("expected 2 reload calls, got %d", runtime.calls)
	}

	_, err = svc.DeleteChannel(ctx, connect.NewRequest(&agentsv1.DeleteChannelRequest{Name: "ch1"}))
	if err != nil {
		t.Fatal(err)
	}
	if runtime.calls != 3 {
		t.Fatalf("expected 3 reload calls, got %d", runtime.calls)
	}
}

func TestChannelServiceServer_ValidatesTriggers(t *testing.T) {
	ctx := testCtx()

	t.Run("rejects unspecified on create", func(t *testing.T) {
		store := memory.New()
		svc := NewChannelServiceServer(store)

		_, err := svc.CreateChannel(ctx, connect.NewRequest(&agentsv1.CreateChannelRequest{
			Channel: &agentsv1.AgentChannel{
				Name:      "ch1",
				AgentName: "agent1",
				Triggers: []*agentsv1.AgentTrigger{{
					Type: agentsv1.AgentTriggerType_AGENT_TRIGGER_TYPE_UNSPECIFIED,
				}},
			},
		}))
		if twerr, ok := err.(*connect.Error); !ok || twerr.Code() != connect.CodeInvalidArgument {
			t.Fatalf("expected InvalidArgument, got %v", err)
		}
		if _, err := store.GetChannel(ctx, wsTest, "ch1"); !errors.Is(err, configrepo.ErrNotFound) {
			t.Fatalf("expected invalid channel not to be persisted, got %v", err)
		}
	})

	t.Run("rejects unimplemented mention on update", func(t *testing.T) {
		store := memory.New()
		if _, err := store.CreateChannel(ctx, wsTest, &agentsv1.AgentChannel{Name: "ch1", AgentName: "agent1"}); err != nil {
			t.Fatalf("seed channel: %v", err)
		}
		svc := NewChannelServiceServer(store)

		_, err := svc.UpdateChannel(ctx, connect.NewRequest(&agentsv1.UpdateChannelRequest{
			Channel: &agentsv1.AgentChannel{
				Name:      "ch1",
				AgentName: "agent1",
				Triggers: []*agentsv1.AgentTrigger{{
					Type: agentsv1.AgentTriggerType_AGENT_TRIGGER_TYPE_MENTION,
				}},
			},
		}))
		if twerr, ok := err.(*connect.Error); !ok || twerr.Code() != connect.CodeInvalidArgument {
			t.Fatalf("expected InvalidArgument, got %v", err)
		}
	})

	t.Run("accepts supported trigger", func(t *testing.T) {
		store := memory.New()
		svc := NewChannelServiceServer(store)

		_, err := svc.CreateChannel(ctx, connect.NewRequest(&agentsv1.CreateChannelRequest{
			Channel: &agentsv1.AgentChannel{
				Name:      "ch1",
				AgentName: "agent1",
				Triggers: []*agentsv1.AgentTrigger{{
					Type: agentsv1.AgentTriggerType_AGENT_TRIGGER_TYPE_MESSAGE,
				}},
			},
		}))
		if err != nil {
			t.Fatalf("expected supported trigger to pass, got %v", err)
		}
	})
}

func TestChannelServiceServer_ReloadError(t *testing.T) {
	store := memory.New()
	svc := NewChannelServiceServer(store)
	svc.SetRuntime(&reloadTracker{err: errors.New("boom")})

	_, err := svc.CreateChannel(testCtx(), connect.NewRequest(&agentsv1.CreateChannelRequest{
		Channel: &agentsv1.AgentChannel{Name: "ch1", AgentName: "agent1"},
	}))
	if twerr, ok := err.(*connect.Error); !ok || twerr.Code() != connect.CodeInternal {
		t.Fatalf("expected Internal, got %v", err)
	}
	if _, err := store.GetChannel(context.Background(), wsTest, "ch1"); !errors.Is(err, configrepo.ErrNotFound) {
		t.Fatalf("expected rollback to remove channel, got %v", err)
	}
}

func TestAgentServiceServer_ReloadsRuntime(t *testing.T) {
	store := memory.New()
	svc := NewAgentServiceServer(store)
	runtime := &reloadTracker{}
	svc.SetRuntime(runtime)

	_, err := svc.CreateAgent(testCtx(), connect.NewRequest(&agentsv1.CreateAgentRequest{
		Agent: &agentsv1.Agent{Name: "a1"},
	}))
	if err != nil {
		t.Fatal(err)
	}
	if runtime.calls != 1 {
		t.Fatalf("expected 1 reload call, got %d", runtime.calls)
	}
}

func TestAgentServiceServer_ReloadErrorRollsBackCreateUpdateDelete(t *testing.T) {
	ctx := testCtx()

	t.Run("create", func(t *testing.T) {
		store := memory.New()
		svc := NewAgentServiceServer(store)
		svc.SetRuntime(&reloadTracker{err: errors.New("boom")})

		_, err := svc.CreateAgent(ctx, connect.NewRequest(&agentsv1.CreateAgentRequest{
			Agent: &agentsv1.Agent{Name: "a1"},
		}))
		if twerr, ok := err.(*connect.Error); !ok || twerr.Code() != connect.CodeInternal {
			t.Fatalf("expected Internal, got %v", err)
		}
		if _, err := store.GetAgent(ctx, wsTest, "a1"); !errors.Is(err, configrepo.ErrNotFound) {
			t.Fatalf("expected rollback to remove agent, got %v", err)
		}
	})

	t.Run("update", func(t *testing.T) {
		store := memory.New()
		if _, err := store.CreateAgent(ctx, wsTest, &agentsv1.Agent{Name: "a1", Description: "before"}); err != nil {
			t.Fatalf("seed agent: %v", err)
		}

		svc := NewAgentServiceServer(store)
		svc.SetRuntime(&reloadTracker{err: errors.New("boom")})

		_, err := svc.UpdateAgent(ctx, connect.NewRequest(&agentsv1.UpdateAgentRequest{
			Agent: &agentsv1.Agent{Name: "a1", Description: "after"},
		}))
		if twerr, ok := err.(*connect.Error); !ok || twerr.Code() != connect.CodeInternal {
			t.Fatalf("expected Internal, got %v", err)
		}

		agent, err := store.GetAgent(ctx, wsTest, "a1")
		if err != nil {
			t.Fatalf("get rolled back agent: %v", err)
		}
		if agent.GetDescription() != "before" {
			t.Fatalf("expected rollback to restore previous description, got %q", agent.GetDescription())
		}
	})

	t.Run("delete", func(t *testing.T) {
		store := memory.New()
		if _, err := store.CreateAgent(ctx, wsTest, &agentsv1.Agent{Name: "a1", Description: "before"}); err != nil {
			t.Fatalf("seed agent: %v", err)
		}

		svc := NewAgentServiceServer(store)
		svc.SetRuntime(&reloadTracker{err: errors.New("boom")})

		_, err := svc.DeleteAgent(ctx, connect.NewRequest(&agentsv1.DeleteAgentRequest{Name: "a1"}))
		if twerr, ok := err.(*connect.Error); !ok || twerr.Code() != connect.CodeInternal {
			t.Fatalf("expected Internal, got %v", err)
		}

		agent, err := store.GetAgent(ctx, wsTest, "a1")
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

	_, err := svc.CreateMCPServer(testCtx(), connect.NewRequest(&agentsv1.CreateMCPServerRequest{
		McpServer: testMCPServer("m1", "mcp1"),
	}))
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

	_, err := svc.CreateMCPServer(testCtx(), connect.NewRequest(&agentsv1.CreateMCPServerRequest{
		McpServer: testMCPServer("m1", "mcp1"),
	}))
	if twerr, ok := err.(*connect.Error); !ok || twerr.Code() != connect.CodeInternal {
		t.Fatalf("expected Internal, got %v", err)
	}
	if _, err := store.GetMCPServer(context.Background(), wsTest, "m1"); !errors.Is(err, configrepo.ErrNotFound) {
		t.Fatalf("expected rollback to remove mcp server, got %v", err)
	}
}

func TestRemoteAgentServiceServer_ReloadsRuntime(t *testing.T) {
	store := memory.New()
	svc := NewRemoteAgentServiceServer(store)
	runtime := &reloadTracker{}
	svc.SetRuntime(runtime)

	_, err := svc.CreateRemoteAgent(testCtx(), connect.NewRequest(&agentsv1.CreateRemoteAgentRequest{
		RemoteAgent: &agentsv1.RemoteAgent{Id: "r1", Name: "ra1", Url: "http://example.com"},
	}))
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

	_, err := svc.CreateRemoteAgent(testCtx(), connect.NewRequest(&agentsv1.CreateRemoteAgentRequest{
		RemoteAgent: &agentsv1.RemoteAgent{Id: "r1", Name: "ra1", Url: "http://example.com"},
	}))
	if twerr, ok := err.(*connect.Error); !ok || twerr.Code() != connect.CodeInternal {
		t.Fatalf("expected Internal, got %v", err)
	}
	if _, err := store.GetRemoteAgent(context.Background(), wsTest, "r1"); !errors.Is(err, configrepo.ErrNotFound) {
		t.Fatalf("expected rollback to remove remote agent, got %v", err)
	}
}
