package discord

import (
	"butterfly.orx.me/core/log"
	"github.com/bwmarrin/discordgo"
)

func (p *Poller) handleClearCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	logger := log.FromContext(p.ctx)
	sessionID := p.deriveSessionID(m)
	userID := m.Author.ID

	logger.Info("clearing session",
		"channel", p.channelName,
		"session_id", sessionID,
		"user_id", userID,
	)

	if err := p.runner.ClearSession(p.ctx, p.channelName, sessionID, userID); err != nil {
		logger.Error("failed to clear session",
			"channel", p.channelName,
			"session_id", sessionID,
			"err", err,
		)
		p.sendReply(s, m.ChannelID, "Failed to clear session. Please try again.")
		return
	}

	logger.Info("session cleared",
		"channel", p.channelName,
		"session_id", sessionID,
	)
	p.sendReply(s, m.ChannelID, "Session cleared.")
}
