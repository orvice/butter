package runner

import (
	"testing"

	"google.golang.org/adk/v2/agent"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func newResolveTestService() *Service {
	return NewServiceForTestAgents(
		&agentsv1.Agent{Name: "researcher", AgentId: "researcher-v2", DisplayName: "Researcher", WorkspaceId: "ws-a"},
		&agentsv1.Agent{Name: "writer", AgentId: "writer", WorkspaceId: "ws-a"},
		&agentsv1.Agent{Name: "reviewer", WorkspaceId: "ws-b"}, // no agent_id assigned
	)
}

func TestResolveAgentRefByID(t *testing.T) {
	s := newResolveTestService()

	name, ok := s.ResolveAgentRef("ws-a", "researcher-v2")
	if !ok || name != "researcher" {
		t.Fatalf("ResolveAgentRef(ws-a, researcher-v2) = %q, %v; want researcher, true", name, ok)
	}

	// agent_id resolution is workspace-scoped.
	if name, ok := s.ResolveAgentRef("ws-b", "researcher-v2"); ok {
		t.Fatalf("ResolveAgentRef(ws-b, researcher-v2) = %q, true; want miss", name)
	}

	// Empty workspace is the admin/system path: resolve across workspaces.
	if name, ok := s.ResolveAgentRef("", "researcher-v2"); !ok || name != "researcher" {
		t.Fatalf("ResolveAgentRef(\"\", researcher-v2) = %q, %v; want researcher, true", name, ok)
	}
}

func TestResolveAgentRefNoNameFallback(t *testing.T) {
	s := newResolveTestService()

	// The runtime name is not a resolvable reference — only agent_id is.
	if name, ok := s.ResolveAgentRef("ws-a", "researcher"); ok {
		t.Fatalf("ResolveAgentRef by runtime name = %q, true; want miss (name is not an id)", name)
	}
	// An agent without an assigned agent_id is simply unreachable by ref.
	if name, ok := s.ResolveAgentRef("ws-b", "reviewer"); ok {
		t.Fatalf("ResolveAgentRef(ws-b, reviewer) = %q, true; want miss (no agent_id)", name)
	}
	if name, ok := s.ResolveAgentRef("ws-a", ""); ok {
		t.Fatalf("ResolveAgentRef(empty id) = %q, true; want miss", name)
	}
}

func TestResolveAgentRefBuiltin(t *testing.T) {
	s := NewServiceForTestAgents(
		&agentsv1.Agent{Name: "helper", AgentId: "helper", WorkspaceId: "ws-a"},
	)
	// Register a builtin (no proto) whose name is its identifier.
	s.agents["system"] = agent.Agent(nil)

	// Resolvable from the system context by name-as-id.
	if name, ok := s.ResolveAgentRef("", "system"); !ok || name != "system" {
		t.Fatalf("ResolveAgentRef(\"\", system) = %q, %v; want system, true", name, ok)
	}
	// Rejected from a tenant context, mirroring HasAgentInWorkspace.
	if name, ok := s.ResolveAgentRef("ws-a", "system"); ok {
		t.Fatalf("ResolveAgentRef(ws-a, system) = %q, true; want miss (builtin is system-only)", name)
	}
}

func TestHasAgentIDInWorkspace(t *testing.T) {
	s := newResolveTestService()

	if !s.HasAgentIDInWorkspace("ws-a", "writer") {
		t.Fatal("HasAgentIDInWorkspace(ws-a, writer) = false; want true")
	}
	if s.HasAgentIDInWorkspace("ws-b", "writer") {
		t.Fatal("HasAgentIDInWorkspace(ws-b, writer) = true; want false")
	}
}

func TestGetAgentIdentity(t *testing.T) {
	s := newResolveTestService()

	id, display, ok := s.GetAgentIdentity("researcher")
	if !ok || id != "researcher-v2" || display != "Researcher" {
		t.Fatalf("GetAgentIdentity(researcher) = %q, %q, %v; want researcher-v2, Researcher, true", id, display, ok)
	}

	// Display name falls back to the agent name when unset.
	id, display, ok = s.GetAgentIdentity("writer")
	if !ok || id != "writer" || display != "writer" {
		t.Fatalf("GetAgentIdentity(writer) = %q, %q, %v; want writer, writer, true", id, display, ok)
	}

	if _, _, ok := s.GetAgentIdentity("ghost"); ok {
		t.Fatal("GetAgentIdentity(ghost) = true; want false")
	}
}

func TestGetAgentIdentityBuiltin(t *testing.T) {
	s := NewServiceForTestAgents()
	s.agents["system"] = agent.Agent(nil)

	// A builtin reports its name as both id and display name.
	id, display, ok := s.GetAgentIdentity("system")
	if !ok || id != "system" || display != "system" {
		t.Fatalf("GetAgentIdentity(system) = %q, %q, %v; want system, system, true", id, display, ok)
	}
}
