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
}

// ConfigStore is a runtime-selectable config repository wrapper.
type ConfigStore struct {
	mu      sync.RWMutex
	backend configBackend
}

func NewConfigStore() *ConfigStore {
	return &ConfigStore{backend: configmemory.New()}
}

func (s *ConfigStore) InitFromConfig(ctx context.Context, cfg *config.AppConfig) error {
	backend, err := s.newBackend(ctx, cfg)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.backend = backend
	s.mu.Unlock()

	if err := s.seedIfNeeded(ctx, cfg, backend); err != nil {
		return err
	}
	return s.SyncToConfig(ctx, cfg)
}

func (s *ConfigStore) newBackend(ctx context.Context, cfg *config.AppConfig) (configBackend, error) {
	switch strings.ToLower(cfg.StorageBackend) {
	case "", "memory":
		return configmemory.New(), nil
	case "mongo":
		db, err := connectMongo(ctx, cfg)
		if err != nil {
			return nil, err
		}
		return configmongo.New(db), nil
	default:
		return nil, fmt.Errorf("unsupported storage backend %q", cfg.StorageBackend)
	}
}

func (s *ConfigStore) seedIfNeeded(ctx context.Context, cfg *config.AppConfig, backend configBackend) error {
	switch store := backend.(type) {
	case *configmemory.Store:
		store.Seed(ctx, cfg.Agents, cfg.MCPServerConfigs, cfg.RemoteAgents, cfg.Channels)
		return nil
	case *configmongo.Store:
		if err := seedMongoCollectionIfEmpty(ctx, cfg.Agents, store.ListAgents, func(items []agentsv1.Agent) error {
			return store.Seed(ctx, items, nil, nil, nil)
		}); err != nil {
			return err
		}
		if err := seedMongoCollectionIfEmpty(ctx, cfg.MCPServerConfigs, store.ListMCPServers, func(items []agentsv1.MCPServer) error {
			return store.Seed(ctx, nil, items, nil, nil)
		}); err != nil {
			return err
		}
		if err := seedMongoCollectionIfEmpty(ctx, cfg.RemoteAgents, store.ListRemoteAgents, func(items []agentsv1.RemoteAgent) error {
			return store.Seed(ctx, nil, nil, items, nil)
		}); err != nil {
			return err
		}
		if err := seedMongoCollectionIfEmpty(ctx, cfg.Channels, store.ListChannels, func(items []agentsv1.AgentChannel) error {
			return store.Seed(ctx, nil, nil, nil, items)
		}); err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("unsupported config backend %T", backend)
	}
}

func seedMongoCollectionIfEmpty[T any](
	ctx context.Context,
	cfgItems []T,
	list func(context.Context) ([]*T, error),
	seed func([]T) error,
) error {
	existing, err := list(ctx)
	if err != nil {
		return err
	}
	if len(existing) > 0 || len(cfgItems) == 0 {
		return nil
	}
	return seed(cfgItems)
}

