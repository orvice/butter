package pipeline

import (
	"context"
	"maps"

	"butterfly.orx.me/core/log"
	"github.com/google/uuid"
	"google.golang.org/adk/v2/session"

	"go.orx.me/apps/butter/internal/runtime/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Handle routes a normalized inbound message through the channel pipeline:
// admission, trigger matching, the empty-message guard, command routing, and
// finally the plain-message path.
func (h *Handler) Handle(ctx context.Context, msg IncomingMessage) {
	logger := log.FromContext(ctx)

	if !Admit(msg.Admission) {
		logger.Debug("message rejected by allowlist",
			"channel", h.cfg.ChannelName,
			"chat_id", msg.ChatID,
			"user_id", msg.UserID,
		)
		return
	}

	if !matchesTrigger(h.cfg.Triggers, msg) {
		logger.Debug("message did not match any trigger",
			"channel", h.cfg.ChannelName,
			"chat_id", msg.ChatID,
			"trigger_count", len(h.cfg.Triggers),
		)
		return
	}

	if msg.Text == "" && !msg.HasMedia {
		logger.Debug("ignoring message with no text or media",
			"channel", h.cfg.ChannelName,
			"chat_id", msg.ChatID,
		)
		return
	}

	if h.routeCommand(ctx, msg) {
		return
	}

	h.handleMessage(ctx, msg)
}

// handleMessage runs the active agent for a plain (non-command) message and
// delivers the response.
func (h *Handler) handleMessage(ctx context.Context, msg IncomingMessage) {
	logger := log.FromContext(ctx)
	agentName := h.ActiveAgent(ctx, msg.SessionID)

	logger.Info("dispatching message to agent",
		"channel", h.cfg.ChannelName,
		"agent", agentName,
		"session_id", msg.SessionID,
		"user_id", msg.UserID,
		"chat_id", msg.ChatID,
	)

	if h.cfg.SendTyping {
		h.transport.SendTyping(ctx, msg)
	}

	onEvent, onCompaction := h.debugCallbacks(ctx, msg)

	ctxInfo := h.buildContextInfo(msg)

	parts, err := msg.BuildParts(ctx)
	if err != nil {
		logger.Error("failed to build message parts", "channel", h.cfg.ChannelName, "err", err)
		h.transport.SendReply(ctx, msg, "⚠️ Sorry, I couldn't process the image in your message.")
		return
	}
	if len(parts) == 0 {
		logger.Debug("no input parts to send", "channel", h.cfg.ChannelName)
		return
	}

	processingMsgID := h.transport.SendProcessing(ctx, msg, agentName)

	modelOverride := h.ActiveModel(ctx, msg.SessionID)
	turn, err := h.runner.RunTurn(ctx, agentName, parts, modelOverride, ctxInfo, onEvent, onCompaction)
	if err != nil {
		logger.Error("agent run failed",
			"channel", h.cfg.ChannelName,
			"agent", agentName,
			"session_id", msg.SessionID,
			"err", err,
		)
		errText := "⚠️ Sorry, something went wrong processing your message."
		if processingMsgID != "" {
			h.transport.EditReply(ctx, msg, processingMsgID, agentName, errText)
		} else {
			h.transport.SendReply(ctx, msg, errText)
		}
		return
	}

	response := TurnResponseText(turn)
	if !TurnHasVisibleText(turn) {
		logger.Warn("agent turn completed without visible text",
			"channel", h.cfg.ChannelName,
			"agent", agentName,
			"session_id", msg.SessionID,
			"invocation_id", ctxInfo.GetUuid(),
			"event_count", turnEventCount(turn),
			"finish_reason", turnFinishReason(turn),
			"error_code", turnErrorCode(turn),
		)
	}

	logger.Info("agent response ready",
		"channel", h.cfg.ChannelName,
		"agent", agentName,
		"session_id", msg.SessionID,
		"response_len", len(response),
	)
	if processingMsgID != "" {
		h.transport.EditReply(ctx, msg, processingMsgID, agentName, response)
	} else {
		h.transport.SendReply(ctx, msg, response)
	}
}

// debugCallbacks returns runner callbacks that stream events/compaction to the
// transport when debug mode is active for this session, else (nil, nil).
func (h *Handler) debugCallbacks(ctx context.Context, msg IncomingMessage) (runner.EventCallback, runner.CompactionCallback) {
	if !h.debugActive(ctx, msg.SessionID) {
		return nil, nil
	}
	onEvent := func(evt *session.Event) {
		h.transport.SendDebugEvent(ctx, msg, evt)
	}
	onCompaction := func(agentName string) {
		h.transport.SendCompaction(ctx, msg, agentName)
	}
	return onEvent, onCompaction
}

// debugActive resolves whether debug mode is on: a per-session override takes
// precedence over the channel default.
func (h *Handler) debugActive(ctx context.Context, sessionID string) bool {
	if h.debug != nil {
		if override, err := h.debug.Get(ctx, h.cfg.ChannelName, sessionID); err == nil && override != nil {
			return *override
		}
	}
	return h.cfg.DebugDefault
}

// buildContextInfo assembles the ContextInfo passed to the runner.
func (h *Handler) buildContextInfo(msg IncomingMessage) *agentsv1.ContextInfo {
	metadata := make(map[string]string, len(msg.Metadata))
	maps.Copy(metadata, msg.Metadata)
	return &agentsv1.ContextInfo{
		Uuid:        uuid.Must(uuid.NewV7()).String(),
		SessionId:   msg.SessionID,
		UserId:      msg.UserID,
		ChannelName: h.cfg.ChannelName,
		Source:      agentsv1.ContextSource_CONTEXT_SOURCE_CHANNEL,
		ChatId:      msg.ChatID,
		ChannelType: msg.ChannelType,
		ChatType:    msg.ChatType,
		WorkspaceId: h.cfg.WorkspaceID,
		Metadata:    metadata,
	}
}

// ActiveAgent resolves the effective agent for a session: a per-session
// selection takes precedence over the channel default agent.
func (h *Handler) ActiveAgent(ctx context.Context, sessionID string) string {
	selected, err := h.agentSel.Get(ctx, h.cfg.ChannelName, sessionID)
	if err != nil {
		log.FromContext(ctx).Warn("failed to get agent selection, using default",
			"channel", h.cfg.ChannelName,
			"session_id", sessionID,
			"default_agent", h.cfg.DefaultAgent,
			"err", err,
		)
		return h.cfg.DefaultAgent
	}
	if selected == "" {
		return h.cfg.DefaultAgent
	}
	return selected
}

// ActiveModel resolves the effective model override for a session: a
// per-session selection takes precedence over the channel default model.
func (h *Handler) ActiveModel(ctx context.Context, sessionID string) string {
	if h.modelSel != nil {
		selected, err := h.modelSel.Get(ctx, h.cfg.ChannelName, sessionID)
		if err != nil {
			log.FromContext(ctx).Warn("failed to get model selection, using channel default",
				"channel", h.cfg.ChannelName,
				"session_id", sessionID,
				"err", err,
			)
		} else if selected != "" {
			return selected
		}
	}
	return h.cfg.DefaultModel
}
