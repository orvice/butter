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

func TestRun_ContextGuardThresholdInheritsModelCapacity(t *testing.T) {
	backend := newFakeBackend(t)
	agents := []agentsv1.Agent{{
		Name:        "guarded-model-default",
		AgentId:     "guarded-model-default",
		WorkspaceId: "ws-a",
		Config: &agentsv1.AgentConfig{
			Model: "model-default",
			ContextGuard: &agentsv1.ContextGuardConfig{
				Strategy: agentsv1.ContextGuardStrategy_CONTEXT_GUARD_STRATEGY_THRESHOLD,
			},
		},
	}}
	agent := &agents[0]

	svc := buildWorkflowServiceWithModels(t, backend, agents, []*agentsv1.ModelConfig{{
		Name:                "model-default",
		ContextWindowTokens: 128,
	}}, session.InMemoryService())
	input := strings.Repeat("This deliberately long request exercises inherited model capacity. ", 20)
	if _, err := svc.Run(context.Background(), agent.GetName(), []*genai.Part{{Text: input}}, "", turnCtxInfo(agent), nil, nil); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if calls := backend.callCount("model-default"); calls < 2 {
		t.Fatalf("model call count = %d, want configured Model capacity to trigger a summary before the turn", calls)
	}
	if input := backend.lastInput("model-default"); !strings.Contains(input, "[System: The conversation was compacted because it exceeded the context window.") {
		t.Fatalf("last model input does not show inherited-capacity compaction: %q", input)
	}
}

func TestRun_ModelCapacityDoesNotEnableContextGuard(t *testing.T) {
	backend := newFakeBackend(t)
	agents := []agentsv1.Agent{{
		Name:        "unguarded",
		AgentId:     "unguarded",
		WorkspaceId: "ws-a",
		Config:      &agentsv1.AgentConfig{Model: "small-unguarded"},
	}}
	agent := &agents[0]
	svc := buildWorkflowServiceWithModels(t, backend, agents, []*agentsv1.ModelConfig{{
		Name:                "small-unguarded",
		ContextWindowTokens: 128,
	}}, session.InMemoryService())

	input := strings.Repeat("Model metadata alone must not compact this request. ", 20)
	if _, err := svc.Run(context.Background(), agent.GetName(), []*genai.Part{{Text: input}}, "", turnCtxInfo(agent), nil, nil); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if calls := backend.callCount("small-unguarded"); calls != 1 {
		t.Fatalf("model call count = %d, want one ordinary unguarded turn", calls)
	}
}

func TestRun_ContextGuardSlidingWindowUsesModelCapacityForSafetyRetries(t *testing.T) {
	backend := newFakeBackend(t)

	runTurns := func(t *testing.T, agentName, modelName string, capacity uint32) int {
		t.Helper()
		agents := []agentsv1.Agent{{
			Name:        agentName,
			AgentId:     agentName,
			WorkspaceId: "ws-a",
			Config: &agentsv1.AgentConfig{
				Model: modelName,
				ContextGuard: &agentsv1.ContextGuardConfig{
					Strategy: agentsv1.ContextGuardStrategy_CONTEXT_GUARD_STRATEGY_SLIDING_WINDOW,
					MaxTurns: 1,
				},
			},
		}}
		agent := &agents[0]
		svc := buildWorkflowServiceWithModels(t, backend, agents, []*agentsv1.ModelConfig{{
			Name:                modelName,
			ContextWindowTokens: capacity,
		}}, session.InMemoryService())
		for i := 0; i < 4; i++ {
			input := strings.Repeat("long sliding-window content ", 20) + string(rune('1'+i))
			if _, err := svc.Run(context.Background(), agent.GetName(), []*genai.Part{{Text: input}}, "", turnCtxInfo(agent), nil, nil); err != nil {
				t.Fatalf("Run turn %d: %v", i+1, err)
			}
		}
		return backend.callCount(modelName)
	}

	fallbackCalls := runTurns(t, "sliding-fallback", "sliding-fallback-model", 0)
	configuredCalls := runTurns(t, "sliding-configured", "sliding-configured-model", 64)
	if configuredCalls <= fallbackCalls {
		t.Fatalf("configured small-window calls = %d, fallback-window calls = %d; want configured capacity to cause additional safety retries", configuredCalls, fallbackCalls)
	}
}

func TestReloadProtoAgentsRebuildsConfiguredModelRegistry(t *testing.T) {
	backend := newFakeBackend(t)
	agents := []agentsv1.Agent{{
		Name:        "reload-guarded",
		AgentId:     "reload-guarded",
		WorkspaceId: "ws-a",
		Config: &agentsv1.AgentConfig{
			Model: "reload-model",
			ContextGuard: &agentsv1.ContextGuardConfig{
				Strategy: agentsv1.ContextGuardStrategy_CONTEXT_GUARD_STRATEGY_THRESHOLD,
			},
		},
	}}
	agent := &agents[0]
	initialModels := []*agentsv1.ModelConfig{{Name: "reload-model", ContextWindowTokens: 128_000}}
	svc := buildWorkflowServiceWithModels(t, backend, agents, initialModels, session.InMemoryService())
	input := strings.Repeat("This request compacts only after the configured capacity reload. ", 20)

	before := turnCtxInfo(agent)
	before.SessionId = "before-reload"
	if _, err := svc.Run(context.Background(), agent.GetName(), []*genai.Part{{Text: input}}, "", before, nil, nil); err != nil {
		t.Fatalf("Run before reload: %v", err)
	}
	if calls := backend.callCount("reload-model"); calls != 1 {
		t.Fatalf("calls before reload = %d, want one ordinary turn", calls)
	}

	updatedProviders := fakeBackendProviders(backend, []*agentsv1.ModelConfig{{
		Name:                "reload-model",
		ContextWindowTokens: 128,
	}})
	if err := svc.ReloadProtoAgents(context.Background(), agents, updatedProviders, nil, nil); err != nil {
		t.Fatalf("ReloadProtoAgents: %v", err)
	}

	after := turnCtxInfo(agent)
	after.SessionId = "after-reload"
	if _, err := svc.Run(context.Background(), agent.GetName(), []*genai.Part{{Text: input}}, "", after, nil, nil); err != nil {
		t.Fatalf("Run after reload: %v", err)
	}
	if calls := backend.callCount("reload-model"); calls < 3 {
		t.Fatalf("total calls after reload = %d, want the future turn to add a summary call and model call", calls)
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
