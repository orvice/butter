package application

// Service-level tests for Destination-based outbound delivery (issue
// #264/#266): notify-group and cron references, deletion guards, legacy-field
// rejection, and the Dashboard test-message action.

import (
	"context"
	"strings"
	"testing"

	"connectrpc.com/connect"

	configmemory "go.orx.me/apps/butter/internal/repo/config/memory"
	cryptokeymemory "go.orx.me/apps/butter/internal/repo/cryptokey/memory"
	"go.orx.me/apps/butter/internal/secretbox"
	"go.orx.me/apps/butter/internal/telegramsend"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// outboundFixture extends the #265 fixture with a working sender and the
// reference repositories the deletion guard consults.
type outboundFixture struct {
	*telegramFixture
	notify  *NotifyGroupServiceServer
	cronJob *fakeCronJobLister
	sender  *telegramsend.Sender
	channel *agentsv1.TelegramChannel
	dest    *agentsv1.TelegramDestination
}

type fakeCronJobLister struct {
	jobs []*agentsv1.CronJob
}

func (f *fakeCronJobLister) List(context.Context, string) ([]*agentsv1.CronJob, error) {
	return f.jobs, nil
}

func newOutboundFixture(t *testing.T) *outboundFixture {
	t.Helper()
	base := newTelegramFixture(t)

	// The channel service and the sender must share a keyring, or the sender
	// cannot decrypt what CreateTelegramChannel stored.
	keyring := secretbox.NewKeyring(cryptokeymemory.New())
	base.channels.SetKeyring(keyring)
	channel := base.createChannel(t, "main", mainBotToken)

	if _, err := base.channels.SetTelegramChannelEnabled(ctxAs("owner", "owner", "ws-a"),
		connect.NewRequest(&agentsv1.SetTelegramChannelEnabledRequest{
			ChannelId: channel.GetId(), Revision: channel.GetRevision(), OutboundEnabled: true,
		})); err != nil {
		t.Fatalf("enable channel outbound: %v", err)
	}

	dest := base.createDestination(t, &agentsv1.TelegramDestination{
		Key: "incidents", ChannelId: channel.GetId(),
		ChatId: "-1001234567890", MessageThreadId: "42", OutboundEnabled: true,
	})

	sender := telegramsend.New(base.repo, keyring, base.bots.Factory())
	notify := NewNotifyGroupServiceServer(configmemory.New())
	notify.SetTelegramRepo(base.repo)
	cronJobs := &fakeCronJobLister{}

	base.destinations.SetSender(sender)
	base.destinations.SetReferenceRepos(notifyGroupRepoOf(notify), cronJobs)

	return &outboundFixture{
		telegramFixture: base,
		notify:          notify,
		cronJob:         cronJobs,
		sender:          sender,
		channel:         channel,
		dest:            dest,
	}
}

// notifyGroupRepoOf exposes the service's repository so the deletion guard
// and the test observe the same store.
func notifyGroupRepoOf(svc *NotifyGroupServiceServer) *configmemory.Store {
	return svc.repo.(*configmemory.Store)
}

// --- Test message ----------------------------------------------------------

func TestSendTestMessageDeliversToTheTopicAndVerifies(t *testing.T) {
	fx := newOutboundFixture(t)

	resp, err := fx.destinations.SendTelegramTestMessage(ctxAs("owner", "owner", "ws-a"),
		connect.NewRequest(&agentsv1.SendTelegramTestMessageRequest{DestinationId: fx.dest.GetId()}))
	if err != nil {
		t.Fatalf("SendTelegramTestMessage: %v", err)
	}
	sent, ok := fx.bots.LastSent()
	if !ok {
		t.Fatal("nothing was sent")
	}
	if sent.Params.MessageThreadID != "42" {
		t.Errorf("message_thread_id = %q, want the destination topic", sent.Params.MessageThreadID)
	}
	if !resp.Msg.GetDestination().GetVerification().GetVerified() {
		t.Error("a successful test message must verify the destination")
	}
	if len(resp.Msg.GetMessageIds()) != 1 {
		t.Errorf("message_ids = %v, want exactly one", resp.Msg.GetMessageIds())
	}
}

// Creating a Destination is configuration, not delivery: it must not touch
// Telegram.
func TestCreatingADestinationSendsNothing(t *testing.T) {
	fx := newOutboundFixture(t)
	fx.bots.Reset()

	fx.createDestination(t, &agentsv1.TelegramDestination{
		Key: "second", ChannelId: fx.channel.GetId(), ChatId: "-1009999", OutboundEnabled: true,
	})
	if got := len(fx.bots.Sent()); got != 0 {
		t.Fatalf("creating a destination sent %d messages", got)
	}
}

func TestTestMessageRequiresManageRole(t *testing.T) {
	fx := newOutboundFixture(t)

	_, err := fx.destinations.SendTelegramTestMessage(ctxAs("member", "member", "ws-a"),
		connect.NewRequest(&agentsv1.SendTelegramTestMessageRequest{DestinationId: fx.dest.GetId()}))
	if code := connectCode(t, err); code != connect.CodePermissionDenied {
		t.Fatalf("code = %v, want PermissionDenied", code)
	}
}

func TestTestMessageToADisabledDestinationFailsExplicitly(t *testing.T) {
	fx := newOutboundFixture(t)
	dest := fx.dest
	dest.OutboundEnabled = false
	if _, err := fx.repo.UpdateDestination(t.Context(), "ws-a", dest, dest.GetRevision()); err != nil {
		t.Fatalf("disable destination: %v", err)
	}

	_, err := fx.destinations.SendTelegramTestMessage(ctxAs("owner", "owner", "ws-a"),
		connect.NewRequest(&agentsv1.SendTelegramTestMessageRequest{DestinationId: fx.dest.GetId()}))
	if code := connectCode(t, err); code != connect.CodeFailedPrecondition {
		t.Fatalf("code = %v, want FailedPrecondition", code)
	}
}

// --- Notify groups ---------------------------------------------------------

func createNotifyGroup(t *testing.T, fx *outboundFixture, group *agentsv1.NotifyGroup) error {
	t.Helper()
	_, err := fx.notify.CreateNotifyGroup(ctxAs("owner", "owner", "ws-a"),
		connect.NewRequest(&agentsv1.CreateNotifyGroupRequest{NotifyGroup: group}))
	return err
}

func telegramTarget(name, destinationID string) *agentsv1.NotifyTarget {
	return &agentsv1.NotifyTarget{
		Name:     name,
		Enabled:  true,
		Type:     agentsv1.NotifyTargetType_NOTIFY_TARGET_TYPE_TELEGRAM,
		Telegram: &agentsv1.TelegramNotifyTarget{DestinationId: destinationID},
	}
}

func TestNotifyGroupAcceptsADestinationReference(t *testing.T) {
	fx := newOutboundFixture(t)

	if err := createNotifyGroup(t, fx, &agentsv1.NotifyGroup{
		Name: "ops", Enabled: true,
		Targets: []*agentsv1.NotifyTarget{telegramTarget("incidents", fx.dest.GetId())},
	}); err != nil {
		t.Fatalf("CreateNotifyGroup: %v", err)
	}
}

func TestNotifyGroupRejectsLegacyTelegramFields(t *testing.T) {
	fx := newOutboundFixture(t)

	for _, tc := range []struct {
		name   string
		target *agentsv1.TelegramNotifyTarget
	}{
		{"bot_token", &agentsv1.TelegramNotifyTarget{BotToken: "123:abc", DestinationId: fx.dest.GetId()}},
		{"chat_id", &agentsv1.TelegramNotifyTarget{ChatId: "-100", DestinationId: fx.dest.GetId()}},
		{"parse_mode", &agentsv1.TelegramNotifyTarget{ParseMode: "MarkdownV2", DestinationId: fx.dest.GetId()}},
		{"message_thread_id", &agentsv1.TelegramNotifyTarget{MessageThreadId: 7, DestinationId: fx.dest.GetId()}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := createNotifyGroup(t, fx, &agentsv1.NotifyGroup{
				Name: "ops-" + tc.name, Enabled: true,
				Targets: []*agentsv1.NotifyTarget{{
					Name: "t", Enabled: true,
					Type:     agentsv1.NotifyTargetType_NOTIFY_TARGET_TYPE_TELEGRAM,
					Telegram: tc.target,
				}},
			})
			if code := connectCode(t, err); code != connect.CodeInvalidArgument {
				t.Fatalf("code = %v, want InvalidArgument", code)
			}
		})
	}
}

