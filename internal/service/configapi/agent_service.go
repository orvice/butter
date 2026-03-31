package configapi

import (
	"context"
	"strings"

	"github.com/twitchtv/twirp"
	"google.golang.org/protobuf/types/known/emptypb"

	"go.orx.me/apps/butter/internal/repo/configstore"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type AgentServiceServer struct {
	store *configstore.Store
}

func NewAgentServiceServer(store *configstore.Store) *AgentServiceServer {
	return &AgentServiceServer{store: store}
}

func (s *AgentServiceServer) ListAgents(_ context.Context, _ *agentsv1.ListAgentsRequest) (*agentsv1.ListAgentsResponse, error) {
	return &agentsv1.ListAgentsResponse{Agents: s.store.ListAgents()}, nil
}

func (s *AgentServiceServer) GetAgent(_ context.Context, req *agentsv1.GetAgentRequest) (*agentsv1.Agent, error) {
	a, err := s.store.GetAgent(req.GetName())
	if err != nil {
		return nil, toTwirpError(err)
	}
	return a, nil
}

func (s *AgentServiceServer) CreateAgent(_ context.Context, req *agentsv1.CreateAgentRequest) (*agentsv1.Agent, error) {
	a, err := s.store.CreateAgent(req.GetAgent())
	if err != nil {
		return nil, toTwirpError(err)
	}
	return a, nil
}

func (s *AgentServiceServer) UpdateAgent(_ context.Context, req *agentsv1.UpdateAgentRequest) (*agentsv1.Agent, error) {
	a, err := s.store.UpdateAgent(req.GetAgent())
	if err != nil {
		return nil, toTwirpError(err)
	}
	return a, nil
}

func (s *AgentServiceServer) DeleteAgent(_ context.Context, req *agentsv1.DeleteAgentRequest) (*emptypb.Empty, error) {
	if err := s.store.DeleteAgent(req.GetName()); err != nil {
		return nil, toTwirpError(err)
	}
	return &emptypb.Empty{}, nil
}

func toTwirpError(err error) twirp.Error {
	msg := err.Error()
	if strings.Contains(msg, "not found") {
		return twirp.NotFoundError(msg)
	}
	if strings.Contains(msg, "already exists") {
		return twirp.NewError(twirp.AlreadyExists, msg)
	}
	return twirp.InternalErrorWith(err)
}
