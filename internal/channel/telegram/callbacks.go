package telegram

import (
	"context"
	"strings"

	"butterfly.orx.me/core/log"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// handleDebugToggleCallback handles the inline button press for debug toggle.
func (p *Poller) handleDebugToggleCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	logger := log.FromContext(ctx)
	cq := update.CallbackQuery
	if cq == nil {
		return
	}
	msg := callbackMessage(cq)
	if msg == nil {
		return
	}
	if !p.isAllowedCallbackQuery(cq) {
		answerCallback(ctx, b, cq.ID, "")
		return
	}

	sessionID := p.deriveSessionIDFromCallback(cq)
	newState, err := p.handler.ToggleDebug(ctx, sessionID)
	if err != nil {
		logger.Error("failed to toggle debug via button", "channel", p.channelName, "session_id", sessionID, "err", err)
		answerCallback(ctx, b, cq.ID, "❌ Failed to toggle debug.")
		return
	}

	logger.Info("debug toggled via button", "channel", p.channelName, "session_id", sessionID, "debug", newState)
	p.sendDebugStatus(ctx, msg.Chat.ID, msg.ID, newState)

	status := "🔴 OFF"
	if newState {
		status = "🟢 ON"
	}
	answerCallback(ctx, b, cq.ID, "🐛 Debug "+status)
}

// handleAgentSelectCallback handles the inline button press for agent selection.
func (p *Poller) handleAgentSelectCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	logger := log.FromContext(ctx)
	cq := update.CallbackQuery
	if cq == nil {
		return
	}
	msg := callbackMessage(cq)
	if msg == nil {
		return
	}
	if !p.isAllowedCallbackQuery(cq) {
		answerCallback(ctx, b, cq.ID, "")
		return
	}

	agentName := strings.TrimPrefix(cq.Data, callbackAgentSelectPrefix)
	sessionID := p.deriveSessionIDFromCallback(cq)
	ok, err := p.handler.SelectAgent(ctx, sessionID, agentName)
	if err != nil {
		logger.Error("failed to set agent via button", "channel", p.channelName, "session_id", sessionID, "err", err)
		answerCallback(ctx, b, cq.ID, "❌ Failed to switch agent.")
		return
	}
	if !ok {
		answerCallback(ctx, b, cq.ID, "❓ Unknown agent.")
		return
	}

	logger.Info("agent switched via button", "channel", p.channelName, "session_id", sessionID, "agent", agentName)
	p.editKeyboard(ctx, msg, agentKeyboard(p.handler.AgentChoices(ctx, sessionID)))
	answerCallback(ctx, b, cq.ID, "✅ Switched to "+agentName)
}

// handleModelSelectCallback handles the inline button press for model selection.
func (p *Poller) handleModelSelectCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	logger := log.FromContext(ctx)
	cq := update.CallbackQuery
	if cq == nil {
		return
	}
	msg := callbackMessage(cq)
	if msg == nil {
		return
	}
	if !p.isAllowedCallbackQuery(cq) {
		answerCallback(ctx, b, cq.ID, "")
		return
	}

	modelAlias := strings.TrimPrefix(cq.Data, callbackModelSelectPrefix)
	sessionID := p.deriveSessionIDFromCallback(cq)
	ok, err := p.handler.SelectModel(ctx, sessionID, modelAlias)
	if err != nil {
		logger.Error("failed to set model via button", "channel", p.channelName, "session_id", sessionID, "err", err)
		answerCallback(ctx, b, cq.ID, "❌ Failed to switch model.")
		return
	}
	if !ok {
		answerCallback(ctx, b, cq.ID, "❓ Unknown model.")
		return
	}

	logger.Info("model switched via button", "channel", p.channelName, "session_id", sessionID, "model", modelAlias)
	p.editKeyboard(ctx, msg, modelKeyboard(p.handler.ModelChoices(ctx, sessionID)))
	answerCallback(ctx, b, cq.ID, "✅ Switched to "+modelAlias)
}

// editKeyboard replaces the inline keyboard on an existing message.
func (p *Poller) editKeyboard(ctx context.Context, msg *models.Message, rows [][]models.InlineKeyboardButton) {
	if _, err := p.bot.EditMessageReplyMarkup(ctx, &bot.EditMessageReplyMarkupParams{
		ChatID:      msg.Chat.ID,
		MessageID:   msg.ID,
		ReplyMarkup: &models.InlineKeyboardMarkup{InlineKeyboard: rows},
	}); err != nil {
		log.FromContext(ctx).Warn("failed to edit inline keyboard", "channel", p.channelName, "err", err)
	}
}

func answerCallback(ctx context.Context, b *bot.Bot, id, text string) {
	if _, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{CallbackQueryID: id, Text: text}); err != nil {
		log.FromContext(ctx).Warn("failed to answer callback query", "err", err)
	}
}