func (s *ConfigStore) loadIntoConfig(ctx context.Context, cfg *config.AppConfig) error {
	cfg.Agents = nil
	cfg.MCPServerConfigs = nil
	cfg.RemoteAgents = nil
	cfg.Channels = nil

	agents, err := s.ListAgents(ctx)
	if err != nil {
		return err
	}
	for _, agent := range agents {
		cfg.Agents = append(cfg.Agents, agentsv1.Agent{})
		proto.Merge(&cfg.Agents[len(cfg.Agents)-1], agent)
	}

	mcpServers, err := s.ListMCPServers(ctx)
	if err != nil {
		return err
	}
	for _, server := range mcpServers {
		cfg.MCPServerConfigs = append(cfg.MCPServerConfigs, agentsv1.MCPServer{})
		proto.Merge(&cfg.MCPServerConfigs[len(cfg.MCPServerConfigs)-1], server)
	}

	remoteAgents, err := s.ListRemoteAgents(ctx)
	if err != nil {
		return err
	}
	for _, agent := range remoteAgents {
		cfg.RemoteAgents = append(cfg.RemoteAgents, agentsv1.RemoteAgent{})
		proto.Merge(&cfg.RemoteAgents[len(cfg.RemoteAgents)-1], agent)
	}

	channels, err := s.ListChannels(ctx)
	if err != nil {
		return err
	}
	for _, channel := range channels {
		cfg.Channels = append(cfg.Channels, agentsv1.AgentChannel{})
		proto.Merge(&cfg.Channels[len(cfg.Channels)-1], channel)
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

func (s *ConfigStore) ListAgents(ctx context.Context) ([]*agentsv1.Agent, error) {
	return s.current().ListAgents(ctx)
}

func (s *ConfigStore) GetAgent(ctx context.Context, name string) (*agentsv1.Agent, error) {
	return s.current().GetAgent(ctx, name)
}

func (s *ConfigStore) CreateAgent(ctx context.Context, agent *agentsv1.Agent) (*agentsv1.Agent, error) {
	return s.current().CreateAgent(ctx, agent)
}

func (s *ConfigStore) UpdateAgent(ctx context.Context, agent *agentsv1.Agent) (*agentsv1.Agent, error) {
	return s.current().UpdateAgent(ctx, agent)
}

func (s *ConfigStore) DeleteAgent(ctx context.Context, name string) error {
	return s.current().DeleteAgent(ctx, name)
}

func (s *ConfigStore) ListMCPServers(ctx context.Context) ([]*agentsv1.MCPServer, error) {
	return s.current().ListMCPServers(ctx)
}

func (s *ConfigStore) GetMCPServer(ctx context.Context, id string) (*agentsv1.MCPServer, error) {
	return s.current().GetMCPServer(ctx, id)
}

func (s *ConfigStore) CreateMCPServer(ctx context.Context, server *agentsv1.MCPServer) (*agentsv1.MCPServer, error) {
	return s.current().CreateMCPServer(ctx, server)
}

func (s *ConfigStore) UpdateMCPServer(ctx context.Context, server *agentsv1.MCPServer) (*agentsv1.MCPServer, error) {
	return s.current().UpdateMCPServer(ctx, server)
}

func (s *ConfigStore) DeleteMCPServer(ctx context.Context, id string) error {
	return s.current().DeleteMCPServer(ctx, id)
}

func (s *ConfigStore) ListRemoteAgents(ctx context.Context) ([]*agentsv1.RemoteAgent, error) {
	return s.current().ListRemoteAgents(ctx)
}

func (s *ConfigStore) GetRemoteAgent(ctx context.Context, id string) (*agentsv1.RemoteAgent, error) {
	return s.current().GetRemoteAgent(ctx, id)
}

func (s *ConfigStore) CreateRemoteAgent(ctx context.Context, agent *agentsv1.RemoteAgent) (*agentsv1.RemoteAgent, error) {
	return s.current().CreateRemoteAgent(ctx, agent)
}

func (s *ConfigStore) UpdateRemoteAgent(ctx context.Context, agent *agentsv1.RemoteAgent) (*agentsv1.RemoteAgent, error) {
	return s.current().UpdateRemoteAgent(ctx, agent)
}

func (s *ConfigStore) DeleteRemoteAgent(ctx context.Context, id string) error {
	return s.current().DeleteRemoteAgent(ctx, id)
}

func (s *ConfigStore) ListChannels(ctx context.Context) ([]*agentsv1.AgentChannel, error) {
	return s.current().ListChannels(ctx)
}

func (s *ConfigStore) GetChannel(ctx context.Context, name string) (*agentsv1.AgentChannel, error) {
	return s.current().GetChannel(ctx, name)
}

func (s *ConfigStore) CreateChannel(ctx context.Context, channel *agentsv1.AgentChannel) (*agentsv1.AgentChannel, error) {
	return s.current().CreateChannel(ctx, channel)
}

func (s *ConfigStore) UpdateChannel(ctx context.Context, channel *agentsv1.AgentChannel) (*agentsv1.AgentChannel, error) {
	return s.current().UpdateChannel(ctx, channel)
}

func (s *ConfigStore) DeleteChannel(ctx context.Context, name string) error {
	return s.current().DeleteChannel(ctx, name)
}
