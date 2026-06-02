package app

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"go.orx.me/apps/butter/internal/application"
	"go.orx.me/apps/butter/internal/repo/auth"
	"go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// These tests cover the GlobalMCPServerService behavior end-to-end against
// the in-memory config store. They previously asserted against REST endpoints
// at /api/(admin/)global-mcp-servers/*; those routes have been removed in
// favor of the Connect service, so the tests now invoke the service methods
// directly. The interesting invariants (admin-only mutation, client_secret
// redaction for non-admins, cross-workspace install boundary) are unchanged.

func newGlobalMCPTest(t *testing.T) (*application.GlobalMCPServerServiceServer, *ConfigStore) {
	t.Helper()
	store := NewConfigStore()
	mcpSvc := application.NewMCPServerServiceServer(store)
	return application.NewGlobalMCPServerServiceServer(store, mcpSvc), store
}

func manualOAuthPreset(id string) *agentsv1.MCPServer {
	return &agentsv1.MCPServer{
		Id:        id,
		Name:      "Manual OAuth",
		Url:       "https://mcp.example.com/mcp",
		Transport: agentsv1.MCPServerTransport_MCP_SERVER_TRANSPORT_STREAMABLE_HTTP,
		Auth: &agentsv1.MCPServerAuth{
			Type: agentsv1.MCPServerAuthType_MCP_SERVER_AUTH_TYPE_OAUTH2,
			Oauth2: &agentsv1.MCPServerOAuth2Config{
				ClientId:         "client-1",
				ClientSecret:     "secret-1",
				AuthorizationUrl: "https://auth.example.com/oauth/authorize",
				TokenUrl:         "https://auth.example.com/oauth/token",
				Scopes:           []string{"read"},
			},
		},
	}
}

