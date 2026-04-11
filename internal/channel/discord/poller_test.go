package discord

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestIsAllowed(t *testing.T) {
	tests := []struct {
		name          string
		allowedGuilds []string
		allowedChans  []string
		guildID       string
		channelID     string
		want          bool
	}{
		{
			name:    "empty allowlists allow all",
			guildID: "111", channelID: "222",
			want: true,
		},
		{
			name:          "guild in allowlist",
			allowedGuilds: []string{"100", "200"},
			guildID:       "100", channelID: "222",
			want: true,
		},
		{
			name:          "guild not in allowlist",
			allowedGuilds: []string{"100", "200"},
			guildID:       "300", channelID: "222",
			want: false,
		},
		{
			name:         "channel in allowlist",
			allowedChans: []string{"10", "20"},
			guildID:      "111", channelID: "10",
			want: true,
		},
		{
			name:         "channel not in allowlist",
			allowedChans: []string{"10", "20"},
			guildID:      "111", channelID: "30",
			want: false,
		},
		{
			name:    "DM with no guild (empty guild ID) passes guild filter",
			guildID: "", channelID: "222",
			allowedGuilds: []string{"100"},
			want:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Poller{
				discordCfg: &agentsv1.DiscordChannelConfig{
					AllowedGuildIds:   tt.allowedGuilds,
					AllowedChannelIds: tt.allowedChans,
				},
			}
			m := &discordgo.MessageCreate{
				Message: &discordgo.Message{
					GuildID:   tt.guildID,
					ChannelID: tt.channelID,
					Author:    &discordgo.User{ID: "user1"},
				},
			}
			if got := p.isAllowed(m); got != tt.want {
				t.Errorf("isAllowed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchesTrigger(t *testing.T) {
	tests := []struct {
		name     string
		triggers []*agentsv1.AgentTrigger
		guildID  string
		content  string
		want     bool
	}{
		{
			name:    "no triggers means accept all",
			content: "hello",
			want:    true,
		},
		{
			name: "message trigger matches text",
			triggers: []*agentsv1.AgentTrigger{
				{Type: agentsv1.AgentTriggerType_AGENT_TRIGGER_TYPE_MESSAGE},
			},
			content: "hello",
			want:    true,
		},
		{
			name: "private chat trigger matches DM (empty guild)",
			triggers: []*agentsv1.AgentTrigger{
				{Type: agentsv1.AgentTriggerType_AGENT_TRIGGER_TYPE_PRIVATE_CHAT},
			},
			guildID: "",
			content: "hello",
			want:    true,
		},
		{
			name: "private chat trigger rejects guild message",
			triggers: []*agentsv1.AgentTrigger{
				{Type: agentsv1.AgentTriggerType_AGENT_TRIGGER_TYPE_PRIVATE_CHAT},
			},
			guildID: "123",
			content: "hello",
			want:    false,
		},
		{
			name: "command trigger matches command",
			triggers: []*agentsv1.AgentTrigger{
				{Type: agentsv1.AgentTriggerType_AGENT_TRIGGER_TYPE_COMMAND},
			},
			content: "/start",
			want:    true,
		},
		{
			name: "command trigger rejects non-command",
			triggers: []*agentsv1.AgentTrigger{
				{Type: agentsv1.AgentTriggerType_AGENT_TRIGGER_TYPE_COMMAND},
			},
			content: "hello",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Poller{
				channelCfg: &agentsv1.AgentChannel{
					Triggers: tt.triggers,
				},
			}
			m := &discordgo.MessageCreate{
				Message: &discordgo.Message{
					GuildID:   tt.guildID,
					ChannelID: "chan1",
					Content:   tt.content,
					Author:    &discordgo.User{ID: "user1"},
				},
			}
			if got := p.matchesTrigger(m); got != tt.want {
				t.Errorf("matchesTrigger() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeriveSessionID(t *testing.T) {
	tests := []struct {
		name      string
		scope     agentsv1.AgentSessionScope
		userID    string
		channelID string
		want      string
	}{
		{
			name:      "user scope",
			scope:     agentsv1.AgentSessionScope_AGENT_SESSION_SCOPE_USER,
			userID:    "12345",
			channelID: "chan1",
			want:      "user:12345",
		},
		{
			name:      "chat scope",
			scope:     agentsv1.AgentSessionScope_AGENT_SESSION_SCOPE_CHAT,
			userID:    "12345",
			channelID: "chan1",
			want:      "chat:chan1",
		},
		{
			name:      "default scope uses chat",
			scope:     agentsv1.AgentSessionScope_AGENT_SESSION_SCOPE_UNSPECIFIED,
			userID:    "12345",
			channelID: "chan1",
			want:      "chat:chan1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Poller{
				channelCfg: &agentsv1.AgentChannel{
					Session: &agentsv1.AgentSessionBinding{
						Scope: tt.scope,
					},
				},
			}
			m := &discordgo.MessageCreate{
				Message: &discordgo.Message{
					ChannelID: tt.channelID,
					Author:    &discordgo.User{ID: tt.userID},
				},
			}
			if got := p.deriveSessionID(m); got != tt.want {
				t.Errorf("deriveSessionID() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSplitMessage(t *testing.T) {
	tests := []struct {
		name   string
		text   string
		maxLen int
		want   int // expected number of chunks
	}{
		{
			name:   "short message",
			text:   "hello",
			maxLen: 10,
			want:   1,
		},
		{
			name:   "exact limit",
			text:   "hello",
			maxLen: 5,
			want:   1,
		},
		{
			name:   "needs splitting",
			text:   "hello world this is a test",
			maxLen: 10,
			want:   3,
		},
		{
			name:   "split at newline",
			text:   "line1\nline2\nline3",
			maxLen: 10,
			want:   3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunks := splitMessage(tt.text, tt.maxLen)
			if len(chunks) != tt.want {
				t.Errorf("splitMessage() returned %d chunks, want %d; chunks=%v", len(chunks), tt.want, chunks)
			}
			// Verify all content is preserved.
			var combined string
			for _, c := range chunks {
				combined += c
			}
			if combined != tt.text {
				t.Errorf("splitMessage() lost content: got %q, want %q", combined, tt.text)
			}
			// Verify no chunk exceeds maxLen.
			for i, c := range chunks {
				if len(c) > tt.maxLen {
					t.Errorf("chunk %d exceeds maxLen: len=%d, maxLen=%d", i, len(c), tt.maxLen)
				}
			}
		})
	}
}
