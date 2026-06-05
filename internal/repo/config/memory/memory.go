package memory

import (
	"context"
	"fmt"
	"sync"

	configrepo "go.orx.me/apps/butter/internal/repo/config"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/proto"
)

// Store provides thread-safe in-memory CRUD for all config entities,
// scoped per-workspace.
type Store struct {
	mu             sync.RWMutex
	agents         map[string]map[string]*agentsv1.Agent
	globalMCP      map[string]*agentsv1.MCPServer
	mcpServers     map[string]map[string]*agentsv1.MCPServer
	remoteAgents   map[string]map[string]*agentsv1.RemoteAgent
	daemonRuntimes map[string]map[string]*agentsv1.DaemonRuntime
	channels       map[string]map[string]*agentsv1.AgentChannel
	modelProviders map[string]map[string]*agentsv1.ModelProvider
	notifyGroups   map[string]map[string]*agentsv1.NotifyGroup
}

func New() *Store {
	return &Store{
		agents:         make(map[string]map[string]*agentsv1.Agent),
		globalMCP:      make(map[string]*agentsv1.MCPServer),
		mcpServers:     make(map[string]map[string]*agentsv1.MCPServer),
		remoteAgents:   make(map[string]map[string]*agentsv1.RemoteAgent),
		daemonRuntimes: make(map[string]map[string]*agentsv1.DaemonRuntime),
		channels:       make(map[string]map[string]*agentsv1.AgentChannel),
		modelProviders: make(map[string]map[string]*agentsv1.ModelProvider),
		notifyGroups:   make(map[string]map[string]*agentsv1.NotifyGroup),
	}
}

func cloneAgent(a *agentsv1.Agent) *agentsv1.Agent { return proto.Clone(a).(*agentsv1.Agent) }
func cloneMCP(m *agentsv1.MCPServer) *agentsv1.MCPServer {
	return proto.Clone(m).(*agentsv1.MCPServer)
}
func cloneRemote(r *agentsv1.RemoteAgent) *agentsv1.RemoteAgent {
	return proto.Clone(r).(*agentsv1.RemoteAgent)
}
func cloneDaemonRuntime(d *agentsv1.DaemonRuntime) *agentsv1.DaemonRuntime {
	return proto.Clone(d).(*agentsv1.DaemonRuntime)
}
func cloneChannel(c *agentsv1.AgentChannel) *agentsv1.AgentChannel {
	return proto.Clone(c).(*agentsv1.AgentChannel)
}
func cloneProvider(p *agentsv1.ModelProvider) *agentsv1.ModelProvider {
	return proto.Clone(p).(*agentsv1.ModelProvider)
}
func cloneNotifyGroup(g *agentsv1.NotifyGroup) *agentsv1.NotifyGroup {
	return proto.Clone(g).(*agentsv1.NotifyGroup)
}

func notFound(entity, ws, key string) error {
	return fmt.Errorf("%s %q (workspace %q): %w", entity, key, ws, configrepo.ErrNotFound)
}

func alreadyExists(entity, ws, key string) error {
	return fmt.Errorf("%s %q (workspace %q): %w", entity, key, ws, configrepo.ErrAlreadyExists)
}

// --- Agents ---

func (s *Store) ListAgents(_ context.Context, workspaceID string) ([]*agentsv1.Agent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	bucket := s.agents[workspaceID]
	out := make([]*agentsv1.Agent, 0, len(bucket))
	for _, a := range bucket {
		out = append(out, cloneAgent(a))
	}
	return out, nil
}

func (s *Store) ListAgentsAcrossWorkspaces(_ context.Context) ([]*agentsv1.Agent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*agentsv1.Agent
	for _, bucket := range s.agents {
		for _, a := range bucket {
			out = append(out, cloneAgent(a))
		}
	}
	return out, nil
}

func (s *Store) GetAgent(_ context.Context, workspaceID, name string) (*agentsv1.Agent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	bucket, ok := s.agents[workspaceID]
	if !ok {
		return nil, notFound("agent", workspaceID, name)
	}
	a, ok := bucket[name]
	if !ok {
		return nil, notFound("agent", workspaceID, name)
	}
	return cloneAgent(a), nil
}

