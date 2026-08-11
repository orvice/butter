package application

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/repo/auth"
	"go.orx.me/apps/butter/internal/repo/invocation"
	invocationmemory "go.orx.me/apps/butter/internal/repo/invocation/memory"
	"go.orx.me/apps/butter/internal/runtime/asyncrun"
	"go.orx.me/apps/butter/internal/runtime/runner"
	"go.orx.me/apps/butter/internal/transport/connectx"
	"go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

// watchTestRunner is a hand-driven fake: RunSSE blocks until the test feeds
// events through emit and a final response through finish, so tests control
// exactly when live frames and the terminal transition happen.
type watchTestRunner struct {
	emit   chan *session.Event
	finish chan string

	mu      sync.Mutex
	started chan struct{} // closed when RunSSE begins
}

func newWatchTestRunner() *watchTestRunner {
	return &watchTestRunner{
		emit:    make(chan *session.Event),
		finish:  make(chan string),
		started: make(chan struct{}),
	}
}

func (r *watchTestRunner) IsReservedAgentName(string) bool      { return false }
func (r *watchTestRunner) CancelInvocation(string, string) bool { return false }
func (r *watchTestRunner) ResolveAgentRef(_, agentID string) (string, bool) {
	return agentID, agentID != ""
}
func (r *watchTestRunner) GetAgentIdentity(name string) (string, string, bool) {
	return name, name, true
}
func (r *watchTestRunner) Run(_ context.Context, _ string, _ []*genai.Part, _ string, _ *agentsv1.ContextInfo, _ runner.EventCallback, _ runner.CompactionCallback) (string, error) {
	return "", nil
}
func (r *watchTestRunner) RunSSE(ctx context.Context, _ string, _ []*genai.Part, _ string, _ *agentsv1.ContextInfo, onEvent runner.EventCallback, _ runner.CompactionCallback) (string, error) {
	r.mu.Lock()
	select {
	case <-r.started:
	default:
		close(r.started)
	}
	r.mu.Unlock()
	for {
		select {
		case evt := <-r.emit:
			if onEvent != nil {
				onEvent(evt)
			}
		case resp := <-r.finish:
			return resp, nil
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
}

func textEvent(text string, partial bool) *session.Event {
	evt := &session.Event{ID: "evt-" + text, Author: "agent", Timestamp: time.Now()}
	evt.Partial = partial
	evt.Content = &genai.Content{Role: "model", Parts: []*genai.Part{{Text: text}}}
	return evt
}

// watchTestEnv serves the real Connect handler over httptest with a real
// asyncrun.Coordinator, so watch tests cross the full RPC seam. Per-request
// identity comes from test headers: X-Test-User (user id), X-Test-Role, and
// X-Test-Workspace (defaults to wsTest).
type watchTestEnv struct {
	invRepo invocation.Repository
	coord   *asyncrun.Coordinator
	runner  *watchTestRunner
	client  agentsv1connect.AgentServiceClient
}

func newWatchTestEnv(t *testing.T) *watchTestEnv {
	t.Helper()
	invRepo := invocationmemory.New()
	fake := newWatchTestRunner()
	coord := asyncrun.New(invRepo, fake, asyncrun.Config{})

	svc := &AgentServiceServer{
		runnerSvc:  fake,
		invRepo:    invRepo,
		asyncCoord: coord,
	}
	path, handler := agentsv1connect.NewAgentServiceHandler(svc, connectx.HandlerOptions()...)
	mux := http.NewServeMux()
	mux.Handle(path, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		wsID := req.Header.Get("X-Test-Workspace")
		if wsID == "" {
			wsID = wsTest
		}
		ctx = workspace.WithID(ctx, wsID)
		if userID := req.Header.Get("X-Test-User"); userID != "" {
			role := req.Header.Get("X-Test-Role")
			if role == "" {
				role = "member"
			}
			ctx = auth.WithAuthenticated(ctx, &agentsv1.User{Id: userID, Role: role}, nil)
		}
		handler.ServeHTTP(w, req.WithContext(ctx))
	}))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return &watchTestEnv{
		invRepo: invRepo,
		coord:   coord,
		runner:  fake,
		client:  agentsv1connect.NewAgentServiceClient(srv.Client(), srv.URL),
	}
}

