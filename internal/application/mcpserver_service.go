package application

import (
	"context"
	"time"

	"butterfly.orx.me/core/log"

	internalagent "go.orx.me/apps/butter/internal/agent"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MCPServerServiceServer struct {
	repo    configrepo.MCPServerRepository
	runtime ConfigRuntime
}

func NewMCPServerServiceServer(repo configrepo.MCPServerRepository) *MCPServerServiceServer {
	return &MCPServerServiceServer{repo: repo}
}

func (s *MCPServerServiceServer) SetRuntime(runtime ConfigRuntime) {
	s.runtime = runtime
}

func (s *MCPServerServiceServer) ListMCPServers(ctx context.Context, _ *agentsv1.ListMCPServersRequest) (*agentsv1.ListMCPServersResponse, error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	servers, err := s.repo.ListMCPServers(ctx, wsID)
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.ListMCPServersResponse{McpServers: servers}, nil
}

func (s *MCPServerServiceServer) GetMCPServer(ctx context.Context, req *agentsv1.GetMCPServerRequest) (*agentsv1.GetMCPServerResponse, error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	m, err := s.repo.GetMCPServer(ctx, wsID, req.GetId())
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.GetMCPServerResponse{McpServer: m}, nil
}

func (s *MCPServerServiceServer) CreateMCPServer(ctx context.Context, req *agentsv1.CreateMCPServerRequest) (*agentsv1.CreateMCPServerResponse, error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	logger := log.FromContext(ctx)
	logger.Info("creating mcp server",
		"workspace_id", wsID,
		"name", req.GetMcpServer().GetName(),
		"transport", req.GetMcpServer().GetTransport().String(),
	)
	m, err := mutateWithRuntime(
		func() (*agentsv1.MCPServer, error) {
			return s.repo.CreateMCPServer(ctx, wsID, req.GetMcpServer())
		},
		func() error {
			return s.reloadRuntime(ctx)
		},
		func() error {
			if err := s.repo.DeleteMCPServer(ctx, wsID, req.GetMcpServer().GetId()); err != nil {
				return err
			}
			return s.reloadRuntime(ctx)
		},
	)
	if err != nil {
		logger.Error("create mcp server failed", "workspace_id", wsID, "name", req.GetMcpServer().GetName(), "err", err)
		return nil, toTwirpError(err)
	}
	logger.Info("mcp server created", "workspace_id", wsID, "id", m.GetId(), "name", m.GetName())
	return &agentsv1.CreateMCPServerResponse{McpServer: m}, nil
}

func (s *MCPServerServiceServer) UpdateMCPServer(ctx context.Context, req *agentsv1.UpdateMCPServerRequest) (*agentsv1.UpdateMCPServerResponse, error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	logger := log.FromContext(ctx)
	prev, err := s.repo.GetMCPServer(ctx, wsID, req.GetMcpServer().GetId())
	if err != nil {
		return nil, toTwirpError(err)
	}
	logger.Info("updating mcp server", "workspace_id", wsID, "id", req.GetMcpServer().GetId(), "name", req.GetMcpServer().GetName())

	m, err := mutateWithRuntime(
		func() (*agentsv1.MCPServer, error) {
			return s.repo.UpdateMCPServer(ctx, wsID, req.GetMcpServer())
		},
		func() error {
			return s.reloadRuntime(ctx)
		},
		func() error {
			if _, err := s.repo.UpdateMCPServer(ctx, wsID, proto.Clone(prev).(*agentsv1.MCPServer)); err != nil {
				return err
			}
			return s.reloadRuntime(ctx)
		},
	)
	if err != nil {
		logger.Error("update mcp server failed", "workspace_id", wsID, "id", req.GetMcpServer().GetId(), "err", err)
		return nil, toTwirpError(err)
	}
	logger.Info("mcp server updated", "workspace_id", wsID, "id", m.GetId(), "name", m.GetName())
	return &agentsv1.UpdateMCPServerResponse{McpServer: m}, nil
}

func (s *MCPServerServiceServer) DeleteMCPServer(ctx context.Context, req *agentsv1.DeleteMCPServerRequest) (*agentsv1.DeleteMCPServerResponse, error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	logger := log.FromContext(ctx)
	prev, err := s.repo.GetMCPServer(ctx, wsID, req.GetId())
	if err != nil {
		return nil, toTwirpError(err)
	}
	logger.Info("deleting mcp server", "workspace_id", wsID, "id", req.GetId(), "name", prev.GetName())

	err = deleteWithRuntime(
		func() error {
			return s.repo.DeleteMCPServer(ctx, wsID, req.GetId())
		},
		func() error {
			return s.reloadRuntime(ctx)
		},
		func() error {
			if _, err := s.repo.CreateMCPServer(ctx, wsID, proto.Clone(prev).(*agentsv1.MCPServer)); err != nil {
				return err
			}
			return s.reloadRuntime(ctx)
		},
	)
	if err != nil {
		logger.Error("delete mcp server failed", "workspace_id", wsID, "id", req.GetId(), "err", err)
		return nil, toTwirpError(err)
	}
	logger.Info("mcp server deleted", "workspace_id", wsID, "id", req.GetId(), "name", prev.GetName())
	return &agentsv1.DeleteMCPServerResponse{}, nil
}

func (s *MCPServerServiceServer) GetMCPServerStatus(ctx context.Context, req *agentsv1.GetMCPServerStatusRequest) (*agentsv1.GetMCPServerStatusResponse, error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	m, err := s.repo.GetMCPServer(ctx, wsID, req.GetId())
	if err != nil {
		return nil, toTwirpError(err)
	}

	status := &agentsv1.MCPServerStatus{
		Id:        m.GetId(),
		Name:      m.GetName(),
		CheckedAt: timestamppb.New(time.Now().UTC()),
	}

	if m.GetTransport() == agentsv1.MCPServerTransport_MCP_SERVER_TRANSPORT_STDIO {
		status.State = agentsv1.MCPServerStatus_STATE_CONFIGURED
		status.ToolCount = int32(len(m.GetToolFilter()))
		status.Detail = "stdio probing not supported"
		return &agentsv1.GetMCPServerStatusResponse{Status: status}, nil
	}

	timeout, err := internalagent.MCPTimeout(m)
	if err != nil {
		status.State = agentsv1.MCPServerStatus_STATE_DISCONNECTED
		status.Detail = err.Error()
		status.ToolCount = int32(len(m.GetToolFilter()))
		return &agentsv1.GetMCPServerStatusResponse{Status: status}, nil
	}
	probeCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	result, err := internalagent.ProbeMCPServer(probeCtx, m)
	if err != nil {
		log.FromContext(ctx).Warn("mcp server probe failed",
			"workspace_id", wsID, "id", m.GetId(), "name", m.GetName(), "err", err)
		status.State = agentsv1.MCPServerStatus_STATE_DISCONNECTED
		status.Detail = err.Error()
		status.ToolCount = int32(len(m.GetToolFilter()))
		return &agentsv1.GetMCPServerStatusResponse{Status: status}, nil
	}
	status.State = agentsv1.MCPServerStatus_STATE_CONNECTED
	status.ToolCount = int32(result.ToolCount)
	return &agentsv1.GetMCPServerStatusResponse{Status: status}, nil
}

func (s *MCPServerServiceServer) ListMCPTools(ctx context.Context, req *agentsv1.ListMCPToolsRequest) (*agentsv1.ListMCPToolsResponse, error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	var servers []*agentsv1.MCPServer
	if id := req.GetServerId(); id != "" {
		srv, err := s.repo.GetMCPServer(ctx, wsID, id)
		if err != nil {
			return nil, toTwirpError(err)
		}
		servers = []*agentsv1.MCPServer{srv}
	} else {
		all, err := s.repo.ListMCPServers(ctx, wsID)
		if err != nil {
			return nil, toTwirpError(err)
		}
		servers = all
	}

	resp := &agentsv1.ListMCPToolsResponse{Errors: map[string]string{}}
	for _, srv := range servers {
		if srv.GetTransport() == agentsv1.MCPServerTransport_MCP_SERVER_TRANSPORT_STDIO {
			resp.Errors[srv.GetId()] = "stdio probing not supported"
			continue
		}
		timeout, err := internalagent.MCPTimeout(srv)
		if err != nil {
			resp.Errors[srv.GetId()] = err.Error()
			continue
		}
		probeCtx, cancel := context.WithTimeout(ctx, timeout)
		result, err := internalagent.ProbeMCPServer(probeCtx, srv)
		cancel()
		if err != nil {
			resp.Errors[srv.GetId()] = err.Error()
			continue
		}
		for _, t := range result.Tools {
			resp.Tools = append(resp.Tools, &agentsv1.MCPTool{
				Name:        t.Name,
				Description: t.Description,
				ServerId:    srv.GetId(),
				ServerName:  srv.GetName(),
				Allowed:     t.Allowed,
			})
		}
	}
	return resp, nil
}

func (s *MCPServerServiceServer) reloadRuntime(ctx context.Context) error {
	if s.runtime == nil {
		return nil
	}
	if err := s.runtime.ReloadRunner(ctx); err != nil {
		return toTwirpError(err)
	}
	return nil
}
