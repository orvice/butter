package application

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"butterfly.orx.me/core/log"
	"github.com/twitchtv/twirp"

	internalagent "go.orx.me/apps/butter/internal/agent"
	"go.orx.me/apps/butter/internal/mcpoauth"
	"go.orx.me/apps/butter/internal/repo/auth"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	mcpoauthrepo "go.orx.me/apps/butter/internal/repo/mcpoauth"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MCPServerServiceServer struct {
	repo         configrepo.MCPServerRepository
	runtime      ConfigRuntime
	oauthService *mcpoauth.Service
	httpFactory  internalagent.MCPHTTPClientFactory
}

const globalMCPPresetMetadataKey = "butter.global_mcp_preset_id"

func NewMCPServerServiceServer(repo configrepo.MCPServerRepository) *MCPServerServiceServer {
	return &MCPServerServiceServer{repo: repo}
}

func (s *MCPServerServiceServer) SetRuntime(runtime ConfigRuntime) {
	s.runtime = runtime
}

func (s *MCPServerServiceServer) SetOAuthService(service *mcpoauth.Service) {
	s.oauthService = service
}

func (s *MCPServerServiceServer) SetMCPHTTPClientFactory(factory internalagent.MCPHTTPClientFactory) {
	s.httpFactory = factory
}

func validateMCPServerConfig(server *agentsv1.MCPServer) error {
	if server == nil {
		return twirp.RequiredArgumentError("mcp_server")
	}
	if !isRemoteMCPTransport(server.GetTransport()) {
		return twirp.InvalidArgumentError("mcp_server.transport", fmt.Sprintf("unsupported MCP transport %s", server.GetTransport()))
	}
	if strings.TrimSpace(server.GetUrl()) == "" {
		return twirp.RequiredArgumentError("mcp_server.url")
	}
	if err := validateHTTPURL("mcp_server.url", server.GetUrl()); err != nil {
		return err
	}
	authType := mcpoauth.AuthType(server)
	if authType != agentsv1.MCPServerAuthType_MCP_SERVER_AUTH_TYPE_OAUTH2 {
		return nil
	}
	oauth := server.GetAuth().GetOauth2()
	if oauth == nil {
		return twirp.RequiredArgumentError("mcp_server.auth.oauth2")
	}
	if strings.TrimSpace(oauth.GetClientSecret()) != "" && strings.TrimSpace(oauth.GetClientId()) == "" {
		return twirp.InvalidArgumentError("mcp_server.auth.oauth2.client_id", "client_id is required when client_secret is set")
	}
	for field, value := range map[string]string{
		"mcp_server.auth.oauth2.authorization_url":        oauth.GetAuthorizationUrl(),
		"mcp_server.auth.oauth2.token_url":                oauth.GetTokenUrl(),
		"mcp_server.auth.oauth2.resource_metadata_url":    oauth.GetResourceMetadataUrl(),
		"mcp_server.auth.oauth2.authorization_server_url": oauth.GetAuthorizationServerUrl(),
		"mcp_server.auth.oauth2.resource":                 oauth.GetResource(),
	} {
		if strings.TrimSpace(value) == "" {
			continue
		}
		if err := validateHTTPURL(field, value); err != nil {
			return err
		}
	}
	return nil
}

func isRemoteMCPTransport(t agentsv1.MCPServerTransport) bool {
	return t == agentsv1.MCPServerTransport_MCP_SERVER_TRANSPORT_STREAMABLE_HTTP ||
		t == agentsv1.MCPServerTransport_MCP_SERVER_TRANSPORT_SSE
}

func validateHTTPURL(field, raw string) error {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Scheme == "" || u.Host == "" {
		return twirp.InvalidArgumentError(field, "must be an absolute http or https URL")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return twirp.InvalidArgumentError(field, "must use http or https")
	}
	return nil
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
	return &agentsv1.ListMCPServersResponse{McpServers: redactInstalledGlobalMCPSecrets(servers)}, nil
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
	return &agentsv1.GetMCPServerResponse{McpServer: redactInstalledGlobalMCPSecret(m)}, nil
}

func MarkInstalledGlobalMCPPreset(server *agentsv1.MCPServer, presetID string) {
	if server == nil || strings.TrimSpace(presetID) == "" {
		return
	}
	if server.Metadata == nil {
		server.Metadata = map[string]string{}
	}
	server.Metadata[globalMCPPresetMetadataKey] = presetID
}