func TestNotifyGroupRejectsDuplicateDestinations(t *testing.T) {
	fx := newOutboundFixture(t)

	err := createNotifyGroup(t, fx, &agentsv1.NotifyGroup{
		Name: "ops", Enabled: true,
		Targets: []*agentsv1.NotifyTarget{
			telegramTarget("a", fx.dest.GetId()),
			telegramTarget("b", fx.dest.GetId()),
		},
	})
	if code := connectCode(t, err); code != connect.CodeInvalidArgument {
		t.Fatalf("code = %v, want InvalidArgument", code)
	}

	// The same Destination in two different groups is intentional.
	if err := createNotifyGroup(t, fx, &agentsv1.NotifyGroup{
		Name: "group-1", Enabled: true,
		Targets: []*agentsv1.NotifyTarget{telegramTarget("a", fx.dest.GetId())},
	}); err != nil {
		t.Fatalf("first group: %v", err)
	}
	if err := createNotifyGroup(t, fx, &agentsv1.NotifyGroup{
		Name: "group-2", Enabled: true,
		Targets: []*agentsv1.NotifyTarget{telegramTarget("a", fx.dest.GetId())},
	}); err != nil {
		t.Fatalf("second group: %v", err)
	}
}

func TestNotifyGroupRejectsUnknownOrUnusableDestinations(t *testing.T) {
	fx := newOutboundFixture(t)

	err := createNotifyGroup(t, fx, &agentsv1.NotifyGroup{
		Name: "unknown", Enabled: true,
		Targets: []*agentsv1.NotifyTarget{telegramTarget("a", "no-such-destination")},
	})
	if code := connectCode(t, err); code != connect.CodeInvalidArgument {
		t.Fatalf("unknown: code = %v, want InvalidArgument", code)
	}

	inboundOnly := fx.createDestination(t, &agentsv1.TelegramDestination{
		Key: "inbound-only", ChannelId: fx.channel.GetId(), ChatId: "-1005555",
		InboundEnabled: true,
		Config:         &agentsv1.TelegramDestinationConfig{AgentId: "support"},
	})
	err = createNotifyGroup(t, fx, &agentsv1.NotifyGroup{
		Name: "inbound", Enabled: true,
		Targets: []*agentsv1.NotifyTarget{telegramTarget("a", inboundOnly.GetId())},
	})
	if code := connectCode(t, err); code != connect.CodeInvalidArgument {
		t.Fatalf("inbound-only: code = %v, want InvalidArgument", code)
	}
}

