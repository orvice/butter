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

// A preset's static headers are credentials ("Authorization: Bearer …"). A
// non-admin reads presets to browse and install them, so neither the listing
// nor the install answer may carry the values; the admin who authored the
// preset still sees what they wrote. (mobile #454)
func TestGlobalMCPServerServiceServer_HidesHeaderValuesFromNonAdmins(t *testing.T) {
	store := memory.New()
	mcpSvc := NewMCPServerServiceServer(store)
	svc := NewGlobalMCPServerServiceServer(store, mcpSvc)

	preset := testMCPServer("preset", "preset")
	preset.Headers = map[string]string{"Authorization": "Bearer top-secret"}
	preset.Auth = &agentsv1.MCPServerAuth{Oauth2: &agentsv1.MCPServerOAuth2Config{ClientId: "cid", ClientSecret: "csecret"}}
	if _, err := svc.CreateGlobalMCPServer(adminCtx(), connect.NewRequest(&agentsv1.CreateGlobalMCPServerRequest{McpServer: preset})); err != nil {
		t.Fatal(err)
	}

	memberCtx := workspace.WithID(context.Background(), wsTest)
	listed, err := svc.ListGlobalMCPServers(memberCtx, connect.NewRequest(&agentsv1.ListGlobalMCPServersRequest{}))
	if err != nil {
		t.Fatal(err)
	}
	got := listed.Msg.GetMcpServers()
	if len(got) != 1 || got[0].GetHeaders()["Authorization"] != redactedHeaderValue {
		t.Fatalf("member listing headers = %v", got)
	}
	if got[0].GetAuth().GetOauth2().GetClientSecret() != "" || got[0].GetAuth().GetOauth2().GetClientId() != "cid" {
		t.Fatalf("member listing oauth = %v", got[0].GetAuth())
	}

	installed, err := svc.InstallGlobalMCPServer(memberCtx, connect.NewRequest(&agentsv1.InstallGlobalMCPServerRequest{Id: "preset"}))
	if err != nil {
		t.Fatal(err)
	}
	if h := installed.Msg.GetMcpServer().GetHeaders()["Authorization"]; h != redactedHeaderValue {
		t.Fatalf("install answer header = %q", h)
	}

	adminListed, err := svc.ListGlobalMCPServers(adminCtx(), connect.NewRequest(&agentsv1.ListGlobalMCPServersRequest{}))
	if err != nil {
		t.Fatal(err)
	}
	if h := adminListed.Msg.GetMcpServers()[0].GetHeaders()["Authorization"]; h != "Bearer top-secret" {
		t.Fatalf("admin must still see what they wrote, got %q", h)
	}
}

// Installing a preset that no longer exists must say so (NotFound), so a
// client can tell "nothing was installed" from a server failure. (mobile #454)
func TestGlobalMCPServerServiceServer_InstallMissingPresetIsNotFound(t *testing.T) {
	store := memory.New()
	svc := NewGlobalMCPServerServiceServer(store, NewMCPServerServiceServer(store))
	ctx := workspace.WithID(context.Background(), wsTest)
	_, err := svc.InstallGlobalMCPServer(ctx, connect.NewRequest(&agentsv1.InstallGlobalMCPServerRequest{Id: "gone"}))
	if twerr, ok := err.(*connect.Error); !ok || twerr.Code() != connect.CodeNotFound {
		t.Fatalf("expected NotFound, got %v", err)
	}
}
