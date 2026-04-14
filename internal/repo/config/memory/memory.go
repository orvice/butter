package memory

import (
	"context"
	"fmt"
	"sync"

	configrepo "go.orx.me/apps/butter/internal/repo/config"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/proto"
)

// Store provides thread-safe in-memory CRUD for all config entities.
// It implements AgentRepository, MCPServerRepository, RemoteAgentRepository, and ChannelRepository.
type Store struct {
	mu           sync.RWMutex
	agents       map[string]*agentsv1.Agent
	mcpServers   map[string]*agentsv1.MCPServer
	remoteAgents map[string]*agentsv1.RemoteAgent
	channels     map[string]*agentsv1.AgentChannel
}

func New() *Store {
	return &Store{
		agents:       make(map[string]*agentsv1.Agent),
		mcpServers:   make(map[string]*agentsv1.MCPServer),
		remoteAgents: make(map[string]*agentsv1.RemoteAgent),
		channels:     make(map[string]*agentsv1.AgentChannel),
	}
}

// --- Agents ---

func (s *Store) ListAgents(_ context.Context) ([]*agentsv1.Agent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*agentsv1.Agent, 0, len(s.agents))
	for _, a := range s.agents {
		result = append(result, proto.Clone(a).(*agentsv1.Agent))
	}
	return result, nil
}

func (s *Store) GetAgent(_ context.Context, name string) (*agentsv1.Agent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.agents[name]
	if !ok {
		return nil, fmt.Errorf("agent %q: %w", name, configrepo.ErrNotFound)
	}
	return proto.Clone(a).(*agentsv1.Agent), nil
}

func (s *Store) CreateAgent(_ context.Context, agent *agentsv1.Agent) (*agentsv1.Agent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.agents[agent.GetName()]; ok {
		return nil, fmt.Errorf("agent %q: %w", agent.GetName(), configrepo.ErrAlreadyExists)
	}
	stored := proto.Clone(agent).(*agentsv1.Agent)
	s.agents[agent.GetName()] = stored
	return proto.Clone(stored).(*agentsv1.Agent), nil
}

func (s *Store) UpdateAgent(_ context.Context, agent *agentsv1.Agent) (*agentsv1.Agent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.agents[agent.GetName()]; !ok {
		return nil, fmt.Errorf("agent %q: %w", agent.GetName(), configrepo.ErrNotFound)
	}
	stored := proto.Clone(agent).(*agentsv1.Agent)
	s.agents[agent.GetName()] = stored
	return proto.Clone(stored).(*agentsv1.Agent), nil
}

func (s *Store) DeleteAgent(_ context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.agents[name]; !ok {
		return fmt.Errorf("agent %q: %w", name, configrepo.ErrNotFound)
	}
	delete(s.agents, name)
	return nil
}

// --- MCP Servers ---

func (s *Store) ListMCPServers(_ context.Context) ([]*agentsv1.MCPServer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*agentsv1.MCPServer, 0, len(s.mcpServers))
	for _, m := range s.mcpServers {
		result = append(result, proto.Clone(m).(*agentsv1.MCPServer))
	}
	return result, nil
}

func (s *Store) GetMCPServer(_ context.Context, id string) (*agentsv1.MCPServer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.mcpServers[id]
	if !ok {
		return nil, fmt.Errorf("mcp server %q: %w", id, configrepo.ErrNotFound)
	}
	return proto.Clone(m).(*agentsv1.MCPServer), nil
}

func (s *Store) CreateMCPServer(_ context.Context, server *agentsv1.MCPServer) (*agentsv1.MCPServer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.mcpServers[server.GetId()]; ok {
		return nil, fmt.Errorf("mcp server %q: %w", server.GetId(), configrepo.ErrAlreadyExists)
	}
	stored := proto.Clone(server).(*agentsv1.MCPServer)
	s.mcpServers[server.GetId()] = stored
	return proto.Clone(stored).(*agentsv1.MCPServer), nil
}

func (s *Store) UpdateMCPServer(_ context.Context, server *agentsv1.MCPServer) (*agentsv1.MCPServer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.mcpServers[server.GetId()]; !ok {
		return nil, fmt.Errorf("mcp server %q: %w", server.GetId(), configrepo.ErrNotFound)
	}
	stored := proto.Clone(server).(*agentsv1.MCPServer)
	s.mcpServers[server.GetId()] = stored
	return proto.Clone(stored).(*agentsv1.MCPServer), nil
}

func (s *Store) DeleteMCPServer(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.mcpServers[id]; !ok {
		return fmt.Errorf("mcp server %q: %w", id, configrepo.ErrNotFound)
	}
	delete(s.mcpServers, id)
	return nil
}

// --- Remote Agents ---