// --- Cron ------------------------------------------------------------------

func TestCronDeliveryValidatesTheDestination(t *testing.T) {
	fx := newOutboundFixture(t)
	svc := &CronJobServiceServer{}
	svc.SetTelegramRepo(fx.repo)

	if err := svc.validateDelivery(t.Context(), "ws-a", &agentsv1.CronJob{
		Delivery: &agentsv1.CronDelivery{
			Type:                  agentsv1.CronDeliveryType_CRON_DELIVERY_TYPE_TELEGRAM_DESTINATION,
			TelegramDestinationId: fx.dest.GetId(),
		},
	}); err != nil {
		t.Fatalf("valid destination: %v", err)
	}

	err := svc.validateDelivery(t.Context(), "ws-a", &agentsv1.CronJob{
		Delivery: &agentsv1.CronDelivery{
			Type:                  agentsv1.CronDeliveryType_CRON_DELIVERY_TYPE_TELEGRAM_DESTINATION,
			TelegramDestinationId: "no-such-destination",
		},
	})
	if code := connectCode(t, err); code != connect.CodeInvalidArgument {
		t.Fatalf("unknown destination: code = %v, want InvalidArgument", code)
	}

	// Cross-workspace references must not resolve.
	err = svc.validateDelivery(t.Context(), "ws-b", &agentsv1.CronJob{
		Delivery: &agentsv1.CronDelivery{
			Type:                  agentsv1.CronDeliveryType_CRON_DELIVERY_TYPE_TELEGRAM_DESTINATION,
			TelegramDestinationId: fx.dest.GetId(),
		},
	})
	if code := connectCode(t, err); code != connect.CodeInvalidArgument {
		t.Fatalf("cross-workspace: code = %v, want InvalidArgument", code)
	}
}

