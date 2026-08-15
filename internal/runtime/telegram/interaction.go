package telegram

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"go.orx.me/apps/butter/internal/telegramapi"
	"go.orx.me/apps/butter/internal/telegramqueue"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// IgnoreReason explains why an accepted update produced no interaction. It is
// a named value rather than a bare bool so logs and tests can distinguish
// "not for us" from "not allowed".
type IgnoreReason string

const (
	IgnoreNone IgnoreReason = ""
	// IgnoreUnsupportedUpdate covers Bot-authored messages, anonymous
	// sender_chat posts, channel posts, automatic forwards, edited messages,
	// and service messages: addressable, but not user input.
	IgnoreUnsupportedUpdate IgnoreReason = "unsupported_update"
	// IgnoreNotTriggered means the Destination's trigger mode did not match.
	IgnoreNotTriggered IgnoreReason = "not_triggered"
	// IgnoreNotAdmitted means the sender is outside allowed_user_ids.
	IgnoreNotAdmitted IgnoreReason = "not_admitted"
	// IgnoreEmpty means there was nothing to send to the Agent.
	IgnoreEmpty IgnoreReason = "empty"
	// IgnoreDestinationUnavailable means the Destination is gone, disabled,
	// or can no longer reply.
	IgnoreDestinationUnavailable IgnoreReason = "destination_unavailable"
)

// Interaction is the decision the orchestrator reached for one update: what
// to run, as whom, in which session, and where to answer.
type Interaction struct {
	// Ignore is set when no Agent should run. Everything else is then unset.
	Ignore IgnoreReason

	// Command is the recognized management command, if any ("status",
	// "debug", "clear", …). Unrecognized commands are deliberately left
	// empty and their text passed to the Agent instead.
	Command string
	// CommandArgs is the remaining text after a recognized command.
	CommandArgs string
	// IsController reports whether the sender may run management commands.
	IsController bool

	// AgentID and Model are the effective routing for this turn.
	AgentID string
	Model   string
	// SessionID isolates history. It always includes the Channel, the
	// Destination, the session subject, and the effective Agent.
	SessionID string
	// SessionSubject is the Destination ID under DESTINATION policy and the
	// Telegram user ID under USER policy.
	SessionSubject string

	// Text is the message with any mention of this Bot removed.
	Text string
	// UserID is the Telegram sender.
	UserID string
	// MessageID is the inbound message, used when reply mode is REPLY.
	MessageID string
	// ReplyToMessageID is set only when the Destination replies by quoting.
	ReplyToMessageID string
	// DebugDefault carries the Destination's debug preference.
	DebugDefault bool
	// Debug is the effective debug state after any stored toggle.
	Debug bool
	// StaleSelection reports that a stored Agent or Model choice is no longer
	// allowed and was dropped, so the caller can clear it.
	StaleSelection bool
	// PreferenceKey addresses this subject's stored selection.
	PreferenceKey string
	// CallbackID and CallbackData are set when the update is an inline
	// keyboard press rather than a message.
	CallbackID   string
	CallbackData string
}