func (s *Store) CreateAgent(_ context.Context, workspaceID string, agent *agentsv1.Agent) (*agentsv1.Agent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	bucket := s.agents[workspaceID]
	if bucket == nil {
		bucket = make(map[string]*agentsv1.Agent)
		s.agents[workspaceID] = bucket
	}
	if _, ok := bucket[agent.GetName()]; ok {
		return nil, alreadyExists("agent", workspaceID, agent.GetName())
	}
	stored := cloneAgent(agent)
	stored.WorkspaceId = workspaceID
	bucket[agent.GetName()] = stored
	return cloneAgent(stored), nil
}

func (s *Store) UpdateAgent(_ context.Context, workspaceID string, agent *agentsv1.Agent) (*agentsv1.Agent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	bucket, ok := s.agents[workspaceID]
	if !ok {
		return nil, notFound("agent", workspaceID, agent.GetName())
	}
	if _, ok := bucket[agent.GetName()]; !ok {
		return nil, notFound("agent", workspaceID, agent.GetName())
	}
	stored := cloneAgent(agent)
	stored.WorkspaceId = workspaceID
	bucket[agent.GetName()] = stored
	return cloneAgent(stored), nil
}

func (s *Store) DeleteAgent(_ context.Context, workspaceID, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	bucket, ok := s.agents[workspaceID]
	if !ok {
		return notFound("agent", workspaceID, name)
	}
	if _, ok := bucket[name]; !ok {
		return notFound("agent", workspaceID, name)
	}
	delete(bucket, name)
	return nil
}

// --- MCP Servers ---

func (s *Store) ListMCPServers(_ context.Context, workspaceID string) ([]*agentsv1.MCPServer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	bucket := s.mcpServers[workspaceID]
	out := make([]*agentsv1.MCPServer, 0, len(bucket))
	for _, m := range bucket {
		out = append(out, cloneMCP(m))
	}
	return out, nil
}

func (s *Store) ListMCPServersAcrossWorkspaces(_ context.Context) ([]*agentsv1.MCPServer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*agentsv1.MCPServer
	for _, bucket := range s.mcpServers {
		for _, m := range bucket {
			out = append(out, cloneMCP(m))
		}
	}
	return out, nil
}

func (s *Store) GetMCPServer(_ context.Context, workspaceID, id string) (*agentsv1.MCPServer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	bucket, ok := s.mcpServers[workspaceID]
	if !ok {
		return nil, notFound("mcp server", workspaceID, id)
	}
	m, ok := bucket[id]
	if !ok {
		return nil, notFound("mcp server", workspaceID, id)
	}
	return cloneMCP(m), nil
}

func (s *Store) CreateMCPServer(_ context.Context, workspaceID string, server *agentsv1.MCPServer) (*agentsv1.MCPServer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	bucket := s.mcpServers[workspaceID]
	if bucket == nil {
		bucket = make(map[string]*agentsv1.MCPServer)
		s.mcpServers[workspaceID] = bucket
	}
	if _, ok := bucket[server.GetId()]; ok {
		return nil, alreadyExists("mcp server", workspaceID, server.GetId())
	}
	stored := cloneMCP(server)
	stored.WorkspaceId = workspaceID
	bucket[server.GetId()] = stored
	return cloneMCP(stored), nil
}

func (s *Store) UpdateMCPServer(_ context.Context, workspaceID string, server *agentsv1.MCPServer) (*agentsv1.MCPServer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	bucket, ok := s.mcpServers[workspaceID]
	if !ok {
		return nil, notFound("mcp server", workspaceID, server.GetId())
	}
	if _, ok := bucket[server.GetId()]; !ok {
		return nil, notFound("mcp server", workspaceID, server.GetId())
	}
	stored := cloneMCP(server)
	stored.WorkspaceId = workspaceID
	bucket[server.GetId()] = stored
	return cloneMCP(stored), nil
}

