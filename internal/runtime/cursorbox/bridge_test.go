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
	"go.mongodb.org/mongo-driver/v2/bson"
	"google.golang.org/adk/v2/agent"
	adkrunner "google.golang.org/adk/v2/runner"
	adksession "google.golang.org/adk/v2/session"
	"google.golang.org/genai"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/proto"

	"go.orx.me/apps/butter/pkg/proto/butterbox/cursor/v1"
	"go.orx.me/apps/butter/pkg/proto/butterbox/cursor/v1/cursorv1connect"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type fakeCursor struct {
	mu           sync.Mutex
	nextID       int
	sessions     map[string]bool
	createReqs   []*cursorv1.CreateSessionRequest
	sendReqs     []*cursorv1.SendMessageRequest
	abortedIDs   []string
	createErr    error
	sendErr      error
	sendDelay    time.Duration
	responseText string
}

func newFakeCursor() *fakeCursor {
	return &fakeCursor{
		sessions:     map[string]bool{},
		responseText: "hello from cursor",
	}
}

func (f *fakeCursor) CreateSession(_ context.Context, req *connect.Request[cursorv1.CreateSessionRequest]) (*connect.Response[cursorv1.CreateSessionResponse], error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.createErr != nil {
		return nil, f.createErr
	}
	f.nextID++
	id := fmt.Sprintf("cur-%d", f.nextID)
	f.sessions[id] = true
	f.createReqs = append(f.createReqs, proto.Clone(req.Msg).(*cursorv1.CreateSessionRequest))
	return connect.NewResponse(&cursorv1.CreateSessionResponse{SessionId: id}), nil
}

func (f *fakeCursor) SendMessage(ctx context.Context, req *connect.Request[cursorv1.SendMessageRequest]) (*connect.Response[cursorv1.SendMessageResponse], error) {
	f.mu.Lock()
	if f.sendErr != nil {
		err := f.sendErr
		f.mu.Unlock()
		return nil, err
	}
	if !f.sessions[req.Msg.GetSessionId()] {
		f.mu.Unlock()
		return nil, connect.NewError(connect.CodeNotFound, errors.New("session not found"))
	}
	f.sendReqs = append(f.sendReqs, proto.Clone(req.Msg).(*cursorv1.SendMessageRequest))
	delay := f.sendDelay
	text := f.responseText
	f.mu.Unlock()

	if delay > 0 {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
		}
	}
	return connect.NewResponse(&cursorv1.SendMessageResponse{Text: text}), nil
}

func (f *fakeCursor) AbortSession(_ context.Context, req *connect.Request[cursorv1.AbortSessionRequest]) (*connect.Response[cursorv1.AbortSessionResponse], error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.abortedIDs = append(f.abortedIDs, req.Msg.GetSessionId())
	return connect.NewResponse(&cursorv1.AbortSessionResponse{}), nil
}

func (f *fakeCursor) ListModels(_ context.Context, _ *connect.Request[cursorv1.ListModelsRequest]) (*connect.Response[cursorv1.ListModelsResponse], error) {
	return connect.NewResponse(&cursorv1.ListModelsResponse{Models: []*cursorv1.Model{
		{Id: "composer-2.5", Name: "Composer 2.5"},
		{Id: "auto-smart", Name: "Auto Smart"},
	}}), nil
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

type staticFactory struct {
	client cursorv1connect.CursorServiceClient
}

func (s staticFactory) ClientFor(context.Context, string, string) (cursorv1connect.CursorServiceClient, error) {
	return s.client, nil
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
				Mode:        "agent",
			},
		},
	}
}

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
	return readBinding(resp.Session.State(), "cursor-coder")
}

func textContent(s string) *genai.Content {
	return genai.NewContentFromText(s, genai.RoleUser)
}

func TestBridge_FirstTurnCreatesSessionAndAnswers(t *testing.T) {
	fake := newFakeCursor()
	b := NewBridge(cursorAgentProto("box-1", "projects/demo"), staticFactory{fake})
	h := newHarness(t, b)

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
	if create.GetCwd() != "projects/demo" || create.GetModel() != "composer-2.5" || create.GetMode() != "agent" {
		t.Fatalf("create request lost config: %+v", create)
	}
	if !strings.HasPrefix(create.GetName(), "butter:cursor-coder:") {
		t.Fatalf("session name: got %q", create.GetName())
	}
	if got := fake.sendReqs[0].GetMessage(); got != "hi cursor" {
		t.Fatalf("submitted message: got %q", got)
	}

	bnd, ok := h.storedBinding()
	if !ok {
		t.Fatal("no binding stored in session state")
	}
	if bnd.CursorSessionID != "cur-1" || bnd.ButterboxID != "box-1" || bnd.WorkingDir != "projects/demo" {
		t.Fatalf("binding: got %+v", bnd)
	}
}