// seedInvocation persists a QUEUED dashboard-async invocation owned by owner.
func (e *watchTestEnv) seedInvocation(t *testing.T, id, owner string) *agentsv1.Invocation {
	t.Helper()
	inv := &agentsv1.Invocation{
		Id:          id,
		AgentName:   "test-agent",
		AgentId:     "test-agent",
		AppName:     "web-chat",
		UserId:      owner,
		SessionId:   "sess-" + id,
		Status:      agentsv1.InvocationStatus_INVOCATION_STATUS_QUEUED,
		Source:      "dashboard-async",
		WorkspaceId: wsTest,
	}
	if err := e.invRepo.Save(context.Background(), inv); err != nil {
		t.Fatal(err)
	}
	return inv
}

// watch opens a WatchAgentInvocation stream as the given user.
func (e *watchTestEnv) watch(t *testing.T, ctx context.Context, invID, userID, role string) *connect.ServerStreamForClient[agentsv1.WatchAgentInvocationResponse] {
	t.Helper()
	req := connect.NewRequest(&agentsv1.WatchAgentInvocationRequest{InvocationId: invID})
	req.Header().Set("X-Test-User", userID)
	if role != "" {
		req.Header().Set("X-Test-Role", role)
	}
	stream, err := e.client.WatchAgentInvocation(ctx, req)
	if err != nil {
		t.Fatalf("WatchAgentInvocation: %v", err)
	}
	return stream
}

// receive advances the stream by one frame or fails the test.
func receiveFrame(t *testing.T, stream *connect.ServerStreamForClient[agentsv1.WatchAgentInvocationResponse]) *agentsv1.WatchAgentInvocationResponse {
	t.Helper()
	if !stream.Receive() {
		t.Fatalf("stream ended early: %v", stream.Err())
	}
	return stream.Msg()
}

func TestWatchAgentInvocation_OrderedLiveFramesAndTerminal(t *testing.T) {
	env := newWatchTestEnv(t)
	inv := env.seedInvocation(t, "inv-live", "user-1")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stream := env.watch(t, ctx, inv.GetId(), "user-1", "")
	defer stream.Close()

	// First frame: authoritative snapshot (QUEUED — not enqueued yet).
	first := receiveFrame(t, stream)
	if first.GetState().GetInvocation().GetStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_QUEUED {
		t.Fatalf("first frame = %v, want QUEUED state", first)
	}

	// Start execution; the watcher must observe the RUNNING transition.
	env.coord.Enqueue(inv, "test-agent", []*genai.Part{{Text: "hi"}}, "")
	running := receiveFrame(t, stream)
	if running.GetState().GetInvocation().GetStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING {
		t.Fatalf("second frame = %v, want RUNNING state", running)
	}

	// Live text delta.
	env.runner.emit <- textEvent("partial text", true)
	delta := receiveFrame(t, stream)
	if delta.GetTextDelta().GetText() != "partial text" {
		t.Fatalf("third frame = %v, want text delta 'partial text'", delta)
	}

	// Terminal state, exactly once, then stream close.
	env.runner.finish <- "final answer"
	terminal := receiveFrame(t, stream)
	if terminal.GetState().GetInvocation().GetStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED {
		t.Fatalf("terminal frame = %v, want SUCCEEDED state", terminal)
	}
	if got := terminal.GetState().GetInvocation().GetOutput(); got != "final answer" {
		t.Fatalf("terminal output = %q, want 'final answer'", got)
	}
	if stream.Receive() {
		t.Fatalf("expected stream end after terminal frame, got %v", stream.Msg())
	}
	if err := stream.Err(); err != nil {
		t.Fatalf("stream error after terminal: %v", err)
	}
}

