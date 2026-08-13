package application

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"butterfly.orx.me/core/log"
	telegramrepo "go.orx.me/apps/butter/internal/repo/telegram"
	"go.orx.me/apps/butter/internal/repo/telegramsetting"
	workspacerepo "go.orx.me/apps/butter/internal/repo/workspace"
	telegramruntime "go.orx.me/apps/butter/internal/runtime/telegram"
	"go.orx.me/apps/butter/internal/secretbox"
	"go.orx.me/apps/butter/internal/telegramapi"
	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// TelegramChannelServiceServer implements
// agentsv1connect.TelegramChannelServiceHandler (issue #264).
//
// A Channel is a Bot transport, not an address. The rules this service
// enforces all follow from that: the Bot identity is resolved once from the
// credential and then pinned, so rotating a token can never silently move
// every Destination to a different Bot; and the token itself is write-only,
// so no read path can leak it.
type TelegramChannelServiceServer struct {
	repo          telegramrepo.Repository
	workspaceRepo workspacerepo.Repository
	keyring       *secretbox.Keyring
	// botFactory builds a Telegram client from a decrypted Bot Token. Tests
	// substitute a fake; production uses telegramapi.NewFactory.
	botFactory telegramapi.Factory
	// settings supplies the platform Webhook base URL, a prerequisite for
	// enabling a Webhook Channel.
	settings telegramsetting.Repository
	// queue reports whether durable Redis Streams infrastructure is present.
	// Webhook mode depends on it, so enabling without it would accept
	// updates the fleet cannot durably hold.
	queue QueueProbe
	// webhookStatus supplies the reconciler's observed registration state.
	webhookStatus WebhookStatusSource
	// pollingStatus supplies this Pod's long-poll observations.
	pollingStatus PollingStatusSource
}

// PollingStatusSource exposes this Pod's long-poll state for a Channel.
type PollingStatusSource interface {
	Status(channelID string) (telegramruntime.PollingStatus, bool)
}

// QueueProbe reports whether the durable update queue is usable.
type QueueProbe interface {
	Available() bool
	Ping(ctx context.Context) error
	CheckReady(ctx context.Context) error
	CheckLeaseReady(ctx context.Context) error
}

// WebhookStatusSource exposes the reconciler's last observation for a Channel.
type WebhookStatusSource interface {
	State(channelID string) (telegramruntime.ReconcileState, bool)
}

func NewTelegramChannelServiceServer(repo telegramrepo.Repository) *TelegramChannelServiceServer {
	return &TelegramChannelServiceServer{
		repo:       repo,
		botFactory: telegramapi.NewFactory(),
	}
}

func (s *TelegramChannelServiceServer) SetRepo(repo telegramrepo.Repository) { s.repo = repo }

func (s *TelegramChannelServiceServer) SetWorkspaceRepo(repo workspacerepo.Repository) {
	s.workspaceRepo = repo
}

func (s *TelegramChannelServiceServer) SetKeyring(keyring *secretbox.Keyring) { s.keyring = keyring }

// SetSettingsRepo wires the platform settings repository.
func (s *TelegramChannelServiceServer) SetSettingsRepo(repo telegramsetting.Repository) {
	s.settings = repo
}

// SetQueueProbe wires the durable-queue readiness check.
func (s *TelegramChannelServiceServer) SetQueueProbe(probe QueueProbe) { s.queue = probe }

// SetWebhookStatusSource wires the reconciler's observed state.
func (s *TelegramChannelServiceServer) SetWebhookStatusSource(source WebhookStatusSource) {
	s.webhookStatus = source
}

// SetPollingStatusSource wires the long-poll supervisor's observations.
func (s *TelegramChannelServiceServer) SetPollingStatusSource(source PollingStatusSource) {
	s.pollingStatus = source
}

// SetBotFactory overrides how Telegram clients are built. Used by tests.
func (s *TelegramChannelServiceServer) SetBotFactory(factory telegramapi.Factory) {
	if factory != nil {
		s.botFactory = factory
	}
}

func (s *TelegramChannelServiceServer) requireReady() error {
	if s.repo == nil {
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("telegram repository not configured"))
	}
	if s.keyring == nil {
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("credential encryption is not configured"))
	}
	return nil
}

// --- Reads -----------------------------------------------------------------