func TestBridge_SecondTurnReusesCursorSession(t *testing.T) {
	fake := newFakeCursor()
	b := NewBridge(cursorAgentProto("box-1", "projects/demo"), staticFactory{fake})
	h := newHarness(t, b)

	if _, err := h.turn(t.Context(), textContent("first")); err != nil {
		t.Fatalf("first turn: %v", err)
	}
	if _, err := h.turn(t.Context(), textContent("second")); err != nil {
		t.Fatalf("second turn: %v", err)
	}
	if fake.createCount() != 1 {
		t.Fatalf("expected one cursor session across turns, got %d creates", fake.createCount())
	}
	if got := fake.sendReqs[1].GetSessionId(); got != "cur-1" {
		t.Fatalf("second send session: got %q", got)
	}
}

func TestBridge_SecondTurnReusesCursorSessionAfterBSONRoundTrip(t *testing.T) {
	fake := newFakeCursor()
	b := NewBridge(cursorAgentProto("box-1", "projects/demo"), staticFactory{fake})
	h := newHarness(t, b)

	if _, err := h.turn(t.Context(), textContent("first")); err != nil {
		t.Fatalf("first turn: %v", err)
	}

	resp, err := h.sessions.Get(t.Context(), &adksession.GetRequest{
		AppName: "test-app", UserID: "u1", SessionID: "s1",
	})
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
		t.Fatalf("marshal session state: %v", err)
	}
	var persisted struct {
		State map[string]any `bson:"state"`
	}
	if err := bson.Unmarshal(encoded, &persisted); err != nil {
		t.Fatalf("unmarshal session state: %v", err)
	}
	roundTripped := persisted.State[stateKey("cursor-coder")]
	if _, ok := roundTripped.(bson.D); !ok {
		t.Fatalf("round-tripped binding type = %T, want bson.D", roundTripped)
	}
	if err := resp.Session.State().Set(stateKey("cursor-coder"), roundTripped); err != nil {
		t.Fatalf("replace binding state: %v", err)
	}

	if _, err := h.turn(t.Context(), textContent("second")); err != nil {
		t.Fatalf("second turn: %v", err)
	}
	if fake.createCount() != 1 {
		t.Fatalf("expected BSON-loaded binding to reuse the cursor session, got %d creates", fake.createCount())
	}
	if got := fake.sendReqs[1].GetSessionId(); got != "cur-1" {
		t.Fatalf("second send session: got %q", got)
	}
}

func TestBridge_RepointedAgentAbandonsAndRecreates(t *testing.T) {
	fake := newFakeCursor()
	b := NewBridge(cursorAgentProto("box-1", "projects/demo"), staticFactory{fake})
	h := newHarness(t, b)
	if _, err := h.turn(t.Context(), textContent("first")); err != nil {
		t.Fatalf("first turn: %v", err)
	}

	b2 := NewBridge(cursorAgentProto("box-1", "projects/other"), staticFactory{fake})
	ag2, err := b2.BuildAgent("cursor-coder", "repointed")
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
	if bnd.CursorSessionID != "cur-2" || bnd.WorkingDir != "projects/other" {
		t.Fatalf("binding after repoint: got %+v", bnd)
	}
}

func TestBridge_BoxLostSessionRecreatesTransparently(t *testing.T) {
	fake := newFakeCursor()
	b := NewBridge(cursorAgentProto("box-1", "projects/demo"), staticFactory{fake})
	h := newHarness(t, b)
	if _, err := h.turn(t.Context(), textContent("first")); err != nil {
		t.Fatalf("first turn: %v", err)
	}

	fake.mu.Lock()
	delete(fake.sessions, "cur-1")
	fake.mu.Unlock()

	out, err := h.turn(t.Context(), textContent("second"))
	if err != nil {
		t.Fatalf("second turn: %v", err)
	}
	if out != "hello from cursor" {
		t.Fatalf("output: got %q", out)
	}
	if fake.createCount() != 2 {
		t.Fatalf("expected transparent recreate, got %d creates", fake.createCount())
	}
	bnd, _ := h.storedBinding()
	if bnd.CursorSessionID != "cur-2" {
		t.Fatalf("binding after recreate: got %+v", bnd)
	}
}

func TestBridge_CapacityIsActionable(t *testing.T) {
	fake := newFakeCursor()
	fake.createErr = connect.NewError(connect.CodeResourceExhausted, errors.New("session limit reached"))
	b := NewBridge(cursorAgentProto("box-1", ""), staticFactory{fake})
	h := newHarness(t, b)

	_, err := h.turn(t.Context(), textContent("hi"))
	if err == nil || !strings.Contains(err.Error(), "capacity") || !strings.Contains(err.Error(), "CURSOR_MAX_SESSIONS") {
		t.Fatalf("expected actionable capacity error, got %v", err)
	}
}

