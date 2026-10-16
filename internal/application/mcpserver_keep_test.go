package application

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"go.orx.me/apps/butter/internal/repo/config/memory"
	"go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// installedPreset installs a preset that holds credentials into wsTest and
// returns what a member reads of it (hidden) together with the services.
func installedPreset(t *testing.T) (*MCPServerServiceServer, context.Context, *agentsv1.MCPServer) {
	t.Helper()
	store := memory.New()
	mcpSvc := NewMCPServerServiceServer(store)
	global := NewGlobalMCPServerServiceServer(store, mcpSvc)

	preset := testMCPServer("preset", "preset")
	preset.Headers = map[string]string{"Authorization": "Bearer top-secret", "X-Team": "alpha"}
	preset.Auth = &agentsv1.MCPServerAuth{
		Type:   agentsv1.MCPServerAuthType_MCP_SERVER_AUTH_TYPE_OAUTH2,
		Oauth2: &agentsv1.MCPServerOAuth2Config{ClientId: "cid", ClientSecret: "csecret"},
	}
	if _, err := global.CreateGlobalMCPServer(adminCtx(), connect.NewRequest(&agentsv1.CreateGlobalMCPServerRequest{McpServer: preset})); err != nil {
		t.Fatal(err)
	}
	ctx := workspace.WithID(context.Background(), wsTest)
	installed, err := global.InstallGlobalMCPServer(ctx, connect.NewRequest(&agentsv1.InstallGlobalMCPServerRequest{Id: "preset"}))
	if err != nil {
		t.Fatal(err)
	}
	return mcpSvc, ctx, installed.Msg.GetMcpServer()
}

