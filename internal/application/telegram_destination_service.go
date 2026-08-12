package application

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"butterfly.orx.me/core/log"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	telegramrepo "go.orx.me/apps/butter/internal/repo/telegram"
	workspacerepo "go.orx.me/apps/butter/internal/repo/workspace"
	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// TelegramDestinationServiceServer implements
// agentsv1connect.TelegramDestinationServiceHandler (issue #264).
//
// A Destination is an address plus the policy for that address. Its address
// fields are immutable on purpose: Cron jobs and Notify Groups persist a
// Destination ID, so an editable address would let a metadata change silently
// redirect every downstream reference to a different chat.
type TelegramDestinationServiceServer struct {
	repo          telegramrepo.Repository
	workspaceRepo workspacerepo.Repository
	agentRepo     configrepo.AgentRepository
	modelRepo     configrepo.ModelProviderRepository
}

func NewTelegramDestinationServiceServer(repo telegramrepo.Repository) *TelegramDestinationServiceServer {
	return &TelegramDestinationServiceServer{repo: repo}
}

func (s *TelegramDestinationServiceServer) SetRepo(repo telegramrepo.Repository) { s.repo = repo }

func (s *TelegramDestinationServiceServer) SetWorkspaceRepo(repo workspacerepo.Repository) {
	s.workspaceRepo = repo
}

func (s *TelegramDestinationServiceServer) SetConfigRepos(agents configrepo.AgentRepository, models configrepo.ModelProviderRepository) {
	s.agentRepo = agents
	s.modelRepo = models
}

func (s *TelegramDestinationServiceServer) requireReady() error {
	if s.repo == nil {
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("telegram repository not configured"))
	}
	return nil
}

// --- Reads -----------------------------------------------------------------

func (s *TelegramDestinationServiceServer) ListTelegramDestinations(ctx context.Context, req *connect.Request[agentsv1.ListTelegramDestinationsRequest]) (*connect.Response[agentsv1.ListTelegramDestinationsResponse], error) {
	if err := s.requireReady(); err != nil {
		return nil, err
	}
	workspaceID, err := telegramWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	destinations, err := s.repo.ListDestinations(ctx, workspaceID, strings.TrimSpace(req.Msg.GetChannelId()))
	if err != nil {
		return nil, mapTelegramRepoErr(err)
	}
	return connect.NewResponse(&agentsv1.ListTelegramDestinationsResponse{Destinations: destinations}), nil
}

func (s *TelegramDestinationServiceServer) GetTelegramDestination(ctx context.Context, req *connect.Request[agentsv1.GetTelegramDestinationRequest]) (*connect.Response[agentsv1.GetTelegramDestinationResponse], error) {
	if err := s.requireReady(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Msg.GetId()) == "" {
		return nil, connectx.RequiredArgument("id")
	}
	workspaceID, err := telegramWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	dest, err := s.repo.GetDestination(ctx, workspaceID, req.Msg.GetId())
	if err != nil {
		return nil, mapTelegramRepoErr(err)
	}
	return connect.NewResponse(&agentsv1.GetTelegramDestinationResponse{Destination: dest}), nil
}

// --- Create ----------------------------------------------------------------

