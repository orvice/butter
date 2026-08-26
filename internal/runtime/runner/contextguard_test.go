package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
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

func TestReloadProtoAgentsUsesNewCapacityForWarmedDefaultAndOverrideCaches(t *testing.T) {
	const marker = "[System: The conversation was compacted because it exceeded the context window."
	backend := newFakeBackend(t)
	agents := []agentsv1.Agent{{
		Name:        "reload-guarded",
		AgentId:     "reload-guarded",
		WorkspaceId: "ws-a",
		Config: &agentsv1.AgentConfig{
			Model: "default-alias",
			ContextGuard: &agentsv1.ContextGuardConfig{
				Strategy: agentsv1.ContextGuardStrategy_CONTEXT_GUARD_STRATEGY_THRESHOLD,
			},
		},
	}}
	agent := &agents[0]
	initialModels := []*agentsv1.ModelConfig{
		{Name: "actual-default", Alias: "default-alias", ContextWindowTokens: 128_000},
		{Name: "actual-override", Alias: "override-alias", ContextWindowTokens: 128_000},
	}
	sessions := session.InMemoryService()
	svc := buildWorkflowServiceWithModels(t, backend, agents, initialModels, sessions)
	input := strings.Repeat("This request compacts only after the configured capacity reload. ", 20)

	defaultCtx := turnCtxInfo(agent)
	defaultCtx.SessionId = "reload-default-session"
	overrideCtx := turnCtxInfo(agent)
	overrideCtx.SessionId = "reload-override-session"
	if _, err := svc.Run(context.Background(), agent.GetName(), []*genai.Part{{Text: input}}, "", defaultCtx, nil, nil); err != nil {
		t.Fatalf("warm default Run: %v", err)
	}
	if _, err := svc.Run(context.Background(), agent.GetName(), []*genai.Part{{Text: input}}, "override-alias", overrideCtx, nil, nil); err != nil {
		t.Fatalf("warm override Run: %v", err)
	}
	if calls := backend.callCount("actual-default"); calls != 1 {
		t.Fatalf("default calls before reload = %d, want one ordinary turn", calls)
	}
	if calls := backend.callCount("actual-override"); calls != 1 {
		t.Fatalf("override calls before reload = %d, want one ordinary turn", calls)
	}

	defaultBefore, err := svc.GetSession(context.Background(), defaultCtx.GetChannelName(), defaultCtx.GetSessionId(), defaultCtx.GetUserId())
	if err != nil {
		t.Fatalf("GetSession default before reload: %v", err)
	}
	overrideBefore, err := svc.GetSession(context.Background(), overrideCtx.GetChannelName(), overrideCtx.GetSessionId(), overrideCtx.GetUserId())
	if err != nil {
		t.Fatalf("GetSession override before reload: %v", err)
	}
	defaultEventCount := defaultBefore.Events().Len()
	overrideEventCount := overrideBefore.Events().Len()

	updatedProviders := fakeBackendProviders(backend, []*agentsv1.ModelConfig{
		{Name: "actual-default", Alias: "default-alias", ContextWindowTokens: 128},
		{Name: "actual-override", Alias: "override-alias", ContextWindowTokens: 128},
	})
	if err := svc.ReloadProtoAgents(context.Background(), agents, updatedProviders, nil, nil); err != nil {
		t.Fatalf("ReloadProtoAgents: %v", err)
	}

	defaultAfterReload, err := svc.GetSession(context.Background(), defaultCtx.GetChannelName(), defaultCtx.GetSessionId(), defaultCtx.GetUserId())
	if err != nil {
		t.Fatalf("GetSession default after reload: %v", err)
	}
	overrideAfterReload, err := svc.GetSession(context.Background(), overrideCtx.GetChannelName(), overrideCtx.GetSessionId(), overrideCtx.GetUserId())
	if err != nil {
		t.Fatalf("GetSession override after reload: %v", err)
	}
	if got := defaultAfterReload.Events().Len(); got != defaultEventCount {
		t.Fatalf("default session events after reload = %d, want preserved %d", got, defaultEventCount)
	}
	if got := overrideAfterReload.Events().Len(); got != overrideEventCount {
		t.Fatalf("override session events after reload = %d, want preserved %d", got, overrideEventCount)
	}

	if _, err := svc.Run(context.Background(), agent.GetName(), []*genai.Part{{Text: input}}, "", defaultCtx, nil, nil); err != nil {
		t.Fatalf("first default Run after reload: %v", err)
	}
	if input := backend.lastInput("actual-default"); !strings.Contains(input, marker) {
		t.Fatalf("first default request after reload did not compact: %q", input)
	}
	if _, err := svc.Run(context.Background(), agent.GetName(), []*genai.Part{{Text: input}}, "override-alias", overrideCtx, nil, nil); err != nil {
		t.Fatalf("first override Run after reload: %v", err)
	}
	if input := backend.lastInput("actual-override"); !strings.Contains(input, marker) {
		t.Fatalf("first override request after reload did not compact: %q", input)
	}
}

