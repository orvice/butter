package service

import (
	"testing"

	"go.orx.me/apps/butter/internal/config"
	"go.orx.me/apps/butter/internal/repo"
)

func TestHealthService_Ping(t *testing.T) {
	svc := NewHealthService(repo.NewHealthRepository(), &config.AppConfig{
		HTTP: config.HTTPConfig{Greeting: "hello"},
	})
	got := svc.Ping()
	if got.Service != "butter" {
		t.Errorf("Ping().Service = %q, want butter", got.Service)
	}
	if got.Message != "hello" {
		t.Errorf("Ping().Message = %q, want hello", got.Message)
	}
}
