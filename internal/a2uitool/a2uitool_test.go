package a2uitool

import (
	"encoding/json"
	"errors"
	"iter"
	"maps"
	"testing"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/session"

	"go.orx.me/apps/butter/internal/a2ui"
	"go.orx.me/apps/butter/internal/testsupport/tooltest"
)

func TestToolParametersAreDescribed(t *testing.T) {
	tooltest.RequireParamDescriptions(t, NewToolset().tool)
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
	_, err := render(ctx, args)
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
