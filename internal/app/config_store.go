package app

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"go.orx.me/apps/butter/internal/config"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	configmemory "go.orx.me/apps/butter/internal/repo/config/memory"
	configmongo "go.orx.me/apps/butter/internal/repo/config/mongo"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/proto"
)

type configBackend interface {
	configrepo.AgentRepository
	configrepo.MCPServerRepository
	configrepo.RemoteAgentRepository
	configrepo.ChannelRepository
	configrepo.ModelProviderRepository
}

// ConfigStore is a runtime-selectable config repository wrapper. All CRUD
// calls require a workspace id; the convenience snapshot helpers
// (SyncToConfig, ListAgents/Channels/... AcrossWorkspaces) flatten every
// workspace's configs for the runtime layers that still consume the global
// AppConfig view.
type ConfigStore struct {
	mu      sync.RWMutex
	backend configBackend
}

func NewConfigStore() *ConfigStore {
	return &ConfigStore{backend: configmemory.New()}
}

func (s *ConfigStore) ActiveBackendName() string {
	switch s.current().(type) {
	case *configmemory.Store:
		return "memory"
	case *configmongo.Store:
		return "mongo"
	default:
		return "unknown"
	}
}

func (s *ConfigStore) InitFromConfig(ctx context.Context, cfg *config.AppConfig) error {
	backend, err := s.newBackend(ctx, cfg)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.backend = backend
	s.mu.Unlock()

	if err := s.ensureIndexes(ctx, backend); err != nil {
		return err
	}
	return s.SyncToConfig(ctx, cfg)
}

func (s *ConfigStore) newBackend(ctx context.Context, cfg *config.AppConfig) (configBackend, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.StorageBackend)) {
	case "", "mongo":
		db, err := connectMongo(ctx, cfg)
		if err != nil {
			return nil, err
		}
		return configmongo.New(db), nil
	case "memory":
		return configmemory.New(), nil
	default:
		return nil, fmt.Errorf("unsupported storage backend %q", cfg.StorageBackend)
	}
}

func (s *ConfigStore) ensureIndexes(ctx context.Context, backend configBackend) error {
	if mongoBackend, ok := backend.(*configmongo.Store); ok {
		return mongoBackend.EnsureIndexes(ctx)
	}
	return nil
}

// loadIntoConfig flattens the persisted configs from all workspaces into
// the legacy AppConfig view. Runtime services (runner, channel manager,
// cron) build their internal indexes from this snapshot and resolve
// workspace ownership via each entity's WorkspaceId field.
func (s *ConfigStore) loadIntoConfig(ctx context.Context, cfg *config.AppConfig) error {
	cfg.Agents = nil
	cfg.MCPServerConfigs = nil
	cfg.RemoteAgents = nil
	cfg.Channels = nil
	cfg.ModelProviders = nil

	agents, err := s.current().ListAgentsAcrossWorkspaces(ctx)
	if err != nil {
		return err
	}
	for _, agent := range agents {
		cfg.Agents = append(cfg.Agents, agentsv1.Agent{})
		proto.Merge(&cfg.Agents[len(cfg.Agents)-1], agent)
	}

	mcpServers, err := s.current().ListMCPServersAcrossWorkspaces(ctx)
	if err != nil {
		return err
	}
	for _, server := range mcpServers {
		cfg.MCPServerConfigs = append(cfg.MCPServerConfigs, agentsv1.MCPServer{})
		proto.Merge(&cfg.MCPServerConfigs[len(cfg.MCPServerConfigs)-1], server)
	}

	remoteAgents, err := s.current().ListRemoteAgentsAcrossWorkspaces(ctx)
	if err != nil {
		return err
	}
	for _, agent := range remoteAgents {
		cfg.RemoteAgents = append(cfg.RemoteAgents, agentsv1.RemoteAgent{})
		proto.Merge(&cfg.RemoteAgents[len(cfg.RemoteAgents)-1], agent)
	}

	channels, err := s.current().ListChannelsAcrossWorkspaces(ctx)
	if err != nil {
		return err
	}
	for _, channel := range channels {
		cfg.Channels = append(cfg.Channels, agentsv1.AgentChannel{})
		proto.Merge(&cfg.Channels[len(cfg.Channels)-1], channel)
	}

	modelProviders, err := s.current().ListModelProvidersAcrossWorkspaces(ctx)
	if err != nil {
		return err
	}
	for _, provider := range modelProviders {
		cfg.ModelProviders = append(cfg.ModelProviders, agentsv1.ModelProvider{})
		proto.Merge(&cfg.ModelProviders[len(cfg.ModelProviders)-1], provider)
	}

	return nil
}

func (s *ConfigStore) SyncToConfig(ctx context.Context, cfg *config.AppConfig) error {
	return s.loadIntoConfig(ctx, cfg)
}

func (s *ConfigStore) current() configBackend {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.backend
}

// --- Agents ---

func (s *ConfigStore) ListAgents(ctx context.Context, workspaceID string) ([]*agentsv1.Agent, error) {
	return s.current().ListAgents(ctx, workspaceID)
}

func (s *ConfigStore) ListAgentsAcrossWorkspaces(ctx context.Context) ([]*agentsv1.Agent, error) {
	return s.current().ListAgentsAcrossWorkspaces(ctx)
}

func (s *ConfigStore) GetAgent(ctx context.Context, workspaceID, name string) (*agentsv1.Agent, error) {
	return s.current().GetAgent(ctx, workspaceID, name)
}

func (s *ConfigStore) CreateAgent(ctx context.Context, workspaceID string, agent *agentsv1.Agent) (*agentsv1.Agent, error) {
	return s.current().CreateAgent(ctx, workspaceID, agent)
}

