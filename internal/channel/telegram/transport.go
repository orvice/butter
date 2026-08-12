package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"butterfly.orx.me/core/log"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"go.orx.me/apps/butter/internal/channel/pipeline"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Poller implements pipeline.Transport.
var _ pipeline.Transport = (*Poller)(nil)

func parseChatID(s string) int64 {
	id, _ := strconv.ParseInt(s, 10, 64)
	return id
}

// SendReply delivers a text reply, rendering Markdown and honoring reply mode,
// with a plain-text fallback if Markdown parsing fails.
//
// The send uses a cancel-detached context so that delivery succeeds even when
// the upstream context (e.g. the Telegram update handler) is already canceled
// by the time the agent finishes processing.
func (p *Poller) SendReply(ctx context.Context, msg pipeline.IncomingMessage, text string) {
	sendCtx := context.WithoutCancel(ctx)
	logger := log.FromContext(ctx)
	chatID := parseChatID(msg.ChatID)
	replyMode := p.channelCfg.GetDelivery().GetReplyMode()

	params := &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      markdownToTelegramMarkdownV2(text),
		ParseMode: models.ParseModeMarkdown,
	}
	if replyMode == agentsv1.AgentReplyMode_AGENT_REPLY_MODE_REPLY {
		if mid, err := strconv.Atoi(msg.MessageID); err == nil {
			params.ReplyParameters = &models.ReplyParameters{MessageID: mid}
		}
	}

	if _, err := p.bot.SendMessage(sendCtx, params); err != nil {
		logger.Warn("MarkdownV2 send failed, falling back to plain text",
			"channel", p.channelName, "chat_id", chatID, "err", err)
		params.Text = text
		params.ParseMode = ""
		if _, err2 := p.bot.SendMessage(sendCtx, params); err2 != nil {
			logger.Error("failed to send telegram message",
				"channel", p.channelName, "chat_id", chatID, "err", err2)
			return
		}
	}

	logger.Debug("telegram message sent",
		"channel", p.channelName, "chat_id", chatID,
		"reply_mode", replyMode.String(), "text_len", len(text))
}

// SendProcessing sends a "processing" placeholder message and returns its
// message ID so it can be edited later with debug progress and the final response.
func (p *Poller) SendProcessing(ctx context.Context, msg pipeline.IncomingMessage, agentName string, debug *pipeline.DebugSummary) string {
	sendCtx := context.WithoutCancel(ctx)
	logger := log.FromContext(ctx)
	chatID := parseChatID(msg.ChatID)

	now := time.Now().Format("15:04:05")
	text := formatProcessingMessage(agentName, now, msg.SessionID, debug)
	params := &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      markdownToTelegramMarkdownV2(text),
		ParseMode: models.ParseModeMarkdown,
	}
	if p.channelCfg.GetDelivery().GetReplyMode() == agentsv1.AgentReplyMode_AGENT_REPLY_MODE_REPLY {
		if mid, err := strconv.Atoi(msg.MessageID); err == nil {
			params.ReplyParameters = &models.ReplyParameters{MessageID: mid}
		}
	}

	sent, err := p.bot.SendMessage(sendCtx, params)
	if err != nil {
		logger.Warn("failed to send processing message, falling back to plain text",
			"channel", p.channelName, "chat_id", chatID, "err", err)
		params.Text = text
		params.ParseMode = ""
		sent, err = p.bot.SendMessage(sendCtx, params)
		if err != nil {
			logger.Error("failed to send processing message",
				"channel", p.channelName, "chat_id", chatID, "err", err)
			return ""
		}
	}

	logger.Debug("processing message sent",
		"channel", p.channelName, "chat_id", chatID, "message_id", sent.ID)
	return strconv.Itoa(sent.ID)
}

// EditDebug refreshes the processing message with aggregate counts and only
// the latest debug-relevant event.
func (p *Poller) EditDebug(ctx context.Context, msg pipeline.IncomingMessage, messageID string, agentName string, debug pipeline.DebugSummary) {
	now := time.Now().Format("15:04:05")
	p.editMessage(ctx, msg, messageID, formatProcessingMessage(agentName, now, msg.SessionID, &debug), "debug")
}

// EditReply edits a previously sent message with the final agent response,
// including agent name, timestamp, and session info.
func (p *Poller) EditReply(ctx context.Context, msg pipeline.IncomingMessage, messageID string, agentName string, text string, debug *pipeline.DebugSummary) {
	now := time.Now().Format("15:04:05")
	fullText := formatFinalMessage(agentName, text, now, msg.SessionID, debug)
	p.editMessage(ctx, msg, messageID, fullText, "final")
}

