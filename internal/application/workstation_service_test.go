package application

import (
	"context"
	cryptokeymemory "go.orx.me/apps/butter/internal/repo/cryptokey/memory"
	"go.orx.me/apps/butter/internal/secretbox"
	workstationmemory "go.orx.me/apps/butter/internal/workstation/memory"
	"go.orx.me/apps/butter/internal/workstation/providerregistry"
	"testing"
)

func TestSetConfiguredServiceWiresRegistryBeforeHandlerUse(t *testing.T) {
	s := NewWorkStationServiceServer()
	err := s.SetConfiguredService(providerregistry.Config{Kind: providerregistry.GitHub}, providerregistry.ServiceDeps{Repo: workstationmemory.New(), Keyring: secretbox.NewKeyring(cryptokeymemory.New()), CheckHost: func(context.Context, string) error { return nil }}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if s.service == nil || s.service.Provider == nil {
		t.Fatal("configured provider was not wired")
	}
}
func TestSetConfiguredServiceRejectsProfileBeforeHandlerUse(t *testing.T) {
	s := NewWorkStationServiceServer()
	err := s.SetConfiguredService(providerregistry.Config{Kind: providerregistry.GitHub, Host: "https://evil.example"}, providerregistry.ServiceDeps{}, nil)
	if err == nil {
		t.Fatal("invalid profile accepted")
	}
	if s.service != nil {
		t.Fatal("invalid configuration mutated handler")
	}
}
