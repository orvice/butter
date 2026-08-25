package pibox

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"strings"
	"sync"
	"testing"
	"time"

	"connectrpc.com/connect"
	piv1 "github.com/orvice/butter-box/pkg/proto/butterbox/pi/v1"
	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"
)

// --- test doubles ---

type fakeState struct {
	mu   sync.Mutex
	data map[string]any
}

func newFakeState() *fakeState {
	return &fakeState{data: make(map[string]any)}
}

func (s *fakeState) Get(key string) (any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.data[key]
	if !ok {
		return nil, session.ErrStateKeyNotExist
	}
	return v, nil
}

func (s *fakeState) Set(key string, val any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = val
	return nil
}

func (s *fakeState) All() iter.Seq2[string, any] {
	return func(yield func(string, any) bool) {
		s.mu.Lock()
		defer s.mu.Unlock()
		for k, v := range s.data {
			if !yield(k, v) {
				return
			}
		}
	}
}

type fakeSession struct {
	id    string
	state *fakeState
}

func (s *fakeSession) ID() string                     { return s.id }
func (s *fakeSession) AppName() string                { return "test-app" }
func (s *fakeSession) UserID() string                 { return "user-1" }
func (s *fakeSession) State() session.State           { return s.state }
func (s *fakeSession) Events() session.Events         { return nil }
func (s *fakeSession) LastUpdateTime() time.Time      { return time.Now() }

type testInvocationContext struct {
	context.Context
	sess         session.Session
	ag           agent.Agent
	userContent  *genai.Content
	invocationID string
	ended        bool
}

func newTestInvocationContext(t *testing.T, userContent *genai.Content) *testInvocationContext {
	t.Helper()
	ag, err := agent.New(agent.Config{Name: "pi-agent", Description: "test"})
	if err != nil {
		t.Fatalf("agent.New: %v", err)
	}
	return &testInvocationContext{
		Context:      context.Background(),
		sess:         &fakeSession{id: "sess-1", state: newFakeState()},
		ag:           ag,
		userContent:  userContent,
		invocationID: "inv-1",
	}
}

func (c *testInvocationContext) Agent() agent.Agent              { return c.ag }
func (c *testInvocationContext) Artifacts() agent.Artifacts      { return nil }
func (c *testInvocationContext) Memory() agent.Memory            { return nil }
func (c *testInvocationContext) Session() session.Session        { return c.sess }
func (c *testInvocationContext) InvocationID() string              { return c.invocationID }
func (c *testInvocationContext) Branch() string                  { return "" }
func (c *testInvocationContext) IsolationScope() string          { return "" }
func (c *testInvocationContext) UserContent() *genai.Content     { return c.userContent }
func (c *testInvocationContext) RunConfig() *agent.RunConfig       { return nil }
func (c *testInvocationContext) EndInvocation()                  { c.ended = true }
func (c *testInvocationContext) Ended() bool                     { return c.ended }
func (c *testInvocationContext) ResumedInput(string) (any, bool) { return nil, false }

func (c *testInvocationContext) WithContext(ctx context.Context) agent.InvocationContext {
	copy := *c
	copy.Context = ctx
	return &copy
}

func (c *testInvocationContext) WithICDelta(d *agent.InvocationContextDelta) agent.InvocationContext {
	copy := *c
	if d.Agent != nil {
		copy.ag = *d.Agent
	}
	if d.Context != nil {
		copy.Context = *d.Context
	}
	return &copy
}

type fakePiService struct {
	mu sync.Mutex

	createSessionCalls int
	getSessionCalls    int
	submitRequests     []*piv1.SubmitMessageRequest

	createSessionFn func(context.Context, *connect.Request[piv1.CreateSessionRequest]) (*connect.Response[piv1.CreateSessionResponse], error)
	getSessionFn    func(context.Context, *connect.Request[piv1.GetSessionRequest]) (*connect.Response[piv1.GetSessionResponse], error)
	submitMessageFn func(context.Context, *connect.Request[piv1.SubmitMessageRequest]) (*connect.Response[piv1.SubmitMessageResponse], error)
	getTurnFn       func(context.Context, *connect.Request[piv1.GetTurnRequest]) (*connect.Response[piv1.GetTurnResponse], error)
	abortSessionFn  func(context.Context, *connect.Request[piv1.AbortSessionRequest]) (*connect.Response[piv1.AbortSessionResponse], error)

	getTurnResponses []*piv1.GetTurnResponse
	getTurnIndex     int

	nextSessionID int
}

