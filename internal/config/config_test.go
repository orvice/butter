package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

// TestRuntimeSnapshotFieldsAreNotYAMLInputs locks in that the runtime snapshot
// fields cannot be populated from the config file. They are rebuilt from the
// config store on every reload (app.ConfigStore.SyncToConfig), so a YAML value
// would be silently discarded before any reader observes it — a config file
// that looks like it defines agents but does nothing. Keeping them excluded
// from unmarshalling makes that inability explicit at the seam.
func TestRuntimeSnapshotFieldsAreNotYAMLInputs(t *testing.T) {
	const doc = `
mongo_db: "butter-test"
agents:
  - name: yaml-agent
    type: 1
model_providers:
  - name: yaml-provider
    type: openai
mcp_server_configs:
  - id: yaml-mcp
    name: yaml-mcp
remote_agents:
  - id: yaml-remote
    name: yaml-remote
channels:
  - name: yaml-channel
`

	var cfg AppConfig
	if err := yaml.Unmarshal([]byte(doc), &cfg); err != nil {
		t.Fatalf("unmarshal config: %v", err)
	}

	// A real key still loads, so the assertions below exercise the tags rather
	// than a failed parse.
	if cfg.MongoDB != "butter-test" {
		t.Fatalf("expected mongo_db to load, got %q", cfg.MongoDB)
	}

	if n := len(cfg.Agents); n != 0 {
		t.Errorf("expected agents to be ignored, got %d", n)
	}
	if n := len(cfg.ModelProviders); n != 0 {
		t.Errorf("expected model_providers to be ignored, got %d", n)
	}
	if n := len(cfg.MCPServerConfigs); n != 0 {
		t.Errorf("expected mcp_server_configs to be ignored, got %d", n)
	}
	if n := len(cfg.RemoteAgents); n != 0 {
		t.Errorf("expected remote_agents to be ignored, got %d", n)
	}
	if n := len(cfg.Channels); n != 0 {
		t.Errorf("expected channels to be ignored, got %d", n)
	}
}
