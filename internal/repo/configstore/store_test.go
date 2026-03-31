package configstore

import (
	"testing"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestAgentCRUD(t *testing.T) {
	s := New()

	// List empty
	if got := s.ListAgents(); len(got) != 0 {
		t.Fatalf("expected 0 agents, got %d", len(got))
	}

	// Create
	a := &agentsv1.Agent{Name: "test-agent", Description: "desc"}
	created, err := s.CreateAgent(a)
	if err != nil {
		t.Fatal(err)
	}
	if created.GetName() != "test-agent" {
		t.Fatalf("expected name test-agent, got %s", created.GetName())
	}

	// Duplicate create
	if _, err := s.CreateAgent(a); err == nil {
		t.Fatal("expected error on duplicate create")
	}

	// Get
	got, err := s.GetAgent("test-agent")
	if err != nil {
		t.Fatal(err)
	}
	if got.GetDescription() != "desc" {
		t.Fatalf("expected desc, got %s", got.GetDescription())
	}

	// Get not found
	if _, err := s.GetAgent("nope"); err == nil {
		t.Fatal("expected error on not found")
	}

	// Update
	a.Description = "updated"
	updated, err := s.UpdateAgent(a)
	if err != nil {
		t.Fatal(err)
	}
	if updated.GetDescription() != "updated" {
		t.Fatalf("expected updated, got %s", updated.GetDescription())
	}

	// Update not found
	if _, err := s.UpdateAgent(&agentsv1.Agent{Name: "nope"}); err == nil {
		t.Fatal("expected error on not found update")
	}

	// List
	if got := s.ListAgents(); len(got) != 1 {
		t.Fatalf("expected 1 agent, got %d", len(got))
	}

	// Delete
	if err := s.DeleteAgent("test-agent"); err != nil {
		t.Fatal(err)
	}
	if got := s.ListAgents(); len(got) != 0 {
		t.Fatalf("expected 0 agents after delete, got %d", len(got))
	}

	// Delete not found
	if err := s.DeleteAgent("nope"); err == nil {
		t.Fatal("expected error on not found delete")
	}
}

func TestMCPServerCRUD(t *testing.T) {
	s := New()

	m := &agentsv1.MCPServer{Id: "mcp1", Name: "test-mcp"}
	created, err := s.CreateMCPServer(m)
	if err != nil {
		t.Fatal(err)
	}
	if created.GetId() != "mcp1" {
		t.Fatalf("expected id mcp1, got %s", created.GetId())
	}

	if _, err := s.CreateMCPServer(m); err == nil {
		t.Fatal("expected error on duplicate")
	}

	got, err := s.GetMCPServer("mcp1")
	if err != nil {
		t.Fatal(err)
	}
	if got.GetName() != "test-mcp" {
		t.Fatalf("expected test-mcp, got %s", got.GetName())
	}

	if _, err := s.GetMCPServer("nope"); err == nil {
		t.Fatal("expected error on not found")
	}

	m.Name = "updated-mcp"
	if _, err := s.UpdateMCPServer(m); err != nil {
		t.Fatal(err)
	}

	if _, err := s.UpdateMCPServer(&agentsv1.MCPServer{Id: "nope"}); err == nil {
		t.Fatal("expected error on not found update")
	}

	if err := s.DeleteMCPServer("mcp1"); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteMCPServer("nope"); err == nil {
		t.Fatal("expected error on not found delete")
	}
}

func TestRemoteAgentCRUD(t *testing.T) {
	s := New()

	r := &agentsv1.RemoteAgent{Id: "ra1", Name: "test-ra", Url: "http://example.com"}
	created, err := s.CreateRemoteAgent(r)
	if err != nil {
		t.Fatal(err)
	}
	if created.GetId() != "ra1" {
		t.Fatalf("expected id ra1, got %s", created.GetId())
	}

	if _, err := s.CreateRemoteAgent(r); err == nil {
		t.Fatal("expected error on duplicate")
	}

	got, err := s.GetRemoteAgent("ra1")
	if err != nil {
		t.Fatal(err)
	}
	if got.GetName() != "test-ra" {
		t.Fatalf("expected test-ra, got %s", got.GetName())
	}

	if _, err := s.GetRemoteAgent("nope"); err == nil {
		t.Fatal("expected error on not found")
	}

	r.Name = "updated-ra"
	if _, err := s.UpdateRemoteAgent(r); err != nil {
		t.Fatal(err)
	}

	if _, err := s.UpdateRemoteAgent(&agentsv1.RemoteAgent{Id: "nope"}); err == nil {
		t.Fatal("expected error on not found update")
	}

	if err := s.DeleteRemoteAgent("ra1"); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteRemoteAgent("nope"); err == nil {
		t.Fatal("expected error on not found delete")
	}
}

func TestSeed(t *testing.T) {
	s := New()
	s.Seed(
		[]agentsv1.Agent{{Name: "a1"}, {Name: "a2"}},
		[]agentsv1.MCPServer{{Id: "m1", Name: "mcp1"}},
		[]agentsv1.RemoteAgent{{Id: "r1", Name: "ra1", Url: "http://example.com"}},
	)

	if got := s.ListAgents(); len(got) != 2 {
		t.Fatalf("expected 2 agents, got %d", len(got))
	}
	if got := s.ListMCPServers(); len(got) != 1 {
		t.Fatalf("expected 1 mcp server, got %d", len(got))
	}
	if got := s.ListRemoteAgents(); len(got) != 1 {
		t.Fatalf("expected 1 remote agent, got %d", len(got))
	}
}