func TestBridge_BusySessionIsActionable(t *testing.T) {
	fake := newFakeCursor()
	fake.sendErr = connect.NewError(connect.CodeFailedPrecondition, errors.New("session is processing another message"))
	b := NewBridge(cursorAgentProto("box-1", ""), staticFactory{fake})
	h := newHarness(t, b)

	_, err := h.turn(t.Context(), textContent("hi"))
	if err == nil || !strings.Contains(err.Error(), "busy") {
		t.Fatalf("expected actionable busy error, got %v", err)
	}
	if _, ok := h.storedBinding(); !ok {
		t.Fatal("binding lost after failed turn")
	}
}

func TestBridge_FreshSessionLostIsActionable(t *testing.T) {
	fake := newFakeCursor()
	fake.sendErr = connect.NewError(connect.CodeNotFound, errors.New("session not found"))
	b := NewBridge(cursorAgentProto("box-1", ""), staticFactory{fake})
	h := newHarness(t, b)

	_, err := h.turn(t.Context(), textContent("hi"))
	if err == nil || !strings.Contains(err.Error(), "freshly created") {
		t.Fatalf("expected unhealthy-box error, got %v", err)
	}
}

func TestBridge_CancelAbortsRunOnBox(t *testing.T) {
	fake := newFakeCursor()
	fake.sendDelay = 5 * time.Second
	b := NewBridge(cursorAgentProto("box-1", ""), staticFactory{fake})
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
	time.Sleep(50 * time.Millisecond) // let abort goroutine fire
	if fake.abortCount() != 1 {
		t.Fatalf("expected one AbortSession on the box, got %d", fake.abortCount())
	}
}

func TestBridge_MaxRunSecondsAbortsWithHint(t *testing.T) {
	fake := newFakeCursor()
	fake.sendDelay = 5 * time.Second
	pb := cursorAgentProto("box-1", "")
	pb.Config.Cursor.MaxRunSeconds = proto.Int32(1)
	b := NewBridge(pb, staticFactory{fake})
	h := newHarness(t, b)

	_, err := h.turn(t.Context(), textContent("hi"))
	if err == nil || !strings.Contains(err.Error(), "max_run_seconds=1") || !strings.Contains(err.Error(), "raise max_run_seconds") {
		t.Fatalf("expected raise-the-limit hint, got %v", err)
	}
}

func TestBridge_ImagesPassThrough(t *testing.T) {
	fake := newFakeCursor()
	b := NewBridge(cursorAgentProto("box-1", ""), staticFactory{fake})
	h := newHarness(t, b)

	content := &genai.Content{Role: genai.RoleUser, Parts: []*genai.Part{
		{Text: "what is in this picture?"},
		{InlineData: &genai.Blob{MIMEType: "image/png", Data: []byte{0x89, 0x50, 0x4e, 0x47}}},
	}}
	if _, err := h.turn(t.Context(), content); err != nil {
		t.Fatalf("turn: %v", err)
	}
	images := fake.sendReqs[0].GetImages()
	if len(images) != 1 || images[0].GetMimeType() != "image/png" || len(images[0].GetData()) != 4 {
		t.Fatalf("images: got %v", images)
	}
}

func TestBridge_UnlimitedRunHasNoDeadline(t *testing.T) {
	pb := cursorAgentProto("box-1", "")
	pb.Config.Cursor.MaxRunSeconds = proto.Int32(0)
	b := NewBridge(pb, staticFactory{newFakeCursor()})
	if b.maxRun != 0 {
		t.Fatalf("explicit 0 must mean unlimited, got %v", b.maxRun)
	}
	if NewBridge(cursorAgentProto("box-1", ""), staticFactory{newFakeCursor()}).maxRun != defaultMaxRunSeconds*time.Second {
		t.Fatal("unset max_run_seconds must default to 1800s")
	}
}

