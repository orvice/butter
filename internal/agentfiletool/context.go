package agentfiletool

import "context"

type runtimeContextKey struct{}

// RuntimeContext is injected by the runner so tools can scope access without
// asking the model to supply tenant or identity fields.
type RuntimeContext struct {
	WorkspaceID string
	AgentName   string
}

func WithRuntimeContext(ctx context.Context, info RuntimeContext) context.Context {
	return context.WithValue(ctx, runtimeContextKey{}, info)
}

func RuntimeContextFrom(ctx context.Context) RuntimeContext {
	info, _ := ctx.Value(runtimeContextKey{}).(RuntimeContext)
	return info
}
