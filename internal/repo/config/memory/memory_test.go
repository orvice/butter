package memory

import (
	"context"
	"errors"
	"testing"

	configrepo "go.orx.me/apps/butter/internal/repo/config"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestAgentCRUD(t *testing.T) {
	s := New()
	ctx := context.Background()

	// List empty
	agents, err := s.ListAgents(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(agents) != 0 {
		t.Fatalf("expected 0 agents, got %d", len(agents))
	}

	// Create
	a := &agentsv1.Agent{Name: "test-agent", Description: "desc"}
	created, err := s.CreateAgent(ctx, a)
	if err != nil {
		t.Fatal(err)
	}
	if created.GetName() != "test-agent" {
		t.Fatalf("expected name test-agent, got %s", created.GetName())
	}

	// Duplicate create
	_, err = s.CreateAgent(ctx, a)
	if !errors.Is(err, configrepo.ErrAlreadyExists) {
		t.Fatalf("expected ErrAlreadyExists, got %v", err)
	}

	// Get
	got, err := s.GetAgent(ctx, "test-agent")
	if err != nil {
		t.Fatal(err)
	}
	if got.GetDescription() != "desc" {
		t.Fatalf("expected desc, got %s", got.GetDescription())
	}

	// Get not found
	_, err = s.GetAgent(ctx, "nope")
	if !errors.Is(err, configrepo.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	// Update
	a.Description = "updated"
	updated, err := s.UpdateAgent(ctx, a)
	if err != nil {
		t.Fatal(err)
	}
	if updated.GetDescription() != "updated" {
		t.Fatalf("expected updated, got %s", updated.GetDescription())
	}

	// Update not found
	_, err = s.UpdateAgent(ctx, &agentsv1.Agent{Name: "nope"})
	if !errors.Is(err, configrepo.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	// List
	agents, err = s.ListAgents(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(agents) != 1 {
		t.Fatalf("expected 1 agent, got %d", len(agents))
	}

	// Delete
	if err := s.DeleteAgent(ctx, "test-agent"); err != nil {
		t.Fatal(err)
	}
	agents, _ = s.ListAgents(ctx)
	if len(agents) != 0 {
		t.Fatalf("expected 0 agents after delete, got %d", len(agents))
	}

	// Delete not found
	if err := s.DeleteAgent(ctx, "nope"); !errors.Is(err, configrepo.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestMCPServerCRUD(t *testing.T) {
	s := New()
	ctx := context.Background()

	m := &agentsv1.MCPServer{Id: "mcp1", Name: "test-mcp"}
	created, err := s.CreateMCPServer(ctx, m)
	if err != nil {
		t.Fatal(err)
	}
	if created.GetId() != "mcp1" {
		t.Fatalf("expected id mcp1, got %s", created.GetId())
	}

	_, err = s.CreateMCPServer(ctx, m)
	if !errors.Is(err, configrepo.ErrAlreadyExists) {
		t.Fatalf("expected ErrAlreadyExists, got %v", err)
	}

	got, err := s.GetMCPServer(ctx, "mcp1")
	if err != nil {
		t.Fatal(err)
	}
	if got.GetName() != "test-mcp" {
		t.Fatalf("expected test-mcp, got %s", got.GetName())
	}

	_, err = s.GetMCPServer(ctx, "nope")
	if !errors.Is(err, configrepo.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	m.Name = "updated-mcp"
	if _, err := s.UpdateMCPServer(ctx, m); err != nil {
		t.Fatal(err)
	}

	_, err = s.UpdateMCPServer(ctx, &agentsv1.MCPServer{Id: "nope"})
	if !errors.Is(err, configrepo.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	if err := s.DeleteMCPServer(ctx, "mcp1"); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteMCPServer(ctx, "nope"); !errors.Is(err, configrepo.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestRemoteAgentCRUD(t *testing.T) {
	s := New()
	ctx := context.Background()

	r := &agentsv1.RemoteAgent{Id: "ra1", Name: "test-ra", Url: "http://example.com"}
	created, err := s.CreateRemoteAgent(ctx, r)
	if err != nil {
		t.Fatal(err)
	}
	if created.GetId() != "ra1" {
		t.Fatalf("expected id ra1, got %s", created.GetId())
	}

	_, err = s.CreateRemoteAgent(ctx, r)
	if !errors.Is(err, configrepo.ErrAlreadyExists) {
		t.Fatalf("expected ErrAlreadyExists, got %v", err)
	}

	got, err := s.GetRemoteAgent(ctx, "ra1")
	if err != nil {
		t.Fatal(err)
	}
	if got.GetName() != "test-ra" {
		t.Fatalf("expected test-ra, got %s", got.GetName())
	}

	_, err = s.GetRemoteAgent(ctx, "nope")
	if !errors.Is(err, configrepo.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	r.Name = "updated-ra"
	if _, err := s.UpdateRemoteAgent(ctx, r); err != nil {
		t.Fatal(err)
	}

	_, err = s.UpdateRemoteAgent(ctx, &agentsv1.RemoteAgent{Id: "nope"})
	if !errors.Is(err, configrepo.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	if err := s.DeleteRemoteAgent(ctx, "ra1"); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteRemoteAgent(ctx, "nope"); !errors.Is(err, configrepo.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestChannelCRUD(t *testing.T) {
	s := New()
	ctx := context.Background()

	ch := &agentsv1.AgentChannel{Name: "tg1", AgentName: "agent1"}
	created, err := s.CreateChannel(ctx, ch)
	if err != nil {
		t.Fatal(err)
	}
	if created.GetName() != "tg1" {
		t.Fatalf("expected name tg1, got %s", created.GetName())
	}

	_, err = s.CreateChannel(ctx, ch)
	if !errors.Is(err, configrepo.ErrAlreadyExists) {
		t.Fatalf("expected ErrAlreadyExists, got %v", err)
	}

	got, err := s.GetChannel(ctx, "tg1")
	if err != nil {
		t.Fatal(err)
	}
	if got.GetAgentName() != "agent1" {
		t.Fatalf("expected agent1, got %s", got.GetAgentName())
	}

	_, err = s.GetChannel(ctx, "nope")
	if !errors.Is(err, configrepo.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	ch.AgentName = "agent2"
	updated, err := s.UpdateChannel(ctx, ch)
	if err != nil {
		t.Fatal(err)
	}
	if updated.GetAgentName() != "agent2" {
		t.Fatalf("expected agent2, got %s", updated.GetAgentName())
	}

	_, err = s.UpdateChannel(ctx, &agentsv1.AgentChannel{Name: "nope"})
	if !errors.Is(err, configrepo.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	if err := s.DeleteChannel(ctx, "tg1"); err != nil {
		t.Fatal(err)
	}
	channels, _ := s.ListChannels(ctx)
	if len(channels) != 0 {
		t.Fatalf("expected 0 channels after delete, got %d", len(channels))
	}

	if err := s.DeleteChannel(ctx, "nope"); !errors.Is(err, configrepo.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestSeed(t *testing.T) {
	s := New()
	ctx := context.Background()
	s.Seed(ctx,
		[]agentsv1.Agent{{Name: "a1"}, {Name: "a2"}},
		[]agentsv1.MCPServer{{Id: "m1", Name: "mcp1"}},
		[]agentsv1.RemoteAgent{{Id: "r1", Name: "ra1", Url: "http://example.com"}},
		[]agentsv1.AgentChannel{{Name: "ch1", AgentName: "a1"}},
	)

	agents, _ := s.ListAgents(ctx)
	if len(agents) != 2 {
		t.Fatalf("expected 2 agents, got %d", len(agents))
	}
	mcps, _ := s.ListMCPServers(ctx)
	if len(mcps) != 1 {
		t.Fatalf("expected 1 mcp server, got %d", len(mcps))
	}
	ras, _ := s.ListRemoteAgents(ctx)
	if len(ras) != 1 {
		t.Fatalf("expected 1 remote agent, got %d", len(ras))
	}
	chs, _ := s.ListChannels(ctx)
	if len(chs) != 1 {
		t.Fatalf("expected 1 channel, got %d", len(chs))
	}
}
