package discord

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"butterfly.orx.me/core/log"
	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"google.golang.org/adk/session"

	"go.orx.me/apps/butter/internal/runtime/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const maxDiscordMessageLen = 2000

// Poller handles Discord gateway events for a single AgentChannel.
type Poller struct {
	ctx           context.Context
	channelName   string
	channelCfg    *agentsv1.AgentChannel
	discordCfg    *agentsv1.DiscordChannelConfig
	session       *discordgo.Session
	runner        *runner.Service
	selector      *AgentSelector
	modelSelector *ModelSelector
	debugToggle   *DebugToggle
	agentNames    []string
	modelNames    []string
}

// NewPoller creates a new Discord gateway poller.
func NewPoller(
	channelCfg *agentsv1.AgentChannel,
	runnerSvc *runner.Service,
	selector *AgentSelector,
	modelSelector *ModelSelector,
	debugToggle *DebugToggle,
	agentNames []string,
	modelNames []string,
) (*Poller, error) {
	dg, err := discordgo.New("Bot " + channelCfg.GetDiscord().GetBotToken())
	if err != nil {
		return nil, fmt.Errorf("creating discord session for channel %q: %w", channelCfg.GetName(), err)
	}

	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages | discordgo.IntentMessageContent

	p := &Poller{
		channelName:   channelCfg.GetName(),
		channelCfg:    channelCfg,
		discordCfg:    channelCfg.GetDiscord(),
		session:       dg,
		runner:        runnerSvc,
		selector:      selector,
		modelSelector: modelSelector,
		debugToggle:   debugToggle,
		agentNames:    agentNames,
		modelNames:    modelNames,
	}

	dg.AddHandler(p.handleMessageCreate)

	return p, nil
}

// Start opens the Discord gateway connection. Blocks until ctx is cancelled.
func (p *Poller) Start(ctx context.Context) {
	logger := log.FromContext(ctx)
	p.ctx = ctx

	logger.Info("starting discord poller", "channel", p.channelName, "agent_default", p.channelCfg.GetAgentName())

	if err := p.session.Open(); err != nil {
		logger.Error("failed to open discord gateway", "channel", p.channelName, "err", err)
		return
	}

	<-ctx.Done()

	logger.Info("stopping discord poller", "channel", p.channelName)
	if err := p.session.Close(); err != nil {
		logger.Error("failed to close discord session", "channel", p.channelName, "err", err)
	}
	logger.Info("discord poller stopped", "channel", p.channelName)
}

func (p *Poller) handleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	logger := log.FromContext(p.ctx)

	// Ignore bot messages (including self).
	if m.Author == nil || m.Author.Bot {
		return
	}

	logger.Debug("received discord message",
		"channel", p.channelName,
		"channel_id", m.ChannelID,
		"guild_id", m.GuildID,
		"author_id", m.Author.ID,
		"text_len", len(m.Content),
	)

	if !p.isAllowed(m) {
		logger.Debug("message rejected by allowlist",
			"channel", p.channelName,
			"channel_id", m.ChannelID,
			"guild_id", m.GuildID,
		)
		return
	}

	if !p.matchesTrigger(m) {
		logger.Debug("message did not match any trigger",
			"channel", p.channelName,
			"channel_id", m.ChannelID,
		)
		return
	}

	text := m.Content
	hasImages := hasImageAttachments(m)
	if text == "" && !hasImages {
		return
	}

	// Route commands.
	if strings.HasPrefix(text, "/agent") {
		p.handleAgentCommand(s, m)
		return
	}
	if strings.HasPrefix(text, "/model") {
		p.handleModelCommand(s, m)
		return
	}
	if strings.HasPrefix(text, "/debug") {
		p.handleDebugCommand(s, m)
		return
	}
	if strings.HasPrefix(text, "/status") {
		sessionID := p.deriveSessionID(m)
		userID := m.Author.ID
		statusText := p.formatStatusMessage(sessionID, userID)
		p.sendReply(s, m.ChannelID, statusText)
		return
	}
	if strings.HasPrefix(text, "/clear") {
		p.handleClearCommand(s, m)
		return
	}

	p.handleMessage(s, m)
}

func (p *Poller) isAllowed(m *discordgo.MessageCreate) bool {
	allowedGuilds := p.discordCfg.GetAllowedGuildIds()
	allowedChannels := p.discordCfg.GetAllowedChannelIds()

	if len(allowedGuilds) > 0 && m.GuildID != "" {
		if !slices.Contains(allowedGuilds, m.GuildID) {
			return false
		}
	}

	if len(allowedChannels) > 0 {
		if !slices.Contains(allowedChannels, m.ChannelID) {
			return false
		}
	}

	return true
}

