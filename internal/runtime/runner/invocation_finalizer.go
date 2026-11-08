package runner

import (
	"context"
	"errors"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// ErrInvocationFinalization means an external terminal arbiter or its
// Invocation write failed. Callers must not confirm that the run settled.
var ErrInvocationFinalization = errors.New("invocation finalization failed")

type invocationFinalizerKey struct{}

// WithInvocationFinalizer lets a coordinator arbitrate the terminal outcome
// before the runner persists it. The callback receives the proposed terminal
// record on a bounded, detached context and may change its status, output and
// error. An error suppresses the terminal write and reaches the Run caller.
// WithoutInvocationRecording takes precedence; without a recorder there is
// no callback. The coordinator must also settle runs that never record.
func WithInvocationFinalizer(ctx context.Context, finalize func(context.Context, *agentsv1.Invocation) error) context.Context {
	return context.WithValue(ctx, invocationFinalizerKey{}, finalize)
}

func invocationFinalizer(ctx context.Context) func(context.Context, *agentsv1.Invocation) error {
	fn, _ := ctx.Value(invocationFinalizerKey{}).(func(context.Context, *agentsv1.Invocation) error)
	return fn
}
