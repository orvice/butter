package telegram

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"butterfly.orx.me/core/log"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"google.golang.org/adk/session"

	"go.orx.me/apps/butter/internal/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const (
	callbackDebugToggle   = "debug_toggle"
	callbackAgentSelectPrefix = "agent_select:"
)

// Poller handles long-polling for a single Telegram AgentChannel.
type Poller struct {
	channelName  string
	channelCfg   *agentsv1.AgentChannel
	telegramCfg  *agentsv1.TelegramChannelConfig
	bot          *bot.Bot
	runner       *runner.Service
	selector     *AgentSelector
	debugToggle  *DebugToggle
	agentNames   []string
}

// NewPoller creates a new Telegram long-polling consumer.
func NewPoller(
	channelCfg *agentsv1.AgentChannel,
	runnerSvc *runner.Service,
	selector *AgentSelector,
	debugToggle *DebugToggle,
	agentNames []string,
) (*Poller, error) {
	p := &Poller{
		channelName: channelCfg.GetName(),
		channelCfg:  channelCfg,
		telegramCfg: channelCfg.GetTelegram(),
		runner:      runnerSvc,
		selector:    selector,
		debugToggle: debugToggle,
		agentNames:  agentNames,
	}

	b, err := bot.New(
		channelCfg.GetTelegram().GetBotToken(),
		bot.WithDefaultHandler(p.handleUpdate),
		bot.WithCallbackQueryDataHandler(callbackDebugToggle, bot.MatchTypeExact, p.handleDebugToggleCallback),
		bot.WithCallbackQueryDataHandler(callbackAgentSelectPrefix, bot.MatchTypePrefix, p.handleAgentSelectCallback),
	)
	if err != nil {
		return nil, fmt.Errorf("creating telegram bot for channel %q: %w", channelCfg.GetName(), err)
	}
	p.bot = b

	return p, nil
}

// Start begins the long-polling loop. Blocks until ctx is cancelled.
func (p *Poller) Start(ctx context.Context) {
	logger := log.FromContext(ctx)
	logger.Info("starting telegram poller", "channel", p.channelName, "agent_default", p.channelCfg.GetAgentName())
	p.bot.Start(ctx)
	logger.Info("telegram poller stopped", "channel", p.channelName)
}

func (p *Poller) handleUpdate(ctx context.Context, b *bot.Bot, update *models.Update) {
	logger := log.FromContext(ctx)

	msg := update.Message
	if msg == nil {
		return
	}

	logger.Debug("received update",
		"channel", p.channelName,
		"update_id", update.ID,
		"chat_id", msg.Chat.ID,
		"from_id", userIDFromMsg(msg),
		"text_len", len(msg.Text),
	)

	if !p.isAllowed(msg) {
		logger.Debug("message rejected by allowlist",
			"channel", p.channelName,
			"chat_id", msg.Chat.ID,
			"from_id", userIDFromMsg(msg),
		)
		return
	}

	if !p.matchesTrigger(msg) {
		logger.Debug("message did not match any trigger",
			"channel", p.channelName,
			"chat_id", msg.Chat.ID,
		)
		return
	}

	text := msg.Text
	if text == "" {
		logger.Debug("ignoring non-text message", "channel", p.channelName)
		return
	}

	// Handle /agent commands.
	if strings.HasPrefix(text, "/agent") {
		p.handleAgentCommand(ctx, b, msg)
		return
	}

	// Handle /debug toggle.
	if strings.HasPrefix(text, "/debug") {
		p.handleDebugCommand(ctx, b, msg)
		return
	}

	// Handle /status.
	if strings.HasPrefix(text, "/status") {
		p.handleStatusCommand(ctx, b, msg)
		return
	}

	p.handleMessage(ctx, b, msg)
}

func (p *Poller) isAllowed(msg *models.Message) bool {
	allowedChats := p.telegramCfg.GetAllowedChatIds()
	allowedUsers := p.telegramCfg.GetAllowedUserIds()

	if len(allowedChats) > 0 {
		if !slices.Contains(allowedChats, msg.Chat.ID) {
			return false
		}
	}

	if len(allowedUsers) > 0 && msg.From != nil {
		if !slices.Contains(allowedUsers, msg.From.ID) {
			return false
		}
	}

	return true
}