func (p *Poller) matchesTrigger(m *discordgo.MessageCreate) bool {
	triggers := p.channelCfg.GetTriggers()
	if len(triggers) == 0 {
		return true
	}

	for _, trigger := range triggers {
		switch trigger.GetType() {
		case agentsv1.AgentTriggerType_AGENT_TRIGGER_TYPE_MESSAGE:
			return true
		case agentsv1.AgentTriggerType_AGENT_TRIGGER_TYPE_PRIVATE_CHAT:
			// DM messages have empty GuildID.
			if m.GuildID == "" {
				return true
			}
		case agentsv1.AgentTriggerType_AGENT_TRIGGER_TYPE_COMMAND:
			if strings.HasPrefix(m.Content, "/") {
				return true
			}
		}
	}

	return false
}

func (p *Poller) handleAgentCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	logger := log.FromContext(p.ctx)
	sessionID := p.deriveSessionID(m)
	sub, arg := parseAgentCommand(m.Content)

	logger.Info("handling agent command",
		"channel", p.channelName,
		"command", sub,
		"arg", arg,
		"session_id", sessionID,
		"from_id", m.Author.ID,
	)

	switch sub {
	case "list":
		activeAgent := p.getActiveAgent(sessionID)
		var lines []string
		for _, name := range p.agentNames {
			if name == activeAgent {
				lines = append(lines, fmt.Sprintf("**> %s** (active)", name))
			} else {
				lines = append(lines, fmt.Sprintf("  %s", name))
			}
		}
		p.sendReply(s, m.ChannelID, "**Agents:**\n"+strings.Join(lines, "\n"))

	case "switch":
		if !p.runner.HasAgent(arg) {
			p.sendReply(s, m.ChannelID, fmt.Sprintf("Unknown agent: %q\nAvailable: %s", arg, strings.Join(p.agentNames, ", ")))
			return
		}
		if err := p.selector.Set(p.ctx, p.channelName, sessionID, arg); err != nil {
			logger.Error("failed to set agent selection",
				"channel", p.channelName,
				"session_id", sessionID,
				"agent", arg,
				"err", err,
			)
			p.sendReply(s, m.ChannelID, "Failed to switch agent. Please try again.")
			return
		}
		logger.Info("agent switched",
			"channel", p.channelName,
			"session_id", sessionID,
			"agent", arg,
		)
		p.sendReply(s, m.ChannelID, fmt.Sprintf("Switched to agent: **%s**", arg))
	}
}

func (p *Poller) handleModelCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	sessionID := p.deriveSessionID(m)
	activeModel := p.getActiveModel(sessionID)

	var lines []string
	for _, name := range p.modelNames {
		if name == activeModel {
			lines = append(lines, fmt.Sprintf("**> %s** (active)", name))
		} else {
			lines = append(lines, fmt.Sprintf("  %s", name))
		}
	}
	p.sendReply(s, m.ChannelID, "**Models:**\n"+strings.Join(lines, "\n"))
}

func (p *Poller) handleDebugCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	logger := log.FromContext(p.ctx)
	sessionID := p.deriveSessionID(m)

	newState, err := p.debugToggle.Toggle(p.ctx, p.channelName, sessionID, p.discordCfg.GetDebug())
	if err != nil {
		logger.Error("failed to toggle debug mode",
			"channel", p.channelName,
			"session_id", sessionID,
			"err", err,
		)
		p.sendReply(s, m.ChannelID, "Failed to toggle debug mode. Please try again.")
		return
	}

	logger.Info("debug mode toggled",
		"channel", p.channelName,
		"session_id", sessionID,
		"debug", newState,
	)

	status := "OFF"
	if newState {
		status = "ON"
	}
	p.sendReply(s, m.ChannelID, fmt.Sprintf("Debug mode: **%s**", status))
}

