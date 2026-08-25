package pibox

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"connectrpc.com/connect"
	piv1 "github.com/orvice/butter-box/pkg/proto/butterbox/pi/v1"
	"github.com/orvice/butter-box/pkg/proto/butterbox/pi/v1/piv1connect"
	"google.golang.org/adk/v2/agent"
	adkrunner "google.golang.org/adk/v2/runner"
	adksession "google.golang.org/adk/v2/session"
	"google.golang.org/genai"
	"google.golang.org/protobuf/proto"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// fakePi is a typed PiService fake served through the real generated Connect
// handler, so the bridge is exercised against actual wire behavior.
type fakePi struct {
	piv1connect.UnimplementedPiServiceHandler

	mu           sync.Mutex
	nextID       int
	sessions     map[string]bool
	createReqs   []*piv1.CreateSessionRequest
	submitReqs   []*piv1.SubmitMessageRequest
	abortedIDs   []string
	createErr    error
	submitErr    error
	turnErr      error
	turnScript   []*piv1.GetTurnResponse // consumed in order; last entry repeats
	turnIdx      int
	gotAuth      string
	turnRequests int
}

func newFakePi() *fakePi {
	return &fakePi{
		sessions: map[string]bool{},
		turnScript: []*piv1.GetTurnResponse{
			{Running: false, Result: &piv1.TurnResult{Text: "hello from pi", StopReason: "stop"}},
		},
	}
}

func (f *fakePi) CreateSession(_ context.Context, req *connect.Request[piv1.CreateSessionRequest]) (*connect.Response[piv1.CreateSessionResponse], error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.gotAuth = req.Header().Get("Authorization")
	if f.createErr != nil {
		return nil, f.createErr
	}
	f.nextID++
	id := fmt.Sprintf("pi-%d", f.nextID)
	f.sessions[id] = true
	f.createReqs = append(f.createReqs, proto.Clone(req.Msg).(*piv1.CreateSessionRequest))
	return connect.NewResponse(&piv1.CreateSessionResponse{
		Session: &piv1.Session{Id: id, Cwd: req.Msg.GetCwd()},
	}), nil
}

func (f *fakePi) SubmitMessage(_ context.Context, req *connect.Request[piv1.SubmitMessageRequest]) (*connect.Response[piv1.SubmitMessageResponse], error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.submitErr != nil {
		return nil, f.submitErr
	}
	if !f.sessions[req.Msg.GetSessionId()] {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("session not found"))
	}
	f.submitReqs = append(f.submitReqs, proto.Clone(req.Msg).(*piv1.SubmitMessageRequest))
	return connect.NewResponse(&piv1.SubmitMessageResponse{TurnCursor: "cursor-1"}), nil
}

func (f *fakePi) GetTurn(_ context.Context, _ *connect.Request[piv1.GetTurnRequest]) (*connect.Response[piv1.GetTurnResponse], error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.turnRequests++
	if f.turnErr != nil {
		return nil, f.turnErr
	}
	resp := f.turnScript[f.turnIdx]
	if f.turnIdx < len(f.turnScript)-1 {
		f.turnIdx++
	}
	return connect.NewResponse(proto.Clone(resp).(*piv1.GetTurnResponse)), nil
}

func (f *fakePi) AbortSession(_ context.Context, req *connect.Request[piv1.AbortSessionRequest]) (*connect.Response[piv1.AbortSessionResponse], error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.abortedIDs = append(f.abortedIDs, req.Msg.GetSessionId())
	return connect.NewResponse(&piv1.AbortSessionResponse{}), nil
}

func (f *fakePi) createCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.createReqs)
}

func (f *fakePi) abortCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.abortedIDs)
}

func serveFake(t *testing.T, f *fakePi) string {
	t.Helper()
	path, handler := piv1connect.NewPiServiceHandler(f)
	mux := http.NewServeMux()
	mux.Handle(path, handler)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv.URL
}

// staticFactory hands out clients against a fixed URL — the box-resolution
// seam is covered separately in factory_test.go.
type staticFactory struct{ url string }

func (s staticFactory) ClientFor(context.Context, string, string) (piv1connect.PiServiceClient, error) {
	return piv1connect.NewPiServiceClient(http.DefaultClient, s.url), nil
}