func (s *Store) DeleteMCPServer(_ context.Context, workspaceID, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	bucket, ok := s.mcpServers[workspaceID]
	if !ok {
		return notFound("mcp server", workspaceID, id)
	}
	if _, ok := bucket[id]; !ok {
		return notFound("mcp server", workspaceID, id)
	}
	delete(bucket, id)
	return nil
}

// --- Global MCP Servers ---

func (s *Store) ListGlobalMCPServers(_ context.Context) ([]*agentsv1.MCPServer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*agentsv1.MCPServer, 0, len(s.globalMCP))
	for _, m := range s.globalMCP {
		out = append(out, cloneMCP(m))
	}
	return out, nil
}

func (s *Store) GetGlobalMCPServer(_ context.Context, id string) (*agentsv1.MCPServer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.globalMCP[id]
	if !ok {
		return nil, notFound("global mcp server", "", id)
	}
	return cloneMCP(m), nil
}

func (s *Store) CreateGlobalMCPServer(_ context.Context, server *agentsv1.MCPServer) (*agentsv1.MCPServer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.globalMCP[server.GetId()]; ok {
		return nil, alreadyExists("global mcp server", "", server.GetId())
	}
	stored := cloneMCP(server)
	stored.WorkspaceId = ""
	s.globalMCP[server.GetId()] = stored
	return cloneMCP(stored), nil
}

func (s *Store) UpdateGlobalMCPServer(_ context.Context, server *agentsv1.MCPServer) (*agentsv1.MCPServer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.globalMCP[server.GetId()]; !ok {
		return nil, notFound("global mcp server", "", server.GetId())
	}
	stored := cloneMCP(server)
	stored.WorkspaceId = ""
	s.globalMCP[server.GetId()] = stored
	return cloneMCP(stored), nil
}

func (s *Store) DeleteGlobalMCPServer(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.globalMCP[id]; !ok {
		return notFound("global mcp server", "", id)
	}
	delete(s.globalMCP, id)
	return nil
}

// --- Remote Agents ---

func (s *Store) ListRemoteAgents(_ context.Context, workspaceID string) ([]*agentsv1.RemoteAgent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	bucket := s.remoteAgents[workspaceID]
	out := make([]*agentsv1.RemoteAgent, 0, len(bucket))
	for _, r := range bucket {
		out = append(out, cloneRemote(r))
	}
	return out, nil
}

func (s *Store) ListRemoteAgentsAcrossWorkspaces(_ context.Context) ([]*agentsv1.RemoteAgent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*agentsv1.RemoteAgent
	for _, bucket := range s.remoteAgents {
		for _, r := range bucket {
			out = append(out, cloneRemote(r))
		}
	}
	return out, nil
}

func (s *Store) GetRemoteAgent(_ context.Context, workspaceID, id string) (*agentsv1.RemoteAgent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	bucket, ok := s.remoteAgents[workspaceID]
	if !ok {
		return nil, notFound("remote agent", workspaceID, id)
	}
	r, ok := bucket[id]
	if !ok {
		return nil, notFound("remote agent", workspaceID, id)
	}
	return cloneRemote(r), nil
}

func (s *Store) CreateRemoteAgent(_ context.Context, workspaceID string, agent *agentsv1.RemoteAgent) (*agentsv1.RemoteAgent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	bucket := s.remoteAgents[workspaceID]
	if bucket == nil {
		bucket = make(map[string]*agentsv1.RemoteAgent)
		s.remoteAgents[workspaceID] = bucket
	}
	if _, ok := bucket[agent.GetId()]; ok {
		return nil, alreadyExists("remote agent", workspaceID, agent.GetId())
	}
	stored := cloneRemote(agent)
	stored.WorkspaceId = workspaceID
	bucket[agent.GetId()] = stored
	return cloneRemote(stored), nil
}

