package channel

import (
	"context"
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

func (r *fakeChannelRepo) ListChannels(context.Context) ([]*agentsv1.AgentChannel, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*agentsv1.AgentChannel, len(r.channels))
	copy(out, r.channels)
	return out, nil
}

func (r *fakeChannelRepo) GetChannel(context.Context, string) (*agentsv1.AgentChannel, error) {
	return nil, configrepo.ErrNotFound
}

func (r *fakeChannelRepo) CreateChannel(context.Context, *agentsv1.AgentChannel) (*agentsv1.AgentChannel, error) {
	return nil, nil
}

func (r *fakeChannelRepo) UpdateChannel(context.Context, *agentsv1.AgentChannel) (*agentsv1.AgentChannel, error) {
	return nil, nil
}

func (r *fakeChannelRepo) DeleteChannel(context.Context, string) error {
	return nil
}

type fakePoller struct {
	started chan string
	stopped chan string
	name    string
}

func (p *fakePoller) Start(ctx context.Context) {
	p.started <- p.name
	<-ctx.Done()
	p.stopped <- p.name
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
			return &fakePoller{name: ch.GetName(), started: started, stopped: stopped}, nil
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
