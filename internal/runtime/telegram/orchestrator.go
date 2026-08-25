package telegram

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	"butterfly.orx.me/core/log"

	telegramrepo "go.orx.me/apps/butter/internal/repo/telegram"
	"go.orx.me/apps/butter/internal/repo/telegramprocessing"
	"go.orx.me/apps/butter/internal/runtime/runner"
	"go.orx.me/apps/butter/internal/telegramapi"
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
	// SupportsModelOverride reports whether Butter may choose a model for the
	// effective Agent. Pi Agents return false because pi owns its model on the
	// ButterBox.
	SupportsModelOverride(workspaceID, agentID string) (bool, bool)
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
	prefs   PreferenceStore
	// processing persists the auditable state of each accepted update.
	processing telegramprocessing.Repository
	// sessions serializes turns within one derived session.
	sessions SessionGuard
	// files downloads user-uploaded media at invocation time.
	files func(ctx context.Context, workspaceID, channelID string) (telegramapi.FileClient, error)
	// appName scopes ADK sessions for the Telegram entry point.
	appName            string
	processingLeaseTTL time.Duration
}

func NewOrchestrator(repo telegramrepo.Repository, sender *telegramsend.Sender, agents AgentRunner) *Orchestrator {
	return &Orchestrator{
		repo: repo, sender: sender, runner: agents, appName: "telegram",
		processingLeaseTTL: processingLeaseTTL,
	}
}

// SetSessionClearer wires `/clear`.
func (o *Orchestrator) SetSessionClearer(clearer SessionClearer) { o.session = clearer }

// SetProcessingRepo wires the durable processing state machine.
func (o *Orchestrator) SetProcessingRepo(repo telegramprocessing.Repository) { o.processing = repo }

// SetSessionGuard wires session serialization.
func (o *Orchestrator) SetSessionGuard(guard SessionGuard) { o.sessions = guard }

// SetFileClientFactory wires media downloads. Without it, photo updates are
// refused with a clear message rather than reaching the Agent as text alone.
func (o *Orchestrator) SetFileClientFactory(factory func(ctx context.Context, workspaceID, channelID string) (telegramapi.FileClient, error)) {
	o.files = factory
}