func (s *Store) UpdateRemoteAgent(_ context.Context, workspaceID string, agent *agentsv1.RemoteAgent) (*agentsv1.RemoteAgent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	bucket, ok := s.remoteAgents[workspaceID]
	if !ok {
		return nil, notFound("remote agent", workspaceID, agent.GetId())
	}
	if _, ok := bucket[agent.GetId()]; !ok {
		return nil, notFound("remote agent", workspaceID, agent.GetId())
	}
	stored := cloneRemote(agent)
	stored.WorkspaceId = workspaceID
	bucket[agent.GetId()] = stored
	return cloneRemote(stored), nil
}

func (s *Store) DeleteRemoteAgent(_ context.Context, workspaceID, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	bucket, ok := s.remoteAgents[workspaceID]
	if !ok {
		return notFound("remote agent", workspaceID, id)
	}
	if _, ok := bucket[id]; !ok {
		return notFound("remote agent", workspaceID, id)
	}
	delete(bucket, id)
	return nil
}

// --- Daemon Configs ---

func (s *Store) ListDaemonRuntimes(_ context.Context, workspaceID string) ([]*agentsv1.DaemonRuntime, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	bucket := s.daemonRuntimes[workspaceID]
	out := make([]*agentsv1.DaemonRuntime, 0, len(bucket))
	for _, d := range bucket {
		out = append(out, cloneDaemonRuntime(d))
	}
	return out, nil
}

func (s *Store) ListDaemonRuntimesAcrossWorkspaces(_ context.Context) ([]*agentsv1.DaemonRuntime, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*agentsv1.DaemonRuntime
	for _, bucket := range s.daemonRuntimes {
		for _, d := range bucket {
			out = append(out, cloneDaemonRuntime(d))
		}
	}
	return out, nil
}

func (s *Store) GetDaemonRuntime(_ context.Context, workspaceID, id string) (*agentsv1.DaemonRuntime, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	bucket, ok := s.daemonRuntimes[workspaceID]
	if !ok {
		return nil, notFound("daemon runtime", workspaceID, id)
	}
	d, ok := bucket[id]
	if !ok {
		return nil, notFound("daemon runtime", workspaceID, id)
	}
	return cloneDaemonRuntime(d), nil
}

func (s *Store) CreateDaemonRuntime(_ context.Context, workspaceID string, daemon *agentsv1.DaemonRuntime) (*agentsv1.DaemonRuntime, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	bucket := s.daemonRuntimes[workspaceID]
	if bucket == nil {
		bucket = make(map[string]*agentsv1.DaemonRuntime)
		s.daemonRuntimes[workspaceID] = bucket
	}
	if _, ok := bucket[daemon.GetId()]; ok {
		return nil, alreadyExists("daemon runtime", workspaceID, daemon.GetId())
	}
	stored := cloneDaemonRuntime(daemon)
	stored.WorkspaceId = workspaceID
	bucket[daemon.GetId()] = stored
	return cloneDaemonRuntime(stored), nil
}

func (s *Store) UpdateDaemonRuntime(_ context.Context, workspaceID string, daemon *agentsv1.DaemonRuntime) (*agentsv1.DaemonRuntime, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	bucket, ok := s.daemonRuntimes[workspaceID]
	if !ok {
		return nil, notFound("daemon runtime", workspaceID, daemon.GetId())
	}
	if _, ok := bucket[daemon.GetId()]; !ok {
		return nil, notFound("daemon runtime", workspaceID, daemon.GetId())
	}
	stored := cloneDaemonRuntime(daemon)
	stored.WorkspaceId = workspaceID
	bucket[daemon.GetId()] = stored
	return cloneDaemonRuntime(stored), nil
}

func (s *Store) DeleteDaemonRuntime(_ context.Context, workspaceID, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	bucket, ok := s.daemonRuntimes[workspaceID]
	if !ok {
		return notFound("daemon runtime", workspaceID, id)
	}
	if _, ok := bucket[id]; !ok {
		return notFound("daemon runtime", workspaceID, id)
	}
	delete(bucket, id)
	return nil
}

// --- Channels ---