func (s *TelegramChannelServiceServer) ListTelegramChannels(ctx context.Context, _ *connect.Request[agentsv1.ListTelegramChannelsRequest]) (*connect.Response[agentsv1.ListTelegramChannelsResponse], error) {
	if err := s.requireReady(); err != nil {
		return nil, err
	}
	workspaceID, err := telegramWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	channels, err := s.repo.ListChannels(ctx, workspaceID)
	if err != nil {
		return nil, mapTelegramRepoErr(err)
	}
	return connect.NewResponse(&agentsv1.ListTelegramChannelsResponse{Channels: channels}), nil
}

func (s *TelegramChannelServiceServer) GetTelegramChannel(ctx context.Context, req *connect.Request[agentsv1.GetTelegramChannelRequest]) (*connect.Response[agentsv1.GetTelegramChannelResponse], error) {
	channel, err := s.loadChannel(ctx, req.Msg.GetId())
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&agentsv1.GetTelegramChannelResponse{Channel: channel}), nil
}

func (s *TelegramChannelServiceServer) loadChannel(ctx context.Context, id string) (*agentsv1.TelegramChannel, error) {
	if err := s.requireReady(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(id) == "" {
		return nil, connectx.RequiredArgument("id")
	}
	workspaceID, err := telegramWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	channel, err := s.repo.GetChannel(ctx, workspaceID, id)
	if err != nil {
		return nil, mapTelegramRepoErr(err)
	}
	return channel, nil
}

// --- Create ----------------------------------------------------------------

func (s *TelegramChannelServiceServer) CreateTelegramChannel(ctx context.Context, req *connect.Request[agentsv1.CreateTelegramChannelRequest]) (*connect.Response[agentsv1.CreateTelegramChannelResponse], error) {
	if err := s.requireReady(); err != nil {
		return nil, err
	}
	workspaceID, err := telegramWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if err := requireTelegramManageRole(ctx, s.workspaceRepo, workspaceID, "create_channel"); err != nil {
		return nil, err
	}

	input := req.Msg.GetChannel()
	key := strings.TrimSpace(input.GetKey())
	if key == "" {
		return nil, connectx.RequiredArgument("channel.key")
	}
	token := strings.TrimSpace(req.Msg.GetBotToken())
	if token == "" {
		return nil, connectx.RequiredArgument("bot_token")
	}

	// Validate before anything is committed: a Channel that exists but has
	// never resolved to a Bot would be an address book with no transport.
	identity, err := s.botFactory(token).GetMe(ctx)
	if err != nil {
		return nil, mapTelegramAPIErr("bot_token", err)
	}

	ciphertext, keyID, err := s.keyring.Encrypt(ctx, []byte(token))
	if err != nil {
		return nil, connectx.InternalWith(fmt.Errorf("encrypt bot token: %w", err))
	}

	receiveMode := input.GetReceiveMode()
	if receiveMode == agentsv1.TelegramReceiveMode_TELEGRAM_RECEIVE_MODE_UNSPECIFIED {
		// Webhook is the multi-Pod-correct default; Long Polling is opt-in.
		receiveMode = agentsv1.TelegramReceiveMode_TELEGRAM_RECEIVE_MODE_WEBHOOK
	}
	name := strings.TrimSpace(input.GetName())
	if name == "" {
		name = key
	}

	channel := &agentsv1.TelegramChannel{
		Id:              uuid.NewString(),
		Key:             key,
		Name:            name,
		BotId:           telegramapi.FormatID(identity.ID),
		BotUsername:     identity.Username,
		BotCapabilities: capabilitiesOf(identity),
		ReceiveMode:     receiveMode,
		// Channels are always created disabled. Enablement runs a separate
		// preflight that later tickets extend with receive-mode checks.
		InboundEnabled:  false,
		OutboundEnabled: false,
		CredentialState: agentsv1.TelegramCredentialState_TELEGRAM_CREDENTIAL_STATE_VALID,
		WorkspaceId:     workspaceID,
	}

	credentials := telegramrepo.ChannelCredentials{
		BotToken: telegramrepo.Credential{Ciphertext: ciphertext, KeyID: keyID},
	}
	if receiveMode == agentsv1.TelegramReceiveMode_TELEGRAM_RECEIVE_MODE_WEBHOOK {
		secret, err := generateWebhookSecret()
		if err != nil {
			return nil, connectx.InternalWith(err)
		}
		secretCiphertext, secretKeyID, err := s.keyring.Encrypt(ctx, []byte(secret))
		if err != nil {
			return nil, connectx.InternalWith(fmt.Errorf("encrypt webhook secret: %w", err))
		}
		credentials.WebhookSecret = telegramrepo.Credential{
			Ciphertext: secretCiphertext,
			KeyID:      secretKeyID,
		}
	}

	created, err := s.repo.CreateChannel(ctx, workspaceID, channel, credentials)
	if err != nil {
		return nil, mapTelegramRepoErr(err)
	}
	log.FromContext(ctx).Info("telegram channel created",
		"workspace_id", workspaceID, "channel_id", created.GetId(),
		"channel_key", created.GetKey(), "bot_id", created.GetBotId(),
		"receive_mode", created.GetReceiveMode().String())
	return connect.NewResponse(&agentsv1.CreateTelegramChannelResponse{Channel: created}), nil
}

// ensureWebhookSecret generates and stores a high-entropy per-Channel secret
// unless one already exists.
//
// The secret is per Channel, not global: Telegram echoes it on every callback,
// so a shared secret would let one compromised Channel authenticate callbacks
// for every other. It is stored through the same encrypted credential seam as
// the Bot Token and is never returned by any read.
func (s *TelegramChannelServiceServer) ensureWebhookSecret(ctx context.Context, workspaceID string, channel *agentsv1.TelegramChannel) error {
	if channel.GetReceiveMode() != agentsv1.TelegramReceiveMode_TELEGRAM_RECEIVE_MODE_WEBHOOK {
		return nil
	}

	secret, err := generateWebhookSecret()
	if err != nil {
		return connectx.InternalWith(err)
	}
	ciphertext, keyID, err := s.keyring.Encrypt(ctx, []byte(secret))
	if err != nil {
		return connectx.InternalWith(fmt.Errorf("encrypt webhook secret: %w", err))
	}
	if _, err := s.repo.SetWebhookSecretIfAbsent(ctx, workspaceID, channel.GetId(),
		telegramrepo.Credential{Ciphertext: ciphertext, KeyID: keyID}); err != nil {
		return mapTelegramRepoErr(err)
	}
	return nil
}

// generateWebhookSecret produces a token within Telegram's allowed alphabet
// (A-Z, a-z, 0-9, _ and -) and length (1..256).
func generateWebhookSecret() (string, error) {
	raw := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, raw); err != nil {
		return "", fmt.Errorf("generate webhook secret: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func capabilitiesOf(identity telegramapi.BotIdentity) *agentsv1.TelegramBotCapabilities {
	return &agentsv1.TelegramBotCapabilities{
		CanJoinGroups:           identity.CanJoinGroups,
		CanReadAllGroupMessages: identity.CanReadAllGroupMessages,
		SupportsInlineQueries:   identity.SupportsInlineQueries,
	}
}

// --- Update ----------------------------------------------------------------

func (s *TelegramChannelServiceServer) UpdateTelegramChannel(ctx context.Context, req *connect.Request[agentsv1.UpdateTelegramChannelRequest]) (*connect.Response[agentsv1.UpdateTelegramChannelResponse], error) {
	if err := s.requireReady(); err != nil {
		return nil, err
	}
	workspaceID, err := telegramWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if err := requireTelegramManageRole(ctx, s.workspaceRepo, workspaceID, "update_channel"); err != nil {
		return nil, err
	}

	input := req.Msg.GetChannel()
	if strings.TrimSpace(input.GetId()) == "" {
		return nil, connectx.RequiredArgument("channel.id")
	}
	current, err := s.repo.GetChannel(ctx, workspaceID, input.GetId())
	if err != nil {
		return nil, mapTelegramRepoErr(err)
	}
	if err := rejectImmutableChannelChange(current, input); err != nil {
		return nil, err
	}

	next := cloneChannelForUpdate(current)
	if name := strings.TrimSpace(input.GetName()); name != "" {
		next.Name = name
	}
	if mode := input.GetReceiveMode(); mode != agentsv1.TelegramReceiveMode_TELEGRAM_RECEIVE_MODE_UNSPECIFIED {
		next.ReceiveMode = mode
	}
	// Provision the Webhook secret before publishing Webhook mode. If the mode
	// CAS later fails, the Channel merely has an unused secret; the reverse
	// order could publish a Webhook Channel that rejects every callback.
	if next.GetReceiveMode() == agentsv1.TelegramReceiveMode_TELEGRAM_RECEIVE_MODE_WEBHOOK &&
		!current.GetWebhookSecretSet() {
		if err := s.ensureWebhookSecret(ctx, workspaceID, next); err != nil {
			return nil, err
		}
		next.WebhookSecretSet = true
	}

	updated, err := s.repo.UpdateChannel(ctx, workspaceID, next, input.GetRevision())
	if err != nil {
		return nil, mapTelegramRepoErr(err)
	}
	log.FromContext(ctx).Info("telegram channel updated",
		"workspace_id", workspaceID, "channel_id", updated.GetId(), "revision", updated.GetRevision())
	return connect.NewResponse(&agentsv1.UpdateTelegramChannelResponse{Channel: updated}), nil
}

// rejectImmutableChannelChange fails the request when the caller sent a
// different value for a field that may never change. Silently ignoring the
// input would let an operator believe a rename or Bot swap had taken effect.
func rejectImmutableChannelChange(current, input *agentsv1.TelegramChannel) error {
	if key := strings.TrimSpace(input.GetKey()); key != "" && key != current.GetKey() {
		return connectx.InvalidArgument("channel.key", "is immutable")
	}
	if botID := strings.TrimSpace(input.GetBotId()); botID != "" && botID != current.GetBotId() {
		return connectx.InvalidArgument("channel.bot_id", "is immutable; rotate the credential instead")
	}
	return nil
}

// cloneChannelForUpdate copies the stored Channel, keeping every field the
// update RPC does not own — notably enablement, which SetTelegramChannelEnabled
// owns so a stale form submission cannot re-enable a Channel an operator just
// turned off.
func cloneChannelForUpdate(current *agentsv1.TelegramChannel) *agentsv1.TelegramChannel {
	next := &agentsv1.TelegramChannel{}
	next.Id = current.GetId()
	next.Key = current.GetKey()
	next.Name = current.GetName()
	next.BotId = current.GetBotId()
	next.BotUsername = current.GetBotUsername()
	next.BotCapabilities = current.GetBotCapabilities()
	next.ReceiveMode = current.GetReceiveMode()
	next.InboundEnabled = current.GetInboundEnabled()
	next.OutboundEnabled = current.GetOutboundEnabled()
	next.CredentialState = current.GetCredentialState()
	next.LastCredentialError = current.GetLastCredentialError()
	next.WorkspaceId = current.GetWorkspaceId()
	return next
}

// --- Credential rotation ---------------------------------------------------

func (s *TelegramChannelServiceServer) RotateTelegramChannelCredential(ctx context.Context, req *connect.Request[agentsv1.RotateTelegramChannelCredentialRequest]) (*connect.Response[agentsv1.RotateTelegramChannelCredentialResponse], error) {
	if err := s.requireReady(); err != nil {
		return nil, err
	}
	workspaceID, err := telegramWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if err := requireTelegramManageRole(ctx, s.workspaceRepo, workspaceID, "rotate_channel_credential"); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Msg.GetChannelId()) == "" {
		return nil, connectx.RequiredArgument("channel_id")
	}
	token := strings.TrimSpace(req.Msg.GetBotToken())
	if token == "" {
		return nil, connectx.RequiredArgument("bot_token")
	}

	current, err := s.repo.GetChannel(ctx, workspaceID, req.Msg.GetChannelId())
	if err != nil {
		return nil, mapTelegramRepoErr(err)
	}

	// Validate first, and only against the pinned Bot ID. A token for a
	// different Bot would otherwise repoint every Destination under this
	// Channel at another account without any of them changing.
	identity, err := s.botFactory(token).GetMe(ctx)
	if err != nil {
		return nil, mapTelegramAPIErr("bot_token", err)
	}
	if got := telegramapi.FormatID(identity.ID); got != current.GetBotId() {
		return nil, connectx.InvalidArgument("bot_token",
			fmt.Sprintf("resolves to bot %s but this channel is pinned to bot %s", got, current.GetBotId()))
	}

	ciphertext, keyID, err := s.keyring.Encrypt(ctx, []byte(token))
	if err != nil {
		return nil, connectx.InternalWith(fmt.Errorf("encrypt bot token: %w", err))
	}
	next := cloneChannelForUpdate(current)
	next.BotUsername = identity.Username
	next.BotCapabilities = capabilitiesOf(identity)
	next.CredentialState = agentsv1.TelegramCredentialState_TELEGRAM_CREDENTIAL_STATE_VALID
	next.LastCredentialError = ""
	updated, err := s.repo.RotateChannelCredential(ctx, workspaceID, next,
		telegramrepo.Credential{Ciphertext: ciphertext, KeyID: keyID}, current.GetRevision())
	if err != nil {
		return nil, mapTelegramRepoErr(err)
	}
	log.FromContext(ctx).Info("telegram channel credential rotated",
		"workspace_id", workspaceID, "channel_id", updated.GetId(), "bot_id", updated.GetBotId())
	return connect.NewResponse(&agentsv1.RotateTelegramChannelCredentialResponse{Channel: updated}), nil
}

// --- Enablement ------------------------------------------------------------

func (s *TelegramChannelServiceServer) SetTelegramChannelEnabled(ctx context.Context, req *connect.Request[agentsv1.SetTelegramChannelEnabledRequest]) (*connect.Response[agentsv1.SetTelegramChannelEnabledResponse], error) {
	if err := s.requireReady(); err != nil {
		return nil, err
	}
	workspaceID, err := telegramWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if err := requireTelegramManageRole(ctx, s.workspaceRepo, workspaceID, "set_channel_enabled"); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Msg.GetChannelId()) == "" {
		return nil, connectx.RequiredArgument("channel_id")
	}
	current, err := s.repo.GetChannel(ctx, workspaceID, req.Msg.GetChannelId())
	if err != nil {
		return nil, mapTelegramRepoErr(err)
	}

	inbound := req.Msg.GetInboundEnabled()
	outbound := req.Msg.GetOutboundEnabled()
	if inbound && !outbound {
		return nil, connectx.InvalidArgument("outbound_enabled",
			"must be true when inbound is enabled: every accepted interaction must be able to reply")
	}

	warnings, err := s.preflight(ctx, workspaceID, current, inbound, outbound)
	if err != nil {
		return nil, err
	}

	next := cloneChannelForUpdate(current)
	next.InboundEnabled = inbound
	next.OutboundEnabled = outbound
	updated, err := s.repo.UpdateChannel(ctx, workspaceID, next, req.Msg.GetRevision())
	if err != nil {
		return nil, mapTelegramRepoErr(err)
	}
	log.FromContext(ctx).Info("telegram channel enablement changed",
		"workspace_id", workspaceID, "channel_id", updated.GetId(),
		"inbound", inbound, "outbound", outbound)
	return connect.NewResponse(&agentsv1.SetTelegramChannelEnabledResponse{
		Channel:  updated,
		Warnings: warnings,
	}), nil
}

