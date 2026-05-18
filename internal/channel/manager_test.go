package channel

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	configrepo "go.orx.me/apps/butter/internal/repo/config"
	"go.orx.me/apps/butter/internal/runtime/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type fakeChannelRepo struct {
	mu       sync.RWMutex
	channels []*agentsv1.AgentChannel
}

func (r *fakeChannelRepo) set(channels ...*agentsv1.AgentChannel) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.channels = channels
}

func (r *fakeChannelRepo) ListChannels(context.Context, string) ([]*agentsv1.AgentChannel, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*agentsv1.AgentChannel, len(r.channels))
	copy(out, r.channels)
	return out, nil
}

func (r *fakeChannelRepo) ListChannelsAcrossWorkspaces(context.Context) ([]*agentsv1.AgentChannel, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*agentsv1.AgentChannel, len(r.channels))
	copy(out, r.channels)
	return out, nil
}

func (r *fakeChannelRepo) GetChannel(context.Context, string, string) (*agentsv1.AgentChannel, error) {
	return nil, configrepo.ErrNotFound
}

func (r *fakeChannelRepo) CreateChannel(context.Context, string, *agentsv1.AgentChannel) (*agentsv1.AgentChannel, error) {
	return nil, nil
}

func (r *fakeChannelRepo) UpdateChannel(context.Context, string, *agentsv1.AgentChannel) (*agentsv1.AgentChannel, error) {
	return nil, nil
}

func (r *fakeChannelRepo) DeleteChannel(context.Context, string, string) error {
	return nil
}

type fakePoller struct {
	started chan string
	stopped chan string
	name    string
	webhook http.Handler
}

func (p *fakePoller) Start(ctx context.Context) {
	p.started <- p.name
	<-ctx.Done()
	p.stopped <- p.name
}

func (p *fakePoller) WebhookHandler() http.Handler {
	return p.webhook
}

func TestManagerReload(t *testing.T) {
	repo := &fakeChannelRepo{}
	repo.set(&agentsv1.AgentChannel{
		Name:     "ch-1",
		Enabled:  true,
		Platform: agentsv1.AgentChannelPlatform_AGENT_CHANNEL_PLATFORM_TELEGRAM,
		Telegram: &agentsv1.TelegramChannelConfig{BotToken: "token"},
	})

	started := make(chan string, 8)
	stopped := make(chan string, 8)
	manager := &Manager{
		repo:       repo,
		runnerSvc:  &runner.Service{},
		rdb:        redis.NewClient(&redis.Options{Addr: "localhost:6379"}),
		modelNames: []string{"test"},
		telegramFactory: func(ch *agentsv1.AgentChannel, _ *runner.Service, _ *redis.Client, _ []string, _ []string) (ChannelPoller, error) {
			return &fakePoller{name: ch.GetName(), started: started, stopped: stopped, webhook: http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})}, nil
		},
		discordFactory: func(ch *agentsv1.AgentChannel, _ *runner.Service, _ *redis.Client, _ []string, _ []string) (ChannelPoller, error) {
			return &fakePoller{name: ch.GetName(), started: started, stopped: stopped}, nil
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go manager.Start(ctx)

	waitForChannel(t, started, "ch-1")

	repo.set(&agentsv1.AgentChannel{
		Name:     "ch-2",
		Enabled:  true,
		Platform: agentsv1.AgentChannelPlatform_AGENT_CHANNEL_PLATFORM_TELEGRAM,
		Telegram: &agentsv1.TelegramChannelConfig{BotToken: "token"},
	})
	if err := manager.Reload(context.Background()); err != nil {
		t.Fatalf("reload failed: %v", err)
	}

	waitForChannel(t, stopped, "ch-1")
	waitForChannel(t, started, "ch-2")

	cancel()
	waitForChannel(t, stopped, "ch-2")
}

func waitForChannel(t *testing.T, ch <-chan string, want string) {
	t.Helper()
	select {
	case got := <-ch:
		if got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for %q", want)
	}
}

