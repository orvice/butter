package application

import (
	"strings"
	"testing"

	"connectrpc.com/connect"
	"go.orx.me/apps/butter/internal/mcpoauth"
	"go.orx.me/apps/butter/internal/repo/config/memory"
	mcpoauthmemory "go.orx.me/apps/butter/internal/repo/mcpoauth/memory"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestMCPServerServiceServer_OAuthValidation(t *testing.T) {
	store := memory.New()
	svc := NewMCPServerServiceServer(store)
	_, err := svc.CreateMCPServer(testCtx(), &agentsv1.CreateMCPServerRequest{McpServer: &agentsv1.MCPServer{
		Id:        "oauth-unspecified",
		Name:      "OAuth Unspecified",
		Transport: agentsv1.MCPServerTransport_MCP_SERVER_TRANSPORT_UNSPECIFIED,
		Auth:      oauthAuth("https://issuer.example.com/authorize", "https://issuer.example.com/token"),
	}})
	if twerr, ok := err.(*connect.Error); !ok || twerr.Code() != connect.CodeInvalidArgument {
		t.Fatalf("expected invalid argument for unsupported transport, got %v", err)
	}

	_, err = svc.CreateMCPServer(testCtx(), &agentsv1.CreateMCPServerRequest{McpServer: &agentsv1.MCPServer{
		Id:        "oauth-http",
		Name:      "OAuth HTTP",
		Transport: agentsv1.MCPServerTransport_MCP_SERVER_TRANSPORT_STREAMABLE_HTTP,
		Url:       "https://mcp.example.com/mcp",
		Auth:      oauthAuth("https://issuer.example.com/authorize", "https://issuer.example.com/token"),
	}})
	if err != nil {
		t.Fatalf("expected remote oauth server to validate: %v", err)
	}
}

func TestMCPServerServiceServer_OAuthStatusAndDisconnect(t *testing.T) {
	store := memory.New()
	svc := NewMCPServerServiceServer(store)
	oauthRepo := mcpoauthmemory.New()
	svc.SetOAuthService(mcpoauth.NewService(oauthRepo, mcpoauth.NewMemoryFlowStore(), func() mcpoauth.Config {
		return mcpoauth.Config{CallbackBaseURL: "http://127.0.0.1:8080", EncryptionKey: "0123456789abcdef0123456789abcdef"}
	}))
	_, err := svc.CreateMCPServer(testCtx(), &agentsv1.CreateMCPServerRequest{McpServer: &agentsv1.MCPServer{
		Id:        "oauth-http",
		Name:      "OAuth HTTP",
		Transport: agentsv1.MCPServerTransport_MCP_SERVER_TRANSPORT_STREAMABLE_HTTP,
		Url:       "https://mcp.example.com/mcp",
		Auth:      oauthAuth("https://issuer.example.com/authorize", "https://issuer.example.com/token"),
	}})
	if err != nil {
		t.Fatalf("create server: %v", err)
	}
	status, err := svc.GetMCPServerOAuthStatus(testCtx(), &agentsv1.GetMCPServerOAuthStatusRequest{ServerId: "oauth-http"})
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if status.GetStatus().GetState() != agentsv1.MCPOAuthConnectionState_MCPO_AUTH_CONNECTION_STATE_DISCONNECTED {
		t.Fatalf("expected disconnected, got %v", status.GetStatus().GetState())
	}
	if _, err := svc.DisconnectMCPServerOAuth(testCtx(), &agentsv1.DisconnectMCPServerOAuthRequest{ServerId: "oauth-http"}); err != nil {
		t.Fatalf("disconnect: %v", err)
	}
	if _, err := svc.GetMCPServer(testCtx(), &agentsv1.GetMCPServerRequest{Id: "oauth-http"}); err != nil {
		t.Fatalf("disconnect deleted MCP config: %v", err)
	}
}

func TestMCPServerServiceServer_ListToolsOAuthMissingConnection(t *testing.T) {
	store := memory.New()
	svc := NewMCPServerServiceServer(store)
	oauthRepo := mcpoauthmemory.New()
	resolver := mcpoauth.NewResolver(oauthRepo, func() mcpoauth.Config {
		return mcpoauth.Config{EncryptionKey: "0123456789abcdef0123456789abcdef"}
	})
	svc.SetMCPHTTPClientFactory(resolver)
	_, err := svc.CreateMCPServer(testCtx(), &agentsv1.CreateMCPServerRequest{McpServer: &agentsv1.MCPServer{
		Id:        "oauth-http",
		Name:      "OAuth HTTP",
		Transport: agentsv1.MCPServerTransport_MCP_SERVER_TRANSPORT_STREAMABLE_HTTP,
		Url:       "http://127.0.0.1:1/mcp",
		Auth:      oauthAuth("https://issuer.example.com/authorize", "https://issuer.example.com/token"),
	}})
	if err != nil {
		t.Fatalf("create server: %v", err)
	}
	resp, err := svc.ListMCPTools(testCtx(), &agentsv1.ListMCPToolsRequest{ServerId: "oauth-http"})
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}
	if got := resp.GetErrors()["oauth-http"]; !strings.Contains(got, "authorization") {
		t.Fatalf("expected authorization error, got %q", got)
	}
}

func oauthAuth(authURL, tokenURL string) *agentsv1.MCPServerAuth {
	return &agentsv1.MCPServerAuth{
		Type: agentsv1.MCPServerAuthType_MCP_SERVER_AUTH_TYPE_OAUTH2,
		Oauth2: &agentsv1.MCPServerOAuth2Config{
			ClientId:         "client",
			AuthorizationUrl: authURL,
			TokenUrl:         tokenURL,
		},
	}
}
