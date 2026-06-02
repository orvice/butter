package application

import (
	"context"
	"errors"
	"strings"

	"butterfly.orx.me/core/log"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	"go.orx.me/apps/butter/internal/repo/auth"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	"go.orx.me/apps/butter/internal/transport/connectx"
	wsctx "go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// GlobalMCPServerServiceServer serves the workspace-agnostic MCP preset
// surface: list/create/update/delete plus install-into-workspace. Mutations
// (create/update/delete) require the caller to be an admin; list and install
// are open to any authenticated user.
//
// The install path delegates to the workspace MCP service so the same
// validation, cloning, and metadata tagging logic stays in one place.
type GlobalMCPServerServiceServer struct {
	repo   configrepo.GlobalMCPServerRepository
	mcpSvc *MCPServerServiceServer
}

func NewGlobalMCPServerServiceServer(repo configrepo.GlobalMCPServerRepository, mcpSvc *MCPServerServiceServer) *GlobalMCPServerServiceServer {
	return &GlobalMCPServerServiceServer{repo: repo, mcpSvc: mcpSvc}
}

func (s *GlobalMCPServerServiceServer) ListGlobalMCPServers(ctx context.Context, _ *agentsv1.ListGlobalMCPServersRequest) (*agentsv1.ListGlobalMCPServersResponse, error) {
	if s.repo == nil {
		return &agentsv1.ListGlobalMCPServersResponse{}, nil
	}
	servers, err := s.repo.ListGlobalMCPServers(ctx)
	if err != nil {
		return nil, connectx.InternalWith(err)
	}
	redact := !auth.IsAdmin(ctx)
	out := make([]*agentsv1.MCPServer, 0, len(servers))
	for _, srv := range servers {
		out = append(out, mcpServerForResponseClone(srv, redact))
	}
	return &agentsv1.ListGlobalMCPServersResponse{McpServers: out}, nil
}

func (s *GlobalMCPServerServiceServer) CreateGlobalMCPServer(ctx context.Context, req *agentsv1.CreateGlobalMCPServerRequest) (*agentsv1.CreateGlobalMCPServerResponse, error) {
	if s.repo == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("global mcp store not available"))
	}
	if !auth.IsAdmin(ctx) {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("admin role required"))
	}
	server := req.GetMcpServer()
	if server == nil {
		return nil, connectx.RequiredArgument("mcp_server")
	}
	server.WorkspaceId = ""
	created, err := s.repo.CreateGlobalMCPServer(ctx, server)
	if err != nil {
		return nil, connectx.InternalWith(err)
	}
	return &agentsv1.CreateGlobalMCPServerResponse{McpServer: mcpServerForResponseClone(created, false)}, nil
}

func (s *GlobalMCPServerServiceServer) UpdateGlobalMCPServer(ctx context.Context, req *agentsv1.UpdateGlobalMCPServerRequest) (*agentsv1.UpdateGlobalMCPServerResponse, error) {
	if s.repo == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("global mcp store not available"))
	}
	if !auth.IsAdmin(ctx) {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("admin role required"))
	}
	server := req.GetMcpServer()
	if server == nil {
		return nil, connectx.RequiredArgument("mcp_server")
	}
	if strings.TrimSpace(server.GetId()) == "" {
		return nil, connectx.RequiredArgument("mcp_server.id")
	}
	server.WorkspaceId = ""
	updated, err := s.repo.UpdateGlobalMCPServer(ctx, server)
	if err != nil {
		return nil, connectx.InternalWith(err)
	}
	return &agentsv1.UpdateGlobalMCPServerResponse{McpServer: mcpServerForResponseClone(updated, false)}, nil
}