func (p *Poller) editMessage(ctx context.Context, msg pipeline.IncomingMessage, messageID, text, kind string) {
	sendCtx := context.WithoutCancel(ctx)
	logger := log.FromContext(ctx)
	chatID := parseChatID(msg.ChatID)

	mid, err := strconv.Atoi(messageID)
	if err != nil {
		logger.Error("invalid message ID for edit, falling back to send",
			"channel", p.channelName, "message_id", messageID, "kind", kind, "err", err)
		p.SendReply(ctx, msg, text)
		return
	}

	editParams := &bot.EditMessageTextParams{
		ChatID:    chatID,
		MessageID: mid,
		Text:      markdownToTelegramMarkdownV2(text),
		ParseMode: models.ParseModeMarkdown,
	}

	if _, err := p.bot.EditMessageText(sendCtx, editParams); err != nil {
		logger.Warn("MarkdownV2 edit failed, falling back to plain text",
			"channel", p.channelName, "chat_id", chatID, "message_id", mid, "kind", kind, "err", err)
		editParams.Text = text
		editParams.ParseMode = ""
		if _, err2 := p.bot.EditMessageText(sendCtx, editParams); err2 != nil {
			logger.Error("failed to edit telegram message",
				"channel", p.channelName, "chat_id", chatID, "message_id", mid, "err", err2)
		}
	}

	logger.Debug("telegram message edited",
		"channel", p.channelName, "chat_id", chatID,
		"message_id", mid, "kind", kind, "text_len", len(text))
}

func formatProcessingMessage(agentName, now, sessionID string, debug *pipeline.DebugSummary) string {
	if debug == nil {
		return fmt.Sprintf("🤖 *%s*\n⏳ Processing...\n\n🕐 `%s`\n💬 `%s`", agentName, now, sessionID)
	}

	sections := []string{fmt.Sprintf("🤖 **%s**\n⏳ Processing...", agentName)}
	sections = append(sections, formatDebugSummary(*debug))
	if latest := formatLatestDebug(*debug); latest != "" {
		sections = append(sections, "**Latest**\n"+latest)
	}
	sections = append(sections, fmt.Sprintf("🕐 `%s` · 💬 `%s`", now, sessionID))
	return strings.Join(sections, "\n\n")
}

func formatFinalMessage(agentName, text, now, sessionID string, debug *pipeline.DebugSummary) string {
	body := fmt.Sprintf("🤖 **%s**\n\n%s", agentName, text)
	footer := make([]string, 0, 2)
	if debug != nil {
		footer = append(footer, formatDebugSummary(*debug))
	}
	footer = append(footer, fmt.Sprintf("🕐 `%s` · 💬 `%s`", now, sessionID))
	return body + "\n\n─────────\n" + strings.Join(footer, "\n")
}

// SendTyping sends a typing chat action.
func (p *Poller) SendTyping(ctx context.Context, msg pipeline.IncomingMessage) {
	if _, err := p.bot.SendChatAction(context.WithoutCancel(ctx), &bot.SendChatActionParams{
		ChatID: parseChatID(msg.ChatID),
		Action: models.ChatActionTyping,
	}); err != nil {
		log.FromContext(ctx).Warn("failed to send typing indicator",
			"channel", p.channelName, "chat_id", msg.ChatID, "err", err)
	}
}

// SendDebugStatus sends a debug on/off status message with a toggle button.
func (p *Poller) SendDebugStatus(ctx context.Context, msg pipeline.IncomingMessage, active bool) {
	p.sendDebugStatus(ctx, parseChatID(msg.ChatID), 0, active)
}

// SendAgentList renders the agent selection inline keyboard.
func (p *Poller) SendAgentList(ctx context.Context, msg pipeline.IncomingMessage, choices []pipeline.AgentChoice) {
	params := &bot.SendMessageParams{
		ChatID:      parseChatID(msg.ChatID),
		Text:        "🤖 Select agent:",
		ReplyMarkup: &models.InlineKeyboardMarkup{InlineKeyboard: agentKeyboard(choices)},
	}
	if p.channelCfg.GetDelivery().GetReplyMode() == agentsv1.AgentReplyMode_AGENT_REPLY_MODE_REPLY {
		if mid, err := strconv.Atoi(msg.MessageID); err == nil {
			params.ReplyParameters = &models.ReplyParameters{MessageID: mid}
		}
	}
	if _, err := p.bot.SendMessage(context.WithoutCancel(ctx), params); err != nil {
		log.FromContext(ctx).Error("failed to send agent list", "channel", p.channelName, "err", err)
	}
}