// DecideInteraction applies the Destination policy to one accepted update.
//
// It is a pure function of the event, the current Destination, the Bot
// identity, and any stored selection, which is what makes the whole
// admission/trigger/selection/session surface testable without Redis, Mongo,
// or a live Bot.
func DecideInteraction(event *telegramqueue.Event, dest *agentsv1.TelegramDestination, botUsername string, stored Preferences) Interaction {
	update, err := telegramapi.ParseUpdate(event.Update)
	if err != nil {
		return Interaction{Ignore: IgnoreUnsupportedUpdate}
	}
	msg, ok := update.RoutableMessage()
	if !ok {
		return Interaction{Ignore: IgnoreUnsupportedUpdate}
	}
	// An inline keyboard press carries the Bot's own message as context, so
	// the sender is the callback's `from`, not the message's.
	if update.CallbackQuery != nil {
		if update.CallbackQuery.From == nil || update.CallbackQuery.From.IsBot {
			return Interaction{Ignore: IgnoreUnsupportedUpdate}
		}
		msg = &telegramapi.Message_{
			MessageID:       msg.MessageID,
			MessageThreadID: msg.MessageThreadID,
			IsTopicMessage:  msg.IsTopicMessage,
			Chat:            msg.Chat,
			From:            update.CallbackQuery.From,
		}
	} else if !isUserInput(update, msg) {
		return Interaction{Ignore: IgnoreUnsupportedUpdate}
	}

	// Re-check the Destination against *current* state, not the snapshot: a
	// Destination disabled or made outbound-only since acceptance must not
	// produce a reply.
	if dest == nil || !dest.GetInboundEnabled() || !dest.GetOutboundEnabled() {
		return Interaction{Ignore: IgnoreDestinationUnavailable}
	}
	currentConfig := dest.GetConfig()
	config := acceptedPolicyConfig(event.Policy, currentConfig)

	userID := telegramapi.FormatID(msg.From.ID)
	// Acceptance policy is frozen, but revocation is immediate: a user must
	// have been admitted when the event entered the queue and still be admitted
	// when it runs. Adding a user later never grants an already queued update.
	if !admits(config.GetAllowedUserIds(), userID) ||
		!admits(currentConfig.GetAllowedUserIds(), userID) {
		return Interaction{Ignore: IgnoreNotAdmitted}
	}
	// A controller must also satisfy ordinary admission, so management
	// commands can never bypass the Destination's user policy. Controller
	// authority also requires both the accepted and current policy.
	isController := slices.Contains(config.GetControllerUserIds(), userID) &&
		slices.Contains(currentConfig.GetControllerUserIds(), userID)

	command, args, isCommand := telegramapi.Command(msg)
	if isCommand && !telegramapi.CommandTargetsBot(msg, botUsername) {
		// `/cmd@otherbot` in a multi-bot group is not ours.
		return Interaction{Ignore: IgnoreUnsupportedUpdate}
	}

	text, entities := messageTextAndEntities(msg)
	// A button press is an explicit invocation, so trigger mode does not
	// apply: requiring a mention to press a button the bot itself rendered
	// would make its own keyboards unusable.
	isCallback := update.CallbackQuery != nil
	mentioned := mentionsBot(msg, botUsername)
	if !isCallback && !triggered(config.GetTriggerMode(), mentioned, isCommand, msg) {
		return Interaction{Ignore: IgnoreNotTriggered}
	}

	// Strip our own mention so the Agent receives the intended prompt.
	// Mentions of other users stay: they are part of what was said.
	text = strings.TrimSpace(stripBotMentions(text, entities, botUsername))

	interaction := Interaction{
		IsController: isController,
		UserID:       userID,
		MessageID:    telegramapi.FormatID(msg.MessageID),
		DebugDefault: debugDefault(config),
		Debug:        debugDefault(config),
		Text:         text,
	}
	if stored.Debug != nil {
		interaction.Debug = *stored.Debug
	}
	if isCallback {
		interaction.CallbackID = update.CallbackQuery.ID
		interaction.CallbackData = update.CallbackQuery.Data
	}
	if isRecognizedCommand(command) {
		interaction.Command = command
		interaction.CommandArgs = args
	} else if isCommand {
		// An unrecognized command is ordinary Agent input, so agent-specific
		// slash commands remain possible.
		interaction.Text = strings.TrimSpace(text)
	}

	// A stored selection must be allowed by both the accepted snapshot and the
	// current configuration. The snapshot prevents later additions from
	// changing queued work; the current check makes removals immediate.
	acceptedStored := stored
	if stored.AgentID != "" && stored.AgentID != currentConfig.GetAgentId() &&
		!agentSelectable(currentConfig, stored.AgentID) {
		acceptedStored.AgentID = ""
	}
	if stored.Model != "" && stored.Model != currentConfig.GetModel() &&
		!modelSelectable(currentConfig, stored.Model) {
		acceptedStored.Model = ""
	}
	effective := ResolveEffective(config, acceptedStored)
	interaction.AgentID = effective.AgentID
	interaction.Model = effective.Model
	interaction.StaleSelection = effective.StaleAgent || effective.StaleModel ||
		acceptedStored.AgentID != stored.AgentID || acceptedStored.Model != stored.Model
	interaction.SessionSubject = sessionSubject(config.GetSessionPolicy(), dest.GetId(), userID)
	interaction.SessionID = SessionID(dest.GetChannelId(), dest.GetId(),
		interaction.SessionSubject, interaction.AgentID)
	interaction.PreferenceKey = PreferenceKey(dest.GetId(), interaction.SessionSubject)

	if config.GetReplyMode() == agentsv1.TelegramReplyMode_TELEGRAM_REPLY_MODE_REPLY {
		interaction.ReplyToMessageID = interaction.MessageID
	}

	if interaction.Command == "" && interaction.Text == "" &&
		interaction.CallbackData == "" && len(msg.Photo) == 0 {
		return Interaction{Ignore: IgnoreEmpty}
	}
	return interaction
}