func (s *TelegramDestinationServiceServer) CreateTelegramDestination(ctx context.Context, req *connect.Request[agentsv1.CreateTelegramDestinationRequest]) (*connect.Response[agentsv1.CreateTelegramDestinationResponse], error) {
	if err := s.requireReady(); err != nil {
		return nil, err
	}
	workspaceID, err := telegramWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if err := requireTelegramManageRole(ctx, s.workspaceRepo, workspaceID, "create_destination"); err != nil {
		return nil, err
	}

	input := req.Msg.GetDestination()
	key := strings.TrimSpace(input.GetKey())
	if key == "" {
		return nil, connectx.RequiredArgument("destination.key")
	}
	channelID := strings.TrimSpace(input.GetChannelId())
	if channelID == "" {
		return nil, connectx.RequiredArgument("destination.channel_id")
	}
	// The Channel must exist in the same workspace: a Destination pointing at
	// another tenant's Bot would leak messages across workspaces.
	if _, err := s.repo.GetChannel(ctx, workspaceID, channelID); err != nil {
		if errors.Is(err, telegramrepo.ErrNotFound) {
			return nil, connectx.InvalidArgument("destination.channel_id", "references unknown channel")
		}
		return nil, mapTelegramRepoErr(err)
	}

	chatID, err := normalizeTelegramID("destination.chat_id", input.GetChatId())
	if err != nil {
		return nil, err
	}
	threadID, err := normalizeThreadID(input.GetMessageThreadId())
	if err != nil {
		return nil, err
	}

	config, err := s.validateConfig(ctx, workspaceID, input.GetConfig(), input.GetInboundEnabled())
	if err != nil {
		return nil, err
	}

	name := strings.TrimSpace(input.GetName())
	if name == "" {
		name = key
	}
	dest := &agentsv1.TelegramDestination{
		Id:              uuid.NewString(),
		Key:             key,
		Name:            name,
		ChannelId:       channelID,
		ChatId:          chatID,
		MessageThreadId: threadID,
		InboundEnabled:  input.GetInboundEnabled(),
		OutboundEnabled: input.GetOutboundEnabled(),
		Config:          config,
		// Verification is evidence, not configuration: it is earned by a real
		// send or receive, never asserted at creation. Creating a Destination
		// deliberately sends no Telegram message.
		Verification: &agentsv1.TelegramDestinationVerification{},
		WorkspaceId:  workspaceID,
	}

	created, err := s.repo.CreateDestination(ctx, workspaceID, dest)
	if err != nil {
		return nil, mapTelegramRepoErr(err)
	}
	log.FromContext(ctx).Info("telegram destination created",
		"workspace_id", workspaceID, "destination_id", created.GetId(),
		"destination_key", created.GetKey(), "channel_id", channelID,
		"chat_id", chatID, "message_thread_id", threadID)
	return connect.NewResponse(&agentsv1.CreateTelegramDestinationResponse{Destination: created}), nil
}

// --- Update ----------------------------------------------------------------

func (s *TelegramDestinationServiceServer) UpdateTelegramDestination(ctx context.Context, req *connect.Request[agentsv1.UpdateTelegramDestinationRequest]) (*connect.Response[agentsv1.UpdateTelegramDestinationResponse], error) {
	if err := s.requireReady(); err != nil {
		return nil, err
	}
	workspaceID, err := telegramWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if err := requireTelegramManageRole(ctx, s.workspaceRepo, workspaceID, "update_destination"); err != nil {
		return nil, err
	}

	input := req.Msg.GetDestination()
	if strings.TrimSpace(input.GetId()) == "" {
		return nil, connectx.RequiredArgument("destination.id")
	}
	current, err := s.repo.GetDestination(ctx, workspaceID, input.GetId())
	if err != nil {
		return nil, mapTelegramRepoErr(err)
	}
	if err := rejectImmutableDestinationChange(current, input); err != nil {
		return nil, err
	}

	config, err := s.validateConfig(ctx, workspaceID, input.GetConfig(), input.GetInboundEnabled())
	if err != nil {
		return nil, err
	}

	next := &agentsv1.TelegramDestination{
		Id:              current.GetId(),
		Key:             current.GetKey(),
		Name:            current.GetName(),
		ChannelId:       current.GetChannelId(),
		ChatId:          current.GetChatId(),
		MessageThreadId: current.GetMessageThreadId(),
		InboundEnabled:  input.GetInboundEnabled(),
		OutboundEnabled: input.GetOutboundEnabled(),
		Config:          config,
		// Verification is server-owned; a caller may not assert it.
		Verification: current.GetVerification(),
		WorkspaceId:  workspaceID,
	}
	if name := strings.TrimSpace(input.GetName()); name != "" {
		next.Name = name
	}

	updated, err := s.repo.UpdateDestination(ctx, workspaceID, next, input.GetRevision())
	if err != nil {
		return nil, mapTelegramRepoErr(err)
	}
	log.FromContext(ctx).Info("telegram destination updated",
		"workspace_id", workspaceID, "destination_id", updated.GetId(), "revision", updated.GetRevision())
	return connect.NewResponse(&agentsv1.UpdateTelegramDestinationResponse{Destination: updated}), nil
}

