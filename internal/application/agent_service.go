package application

import (
	"context"
	"errors"

	"github.com/twitchtv/twirp"

	configrepo "go.orx.me/apps/butter/internal/repo/config"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type AgentServiceServer struct {
	repo    configrepo.AgentRepository
	runtime ConfigRuntime
}

func NewAgentServiceServer(repo configrepo.AgentRepository) *AgentServiceServer {
	return &AgentServiceServer{repo: repo}
}

func (s *AgentServiceServer) SetRuntime(runtime ConfigRuntime) {
	s.runtime = runtime
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
	a, err := s.repo.CreateAgent(ctx, req.GetAgent())
	if err != nil {
		return nil, toTwirpError(err)
	}
	if err := s.reloadRuntime(ctx); err != nil {
		return nil, err
	}
	return &agentsv1.CreateAgentResponse{Agent: a}, nil
}

func (s *AgentServiceServer) UpdateAgent(ctx context.Context, req *agentsv1.UpdateAgentRequest) (*agentsv1.UpdateAgentResponse, error) {
	a, err := s.repo.UpdateAgent(ctx, req.GetAgent())
	if err != nil {
		return nil, toTwirpError(err)
	}
	if err := s.reloadRuntime(ctx); err != nil {
		return nil, err
	}
	return &agentsv1.UpdateAgentResponse{Agent: a}, nil
}

func (s *AgentServiceServer) DeleteAgent(ctx context.Context, req *agentsv1.DeleteAgentRequest) (*agentsv1.DeleteAgentResponse, error) {
	if err := s.repo.DeleteAgent(ctx, req.GetName()); err != nil {
		return nil, toTwirpError(err)
	}
	if err := s.reloadRuntime(ctx); err != nil {
		return nil, err
	}
	return &agentsv1.DeleteAgentResponse{}, nil
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
	if errors.Is(err, configrepo.ErrNotFound) {
		return twirp.NotFoundError(err.Error())
	}
	if errors.Is(err, configrepo.ErrAlreadyExists) {
		return twirp.NewError(twirp.AlreadyExists, err.Error())
	}
	return twirp.InternalErrorWith(err)
}
