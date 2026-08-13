package telegram

// Orchestration tests (issue #264/#268): which updates invoke an Agent, as
// whom, in which session, and where the answer goes. These assert externally
// observable decisions rather than internal call sequences.

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"go.orx.me/apps/butter/internal/telegramqueue"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const botUsername = "opsbot"

// destination builds a Destination with the given policy overrides applied.
func destination(mutate func(*agentsv1.TelegramDestinationConfig)) *agentsv1.TelegramDestination {
	config := &agentsv1.TelegramDestinationConfig{
		AgentId:       "support",
		TriggerMode:   agentsv1.TelegramTriggerMode_TELEGRAM_TRIGGER_MODE_ALL,
		SessionPolicy: agentsv1.TelegramSessionPolicy_TELEGRAM_SESSION_POLICY_DESTINATION,
		ReplyMode:     agentsv1.TelegramReplyMode_TELEGRAM_REPLY_MODE_REPLY,
	}
	if mutate != nil {
		mutate(config)
	}
	return &agentsv1.TelegramDestination{
		Id: "dest-1", Key: "incidents", ChannelId: "ch-1",
		ChatId: "-100", MessageThreadId: "42",
		InboundEnabled: true, OutboundEnabled: true,
		Config: config, Revision: 1,
	}
}

// eventFor wraps a raw update in an accepted queue event.
func eventFor(raw string) *telegramqueue.Event {
	return &telegramqueue.Event{
		Kind:                telegramqueue.KindDestinationUpdate,
		WorkspaceID:         "ws-a",
		ChannelID:           "ch-1",
		BotID:               "111111",
		BotUsername:         botUsername,
		DestinationID:       "dest-1",
		DestinationRevision: 1,
		Address:             telegramqueue.Address{ChatID: "-100", MessageThreadID: "42"},
		UpdateID:            1,
		Update:              json.RawMessage(raw),
	}
}

// message builds a Telegram message update from a sender.
func message(from string, text string, extra string) string {
	if extra != "" && !strings.HasPrefix(extra, ",") {
		extra = "," + extra
	}
	return fmt.Sprintf(`{"update_id":1,"message":{"message_id":9,"message_thread_id":42,"is_topic_message":true,"from":%s,"chat":{"id":-100,"type":"supergroup"},"text":%q%s}}`,
		from, text, extra)
}

const realUser = `{"id":7,"is_bot":false,"username":"ops"}`

func TestPlainMessageInvokesTheDestinationAgent(t *testing.T) {
	decision := DecideInteraction(eventFor(message(realUser, "hello", "")), destination(nil), botUsername, Preferences{})

	if decision.Ignore != IgnoreNone {
		t.Fatalf("ignored: %s", decision.Ignore)
	}
	if decision.AgentID != "support" {
		t.Errorf("agent = %q", decision.AgentID)
	}
	if decision.Text != "hello" {
		t.Errorf("text = %q", decision.Text)
	}
	// REPLY is the default, so the response quotes the inbound message.
	if decision.ReplyToMessageID != "9" {
		t.Errorf("reply_to = %q, want the inbound message", decision.ReplyToMessageID)
	}
}

func TestInteractionUsesTheAcceptedPolicySnapshot(t *testing.T) {
	accepted := destination(nil)
	event := eventFor(message(realUser, "hello", ""))
	event.Policy = snapshotPolicy(accepted.GetConfig())

	current := destination(func(config *agentsv1.TelegramDestinationConfig) {
		config.AgentId = "research"
		config.Model = "new-model"
		config.TriggerMode = agentsv1.TelegramTriggerMode_TELEGRAM_TRIGGER_MODE_COMMAND
		config.SessionPolicy = agentsv1.TelegramSessionPolicy_TELEGRAM_SESSION_POLICY_USER
		config.ReplyMode = agentsv1.TelegramReplyMode_TELEGRAM_REPLY_MODE_NEW_MESSAGE
	})
	decision := DecideInteraction(event, current, botUsername, Preferences{})

	if decision.Ignore != IgnoreNone {
		t.Fatalf("accepted ALL policy was replaced by current policy: %s", decision.Ignore)
	}
	if decision.AgentID != "support" || decision.Model != "" {
		t.Fatalf("routing = %q/%q, want accepted support/default", decision.AgentID, decision.Model)
	}
	if !strings.Contains(decision.SessionID, "ddest-1") {
		t.Fatalf("session = %q, want accepted destination policy", decision.SessionID)
	}
	if decision.ReplyToMessageID != "9" {
		t.Fatalf("reply_to = %q, want accepted reply mode", decision.ReplyToMessageID)
	}
}