// TestBridge_GeneratedHandlerOverWire drives the full bridge through the
// generated Connect handler mounted on an httptest server — the wire path the
// production Factory produces — proving the cursor.v1 messages actually
// serialize and the session lifecycle survives the round trip. This is the
// check #316 requires (typed fake CursorService over httptest); the previous
// bridge tests use the fake as an in-process client, which never exercises
// the codec.
func TestBridge_GeneratedHandlerOverWire(t *testing.T) {
	serverCursor := newFakeCursor()
	path, handler := cursorv1connect.NewCursorServiceHandler(serverCursor)
	mux := http.NewServeMux()
	mux.Handle(path, handler)
	server := httptest.NewServer(mux)
	defer server.Close()

	client := cursorv1connect.NewCursorServiceClient(server.Client(), server.URL)
	b := NewBridge(cursorAgentProto("box-1", "projects/demo"), staticFactory{client})
	h := newHarness(t, b)

	if _, err := h.turn(t.Context(), textContent("first")); err != nil {
		t.Fatalf("first turn: %v", err)
	}
	if _, err := h.turn(t.Context(), textContent("second")); err != nil {
		t.Fatalf("second turn: %v", err)
	}
	if serverCursor.createCount() != 1 {
		t.Fatalf("expected one cursor session across wire turns, got %d creates", serverCursor.createCount())
	}
	if got := serverCursor.sendReqs[1].GetSessionId(); got != "cur-1" {
		t.Fatalf("second wire send session: got %q", got)
	}
	if got := serverCursor.sendReqs[0].GetMessage(); got != "first" {
		t.Fatalf("wire message: got %q", got)
	}
}

// TestBridge_CursorAPIKeyInvalidIsActionable verifies the #316/#317 error
// path: a box that signals a missing or invalid CURSOR_API_KEY (Unauthenticated
// carrying the documented google.rpc.ErrorInfo reason) yields guidance to
// configure the key on the box, while a plain rejected access token still maps
// to rotating the ButterBox token.
func TestBridge_CursorAPIKeyInvalidIsActionable(t *testing.T) {
	fake := newFakeCursor()
	detail, derr := connect.NewErrorDetail(&errdetails.ErrorInfo{
		Reason: cursorAPIKeyInvalidReason,
		Domain: "butterbox.cursor.v1",
	})
	if derr != nil {
		t.Fatalf("error detail: %v", derr)
	}
	fake.sendErr = connect.NewError(connect.CodeUnauthenticated, errors.New("cursor api key rejected"))
	fake.sendErr.(*connect.Error).AddDetail(detail)
	b := NewBridge(cursorAgentProto("box-1", ""), staticFactory{fake})
	h := newHarness(t, b)

	_, err := h.turn(t.Context(), textContent("hi"))
	if err == nil || !strings.Contains(err.Error(), "CURSOR_API_KEY") {
		t.Fatalf("expected configure-the-angel-key guidance, got %v", err)
	}
	if strings.Contains(err.Error(), "SetButterBoxToken") {
		t.Fatalf("API-key failure must not be reported as a box-token problem: %v", err)
	}

	// A rejected box access token (no ErrorInfo reason) still maps to the
	// rotate-the-token guidance.
	tokenFake := newFakeCursor()
	tokenFake.createErr = connect.NewError(connect.CodeUnauthenticated, errors.New("bad bearer"))
	tokenBridge := NewBridge(cursorAgentProto("box-1", ""), staticFactory{tokenFake})
	tokenHarness := newHarness(t, tokenBridge)
	_, tokenErr := tokenHarness.turn(t.Context(), textContent("hi"))
	if tokenErr == nil || !strings.Contains(tokenErr.Error(), "SetButterBoxToken") {
		t.Fatalf("box token rejection must map to rotate-the-token, got %v", tokenErr)
	}
}

// TestBridge_CursorAPIKeyInvalidOverWire is the same distinguishable-error
// check over the generated Connect handler: the box answers Unauthenticated
// with the CURSOR_API_KEY ErrorInfo reason, the client decodes the wire
// error, and the bridge still maps it to configure-the-key guidance.
func TestBridge_CursorAPIKeyInvalidOverWire(t *testing.T) {
	serverCursor := newFakeCursor()
	detail, derr := connect.NewErrorDetail(&errdetails.ErrorInfo{
		Reason: cursorAPIKeyInvalidReason,
		Domain: "butterbox.cursor.v1",
	})
	if derr != nil {
		t.Fatalf("error detail: %v", derr)
	}
	apikeyErr := connect.NewError(connect.CodeUnauthenticated, errors.New("cursor api key rejected"))
	apikeyErr.AddDetail(detail)
	serverCursor.sendErr = apikeyErr

	path, handler := cursorv1connect.NewCursorServiceHandler(serverCursor)
	mux := http.NewServeMux()
	mux.Handle(path, handler)
	server := httptest.NewServer(mux)
	defer server.Close()

	client := cursorv1connect.NewCursorServiceClient(server.Client(), server.URL)
	b := NewBridge(cursorAgentProto("box-1", ""), staticFactory{client})
	h := newHarness(t, b)

	_, err := h.turn(t.Context(), textContent("hi"))
	if err == nil || !strings.Contains(err.Error(), "CURSOR_API_KEY") {
		t.Fatalf("expected configure-the-angel-key guidance over the wire, got %v", err)
	}
	if strings.Contains(err.Error(), "SetButterBoxToken") {
		t.Fatalf("API-key failure must not be reported as a box-token problem: %v", err)
	}
}
