package config

import (
	"time"

	"github.com/achetronic/adk-utils-go/plugin/langfuse"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type AppConfig struct {
	Agents           []agentsv1.Agent         `yaml:"agents"`
	Channels         []agentsv1.AgentChannel  `yaml:"channels"`
	ModelProviders   []agentsv1.ModelProvider `yaml:"model_providers"`
	MCPServerConfigs []agentsv1.MCPServer     `yaml:"mcp_server_configs"`
	RemoteAgents     []agentsv1.RemoteAgent   `yaml:"remote_agents"`
	APIToken         string                   `yaml:"apiToken"`
	Auth             AuthConfig               `yaml:"auth"`
	SystemAgentModel string                   `yaml:"system_agent_model"`

	Langfuse langfuse.Config `yaml:"langfuse"`

	MongoURI      string `yaml:"mongo_uri"`
	MongoDB       string `yaml:"mongo_db"`
	RedisAddr     string `yaml:"redis_addr"`
	RedisPassword string `yaml:"redis_password"`

	HTTP           HTTPConfig `yaml:"http"`
	StorageBackend string     `yaml:"storage_backend"` // "mongo" (default) or "memory"
	GRPCPort       int        `yaml:"grpc_port"`       // daemon gRPC server port (default 9090)
}

type HTTPConfig struct {
	Greeting string `yaml:"greeting"`
}

type AuthConfig struct {
	InitialAdminUsername string        `yaml:"initial_admin_username"`
	InitialAdminPassword string        `yaml:"initial_admin_password"`
	SessionTTL           time.Duration `yaml:"session_ttl"`
}

func (c AuthConfig) EffectiveSessionTTL() time.Duration {
	if c.SessionTTL <= 0 {
		return 7 * 24 * time.Hour
	}
	return c.SessionTTL
}

func (c *AppConfig) Print() {}
