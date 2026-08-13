package application

import (
	"context"
	"iter"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type generateReq = agentsv1.GenerateSessionTitleRequest

func unauthenticatedCtx() context.Context { return context.Background() }

func makeEvent(author string, parts ...*genai.Part) *session.Event {
	return &session.Event{
		LLMResponse: model.LLMResponse{
			Content: &genai.Content{Parts: parts},
		},
		Author: author,
	}
}

func textPart(s string) *genai.Part { return &genai.Part{Text: s} }
func funcCallPart(name string) *genai.Part {
	return &genai.Part{FunctionCall: &genai.FunctionCall{Name: name}}
}
func funcRespPart(name string) *genai.Part {
	return &genai.Part{FunctionResponse: &genai.FunctionResponse{Name: name}}
}
func imagePart() *genai.Part { return &genai.Part{InlineData: &genai.Blob{MIMEType: "image/png"}} }

// fakeSessionWithEvents extends fakeSession with real events.
type fakeSessionWithEvents struct {
	fakeSession
	events []*session.Event
}

func (s *fakeSessionWithEvents) Events() session.Events {
	return &fakeEventsImpl{events: s.events}
}

type fakeEventsImpl struct {
	events []*session.Event
}

func (e *fakeEventsImpl) All() iter.Seq[*session.Event] {
	return func(yield func(*session.Event) bool) {
		for _, evt := range e.events {
			if !yield(evt) {
				return
			}
		}
	}
}

func (e *fakeEventsImpl) Len() int                { return len(e.events) }
func (e *fakeEventsImpl) At(i int) *session.Event { return e.events[i] }

func (s *fakeSessionWithEvents) LastUpdateTime() time.Time { return time.Now() }

// --- deriveAutoTitle unit tests ---

func TestDeriveAutoTitle_PrefersFirstUserText(t *testing.T) {
	events := []*session.Event{
		makeEvent("user", textPart("Hello, how are you?")),
		makeEvent("agent", textPart("I'm doing well!")),
	}
	got := deriveAutoTitle(events)
	if got != "Hello, how are you?" {
		t.Fatalf("expected user text, got %q", got)
	}
}

func TestDeriveAutoTitle_FallsBackToAssistantText(t *testing.T) {
	events := []*session.Event{
		makeEvent("user", imagePart()),
		makeEvent("agent", textPart("Based on the image, I can see...")),
	}
	got := deriveAutoTitle(events)
	want := "Based on the image, I can see."
	if got != want {
		t.Fatalf("expected assistant text truncated to 30 cp %q, got %q", want, got)
	}
}

func TestDeriveAutoTitle_ImageChatFallback(t *testing.T) {
	events := []*session.Event{
		makeEvent("user", imagePart()),
		makeEvent("agent", funcCallPart("analyze")),
	}
	got := deriveAutoTitle(events)
	if got != "Image chat" {
		t.Fatalf("expected 'Image chat', got %q", got)
	}
}

func TestDeriveAutoTitle_EmptyWhenNoContent(t *testing.T) {
	events := []*session.Event{}
	got := deriveAutoTitle(events)
	if got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}

func TestDeriveAutoTitle_SkipsToolEvents(t *testing.T) {
	events := []*session.Event{
		makeEvent("agent", funcCallPart("search")),
		makeEvent("agent", funcRespPart("search")),
		makeEvent("user", textPart("My question")),
		makeEvent("agent", textPart("Here's the answer")),
	}
	got := deriveAutoTitle(events)
	if got != "My question" {
		t.Fatalf("expected user text (skipping tool events), got %q", got)
	}
}

func TestDeriveAutoTitle_Truncates30CodePoints(t *testing.T) {
	longText := strings.Repeat("あ", 40)
	events := []*session.Event{
		makeEvent("user", textPart(longText)),
	}
	got := deriveAutoTitle(events)
	if count := len([]rune(got)); count != 30 {
		t.Fatalf("expected 30 code points, got %d (%q)", count, got)
	}
}

func TestDeriveAutoTitle_NormalizesNewlines(t *testing.T) {
	events := []*session.Event{
		makeEvent("user", textPart("  hello\nworld\r\nfoo  ")),
	}
	got := deriveAutoTitle(events)
	if got != "hello world foo" {
		t.Fatalf("expected normalized text, got %q", got)
	}
}

// --- normalizeAutoTitle tests ---

func TestNormalizeAutoTitle_Empty(t *testing.T) {
	if got := normalizeAutoTitle(""); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}

func TestNormalizeAutoTitle_TruncatesTo30(t *testing.T) {
	input := strings.Repeat("x", 50)
	got := normalizeAutoTitle(input)
	if len([]rune(got)) != 30 {
		t.Fatalf("expected 30 code points, got %d", len([]rune(got)))
	}
}

func TestNormalizeAutoTitle_Exactly30Unchanged(t *testing.T) {
	input := strings.Repeat("x", 30)
	got := normalizeAutoTitle(input)
	if got != input {
		t.Fatalf("expected unchanged, got %q", got)
	}
}

// --- truncateCodePoints tests ---

func TestTruncateCodePoints_ASCII(t *testing.T) {
	if got := truncateCodePoints("hello world", 5); got != "hello" {
		t.Fatalf("got %q", got)
	}
}

func TestTruncateCodePoints_Unicode(t *testing.T) {
	input := "日本語テスト"
	if got := truncateCodePoints(input, 3); got != "日本語" {
		t.Fatalf("got %q", got)
	}
}

func TestTruncateCodePoints_NoTruncation(t *testing.T) {
	if got := truncateCodePoints("short", 100); got != "short" {
		t.Fatalf("got %q", got)
	}
}

// --- isToolOnlyEvent tests ---

func TestIsToolOnlyEvent_FunctionCallOnly(t *testing.T) {
	evt := makeEvent("agent", funcCallPart("search"))
	if !isToolOnlyEvent(evt) {
		t.Fatal("expected tool-only")
	}
}

func TestIsToolOnlyEvent_FunctionResponseOnly(t *testing.T) {
	evt := makeEvent("agent", funcRespPart("search"))
	if !isToolOnlyEvent(evt) {
		t.Fatal("expected tool-only")
	}
}

func TestIsToolOnlyEvent_MixedWithText(t *testing.T) {
	evt := makeEvent("agent", funcCallPart("search"), textPart("Here's what I found"))
	if isToolOnlyEvent(evt) {
		t.Fatal("expected NOT tool-only when text is present")
	}
}

func TestIsToolOnlyEvent_TextOnly(t *testing.T) {
	evt := makeEvent("agent", textPart("Hello"))
	if isToolOnlyEvent(evt) {
		t.Fatal("text-only event is not tool-only")
	}
}

func TestIsToolOnlyEvent_NilContent(t *testing.T) {
	evt := &session.Event{Author: "agent"}
	if isToolOnlyEvent(evt) {
		t.Fatal("nil content is not tool-only")
	}
}

// --- GenerateSessionTitle handler tests ---

func newGenerateTestService(store *stubTitleStore, sessions ...session.Session) *SessionServiceServer {
	svc := NewSessionServiceServer()
	svc.titleStore = store
	svc.sessionSvc = &titleSeamSessionService{sessions: sessions}
	return svc
}

func TestGenerateSessionTitle_RequiredFields(t *testing.T) {
	store := &stubTitleStore{}
	svc := newGenerateTestService(store)

	tests := []struct {
		name string
		req  *generateReq
		want string
	}{
		{"missing app_name", &generateReq{UserId: "u1", SessionId: "s1"}, "app_name"},
		{"missing user_id", &generateReq{AppName: "web", SessionId: "s1"}, "user_id"},
		{"missing session_id", &generateReq{AppName: "web", UserId: "u1"}, "session_id"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.GenerateSessionTitle(titleAdminCtx(), connect.NewRequest(tt.req))
			assertInvalidArgument(t, err)
		})
	}
}

