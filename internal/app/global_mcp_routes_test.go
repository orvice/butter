package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"go.orx.me/apps/butter/internal/application"
	"go.orx.me/apps/butter/internal/repo/auth"
	"go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func newGlobalMCPTestRouter(t *testing.T) (*gin.Engine, *Handlers) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	store := NewConfigStore()
	mcpSvc := application.NewMCPServerServiceServer(store)
	handlers := &Handlers{
		configStore:         store,
		globalMCPServerRepo: store,
	}
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		ctx := c.Request.Context()
		if c.GetHeader("X-Test-Admin") == "true" {
			ctx = auth.WithAdmin(ctx)
		}
		if ws := c.GetHeader(workspace.HeaderName); ws != "" {
			ctx = workspace.WithID(ctx, ws)
		}
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	registerGlobalMCPServerRoutes(engine, handlers, mcpSvc)
	return engine, handlers
}

func manualOAuthPresetPayload(id string) []byte {
	return []byte(`{
		"id": "` + id + `",
		"name": "Manual OAuth",
		"url": "https://mcp.example.com/mcp",
		"transport": "MCP_SERVER_TRANSPORT_STREAMABLE_HTTP",
		"auth": {
			"type": "MCP_SERVER_AUTH_TYPE_OAUTH2",
			"oauth2": {
				"client_id": "client-1",
				"client_secret": "secret-1",
				"scopes": ["read"],
				"authorization_url": "https://auth.example.com/oauth/authorize",
				"token_url": "https://auth.example.com/oauth/token"
			}
		}
	}`)
}

func doJSON(r http.Handler, method, path string, body []byte, workspaceID string) *httptest.ResponseRecorder {
	return doJSONWithAdmin(r, method, path, body, workspaceID, false)
}

