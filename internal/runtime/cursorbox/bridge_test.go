package cursorbox

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
	cursorv1 "github.com/orvice/butter-box/pkg/proto/butterbox/cursor/v1"
	"github.com/orvice/butter-box/pkg/proto/butterbox/cursor/v1/cursorv1connect"
	"go.mongodb.org/mongo-driver/v2/bson"
	"google.golang.org/adk/v2/agent"
	adkrunner "google.golang.org/adk/v2/runner"
	adksession "google.golang.org/adk/v2/session"
	"google.golang.org/genai"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/proto"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// fakeCursor is a typed CursorService fake served through the real generated
// Connect handler, so the bridge is exercised against actual wire behavior.
type fakeCursor struct {
	cursorv1connect.UnimplementedCursorServiceHandler

	mu         sync.Mutex
	nextID     int
	sessions   map[string]bool
	createReqs []*cursorv1.CreateSessionRequest
	sendReqs   []*cursorv1.SendMessageRequest
	abortedIDs []string
	createErr  error
	sendErr    error
	// block makes SendMessage hold until the caller's context ends, the way a
	// long Cursor run holds the call on the real box.
	block    bool
	sendText string
	gotAuth  string
}

func newFakeCursor() *fakeCursor {
	return &fakeCursor{sessions: map[string]bool{}, sendText: "hello from cursor"}
}

func (f *fakeCursor) CreateSession(_ context.Context, req *connect.Request[cursorv1.CreateSessionRequest]) (*connect.Response[cursorv1.CreateSessionResponse], error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.gotAuth = req.Header().Get("Authorization")
	if f.createErr != nil {
		return nil, f.createErr
	}
	f.nextID++
	id := fmt.Sprintf("agent-%d", f.nextID)
	f.sessions[id] = true
	f.createReqs = append(f.createReqs, proto.Clone(req.Msg).(*cursorv1.CreateSessionRequest))
	return connect.NewResponse(&cursorv1.CreateSessionResponse{SessionId: id}), nil
}

func (f *fakeCursor) SendMessage(ctx context.Context, req *connect.Request[cursorv1.SendMessageRequest]) (*connect.Response[cursorv1.SendMessageResponse], error) {
	f.mu.Lock()
	f.sendReqs = append(f.sendReqs, proto.Clone(req.Msg).(*cursorv1.SendMessageRequest))
	sendErr, block, known, text := f.sendErr, f.block, f.sessions[req.Msg.GetSessionId()], f.sendText
	f.mu.Unlock()
	if sendErr != nil {
		return nil, sendErr
	}
	if !known {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("cursor session not found"))
	}
	if block {
		<-ctx.Done()
		return nil, connect.NewError(connect.CodeCanceled, ctx.Err())
	}
	return connect.NewResponse(&cursorv1.SendMessageResponse{Text: text}), nil
}

func (f *fakeCursor) AbortSession(_ context.Context, req *connect.Request[cursorv1.AbortSessionRequest]) (*connect.Response[cursorv1.AbortSessionResponse], error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.abortedIDs = append(f.abortedIDs, req.Msg.GetSessionId())
	return connect.NewResponse(&cursorv1.AbortSessionResponse{}), nil
}

func (f *fakeCursor) createCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.createReqs)
}

func (f *fakeCursor) abortCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.abortedIDs)
}

func serveFake(t *testing.T, f *fakeCursor) string {
	t.Helper()
	path, handler := cursorv1connect.NewCursorServiceHandler(f)
	mux := http.NewServeMux()
	mux.Handle(path, handler)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv.URL
}

// staticFactory hands out clients against a fixed URL — the box-resolution
// seam is covered separately in factory_test.go.
type staticFactory struct{ url string }

func (s staticFactory) ClientFor(context.Context, string, string) (cursorv1connect.CursorServiceClient, error) {
	return cursorv1connect.NewCursorServiceClient(http.DefaultClient, s.url), nil
}

