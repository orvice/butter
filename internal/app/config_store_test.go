package app

import (
	"context"
	"testing"

	"go.orx.me/apps/butter/internal/config"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestConfigStoreRuntimeConfigComesFromStoreNotYAML(t *testing.T) {
	ctx := context.Background()
	cfg := &config.AppConfig{
		StorageBackend: "memory",
		Agents: []agentsv1.Agent{{
			Name: "yaml-agent",
		}},
		MCPServerConfigs: []agentsv1.MCPServer{{
			Id:   "yaml-mcp",
			Name: "yaml-mcp",
		}},
		RemoteAgents: []agentsv1.RemoteAgent{{
			Id:   "yaml-remote",
			Name: "yaml-remote",
		}},
		Channels: []agentsv1.AgentChannel{{
			Name: "yaml-channel",
		}},
		ModelProviders: []agentsv1.ModelProvider{{
			Name: "yaml-provider",
			Type: "openai",
		}},
	}
	store := NewConfigStore()

	if err := store.InitFromConfig(ctx, cfg); err != nil {
		t.Fatalf("init config store: %v", err)
	}
	if len(cfg.Agents) != 0 || len(cfg.MCPServerConfigs) != 0 || len(cfg.RemoteAgents) != 0 || len(cfg.Channels) != 0 || len(cfg.ModelProviders) != 0 {
		t.Fatalf("expected yaml runtime config to be ignored, got agents=%d mcp=%d remote=%d channels=%d providers=%d",
			len(cfg.Agents), len(cfg.MCPServerConfigs), len(cfg.RemoteAgents), len(cfg.Channels), len(cfg.ModelProviders))
	}

	if _, err := store.CreateAgent(ctx, "ws-test", &agentsv1.Agent{Name: "db-agent"}); err != nil {
		t.Fatalf("create agent: %v", err)
	}
	if _, err := store.CreateMCPServer(ctx, "ws-test", &agentsv1.MCPServer{Id: "db-mcp", Name: "db-mcp"}); err != nil {
		t.Fatalf("create mcp server: %v", err)
	}
	if _, err := store.CreateRemoteAgent(ctx, "ws-test", &agentsv1.RemoteAgent{Id: "db-remote", Name: "db-remote"}); err != nil {
		t.Fatalf("create remote agent: %v", err)
	}
	if _, err := store.CreateChannel(ctx, "ws-test", &agentsv1.AgentChannel{Name: "db-channel"}); err != nil {
		t.Fatalf("create channel: %v", err)
	}
	if _, err := store.CreateModelProvider(ctx, "ws-test", &agentsv1.ModelProvider{
		Name:   "db-provider",
		Type:   "openai",
		Models: []*agentsv1.ModelConfig{{Name: "gpt-4o", Alias: "4o"}},
	}); err != nil {
		t.Fatalf("create model provider: %v", err)
	}

	if err := store.SyncToConfig(ctx, cfg); err != nil {
		t.Fatalf("sync to config: %v", err)
	}
	if len(cfg.Agents) != 1 || cfg.Agents[0].GetName() != "db-agent" {
		t.Fatalf("expected db agent synced into config, got %+v", cfg.Agents)
	}
	if len(cfg.MCPServerConfigs) != 1 || cfg.MCPServerConfigs[0].GetId() != "db-mcp" {
		t.Fatalf("expected db mcp synced into config, got %+v", cfg.MCPServerConfigs)
	}
	if len(cfg.RemoteAgents) != 1 || cfg.RemoteAgents[0].GetId() != "db-remote" {
		t.Fatalf("expected db remote agent synced into config, got %+v", cfg.RemoteAgents)
	}
	if len(cfg.Channels) != 1 || cfg.Channels[0].GetName() != "db-channel" {
		t.Fatalf("expected db channel synced into config, got %+v", cfg.Channels)
	}
	if len(cfg.ModelProviders) != 1 || cfg.ModelProviders[0].GetName() != "db-provider" {
		t.Fatalf("expected db provider synced into config, got %+v", cfg.ModelProviders)
	}
}
