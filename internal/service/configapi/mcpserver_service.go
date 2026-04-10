package configapi

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	"go.orx.me/apps/butter/internal/repo/configstore"
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

func (s *MCPServerServiceServer) GetMCPServer(_ context.Context, req *agentsv1.GetMCPServerRequest) (*agentsv1.MCPServer, error) {
	m, err := s.store.GetMCPServer(req.GetId())
	if err != nil {
		return nil, toTwirpError(err)
	}
	return m, nil
}

func (s *MCPServerServiceServer) CreateMCPServer(_ context.Context, req *agentsv1.CreateMCPServerRequest) (*agentsv1.MCPServer, error) {
	m, err := s.store.CreateMCPServer(req.GetMcpServer())
	if err != nil {
		return nil, toTwirpError(err)
	}
	return m, nil
}

func (s *MCPServerServiceServer) UpdateMCPServer(_ context.Context, req *agentsv1.UpdateMCPServerRequest) (*agentsv1.MCPServer, error) {
	m, err := s.store.UpdateMCPServer(req.GetMcpServer())
	if err != nil {
		return nil, toTwirpError(err)
	}
	return m, nil
}

func (s *MCPServerServiceServer) DeleteMCPServer(_ context.Context, req *agentsv1.DeleteMCPServerRequest) (*emptypb.Empty, error) {
	if err := s.store.DeleteMCPServer(req.GetId()); err != nil {
		return nil, toTwirpError(err)
	}
	return &emptypb.Empty{}, nil
}
