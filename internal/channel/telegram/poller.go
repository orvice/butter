package telegram

import (
	"context"
	"fmt"
	"strconv"

	"butterfly.orx.me/core/log"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/channel/pipeline"
	"go.orx.me/apps/butter/internal/runtime/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const (
	callbackDebugToggle       = "debug_toggle"
	callbackAgentSelectPrefix = "agent_select:"
	callbackModelSelectPrefix = "model_select:"
)

// Poller is the Telegram transport adapter: it long-polls the Telegram API,
// normalizes updates into pipeline.IncomingMessage, delegates all routing to a
// pipeline.Handler, and implements pipeline.Transport for outbound I/O.
type Poller struct {
	channelName   string
	channelCfg    *agentsv1.AgentChannel
	telegramCfg   *agentsv1.TelegramChannelConfig
	bot           *bot.Bot
	handler       *pipeline.Handler
	selector      *AgentSelector
	modelSelector *ModelSelector
	debugToggle   *DebugToggle
	agentNames    []string
	modelNames    []string // available model aliases
}

// NewPoller creates a new Telegram long-polling consumer.
func NewPoller(
	channelCfg *agentsv1.AgentChannel,
	runnerSvc *runner.Service,
	selector *AgentSelector,
	modelSelector *ModelSelector,
	debugToggle *DebugToggle,
	agentNames []string,
	modelNames []string,
) (*Poller, error) {
	p := &Poller{
		channelName:   channelCfg.GetName(),
		channelCfg:    channelCfg,
		telegramCfg:   channelCfg.GetTelegram(),
		selector:      selector,
		modelSelector: modelSelector,
		debugToggle:   debugToggle,
		agentNames:    agentNames,
		modelNames:    modelNames,
	}

	b, err := bot.New(
		channelCfg.GetTelegram().GetBotToken(),
		bot.WithDefaultHandler(p.handleUpdate),
		bot.WithCallbackQueryDataHandler(callbackDebugToggle, bot.MatchTypeExact, p.handleDebugToggleCallback),
		bot.WithCallbackQueryDataHandler(callbackAgentSelectPrefix, bot.MatchTypePrefix, p.handleAgentSelectCallback),
		bot.WithCallbackQueryDataHandler(callbackModelSelectPrefix, bot.MatchTypePrefix, p.handleModelSelectCallback),
	)
	if err != nil {
		return nil, fmt.Errorf("creating telegram bot for channel %q: %w", channelCfg.GetName(), err)
	}
	p.bot = b

	p.handler = pipeline.NewHandler(
		pipeline.Config{
			ChannelName:  channelCfg.GetName(),
			WorkspaceID:  channelCfg.GetWorkspaceId(),
			DefaultAgent: channelCfg.GetAgentName(),
			DefaultModel: channelCfg.GetModel(),
			ChannelType:  "telegram",
			Triggers:     channelCfg.GetTriggers(),
			SendTyping:   channelCfg.GetDelivery().GetSendTyping(),
			DebugDefault: channelCfg.GetTelegram().GetDebug(),
			AgentNames:   agentNames,
			ModelNames:   modelNames,
		},
		runnerSvc, selector, modelSelector, debugToggle, p,
	)

	return p, nil
}

// Start begins the long-polling loop. Blocks until ctx is cancelled.
func (p *Poller) Start(ctx context.Context) {
	logger := log.FromContext(ctx)
	logger.Info("starting telegram poller", "channel", p.channelName, "agent_default", p.channelCfg.GetAgentName())
	p.bot.Start(ctx)
	logger.Info("telegram poller stopped", "channel", p.channelName)
}

func (p *Poller) handleUpdate(ctx context.Context, _ *bot.Bot, update *models.Update) {
	msg := update.Message
	if msg == nil {
		return
	}

	log.FromContext(ctx).Debug("received update",
		"channel", p.channelName,
		"update_id", update.ID,
		"chat_id", msg.Chat.ID,
		"chat_type", string(msg.Chat.Type),
		"from_id", userIDFromMsg(msg),
		"text_len", len(msg.Text),
		"photo_count", len(msg.Photo),
	)

	p.handler.Handle(ctx, p.toIncoming(msg))
}