func (p *Poller) matchesTrigger(msg *models.Message) bool {
	triggers := p.channelCfg.GetTriggers()
	if len(triggers) == 0 {
		// No triggers configured means accept all messages.
		return true
	}

	for _, trigger := range triggers {
		switch trigger.GetType() {
		case agentsv1.AgentTriggerType_AGENT_TRIGGER_TYPE_MESSAGE:
			return true
		case agentsv1.AgentTriggerType_AGENT_TRIGGER_TYPE_PRIVATE_CHAT:
			if msg.Chat.Type == models.ChatTypePrivate {
				return true
			}
		case agentsv1.AgentTriggerType_AGENT_TRIGGER_TYPE_COMMAND:
			if strings.HasPrefix(msg.Text, "/") {
				return true
			}
		}
	}

	return false
}

func (p *Poller) handleAgentCommand(ctx context.Context, b *bot.Bot, msg *models.Message) {
	logger := log.FromContext(ctx)
	sessionID := p.deriveSessionID(msg)
	sub, arg := parseAgentCommand(msg.Text)

	logger.Info("handling agent command",
		"channel", p.channelName,
		"command", sub,
		"arg", arg,
		"session_id", sessionID,
		"from_id", userIDFromMsg(msg),
	)

	switch sub {
	case "list":
		activeAgent := p.getActiveAgent(ctx, sessionID)
		p.sendAgentList(ctx, b, msg, activeAgent)

	case "switch":
		if !p.runner.HasAgent(arg) {
			logger.Warn("agent switch failed: unknown agent",
				"channel", p.channelName,
				"requested_agent", arg,
			)
			p.sendReply(ctx, b, msg, fmt.Sprintf("Unknown agent: %q\n\nAvailable: %s", arg, strings.Join(p.agentNames, ", ")))
			return
		}
		if err := p.selector.Set(ctx, p.channelName, sessionID, arg); err != nil {
			logger.Error("failed to set agent selection in redis",
				"channel", p.channelName,
				"session_id", sessionID,
				"agent", arg,
				"err", err,
			)
			p.sendReply(ctx, b, msg, "Failed to switch agent. Please try again.")
			return
		}
		logger.Info("agent switched",
			"channel", p.channelName,
			"session_id", sessionID,
			"agent", arg,
		)
		p.sendReply(ctx, b, msg, fmt.Sprintf("Switched to agent: %s", arg))
	}
}

func (p *Poller) handleDebugCommand(ctx context.Context, b *bot.Bot, msg *models.Message) {
	logger := log.FromContext(ctx)
	sessionID := p.deriveSessionID(msg)

	newState, err := p.debugToggle.Toggle(ctx, p.channelName, sessionID, p.telegramCfg.GetDebug())
	if err != nil {
		logger.Error("failed to toggle debug mode",
			"channel", p.channelName,
			"session_id", sessionID,
			"err", err,
		)
		p.sendReply(ctx, b, msg, "Failed to toggle debug mode. Please try again.")
		return
	}

	logger.Info("debug mode toggled",
		"channel", p.channelName,
		"session_id", sessionID,
		"debug", newState,
	)
	p.sendDebugStatus(ctx, b, msg.Chat.ID, 0, newState)
}

func (p *Poller) handleMessage(ctx context.Context, b *bot.Bot, msg *models.Message) {
	logger := log.FromContext(ctx)
	sessionID := p.deriveSessionID(msg)
	agentName := p.getActiveAgent(ctx, sessionID)
	userID := fmt.Sprintf("%d", msg.From.ID)

	logger.Info("dispatching message to agent",
		"channel", p.channelName,
		"agent", agentName,
		"session_id", sessionID,
		"user_id", userID,
		"chat_id", msg.Chat.ID,
		"text_len", len(msg.Text),
	)

	// Send typing indicator.
	if p.channelCfg.GetDelivery().GetSendTyping() {
		if _, err := b.SendChatAction(ctx, &bot.SendChatActionParams{
			ChatID: msg.Chat.ID,
			Action: models.ChatActionTyping,
		}); err != nil {
			logger.Warn("failed to send typing indicator",
				"channel", p.channelName,
				"chat_id", msg.Chat.ID,
				"err", err,
			)
		}
	}

	// Build event callback for debug mode.
	var onEvent runner.EventCallback
	if IsDebugActive(ctx, p.debugToggle, p.channelName, sessionID, p.telegramCfg) {
		onEvent = func(evt *session.Event) {
			text := FormatDebugEvent(evt)
			if text == "" {
				return
			}
			if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: msg.Chat.ID,
				Text:   text,
			}); err != nil {
				logger.Warn("failed to send debug message",
					"channel", p.channelName,
					"chat_id", msg.Chat.ID,
					"err", err,
				)
			}
		}
	}

	// Build compaction callback for debug mode.
	var onCompaction runner.CompactionCallback
	if onEvent != nil {
		onCompaction = func(agentName string) {
			text := fmt.Sprintf("[DEBUG] Context compacted (agent: %s)", agentName)
			if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: msg.Chat.ID,
				Text:   text,
			}); err != nil {
				logger.Warn("failed to send compaction debug message",
					"channel", p.channelName,
					"chat_id", msg.Chat.ID,
					"err", err,
				)
			}
		}
	}

	response, err := p.runner.Run(ctx, p.channelName, agentName, sessionID, userID, msg.Text, onEvent, onCompaction)
	if err != nil {
		logger.Error("agent run failed",
			"channel", p.channelName,
			"agent", agentName,
			"session_id", sessionID,
			"err", err,
		)
		p.sendReply(ctx, b, msg, "Sorry, something went wrong processing your message.")
		return
	}

	if response == "" {
		response = "(no response)"
	}

	logger.Info("agent response ready",
		"channel", p.channelName,
		"agent", agentName,
		"session_id", sessionID,
		"response_len", len(response),
	)

	p.sendReply(ctx, b, msg, response)
}