func storedServer(t *testing.T, svc *MCPServerServiceServer, id string) *agentsv1.MCPServer {
	t.Helper()
	got, err := svc.repo.GetMCPServer(context.Background(), wsTest, id)
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func TestUpdateMCPServer_KeepsHiddenCredentialsOfAnInstalledPreset(t *testing.T) {
	svc, ctx, hidden := installedPreset(t)
	edited := hidden
	edited.Name = "Renamed"
	edited.TimeoutSeconds = 45
	edited.ToolFilter = []string{"echo"}
	delete(edited.Headers, "Authorization")
	delete(edited.Headers, "X-Team")
	edited.Headers["X-Team"] = "beta" // an explicit replacement of a visible-by-key header
	edited.Auth.Oauth2.ClientSecret = ""

	resp, err := svc.UpdateMCPServerKeepingCredentials(ctx, connect.NewRequest(&agentsv1.UpdateMCPServerKeepingCredentialsRequest{
		McpServer:             edited,
		KeepHeaderKeys:        []string{"Authorization"},
		KeepOauthClientSecret: true,
	}))
	if err != nil {
		t.Fatal(err)
	}
	stored := storedServer(t, svc, hidden.GetId())
	if stored.GetName() != "Renamed" || stored.GetTimeoutSeconds() != 45 || len(stored.GetToolFilter()) != 1 {
		t.Fatalf("explicit edits were not applied: %v", stored)
	}
	if stored.GetHeaders()["Authorization"] != "Bearer top-secret" {
		t.Fatalf("kept header lost: %v", stored.GetHeaders())
	}
	if stored.GetHeaders()["X-Team"] != "beta" {
		t.Fatalf("explicit header replacement lost: %v", stored.GetHeaders())
	}
	if stored.GetAuth().GetOauth2().GetClientSecret() != "csecret" {
		t.Fatalf("kept client secret lost: %q", stored.GetAuth().GetOauth2().GetClientSecret())
	}
	if stored.GetMetadata()[globalMCPPresetMetadataKey] != "preset" {
		t.Fatalf("preset marker lost: %v", stored.GetMetadata())
	}
	// Nothing the caller could not read comes back.
	got := resp.Msg.GetMcpServer()
	for k, v := range got.GetHeaders() {
		if v != redactedHeaderValue {
			t.Fatalf("response leaks header %q = %q", k, v)
		}
	}
	if got.GetAuth().GetOauth2().GetClientSecret() != "" {
		t.Fatalf("response leaks the client secret")
	}
}

func TestUpdateMCPServer_KeepCannotBeUsedToStripTheMarkerAndReadSecretsBack(t *testing.T) {
	svc, ctx, hidden := installedPreset(t)
	edited := hidden
	edited.Metadata = nil // a hostile or careless client drops the marker
	delete(edited.Headers, "Authorization")
	if _, err := svc.UpdateMCPServerKeepingCredentials(ctx, connect.NewRequest(&agentsv1.UpdateMCPServerKeepingCredentialsRequest{
		McpServer: edited, KeepHeaderKeys: []string{"Authorization"},
	})); err != nil {
		t.Fatal(err)
	}
	read, err := svc.GetMCPServer(ctx, connect.NewRequest(&agentsv1.GetMCPServerRequest{Id: hidden.GetId()}))
	if err != nil {
		t.Fatal(err)
	}
	if h := read.Msg.GetMcpServer().GetHeaders()["Authorization"]; h != redactedHeaderValue {
		t.Fatalf("a later read leaks the kept header: %q", h)
	}
}

func TestUpdateMCPServer_ExplicitReplaceAndClearApplyWithTheOrdinaryRequest(t *testing.T) {
	svc, ctx, hidden := installedPreset(t)
	edited := hidden
	delete(edited.Headers, "Authorization") // removed: neither kept nor given
	edited.Headers["X-Team"] = "gamma"
	edited.Auth.Oauth2.ClientSecret = "fresh-secret" // replaced

	if _, err := svc.UpdateMCPServer(ctx, connect.NewRequest(&agentsv1.UpdateMCPServerRequest{McpServer: edited})); err != nil {
		t.Fatal(err)
	}
	stored := storedServer(t, svc, hidden.GetId())
	if _, ok := stored.GetHeaders()["Authorization"]; ok {
		t.Fatalf("an omitted header must be removed, as before: %v", stored.GetHeaders())
	}
	if stored.GetHeaders()["X-Team"] != "gamma" || stored.GetAuth().GetOauth2().GetClientSecret() != "fresh-secret" {
		t.Fatalf("explicit values not stored: %v", stored)
	}
}

func TestUpdateMCPServer_OldWholeObjectRequestBehavesAsBefore(t *testing.T) {
	svc, ctx, hidden := installedPreset(t)
	// The pre-seam client round-trips the masks it was shown. Without a keep
	// field the server stores exactly what it is sent and never guesses.
	resp, err := svc.UpdateMCPServer(ctx, connect.NewRequest(&agentsv1.UpdateMCPServerRequest{McpServer: hidden}))
	if err != nil {
		t.Fatal(err)
	}
	stored := storedServer(t, svc, hidden.GetId())
	if stored.GetHeaders()["Authorization"] != redactedHeaderValue {
		t.Fatalf("whole-object semantics changed: %v", stored.GetHeaders())
	}
	if resp.Msg.GetMcpServer().GetHeaders()["Authorization"] != redactedHeaderValue {
		t.Fatalf("unexpected answer: %v", resp.Msg.GetMcpServer().GetHeaders())
	}
}

func TestUpdateMCPServer_KeepRefusals(t *testing.T) {
	cases := map[string]func(*agentsv1.MCPServer) *agentsv1.UpdateMCPServerKeepingCredentialsRequest{
		"a header that is not stored": func(s *agentsv1.MCPServer) *agentsv1.UpdateMCPServerKeepingCredentialsRequest {
			return &agentsv1.UpdateMCPServerKeepingCredentialsRequest{McpServer: s, KeepHeaderKeys: []string{"X-Missing"}}
		},
		"a header both kept and given": func(s *agentsv1.MCPServer) *agentsv1.UpdateMCPServerKeepingCredentialsRequest {
			return &agentsv1.UpdateMCPServerKeepingCredentialsRequest{McpServer: s, KeepHeaderKeys: []string{"Authorization"}}
		},
		"a secret kept for a changed client": func(s *agentsv1.MCPServer) *agentsv1.UpdateMCPServerKeepingCredentialsRequest {
			s.Auth.Oauth2.ClientId = "other"
			s.Auth.Oauth2.ClientSecret = ""
			return &agentsv1.UpdateMCPServerKeepingCredentialsRequest{McpServer: s, KeepOauthClientSecret: true}
		},
		"a secret both kept and given": func(s *agentsv1.MCPServer) *agentsv1.UpdateMCPServerKeepingCredentialsRequest {
			s.Auth.Oauth2.ClientSecret = "x"
			return &agentsv1.UpdateMCPServerKeepingCredentialsRequest{McpServer: s, KeepOauthClientSecret: true}
		},
		"a secret kept without OAuth": func(s *agentsv1.MCPServer) *agentsv1.UpdateMCPServerKeepingCredentialsRequest {
			s.Auth = &agentsv1.MCPServerAuth{Type: agentsv1.MCPServerAuthType_MCP_SERVER_AUTH_TYPE_NONE}
			return &agentsv1.UpdateMCPServerKeepingCredentialsRequest{McpServer: s, KeepOauthClientSecret: true}
		},
	}
	for name, build := range cases {
		t.Run(name, func(t *testing.T) {
			svc, ctx, hidden := installedPreset(t)
			before := storedServer(t, svc, hidden.GetId())
			_, err := svc.UpdateMCPServerKeepingCredentials(ctx, connect.NewRequest(build(hidden)))
			if connect.CodeOf(err) != connect.CodeInvalidArgument {
				t.Fatalf("code = %v (%v)", connect.CodeOf(err), err)
			}
			after := storedServer(t, svc, hidden.GetId())
			if after.GetHeaders()["Authorization"] != before.GetHeaders()["Authorization"] || after.GetName() != before.GetName() {
				t.Fatalf("a refused update changed the server")
			}
		})
	}
}

func TestUpdateMCPServer_KeepNeverReadsAnotherWorkspace(t *testing.T) {
	svc, _, hidden := installedPreset(t)
	other := workspace.WithID(context.Background(), "ws-other")
	_, err := svc.UpdateMCPServerKeepingCredentials(other, connect.NewRequest(&agentsv1.UpdateMCPServerKeepingCredentialsRequest{
		McpServer: hidden, KeepHeaderKeys: []string{"Authorization"},
	}))
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("code = %v (%v)", connect.CodeOf(err), err)
	}
}