func newFakePiService() *fakePiService {
	return &fakePiService{nextSessionID: 1}
}

func (f *fakePiService) defaultCreateSession(_ context.Context, _ *connect.Request[piv1.CreateSessionRequest]) (*connect.Response[piv1.CreateSessionResponse], error) {
	f.mu.Lock()
	id := fmt.Sprintf("pi-sess-%d", f.nextSessionID)
	f.nextSessionID++
	f.mu.Unlock()
	return connect.NewResponse(&piv1.CreateSessionResponse{
		Session: &piv1.Session{Id: id},
	}), nil
}

func (f *fakePiService) defaultGetSession(_ context.Context, req *connect.Request[piv1.GetSessionRequest]) (*connect.Response[piv1.GetSessionResponse], error) {
	return connect.NewResponse(&piv1.GetSessionResponse{
		Session: &piv1.Session{Id: req.Msg.GetSessionId()},
	}), nil
}

func (f *fakePiService) defaultSubmitMessage(_ context.Context, _ *connect.Request[piv1.SubmitMessageRequest]) (*connect.Response[piv1.SubmitMessageResponse], error) {
	return connect.NewResponse(&piv1.SubmitMessageResponse{TurnCursor: "cursor-1"}), nil
}

func (f *fakePiService) CreateSession(ctx context.Context, req *connect.Request[piv1.CreateSessionRequest]) (*connect.Response[piv1.CreateSessionResponse], error) {
	f.mu.Lock()
	f.createSessionCalls++
	fn := f.createSessionFn
	f.mu.Unlock()
	if fn != nil {
		return fn(ctx, req)
	}
	return f.defaultCreateSession(ctx, req)
}

func (f *fakePiService) GetSession(ctx context.Context, req *connect.Request[piv1.GetSessionRequest]) (*connect.Response[piv1.GetSessionResponse], error) {
	f.mu.Lock()
	f.getSessionCalls++
	fn := f.getSessionFn
	f.mu.Unlock()
	if fn != nil {
		return fn(ctx, req)
	}
	return f.defaultGetSession(ctx, req)
}

func (f *fakePiService) SubmitMessage(ctx context.Context, req *connect.Request[piv1.SubmitMessageRequest]) (*connect.Response[piv1.SubmitMessageResponse], error) {
	f.mu.Lock()
	f.submitRequests = append(f.submitRequests, req.Msg)
	fn := f.submitMessageFn
	f.mu.Unlock()
	if fn != nil {
		return fn(ctx, req)
	}
	return f.defaultSubmitMessage(ctx, req)
}