// rejectImmutableDestinationChange fails the request when the caller sent a
// different address. Accepting it silently — even by ignoring it — would let
// an operator believe traffic had been repointed.
func rejectImmutableDestinationChange(current, input *agentsv1.TelegramDestination) error {
	if key := strings.TrimSpace(input.GetKey()); key != "" && key != current.GetKey() {
		return connectx.InvalidArgument("destination.key", "is immutable")
	}
	if channelID := strings.TrimSpace(input.GetChannelId()); channelID != "" && channelID != current.GetChannelId() {
		return connectx.InvalidArgument("destination.channel_id",
			"is immutable; create a new destination for a different channel")
	}
	// Compare canonicalized values so "007" is not reported as a change
	// from "7".
	if raw := strings.TrimSpace(input.GetChatId()); raw != "" {
		chatID, err := normalizeTelegramID("destination.chat_id", raw)
		if err != nil {
			return err
		}
		if chatID != current.GetChatId() {
			return connectx.InvalidArgument("destination.chat_id",
				"is immutable; create a new destination for a different chat")
		}
	}
	if raw := strings.TrimSpace(input.GetMessageThreadId()); raw != "" {
		threadID, err := normalizeThreadID(raw)
		if err != nil {
			return err
		}
		if threadID != current.GetMessageThreadId() {
			return connectx.InvalidArgument("destination.message_thread_id",
				"is immutable; create a new destination for a different topic")
		}
	}
	return nil
}

// --- Delete ----------------------------------------------------------------

func (s *TelegramDestinationServiceServer) DeleteTelegramDestination(ctx context.Context, req *connect.Request[agentsv1.DeleteTelegramDestinationRequest]) (*connect.Response[agentsv1.DeleteTelegramDestinationResponse], error) {
	if err := s.requireReady(); err != nil {
		return nil, err
	}
	workspaceID, err := telegramWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if err := requireTelegramManageRole(ctx, s.workspaceRepo, workspaceID, "delete_destination"); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Msg.GetId()) == "" {
		return nil, connectx.RequiredArgument("id")
	}
	if err := s.repo.DeleteDestination(ctx, workspaceID, req.Msg.GetId()); err != nil {
		return nil, mapTelegramRepoErr(err)
	}
	log.FromContext(ctx).Info("telegram destination deleted",
		"workspace_id", workspaceID, "destination_id", req.Msg.GetId())
	return connect.NewResponse(&agentsv1.DeleteTelegramDestinationResponse{}), nil
}

// --- Config validation -----------------------------------------------------

