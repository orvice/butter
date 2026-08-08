package runner

import (
	"testing"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func newResolveTestService() *Service {
	return NewServiceForTestAgents(
		&agentsv1.Agent{Name: "researcher", AgentId: "researcher-v2", DisplayName: "Researcher", WorkspaceId: "ws-a"},
		&agentsv1.Agent{Name: "writer", AgentId: "writer", WorkspaceId: "ws-a"},
		&agentsv1.Agent{Name: "reviewer", WorkspaceId: "ws-b"}, // no agent_id assigned yet
	)
}

func TestResolveAgentRefByID(t *testing.T) {
	s := newResolveTestService()

	name, ok := s.ResolveAgentRef("ws-a", "researcher-v2", "")
	if !ok || name != "researcher" {
		t.Fatalf("ResolveAgentRef(ws-a, researcher-v2) = %q, %v; want researcher, true", name, ok)
	}

	// agent_id resolution is workspace-scoped.
	if name, ok := s.ResolveAgentRef("ws-b", "researcher-v2", ""); ok {
		t.Fatalf("ResolveAgentRef(ws-b, researcher-v2) = %q, true; want miss", name)
	}

	// Empty workspace is the admin/system path: resolve across workspaces.
	if name, ok := s.ResolveAgentRef("", "researcher-v2", ""); !ok || name != "researcher" {
		t.Fatalf("ResolveAgentRef(\"\", researcher-v2) = %q, %v; want researcher, true", name, ok)
	}
}

func TestResolveAgentRefPrefersIDOverLegacyName(t *testing.T) {
	s := newResolveTestService()

	// When both are set, agent_id wins even if the legacy name points elsewhere.
	name, ok := s.ResolveAgentRef("ws-a", "researcher-v2", "writer")
	if !ok || name != "researcher" {
		t.Fatalf("ResolveAgentRef(id+name) = %q, %v; want researcher, true", name, ok)
	}

	// A set-but-unknown agent_id is a miss; it must not fall through to the
	// legacy name, otherwise a caller could silently invoke the wrong agent.
	if name, ok := s.ResolveAgentRef("ws-a", "nope", "writer"); ok {
		t.Fatalf("ResolveAgentRef(unknown id) = %q, true; want miss", name)
	}
}

func TestResolveAgentRefLegacyNameFallback(t *testing.T) {
	s := newResolveTestService()

	name, ok := s.ResolveAgentRef("ws-b", "", "reviewer")
	if !ok || name != "reviewer" {
		t.Fatalf("ResolveAgentRef(ws-b, name=reviewer) = %q, %v; want reviewer, true", name, ok)
	}

	// Legacy name resolution enforces the workspace boundary like
	// HasAgentInWorkspace.
	if name, ok := s.ResolveAgentRef("ws-a", "", "reviewer"); ok {
		t.Fatalf("ResolveAgentRef(ws-a, name=reviewer) = %q, true; want miss", name)
	}

	if name, ok := s.ResolveAgentRef("ws-a", "", ""); ok {
		t.Fatalf("ResolveAgentRef(empty ref) = %q, true; want miss", name)
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

	// Agents without an assigned ID report an empty id but still resolve.
	id, display, ok = s.GetAgentIdentity("reviewer")
	if !ok || id != "" || display != "reviewer" {
		t.Fatalf("GetAgentIdentity(reviewer) = %q, %q, %v; want \"\", reviewer, true", id, display, ok)
	}

	if _, _, ok := s.GetAgentIdentity("ghost"); ok {
		t.Fatal("GetAgentIdentity(ghost) = true; want false")
	}
}