func (f *fakePiService) GetTurn(ctx context.Context, req *connect.Request[piv1.GetTurnRequest]) (*connect.Response[piv1.GetTurnResponse], error) {
	if f.getTurnFn != nil {
		return f.getTurnFn(ctx, req)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.getTurnIndex >= len(f.getTurnResponses) {
		return connect.NewResponse(&piv1.GetTurnResponse{
			Running: false,
			Result:  &piv1.TurnResult{Text: "unexpected extra poll"},
		}), nil
	}
	resp := f.getTurnResponses[f.getTurnIndex]
	f.getTurnIndex++
	return connect.NewResponse(resp), nil
}

func (f *fakePiService) AbortSession(context.Context, *connect.Request[piv1.AbortSessionRequest]) (*connect.Response[piv1.AbortSessionResponse], error) {
	if f.abortSessionFn != nil {
		return f.abortSessionFn(context.Background(), connect.NewRequest(&piv1.AbortSessionRequest{}))
	}
	return connect.NewResponse(&piv1.AbortSessionResponse{}), nil
}

func testBridge(fake *fakePiService, agentID, boxID string) *Bridge {
	return NewBridge(Config{
		Client:      fake,
		AgentID:     agentID,
		ButterBoxID: boxID,
		WorkingDir:  "/workspace",
		Provider:    "openai",
		Model:       "gpt-4",
	})
}

func collectRun(t *testing.T, b *Bridge, ictx agent.InvocationContext) (string, error) {
	t.Helper()
	var (
		gotText string
		gotErr  error
		count   int
	)
	for event, err := range b.run(ictx) {
		count++
		if err != nil {
			gotErr = err
			break
		}
		if event != nil && event.Content != nil {
			gotText, _ = extractContent(event.Content)
		}
	}
	if count == 0 {
		t.Fatal("run yielded no events")
	}
	return gotText, gotErr
}

// --- extractContent ---

func TestExtractContent(t *testing.T) {
	t.Run("nil content", func(t *testing.T) {
		text, images := extractContent(nil)
		if text != "" || images != nil {
			t.Fatalf("got text=%q images=%v, want empty", text, images)
		}
	})

	t.Run("text only", func(t *testing.T) {
		c := genai.NewContentFromText("hello\nworld", genai.RoleUser)
		text, images := extractContent(c)
		if text != "hello\nworld" {
			t.Fatalf("text = %q, want %q", text, "hello\nworld")
		}
		if len(images) != 0 {
			t.Fatalf("images = %v, want none", images)
		}
	})

	t.Run("image only", func(t *testing.T) {
		c := &genai.Content{
			Role: genai.RoleUser,
			Parts: []*genai.Part{{
				InlineData: &genai.Blob{MIMEType: "image/png", Data: []byte("png-bytes")},
			}},
		}
		text, images := extractContent(c)
		if text != "" {
			t.Fatalf("text = %q, want empty", text)
		}
		if len(images) != 1 || images[0].GetMimeType() != "image/png" {
			t.Fatalf("images = %v, want one image/png", images)
		}
	})

	t.Run("non-image inline data ignored", func(t *testing.T) {
		c := &genai.Content{
			Role: genai.RoleUser,
			Parts: []*genai.Part{{
				InlineData: &genai.Blob{MIMEType: "application/pdf", Data: []byte("pdf")},
			}},
		}
		_, images := extractContent(c)
		if len(images) != 0 {
			t.Fatalf("images = %v, want none", images)
		}
	})
}

// --- submitAndPoll ---

func TestSubmitAndPollSuccess(t *testing.T) {
	fake := newFakePiService()
	fake.getTurnResponses = []*piv1.GetTurnResponse{
		{Running: true},
		{Running: false, Result: &piv1.TurnResult{Text: "hello from pi"}},
	}

	b := testBridge(fake, "agent-1", "box-1")
	text, err := b.submitAndPoll(context.Background(), "pi-sess-1", "do work", nil)
	if err != nil {
		t.Fatalf("submitAndPoll: %v", err)
	}
	if text != "hello from pi" {
		t.Fatalf("text = %q, want %q", text, "hello from pi")
	}
	if fake.getTurnIndex != 2 {
		t.Fatalf("GetTurn calls = %d, want 2", fake.getTurnIndex)
	}

	// Full run cycle yields a model event with the same text.
	fake2 := newFakePiService()
	fake2.getTurnResponses = []*piv1.GetTurnResponse{
		{Running: true},
		{Running: false, Result: &piv1.TurnResult{Text: "hello from pi"}},
	}
	b2 := testBridge(fake2, "agent-1", "box-1")
	ictx := newTestInvocationContext(t, genai.NewContentFromText("do work", genai.RoleUser))
	gotText, err := collectRun(t, b2, ictx)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if gotText != "hello from pi" {
		t.Fatalf("run text = %q, want %q", gotText, "hello from pi")
	}
}

func TestDidNotFinish(t *testing.T) {
	fake := newFakePiService()
	fake.getTurnResponses = []*piv1.GetTurnResponse{
		{Running: false},
	}

	b := testBridge(fake, "agent-1", "box-1")
	_, err := b.submitAndPoll(context.Background(), "pi-sess-1", "do work", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "did not finish") {
		t.Fatalf("error = %q, want substring %q", err.Error(), "did not finish")
	}
}

func TestImagePassthrough(t *testing.T) {
	fake := newFakePiService()
	fake.getTurnResponses = []*piv1.GetTurnResponse{
		{Running: false, Result: &piv1.TurnResult{Text: "seen image"}},
	}

	imageData := []byte{0x89, 0x50, 0x4e, 0x47}
	images := []*piv1.ImageContent{{
		MimeType: "image/png",
		Data:     imageData,
	}}

	b := testBridge(fake, "agent-1", "box-1")
	_, err := b.submitAndPoll(context.Background(), "pi-sess-1", "describe this", images)
	if err != nil {
		t.Fatalf("submitAndPoll: %v", err)
	}
	if len(fake.submitRequests) != 1 {
		t.Fatalf("SubmitMessage calls = %d, want 1", len(fake.submitRequests))
	}
	req := fake.submitRequests[0]
	if req.GetMessage() != "describe this" {
		t.Fatalf("message = %q, want %q", req.GetMessage(), "describe this")
	}
	if len(req.GetImages()) != 1 {
		t.Fatalf("images = %v, want 1", req.GetImages())
	}
	if req.GetImages()[0].GetMimeType() != "image/png" {
		t.Fatalf("mime = %q, want image/png", req.GetImages()[0].GetMimeType())
	}
	if string(req.GetImages()[0].GetData()) != string(imageData) {
		t.Fatalf("image data mismatch")
	}

	// End-to-end: inline image in user content reaches SubmitMessage.
	fake2 := newFakePiService()
	fake2.getTurnResponses = []*piv1.GetTurnResponse{
		{Running: false, Result: &piv1.TurnResult{Text: "ok"}},
	}
	b2 := testBridge(fake2, "agent-1", "box-1")
	content := &genai.Content{
		Role: genai.RoleUser,
		Parts: []*genai.Part{
			{Text: "look at this"},
			{InlineData: &genai.Blob{MIMEType: "image/jpeg", Data: []byte("jpeg-data")}},
		},
	}
	ictx := newTestInvocationContext(t, content)
	if _, err := collectRun(t, b2, ictx); err != nil {
		t.Fatalf("run: %v", err)
	}
	if len(fake2.submitRequests) != 1 {
		t.Fatalf("SubmitMessage calls = %d, want 1", len(fake2.submitRequests))
	}
	submitted := fake2.submitRequests[0]
	if len(submitted.GetImages()) != 1 || submitted.GetImages()[0].GetMimeType() != "image/jpeg" {
		t.Fatalf("submitted images = %v, want one image/jpeg", submitted.GetImages())
	}
}

// --- ensureSession ---

func TestSessionReuse(t *testing.T) {
	fake := newFakePiService()
	b := testBridge(fake, "agent-1", "box-a")
	ictx := newTestInvocationContext(t, genai.NewContentFromText("hi", genai.RoleUser))

	id1, err := b.ensureSession(ictx)
	if err != nil {
		t.Fatalf("first ensureSession: %v", err)
	}
	if id1 != "pi-sess-1" {
		t.Fatalf("first session id = %q, want pi-sess-1", id1)
	}
	if fake.createSessionCalls != 1 {
		t.Fatalf("CreateSession calls after first = %d, want 1", fake.createSessionCalls)
	}

	id2, err := b.ensureSession(ictx)
	if err != nil {
		t.Fatalf("second ensureSession: %v", err)
	}
	if id2 != id1 {
		t.Fatalf("second session id = %q, want reuse of %q", id2, id1)
	}
	if fake.createSessionCalls != 1 {
		t.Fatalf("CreateSession calls after second = %d, want 1 (reused)", fake.createSessionCalls)
	}
	if fake.getSessionCalls != 1 {
		t.Fatalf("GetSession calls = %d, want 1", fake.getSessionCalls)
	}

	// Two consecutive runs through the bridge also reuse the pi session.
	fake2 := newFakePiService()
	fake2.getTurnResponses = []*piv1.GetTurnResponse{
		{Running: false, Result: &piv1.TurnResult{Text: "first"}},
	}
	b2 := testBridge(fake2, "agent-1", "box-a")
	ictx2 := newTestInvocationContext(t, genai.NewContentFromText("first", genai.RoleUser))
	if _, err := collectRun(t, b2, ictx2); err != nil {
		t.Fatalf("first run: %v", err)
	}
	firstCreates := fake2.createSessionCalls

	fake2.getTurnResponses = []*piv1.GetTurnResponse{
		{Running: false, Result: &piv1.TurnResult{Text: "second"}},
	}
	fake2.getTurnIndex = 0
	ictx2.userContent = genai.NewContentFromText("second", genai.RoleUser)
	if _, err := collectRun(t, b2, ictx2); err != nil {
		t.Fatalf("second run: %v", err)
	}
	if fake2.createSessionCalls != firstCreates {
		t.Fatalf("CreateSession after second run = %d, want %d (reused)", fake2.createSessionCalls, firstCreates)
	}
}

func TestSessionRepointRecreates(t *testing.T) {
	fake := newFakePiService()
	ictx := newTestInvocationContext(t, genai.NewContentFromText("hi", genai.RoleUser))

	bridgeA := testBridge(fake, "agent-1", "box-a")
	idA, err := bridgeA.ensureSession(ictx)
	if err != nil {
		t.Fatalf("ensureSession box-a: %v", err)
	}
	if fake.createSessionCalls != 1 {
		t.Fatalf("CreateSession after box-a = %d, want 1", fake.createSessionCalls)
	}

	bridgeB := testBridge(fake, "agent-1", "box-b")
	idB, err := bridgeB.ensureSession(ictx)
	if err != nil {
		t.Fatalf("ensureSession box-b: %v", err)
	}
	if fake.createSessionCalls != 2 {
		t.Fatalf("CreateSession after repoint = %d, want 2", fake.createSessionCalls)
	}
	if idB == idA {
		t.Fatalf("expected new session after repoint, got same id %q", idB)
	}
	if fake.getSessionCalls != 0 {
		t.Fatalf("GetSession calls = %d, want 0 (repoint skips reuse check)", fake.getSessionCalls)
	}

	// Working dir change also triggers recreate.
	fake2 := newFakePiService()
	ictx2 := newTestInvocationContext(t, genai.NewContentFromText("hi", genai.RoleUser))
	b1 := NewBridge(Config{
		Client:      fake2,
		AgentID:     "agent-1",
		ButterBoxID: "box-a",
		WorkingDir:  "/dir-a",
	})
	if _, err := b1.ensureSession(ictx2); err != nil {
		t.Fatalf("ensureSession dir-a: %v", err)
	}
	b2 := NewBridge(Config{
		Client:      fake2,
		AgentID:     "agent-1",
		ButterBoxID: "box-a",
		WorkingDir:  "/dir-b",
	})
	if _, err := b2.ensureSession(ictx2); err != nil {
		t.Fatalf("ensureSession dir-b: %v", err)
	}
	if fake2.createSessionCalls != 2 {
		t.Fatalf("CreateSession after working_dir repoint = %d, want 2", fake2.createSessionCalls)
	}
}

func TestCapacityError(t *testing.T) {
	fake := newFakePiService()
	fake.createSessionFn = func(_ context.Context, _ *connect.Request[piv1.CreateSessionRequest]) (*connect.Response[piv1.CreateSessionResponse], error) {
		return nil, connect.NewError(connect.CodeResourceExhausted, errors.New("max sessions reached"))
	}

	b := testBridge(fake, "agent-1", "box-a")
	ictx := newTestInvocationContext(t, genai.NewContentFromText("hi", genai.RoleUser))
	_, err := b.ensureSession(ictx)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "at capacity") {
		t.Fatalf("error = %q, want substring %q", err.Error(), "at capacity")
	}

	// Propagates through the full run path as well.
	fake2 := newFakePiService()
	fake2.createSessionFn = fake.createSessionFn
	b2 := testBridge(fake2, "agent-1", "box-a")
	_, err = collectRun(t, b2, ictx)
	if err == nil {
		t.Fatal("expected run error, got nil")
	}
	if !strings.Contains(err.Error(), "at capacity") {
		t.Fatalf("run error = %q, want substring %q", err.Error(), "at capacity")
	}
}

