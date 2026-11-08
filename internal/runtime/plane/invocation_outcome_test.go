package plane

import (
	"context"
	"errors"
	"iter"
	"sync"
	"testing"
	"time"

	"go.orx.me/apps/butter/internal/a2ui"
	internalagent "go.orx.me/apps/butter/internal/agent"
	"go.orx.me/apps/butter/internal/repo/invocation"
	invocationmem "go.orx.me/apps/butter/internal/repo/invocation/memory"
	"go.orx.me/apps/butter/internal/runtime/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/adk/v2/agent"
	adkrunner "google.golang.org/adk/v2/runner"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"
)

// Exercise the real Runner and recorder, with a deterministic leaf in place
// of the remote box. The immutable id deliberately differs from its name.
func invocationProcessor(t *testing.T, event func(agent.InvocationContext) (string, error), rec runner.InvocationRecorder) (*Processor, *Event, *agentsv1.PlaneApp) {
	t.Helper()
	pb := agentsv1.Agent{AgentId: "agent-id", Name: "agent-name", WorkspaceId: "ws",
		Type:   agentsv1.AgentType_AGENT_TYPE_PI,
		Config: &agentsv1.AgentConfig{Pi: &agentsv1.PiAgentConfig{ButterboxId: "box"}}}
	builders := &internalagent.BoxAgentBuilders{Pi: func(pb *agentsv1.Agent, _ a2ui.CardPolicy) (agent.Agent, error) {
		return agent.New(agent.Config{Name: pb.GetName(), Run: func(ctx agent.InvocationContext) iter.Seq2[*session.Event, error] {
			return func(yield func(*session.Event, error) bool) {
				text, err := event(ctx)
				if err != nil {
					yield(nil, err)
					return
				}
				e := session.NewEvent(ctx, ctx.InvocationID())
				e.Author = pb.GetName()
				e.Content = genai.NewContentFromText(text, genai.RoleModel)
				yield(e, nil)
			}
		}})
	}}
	// The runner's legacy constructor accepts generated Agent values. Keep the
	// copy local to this integration fixture; production configuration already
	// owns the value slice before constructing the runner.
	//nolint:govet // generated protobuf messages contain an internal mutex.
	svc, err := runner.NewServiceWithMCPHTTPClientFactory(t.Context(), []agentsv1.Agent{pb}, nil, nil, nil, nil,
		session.InMemoryService(), nil, nil, nil, 0, nil, adkrunner.PluginConfig{}, nil, builders)
	if err != nil {
		t.Fatal(err)
	}
	svc.SetInvocationRecorder(rec)
	p := NewProcessor(svc, &fakePublisher{})
	p.controlPoll = time.Millisecond
	return p, &Event{WorkspaceID: "ws", AppID: "app", InstallationID: "install", AgentRunID: "run",
		ActivityID: "prompt", BotUserID: "user", Prompt: "work"}, &agentsv1.PlaneApp{AgentId: pb.GetAgentId()}
}

func TestInvocationUsesArbitratedPlaneOutcome(t *testing.T) {
	for _, tc := range []struct {
		name    string
		stop    bool
		text    string
		runErr  error
		status  agentsv1.InvocationStatus
		outcome RunState
	}{
		{"success", false, "answer", nil, agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED, RunStateSucceeded},
		{"failure", false, "", errors.New("box failed"), agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED, RunStateFailed},
		{"empty", false, "", nil, agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED, RunStateFailed},
		{"stop-clean-return", true, "late answer", nil, agentsv1.InvocationStatus_INVOCATION_STATUS_CANCELLED, RunStateStopped},
		{"stop-error-return", true, "", context.Canceled, agentsv1.InvocationStatus_INVOCATION_STATUS_CANCELLED, RunStateStopped},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec := invocationmem.New()
			started, release := make(chan struct{}), make(chan struct{})
			p, e, app := invocationProcessor(t, func(agent.InvocationContext) (string, error) {
				close(started)
				<-release
				return tc.text, tc.runErr
			}, rec)
			ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
			defer cancel()
			done := make(chan error, 1)
			go func() { _, err := p.Process(ctx, app, e); done <- err }()
			awaitPlane(t, started, "real runner start")
			var target string
			if tc.stop {
				var err error
				target, err = p.control.RequestStop(ctx, e)
				if err != nil || target == "" {
					t.Fatalf("Stop target=%q err=%v", target, err)
				}
			}
			close(release)
			err := <-done
			if tc.status == agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil {
				t.Fatal("expected unsuccessful execution")
			}
			rows, _, _, err := rec.List(ctx, invocation.ListFilter{WorkspaceID: e.WorkspaceID}, 10, "")
			if err != nil || len(rows) != 1 {
				t.Fatalf("records=%v err=%v", rows, err)
			}
			inv := rows[0]
			if inv.GetStatus() != tc.status || inv.GetAgentId() != app.GetAgentId() || inv.GetWorkspaceId() != e.WorkspaceID {
				t.Fatalf("invocation=%v", inv)
			}
			if tc.stop {
				if inv.GetOutput() != "" || inv.GetError() != "stopped by Plane" {
					t.Fatalf("stopped record contains late output: %v", inv)
				}
				if outcome, ended, err := p.control.Outcome(ctx, e, target); err != nil || !ended || outcome != tc.outcome {
					t.Fatalf("outcome=%s ended=%v err=%v", outcome, ended, err)
				}
			}
			if inv.GetFinishedAt() == nil {
				t.Fatal("terminal timestamp missing")
			}
			// Completion is already claimed. A later Stop must not rewrite it.
			if target, err := p.control.RequestStop(ctx, e); err != nil || target != "" {
				t.Fatalf("late Stop target=%q err=%v", target, err)
			}
		})
	}
}

