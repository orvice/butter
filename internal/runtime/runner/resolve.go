package runner

import (
	"google.golang.org/adk/v2/agent"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// ResolveAgentRef resolves an immutable agent_id to its registered runtime
// name. There is no legacy-name resolution path: every interface now
// addresses agents by agent_id (issue #213 contract step), so a miss is a
// miss — a caller holding a stale or wrong id can never accidentally hit an
// agent by name.
//
// An empty workspaceID is the admin/system path and matches any workspace.
// Built-in agents (dynamically registered, no proto config, e.g. the system
// agent) have no workspace binding and their name doubles as their agent_id;
// they resolve only from the system context, mirroring HasAgentInWorkspace.
func (s *Service) ResolveAgentRef(workspaceID, agentID string) (string, bool) {
	if agentID == "" {
		return "", false
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	for name, p := range s.agentsProto {
		if p.GetAgentId() != agentID {
			continue
		}
		if workspaceID != "" && p.GetWorkspaceId() != workspaceID {
			continue
		}
		if _, registered := s.agents[name]; !registered {
			continue
		}
		return name, true
	}

	// Built-in agent: name-as-id, no workspace binding — system context only.
	if workspaceID == "" {
		if _, registered := s.agents[agentID]; registered {
			if _, hasProto := s.agentsProto[agentID]; !hasProto {
				return agentID, true
			}
		}
	}
	return "", false
}

// SupportsModelOverride reports whether the referenced Agent accepts a Butter
// model override. Pi owns model selection on its ButterBox, so policy-driven
// entry points expose no Butter model candidates while a Pi Agent is active.
func (s *Service) SupportsModelOverride(workspaceID, agentID string) (bool, bool) {
	name, ok := s.ResolveAgentRef(workspaceID, agentID)
	if !ok {
		return false, false
	}
	_, _, agentType, hasProto := s.GetAgentMeta(name)
	if !hasProto {
		// Built-in Agents have no proto type and retain their existing override
		// behavior on the system path.
		return true, true
	}
	return agentType != agentsv1.AgentType_AGENT_TYPE_PI, true
}

// HasAgentIDInWorkspace reports whether an agent with the given agent_id is
// registered and belongs to the given workspace. An empty workspaceID is the
// admin/system path.
func (s *Service) HasAgentIDInWorkspace(workspaceID, agentID string) bool {
	_, ok := s.ResolveAgentRef(workspaceID, agentID)
	return ok
}

// GetAgentIdentity returns the named agent's immutable agent_id and its
// human-readable display name (falling back to the agent name). Built-in
// agents (no proto config, e.g. the system agent) report their name as both
// id and display name, since their name is their stable identifier. ok is
// false only when the agent is not registered at all.
func (s *Service) GetAgentIdentity(name string) (agentID, displayName string, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, found := s.agentsProto[name]
	if !found {
		if _, registered := s.agents[name]; registered {
			return name, name, true
		}
		return "", "", false
	}
	displayName = p.GetDisplayName()
	if displayName == "" {
		displayName = p.GetName()
	}
	return p.GetAgentId(), displayName, true
}

// NewServiceForTestAgents builds a Service whose registry contains the given
// proto agents. No ADK runtime is initialized — only the lookup and
// resolution helpers are usable. Exposed for cross-package tests.
func NewServiceForTestAgents(agents ...*agentsv1.Agent) *Service {
	s := &Service{
		agents:      map[string]agent.Agent{},
		agentsProto: map[string]*agentsv1.Agent{},
	}
	for _, a := range agents {
		s.agents[a.GetName()] = nil
		s.agentsProto[a.GetName()] = a
	}
	return s
}
