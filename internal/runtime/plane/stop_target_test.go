package plane

import (
	"context"
	"errors"
	"testing"
	"time"

	planeprocessing "go.orx.me/apps/butter/internal/repo/planeprocessing"
	"go.orx.me/apps/butter/internal/runtime/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/genai"
)

type rejectedProcessingClaims struct{ completionConflictClaims }

func (rejectedProcessingClaims) Claim(context.Context, string, string, string, string, string, time.Time) (planeprocessing.Claim, error) {
	return planeprocessing.Claim{}, planeprocessing.ErrClaimConflict
}

func TestRejectedProcessingClaimReleasesMutex(t *testing.T) {
	for _, owner := range []string{"", "pod"} {
		t.Run("owner="+owner, func(t *testing.T) {
			p := NewProcessor(spyRunner{}, &spyPublisher{})
			p.SetClaimRepository(rejectedProcessingClaims{}, owner)
			e := &Event{InstallationID: "i", AgentRunID: "r", ActivityID: "prompt", Prompt: "work"}
			_, err := p.Process(t.Context(), &agentsv1.PlaneApp{AgentId: "agent"}, e)
			if !errors.Is(err, planeprocessing.ErrClaimConflict) {
				t.Fatalf("claim error=%v", err)
			}
			done := make(chan struct{})
			go func() {
				_, _ = p.Stop(t.Context(), e)
				close(done)
			}()
			select {
			case <-done:
			case <-time.After(time.Second):
				t.Fatal("rejected claim left processor mutex locked")
			}
		})
	}
}

type stopTargetRunner struct {
	started chan struct{}
}

func (r stopTargetRunner) Run(ctx context.Context, _ string, _ []*genai.Part, _ string, _ *agentsv1.ContextInfo, _ runner.EventCallback, _ runner.CompactionCallback) (string, error) {
	close(r.started)
	<-ctx.Done()
	return "", ctx.Err()
}

func TestStopUsesRunIdentityInsteadOfItsOwnActivity(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	r := stopTargetRunner{started: make(chan struct{})}
	pub := &fakePublisher{}
	p := NewProcessor(r, pub)
	control := NewMemoryRunControl()
	p.SetRunControl(control)
	e := &Event{WorkspaceID: "ws", AppID: "app", InstallationID: "i", AgentRunID: "r", ActivityID: "prompt", Prompt: "work"}
	done := make(chan error, 1)
	go func() {
		_, err := p.Process(ctx, &agentsv1.PlaneApp{AgentId: "agent"}, e)
		done <- err
	}()
	<-r.started
	// Identical external IDs in another scope must not cancel this Prompt,
	// even when the foreign Stop happens to reuse the Prompt Activity ID.
	for _, scope := range []string{"workspace", "app", "installation"} {
		foreign := *e
		switch scope {
		case "workspace":
			foreign.WorkspaceID = "other"
		case "app":
			foreign.AppID = "other"
		case "installation":
			foreign.InstallationID = "other"
		}
		if _, err := p.Stop(ctx, &foreign); err != nil {
			t.Fatal(err)
		}
		control.mu.Lock()
		stopped := control.current(e).stopped
		control.mu.Unlock()
		if stopped {
			t.Fatalf("Stop crossed %s boundary", scope)
		}
	}
	stop := *e
	stop.ActivityID = "stop-activity"
	stop.Stop = true
	audit, err := p.Stop(ctx, &stop)
	if err != nil || audit.State != RunStateStopping {
		t.Errorf("stop audit=%+v err=%v; want running Prompt targeted", audit, err)
	}
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("run error=%v, want cancellation by Stop", err)
		}
	case <-ctx.Done():
		t.Error("Stop did not cancel the running Prompt")
	}
}
