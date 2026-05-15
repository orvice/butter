package application

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/twitchtv/twirp"
	"google.golang.org/genai"

	configrepo "go.orx.me/apps/butter/internal/repo/config"
	"go.orx.me/apps/butter/internal/repo/invocation"
	"go.orx.me/apps/butter/internal/runtime/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/proto"
)

type AgentServiceServer struct {
	repo      configrepo.AgentRepository
	runtime   ConfigRuntime
	runnerSvc *runner.Service
	invRepo   invocation.Repository
}

func NewAgentServiceServer(repo configrepo.AgentRepository) *AgentServiceServer {
	return &AgentServiceServer{repo: repo}
}

func (s *AgentServiceServer) SetRuntime(runtime ConfigRuntime) {
	s.runtime = runtime
}

// SetRunnerService wires the runner so InvokeAgent can execute agents.
func (s *AgentServiceServer) SetRunnerService(svc *runner.Service) {
	s.runnerSvc = svc
}

// SetInvocationRepo wires the invocation repository used by
// ListAgentInvocations.
func (s *AgentServiceServer) SetInvocationRepo(repo invocation.Repository) {
	s.invRepo = repo
}

func (s *AgentServiceServer) ListAgents(ctx context.Context, _ *agentsv1.ListAgentsRequest) (*agentsv1.ListAgentsResponse, error) {
	agents, err := s.repo.ListAgents(ctx)
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.ListAgentsResponse{Agents: agents}, nil
}

func (s *AgentServiceServer) GetAgent(ctx context.Context, req *agentsv1.GetAgentRequest) (*agentsv1.GetAgentResponse, error) {
	a, err := s.repo.GetAgent(ctx, req.GetName())
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.GetAgentResponse{Agent: a}, nil
}

func (s *AgentServiceServer) CreateAgent(ctx context.Context, req *agentsv1.CreateAgentRequest) (*agentsv1.CreateAgentResponse, error) {
	a, err := mutateWithRuntime(
		func() (*agentsv1.Agent, error) {
			return s.repo.CreateAgent(ctx, req.GetAgent())
		},
		func() error {
			return s.reloadRuntime(ctx)
		},
		func() error {
			if err := s.repo.DeleteAgent(ctx, req.GetAgent().GetName()); err != nil {
				return err
			}
			return s.reloadRuntime(ctx)
		},
	)
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.CreateAgentResponse{Agent: a}, nil
}

func (s *AgentServiceServer) UpdateAgent(ctx context.Context, req *agentsv1.UpdateAgentRequest) (*agentsv1.UpdateAgentResponse, error) {
	prev, err := s.repo.GetAgent(ctx, req.GetAgent().GetName())
	if err != nil {
		return nil, toTwirpError(err)
	}

	a, err := mutateWithRuntime(
		func() (*agentsv1.Agent, error) {
			return s.repo.UpdateAgent(ctx, req.GetAgent())
		},
		func() error {
			return s.reloadRuntime(ctx)
		},
		func() error {
			if _, err := s.repo.UpdateAgent(ctx, proto.Clone(prev).(*agentsv1.Agent)); err != nil {
				return err
			}
			return s.reloadRuntime(ctx)
		},
	)
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.UpdateAgentResponse{Agent: a}, nil
}

func (s *AgentServiceServer) DeleteAgent(ctx context.Context, req *agentsv1.DeleteAgentRequest) (*agentsv1.DeleteAgentResponse, error) {
	prev, err := s.repo.GetAgent(ctx, req.GetName())
	if err != nil {
		return nil, toTwirpError(err)
	}

	err = deleteWithRuntime(
		func() error {
			return s.repo.DeleteAgent(ctx, req.GetName())
		},
		func() error {
			return s.reloadRuntime(ctx)
		},
		func() error {
			if _, err := s.repo.CreateAgent(ctx, proto.Clone(prev).(*agentsv1.Agent)); err != nil {
				return err
			}
			return s.reloadRuntime(ctx)
		},
	)
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.DeleteAgentResponse{}, nil
}

func (s *AgentServiceServer) InvokeAgent(ctx context.Context, req *agentsv1.InvokeAgentRequest) (*agentsv1.InvokeAgentResponse, error) {
	if s.runnerSvc == nil {
		return nil, twirp.NewError(twirp.FailedPrecondition, "runner service not available")
	}
	if req.GetAgentName() == "" {
		return nil, twirp.RequiredArgumentError("agent_name")
	}
	if req.GetInput() == "" {
		return nil, twirp.RequiredArgumentError("input")
	}

	appName := req.GetAppName()
	if appName == "" {
		appName = "api"
	}
	userID := req.GetUserId()
	if userID == "" {
		userID = "api"
	}
	sessionID := req.GetSessionId()
	if sessionID == "" {
		sessionID = "invoke-" + uuid.NewString()
	}

	ctxInfo := &agentsv1.ContextInfo{
		Uuid:        uuid.NewString(),
		SessionId:   sessionID,
		UserId:      userID,
		ChannelName: appName,
		Source:      agentsv1.ContextSource_CONTEXT_SOURCE_API,
	}

	parts := []*genai.Part{{Text: req.GetInput()}}
	response, err := s.runnerSvc.Run(ctx, req.GetAgentName(), parts, req.GetModelOverride(), ctxInfo, nil, nil)
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}
	return &agentsv1.InvokeAgentResponse{SessionId: sessionID, Response: response}, nil
}

func (s *AgentServiceServer) ListAgentInvocations(ctx context.Context, req *agentsv1.ListAgentInvocationsRequest) (*agentsv1.ListAgentInvocationsResponse, error) {
	if s.invRepo == nil {
		return &agentsv1.ListAgentInvocationsResponse{}, nil
	}
	invs, next, total, err := s.invRepo.List(ctx, invocation.ListFilter{
		AgentName: req.GetAgentName(),
		SessionID: req.GetSessionId(),
	}, req.GetPageSize(), req.GetPageToken())
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}
	return &agentsv1.ListAgentInvocationsResponse{
		Invocations:   invs,
		NextPageToken: next,
		Total:         total,
	}, nil
}

func (s *AgentServiceServer) reloadRuntime(ctx context.Context) error {
	if s.runtime == nil {
		return nil
	}
	if err := s.runtime.ReloadRunner(ctx); err != nil {
		return toTwirpError(err)
	}
	return nil
}

type ConfigRuntime interface {
	ReloadRunner(ctx context.Context) error
	ReloadChannels(ctx context.Context) error
}

func toTwirpError(err error) twirp.Error {
	if twerr, ok := err.(twirp.Error); ok {
		return twerr
	}
	if errors.Is(err, configrepo.ErrNotFound) {
		return twirp.NotFoundError(err.Error())
	}
	if errors.Is(err, configrepo.ErrAlreadyExists) {
		return twirp.NewError(twirp.AlreadyExists, err.Error())
	}
	return twirp.InternalErrorWith(err)
}