func TestWatchAgentInvocation_MultipleObservers(t *testing.T) {
	env := newWatchTestEnv(t)
	inv := env.seedInvocation(t, "inv-multi", "user-1")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	s1 := env.watch(t, ctx, inv.GetId(), "user-1", "")
	defer s1.Close()
	s2 := env.watch(t, ctx, inv.GetId(), "user-1", "")
	defer s2.Close()

	receiveFrame(t, s1) // QUEUED snapshot
	receiveFrame(t, s2)

	env.coord.Enqueue(inv, "test-agent", []*genai.Part{{Text: "hi"}}, "")
	receiveFrame(t, s1) // RUNNING
	receiveFrame(t, s2)

	env.runner.emit <- textEvent("chunk", true)
	for i, s := range []*connect.ServerStreamForClient[agentsv1.WatchAgentInvocationResponse]{s1, s2} {
		if got := receiveFrame(t, s).GetTextDelta().GetText(); got != "chunk" {
			t.Fatalf("observer %d delta = %q, want 'chunk'", i+1, got)
		}
	}

	env.runner.finish <- "done"
	for i, s := range []*connect.ServerStreamForClient[agentsv1.WatchAgentInvocationResponse]{s1, s2} {
		if got := receiveFrame(t, s).GetState().GetInvocation().GetStatus(); got != agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED {
			t.Fatalf("observer %d terminal = %v, want SUCCEEDED", i+1, got)
		}
	}
}

func TestWatchAgentInvocation_DisconnectDoesNotCancelRun(t *testing.T) {
	env := newWatchTestEnv(t)
	inv := env.seedInvocation(t, "inv-disc", "user-1")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Observer A attaches with its own cancellable context, then disconnects.
	aCtx, aCancel := context.WithCancel(ctx)
	sA := env.watch(t, aCtx, inv.GetId(), "user-1", "")
	receiveFrame(t, sA) // QUEUED snapshot

	// Observer B stays attached.
	sB := env.watch(t, ctx, inv.GetId(), "user-1", "")
	defer sB.Close()
	receiveFrame(t, sB)

	env.coord.Enqueue(inv, "test-agent", []*genai.Part{{Text: "hi"}}, "")
	receiveFrame(t, sB) // RUNNING

	// Disconnect all of observer A mid-run.
	sA.Close()
	aCancel()

	// The run continues: B still receives live frames and the terminal state.
	env.runner.emit <- textEvent("after disconnect", true)
	if got := receiveFrame(t, sB).GetTextDelta().GetText(); got != "after disconnect" {
		t.Fatalf("delta after disconnect = %q", got)
	}
	env.runner.finish <- "survived"
	terminal := receiveFrame(t, sB)
	if terminal.GetState().GetInvocation().GetStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED {
		t.Fatalf("terminal = %v, want SUCCEEDED", terminal)
	}

	// And the persisted record is complete.
	stored, err := env.invRepo.Get(context.Background(), inv.GetId())
	if err != nil {
		t.Fatal(err)
	}
	if stored.GetOutput() != "survived" {
		t.Fatalf("persisted output = %q, want 'survived'", stored.GetOutput())
	}
}

func TestWatchAgentInvocation_AllObserversGone_RunStillCompletes(t *testing.T) {
	env := newWatchTestEnv(t)
	inv := env.seedInvocation(t, "inv-none", "user-1")

	ctx, cancel := context.WithCancel(context.Background())
	stream := env.watch(t, ctx, inv.GetId(), "user-1", "")
	receiveFrame(t, stream)

	env.coord.Enqueue(inv, "test-agent", []*genai.Part{{Text: "hi"}}, "")

	// Drop the only observer.
	stream.Close()
	cancel()

	// Finish the run with nobody watching.
	env.runner.finish <- "unobserved result"

	deadline := time.After(5 * time.Second)
	for {
		stored, err := env.invRepo.Get(context.Background(), inv.GetId())
		if err == nil && stored.GetStatus() == agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED {
			if stored.GetOutput() != "unobserved result" {
				t.Fatalf("output = %q, want 'unobserved result'", stored.GetOutput())
			}
			return
		}
		select {
		case <-deadline:
			t.Fatalf("invocation never reached SUCCEEDED; last = %v", stored)
		case <-time.After(10 * time.Millisecond):
		}
	}
}