func cursorAgentProto(butterboxID, workingDir string) *agentsv1.Agent {
	return &agentsv1.Agent{
		Name:        "cursor-coder",
		AgentId:     "cursor-coder",
		WorkspaceId: "ws-1",
		Type:        agentsv1.AgentType_AGENT_TYPE_CURSOR,
		Config: &agentsv1.AgentConfig{
			Cursor: &agentsv1.CursorAgentConfig{
				ButterboxId: butterboxID,
				WorkingDir:  workingDir,
				Model:       "composer-2.5",
				Mode:        "plan",
			},
		},
	}
}

// harness runs the bridge through a real ADK runner over an in-memory
// session service.
type harness struct {
	t        *testing.T
	runner   *adkrunner.Runner
	sessions adksession.Service
}

func newHarness(t *testing.T, b *Bridge) *harness {
	t.Helper()
	ag, err := b.BuildAgent("cursor-coder", "test cursor agent")
	if err != nil {
		t.Fatalf("BuildAgent: %v", err)
	}
	sessions := adksession.InMemoryService()
	r, err := adkrunner.New(adkrunner.Config{AppName: "test-app", Agent: ag, SessionService: sessions})
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

func (h *harness) turnWith(ctx context.Context, r *adkrunner.Runner, content *genai.Content) (string, error) {
	h.t.Helper()
	var out strings.Builder
	for evt, err := range r.Run(ctx, "u1", "s1", content, agent.RunConfig{}) {
		if err != nil {
			return out.String(), err
		}
		if evt.Content != nil {
			for _, p := range evt.Content.Parts {
				out.WriteString(p.Text)
			}
		}
	}
	return out.String(), nil
}

func (h *harness) turn(ctx context.Context, content *genai.Content) (string, error) {
	h.t.Helper()
	return h.turnWith(ctx, h.runner, content)
}

func (h *harness) storedBinding() (binding, bool) {
	h.t.Helper()
	resp, err := h.sessions.Get(context.Background(), &adksession.GetRequest{
		AppName: "test-app", UserID: "u1", SessionID: "s1",
	})
	if err != nil {
		h.t.Fatalf("get session: %v", err)
	}
	return readBinding(resp.Session.State(), "cursor-coder")
}

func textContent(s string) *genai.Content {
	return genai.NewContentFromText(s, genai.RoleUser)
}

func TestBridge_FirstTurnCreatesSessionAndAnswers(t *testing.T) {
	fake := newFakeCursor()
	h := newHarness(t, NewBridge(cursorAgentProto("box-1", "projects/demo"), staticFactory{serveFake(t, fake)}))

	out, err := h.turn(t.Context(), textContent("hi cursor"))
	if err != nil {
		t.Fatalf("turn: %v", err)
	}
	if out != "hello from cursor" {
		t.Fatalf("output: got %q", out)
	}
	if fake.createCount() != 1 {
		t.Fatalf("create count: got %d", fake.createCount())
	}
	create := fake.createReqs[0]
	if create.GetCwd() != "projects/demo" || create.GetModel() != "composer-2.5" || create.GetMode() != "plan" {
		t.Fatalf("create request lost config: %v", create)
	}
	if !strings.HasPrefix(create.GetName(), "butter:cursor-coder:") {
		t.Fatalf("session name: got %q", create.GetName())
	}
	if got := fake.sendReqs[0]; got.GetMessage() != "hi cursor" || got.GetSessionId() != "agent-1" {
		t.Fatalf("send request: got %v", got)
	}
	bnd, ok := h.storedBinding()
	if !ok {
		t.Fatal("no binding stored in session state")
	}
	if bnd.CursorSessionID != "agent-1" || bnd.ButterboxID != "box-1" || bnd.WorkingDir != "projects/demo" {
		t.Fatalf("binding: got %+v", bnd)
	}
}

func TestBridge_SecondTurnReusesCursorSession(t *testing.T) {
	fake := newFakeCursor()
	h := newHarness(t, NewBridge(cursorAgentProto("box-1", "projects/demo"), staticFactory{serveFake(t, fake)}))

	for _, msg := range []string{"first", "second"} {
		if _, err := h.turn(t.Context(), textContent(msg)); err != nil {
			t.Fatalf("turn %q: %v", msg, err)
		}
	}
	if fake.createCount() != 1 {
		t.Fatalf("expected one cursor session across turns, got %d creates", fake.createCount())
	}
	if got := fake.sendReqs[1].GetSessionId(); got != "agent-1" {
		t.Fatalf("second send session: got %q", got)
	}
}

func TestBridge_SecondTurnReusesSessionAfterBSONRoundTrip(t *testing.T) {
	fake := newFakeCursor()
	h := newHarness(t, NewBridge(cursorAgentProto("box-1", "projects/demo"), staticFactory{serveFake(t, fake)}))
	if _, err := h.turn(t.Context(), textContent("first")); err != nil {
		t.Fatalf("first turn: %v", err)
	}

	resp, err := h.sessions.Get(t.Context(), &adksession.GetRequest{AppName: "test-app", UserID: "u1", SessionID: "s1"})
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	value, err := resp.Session.State().Get(stateKey("cursor-coder"))
	if err != nil {
		t.Fatalf("get binding state: %v", err)
	}
	encoded, err := bson.Marshal(struct {
		State map[string]any `bson:"state"`
	}{State: map[string]any{stateKey("cursor-coder"): value}})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var persisted struct {
		State map[string]any `bson:"state"`
	}
	if err := bson.Unmarshal(encoded, &persisted); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if err := resp.Session.State().Set(stateKey("cursor-coder"), persisted.State[stateKey("cursor-coder")]); err != nil {
		t.Fatalf("replace binding state: %v", err)
	}

	if _, err := h.turn(t.Context(), textContent("second")); err != nil {
		t.Fatalf("second turn: %v", err)
	}
	if fake.createCount() != 1 {
		t.Fatalf("expected BSON-loaded binding to reuse the session, got %d creates", fake.createCount())
	}
}

func TestBridge_RepointedAgentAbandonsAndRecreates(t *testing.T) {
	fake := newFakeCursor()
	url := serveFake(t, fake)
	h := newHarness(t, NewBridge(cursorAgentProto("box-1", "projects/demo"), staticFactory{url}))
	if _, err := h.turn(t.Context(), textContent("first")); err != nil {
		t.Fatalf("first turn: %v", err)
	}

	ag2, err := NewBridge(cursorAgentProto("box-1", "projects/other"), staticFactory{url}).BuildAgent("cursor-coder", "repointed")
	if err != nil {
		t.Fatalf("BuildAgent: %v", err)
	}
	r2, err := adkrunner.New(adkrunner.Config{AppName: "test-app", Agent: ag2, SessionService: h.sessions})
	if err != nil {
		t.Fatalf("runner.New: %v", err)
	}
	if _, err := h.turnWith(t.Context(), r2, textContent("second")); err != nil {
		t.Fatalf("second turn: %v", err)
	}
	if fake.createCount() != 2 || fake.createReqs[1].GetCwd() != "projects/other" {
		t.Fatalf("expected recreate in the new directory, got %d creates", fake.createCount())
	}
	if bnd, _ := h.storedBinding(); bnd.CursorSessionID != "agent-2" || bnd.WorkingDir != "projects/other" {
		t.Fatalf("binding after repoint: got %+v", bnd)
	}
}

func TestBridge_BoxLostSessionRecreatesTransparently(t *testing.T) {
	fake := newFakeCursor()
	h := newHarness(t, NewBridge(cursorAgentProto("box-1", "projects/demo"), staticFactory{serveFake(t, fake)}))
	if _, err := h.turn(t.Context(), textContent("first")); err != nil {
		t.Fatalf("first turn: %v", err)
	}
	fake.mu.Lock()
	delete(fake.sessions, "agent-1") // box restart: the session is unknown
	fake.mu.Unlock()

	out, err := h.turn(t.Context(), textContent("second"))
	if err != nil {
		t.Fatalf("second turn: %v", err)
	}
	if out != "hello from cursor" || fake.createCount() != 2 {
		t.Fatalf("expected transparent recreate, got out=%q creates=%d", out, fake.createCount())
	}
	if bnd, _ := h.storedBinding(); bnd.CursorSessionID != "agent-2" {
		t.Fatalf("binding after recreate: got %+v", bnd)
	}
}

func TestBridge_FreshSessionLostIsActionable(t *testing.T) {
	fake := newFakeCursor()
	fake.sendErr = connect.NewError(connect.CodeNotFound, errors.New("cursor session not found"))
	h := newHarness(t, NewBridge(cursorAgentProto("box-1", ""), staticFactory{serveFake(t, fake)}))

	_, err := h.turn(t.Context(), textContent("hi"))
	if err == nil || !strings.Contains(err.Error(), "freshly created") {
		t.Fatalf("expected unhealthy-box error, got %v", err)
	}
	if fake.createCount() != 1 {
		t.Fatalf("must not loop on recreates, got %d creates", fake.createCount())
	}
}

func TestBridge_BoxErrorsAreActionable(t *testing.T) {
	apiKeyErr := connect.NewError(connect.CodeUnauthenticated, errors.New("cursor API key is missing or invalid"))
	detail, err := connect.NewErrorDetail(&errdetails.ErrorInfo{Reason: "CURSOR_API_KEY_MISSING_OR_INVALID"})
	if err != nil {
		t.Fatal(err)
	}
	apiKeyErr.AddDetail(detail)

	cases := []struct {
		name      string
		createErr error
		sendErr   error
		want      []string
	}{
		{"capacity", connect.NewError(connect.CodeResourceExhausted, errors.New("active cursor session limit reached")), nil, []string{"capacity", "CURSOR_MAX_SESSIONS"}},
		{"busy", nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("busy")), []string{"busy"}},
		{"api key", apiKeyErr, nil, []string{"CURSOR_API_KEY", "on the box"}},
		{"box token", connect.NewError(connect.CodeUnauthenticated, errors.New("unauthorized")), nil, []string{"access token", "SetButterBoxToken"}},
		{"bad cwd", connect.NewError(connect.CodeInvalidArgument, errors.New("invalid cursor working directory")), nil, []string{"working_dir"}},
		{"bridge exited", nil, connect.NewError(connect.CodeUnavailable, errors.New("bridge exited")), []string{"did not finish"}},
		{"box cancelled", nil, connect.NewError(connect.CodeCanceled, errors.New("cursor run cancelled")), []string{"cancelled on the box"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fake := newFakeCursor()
			fake.createErr, fake.sendErr = tc.createErr, tc.sendErr
			h := newHarness(t, NewBridge(cursorAgentProto("box-1", ""), staticFactory{serveFake(t, fake)}))
			_, err := h.turn(t.Context(), textContent("hi"))
			if err == nil {
				t.Fatal("expected an error")
			}
			for _, w := range tc.want {
				if !strings.Contains(err.Error(), w) {
					t.Fatalf("error %q does not mention %q", err, w)
				}
			}
		})
	}
}

