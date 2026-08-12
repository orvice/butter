package runner

import (
	"context"
	"iter"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"google.golang.org/adk/v2/agent"
	adkrunner "google.golang.org/adk/v2/runner"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestRunTurn_UsesStructuredEventOutputWhenNoVisibleText(t *testing.T) {
	svc := newEventService(t, func(ctx agent.InvocationContext) *session.Event {
		evt := session.NewEvent(ctx, ctx.InvocationID())
		evt.Output = map[string]any{"status": "ok", "count": 2}
		return evt
	})

	turn, err := svc.RunTurn(context.Background(), "test-agent",
		[]*genai.Part{genai.NewPartFromText("hello")}, "", testTurnContext("s1"), nil, nil)
	if err != nil {
		t.Fatalf("RunTurn: %v", err)
	}
	if got, want := turn.Output, `{"count":2,"status":"ok"}`; got != want {
		t.Fatalf("Output = %q, want %q", got, want)
	}
}

func TestRunTurn_DoesNotExposeThoughtText(t *testing.T) {
	svc := newEventService(t, func(ctx agent.InvocationContext) *session.Event {
		evt := session.NewEvent(ctx, ctx.InvocationID())
		evt.Content = &genai.Content{Role: genai.RoleModel, Parts: []*genai.Part{
			{Text: "private reasoning", Thought: true},
			{Text: "visible answer"},
		}}
		return evt
	})

	turn, err := svc.RunTurn(context.Background(), "test-agent",
		[]*genai.Part{genai.NewPartFromText("hello")}, "", testTurnContext("s1"), nil, nil)
	if err != nil {
		t.Fatalf("RunTurn: %v", err)
	}
	if got, want := turn.Output, "visible answer"; got != want {
		t.Fatalf("Output = %q, want %q", got, want)
	}
}

func TestRunTurn_PreservesEmptyTerminalDiagnostics(t *testing.T) {
	svc := newEventService(t, func(ctx agent.InvocationContext) *session.Event {
		evt := session.NewEvent(ctx, ctx.InvocationID())
		evt.FinishReason = genai.FinishReasonMaxTokens
		evt.ErrorCode = "MAX_TOKENS"
		evt.ErrorMessage = "output limit reached"
		return evt
	})

	turn, err := svc.RunTurn(context.Background(), "test-agent",
		[]*genai.Part{genai.NewPartFromText("hello")}, "", testTurnContext("s1"), nil, nil)
	if err != nil {
		t.Fatalf("RunTurn: %v", err)
	}
	if turn.Output != "" {
		t.Fatalf("Output = %q, want empty", turn.Output)
	}
	if turn.EventCount != 1 {
		t.Fatalf("EventCount = %d, want 1", turn.EventCount)
	}
	if turn.FinishReason != genai.FinishReasonMaxTokens {
		t.Fatalf("FinishReason = %q, want %q", turn.FinishReason, genai.FinishReasonMaxTokens)
	}
	if turn.ErrorCode != "MAX_TOKENS" || turn.ErrorMessage != "output limit reached" {
		t.Fatalf("error diagnostics = (%q, %q)", turn.ErrorCode, turn.ErrorMessage)
	}
}

func TestRunTurn_SerializesSameSession(t *testing.T) {
	var inFlight atomic.Int32
	var maxInFlight atomic.Int32
	started := make(chan struct{}, 2)
	release := make(chan struct{})
	svc := newEventService(t, func(ctx agent.InvocationContext) *session.Event {
		current := inFlight.Add(1)
		for current > maxInFlight.Load() && !maxInFlight.CompareAndSwap(maxInFlight.Load(), current) {
		}
		started <- struct{}{}
		<-release
		inFlight.Add(-1)
		evt := session.NewEvent(ctx, ctx.InvocationID())
		evt.Content = genai.NewContentFromText("done", genai.RoleModel)
		return evt
	})

	var wg sync.WaitGroup
	errs := make(chan error, 2)
	run := func() {
		defer wg.Done()
		_, err := svc.RunTurn(context.Background(), "test-agent",
			[]*genai.Part{genai.NewPartFromText("hello")}, "", testTurnContext("same"), nil, nil)
		errs <- err
	}

	wg.Add(1)
	go run()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("first turn did not start")
	}
	wg.Add(1)
	go run()
	select {
	case <-started:
		t.Fatal("second turn entered the same session concurrently")
	case <-time.After(100 * time.Millisecond):
	}
	close(release)
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("RunTurn: %v", err)
		}
	}
	if got := maxInFlight.Load(); got != 1 {
		t.Fatalf("max concurrent turns = %d, want 1", got)
	}
}

func TestRunTurn_AllowsDifferentSessionsToOverlap(t *testing.T) {
	started := make(chan struct{}, 2)
	release := make(chan struct{})
	svc := newEventService(t, func(ctx agent.InvocationContext) *session.Event {
		started <- struct{}{}
		<-release
		evt := session.NewEvent(ctx, ctx.InvocationID())
		evt.Content = genai.NewContentFromText("done", genai.RoleModel)
		return evt
	})

	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for _, sessionID := range []string{"s1", "s2"} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := svc.RunTurn(context.Background(), "test-agent",
				[]*genai.Part{genai.NewPartFromText("hello")}, "", testTurnContext(sessionID), nil, nil)
			errs <- err
		}()
	}
	for range 2 {
		select {
		case <-started:
		case <-time.After(time.Second):
			close(release)
			t.Fatal("different sessions did not overlap")
		}
	}
	close(release)
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("RunTurn: %v", err)
		}
	}
}

func newEventService(t *testing.T, event func(agent.InvocationContext) *session.Event) *Service {
	t.Helper()
	ag, err := agent.New(agent.Config{
		Name: "test-agent",
		Run: func(ctx agent.InvocationContext) iter.Seq2[*session.Event, error] {
			return func(yield func(*session.Event, error) bool) {
				evt := event(ctx)
				evt.Author = "test-agent"
				yield(evt, nil)
			}
		},
	})
	if err != nil {
		t.Fatalf("agent.New: %v", err)
	}

	svc, err := NewService(context.Background(), nil, nil, nil, nil, nil,
		session.InMemoryService(), nil, nil, adkrunner.PluginConfig{})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	svc.RegisterAgent("test-agent", ag)
	return svc
}

func testTurnContext(sessionID string) *agentsv1.ContextInfo {
	return &agentsv1.ContextInfo{
		Uuid:        "test-uuid-" + sessionID,
		SessionId:   sessionID,
		UserId:      "u1",
		ChannelName: "test-app",
		Source:      agentsv1.ContextSource_CONTEXT_SOURCE_API,
	}
}
