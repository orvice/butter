package mongo

import (
	"testing"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestUnmarshalMCPServerIgnoresLegacyStdioJSON(t *testing.T) {
	const legacy = `{
		"id": "legacy-stdio",
		"name": "Legacy stdio",
		"transport": "MCP_SERVER_TRANSPORT_STDIO",
		"command": "node",
		"args": ["server.js"],
		"env": {"TOKEN": "secret"}
	}`

	server := &agentsv1.MCPServer{}
	if err := unmarshalMCPServer(legacy, server); err != nil {
		t.Fatalf("unmarshal legacy stdio MCP server: %v", err)
	}
	if server.GetId() != "legacy-stdio" {
		t.Fatalf("expected legacy id, got %q", server.GetId())
	}
	if server.GetTransport() != agentsv1.MCPServerTransport_MCP_SERVER_TRANSPORT_UNSPECIFIED {
		t.Fatalf("expected discarded legacy transport to become unspecified, got %v", server.GetTransport())
	}
}

func TestUnmarshalMCPServerPreservesRemoteJSON(t *testing.T) {
	const remote = `{
		"id": "remote",
		"name": "Remote",
		"transport": "MCP_SERVER_TRANSPORT_STREAMABLE_HTTP",
		"url": "https://mcp.example.com/mcp"
	}`

	server := &agentsv1.MCPServer{}
	if err := unmarshalMCPServer(remote, server); err != nil {
		t.Fatalf("unmarshal remote MCP server: %v", err)
	}
	if server.GetTransport() != agentsv1.MCPServerTransport_MCP_SERVER_TRANSPORT_STREAMABLE_HTTP {
		t.Fatalf("expected streamable HTTP transport, got %v", server.GetTransport())
	}
	if server.GetUrl() != "https://mcp.example.com/mcp" {
		t.Fatalf("expected remote URL, got %q", server.GetUrl())
	}
}