// SetPreferenceStore wires persisted Agent/Model selection. Without one,
// selection commands report that switching is unavailable rather than
// pretending to take effect for a single message.
func (o *Orchestrator) SetPreferenceStore(store PreferenceStore) { o.prefs = store }

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

	// Admission, trigger, and the accepted session subject do not depend on a
	// stored Agent/Model selection. Resolve them first so ignored updates never
	// contend on a session lease.
	decision := DecideInteraction(event, dest, channel.GetBotUsername(), Preferences{})
	if decision.Ignore != IgnoreNone {
		logger.Debug("telegram update produced no interaction", "reason", string(decision.Ignore))
		return nil
	}
	// Preference resolution needs its own short lease because changing Agent
	// changes the final session ID. Without this handoff lease, a message could
	// read the old preference, pause, let /agent complete, and then run the old
	// Agent after the switch had already been acknowledged.
	routingCtx := ctx
	releaseRouting := func() {}
	if o.sessions != nil {
		leaseCtx, release, ok, leaseErr := o.sessions.Acquire(ctx,
			RoutingLeaseID(event.ChannelID, dest.GetId(), decision.SessionSubject))
		if leaseErr != nil {
			return leaseErr
		}
		if !ok {
			logger.Debug("telegram session routing is busy; deferring",
				"session_subject", decision.SessionSubject)
			return ErrSessionBusy
		}
		routingCtx = leaseCtx
		releaseRouting = release
		defer releaseRouting()
	}

	// Load the stored selection while routing is serialized, so the effective
	// Agent and the final session lease are one coherent decision.
	stored := Preferences{}
	if o.prefs != nil {
		if loaded, prefErr := o.prefs.Get(routingCtx, decision.PreferenceKey); prefErr != nil {
			logger.Warn("could not read telegram preferences", "err", prefErr)
		} else {
			stored = loaded
		}
	}

	decision = DecideInteraction(event, dest, channel.GetBotUsername(), stored)
	if decision.Ignore != IgnoreNone {
		logger.Debug("telegram update produced no interaction", "reason", string(decision.Ignore))
		return nil
	}
	decision = o.applyAgentModelPolicy(event.WorkspaceID, decision)

	// Claiming derives the only safe recovery action from persisted state.
	record, action, leaseToken, err := o.claimRecord(routingCtx, event)
	if err != nil {
		return err
	}
	if action == telegramprocessing.ClaimAcknowledge {
		logger.Info("acknowledging telegram update without repeating completed or uncertain work",
			"record_id", record.GetId())
		return nil
	}
	defer o.releaseProcessingClaim(ctx, record, leaseToken)
	processingCtx, stopProcessingHeartbeat := o.withProcessingHeartbeat(ctx, record, leaseToken)
	defer stopProcessingHeartbeat()
	ctx = processingCtx
	if record != nil {
		logger = logger.With("record_id", record.GetId(), "invocation_id", record.GetInvocationId())
	}
	if dest.GetRevision() != event.DestinationRevision {
		// Not fatal — the re-read above already applied current policy — but
		// worth correlating when behavior surprises an operator.
		logger.Info("telegram destination changed after acceptance",
			"accepted_revision", event.DestinationRevision, "current_revision", dest.GetRevision())
	}
	if action == telegramprocessing.ClaimResumeDelivery {
		return o.resumeDelivery(ctx, event, record, leaseToken)
	}

	// A stored choice current configuration no longer allows is cleared, not
	// merely ignored, so the next turn does not pay to re-discover it.
	if decision.StaleSelection && o.prefs != nil {
		if err := o.prefs.Delete(ctx, decision.PreferenceKey); err != nil {
			logger.Warn("could not clear a stale telegram selection", "err", err)
		}
	}

	// Serialize every interaction for the derived session. Management commands
	// and callbacks can clear history or change routing while an Agent turn is
	// active, so letting them bypass this lease would race the conversation they
	// are intended to manage.
	if o.sessions != nil {
		leaseCtx, release, ok, leaseErr := o.sessions.Acquire(ctx, decision.SessionID)
		if leaseErr != nil {
			_ = o.recordFailure(ctx, record, leaseToken, leaseErr, false)
			return leaseErr
		}
		if !ok {
			// Leave it unacknowledged: it is reclaimed once the other worker
			// finishes, rather than interleaving two actions in one session.
			logger.Debug("telegram session is busy; deferring", "session_id", decision.SessionID)
			return ErrSessionBusy
		}
		defer release()
		ctx = leaseCtx
	}
	if decision.CallbackData != "" {
		ctx, stop := contextWithPeerCancellation(ctx, routingCtx)
		defer stop()
		err := o.handleCallback(ctx, event, dest, decision)
		return o.finishNonAgent(ctx, record, leaseToken, err)
	}
	if decision.Command != "" {
		ctx, stop := contextWithPeerCancellation(ctx, routingCtx)
		defer stop()
		err := o.handleCommand(ctx, event, dest, decision)
		return o.finishNonAgent(ctx, record, leaseToken, err)
	}

	if err := routingCtx.Err(); err != nil {
		return err
	}
	// The final Agent session lease now fences the resolved preference. Release
	// the short routing lease so a different Agent session can proceed in
	// parallel while this turn runs.
	releaseRouting()
	return o.runAgent(ctx, event, dest, decision, record, leaseToken)
}