// The legacy channel-name + chat-id form is no longer a working Telegram path
// for new or edited jobs.
func TestCronRejectsLegacyChannelDelivery(t *testing.T) {
	svc := &CronJobServiceServer{}

	err := svc.validateDelivery(t.Context(), "ws-a", &agentsv1.CronJob{
		Delivery: &agentsv1.CronDelivery{
			Type:        agentsv1.CronDeliveryType_CRON_DELIVERY_TYPE_CHANNEL,
			ChannelName: "legacy", ChatId: "-100",
		},
	})
	if code := connectCode(t, err); code != connect.CodeInvalidArgument {
		t.Fatalf("code = %v, want InvalidArgument", code)
	}
}

// Non-Telegram delivery types are untouched.
func TestCronStillAcceptsOtherDeliveryTypes(t *testing.T) {
	svc := &CronJobServiceServer{}

	for _, deliveryType := range []agentsv1.CronDeliveryType{
		agentsv1.CronDeliveryType_CRON_DELIVERY_TYPE_LOG,
		agentsv1.CronDeliveryType_CRON_DELIVERY_TYPE_WEBHOOK,
		agentsv1.CronDeliveryType_CRON_DELIVERY_TYPE_NOTIFY_GROUP,
	} {
		if err := svc.validateDelivery(t.Context(), "ws-a", &agentsv1.CronJob{
			Delivery: &agentsv1.CronDelivery{Type: deliveryType},
		}); err != nil {
			t.Errorf("%v: %v", deliveryType, err)
		}
	}
}

// --- Deletion guard --------------------------------------------------------

func TestDeleteDestinationIsBlockedByReferences(t *testing.T) {
	fx := newOutboundFixture(t)
	if err := createNotifyGroup(t, fx, &agentsv1.NotifyGroup{
		Name: "ops", Enabled: true,
		Targets: []*agentsv1.NotifyTarget{telegramTarget("incidents", fx.dest.GetId())},
	}); err != nil {
		t.Fatalf("CreateNotifyGroup: %v", err)
	}
	fx.cronJob.jobs = []*agentsv1.CronJob{{
		Name: "nightly",
		Delivery: &agentsv1.CronDelivery{
			Type:                  agentsv1.CronDeliveryType_CRON_DELIVERY_TYPE_TELEGRAM_DESTINATION,
			TelegramDestinationId: fx.dest.GetId(),
		},
	}}

	_, err := fx.destinations.DeleteTelegramDestination(ctxAs("owner", "owner", "ws-a"),
		connect.NewRequest(&agentsv1.DeleteTelegramDestinationRequest{Id: fx.dest.GetId()}))
	if code := connectCode(t, err); code != connect.CodeFailedPrecondition {
		t.Fatalf("code = %v, want FailedPrecondition", code)
	}
	// The operator needs to know what to repair.
	if !strings.Contains(err.Error(), "notify group ops") {
		t.Errorf("error should name the notify group, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "cron job nightly") {
		t.Errorf("error should name the cron job, got %q", err.Error())
	}
}

func TestDeleteDestinationSucceedsOnceReferencesAreRemoved(t *testing.T) {
	fx := newOutboundFixture(t)
	fx.cronJob.jobs = []*agentsv1.CronJob{{
		Name: "nightly",
		Delivery: &agentsv1.CronDelivery{
			Type:                  agentsv1.CronDeliveryType_CRON_DELIVERY_TYPE_TELEGRAM_DESTINATION,
			TelegramDestinationId: fx.dest.GetId(),
		},
	}}
	if _, err := fx.destinations.DeleteTelegramDestination(ctxAs("owner", "owner", "ws-a"),
		connect.NewRequest(&agentsv1.DeleteTelegramDestinationRequest{Id: fx.dest.GetId()})); err == nil {
		t.Fatal("expected the referenced destination to be protected")
	}

	fx.cronJob.jobs = nil
	if _, err := fx.destinations.DeleteTelegramDestination(ctxAs("owner", "owner", "ws-a"),
		connect.NewRequest(&agentsv1.DeleteTelegramDestinationRequest{Id: fx.dest.GetId()})); err != nil {
		t.Fatalf("DeleteTelegramDestination: %v", err)
	}
	if _, err := fx.repo.GetDestination(t.Context(), "ws-a", fx.dest.GetId()); err == nil {
		t.Error("expected the destination to be gone")
	}
}
