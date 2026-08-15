// Package aguitool exposes AG-UI client-declared frontend tools to an agent
// for exactly one run.
//
// A frontend tool executes on the AG-UI client, not the server: the agent
// declares the call and the run ends with it pending; the client answers on a
// later request with a tool-result message. That is precisely ADK's
// long-running function tool contract — a tool whose Run returns nil produces
// no FunctionResponse, marks the call in Event.LongRunningToolIDs, and ends
// the invocation — so each declaration becomes one long-running tool.
//
// Declarations travel per request through the context (the same seam
// agentfiletool uses for tenant scoping): ADK resolves Toolsets per
// invocation via Tools(ReadonlyContext), and ReadonlyContext chains the
// caller's context.Context, so a Toolset attached once at agent build time
// yields the request's tools without rebuilding agents. A run without
// declarations sees no tools and pays nothing.
package aguitool

import (
	"context"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/genai"
)

// Declaration is one client-declared frontend tool: a name the model calls,
// prose for the model, and a raw JSON-schema value for the arguments. The
// schema is passed through to the provider untouched
// (FunctionDeclaration.ParametersJsonSchema); the server never validates
// arguments because it never executes the tool.
type Declaration struct {
	Name        string
	Description string
	Parameters  any
}

type clientToolsKey struct{}

// WithClientTools returns a context carrying this run's frontend tool
// declarations.
func WithClientTools(ctx context.Context, decls []Declaration) context.Context {
	return context.WithValue(ctx, clientToolsKey{}, decls)
}

// ClientToolsFrom returns the frontend tool declarations carried by ctx, if
// any.
func ClientToolsFrom(ctx context.Context) []Declaration {
	decls, _ := ctx.Value(clientToolsKey{}).([]Declaration)
	return decls
}

// Toolset resolves the request's client-declared tools per invocation. Attach
// it once per LLM agent; it is inert for runs whose context carries no
// declarations.
type Toolset struct{}

func NewToolset() Toolset { return Toolset{} }

var _ tool.Toolset = Toolset{}

func (Toolset) Name() string { return "agui_client_tools" }

func (Toolset) Tools(ctx agent.ReadonlyContext) ([]tool.Tool, error) {
	decls := ClientToolsFrom(ctx)
	if len(decls) == 0 {
		return nil, nil
	}
	tools := make([]tool.Tool, 0, len(decls))
	for _, decl := range decls {
		tools = append(tools, &clientTool{decl: decl})
	}
	return tools, nil
}

// clientTool satisfies ADK's structural FunctionTool shape
// (internal/toolinternal): Declaration() feeds the model request, and Run
// returning (nil, nil) with IsLongRunning() true is the pause — no
// FunctionResponse is generated and the run ends with the call pending.
type clientTool struct {
	decl Declaration
}

func (t *clientTool) Name() string        { return t.decl.Name }
func (t *clientTool) Description() string { return t.decl.Description }
func (t *clientTool) IsLongRunning() bool { return true }

func (t *clientTool) Declaration() *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{
		Name:                 t.decl.Name,
		Description:          t.decl.Description,
		ParametersJsonSchema: t.decl.Parameters,
	}
}

func (t *clientTool) Run(agent.Context, any) (map[string]any, error) {
	return nil, nil
}