func (s *ConfigStore) UpdateAgent(ctx context.Context, workspaceID string, agent *agentsv1.Agent) (*agentsv1.Agent, error) {
	return s.current().UpdateAgent(ctx, workspaceID, agent)
}

func (s *ConfigStore) DeleteAgent(ctx context.Context, workspaceID, name string) error {
	return s.current().DeleteAgent(ctx, workspaceID, name)
}

// --- MCP Servers ---

func (s *ConfigStore) ListMCPServers(ctx context.Context, workspaceID string) ([]*agentsv1.MCPServer, error) {
	return s.current().ListMCPServers(ctx, workspaceID)
}

func (s *ConfigStore) ListMCPServersAcrossWorkspaces(ctx context.Context) ([]*agentsv1.MCPServer, error) {
	return s.current().ListMCPServersAcrossWorkspaces(ctx)
}

func (s *ConfigStore) GetMCPServer(ctx context.Context, workspaceID, id string) (*agentsv1.MCPServer, error) {
	return s.current().GetMCPServer(ctx, workspaceID, id)
}

func (s *ConfigStore) CreateMCPServer(ctx context.Context, workspaceID string, server *agentsv1.MCPServer) (*agentsv1.MCPServer, error) {
	return s.current().CreateMCPServer(ctx, workspaceID, server)
}

func (s *ConfigStore) UpdateMCPServer(ctx context.Context, workspaceID string, server *agentsv1.MCPServer) (*agentsv1.MCPServer, error) {
	return s.current().UpdateMCPServer(ctx, workspaceID, server)
}

func (s *ConfigStore) DeleteMCPServer(ctx context.Context, workspaceID, id string) error {
	return s.current().DeleteMCPServer(ctx, workspaceID, id)
}

// --- Remote Agents ---

func (s *ConfigStore) ListRemoteAgents(ctx context.Context, workspaceID string) ([]*agentsv1.RemoteAgent, error) {
	return s.current().ListRemoteAgents(ctx, workspaceID)
}

func (s *ConfigStore) ListRemoteAgentsAcrossWorkspaces(ctx context.Context) ([]*agentsv1.RemoteAgent, error) {
	return s.current().ListRemoteAgentsAcrossWorkspaces(ctx)
}

func (s *ConfigStore) GetRemoteAgent(ctx context.Context, workspaceID, id string) (*agentsv1.RemoteAgent, error) {
	return s.current().GetRemoteAgent(ctx, workspaceID, id)
}

func (s *ConfigStore) CreateRemoteAgent(ctx context.Context, workspaceID string, agent *agentsv1.RemoteAgent) (*agentsv1.RemoteAgent, error) {
	return s.current().CreateRemoteAgent(ctx, workspaceID, agent)
}

func (s *ConfigStore) UpdateRemoteAgent(ctx context.Context, workspaceID string, agent *agentsv1.RemoteAgent) (*agentsv1.RemoteAgent, error) {
	return s.current().UpdateRemoteAgent(ctx, workspaceID, agent)
}

func (s *ConfigStore) DeleteRemoteAgent(ctx context.Context, workspaceID, id string) error {
	return s.current().DeleteRemoteAgent(ctx, workspaceID, id)
}

// --- Channels ---

func (s *ConfigStore) ListChannels(ctx context.Context, workspaceID string) ([]*agentsv1.AgentChannel, error) {
	return s.current().ListChannels(ctx, workspaceID)
}

func (s *ConfigStore) ListChannelsAcrossWorkspaces(ctx context.Context) ([]*agentsv1.AgentChannel, error) {
	return s.current().ListChannelsAcrossWorkspaces(ctx)
}

func (s *ConfigStore) GetChannel(ctx context.Context, workspaceID, name string) (*agentsv1.AgentChannel, error) {
	return s.current().GetChannel(ctx, workspaceID, name)
}

func (s *ConfigStore) CreateChannel(ctx context.Context, workspaceID string, channel *agentsv1.AgentChannel) (*agentsv1.AgentChannel, error) {
	return s.current().CreateChannel(ctx, workspaceID, channel)
}

func (s *ConfigStore) UpdateChannel(ctx context.Context, workspaceID string, channel *agentsv1.AgentChannel) (*agentsv1.AgentChannel, error) {
	return s.current().UpdateChannel(ctx, workspaceID, channel)
}

func (s *ConfigStore) DeleteChannel(ctx context.Context, workspaceID, name string) error {
	return s.current().DeleteChannel(ctx, workspaceID, name)
}

// --- Model Providers ---

func (s *ConfigStore) ListModelProviders(ctx context.Context, workspaceID string) ([]*agentsv1.ModelProvider, error) {
	return s.current().ListModelProviders(ctx, workspaceID)
}

func (s *ConfigStore) ListModelProvidersAcrossWorkspaces(ctx context.Context) ([]*agentsv1.ModelProvider, error) {
	return s.current().ListModelProvidersAcrossWorkspaces(ctx)
}

func (s *ConfigStore) GetModelProvider(ctx context.Context, workspaceID, name string) (*agentsv1.ModelProvider, error) {
	return s.current().GetModelProvider(ctx, workspaceID, name)
}

func (s *ConfigStore) CreateModelProvider(ctx context.Context, workspaceID string, provider *agentsv1.ModelProvider) (*agentsv1.ModelProvider, error) {
	return s.current().CreateModelProvider(ctx, workspaceID, provider)
}

func (s *ConfigStore) UpdateModelProvider(ctx context.Context, workspaceID string, provider *agentsv1.ModelProvider) (*agentsv1.ModelProvider, error) {
	return s.current().UpdateModelProvider(ctx, workspaceID, provider)
}

func (s *ConfigStore) DeleteModelProvider(ctx context.Context, workspaceID, name string) error {
	return s.current().DeleteModelProvider(ctx, workspaceID, name)
}