func (s *Store) ListRemoteAgents(_ context.Context) ([]*agentsv1.RemoteAgent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*agentsv1.RemoteAgent, 0, len(s.remoteAgents))
	for _, r := range s.remoteAgents {
		result = append(result, proto.Clone(r).(*agentsv1.RemoteAgent))
	}
	return result, nil
}

func (s *Store) GetRemoteAgent(_ context.Context, id string) (*agentsv1.RemoteAgent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.remoteAgents[id]
	if !ok {
		return nil, fmt.Errorf("remote agent %q: %w", id, configrepo.ErrNotFound)
	}
	return proto.Clone(r).(*agentsv1.RemoteAgent), nil
}

func (s *Store) CreateRemoteAgent(_ context.Context, agent *agentsv1.RemoteAgent) (*agentsv1.RemoteAgent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.remoteAgents[agent.GetId()]; ok {
		return nil, fmt.Errorf("remote agent %q: %w", agent.GetId(), configrepo.ErrAlreadyExists)
	}
	stored := proto.Clone(agent).(*agentsv1.RemoteAgent)
	s.remoteAgents[agent.GetId()] = stored
	return proto.Clone(stored).(*agentsv1.RemoteAgent), nil
}

func (s *Store) UpdateRemoteAgent(_ context.Context, agent *agentsv1.RemoteAgent) (*agentsv1.RemoteAgent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.remoteAgents[agent.GetId()]; !ok {
		return nil, fmt.Errorf("remote agent %q: %w", agent.GetId(), configrepo.ErrNotFound)
	}
	stored := proto.Clone(agent).(*agentsv1.RemoteAgent)
	s.remoteAgents[agent.GetId()] = stored
	return proto.Clone(stored).(*agentsv1.RemoteAgent), nil
}

func (s *Store) DeleteRemoteAgent(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.remoteAgents[id]; !ok {
		return fmt.Errorf("remote agent %q: %w", id, configrepo.ErrNotFound)
	}
	delete(s.remoteAgents, id)
	return nil
}

// --- Channels ---

func (s *Store) ListChannels(_ context.Context) ([]*agentsv1.AgentChannel, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*agentsv1.AgentChannel, 0, len(s.channels))
	for _, c := range s.channels {
		result = append(result, proto.Clone(c).(*agentsv1.AgentChannel))
	}
	return result, nil
}

func (s *Store) GetChannel(_ context.Context, name string) (*agentsv1.AgentChannel, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.channels[name]
	if !ok {
		return nil, fmt.Errorf("channel %q: %w", name, configrepo.ErrNotFound)
	}
	return proto.Clone(c).(*agentsv1.AgentChannel), nil
}

func (s *Store) CreateChannel(_ context.Context, channel *agentsv1.AgentChannel) (*agentsv1.AgentChannel, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.channels[channel.GetName()]; ok {
		return nil, fmt.Errorf("channel %q: %w", channel.GetName(), configrepo.ErrAlreadyExists)
	}
	stored := proto.Clone(channel).(*agentsv1.AgentChannel)
	s.channels[channel.GetName()] = stored
	return proto.Clone(stored).(*agentsv1.AgentChannel), nil
}

func (s *Store) UpdateChannel(_ context.Context, channel *agentsv1.AgentChannel) (*agentsv1.AgentChannel, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.channels[channel.GetName()]; !ok {
		return nil, fmt.Errorf("channel %q: %w", channel.GetName(), configrepo.ErrNotFound)
	}
	stored := proto.Clone(channel).(*agentsv1.AgentChannel)
	s.channels[channel.GetName()] = stored
	return proto.Clone(stored).(*agentsv1.AgentChannel), nil
}

func (s *Store) DeleteChannel(_ context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.channels[name]; !ok {
		return fmt.Errorf("channel %q: %w", name, configrepo.ErrNotFound)
	}
	delete(s.channels, name)
	return nil
}

// Seed populates the store from config data. Existing entries with matching keys are overwritten.
func (s *Store) Seed(ctx context.Context, agents []agentsv1.Agent, mcpServers []agentsv1.MCPServer, remoteAgents []agentsv1.RemoteAgent, channels []agentsv1.AgentChannel) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range agents {
		s.agents[agents[i].GetName()] = proto.Clone(&agents[i]).(*agentsv1.Agent)
	}
	for i := range mcpServers {
		s.mcpServers[mcpServers[i].GetId()] = proto.Clone(&mcpServers[i]).(*agentsv1.MCPServer)
	}
	for i := range remoteAgents {
		s.remoteAgents[remoteAgents[i].GetId()] = proto.Clone(&remoteAgents[i]).(*agentsv1.RemoteAgent)
	}
	for i := range channels {
		s.channels[channels[i].GetName()] = proto.Clone(&channels[i]).(*agentsv1.AgentChannel)
	}
}