// preflight refuses enablement whose prerequisites are known to be missing,
// and returns non-blocking warnings for degraded — but workable —
// configurations. Disabling is never blocked.
func (s *TelegramChannelServiceServer) preflight(ctx context.Context, workspaceID string, channel *agentsv1.TelegramChannel, inbound, outbound bool) ([]string, error) {
	if !inbound && !outbound {
		return nil, nil
	}
	blockers, warnings, err := s.evaluate(ctx, workspaceID, channel, inbound)
	if err != nil {
		return nil, err
	}
	if len(blockers) > 0 {
		return nil, connect.NewError(connect.CodeFailedPrecondition,
			fmt.Errorf("channel cannot be enabled: %s", strings.Join(blockers, "; ")))
	}
	return warnings, nil
}

// evaluate collects the prerequisites and degradations for a Channel.
// Later tickets extend it with receive-mode prerequisites (global Webhook
// base URL, Redis durability) rather than adding a second preflight path.
func (s *TelegramChannelServiceServer) evaluate(ctx context.Context, workspaceID string, channel *agentsv1.TelegramChannel, inbound bool) (blockers, warnings []string, err error) {
	if channel.GetCredentialState() == agentsv1.TelegramCredentialState_TELEGRAM_CREDENTIAL_STATE_MISSING {
		blockers = append(blockers, "no bot token is stored")
	}
	if channel.GetCredentialState() == agentsv1.TelegramCredentialState_TELEGRAM_CREDENTIAL_STATE_INVALID {
		blockers = append(blockers, "the stored bot token was rejected by Telegram")
	}
	if inbound && channel.GetReceiveMode() == agentsv1.TelegramReceiveMode_TELEGRAM_RECEIVE_MODE_LONG_POLLING {
		// Long Polling needs the same durable queue as Webhook mode — a
		// fetched update is only confirmed to Telegram once it is safely
		// enqueued — plus a lease store to elect the single consumer.
		if s.queue == nil || !s.queue.Available() {
			blockers = append(blockers,
				"redis is not configured, which long polling requires for the update queue and the consumer lease")
		} else if pingErr := s.queue.Ping(ctx); pingErr != nil {
			blockers = append(blockers, "the update queue is unreachable")
		} else if readyErr := s.queue.CheckReady(ctx); readyErr != nil {
			blockers = append(blockers, "the update queue is not durable: "+readyErr.Error())
		} else if leaseErr := s.queue.CheckLeaseReady(ctx); leaseErr != nil {
			blockers = append(blockers, "the long polling consumer lease is unavailable: "+leaseErr.Error())
		}
	}
	if inbound && channel.GetReceiveMode() == agentsv1.TelegramReceiveMode_TELEGRAM_RECEIVE_MODE_WEBHOOK {
		// Webhook mode has infrastructure prerequisites that Long Polling
		// does not: Telegram must have somewhere to deliver to, and the
		// delivery must land somewhere durable before it is acknowledged.
		if s.settings == nil {
			blockers = append(blockers, "telegram platform settings are not available")
		} else {
			settings, settingsErr := s.settings.Get(ctx)
			if settingsErr != nil {
				return nil, nil, connectx.InternalWith(settingsErr)
			}
			if strings.TrimSpace(settings.GetWebhookBaseUrl()) == "" {
				blockers = append(blockers,
					"no public webhook base URL is configured; a global administrator must set one")
			}
		}
		if s.queue == nil || !s.queue.Available() {
			blockers = append(blockers,
				"redis is not configured as a durable update queue, which webhook mode requires")
		} else if readyErr := s.queue.CheckReady(ctx); readyErr != nil {
			blockers = append(blockers, "the update queue is not durable: "+readyErr.Error())
		}
	}
	if inbound {
		destinations, listErr := s.repo.ListDestinations(ctx, workspaceID, channel.GetId())
		if listErr != nil {
			return nil, nil, mapTelegramRepoErr(listErr)
		}
		enabled := 0
		for _, dest := range destinations {
			if dest.GetInboundEnabled() {
				enabled++
			}
		}
		if enabled == 0 {
			blockers = append(blockers, "no inbound destination is enabled for this channel")
		}
		if !channel.GetBotCapabilities().GetCanReadAllGroupMessages() {
			// Not a blocker: private chats and command/mention triggers still
			// work. Butter cannot turn Group Privacy off, only report it.
			warnings = append(warnings,
				"BotFather Group Privacy is enabled: in groups the bot only receives commands and replies, so ALL trigger mode will not see ordinary messages")
		}
	}
	return blockers, warnings, nil
}