// SendModelList renders the model selection inline keyboard.
func (p *Poller) SendModelList(ctx context.Context, msg pipeline.IncomingMessage, choices []pipeline.ModelChoice) {
	params := &bot.SendMessageParams{
		ChatID:      parseChatID(msg.ChatID),
		Text:        "🧠 Select model:",
		ReplyMarkup: &models.InlineKeyboardMarkup{InlineKeyboard: modelKeyboard(choices)},
	}
	if p.channelCfg.GetDelivery().GetReplyMode() == agentsv1.AgentReplyMode_AGENT_REPLY_MODE_REPLY {
		if mid, err := strconv.Atoi(msg.MessageID); err == nil {
			params.ReplyParameters = &models.ReplyParameters{MessageID: mid}
		}
	}
	if _, err := p.bot.SendMessage(context.WithoutCancel(ctx), params); err != nil {
		log.FromContext(ctx).Error("failed to send model list", "channel", p.channelName, "err", err)
	}
}

// SendStatus renders and sends the /status view.
func (p *Poller) SendStatus(ctx context.Context, msg pipeline.IncomingMessage, view pipeline.StatusView) {
	p.SendReply(ctx, msg, formatStatusView(view))
}

// formatStatusView adapts the platform-neutral StatusView into the Telegram
// Markdown status message.
func formatStatusView(view pipeline.StatusView) string {
	var modelText string
	switch {
	case view.ActiveModel != "" && view.ResolvedModel != "":
		modelText = fmt.Sprintf("`%s` -> `%s`", view.ActiveModel, view.ResolvedModel)
	case view.ActiveModel != "":
		modelText = fmt.Sprintf("`%s`", view.ActiveModel)
	case view.AgentModel != "":
		modelText = fmt.Sprintf("`%s` (agent default)", view.AgentModel)
	}

	var sess *sessionStatus
	if view.HasSession {
		sess = &sessionStatus{eventCount: view.EventCount, lastUpdate: view.LastUpdate}
	}

	return formatStatusMessage(view.AgentStatus, view.ActiveAgent, modelText, view.SessionID, sess, view.SessionErr, view.Now)
}

// sendDebugStatus sends (or edits) a message showing debug state with a toggle
// button. A non-zero editMsgID edits the existing message instead of sending.
func (p *Poller) sendDebugStatus(ctx context.Context, chatID any, editMsgID int, active bool) {
	label := "🔴 Debug: OFF"
	if active {
		label = "🟢 Debug: ON"
	}
	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: label, CallbackData: callbackDebugToggle}},
		},
	}
	sendCtx := context.WithoutCancel(ctx)
	if editMsgID != 0 {
		if _, err := p.bot.EditMessageText(sendCtx, &bot.EditMessageTextParams{
			ChatID:      chatID,
			MessageID:   editMsgID,
			Text:        "🐛 Debug mode",
			ReplyMarkup: kb,
		}); err != nil {
			log.FromContext(ctx).Warn("failed to edit debug status message", "err", err)
		}
		return
	}
	if _, err := p.bot.SendMessage(sendCtx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        "Debug mode",
		ReplyMarkup: kb,
	}); err != nil {
		log.FromContext(ctx).Warn("failed to send debug status message", "err", err)
	}
}

// agentKeyboard lays out agent choices two per row, flagging the active one.
func agentKeyboard(choices []pipeline.AgentChoice) [][]models.InlineKeyboardButton {
	var rows [][]models.InlineKeyboardButton
	for i := 0; i < len(choices); i += 2 {
		var row []models.InlineKeyboardButton
		for j := i; j < i+2 && j < len(choices); j++ {
			c := choices[j]
			label := c.Name
			if c.Active {
				label = "✅ " + c.Name
			}
			row = append(row, models.InlineKeyboardButton{
				Text:         label,
				CallbackData: callbackAgentSelectPrefix + c.Name,
			})
		}
		rows = append(rows, row)
	}
	return rows
}

// modelKeyboard lays out model choices two per row, flagging the active one.
func modelKeyboard(choices []pipeline.ModelChoice) [][]models.InlineKeyboardButton {
	var rows [][]models.InlineKeyboardButton
	for i := 0; i < len(choices); i += 2 {
		var row []models.InlineKeyboardButton
		for j := i; j < i+2 && j < len(choices); j++ {
			c := choices[j]
			label := c.Alias
			if c.Active {
				label = "✅ " + c.Alias
			}
			row = append(row, models.InlineKeyboardButton{
				Text:         label,
				CallbackData: callbackModelSelectPrefix + c.Alias,
			})
		}
		rows = append(rows, row)
	}
	return rows
}
