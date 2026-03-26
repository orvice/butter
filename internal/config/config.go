package config

import agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"

type AppConfig struct {
	Agents   []agentsv1.Agent
	Channels []agentsv1.AgentChannel
}

func (c *AppConfig) Print() {}