func TestGlobalMCPAdminCRUD(t *testing.T) {
	svc, _ := newGlobalMCPTest(t)
	adminCtx := auth.WithAdmin(context.Background())

	createRes, err := svc.CreateGlobalMCPServer(adminCtx, connect.NewRequest(&agentsv1.CreateGlobalMCPServerRequest{
		McpServer: manualOAuthPreset("manual"),
	}))
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if got := createRes.Msg.GetMcpServer().GetAuth().GetOauth2().GetClientSecret(); got != "secret-1" {
		t.Fatalf("admin create response should include secret, got %q", got)
	}

	if _, err := svc.UpdateGlobalMCPServer(adminCtx, connect.NewRequest(&agentsv1.UpdateGlobalMCPServerRequest{
		McpServer: manualOAuthPreset("manual"),
	})); err != nil {
		t.Fatalf("update: %v", err)
	}

	if _, err := svc.DeleteGlobalMCPServer(adminCtx, connect.NewRequest(&agentsv1.DeleteGlobalMCPServerRequest{Id: "manual"})); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestGlobalMCPNonAdminMutationsRejected(t *testing.T) {
	svc, _ := newGlobalMCPTest(t)
	ctx := context.Background()

	if _, err := svc.CreateGlobalMCPServer(ctx, connect.NewRequest(&agentsv1.CreateGlobalMCPServerRequest{
		McpServer: manualOAuthPreset("manual"),
	})); err == nil {
		t.Fatal("expected non-admin create to be rejected")
	}
	if _, err := svc.UpdateGlobalMCPServer(ctx, connect.NewRequest(&agentsv1.UpdateGlobalMCPServerRequest{
		McpServer: manualOAuthPreset("manual"),
	})); err == nil {
		t.Fatal("expected non-admin update to be rejected")
	}
	if _, err := svc.DeleteGlobalMCPServer(ctx, connect.NewRequest(&agentsv1.DeleteGlobalMCPServerRequest{Id: "manual"})); err == nil {
		t.Fatal("expected non-admin delete to be rejected")
	}
}

func TestGlobalMCPListRedactsSecretsForNonAdmins(t *testing.T) {
	svc, store := newGlobalMCPTest(t)
	if _, err := store.CreateGlobalMCPServer(context.Background(), manualOAuthPreset("manual")); err != nil {
		t.Fatalf("seed: %v", err)
	}

	// Non-admin sees the preset but with the secret zeroed.
	res, err := svc.ListGlobalMCPServers(context.Background(), connect.NewRequest(&agentsv1.ListGlobalMCPServersRequest{}))
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if got := res.Msg.GetMcpServers()[0].GetAuth().GetOauth2().GetClientSecret(); got != "" {
		t.Fatalf("non-admin list leaked client_secret: %q", got)
	}

	// Admin retains the secret.
	adminRes, err := svc.ListGlobalMCPServers(auth.WithAdmin(context.Background()), connect.NewRequest(&agentsv1.ListGlobalMCPServersRequest{}))
	if err != nil {
		t.Fatalf("admin list: %v", err)
	}
	if got := adminRes.Msg.GetMcpServers()[0].GetAuth().GetOauth2().GetClientSecret(); got != "secret-1" {
		t.Fatalf("admin list dropped client_secret: got %q", got)
	}
}

func TestGlobalMCPInstallRedactsAndStoresSecret(t *testing.T) {
	svc, store := newGlobalMCPTest(t)
	if _, err := store.CreateGlobalMCPServer(context.Background(), manualOAuthPreset("manual")); err != nil {
		t.Fatalf("seed: %v", err)
	}

	wsCtx := workspace.WithID(context.Background(), "ws-a")
	res, err := svc.InstallGlobalMCPServer(wsCtx, connect.NewRequest(&agentsv1.InstallGlobalMCPServerRequest{Id: "manual"}))
	if err != nil {
		t.Fatalf("install: %v", err)
	}
	if got := res.Msg.GetMcpServer().GetAuth().GetOauth2().GetClientSecret(); got != "" {
		t.Fatalf("install response leaked client_secret: %q", got)
	}

	// The stored workspace copy keeps the secret so the OAuth flow can run;
	// only the dashboard-bound response is redacted.
	stored, err := store.GetMCPServer(context.Background(), "ws-a", "manual")
	if err != nil {
		t.Fatalf("get installed server: %v", err)
	}
	if stored.GetAuth().GetOauth2().GetClientSecret() != "secret-1" {
		t.Fatalf("stored workspace server should retain secret for OAuth flow")
	}
}

func TestGlobalMCPInstallWorkspaceBoundary(t *testing.T) {
	svc, store := newGlobalMCPTest(t)
	if _, err := store.CreateGlobalMCPServer(context.Background(), manualOAuthPreset("manual")); err != nil {
		t.Fatalf("seed: %v", err)
	}

	// Non-admin trying to install into a workspace they're not entering
	// gets rejected outright.
	wsACtx := workspace.WithID(context.Background(), "ws-a")
	if _, err := svc.InstallGlobalMCPServer(wsACtx, connect.NewRequest(&agentsv1.InstallGlobalMCPServerRequest{
		Id:          "manual",
		WorkspaceId: "ws-b",
	})); err == nil {
		t.Fatal("non-admin cross-workspace install should be rejected")
	}
	if _, err := store.GetMCPServer(context.Background(), "ws-b", "manual"); err == nil {
		t.Fatal("non-admin cross-workspace install actually created the server")
	}

	// Admin can install across workspaces (cross-workspace audit log fires
	// in the implementation; not asserted here).
	adminCtx := auth.WithAdmin(context.Background())
	if _, err := svc.InstallGlobalMCPServer(adminCtx, connect.NewRequest(&agentsv1.InstallGlobalMCPServerRequest{
		Id:          "manual",
		WorkspaceId: "ws-b",
	})); err != nil {
		t.Fatalf("admin cross-workspace install: %v", err)
	}
	if _, err := store.GetMCPServer(context.Background(), "ws-b", "manual"); err != nil {
		t.Fatalf("admin cross-workspace install missing server: %v", err)
	}
}
