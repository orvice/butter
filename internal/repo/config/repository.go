package configrepo

import (
	"context"
	"errors"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

// AgentRepository defines CRUD operations for Agent configurations.
// All methods are scoped to a single workspace.
type AgentRepository interface {
	ListAgents(ctx context.Context, workspaceID string) ([]*agentsv1.Agent, error)
	GetAgent(ctx context.Context, workspaceID, name string) (*agentsv1.Agent, error)
	CreateAgent(ctx context.Context, workspaceID string, agent *agentsv1.Agent) (*agentsv1.Agent, error)
	UpdateAgent(ctx context.Context, workspaceID string, agent *agentsv1.Agent) (*agentsv1.Agent, error)
	DeleteAgent(ctx context.Context, workspaceID, name string) error

	// ListAgentsAcrossWorkspaces returns agents from every workspace, used by
	// the runtime to (re)build runners across all configured tenants.
	ListAgentsAcrossWorkspaces(ctx context.Context) ([]*agentsv1.Agent, error)
}

// MCPServerRepository defines CRUD operations for MCPServer configurations.
type MCPServerRepository interface {
	ListMCPServers(ctx context.Context, workspaceID string) ([]*agentsv1.MCPServer, error)
	GetMCPServer(ctx context.Context, workspaceID, id string) (*agentsv1.MCPServer, error)
	CreateMCPServer(ctx context.Context, workspaceID string, server *agentsv1.MCPServer) (*agentsv1.MCPServer, error)
	UpdateMCPServer(ctx context.Context, workspaceID string, server *agentsv1.MCPServer) (*agentsv1.MCPServer, error)
	DeleteMCPServer(ctx context.Context, workspaceID, id string) error
	ListMCPServersAcrossWorkspaces(ctx context.Context) ([]*agentsv1.MCPServer, error)
}

// GlobalMCPServerRepository defines CRUD operations for admin-managed MCP
// server presets that can be installed into workspaces.
type GlobalMCPServerRepository interface {
	ListGlobalMCPServers(ctx context.Context) ([]*agentsv1.MCPServer, error)
	GetGlobalMCPServer(ctx context.Context, id string) (*agentsv1.MCPServer, error)
	CreateGlobalMCPServer(ctx context.Context, server *agentsv1.MCPServer) (*agentsv1.MCPServer, error)
	UpdateGlobalMCPServer(ctx context.Context, server *agentsv1.MCPServer) (*agentsv1.MCPServer, error)
	DeleteGlobalMCPServer(ctx context.Context, id string) error
}

// RemoteAgentRepository defines CRUD operations for RemoteAgent configurations.
type RemoteAgentRepository interface {
	ListRemoteAgents(ctx context.Context, workspaceID string) ([]*agentsv1.RemoteAgent, error)
	GetRemoteAgent(ctx context.Context, workspaceID, id string) (*agentsv1.RemoteAgent, error)
	CreateRemoteAgent(ctx context.Context, workspaceID string, agent *agentsv1.RemoteAgent) (*agentsv1.RemoteAgent, error)
	UpdateRemoteAgent(ctx context.Context, workspaceID string, agent *agentsv1.RemoteAgent) (*agentsv1.RemoteAgent, error)
	DeleteRemoteAgent(ctx context.Context, workspaceID, id string) error
	ListRemoteAgentsAcrossWorkspaces(ctx context.Context) ([]*agentsv1.RemoteAgent, error)
}

// DaemonConfigRepository defines workspace-scoped daemon worker definitions.
type DaemonConfigRepository interface {
	ListDaemonConfigs(ctx context.Context, workspaceID string) ([]*agentsv1.DaemonConfig, error)
	GetDaemonConfig(ctx context.Context, workspaceID, id string) (*agentsv1.DaemonConfig, error)
	CreateDaemonConfig(ctx context.Context, workspaceID string, daemon *agentsv1.DaemonConfig) (*agentsv1.DaemonConfig, error)
	UpdateDaemonConfig(ctx context.Context, workspaceID string, daemon *agentsv1.DaemonConfig) (*agentsv1.DaemonConfig, error)
	DeleteDaemonConfig(ctx context.Context, workspaceID, id string) error
	ListDaemonConfigsAcrossWorkspaces(ctx context.Context) ([]*agentsv1.DaemonConfig, error)
}

// ChannelRepository defines CRUD operations for AgentChannel configurations.
type ChannelRepository interface {
	ListChannels(ctx context.Context, workspaceID string) ([]*agentsv1.AgentChannel, error)
	GetChannel(ctx context.Context, workspaceID, name string) (*agentsv1.AgentChannel, error)
	CreateChannel(ctx context.Context, workspaceID string, channel *agentsv1.AgentChannel) (*agentsv1.AgentChannel, error)
	UpdateChannel(ctx context.Context, workspaceID string, channel *agentsv1.AgentChannel) (*agentsv1.AgentChannel, error)
	DeleteChannel(ctx context.Context, workspaceID, name string) error
	ListChannelsAcrossWorkspaces(ctx context.Context) ([]*agentsv1.AgentChannel, error)
}

// ModelProviderRepository defines CRUD operations for LLM model provider configurations.
type ModelProviderRepository interface {
	ListModelProviders(ctx context.Context, workspaceID string) ([]*agentsv1.ModelProvider, error)
	GetModelProvider(ctx context.Context, workspaceID, name string) (*agentsv1.ModelProvider, error)
	CreateModelProvider(ctx context.Context, workspaceID string, provider *agentsv1.ModelProvider) (*agentsv1.ModelProvider, error)
	UpdateModelProvider(ctx context.Context, workspaceID string, provider *agentsv1.ModelProvider) (*agentsv1.ModelProvider, error)
	DeleteModelProvider(ctx context.Context, workspaceID, name string) error
	ListModelProvidersAcrossWorkspaces(ctx context.Context) ([]*agentsv1.ModelProvider, error)
}

// NotifyGroupRepository defines CRUD operations for outbound notification groups.
type NotifyGroupRepository interface {
	ListNotifyGroups(ctx context.Context, workspaceID string) ([]*agentsv1.NotifyGroup, error)
	GetNotifyGroup(ctx context.Context, workspaceID, name string) (*agentsv1.NotifyGroup, error)
	CreateNotifyGroup(ctx context.Context, workspaceID string, group *agentsv1.NotifyGroup) (*agentsv1.NotifyGroup, error)
	UpdateNotifyGroup(ctx context.Context, workspaceID string, group *agentsv1.NotifyGroup) (*agentsv1.NotifyGroup, error)
	DeleteNotifyGroup(ctx context.Context, workspaceID, name string) error
	ListNotifyGroupsAcrossWorkspaces(ctx context.Context) ([]*agentsv1.NotifyGroup, error)
}