// debugDefault resolves a Destination's debug preference. Unset means
// enabled: debug output is on unless an operator explicitly turns it off.
func debugDefault(config *agentsv1.TelegramDestinationConfig) bool {
	if config != nil && config.DebugDefault != nil {
		return *config.DebugDefault
	}
	return true
}

func acceptedPolicyConfig(policy *telegramqueue.DestinationPolicy, fallback *agentsv1.TelegramDestinationConfig) *agentsv1.TelegramDestinationConfig {
	if policy == nil {
		return fallback
	}
	return &agentsv1.TelegramDestinationConfig{
		AgentId:            policy.AgentID,
		Model:              policy.Model,
		SelectableAgentIds: slices.Clone(policy.SelectableAgentIDs),
		SelectableModels:   slices.Clone(policy.SelectableModels),
		TriggerMode:        triggerModeFromName(policy.TriggerMode),
		SessionPolicy:      sessionPolicyFromName(policy.SessionPolicy),
		AllowedUserIds:     slices.Clone(policy.AllowedUserIDs),
		ControllerUserIds:  slices.Clone(policy.ControllerUserIDs),
		ReplyMode:          replyModeFromName(policy.ReplyMode),
		DebugDefault:       policy.DebugDefault,
	}
}

func triggerModeFromName(name string) agentsv1.TelegramTriggerMode {
	return agentsv1.TelegramTriggerMode(agentsv1.TelegramTriggerMode_value[name])
}

func sessionPolicyFromName(name string) agentsv1.TelegramSessionPolicy {
	return agentsv1.TelegramSessionPolicy(agentsv1.TelegramSessionPolicy_value[name])
}

func replyModeFromName(name string) agentsv1.TelegramReplyMode {
	return agentsv1.TelegramReplyMode(agentsv1.TelegramReplyMode_value[name])
}

// isUserInput rejects every update shape that is addressable but is not a
// person speaking. Treating any of these as input would re-run agents on
// history edits, echo the Bot's own messages back into a conversation, or
// attribute an anonymous admin post to a user that does not exist.
func isUserInput(update *telegramapi.Update, msg *telegramapi.Message_) bool {
	switch {
	case update.EditedMessage != nil, update.ChannelPost != nil:
		return false
	case msg.IsAutomaticForward:
		return false
	case msg.SenderChat != nil:
		// An anonymous administrator or a channel speaking as itself.
		return false
	case msg.From == nil || msg.From.IsBot:
		return false
	case len(msg.NewChatMembers) > 0 || msg.LeftChatMember != nil || msg.PinnedMessage != nil:
		// Service messages.
		return false
	}
	return true
}

// admits applies the Destination's user policy. An empty list admits every
// real Telegram user that can reach the address, so an open topic needs no
// enumerated membership list.
func admits(allowed []string, userID string) bool {
	if len(allowed) == 0 {
		return true
	}
	return slices.Contains(allowed, userID)
}

// triggered evaluates the Destination's trigger mode.
func triggered(mode agentsv1.TelegramTriggerMode, mentioned, isCommand bool, msg *telegramapi.Message_) bool {
	switch mode {
	case agentsv1.TelegramTriggerMode_TELEGRAM_TRIGGER_MODE_MENTION:
		return mentioned
	case agentsv1.TelegramTriggerMode_TELEGRAM_TRIGGER_MODE_COMMAND:
		return isCommand
	case agentsv1.TelegramTriggerMode_TELEGRAM_TRIGGER_MODE_MENTION_OR_COMMAND:
		return mentioned || isCommand
	default:
		// ALL, and UNSPECIFIED which the service normalizes to ALL.
		return msg.Text != "" || msg.Caption != "" || len(msg.Photo) > 0
	}
}