// toIncoming normalizes a Telegram message into a pipeline.IncomingMessage.
func (p *Poller) toIncoming(msg *models.Message) pipeline.IncomingMessage {
	chatID := strconv.FormatInt(msg.Chat.ID, 10)

	var userID string
	if msg.From != nil {
		userID = strconv.FormatInt(msg.From.ID, 10)
	}

	chatType := agentsv1.ChatType_CHAT_TYPE_PRIVATE
	if msg.Chat.Type == models.ChatTypeGroup || msg.Chat.Type == models.ChatTypeSupergroup {
		chatType = agentsv1.ChatType_CHAT_TYPE_GROUP
	}

	metadata := map[string]string{"chat_id": chatID}
	if msg.From != nil {
		if msg.From.Username != "" {
			metadata["username"] = msg.From.Username
		}
		if msg.From.FirstName != "" {
			metadata["first_name"] = msg.From.FirstName
		}
		if msg.From.LastName != "" {
			metadata["last_name"] = msg.From.LastName
		}
	}

	admission := []pipeline.AdmissionRule{
		{Value: chatID, Allowlist: int64sToStrings(p.telegramCfg.GetAllowedChatIds())},
		// A missing sender skips the user allowlist, matching the prior poller.
		{Value: userID, Allowlist: int64sToStrings(p.telegramCfg.GetAllowedUserIds()), SkipWhenEmpty: true},
	}

	return pipeline.IncomingMessage{
		SessionID:   p.deriveSessionID(msg),
		UserID:      userID,
		ChatID:      chatID,
		MessageID:   strconv.Itoa(msg.ID),
		Text:        msg.Text,
		HasMedia:    len(msg.Photo) > 0,
		IsPrivate:   msg.Chat.Type == models.ChatTypePrivate,
		ChatType:    chatType,
		ChannelType: "telegram",
		Metadata:    metadata,
		Admission:   admission,
		BuildParts: func(ctx context.Context) ([]*genai.Part, error) {
			return buildMessageParts(ctx, p.bot, msg)
		},
	}
}

func (p *Poller) deriveSessionID(msg *models.Message) string {
	var userID int64
	if msg.From != nil {
		userID = msg.From.ID
	}
	scope := p.channelCfg.GetSession().GetScope()
	return runner.DeriveSessionID(scope, msg.Chat.ID, userID)
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

// callbackIncoming builds a minimal IncomingMessage from a callback query so
// Transport rendering (which keys off ChatID/MessageID) works for button edits.
func (p *Poller) callbackIncoming(cq *models.CallbackQuery) pipeline.IncomingMessage {
	msg := callbackMessage(cq)
	var chatID, msgID string
	if msg != nil {
		chatID = strconv.FormatInt(msg.Chat.ID, 10)
		msgID = strconv.Itoa(msg.ID)
	}
	return pipeline.IncomingMessage{
		SessionID: p.deriveSessionIDFromCallback(cq),
		ChatID:    chatID,
		MessageID: msgID,
	}
}

func (p *Poller) isAllowedCallbackQuery(cq *models.CallbackQuery) bool {
	msg := callbackMessage(cq)
	rules := []pipeline.AdmissionRule{
		{Value: strconv.FormatInt(cq.From.ID, 10), Allowlist: int64sToStrings(p.telegramCfg.GetAllowedUserIds()), SkipWhenEmpty: true},
	}
	if msg != nil {
		rules = append(rules, pipeline.AdmissionRule{
			Value:     strconv.FormatInt(msg.Chat.ID, 10),
			Allowlist: int64sToStrings(p.telegramCfg.GetAllowedChatIds()),
		})
	}
	return pipeline.Admit(rules)
}

func userIDFromMsg(msg *models.Message) int64 {
	if msg.From != nil {
		return msg.From.ID
	}
	return 0
}

// callbackMessage returns the Message from a CallbackQuery.
func callbackMessage(cq *models.CallbackQuery) *models.Message {
	return cq.Message.Message
}

func int64sToStrings(ids []int64) []string {
	if len(ids) == 0 {
		return nil
	}
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = strconv.FormatInt(id, 10)
	}
	return out
}
