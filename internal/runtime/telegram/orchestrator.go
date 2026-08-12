package telegram

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	"butterfly.orx.me/core/log"

	telegramrepo "go.orx.me/apps/butter/internal/repo/telegram"
	"go.orx.me/apps/butter/internal/runtime/runner"
	"go.orx.me/apps/butter/internal/telegramqueue"
	"go.orx.me/apps/butter/internal/telegramsend"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// AgentRunner is the slice of the runner service the Telegram orchestrator
// needs. Declaring it here keeps the orchestrator testable without an ADK
// runtime — the routing and policy decisions are the part worth testing.
type AgentRunner interface {
	// ResolveAgentRef maps a workspace-scoped agent_id to a runnable name.
	ResolveAgentRef(workspaceID, agentID string) (string, bool)
	// RunTurnSSE runs one turn.
	RunTurnSSE(ctx context.Context, agentName string, parts []*genai.Part, modelOverride string,
		ctxInfo *agentsv1.ContextInfo, onEvent runner.EventCallback,
		onCompaction runner.CompactionCallback) (*runner.TurnResult, error)
}

// SessionClearer removes a session's history, backing `/clear`. It matches
// the ADK session service so the runtime's own store satisfies it directly.
type SessionClearer interface {
	Delete(ctx context.Context, req *session.DeleteRequest) error
}

// Orchestrator turns an accepted event into a Destination-scoped Agent
// interaction and delivers the response.
//
// It re-reads the Channel and Destination rather than trusting the queued
// snapshot for *authorization*: the snapshot records what was true at
// acceptance, but a Destination disabled or a user removed in the meantime
// must not get an Agent run. The snapshot still travels so a worker can see
// what changed.
type Orchestrator struct {
	repo    telegramrepo.Repository
	sender  *telegramsend.Sender
	runner  AgentRunner
	session SessionClearer
	// appName scopes ADK sessions for the Telegram entry point.
	appName string
}

func NewOrchestrator(repo telegramrepo.Repository, sender *telegramsend.Sender, agents AgentRunner) *Orchestrator {
	return &Orchestrator{repo: repo, sender: sender, runner: agents, appName: "telegram"}
}

// SetSessionClearer wires `/clear`.
func (o *Orchestrator) SetSessionClearer(clearer SessionClearer) { o.session = clearer }

// Handle implements EventHandler for destination-scoped updates.
func (o *Orchestrator) Handle(ctx context.Context, event *telegramqueue.Event) error {
	if event.Kind != telegramqueue.KindDestinationUpdate {
		return nil
	}
	logger := log.FromContext(ctx).With(
		"workspace_id", event.WorkspaceID,
		"channel_id", event.ChannelID,
		"destination_id", event.DestinationID,
		"update_id", event.UpdateID,
	)

	channel, err := o.repo.GetChannel(ctx, event.WorkspaceID, event.ChannelID)
	if err != nil {
		if errors.Is(err, telegramrepo.ErrNotFound) {
			logger.Info("skipping telegram update for a deleted channel")
			return nil
		}
		return err
	}
	dest, err := o.repo.GetDestination(ctx, event.WorkspaceID, event.DestinationID)
	if err != nil {
		if errors.Is(err, telegramrepo.ErrNotFound) {
			// A deleted Destination must not produce a reply, and retrying
			// cannot bring it back.
			logger.Info("skipping telegram update for a deleted destination")
			return nil
		}
		return err
	}

	decision := DecideInteraction(event, dest, channel.GetBotUsername())
	if decision.Ignore != IgnoreNone {
		logger.Debug("telegram update produced no interaction", "reason", string(decision.Ignore))
		return nil
	}
	if dest.GetRevision() != event.DestinationRevision {
		// Not fatal — the re-read above already applied current policy — but
		// worth correlating when behavior surprises an operator.
		logger.Info("telegram destination changed after acceptance",
			"accepted_revision", event.DestinationRevision, "current_revision", dest.GetRevision())
	}

	if decision.Command != "" {
		return o.handleCommand(ctx, event, dest, decision)
	}
	return o.runAgent(ctx, event, dest, decision)
}

