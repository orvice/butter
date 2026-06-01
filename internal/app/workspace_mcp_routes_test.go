package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/application"
	"go.orx.me/apps/butter/internal/config"
	agentfilememory "go.orx.me/apps/butter/internal/repo/agentfile/memory"
	apitokenmemory "go.orx.me/apps/butter/internal/repo/apitoken/memory"
	workspacememory "go.orx.me/apps/butter/internal/repo/workspace/memory"
	"go.orx.me/apps/butter/internal/runtime/daemon"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type staticOAuthHandler struct {
	token string
	base  http.RoundTripper
}

func (h staticOAuthHandler) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.Header = req.Header.Clone()
	clone.Header.Set("Authorization", "Bearer "+h.token)
	base := h.base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(clone)
}

func TestWorkspaceMCPServerExposesReadOnlyWorkspaceTools(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx := context.Background()
	cfg := &config.AppConfig{APIToken: "root-token"}
	routerFn, handlers := SetupRoutes(cfg, daemon.NewRegistry())

	wsRepo := workspacememory.New()
	if _, err := wsRepo.CreateWorkspace(ctx, &agentsv1.Workspace{
		Id:        "ws-a",
		Name:      "Workspace A",
		Slug:      "ws-a",
		CreatedAt: timestamppb.Now(),
		UpdatedAt: timestamppb.Now(),
	}); err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	fileRepo := agentfilememory.New()
	space, err := fileRepo.CreateSpace(ctx, "ws-a", &agentsv1.AgentFileSpace{Name: "Docs"})
	if err != nil {
		t.Fatalf("create file space: %v", err)
	}
	if _, err := fileRepo.WriteFile(ctx, "ws-a", space.GetId(), "/notes/intro.md", "hello workspace mcp", "text/markdown", map[string]string{"api_token": "hide-me"}); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if _, err := handlers.configStore.CreateAgent(ctx, "ws-a", &agentsv1.Agent{
		Name:        "planner",
		Description: "Plans work",
		Metadata:    map[string]string{"secret": "hide-me", "public": "ok"},
		Type:        agentsv1.AgentType_AGENT_TYPE_LLM,
	}); err != nil {
		t.Fatalf("create agent: %v", err)
	}
	if _, err := handlers.configStore.CreateMCPServer(ctx, "ws-a", &agentsv1.MCPServer{
		Id:        "external",
		Name:      "External MCP",
		Url:       "https://example.com/mcp",
		Transport: agentsv1.MCPServerTransport_MCP_SERVER_TRANSPORT_STREAMABLE_HTTP,
		Headers:   map[string]string{"Authorization": "Bearer hide-me"},
		Auth:      &agentsv1.MCPServerAuth{Type: agentsv1.MCPServerAuthType_MCP_SERVER_AUTH_TYPE_STATIC_HEADERS},
	}); err != nil {
		t.Fatalf("create mcp server: %v", err)
	}

	handlers.Wire(&BootstrapResult{WorkspaceRepo: wsRepo, AgentFileRepo: fileRepo})
	engine := gin.New()
	routerFn(engine)
	server := httptest.NewServer(engine)
	t.Cleanup(server.Close)

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v0.1.0"}, nil)
	session, err := client.Connect(ctx, &mcp.StreamableClientTransport{
		Endpoint:             server.URL + "/api/workspaces/ws-a/mcp",
		DisableStandaloneSSE: true,
		HTTPClient:           &http.Client{Transport: staticOAuthHandler{token: "root-token"}},
	}, nil)
	if err != nil {
		t.Fatalf("connect mcp client: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })

	tools, err := session.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}
	if !hasTool(tools.Tools, "workspace_info") || !hasTool(tools.Tools, "read_file") {
		t.Fatalf("expected workspace tools, got %+v", tools.Tools)
	}

	info, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "workspace_info", Arguments: map[string]any{}})
	if err != nil {
		t.Fatalf("workspace_info: %v", err)
	}
	infoContent := info.StructuredContent.(map[string]any)
	if infoContent["id"] != "ws-a" || infoContent["name"] != "Workspace A" {
		t.Fatalf("unexpected workspace_info: %+v", infoContent)
	}

	agents, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "list_agents", Arguments: map[string]any{}})
	if err != nil {
		t.Fatalf("list_agents: %v", err)
	}
	if containsStructuredValue(agents.StructuredContent, "hide-me") {
		t.Fatalf("list_agents leaked secret metadata: %+v", agents.StructuredContent)
	}

	mcpServers, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "list_mcp_servers", Arguments: map[string]any{}})
	if err != nil {
		t.Fatalf("list_mcp_servers: %v", err)
	}
	if containsStructuredValue(mcpServers.StructuredContent, "Bearer hide-me") {
		t.Fatalf("list_mcp_servers leaked header value: %+v", mcpServers.StructuredContent)
	}
	if !containsStructuredValue(mcpServers.StructuredContent, "Authorization") {
		t.Fatalf("list_mcp_servers should expose header names: %+v", mcpServers.StructuredContent)
	}

	read, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "read_file",
		Arguments: map[string]any{
			"space_id": space.GetId(),
			"path":     "/notes/intro.md",
		},
	})
	if err != nil {
		t.Fatalf("read_file: %v", err)
	}
	if !containsStructuredValue(read.StructuredContent, "hello workspace mcp") {
		t.Fatalf("read_file missing content: %+v", read.StructuredContent)
	}
	if containsStructuredValue(read.StructuredContent, "hide-me") {
		t.Fatalf("read_file leaked secret metadata: %+v", read.StructuredContent)
	}
}

func TestWorkspaceMCPServerRejectsAPITokenWorkspaceMismatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx := context.Background()
	cfg := &config.AppConfig{}
	routerFn, handlers := SetupRoutes(cfg, daemon.NewRegistry())
	tokenRepo := apitokenmemory.New()
	token := &agentsv1.APIToken{
		Id:          "token-a",
		Name:        "Token A",
		Prefix:      "bt_test",
		CreatedAt:   timestamppb.Now(),
		WorkspaceId: "ws-a",
	}
	if err := tokenRepo.Create(ctx, token, application.HashAPITokenSecret("bt_test")); err != nil {
		t.Fatalf("create token: %v", err)
	}
	handlers.Wire(&BootstrapResult{APITokenRepo: tokenRepo, AgentFileRepo: agentfilememory.New()})

	engine := gin.New()
	routerFn(engine)
	server := httptest.NewServer(engine)
	t.Cleanup(server.Close)

	req := httptest.NewRequest(http.MethodPost, server.URL+"/api/workspaces/ws-b/mcp", nil)
	req.Header.Set("Authorization", "Bearer bt_test")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403, body=%s", w.Code, w.Body.String())
	}
}

func hasTool(tools []*mcp.Tool, name string) bool {
	for _, tool := range tools {
		if tool.Name == name {
			return true
		}
	}
	return false
}

func containsStructuredValue(v any, want string) bool {
	switch x := v.(type) {
	case string:
		return x == want
	case []any:
		for _, item := range x {
			if containsStructuredValue(item, want) {
				return true
			}
		}
	case map[string]any:
		for _, item := range x {
			if containsStructuredValue(item, want) {
				return true
			}
		}
	}
	return false
}