func piAgentProto(butterboxID, workingDir string) *agentsv1.Agent {
	return &agentsv1.Agent{
		Name:        "pi-coder",
		AgentId:     "pi-coder",
		WorkspaceId: "ws-1",
		Type:        agentsv1.AgentType_AGENT_TYPE_PI,
		Config: &agentsv1.AgentConfig{
			Pi: &agentsv1.PiAgentConfig{
				ButterboxId:   butterboxID,
				WorkingDir:    workingDir,
				Provider:      "anthropic",
				Model:         "claude-fable-5",
				ThinkingLevel: "high",
			},
		},
	}
}

// harness runs the bridge through a real ADK runner over an in-memory
// session service, so state deltas persist exactly as they do in production.
type harness struct {
	t        *testing.T
	runner   *adkrunner.Runner
	sessions adksession.Service
}

func newHarness(t *testing.T, b *Bridge) *harness {
	t.Helper()
	ag, err := b.BuildAgent("pi-coder", "test pi agent")
	if err != nil {
		t.Fatalf("BuildAgent: %v", err)
	}
	sessions := adksession.InMemoryService()
	r, err := adkrunner.New(adkrunner.Config{
		AppName:        "test-app",
		Agent:          ag,
		SessionService: sessions,
	})
	if err != nil {
		t.Fatalf("runner.New: %v", err)
	}
	if _, err := sessions.Create(context.Background(), &adksession.CreateRequest{
		AppName: "test-app", UserID: "u1", SessionID: "s1",
	}); err != nil {
		t.Fatalf("create session: %v", err)
	}
	return &harness{t: t, runner: r, sessions: sessions}
}

// turn sends one user content and returns the joined final text.
func (h *harness) turn(ctx context.Context, content *genai.Content) (string, error) {
	h.t.Helper()
	var out strings.Builder
	for evt, err := range h.runner.Run(ctx, "u1", "s1", content, agent.RunConfig{}) {
		if err != nil {
			return out.String(), err
		}
		if evt.Content != nil {
			for _, p := range evt.Content.Parts {
				if p.Text != "" {
					out.WriteString(p.Text)
				}
			}
		}
	}
	return out.String(), nil
}

func (h *harness) storedBinding() (binding, bool) {
	h.t.Helper()
	resp, err := h.sessions.Get(context.Background(), &adksession.GetRequest{
		AppName: "test-app", UserID: "u1", SessionID: "s1",
	})
	if err != nil {
		h.t.Fatalf("get session: %v", err)
	}
	return readBinding(resp.Session.State(), "pi-coder")
}

func textContent(s string) *genai.Content {
	return genai.NewContentFromText(s, genai.RoleUser)
}

func TestBridge_FirstTurnCreatesSessionAndAnswers(t *testing.T) {
	fake := newFakePi()
	b := NewBridge(piAgentProto("box-1", "projects/demo"), staticFactory{serveFake(t, fake)})
	h := newHarness(t, b)

	out, err := h.turn(t.Context(), textContent("hi pi"))
	if err != nil {
		t.Fatalf("turn: %v", err)
	}
	if out != "hello from pi" {
		t.Fatalf("output: got %q", out)
	}

	if fake.createCount() != 1 {
		t.Fatalf("create count: got %d", fake.createCount())
	}
	create := fake.createReqs[0]
	if create.GetCwd() != "projects/demo" || create.GetProvider() != "anthropic" ||
		create.GetModel() != "claude-fable-5" || create.GetThinkingLevel() != "high" {
		t.Fatalf("create request lost config: %v", create)
	}
	if !strings.HasPrefix(create.GetName(), "butter:pi-coder:") {
		t.Fatalf("session name: got %q", create.GetName())
	}
	if got := fake.submitReqs[0].GetMessage(); got != "hi pi" {
		t.Fatalf("submitted message: got %q", got)
	}

	bnd, ok := h.storedBinding()
	if !ok {
		t.Fatal("no binding stored in session state")
	}
	if bnd.PiSessionID != "pi-1" || bnd.ButterboxID != "box-1" || bnd.WorkingDir != "projects/demo" {
		t.Fatalf("binding: got %+v", bnd)
	}
}

func TestBridge_SecondTurnReusesPiSession(t *testing.T) {
	fake := newFakePi()
	b := NewBridge(piAgentProto("box-1", "projects/demo"), staticFactory{serveFake(t, fake)})
	h := newHarness(t, b)

	if _, err := h.turn(t.Context(), textContent("first")); err != nil {
		t.Fatalf("first turn: %v", err)
	}
	if _, err := h.turn(t.Context(), textContent("second")); err != nil {
		t.Fatalf("second turn: %v", err)
	}
	if fake.createCount() != 1 {
		t.Fatalf("expected one pi session across turns, got %d creates", fake.createCount())
	}
	if got := fake.submitReqs[1].GetSessionId(); got != "pi-1" {
		t.Fatalf("second submit session: got %q", got)
	}
}

