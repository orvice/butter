package config

import (
	"os"
	"testing"
	"time"

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

// The AG-UI maximum run duration loads from agui.max_run_duration and falls
// back to 30 minutes, as the async chat's does.
func TestAGUIMaxRunDuration(t *testing.T) {
	var cfg AppConfig
	if err := yaml.Unmarshal([]byte("agui:\n  max_run_duration: 45m\n"), &cfg); err != nil {
		t.Fatalf("unmarshal config: %v", err)
	}
	if got := cfg.AGUI.EffectiveMaxRunDuration(); got != 45*time.Minute {
		t.Fatalf("configured max run duration = %v, want 45m", got)
	}
	if got := (AGUIConfig{}).EffectiveMaxRunDuration(); got != 30*time.Minute {
		t.Fatalf("default max run duration = %v, want 30m", got)
	}
}

// The startup web-chat cleanup (#411) runs unless maintenance.delete_web_chat
// is false, and the sample config.yaml shows the flag, on.
func TestMaintenanceDeleteWebChatDefaultsToTrue(t *testing.T) {
	sample, err := os.ReadFile("../../config.yaml")
	if err != nil {
		t.Fatalf("read the sample config: %v", err)
	}
	for _, c := range []struct {
		name, doc string
		set, want bool
	}{
		{name: "absent", doc: "mongo_db: \"butter\"\n", want: true},
		{name: "empty section", doc: "maintenance: {}\n", want: true},
		{name: "true", doc: "maintenance:\n  delete_web_chat: true\n", set: true, want: true},
		{name: "false", doc: "maintenance:\n  delete_web_chat: false\n", set: true, want: false},
		{name: "the sample config.yaml", doc: string(sample), set: true, want: true},
	} {
		t.Run(c.name, func(t *testing.T) {
			var cfg AppConfig
			if err := yaml.Unmarshal([]byte(c.doc), &cfg); err != nil {
				t.Fatalf("unmarshal config: %v", err)
			}
			if set := cfg.Maintenance.DeleteWebChat != nil; set != c.set {
				t.Errorf("delete_web_chat set = %v, want %v", set, c.set)
			}
			if got := cfg.Maintenance.EffectiveDeleteWebChat(); got != c.want {
				t.Errorf("EffectiveDeleteWebChat() = %v, want %v", got, c.want)
			}
		})
	}
}