func TestRun_ContextGuardUsesActualSelectedModelIDAndCapacity(t *testing.T) {
	const marker = "[System: The conversation was compacted because it exceeded the context window."
	longInput := strings.Repeat("This input is long enough to require context compaction. ", 20)

	tests := []struct {
		name          string
		defaultModel  string
		modelOverride string
		maxTokens     int32
		models        []*agentsv1.ModelConfig
		actualModelID string
	}{
		{
			name:         "default alias resolves before configured metadata lookup",
			defaultModel: "default-alias",
			models: []*agentsv1.ModelConfig{
				{Name: "actual-default", Alias: "default-alias", ContextWindowTokens: 128},
			},
			actualModelID: "actual-default",
		},
		{
			name:          "per-turn alias selects overridden model metadata",
			defaultModel:  "large-alias",
			modelOverride: "small-alias",
			models: []*agentsv1.ModelConfig{
				{Name: "actual-large", Alias: "large-alias", ContextWindowTokens: 128_000},
				{Name: "actual-small", Alias: "small-alias", ContextWindowTokens: 128},
			},
			actualModelID: "actual-small",
		},
		{
			name:          "agent override remains authoritative for overridden model",
			defaultModel:  "summary-alias",
			modelOverride: "override-alias",
			maxTokens:     128,
			models: []*agentsv1.ModelConfig{
				{Name: "actual-summary", Alias: "summary-alias", ContextWindowTokens: 128_000},
				{Name: "actual-override", Alias: "override-alias", ContextWindowTokens: 128_000},
			},
			actualModelID: "actual-override",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backend := newFakeBackend(t)
			agents := []agentsv1.Agent{{
				Name:        "selected-model-agent",
				AgentId:     "selected-model-agent",
				WorkspaceId: "ws-a",
				Config: &agentsv1.AgentConfig{
					Model: tt.defaultModel,
					ContextGuard: &agentsv1.ContextGuardConfig{
						Strategy:  agentsv1.ContextGuardStrategy_CONTEXT_GUARD_STRATEGY_THRESHOLD,
						MaxTokens: tt.maxTokens,
					},
				},
			}}
			svc := buildWorkflowServiceWithModels(t, backend, agents, tt.models, session.InMemoryService())
			if _, err := svc.Run(context.Background(), agents[0].GetName(), []*genai.Part{{Text: longInput}}, tt.modelOverride, turnCtxInfo(&agents[0]), nil, nil); err != nil {
				t.Fatalf("Run: %v", err)
			}

			request, ok := backend.lastRequest(tt.actualModelID)
			if !ok {
				t.Fatalf("provider received no request for actual model ID %q", tt.actualModelID)
			}
			if request.Model != tt.actualModelID || request.Decoded["model"] != tt.actualModelID {
				t.Fatalf("complete request model = %q (%v), want actual ID %q", request.Model, request.Decoded["model"], tt.actualModelID)
			}
			if len(request.Messages) == 0 {
				t.Fatal("complete request captured no messages")
			}
			if input := backend.lastInput(tt.actualModelID); !strings.Contains(input, marker) {
				t.Fatalf("selected-model request did not contain ContextGuard compaction marker: %q", input)
			}
			if tt.modelOverride != "" && backend.callCount(tt.modelOverride) != 0 {
				t.Fatalf("provider was called with alias %q; aliases must not be provider model IDs", tt.modelOverride)
			}
			if tt.modelOverride == "" && backend.callCount(tt.defaultModel) != 0 {
				t.Fatalf("provider was called with alias %q; aliases must not be provider model IDs", tt.defaultModel)
			}
		})
	}
}

