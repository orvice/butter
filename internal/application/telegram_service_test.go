package application

// Service-level tests for TelegramChannelService and
// TelegramDestinationService (issue #264/#265): role enforcement, credential
// write-only handling, immutable Bot identity, optimistic revisions, address
// uniqueness, and strong Agent/Model references.

import (
	"context"
	"strings"
	"testing"

	"connectrpc.com/connect"

	configmemory "go.orx.me/apps/butter/internal/repo/config/memory"
	cryptokeymemory "go.orx.me/apps/butter/internal/repo/cryptokey/memory"
	telegramrepo "go.orx.me/apps/butter/internal/repo/telegram"
	telegrammemory "go.orx.me/apps/butter/internal/repo/telegram/memory"
	telegramsettingmemory "go.orx.me/apps/butter/internal/repo/telegramsetting/memory"
	workspacememory "go.orx.me/apps/butter/internal/repo/workspace/memory"
	"go.orx.me/apps/butter/internal/secretbox"
	"go.orx.me/apps/butter/internal/telegramapi"
	"go.orx.me/apps/butter/internal/telegramapi/telegramtest"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const (
	mainBotToken    = "111111:main-bot-token"
	rotatedBotToken = "111111:rotated-bot-token"
	otherBotToken   = "222222:other-bot-token"
)

type telegramFixture struct {
	channels     *TelegramChannelServiceServer
	destinations *TelegramDestinationServiceServer
	repo         telegramrepo.Repository
	config       *configmemory.Store
	bots         *telegramtest.Fake
	settings     *telegramsettingmemory.Store
	queue        *stubQueueProbe
}

// stubQueueProbe stands in for the Redis Streams queue in service tests.
type stubQueueProbe struct {
	available bool
	pingErr   error
}

func (q *stubQueueProbe) Available() bool            { return q.available }
func (q *stubQueueProbe) Ping(context.Context) error { return q.pingErr }

func newTelegramFixture(t *testing.T) *telegramFixture {
	t.Helper()

	wsRepo := workspacememory.New()
	for _, ws := range []string{"ws-a", "ws-b"} {
		if _, err := wsRepo.CreateWorkspace(t.Context(), &agentsv1.Workspace{Id: ws, Name: ws, Slug: ws}); err != nil {
			t.Fatalf("seed workspace %s: %v", ws, err)
		}
	}
	for _, m := range []struct{ user, role, ws string }{
		{"owner", "owner", "ws-a"},
		{"member", "member", "ws-a"},
		{"owner-b", "owner", "ws-b"},
	} {
		if _, err := wsRepo.AddMember(t.Context(), &agentsv1.WorkspaceMember{
			WorkspaceId: m.ws, UserId: m.user, Role: m.role,
		}); err != nil {
			t.Fatalf("seed member %s: %v", m.user, err)
		}
	}

	bots := telegramtest.NewFake()
	bots.Register(mainBotToken, telegramtest.Identity(111111, "mainbot"))
	bots.Register(rotatedBotToken, telegramtest.Identity(111111, "mainbot_renamed"))
	bots.Register(otherBotToken, telegramtest.Identity(222222, "otherbot"))

	repo := telegrammemory.New()
	configStore := configmemory.New()

	channels := NewTelegramChannelServiceServer(repo)
	channels.SetWorkspaceRepo(wsRepo)
	channels.SetKeyring(secretbox.NewKeyring(cryptokeymemory.New()))
	channels.SetBotFactory(bots.Factory())
	// Webhook mode has real infrastructure prerequisites (#267): a public
	// base URL and a durable queue. The fixture satisfies both so tests about
	// other rules are not dominated by them; the tests that care about the
	// prerequisites remove them explicitly.
	settings := telegramsettingmemory.New()
	if _, err := settings.Put(t.Context(), &agentsv1.TelegramSettings{
		WebhookBaseUrl: "https://butter.test",
	}); err != nil {
		t.Fatalf("seed telegram settings: %v", err)
	}
	channels.SetSettingsRepo(settings)
	queue := &stubQueueProbe{available: true}
	channels.SetQueueProbe(queue)

	destinations := NewTelegramDestinationServiceServer(repo)
	destinations.SetWorkspaceRepo(wsRepo)
	destinations.SetConfigRepos(configStore, configStore)

	fx := &telegramFixture{
		channels:     channels,
		destinations: destinations,
		repo:         repo,
		config:       configStore,
		bots:         bots,
		settings:     settings,
		queue:        queue,
	}
	fx.seedAgent(t, "ws-a", "support", agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_ACTIVE)
	fx.seedAgent(t, "ws-a", "research", agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_ACTIVE)
	fx.seedAgent(t, "ws-a", "retired", agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_DELETED)
	if _, err := configStore.CreateModelProvider(t.Context(), "ws-a", &agentsv1.ModelProvider{
		Name: "openai",
		Type: "openai",
		Models: []*agentsv1.ModelConfig{
			{Name: "gpt-5", Alias: "fast"},
			{Name: "gpt-5-pro", Alias: "pro"},
		},
	}); err != nil {
		t.Fatalf("seed model provider: %v", err)
	}
	return fx
}

func (fx *telegramFixture) seedAgent(t *testing.T, workspaceID, agentID string, status agentsv1.AgentLifecycleStatus) {
	t.Helper()
	if _, err := fx.config.CreateAgent(t.Context(), workspaceID, &agentsv1.Agent{
		Name:            agentID,
		AgentId:         agentID,
		LifecycleStatus: status,
	}); err != nil {
		t.Fatalf("seed agent %s: %v", agentID, err)
	}
}

// createChannel registers a working Channel as the ws-a owner.
func (fx *telegramFixture) createChannel(t *testing.T, key, token string) *agentsv1.TelegramChannel {
	t.Helper()
	resp, err := fx.channels.CreateTelegramChannel(ctxAs("owner", "owner", "ws-a"),
		connect.NewRequest(&agentsv1.CreateTelegramChannelRequest{
			Channel:  &agentsv1.TelegramChannel{Key: key},
			BotToken: token,
		}))
	if err != nil {
		t.Fatalf("CreateTelegramChannel(%s): %v", key, err)
	}
	return resp.Msg.GetChannel()
}

func (fx *telegramFixture) createDestination(t *testing.T, dest *agentsv1.TelegramDestination) *agentsv1.TelegramDestination {
	t.Helper()
	resp, err := fx.destinations.CreateTelegramDestination(ctxAs("owner", "owner", "ws-a"),
		connect.NewRequest(&agentsv1.CreateTelegramDestinationRequest{Destination: dest}))
	if err != nil {
		t.Fatalf("CreateTelegramDestination(%s): %v", dest.GetKey(), err)
	}
	return resp.Msg.GetDestination()
}

func connectCode(t *testing.T, err error) connect.Code {
	t.Helper()
	if err == nil {
		t.Fatal("expected an error")
	}
	return connect.CodeOf(err)
}

// --- Channel creation ------------------------------------------------------

func TestCreateChannelPinsTheValidatedBotIdentity(t *testing.T) {
	fx := newTelegramFixture(t)

	channel := fx.createChannel(t, "main", mainBotToken)

	if channel.GetBotId() != "111111" {
		t.Errorf("bot_id = %q, want the id resolved by getMe", channel.GetBotId())
	}
	if channel.GetBotUsername() != "mainbot" {
		t.Errorf("bot_username = %q", channel.GetBotUsername())
	}
	if channel.GetReceiveMode() != agentsv1.TelegramReceiveMode_TELEGRAM_RECEIVE_MODE_WEBHOOK {
		t.Errorf("receive_mode = %v, want WEBHOOK by default", channel.GetReceiveMode())
	}
	if channel.GetInboundEnabled() || channel.GetOutboundEnabled() {
		t.Error("channels must be created disabled")
	}
	if channel.GetRevision() != 1 {
		t.Errorf("revision = %d, want 1", channel.GetRevision())
	}
	if fx.bots.GetMeCalls() != 1 {
		t.Errorf("getMe calls = %d, want exactly one validation", fx.bots.GetMeCalls())
	}
}

func TestCreateChannelRejectsAnInvalidTokenBeforeCommitting(t *testing.T) {
	fx := newTelegramFixture(t)

	_, err := fx.channels.CreateTelegramChannel(ctxAs("owner", "owner", "ws-a"),
		connect.NewRequest(&agentsv1.CreateTelegramChannelRequest{
			Channel:  &agentsv1.TelegramChannel{Key: "main"},
			BotToken: "999999:not-a-real-token",
		}))
	if code := connectCode(t, err); code != connect.CodeInvalidArgument {
		t.Fatalf("code = %v, want InvalidArgument", code)
	}

	// Nothing may be left behind: no Channel, and therefore no credential.
	channels, listErr := fx.repo.ListChannels(t.Context(), "ws-a")
	if listErr != nil {
		t.Fatalf("ListChannels: %v", listErr)
	}
	if len(channels) != 0 {
		t.Fatalf("expected no channel to be committed, got %d", len(channels))
	}
}

// One Telegram Bot must never back two Channels, in any workspace: both
// would consume its updates.
func TestCreateChannelRejectsABotAlreadyRegisteredInAnotherWorkspace(t *testing.T) {
	fx := newTelegramFixture(t)
	fx.createChannel(t, "main", mainBotToken)

	_, err := fx.channels.CreateTelegramChannel(ctxAs("owner-b", "owner", "ws-b"),
		connect.NewRequest(&agentsv1.CreateTelegramChannelRequest{
			Channel:  &agentsv1.TelegramChannel{Key: "main"},
			BotToken: mainBotToken,
		}))
	if code := connectCode(t, err); code != connect.CodeAlreadyExists {
		t.Fatalf("code = %v, want AlreadyExists", code)
	}
}

func TestChannelReadsNeverReturnTheBotToken(t *testing.T) {
	fx := newTelegramFixture(t)
	created := fx.createChannel(t, "main", mainBotToken)

	resp, err := fx.channels.GetTelegramChannel(ctxAs("member", "member", "ws-a"),
		connect.NewRequest(&agentsv1.GetTelegramChannelRequest{Id: created.GetId()}))
	if err != nil {
		t.Fatalf("GetTelegramChannel: %v", err)
	}
	// The proto has no token field at all; assert on the rendered message so
	// the test still fails if one is ever added.
	if rendered := resp.Msg.GetChannel().String(); strings.Contains(rendered, "main-bot-token") {
		t.Fatalf("channel read leaked the bot token: %s", rendered)
	}
	if resp.Msg.GetChannel().GetCredentialState() != agentsv1.TelegramCredentialState_TELEGRAM_CREDENTIAL_STATE_VALID {
		t.Errorf("credential_state = %v, want VALID", resp.Msg.GetChannel().GetCredentialState())
	}
}

// --- Roles -----------------------------------------------------------------

func TestChannelReadsAreOpenToMembersButMutationsAreNot(t *testing.T) {
	fx := newTelegramFixture(t)
	created := fx.createChannel(t, "main", mainBotToken)

	if _, err := fx.channels.ListTelegramChannels(ctxAs("member", "member", "ws-a"),
		connect.NewRequest(&agentsv1.ListTelegramChannelsRequest{})); err != nil {
		t.Fatalf("member list: %v", err)
	}

	_, err := fx.channels.UpdateTelegramChannel(ctxAs("member", "member", "ws-a"),
		connect.NewRequest(&agentsv1.UpdateTelegramChannelRequest{
			Channel: &agentsv1.TelegramChannel{Id: created.GetId(), Name: "Hijacked", Revision: 1},
		}))
	if code := connectCode(t, err); code != connect.CodePermissionDenied {
		t.Fatalf("code = %v, want PermissionDenied", code)
	}
}

// --- Rotation --------------------------------------------------------------

func TestRotateCredentialAcceptsOnlyTheSameBot(t *testing.T) {
	fx := newTelegramFixture(t)
	created := fx.createChannel(t, "main", mainBotToken)

	resp, err := fx.channels.RotateTelegramChannelCredential(ctxAs("owner", "owner", "ws-a"),
		connect.NewRequest(&agentsv1.RotateTelegramChannelCredentialRequest{
			ChannelId: created.GetId(),
			BotToken:  rotatedBotToken,
		}))
	if err != nil {
		t.Fatalf("RotateTelegramChannelCredential: %v", err)
	}
	if resp.Msg.GetChannel().GetBotId() != "111111" {
		t.Errorf("bot_id changed to %q", resp.Msg.GetChannel().GetBotId())
	}
	if resp.Msg.GetChannel().GetBotUsername() != "mainbot_renamed" {
		t.Errorf("bot_username = %q, want the refreshed username", resp.Msg.GetChannel().GetBotUsername())
	}
}

func TestRotateCredentialRejectsADifferentBotAndKeepsThePriorCredential(t *testing.T) {
	fx := newTelegramFixture(t)
	created := fx.createChannel(t, "main", mainBotToken)
	before, err := fx.repo.GetChannelCredential(t.Context(), "ws-a", created.GetId())
	if err != nil {
		t.Fatalf("GetChannelCredential: %v", err)
	}

	_, err = fx.channels.RotateTelegramChannelCredential(ctxAs("owner", "owner", "ws-a"),
		connect.NewRequest(&agentsv1.RotateTelegramChannelCredentialRequest{
			ChannelId: created.GetId(),
			BotToken:  otherBotToken,
		}))
	if code := connectCode(t, err); code != connect.CodeInvalidArgument {
		t.Fatalf("code = %v, want InvalidArgument", code)
	}

	after, err := fx.repo.GetChannelCredential(t.Context(), "ws-a", created.GetId())
	if err != nil {
		t.Fatalf("GetChannelCredential after rejected rotation: %v", err)
	}
	if after.Ciphertext != before.Ciphertext {
		t.Error("a rejected rotation must leave the prior credential in place")
	}
}

// --- Revisions and immutability -------------------------------------------

func TestUpdateChannelRejectsAStaleRevision(t *testing.T) {
	fx := newTelegramFixture(t)
	created := fx.createChannel(t, "main", mainBotToken)

	first, err := fx.channels.UpdateTelegramChannel(ctxAs("owner", "owner", "ws-a"),
		connect.NewRequest(&agentsv1.UpdateTelegramChannelRequest{
			Channel: &agentsv1.TelegramChannel{Id: created.GetId(), Name: "First", Revision: 1},
		}))
	if err != nil {
		t.Fatalf("first update: %v", err)
	}

	_, err = fx.channels.UpdateTelegramChannel(ctxAs("owner", "owner", "ws-a"),
		connect.NewRequest(&agentsv1.UpdateTelegramChannelRequest{
			Channel: &agentsv1.TelegramChannel{Id: created.GetId(), Name: "Second", Revision: 1},
		}))
	if code := connectCode(t, err); code != connect.CodeAborted {
		t.Fatalf("code = %v, want Aborted", code)
	}

	current, err := fx.repo.GetChannel(t.Context(), "ws-a", created.GetId())
	if err != nil {
		t.Fatalf("GetChannel: %v", err)
	}
	if current.GetName() != "First" {
		t.Errorf("name = %q; the stale write overwrote newer state", current.GetName())
	}
	if current.GetRevision() != first.Msg.GetChannel().GetRevision() {
		t.Errorf("revision moved to %d", current.GetRevision())
	}
}

func TestUpdateChannelRejectsImmutableFieldChanges(t *testing.T) {
	fx := newTelegramFixture(t)
	created := fx.createChannel(t, "main", mainBotToken)

	for _, tc := range []struct {
		name    string
		channel *agentsv1.TelegramChannel
	}{
		{"key", &agentsv1.TelegramChannel{Id: created.GetId(), Key: "renamed", Revision: 1}},
		{"bot_id", &agentsv1.TelegramChannel{Id: created.GetId(), BotId: "999", Revision: 1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := fx.channels.UpdateTelegramChannel(ctxAs("owner", "owner", "ws-a"),
				connect.NewRequest(&agentsv1.UpdateTelegramChannelRequest{Channel: tc.channel}))
			if code := connectCode(t, err); code != connect.CodeInvalidArgument {
				t.Fatalf("code = %v, want InvalidArgument", code)
			}
		})
	}
}

// --- Enablement ------------------------------------------------------------

func TestEnableRejectsInboundWithoutOutbound(t *testing.T) {
	fx := newTelegramFixture(t)
	created := fx.createChannel(t, "main", mainBotToken)

	_, err := fx.channels.SetTelegramChannelEnabled(ctxAs("owner", "owner", "ws-a"),
		connect.NewRequest(&agentsv1.SetTelegramChannelEnabledRequest{
			ChannelId: created.GetId(), Revision: 1, InboundEnabled: true, OutboundEnabled: false,
		}))
	if code := connectCode(t, err); code != connect.CodeInvalidArgument {
		t.Fatalf("code = %v, want InvalidArgument", code)
	}
}

func TestEnableInboundRequiresAnEnabledInboundDestination(t *testing.T) {
	fx := newTelegramFixture(t)
	created := fx.createChannel(t, "main", mainBotToken)

	_, err := fx.channels.SetTelegramChannelEnabled(ctxAs("owner", "owner", "ws-a"),
		connect.NewRequest(&agentsv1.SetTelegramChannelEnabledRequest{
			ChannelId: created.GetId(), Revision: 1, InboundEnabled: true, OutboundEnabled: true,
		}))
	if code := connectCode(t, err); code != connect.CodeFailedPrecondition {
		t.Fatalf("code = %v, want FailedPrecondition", code)
	}

	fx.createDestination(t, &agentsv1.TelegramDestination{
		Key:             "ops",
		ChannelId:       created.GetId(),
		ChatId:          "-1001234567890",
		InboundEnabled:  true,
		OutboundEnabled: true,
		Config:          &agentsv1.TelegramDestinationConfig{AgentId: "support"},
	})

	if _, err := fx.channels.SetTelegramChannelEnabled(ctxAs("owner", "owner", "ws-a"),
		connect.NewRequest(&agentsv1.SetTelegramChannelEnabledRequest{
			ChannelId: created.GetId(), Revision: 1, InboundEnabled: true, OutboundEnabled: true,
		})); err != nil {
		t.Fatalf("enable after adding an inbound destination: %v", err)
	}
}

// Outbound-only Channels must be enablable without any inbound Destination:
// a notification Bot never receives anything.
func TestEnableOutboundOnlyNeedsNoInboundDestination(t *testing.T) {
	fx := newTelegramFixture(t)
	created := fx.createChannel(t, "main", mainBotToken)

	if _, err := fx.channels.SetTelegramChannelEnabled(ctxAs("owner", "owner", "ws-a"),
		connect.NewRequest(&agentsv1.SetTelegramChannelEnabledRequest{
			ChannelId: created.GetId(), Revision: 1, InboundEnabled: false, OutboundEnabled: true,
		})); err != nil {
		t.Fatalf("enable outbound only: %v", err)
	}
}

func TestGroupPrivacyIsReportedAsAWarningNotABlocker(t *testing.T) {
	fx := newTelegramFixture(t)
	// A Bot whose Group Privacy is on: Butter can diagnose this but cannot
	// change it, so it must not block enablement.
	fx.bots.Register("333333:private-bot", telegramapiIdentityWithPrivacy())
	created := fx.createChannel(t, "private", "333333:private-bot")
	fx.createDestination(t, &agentsv1.TelegramDestination{
		Key:             "private-ops",
		ChannelId:       created.GetId(),
		ChatId:          "-1009876543210",
		InboundEnabled:  true,
		OutboundEnabled: true,
		Config:          &agentsv1.TelegramDestinationConfig{AgentId: "support"},
	})

	resp, err := fx.channels.SetTelegramChannelEnabled(ctxAs("owner", "owner", "ws-a"),
		connect.NewRequest(&agentsv1.SetTelegramChannelEnabledRequest{
			ChannelId: created.GetId(), Revision: 1, InboundEnabled: true, OutboundEnabled: true,
		}))
	if err != nil {
		t.Fatalf("SetTelegramChannelEnabled: %v", err)
	}
	if len(resp.Msg.GetWarnings()) == 0 {
		t.Fatal("expected a group privacy warning")
	}
	if !strings.Contains(strings.ToLower(resp.Msg.GetWarnings()[0]), "group privacy") {
		t.Errorf("warning = %q", resp.Msg.GetWarnings()[0])
	}
}

// --- Channel deletion ------------------------------------------------------

func TestDeleteChannelIsBlockedWhileADestinationReferencesIt(t *testing.T) {
	fx := newTelegramFixture(t)
	created := fx.createChannel(t, "main", mainBotToken)
	dest := fx.createDestination(t, &agentsv1.TelegramDestination{
		Key: "ops", ChannelId: created.GetId(), ChatId: "-1001234567890", OutboundEnabled: true,
	})

	_, err := fx.channels.DeleteTelegramChannel(ctxAs("owner", "owner", "ws-a"),
		connect.NewRequest(&agentsv1.DeleteTelegramChannelRequest{Id: created.GetId()}))
	if code := connectCode(t, err); code != connect.CodeFailedPrecondition {
		t.Fatalf("code = %v, want FailedPrecondition", code)
	}
	// The operator needs to know which Destinations to repair.
	if !strings.Contains(err.Error(), dest.GetId()) {
		t.Errorf("error should name the referencing destination, got %q", err.Error())
	}
}

// --- Destinations ----------------------------------------------------------

func TestCreateDestinationCanonicalizesTelegramIdentifiers(t *testing.T) {
	fx := newTelegramFixture(t)
	channel := fx.createChannel(t, "main", mainBotToken)

	dest := fx.createDestination(t, &agentsv1.TelegramDestination{
		Key:             "ops",
		ChannelId:       channel.GetId(),
		ChatId:          " -1001234567890 ",
		MessageThreadId: "007",
		OutboundEnabled: true,
	})
	if dest.GetChatId() != "-1001234567890" {
		t.Errorf("chat_id = %q, want the canonical form", dest.GetChatId())
	}
	if dest.GetMessageThreadId() != "7" {
		t.Errorf("message_thread_id = %q, want the canonical form", dest.GetMessageThreadId())
	}
}

func TestCreateDestinationRejectsANonPositiveThreadID(t *testing.T) {
	fx := newTelegramFixture(t)
	channel := fx.createChannel(t, "main", mainBotToken)

	_, err := fx.destinations.CreateTelegramDestination(ctxAs("owner", "owner", "ws-a"),
		connect.NewRequest(&agentsv1.CreateTelegramDestinationRequest{
			Destination: &agentsv1.TelegramDestination{
				Key: "ops", ChannelId: channel.GetId(), ChatId: "-100", MessageThreadId: "0",
			},
		}))
	if code := connectCode(t, err); code != connect.CodeInvalidArgument {
		t.Fatalf("code = %v, want InvalidArgument", code)
	}
}

func TestCreateDestinationRejectsADuplicateAddress(t *testing.T) {
	fx := newTelegramFixture(t)
	channel := fx.createChannel(t, "main", mainBotToken)
	fx.createDestination(t, &agentsv1.TelegramDestination{
		Key: "ops", ChannelId: channel.GetId(), ChatId: "-100", MessageThreadId: "7",
	})

	_, err := fx.destinations.CreateTelegramDestination(ctxAs("owner", "owner", "ws-a"),
		connect.NewRequest(&agentsv1.CreateTelegramDestinationRequest{
			Destination: &agentsv1.TelegramDestination{
				Key: "ops-copy", ChannelId: channel.GetId(), ChatId: "-100", MessageThreadId: "7",
			},
		}))
	if code := connectCode(t, err); code != connect.CodeAlreadyExists {
		t.Fatalf("code = %v, want AlreadyExists", code)
	}
}

func TestCreateDestinationRejectsAChannelFromAnotherWorkspace(t *testing.T) {
	fx := newTelegramFixture(t)
	channel := fx.createChannel(t, "main", mainBotToken)

	_, err := fx.destinations.CreateTelegramDestination(ctxAs("owner-b", "owner", "ws-b"),
		connect.NewRequest(&agentsv1.CreateTelegramDestinationRequest{
			Destination: &agentsv1.TelegramDestination{
				Key: "ops", ChannelId: channel.GetId(), ChatId: "-100",
			},
		}))
	if code := connectCode(t, err); code != connect.CodeInvalidArgument {
		t.Fatalf("code = %v, want InvalidArgument", code)
	}
}

func TestInboundDestinationRequiresAnActiveAgent(t *testing.T) {
	fx := newTelegramFixture(t)
	channel := fx.createChannel(t, "main", mainBotToken)

	for _, tc := range []struct {
		name    string
		agentID string
	}{
		{"missing", ""},
		{"unknown", "no-such-agent"},
		{"tombstoned", "retired"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := fx.destinations.CreateTelegramDestination(ctxAs("owner", "owner", "ws-a"),
				connect.NewRequest(&agentsv1.CreateTelegramDestinationRequest{
					Destination: &agentsv1.TelegramDestination{
						Key:            "ops-" + tc.name,
						ChannelId:      channel.GetId(),
						ChatId:         "-100",
						InboundEnabled: true,
						Config:         &agentsv1.TelegramDestinationConfig{AgentId: tc.agentID},
					},
				}))
			if code := connectCode(t, err); code != connect.CodeInvalidArgument {
				t.Fatalf("code = %v, want InvalidArgument", code)
			}
		})
	}
}

