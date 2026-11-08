package plane

import (
	"context"
	"errors"
	"go.orx.me/apps/butter/internal/runtime/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/genai"
	"sync"
	"testing"
)

type fakeRunner struct {
	mu    sync.Mutex
	calls int
	out   string
	err   error
	block chan struct{}
}

func (f *fakeRunner) Run(ctx context.Context, _ string, p []*genai.Part, _ string, _ *agentsv1.ContextInfo, _ runner.EventCallback, _ runner.CompactionCallback) (string, error) {
	f.mu.Lock()
	f.calls++
	f.mu.Unlock()
	if f.block != nil {
		select {
		case <-f.block:
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	return f.out, f.err
}

type fakePublisher struct {
	mu    sync.Mutex
	calls int
	reply string
}

func (f *fakePublisher) PublishReply(_ context.Context, _ *Event, s string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	f.reply = s
	return nil
}
func TestProcessorValidPromptOnceAndShortReply(t *testing.T) {
	r := &fakeRunner{out: "  answer  "}
	pub := &fakePublisher{}
	p := NewProcessor(r, pub)
	e := &Event{InstallationID: "i", AgentRunID: "r", ActivityID: "a", Prompt: " hello "}
	out, err := p.Process(context.Background(), &agentsv1.PlaneApp{AgentId: "agent"}, e)
	if err != nil || out != "answer" || pub.calls != 1 || r.calls != 1 {
		t.Fatalf("out=%q err=%v calls=%d/%d", out, err, r.calls, pub.calls)
	}
	if _, err = p.Process(context.Background(), &agentsv1.PlaneApp{AgentId: "agent"}, e); !errors.Is(err, ErrDuplicate) {
		t.Fatalf("duplicate err=%v", err)
	}
}
func TestProcessorRejectsMissingPrompt(t *testing.T) {
	_, err := NewProcessor(&fakeRunner{out: "x"}, &fakePublisher{}).Process(context.Background(), &agentsv1.PlaneApp{AgentId: "a"}, &Event{})
	if !errors.Is(err, ErrPromptRequired) {
		t.Fatal(err)
	}
}
func TestProcessorStopCancelsRunningRunner(t *testing.T) {
	r := &fakeRunner{out: "x", block: make(chan struct{})}
	p := NewProcessor(r, &fakePublisher{})
	e := &Event{InstallationID: "i", AgentRunID: "r", ActivityID: "a", Prompt: "x"}
	done := make(chan error, 1)
	go func() { _, err := p.Process(context.Background(), &agentsv1.PlaneApp{AgentId: "a"}, e); done <- err }()
	for {
		r.mu.Lock()
		n := r.calls
		r.mu.Unlock()
		if n == 1 {
			break
		}
	}
	audit, err := p.Stop(context.Background(), e)
	if err != nil || audit.State != RunStateStopping {
		t.Fatalf("audit=%+v err=%v", audit, err)
	}
	if audit, err = p.Stop(context.Background(), e); err != nil || audit.State != RunStateStopping {
		t.Fatalf("repeat audit=%+v err=%v", audit, err)
	}
	close(r.block)
	if err = <-done; err == nil {
		t.Fatal("expected cancelled runner")
	}
}
func TestProcessorConcurrentDuplicate(t *testing.T) {
	r := &fakeRunner{out: "x", block: make(chan struct{})}
	p := NewProcessor(r, &fakePublisher{})
	e := &Event{InstallationID: "i", AgentRunID: "r", ActivityID: "a", Prompt: "x"}
	done := make(chan error, 1)
	go func() { _, err := p.Process(context.Background(), &agentsv1.PlaneApp{AgentId: "a"}, e); done <- err }()
	for {
		r.mu.Lock()
		n := r.calls
		r.mu.Unlock()
		if n == 1 {
			break
		}
	}
	if _, err := p.Process(context.Background(), &agentsv1.PlaneApp{AgentId: "a"}, e); !errors.Is(err, ErrDuplicate) {
		t.Fatal(err)
	}
	close(r.block)
	<-done
}