func (s *Store) ListChannels(_ context.Context, workspaceID string) ([]*agentsv1.AgentChannel, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	bucket := s.channels[workspaceID]
	out := make([]*agentsv1.AgentChannel, 0, len(bucket))
	for _, c := range bucket {
		out = append(out, cloneChannel(c))
	}
	return out, nil
}

func (s *Store) ListChannelsAcrossWorkspaces(_ context.Context) ([]*agentsv1.AgentChannel, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*agentsv1.AgentChannel
	for _, bucket := range s.channels {
		for _, c := range bucket {
			out = append(out, cloneChannel(c))
		}
	}
	return out, nil
}

func (s *Store) GetChannel(_ context.Context, workspaceID, name string) (*agentsv1.AgentChannel, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	bucket, ok := s.channels[workspaceID]
	if !ok {
		return nil, notFound("channel", workspaceID, name)
	}
	c, ok := bucket[name]
	if !ok {
		return nil, notFound("channel", workspaceID, name)
	}
	return cloneChannel(c), nil
}

func (s *Store) CreateChannel(_ context.Context, workspaceID string, channel *agentsv1.AgentChannel) (*agentsv1.AgentChannel, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	bucket := s.channels[workspaceID]
	if bucket == nil {
		bucket = make(map[string]*agentsv1.AgentChannel)
		s.channels[workspaceID] = bucket
	}
	if _, ok := bucket[channel.GetName()]; ok {
		return nil, alreadyExists("channel", workspaceID, channel.GetName())
	}
	stored := cloneChannel(channel)
	stored.WorkspaceId = workspaceID
	bucket[channel.GetName()] = stored
	return cloneChannel(stored), nil
}

func (s *Store) UpdateChannel(_ context.Context, workspaceID string, channel *agentsv1.AgentChannel) (*agentsv1.AgentChannel, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	bucket, ok := s.channels[workspaceID]
	if !ok {
		return nil, notFound("channel", workspaceID, channel.GetName())
	}
	if _, ok := bucket[channel.GetName()]; !ok {
		return nil, notFound("channel", workspaceID, channel.GetName())
	}
	stored := cloneChannel(channel)
	stored.WorkspaceId = workspaceID
	bucket[channel.GetName()] = stored
	return cloneChannel(stored), nil
}

func (s *Store) DeleteChannel(_ context.Context, workspaceID, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	bucket, ok := s.channels[workspaceID]
	if !ok {
		return notFound("channel", workspaceID, name)
	}
	if _, ok := bucket[name]; !ok {
		return notFound("channel", workspaceID, name)
	}
	delete(bucket, name)
	return nil
}

// --- Model Providers ---

func (s *Store) ListModelProviders(_ context.Context, workspaceID string) ([]*agentsv1.ModelProvider, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	bucket := s.modelProviders[workspaceID]
	out := make([]*agentsv1.ModelProvider, 0, len(bucket))
	for _, p := range bucket {
		out = append(out, cloneProvider(p))
	}
	return out, nil
}

func (s *Store) ListModelProvidersAcrossWorkspaces(_ context.Context) ([]*agentsv1.ModelProvider, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*agentsv1.ModelProvider
	for _, bucket := range s.modelProviders {
		for _, p := range bucket {
			out = append(out, cloneProvider(p))
		}
	}
	return out, nil
}

func (s *Store) GetModelProvider(_ context.Context, workspaceID, name string) (*agentsv1.ModelProvider, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	bucket, ok := s.modelProviders[workspaceID]
	if !ok {
		return nil, notFound("model provider", workspaceID, name)
	}
	p, ok := bucket[name]
	if !ok {
		return nil, notFound("model provider", workspaceID, name)
	}
	return cloneProvider(p), nil
}

func (s *Store) CreateModelProvider(_ context.Context, workspaceID string, provider *agentsv1.ModelProvider) (*agentsv1.ModelProvider, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	bucket := s.modelProviders[workspaceID]
	if bucket == nil {
		bucket = make(map[string]*agentsv1.ModelProvider)
		s.modelProviders[workspaceID] = bucket
	}
	if _, ok := bucket[provider.GetName()]; ok {
		return nil, alreadyExists("model provider", workspaceID, provider.GetName())
	}
	stored := cloneProvider(provider)
	stored.WorkspaceId = workspaceID
	bucket[provider.GetName()] = stored
	return cloneProvider(stored), nil
}

