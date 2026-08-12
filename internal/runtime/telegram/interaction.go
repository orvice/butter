package telegram

import (
	"fmt"
	"slices"
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
}

// DecideInteraction applies the Destination policy to one accepted update.
//
// It is a pure function of the event, the current Destination, and the Bot
// identity, which is what makes the whole admission/trigger/session surface
// testable without Redis, Mongo, or a live Bot.
func DecideInteraction(event *telegramqueue.Event, dest *agentsv1.TelegramDestination, botUsername string) Interaction {
	update, err := telegramapi.ParseUpdate(event.Update)
	if err != nil {
		return Interaction{Ignore: IgnoreUnsupportedUpdate}
	}
	msg, ok := update.RoutableMessage()
	if !ok || !isUserInput(update, msg) {
		return Interaction{Ignore: IgnoreUnsupportedUpdate}
	}

	// Re-check the Destination against *current* state, not the snapshot: a
	// Destination disabled or made outbound-only since acceptance must not
	// produce a reply.
	if dest == nil || !dest.GetInboundEnabled() || !dest.GetOutboundEnabled() {
		return Interaction{Ignore: IgnoreDestinationUnavailable}
	}
	config := dest.GetConfig()

	userID := telegramapi.FormatID(msg.From.ID)
	if !admits(config.GetAllowedUserIds(), userID) {
		return Interaction{Ignore: IgnoreNotAdmitted}
	}
	// A controller must also satisfy ordinary admission, so management
	// commands can never bypass the Destination's user policy.
	isController := slices.Contains(config.GetControllerUserIds(), userID)

	command, args, isCommand := telegramapi.Command(msg)
	if isCommand && !telegramapi.CommandTargetsBot(msg, botUsername) {
		// `/cmd@otherbot` in a multi-bot group is not ours.
		return Interaction{Ignore: IgnoreUnsupportedUpdate}
	}

	text := msg.Text
	if text == "" {
		text = msg.Caption
	}
	mentioned := mentionsBot(msg, botUsername)
	if !triggered(config.GetTriggerMode(), mentioned, isCommand, msg) {
		return Interaction{Ignore: IgnoreNotTriggered}
	}

	// Strip our own mention so the Agent receives the intended prompt.
	// Mentions of other users stay: they are part of what was said.
	text = strings.TrimSpace(stripBotMention(text, botUsername))

	interaction := Interaction{
		IsController: isController,
		UserID:       userID,
		MessageID:    telegramapi.FormatID(msg.MessageID),
		DebugDefault: config.GetDebugDefault(),
		Text:         text,
	}
	if isRecognizedCommand(command) {
		interaction.Command = command
		interaction.CommandArgs = args
	} else if isCommand {
		// An unrecognized command is ordinary Agent input, so agent-specific
		// slash commands remain possible.
		interaction.Text = strings.TrimSpace(text)
	}

	interaction.AgentID = config.GetAgentId()
	interaction.Model = config.GetModel()
	interaction.SessionSubject = sessionSubject(config.GetSessionPolicy(), dest.GetId(), userID)
	interaction.SessionID = SessionID(dest.GetChannelId(), dest.GetId(),
		interaction.SessionSubject, interaction.AgentID)

	if config.GetReplyMode() == agentsv1.TelegramReplyMode_TELEGRAM_REPLY_MODE_REPLY {
		interaction.ReplyToMessageID = interaction.MessageID
	}

	if interaction.Command == "" && interaction.Text == "" && len(msg.Photo) == 0 {
		return Interaction{Ignore: IgnoreEmpty}
	}
	return interaction
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
	text, entities := msg.Text, msg.Entities
	if text == "" {
		text, entities = msg.Caption, msg.CaptionEntities
	}
	runes := []rune(text)
	for _, entity := range entities {
		if entity.Type != "mention" {
			continue
		}
		end := min(entity.Offset+entity.Length, len(runes))
		if entity.Offset < 0 || entity.Offset > end {
			continue
		}
		if strings.EqualFold(strings.TrimPrefix(string(runes[entity.Offset:end]), "@"), botUsername) {
			return true
		}
	}
	return false
}

// stripBotMention removes "@botname" and the whitespace around it.
func stripBotMention(text, botUsername string) string {
	if botUsername == "" {
		return text
	}
	mention := "@" + botUsername
	lower := strings.ToLower(text)
	target := strings.ToLower(mention)
	for {
		at := strings.Index(lower, target)
		if at < 0 {
			return text
		}
		text = text[:at] + " " + text[at+len(mention):]
		lower = strings.ToLower(text)
	}
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