func doJSONWithAdmin(r http.Handler, method, path string, body []byte, workspaceID string, admin bool) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if workspaceID != "" {
		req.Header.Set(workspace.HeaderName, workspaceID)
	}
	if admin {
		req.Header.Set("X-Test-Admin", "true")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestGlobalMCPAdminCRUDUsesProtoJSONEnums(t *testing.T) {
	r, _ := newGlobalMCPTestRouter(t)

	w := doJSONWithAdmin(r, http.MethodPost, "/api/admin/global-mcp-servers", manualOAuthPresetPayload("manual"), "", true)
	if w.Code != http.StatusCreated {
		t.Fatalf("create status = %d body=%s", w.Code, w.Body.String())
	}

	var created struct {
		MCPServer struct {
			ID        string `json:"id"`
			Transport string `json:"transport"`
			Auth      struct {
				Type   string `json:"type"`
				OAuth2 struct {
					ClientSecret string `json:"client_secret"`
				} `json:"oauth2"`
			} `json:"auth"`
		} `json:"mcp_server"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if created.MCPServer.Transport != "MCP_SERVER_TRANSPORT_STREAMABLE_HTTP" {
		t.Fatalf("transport should be proto enum string, got %q", created.MCPServer.Transport)
	}
	if created.MCPServer.Auth.Type != "MCP_SERVER_AUTH_TYPE_OAUTH2" {
		t.Fatalf("auth type should be proto enum string, got %q", created.MCPServer.Auth.Type)
	}
	if created.MCPServer.Auth.OAuth2.ClientSecret != "secret-1" {
		t.Fatalf("admin response should include secret, got %q", created.MCPServer.Auth.OAuth2.ClientSecret)
	}

	updatePayload := manualOAuthPresetPayload("manual")
	w = doJSONWithAdmin(r, http.MethodPut, "/api/admin/global-mcp-servers/manual", updatePayload, "", true)
	if w.Code != http.StatusOK {
		t.Fatalf("update status = %d body=%s", w.Code, w.Body.String())
	}

	w = doJSONWithAdmin(r, http.MethodDelete, "/api/admin/global-mcp-servers/manual", nil, "", true)
	if w.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d body=%s", w.Code, w.Body.String())
	}
}

func TestGlobalMCPListAndInstallRedactSecretsForWorkspaceUsers(t *testing.T) {
	r, handlers := newGlobalMCPTestRouter(t)
	ctx := httptest.NewRequest(http.MethodGet, "/", nil).Context()
	preset := &agentsv1.MCPServer{
		Id:        "manual",
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
			},
		},
	}
	if _, err := handlers.configStore.CreateGlobalMCPServer(ctx, preset); err != nil {
		t.Fatalf("seed global preset: %v", err)
	}

	w := doJSON(r, http.MethodGet, "/api/global-mcp-servers", nil, "ws-a")
	if w.Code != http.StatusOK {
		t.Fatalf("list status = %d body=%s", w.Code, w.Body.String())
	}
	if bytes.Contains(w.Body.Bytes(), []byte("secret-1")) {
		t.Fatalf("non-admin list leaked client secret: %s", w.Body.String())
	}

	w = doJSON(r, http.MethodPost, "/api/global-mcp-servers/manual/install", []byte(`{}`), "ws-a")
	if w.Code != http.StatusCreated {
		t.Fatalf("install status = %d body=%s", w.Code, w.Body.String())
	}
	if bytes.Contains(w.Body.Bytes(), []byte("secret-1")) {
		t.Fatalf("install response leaked client secret: %s", w.Body.String())
	}
	stored, err := handlers.configStore.GetMCPServer(ctx, "ws-a", "manual")
	if err != nil {
		t.Fatalf("get installed server: %v", err)
	}
	if stored.GetAuth().GetOauth2().GetClientSecret() != "secret-1" {
		t.Fatalf("stored workspace server should retain secret for OAuth flow")
	}
}

func TestGlobalMCPInstallThenWorkspaceListGetRedactsAdminOwnedSecret(t *testing.T) {
	r, handlers := newGlobalMCPTestRouter(t)
	ctx := httptest.NewRequest(http.MethodGet, "/", nil).Context()
	preset := &agentsv1.MCPServer{
		Id:        "manual",
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
			},
		},
	}
	if _, err := handlers.configStore.CreateGlobalMCPServer(ctx, preset); err != nil {
		t.Fatalf("seed global preset: %v", err)
	}

	w := doJSON(r, http.MethodPost, "/api/global-mcp-servers/manual/install", []byte(`{}`), "ws-a")
	if w.Code != http.StatusCreated {
		t.Fatalf("install status = %d body=%s", w.Code, w.Body.String())
	}

	stored, err := handlers.configStore.GetMCPServer(ctx, "ws-a", "manual")
	if err != nil {
		t.Fatalf("get installed server from store: %v", err)
	}
	if stored.GetAuth().GetOauth2().GetClientSecret() != "secret-1" {
		t.Fatalf("stored workspace server should retain secret for OAuth flow")
	}

	mcpSvc := application.NewMCPServerServiceServer(handlers.configStore)
	workspaceCtx := workspace.WithID(context.Background(), "ws-a")
	listResp, err := mcpSvc.ListMCPServers(workspaceCtx, &agentsv1.ListMCPServersRequest{})
	if err != nil {
		t.Fatalf("list workspace mcp servers: %v", err)
	}
	if got := listResp.GetMcpServers()[0].GetAuth().GetOauth2().GetClientSecret(); got != "" {
		t.Fatalf("workspace list leaked installed global client secret: %q", got)
	}

	getResp, err := mcpSvc.GetMCPServer(workspaceCtx, &agentsv1.GetMCPServerRequest{Id: "manual"})
	if err != nil {
		t.Fatalf("get workspace mcp server: %v", err)
	}
	if got := getResp.GetMcpServer().GetAuth().GetOauth2().GetClientSecret(); got != "" {
		t.Fatalf("workspace get leaked installed global client secret: %q", got)
	}
}

func TestGlobalMCPInstallWorkspaceBoundary(t *testing.T) {
	r, handlers := newGlobalMCPTestRouter(t)
	ctx := httptest.NewRequest(http.MethodGet, "/", nil).Context()
	preset := &agentsv1.MCPServer{
		Id:        "manual",
		Name:      "Manual OAuth",
		Url:       "https://mcp.example.com/mcp",
		Transport: agentsv1.MCPServerTransport_MCP_SERVER_TRANSPORT_STREAMABLE_HTTP,
		Auth: &agentsv1.MCPServerAuth{
			Type:   agentsv1.MCPServerAuthType_MCP_SERVER_AUTH_TYPE_OAUTH2,
			Oauth2: &agentsv1.MCPServerOAuth2Config{ClientId: "client-1"},
		},
	}
	if _, err := handlers.configStore.CreateGlobalMCPServer(ctx, preset); err != nil {
		t.Fatalf("seed global preset: %v", err)
	}

	w := doJSON(r, http.MethodPost, "/api/global-mcp-servers/manual/install", []byte(`{"workspace_id":"ws-b"}`), "ws-a")
	if w.Code != http.StatusForbidden {
		t.Fatalf("non-admin cross-workspace install status = %d body=%s", w.Code, w.Body.String())
	}
	if _, err := handlers.configStore.GetMCPServer(ctx, "ws-b", "manual"); err == nil {
		t.Fatal("non-admin cross-workspace install created server")
	}

	w = doJSONWithAdmin(r, http.MethodPost, "/api/global-mcp-servers/manual/install", []byte(`{"workspace_id":"ws-b"}`), "", true)
	if w.Code != http.StatusCreated {
		t.Fatalf("admin cross-workspace install status = %d body=%s", w.Code, w.Body.String())
	}
	if _, err := handlers.configStore.GetMCPServer(ctx, "ws-b", "manual"); err != nil {
		t.Fatalf("admin cross-workspace install missing server: %v", err)
	}
}
