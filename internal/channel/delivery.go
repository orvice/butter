package channel

import (
	"context"
	"fmt"
	"strings"

	"go.orx.me/apps/butter/internal/channel/discord"
	"go.orx.me/apps/butter/internal/channel/telegram"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type textSender func(ctx context.Context, token, targetID, text string) error

// DeliverCronResult sends a cron execution result to the configured AgentChannel.
func (m *Manager) DeliverCronResult(ctx context.Context, job *agentsv1.CronJob, exec *agentsv1.CronExecution) error {
	delivery := job.GetDelivery()
	channelName := strings.TrimSpace(delivery.GetChannelName())
	chatID := strings.TrimSpace(delivery.GetChatId())
	if channelName == "" {
		return fmt.Errorf("channel_name is required")
	}
	if chatID == "" {
		return fmt.Errorf("chat_id is required")
	}

	ch, err := m.repo.GetChannel(ctx, job.GetWorkspaceId(), channelName)
	if err != nil {
		return fmt.Errorf("get channel %q: %w", channelName, err)
	}
	if !ch.GetEnabled() {
		return fmt.Errorf("channel %q is disabled", channelName)
	}

	text := formatCronResult(job, exec)
	switch ch.GetPlatform() {
	case agentsv1.AgentChannelPlatform_AGENT_CHANNEL_PLATFORM_TELEGRAM:
		token := ch.GetTelegram().GetBotToken()
		if token == "" {
			return fmt.Errorf("telegram channel %q has empty bot token", channelName)
		}
		return m.sendTelegram(ctx, token, chatID, text)
	case agentsv1.AgentChannelPlatform_AGENT_CHANNEL_PLATFORM_DISCORD:
		token := ch.GetDiscord().GetBotToken()
		if token == "" {
			return fmt.Errorf("discord channel %q has empty bot token", channelName)
		}
		return m.sendDiscord(ctx, token, chatID, text)
	default:
		return fmt.Errorf("channel %q uses unsupported platform %s", channelName, ch.GetPlatform().String())
	}
}

func (m *Manager) sendTelegram(ctx context.Context, token, chatID, text string) error {
	if m.telegramSender != nil {
		return m.telegramSender(ctx, token, chatID, text)
	}
	return telegram.SendText(ctx, token, chatID, text)
}

func (m *Manager) sendDiscord(ctx context.Context, token, channelID, text string) error {
	if m.discordSender != nil {
		return m.discordSender(ctx, token, channelID, text)
	}
	return discord.SendText(ctx, token, channelID, text)
}

func formatCronResult(job *agentsv1.CronJob, exec *agentsv1.CronExecution) string {
	status := "success"
	if exec.GetStatus() == agentsv1.CronExecutionStatus_CRON_EXECUTION_STATUS_ERROR {
		status = "error"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Cron job %s finished with %s\n", job.GetName(), status)
	fmt.Fprintf(&b, "Agent: %s\n", exec.GetAgentName())
	if exec.GetStartedAt() != nil {
		fmt.Fprintf(&b, "Started: %s\n", exec.GetStartedAt().AsTime().Format("2006-01-02 15:04:05 MST"))
	}
	if exec.GetFinishedAt() != nil {
		fmt.Fprintf(&b, "Finished: %s\n", exec.GetFinishedAt().AsTime().Format("2006-01-02 15:04:05 MST"))
	}
	if output := strings.TrimSpace(exec.GetOutput()); output != "" {
		fmt.Fprintf(&b, "\n%s", output)
	}
	return b.String()
}
