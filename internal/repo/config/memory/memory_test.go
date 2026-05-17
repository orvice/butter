package memory

import (
	"context"
	"errors"
	"testing"

	configrepo "go.orx.me/apps/butter/internal/repo/config"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const wsTest = "ws-test"

func TestAgentCRUD(t *testing.T) {
	s := New()
	ctx := context.Background()

	// List empty
	agents, err := s.ListAgents(ctx, wsTest)
	if err != nil {
		t.Fatal(err)
	}
	if len(agents) != 0 {
		t.Fatalf("expected 0 agents, got %d", len(agents))
	}

	// Create
	a := &agentsv1.Agent{Name: "test-agent", Description: "desc"}
	created, err := s.CreateAgent(ctx, wsTest, a)
	if err != nil {
		t.Fatal(err)
	}
	if created.GetName() != "test-agent" {
		t.Fatalf("expected name test-agent, got %s", created.GetName())
	}

	// Duplicate create
	_, err = s.CreateAgent(ctx, wsTest, a)
	if !errors.Is(err, configrepo.ErrAlreadyExists) {
		t.Fatalf("expected ErrAlreadyExists, got %v", err)
	}

	// Get
	got, err := s.GetAgent(ctx, wsTest, "test-agent")
	if err != nil {
		t.Fatal(err)
	}
	if got.GetDescription() != "desc" {
		t.Fatalf("expected desc, got %s", got.GetDescription())
	}

	// Get not found
	_, err = s.GetAgent(ctx, wsTest, "nope")
	if !errors.Is(err, configrepo.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	// Update
	a.Description = "updated"
	updated, err := s.UpdateAgent(ctx, wsTest, a)
	if err != nil {
		t.Fatal(err)
	}
	if updated.GetDescription() != "updated" {
		t.Fatalf("expected updated, got %s", updated.GetDescription())
	}

	// Update not found
	_, err = s.UpdateAgent(ctx, wsTest, &agentsv1.Agent{Name: "nope"})
	if !errors.Is(err, configrepo.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	// List
	agents, err = s.ListAgents(ctx, wsTest)
	if err != nil {
		t.Fatal(err)
	}
	if len(agents) != 1 {
		t.Fatalf("expected 1 agent, got %d", len(agents))
	}

	// Delete
	if err := s.DeleteAgent(ctx, wsTest, "test-agent"); err != nil {
		t.Fatal(err)
	}
	agents, _ = s.ListAgents(ctx, wsTest)
	if len(agents) != 0 {
		t.Fatalf("expected 0 agents after delete, got %d", len(agents))
	}

	// Delete not found
	if err := s.DeleteAgent(ctx, wsTest, "nope"); !errors.Is(err, configrepo.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestMCPServerCRUD(t *testing.T) {
	s := New()
	ctx := context.Background()

	m := &agentsv1.MCPServer{Id: "mcp1", Name: "test-mcp"}
	created, err := s.CreateMCPServer(ctx, wsTest, m)
	if err != nil {
		t.Fatal(err)
	}
	if created.GetId() != "mcp1" {
		t.Fatalf("expected id mcp1, got %s", created.GetId())
	}

	_, err = s.CreateMCPServer(ctx, wsTest, m)
	if !errors.Is(err, configrepo.ErrAlreadyExists) {
		t.Fatalf("expected ErrAlreadyExists, got %v", err)
	}

	got, err := s.GetMCPServer(ctx, wsTest, "mcp1")
	if err != nil {
		t.Fatal(err)
	}
	if got.GetName() != "test-mcp" {
		t.Fatalf("expected test-mcp, got %s", got.GetName())
	}

	_, err = s.GetMCPServer(ctx, wsTest, "nope")
	if !errors.Is(err, configrepo.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	m.Name = "updated-mcp"
	if _, err := s.UpdateMCPServer(ctx, wsTest, m); err != nil {
		t.Fatal(err)
	}

	_, err = s.UpdateMCPServer(ctx, wsTest, &agentsv1.MCPServer{Id: "nope"})
	if !errors.Is(err, configrepo.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	if err := s.DeleteMCPServer(ctx, wsTest, "mcp1"); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteMCPServer(ctx, wsTest, "nope"); !errors.Is(err, configrepo.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestRemoteAgentCRUD(t *testing.T) {
	s := New()
	ctx := context.Background()

	r := &agentsv1.RemoteAgent{Id: "ra1", Name: "test-ra", Url: "http://example.com"}
	created, err := s.CreateRemoteAgent(ctx, wsTest, r)
	if err != nil {
		t.Fatal(err)
	}
	if created.GetId() != "ra1" {
		t.Fatalf("expected id ra1, got %s", created.GetId())
	}

	_, err = s.CreateRemoteAgent(ctx, wsTest, r)
	if !errors.Is(err, configrepo.ErrAlreadyExists) {
		t.Fatalf("expected ErrAlreadyExists, got %v", err)
	}

	got, err := s.GetRemoteAgent(ctx, wsTest, "ra1")
	if err != nil {
		t.Fatal(err)
	}
	if got.GetName() != "test-ra" {
		t.Fatalf("expected test-ra, got %s", got.GetName())
	}

	_, err = s.GetRemoteAgent(ctx, wsTest, "nope")
	if !errors.Is(err, configrepo.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	r.Name = "updated-ra"
	if _, err := s.UpdateRemoteAgent(ctx, wsTest, r); err != nil {
		t.Fatal(err)
	}

	_, err = s.UpdateRemoteAgent(ctx, wsTest, &agentsv1.RemoteAgent{Id: "nope"})
	if !errors.Is(err, configrepo.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	if err := s.DeleteRemoteAgent(ctx, wsTest, "ra1"); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteRemoteAgent(ctx, wsTest, "nope"); !errors.Is(err, configrepo.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestChannelCRUD(t *testing.T) {
	s := New()
	ctx := context.Background()

	ch := &agentsv1.AgentChannel{Name: "tg1", AgentName: "agent1"}
	created, err := s.CreateChannel(ctx, wsTest, ch)
	if err != nil {
		t.Fatal(err)
	}
	if created.GetName() != "tg1" {
		t.Fatalf("expected name tg1, got %s", created.GetName())
	}

	_, err = s.CreateChannel(ctx, wsTest, ch)
	if !errors.Is(err, configrepo.ErrAlreadyExists) {
		t.Fatalf("expected ErrAlreadyExists, got %v", err)
	}

	got, err := s.GetChannel(ctx, wsTest, "tg1")
	if err != nil {
		t.Fatal(err)
	}
	if got.GetAgentName() != "agent1" {
		t.Fatalf("expected agent1, got %s", got.GetAgentName())
	}

	_, err = s.GetChannel(ctx, wsTest, "nope")
	if !errors.Is(err, configrepo.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	ch.AgentName = "agent2"
	updated, err := s.UpdateChannel(ctx, wsTest, ch)
	if err != nil {
		t.Fatal(err)
	}
	if updated.GetAgentName() != "agent2" {
		t.Fatalf("expected agent2, got %s", updated.GetAgentName())
	}

	_, err = s.UpdateChannel(ctx, wsTest, &agentsv1.AgentChannel{Name: "nope"})
	if !errors.Is(err, configrepo.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	if err := s.DeleteChannel(ctx, wsTest, "tg1"); err != nil {
		t.Fatal(err)
	}
	channels, _ := s.ListChannels(ctx, wsTest)
	if len(channels) != 0 {
		t.Fatalf("expected 0 channels after delete, got %d", len(channels))
	}

	if err := s.DeleteChannel(ctx, wsTest, "nope"); !errors.Is(err, configrepo.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestModelProviderCRUD(t *testing.T) {
	s := New()
	ctx := context.Background()

	provider := &agentsv1.ModelProvider{
		Name:   "openai",
		Type:   "openai",
		Models: []*agentsv1.ModelConfig{{Name: "gpt-4o", Alias: "4o"}},
	}
	created, err := s.CreateModelProvider(ctx, wsTest, provider)
	if err != nil {
		t.Fatal(err)
	}
	if created.GetName() != "openai" {
		t.Fatalf("expected name openai, got %s", created.GetName())
	}

	_, err = s.CreateModelProvider(ctx, wsTest, provider)
	if !errors.Is(err, configrepo.ErrAlreadyExists) {
		t.Fatalf("expected ErrAlreadyExists, got %v", err)
	}

	got, err := s.GetModelProvider(ctx, wsTest, "openai")
	if err != nil {
		t.Fatal(err)
	}
	if got.GetType() != "openai" {
		t.Fatalf("expected openai type, got %s", got.GetType())
	}

	_, err = s.GetModelProvider(ctx, wsTest, "nope")
	if !errors.Is(err, configrepo.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	provider.Type = "gemini"
	updated, err := s.UpdateModelProvider(ctx, wsTest, provider)
	if err != nil {
		t.Fatal(err)
	}
	if updated.GetType() != "gemini" {
		t.Fatalf("expected gemini type, got %s", updated.GetType())
	}

	_, err = s.UpdateModelProvider(ctx, wsTest, &agentsv1.ModelProvider{Name: "nope"})
	if !errors.Is(err, configrepo.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	providers, err := s.ListModelProviders(ctx, wsTest)
	if err != nil {
		t.Fatal(err)
	}
	if len(providers) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(providers))
	}

	if err := s.DeleteModelProvider(ctx, wsTest, "openai"); err != nil {
		t.Fatal(err)
	}
	providers, _ = s.ListModelProviders(ctx, wsTest)
	if len(providers) != 0 {
		t.Fatalf("expected 0 providers after delete, got %d", len(providers))
	}

	if err := s.DeleteModelProvider(ctx, wsTest, "nope"); !errors.Is(err, configrepo.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
