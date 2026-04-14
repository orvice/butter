package application

import (
	"context"

	configrepo "go.orx.me/apps/butter/internal/repo/config"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type RemoteAgentServiceServer struct {
	repo configrepo.RemoteAgentRepository
}

func NewRemoteAgentServiceServer(repo configrepo.RemoteAgentRepository) *RemoteAgentServiceServer {
	return &RemoteAgentServiceServer{repo: repo}
}

func (s *RemoteAgentServiceServer) ListRemoteAgents(ctx context.Context, _ *agentsv1.ListRemoteAgentsRequest) (*agentsv1.ListRemoteAgentsResponse, error) {
	agents, err := s.repo.ListRemoteAgents(ctx)
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.ListRemoteAgentsResponse{RemoteAgents: agents}, nil
}

func (s *RemoteAgentServiceServer) GetRemoteAgent(ctx context.Context, req *agentsv1.GetRemoteAgentRequest) (*agentsv1.GetRemoteAgentResponse, error) {
	r, err := s.repo.GetRemoteAgent(ctx, req.GetId())
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.GetRemoteAgentResponse{RemoteAgent: r}, nil
}

func (s *RemoteAgentServiceServer) CreateRemoteAgent(ctx context.Context, req *agentsv1.CreateRemoteAgentRequest) (*agentsv1.CreateRemoteAgentResponse, error) {
	r, err := s.repo.CreateRemoteAgent(ctx, req.GetRemoteAgent())
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.CreateRemoteAgentResponse{RemoteAgent: r}, nil
}

func (s *RemoteAgentServiceServer) UpdateRemoteAgent(ctx context.Context, req *agentsv1.UpdateRemoteAgentRequest) (*agentsv1.UpdateRemoteAgentResponse, error) {
	r, err := s.repo.UpdateRemoteAgent(ctx, req.GetRemoteAgent())
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.UpdateRemoteAgentResponse{RemoteAgent: r}, nil
}

func (s *RemoteAgentServiceServer) DeleteRemoteAgent(ctx context.Context, req *agentsv1.DeleteRemoteAgentRequest) (*agentsv1.DeleteRemoteAgentResponse, error) {
	if err := s.repo.DeleteRemoteAgent(ctx, req.GetId()); err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.DeleteRemoteAgentResponse{}, nil
}
