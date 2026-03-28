package runner

import (
	"testing"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestDeriveSessionID(t *testing.T) {
	tests := []struct {
		name   string
		scope  agentsv1.AgentSessionScope
		chatID int64
		userID int64
		want   string
	}{
		{
			name:   "user scope",
			scope:  agentsv1.AgentSessionScope_AGENT_SESSION_SCOPE_USER,
			chatID: 100, userID: 42,
			want: "user:42",
		},
		{
			name:   "chat scope",
			scope:  agentsv1.AgentSessionScope_AGENT_SESSION_SCOPE_CHAT,
			chatID: 100, userID: 42,
			want: "chat:100",
		},
		{
			name:   "unspecified defaults to chat",
			scope:  agentsv1.AgentSessionScope_AGENT_SESSION_SCOPE_UNSPECIFIED,
			chatID: 100, userID: 42,
			want: "chat:100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DeriveSessionID(tt.scope, tt.chatID, tt.userID)
			if got != tt.want {
				t.Errorf("DeriveSessionID() = %q, want %q", got, tt.want)
			}
		})
	}
}