// validateConfig normalizes and checks the Destination policy. It returns a
// fresh message so caller-supplied values can never bypass normalization.
//
// Every reference it checks is a strong reference: an Agent or Model that
// disappears leaves the topic silently broken, which is the failure mode the
// PRD calls out, so all of them are resolved at write time.
func (s *TelegramDestinationServiceServer) validateConfig(ctx context.Context, workspaceID string, in *agentsv1.TelegramDestinationConfig, inboundEnabled bool) (*agentsv1.TelegramDestinationConfig, error) {
	out := &agentsv1.TelegramDestinationConfig{
		TriggerMode:   in.GetTriggerMode(),
		SessionPolicy: in.GetSessionPolicy(),
		ReplyMode:     in.GetReplyMode(),
		DebugDefault:  in.GetDebugDefault(),
	}
	if out.TriggerMode == agentsv1.TelegramTriggerMode_TELEGRAM_TRIGGER_MODE_UNSPECIFIED {
		out.TriggerMode = agentsv1.TelegramTriggerMode_TELEGRAM_TRIGGER_MODE_ALL
	}
	if out.SessionPolicy == agentsv1.TelegramSessionPolicy_TELEGRAM_SESSION_POLICY_UNSPECIFIED {
		out.SessionPolicy = agentsv1.TelegramSessionPolicy_TELEGRAM_SESSION_POLICY_DESTINATION
	}
	if out.ReplyMode == agentsv1.TelegramReplyMode_TELEGRAM_REPLY_MODE_UNSPECIFIED {
		out.ReplyMode = agentsv1.TelegramReplyMode_TELEGRAM_REPLY_MODE_REPLY
	}

	allowed, err := normalizeUserIDs("config.allowed_user_ids", in.GetAllowedUserIds())
	if err != nil {
		return nil, err
	}
	controllers, err := normalizeUserIDs("config.controller_user_ids", in.GetControllerUserIds())
	if err != nil {
		return nil, err
	}
	// A controller who is not admitted could never reach the Destination, so
	// the pairing would look configured but do nothing.
	if len(allowed) > 0 {
		for _, controller := range controllers {
			if !slices.Contains(allowed, controller) {
				return nil, connectx.InvalidArgument("config.controller_user_ids",
					fmt.Sprintf("user %s must also appear in allowed_user_ids", controller))
			}
		}
	}
	out.AllowedUserIds = allowed
	out.ControllerUserIds = controllers

	agentID := strings.TrimSpace(in.GetAgentId())
	selectableAgents := trimmedUnique(in.GetSelectableAgentIds())
	if inboundEnabled && agentID == "" {
		return nil, connectx.RequiredArgument("config.agent_id")
	}
	if agentID != "" {
		if _, err := resolveActiveAgent(ctx, s.agentRepo, workspaceID, "config.agent_id", agentID); err != nil {
			return nil, err
		}
	}
	for _, candidate := range selectableAgents {
		if _, err := resolveActiveAgent(ctx, s.agentRepo, workspaceID, "config.selectable_agent_ids", candidate); err != nil {
			return nil, err
		}
	}
	// A default outside the candidate list would be unreachable after the
	// first switch: /agent could never return to it.
	if len(selectableAgents) > 0 && agentID != "" && !slices.Contains(selectableAgents, agentID) {
		return nil, connectx.InvalidArgument("config.selectable_agent_ids",
			fmt.Sprintf("must include the default agent %q when agent selection is enabled", agentID))
	}
	out.AgentId = agentID
	out.SelectableAgentIds = selectableAgents

	model := strings.TrimSpace(in.GetModel())
	selectableModels := trimmedUnique(in.GetSelectableModels())
	if model != "" || len(selectableModels) > 0 {
		aliases, err := workspaceModelAliases(ctx, s.modelRepo, workspaceID)
		if err != nil {
			return nil, err
		}
		if model != "" {
			if _, ok := aliases[model]; !ok {
				return nil, connectx.InvalidArgument("config.model",
					fmt.Sprintf("references unknown model alias %q", model))
			}
		}
		for _, candidate := range selectableModels {
			if _, ok := aliases[candidate]; !ok {
				return nil, connectx.InvalidArgument("config.selectable_models",
					fmt.Sprintf("references unknown model alias %q", candidate))
			}
		}
	}
	if len(selectableModels) > 0 && model != "" && !slices.Contains(selectableModels, model) {
		return nil, connectx.InvalidArgument("config.selectable_models",
			fmt.Sprintf("must include the default model %q when model selection is enabled", model))
	}
	out.Model = model
	out.SelectableModels = selectableModels

	return out, nil
}

// trimmedUnique drops blanks and duplicates while preserving order.
func trimmedUnique(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, dup := seen[trimmed]; dup {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}