func TestGenerateSessionTitle_OwnerCanGenerate(t *testing.T) {
	sess := &fakeSessionWithEvents{
		fakeSession: fakeSession{id: "s1", title: "", state: &fakeState{data: map[string]any{}}},
		events:      []*session.Event{makeEvent("user", textPart("Hello"))},
	}
	store := &stubTitleStore{}
	svc := newGenerateTestService(store, sess)

	ctx := titleUserCtx("u1", "user")
	resp, err := svc.GenerateSessionTitle(ctx, connect.NewRequest(&generateReq{
		AppName: "test", UserId: "u1", SessionId: "s1",
	}))
	if err != nil {
		t.Fatalf("owner should be allowed: %v", err)
	}
	if !resp.Msg.GetGenerated() {
		t.Fatal("expected generated=true")
	}
	if resp.Msg.GetSession().GetTitle() != "Hello" {
		t.Fatalf("expected title 'Hello', got %q", resp.Msg.GetSession().GetTitle())
	}
}

func TestGenerateSessionTitle_NonOwnerRejected(t *testing.T) {
	store := &stubTitleStore{}
	svc := newGenerateTestService(store)

	ctx := titleUserCtx("u1", "user")
	_, err := svc.GenerateSessionTitle(ctx, connect.NewRequest(&generateReq{
		AppName: "web", UserId: "other-user", SessionId: "s1",
	}))
	if connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("expected PermissionDenied, got %v", err)
	}
}

