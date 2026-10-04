package application

import (
	"context"
	"errors"
	"slices"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/adk/v2/session"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/repo/invocation"
	invocationmemory "go.orx.me/apps/butter/internal/repo/invocation/memory"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// AG-UI threads and the Stop (#402, ADR-0016 decisions 4 and 7), as the RPCs
// route them. The handler-level tests in internal/handler/http drive the
// same paths against real runs on two instances.

// fakeAGUIStopper records the Stops CancelAgentInvocation hands the AG-UI
// path.
type fakeAGUIStopper struct {
	stopped bool
	err     error
	calls   []string
}

func (f *fakeAGUIStopper) StopInvocation(_ context.Context, userID, sessionID, invocationID string) (bool, error) {
	f.calls = append(f.calls, userID+"/"+sessionID+"/"+invocationID)
	return f.stopped, f.err
}

// cancelCountingRunner counts the cancels that reach the runner.
type cancelCountingRunner struct {
	asyncTestRunner
	cancels int
}

func (r *cancelCountingRunner) CancelInvocation(string, string) bool {
	r.cancels++
	return true
}

func TestCancelAgentInvocation_AGUIRunsGoThroughTheAGUIStop(t *testing.T) {
	invRepo := invocationmemory.New()
	for _, inv := range []*agentsv1.Invocation{
		{
			Id: "inv-detached", WorkspaceId: wsTest, UserId: "user-1", AppName: "agui", SessionId: "agui-t-1",
			Source: invocation.SourceAGUIDetached, Status: agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING,
			StartedAt: timestamppb.Now(),
		},
		{
			// A run without the opt-in: the runner recorded and registered it.
			Id: "inv-in-request", WorkspaceId: wsTest, UserId: "user-1", AppName: "agui", SessionId: "agui-t-2",
			Source: "CONTEXT_SOURCE_API", Status: agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING,
			StartedAt: timestamppb.Now(),
		},
	} {
		if err := invRepo.Save(context.Background(), inv); err != nil {
			t.Fatal(err)
		}
	}
	runner := &cancelCountingRunner{}
	stopper := &fakeAGUIStopper{stopped: true}
	svc := &AgentServiceServer{runnerSvc: runner, invRepo: invRepo}
	svc.SetAGUIRunStopper(stopper)
	cancel := func(id string) (*connect.Response[agentsv1.CancelAgentInvocationResponse], error) {
		return svc.CancelAgentInvocation(testContextWithUser(wsTest, "user-1"),
			connect.NewRequest(&agentsv1.CancelAgentInvocationRequest{InvocationId: id}))
	}

	resp, err := cancel("inv-detached")
	if err != nil || !resp.Msg.GetCancelled() {
		t.Fatalf("CancelAgentInvocation = %v, %v; want cancelled", resp, err)
	}
	if !slices.Equal(stopper.calls, []string{"user-1/agui-t-1/inv-detached"}) || runner.cancels != 0 {
		t.Fatalf("stops %v, runner cancels %d; want the AG-UI Stop for the run, the runner untouched", stopper.calls, runner.cancels)
	}

	// A run that already ended: the Stop finds nothing to stop.
	stopper.stopped = false
	if resp, err := cancel("inv-detached"); err != nil || resp.Msg.GetCancelled() || runner.cancels != 0 {
		t.Fatalf("CancelAgentInvocation of an ended run = %v, %v (runner cancels %d)", resp, err, runner.cancels)
	}

	// The Stop cannot be reached: a retryable error, not a false "nothing ran".
	stopper.err = errors.New("redis: connection refused")
	if _, err := cancel("inv-detached"); connect.CodeOf(err) != connect.CodeUnavailable {
		t.Fatalf("CancelAgentInvocation with the Stop unreachable = %v, want unavailable", err)
	}
	stopper.err = nil

	// Every other Invocation still goes to the runner.
	if resp, err := cancel("inv-in-request"); err != nil || !resp.Msg.GetCancelled() || runner.cancels != 1 || len(stopper.calls) != 3 {
		t.Fatalf("CancelAgentInvocation of a runner-recorded run = %v, %v (runner cancels %d, stops %v)", resp, err, runner.cancels, stopper.calls)
	}

	// Without the AG-UI Stop wired, an AG-UI-owned run cannot be cancelled.
	svc.aguiRuns = nil
	if _, err := cancel("inv-detached"); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("CancelAgentInvocation without the AG-UI Stop = %v, want failed_precondition", err)
	}
}

// fakeAGUIThreads holds AG-UI threads for DeleteSession.
type fakeAGUIThreads struct {
	held     bool
	err      error
	holds    []string
	released int
}

type holdKey struct{}

func (f *fakeAGUIThreads) HoldThreadForDelete(ctx context.Context, userID, sessionID string) (context.Context, func(), bool, error) {
	f.holds = append(f.holds, userID+"/"+sessionID)
	if f.err != nil || !f.held {
		return nil, nil, false, f.err
	}
	return context.WithValue(ctx, holdKey{}, true), func() { f.released++ }, true, nil
}

// heldDeleteSessions notes how each session delete ran.
type heldDeleteSessions struct {
	stubSessionService
	threads *fakeAGUIThreads
	svc     *SessionServiceServer
	// underHold: the delete ran on the hold's context, before its release.
	underHold bool
	// deletingMarked: the process-local deleting guard was set.
	deletingMarked bool
}