func TestManagerTelegramWebhookHandlerReload(t *testing.T) {
	repo := &fakeChannelRepo{}
	repo.set(&agentsv1.AgentChannel{
		Name:     "telegram-main",
		Enabled:  true,
		Platform: agentsv1.AgentChannelPlatform_AGENT_CHANNEL_PLATFORM_TELEGRAM,
		Telegram: &agentsv1.TelegramChannelConfig{
			BotToken:   "token",
			WebhookUrl: "https://example.com/webhooks/telegram/telegram-main",
		},
	})

	started := make(chan string, 8)
	stopped := make(chan string, 8)
	handlerA := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})
	handlerB := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	nextHandler := handlerA

	manager := &Manager{
		repo:       repo,
		runnerSvc:  &runner.Service{},
		rdb:        redis.NewClient(&redis.Options{Addr: "localhost:6379"}),
		modelNames: []string{"test"},
		telegramFactory: func(ch *agentsv1.AgentChannel, _ *runner.Service, _ *redis.Client, _ []string, _ []string) (ChannelPoller, error) {
			return &fakePoller{name: ch.GetName(), started: started, stopped: stopped, webhook: nextHandler}, nil
		},
		discordFactory: func(ch *agentsv1.AgentChannel, _ *runner.Service, _ *redis.Client, _ []string, _ []string) (ChannelPoller, error) {
			return &fakePoller{name: ch.GetName(), started: started, stopped: stopped}, nil
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go manager.Start(ctx)
	waitForChannel(t, started, "telegram-main")

	h, ok := manager.TelegramWebhookHandler("telegram-main")
	if !ok {
		t.Fatal("expected webhook handler")
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/webhooks/telegram/telegram-main", nil))
	if rec.Code != http.StatusAccepted {
		t.Fatalf("handler status = %d, want %d", rec.Code, http.StatusAccepted)
	}

	nextHandler = handlerB
	if err := manager.Reload(context.Background()); err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	waitForChannel(t, stopped, "telegram-main")
	waitForChannel(t, started, "telegram-main")

	h, ok = manager.TelegramWebhookHandler("telegram-main")
	if !ok {
		t.Fatal("expected webhook handler after reload")
	}
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/webhooks/telegram/telegram-main", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("handler status after reload = %d, want %d", rec.Code, http.StatusNoContent)
	}

	cancel()
	waitForChannel(t, stopped, "telegram-main")
}

// TestManagerScopesAgentsPerWorkspace verifies that buildPollers asks the
// runner for a per-channel workspace-scoped agent list instead of the global
// view — closes the tenant-boundary leak where workspace A's bot could see
// (and route to) workspace B's agents.
func TestManagerScopesAgentsPerWorkspace(t *testing.T) {
	repo := &fakeChannelRepo{}
	repo.set(
		&agentsv1.AgentChannel{
			Name:        "ch-a",
			WorkspaceId: "ws-a",
			Enabled:     true,
			Platform:    agentsv1.AgentChannelPlatform_AGENT_CHANNEL_PLATFORM_TELEGRAM,
			Telegram:    &agentsv1.TelegramChannelConfig{BotToken: "token-a"},
		},
		&agentsv1.AgentChannel{
			Name:        "ch-b",
			WorkspaceId: "ws-b",
			Enabled:     true,
			Platform:    agentsv1.AgentChannelPlatform_AGENT_CHANNEL_PLATFORM_TELEGRAM,
			Telegram:    &agentsv1.TelegramChannelConfig{BotToken: "token-b"},
		},
	)

	svc := runner.NewServiceForTest(map[string]string{
		"agent-a1": "ws-a",
		"agent-a2": "ws-a",
		"agent-b1": "ws-b",
	})

	type captured struct {
		channel string
		agents  []string
	}
	captures := make(chan captured, 4)

	manager := &Manager{
		repo:       repo,
		runnerSvc:  svc,
		rdb:        redis.NewClient(&redis.Options{Addr: "localhost:6379"}),
		modelNames: []string{"test"},
		telegramFactory: func(ch *agentsv1.AgentChannel, _ *runner.Service, _ *redis.Client, agentNames []string, _ []string) (ChannelPoller, error) {
			captures <- captured{channel: ch.GetName(), agents: agentNames}
			return &fakePoller{name: ch.GetName(), started: make(chan string, 1), stopped: make(chan string, 1)}, nil
		},
		discordFactory: func(ch *agentsv1.AgentChannel, _ *runner.Service, _ *redis.Client, _ []string, _ []string) (ChannelPoller, error) {
			return nil, nil
		},
	}

	if _, err := manager.buildPollers(context.Background()); err != nil {
		t.Fatalf("buildPollers: %v", err)
	}

	close(captures)
	got := map[string][]string{}
	for c := range captures {
		sortedAgents := append([]string(nil), c.agents...)
		// Stable comparison.
		for i := 1; i < len(sortedAgents); i++ {
			for j := i; j > 0 && sortedAgents[j-1] > sortedAgents[j]; j-- {
				sortedAgents[j-1], sortedAgents[j] = sortedAgents[j], sortedAgents[j-1]
			}
		}
		got[c.channel] = sortedAgents
	}

	wantA := []string{"agent-a1", "agent-a2"}
	wantB := []string{"agent-b1"}
	if !equalStringSlice(got["ch-a"], wantA) {
		t.Errorf("ch-a agents = %v, want %v", got["ch-a"], wantA)
	}
	if !equalStringSlice(got["ch-b"], wantB) {
		t.Errorf("ch-b agents = %v, want %v", got["ch-b"], wantB)
	}
}

func equalStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
