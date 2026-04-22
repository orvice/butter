package app

import (
	"context"
	"errors"
	"testing"

	"go.orx.me/apps/butter/internal/config"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type fakeConfigRuntimeStore struct {
	syncCalls int
	syncErr   error
}

func (s *fakeConfigRuntimeStore) SyncToConfig(context.Context, *config.AppConfig) error {
	s.syncCalls++
	return s.syncErr
}

type fakeRunnerReloader struct {
	calls int
	err   error
}

func (r *fakeRunnerReloader) ReloadProtoAgents(context.Context, []agentsv1.Agent, []agentsv1.MCPServer, []agentsv1.RemoteAgent) error {
	r.calls++
	return r.err
}

type fakeChannelReloader struct {
	calls int
	err   error
}

func (r *fakeChannelReloader) Reload(context.Context) error {
	r.calls++
	return r.err
}

func TestConfigRuntimeReloadRunnerAlsoReloadsChannels(t *testing.T) {
	store := &fakeConfigRuntimeStore{}
	runnerSvc := &fakeRunnerReloader{}
	channelMgr := &fakeChannelReloader{}
	runtime := &ConfigRuntime{
		store:      store,
		cfg:        &config.AppConfig{},
		runnerSvc:  runnerSvc,
		channelMgr: channelMgr,
	}

	if err := runtime.ReloadRunner(context.Background()); err != nil {
		t.Fatalf("reload runner: %v", err)
	}

	if store.syncCalls != 1 {
		t.Fatalf("expected 1 sync call, got %d", store.syncCalls)
	}
	if runnerSvc.calls != 1 {
		t.Fatalf("expected 1 runner reload, got %d", runnerSvc.calls)
	}
	if channelMgr.calls != 1 {
		t.Fatalf("expected 1 channel reload, got %d", channelMgr.calls)
	}
}

func TestConfigRuntimeReloadRunnerReturnsChannelReloadError(t *testing.T) {
	channelMgr := &fakeChannelReloader{err: errors.New("boom")}
	runtime := &ConfigRuntime{
		store:      &fakeConfigRuntimeStore{},
		cfg:        &config.AppConfig{},
		runnerSvc:  &fakeRunnerReloader{},
		channelMgr: channelMgr,
	}

	err := runtime.ReloadRunner(context.Background())
	if err == nil || err.Error() != "boom" {
		t.Fatalf("expected channel reload error, got %v", err)
	}
}
