package application

import (
	"errors"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	"go.orx.me/apps/butter/internal/repo/auth"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	"go.orx.me/apps/butter/internal/repo/config/memory"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func contextGuardAgent(id string, typ agentsv1.AgentType, guard *agentsv1.ContextGuardConfig) *agentsv1.Agent {
	cfg := &agentsv1.AgentConfig{
		Model:        "model-a",
		ContextGuard: guard,
	}
	if typ == agentsv1.AgentType_AGENT_TYPE_WORKFLOW {
		cfg.Workflow = &agentsv1.WorkflowConfig{
			Nodes: []*agentsv1.WorkflowNode{{
				Name:     "human",
				Kind:     agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_HUMAN_INPUT,
				Question: "continue?",
			}},
			Edges: []*agentsv1.WorkflowEdge{{From: "START", To: "human"}},
		}
	}
	return &agentsv1.Agent{
		Name:    id,
		AgentId: id,
		Type:    typ,
		Config:  cfg,
	}
}

func thresholdGuard(maxTokens int32) *agentsv1.ContextGuardConfig {
	return &agentsv1.ContextGuardConfig{
		Strategy:  agentsv1.ContextGuardStrategy_CONTEXT_GUARD_STRATEGY_THRESHOLD,
		MaxTokens: maxTokens,
	}
}

func slidingGuard(maxTurns int32) *agentsv1.ContextGuardConfig {
	return &agentsv1.ContextGuardConfig{
		Strategy: agentsv1.ContextGuardStrategy_CONTEXT_GUARD_STRATEGY_SLIDING_WINDOW,
		MaxTurns: maxTurns,
	}
}

func requireInvalidArgument(t *testing.T, err error, field string) {
	t.Helper()
	cerr, ok := err.(*connect.Error)
	if !ok || cerr.Code() != connect.CodeInvalidArgument {
		t.Fatalf("error = %v, want InvalidArgument", err)
	}
	if !strings.Contains(cerr.Message(), field) {
		t.Fatalf("error = %q, want field %q", cerr.Message(), field)
	}
}

func TestAgentService_ContextGuardValidConfigRoundTripsOnCreateAndUpdate(t *testing.T) {
	store := memory.New()
	svc := NewAgentServiceServer(store)
	ctx := testCtx()

	created, err := svc.CreateAgent(ctx, connect.NewRequest(&agentsv1.CreateAgentRequest{
		Agent: contextGuardAgent("guarded", agentsv1.AgentType_AGENT_TYPE_LLM, thresholdGuard(32000)),
	}))
	if err != nil {
		t.Fatalf("CreateAgent: %v", err)
	}
	if got := created.Msg.GetAgent().GetConfig().GetContextGuard().GetMaxTokens(); got != 32000 {
		t.Fatalf("created max_tokens = %d, want 32000", got)
	}

	updated := proto.Clone(created.Msg.GetAgent()).(*agentsv1.Agent)
	updated.Config.ContextGuard = slidingGuard(6)
	response, err := svc.UpdateAgent(ctx, connect.NewRequest(&agentsv1.UpdateAgentRequest{Agent: updated}))
	if err != nil {
		t.Fatalf("UpdateAgent: %v", err)
	}
	guard := response.Msg.GetAgent().GetConfig().GetContextGuard()
	if guard.GetStrategy() != agentsv1.ContextGuardStrategy_CONTEXT_GUARD_STRATEGY_SLIDING_WINDOW || guard.GetMaxTurns() != 6 {
		t.Fatalf("updated ContextGuard = %v, want sliding window with 6 turns", guard)
	}
	if guard.GetMaxTokens() != 0 {
		t.Fatalf("updated max_tokens = %d, want zero", guard.GetMaxTokens())
	}
}

func TestAgentService_ContextGuardValidationRunsOnEveryWritePath(t *testing.T) {
	invalid := func() *agentsv1.ContextGuardConfig {
		return &agentsv1.ContextGuardConfig{
			Strategy: agentsv1.ContextGuardStrategy_CONTEXT_GUARD_STRATEGY_THRESHOLD,
			MaxTurns: 2,
		}
	}

	t.Run("create", func(t *testing.T) {
		store := memory.New()
		svc := NewAgentServiceServer(store)
		_, err := svc.CreateAgent(testCtx(), connect.NewRequest(&agentsv1.CreateAgentRequest{
			Agent: contextGuardAgent("invalid-create", agentsv1.AgentType_AGENT_TYPE_LLM, invalid()),
		}))
		requireInvalidArgument(t, err, "config.context_guard.max_turns")
		if _, getErr := store.GetAgent(testCtx(), wsTest, "invalid-create"); !errors.Is(getErr, configrepo.ErrNotFound) {
			t.Fatalf("invalid create was persisted: %v", getErr)
		}
	})

	t.Run("direct update", func(t *testing.T) {
		store := memory.New()
		seed := contextGuardAgent("invalid-update", agentsv1.AgentType_AGENT_TYPE_LLM, nil)
		if _, err := store.CreateAgent(testCtx(), wsTest, seed); err != nil {
			t.Fatalf("seed agent: %v", err)
		}
		svc := NewAgentServiceServer(store)
		_, err := svc.UpdateAgent(testCtx(), connect.NewRequest(&agentsv1.UpdateAgentRequest{
			Agent: contextGuardAgent("invalid-update", agentsv1.AgentType_AGENT_TYPE_LLM, invalid()),
		}))
		requireInvalidArgument(t, err, "config.context_guard.max_turns")
		stored, getErr := store.GetAgent(testCtx(), wsTest, "invalid-update")
		if getErr != nil {
			t.Fatalf("get stored agent: %v", getErr)
		}
		if stored.GetConfig().GetContextGuard() != nil {
			t.Fatal("invalid update changed the stored configuration")
		}
	})

	t.Run("repository-bound composite update", func(t *testing.T) {
		// The semantic validator must run before the coordinator dependency is
		// checked. An invalid patch therefore returns InvalidArgument rather
		// than the no-coordinator FailedPrecondition error.
		svc := NewAgentServiceServer(memory.New())
		_, err := svc.UpdateAgentConfiguration(auth.WithAdmin(testCtx()), connect.NewRequest(&agentsv1.UpdateAgentConfigurationRequest{
			AgentPatch: contextGuardAgent("invalid-composite", agentsv1.AgentType_AGENT_TYPE_LLM, invalid()),
		}))
		requireInvalidArgument(t, err, "config.context_guard.max_turns")
	})
}

func TestAgentService_ContextGuardRejectsUnsupportedTypes(t *testing.T) {
	for _, typ := range []agentsv1.AgentType{
		agentsv1.AgentType_AGENT_TYPE_LOOP,
		agentsv1.AgentType_AGENT_TYPE_SEQUENTIAL,
		agentsv1.AgentType_AGENT_TYPE_PARALLEL,
		agentsv1.AgentType_AGENT_TYPE_WORKFLOW,
		agentsv1.AgentType_AGENT_TYPE_PI,
	} {
		t.Run(typ.String(), func(t *testing.T) {
			svc := NewAgentServiceServer(memory.New())
			_, err := svc.CreateAgent(testCtx(), connect.NewRequest(&agentsv1.CreateAgentRequest{
				Agent: contextGuardAgent("unsupported-"+strings.ToLower(typ.String()), typ, thresholdGuard(32000)),
			}))
			requireInvalidArgument(t, err, "config.context_guard")
		})
	}
}

func TestAgentService_ContextGuardAcceptsLegacyUnspecifiedLLMAgent(t *testing.T) {
	svc := NewAgentServiceServer(memory.New())
	created, err := svc.CreateAgent(testCtx(), connect.NewRequest(&agentsv1.CreateAgentRequest{
		Agent: contextGuardAgent("legacy-llm", agentsv1.AgentType_AGENT_TYPE_UNSPECIFIED, thresholdGuard(64000)),
	}))
	if err != nil {
		t.Fatalf("CreateAgent: %v", err)
	}
	if created.Msg.GetAgent().GetConfig().GetContextGuard().GetMaxTokens() != 64000 {
		t.Fatalf("legacy unspecified Agent Context Override was not persisted")
	}
}

func TestAgentService_ContextGuardRejectsMalformedConfigs(t *testing.T) {
	cases := []struct {
		name  string
		guard *agentsv1.ContextGuardConfig
		field string
	}{
		{
			name:  "unspecified strategy",
			guard: &agentsv1.ContextGuardConfig{},
			field: "config.context_guard.strategy",
		},
		{
			name: "negative tokens",
			guard: &agentsv1.ContextGuardConfig{
				Strategy:  agentsv1.ContextGuardStrategy_CONTEXT_GUARD_STRATEGY_THRESHOLD,
				MaxTokens: -1,
			},
			field: "config.context_guard.max_tokens",
		},
		{
			name: "negative turns",
			guard: &agentsv1.ContextGuardConfig{
				Strategy: agentsv1.ContextGuardStrategy_CONTEXT_GUARD_STRATEGY_SLIDING_WINDOW,
				MaxTurns: -1,
			},
			field: "config.context_guard.max_turns",
		},
		{
			name: "sliding tokens",
			guard: &agentsv1.ContextGuardConfig{
				Strategy:  agentsv1.ContextGuardStrategy_CONTEXT_GUARD_STRATEGY_SLIDING_WINDOW,
				MaxTokens: 1,
			},
			field: "config.context_guard.max_tokens",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewAgentServiceServer(memory.New())
			_, err := svc.CreateAgent(testCtx(), connect.NewRequest(&agentsv1.CreateAgentRequest{
				Agent: contextGuardAgent("malformed-"+tc.name, agentsv1.AgentType_AGENT_TYPE_LLM, tc.guard),
			}))
			requireInvalidArgument(t, err, tc.field)
		})
	}
}
