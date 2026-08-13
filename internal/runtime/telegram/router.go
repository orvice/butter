// Package telegram is the Telegram receive runtime (issue #264): routing
// accepted updates into the durable queue, reconciling Webhook registration,
// and running the workers that answer them.
package telegram

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"butterfly.orx.me/core/log"

	telegramrepo "go.orx.me/apps/butter/internal/repo/telegram"
	"go.orx.me/apps/butter/internal/telegramapi"
	"go.orx.me/apps/butter/internal/telegramqueue"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Decision is what the receive path did with one update.
type Decision string

const (
	// DecisionAccepted means the event is durably queued.
	DecisionAccepted Decision = "accepted"
	// DecisionDuplicate means this (channel, update) pair was already
	// accepted. The caller acknowledges: re-delivering would double-run it.
	DecisionDuplicate Decision = "duplicate"
	// DecisionIgnored means the update is valid but not actionable — an
	// unknown address, or an update type Butter does not process. The caller
	// acknowledges so Telegram stops retrying.
	DecisionIgnored Decision = "ignored"
)

// whereCommand reports the Telegram identifiers of wherever it is run. It is
// handled before Destination matching precisely because its purpose is to let
// an administrator configure a Destination that does not exist yet.
const whereCommand = "where"

// Router turns a raw Telegram update into a queued event.
//
// It reloads Channel and Destination state from the database on every update
// rather than caching: a Destination that was just disabled must stop
// accepting immediately, and a cross-request cache would decide otherwise.
type Router struct {
	repo  telegramrepo.Repository
	queue Acceptor
	// now is injectable so tests can assert the accepted timestamp.
	now func() time.Time
}

// Acceptor is the durable hand-off the router writes to. Declaring it as an
// interface keeps routing decisions testable without Redis — the rules about
// which updates are accepted are the part worth testing, and they are
// independent of how the queue is implemented.
type Acceptor interface {
	Accept(ctx context.Context, event *telegramqueue.Event) (string, error)
}

func NewRouter(repo telegramrepo.Repository, queue Acceptor) *Router {
	return &Router{repo: repo, queue: queue, now: time.Now}
}

// SetClock overrides the timestamp source. Used by tests.
func (r *Router) SetClock(now func() time.Time) {
	if now != nil {
		r.now = now
	}
}