// --- Status ----------------------------------------------------------------

func (s *TelegramChannelServiceServer) GetTelegramChannelStatus(ctx context.Context, req *connect.Request[agentsv1.GetTelegramChannelStatusRequest]) (*connect.Response[agentsv1.GetTelegramChannelStatusResponse], error) {
	if err := s.requireReady(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Msg.GetChannelId()) == "" {
		return nil, connectx.RequiredArgument("channel_id")
	}
	workspaceID, err := telegramWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	channel, err := s.repo.GetChannel(ctx, workspaceID, req.Msg.GetChannelId())
	if err != nil {
		return nil, mapTelegramRepoErr(err)
	}
	destinations, err := s.repo.ListDestinations(ctx, workspaceID, channel.GetId())
	if err != nil {
		return nil, mapTelegramRepoErr(err)
	}
	var inboundCount, outboundCount int32
	for _, dest := range destinations {
		if dest.GetInboundEnabled() {
			inboundCount++
		}
		if dest.GetOutboundEnabled() {
			outboundCount++
		}
	}
	// Evaluate against inbound=true so the Dashboard can show what would
	// block enabling, not merely what blocks the current state.
	blockers, warnings, err := s.evaluate(ctx, workspaceID, channel, true)
	if err != nil {
		return nil, err
	}
	status := &agentsv1.TelegramChannelStatus{
		ChannelId:                channel.GetId(),
		InboundDesired:           channel.GetInboundEnabled(),
		OutboundDesired:          channel.GetOutboundEnabled(),
		CredentialState:          channel.GetCredentialState(),
		ReceiveMode:              channel.GetReceiveMode(),
		InboundDestinationCount:  inboundCount,
		OutboundDestinationCount: outboundCount,
		Blockers:                 blockers,
		Warnings:                 warnings,
		LastCredentialError:      channel.GetLastCredentialError(),
		QueueReady:               s.queue != nil && s.queue.Available(),
	}
	// Long Polling leadership is per Pod, so this reports what *this* Pod
	// currently holds rather than a global fact.
	if s.pollingStatus != nil {
		if observed, ok := s.pollingStatus.Status(channel.GetId()); ok {
			status.PollingLeader = observed.Leader
			status.LastFetchedUpdateId = observed.LastFetchedUpdateID
			status.LastAcceptedUpdateId = observed.LastAcceptedUpdateID
			if observed.LastError != "" {
				status.LastReceiveError = observed.LastError
			}
			if !observed.LastPolledAt.IsZero() {
				status.LastPolledAt = timestamppb.New(observed.LastPolledAt)
			}
		}
	}
	// Registration state is observed by the reconciler, never persisted:
	// availability is a runtime fact, not configuration.
	if s.webhookStatus != nil {
		if observed, ok := s.webhookStatus.State(channel.GetId()); ok {
			status.WebhookState = observed.State
			status.WebhookUrl = observed.URL
			status.LastWebhookError = observed.Error
			if !observed.ReconciledAt.IsZero() {
				status.LastReconciledAt = timestamppb.New(observed.ReconciledAt)
			}
		}
	}
	if status.GetWebhookUrl() == "" && s.settings != nil {
		if settings, err := s.settings.Get(ctx); err == nil &&
			strings.TrimSpace(settings.GetWebhookBaseUrl()) != "" {
			status.WebhookUrl = telegramruntime.CallbackURL(settings.GetWebhookBaseUrl(), channel.GetId())
		}
	}
	return connect.NewResponse(&agentsv1.GetTelegramChannelStatusResponse{Status: status}), nil
}

