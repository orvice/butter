package config

import agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"

type AppConfig struct {
	Agents   []agentsv1.Agent        `yaml:"agents"`
	Channels []agentsv1.AgentChannel `yaml:"channels"`

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
