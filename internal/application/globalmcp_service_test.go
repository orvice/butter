package application

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"go.orx.me/apps/butter/internal/repo/auth"
	"go.orx.me/apps/butter/internal/repo/config/memory"
	"go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func adminCtx() context.Context {
	return auth.WithAdmin(context.Background())
}

func TestGlobalMCPServerServiceServer_ValidationRejectsInvalidPreset(t *testing.T) {
	store := memory.New()
	svc := NewGlobalMCPServerServiceServer(store, NewMCPServerServiceServer(store))

	_, err := svc.CreateGlobalMCPServer(adminCtx(), connect.NewRequest(&agentsv1.CreateGlobalMCPServerRequest{
		McpServer: &agentsv1.MCPServer{
			Id:        "preset",
			Name:      "preset",
			Transport: agentsv1.MCPServerTransport_MCP_SERVER_TRANSPORT_STREAMABLE_HTTP,
		},
	}))
	if twerr, ok := err.(*connect.Error); !ok || twerr.Code() != connect.CodeInvalidArgument {
		t.Fatalf("expected InvalidArgument for missing URL, got %v", err)
	}

	_, err = svc.CreateGlobalMCPServer(adminCtx(), connect.NewRequest(&agentsv1.CreateGlobalMCPServerRequest{
		McpServer: testMCPServer("preset", "preset"),
	}))
	if err != nil {
		t.Fatal(err)
	}

	_, err = svc.UpdateGlobalMCPServer(adminCtx(), connect.NewRequest(&agentsv1.UpdateGlobalMCPServerRequest{
		McpServer: &agentsv1.MCPServer{
			Id:        "preset",
			Name:      "preset",
			Transport: agentsv1.MCPServerTransport_MCP_SERVER_TRANSPORT_UNSPECIFIED,
			Url:       "https://mcp.example.com/mcp",
		},
	}))
	if twerr, ok := err.(*connect.Error); !ok || twerr.Code() != connect.CodeInvalidArgument {
		t.Fatalf("expected InvalidArgument for unsupported transport, got %v", err)
	}
}

func TestGlobalMCPServerServiceServer_InstallReinstallUpdatesInstalledPreset(t *testing.T) {
	store := memory.New()
	mcpSvc := NewMCPServerServiceServer(store)
	runtime := &reloadTracker{}
	mcpSvc.SetRuntime(runtime)
	svc := NewGlobalMCPServerServiceServer(store, mcpSvc)

	_, err := svc.CreateGlobalMCPServer(adminCtx(), connect.NewRequest(&agentsv1.CreateGlobalMCPServerRequest{
		McpServer: testMCPServer("preset", "old name"),
	}))
	if err != nil {
		t.Fatal(err)
	}

	ctx := workspace.WithID(context.Background(), wsTest)
	first, err := svc.InstallGlobalMCPServer(ctx, connect.NewRequest(&agentsv1.InstallGlobalMCPServerRequest{Id: "preset"}))
	if err != nil {
		t.Fatal(err)
	}
	if first.Msg.GetMcpServer().GetName() != "old name" {
		t.Fatalf("expected initial install name old name, got %s", first.Msg.GetMcpServer().GetName())
	}

	_, err = svc.UpdateGlobalMCPServer(adminCtx(), connect.NewRequest(&agentsv1.UpdateGlobalMCPServerRequest{
		McpServer: testMCPServer("preset", "new name"),
	}))
	if err != nil {
		t.Fatal(err)
	}

	second, err := svc.InstallGlobalMCPServer(ctx, connect.NewRequest(&agentsv1.InstallGlobalMCPServerRequest{Id: "preset"}))
	if err != nil {
		t.Fatal(err)
	}
	if second.Msg.GetMcpServer().GetName() != "new name" {
		t.Fatalf("expected reinstall to update name to new name, got %s", second.Msg.GetMcpServer().GetName())
	}
	if second.Msg.GetMcpServer().GetMetadata()[globalMCPPresetMetadataKey] != "preset" {
		t.Fatalf("expected installed preset metadata, got %v", second.Msg.GetMcpServer().GetMetadata())
	}
	if runtime.calls != 2 {
		t.Fatalf("expected create and update reloads, got %d", runtime.calls)
	}
}

func TestGlobalMCPServerServiceServer_InstallRejectsWorkspaceIDCollision(t *testing.T) {
	store := memory.New()
	mcpSvc := NewMCPServerServiceServer(store)
	svc := NewGlobalMCPServerServiceServer(store, mcpSvc)

	_, err := svc.CreateGlobalMCPServer(adminCtx(), connect.NewRequest(&agentsv1.CreateGlobalMCPServerRequest{
		McpServer: testMCPServer("preset", "preset"),
	}))
	if err != nil {
		t.Fatal(err)
	}

	ctx := workspace.WithID(context.Background(), wsTest)
	_, err = mcpSvc.CreateMCPServer(ctx, connect.NewRequest(&agentsv1.CreateMCPServerRequest{
		McpServer: testMCPServer("preset", "workspace server"),
	}))
	if err != nil {
		t.Fatal(err)
	}

	_, err = svc.InstallGlobalMCPServer(ctx, connect.NewRequest(&agentsv1.InstallGlobalMCPServerRequest{Id: "preset"}))
	if twerr, ok := err.(*connect.Error); !ok || twerr.Code() != connect.CodeAlreadyExists {
		t.Fatalf("expected AlreadyExists for workspace id collision, got %v", err)
	}
}