func (p *Poller) handleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	logger := log.FromContext(p.ctx)
	sessionID := p.deriveSessionID(m)
	agentName := p.getActiveAgent(sessionID)
	userID := m.Author.ID

	logger.Info("dispatching message to agent",
		"channel", p.channelName,
		"agent", agentName,
		"session_id", sessionID,
		"user_id", userID,
		"channel_id", m.ChannelID,
		"text_len", len(m.Content),
	)

	// Send typing indicator.
	if p.channelCfg.GetDelivery().GetSendTyping() {
		if err := s.ChannelTyping(m.ChannelID); err != nil {
			logger.Warn("failed to send typing indicator",
				"channel", p.channelName,
				"channel_id", m.ChannelID,
				"err", err,
			)
		}
	}

	// Build event callback for debug mode.
	var onEvent runner.EventCallback
	if IsDebugActive(p.ctx, p.debugToggle, p.channelName, sessionID, p.discordCfg) {
		onEvent = func(evt *session.Event) {
			text := FormatDebugEvent(evt)
			if text == "" {
				return
			}
			p.sendReply(s, m.ChannelID, text)
		}
	}

	// Build compaction callback for debug mode.
	var onCompaction runner.CompactionCallback
	if onEvent != nil {
		onCompaction = func(agentName string) {
			text := fmt.Sprintf("[DEBUG] Context compacted (agent: %s)", agentName)
			p.sendReply(s, m.ChannelID, text)
		}
	}

	chatType := agentsv1.ChatType_CHAT_TYPE_GROUP
	if m.GuildID == "" {
		chatType = agentsv1.ChatType_CHAT_TYPE_PRIVATE
	}

	ctxInfo := &agentsv1.ContextInfo{
		Uuid:        uuid.Must(uuid.NewV7()).String(),
		SessionId:   sessionID,
		UserId:      userID,
		ChannelName: p.channelName,
		Source:      agentsv1.ContextSource_CONTEXT_SOURCE_CHANNEL,
		ChatId:      m.ChannelID,
		ChannelType: "discord",
		ChatType:    chatType,
		Metadata: map[string]string{
			"channel_id": m.ChannelID,
		},
	}
	if m.GuildID != "" {
		ctxInfo.Metadata["guild_id"] = m.GuildID
	}
	if m.Author != nil {
		if m.Author.Username != "" {
			ctxInfo.Metadata["username"] = m.Author.Username
		}
	}

	// Build multimodal input parts from message (text + optional image attachments).
	parts := buildMessageParts(p.ctx, m)
	if len(parts) == 0 {
		logger.Debug("no input parts to send", "channel", p.channelName)
		return
	}

	modelOverride := p.getActiveModel(sessionID)
	response, err := p.runner.Run(p.ctx, agentName, parts, modelOverride, ctxInfo, onEvent, onCompaction)
	if err != nil {
		logger.Error("agent run failed",
			"channel", p.channelName,
			"agent", agentName,
			"session_id", sessionID,
			"err", err,
		)
		p.sendReply(s, m.ChannelID, "Sorry, something went wrong processing your message.")
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

	p.sendMessage(s, m.ChannelID, response)
}

func (p *Poller) getActiveAgent(sessionID string) string {
	logger := log.FromContext(p.ctx)
	selected, err := p.selector.Get(p.ctx, p.channelName, sessionID)
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

func (p *Poller) getActiveModel(sessionID string) string {
	if p.modelSelector != nil {
		logger := log.FromContext(p.ctx)
		selected, err := p.modelSelector.Get(p.ctx, p.channelName, sessionID)
		if err != nil {
			logger.Warn("failed to get model selection from redis, using channel default",
				"channel", p.channelName,
				"session_id", sessionID,
				"err", err,
			)
		} else if selected != "" {
			return selected
		}
	}
	return p.channelCfg.GetModel()
}

func (p *Poller) deriveSessionID(m *discordgo.MessageCreate) string {
	scope := p.channelCfg.GetSession().GetScope()
	switch scope {
	case agentsv1.AgentSessionScope_AGENT_SESSION_SCOPE_USER:
		return "user:" + m.Author.ID
	case agentsv1.AgentSessionScope_AGENT_SESSION_SCOPE_CHAT:
		return "chat:" + m.ChannelID
	default:
		return "chat:" + m.ChannelID
	}
}

func (p *Poller) sendReply(s *discordgo.Session, channelID, text string) {
	p.sendMessage(s, channelID, text)
}

// sendMessage sends a message, splitting it if it exceeds Discord's 2000-character limit.
func (p *Poller) sendMessage(s *discordgo.Session, channelID, text string) {
	logger := log.FromContext(p.ctx)

	chunks := splitMessage(text, maxDiscordMessageLen)
	for _, chunk := range chunks {
		if _, err := s.ChannelMessageSend(channelID, chunk); err != nil {
			logger.Error("failed to send discord message",
				"channel", p.channelName,
				"channel_id", channelID,
				"err", err,
			)
			return
		}
	}
}

// splitMessage splits text into chunks of at most maxLen characters,
// preferring to split at newline boundaries.
func splitMessage(text string, maxLen int) []string {
	if len(text) <= maxLen {
		return []string{text}
	}

	var chunks []string
	for len(text) > 0 {
		if len(text) <= maxLen {
			chunks = append(chunks, text)
			break
		}

		// Try to find a newline to split at.
		splitAt := maxLen
		if idx := strings.LastIndex(text[:maxLen], "\n"); idx > 0 {
			splitAt = idx + 1
		}

		chunks = append(chunks, text[:splitAt])
		text = text[splitAt:]
	}

	return chunks
}