func contextWithPeerCancellation(ctx, peer context.Context) (context.Context, func()) {
	merged, cancel := context.WithCancel(ctx)
	stopPeer := context.AfterFunc(peer, cancel)
	return merged, func() {
		stopPeer()
		cancel()
	}
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
		switch decision.Command {
		case "clear":
			return o.clearSession(ctx, event, decision)
		case "agent":
			return o.handleAgentCommand(ctx, event, dest, decision)
		case "model":
			return o.handleModelCommand(ctx, event, dest, decision)
		default:
			return o.handleDebugCommand(ctx, event, decision)
		}
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
func (o *Orchestrator) runAgent(ctx context.Context, event *telegramqueue.Event, dest *agentsv1.TelegramDestination, decision Interaction, record *agentsv1.TelegramProcessingRecord, leaseToken string) error {
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

	// Media download is pre-Agent work: transient failures retry, and a
	// permanent one fails without ever having started the Agent.
	var parts []*genai.Part
	partsErr := o.retryPreAgent(ctx, "prepare telegram agent input", func() error {
		built, err := o.inputParts(ctx, event, decision)
		if err != nil {
			return err
		}
		parts = built
		return nil
	})
	if partsErr != nil {
		// A photo that cannot be fetched must not reach the Agent as a
		// caption alone: the Agent would answer confidently about an image it
		// never saw. Report it in the topic and stop.
		logger.Warn("could not build telegram agent input", "err", partsErr)
		if recordErr := o.recordFailure(ctx, record, leaseToken, partsErr, false); recordErr != nil {
			return recordErr
		}
		_ = o.reply(ctx, event, decision, partsErr.Error())
		return nil
	}
	if len(parts) == 0 {
		return nil
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

	// A placeholder both acknowledges the message and becomes the first
	// segment of the answer, so a long reply never leaves a stranded
	// "working on it" above it.
	placeholder, placeholderErr := o.sender.SendProcessing(ctx, event.WorkspaceID, event.DestinationID,
		fmt.Sprintf("Working on it with %s…", agentName), decision.ReplyToMessageID)
	if placeholderErr != nil {
		logger.Debug("could not send a telegram placeholder", "err", placeholderErr)
	}

	// From here the Agent may run tools with external side effects, so a
	// crash is no longer safely retryable.
	if err := o.recordProgress(ctx, record, leaseToken, agentsv1.TelegramProcessingStatus_TELEGRAM_PROCESSING_STATUS_PROCESSING); err != nil {
		return err
	}

	turn, err := o.runner.RunTurnSSE(ctx, agentName, parts, decision.Model, ctxInfo, nil, nil)
	if err != nil {
		// Report the failure in the originating topic rather than leaving the
		// user waiting on silence, and mark the record uncertain: an
		// automatic rerun could repeat whatever the agent already did.
		_ = o.recordFailure(ctx, record, leaseToken, err, true)
		_ = o.deliver(ctx, event, decision, placeholder, "The agent could not complete this request.")
		return fmt.Errorf("run telegram agent turn: %w", err)
	}
	output := strings.TrimSpace(turn.Output)
	if output == "" {
		output = "(the agent produced no output)"
	}

	// Persist the complete response before delivery starts. A delivery
	// failure then resends text that already exists rather than re-running
	// the agent to reproduce it.
	delivery := telegramsend.NewDelivery(output, placeholder, decision.ReplyToMessageID)
	if len(delivery.Segments) == 0 {
		return o.recordProgress(ctx, record, leaseToken, agentsv1.TelegramProcessingStatus_TELEGRAM_PROCESSING_STATUS_SUCCEEDED)
	}
	if err := o.persistOutput(ctx, record, leaseToken, delivery); err != nil {
		_ = o.recordFailure(ctx, record, leaseToken, err, true)
		return err
	}

	deliverErr := o.sender.DeliverSegments(ctx, event.WorkspaceID, event.DestinationID, delivery,
		func(current *telegramsend.Delivery) error {
			return o.persistDeliveryProgress(ctx, record, leaseToken, current)
		})
	if syncErr := o.syncDelivery(ctx, record, leaseToken, delivery, deliverErr); syncErr != nil {
		return syncErr
	}
	if deliverErr != nil {
		return fmt.Errorf("deliver telegram response: %w", deliverErr)
	}
	return nil
}

// reply delivers text to the originating Destination. Every response —
// agent output, command answers, and errors alike — goes through the unified
// sender, which is what guarantees it stays in the originating Forum Topic.
func (o *Orchestrator) reply(ctx context.Context, event *telegramqueue.Event, decision Interaction, text string) error {
	// Command answers, status views, and errors take the same segmenting
	// path as agent output, so any of them exceeding Telegram's limit is
	// split rather than rejected.
	return o.deliver(ctx, event, decision, "", text)
}

// --- Selection ---------------------------------------------------------------

// resetArg is the argument that returns a choice to the Destination default.
const resetArg = "reset"

// handleAgentCommand switches, resets, or lists the selectable Agents.
//
// Selection is a controller action because it changes shared routing: in a
// DESTINATION-policy topic every admitted user's next message goes to the
// Agent that was chosen.
func (o *Orchestrator) handleAgentCommand(ctx context.Context, event *telegramqueue.Event, dest *agentsv1.TelegramDestination, decision Interaction) error {
	config := dest.GetConfig()
	if !AgentSelectionEnabled(config) {
		// An empty selectable list locks routing, which is the default.
		return o.reply(ctx, event, decision,
			"Agent selection is locked at this destination. Add selectable agents in the dashboard to enable it.")
	}
	if o.prefs == nil {
		return o.reply(ctx, event, decision, "Selection storage is not available.")
	}

	arg := strings.TrimSpace(decision.CommandArgs)
	switch {
	case arg == "":
		return o.reply(ctx, event, decision, listChoices("agent",
			config.GetSelectableAgentIds(), decision.AgentID, config.GetAgentId()))
	case strings.EqualFold(arg, resetArg):
		return o.applySelection(ctx, event, decision, func(p *Preferences) { p.AgentID = "" },
			fmt.Sprintf("Agent reset to the destination default (%s).", config.GetAgentId()))
	case !agentSelectable(config, arg):
		return o.reply(ctx, event, decision,
			fmt.Sprintf("%q is not selectable here. Choose one of: %s",
				arg, strings.Join(config.GetSelectableAgentIds(), ", ")))
	default:
		// Switching Agent moves the session, because the session key includes
		// the effective Agent — that is what keeps histories separate and lets
		// switching back resume the earlier conversation.
		return o.applySelection(ctx, event, decision, func(p *Preferences) { p.AgentID = arg },
			fmt.Sprintf("Agent set to %s. This agent keeps its own conversation history.", arg))
	}
}

// handleModelCommand switches, resets, or lists the selectable Models.
func (o *Orchestrator) handleModelCommand(ctx context.Context, event *telegramqueue.Event, dest *agentsv1.TelegramDestination, decision Interaction) error {
	config := dest.GetConfig()
	if !o.agentSupportsModelOverride(event.WorkspaceID, decision.AgentID) {
		return o.reply(ctx, event, decision,
			"Model selection is unavailable for Pi agents. Configure the model in the Agent's ButterBox settings.")
	}
	if !ModelSelectionEnabled(config) {
		return o.reply(ctx, event, decision,
			"Model selection is locked at this destination. Add selectable models in the dashboard to enable it.")
	}
	if o.prefs == nil {
		return o.reply(ctx, event, decision, "Selection storage is not available.")
	}

	arg := strings.TrimSpace(decision.CommandArgs)
	switch {
	case arg == "":
		return o.reply(ctx, event, decision, listChoices("model",
			config.GetSelectableModels(), decision.Model, config.GetModel()))
	case strings.EqualFold(arg, resetArg):
		return o.applySelection(ctx, event, decision, func(p *Preferences) { p.Model = "" },
			"Model reset to the destination default.")
	case !modelSelectable(config, arg):
		return o.reply(ctx, event, decision,
			fmt.Sprintf("%q is not selectable here. Choose one of: %s",
				arg, strings.Join(config.GetSelectableModels(), ", ")))
	default:
		// A Model switch deliberately does *not* move the session: changing
		// model should not silently start a new conversation.
		return o.applySelection(ctx, event, decision, func(p *Preferences) { p.Model = arg },
			fmt.Sprintf("Model set to %s. The current conversation continues.", arg))
	}
}

// handleDebugCommand toggles per-session debug output.
func (o *Orchestrator) handleDebugCommand(ctx context.Context, event *telegramqueue.Event, decision Interaction) error {
	if o.prefs == nil {
		return o.reply(ctx, event, decision, "Selection storage is not available.")
	}
	arg := strings.ToLower(strings.TrimSpace(decision.CommandArgs))
	enabled := !decision.Debug
	switch arg {
	case "on":
		enabled = true
	case "off":
		enabled = false
	case resetArg:
		return o.applySelection(ctx, event, decision, func(p *Preferences) { p.Debug = nil },
			"Debug reset to the destination default.")
	}
	state := "off"
	if enabled {
		state = "on"
	}
	return o.applySelection(ctx, event, decision,
		func(p *Preferences) { p.Debug = &enabled },
		fmt.Sprintf("Debug output is now %s for this conversation.", state))
}

// applySelection reads, mutates, and stores the selection, then confirms.
func (o *Orchestrator) applySelection(ctx context.Context, event *telegramqueue.Event, decision Interaction, mutate func(*Preferences), confirmation string) error {
	stored, err := o.prefs.Get(ctx, decision.PreferenceKey)
	if err != nil {
		return fmt.Errorf("read telegram preferences: %w", err)
	}
	mutate(&stored)
	if err := o.prefs.Put(ctx, decision.PreferenceKey, stored); err != nil {
		return fmt.Errorf("store telegram preferences: %w", err)
	}
	return o.reply(ctx, event, decision, confirmation)
}

// handleCallback revalidates an inline keyboard press against *current*
// policy before changing anything.
//
// A button rendered minutes ago encodes what was allowed then. Re-checking
// controller rights and candidate membership is what stops a stale button
// from switching routing after an operator revoked the permission.
func (o *Orchestrator) handleCallback(ctx context.Context, event *telegramqueue.Event, dest *agentsv1.TelegramDestination, decision Interaction) error {
	kind, value, ok := strings.Cut(decision.CallbackData, ":")
	if !ok {
		return nil
	}
	if !decision.IsController {
		return o.reply(ctx, event, decision,
			"Only a configured controller can change this destination's selection.")
	}
	config := dest.GetConfig()

	switch kind {
	case "agent":
		if !AgentSelectionEnabled(config) || !agentSelectable(config, value) {
			return o.reply(ctx, event, decision,
				"That agent is no longer selectable at this destination.")
		}
		return o.applySelection(ctx, event, decision, func(p *Preferences) { p.AgentID = value },
			fmt.Sprintf("Agent set to %s.", value))
	case "model":
		if !o.agentSupportsModelOverride(event.WorkspaceID, decision.AgentID) ||
			!ModelSelectionEnabled(config) || !modelSelectable(config, value) {
			return o.reply(ctx, event, decision,
				"That model is no longer selectable at this destination.")
		}
		return o.applySelection(ctx, event, decision, func(p *Preferences) { p.Model = value },
			fmt.Sprintf("Model set to %s.", value))
	default:
		return nil
	}
}

// applyAgentModelPolicy removes the Destination's Butter model choice when the
// effective Agent owns model selection elsewhere. Preferences stay intact so
// switching back to a normal Agent restores its model choice without changing
// either Agent's conversation history.
func (o *Orchestrator) applyAgentModelPolicy(workspaceID string, decision Interaction) Interaction {
	if !o.agentSupportsModelOverride(workspaceID, decision.AgentID) {
		decision.Model = ""
	}
	return decision
}

func (o *Orchestrator) agentSupportsModelOverride(workspaceID, agentID string) bool {
	if o.runner == nil {
		return true
	}
	supported, known := o.runner.SupportsModelOverride(workspaceID, agentID)
	return !known || supported
}

// listChoices renders the available options with the current one marked.
func listChoices(kind string, choices []string, current, fallback string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Selectable %ss:\n", kind)
	for _, choice := range choices {
		marker := "  "
		if choice == current {
			marker = "* "
		}
		fmt.Fprintf(&b, "%s%s\n", marker, choice)
	}
	if fallback != "" {
		fmt.Fprintf(&b, "\nDefault: %s", fallback)
	}
	fmt.Fprintf(&b, "\nUse /%s <name> to switch, or /%s reset.", kind, kind)
	return b.String()
}

// --- Media and segmented delivery -------------------------------------------

// inputParts builds the Agent input for a turn, downloading any photo at the
// last possible moment.
func (o *Orchestrator) inputParts(ctx context.Context, event *telegramqueue.Event, decision Interaction) ([]*genai.Part, error) {
	update, err := telegramapi.ParseUpdate(event.Update)
	if err != nil {
		return nil, nil
	}
	msg, ok := update.RoutableMessage()
	if !ok || len(msg.Photo) == 0 {
		if decision.Text == "" {
			return nil, nil
		}
		return []*genai.Part{{Text: decision.Text}}, nil
	}

	if o.files == nil {
		return nil, errors.New("Image input is not available on this deployment.")
	}
	client, err := o.files(ctx, event.WorkspaceID, event.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("Could not reach Telegram to download the image: %w", err)
	}
	photo, _ := LargestPhoto(msg.Photo)
	image, err := DownloadPhoto(ctx, client, photo)
	if err != nil {
		if errors.Is(err, ErrUnsupportedImage) {
			return nil, errors.New("That image type is not supported.")
		}
		return nil, fmt.Errorf("Could not download the image: %w", err)
	}
	return BuildInputParts(decision.Text, image), nil
}

// deliver persists nothing itself but splits the response and delivers it in
// order, editing the placeholder into the first segment.
//
// Segmentation lives behind the sender so debug, status, and error responses
// that happen to exceed Telegram's limit take the same safe path as agent
// output.
func (o *Orchestrator) deliver(ctx context.Context, event *telegramqueue.Event, decision Interaction, placeholder, text string) error {
	delivery := telegramsend.NewDelivery(text, placeholder, decision.ReplyToMessageID)
	if len(delivery.Segments) == 0 {
		return nil
	}
	if err := o.sender.DeliverSegments(ctx, event.WorkspaceID, event.DestinationID, delivery); err != nil {
		// Later segments stay pending, so a retry continues rather than
		// duplicating what already landed.
		return fmt.Errorf("deliver telegram response: %w", err)
	}
	return nil
}

// finishNonAgent closes out a command or callback, which never runs an Agent
// and therefore is always safely retryable on failure.
func (o *Orchestrator) finishNonAgent(ctx context.Context, record *agentsv1.TelegramProcessingRecord, leaseToken string, err error) error {
	if err != nil {
		if recordErr := o.recordFailure(ctx, record, leaseToken, err, false); recordErr != nil {
			return recordErr
		}
		return err
	}
	return o.recordProgress(ctx, record, leaseToken, agentsv1.TelegramProcessingStatus_TELEGRAM_PROCESSING_STATUS_SUCCEEDED)
}
