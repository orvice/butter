package runner

import (
	"context"
	"strings"
	"testing"

	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestRun_ContextGuardThresholdAgentOverrideCompacts(t *testing.T) {
	backend := newFakeBackend(t)
	agents := []agentsv1.Agent{{
		Name:        "guarded",
		AgentId:     "guarded",
		WorkspaceId: "ws-a",
		Config: &agentsv1.AgentConfig{
			Model: "model-a",
			ContextGuard: &agentsv1.ContextGuardConfig{
				Strategy:  agentsv1.ContextGuardStrategy_CONTEXT_GUARD_STRATEGY_THRESHOLD,
				MaxTokens: 128,
			},
		},
	}}
	agent := &agents[0]

	svc := buildWorkflowService(t, backend, agents, []string{"model-a"}, session.InMemoryService())
	input := strings.Repeat("This deliberately long request exercises the configured context window. ", 20)
	got, err := svc.Run(
		context.Background(),
		agent.GetName(),
		[]*genai.Part{{Text: input}},
		"",
		turnCtxInfo(agent),
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got == "" {
		t.Fatal("Run returned an empty response")
	}

	// A threshold compaction makes one summarization request, then sends the
	// actual turn with the summary and continuation injected by ContextGuard.
	if calls := backend.callCount("model-a"); calls < 2 {
		t.Fatalf("model call count = %d, want a summarization call plus the turn", calls)
	}
	if input := backend.lastInput("model-a"); !strings.Contains(input, "[System: The conversation was compacted because it exceeded the context window.") {
		t.Fatalf("last model input does not show ContextGuard compaction: %q", input)
	}
}

func TestRun_ContextGuardSlidingWindowUsesContentEntryLimit(t *testing.T) {
	backend := newFakeBackend(t)
	agents := []agentsv1.Agent{{
		Name:        "sliding",
		AgentId:     "sliding",
		WorkspaceId: "ws-a",
		Config: &agentsv1.AgentConfig{
			Model: "model-a",
			ContextGuard: &agentsv1.ContextGuardConfig{
				Strategy: agentsv1.ContextGuardStrategy_CONTEXT_GUARD_STRATEGY_SLIDING_WINDOW,
				MaxTurns: 1,
			},
		},
	}}
	agent := &agents[0]
	svc := buildWorkflowService(t, backend, agents, []string{"model-a"}, session.InMemoryService())

	for i := 0; i < 4; i++ {
		_, err := svc.Run(
			context.Background(),
			agent.GetName(),
			[]*genai.Part{{Text: "turn " + string(rune('1'+i))}},
			"",
			turnCtxInfo(agent),
			nil,
			nil,
		)
		if err != nil {
			t.Fatalf("Run turn %d: %v", i+1, err)
		}
	}

	// Four ordinary turns produce four model calls. The sliding-window
	// strategy's content-entry limit is exceeded once there is enough history
	// to retain its recent window, adding a summarization call and rewriting
	// the final request.
	if calls := backend.callCount("model-a"); calls < 5 {
		t.Fatalf("model call count = %d, want at least one sliding-window summary call", calls)
	}
	if input := backend.lastInput("model-a"); !strings.Contains(input, "[System: The conversation was compacted because it exceeded the context window.") {
		t.Fatalf("last model input does not show sliding-window compaction: %q", input)
	}
}
