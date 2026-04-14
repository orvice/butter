package application

import (
	"context"

	configrepo "go.orx.me/apps/butter/internal/repo/config"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type MCPServerServiceServer struct {
	repo configrepo.MCPServerRepository
}

func NewMCPServerServiceServer(repo configrepo.MCPServerRepository) *MCPServerServiceServer {
	return &MCPServerServiceServer{repo: repo}
}

func (s *MCPServerServiceServer) ListMCPServers(ctx context.Context, _ *agentsv1.ListMCPServersRequest) (*agentsv1.ListMCPServersResponse, error) {
	servers, err := s.repo.ListMCPServers(ctx)
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.ListMCPServersResponse{McpServers: servers}, nil
}

func (s *MCPServerServiceServer) GetMCPServer(ctx context.Context, req *agentsv1.GetMCPServerRequest) (*agentsv1.GetMCPServerResponse, error) {
	m, err := s.repo.GetMCPServer(ctx, req.GetId())
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.GetMCPServerResponse{McpServer: m}, nil
}

func (s *MCPServerServiceServer) CreateMCPServer(ctx context.Context, req *agentsv1.CreateMCPServerRequest) (*agentsv1.CreateMCPServerResponse, error) {
	m, err := s.repo.CreateMCPServer(ctx, req.GetMcpServer())
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.CreateMCPServerResponse{McpServer: m}, nil
}

func (s *MCPServerServiceServer) UpdateMCPServer(ctx context.Context, req *agentsv1.UpdateMCPServerRequest) (*agentsv1.UpdateMCPServerResponse, error) {
	m, err := s.repo.UpdateMCPServer(ctx, req.GetMcpServer())
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.UpdateMCPServerResponse{McpServer: m}, nil
}

func (s *MCPServerServiceServer) DeleteMCPServer(ctx context.Context, req *agentsv1.DeleteMCPServerRequest) (*agentsv1.DeleteMCPServerResponse, error) {
	if err := s.repo.DeleteMCPServer(ctx, req.GetId()); err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.DeleteMCPServerResponse{}, nil
}