func TestBridge_RepointedAgentAbandonsAndRecreates(t *testing.T) {
	fake := newFakePi()
	url := serveFake(t, fake)
	b := NewBridge(piAgentProto("box-1", "projects/demo"), staticFactory{url})
	h := newHarness(t, b)
	if _, err := h.turn(t.Context(), textContent("first")); err != nil {
		t.Fatalf("first turn: %v", err)
	}

	// Repoint the agent's working directory: same butter session, new bridge.
	b2 := NewBridge(piAgentProto("box-1", "projects/other"), staticFactory{url})
	ag2, err := b2.BuildAgent("pi-coder", "repointed")
	if err != nil {
		t.Fatalf("BuildAgent: %v", err)
	}
	r2, err := adkrunner.New(adkrunner.Config{AppName: "test-app", Agent: ag2, SessionService: h.sessions})
	if err != nil {
		t.Fatalf("runner.New: %v", err)
	}
	for _, err := range r2.Run(t.Context(), "u1", "s1", textContent("second"), agent.RunConfig{}) {
		if err != nil {
			t.Fatalf("second turn: %v", err)
		}
	}

	if fake.createCount() != 2 {
		t.Fatalf("expected recreate after repoint, got %d creates", fake.createCount())
	}
	if got := fake.createReqs[1].GetCwd(); got != "projects/other" {
		t.Fatalf("recreate cwd: got %q", got)
	}
	bnd, _ := h.storedBinding()
	if bnd.PiSessionID != "pi-2" || bnd.WorkingDir != "projects/other" {
		t.Fatalf("binding after repoint: got %+v", bnd)
	}
}

func TestBridge_BoxLostSessionRecreatesTransparently(t *testing.T) {
	fake := newFakePi()
	b := NewBridge(piAgentProto("box-1", "projects/demo"), staticFactory{serveFake(t, fake)})
	h := newHarness(t, b)
	if _, err := h.turn(t.Context(), textContent("first")); err != nil {
		t.Fatalf("first turn: %v", err)
	}

	// The box forgets the session (restart + purge): next submit is NotFound.
	fake.mu.Lock()
	delete(fake.sessions, "pi-1")
	fake.mu.Unlock()

	out, err := h.turn(t.Context(), textContent("second"))
	if err != nil {
		t.Fatalf("second turn: %v", err)
	}
	if out != "hello from pi" {
		t.Fatalf("output: got %q", out)
	}
	if fake.createCount() != 2 {
		t.Fatalf("expected transparent recreate, got %d creates", fake.createCount())
	}
	bnd, _ := h.storedBinding()
	if bnd.PiSessionID != "pi-2" {
		t.Fatalf("binding after recreate: got %+v", bnd)
	}
}

func TestBridge_CapacityIsActionable(t *testing.T) {
	fake := newFakePi()
	fake.createErr = connect.NewError(connect.CodeResourceExhausted, errors.New("session limit reached"))
	b := NewBridge(piAgentProto("box-1", ""), staticFactory{serveFake(t, fake)})
	h := newHarness(t, b)

	_, err := h.turn(t.Context(), textContent("hi"))
	if err == nil || !strings.Contains(err.Error(), "capacity") || !strings.Contains(err.Error(), "PI_API_MAX_SESSIONS") {
		t.Fatalf("expected actionable capacity error, got %v", err)
	}
}

func TestBridge_BusySessionIsActionable(t *testing.T) {
	fake := newFakePi()
	fake.submitErr = connect.NewError(connect.CodeFailedPrecondition, errors.New("session is processing another message"))
	b := NewBridge(piAgentProto("box-1", ""), staticFactory{serveFake(t, fake)})
	h := newHarness(t, b)

	_, err := h.turn(t.Context(), textContent("hi"))
	if err == nil || !strings.Contains(err.Error(), "busy") {
		t.Fatalf("expected actionable busy error, got %v", err)
	}
	// The session was created before the busy submit; the binding must
	// survive the failed turn so the next one reuses it.
	if _, ok := h.storedBinding(); !ok {
		t.Fatal("binding lost after failed turn")
	}
}

func TestBridge_DidNotFinishIsHonest(t *testing.T) {
	fake := newFakePi()
	fake.turnScript = []*piv1.GetTurnResponse{{Running: false, Result: nil}}
	b := NewBridge(piAgentProto("box-1", ""), staticFactory{serveFake(t, fake)})
	h := newHarness(t, b)

	_, err := h.turn(t.Context(), textContent("hi"))
	if err == nil || !strings.Contains(err.Error(), "did not finish") {
		t.Fatalf("expected did-not-finish error, got %v", err)
	}
}