func (s *GlobalMCPServerServiceServer) DeleteGlobalMCPServer(ctx context.Context, req *agentsv1.DeleteGlobalMCPServerRequest) (*agentsv1.DeleteGlobalMCPServerResponse, error) {
	if s.repo == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("global mcp store not available"))
	}
	if !auth.IsAdmin(ctx) {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("admin role required"))
	}
	id := strings.TrimSpace(req.GetId())
	if id == "" {
		return nil, connectx.RequiredArgument("id")
	}
	if err := s.repo.DeleteGlobalMCPServer(ctx, id); err != nil {
		return nil, connectx.InternalWith(err)
	}
	return &agentsv1.DeleteGlobalMCPServerResponse{}, nil
}

func (s *GlobalMCPServerServiceServer) InstallGlobalMCPServer(ctx context.Context, req *agentsv1.InstallGlobalMCPServerRequest) (*agentsv1.InstallGlobalMCPServerResponse, error) {
	if s.repo == nil || s.mcpSvc == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("global mcp store not available"))
	}
	presetID := strings.TrimSpace(req.GetId())
	if presetID == "" {
		return nil, connectx.RequiredArgument("id")
	}

	contextWorkspaceID, _ := wsctx.FromContext(ctx)
	targetWorkspaceID := contextWorkspaceID
	requested := strings.TrimSpace(req.GetWorkspaceId())
	if requested != "" {
		if !auth.IsAdmin(ctx) && requested != contextWorkspaceID {
			return nil, connect.NewError(connect.CodePermissionDenied, errors.New("admin role required for cross-workspace install"))
		}
		if requested != contextWorkspaceID {
			auditCrossWorkspaceInstall(ctx, contextWorkspaceID, requested, presetID)
		}
		targetWorkspaceID = requested
	}
	if targetWorkspaceID == "" && !auth.IsAdmin(ctx) {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("workspace required"))
	}

	preset, err := s.repo.GetGlobalMCPServer(ctx, presetID)
	if err != nil {
		return nil, connectx.InternalWith(err)
	}
	server := proto.Clone(preset).(*agentsv1.MCPServer)
	server.WorkspaceId = ""
	MarkInstalledGlobalMCPPreset(server, preset.GetId())

	installCtx := wsctx.WithID(ctx, targetWorkspaceID)
	created, err := s.mcpSvc.CreateMCPServer(installCtx, &agentsv1.CreateMCPServerRequest{McpServer: server})
	if err != nil {
		return nil, err
	}
	return &agentsv1.InstallGlobalMCPServerResponse{McpServer: mcpServerForResponseClone(created.GetMcpServer(), true)}, nil
}

// mcpServerForResponseClone clones the server and optionally clears the
// OAuth2 client_secret. Mirrors the helper that the REST handler used; kept
// package-local here so callers don't reach into internal/app.
func mcpServerForResponseClone(server *agentsv1.MCPServer, redact bool) *agentsv1.MCPServer {
	if server == nil {
		return nil
	}
	clone := proto.Clone(server).(*agentsv1.MCPServer)
	if redact {
		if oauth := clone.GetAuth().GetOauth2(); oauth != nil {
			oauth.ClientSecret = ""
		}
	}
	return clone
}

// auditCrossWorkspaceInstall logs an admin install that targets a workspace
// the caller did not enter via X-Workspace-ID. Cross-workspace installs are
// intentionally allowed for ops/automation but need a paper trail because a
// compromised admin can use this path to plant SSRF-capable MCP servers in
// any tenant.
func auditCrossWorkspaceInstall(ctx context.Context, contextWorkspaceID, targetWorkspaceID, presetID string) {
	logger := log.FromContext(ctx).With("audit", "admin_cross_workspace_install")
	fields := []any{
		"preset_id", presetID,
		"context_workspace_id", contextWorkspaceID,
		"target_workspace_id", targetWorkspaceID,
	}
	if user, ok := auth.UserFromContext(ctx); ok {
		fields = append(fields, "user_id", user.GetId(), "user_role", user.GetRole())
	}
	logger.Warn("admin installed global MCP preset into another workspace", fields...)
}
