package telegram

import (
	"context"
	"fmt"

	"butterfly.orx.me/core/log"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (p *Poller) handleClearCommand(ctx context.Context, b *bot.Bot, msg *models.Message) {
	logger := log.FromContext(ctx)
	sessionID := p.deriveSessionID(msg)
	userID := fmt.Sprintf("%d", userIDFromMsg(msg))

	logger.Info("clearing session",
		"channel", p.channelName,
		"session_id", sessionID,
		"user_id", userID,
	)

	if err := p.runner.ClearSession(ctx, p.channelName, sessionID, userID); err != nil {
		logger.Error("failed to clear session",
			"channel", p.channelName,
			"session_id", sessionID,
			"err", err,
		)
		p.sendReply(ctx, b, msg, "❌ Failed to clear session. Please try again.")
		return
	}

	logger.Info("session cleared",
		"channel", p.channelName,
		"session_id", sessionID,
	)
	p.sendReply(ctx, b, msg, "🧹 Session cleared.")
}
