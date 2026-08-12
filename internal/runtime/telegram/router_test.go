package telegram

// Routing tests for the Telegram receive path (issue #264/#267): exact Topic
// matching, `/where` before Destination matching, unknown addresses, ignored
// update types, and the retryable/permanent error boundary.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	telegramrepo "go.orx.me/apps/butter/internal/repo/telegram"
	telegrammemory "go.orx.me/apps/butter/internal/repo/telegram/memory"
	"go.orx.me/apps/butter/internal/telegramqueue"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// fakeQueue records accepted events without Redis. It implements the same
// contract the router depends on: duplicates are reported, failures are not
// swallowed.
type fakeQueue struct {
	accepted []*telegramqueue.Event
	seen     map[string]bool
	failWith error
}

func newFakeQueue() *fakeQueue { return &fakeQueue{seen: map[string]bool{}} }

func (q *fakeQueue) Accept(_ context.Context, event *telegramqueue.Event) (string, error) {
	if q.failWith != nil {
		return "", q.failWith
	}
	key := fmt.Sprintf("%s:%d", event.ChannelID, event.UpdateID)
	if q.seen[key] {
		return "", telegramqueue.ErrDuplicate
	}
	q.seen[key] = true
	q.accepted = append(q.accepted, event)
	return fmt.Sprintf("stream-%d", len(q.accepted)), nil
}

// routerFixture wires a router over a memory repository and a fake queue.
type routerFixture struct {
	router  *Router
	repo    *telegrammemory.Store
	queue   *fakeQueue
	channel *agentsv1.TelegramChannel
}

func newRouterFixture(t *testing.T) *routerFixture {
	t.Helper()
	repo := telegrammemory.New()
	channel, err := repo.CreateChannel(t.Context(), "ws-a", &agentsv1.TelegramChannel{
		Id: "ch-1", Key: "main", BotId: "111111", BotUsername: "opsbot",
		InboundEnabled: true, OutboundEnabled: true,
		ReceiveMode: agentsv1.TelegramReceiveMode_TELEGRAM_RECEIVE_MODE_WEBHOOK,
	}, telegramrepo.Credential{Ciphertext: "cipher", KeyID: "k1"})
	if err != nil {
		t.Fatalf("create channel: %v", err)
	}
	if _, err := repo.CreateDestination(t.Context(), "ws-a", &agentsv1.TelegramDestination{
		Id: "dest-topic", Key: "incidents", ChannelId: "ch-1",
		ChatId: "-1001234567890", MessageThreadId: "42",
		InboundEnabled: true, OutboundEnabled: true,
		Config: &agentsv1.TelegramDestinationConfig{
			AgentId:           "support",
			TriggerMode:       agentsv1.TelegramTriggerMode_TELEGRAM_TRIGGER_MODE_MENTION,
			ControllerUserIds: []string{"7"},
		},
	}); err != nil {
		t.Fatalf("create topic destination: %v", err)
	}
	if _, err := repo.CreateDestination(t.Context(), "ws-a", &agentsv1.TelegramDestination{
		Id: "dest-general", Key: "general", ChannelId: "ch-1",
		ChatId:         "-1001234567890",
		InboundEnabled: true, OutboundEnabled: true,
		Config: &agentsv1.TelegramDestinationConfig{AgentId: "general-agent"},
	}); err != nil {
		t.Fatalf("create general destination: %v", err)
	}

	queue := newFakeQueue()
	router := NewRouter(repo, queue)
	return &routerFixture{router: router, repo: repo, queue: queue, channel: channel}
}

// textUpdate builds a Telegram update JSON for a message.
func textUpdate(updateID int64, chatID int64, threadID int64, text string, entities string) []byte {
	thread := ""
	if threadID > 0 {
		thread = fmt.Sprintf(`"message_thread_id":%d,"is_topic_message":true,`, threadID)
	}
	ent := ""
	if entities != "" {
		ent = fmt.Sprintf(`"entities":%s,`, entities)
	}
	return fmt.Appendf(nil, `{"update_id":%d,"message":{"message_id":9,%s%s"from":{"id":7,"is_bot":false,"username":"ops"},"chat":{"id":%d,"type":"supergroup"},"text":%q}}`,
		updateID, thread, ent, chatID, text)
}

const commandEntity = `[{"type":"bot_command","offset":0,"length":6}]`

func TestRouteMatchesTheExactTopic(t *testing.T) {
	fx := newRouterFixture(t)

	decision, err := fx.router.Route(t.Context(), fx.channel,
		textUpdate(1, -1001234567890, 42, "hello", ""))
	if err != nil {
		t.Fatalf("Route: %v", err)
	}
	if decision != DecisionAccepted {
		t.Fatalf("decision = %s", decision)
	}
	event := fx.queue.accepted[0]
	if event.DestinationID != "dest-topic" {
		t.Errorf("destination = %q, want the topic destination", event.DestinationID)
	}
	if event.Address.MessageThreadID != "42" {
		t.Errorf("thread = %q", event.Address.MessageThreadID)
	}
	if event.WorkspaceID != "ws-a" || event.ChannelID != "ch-1" || event.BotID != "111111" {
		t.Errorf("transport identity not frozen: %+v", event)
	}
}