func (s *heldDeleteSessions) Delete(ctx context.Context, req *session.DeleteRequest) error {
	held, _ := ctx.Value(holdKey{}).(bool)
	s.underHold = held && s.threads.released == 0
	s.deletingMarked = s.svc.IsSessionDeleting(req.SessionID)
	return s.stubSessionService.Delete(ctx, req)
}

func newAGUIDeleteFixture(t *testing.T, threads *fakeAGUIThreads) (*SessionServiceServer, *heldDeleteSessions, *invocationmemory.Store, *stubDeleteCoordinator) {
	t.Helper()
	invRepo := invocationmemory.New()
	if err := invRepo.Save(context.Background(), &agentsv1.Invocation{
		Id: "inv-1", WorkspaceId: wsTest, UserId: "user-1", AppName: "agui", SessionId: "agui-t-1",
		Source: invocation.SourceAGUIDetached, Status: agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING,
		Input: "hello", StartedAt: timestamppb.Now(),
	}); err != nil {
		t.Fatal(err)
	}
	coord := newStubDeleteCoordinator()
	svc := NewSessionServiceServer()
	sessions := &heldDeleteSessions{threads: threads, svc: svc}
	svc.SetSessionService(sessions)
	svc.SetInvocationRepo(invRepo)
	svc.SetAsyncCoordinator(coord)
	svc.SetAGUIThreads(threads)
	return svc, sessions, invRepo, coord
}

func deleteAGUIThread(svc *SessionServiceServer) error {
	_, err := svc.DeleteSession(testContextWithUser(wsTest, "user-1"), connect.NewRequest(&agentsv1.DeleteSessionRequest{
		AppName: "agui", UserId: "user-1", SessionId: "agui-t-1",
	}))
	return err
}

// An AG-UI thread is deleted under its lease, held by the AG-UI path, in
// place of the process-local deleting guard and the async coordinator.
func TestDeleteSession_AGUIThreadIsDeletedUnderItsLease(t *testing.T) {
	threads := &fakeAGUIThreads{held: true}
	svc, sessions, invRepo, coord := newAGUIDeleteFixture(t, threads)
	var notified int
	svc.AddSessionDeleteListener(func(string, string, string) { notified++ })

	if err := deleteAGUIThread(svc); err != nil {
		t.Fatalf("DeleteSession: %v", err)
	}
	if !slices.Equal(threads.holds, []string{"user-1/agui-t-1"}) || threads.released != 1 {
		t.Fatalf("holds %v, released %d; want the thread held once and released", threads.holds, threads.released)
	}
	if len(sessions.deleted) != 1 || !sessions.underHold || sessions.deletingMarked {
		t.Fatalf("deleted %d, under the hold %v, deleting guard set %v", len(sessions.deleted), sessions.underHold, sessions.deletingMarked)
	}
	if coord.wasCancelled("inv-1") {
		t.Fatal("the async coordinator was asked to cancel an AG-UI run")
	}
	if inv, err := invRepo.Get(context.Background(), wsTest, "inv-1"); err != nil || inv.GetInput() != "" {
		t.Fatalf("record = %+v, %v; want its content redacted", inv, err)
	}
	if notified != 1 {
		t.Fatalf("listeners notified %d times, want 1", notified)
	}
}

// A thread whose lease does not come free in time, or cannot be held at all,
// is left whole and the delete fails with a retryable error.
func TestDeleteSession_AGUIThreadThatCannotBeHeldIsLeftWhole(t *testing.T) {
	for name, threads := range map[string]*fakeAGUIThreads{
		"lease not free in time": {held: false},
		"lease unreachable":      {err: errors.New("redis: connection refused")},
	} {
		t.Run(name, func(t *testing.T) {
			svc, sessions, invRepo, _ := newAGUIDeleteFixture(t, threads)
			var notified int
			svc.AddSessionDeleteListener(func(string, string, string) { notified++ })

			if err := deleteAGUIThread(svc); connect.CodeOf(err) != connect.CodeUnavailable {
				t.Fatalf("DeleteSession = %v, want unavailable", err)
			}
			if len(sessions.deleted) != 0 || notified != 0 {
				t.Fatalf("deleted %d, notified %d; want nothing deleted", len(sessions.deleted), notified)
			}
			if inv, err := invRepo.Get(context.Background(), wsTest, "inv-1"); err != nil || inv.GetInput() != "hello" {
				t.Fatalf("record = %+v, %v; want it untouched", inv, err)
			}
		})
	}
}

// Sessions of every other app keep their delete path: the AG-UI path is not
// consulted.
func TestDeleteSession_OtherAppsDoNotHoldAnAGUIThread(t *testing.T) {
	threads := &fakeAGUIThreads{held: true}
	svc, sessions, _, _ := newAGUIDeleteFixture(t, threads)
	if _, err := svc.DeleteSession(testContextWithUser(wsTest, "user-1"), connect.NewRequest(&agentsv1.DeleteSessionRequest{
		AppName: "web-chat", UserId: "user-1", SessionId: "sess-1",
	})); err != nil {
		t.Fatalf("DeleteSession: %v", err)
	}
	if len(threads.holds) != 0 || !sessions.deletingMarked || sessions.underHold {
		t.Fatalf("holds %v, deleting guard set %v, under a hold %v; want the web-chat path", threads.holds, sessions.deletingMarked, sessions.underHold)
	}
}
