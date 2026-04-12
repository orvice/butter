package application

import (
	"context"

	"go.orx.me/apps/butter/internal/store/config"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type MCPServerServiceServer struct {
	store *configstore.Store
}

func NewMCPServerServiceServer(store *configstore.Store) *MCPServerServiceServer {
	return &MCPServerServiceServer{store: store}
}

func (s *MCPServerServiceServer) ListMCPServers(_ context.Context, _ *agentsv1.ListMCPServersRequest) (*agentsv1.ListMCPServersResponse, error) {
	return &agentsv1.ListMCPServersResponse{McpServers: s.store.ListMCPServers()}, nil
}

func (s *MCPServerServiceServer) GetMCPServer(_ context.Context, req *agentsv1.GetMCPServerRequest) (*agentsv1.GetMCPServerResponse, error) {
	m, err := s.store.GetMCPServer(req.GetId())
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.GetMCPServerResponse{McpServer: m}, nil
}

func (s *MCPServerServiceServer) CreateMCPServer(_ context.Context, req *agentsv1.CreateMCPServerRequest) (*agentsv1.CreateMCPServerResponse, error) {
	m, err := s.store.CreateMCPServer(req.GetMcpServer())
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.CreateMCPServerResponse{McpServer: m}, nil
}

func (s *MCPServerServiceServer) UpdateMCPServer(_ context.Context, req *agentsv1.UpdateMCPServerRequest) (*agentsv1.UpdateMCPServerResponse, error) {
	m, err := s.store.UpdateMCPServer(req.GetMcpServer())
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.UpdateMCPServerResponse{McpServer: m}, nil
}

func (s *MCPServerServiceServer) DeleteMCPServer(_ context.Context, req *agentsv1.DeleteMCPServerRequest) (*agentsv1.DeleteMCPServerResponse, error) {
	if err := s.store.DeleteMCPServer(req.GetId()); err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.DeleteMCPServerResponse{}, nil
}