// mentionsBot reports whether the message mentions this Bot, using Telegram's
// own entities and the current Bot username. Matching on raw text would fire
// on a similarly named user or on the username appearing inside a quote.
func mentionsBot(msg *telegramapi.Message_, botUsername string) bool {
	if botUsername == "" {
		return false
	}
	text, entities := messageTextAndEntities(msg)
	for _, entity := range entities {
		if entity.Type != "mention" {
			continue
		}
		mention, ok := telegramapi.SliceUTF16(text, entity.Offset, entity.Length)
		if !ok {
			continue
		}
		if strings.EqualFold(strings.TrimPrefix(mention, "@"), botUsername) {
			return true
		}
	}
	return false
}

func messageTextAndEntities(msg *telegramapi.Message_) (string, []telegramapi.MessageEntity) {
	if msg.Text != "" {
		return msg.Text, msg.Entities
	}
	return msg.Caption, msg.CaptionEntities
}

// stripBotMentions removes only Telegram entities that exactly name this Bot.
// Entity offsets are UTF-16 units, while Go string indexes are bytes.
func stripBotMentions(text string, entities []telegramapi.MessageEntity, botUsername string) string {
	if botUsername == "" {
		return text
	}
	type byteRange struct {
		start int
		end   int
	}
	ranges := make([]byteRange, 0, len(entities))
	seen := make(map[byteRange]struct{}, len(entities))
	for _, entity := range entities {
		if entity.Type != "mention" {
			continue
		}
		mention, ok := telegramapi.SliceUTF16(text, entity.Offset, entity.Length)
		if !ok || !strings.EqualFold(strings.TrimPrefix(mention, "@"), botUsername) {
			continue
		}
		prefix, prefixOK := telegramapi.SliceUTF16(text, 0, entity.Offset)
		through, throughOK := telegramapi.SliceUTF16(text, 0, entity.Offset+entity.Length)
		if !prefixOK || !throughOK {
			continue
		}
		span := byteRange{start: len(prefix), end: len(through)}
		if _, duplicate := seen[span]; duplicate {
			continue
		}
		seen[span] = struct{}{}
		ranges = append(ranges, span)
	}
	sort.Slice(ranges, func(i, j int) bool { return ranges[i].start > ranges[j].start })
	rightBoundary := len(text)
	for _, span := range ranges {
		if span.end > rightBoundary {
			continue
		}
		text = text[:span.start] + text[span.end:]
		rightBoundary = span.start
	}
	return text
}

// recognizedCommands are the management commands the runtime owns. Anything
// else is Agent input.
var recognizedCommands = []string{"status", "debug", "clear", "agent", "model"}

func isRecognizedCommand(command string) bool {
	return command != "" && slices.Contains(recognizedCommands, command)
}

// sessionSubject selects who the conversation belongs to inside one
// Destination.
func sessionSubject(policy agentsv1.TelegramSessionPolicy, destinationID, userID string) string {
	if policy == agentsv1.TelegramSessionPolicy_TELEGRAM_SESSION_POLICY_USER {
		return "u" + userID
	}
	return "d" + destinationID
}

// SessionID builds the session namespace.
//
// The Agent ID is part of the key on purpose: switching Agents must not let
// one Agent inherit another's conversation, and switching back must resume
// the earlier one. Channel and Destination are included so two topics under
// one Bot can never share history.
func SessionID(channelID, destinationID, subject, agentID string) string {
	return fmt.Sprintf("tg:%s:%s:%s:%s", channelID, destinationID, subject, agentID)
}

// RoutingLeaseID serializes preference resolution for one session subject.
// It is separate from the Agent session lease because changing Agent changes
// the session ID itself; the short routing lease closes that handoff window.
func RoutingLeaseID(channelID, destinationID, subject string) string {
	return fmt.Sprintf("tg-routing:%s:%s:%s", channelID, destinationID, subject)
}
