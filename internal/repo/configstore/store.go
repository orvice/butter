package configstore

import (
	"fmt"
	"sync"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/proto"
)

// Store provides thread-safe in-memory CRUD for Agent, MCPServer, and RemoteAgent configs.
type Store struct {
	mu           sync.RWMutex
	agents       map[string]*agentsv1.Agent
	mcpServers   map[string]*agentsv1.MCPServer
	remoteAgents map[string]*agentsv1.RemoteAgent
}

func New() *Store {
	return &Store{
		agents:       make(map[string]*agentsv1.Agent),
		mcpServers:   make(map[string]*agentsv1.MCPServer),
		remoteAgents: make(map[string]*agentsv1.RemoteAgent),
	}
}

// Seed populates the store from AppConfig data. Existing entries are overwritten.
func (s *Store) Seed(agents []agentsv1.Agent, mcpServers []agentsv1.MCPServer, remoteAgents []agentsv1.RemoteAgent) {
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
}

// --- Agents ---

func (s *Store) ListAgents() []*agentsv1.Agent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*agentsv1.Agent, 0, len(s.agents))
	for _, a := range s.agents {
		result = append(result, proto.Clone(a).(*agentsv1.Agent))
	}
	return result
}

func (s *Store) GetAgent(name string) (*agentsv1.Agent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.agents[name]
	if !ok {
		return nil, fmt.Errorf("agent %q not found", name)
	}
	return proto.Clone(a).(*agentsv1.Agent), nil
}

func (s *Store) CreateAgent(agent *agentsv1.Agent) (*agentsv1.Agent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.agents[agent.GetName()]; ok {
		return nil, fmt.Errorf("agent %q already exists", agent.GetName())
	}
	stored := proto.Clone(agent).(*agentsv1.Agent)
	s.agents[agent.GetName()] = stored
	return proto.Clone(stored).(*agentsv1.Agent), nil
}

func (s *Store) UpdateAgent(agent *agentsv1.Agent) (*agentsv1.Agent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.agents[agent.GetName()]; !ok {
		return nil, fmt.Errorf("agent %q not found", agent.GetName())
	}
	stored := proto.Clone(agent).(*agentsv1.Agent)
	s.agents[agent.GetName()] = stored
	return proto.Clone(stored).(*agentsv1.Agent), nil
}

func (s *Store) DeleteAgent(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.agents[name]; !ok {
		return fmt.Errorf("agent %q not found", name)
	}
	delete(s.agents, name)
	return nil
}

// --- MCP Servers ---

func (s *Store) ListMCPServers() []*agentsv1.MCPServer {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*agentsv1.MCPServer, 0, len(s.mcpServers))
	for _, m := range s.mcpServers {
		result = append(result, proto.Clone(m).(*agentsv1.MCPServer))
	}
	return result
}

func (s *Store) GetMCPServer(id string) (*agentsv1.MCPServer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.mcpServers[id]
	if !ok {
		return nil, fmt.Errorf("mcp server %q not found", id)
	}
	return proto.Clone(m).(*agentsv1.MCPServer), nil
}

func (s *Store) CreateMCPServer(server *agentsv1.MCPServer) (*agentsv1.MCPServer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.mcpServers[server.GetId()]; ok {
		return nil, fmt.Errorf("mcp server %q already exists", server.GetId())
	}
	stored := proto.Clone(server).(*agentsv1.MCPServer)
	s.mcpServers[server.GetId()] = stored
	return proto.Clone(stored).(*agentsv1.MCPServer), nil
}

func (s *Store) UpdateMCPServer(server *agentsv1.MCPServer) (*agentsv1.MCPServer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.mcpServers[server.GetId()]; !ok {
		return nil, fmt.Errorf("mcp server %q not found", server.GetId())
	}
	stored := proto.Clone(server).(*agentsv1.MCPServer)
	s.mcpServers[server.GetId()] = stored
	return proto.Clone(stored).(*agentsv1.MCPServer), nil
}

func (s *Store) DeleteMCPServer(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.mcpServers[id]; !ok {
		return fmt.Errorf("mcp server %q not found", id)
	}
	delete(s.mcpServers, id)
	return nil
}

// --- Remote Agents ---

func (s *Store) ListRemoteAgents() []*agentsv1.RemoteAgent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*agentsv1.RemoteAgent, 0, len(s.remoteAgents))
	for _, r := range s.remoteAgents {
		result = append(result, proto.Clone(r).(*agentsv1.RemoteAgent))
	}
	return result
}

func (s *Store) GetRemoteAgent(id string) (*agentsv1.RemoteAgent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.remoteAgents[id]
	if !ok {
		return nil, fmt.Errorf("remote agent %q not found", id)
	}
	return proto.Clone(r).(*agentsv1.RemoteAgent), nil
}

func (s *Store) CreateRemoteAgent(agent *agentsv1.RemoteAgent) (*agentsv1.RemoteAgent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.remoteAgents[agent.GetId()]; ok {
		return nil, fmt.Errorf("remote agent %q already exists", agent.GetId())
	}
	stored := proto.Clone(agent).(*agentsv1.RemoteAgent)
	s.remoteAgents[agent.GetId()] = stored
	return proto.Clone(stored).(*agentsv1.RemoteAgent), nil
}

func (s *Store) UpdateRemoteAgent(agent *agentsv1.RemoteAgent) (*agentsv1.RemoteAgent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.remoteAgents[agent.GetId()]; !ok {
		return nil, fmt.Errorf("remote agent %q not found", agent.GetId())
	}
	stored := proto.Clone(agent).(*agentsv1.RemoteAgent)
	s.remoteAgents[agent.GetId()] = stored
	return proto.Clone(stored).(*agentsv1.RemoteAgent), nil
}

func (s *Store) DeleteRemoteAgent(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.remoteAgents[id]; !ok {
		return fmt.Errorf("remote agent %q not found", id)
	}
	delete(s.remoteAgents, id)
	return nil
}