func TestDestinationConfigDefaultsAreApplied(t *testing.T) {
	fx := newTelegramFixture(t)
	channel := fx.createChannel(t, "main", mainBotToken)

	dest := fx.createDestination(t, &agentsv1.TelegramDestination{
		Key: "ops", ChannelId: channel.GetId(), ChatId: "-100",
		InboundEnabled: true,
		Config:         &agentsv1.TelegramDestinationConfig{AgentId: "support"},
	})
	cfg := dest.GetConfig()
	if cfg.GetTriggerMode() != agentsv1.TelegramTriggerMode_TELEGRAM_TRIGGER_MODE_ALL {
		t.Errorf("trigger_mode = %v, want ALL", cfg.GetTriggerMode())
	}
	if cfg.GetSessionPolicy() != agentsv1.TelegramSessionPolicy_TELEGRAM_SESSION_POLICY_DESTINATION {
		t.Errorf("session_policy = %v, want DESTINATION", cfg.GetSessionPolicy())
	}
	if cfg.GetReplyMode() != agentsv1.TelegramReplyMode_TELEGRAM_REPLY_MODE_REPLY {
		t.Errorf("reply_mode = %v, want REPLY", cfg.GetReplyMode())
	}
}

// A controller who cannot be admitted could never reach the Destination.
func TestControllersMustAlsoBeAdmitted(t *testing.T) {
	fx := newTelegramFixture(t)
	channel := fx.createChannel(t, "main", mainBotToken)

	_, err := fx.destinations.CreateTelegramDestination(ctxAs("owner", "owner", "ws-a"),
		connect.NewRequest(&agentsv1.CreateTelegramDestinationRequest{
			Destination: &agentsv1.TelegramDestination{
				Key: "ops", ChannelId: channel.GetId(), ChatId: "-100", InboundEnabled: true,
				Config: &agentsv1.TelegramDestinationConfig{
					AgentId:           "support",
					AllowedUserIds:    []string{"10", "20"},
					ControllerUserIds: []string{"30"},
				},
			},
		}))
	if code := connectCode(t, err); code != connect.CodeInvalidArgument {
		t.Fatalf("code = %v, want InvalidArgument", code)
	}

	// With an empty allow-list every real user is admitted, so any
	// controller is consistent.
	fx.createDestination(t, &agentsv1.TelegramDestination{
		Key: "open", ChannelId: channel.GetId(), ChatId: "-200", InboundEnabled: true,
		Config: &agentsv1.TelegramDestinationConfig{
			AgentId:           "support",
			ControllerUserIds: []string{"30"},
		},
	})
}