func (p *Poller) getActiveAgent(ctx context.Context, sessionID string) string {
	logger := log.FromContext(ctx)
	selected, err := p.selector.Get(ctx, p.channelName, sessionID)
	if err != nil {
		logger.Warn("failed to get agent selection from redis, using default",
			"channel", p.channelName,
			"session_id", sessionID,
			"default_agent", p.channelCfg.GetAgentName(),
			"err", err,
		)
		return p.channelCfg.GetAgentName()
	}
	if selected == "" {
		return p.channelCfg.GetAgentName()
	}
	return selected
}

func (p *Poller) deriveSessionID(msg *models.Message) string {
	var userID int64
	if msg.From != nil {
		userID = msg.From.ID
	}
	scope := p.channelCfg.GetSession().GetScope()
	return runner.DeriveSessionID(scope, msg.Chat.ID, userID)
}

func (p *Poller) sendReply(ctx context.Context, b *bot.Bot, msg *models.Message, text string) {
	logger := log.FromContext(ctx)

	params := &bot.SendMessageParams{
		ChatID: msg.Chat.ID,
		Text:   text,
	}

	replyMode := p.channelCfg.GetDelivery().GetReplyMode()
	if replyMode == agentsv1.AgentReplyMode_AGENT_REPLY_MODE_REPLY {
		params.ReplyParameters = &models.ReplyParameters{
			MessageID: msg.ID,
		}
	}

	if _, err := b.SendMessage(ctx, params); err != nil {
		logger.Error("failed to send telegram message",
			"channel", p.channelName,
			"chat_id", msg.Chat.ID,
			"err", err,
		)
	} else {
		logger.Debug("telegram message sent",
			"channel", p.channelName,
			"chat_id", msg.Chat.ID,
			"reply_mode", replyMode.String(),
			"text_len", len(text),
		)
	}
}

func userIDFromMsg(msg *models.Message) int64 {
	if msg.From != nil {
		return msg.From.ID
	}
	return 0
}

// sendDebugStatus sends (or edits) a message showing debug state with a toggle button.
// If editMsgID is non-zero the existing message is edited; otherwise a new message is sent.
func (p *Poller) sendDebugStatus(ctx context.Context, b *bot.Bot, chatID any, editMsgID int, active bool) {
	label := "Debug: OFF"
	if active {
		label = "Debug: ON"
	}
	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: label, CallbackData: callbackDebugToggle}},
		},
	}
	if editMsgID != 0 {
		if _, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:      chatID,
			MessageID:   editMsgID,
			Text:        "Debug mode",
			ReplyMarkup: kb,
		}); err != nil {
			log.FromContext(ctx).Warn("failed to edit debug status message", "err", err)
		}
		return
	}
	if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        "Debug mode",
		ReplyMarkup: kb,
	}); err != nil {
		log.FromContext(ctx).Warn("failed to send debug status message", "err", err)
	}
}