func (s *Store) UpdateModelProvider(_ context.Context, workspaceID string, provider *agentsv1.ModelProvider) (*agentsv1.ModelProvider, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	bucket, ok := s.modelProviders[workspaceID]
	if !ok {
		return nil, notFound("model provider", workspaceID, provider.GetName())
	}
	if _, ok := bucket[provider.GetName()]; !ok {
		return nil, notFound("model provider", workspaceID, provider.GetName())
	}
	stored := cloneProvider(provider)
	stored.WorkspaceId = workspaceID
	bucket[provider.GetName()] = stored
	return cloneProvider(stored), nil
}

func (s *Store) DeleteModelProvider(_ context.Context, workspaceID, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	bucket, ok := s.modelProviders[workspaceID]
	if !ok {
		return notFound("model provider", workspaceID, name)
	}
	if _, ok := bucket[name]; !ok {
		return notFound("model provider", workspaceID, name)
	}
	delete(bucket, name)
	return nil
}

// --- Notify Groups ---

func (s *Store) ListNotifyGroups(_ context.Context, workspaceID string) ([]*agentsv1.NotifyGroup, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	bucket := s.notifyGroups[workspaceID]
	out := make([]*agentsv1.NotifyGroup, 0, len(bucket))
	for _, g := range bucket {
		out = append(out, cloneNotifyGroup(g))
	}
	return out, nil
}

func (s *Store) ListNotifyGroupsAcrossWorkspaces(_ context.Context) ([]*agentsv1.NotifyGroup, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*agentsv1.NotifyGroup
	for _, bucket := range s.notifyGroups {
		for _, g := range bucket {
			out = append(out, cloneNotifyGroup(g))
		}
	}
	return out, nil
}

func (s *Store) GetNotifyGroup(_ context.Context, workspaceID, name string) (*agentsv1.NotifyGroup, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	bucket, ok := s.notifyGroups[workspaceID]
	if !ok {
		return nil, notFound("notify group", workspaceID, name)
	}
	g, ok := bucket[name]
	if !ok {
		return nil, notFound("notify group", workspaceID, name)
	}
	return cloneNotifyGroup(g), nil
}

func (s *Store) CreateNotifyGroup(_ context.Context, workspaceID string, group *agentsv1.NotifyGroup) (*agentsv1.NotifyGroup, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	bucket := s.notifyGroups[workspaceID]
	if bucket == nil {
		bucket = make(map[string]*agentsv1.NotifyGroup)
		s.notifyGroups[workspaceID] = bucket
	}
	if _, ok := bucket[group.GetName()]; ok {
		return nil, alreadyExists("notify group", workspaceID, group.GetName())
	}
	stored := cloneNotifyGroup(group)
	stored.WorkspaceId = workspaceID
	bucket[group.GetName()] = stored
	return cloneNotifyGroup(stored), nil
}

func (s *Store) UpdateNotifyGroup(_ context.Context, workspaceID string, group *agentsv1.NotifyGroup) (*agentsv1.NotifyGroup, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	bucket, ok := s.notifyGroups[workspaceID]
	if !ok {
		return nil, notFound("notify group", workspaceID, group.GetName())
	}
	if _, ok := bucket[group.GetName()]; !ok {
		return nil, notFound("notify group", workspaceID, group.GetName())
	}
	stored := cloneNotifyGroup(group)
	stored.WorkspaceId = workspaceID
	bucket[group.GetName()] = stored
	return cloneNotifyGroup(stored), nil
}

func (s *Store) DeleteNotifyGroup(_ context.Context, workspaceID, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	bucket, ok := s.notifyGroups[workspaceID]
	if !ok {
		return notFound("notify group", workspaceID, name)
	}
	if _, ok := bucket[name]; !ok {
		return notFound("notify group", workspaceID, name)
	}
	delete(bucket, name)
	return nil
}
