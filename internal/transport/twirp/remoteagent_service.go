package twirp

import (
	"context"

	"go.orx.me/apps/butter/internal/store/config"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type RemoteAgentServiceServer struct {
	store *configstore.Store
}

func NewRemoteAgentServiceServer(store *configstore.Store) *RemoteAgentServiceServer {
	return &RemoteAgentServiceServer{store: store}
}

func (s *RemoteAgentServiceServer) ListRemoteAgents(_ context.Context, _ *agentsv1.ListRemoteAgentsRequest) (*agentsv1.ListRemoteAgentsResponse, error) {
	return &agentsv1.ListRemoteAgentsResponse{RemoteAgents: s.store.ListRemoteAgents()}, nil
}

func (s *RemoteAgentServiceServer) GetRemoteAgent(_ context.Context, req *agentsv1.GetRemoteAgentRequest) (*agentsv1.GetRemoteAgentResponse, error) {
	r, err := s.store.GetRemoteAgent(req.GetId())
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.GetRemoteAgentResponse{RemoteAgent: r}, nil
}

func (s *RemoteAgentServiceServer) CreateRemoteAgent(_ context.Context, req *agentsv1.CreateRemoteAgentRequest) (*agentsv1.CreateRemoteAgentResponse, error) {
	r, err := s.store.CreateRemoteAgent(req.GetRemoteAgent())
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.CreateRemoteAgentResponse{RemoteAgent: r}, nil
}

func (s *RemoteAgentServiceServer) UpdateRemoteAgent(_ context.Context, req *agentsv1.UpdateRemoteAgentRequest) (*agentsv1.UpdateRemoteAgentResponse, error) {
	r, err := s.store.UpdateRemoteAgent(req.GetRemoteAgent())
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.UpdateRemoteAgentResponse{RemoteAgent: r}, nil
}

func (s *RemoteAgentServiceServer) DeleteRemoteAgent(_ context.Context, req *agentsv1.DeleteRemoteAgentRequest) (*agentsv1.DeleteRemoteAgentResponse, error) {
	if err := s.store.DeleteRemoteAgent(req.GetId()); err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.DeleteRemoteAgentResponse{}, nil
}
