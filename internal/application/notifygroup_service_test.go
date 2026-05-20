package application

import (
	"errors"
	"testing"

	"github.com/twitchtv/twirp"

	configrepo "go.orx.me/apps/butter/internal/repo/config"
	"go.orx.me/apps/butter/internal/repo/config/memory"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestNotifyGroupServiceServer_ValidatesNotifyGroup(t *testing.T) {
	ctx := testCtx()

	tests := []struct {
		name  string
		group *agentsv1.NotifyGroup
	}{
		{
			name:  "missing group",
			group: nil,
		},
		{
			name:  "empty name",
			group: &agentsv1.NotifyGroup{},
		},
		{
			name: "empty target type",
			group: &agentsv1.NotifyGroup{
				Name: "ops",
				Targets: []*agentsv1.NotifyTarget{{
					Enabled: true,
				}},
			},
		},
		{
			name: "telegram missing config",
			group: &agentsv1.NotifyGroup{
				Name: "ops",
				Targets: []*agentsv1.NotifyTarget{{
					Enabled: true,
					Type:    agentsv1.NotifyTargetType_NOTIFY_TARGET_TYPE_TELEGRAM,
				}},
			},
		},
		{
			name: "telegram missing chat",
			group: &agentsv1.NotifyGroup{
				Name: "ops",
				Targets: []*agentsv1.NotifyTarget{{
					Enabled:  true,
					Type:     agentsv1.NotifyTargetType_NOTIFY_TARGET_TYPE_TELEGRAM,
					Telegram: &agentsv1.TelegramNotifyTarget{BotToken: "token"},
				}},
			},
		},
		{
			name: "lark missing webhook",
			group: &agentsv1.NotifyGroup{
				Name: "ops",
				Targets: []*agentsv1.NotifyTarget{{
					Enabled: true,
					Type:    agentsv1.NotifyTargetType_NOTIFY_TARGET_TYPE_LARK_WEBHOOK,
					Lark:    &agentsv1.LarkNotifyTarget{},
				}},
			},
		},
		{
			name: "discord missing webhook",
			group: &agentsv1.NotifyGroup{
				Name: "ops",
				Targets: []*agentsv1.NotifyTarget{{
					Enabled: true,
					Type:    agentsv1.NotifyTargetType_NOTIFY_TARGET_TYPE_DISCORD_WEBHOOK,
					Discord: &agentsv1.DiscordNotifyTarget{},
				}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := memory.New()
			svc := NewNotifyGroupServiceServer(store)
			_, err := svc.CreateNotifyGroup(ctx, &agentsv1.CreateNotifyGroupRequest{NotifyGroup: tt.group})
			if twerr, ok := err.(twirp.Error); !ok || twerr.Code() != twirp.InvalidArgument {
				t.Fatalf("expected validation error, got %v", err)
			}
			if _, err := store.GetNotifyGroup(ctx, wsTest, "ops"); !errors.Is(err, configrepo.ErrNotFound) {
				t.Fatalf("expected invalid group not to be persisted, got %v", err)
			}
		})
	}
}

func TestNotifyGroupServiceServer_AcceptsSupportedTargets(t *testing.T) {
	store := memory.New()
	svc := NewNotifyGroupServiceServer(store)

	_, err := svc.CreateNotifyGroup(testCtx(), &agentsv1.CreateNotifyGroupRequest{
		NotifyGroup: &agentsv1.NotifyGroup{
			Name:    "ops",
			Enabled: true,
			Targets: []*agentsv1.NotifyTarget{
				{
					Name:     "telegram",
					Enabled:  true,
					Type:     agentsv1.NotifyTargetType_NOTIFY_TARGET_TYPE_TELEGRAM,
					Telegram: &agentsv1.TelegramNotifyTarget{BotToken: "token", ChatId: "chat"},
				},
				{
					Name:    "lark",
					Enabled: true,
					Type:    agentsv1.NotifyTargetType_NOTIFY_TARGET_TYPE_LARK_WEBHOOK,
					Lark:    &agentsv1.LarkNotifyTarget{WebhookUrl: "https://example.com/lark"},
				},
				{
					Name:    "discord",
					Enabled: true,
					Type:    agentsv1.NotifyTargetType_NOTIFY_TARGET_TYPE_DISCORD_WEBHOOK,
					Discord: &agentsv1.DiscordNotifyTarget{
						WebhookUrl: "https://example.com/discord",
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("expected valid notify group to pass, got %v", err)
	}
}