func TestBridge_BindingSurvivesFailedTurn(t *testing.T) {
	fake := newFakeCursor()
	fake.sendErr = connect.NewError(connect.CodeFailedPrecondition, errors.New("busy"))
	h := newHarness(t, NewBridge(cursorAgentProto("box-1", ""), staticFactory{serveFake(t, fake)}))
	if _, err := h.turn(t.Context(), textContent("hi")); err == nil {
		t.Fatal("expected busy error")
	}
	if _, ok := h.storedBinding(); !ok {
		t.Fatal("binding lost after failed turn")
	}
}

func TestBridge_UnreachableBoxIsActionable(t *testing.T) {
	h := newHarness(t, NewBridge(cursorAgentProto("box-1", ""), staticFactory{"http://127.0.0.1:1"}))
	_, err := h.turn(t.Context(), textContent("hi"))
	if err == nil || !strings.Contains(err.Error(), "unreachable") {
		t.Fatalf("expected unreachable-box error, got %v", err)
	}
}

func TestBridge_CancelAbortsRunOnBox(t *testing.T) {
	fake := newFakeCursor()
	fake.block = true
	h := newHarness(t, NewBridge(cursorAgentProto("box-1", ""), staticFactory{serveFake(t, fake)}))

	ctx, cancel := context.WithCancel(t.Context())
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()
	_, err := h.turn(ctx, textContent("hi"))
	if err == nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if fake.abortCount() != 1 || fake.abortedIDs[0] != "agent-1" {
		t.Fatalf("expected one AbortSession for agent-1, got %v", fake.abortedIDs)
	}
}