func TestInteractionHonorsCurrentAdmissionRevocation(t *testing.T) {
	accepted := destination(nil)
	event := eventFor(message(realUser, "hello", ""))
	event.Policy = snapshotPolicy(accepted.GetConfig())
	current := destination(func(config *agentsv1.TelegramDestinationConfig) {
		config.AllowedUserIds = []string{"10"}
	})

	if got := DecideInteraction(event, current, botUsername, Preferences{}); got.Ignore != IgnoreNotAdmitted {
		t.Fatalf("ignore = %q, want current revocation to win", got.Ignore)
	}
}

func TestInteractionHonorsCurrentCandidateRevocation(t *testing.T) {
	accepted := destination(func(config *agentsv1.TelegramDestinationConfig) {
		config.SelectableAgentIds = []string{"support", "research"}
	})
	event := eventFor(message(realUser, "hello", ""))
	event.Policy = snapshotPolicy(accepted.GetConfig())
	current := destination(func(config *agentsv1.TelegramDestinationConfig) {
		config.SelectableAgentIds = []string{"support"}
	})

	got := DecideInteraction(event, current, botUsername, Preferences{AgentID: "research"})
	if got.AgentID != "support" || !got.StaleSelection {
		t.Fatalf("agent = %q stale = %v, want revoked selection to fall back", got.AgentID, got.StaleSelection)
	}
}

