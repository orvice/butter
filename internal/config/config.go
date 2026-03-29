package config

import (
	"github.com/achetronic/adk-utils-go/plugin/langfuse"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type AppConfig struct {
	Agents           []agentsv1.Agent         `yaml:"agents"`
	Channels         []agentsv1.AgentChannel  `yaml:"channels"`
	ModelProviders   []agentsv1.ModelProvider  `yaml:"model_providers"`
	MCPServerConfigs []agentsv1.MCPServer      `yaml:"mcp_server_configs"`
	RemoteAgents     []agentsv1.RemoteAgent    `yaml:"remote_agents"`

	Langfuse langfuse.Config `yaml:"langfuse"`

	MongoURI      string `yaml:"mongo_uri"`
	MongoDB       string `yaml:"mongo_db"`
	RedisAddr     string `yaml:"redis_addr"`
	RedisPassword string `yaml:"redis_password"`

	HTTP HTTPConfig `yaml:"http"`
}

type HTTPConfig struct {
	Greeting string `yaml:"greeting"`
}

func (c *AppConfig) Print() {}