func TestBridge_MaxRunSecondsAbortsWithHint(t *testing.T) {
	fake := newFakeCursor()
	fake.block = true
	pb := cursorAgentProto("box-1", "")
	pb.Config.Cursor.MaxRunSeconds = proto.Int32(1)
	h := newHarness(t, NewBridge(pb, staticFactory{serveFake(t, fake)}))

	_, err := h.turn(t.Context(), textContent("hi"))
	if err == nil || !strings.Contains(err.Error(), "max_run_seconds=1") || !strings.Contains(err.Error(), "raise max_run_seconds") {
		t.Fatalf("expected raise-the-limit hint, got %v", err)
	}
	if fake.abortCount() != 1 {
		t.Fatalf("expected one AbortSession on the box, got %d", fake.abortCount())
	}
}

func TestBridge_ImagesPassThrough(t *testing.T) {
	fake := newFakeCursor()
	h := newHarness(t, NewBridge(cursorAgentProto("box-1", ""), staticFactory{serveFake(t, fake)}))
	content := &genai.Content{Role: genai.RoleUser, Parts: []*genai.Part{
		{Text: "what is in this picture?"},
		{InlineData: &genai.Blob{MIMEType: "image/png", Data: []byte{0x89, 0x50, 0x4e, 0x47}}},
		{InlineData: &genai.Blob{MIMEType: "application/pdf", Data: []byte{1}}},
	}}
	if _, err := h.turn(t.Context(), content); err != nil {
		t.Fatalf("turn: %v", err)
	}
	images := fake.sendReqs[0].GetImages()
	if len(images) != 1 || images[0].GetMimeType() != "image/png" || len(images[0].GetData()) != 4 {
		t.Fatalf("images: got %v", images)
	}
}

func TestBridge_EmptyInputIsRejected(t *testing.T) {
	fake := newFakeCursor()
	h := newHarness(t, NewBridge(cursorAgentProto("box-1", ""), staticFactory{serveFake(t, fake)}))
	if _, err := h.turn(t.Context(), &genai.Content{Role: genai.RoleUser, Parts: []*genai.Part{{Text: ""}}}); err == nil {
		t.Fatal("expected empty-input error")
	}
	if fake.createCount() != 0 {
		t.Fatal("empty input must not create a session")
	}
}

func TestNewBridge_MaxRunDefaults(t *testing.T) {
	pb := cursorAgentProto("box-1", "")
	pb.Config.Cursor.MaxRunSeconds = proto.Int32(0)
	if b := NewBridge(pb, staticFactory{}); b.maxRun != 0 {
		t.Fatalf("explicit 0 must mean unlimited, got %v", b.maxRun)
	}
	if b := NewBridge(cursorAgentProto("box-1", ""), staticFactory{}); b.maxRun != defaultMaxRunSeconds*time.Second {
		t.Fatal("unset max_run_seconds must default to 1800s")
	}
}
