package plane

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"go.orx.me/apps/butter/internal/runtime/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/genai"
)

var ErrPromptRequired = errors.New("plane activity prompt is required")
var ErrRunnerUnavailable = errors.New("plane runner is unavailable")
var ErrReplyUnavailable = errors.New("plane reply publisher is unavailable")

type Runner interface {
	Run(context.Context, string, []*genai.Part, string, *agentsv1.ContextInfo, runner.EventCallback, runner.CompactionCallback) (string, error)
}
type ReplyPublisher interface {
	PublishReply(context.Context, *Event, string) error
}
type Processor struct {
	runner    Runner
	publisher ReplyPublisher
	mu        sync.Mutex
	active    map[string]struct{}
	seen      map[string]struct{}
	maxReply  int
}

func NewProcessor(r Runner, p ReplyPublisher) *Processor {
	return &Processor{runner: r, publisher: p, active: map[string]struct{}{}, seen: map[string]struct{}{}, maxReply: 500}
}
func (p *Processor) Process(ctx context.Context, app *agentsv1.PlaneApp, e *Event) (string, error) {
	if p == nil || p.runner == nil {
		return "", ErrRunnerUnavailable
	}
	if p.publisher == nil {
		return "", ErrReplyUnavailable
	}
	if app == nil || app.GetAgentId() == "" {
		return "", errors.New("plane app agent is not configured")
	}
	prompt := strings.TrimSpace(e.Prompt)
	if prompt == "" {
		return "", ErrPromptRequired
	}
	key := e.InstallationID + ":" + e.AgentRunID + ":" + e.ActivityID
	if !p.begin(key) {
		return "", ErrDuplicate
	}
	defer p.end(key)
	out, err := p.runner.Run(ctx, app.GetAgentId(), []*genai.Part{{Text: prompt}}, "", &agentsv1.ContextInfo{UserId: e.BotUserID, SessionId: e.AgentRunID, ChannelName: "plane"}, nil, nil)
	if err != nil {
		return "", fmt.Errorf("plane runner: %w", err)
	}
	out = strings.TrimSpace(out)
	if out == "" {
		return "", errors.New("plane runner returned empty reply")
	}
	if len(out) > p.maxReply {
		out = out[:p.maxReply]
	}
	if err := p.publisher.PublishReply(ctx, e, out); err != nil {
		return "", fmt.Errorf("plane reply: %w", err)
	}
	p.mu.Lock()
	p.seen[key] = struct{}{}
	p.mu.Unlock()
	return out, nil
}
func (p *Processor) begin(k string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.seen[k]; ok {
		return false
	}
	if _, ok := p.active[k]; ok {
		return false
	}
	p.active[k] = struct{}{}
	return true
}
func (p *Processor) end(k string) { p.mu.Lock(); delete(p.active, k); p.mu.Unlock() }
