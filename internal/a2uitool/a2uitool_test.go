package a2uitool

import (
	"encoding/json"
	"errors"
	"iter"
	"maps"
	"strings"
	"testing"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/session"

	"go.orx.me/apps/butter/internal/a2ui"
	"go.orx.me/apps/butter/internal/testsupport/tooltest"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestToolParametersAreDescribed(t *testing.T) {
	tooltest.RequireParamDescriptions(t, NewToolset(a2ui.CardPolicy{}).tool)
}

func disabledPolicy() a2ui.CardPolicy {
	return a2ui.CardPolicy{}.Narrow(&agentsv1.ResultCardConfig{
		Generation: agentsv1.ResultCardGeneration_RESULT_CARD_GENERATION_DISABLED,
	})
}

// PREFERRED adds the hint to render_ui's description, once; AUTO, a run's
// root, leaves the description without it.
func TestDescriptionCarriesThePreferredHint(t *testing.T) {
	auto := NewToolset(a2ui.CardPolicy{}).tool.Description()
	preferred := NewToolset(a2ui.CardPolicy{}.Narrow(&agentsv1.ResultCardConfig{
		Presentation: agentsv1.ResultCardPresentation_RESULT_CARD_PRESENTATION_PREFERRED,
	})).tool.Description()
	if strings.Contains(auto, PreferredHint) {
		t.Errorf("the AUTO description carries the hint:\n%s", auto)
	}
	if strings.Count(preferred, PreferredHint) != 1 || strings.Replace(preferred, " "+PreferredHint, "", 1) != auto {
		t.Errorf("the PREFERRED description is not the AUTO one plus the hint:\n%s", preferred)
	}
}

// render_ui is offered only in an A2UI run, and only where the agent's Card
// Policy allows cards.
func TestToolsetOffersRenderUIWhereThePolicyAllows(t *testing.T) {
	runCtx, _ := newRunCtx(t)
	plainCtx := &toolCtx{StrictContextMock: agent.NewStrictContextMock(t.Context()), state: mapState{}}
	cases := []struct {
		name   string
		policy a2ui.CardPolicy
		ctx    agent.ReadonlyContext
		want   int
	}{
		{name: "a run's root, A2UI run", policy: a2ui.CardPolicy{}, ctx: runCtx, want: 1},
		{name: "disabled, A2UI run", policy: disabledPolicy(), ctx: runCtx, want: 0},
		{name: "a run's root, no A2UI run", policy: a2ui.CardPolicy{}, ctx: plainCtx, want: 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tools, err := NewToolset(tc.policy).Tools(tc.ctx)
			if err != nil {
				t.Fatal(err)
			}
			if len(tools) != tc.want {
				t.Fatalf("offered %d tools, want %d", len(tools), tc.want)
			}
			if tc.want == 1 && tools[0].Name() != ToolName {
				t.Fatalf("offered %q, want %q", tools[0].Name(), ToolName)
			}
		})
	}
}

// The tool runs under the policy it was built with: where cards are off it
// refuses and writes nothing, even inside an A2UI run.
func TestRenderRefusesUnderADisabledPolicy(t *testing.T) {
	ctx, st := newRunCtx(t)
	var args renderArgs
	if err := json.Unmarshal([]byte(validCard), &args); err != nil {
		t.Fatal(err)
	}
	if _, err := render(ctx, disabledPolicy(), args); err == nil {
		t.Fatal("render accepted a card under a disabled policy")
	}
	if len(st) != 0 {
		t.Fatalf("a refused call wrote state: %v", st)
	}
}

type mapState map[string]any

func (s mapState) Get(key string) (any, error) {
	v, ok := s[key]
	if !ok {
		return nil, session.ErrStateKeyNotExist
	}
	return v, nil
}

func (s mapState) Set(key string, v any) error {
	s[key] = v
	return nil
}

func (s mapState) All() iter.Seq2[string, any] { return maps.All(s) }

// toolCtx is the slice of agent.Context render uses: context values (for
// the A2UI run), session state and the invocation ID.
type toolCtx struct {
	agent.StrictContextMock
	state mapState
}

func (c *toolCtx) State() session.State { return c.state }
func (c *toolCtx) InvocationID() string { return "inv-1" }

// newRunCtx returns the context of one A2UI run's tool calls and the
// session state they write.
func newRunCtx(t *testing.T) (*toolCtx, mapState) {
	t.Helper()
	ctx := a2ui.WithRun(t.Context(), &a2ui.Run{ThreadID: "thread-1", RunID: "run-1", MessageID: "msg-1"})
	st := mapState{}
	return &toolCtx{StrictContextMock: agent.NewStrictContextMock(ctx), state: st}, st
}

const (
	validCard = `{"messages": [{"updateComponents": {"components": [
		{"id": "root", "component": "Card", "child": "title"},
		{"id": "title", "component": "Text", "text": "Deploy summary"}]}}],
		"fallback": "Deploy summary"}`
	invalidCard = `{"messages": [{"updateComponents": {"components": [
		{"id": "root", "component": "Marquee"}]}}],
		"fallback": "Deploy summary"}`
)

func call(t *testing.T, ctx *toolCtx, raw string) error {
	t.Helper()
	var args renderArgs
	if err := json.Unmarshal([]byte(raw), &args); err != nil {
		t.Fatal(err)
	}
	_, err := render(ctx, a2ui.CardPolicy{}, args)
	return err
}

func TestRenderStopsAfterRepeatedFailures(t *testing.T) {
	ctx, st := newRunCtx(t)
	for i := 1; i <= maxConsecutiveFailures; i++ {
		err := call(t, ctx, invalidCard)
		if err == nil {
			t.Fatalf("call %d: invalid card accepted", i)
		}
		if stopped := errors.Is(err, errRenderStopped); stopped != (i == maxConsecutiveFailures) {
			t.Fatalf("call %d: %v; want the stop notice on failure %d only", i, err, maxConsecutiveFailures)
		}
	}
	if err := call(t, ctx, validCard); !errors.Is(err, errRenderStopped) {
		t.Fatalf("valid card after the limit: %v; want it refused", err)
	}
	if len(st) != 0 {
		t.Fatalf("a refused call wrote state: %v", st)
	}
}

func TestRenderSuccessResetsFailures(t *testing.T) {
	ctx, st := newRunCtx(t)
	for range 2 {
		for i := 1; i < maxConsecutiveFailures; i++ {
			if err := call(t, ctx, invalidCard); err == nil || errors.Is(err, errRenderStopped) {
				t.Fatalf("failure %d: %v; want a plain validation error", i, err)
			}
		}
		if err := call(t, ctx, validCard); err != nil {
			t.Fatalf("valid card: %v", err)
		}
	}
	if len(st) != 2 {
		t.Fatalf("state holds %d cards, want 2", len(st))
	}
}