// --- Delete ----------------------------------------------------------------

func (s *TelegramChannelServiceServer) DeleteTelegramChannel(ctx context.Context, req *connect.Request[agentsv1.DeleteTelegramChannelRequest]) (*connect.Response[agentsv1.DeleteTelegramChannelResponse], error) {
	if err := s.requireReady(); err != nil {
		return nil, err
	}
	workspaceID, err := telegramWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if err := requireTelegramManageRole(ctx, s.workspaceRepo, workspaceID, "delete_channel"); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Msg.GetId()) == "" {
		return nil, connectx.RequiredArgument("id")
	}

	// Name the blocking Destinations so the operator can repair references
	// deliberately instead of guessing which ones to remove.
	destinations, err := s.repo.ListDestinations(ctx, workspaceID, req.Msg.GetId())
	if err != nil {
		return nil, mapTelegramRepoErr(err)
	}
	if len(destinations) > 0 {
		ids := make([]string, 0, len(destinations))
		for _, dest := range destinations {
			ids = append(ids, dest.GetId())
		}
		return nil, connect.NewError(connect.CodeFailedPrecondition,
			fmt.Errorf("channel is referenced by destinations: %s", strings.Join(ids, ", ")))
	}

	if err := s.repo.DeleteChannel(ctx, workspaceID, req.Msg.GetId()); err != nil {
		return nil, mapTelegramRepoErr(err)
	}
	log.FromContext(ctx).Info("telegram channel deleted",
		"workspace_id", workspaceID, "channel_id", req.Msg.GetId())
	return connect.NewResponse(&agentsv1.DeleteTelegramChannelResponse{}), nil
}