func TestWatchAgentInvocation_ReconnectSeesTerminalState(t *testing.T) {
	env := newWatchTestEnv(t)
	inv := env.seedInvocation(t, "inv-reconnect", "user-1")
	inv.Status = agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED
	inv.Output = "persisted result"
	if err := env.invRepo.Save(context.Background(), inv); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	stream := env.watch(t, ctx, inv.GetId(), "user-1", "")
	defer stream.Close()

	// A watcher attaching after the run gets one authoritative terminal
	// state frame and a clean close — no transient delta replay required.
	frame := receiveFrame(t, stream)
	state := frame.GetState().GetInvocation()
	if state.GetStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED || state.GetOutput() != "persisted result" {
		t.Fatalf("reconnect frame = %v, want terminal SUCCEEDED with output", frame)
	}
	if stream.Receive() {
		t.Fatalf("expected immediate close after terminal snapshot, got %v", stream.Msg())
	}
	if err := stream.Err(); err != nil {
		t.Fatalf("stream err: %v", err)
	}
}

func TestWatchAgentInvocation_UnauthorizedObservers(t *testing.T) {
	env := newWatchTestEnv(t)
	inv := env.seedInvocation(t, "inv-private", "owner-user")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Another workspace member cannot watch a private dashboard chat.
	stream := env.watch(t, ctx, inv.GetId(), "other-user", "")
	if stream.Receive() {
		t.Fatalf("expected no frames for non-owner, got %v", stream.Msg())
	}
	if connect.CodeOf(stream.Err()) != connect.CodeNotFound {
		t.Fatalf("non-owner err = %v, want NotFound", stream.Err())
	}

	// A caller from another workspace cannot see it either.
	req := connect.NewRequest(&agentsv1.WatchAgentInvocationRequest{InvocationId: inv.GetId()})
	req.Header().Set("X-Test-User", "owner-user")
	req.Header().Set("X-Test-Workspace", "other-workspace")
	other, err := env.client.WatchAgentInvocation(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	if other.Receive() {
		t.Fatalf("expected no frames across workspaces, got %v", other.Msg())
	}
	if connect.CodeOf(other.Err()) != connect.CodeNotFound {
		t.Fatalf("cross-workspace err = %v, want NotFound", other.Err())
	}

	// The global-admin support path still works.
	admin := env.watch(t, ctx, inv.GetId(), "admin-user", "admin")
	defer admin.Close()
	frame := receiveFrame(t, admin)
	if frame.GetState().GetInvocation().GetId() != inv.GetId() {
		t.Fatalf("admin frame = %v, want state for %s", frame, inv.GetId())
	}
}

// slowWatchCoordinator simulates the hub dropping a lagged observer: Watch
// returns an already-closed channel with no terminal frame.
type slowWatchCoordinator struct{ fakeAsyncCoordinator }

func (c *slowWatchCoordinator) Watch(string) (<-chan asyncrun.Frame, func()) {
	ch := make(chan asyncrun.Frame)
	close(ch)
	return ch, func() {}
}

func TestWatchAgentInvocation_LaggedObserverDisconnectedWithResourceExhausted(t *testing.T) {
	invRepo := invocationmemory.New()
	inv := &agentsv1.Invocation{
		Id:          "inv-lag",
		UserId:      "user-1",
		Status:      agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING,
		Source:      "dashboard-async",
		WorkspaceId: wsTest,
	}
	if err := invRepo.Save(context.Background(), inv); err != nil {
		t.Fatal(err)
	}

	svc := &AgentServiceServer{invRepo: invRepo, asyncCoord: &slowWatchCoordinator{}}
	path, handler := agentsv1connect.NewAgentServiceHandler(svc, connectx.HandlerOptions()...)
	mux := http.NewServeMux()
	mux.Handle(path, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := workspace.WithID(req.Context(), wsTest)
		ctx = auth.WithAuthenticated(ctx, &agentsv1.User{Id: "user-1", Role: "member"}, nil)
		handler.ServeHTTP(w, req.WithContext(ctx))
	}))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	client := agentsv1connect.NewAgentServiceClient(srv.Client(), srv.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	stream, err := client.WatchAgentInvocation(ctx, connect.NewRequest(&agentsv1.WatchAgentInvocationRequest{InvocationId: "inv-lag"}))
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()

	// The authoritative snapshot still arrives first.
	if !stream.Receive() {
		t.Fatalf("expected snapshot frame, got err %v", stream.Err())
	}
	// Then the lagged observer is disconnected with RESOURCE_EXHAUSTED.
	if stream.Receive() {
		t.Fatalf("expected disconnect, got %v", stream.Msg())
	}
	if connect.CodeOf(stream.Err()) != connect.CodeResourceExhausted {
		t.Fatalf("err = %v, want ResourceExhausted", stream.Err())
	}
}

