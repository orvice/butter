package aguitool

import (
	"context"
	"testing"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"
)

// fakeReadonlyContext wraps a plain context.Context as the ReadonlyContext the
// toolset receives; only the embedded context is ever used.
type fakeReadonlyContext struct{ context.Context }

func (fakeReadonlyContext) UserContent() *genai.Content          { return nil }
func (fakeReadonlyContext) InvocationID() string                 { return "inv-1" }
func (fakeReadonlyContext) AgentName() string                    { return "Writer" }
func (fakeReadonlyContext) ReadonlyState() session.ReadonlyState { return nil }
func (fakeReadonlyContext) UserID() string                       { return "u1" }
func (fakeReadonlyContext) AppName() string                      { return "agui" }
func (fakeReadonlyContext) SessionID() string                    { return "agui-t-1" }
func (fakeReadonlyContext) Branch() string                       { return "" }

var _ agent.ReadonlyContext = fakeReadonlyContext{}

func TestToolsetResolvesDeclarationsFromContext(t *testing.T) {
	ts := NewToolset()

	// No declarations: inert.
	tools, err := ts.Tools(fakeReadonlyContext{t.Context()})
	if err != nil || len(tools) != 0 {
		t.Fatalf("Tools without declarations = %v, %v", tools, err)
	}

	schema := map[string]any{"type": "object"}
	ctx := WithClientTools(t.Context(), []Declaration{
		{Name: "confirm", Description: "ask the user", Parameters: schema},
	})
	tools, err = ts.Tools(fakeReadonlyContext{ctx})
	if err != nil || len(tools) != 1 {
		t.Fatalf("Tools = %v, %v", tools, err)
	}

	tl := tools[0]
	if tl.Name() != "confirm" || tl.Description() != "ask the user" {
		t.Errorf("tool = %q / %q", tl.Name(), tl.Description())
	}
	// The pause contract: long-running, and Run yields no result so no
	// FunctionResponse is generated server-side.
	if !tl.IsLongRunning() {
		t.Error("client tool must be long-running")
	}
	ct := tl.(*clientTool)
	decl := ct.Declaration()
	if decl.Name != "confirm" || decl.ParametersJsonSchema == nil {
		t.Errorf("declaration = %+v", decl)
	}
	if result, err := ct.Run(nil, nil); result != nil || err != nil {
		t.Errorf("Run = %v, %v; want nil, nil", result, err)
	}
}