// handleCommand answers a management command.
//
// `/status` is available to any admitted user because reading the effective
// configuration is not a privilege; everything that *changes* shared state
// requires controller authorization.
func (o *Orchestrator) handleCommand(ctx context.Context, event *telegramqueue.Event, dest *agentsv1.TelegramDestination, decision Interaction) error {
	switch decision.Command {
	case "status":
		return o.reply(ctx, event, decision, o.formatStatus(dest, decision))
	case "debug", "clear", "agent", "model":
		if !decision.IsController {
			return o.reply(ctx, event, decision,
				"Only a configured controller can run this command at this destination.")
		}
		if decision.Command == "clear" {
			return o.clearSession(ctx, event, decision)
		}
		// `/debug`, `/agent`, and `/model` gain behavior in a later ticket;
		// answering explicitly beats silently ignoring a controller.
		return o.reply(ctx, event, decision,
			fmt.Sprintf("`/%s` is not available yet at this destination.", decision.Command))
	default:
		return nil
	}
}

func (o *Orchestrator) clearSession(ctx context.Context, event *telegramqueue.Event, decision Interaction) error {
	if o.session == nil {
		return o.reply(ctx, event, decision, "Session storage is not available.")
	}
	if err := o.session.Delete(ctx, &session.DeleteRequest{
		AppName:   o.appName,
		UserID:    decision.SessionSubject,
		SessionID: decision.SessionID,
	}); err != nil {
		log.FromContext(ctx).Warn("could not clear telegram session",
			"session_id", decision.SessionID, "err", err)
		return o.reply(ctx, event, decision, "Could not clear the conversation.")
	}
	return o.reply(ctx, event, decision, "Conversation cleared.")
}

func (o *Orchestrator) formatStatus(dest *agentsv1.TelegramDestination, decision Interaction) string {
	var b strings.Builder
	b.WriteString("*Destination status*\n")
	fmt.Fprintf(&b, "destination: %s\n", dest.GetKey())
	fmt.Fprintf(&b, "agent: %s\n", decision.AgentID)
	if decision.Model != "" {
		fmt.Fprintf(&b, "model: %s\n", decision.Model)
	} else {
		b.WriteString("model: (agent default)\n")
	}
	fmt.Fprintf(&b, "session: %s\n", decision.SessionID)
	fmt.Fprintf(&b, "trigger: %s\n", dest.GetConfig().GetTriggerMode().String())
	fmt.Fprintf(&b, "history: %s", dest.GetConfig().GetSessionPolicy().String())
	return b.String()
}

// runAgent invokes the Destination's Agent and delivers the reply.
func (o *Orchestrator) runAgent(ctx context.Context, event *telegramqueue.Event, dest *agentsv1.TelegramDestination, decision Interaction) error {
	logger := log.FromContext(ctx)
	if o.runner == nil {
		return errors.New("agent runner is not configured")
	}
	agentName, ok := o.runner.ResolveAgentRef(event.WorkspaceID, decision.AgentID)
	if !ok {
		// The Destination service blocks deleting a referenced Agent, so this
		// means the runtime has not loaded it yet rather than that it is gone.
		return fmt.Errorf("agent %q is not available in workspace %q", decision.AgentID, event.WorkspaceID)
	}

	ctxInfo := &agentsv1.ContextInfo{
		SessionId:   decision.SessionID,
		UserId:      decision.UserID,
		ChatId:      event.Address.ChatID,
		ChannelName: dest.GetKey(),
		ChannelType: "telegram",
	}
	logger.Info("invoking telegram agent",
		"agent", agentName, "agent_id", decision.AgentID, "model", decision.Model,
		"session_id", decision.SessionID, "user_id", decision.UserID)

	turn, err := o.runner.RunTurnSSE(ctx, agentName,
		[]*genai.Part{{Text: decision.Text}}, decision.Model, ctxInfo, nil, nil)
	if err != nil {
		// Report the failure in the originating topic rather than leaving the
		// user waiting on silence.
		_ = o.reply(ctx, event, decision, "The agent could not complete this request.")
		return fmt.Errorf("run telegram agent turn: %w", err)
	}
	output := strings.TrimSpace(turn.Output)
	if output == "" {
		output = "(the agent produced no output)"
	}
	return o.reply(ctx, event, decision, output)
}

// reply delivers text to the originating Destination. Every response —
// agent output, command answers, and errors alike — goes through the unified
// sender, which is what guarantees it stays in the originating Forum Topic.
func (o *Orchestrator) reply(ctx context.Context, event *telegramqueue.Event, decision Interaction, text string) error {
	_, err := o.sender.Send(ctx, event.WorkspaceID, event.DestinationID, telegramsend.Message{
		Text:             text,
		ReplyToMessageID: decision.ReplyToMessageID,
	})
	if err != nil {
		return fmt.Errorf("deliver telegram reply: %w", err)
	}
	return nil
}