func TestGetAgentInvocation_ActiveBySession(t *testing.T) {
	invRepo := invocationmemory.New()
	svc := &AgentServiceServer{invRepo: invRepo}
	ctx := testContextWithUser(wsTest, "user-1")

	inv := &agentsv1.Invocation{
		Id:          "inv-active",
		UserId:      "user-1",
		SessionId:   "sess-a",
		Status:      agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING,
		Source:      "dashboard-async",
		WorkspaceId: wsTest,
	}
	if err := invRepo.Save(context.Background(), inv); err != nil {
		t.Fatal(err)
	}

	resp, err := svc.GetAgentInvocation(ctx, connect.NewRequest(&agentsv1.GetAgentInvocationRequest{SessionId: "sess-a"}))
	if err != nil {
		t.Fatal(err)
	}
	if resp.Msg.GetInvocation().GetId() != "inv-active" {
		t.Fatalf("got %q, want inv-active", resp.Msg.GetInvocation().GetId())
	}

	// A session without an active invocation reports NotFound.
	_, err = svc.GetAgentInvocation(ctx, connect.NewRequest(&agentsv1.GetAgentInvocationRequest{SessionId: "sess-idle"}))
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("idle session err = %v, want NotFound", err)
	}
}

func TestGetAgentInvocation_PrivateOwnershipEnforced(t *testing.T) {
	invRepo := invocationmemory.New()
	svc := &AgentServiceServer{invRepo: invRepo}

	inv := &agentsv1.Invocation{
		Id:          "inv-owned",
		UserId:      "owner-user",
		Status:      agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING,
		Source:      "dashboard-async",
		WorkspaceId: wsTest,
	}
	if err := invRepo.Save(context.Background(), inv); err != nil {
		t.Fatal(err)
	}

	// Another member of the same workspace is refused with NotFound.
	otherCtx := testContextWithUser(wsTest, "other-user")
	_, err := svc.GetAgentInvocation(otherCtx, connect.NewRequest(&agentsv1.GetAgentInvocationRequest{InvocationId: "inv-owned"}))
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("non-owner err = %v, want NotFound", err)
	}

	// The owner can read it.
	ownerCtx := testContextWithUser(wsTest, "owner-user")
	if _, err := svc.GetAgentInvocation(ownerCtx, connect.NewRequest(&agentsv1.GetAgentInvocationRequest{InvocationId: "inv-owned"})); err != nil {
		t.Fatalf("owner read failed: %v", err)
	}

	// A global admin retains the support path.
	adminCtx := workspace.WithID(context.Background(), wsTest)
	adminCtx = auth.WithAuthenticated(adminCtx, &agentsv1.User{Id: "admin-user", Role: "admin"}, nil)
	if _, err := svc.GetAgentInvocation(adminCtx, connect.NewRequest(&agentsv1.GetAgentInvocationRequest{InvocationId: "inv-owned"})); err != nil {
		t.Fatalf("admin read failed: %v", err)
	}
}
