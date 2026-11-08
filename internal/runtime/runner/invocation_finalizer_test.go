package runner

import (
	"context"
	"errors"
	"testing"

	invocationmem "go.orx.me/apps/butter/internal/repo/invocation/memory"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"
)

func TestInvocationFinalizerRejectionSuppressesTerminalWrite(t *testing.T) {
	svc := newEventService(t, func(ctx agent.InvocationContext) *session.Event {
		e := session.NewEvent(ctx, ctx.InvocationID())
		e.Content = genai.NewContentFromText("answer", genai.RoleModel)
		return e
	})
	rec := invocationmem.New()
	svc.SetInvocationRecorder(rec)
	rejected := errors.New("terminal fence lost")
	ctx := WithInvocationFinalizer(t.Context(), func(context.Context, *agentsv1.Invocation) error {
		return rejected
	})
	_, err := svc.Run(ctx, "test-agent", []*genai.Part{{Text: "hello"}}, "", testTurnContext("rejected"), nil, nil)
	if !errors.Is(err, ErrInvocationFinalization) || !errors.Is(err, rejected) {
		t.Fatalf("Run error=%v", err)
	}
	inv, err := rec.Get(t.Context(), "", "test-uuid-rejected")
	if err != nil || inv.GetStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING {
		t.Fatalf("rejected terminal write: invocation=%v err=%v", inv, err)
	}
}

func TestOrdinaryCancellationDoesNotBecomeExplicitStop(t *testing.T) {
	svc := newEventService(t, func(ctx agent.InvocationContext) *session.Event {
		e := session.NewEvent(ctx, ctx.InvocationID())
		e.Content = genai.NewContentFromText("answer", genai.RoleModel)
		return e
	})
	rec := invocationmem.New()
	svc.SetInvocationRecorder(rec)
	ctx, cancel := context.WithCancel(t.Context())
	inv := svc.startInvocation(ctx, "test-agent", nil, "", testTurnContext("ordinary"))
	cancel()
	output := ""
	runErr := context.Canceled
	svc.finishInvocation(ctx, inv, &output, &runErr)
	saved, err := rec.Get(t.Context(), "", inv.GetId())
	if err != nil || saved.GetStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED {
		t.Fatalf("ordinary cancellation status=%v err=%v", saved, err)
	}
}

func TestWithoutRecordingSkipsInvocationFinalizer(t *testing.T) {
	svc := newEventService(t, func(ctx agent.InvocationContext) *session.Event {
		e := session.NewEvent(ctx, ctx.InvocationID())
		e.Content = genai.NewContentFromText("answer", genai.RoleModel)
		return e
	})
	svc.SetInvocationRecorder(invocationmem.New())
	ctx := WithInvocationFinalizer(t.Context(), func(context.Context, *agentsv1.Invocation) error {
		t.Error("outer-owned invocation must not be finalized by Runner")
		return nil
	})
	out, err := svc.Run(WithoutInvocationRecording(ctx), "test-agent",
		[]*genai.Part{{Text: "hello"}}, "", testTurnContext("outer"), nil, nil)
	if err != nil || out != "answer" {
		t.Fatalf("out=%q err=%v", out, err)
	}
}