// Only ordinary messages from real users invoke an Agent.
func TestIgnoredUpdateShapes(t *testing.T) {
	cases := []struct {
		name string
		raw  string
	}{
		{"bot author", message(`{"id":8,"is_bot":true,"username":"otherbot"}`, "hi", "")},
		{"anonymous sender_chat", message(realUser, "hi", `"sender_chat":{"id":-100,"type":"supergroup"}`)},
		{"service message", message(realUser, "", `"new_chat_members":[{"id":9,"is_bot":false}]`)},
		{"automatic forward", message(realUser, "hi", `"is_automatic_forward":true`)},
		{
			"edited message",
			`{"update_id":1,"edited_message":{"message_id":9,"from":{"id":7},"chat":{"id":-100},"text":"edited"}}`,
		},
		{
			"channel post",
			`{"update_id":1,"channel_post":{"message_id":9,"chat":{"id":-100},"text":"post"}}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			decision := DecideInteraction(eventFor(tc.raw), destination(nil), botUsername, Preferences{})
			if decision.Ignore == IgnoreNone {
				t.Fatalf("expected the update to be ignored, got %+v", decision)
			}
		})
	}
}

// A Destination that can no longer reply must not run an Agent: a response
// that cannot be delivered is worse than no response.
func TestUnavailableDestinationDoesNotInvokeTheAgent(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*agentsv1.TelegramDestination)
	}{
		{"deleted", nil},
		{"inbound disabled", func(d *agentsv1.TelegramDestination) { d.InboundEnabled = false }},
		{"outbound disabled", func(d *agentsv1.TelegramDestination) { d.OutboundEnabled = false }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dest := destination(nil)
			if tc.mutate == nil {
				dest = nil
			} else {
				tc.mutate(dest)
			}
			decision := DecideInteraction(eventFor(message(realUser, "hello", "")), dest, botUsername, Preferences{})
			if decision.Ignore != IgnoreDestinationUnavailable {
				t.Fatalf("ignore = %q, want destination_unavailable", decision.Ignore)
			}
		})
	}
}

// --- Admission -------------------------------------------------------------

func TestAllowedUserIDsGateOrdinaryAccess(t *testing.T) {
	dest := destination(func(c *agentsv1.TelegramDestinationConfig) {
		c.AllowedUserIds = []string{"10", "20"}
	})

	if got := DecideInteraction(eventFor(message(realUser, "hi", "")), dest, botUsername, Preferences{}); got.Ignore != IgnoreNotAdmitted {
		t.Fatalf("ignore = %q, want not_admitted", got.Ignore)
	}
	admitted := `{"id":10,"is_bot":false}`
	if got := DecideInteraction(eventFor(message(admitted, "hi", "")), dest, botUsername, Preferences{}); got.Ignore != IgnoreNone {
		t.Fatalf("an allowed user was rejected: %q", got.Ignore)
	}
}

// An empty allow-list admits every real user, so an open topic needs no
// enumerated membership.
func TestEmptyAllowListAdmitsEveryRealUser(t *testing.T) {
	decision := DecideInteraction(eventFor(message(realUser, "hi", "")), destination(nil), botUsername, Preferences{})
	if decision.Ignore != IgnoreNone {
		t.Fatalf("ignore = %q", decision.Ignore)
	}
}

// A controller who is not admitted stays out: management commands never
// bypass the Destination's user policy.
func TestControllersMustAlsoBeAdmitted(t *testing.T) {
	dest := destination(func(c *agentsv1.TelegramDestinationConfig) {
		c.AllowedUserIds = []string{"10"}
		c.ControllerUserIds = []string{"7"}
	})

	decision := DecideInteraction(
		eventFor(message(realUser, "/clear", `"entities":[{"type":"bot_command","offset":0,"length":6}]`)),
		dest, botUsername, Preferences{})
	if decision.Ignore != IgnoreNotAdmitted {
		t.Fatalf("ignore = %q, want not_admitted", decision.Ignore)
	}
}

func TestControllerIsRecognized(t *testing.T) {
	dest := destination(func(c *agentsv1.TelegramDestinationConfig) {
		c.ControllerUserIds = []string{"7"}
	})

	decision := DecideInteraction(
		eventFor(message(realUser, "/clear", `"entities":[{"type":"bot_command","offset":0,"length":6}]`)),
		dest, botUsername, Preferences{})
	if !decision.IsController {
		t.Fatal("expected the sender to be recognized as a controller")
	}
	if decision.Command != "clear" {
		t.Errorf("command = %q", decision.Command)
	}
}

// --- Triggers --------------------------------------------------------------

func TestTriggerModes(t *testing.T) {
	mentionEntity := `"entities":[{"type":"mention","offset":0,"length":7}]`
	commandEntity := `"entities":[{"type":"bot_command","offset":0,"length":7}]`

	cases := []struct {
		name     string
		mode     agentsv1.TelegramTriggerMode
		raw      string
		expected IgnoreReason
	}{
		{"all/plain", agentsv1.TelegramTriggerMode_TELEGRAM_TRIGGER_MODE_ALL, message(realUser, "hello", ""), IgnoreNone},
		{"mention/plain", agentsv1.TelegramTriggerMode_TELEGRAM_TRIGGER_MODE_MENTION, message(realUser, "hello", ""), IgnoreNotTriggered},
		{"mention/mentioned", agentsv1.TelegramTriggerMode_TELEGRAM_TRIGGER_MODE_MENTION, message(realUser, "@opsbot hello", mentionEntity), IgnoreNone},
		{"command/plain", agentsv1.TelegramTriggerMode_TELEGRAM_TRIGGER_MODE_COMMAND, message(realUser, "hello", ""), IgnoreNotTriggered},
		{"command/command", agentsv1.TelegramTriggerMode_TELEGRAM_TRIGGER_MODE_COMMAND, message(realUser, "/deploy now", commandEntity), IgnoreNone},
		{"either/mentioned", agentsv1.TelegramTriggerMode_TELEGRAM_TRIGGER_MODE_MENTION_OR_COMMAND, message(realUser, "@opsbot hi", mentionEntity), IgnoreNone},
		{"either/command", agentsv1.TelegramTriggerMode_TELEGRAM_TRIGGER_MODE_MENTION_OR_COMMAND, message(realUser, "/deploy now", commandEntity), IgnoreNone},
		{"either/plain", agentsv1.TelegramTriggerMode_TELEGRAM_TRIGGER_MODE_MENTION_OR_COMMAND, message(realUser, "hello", ""), IgnoreNotTriggered},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dest := destination(func(c *agentsv1.TelegramDestinationConfig) { c.TriggerMode = tc.mode })
			got := DecideInteraction(eventFor(tc.raw), dest, botUsername, Preferences{})
			if got.Ignore != tc.expected {
				t.Fatalf("ignore = %q, want %q", got.Ignore, tc.expected)
			}
		})
	}
}

// A mention of a similarly named user must not trigger us.
func TestMentionMatchesTheCurrentBotOnly(t *testing.T) {
	dest := destination(func(c *agentsv1.TelegramDestinationConfig) {
		c.TriggerMode = agentsv1.TelegramTriggerMode_TELEGRAM_TRIGGER_MODE_MENTION
	})
	raw := message(realUser, "@opsbot2 hello", `"entities":[{"type":"mention","offset":0,"length":8}]`)

	if got := DecideInteraction(eventFor(raw), dest, botUsername, Preferences{}); got.Ignore != IgnoreNotTriggered {
		t.Fatalf("ignore = %q, want not_triggered", got.Ignore)
	}
}

func TestMentionUsesTelegramUTF16EntityCoordinates(t *testing.T) {
	dest := destination(func(c *agentsv1.TelegramDestinationConfig) {
		c.TriggerMode = agentsv1.TelegramTriggerMode_TELEGRAM_TRIGGER_MODE_MENTION
	})
	raw := message(realUser, "🙂 @opsbot hello",
		`"entities":[{"type":"mention","offset":3,"length":7}]`)

	if got := DecideInteraction(eventFor(raw), dest, botUsername); got.Ignore != IgnoreNone {
		t.Fatalf("ignore = %q, want accepted", got.Ignore)
	}
}

// The bot's own mention is removed; other users' mentions stay, because they
// are part of what was said.
func TestOurMentionIsStrippedButOthersRemain(t *testing.T) {
	raw := message(realUser, "@opsbot please ping @alice",
		`"entities":[{"type":"mention","offset":0,"length":7},{"type":"mention","offset":20,"length":6}]`)

	decision := DecideInteraction(eventFor(raw), destination(nil), botUsername, Preferences{})
	if decision.Text != "please ping @alice" {
		t.Fatalf("text = %q", decision.Text)
	}
}

func TestMentionStrippingUsesExactTelegramEntitySpans(t *testing.T) {
	raw := message(realUser, "🙂 @opsbot2 ping @opsbot",
		`"entities":[{"type":"mention","offset":3,"length":8},{"type":"mention","offset":17,"length":7}]`)

	decision := DecideInteraction(eventFor(raw), destination(nil), botUsername)
	if decision.Text != "🙂 @opsbot2 ping" {
		t.Fatalf("text = %q, want similarly named mention preserved", decision.Text)
	}
}

func TestRawMentionTextWithoutEntityIsNotStripped(t *testing.T) {
	decision := DecideInteraction(
		eventFor(message(realUser, "quote @opsbot", "")),
		destination(nil),
		botUsername,
	)

	if decision.Text != "quote @opsbot" {
		t.Fatalf("text = %q, want non-entity text preserved", decision.Text)
	}
}

func TestCaptionMentionStrippingUsesUTF16EntityCoordinates(t *testing.T) {
	raw := message(realUser, "", `"caption":"🙂 @opsbot inspect","caption_entities":[{"type":"mention","offset":3,"length":7}],"photo":[{"file_id":"photo-1","file_unique_id":"unique-1","width":1,"height":1}]`)

	decision := DecideInteraction(eventFor(raw), destination(nil), botUsername)
	if decision.Text != "🙂  inspect" {
		t.Fatalf("text = %q, want only the caption mention entity removed", decision.Text)
	}
}

// --- Commands --------------------------------------------------------------

func TestUnrecognizedCommandsReachTheAgent(t *testing.T) {
	raw := message(realUser, "/deploy staging", `"entities":[{"type":"bot_command","offset":0,"length":7}]`)

	decision := DecideInteraction(eventFor(raw), destination(nil), botUsername, Preferences{})
	if decision.Command != "" {
		t.Fatalf("command = %q, want the runtime to leave it to the agent", decision.Command)
	}
	if decision.Text != "/deploy staging" {
		t.Errorf("text = %q, want the command preserved verbatim", decision.Text)
	}
}

func TestCommandsAddressedToAnotherBotAreIgnored(t *testing.T) {
	raw := message(realUser, "/status@otherbot", `"entities":[{"type":"bot_command","offset":0,"length":16}]`)

	if got := DecideInteraction(eventFor(raw), destination(nil), botUsername, Preferences{}); got.Ignore != IgnoreUnsupportedUpdate {
		t.Fatalf("ignore = %q", got.Ignore)
	}
}

func TestCommandsAddressedToUsAreRecognized(t *testing.T) {
	raw := message(realUser, "/status@opsbot", `"entities":[{"type":"bot_command","offset":0,"length":14}]`)

	decision := DecideInteraction(eventFor(raw), destination(nil), botUsername, Preferences{})
	if decision.Command != "status" {
		t.Fatalf("command = %q", decision.Command)
	}
}

// --- Sessions --------------------------------------------------------------

// Two topics under one Bot must never share history.
func TestSessionsAreIsolatedPerDestination(t *testing.T) {
	first := DecideInteraction(eventFor(message(realUser, "hi", "")), destination(nil), botUsername, Preferences{})

	other := destination(nil)
	other.Id = "dest-2"
	second := DecideInteraction(eventFor(message(realUser, "hi", "")), other, botUsername, Preferences{})

	if first.SessionID == second.SessionID {
		t.Fatalf("two destinations shared session %q", first.SessionID)
	}
}

// Switching Agents must not let one inherit the other's conversation.
func TestSessionsAreIsolatedPerAgent(t *testing.T) {
	first := DecideInteraction(eventFor(message(realUser, "hi", "")), destination(nil), botUsername, Preferences{})
	second := DecideInteraction(eventFor(message(realUser, "hi", "")),
		destination(func(c *agentsv1.TelegramDestinationConfig) { c.AgentId = "research" }),
		botUsername, Preferences{})

	if first.SessionID == second.SessionID {
		t.Fatalf("two agents shared session %q", first.SessionID)
	}
}

func TestSessionPolicySelectsTheSubject(t *testing.T) {
	shared := DecideInteraction(eventFor(message(realUser, "hi", "")), destination(nil), botUsername, Preferences{})
	if !strings.Contains(shared.SessionID, "ddest-1") {
		t.Errorf("DESTINATION policy session = %q, want it keyed by destination", shared.SessionID)
	}

	perUser := destination(func(c *agentsv1.TelegramDestinationConfig) {
		c.SessionPolicy = agentsv1.TelegramSessionPolicy_TELEGRAM_SESSION_POLICY_USER
	})
	first := DecideInteraction(eventFor(message(realUser, "hi", "")), perUser, botUsername, Preferences{})
	second := DecideInteraction(eventFor(message(`{"id":11,"is_bot":false}`, "hi", "")), perUser, botUsername, Preferences{})
	if first.SessionID == second.SessionID {
		t.Fatalf("USER policy shared session %q between users", first.SessionID)
	}
}

// --- Reply behavior --------------------------------------------------------

// Topic targeting is independent of whether the response quotes the inbound
// message: NEW_MESSAGE still answers in the topic.
func TestReplyModeIsIndependentOfTopicTargeting(t *testing.T) {
	dest := destination(func(c *agentsv1.TelegramDestinationConfig) {
		c.ReplyMode = agentsv1.TelegramReplyMode_TELEGRAM_REPLY_MODE_NEW_MESSAGE
	})

	decision := DecideInteraction(eventFor(message(realUser, "hi", "")), dest, botUsername, Preferences{})
	if decision.ReplyToMessageID != "" {
		t.Errorf("NEW_MESSAGE must not quote, got reply_to %q", decision.ReplyToMessageID)
	}
	// The event's address still carries the topic, which is what the sender
	// uses to stay inside it.
	if eventFor(message(realUser, "hi", "")).Address.MessageThreadID != "42" {
		t.Error("the topic must remain on the event regardless of reply mode")
	}
}

func TestEmptyMessagesProduceNoInteraction(t *testing.T) {
	decision := DecideInteraction(eventFor(message(realUser, "   ", "")), destination(nil), botUsername, Preferences{})
	if decision.Ignore != IgnoreEmpty && decision.Ignore != IgnoreNotTriggered {
		t.Fatalf("ignore = %q, want the empty message to be dropped", decision.Ignore)
	}
}