func redactInstalledGlobalMCPSecrets(servers []*agentsv1.MCPServer) []*agentsv1.MCPServer {
	out := make([]*agentsv1.MCPServer, 0, len(servers))
	for _, server := range servers {
		out = append(out, redactInstalledGlobalMCPSecret(server))
	}
	return out
}

// redactedHeaderValue replaces secret-bearing header values returned to
// installed-preset consumers. Header keys are kept so users can still see
// which headers the preset configures.
const redactedHeaderValue = "***"

func redactInstalledGlobalMCPSecret(server *agentsv1.MCPServer) *agentsv1.MCPServer {
	if server == nil || server.GetMetadata()[globalMCPPresetMetadataKey] == "" {
		return server
	}
	clone := proto.Clone(server).(*agentsv1.MCPServer)
	if oauth := clone.GetAuth().GetOauth2(); oauth != nil {
		oauth.ClientSecret = ""
	}
	// STATIC_HEADERS auth puts secrets directly in the headers map (e.g.
	// "Authorization: Bearer …"). The OAuth2 client_secret was already
	// covered above; mask the headers map too so installed presets do not
	// leak whatever the admin set on the global record.
	for k := range clone.Headers {
		clone.Headers[k] = redactedHeaderValue
	}
	return clone
}

func (s *MCPServerServiceServer) CreateMCPServer(ctx context.Context, req *agentsv1.CreateMCPServerRequest) (*agentsv1.CreateMCPServerResponse, error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if err := validateMCPServerConfig(req.GetMcpServer()); err != nil {
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
	if err := validateMCPServerConfig(req.GetMcpServer()); err != nil {
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
	if s.oauthService != nil {
		_ = s.oauthService.Disconnect(ctx, wsID, req.GetId())
	}
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

	timeout, err := internalagent.MCPTimeout(m)
	if err != nil {
		status.State = agentsv1.MCPServerStatus_STATE_DISCONNECTED
		status.Detail = err.Error()
		status.ToolCount = int32(len(m.GetToolFilter()))
		return &agentsv1.GetMCPServerStatusResponse{Status: status}, nil
	}
	probeCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	result, err := internalagent.ProbeMCPServerWithFactory(probeCtx, m, s.mcpHTTPClientFactory())
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
		timeout, err := internalagent.MCPTimeout(srv)
		if err != nil {
			resp.Errors[srv.GetId()] = err.Error()
			continue
		}
		probeCtx, cancel := context.WithTimeout(ctx, timeout)
		result, err := internalagent.ProbeMCPServerWithFactory(probeCtx, srv, s.mcpHTTPClientFactory())
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

func (s *MCPServerServiceServer) StartMCPServerOAuth(ctx context.Context, req *agentsv1.StartMCPServerOAuthRequest) (*agentsv1.StartMCPServerOAuthResponse, error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if s.oauthService == nil {
		return nil, twirp.NewError(twirp.FailedPrecondition, "mcp oauth service is not configured")
	}
	if strings.TrimSpace(req.GetServerId()) == "" {
		return nil, twirp.RequiredArgumentError("server_id")
	}
	srv, err := s.repo.GetMCPServer(ctx, wsID, req.GetServerId())
	if err != nil {
		return nil, toTwirpError(err)
	}
	if err := validateMCPServerConfig(srv); err != nil {
		return nil, err
	}
	userID := "api"
	if user, ok := auth.UserFromContext(ctx); ok {
		userID = user.GetId()
	}
	start, err := s.oauthService.Start(ctx, wsID, userID, srv, req.GetReturnUrl())
	if err != nil {
		return nil, twirp.NewError(twirp.FailedPrecondition, err.Error())
	}
	return &agentsv1.StartMCPServerOAuthResponse{AuthorizationUrl: start.AuthorizationURL, FlowId: start.FlowID}, nil
}

func (s *MCPServerServiceServer) CompleteMCPServerOAuth(ctx context.Context, req *agentsv1.CompleteMCPServerOAuthRequest) (*agentsv1.CompleteMCPServerOAuthResponse, error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if s.oauthService == nil {
		return nil, twirp.NewError(twirp.FailedPrecondition, "mcp oauth service is not configured")
	}
	if strings.TrimSpace(req.GetFlowId()) == "" {
		return nil, twirp.RequiredArgumentError("flow_id")
	}
	conn, err := s.oauthService.Complete(ctx, wsID, req.GetFlowId(), req.GetCode(), req.GetState())
	if err != nil {
		return nil, twirp.NewError(twirp.InvalidArgument, err.Error())
	}
	return &agentsv1.CompleteMCPServerOAuthResponse{Status: oauthStatusFromConnection(conn.ServerID, conn, nil)}, nil
}

func (s *MCPServerServiceServer) GetMCPServerOAuthStatus(ctx context.Context, req *agentsv1.GetMCPServerOAuthStatusRequest) (*agentsv1.GetMCPServerOAuthStatusResponse, error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.GetServerId()) == "" {
		return nil, twirp.RequiredArgumentError("server_id")
	}
	srv, err := s.repo.GetMCPServer(ctx, wsID, req.GetServerId())
	if err != nil {
		return nil, toTwirpError(err)
	}
	if s.oauthService == nil {
		return &agentsv1.GetMCPServerOAuthStatusResponse{Status: disconnectedOAuthStatus(srv.GetId(), "mcp oauth service is not configured")}, nil
	}
	conn, err := s.oauthService.Status(ctx, wsID, srv.GetId())
	return &agentsv1.GetMCPServerOAuthStatusResponse{Status: oauthStatusFromConnection(srv.GetId(), conn, err)}, nil
}

func (s *MCPServerServiceServer) DisconnectMCPServerOAuth(ctx context.Context, req *agentsv1.DisconnectMCPServerOAuthRequest) (*agentsv1.DisconnectMCPServerOAuthResponse, error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.GetServerId()) == "" {
		return nil, twirp.RequiredArgumentError("server_id")
	}
	if _, err := s.repo.GetMCPServer(ctx, wsID, req.GetServerId()); err != nil {
		return nil, toTwirpError(err)
	}
	if s.oauthService != nil {
		if err := s.oauthService.Disconnect(ctx, wsID, req.GetServerId()); err != nil {
			return nil, twirp.InternalErrorWith(err)
		}
	}
	return &agentsv1.DisconnectMCPServerOAuthResponse{Status: disconnectedOAuthStatus(req.GetServerId(), "disconnected")}, nil
}

func (s *MCPServerServiceServer) CompleteMCPServerOAuthCallback(ctx context.Context, state, code string) (string, *agentsv1.MCPOAuthConnectionStatus, error) {
	if s.oauthService == nil {
		return "", nil, fmt.Errorf("mcp oauth service is not configured")
	}
	result, err := s.oauthService.CompleteByState(ctx, state, code)
	if err != nil {
		return "", nil, err
	}
	return result.ReturnURL, oauthStatusFromConnection(result.Connection.ServerID, result.Connection, nil), nil
}

func (s *MCPServerServiceServer) mcpHTTPClientFactory() internalagent.MCPHTTPClientFactory {
	return s.httpFactory
}

func oauthStatusFromConnection(serverID string, conn *mcpoauthrepo.Connection, err error) *agentsv1.MCPOAuthConnectionStatus {
	if err != nil {
		if errors.Is(err, mcpoauthrepo.ErrNotFound) {
			return disconnectedOAuthStatus(serverID, "")
		}
		return &agentsv1.MCPOAuthConnectionStatus{
			ServerId:  serverID,
			State:     agentsv1.MCPOAuthConnectionState_MCPO_AUTH_CONNECTION_STATE_ERROR,
			Detail:    err.Error(),
			CheckedAt: timestamppb.New(time.Now().UTC()),
		}
	}
	if conn == nil {
		return disconnectedOAuthStatus(serverID, "")
	}
	status := &agentsv1.MCPOAuthConnectionStatus{
		ServerId:  conn.ServerID,
		State:     conn.State,
		Detail:    conn.LastError,
		Scopes:    append([]string(nil), conn.Scopes...),
		CheckedAt: timestamppb.New(time.Now().UTC()),
	}
	if !conn.ConnectedAt.IsZero() {
		status.ConnectedAt = timestamppb.New(conn.ConnectedAt)
	}
	if !conn.ExpiresAt.IsZero() {
		status.ExpiresAt = timestamppb.New(conn.ExpiresAt)
	}
	return status
}

func disconnectedOAuthStatus(serverID, detail string) *agentsv1.MCPOAuthConnectionStatus {
	return &agentsv1.MCPOAuthConnectionStatus{
		ServerId:  serverID,
		State:     agentsv1.MCPOAuthConnectionState_MCPO_AUTH_CONNECTION_STATE_DISCONNECTED,
		Detail:    detail,
		CheckedAt: timestamppb.New(time.Now().UTC()),
	}
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
