package telegram

import (
	"testing"

	"github.com/go-telegram/bot/models"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestIsAllowed(t *testing.T) {
	tests := []struct {
		name        string
		allowedChat []int64
		allowedUser []int64
		chatID      int64
		userID      int64
		want        bool
	}{
		{
			name:   "empty allowlists allow all",
			chatID: 123, userID: 456,
			want: true,
		},
		{
			name:        "chat in allowlist",
			allowedChat: []int64{100, 200},
			chatID:      100, userID: 456,
			want: true,
		},
		{
			name:        "chat not in allowlist",
			allowedChat: []int64{100, 200},
			chatID:      300, userID: 456,
			want: false,
		},
		{
			name:        "user in allowlist",
			allowedUser: []int64{10, 20},
			chatID:      123, userID: 10,
			want: true,
		},
		{
			name:        "user not in allowlist",
			allowedUser: []int64{10, 20},
			chatID:      123, userID: 30,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Poller{
				telegramCfg: &agentsv1.TelegramChannelConfig{
					AllowedChatIds: tt.allowedChat,
					AllowedUserIds: tt.allowedUser,
				},
			}
			msg := &models.Message{
				Chat: models.Chat{ID: tt.chatID},
				From: &models.User{ID: tt.userID},
			}
			if got := p.isAllowed(msg); got != tt.want {
				t.Errorf("isAllowed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchesTrigger(t *testing.T) {
	tests := []struct {
		name     string
		triggers []*agentsv1.AgentTrigger
		msg      *models.Message
		want     bool
	}{
		{
			name: "no triggers means accept all",
			msg:  &models.Message{Text: "hello"},
			want: true,
		},
		{
			name: "message trigger matches text",
			triggers: []*agentsv1.AgentTrigger{
				{Type: agentsv1.AgentTriggerType_AGENT_TRIGGER_TYPE_MESSAGE},
			},
			msg:  &models.Message{Text: "hello"},
			want: true,
		},
		{
			name: "private chat trigger matches private",
			triggers: []*agentsv1.AgentTrigger{
				{Type: agentsv1.AgentTriggerType_AGENT_TRIGGER_TYPE_PRIVATE_CHAT},
			},
			msg:  &models.Message{Chat: models.Chat{Type: models.ChatTypePrivate}},
			want: true,
		},
		{
			name: "private chat trigger rejects group",
			triggers: []*agentsv1.AgentTrigger{
				{Type: agentsv1.AgentTriggerType_AGENT_TRIGGER_TYPE_PRIVATE_CHAT},
			},
			msg:  &models.Message{Chat: models.Chat{Type: models.ChatTypeGroup}},
			want: false,
		},
		{
			name: "command trigger matches command",
			triggers: []*agentsv1.AgentTrigger{
				{Type: agentsv1.AgentTriggerType_AGENT_TRIGGER_TYPE_COMMAND},
			},
			msg:  &models.Message{Text: "/start"},
			want: true,
		},
		{
			name: "command trigger rejects non-command",
			triggers: []*agentsv1.AgentTrigger{
				{Type: agentsv1.AgentTriggerType_AGENT_TRIGGER_TYPE_COMMAND},
			},
			msg:  &models.Message{Text: "hello"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Poller{
				channelCfg: &agentsv1.AgentChannel{
					Triggers: tt.triggers,
				},
			}
			if got := p.matchesTrigger(tt.msg); got != tt.want {
				t.Errorf("matchesTrigger() = %v, want %v", got, tt.want)
			}
		})
	}
}
