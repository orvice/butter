package application

import (
	"testing"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestRedactInstalledGlobalMCPSecret(t *testing.T) {
	server := &agentsv1.MCPServer{
		Name: "preset",
		Headers: map[string]string{
			"Authorization": "Bearer secret-token",
			"X-Api-Key":     "abc123",
		},
		Auth: &agentsv1.MCPServerAuth{
			Oauth2: &agentsv1.MCPServerOAuth2Config{
				ClientId:     "client",
				ClientSecret: "do-not-leak",
			},
		},
		Metadata: map[string]string{globalMCPPresetMetadataKey: "preset-id"},
	}

	got := redactInstalledGlobalMCPSecret(server)

	if got.GetAuth().GetOauth2().GetClientSecret() != "" {
		t.Errorf("client_secret not redacted: %q", got.GetAuth().GetOauth2().GetClientSecret())
	}
	if got.GetAuth().GetOauth2().GetClientId() != "client" {
		t.Errorf("client_id should remain: got %q", got.GetAuth().GetOauth2().GetClientId())
	}
	for k, v := range got.GetHeaders() {
		if v != redactedHeaderValue {
			t.Errorf("header %q value not redacted: %q", k, v)
		}
	}
	// Source object must be untouched (proto.Clone semantics).
	if server.GetAuth().GetOauth2().GetClientSecret() != "do-not-leak" {
		t.Errorf("source mutated: client_secret = %q", server.GetAuth().GetOauth2().GetClientSecret())
	}
	if server.GetHeaders()["Authorization"] != "Bearer secret-token" {
		t.Errorf("source header mutated: %q", server.GetHeaders()["Authorization"])
	}
}

func TestRedactInstalledGlobalMCPSecret_NotInstalledPreset(t *testing.T) {
	// Non-preset servers (no metadata marker) must be returned as-is so
	// regular user-owned MCP server records keep their own headers/secrets.
	server := &agentsv1.MCPServer{
		Name:    "user-owned",
		Headers: map[string]string{"X": "y"},
	}
	got := redactInstalledGlobalMCPSecret(server)
	if got.GetHeaders()["X"] != "y" {
		t.Fatalf("expected headers untouched, got %v", got.GetHeaders())
	}
}