// The same chat without a topic is a different address, not the same one.
func TestRouteDistinguishesTheGeneralConversation(t *testing.T) {
	fx := newRouterFixture(t)

	if _, err := fx.router.Route(t.Context(), fx.channel,
		textUpdate(1, -1001234567890, 0, "hello", "")); err != nil {
		t.Fatalf("Route: %v", err)
	}
	if got := fx.queue.accepted[0].DestinationID; got != "dest-general" {
		t.Errorf("destination = %q, want the non-topic destination", got)
	}
}

// The snapshot must carry the policy as it stood at acceptance, so a later
// edit cannot change the rules an in-flight message is judged by.
func TestRouteFreezesTheDestinationPolicy(t *testing.T) {
	fx := newRouterFixture(t)

	if _, err := fx.router.Route(t.Context(), fx.channel,
		textUpdate(1, -1001234567890, 42, "hello", "")); err != nil {
		t.Fatalf("Route: %v", err)
	}
	event := fx.queue.accepted[0]
	if event.Policy.AgentID != "support" {
		t.Errorf("agent = %q", event.Policy.AgentID)
	}
	if event.Policy.TriggerMode != agentsv1.TelegramTriggerMode_TELEGRAM_TRIGGER_MODE_MENTION.String() {
		t.Errorf("trigger = %q", event.Policy.TriggerMode)
	}
	if len(event.Policy.ControllerUserIDs) != 1 || event.Policy.ControllerUserIDs[0] != "7" {
		t.Errorf("controllers = %v", event.Policy.ControllerUserIDs)
	}
	if event.DestinationRevision == 0 {
		t.Error("expected the destination revision to travel with the event")
	}
}

func TestRouteIgnoresUnknownAddresses(t *testing.T) {
	fx := newRouterFixture(t)

	decision, err := fx.router.Route(t.Context(), fx.channel,
		textUpdate(1, -1009999999999, 0, "hello", ""))
	if err != nil {
		t.Fatalf("Route: %v", err)
	}
	if decision != DecisionIgnored {
		t.Fatalf("decision = %s, want ignored", decision)
	}
	if len(fx.queue.accepted) != 0 {
		t.Error("an unknown address must not reach the queue")
	}
}

func TestRouteIgnoresDisabledDestinations(t *testing.T) {
	fx := newRouterFixture(t)
	dest, err := fx.repo.GetDestination(t.Context(), "ws-a", "dest-topic")
	if err != nil {
		t.Fatalf("GetDestination: %v", err)
	}
	dest.InboundEnabled = false
	if _, err := fx.repo.UpdateDestination(t.Context(), "ws-a", dest, dest.GetRevision()); err != nil {
		t.Fatalf("disable destination: %v", err)
	}

	decision, err := fx.router.Route(t.Context(), fx.channel,
		textUpdate(1, -1001234567890, 42, "hello", ""))
	if err != nil {
		t.Fatalf("Route: %v", err)
	}
	if decision != DecisionIgnored || len(fx.queue.accepted) != 0 {
		t.Fatalf("decision = %s, accepted = %d", decision, len(fx.queue.accepted))
	}
}

// `/where` is recognized before Destination matching — that is the whole
// point: it has to work where nothing is configured yet.
func TestWhereIsAcceptedAtAnUnconfiguredAddress(t *testing.T) {
	fx := newRouterFixture(t)

	decision, err := fx.router.Route(t.Context(), fx.channel,
		textUpdate(1, -1009999999999, 0, "/where", commandEntity))
	if err != nil {
		t.Fatalf("Route: %v", err)
	}
	if decision != DecisionAccepted {
		t.Fatalf("decision = %s", decision)
	}
	event := fx.queue.accepted[0]
	if event.Kind != telegramqueue.KindWhere {
		t.Errorf("kind = %s", event.Kind)
	}
	if event.DestinationID != "" || event.Policy != nil {
		t.Error("/where must not carry destination routing")
	}
}