// Route accepts one raw update for a Channel.
//
// A returned error means "not accepted": the caller must answer Telegram with
// a retryable status rather than acknowledging, because acknowledging an
// update we failed to enqueue loses it permanently.
func (r *Router) Route(ctx context.Context, channel *agentsv1.TelegramChannel, raw []byte) (Decision, error) {
	logger := log.FromContext(ctx)

	update, err := telegramapi.ParseUpdate(raw)
	if err != nil {
		// Malformed JSON can never succeed on retry, so it is reported as
		// permanent and acknowledged rather than asked for again.
		return DecisionIgnored, fmt.Errorf("%w: %w", ErrMalformedUpdate, err)
	}
	msg, ok := update.RoutableMessage()
	if !ok {
		logger.Debug("ignoring unsupported telegram update type",
			"channel_id", channel.GetId(), "update_id", update.UpdateID)
		return DecisionIgnored, nil
	}
	chatID, threadID := telegramapi.AddressOf(msg)
	if chatID == "" {
		return DecisionIgnored, nil
	}

	// `/where` is recognized before Destination matching so an unconfigured
	// chat or topic can identify itself. It is unrestricted and invokes no
	// Agent, and it is the only path allowed to answer an address that no
	// Destination represents.
	if command, _, isCommand := telegramapi.Command(msg); isCommand && command == whereCommand &&
		telegramapi.CommandTargetsBot(msg, channel.GetBotUsername()) {
		return r.accept(ctx, &telegramqueue.Event{
			Kind:        telegramqueue.KindWhere,
			WorkspaceID: channel.GetWorkspaceId(),
			ChannelID:   channel.GetId(),
			BotID:       channel.GetBotId(),
			BotUsername: channel.GetBotUsername(),
			Address:     telegramqueue.Address{ChatID: chatID, MessageThreadID: threadID},
			UpdateID:    update.UpdateID,
			Update:      json.RawMessage(raw),
		})
	}

	dest, err := r.repo.FindDestinationByAddress(ctx, channel.GetId(), chatID, threadID)
	if err != nil {
		if errors.Is(err, telegramrepo.ErrNotFound) {
			// Unknown addresses are acknowledged and dropped rather than
			// queued: queueing them would let anyone who can find the bot
			// fill the stream.
			logger.Debug("ignoring telegram update for an unconfigured address",
				"channel_id", channel.GetId(), "chat_id", chatID, "message_thread_id", threadID)
			return DecisionIgnored, nil
		}
		// A database failure is not "unknown address" — it is "we could not
		// tell", which must be retried rather than silently dropped.
		return DecisionIgnored, fmt.Errorf("resolve telegram destination: %w", err)
	}
	if !dest.GetInboundEnabled() {
		logger.Debug("ignoring telegram update for an inbound-disabled destination",
			"channel_id", channel.GetId(), "destination_id", dest.GetId())
		return DecisionIgnored, nil
	}

	return r.accept(ctx, &telegramqueue.Event{
		Kind:                telegramqueue.KindDestinationUpdate,
		WorkspaceID:         channel.GetWorkspaceId(),
		ChannelID:           channel.GetId(),
		BotID:               channel.GetBotId(),
		BotUsername:         channel.GetBotUsername(),
		DestinationID:       dest.GetId(),
		DestinationRevision: dest.GetRevision(),
		Policy:              snapshotPolicy(dest.GetConfig()),
		Address:             telegramqueue.Address{ChatID: chatID, MessageThreadID: threadID},
		UpdateID:            update.UpdateID,
		Update:              json.RawMessage(raw),
	})
}

func (r *Router) accept(ctx context.Context, event *telegramqueue.Event) (Decision, error) {
	event.ReceivedAtUnixMs = r.now().UTC().UnixMilli()
	id, err := r.queue.Accept(ctx, event)
	if errors.Is(err, telegramqueue.ErrDuplicate) {
		log.FromContext(ctx).Debug("telegram update already accepted",
			"channel_id", event.ChannelID, "update_id", event.UpdateID)
		return DecisionDuplicate, nil
	}
	if err != nil {
		return DecisionIgnored, err
	}
	log.FromContext(ctx).Info("telegram update accepted",
		"channel_id", event.ChannelID, "destination_id", event.DestinationID,
		"update_id", event.UpdateID, "kind", string(event.Kind), "stream_id", id)
	return DecisionAccepted, nil
}

// snapshotPolicy freezes the Destination's inbound policy into the event.
//
// The worker uses the snapshot rather than re-reading configuration so a
// policy edit between acceptance and execution cannot change the rules an
// in-flight message is judged by. The revision travels alongside it, so a
// worker that does care about staleness can tell.
func snapshotPolicy(config *agentsv1.TelegramDestinationConfig) *telegramqueue.DestinationPolicy {
	if config == nil {
		return &telegramqueue.DestinationPolicy{}
	}
	return &telegramqueue.DestinationPolicy{
		AgentID:            config.GetAgentId(),
		Model:              config.GetModel(),
		SelectableAgentIDs: config.GetSelectableAgentIds(),
		SelectableModels:   config.GetSelectableModels(),
		TriggerMode:        config.GetTriggerMode().String(),
		SessionPolicy:      config.GetSessionPolicy().String(),
		AllowedUserIDs:     config.GetAllowedUserIds(),
		ControllerUserIDs:  config.GetControllerUserIds(),
		ReplyMode:          config.GetReplyMode().String(),
		DebugDefault:       config.GetDebugDefault(),
	}
}