func TestGenerateSessionTitle_UnauthenticatedRejected(t *testing.T) {
	store := &stubTitleStore{}
	svc := newGenerateTestService(store)

	_, err := svc.GenerateSessionTitle(unauthenticatedCtx(), connect.NewRequest(&generateReq{
		AppName: "web", UserId: "u1", SessionId: "s1",
	}))
	if connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("expected Unauthenticated, got %v", err)
	}
}

func TestGenerateSessionTitle_AdminCanGenerate(t *testing.T) {
	sess := &fakeSessionWithEvents{
		fakeSession: fakeSession{id: "s1", title: "", state: &fakeState{data: map[string]any{}}},
		events:      []*session.Event{makeEvent("user", textPart("Admin test"))},
	}
	store := &stubTitleStore{}
	svc := newGenerateTestService(store, sess)

	resp, err := svc.GenerateSessionTitle(titleAdminCtx(), connect.NewRequest(&generateReq{
		AppName: "test", UserId: "u1", SessionId: "s1",
	}))
	if err != nil {
		t.Fatalf("admin should be allowed: %v", err)
	}
	if !resp.Msg.GetGenerated() {
		t.Fatal("expected generated=true")
	}
}

func TestGenerateSessionTitle_ExistingFirstClassTitleNoOp(t *testing.T) {
	sess := &fakeSessionWithEvents{
		fakeSession: fakeSession{id: "s1", title: "My Chat", state: &fakeState{data: map[string]any{}}},
		events:      []*session.Event{makeEvent("user", textPart("Hello"))},
	}
	store := &stubTitleStore{}
	svc := newGenerateTestService(store, sess)

	resp, err := svc.GenerateSessionTitle(titleAdminCtx(), connect.NewRequest(&generateReq{
		AppName: "test", UserId: "u1", SessionId: "s1",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Msg.GetGenerated() {
		t.Fatal("expected generated=false for existing title")
	}
	if resp.Msg.GetSession().GetTitle() != "My Chat" {
		t.Fatalf("expected existing title, got %q", resp.Msg.GetSession().GetTitle())
	}
	if store.casCalled > 0 {
		t.Fatal("CAS should not be called when title already exists")
	}
}

func TestGenerateSessionTitle_LegacyTitleNoOp(t *testing.T) {
	sess := &fakeSessionWithEvents{
		fakeSession: fakeSession{id: "s1", title: "", state: &fakeState{data: map[string]any{"title": "Legacy Title"}}},
		events:      []*session.Event{makeEvent("user", textPart("Hello"))},
	}
	store := &stubTitleStore{}
	svc := newGenerateTestService(store, sess)

	resp, err := svc.GenerateSessionTitle(titleAdminCtx(), connect.NewRequest(&generateReq{
		AppName: "test", UserId: "u1", SessionId: "s1",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Msg.GetGenerated() {
		t.Fatal("expected generated=false for legacy title")
	}
	if resp.Msg.GetSession().GetTitle() != "Legacy Title" {
		t.Fatalf("expected legacy title, got %q", resp.Msg.GetSession().GetTitle())
	}
}

func TestGenerateSessionTitle_DuplicateCallReturnsExisting(t *testing.T) {
	sess := &fakeSessionWithEvents{
		fakeSession: fakeSession{id: "s1", title: "", state: &fakeState{data: map[string]any{}}},
		events:      []*session.Event{makeEvent("user", textPart("Hello"))},
	}
	store := &stubTitleStore{existingTitle: "Hello"}
	svc := newGenerateTestService(store, sess)

	resp, err := svc.GenerateSessionTitle(titleAdminCtx(), connect.NewRequest(&generateReq{
		AppName: "test", UserId: "u1", SessionId: "s1",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Msg.GetGenerated() {
		t.Fatal("expected generated=false for duplicate call")
	}
	if resp.Msg.GetSession().GetTitle() != "Hello" {
		t.Fatalf("expected existing title, got %q", resp.Msg.GetSession().GetTitle())
	}
}

func TestGenerateSessionTitle_ConcurrentManualUpdateWins(t *testing.T) {
	sess := &fakeSessionWithEvents{
		fakeSession: fakeSession{id: "s1", title: "", state: &fakeState{data: map[string]any{}}},
		events:      []*session.Event{makeEvent("user", textPart("Hello"))},
	}
	store := &stubTitleStore{existingTitle: "Manually Set"}
	svc := newGenerateTestService(store, sess)

	resp, err := svc.GenerateSessionTitle(titleAdminCtx(), connect.NewRequest(&generateReq{
		AppName: "test", UserId: "u1", SessionId: "s1",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Msg.GetGenerated() {
		t.Fatal("expected generated=false when manual update won the race")
	}
	if resp.Msg.GetSession().GetTitle() != "Manually Set" {
		t.Fatalf("expected manual title, got %q", resp.Msg.GetSession().GetTitle())
	}
}

func TestGenerateSessionTitle_SessionNotFound(t *testing.T) {
	store := &stubTitleStore{}
	svc := newGenerateTestService(store)

	_, err := svc.GenerateSessionTitle(titleAdminCtx(), connect.NewRequest(&generateReq{
		AppName: "web", UserId: "u1", SessionId: "does-not-exist",
	}))
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("expected NotFound, got %v", err)
	}
}

func TestGenerateSessionTitle_ImageOnlyFallback(t *testing.T) {
	sess := &fakeSessionWithEvents{
		fakeSession: fakeSession{id: "s1", title: "", state: &fakeState{data: map[string]any{}}},
		events: []*session.Event{
			makeEvent("user", imagePart()),
			makeEvent("agent", funcCallPart("analyze"), funcRespPart("analyze")),
		},
	}
	store := &stubTitleStore{}
	svc := newGenerateTestService(store, sess)

	resp, err := svc.GenerateSessionTitle(titleAdminCtx(), connect.NewRequest(&generateReq{
		AppName: "test", UserId: "u1", SessionId: "s1",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Msg.GetGenerated() {
		t.Fatal("expected generated=true")
	}
	if resp.Msg.GetSession().GetTitle() != "Image chat" {
		t.Fatalf("expected 'Image chat', got %q", resp.Msg.GetSession().GetTitle())
	}
}

func TestGenerateSessionTitle_NormalizesAndTruncates(t *testing.T) {
	longText := "  " + strings.Repeat("x", 50) + "\n"
	sess := &fakeSessionWithEvents{
		fakeSession: fakeSession{id: "s1", title: "", state: &fakeState{data: map[string]any{}}},
		events:      []*session.Event{makeEvent("user", textPart(longText))},
	}
	store := &stubTitleStore{}
	svc := newGenerateTestService(store, sess)

	resp, err := svc.GenerateSessionTitle(titleAdminCtx(), connect.NewRequest(&generateReq{
		AppName: "test", UserId: "u1", SessionId: "s1",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	title := resp.Msg.GetSession().GetTitle()
	if count := len([]rune(title)); count > 30 {
		t.Fatalf("title exceeds 30 code points: %d (%q)", count, title)
	}
	if strings.Contains(title, "\n") || strings.HasPrefix(title, " ") {
		t.Fatalf("title not normalized: %q", title)
	}
}

func TestGenerateSessionTitle_RetryAfterFailure(t *testing.T) {
	sess := &fakeSessionWithEvents{
		fakeSession: fakeSession{id: "s1", title: "", state: &fakeState{data: map[string]any{}}},
		events:      []*session.Event{makeEvent("user", textPart("Retry me"))},
	}
	store := &stubTitleStore{}
	svc := newGenerateTestService(store, sess)

	resp, err := svc.GenerateSessionTitle(titleAdminCtx(), connect.NewRequest(&generateReq{
		AppName: "test", UserId: "u1", SessionId: "s1",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Msg.GetGenerated() {
		t.Fatal("first call should generate")
	}

	// Simulate the title now existing for the second call.
	store.existingTitle = "Retry me"
	resp, err = svc.GenerateSessionTitle(titleAdminCtx(), connect.NewRequest(&generateReq{
		AppName: "test", UserId: "u1", SessionId: "s1",
	}))
	if err != nil {
		t.Fatalf("retry unexpected error: %v", err)
	}
	if resp.Msg.GetGenerated() {
		t.Fatal("retry should not re-generate")
	}
}

func TestGenerateSessionTitle_NoEventsNoGeneration(t *testing.T) {
	sess := &fakeSessionWithEvents{
		fakeSession: fakeSession{id: "s1", title: "", state: &fakeState{data: map[string]any{}}},
		events:      nil,
	}
	store := &stubTitleStore{}
	svc := newGenerateTestService(store, sess)

	resp, err := svc.GenerateSessionTitle(titleAdminCtx(), connect.NewRequest(&generateReq{
		AppName: "test", UserId: "u1", SessionId: "s1",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Msg.GetGenerated() {
		t.Fatal("expected generated=false when no events")
	}
	if store.casCalled > 0 {
		t.Fatal("CAS should not be called when no title could be derived")
	}
}
