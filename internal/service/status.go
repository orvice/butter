package service

import (
	"context"
	"strings"

	appconfig "go.orx.me/apps/butter/internal/config"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
)

type ConfigStatusProvider interface {
	configrepo.AgentRepository
	configrepo.MCPServerRepository
	configrepo.RemoteAgentRepository
	configrepo.ChannelRepository
	ActiveBackendName() string
}

type StatusService struct {
	cfg   *appconfig.AppConfig
	store ConfigStatusProvider
}

type StatusResponse struct {
	Service string        `json:"service"`
	Storage StorageStatus `json:"storage"`
}

type StorageStatus struct {
	ConfiguredBackend string             `json:"configured_backend"`
	ActiveBackend     string             `json:"active_backend"`
	Persistent        bool               `json:"persistent"`
	MongoDatabase     string             `json:"mongo_database,omitempty"`
	Collections       ConfigEntityCounts `json:"collections"`
}

type ConfigEntityCounts struct {
	Agents       int `json:"agents"`
	MCPServers   int `json:"mcp_servers"`
	RemoteAgents int `json:"remote_agents"`
	Channels     int `json:"channels"`
}

func NewStatusService(cfg *appconfig.AppConfig, store ConfigStatusProvider) *StatusService {
	return &StatusService{
		cfg:   cfg,
		store: store,
	}
}

func (s *StatusService) Status(ctx context.Context) (StatusResponse, error) {
	agents, err := s.store.ListAgents(ctx)
	if err != nil {
		return StatusResponse{}, err
	}
	mcpServers, err := s.store.ListMCPServers(ctx)
	if err != nil {
		return StatusResponse{}, err
	}
	remoteAgents, err := s.store.ListRemoteAgents(ctx)
	if err != nil {
		return StatusResponse{}, err
	}
	channels, err := s.store.ListChannels(ctx)
	if err != nil {
		return StatusResponse{}, err
	}

	active := s.store.ActiveBackendName()
	storage := StorageStatus{
		ConfiguredBackend: configuredBackend(s.cfg),
		ActiveBackend:     active,
		Persistent:        active == "mongo",
		Collections: ConfigEntityCounts{
			Agents:       len(agents),
			MCPServers:   len(mcpServers),
			RemoteAgents: len(remoteAgents),
			Channels:     len(channels),
		},
	}
	if active == "mongo" || storage.ConfiguredBackend == "mongo" {
		storage.MongoDatabase = mongoDatabase(s.cfg)
	}

	return StatusResponse{
		Service: "butter",
		Storage: storage,
	}, nil
}

func configuredBackend(cfg *appconfig.AppConfig) string {
	if cfg == nil {
		return "memory"
	}
	backend := strings.TrimSpace(strings.ToLower(cfg.StorageBackend))
	if backend == "" {
		return "memory"
	}
	return backend
}

func mongoDatabase(cfg *appconfig.AppConfig) string {
	if cfg == nil || strings.TrimSpace(cfg.MongoDB) == "" {
		return "butter"
	}
	return cfg.MongoDB
}
