package app

import (
	"context"
	"errors"
	"strings"
	"testing"

	"go.orx.me/apps/butter/internal/agentcontent"
	"go.orx.me/apps/butter/internal/config"
	agentcontentmemory "go.orx.me/apps/butter/internal/repo/agentcontent/memory"
	repobindingmemory "go.orx.me/apps/butter/internal/repo/repobinding/memory"
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
	calls  int
	err    error
	agents []agentsv1.Agent
}

func (r *fakeRunnerReloader) ReloadProtoAgents(_ context.Context, agents []agentsv1.Agent, _ []agentsv1.ModelProvider, _ []agentsv1.MCPServer, _ []agentsv1.RemoteAgent) error {
	r.calls++
	r.agents = agents
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

func TestConfigRuntimeLoadsExactActiveContentRevision(t *testing.T) {
	ctx := context.Background()
	bindingRepo := repobindingmemory.New()
	if _, err := bindingRepo.Put(ctx, "ws-a", &agentsv1.WorkspaceRepoBinding{ActiveCommitSha: "active-sha"}); err != nil {
		t.Fatal(err)
	}
	contentRepo := agentcontentmemory.New()
	if err := contentRepo.PutSnapshot(ctx, "ws-a", agentcontent.Snapshot{
		CommitSHA: "inactive-sha",
		Entries: map[string]agentcontent.AgentContent{
			"agent-1": {AgentID: "agent-1", Instruction: "inactive"},
		},
	}); err != nil {
		t.Fatal(err)
	}
	if err := contentRepo.PutSnapshot(ctx, "ws-a", agentcontent.Snapshot{
		CommitSHA: "active-sha",
		Entries: map[string]agentcontent.AgentContent{
			"agent-1": {AgentID: "agent-1", Instruction: "active"},
		},
	}); err != nil {
		t.Fatal(err)
	}
	runnerSvc := &fakeRunnerReloader{}
	runtime := &ConfigRuntime{
		store:       &fakeConfigRuntimeStore{},
		cfg:         &config.AppConfig{Agents: []agentsv1.Agent{{AgentId: "agent-1", WorkspaceId: "ws-a"}}},
		runnerSvc:   runnerSvc,
		bindingRepo: bindingRepo,
		contentRepo: contentRepo,
	}
	if err := runtime.ReloadRunner(ctx); err != nil {
		t.Fatalf("ReloadRunner: %v", err)
	}
	if got := runnerSvc.agents[0].GetConfig().GetInstruction(); got != "active" {
		t.Fatalf("runner instruction = %q, want active revision", got)
	}
}

func TestConfigRuntimeRejectsMissingActiveContentRevision(t *testing.T) {
	ctx := context.Background()
	bindingRepo := repobindingmemory.New()
	if _, err := bindingRepo.Put(ctx, "ws-a", &agentsv1.WorkspaceRepoBinding{ActiveCommitSha: "missing-sha"}); err != nil {
		t.Fatal(err)
	}
	runtime := &ConfigRuntime{
		store:       &fakeConfigRuntimeStore{},
		cfg:         &config.AppConfig{Agents: []agentsv1.Agent{{AgentId: "agent-1", WorkspaceId: "ws-a"}}},
		bindingRepo: bindingRepo,
		contentRepo: agentcontentmemory.New(),
	}
	err := runtime.ReloadRunner(ctx)
	if err == nil || !strings.Contains(err.Error(), "missing-sha") {
		t.Fatalf("ReloadRunner error = %v, want missing Active Revision", err)
	}
}