func TestSelectableListsMustContainTheDefaults(t *testing.T) {
	fx := newTelegramFixture(t)
	channel := fx.createChannel(t, "main", mainBotToken)

	_, err := fx.destinations.CreateTelegramDestination(ctxAs("owner", "owner", "ws-a"),
		connect.NewRequest(&agentsv1.CreateTelegramDestinationRequest{
			Destination: &agentsv1.TelegramDestination{
				Key: "ops", ChannelId: channel.GetId(), ChatId: "-100", InboundEnabled: true,
				Config: &agentsv1.TelegramDestinationConfig{
					AgentId:            "support",
					SelectableAgentIds: []string{"research"},
				},
			},
		}))
	if code := connectCode(t, err); code != connect.CodeInvalidArgument {
		t.Fatalf("agent list: code = %v, want InvalidArgument", code)
	}

	_, err = fx.destinations.CreateTelegramDestination(ctxAs("owner", "owner", "ws-a"),
		connect.NewRequest(&agentsv1.CreateTelegramDestinationRequest{
			Destination: &agentsv1.TelegramDestination{
				Key: "ops2", ChannelId: channel.GetId(), ChatId: "-200", InboundEnabled: true,
				Config: &agentsv1.TelegramDestinationConfig{
					AgentId:          "support",
					Model:            "fast",
					SelectableModels: []string{"pro"},
				},
			},
		}))
	if code := connectCode(t, err); code != connect.CodeInvalidArgument {
		t.Fatalf("model list: code = %v, want InvalidArgument", code)
	}
}

