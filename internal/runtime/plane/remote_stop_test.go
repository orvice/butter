package plane

import (
	"context"
	"errors"
	"testing"
	"time"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestStopOnAnotherProcessorCancelsHolder(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
	defer cancel()
	r := stopTargetRunner{started: make(chan struct{})}
	holder := NewProcessor(r, &fakePublisher{})
	other := NewProcessor(r, &fakePublisher{})
	control := NewMemoryRunControl()
	holder.SetRunControl(control)
	other.SetRunControl(control)
	e := &Event{WorkspaceID: "w", AppID: "a", InstallationID: "i", AgentRunID: "run", ActivityID: "prompt", Prompt: "work"}
	done := make(chan error, 1)
	go func() {
		_, err := holder.Process(ctx, &agentsv1.PlaneApp{AgentId: "agent"}, e)
		done <- err
	}()
	<-r.started
	stop := *e
	stop.ActivityID = "stop"
	if _, err := other.Stop(ctx, &stop); err != nil {
		t.Fatal(err)
	}
	select {
	case <-done:
		if ctx.Err() != nil {
			t.Fatal("remote Stop did not cancel holder before the deadline")
		}
	case <-ctx.Done():
		t.Fatal("remote Stop did not reach the running Prompt")
	}
}

func TestProcessorLosingSharedLeaseCancelsAndSuppressesLateWrites(t *testing.T) {
	for _, backend := range []string{"memory", "redis"} {
		t.Run(backend, func(t *testing.T) {
			control, e, steal := controlFixture(t, backend)
			r := &controlledTestRunner{started: make(chan struct{}), canceled: make(chan struct{}), settle: make(chan struct{})}
			pub := &fakePublisher{}
			progress := &progressRecorder{}
			p := NewProcessor(r, pub)
			p.SetRunControl(control)
			p.SetProgressSink(progress)
			p.controlPoll = 10 * time.Millisecond
			ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
			defer cancel()
			done := make(chan error, 1)
			go func() {
				_, err := p.Process(ctx, &agentsv1.PlaneApp{AgentId: "agent"}, e)
				done <- err
			}()
			awaitPlane(t, r.started, "holder start")
			steal()
			next := *e
			next.ActivityID = "next"
			if err := control.Begin(ctx, &next, "new-holder"); err != nil {
				close(r.settle)
				t.Fatal(err)
			}
			awaitPlane(t, r.canceled, "lease loss cancellation")
			close(r.settle)
			if err := <-done; !errors.Is(err, ErrRunLeaseLost) {
				t.Fatalf("lost holder result=%v", err)
			}
			if pub.calls != 0 || len(progress.events) != 0 {
				t.Fatalf("late writes: replies=%d progress=%+v", pub.calls, progress.events)
			}
			if stopped, err := control.Renew(ctx, &next, "new-holder"); err != nil || stopped {
				t.Fatalf("new holder was changed by old cleanup: stopped=%v err=%v", stopped, err)
			}
		})
	}
}
