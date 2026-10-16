package agent

import (
	"strings"
	"testing"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// The MCP and Remote Agent registries span Workspaces, but an Agent resolves
// referenced IDs only among its own Workspace's entries: two Workspaces may
// each own a resource with the same ID, and a run must never reach the other
// one's endpoint, headers or daemon runtime (#455).

func TestResolveMCPServersStaysInTheHostWorkspace(t *testing.T) {
	registry := []agentsv1.MCPServer{
		{Id: "docs", Name: "docs-alpha", Url: "https://alpha.example/mcp", WorkspaceId: "alpha"},
		{Id: "docs", Name: "docs-beta", Url: "https://beta.example/mcp", WorkspaceId: "beta"},
		{Id: "only-beta", Name: "only-beta", Url: "https://beta.example/x", WorkspaceId: "beta"},
	}
	host := func(ws string, ids ...string) *agentsv1.Agent {
		return &agentsv1.Agent{WorkspaceId: ws, Config: &agentsv1.AgentConfig{McpServerIds: ids}}
	}

	for _, tc := range []struct{ ws, want string }{{"alpha", "https://alpha.example/mcp"}, {"beta", "https://beta.example/mcp"}} {
		got, err := resolveMCPServers(host(tc.ws, "docs"), registry)
		if err != nil || len(got) != 1 || got[0].GetUrl() != tc.want {
			t.Fatalf("workspace %s resolved %v, %v; want %s", tc.ws, got, err, tc.want)
		}
	}

	// An ID that exists only in another Workspace does not resolve.
	if _, err := resolveMCPServers(host("alpha", "only-beta"), registry); err == nil ||
		!strings.Contains(err.Error(), "unknown mcp_server_id") {
		t.Fatalf("a foreign MCP server resolved: %v", err)
	}

	// An Agent without a Workspace (a legacy record) keeps the flat lookup.
	if got, err := resolveMCPServers(host("", "only-beta"), registry); err != nil || len(got) != 1 {
		t.Fatalf("legacy agent resolved %v, %v", got, err)
	}
}

func TestResolveRemoteAgentsStaysInTheHostWorkspace(t *testing.T) {
	registry := []agentsv1.RemoteAgent{
		{Id: "only-beta", Name: "only-beta", WorkspaceId: "beta",
			Protocol: agentsv1.RemoteAgentProtocol_REMOTE_AGENT_PROTOCOL_DAEMON, DaemonRuntimeId: "rt", AcpRuntime: "codex"},
	}
	host := &agentsv1.Agent{WorkspaceId: "alpha", Config: &agentsv1.AgentConfig{RemoteAgentIds: []string{"only-beta"}}}
	if _, err := resolveRemoteAgents(host, registry, nil); err == nil ||
		!strings.Contains(err.Error(), "unknown remote_agent_id") {
		t.Fatalf("a foreign Remote Agent resolved: %v", err)
	}
}