// sendAgentList sends a message listing all agents as inline buttons.
func (p *Poller) sendAgentList(ctx context.Context, b *bot.Bot, msg *models.Message, activeAgent string) {
	var rows [][]models.InlineKeyboardButton
	for i := 0; i < len(p.agentNames); i += 2 {
		var row []models.InlineKeyboardButton
		for j := i; j < i+2 && j < len(p.agentNames); j++ {
			name := p.agentNames[j]
			label := name
			if name == activeAgent {
				label = "✅ " + name
			}
			row = append(row, models.InlineKeyboardButton{
				Text:         label,
				CallbackData: callbackAgentSelectPrefix + name,
			})
		}
		rows = append(rows, row)
	}
	params := &bot.SendMessageParams{
		ChatID:      msg.Chat.ID,
		Text:        "Select agent:",
		ReplyMarkup: &models.InlineKeyboardMarkup{InlineKeyboard: rows},
	}
	replyMode := p.channelCfg.GetDelivery().GetReplyMode()
	if replyMode == agentsv1.AgentReplyMode_AGENT_REPLY_MODE_REPLY {
		params.ReplyParameters = &models.ReplyParameters{MessageID: msg.ID}
	}
	if _, err := b.SendMessage(ctx, params); err != nil {
		log.FromContext(ctx).Error("failed to send agent list", "channel", p.channelName, "err", err)
	}
}

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
		if _, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{CallbackQueryID: cq.ID}); err != nil {
			logger.Warn("failed to answer callback query", "err", err)
		}
		return
	}

	sessionID := p.deriveSessionIDFromCallback(cq)
	newState, err := p.debugToggle.Toggle(ctx, p.channelName, sessionID, p.telegramCfg.GetDebug())
	if err != nil {
		logger.Error("failed to toggle debug via button", "channel", p.channelName, "session_id", sessionID, "err", err)
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{CallbackQueryID: cq.ID, Text: "Failed to toggle debug."}) //nolint:errcheck
		return
	}

	logger.Info("debug toggled via button", "channel", p.channelName, "session_id", sessionID, "debug", newState)

	p.sendDebugStatus(ctx, b, msg.Chat.ID, msg.ID, newState)

	status := "OFF"
	if newState {
		status = "ON"
	}
	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{CallbackQueryID: cq.ID, Text: "Debug " + status}) //nolint:errcheck
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
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{CallbackQueryID: cq.ID}) //nolint:errcheck
		return
	}

	agentName := strings.TrimPrefix(cq.Data, callbackAgentSelectPrefix)
	if !p.runner.HasAgent(agentName) {
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{CallbackQueryID: cq.ID, Text: "Unknown agent."}) //nolint:errcheck
		return
	}

	sessionID := p.deriveSessionIDFromCallback(cq)
	if err := p.selector.Set(ctx, p.channelName, sessionID, agentName); err != nil {
		logger.Error("failed to set agent via button", "channel", p.channelName, "session_id", sessionID, "err", err)
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{CallbackQueryID: cq.ID, Text: "Failed to switch agent."}) //nolint:errcheck
		return
	}

	logger.Info("agent switched via button", "channel", p.channelName, "session_id", sessionID, "agent", agentName)

	// Rebuild the keyboard to reflect the new active agent.
	var rows [][]models.InlineKeyboardButton
	for i := 0; i < len(p.agentNames); i += 2 {
		var row []models.InlineKeyboardButton
		for j := i; j < i+2 && j < len(p.agentNames); j++ {
			name := p.agentNames[j]
			label := name
			if name == agentName {
				label = "✅ " + name
			}
			row = append(row, models.InlineKeyboardButton{
				Text:         label,
				CallbackData: callbackAgentSelectPrefix + name,
			})
		}
		rows = append(rows, row)
	}
	if _, err := b.EditMessageReplyMarkup(ctx, &bot.EditMessageReplyMarkupParams{
		ChatID:      msg.Chat.ID,
		MessageID:   msg.ID,
		ReplyMarkup: &models.InlineKeyboardMarkup{InlineKeyboard: rows},
	}); err != nil {
		logger.Warn("failed to edit agent list keyboard", "err", err)
	}

	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{CallbackQueryID: cq.ID, Text: "Switched to " + agentName}) //nolint:errcheck
}

func (p *Poller) deriveSessionIDFromCallback(cq *models.CallbackQuery) string {
	msg := callbackMessage(cq)
	var chatID int64
	if msg != nil {
		chatID = msg.Chat.ID
	}
	userID := cq.From.ID
	scope := p.channelCfg.GetSession().GetScope()
	return runner.DeriveSessionID(scope, chatID, userID)
}

func (p *Poller) isAllowedCallbackQuery(cq *models.CallbackQuery) bool {
	msg := callbackMessage(cq)
	allowedChats := p.telegramCfg.GetAllowedChatIds()
	allowedUsers := p.telegramCfg.GetAllowedUserIds()

	if msg != nil && len(allowedChats) > 0 {
		if !slices.Contains(allowedChats, msg.Chat.ID) {
			return false
		}
	}
	if len(allowedUsers) > 0 {
		if !slices.Contains(allowedUsers, cq.From.ID) {
			return false
		}
	}
	return true
}

// callbackMessage returns the Message from a CallbackQuery.
func callbackMessage(cq *models.CallbackQuery) *models.Message {
	return cq.Message.Message
}