// The ordinary request is byte-for-byte what it was: an older server and the
// Web console never see a keep field, so keeping credentials cannot silently
// degrade into a whole-object write on a server that ignores unknown fields.
func TestUpdateMCPServerRequestCarriesNoKeepFields(t *testing.T) {
	fields := (&agentsv1.UpdateMCPServerRequest{}).ProtoReflect().Descriptor().Fields()
	if fields.Len() != 1 || fields.ByName("mcp_server") == nil {
		t.Fatalf("UpdateMCPServerRequest must stay {mcp_server}; fields = %d", fields.Len())
	}
}

func TestUpdateMCPServerKeepingCredentials_WithoutKeepsStoresTheWholeObject(t *testing.T) {
	svc, ctx, hidden := installedPreset(t)
	edited := hidden
	delete(edited.Headers, "Authorization")
	edited.Auth.Oauth2.ClientSecret = ""
	resp, err := svc.UpdateMCPServerKeepingCredentials(ctx, connect.NewRequest(&agentsv1.UpdateMCPServerKeepingCredentialsRequest{McpServer: edited}))
	if err != nil {
		t.Fatal(err)
	}
	stored := storedServer(t, svc, hidden.GetId())
	if _, ok := stored.GetHeaders()["Authorization"]; ok {
		t.Fatalf("an unkept header must be removed: %v", stored.GetHeaders())
	}
	if stored.GetAuth().GetOauth2().GetClientSecret() != "" {
		t.Fatalf("an unkept secret must be cleared")
	}
	if got := resp.Msg.GetMcpServer().GetHeaders()["X-Team"]; got != redactedHeaderValue {
		t.Fatalf("the answer must stay redacted: %q", got)
	}
}

func TestUpdateMCPServerKeepingCredentials_RequiresAServer(t *testing.T) {
	svc, ctx, _ := installedPreset(t)
	_, err := svc.UpdateMCPServerKeepingCredentials(ctx, connect.NewRequest(&agentsv1.UpdateMCPServerKeepingCredentialsRequest{}))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("code = %v (%v)", connect.CodeOf(err), err)
	}
}