func TestDestinationModelMustNameAKnownAlias(t *testing.T) {
	fx := newTelegramFixture(t)
	channel := fx.createChannel(t, "main", mainBotToken)

	_, err := fx.destinations.CreateTelegramDestination(ctxAs("owner", "owner", "ws-a"),
		connect.NewRequest(&agentsv1.CreateTelegramDestinationRequest{
			Destination: &agentsv1.TelegramDestination{
				Key: "ops", ChannelId: channel.GetId(), ChatId: "-100", InboundEnabled: true,
				Config: &agentsv1.TelegramDestinationConfig{AgentId: "support", Model: "nonexistent"},
			},
		}))
	if code := connectCode(t, err); code != connect.CodeInvalidArgument {
		t.Fatalf("code = %v, want InvalidArgument", code)
	}

	// A real alias is accepted.
	fx.createDestination(t, &agentsv1.TelegramDestination{
		Key: "ops-ok", ChannelId: channel.GetId(), ChatId: "-200", InboundEnabled: true,
		Config: &agentsv1.TelegramDestinationConfig{AgentId: "support", Model: "pro"},
	})
}

func TestUpdateDestinationRejectsAddressChanges(t *testing.T) {
	fx := newTelegramFixture(t)
	channel := fx.createChannel(t, "main", mainBotToken)
	dest := fx.createDestination(t, &agentsv1.TelegramDestination{
		Key: "ops", ChannelId: channel.GetId(), ChatId: "-100", MessageThreadId: "7",
	})

	for _, tc := range []struct {
		name   string
		mutate func(*agentsv1.TelegramDestination)
	}{
		{"chat", func(d *agentsv1.TelegramDestination) { d.ChatId = "-200" }},
		{"thread", func(d *agentsv1.TelegramDestination) { d.MessageThreadId = "9" }},
		{"channel", func(d *agentsv1.TelegramDestination) { d.ChannelId = "other-channel" }},
		{"key", func(d *agentsv1.TelegramDestination) { d.Key = "renamed" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := &agentsv1.TelegramDestination{
				Id:              dest.GetId(),
				Key:             dest.GetKey(),
				ChannelId:       dest.GetChannelId(),
				ChatId:          dest.GetChatId(),
				MessageThreadId: dest.GetMessageThreadId(),
				Revision:        dest.GetRevision(),
			}
			tc.mutate(input)
			_, err := fx.destinations.UpdateTelegramDestination(ctxAs("owner", "owner", "ws-a"),
				connect.NewRequest(&agentsv1.UpdateTelegramDestinationRequest{Destination: input}))
			if code := connectCode(t, err); code != connect.CodeInvalidArgument {
				t.Fatalf("code = %v, want InvalidArgument", code)
			}
		})
	}

	// Canonicalization means an equivalent-but-differently-written address
	// is not treated as a change.
	input := &agentsv1.TelegramDestination{
		Id: dest.GetId(), ChatId: "-100", MessageThreadId: "007",
		Name: "Ops room", Revision: dest.GetRevision(),
	}
	if _, err := fx.destinations.UpdateTelegramDestination(ctxAs("owner", "owner", "ws-a"),
		connect.NewRequest(&agentsv1.UpdateTelegramDestinationRequest{Destination: input})); err != nil {
		t.Fatalf("update with an equivalent address: %v", err)
	}
}

