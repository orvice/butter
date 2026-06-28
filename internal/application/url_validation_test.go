package application

import (
	"testing"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestValidateRemoteAgentURL(t *testing.T) {
	cases := []struct {
		name    string
		ra      *agentsv1.RemoteAgent
		wantErr bool
	}{
		{
			name: "a2a valid https",
			ra: &agentsv1.RemoteAgent{
				Protocol: agentsv1.RemoteAgentProtocol_REMOTE_AGENT_PROTOCOL_A2A,
				Url:      "https://example.com/agent",
			},
		},
		{
			name: "a2a missing url",
			ra: &agentsv1.RemoteAgent{
				Protocol: agentsv1.RemoteAgentProtocol_REMOTE_AGENT_PROTOCOL_A2A,
			},
			wantErr: true,
		},
		{
			name: "a2a file scheme rejected",
			ra: &agentsv1.RemoteAgent{
				Protocol: agentsv1.RemoteAgentProtocol_REMOTE_AGENT_PROTOCOL_A2A,
				Url:      "file:///etc/passwd",
			},
			wantErr: true,
		},
		{
			name: "a2a empty host rejected",
			ra: &agentsv1.RemoteAgent{
				Protocol: agentsv1.RemoteAgentProtocol_REMOTE_AGENT_PROTOCOL_A2A,
				Url:      "http:///",
			},
			wantErr: true,
		},
		{
			name: "daemon ignores empty url",
			ra: &agentsv1.RemoteAgent{
				Protocol:        agentsv1.RemoteAgentProtocol_REMOTE_AGENT_PROTOCOL_DAEMON,
				DaemonRuntimeId: "runtime-1",
				AcpRuntime:      "opencode",
			},
		},
		{
			name: "daemon validates non-empty url",
			ra: &agentsv1.RemoteAgent{
				Protocol:        agentsv1.RemoteAgentProtocol_REMOTE_AGENT_PROTOCOL_DAEMON,
				DaemonRuntimeId: "runtime-1",
				AcpRuntime:      "opencode",
				Url:             "gopher://localhost",
			},
			wantErr: true,
		},
		{
			name: "opencode_http valid http",
			ra: &agentsv1.RemoteAgent{
				Protocol: agentsv1.RemoteAgentProtocol_REMOTE_AGENT_PROTOCOL_OPENCODE_HTTP,
				Url:      "http://localhost:4096",
			},
		},
		{
			name: "opencode_http missing url",
			ra: &agentsv1.RemoteAgent{
				Protocol: agentsv1.RemoteAgentProtocol_REMOTE_AGENT_PROTOCOL_OPENCODE_HTTP,
			},
			wantErr: true,
		},
		{
			name: "opencode_http file scheme rejected",
			ra: &agentsv1.RemoteAgent{
				Protocol: agentsv1.RemoteAgentProtocol_REMOTE_AGENT_PROTOCOL_OPENCODE_HTTP,
				Url:      "file:///etc/passwd",
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateRemoteAgentURL(tc.ra)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateModelProviderBaseURL(t *testing.T) {
	cases := []struct {
		name    string
		baseURL string
		wantErr bool
	}{
		{name: "empty allowed", baseURL: ""},
		{name: "https ok", baseURL: "https://api.openai.com/v1"},
		{name: "http ok", baseURL: "http://localhost:11434/v1"},
		{name: "missing scheme rejected", baseURL: "api.openai.com/v1", wantErr: true},
		{name: "file scheme rejected", baseURL: "file:///etc/passwd", wantErr: true},
		{name: "ftp scheme rejected", baseURL: "ftp://example.com", wantErr: true},
		{name: "empty host rejected", baseURL: "http:///path", wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateModelProviderBaseURL(&agentsv1.ModelProvider{BaseUrl: tc.baseURL})
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
