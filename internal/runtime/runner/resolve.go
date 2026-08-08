package runner

import (
	"google.golang.org/adk/v2/agent"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// ResolveAgentRef resolves an agent reference to its registered runtime name.
//
// agentID takes precedence: when set, the agent is looked up by its immutable
// workspace-unique agent_id and a miss never falls through to legacyName —
// otherwise a caller holding a stale ID could silently invoke a different
// agent. When agentID is empty, legacyName is resolved with the same
// workspace semantics as HasAgentInWorkspace. An empty workspaceID is the
// admin/system path and matches any workspace.
func (s *Service) ResolveAgentRef(workspaceID, agentID, legacyName string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if agentID != "" {
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
		return "", false
	}

	if legacyName == "" {
		return "", false
	}
	if _, ok := s.agents[legacyName]; !ok {
		return "", false
	}
	if workspaceID == "" {
		return legacyName, true
	}
	p, ok := s.agentsProto[legacyName]
	if !ok {
		// Dynamically registered agents (no proto) have no workspace binding —
		// reject them from a tenant context, mirroring HasAgentInWorkspace.
		return "", false
	}
	if p.GetWorkspaceId() != workspaceID {
		return "", false
	}
	return legacyName, true
}

// AgentRefLabel renders the human-facing identifier for an agent reference —
// the agent_id when set, otherwise the legacy name — for use in error
// messages and logs. It mirrors the precedence of ResolveAgentRef.
func AgentRefLabel(agentID, legacyName string) string {
	if agentID != "" {
		return agentID
	}
	return legacyName
}

// HasAgentIDInWorkspace reports whether an agent with the given agent_id is
// registered and belongs to the given workspace. An empty workspaceID is the
// admin/system path.
func (s *Service) HasAgentIDInWorkspace(workspaceID, agentID string) bool {
	_, ok := s.ResolveAgentRef(workspaceID, agentID, "")
	return ok
}

// GetAgentIdentity returns the named agent's immutable agent_id (empty when
// not yet assigned) and its human-readable display name (falling back to the
// agent name). ok is false when the agent has no proto config.
func (s *Service) GetAgentIdentity(name string) (agentID, displayName string, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, found := s.agentsProto[name]
	if !found {
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