type terminalGateRecorder struct {
	runner.InvocationRecorder
	entered chan struct{}
	release chan struct{}
	err     error
	once    sync.Once
}

func (r *terminalGateRecorder) Save(ctx context.Context, inv *agentsv1.Invocation) error {
	if inv.GetStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING {
		r.once.Do(func() { close(r.entered) })
		<-r.release
		if r.err != nil {
			return r.err
		}
	}
	return r.InvocationRecorder.Save(ctx, inv)
}

func TestPlaneConfirmationWaitsForInvocationPersistence(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "saved", true: "save-failed"}[fail], func(t *testing.T) {
			store := invocationmem.New()
			rec := &terminalGateRecorder{InvocationRecorder: store, entered: make(chan struct{}), release: make(chan struct{})}
			if fail {
				rec.err = errors.New("storage unavailable")
			}
			started, release := make(chan struct{}), make(chan struct{})
			p, e, app := invocationProcessor(t, func(agent.InvocationContext) (string, error) {
				close(started)
				<-release
				return "late answer", nil
			}, rec)
			ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
			defer cancel()
			done := make(chan error, 1)
			go func() { _, err := p.Process(ctx, app, e); done <- err }()
			awaitPlane(t, started, "runner start")
			target, err := p.control.RequestStop(ctx, e)
			if err != nil || target == "" {
				t.Fatalf("target=%q err=%v", target, err)
			}
			close(release)
			awaitPlane(t, rec.entered, "terminal write")
			if _, ended, err := p.control.Outcome(ctx, e, target); err != nil || ended {
				t.Fatalf("confirmed before Invocation settled: ended=%v err=%v", ended, err)
			}
			close(rec.release)
			err = <-done
			if fail && !errors.Is(err, runner.ErrInvocationFinalization) {
				t.Fatalf("missing persistence error: %v", err)
			}
			if _, ended, err := p.control.Outcome(ctx, e, target); err != nil || ended == fail {
				t.Fatalf("ended=%v fail=%v err=%v", ended, fail, err)
			}
			rows, _, _, err := store.List(ctx, invocation.ListFilter{WorkspaceID: e.WorkspaceID}, 10, "")
			if err != nil || len(rows) != 1 {
				t.Fatalf("records=%v err=%v", rows, err)
			}
			want := agentsv1.InvocationStatus_INVOCATION_STATUS_CANCELLED
			if fail {
				want = agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING
			}
			if rows[0].GetStatus() != want {
				t.Fatalf("status=%v want=%v", rows[0].GetStatus(), want)
			}
		})
	}
}

func TestPlaneCompletionBeforeStopPreservesInvocation(t *testing.T) {
	store := invocationmem.New()
	rec := &terminalGateRecorder{InvocationRecorder: store, entered: make(chan struct{}), release: make(chan struct{})}
	p, e, app := invocationProcessor(t, func(agent.InvocationContext) (string, error) { return "answer", nil }, rec)
	ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() { _, err := p.Process(ctx, app, e); done <- err }()
	awaitPlane(t, rec.entered, "completion claimed before Invocation Save")
	if target, err := p.control.RequestStop(ctx, e); err != nil || target != "" {
		t.Fatalf("late Stop target=%q err=%v", target, err)
	}
	close(rec.release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	rows, _, _, err := store.List(ctx, invocation.ListFilter{WorkspaceID: e.WorkspaceID}, 10, "")
	if err != nil || len(rows) != 1 || rows[0].GetStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED {
		t.Fatalf("late Stop changed completion: rows=%v err=%v", rows, err)
	}
}

func TestPlaneLeaseLossSuppressesInvocationTerminalWrite(t *testing.T) {
	store := invocationmem.New()
	started, release := make(chan struct{}), make(chan struct{})
	p, e, app := invocationProcessor(t, func(agent.InvocationContext) (string, error) {
		close(started)
		<-release
		return "late answer", nil
	}, store)
	ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() { _, err := p.Process(ctx, app, e); done <- err }()
	awaitPlane(t, started, "runner start")
	c := p.control.(*MemoryRunControl)
	c.mu.Lock()
	delete(c.runs, controlKey(e))
	c.mu.Unlock()
	close(release)
	if err := <-done; !errors.Is(err, ErrRunLeaseLost) {
		t.Fatalf("lost fence error=%v", err)
	}
	rows, _, _, err := store.List(ctx, invocation.ListFilter{WorkspaceID: e.WorkspaceID}, 10, "")
	if err != nil || len(rows) != 1 || rows[0].GetStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING {
		t.Fatalf("lost holder wrote terminal outcome: rows=%v err=%v", rows, err)
	}
}

func TestPlaneAgentResolutionRejectsForeignWorkspace(t *testing.T) {
	store := invocationmem.New()
	p, e, app := invocationProcessor(t, func(agent.InvocationContext) (string, error) {
		t.Error("foreign workspace reached the Agent")
		return "answer", nil
	}, store)
	e.WorkspaceID = "foreign"
	if _, err := p.Process(t.Context(), app, e); err == nil {
		t.Fatal("expected workspace rejection")
	}
	rows, _, _, err := store.List(t.Context(), invocation.ListFilter{}, 10, "")
	if err != nil || len(rows) != 0 {
		t.Fatalf("unauthorized invocation created: rows=%v err=%v", rows, err)
	}
}
