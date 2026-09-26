package mem0memory

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"google.golang.org/adk/v2/memory"
	"google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/mem0"
	cryptokeymemory "go.orx.me/apps/butter/internal/repo/cryptokey/memory"
	memoryconfigmem "go.orx.me/apps/butter/internal/repo/memoryconfig/memory"
	"go.orx.me/apps/butter/internal/runtime/memoryconn"
	"go.orx.me/apps/butter/internal/secretbox"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// fakeMem0 records every request body and answers searches from a table
// keyed by the identity filter value.
type fakeMem0 struct {
	srv *httptest.Server

	mu       sync.Mutex
	searches []map[string]any
	adds     []map[string]any
	results  map[string][]mem0.Memory // filter identity value -> results
	status   int                      // non-zero forces an error status
}

func newFakeMem0(t *testing.T) *fakeMem0 {
	f := &fakeMem0{results: map[string][]mem0.Memory{}}
	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		f.mu.Lock()
		defer f.mu.Unlock()
		if f.status != 0 {
			http.Error(w, `{"detail":"boom"}`, f.status)
			return
		}
		switch r.URL.Path {
		case "/search":
			f.searches = append(f.searches, body)
			filters, _ := body["filters"].(map[string]any)
			key, _ := filters["user_id"].(string)
			if key == "" {
				key, _ = filters["agent_id"].(string)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"results": f.results[key]})
		case "/memories":
			f.adds = append(f.adds, body)
			_, _ = w.Write([]byte(`{"results":[]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(f.srv.Close)
	return f
}

func (f *fakeMem0) recorded() (searches, adds []map[string]any) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]map[string]any(nil), f.searches...), append([]map[string]any(nil), f.adds...)
}

// newService wires a Service to the fake with an enabled config for ws1.
func newService(t *testing.T, f *fakeMem0, enabled bool) *Service {
	t.Helper()
	repo := memoryconfigmem.New()
	cfg := &agentsv1.WorkspaceMemoryConfig{BaseUrl: f.srv.URL, Enabled: enabled}
	if _, err := repo.Put(t.Context(), "ws1", cfg, nil); err != nil {
		t.Fatalf("Put: %v", err)
	}
	svc := New(memoryconn.NewResolver(repo, secretbox.NewKeyring(cryptokeymemory.New())))
	svc.HTTPClient = f.srv.Client()
	return svc
}

var testScope = Scope{
	WorkspaceID:  "ws1",
	AgentID:      "helper",
	SessionID:    "s1",
	InvocationID: "inv-2",
	Channel:      "telegram",
	Principal:    "tg:42",
}

func TestScopeIdentityEncoding(t *testing.T) {
	if got := WorkspaceUserID("ws1"); got != "ws:ws1" {
		t.Fatalf("WorkspaceUserID = %q", got)
	}
	if got := AgentMemoryID("ws1", "helper"); got != "ws:ws1:agent:helper" {
		t.Fatalf("AgentMemoryID = %q", got)
	}
	if f, err := testScope.filters(TargetWorkspace); err != nil || len(f) != 1 || f["user_id"] != "ws:ws1" {
		t.Fatalf("workspace filters = %v, %v", f, err)
	}
	if f, err := testScope.filters(TargetAgent); err != nil || len(f) != 1 || f["agent_id"] != "ws:ws1:agent:helper" {
		t.Fatalf("agent filters = %v, %v", f, err)
	}
	if _, err := (Scope{AgentID: "a"}).filters(TargetWorkspace); err == nil {
		t.Fatal("scope without workspace accepted")
	}
	if _, err := (Scope{WorkspaceID: "ws1"}).filters(TargetAgent); err == nil {
		t.Fatal("agent target without agent accepted")
	}
}

func TestAddWritesWorkspaceIdentityOnlyWithProvenance(t *testing.T) {
	f := newFakeMem0(t)
	svc := newService(t, f, true)

	_, err := svc.Add(t.Context(), testScope, TargetWorkspace, []mem0.Message{{Role: "user", Content: "we deploy on fridays"}})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	_, adds := f.recorded()
	if len(adds) != 1 {
		t.Fatalf("adds = %d", len(adds))
	}
	body := adds[0]
	if body["user_id"] != "ws:ws1" || body["infer"] != true {
		t.Fatalf("body = %v", body)
	}
	for _, absent := range []string{"agent_id", "run_id"} {
		if _, ok := body[absent]; ok {
			t.Fatalf("workspace add carries %s: %v", absent, body)
		}
	}
	meta, _ := body["metadata"].(map[string]any)
	want := map[string]string{
		"butter_workspace_id":  "ws1",
		"butter_agent_id":      "helper",
		"butter_session_id":    "s1",
		"butter_invocation_id": "inv-2",
		"butter_channel":       "telegram",
		"butter_principal":     "tg:42",
	}
	for k, v := range want {
		if meta[k] != v {
			t.Fatalf("metadata[%s] = %v, want %q (metadata = %v)", k, meta[k], v, meta)
		}
	}
}

func TestAddAgentTargetWritesAgentIdentityOnly(t *testing.T) {
	f := newFakeMem0(t)
	svc := newService(t, f, true)

	if _, err := svc.Add(t.Context(), testScope, TargetAgent, []mem0.Message{{Role: "user", Content: "x"}}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	_, adds := f.recorded()
	if adds[0]["agent_id"] != "ws:ws1:agent:helper" {
		t.Fatalf("body = %v", adds[0])
	}
	if _, ok := adds[0]["user_id"]; ok {
		t.Fatalf("agent add carries user_id: %v", adds[0])
	}
}

func TestAddWithoutMessagesSendsNothing(t *testing.T) {
	f := newFakeMem0(t)
	svc := newService(t, f, true)
	if _, err := svc.Add(t.Context(), testScope, TargetWorkspace, nil); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if _, adds := f.recorded(); len(adds) != 0 {
		t.Fatalf("empty add reached the server: %v", adds)
	}
}

func TestRecallMergesScopesDedupesAndKeepsTopK(t *testing.T) {
	f := newFakeMem0(t)
	f.results["ws:ws1"] = []mem0.Memory{
		{ID: "w1", Memory: "team uses pnpm", Score: 0.9},
		{ID: "shared", Memory: "dup", Score: 0.5},
		{ID: "w2", Memory: "deploys friday", Score: 0.4},
	}
	f.results["ws:ws1:agent:helper"] = []mem0.Memory{
		{ID: "a1", Memory: "answer tersely", Score: 0.8},
		{ID: "shared", Memory: "dup", Score: 0.5},
	}
	svc := newService(t, f, true)

	threshold := 0.2
	got, err := svc.Recall(t.Context(), testScope, "how do we build?", SearchOptions{TopK: 3, Threshold: &threshold})
	if err != nil {
		t.Fatalf("Recall: %v", err)
	}
	var ids []string
	for _, m := range got {
		ids = append(ids, m.ID+"/"+m.Target.String())
	}
	if strings.Join(ids, ",") != "w1/workspace,a1/agent,shared/workspace" {
		t.Fatalf("recalled = %v", ids)
	}

	searches, _ := f.recorded()
	if len(searches) != 2 {
		t.Fatalf("searches = %d, want one per scope", len(searches))
	}
	for _, s := range searches {
		if s["query"] != "how do we build?" || s["top_k"] != float64(3) || s["threshold"] != 0.2 {
			t.Fatalf("search body = %v", s)
		}
	}
}

func TestRecallWithoutAgentSearchesWorkspaceOnlyWithDefaults(t *testing.T) {
	f := newFakeMem0(t)
	svc := newService(t, f, true)

	scope := testScope
	scope.AgentID = ""
	if _, err := svc.Recall(t.Context(), scope, "q", SearchOptions{}); err != nil {
		t.Fatalf("Recall: %v", err)
	}
	searches, _ := f.recorded()
	if len(searches) != 1 {
		t.Fatalf("searches = %d, want workspace only", len(searches))
	}
	if searches[0]["top_k"] != float64(DefaultTopK) || searches[0]["threshold"] != DefaultThreshold {
		t.Fatalf("defaults not applied: %v", searches[0])
	}
}

func TestExtensionMethodsReportNotConfigured(t *testing.T) {
	f := newFakeMem0(t)
	for name, svc := range map[string]*Service{
		"disabled":  newService(t, f, false),
		"no config": New(memoryconn.NewResolver(memoryconfigmem.New(), nil)),
	} {
		if _, err := svc.Recall(t.Context(), testScope, "q", SearchOptions{}); !errors.Is(err, memoryconn.ErrNotConfigured) {
			t.Fatalf("%s Recall = %v, want ErrNotConfigured", name, err)
		}
		if _, err := svc.Add(t.Context(), testScope, TargetWorkspace, []mem0.Message{{Role: "user", Content: "x"}}); !errors.Is(err, memoryconn.ErrNotConfigured) {
			t.Fatalf("%s Add = %v, want ErrNotConfigured", name, err)
		}
	}
	if searches, adds := f.recorded(); len(searches)+len(adds) != 0 {
		t.Fatal("an unconfigured workspace reached mem0")
	}
}

func TestServerErrorsPropagate(t *testing.T) {
	f := newFakeMem0(t)
	f.status = http.StatusBadGateway
	svc := newService(t, f, true)

	_, err := svc.Recall(t.Context(), testScope, "q", SearchOptions{})
	var apiErr *mem0.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusBadGateway {
		t.Fatalf("Recall = %v, want the mem0 APIError", err)
	}
}

func TestADKInterfaceUsesContextScope(t *testing.T) {
	f := newFakeMem0(t)
	f.results["ws:ws1"] = []mem0.Memory{{ID: "w1", Memory: "team uses pnpm", Score: 0.9, CreatedAt: "2026-09-01T10:00:00.123456+00:00"}}
	svc := newService(t, f, true)

	resp, err := svc.SearchMemory(WithScope(t.Context(), testScope), &memory.SearchRequest{Query: "build", AppName: "ignored", UserID: "ignored"})
	if err != nil {
		t.Fatalf("SearchMemory: %v", err)
	}
	if len(resp.Memories) != 1 {
		t.Fatalf("memories = %v", resp.Memories)
	}
	e := resp.Memories[0]
	if e.ID != "w1" || e.Content.Parts[0].Text != "team uses pnpm" || e.CustomMetadata["scope"] != "workspace" || e.Timestamp.IsZero() {
		t.Fatalf("entry = %+v", e)
	}
}

func TestADKInterfaceIsANoOpWithoutScopeOrConfig(t *testing.T) {
	f := newFakeMem0(t)
	sess := newTestSession(t, nil)

	enabled := newService(t, f, true)
	if resp, err := enabled.SearchMemory(t.Context(), &memory.SearchRequest{Query: "q"}); err != nil || len(resp.Memories) != 0 {
		t.Fatalf("no-scope SearchMemory = %v, %v", resp, err)
	}
	if err := enabled.AddSessionToMemory(t.Context(), sess); err != nil {
		t.Fatalf("no-scope AddSessionToMemory = %v", err)
	}

	disabled := newService(t, f, false)
	ctx := WithScope(t.Context(), testScope)
	if resp, err := disabled.SearchMemory(ctx, &memory.SearchRequest{Query: "q"}); err != nil || len(resp.Memories) != 0 {
		t.Fatalf("disabled SearchMemory = %v, %v", resp, err)
	}
	if err := disabled.AddSessionToMemory(ctx, sess); err != nil {
		t.Fatalf("disabled AddSessionToMemory = %v", err)
	}
	if searches, adds := f.recorded(); len(searches)+len(adds) != 0 {
		t.Fatal("a no-op call reached mem0")
	}
}

// --- capture ---------------------------------------------------------------

func textEvent(inv, author string, parts ...*genai.Part) *session.Event {
	return &session.Event{
		InvocationID: inv,
		Author:       author,
		LLMResponse:  model.LLMResponse{Content: &genai.Content{Parts: parts}},
	}
}

func newTestSession(t *testing.T, events []*session.Event) session.Session {
	t.Helper()
	svc := session.InMemoryService()
	created, err := svc.Create(t.Context(), &session.CreateRequest{AppName: "app", UserID: "u", SessionID: "s1"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	for _, ev := range events {
		if err := svc.AppendEvent(t.Context(), created.Session, ev); err != nil {
			t.Fatalf("append event: %v", err)
		}
	}
	got, err := svc.Get(t.Context(), &session.GetRequest{AppName: "app", UserID: "u", SessionID: "s1"})
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	return got.Session
}

func TestTurnMessagesSelectsTheInvocationsText(t *testing.T) {
	sess := newTestSession(t, []*session.Event{
		// A previous turn: never re-sent.
		textEvent("inv-1", "user", genai.NewPartFromText("old question")),
		textEvent("inv-1", "helper", genai.NewPartFromText("old answer")),
		// This turn.
		textEvent("inv-2", "user", genai.NewPartFromText("what did we decide?"), genai.NewPartFromBytes([]byte{1}, "image/png")),
		textEvent("inv-2", "helper",
			genai.NewPartFromText("let me look"),
			genai.NewPartFromFunctionCall("search", map[string]any{"q": "decision"})),
		textEvent("inv-2", "user", genai.NewPartFromFunctionResponse("search", map[string]any{"r": "pnpm"})),
		textEvent("inv-2", "helper", &genai.Part{Text: "thinking about it", Thought: true}, genai.NewPartFromText("We chose pnpm.")),
	})

	got := TurnMessages(sess, "inv-2")
	want := []mem0.Message{
		{Role: "user", Content: "what did we decide?\n[image]"},
		{Role: "assistant", Content: "We chose pnpm."},
	}
	if len(got) != len(want) {
		t.Fatalf("messages = %+v, want %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("message %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestCaptureSendsOnlyTheTurnToWorkspaceMemory(t *testing.T) {
	f := newFakeMem0(t)
	svc := newService(t, f, true)
	sess := newTestSession(t, []*session.Event{
		textEvent("inv-1", "user", genai.NewPartFromText("old")),
		textEvent("inv-2", "user", genai.NewPartFromText("remember we use pnpm")),
		textEvent("inv-2", "helper", genai.NewPartFromText("Noted.")),
	})

	if err := svc.AddSessionToMemory(WithScope(t.Context(), testScope), sess); err != nil {
		t.Fatalf("AddSessionToMemory: %v", err)
	}
	_, adds := f.recorded()
	if len(adds) != 1 || adds[0]["user_id"] != "ws:ws1" {
		t.Fatalf("adds = %v", adds)
	}
	msgs, _ := adds[0]["messages"].([]any)
	if len(msgs) != 2 || !strings.Contains(adds[0]["messages"].([]any)[0].(map[string]any)["content"].(string), "pnpm") {
		t.Fatalf("messages = %v", adds[0]["messages"])
	}
}

func TestCaptureWithoutTextOrInvocation(t *testing.T) {
	f := newFakeMem0(t)
	svc := newService(t, f, true)
	sess := newTestSession(t, []*session.Event{
		textEvent("inv-2", "user", genai.NewPartFromFunctionResponse("approve", map[string]any{"ok": true})),
	})

	if err := svc.Capture(t.Context(), testScope, sess); err != nil {
		t.Fatalf("Capture of a text-less turn: %v", err)
	}
	if _, adds := f.recorded(); len(adds) != 0 {
		t.Fatalf("text-less turn reached mem0: %v", adds)
	}

	scope := testScope
	scope.InvocationID = ""
	if err := svc.Capture(t.Context(), scope, sess); err == nil {
		t.Fatal("Capture without an invocation ID succeeded")
	}
}

func TestTruncateRunes(t *testing.T) {
	if got := truncateRunes("héllo", 3); got != "hél…" {
		t.Fatalf("truncateRunes = %q", got)
	}
	if got := truncateRunes("héllo", 5); got != "héllo" {
		t.Fatalf("truncateRunes at limit = %q", got)
	}
}

// A workflow resume reuses the paused invocation's ID and rewraps the reply
// as a FunctionResponse: capture must send the typed reply and only what
// this turn appended.
func TestCaptureTurnHandlesAWorkflowResume(t *testing.T) {
	f := newFakeMem0(t)
	svc := newService(t, f, true)
	events := []*session.Event{
		// The turn that paused (same invocation ID as the resume).
		textEvent("inv-2", "user", genai.NewPartFromText("ship the release")),
		textEvent("inv-2", "helper", genai.NewPartFromText("Which version should I tag?")),
	}
	fromEvent := len(events)
	events = append(events,
		// The resume turn: the reply arrives as a FunctionResponse.
		textEvent("inv-2", "user", genai.NewPartFromFunctionResponse("human_input", map[string]any{"answer": "v2.1.0"})),
		textEvent("inv-2", "helper", genai.NewPartFromText("Tagged v2.1.0.")),
	)
	sess := newTestSession(t, events)

	err := svc.CaptureTurn(t.Context(), testScope, TurnInput{Session: sess, FromEvent: fromEvent, UserText: "v2.1.0"})
	if err != nil {
		t.Fatalf("CaptureTurn: %v", err)
	}
	_, adds := f.recorded()
	if len(adds) != 1 {
		t.Fatalf("adds = %d", len(adds))
	}
	var got []string
	for _, m := range adds[0]["messages"].([]any) {
		msg := m.(map[string]any)
		got = append(got, msg["role"].(string)+": "+msg["content"].(string))
	}
	if strings.Join(got, " | ") != "user: v2.1.0 | assistant: Tagged v2.1.0." {
		t.Fatalf("captured = %q", got)
	}
}

func TestAddRedactsSecrets(t *testing.T) {
	f := newFakeMem0(t)
	svc := newService(t, f, true)
	_, err := svc.Add(t.Context(), testScope, TargetWorkspace, []mem0.Message{{Role: "user", Content: "deploy key is ghp_abcdefghijklmnopqrstuvwxyz0123456789"}})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	_, adds := f.recorded()
	content := adds[0]["messages"].([]any)[0].(map[string]any)["content"].(string)
	if strings.Contains(content, "ghp_") || !strings.Contains(content, "[REDACTED]") {
		t.Fatalf("content = %q", content)
	}
}
