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

func (s *GlobalMCPServerServiceServer) ListGlobalMCPServers(ctx context.Context, _ *connect.Request[agentsv1.ListGlobalMCPServersRequest]) (*connect.Response[agentsv1.ListGlobalMCPServersResponse], error) {
	if s.repo == nil {
		return connect.NewResponse(&agentsv1.ListGlobalMCPServersResponse{}), nil
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
	return connect.NewResponse(&agentsv1.ListGlobalMCPServersResponse{McpServers: out}), nil
}

func (s *GlobalMCPServerServiceServer) CreateGlobalMCPServer(ctx context.Context, req *connect.Request[agentsv1.CreateGlobalMCPServerRequest]) (*connect.Response[agentsv1.CreateGlobalMCPServerResponse], error) {
	if s.repo == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("global mcp store not available"))
	}
	if !auth.IsAdmin(ctx) {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("admin role required"))
	}
	server := req.Msg.GetMcpServer()
	if server == nil {
		return nil, connectx.RequiredArgument("mcp_server")
	}
	server.WorkspaceId = ""
	if err := validateMCPServerConfig(server); err != nil {
		return nil, err
	}
	created, err := s.repo.CreateGlobalMCPServer(ctx, server)
	if err != nil {
		return nil, connectx.InternalWith(err)
	}
	return connect.NewResponse(&agentsv1.CreateGlobalMCPServerResponse{McpServer: mcpServerForResponseClone(created, false)}), nil
}

func (s *GlobalMCPServerServiceServer) UpdateGlobalMCPServer(ctx context.Context, req *connect.Request[agentsv1.UpdateGlobalMCPServerRequest]) (*connect.Response[agentsv1.UpdateGlobalMCPServerResponse], error) {
	if s.repo == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("global mcp store not available"))
	}
	if !auth.IsAdmin(ctx) {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("admin role required"))
	}
	server := req.Msg.GetMcpServer()
	if server == nil {
		return nil, connectx.RequiredArgument("mcp_server")
	}
	if strings.TrimSpace(server.GetId()) == "" {
		return nil, connectx.RequiredArgument("mcp_server.id")
	}
	server.WorkspaceId = ""
	if err := validateMCPServerConfig(server); err != nil {
		return nil, err
	}
	updated, err := s.repo.UpdateGlobalMCPServer(ctx, server)
	if err != nil {
		return nil, connectx.InternalWith(err)
	}
	return connect.NewResponse(&agentsv1.UpdateGlobalMCPServerResponse{McpServer: mcpServerForResponseClone(updated, false)}), nil
}

func (s *GlobalMCPServerServiceServer) DeleteGlobalMCPServer(ctx context.Context, req *connect.Request[agentsv1.DeleteGlobalMCPServerRequest]) (*connect.Response[agentsv1.DeleteGlobalMCPServerResponse], error) {
	if s.repo == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("global mcp store not available"))
	}
	if !auth.IsAdmin(ctx) {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("admin role required"))
	}
	id := strings.TrimSpace(req.Msg.GetId())
	if id == "" {
		return nil, connectx.RequiredArgument("id")
	}
	if err := s.repo.DeleteGlobalMCPServer(ctx, id); err != nil {
		return nil, connectx.InternalWith(err)
	}
	return connect.NewResponse(&agentsv1.DeleteGlobalMCPServerResponse{}), nil
}

func (s *GlobalMCPServerServiceServer) InstallGlobalMCPServer(ctx context.Context, req *connect.Request[agentsv1.InstallGlobalMCPServerRequest]) (*connect.Response[agentsv1.InstallGlobalMCPServerResponse], error) {
	if s.repo == nil || s.mcpSvc == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("global mcp store not available"))
	}
	presetID := strings.TrimSpace(req.Msg.GetId())
	if presetID == "" {
		return nil, connectx.RequiredArgument("id")
	}

	contextWorkspaceID, _ := wsctx.FromContext(ctx)
	targetWorkspaceID := contextWorkspaceID
	requested := strings.TrimSpace(req.Msg.GetWorkspaceId())
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
	existing, err := s.mcpSvc.repo.GetMCPServer(ctx, targetWorkspaceID, server.GetId())
	if err == nil {
		if existing.GetMetadata()[globalMCPPresetMetadataKey] != preset.GetId() {
			return nil, connect.NewError(connect.CodeAlreadyExists, errors.New("mcp server id is already used by a workspace server not installed from this global preset"))
		}
		updated, err := s.mcpSvc.UpdateMCPServer(installCtx, connect.NewRequest(&agentsv1.UpdateMCPServerRequest{McpServer: server}))
		if err != nil {
			return nil, err
		}
		return connect.NewResponse(&agentsv1.InstallGlobalMCPServerResponse{McpServer: mcpServerForResponseClone(updated.Msg.GetMcpServer(), true)}), nil
	}
	if !errors.Is(err, configrepo.ErrNotFound) {
		return nil, toConnectError(err)
	}

	created, err := s.mcpSvc.CreateMCPServer(installCtx, connect.NewRequest(&agentsv1.CreateMCPServerRequest{McpServer: server}))
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&agentsv1.InstallGlobalMCPServerResponse{McpServer: mcpServerForResponseClone(created.Msg.GetMcpServer(), true)}), nil
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