func TestRun_LogsEffectiveContextWindowMetadataForAllSources(t *testing.T) {
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	backend := newFakeBackend(t)
	agents := []agentsv1.Agent{
		contextGuardRuntimeAgent("agent-source", "agent-alias", 32_000),
		contextGuardRuntimeAgent("model-source", "model-alias", 0),
		contextGuardRuntimeAgent("embedded-source", "embedded-alias", 0),
		contextGuardRuntimeAgent("fallback-source", "fallback-alias", 0),
	}
	models := []*agentsv1.ModelConfig{
		{Name: "actual-agent-model", Alias: "agent-alias", ContextWindowTokens: 64_000},
		{Name: "actual-agent-override-model", Alias: "agent-turn-override", ContextWindowTokens: 80_000},
		{Name: "actual-configured-model", Alias: "model-alias", ContextWindowTokens: 96_000},
		{Name: "gpt-4o", Alias: "embedded-alias"},
		{Name: "unknown-effective-window-model-324", Alias: "fallback-alias"},
	}
	svc := buildWorkflowServiceWithModels(t, backend, agents, models, session.InMemoryService())

	secrets := map[string]string{}
	modelOverrides := map[string]string{"agent-source": "agent-turn-override"}
	for i := range agents {
		secret := "private-prompt-for-" + agents[i].GetName()
		secrets[agents[i].GetName()] = secret
		ctxInfo := turnCtxInfo(&agents[i])
		ctxInfo.SessionId = "log-source-" + agents[i].GetName()
		if _, err := svc.Run(context.Background(), agents[i].GetName(), []*genai.Part{{Text: secret}}, modelOverrides[agents[i].GetName()], ctxInfo, nil, nil); err != nil {
			t.Fatalf("Run %s: %v", agents[i].GetName(), err)
		}
	}

	expected := map[string]map[string]any{
		"agent-source": {
			"strategy":                  agentsv1.ContextGuardStrategy_CONTEXT_GUARD_STRATEGY_THRESHOLD.String(),
			"selected_model_id":         "actual-agent-override-model",
			"metadata_source":           "agent",
			"configured_agent_override": float64(32_000),
			"configured_model_capacity": float64(80_000),
			"effective_context_window":  float64(32_000),
		},
		"model-source": {
			"strategy":                  agentsv1.ContextGuardStrategy_CONTEXT_GUARD_STRATEGY_THRESHOLD.String(),
			"selected_model_id":         "actual-configured-model",
			"metadata_source":           "model",
			"configured_agent_override": float64(0),
			"configured_model_capacity": float64(96_000),
			"effective_context_window":  float64(96_000),
		},
		"embedded-source": {
			"strategy":                  agentsv1.ContextGuardStrategy_CONTEXT_GUARD_STRATEGY_THRESHOLD.String(),
			"selected_model_id":         "gpt-4o",
			"metadata_source":           "embedded",
			"configured_agent_override": float64(0),
			"configured_model_capacity": float64(0),
			"effective_context_window":  float64(128_000),
		},
		"fallback-source": {
			"strategy":                  agentsv1.ContextGuardStrategy_CONTEXT_GUARD_STRATEGY_THRESHOLD.String(),
			"selected_model_id":         "unknown-effective-window-model-324",
			"metadata_source":           "fallback",
			"configured_agent_override": float64(0),
			"configured_model_capacity": float64(0),
			"effective_context_window":  float64(128_000),
		},
	}

	seen := make(map[string]bool)
	for _, line := range bytes.Split(bytes.TrimSpace(logs.Bytes()), []byte("\n")) {
		var record map[string]any
		if err := json.Unmarshal(line, &record); err != nil {
			t.Fatalf("decode slog record: %v", err)
		}
		if record["msg"] != "effective context window resolved" {
			continue
		}
		agentName, _ := record["agent"].(string)
		want, ok := expected[agentName]
		if !ok {
			t.Fatalf("unexpected resolution log for agent %q: %v", agentName, record)
		}
		if seen[agentName] {
			t.Fatalf("duplicate resolution log for agent %q", agentName)
		}
		seen[agentName] = true

		allowed := map[string]bool{"time": true, "level": true, "msg": true, "agent": true}
		for key, value := range want {
			allowed[key] = true
			if record[key] != value {
				t.Errorf("%s log field %s = %v, want %v", agentName, key, record[key], value)
			}
		}
		for key := range record {
			if !allowed[key] {
				t.Errorf("%s resolution log contains non-metadata field %q", agentName, key)
			}
		}
		if bytes.Contains(line, []byte(secrets[agentName])) {
			t.Errorf("%s resolution log contains conversation content", agentName)
		}
	}

	for agentName := range expected {
		if !seen[agentName] {
			t.Errorf("missing resolution log for agent %q", agentName)
		}
	}
}

func contextGuardRuntimeAgent(name, modelRef string, maxTokens int32) agentsv1.Agent {
	return agentsv1.Agent{
		Name:        name,
		AgentId:     name,
		WorkspaceId: "ws-a",
		Config: &agentsv1.AgentConfig{
			Model: modelRef,
			ContextGuard: &agentsv1.ContextGuardConfig{
				Strategy:  agentsv1.ContextGuardStrategy_CONTEXT_GUARD_STRATEGY_THRESHOLD,
				MaxTokens: maxTokens,
			},
		},
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