func TestEnsureSessionNotFoundRecreates(t *testing.T) {
	fake := newFakePiService()
	ictx := newTestInvocationContext(t, genai.NewContentFromText("hi", genai.RoleUser))
	b := testBridge(fake, "agent-1", "box-a")

	id1, err := b.ensureSession(ictx)
	if err != nil {
		t.Fatalf("first ensureSession: %v", err)
	}

	fake.getSessionFn = func(_ context.Context, _ *connect.Request[piv1.GetSessionRequest]) (*connect.Response[piv1.GetSessionResponse], error) {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("session gone"))
	}
	id2, err := b.ensureSession(ictx)
	if err != nil {
		t.Fatalf("second ensureSession: %v", err)
	}
	if id2 == id1 {
		t.Fatalf("expected new session after not-found, got same id %q", id2)
	}
	if fake.createSessionCalls != 2 {
		t.Fatalf("CreateSession calls = %d, want 2", fake.createSessionCalls)
	}
}

func TestBuildAgent(t *testing.T) {
	fake := newFakePiService()
	b := testBridge(fake, "agent-1", "box-a")
	ag, err := b.BuildAgent("my-pi", "A pi agent")
	if err != nil {
		t.Fatalf("BuildAgent: %v", err)
	}
	if ag.Name() != "my-pi" {
		t.Fatalf("name = %q, want my-pi", ag.Name())
	}
	if ag.Description() != "A pi agent" {
		t.Fatalf("description = %q, want %q", ag.Description(), "A pi agent")
	}
}
