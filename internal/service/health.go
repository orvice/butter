package service

import (
	appconfig "go.orx.me/apps/butter/internal/config"
	"go.orx.me/apps/butter/internal/repo"
)

type HealthService struct {
	repo *repo.HealthRepository
	cfg  *appconfig.AppConfig
}

type PingResponse struct {
	Service string `json:"service"`
	Message string `json:"message"`
}

func NewHealthService(repo *repo.HealthRepository, cfg *appconfig.AppConfig) *HealthService {
	return &HealthService{
		repo: repo,
		cfg:  cfg,
	}
}

func (s *HealthService) Ping() PingResponse {
	return PingResponse{
		Service: s.repo.ServiceName(),
		Message: s.cfg.HTTP.Greeting,
	}
}