func TestBridge_SessionLostMidRunIsDidNotFinish(t *testing.T) {
	fake := newFakePi()
	fake.turnErr = connect.NewError(connect.CodeNotFound, errors.New("session not found"))
	b := NewBridge(piAgentProto("box-1", ""), staticFactory{serveFake(t, fake)})
	h := newHarness(t, b)

	_, err := h.turn(t.Context(), textContent("hi"))
	if err == nil || !strings.Contains(err.Error(), "did not finish") {
		t.Fatalf("expected did-not-finish for a session lost mid-run, got %v", err)
	}
}

func TestBridge_FreshSessionLostIsActionable(t *testing.T) {
	fake := newFakePi()
	fake.submitErr = connect.NewError(connect.CodeNotFound, errors.New("session not found"))
	b := NewBridge(piAgentProto("box-1", ""), staticFactory{serveFake(t, fake)})
	h := newHarness(t, b)

	// The session is created and then the box immediately forgets it: the
	// bridge must not loop on recreates, and must not leak a raw not_found.
	_, err := h.turn(t.Context(), textContent("hi"))
	if err == nil || !strings.Contains(err.Error(), "freshly created") {
		t.Fatalf("expected unhealthy-box error, got %v", err)
	}
}

func TestBridge_CancelAbortsRunOnBox(t *testing.T) {
	fake := newFakePi()
	fake.turnScript = []*piv1.GetTurnResponse{{Running: true}}
	b := NewBridge(piAgentProto("box-1", ""), staticFactory{serveFake(t, fake)})
	b.pollWaitSeconds = 0
	b.pollRetryDelay = 10 * time.Millisecond
	h := newHarness(t, b)

	ctx, cancel := context.WithCancel(t.Context())
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()
	_, err := h.turn(ctx, textContent("hi"))
	if err == nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if fake.abortCount() != 1 {
		t.Fatalf("expected one AbortSession on the box, got %d", fake.abortCount())
	}
}

func TestBridge_MaxRunSecondsAbortsWithHint(t *testing.T) {
	fake := newFakePi()
	fake.turnScript = []*piv1.GetTurnResponse{{Running: true}}
	pb := piAgentProto("box-1", "")
	pb.Config.Pi.MaxRunSeconds = proto.Int32(1)
	b := NewBridge(pb, staticFactory{serveFake(t, fake)})
	b.pollWaitSeconds = 0
	b.pollRetryDelay = 10 * time.Millisecond
	h := newHarness(t, b)

	_, err := h.turn(t.Context(), textContent("hi"))
	if err == nil || !strings.Contains(err.Error(), "max_run_seconds=1") || !strings.Contains(err.Error(), "raise max_run_seconds") {
		t.Fatalf("expected raise-the-limit hint, got %v", err)
	}
	if fake.abortCount() != 1 {
		t.Fatalf("expected one AbortSession on the box, got %d", fake.abortCount())
	}
}

func TestBridge_ImagesPassThrough(t *testing.T) {
	fake := newFakePi()
	b := NewBridge(piAgentProto("box-1", ""), staticFactory{serveFake(t, fake)})
	h := newHarness(t, b)

	content := &genai.Content{Role: genai.RoleUser, Parts: []*genai.Part{
		{Text: "what is in this picture?"},
		{InlineData: &genai.Blob{MIMEType: "image/png", Data: []byte{0x89, 0x50, 0x4e, 0x47}}},
	}}
	if _, err := h.turn(t.Context(), content); err != nil {
		t.Fatalf("turn: %v", err)
	}
	images := fake.submitReqs[0].GetImages()
	if len(images) != 1 || images[0].GetMimeType() != "image/png" || len(images[0].GetData()) != 4 {
		t.Fatalf("images: got %v", images)
	}
}

func TestBridge_UnlimitedRunHasNoDeadline(t *testing.T) {
	pb := piAgentProto("box-1", "")
	pb.Config.Pi.MaxRunSeconds = proto.Int32(0)
	b := NewBridge(pb, staticFactory{""})
	if b.maxRun != 0 {
		t.Fatalf("explicit 0 must mean unlimited, got %v", b.maxRun)
	}
	if NewBridge(piAgentProto("box-1", ""), staticFactory{""}).maxRun != defaultMaxRunSeconds*time.Second {
		t.Fatal("unset max_run_seconds must default to 1800s")
	}
}
