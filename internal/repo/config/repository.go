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
type AgentRepository interface {
	ListAgents(ctx context.Context) ([]*agentsv1.Agent, error)
	GetAgent(ctx context.Context, name string) (*agentsv1.Agent, error)
	CreateAgent(ctx context.Context, agent *agentsv1.Agent) (*agentsv1.Agent, error)
	UpdateAgent(ctx context.Context, agent *agentsv1.Agent) (*agentsv1.Agent, error)
	DeleteAgent(ctx context.Context, name string) error
}

// MCPServerRepository defines CRUD operations for MCPServer configurations.
type MCPServerRepository interface {
	ListMCPServers(ctx context.Context) ([]*agentsv1.MCPServer, error)
	GetMCPServer(ctx context.Context, id string) (*agentsv1.MCPServer, error)
	CreateMCPServer(ctx context.Context, server *agentsv1.MCPServer) (*agentsv1.MCPServer, error)
	UpdateMCPServer(ctx context.Context, server *agentsv1.MCPServer) (*agentsv1.MCPServer, error)
	DeleteMCPServer(ctx context.Context, id string) error
}

// RemoteAgentRepository defines CRUD operations for RemoteAgent configurations.
type RemoteAgentRepository interface {
	ListRemoteAgents(ctx context.Context) ([]*agentsv1.RemoteAgent, error)
	GetRemoteAgent(ctx context.Context, id string) (*agentsv1.RemoteAgent, error)
	CreateRemoteAgent(ctx context.Context, agent *agentsv1.RemoteAgent) (*agentsv1.RemoteAgent, error)
	UpdateRemoteAgent(ctx context.Context, agent *agentsv1.RemoteAgent) (*agentsv1.RemoteAgent, error)
	DeleteRemoteAgent(ctx context.Context, id string) error
}

// ChannelRepository defines CRUD operations for AgentChannel configurations.
type ChannelRepository interface {
	ListChannels(ctx context.Context) ([]*agentsv1.AgentChannel, error)
	GetChannel(ctx context.Context, name string) (*agentsv1.AgentChannel, error)
	CreateChannel(ctx context.Context, channel *agentsv1.AgentChannel) (*agentsv1.AgentChannel, error)
	UpdateChannel(ctx context.Context, channel *agentsv1.AgentChannel) (*agentsv1.AgentChannel, error)
	DeleteChannel(ctx context.Context, name string) error
}

// ModelProviderRepository defines CRUD operations for LLM model provider configurations.
type ModelProviderRepository interface {
	ListModelProviders(ctx context.Context) ([]*agentsv1.ModelProvider, error)
	GetModelProvider(ctx context.Context, name string) (*agentsv1.ModelProvider, error)
	CreateModelProvider(ctx context.Context, provider *agentsv1.ModelProvider) (*agentsv1.ModelProvider, error)
	UpdateModelProvider(ctx context.Context, provider *agentsv1.ModelProvider) (*agentsv1.ModelProvider, error)
	DeleteModelProvider(ctx context.Context, name string) error
}
