package pipeline

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"butterfly.orx.me/core/log"

	internalagent "go.orx.me/apps/butter/internal/agent"
)

// routeCommand handles slash commands and reports whether the message was a
// recognized command (and therefore fully handled here).
func (h *Handler) routeCommand(ctx context.Context, msg IncomingMessage) bool {
	switch {
	case strings.HasPrefix(msg.Text, "/agent"):
		h.handleAgentCommand(ctx, msg)
	case strings.HasPrefix(msg.Text, "/model"):
		h.transport.SendModelList(ctx, msg, h.ModelChoices(ctx, msg.SessionID))
	case strings.HasPrefix(msg.Text, "/debug"):
		h.handleDebugCommand(ctx, msg)
	case strings.HasPrefix(msg.Text, "/status"):
		h.transport.SendStatus(ctx, msg, h.buildStatusView(ctx, msg.SessionID, msg.UserID))
	case strings.HasPrefix(msg.Text, "/clear"):
		h.handleClearCommand(ctx, msg)
	default:
		return false
	}
	return true
}

func (h *Handler) handleAgentCommand(ctx context.Context, msg IncomingMessage) {
	logger := log.FromContext(ctx)
	sub, arg := parseAgentCommand(msg.Text)

	switch sub {
	case "list":
		h.transport.SendAgentList(ctx, msg, h.AgentChoices(ctx, msg.SessionID))
	case "switch":
		ok, err := h.SelectAgent(ctx, msg.SessionID, arg)
		if err != nil {
			logger.Error("failed to set agent selection",
				"channel", h.cfg.ChannelName, "session_id", msg.SessionID, "agent", arg, "err", err)
			h.transport.SendReply(ctx, msg, "❌ Failed to switch agent. Please try again.")
			return
		}
		if !ok {
			h.transport.SendReply(ctx, msg,
				fmt.Sprintf("❓ Unknown agent: %q\n\n📋 Available: %s", arg, strings.Join(h.cfg.AgentNames, ", ")))
			return
		}
		logger.Info("agent switched", "channel", h.cfg.ChannelName, "session_id", msg.SessionID, "agent", arg)
		h.transport.SendReply(ctx, msg, fmt.Sprintf("✅ Switched to agent: %s", arg))
	}
}

// ToggleDebug flips the per-session debug state and returns the new value. It
// is shared by the /debug command and the platform debug-toggle button.
func (h *Handler) ToggleDebug(ctx context.Context, sessionID string) (bool, error) {
	return h.debug.Toggle(ctx, h.cfg.ChannelName, sessionID, h.cfg.DebugDefault)
}

func (h *Handler) handleDebugCommand(ctx context.Context, msg IncomingMessage) {
	logger := log.FromContext(ctx)
	newState, err := h.ToggleDebug(ctx, msg.SessionID)
	if err != nil {
		logger.Error("failed to toggle debug mode",
			"channel", h.cfg.ChannelName, "session_id", msg.SessionID, "err", err)
		h.transport.SendReply(ctx, msg, "❌ Failed to toggle debug mode. Please try again.")
		return
	}
	logger.Info("debug mode toggled", "channel", h.cfg.ChannelName, "session_id", msg.SessionID, "debug", newState)
	h.transport.SendDebugStatus(ctx, msg, newState)
}

func (h *Handler) handleClearCommand(ctx context.Context, msg IncomingMessage) {
	logger := log.FromContext(ctx)
	if err := h.runner.ClearSession(ctx, h.cfg.ChannelName, msg.SessionID, msg.UserID); err != nil {
		logger.Error("failed to clear session",
			"channel", h.cfg.ChannelName, "session_id", msg.SessionID, "err", err)
		h.transport.SendReply(ctx, msg, "❌ Failed to clear session. Please try again.")
		return
	}
	logger.Info("session cleared", "channel", h.cfg.ChannelName, "session_id", msg.SessionID)
	h.transport.SendReply(ctx, msg, "🧹 Session cleared.")
}

// SelectAgent validates and stores a per-session agent selection. It returns
// ok=false (without storing) when the agent is unknown in the workspace.
func (h *Handler) SelectAgent(ctx context.Context, sessionID, agentName string) (ok bool, err error) {
	if !h.runner.HasAgentInWorkspace(h.cfg.WorkspaceID, agentName) {
		return false, nil
	}
	if err := h.agentSel.Set(ctx, h.cfg.ChannelName, sessionID, agentName); err != nil {
		return false, err
	}
	return true, nil
}

// SelectModel validates and stores a per-session model selection. It returns
// ok=false (without storing) when the alias is not among the configured models.
func (h *Handler) SelectModel(ctx context.Context, sessionID, alias string) (ok bool, err error) {
	if !slices.Contains(h.cfg.ModelNames, alias) {
		return false, nil
	}
	if h.modelSel == nil {
		return false, nil
	}
	if err := h.modelSel.Set(ctx, h.cfg.ChannelName, sessionID, alias); err != nil {
		return false, err
	}
	return true, nil
}

// AgentChoices lists all agents with the session's active one flagged.
func (h *Handler) AgentChoices(ctx context.Context, sessionID string) []AgentChoice {
	active := h.ActiveAgent(ctx, sessionID)
	choices := make([]AgentChoice, 0, len(h.cfg.AgentNames))
	for _, name := range h.cfg.AgentNames {
		choices = append(choices, AgentChoice{Name: name, Active: name == active})
	}
	return choices
}

// ModelChoices lists all models with the session's active one flagged.
func (h *Handler) ModelChoices(ctx context.Context, sessionID string) []ModelChoice {
	active := h.ActiveModel(ctx, sessionID)
	choices := make([]ModelChoice, 0, len(h.cfg.ModelNames))
	for _, alias := range h.cfg.ModelNames {
		choices = append(choices, ModelChoice{Alias: alias, Active: alias == active})
	}
	return choices
}

// buildStatusView gathers the platform-neutral data for a /status reply.
func (h *Handler) buildStatusView(ctx context.Context, sessionID, userID string) StatusView {
	activeAgent := h.ActiveAgent(ctx, sessionID)
	view := StatusView{
		AgentStatus: h.runner.GetAgentStatus(activeAgent),
		ActiveAgent: activeAgent,
		SessionID:   sessionID,
		Now:         time.Now(),
	}

	activeModel := h.ActiveModel(ctx, sessionID)
	view.ActiveModel = activeModel
	if activeModel != "" {
		if resolved, found := internalagent.ResolveModelAlias(activeModel, h.runner.ModelProviders()); found && resolved != activeModel {
			view.ResolvedModel = resolved
		}
	} else {
		view.AgentModel = h.runner.GetAgentModel(activeAgent)
	}

	sess, err := h.runner.GetSession(ctx, h.cfg.ChannelName, sessionID, userID)
	if err != nil {
		view.SessionErr = err
		return view
	}
	if sess == nil {
		return view
	}
	view.HasSession = true
	view.EventCount = sess.Events().Len()
	view.LastUpdate = sess.LastUpdateTime()
	return view
}

// parseAgentCommand parses "/agent <subcommand>" text.
// "/agent"/"/agent list" → ("list",""); "/agent foo" → ("switch","foo");
// non-/agent text → ("","").
func parseAgentCommand(text string) (subcommand, arg string) {
	text = strings.TrimSpace(text)
	if !strings.HasPrefix(text, "/agent") {
		return "", ""
	}
	fields := strings.Fields(text)
	if len(fields) == 1 {
		return "list", ""
	}
	if fields[1] == "list" {
		return "list", ""
	}
	return "switch", fields[1]
}
