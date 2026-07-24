package telegram

import (
	"slices"
	"testing"

	"github.com/go-telegram/bot/models"

	"go.orx.me/apps/butter/internal/channel/pipeline"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// newTestPoller builds a Poller with just the fields toIncoming reads, without
// constructing a real Telegram bot.
func newTestPoller(tg *agentsv1.TelegramChannelConfig, session *agentsv1.AgentSessionBinding) *Poller {
	return &Poller{
		channelName: "tg",
		channelCfg:  &agentsv1.AgentChannel{Name: "tg", Telegram: tg, Session: session},
		telegramCfg: tg,
	}
}

func TestToIncoming_NormalizesFields(t *testing.T) {
	p := newTestPoller(&agentsv1.TelegramChannelConfig{}, nil)
	msg := &models.Message{
		ID:   99,
		Chat: models.Chat{ID: 123, Type: models.ChatTypePrivate},
		From: &models.User{ID: 456, Username: "orx", FirstName: "Or"},
		Text: "hello",
	}

	in := p.toIncoming(msg)

	if in.ChatID != "123" || in.UserID != "456" || in.MessageID != "99" {
		t.Errorf("ids: chat=%q user=%q msg=%q", in.ChatID, in.UserID, in.MessageID)
	}
	if in.Text != "hello" || in.HasMedia {
		t.Errorf("text=%q hasMedia=%v", in.Text, in.HasMedia)
	}
	if !in.IsPrivate || in.ChatType != agentsv1.ChatType_CHAT_TYPE_PRIVATE {
		t.Errorf("private=%v chatType=%v", in.IsPrivate, in.ChatType)
	}
	if in.ChannelType != "telegram" {
		t.Errorf("channelType=%q", in.ChannelType)
	}
	if in.Metadata["username"] != "orx" || in.Metadata["first_name"] != "Or" || in.Metadata["chat_id"] != "123" {
		t.Errorf("metadata=%v", in.Metadata)
	}
}

func TestToIncoming_GroupChatWithPhoto(t *testing.T) {
	p := newTestPoller(&agentsv1.TelegramChannelConfig{}, nil)
	msg := &models.Message{
		Chat:    models.Chat{ID: 1, Type: models.ChatTypeSupergroup},
		From:    &models.User{ID: 2},
		Photo:   []models.PhotoSize{{FileID: "f"}},
		Caption: "look",
	}

	in := p.toIncoming(msg)

	if in.IsPrivate || in.ChatType != agentsv1.ChatType_CHAT_TYPE_GROUP {
		t.Errorf("expected group chat, got private=%v chatType=%v", in.IsPrivate, in.ChatType)
	}
	if !in.HasMedia {
		t.Errorf("expected HasMedia for a photo message")
	}
}

// Admission rules are built so the pipeline's Admit reproduces the old
// allowlist behavior, including that a missing sender skips the user check.
func TestToIncoming_AdmissionRules(t *testing.T) {
	p := newTestPoller(&agentsv1.TelegramChannelConfig{
		AllowedChatIds: []int64{123},
		AllowedUserIds: []int64{456},
	}, nil)

	allowed := p.toIncoming(&models.Message{Chat: models.Chat{ID: 123}, From: &models.User{ID: 456}})
	if !pipeline.Admit(allowed.Admission) {
		t.Errorf("expected allowlisted chat+user to be admitted")
	}

	wrongUser := p.toIncoming(&models.Message{Chat: models.Chat{ID: 123}, From: &models.User{ID: 999}})
	if pipeline.Admit(wrongUser.Admission) {
		t.Errorf("expected non-allowlisted user to be rejected")
	}

	wrongChat := p.toIncoming(&models.Message{Chat: models.Chat{ID: 500}, From: &models.User{ID: 456}})
	if pipeline.Admit(wrongChat.Admission) {
		t.Errorf("expected non-allowlisted chat to be rejected")
	}

	// No sender: the user allowlist is skipped, so admission rests on the chat.
	noSender := p.toIncoming(&models.Message{Chat: models.Chat{ID: 123}})
	if !pipeline.Admit(noSender.Admission) {
		t.Errorf("expected missing sender to skip the user allowlist")
	}
}

func TestToIncoming_SessionScope(t *testing.T) {
	p := newTestPoller(&agentsv1.TelegramChannelConfig{}, &agentsv1.AgentSessionBinding{
		Scope: agentsv1.AgentSessionScope_AGENT_SESSION_SCOPE_USER,
	})
	in := p.toIncoming(&models.Message{Chat: models.Chat{ID: 123}, From: &models.User{ID: 456}})
	if !slices.Contains([]string{"user:456"}, in.SessionID) {
		t.Errorf("session id = %q, want user:456", in.SessionID)
	}
}
