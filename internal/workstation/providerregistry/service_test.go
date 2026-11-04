package providerregistry

import (
	"context"
	cryptokeymemory "go.orx.me/apps/butter/internal/repo/cryptokey/memory"
	"go.orx.me/apps/butter/internal/secretbox"
	ws "go.orx.me/apps/butter/internal/workstation"
	workstationmemory "go.orx.me/apps/butter/internal/workstation/memory"
	"testing"
)

func TestNewServiceWiresConfiguredProvider(t *testing.T) {
	s, e := NewService(Config{Kind: GitHub}, ServiceDeps{Repo: workstationmemory.New(), Keyring: secretbox.NewKeyring(cryptokeymemory.New()), CheckHost: func(context.Context, string) error { return nil }})
	if e != nil {
		t.Fatal(e)
	}
	if s.Provider == nil {
		t.Fatal("provider was not wired")
	}
}
func TestNewServiceMapsMissingDependencies(t *testing.T) {
	_, e := NewService(Config{Kind: GitHub}, ServiceDeps{})
	if e != ws.ErrUnsupported {
		t.Fatalf("error=%v", e)
	}
}