func TestWhereRespectsCommandAddressing(t *testing.T) {
	fx := newRouterFixture(t)

	// `/where@otherbot` addresses a different bot in the same group.
	raw := fmt.Appendf(nil, `{"update_id":1,"message":{"message_id":9,"from":{"id":7},"chat":{"id":-1009999999999,"type":"supergroup"},"text":"/where@otherbot","entities":[{"type":"bot_command","offset":0,"length":15}]}}`)
	decision, err := fx.router.Route(t.Context(), fx.channel, raw)
	if err != nil {
		t.Fatalf("Route: %v", err)
	}
	if decision != DecisionIgnored {
		t.Fatalf("decision = %s, want the other bot's command to be ignored", decision)
	}

	// `/where@opsbot` addresses us.
	raw = fmt.Appendf(nil, `{"update_id":2,"message":{"message_id":9,"from":{"id":7},"chat":{"id":-1009999999999,"type":"supergroup"},"text":"/where@opsbot","entities":[{"type":"bot_command","offset":0,"length":13}]}}`)
	if decision, err = fx.router.Route(t.Context(), fx.channel, raw); err != nil {
		t.Fatalf("Route: %v", err)
	}
	if decision != DecisionAccepted {
		t.Fatalf("decision = %s, want accepted", decision)
	}
}

// A message that merely contains "/where" is not a command: Telegram marks
// commands with an entity, and matching on text would misfire.
func TestWhereRequiresACommandEntity(t *testing.T) {
	fx := newRouterFixture(t)

	decision, err := fx.router.Route(t.Context(), fx.channel,
		textUpdate(1, -1009999999999, 0, "ask /where it went", ""))
	if err != nil {
		t.Fatalf("Route: %v", err)
	}
	if decision != DecisionIgnored {
		t.Fatalf("decision = %s, want ignored", decision)
	}
}

func TestRouteDeduplicatesByChannelAndUpdate(t *testing.T) {
	fx := newRouterFixture(t)
	raw := textUpdate(77, -1001234567890, 42, "hello", "")

	first, err := fx.router.Route(t.Context(), fx.channel, raw)
	if err != nil || first != DecisionAccepted {
		t.Fatalf("first: %s %v", first, err)
	}
	second, err := fx.router.Route(t.Context(), fx.channel, raw)
	if err != nil {
		t.Fatalf("second: %v", err)
	}
	if second != DecisionDuplicate {
		t.Fatalf("decision = %s, want duplicate", second)
	}
	if len(fx.queue.accepted) != 1 {
		t.Errorf("accepted %d events, want 1", len(fx.queue.accepted))
	}
}

// Malformed JSON is permanent: acknowledging it stops Telegram retrying
// something that can never work.
func TestMalformedUpdateIsPermanent(t *testing.T) {
	fx := newRouterFixture(t)

	_, err := fx.router.Route(t.Context(), fx.channel, []byte("{not json"))
	if !errors.Is(err, ErrMalformedUpdate) {
		t.Fatalf("err = %v, want ErrMalformedUpdate", err)
	}
}

// A queue failure is retryable: the caller must not acknowledge, or the
// update is lost.
func TestQueueFailureIsReportedAsNotAccepted(t *testing.T) {
	fx := newRouterFixture(t)
	fx.queue.failWith = errors.New("redis unavailable")

	_, err := fx.router.Route(t.Context(), fx.channel,
		textUpdate(1, -1001234567890, 42, "hello", ""))
	if err == nil {
		t.Fatal("expected the failure to surface")
	}
	if errors.Is(err, ErrMalformedUpdate) {
		t.Fatal("an infrastructure failure must not be reported as permanent")
	}
}

func TestIgnoredUpdateTypes(t *testing.T) {
	fx := newRouterFixture(t)

	for _, tc := range []struct {
		name string
		raw  string
	}{
		{"edited message", `{"update_id":1,"edited_message":{"message_id":9,"chat":{"id":-1001234567890},"text":"edited"}}`},
		{"channel post", `{"update_id":2,"channel_post":{"message_id":9,"chat":{"id":-1001234567890},"text":"post"}}`},
		{"no message", `{"update_id":3,"poll":{"id":"p"}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			decision, err := fx.router.Route(t.Context(), fx.channel, []byte(tc.raw))
			if err != nil {
				t.Fatalf("Route: %v", err)
			}
			if decision != DecisionIgnored {
				t.Fatalf("decision = %s, want ignored", decision)
			}
		})
	}
	if len(fx.queue.accepted) != 0 {
		t.Errorf("ignored types reached the queue: %d", len(fx.queue.accepted))
	}
}

// The raw update travels intact so a worker can read fields the receive path
// deliberately does not model.
func TestAcceptedEventCarriesTheRawUpdate(t *testing.T) {
	fx := newRouterFixture(t)
	raw := textUpdate(1, -1001234567890, 42, "hello", "")

	if _, err := fx.router.Route(t.Context(), fx.channel, raw); err != nil {
		t.Fatalf("Route: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(fx.queue.accepted[0].Update, &decoded); err != nil {
		t.Fatalf("decode carried update: %v", err)
	}
	if !strings.Contains(string(fx.queue.accepted[0].Update), `"text":"hello"`) {
		t.Error("expected the original update body to be preserved")
	}
}