func TestUpdateDestinationRejectsAStaleRevision(t *testing.T) {
	fx := newTelegramFixture(t)
	channel := fx.createChannel(t, "main", mainBotToken)
	dest := fx.createDestination(t, &agentsv1.TelegramDestination{
		Key: "ops", ChannelId: channel.GetId(), ChatId: "-100",
	})

	if _, err := fx.destinations.UpdateTelegramDestination(ctxAs("owner", "owner", "ws-a"),
		connect.NewRequest(&agentsv1.UpdateTelegramDestinationRequest{
			Destination: &agentsv1.TelegramDestination{Id: dest.GetId(), Name: "First", Revision: 1},
		})); err != nil {
		t.Fatalf("first update: %v", err)
	}
	_, err := fx.destinations.UpdateTelegramDestination(ctxAs("owner", "owner", "ws-a"),
		connect.NewRequest(&agentsv1.UpdateTelegramDestinationRequest{
			Destination: &agentsv1.TelegramDestination{Id: dest.GetId(), Name: "Second", Revision: 1},
		}))
	if code := connectCode(t, err); code != connect.CodeAborted {
		t.Fatalf("code = %v, want Aborted", code)
	}
}

func TestChannelStatusReportsBlockersAndCounts(t *testing.T) {
	fx := newTelegramFixture(t)
	channel := fx.createChannel(t, "main", mainBotToken)

	resp, err := fx.channels.GetTelegramChannelStatus(ctxAs("member", "member", "ws-a"),
		connect.NewRequest(&agentsv1.GetTelegramChannelStatusRequest{ChannelId: channel.GetId()}))
	if err != nil {
		t.Fatalf("GetTelegramChannelStatus: %v", err)
	}
	status := resp.Msg.GetStatus()
	if len(status.GetBlockers()) == 0 {
		t.Fatal("expected the missing inbound destination to be reported as a blocker")
	}
	if status.GetInboundDesired() || status.GetOutboundDesired() {
		t.Error("a freshly created channel must report both desired states as off")
	}

	fx.createDestination(t, &agentsv1.TelegramDestination{
		Key: "ops", ChannelId: channel.GetId(), ChatId: "-100",
		InboundEnabled: true, OutboundEnabled: true,
		Config: &agentsv1.TelegramDestinationConfig{AgentId: "support"},
	})
	resp, err = fx.channels.GetTelegramChannelStatus(ctxAs("member", "member", "ws-a"),
		connect.NewRequest(&agentsv1.GetTelegramChannelStatusRequest{ChannelId: channel.GetId()}))
	if err != nil {
		t.Fatalf("GetTelegramChannelStatus: %v", err)
	}
	if got := resp.Msg.GetStatus().GetInboundDestinationCount(); got != 1 {
		t.Errorf("inbound_destination_count = %d, want 1", got)
	}
	if len(resp.Msg.GetStatus().GetBlockers()) != 0 {
		t.Errorf("blockers = %v, want none", resp.Msg.GetStatus().GetBlockers())
	}
}

func TestListDestinationsIsWorkspaceScoped(t *testing.T) {
	fx := newTelegramFixture(t)
	channel := fx.createChannel(t, "main", mainBotToken)
	fx.createDestination(t, &agentsv1.TelegramDestination{
		Key: "ops", ChannelId: channel.GetId(), ChatId: "-100",
	})

	resp, err := fx.destinations.ListTelegramDestinations(ctxAs("owner-b", "owner", "ws-b"),
		connect.NewRequest(&agentsv1.ListTelegramDestinationsRequest{}))
	if err != nil {
		t.Fatalf("ListTelegramDestinations: %v", err)
	}
	if len(resp.Msg.GetDestinations()) != 0 {
		t.Fatalf("another workspace saw %d destinations", len(resp.Msg.GetDestinations()))
	}
}

// telegramapiIdentityWithPrivacy builds a Bot with BotFather Group Privacy
// enabled, which Telegram reports as can_read_all_group_messages = false.
func telegramapiIdentityWithPrivacy() telegramapi.BotIdentity {
	identity := telegramtest.Identity(333333, "privatebot")
	identity.CanReadAllGroupMessages = false
	return identity
}
