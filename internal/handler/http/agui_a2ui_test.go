package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	adkrunner "google.golang.org/adk/v2/runner"
	adksession "google.golang.org/adk/v2/session"

	"go.orx.me/apps/butter/internal/a2uitool"
	"go.orx.me/apps/butter/internal/repo/auth"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	"go.orx.me/apps/butter/internal/runtime/runner"
	"go.orx.me/apps/butter/internal/testsupport/openaifake"
	wsctx "go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// These tests drive the A2UI extension of the AG-UI endpoint through its
// HTTP contract with a real runner, a real in-memory session store, and an
// OpenAI-compatible fake model whose replies each test scripts. What they
// observe is what a client observes: the SSE stream, pre-stream errors, and
// the UI snapshot read back from the persisted session.

// wsAgentRepo resolves agents per workspace, so the same agent_id in another
// workspace is a different (or missing) agent — the isolation tests need it.
type wsAgentRepo struct {
	configrepo.AgentRepository
	agents []*agentsv1.Agent
}

func (r *wsAgentRepo) GetAgent(_ context.Context, workspaceID, agentID string) (*agentsv1.Agent, error) {
	for _, a := range r.agents {
		if a.GetWorkspaceId() == workspaceID && a.GetAgentId() == agentID {
			return a, nil
		}
	}
	return nil, configrepo.ErrNotFound
}

type a2uiHarness struct {
	t        *testing.T
	backend  *openaifake.Backend
	sessions adksession.Service
	router   *gin.Engine
	handler  *AGUIHandler
	guard    *fakeSessionGuard
	// titler, when set before build, titles threads after successful runs.
	titler AGUISessionTitler
	// recorder, when set before build, records the runner's Invocations.
	recorder runner.InvocationRecorder
	// dropCustom makes every CUSTOM frame fail to send, as a client that
	// disconnects mid-stream would, after the server already persisted it.
	dropCustom bool
}

// dropCustomWriter fails the write of every CUSTOM frame.
type dropCustomWriter struct {
	gin.ResponseWriter
}

func (w dropCustomWriter) Write(b []byte) (int, error) {
	if strings.Contains(string(b), `"type":"CUSTOM"`) {
		return 0, errors.New("client connection lost")
	}
	return w.ResponseWriter.Write(b)
}

// newA2UIHarness wires the AG-UI handler to a real runner over the given
// agents, as they are: requests come from a signed-in user, who reaches every
// agent the runner can run, unless asAPIToken or asRootToken says otherwise.
// Models are served by the fake.
func newA2UIHarness(t *testing.T, agents []agentsv1.Agent, models ...string) *a2uiHarness {
	t.Helper()
	h := &a2uiHarness{t: t, backend: openaifake.New(t), sessions: newStoreLikeSessions(), guard: &fakeSessionGuard{}}
	h.router = h.build(agents, models)
	return h
}

// restart rebuilds the handler and runner over the same session store, as a
// Pod restart would.
func (h *a2uiHarness) restart(agents []agentsv1.Agent, models ...string) {
	h.router = h.build(agents, models)
}

func (h *a2uiHarness) build(agents []agentsv1.Agent, models []string) *gin.Engine {
	t := h.t
	modelCfgs := make([]*agentsv1.ModelConfig, 0, len(models))
	for _, m := range models {
		modelCfgs = append(modelCfgs, &agentsv1.ModelConfig{Name: m})
	}
	providers := []agentsv1.ModelProvider{{Name: "fake", Type: "openai", BaseUrl: h.backend.URL(), Models: modelCfgs}}
	svc, err := runner.NewService(context.Background(), agents, providers,
		nil, nil, nil, h.sessions, nil, nil, adkrunner.PluginConfig{})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	if h.recorder != nil {
		svc.SetInvocationRecorder(h.recorder)
	}
	repo := &wsAgentRepo{}
	for i := range agents {
		repo.agents = append(repo.agents, &agents[i])
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	// Stand-in for the auth middleware: the caller and workspace come from
	// test headers so isolation tests can vary them per request. A token
	// caller has no user, as the real middleware leaves it.
	r.Use(func(c *gin.Context) {
		ws := c.GetHeader("X-Workspace-ID")
		if ws == "" {
			ws = "ws-a"
		}
		ctx := wsctx.WithID(c.Request.Context(), ws)
		switch c.GetHeader("X-Test-Token") {
		case "api":
			ctx = auth.WithAPIToken(ctx, "token-1")
		case "root":
			ctx = auth.WithAdmin(ctx)
		default:
			user := c.GetHeader("X-Test-User")
			if user == "" {
				user = "u1"
			}
			ctx = auth.WithAuthenticated(ctx, &agentsv1.User{Id: user}, nil)
		}
		c.Request = c.Request.WithContext(ctx)
		if h.dropCustom {
			c.Writer = dropCustomWriter{c.Writer}
		}
		c.Next()
	})
	handler := NewAGUIHandler(repo)
	handler.SetRunnerService(svc)
	handler.SetSessionService(h.sessions)
	handler.SetSessionGuard(h.guard)
	if h.titler != nil {
		handler.SetSessionTitler(h.titler)
	}
	handler.Register(r)
	h.handler = handler
	return r
}

type a2uiRequest struct {
	workspace string
	user      string
	token     string
}

type a2uiOpt func(*a2uiRequest)

func asUser(user string) a2uiOpt    { return func(r *a2uiRequest) { r.user = user } }
func inWorkspace(ws string) a2uiOpt { return func(r *a2uiRequest) { r.workspace = ws } }

// asAPIToken and asRootToken send the request with a token instead of a
// signed-in user's session.
func asAPIToken() a2uiOpt  { return func(r *a2uiRequest) { r.token = "api" } }
func asRootToken() a2uiOpt { return func(r *a2uiRequest) { r.token = "root" } }

func (r a2uiRequest) apply(req *http.Request) {
	if r.workspace != "" {
		req.Header.Set("X-Workspace-ID", r.workspace)
	}
	if r.user != "" {
		req.Header.Set("X-Test-User", r.user)
	}
	if r.token != "" {
		req.Header.Set("X-Test-Token", r.token)
	}
}

func (h *a2uiHarness) post(agentID string, body map[string]any, opts ...a2uiOpt) *httptest.ResponseRecorder {
	h.t.Helper()
	var ro a2uiRequest
	for _, o := range opts {
		o(&ro)
	}
	payload, err := json.Marshal(body)
	if err != nil {
		h.t.Fatalf("marshal: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/agui/"+agentID, strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/json")
	ro.apply(req)
	w := httptest.NewRecorder()
	h.router.ServeHTTP(w, req)
	return w
}

// a2uiCapability is the forwardedProps declaration of an A2UI-capable client.
func a2uiCapability() map[string]any {
	return map[string]any{"butterA2UI": map[string]any{
		"version":  "v0.9.1",
		"catalogs": []any{"butter-basic-v1"},
	}}
}

// a2uiBody is a user turn from an A2UI-capable client.
func a2uiBody(threadID, text string) map[string]any {
	body := minimalAGUIBody(threadID, text)
	body["forwardedProps"] = a2uiCapability()
	return body
}

// sseEvents decodes an AG-UI SSE body into its events, in order.
func sseEvents(t *testing.T, body string) []map[string]any {
	t.Helper()
	var events []map[string]any
	for _, frame := range strings.Split(body, "\n\n") {
		for _, line := range strings.Split(frame, "\n") {
			data, ok := strings.CutPrefix(line, "data: ")
			if !ok {
				continue
			}
			var ev map[string]any
			if err := json.Unmarshal([]byte(data), &ev); err != nil {
				t.Fatalf("decode SSE frame %q: %v", data, err)
			}
			events = append(events, ev)
		}
	}
	return events
}

// a2uiValues returns the value of every butter.a2ui CUSTOM event, in order.
func a2uiValues(events []map[string]any) []map[string]any {
	var out []map[string]any
	for _, ev := range events {
		if ev["type"] == "CUSTOM" && ev["name"] == "butter.a2ui" {
			v, _ := ev["value"].(map[string]any)
			out = append(out, v)
		}
	}
	return out
}

// envelopeKind names the A2UI message an event value carries.
func envelopeKind(v map[string]any) string {
	env, _ := v["envelope"].(map[string]any)
	for _, k := range []string{"createSurface", "updateComponents", "updateDataModel", "deleteSurface"} {
		if _, ok := env[k]; ok {
			return k
		}
	}
	return ""
}

func eventIndex(events []map[string]any, match func(map[string]any) bool) int {
	for i, ev := range events {
		if match(ev) {
			return i
		}
	}
	return -1
}

func cardAgent() agentsv1.Agent {
	return agentsv1.Agent{
		Name: "Carder", AgentId: "carder", WorkspaceId: "ws-a",
		Config: &agentsv1.AgentConfig{Model: "card-model"},
	}
}

// deployCardArgs is a valid read-only card: a title, a body, a key-value
// result, and a status, with the body text bound to the data model.
func deployCardArgs() map[string]any {
	return map[string]any{
		"messages": []any{
			map[string]any{"updateComponents": map[string]any{"components": []any{
				map[string]any{"id": "root", "component": "Card", "child": "col"},
				map[string]any{"id": "col", "component": "Column", "children": []any{"title", "body", "env", "state"}},
				map[string]any{"id": "title", "component": "Text", "text": "Deploy summary", "variant": "h3"},
				map[string]any{"id": "body", "component": "Text", "text": map[string]any{"path": "/summary"}},
				map[string]any{"id": "env", "component": "KeyValue", "label": "Environment", "value": "production"},
				map[string]any{"id": "state", "component": "Status", "text": "Healthy", "tone": "success"},
			}}},
			map[string]any{"updateDataModel": map[string]any{"path": "/", "value": map[string]any{"summary": "3 services rolled out."}}},
		},
		"fallback": "Deploy summary: 3 services rolled out to production (healthy).",
	}
}

// scriptToolThenText makes model call tool once with args, then answer text
// once the tool result is in the conversation.
func (h *a2uiHarness) scriptToolThenText(model, tool string, args map[string]any, text string) {
	raw, err := json.Marshal(args)
	if err != nil {
		h.t.Fatalf("marshal args: %v", err)
	}
	h.backend.ScriptRequest(model, func(w http.ResponseWriter, req openaifake.ChatCompletionRequest) {
		if req.LastRole() == "tool" {
			openaifake.WriteReply(w, req, text)
			return
		}
		openaifake.WriteReply(w, req, "", openaifake.ToolCall{ID: "call-1", Name: tool, Arguments: string(raw)})
	})
}

// lastToolResult returns the tool-result content the model received last.
func (h *a2uiHarness) lastToolResult(model string) map[string]any {
	h.t.Helper()
	req, ok := h.backend.LastRequest(model)
	if !ok {
		h.t.Fatalf("model %s was never called", model)
	}
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role != "tool" {
			continue
		}
		var content string
		if err := json.Unmarshal(req.Messages[i].Content, &content); err != nil {
			h.t.Fatalf("tool content: %v", err)
		}
		var out map[string]any
		if err := json.Unmarshal([]byte(content), &out); err != nil {
			h.t.Fatalf("tool result %q is not JSON: %v", content, err)
		}
		return out
	}
	h.t.Fatal("no tool result reached the model")
	return nil
}

// --- #351: the model renders a read-only result card ---

// An A2UI-capable client's run offers render_ui; the card the model renders
// is persisted and then streamed as ordered butter.a2ui CUSTOM events — the
// create, component and data envelopes of one surface — alongside the
// tool-call events and the text answer of the same run.
func TestAGUIA2UI_ModelRendersReadOnlyCard(t *testing.T) {
	h := newA2UIHarness(t, []agentsv1.Agent{cardAgent()}, "card-model")
	h.scriptToolThenText("card-model", "render_ui", deployCardArgs(), "Deployed.")

	w := h.post("carder", a2uiBody("t-card", "deploy please"))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	events := sseEvents(t, w.Body.String())
	values := a2uiValues(events)
	if len(values) != 3 {
		t.Fatalf("butter.a2ui events = %d, want create+components+data (3)\n%s", len(values), w.Body.String())
	}
	wantKinds := []string{"createSurface", "updateComponents", "updateDataModel"}
	surfaceID, _ := values[0]["surfaceId"].(string)
	if surfaceID == "" {
		t.Fatalf("event carries no server-assigned surfaceId: %+v", values[0])
	}
	var lastSeq float64 = -1
	for i, v := range values {
		if got := envelopeKind(v); got != wantKinds[i] {
			t.Errorf("event %d envelope = %s, want %s", i, got, wantKinds[i])
		}
		if v["version"] != "v0.9.1" || v["kind"] != "card" || v["surfaceId"] != surfaceID {
			t.Errorf("event %d header = %+v", i, v)
		}
		if v["threadId"] != "t-card" || v["runId"] != "run-1" || v["messageId"] == "" {
			t.Errorf("event %d lacks message association: %+v", i, v)
		}
		env, _ := v["envelope"].(map[string]any)
		if env["version"] != "v0.9.1" {
			t.Errorf("event %d envelope version = %v", i, env["version"])
		}
		if v["revision"] != float64(1) {
			t.Errorf("event %d revision = %v, want 1", i, v["revision"])
		}
		seq, _ := v["seq"].(float64)
		if seq <= lastSeq {
			t.Errorf("event %d seq %v does not increase", i, seq)
		}
		lastSeq = seq
	}
	create := values[0]["envelope"].(map[string]any)["createSurface"].(map[string]any)
	if create["catalogId"] != "butter-basic-v1" || create["surfaceId"] != surfaceID {
		t.Errorf("createSurface = %+v", create)
	}
	if values[0]["fallback"] != "Deploy summary: 3 services rolled out to production (healthy)." {
		t.Errorf("fallback = %v", values[0]["fallback"])
	}

	// The tool call streams too, and the card precedes the run's end.
	toolStart := eventIndex(events, func(ev map[string]any) bool {
		return ev["type"] == "TOOL_CALL_START" && ev["toolCallName"] == "render_ui"
	})
	finished := eventIndex(events, func(ev map[string]any) bool { return ev["type"] == "RUN_FINISHED" })
	firstCard := eventIndex(events, func(ev map[string]any) bool { return ev["type"] == "CUSTOM" })
	if toolStart < 0 || !(toolStart < firstCard && firstCard < finished) {
		t.Fatalf("event order: tool start %d, first card %d, finished %d\n%s", toolStart, firstCard, finished, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"delta":"Deployed."`) {
		t.Errorf("text answer missing:\n%s", w.Body.String())
	}

	// The model learns the handle it can update later.
	result := h.lastToolResult("card-model")
	if result["surface_id"] != surfaceID {
		t.Errorf("tool result = %+v, want surface_id %s", result, surfaceID)
	}
}

func (h *a2uiHarness) snapshot(agentID, threadID string, opts ...a2uiOpt) (int, map[string]any) {
	h.t.Helper()
	var ro a2uiRequest
	for _, o := range opts {
		o(&ro)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/agui/"+agentID+"/threads/"+threadID+"/ui", nil)
	ro.apply(req)
	w := httptest.NewRecorder()
	h.router.ServeHTTP(w, req)
	var body map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	return w.Code, body
}

// snapshotSurfaces returns the snapshot's surfaces, failing on a non-200.
func (h *a2uiHarness) snapshotSurfaces(agentID, threadID string, opts ...a2uiOpt) []map[string]any {
	h.t.Helper()
	code, body := h.snapshot(agentID, threadID, opts...)
	if code != http.StatusOK {
		h.t.Fatalf("snapshot status = %d, body = %+v", code, body)
	}
	list, _ := body["surfaces"].([]any)
	out := make([]map[string]any, 0, len(list))
	for _, item := range list {
		m, _ := item.(map[string]any)
		out = append(out, m)
	}
	return out
}

// offeredTools lists the tool names the model was offered on its last call.
func (h *a2uiHarness) offeredTools(model string) []string {
	h.t.Helper()
	req, ok := h.backend.LastRequest(model)
	if !ok {
		h.t.Fatalf("model %s was never called", model)
	}
	var names []string
	tools, _ := req.Decoded["tools"].([]any)
	for _, t := range tools {
		fn, _ := t.(map[string]any)["function"].(map[string]any)
		if name, ok := fn["name"].(string); ok {
			names = append(names, name)
		}
	}
	return names
}

func containsString(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// A client that does not declare A2UI — or declares a version or catalog
// butter does not serve — gets plain chat: no render_ui for the model and no
// butter.a2ui events.
func TestAGUIA2UI_WithoutCapabilityStaysPlainChat(t *testing.T) {
	cases := map[string]any{
		"no declaration":      nil,
		"unsupported version": map[string]any{"butterA2UI": map[string]any{"version": "v1.0", "catalogs": []any{"butter-basic-v1"}}},
		"unsupported catalog": map[string]any{"butterA2UI": map[string]any{"version": "v0.9.1", "catalogs": []any{"https://a2ui.org/basic"}}},
	}
	for name, props := range cases {
		t.Run(name, func(t *testing.T) {
			h := newA2UIHarness(t, []agentsv1.Agent{cardAgent()}, "card-model")
			h.backend.ScriptRequest("card-model", func(w http.ResponseWriter, req openaifake.ChatCompletionRequest) {
				openaifake.WriteReply(w, req, "plain answer")
			})
			body := minimalAGUIBody("t-plain", "hi")
			body["forwardedProps"] = props
			w := h.post("carder", body)
			if w.Code != http.StatusOK {
				t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
			}
			if containsString(h.offeredTools("card-model"), "render_ui") {
				t.Errorf("render_ui offered without A2UI capability: %v", h.offeredTools("card-model"))
			}
			if got := a2uiValues(sseEvents(t, w.Body.String())); len(got) != 0 {
				t.Errorf("butter.a2ui events sent to a plain client: %+v", got)
			}
			if !strings.Contains(w.Body.String(), `"delta":"plain answer"`) {
				t.Errorf("text answer missing:\n%s", w.Body.String())
			}
		})
	}
}

// A declaration that is not shaped like one is refused before the stream,
// rather than silently read as "no A2UI".
func TestAGUIA2UI_MalformedCapabilityIsBadRequest(t *testing.T) {
	h := newA2UIHarness(t, []agentsv1.Agent{cardAgent()}, "card-model")
	body := minimalAGUIBody("t-bad", "hi")
	body["forwardedProps"] = map[string]any{"butterA2UI": "yes please"}
	w := h.post("carder", body)
	if w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), "butterA2UI") {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	if h.backend.CallCount("card-model") != 0 {
		t.Fatal("the agent ran despite a malformed declaration")
	}
}

// An invalid card is rejected whole: the model gets a readable tool error,
// nothing is streamed as UI, nothing is persisted, and the run still
// finishes with the model's text answer.
func TestAGUIA2UI_InvalidCardIsReadableToolError(t *testing.T) {
	component := func(c map[string]any) map[string]any {
		return map[string]any{"messages": []any{map[string]any{"updateComponents": map[string]any{"components": []any{c}}}}, "fallback": "x"}
	}
	// root plus 100 rows: one component over the limit.
	many := make([]any, 0, 101)
	children := make([]any, 0, 100)
	for i := 0; i < 100; i++ {
		id := "row" + strconv.Itoa(i)
		children = append(children, id)
		many = append(many, map[string]any{"id": id, "component": "Text", "text": "row"})
	}
	many = append([]any{map[string]any{"id": "root", "component": "Column", "children": children}}, many...)

	cases := map[string]struct {
		args map[string]any
		want string
	}{
		"raw HTML":    {component(map[string]any{"id": "root", "component": "Text", "text": "<script>alert(1)</script>"}), "HTML"},
		"URL in text": {component(map[string]any{"id": "root", "component": "Text", "text": "details at https://evil.example/login"}), "URL"},
		"script URL in data": {map[string]any{"messages": []any{
			map[string]any{"updateComponents": map[string]any{"components": []any{map[string]any{"id": "root", "component": "Text", "text": map[string]any{"path": "/t"}}}}},
			map[string]any{"updateDataModel": map[string]any{"path": "/", "value": map[string]any{"t": "javascript:alert(1)"}}},
		}, "fallback": "x"}, "URL"},
		"unknown component":    {component(map[string]any{"id": "root", "component": "Image", "url": "https://example.com/x.png"}), "unknown component"},
		"model-defined action": {component(map[string]any{"id": "root", "component": "Button", "child": "x", "action": map[string]any{"event": map[string]any{"name": "delete_all"}}}), "read-only"},
		"input component":      {component(map[string]any{"id": "root", "component": "TextField", "label": "Name", "value": map[string]any{"path": "/n"}}), "read-only"},
		"unknown property":     {component(map[string]any{"id": "root", "component": "Text", "text": "hi", "onClick": "x"}), "no property"},
		"function binding":     {component(map[string]any{"id": "root", "component": "Text", "text": map[string]any{"call": "fetch", "args": map[string]any{}}}), "path"},
		"missing root":         {component(map[string]any{"id": "top", "component": "Text", "text": "hi"}), "root"},
		"dangling reference":   {component(map[string]any{"id": "root", "component": "Card", "child": "nowhere"}), "does not exist"},
		"cycle": {map[string]any{"messages": []any{map[string]any{"updateComponents": map[string]any{"components": []any{
			map[string]any{"id": "root", "component": "Column", "children": []any{"a"}},
			map[string]any{"id": "a", "component": "Card", "child": "root"},
		}}}}, "fallback": "x"}, "ancestor"},
		"too many components":    {map[string]any{"messages": []any{map[string]any{"updateComponents": map[string]any{"components": many}}}, "fallback": "x"}, "at most 100"},
		"missing fallback":       {map[string]any{"messages": deployCardArgs()["messages"]}, "fallback"},
		"model picks the handle": {map[string]any{"surface_id": "card-mine", "messages": deployCardArgs()["messages"], "fallback": "x"}, "not a card"},
		"createSurface":          {map[string]any{"messages": []any{map[string]any{"createSurface": map[string]any{"surfaceId": "x", "catalogId": "butter-basic-v1"}}}, "fallback": "x"}, "createSurface"},
	}
	// The batch limit is on the encoded messages: 22 texts of 3000 bytes.
	big := make([]any, 0, 23)
	kids := make([]any, 0, 22)
	for i := 0; i < 22; i++ {
		id := "p" + strconv.Itoa(i)
		kids = append(kids, id)
		big = append(big, map[string]any{"id": id, "component": "Text", "text": strings.Repeat("a", 3000)})
	}
	big = append(big, map[string]any{"id": "root", "component": "Column", "children": kids})
	cases["batch too large"] = struct {
		args map[string]any
		want string
	}{map[string]any{"messages": []any{map[string]any{"updateComponents": map[string]any{"components": big}}}, "fallback": "x"}, "at most 65536"}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			h := newA2UIHarness(t, []agentsv1.Agent{cardAgent()}, "card-model")
			h.scriptToolThenText("card-model", "render_ui", tc.args, "Here it is in text.")
			w := h.post("carder", a2uiBody("t-bad-card", "show me"))
			if w.Code != http.StatusOK {
				t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
			}
			if got := a2uiValues(sseEvents(t, w.Body.String())); len(got) != 0 {
				t.Fatalf("an invalid card was streamed: %+v", got)
			}
			result := h.lastToolResult("card-model")
			msg, _ := result["error"].(string)
			if msg == "" || !strings.Contains(msg, tc.want) || !strings.Contains(msg, "nothing was changed") {
				t.Fatalf("tool result = %+v, want a readable error mentioning %q", result, tc.want)
			}
			if !strings.Contains(w.Body.String(), `"delta":"Here it is in text."`) {
				t.Errorf("the run did not continue to the text answer:\n%s", w.Body.String())
			}
			if got := h.snapshotSurfaces("carder", "t-bad-card"); len(got) != 0 {
				t.Errorf("an invalid card was persisted: %+v", got)
			}
		})
	}
}

// The persisted card is what a later snapshot returns — the same envelopes,
// revision and message association — and a Pod restart (a new handler and
// runner over the same store) loses nothing.
func TestAGUIA2UI_CardSurvivesInSnapshotAndRestart(t *testing.T) {
	agents := []agentsv1.Agent{cardAgent()}
	h := newA2UIHarness(t, agents, "card-model")
	h.scriptToolThenText("card-model", "render_ui", deployCardArgs(), "Deployed.")
	w := h.post("carder", a2uiBody("t-snap", "deploy"))
	streamed := a2uiValues(sseEvents(t, w.Body.String()))
	if len(streamed) != 3 {
		t.Fatalf("streamed %d events", len(streamed))
	}

	h.restart([]agentsv1.Agent{cardAgent()}, "card-model")
	calls := h.backend.CallCount("card-model")
	surfaces := h.snapshotSurfaces("carder", "t-snap")
	if len(surfaces) != 1 {
		t.Fatalf("snapshot surfaces = %+v", surfaces)
	}
	s := surfaces[0]
	if s["surfaceId"] != streamed[0]["surfaceId"] || s["kind"] != "card" || s["revision"] != float64(1) {
		t.Errorf("snapshot surface = %+v", s)
	}
	if s["messageId"] != streamed[0]["messageId"] || s["runId"] != "run-1" {
		t.Errorf("snapshot lost the message association: %+v", s)
	}
	envs, _ := s["envelopes"].([]any)
	if len(envs) != 3 {
		t.Fatalf("snapshot envelopes = %+v", envs)
	}
	for i, env := range envs {
		want, _ := json.Marshal(streamed[i]["envelope"])
		got, _ := json.Marshal(env)
		if string(got) != string(want) {
			t.Errorf("snapshot envelope %d = %s, streamed %s", i, got, want)
		}
	}
	if h.backend.CallCount("card-model") != calls {
		t.Error("reading the snapshot ran the agent")
	}
}

// The same caller reusing a threadId under another agent lands on the same
// session (its key is caller + thread), so the thread's binding decides:
// reusing it with another agent, from another workspace, or as another user
// is refused before anything runs, and their snapshots are empty. The UI
// stays with the context that created it.
func TestAGUIA2UI_ThreadIDReuseIsIsolated(t *testing.T) {
	h := newA2UIHarness(t, []agentsv1.Agent{
		cardAgent(),
		{Name: "Other", AgentId: "other", WorkspaceId: "ws-a", Config: &agentsv1.AgentConfig{Model: "other-model"}},
		// The same agent_id in another workspace is another agent.
		{Name: "CarderB", AgentId: "carder", WorkspaceId: "ws-b", Config: &agentsv1.AgentConfig{Model: "other-model"}},
	}, "card-model", "other-model")
	h.scriptToolThenText("card-model", "render_ui", deployCardArgs(), "Deployed.")
	h.backend.ScriptRequest("other-model", func(w http.ResponseWriter, req openaifake.ChatCompletionRequest) {
		openaifake.WriteReply(w, req, "other answer")
	})
	if w := h.post("carder", a2uiBody("t-shared", "deploy")); len(a2uiValues(sseEvents(t, w.Body.String()))) != 3 {
		t.Fatalf("setup: card not rendered\n%s", w.Body.String())
	}

	for name, tc := range map[string]struct {
		agentID string
		opts    []a2uiOpt
		want    string
	}{
		"another agent":     {agentID: "other", want: errThreadOfAnotherAgent.Error()},
		"another workspace": {agentID: "carder", opts: []a2uiOpt{inWorkspace("ws-b")}, want: errThreadUnavailable.Error()},
		"another user":      {agentID: "carder", opts: []a2uiOpt{asUser("u2")}, want: errThreadUnavailable.Error()},
	} {
		t.Run(name, func(t *testing.T) {
			calls := h.backend.CallCount("card-model") + h.backend.CallCount("other-model")
			w := h.post(tc.agentID, a2uiBody("t-shared", "hello"), tc.opts...)
			if w.Code != http.StatusForbidden || !strings.Contains(w.Body.String(), tc.want) {
				t.Fatalf("status = %d, body = %s; want 403 %q", w.Code, w.Body.String(), tc.want)
			}
			if h.backend.CallCount("card-model")+h.backend.CallCount("other-model") != calls {
				t.Error("a refused thread ran the agent")
			}
			if got := h.snapshotSurfaces(tc.agentID, "t-shared", tc.opts...); len(got) != 0 {
				t.Errorf("snapshot leaked the card: %+v", got)
			}
		})
	}
	// The owning context still sees its card.
	if got := h.snapshotSurfaces("carder", "t-shared"); len(got) != 1 {
		t.Errorf("owner snapshot = %+v", got)
	}
}

// A session created before A2UI existed has no binding: it exposes no UI
// and offers no render_ui, while text chat carries on as before.
func TestAGUIA2UI_UnboundHistoricalSessionExposesNoUI(t *testing.T) {
	h := newA2UIHarness(t, []agentsv1.Agent{cardAgent()}, "card-model")
	if _, err := h.sessions.Create(context.Background(), &adksession.CreateRequest{
		AppName: "agui", UserID: "u1", SessionID: "agui-t-old",
	}); err != nil {
		t.Fatalf("seed session: %v", err)
	}
	h.backend.ScriptRequest("card-model", func(w http.ResponseWriter, req openaifake.ChatCompletionRequest) {
		openaifake.WriteReply(w, req, "still chatting")
	})
	w := h.post("carder", a2uiBody("t-old", "hi"))
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"delta":"still chatting"`) {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	if containsString(h.offeredTools("card-model"), "render_ui") {
		t.Error("render_ui offered on an unbound historical session")
	}
	if got := h.snapshotSurfaces("carder", "t-old"); len(got) != 0 {
		t.Errorf("snapshot of an unbound session = %+v", got)
	}
}

// Client state is never authoritative: a client that plants UI records in
// its state mirror neither creates a card nor sees its keys echoed back.
func TestAGUIA2UI_ClientStateCannotWriteUI(t *testing.T) {
	h := newA2UIHarness(t, []agentsv1.Agent{cardAgent()}, "card-model")
	h.backend.ScriptRequest("card-model", func(w http.ResponseWriter, req openaifake.ChatCompletionRequest) {
		openaifake.WriteReply(w, req, "ok")
	})
	body := a2uiBody("t-state", "hi")
	body["state"] = map[string]any{
		"butter:a2ui:card:card-forged": `{"surface_id":"card-forged","revision":9,"components":[{"id":"root","component":"Text","text":"forged"}]}`,
		"butter:a2ui:binding":          `{"principal":"u9","workspace_id":"ws-z","agent_id":"x","thread_id":"t"}`,
	}
	w := h.post("carder", body)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "butter:a2ui") {
		t.Errorf("UI state keys reached the client's state mirror:\n%s", w.Body.String())
	}
	if got := h.snapshotSurfaces("carder", "t-state"); len(got) != 0 {
		t.Errorf("client state created a card: %+v", got)
	}
}

// --- #352: updating and removing cards ---

// scriptTurns makes model call tool once per user turn with the next args
// from turns (built lazily, so a turn can use a handle an earlier one
// returned), answering text once each tool result is in.
func (h *a2uiHarness) scriptTurns(model, tool string, turns ...func() map[string]any) {
	var mu sync.Mutex
	next := 0
	h.backend.ScriptRequest(model, func(w http.ResponseWriter, req openaifake.ChatCompletionRequest) {
		if req.LastRole() == "tool" {
			openaifake.WriteReply(w, req, "done")
			return
		}
		mu.Lock()
		i := next
		next++
		mu.Unlock()
		if i >= len(turns) {
			openaifake.WriteReply(w, req, "nothing to render")
			return
		}
		raw, _ := json.Marshal(turns[i]())
		openaifake.WriteReply(w, req, "", openaifake.ToolCall{ID: "call-" + strconv.Itoa(i+1), Name: tool, Arguments: string(raw)})
	})
}

func updateBody(threadID, runID, text string) map[string]any {
	body := a2uiBody(threadID, text)
	body["runId"] = runID
	return body
}

// An update to an existing card re-renders the same surface: the stream
// carries only the changed component and the new data at the next
// revision, and the snapshot holds the merged result. A delete removes the
// card from the snapshot and frees its slot.
func TestAGUIA2UI_CardUpdateAndDelete(t *testing.T) {
	h := newA2UIHarness(t, []agentsv1.Agent{cardAgent()}, "card-model")
	var surfaceID string
	h.scriptTurns("card-model", "render_ui",
		deployCardArgs,
		func() map[string]any {
			return map[string]any{"surface_id": surfaceID, "messages": []any{
				map[string]any{"updateComponents": map[string]any{"components": []any{
					map[string]any{"id": "state", "component": "Status", "text": "Degraded", "tone": "warning"},
				}}},
				map[string]any{"updateDataModel": map[string]any{"path": "/summary", "value": "1 service rolled back."}},
			}}
		},
		func() map[string]any {
			return map[string]any{"surface_id": surfaceID, "messages": []any{map[string]any{"deleteSurface": map[string]any{}}}}
		},
	)

	created := a2uiValues(sseEvents(t, h.post("carder", updateBody("t-life", "r1", "deploy")).Body.String()))
	if len(created) != 3 {
		t.Fatalf("create events = %+v", created)
	}
	surfaceID = created[0]["surfaceId"].(string)

	updated := a2uiValues(sseEvents(t, h.post("carder", updateBody("t-life", "r2", "status?")).Body.String()))
	if len(updated) != 2 || envelopeKind(updated[0]) != "updateComponents" || envelopeKind(updated[1]) != "updateDataModel" {
		t.Fatalf("update events = %+v", updated)
	}
	for _, v := range updated {
		if v["surfaceId"] != surfaceID || v["revision"] != float64(2) || v["runId"] != "r2" {
			t.Errorf("update event header = %+v", v)
		}
	}
	comps := updated[0]["envelope"].(map[string]any)["updateComponents"].(map[string]any)["components"].([]any)
	if len(comps) != 1 || comps[0].(map[string]any)["id"] != "state" {
		t.Errorf("update sent components %+v, want only the changed one", comps)
	}
	data := updated[1]["envelope"].(map[string]any)["updateDataModel"].(map[string]any)
	if data["path"] != "/" || data["value"].(map[string]any)["summary"] != "1 service rolled back." {
		t.Errorf("update data = %+v", data)
	}
	if got := h.lastToolResult("card-model"); got["surface_id"] != surfaceID || got["status"] != "updated" || got["revision"] != float64(2) {
		t.Errorf("update tool result = %+v", got)
	}

	snap := h.snapshotSurfaces("carder", "t-life")
	if len(snap) != 1 || snap[0]["revision"] != float64(2) || snap[0]["messageId"] != created[0]["messageId"] {
		t.Fatalf("snapshot after update = %+v", snap)
	}
	envs, _ := json.Marshal(snap[0]["envelopes"])
	if !strings.Contains(string(envs), "Degraded") || !strings.Contains(string(envs), "Deploy summary") || strings.Contains(string(envs), "Healthy") {
		t.Errorf("snapshot does not hold the merged card: %s", envs)
	}

	deleted := a2uiValues(sseEvents(t, h.post("carder", updateBody("t-life", "r3", "clear it")).Body.String()))
	if len(deleted) != 1 || envelopeKind(deleted[0]) != "deleteSurface" || deleted[0]["revision"] != float64(3) {
		t.Fatalf("delete events = %+v", deleted)
	}
	if got := h.snapshotSurfaces("carder", "t-life"); len(got) != 0 {
		t.Errorf("deleted card still in snapshot: %+v", got)
	}
}

// A batch that is invalid anywhere changes nothing: a valid first message
// followed by an invalid one leaves the card exactly as it was.
func TestAGUIA2UI_InvalidUpdateChangesNothing(t *testing.T) {
	h := newA2UIHarness(t, []agentsv1.Agent{cardAgent()}, "card-model")
	var surfaceID string
	h.scriptTurns("card-model", "render_ui",
		deployCardArgs,
		func() map[string]any {
			return map[string]any{"surface_id": surfaceID, "messages": []any{
				map[string]any{"updateComponents": map[string]any{"components": []any{
					map[string]any{"id": "title", "component": "Text", "text": "Changed title"},
				}}},
				map[string]any{"updateComponents": map[string]any{"components": []any{
					map[string]any{"id": "state", "component": "Status", "text": "x", "tone": "purple"},
				}}},
			}}
		},
		func() map[string]any {
			// A card from another thread is not addressable here.
			return map[string]any{"surface_id": surfaceID, "messages": []any{map[string]any{"deleteSurface": map[string]any{}}}}
		},
	)
	created := a2uiValues(sseEvents(t, h.post("carder", updateBody("t-inv", "r1", "deploy")).Body.String()))
	surfaceID = created[0]["surfaceId"].(string)

	w := h.post("carder", updateBody("t-inv", "r2", "retitle"))
	if got := a2uiValues(sseEvents(t, w.Body.String())); len(got) != 0 {
		t.Fatalf("a rejected batch streamed UI: %+v", got)
	}
	if msg, _ := h.lastToolResult("card-model")["error"].(string); !strings.Contains(msg, "tone") {
		t.Errorf("tool error = %q, want it to name the bad property", msg)
	}
	snap := h.snapshotSurfaces("carder", "t-inv")
	envs, _ := json.Marshal(snap[0]["envelopes"])
	if snap[0]["revision"] != float64(1) || strings.Contains(string(envs), "Changed title") {
		t.Errorf("rejected batch partially applied: revision %v, %s", snap[0]["revision"], envs)
	}

	// Another thread cannot address this thread's card.
	w = h.post("carder", updateBody("t-other-thread", "r3", "delete that"))
	if got := a2uiValues(sseEvents(t, w.Body.String())); len(got) != 0 {
		t.Fatalf("cross-thread delete streamed UI: %+v", got)
	}
	if msg, _ := h.lastToolResult("card-model")["error"].(string); !strings.Contains(msg, "not a card in this conversation") {
		t.Errorf("cross-thread tool error = %q", msg)
	}
	if got := h.snapshotSurfaces("carder", "t-inv"); len(got) != 1 {
		t.Errorf("card deleted from another thread: %+v", got)
	}
}

// A session holds at most 20 read-only cards; the 21st is refused with a
// readable error, and deleting one frees the slot again.
func TestAGUIA2UI_CardQuotaAndRelease(t *testing.T) {
	h := newA2UIHarness(t, []agentsv1.Agent{cardAgent()}, "card-model")
	var first string
	turns := make([]func() map[string]any, 0, 23)
	for i := 0; i < 21; i++ {
		turns = append(turns, deployCardArgs)
	}
	turns = append(turns,
		func() map[string]any {
			return map[string]any{"surface_id": first, "messages": []any{map[string]any{"deleteSurface": map[string]any{}}}}
		},
		deployCardArgs,
	)
	h.scriptTurns("card-model", "render_ui", turns...)

	for i := 0; i < 20; i++ {
		values := a2uiValues(sseEvents(t, h.post("carder", updateBody("t-quota", "r"+strconv.Itoa(i), "card")).Body.String()))
		if len(values) != 3 {
			t.Fatalf("card %d not created: %+v", i+1, values)
		}
		if i == 0 {
			first = values[0]["surfaceId"].(string)
		}
	}
	if got := a2uiValues(sseEvents(t, h.post("carder", updateBody("t-quota", "r20", "one more")).Body.String())); len(got) != 0 {
		t.Fatalf("21st card streamed: %+v", got)
	}
	if msg, _ := h.lastToolResult("card-model")["error"].(string); !strings.Contains(msg, "20 cards") {
		t.Errorf("quota error = %q", msg)
	}
	if got := a2uiValues(sseEvents(t, h.post("carder", updateBody("t-quota", "r21", "drop the first")).Body.String())); len(got) != 1 {
		t.Fatalf("delete events = %+v", got)
	}
	if got := a2uiValues(sseEvents(t, h.post("carder", updateBody("t-quota", "r22", "now one more")).Body.String())); len(got) != 3 {
		t.Fatalf("card after freeing a slot not created: %+v", got)
	}
	if got := h.snapshotSurfaces("carder", "t-quota"); len(got) != 20 {
		t.Errorf("snapshot holds %d cards, want 20", len(got))
	}
}

// --- #353/#354: Human Input forms ---

func deployForm() *agentsv1.HumanInputForm {
	return &agentsv1.HumanInputForm{
		Title: "Deploy approval",
		Fields: []*agentsv1.HumanInputFormField{
			{
				Name: "env", Label: "Environment", Required: true,
				Type: agentsv1.HumanInputFormFieldType_HUMAN_INPUT_FORM_FIELD_TYPE_SINGLE_CHOICE,
				Options: []*agentsv1.HumanInputFormOption{
					{Value: "prod", Label: "Production"},
					{Value: "staging", Label: "Staging"},
				},
			},
			{
				Name: "reason", Label: "Reason", Hint: "Why now?", Required: true, MaxLength: 20,
				Type: agentsv1.HumanInputFormFieldType_HUMAN_INPUT_FORM_FIELD_TYPE_TEXT,
			},
			{
				Name: "note", Label: "Note",
				Type: agentsv1.HumanInputFormFieldType_HUMAN_INPUT_FORM_FIELD_TYPE_TEXT,
			},
		},
	}
}

// approvalWorkflow is draft → ask (Human Input, optionally with a form) →
// publish, the drafter and publisher being LLM agents on the fake models.
func approvalWorkflow(form *agentsv1.HumanInputForm) []agentsv1.Agent {
	return []agentsv1.Agent{{
		Name: "approval", AgentId: "approval", WorkspaceId: "ws-a",
		Type:          agentsv1.AgentType_AGENT_TYPE_WORKFLOW,
		ChildAgentIds: []string{"draft", "publish"},
		Config: &agentsv1.AgentConfig{Workflow: &agentsv1.WorkflowConfig{
			Nodes: []*agentsv1.WorkflowNode{
				{Name: "draft", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, AgentId: "draft"},
				{Name: "ask", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_HUMAN_INPUT, Question: "Approve this deploy?", Form: form},
				{Name: "publish", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, AgentId: "publish"},
			},
			Edges: []*agentsv1.WorkflowEdge{
				{From: "START", To: "draft"},
				{From: "draft", To: "ask"},
				{From: "ask", To: "publish"},
			},
		}},
	},
		{Name: "draft", AgentId: "draft", WorkspaceId: "ws-a", Config: &agentsv1.AgentConfig{Model: "drafter"}},
		{Name: "publish", AgentId: "publish", WorkspaceId: "ws-a", Config: &agentsv1.AgentConfig{Model: "publisher"}},
	}
}

// echoModels makes each model answer "<model>(<last user input>)".
func (h *a2uiHarness) echoModels(models ...string) {
	for _, m := range models {
		model := m
		h.backend.ScriptRequest(model, func(w http.ResponseWriter, req openaifake.ChatCompletionRequest) {
			openaifake.WriteReply(w, req, model+"("+h.backend.LastInput(model)+")")
		})
	}
}

// pausedForm runs the workflow to its Human Input node and returns the form
// events and the interrupt outcome.
func (h *a2uiHarness) pausedForm(threadID string) ([]map[string]any, map[string]any) {
	h.t.Helper()
	w := h.post("approval", a2uiBody(threadID, "write the release"))
	if w.Code != http.StatusOK {
		h.t.Fatalf("run 1 status = %d, body = %s", w.Code, w.Body.String())
	}
	events := sseEvents(h.t, w.Body.String())
	var finished map[string]any
	for _, ev := range events {
		if ev["type"] == "RUN_FINISHED" {
			finished = ev
		}
	}
	return a2uiValues(events), finished
}

func formSubmission(form map[string]any, surfaceID string, values map[string]any) map[string]any {
	return map[string]any{"butterForm": map[string]any{
		"version":   "v0.9.1",
		"surfaceId": surfaceID,
		"revision":  form["revision"],
		"token":     form["token"],
		"values":    values,
	}}
}

func resumeBody(threadID, runID, interruptID string, payload any) map[string]any {
	return map[string]any{
		"threadId":       threadID,
		"runId":          runID,
		"messages":       []any{},
		"forwardedProps": a2uiCapability(),
		"resume":         []any{map[string]any{"interruptId": interruptID, "status": "resolved", "payload": payload}},
	}
}

// storedAnswers returns the payloads of the request-input FunctionResponses
// persisted on a thread's session.
func (h *a2uiHarness) storedAnswers(threadID string) []any {
	h.t.Helper()
	resp, err := h.sessions.Get(context.Background(), &adksession.GetRequest{AppName: "agui", UserID: "u1", SessionID: "agui-" + threadID})
	if err != nil {
		h.t.Fatalf("get session: %v", err)
	}
	var out []any
	for ev := range resp.Session.Events().All() {
		if ev.Content == nil {
			continue
		}
		for _, p := range ev.Content.Parts {
			if fr := p.FunctionResponse; fr != nil && fr.Name == "adk_request_input" {
				out = append(out, fr.Response["payload"])
			}
		}
	}
	return out
}

// The whole form path on a real Workflow Agent: the Human Input node pauses
// and the A2UI client receives a server-built form bound to the Interrupt;
// submitting it resumes exactly that Interrupt with the fields encoded as a
// JSON object text in configured order; the successor runs on that text,
// the form is marked answered, and the answer is persisted on the session.
func TestAGUIA2UI_FormResumesWorkflowWithStructuredAnswer(t *testing.T) {
	h := newA2UIHarness(t, approvalWorkflow(deployForm()), "drafter", "publisher")
	h.echoModels("drafter", "publisher")

	values, finished := h.pausedForm("t-form")
	if len(values) != 3 {
		t.Fatalf("form events = %+v", values)
	}
	outcome := finished["outcome"].(map[string]any)
	interrupts := outcome["interrupts"].([]any)
	if outcome["type"] != "interrupt" || len(interrupts) != 1 {
		t.Fatalf("outcome = %+v", outcome)
	}
	interruptID := interrupts[0].(map[string]any)["id"].(string)
	// Clients without A2UI read the fields from the question text.
	if msg := interrupts[0].(map[string]any)["message"].(string); !strings.HasPrefix(msg, "Approve this deploy?") ||
		!strings.Contains(msg, "env: Environment (required; one of: prod (Production), staging (Staging))") ||
		!strings.Contains(msg, "reason: Reason (required; up to 20 characters) — Why now?") {
		t.Errorf("interrupt message lacks field instructions: %q", msg)
	}

	surfaceID := values[0]["surfaceId"].(string)
	for i, want := range []string{"createSurface", "updateComponents", "updateDataModel"} {
		v := values[i]
		if envelopeKind(v) != want || v["kind"] != "form" || v["surfaceId"] != surfaceID || v["revision"] != float64(1) {
			t.Errorf("form event %d = %+v", i, v)
		}
	}
	form := values[0]["form"].(map[string]any)
	if form["interruptId"] != interruptID || form["token"] == "" || form["title"] != "Deploy approval" || form["question"] != "Approve this deploy?" {
		t.Fatalf("form binding = %+v", form)
	}
	fields := form["fields"].([]any)
	if len(fields) != 3 || fields[0].(map[string]any)["name"] != "env" || fields[0].(map[string]any)["type"] != "single_choice" {
		t.Errorf("form fields = %+v", fields)
	}
	comps, _ := json.Marshal(values[1]["envelope"])
	for _, want := range []string{`"component":"ChoicePicker"`, `"name":"env"`, `"component":"TextField"`, `"component":"Button"`, `"name":"butter.submitForm"`} {
		if !strings.Contains(string(comps), want) {
			t.Errorf("form components lack %s: %s", want, comps)
		}
	}
	if h.backend.CallCount("publisher") != 0 {
		t.Fatal("publisher ran before the form was answered")
	}

	// Submit: fields arrive in a different order and with padding; the
	// answer is canonical.
	w := h.post("approval", resumeBody("t-form", "run-2", interruptID, formSubmission(form, surfaceID, map[string]any{
		"reason": "  hotfix  ",
		"env":    "prod",
	})))
	if w.Code != http.StatusOK {
		t.Fatalf("submit status = %d, body = %s", w.Code, w.Body.String())
	}
	// Butter delivers the fields in configured order, every field present
	// and nothing else; that text is what the session stores. ADK's
	// workflow engine parses a JSON answer when it resumes and hands the
	// successor the same object re-encoded, so compare it as an object.
	const want = `{"env":"prod","reason":"hotfix","note":""}`
	var gotObj, wantObj map[string]any
	if err := json.Unmarshal([]byte(h.backend.LastInput("publisher")), &gotObj); err != nil {
		t.Fatalf("publisher input %q is not a JSON object: %v", h.backend.LastInput("publisher"), err)
	}
	_ = json.Unmarshal([]byte(want), &wantObj)
	if !reflect.DeepEqual(gotObj, wantObj) {
		t.Fatalf("publisher input = %v, want %v", gotObj, wantObj)
	}
	events := sseEvents(t, w.Body.String())
	answered := a2uiValues(events)
	if len(answered) != 1 || answered[0]["surfaceId"] != surfaceID || answered[0]["revision"] != float64(2) {
		t.Fatalf("answered events = %+v", answered)
	}
	data := answered[0]["envelope"].(map[string]any)["updateDataModel"].(map[string]any)
	if data["path"] != "/status" || data["value"] != "answered" {
		t.Errorf("answered envelope = %+v", data)
	}
	last := events[len(events)-1]
	if last["type"] != "RUN_FINISHED" || last["outcome"].(map[string]any)["type"] != "success" {
		t.Errorf("run did not finish successfully: %+v", last)
	}
	if !strings.Contains(w.Body.String(), "publisher(") {
		t.Errorf("successor output missing:\n%s", w.Body.String())
	}
	if got := h.storedAnswers("t-form"); len(got) != 1 || got[0] != want {
		t.Errorf("persisted answers = %+v, want [%s]", got, want)
	}
	if got := h.snapshotSurfaces("approval", "t-form"); len(got) != 0 {
		t.Errorf("answered form still in snapshot: %+v", got)
	}
}

// formFixture pauses the approval workflow and returns what a client needs
// to submit its form.
type formFixture struct {
	interruptID string
	surfaceID   string
	form        map[string]any
}

func (h *a2uiHarness) pauseWithForm(threadID string) formFixture {
	h.t.Helper()
	values, _ := h.pausedForm(threadID)
	if len(values) == 0 {
		h.t.Fatal("no form was shown")
	}
	form := values[0]["form"].(map[string]any)
	return formFixture{interruptID: form["interruptId"].(string), surfaceID: values[0]["surfaceId"].(string), form: form}
}

func validDeployValues() map[string]any {
	return map[string]any{"env": "staging", "reason": "weekly"}
}

// A submission that does not match a pending form of this very context is
// refused before the stream opens: no answer is appended, the agent does
// not run, and the Interrupt stays pending — a later valid submission still
// resumes it. Field errors name the field.
func TestAGUIA2UI_FormRejectsBadSubmissions(t *testing.T) {
	h := newA2UIHarness(t, approvalWorkflow(deployForm()), "drafter", "publisher")
	h.echoModels("drafter", "publisher")
	fx := h.pauseWithForm("t-rej")

	sub := func(mut func(m map[string]any)) map[string]any {
		payload := formSubmission(fx.form, fx.surfaceID, validDeployValues())
		mut(payload["butterForm"].(map[string]any))
		return payload
	}
	cases := []struct {
		name        string
		interruptID string
		payload     map[string]any
		opts        []a2uiOpt
		status      int
		want        string
		field       string
		code        string
	}{
		{name: "forged token", payload: sub(func(m map[string]any) { m["token"] = "guess" }), status: http.StatusBadRequest, want: "unknown or expired form", code: "form_unknown"},
		{name: "unknown surface", payload: sub(func(m map[string]any) { m["surfaceId"] = "form-nope" }), status: http.StatusBadRequest, want: "unknown or expired form", code: "form_unknown"},
		{name: "other interrupt", interruptID: "ask-made-up", payload: sub(func(map[string]any) {}), status: http.StatusBadRequest, want: "unknown or expired form", code: "form_unknown"},
		{name: "old revision", payload: sub(func(m map[string]any) { m["revision"] = 0 }), status: http.StatusConflict, want: "changed", code: "form_stale"},
		{name: "wrong version", payload: sub(func(m map[string]any) { m["version"] = "v0.8" }), status: http.StatusBadRequest, want: "version"},
		{name: "missing required", payload: sub(func(m map[string]any) { m["values"] = map[string]any{"env": "prod"} }), status: http.StatusUnprocessableEntity, field: "reason", code: "form_invalid"},
		{name: "extra field", payload: sub(func(m map[string]any) { m["values"] = map[string]any{"env": "prod", "reason": "x", "admin": "yes"} }), status: http.StatusUnprocessableEntity, field: "admin"},
		{name: "non-string value", payload: sub(func(m map[string]any) { m["values"] = map[string]any{"env": []any{"prod"}, "reason": "x"} }), status: http.StatusUnprocessableEntity, field: "env"},
		{name: "not an option", payload: sub(func(m map[string]any) { m["values"] = map[string]any{"env": "moon", "reason": "x"} }), status: http.StatusUnprocessableEntity, field: "env"},
		{name: "too long", payload: sub(func(m map[string]any) {
			m["values"] = map[string]any{"env": "prod", "reason": strings.Repeat("r", 21)}
		}), status: http.StatusUnprocessableEntity, field: "reason", code: "form_invalid"},
		{name: "another workspace", payload: sub(func(map[string]any) {}), opts: []a2uiOpt{inWorkspace("ws-b")}, status: http.StatusNotFound},
		// Another user holds no session under this threadId and may not take it.
		{name: "another user", payload: sub(func(map[string]any) {}), opts: []a2uiOpt{asUser("u2")}, status: http.StatusForbidden, want: "threadId is not available"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			interruptID := fx.interruptID
			if tc.interruptID != "" {
				interruptID = tc.interruptID
			}
			w := h.post("approval", resumeBody("t-rej", "run-x", interruptID, tc.payload), tc.opts...)
			if w.Code != tc.status {
				t.Fatalf("status = %d, want %d (body %s)", w.Code, tc.status, w.Body.String())
			}
			if strings.Contains(w.Body.String(), "RUN_STARTED") {
				t.Fatal("a rejected submission opened a stream")
			}
			var body aguiCodedError
			_ = json.Unmarshal(w.Body.Bytes(), &body)
			if tc.want != "" && !strings.Contains(body.Error, tc.want) {
				t.Errorf("error = %q, want it to mention %q", body.Error, tc.want)
			}
			if tc.code != "" && body.Code != tc.code {
				t.Errorf("code = %q, want %q", body.Code, tc.code)
			}
			if tc.field != "" && body.FieldErrors[tc.field] == "" {
				t.Errorf("fieldErrors = %+v, want an error for %q", body.FieldErrors, tc.field)
			}
		})
	}
	if h.backend.CallCount("publisher") != 0 {
		t.Fatal("a rejected submission ran the workflow")
	}
	if got := h.storedAnswers("t-rej"); len(got) != 0 {
		t.Fatalf("a rejected submission was appended: %+v", got)
	}
	if got := h.snapshotSurfaces("approval", "t-rej"); len(got) != 1 {
		t.Fatalf("pending form lost after rejections: %+v", got)
	}

	w := h.post("approval", resumeBody("t-rej", "run-ok", fx.interruptID, formSubmission(fx.form, fx.surfaceID, validDeployValues())))
	if w.Code != http.StatusOK || h.backend.CallCount("publisher") != 1 {
		t.Fatalf("valid submission after rejections: status %d, publisher calls %d", w.Code, h.backend.CallCount("publisher"))
	}
}

// Submitting the same form twice resumes the workflow once; the repeat is
// told the form was already submitted.
func TestAGUIA2UI_FormDuplicateSubmission(t *testing.T) {
	h := newA2UIHarness(t, approvalWorkflow(deployForm()), "drafter", "publisher")
	h.echoModels("drafter", "publisher")
	fx := h.pauseWithForm("t-dup")
	body := resumeBody("t-dup", "run-2", fx.interruptID, formSubmission(fx.form, fx.surfaceID, validDeployValues()))

	if w := h.post("approval", body); w.Code != http.StatusOK {
		t.Fatalf("first submit status = %d, body = %s", w.Code, w.Body.String())
	}
	w := h.post("approval", body)
	if w.Code != http.StatusConflict || !strings.Contains(w.Body.String(), "already submitted") ||
		!strings.Contains(w.Body.String(), `"code":"form_answered"`) {
		t.Fatalf("repeat status = %d, body = %s", w.Code, w.Body.String())
	}
	if h.backend.CallCount("publisher") != 1 || len(h.storedAnswers("t-dup")) != 1 {
		t.Fatalf("publisher calls %d, stored answers %d, want 1 each", h.backend.CallCount("publisher"), len(h.storedAnswers("t-dup")))
	}
}

// While a run holds the thread, a submission is refused with 409 before the
// stream opens and consumes nothing. The snapshot takes no lease, so it still
// answers, with the form open.
func TestAGUIA2UI_FormSubmissionOnBusyThread(t *testing.T) {
	h := newA2UIHarness(t, approvalWorkflow(deployForm()), "drafter", "publisher")
	h.echoModels("drafter", "publisher")
	fx := h.pauseWithForm("t-busy")

	h.guard.busy = true
	w := h.post("approval", resumeBody("t-busy", "run-2", fx.interruptID, formSubmission(fx.form, fx.surfaceID, validDeployValues())))
	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	if got := h.snapshotSurfaces("approval", "t-busy"); len(got) != 1 || got[0]["surfaceId"] != fx.surfaceID {
		t.Errorf("snapshot on a busy thread = %+v, want the open form", got)
	}
	h.guard.busy = false
	if len(h.storedAnswers("t-busy")) != 0 || len(h.snapshotSurfaces("approval", "t-busy")) != 1 {
		t.Fatal("a refused submission consumed the Interrupt")
	}
}

// parallelApprovals is START → ask_a → pub_a and START → ask_b → pub_b: two
// Human Input nodes pause side by side.
func parallelApprovals() []agentsv1.Agent {
	form := func(title string) *agentsv1.HumanInputForm {
		return &agentsv1.HumanInputForm{Title: title, Fields: []*agentsv1.HumanInputFormField{{
			Name: "answer", Label: "Answer", Required: true,
			Type: agentsv1.HumanInputFormFieldType_HUMAN_INPUT_FORM_FIELD_TYPE_TEXT,
		}}}
	}
	return []agentsv1.Agent{{
		Name: "approval", AgentId: "approval", WorkspaceId: "ws-a",
		Type:          agentsv1.AgentType_AGENT_TYPE_WORKFLOW,
		ChildAgentIds: []string{"pub_a", "pub_b"},
		Config: &agentsv1.AgentConfig{Workflow: &agentsv1.WorkflowConfig{
			Nodes: []*agentsv1.WorkflowNode{
				{Name: "ask_a", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_HUMAN_INPUT, Question: "Question A?", Form: form("Form A")},
				{Name: "ask_b", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_HUMAN_INPUT, Question: "Question B?", Form: form("Form B")},
				{Name: "pub_a", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, AgentId: "pub_a"},
				{Name: "pub_b", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, AgentId: "pub_b"},
			},
			Edges: []*agentsv1.WorkflowEdge{
				{From: "START", To: "ask_a"},
				{From: "START", To: "ask_b"},
				{From: "ask_a", To: "pub_a"},
				{From: "ask_b", To: "pub_b"},
			},
		}},
	},
		{Name: "pub_a", AgentId: "pub_a", WorkspaceId: "ws-a", Config: &agentsv1.AgentConfig{Model: "model-a"}},
		{Name: "pub_b", AgentId: "pub_b", WorkspaceId: "ws-a", Config: &agentsv1.AgentConfig{Model: "model-b"}},
	}
}

// Two parallel Human Input nodes each get their own form. Answering B
// resumes B only — A's successor does not run, A's form stays pending and is
// still reported as an open interrupt — and B's token cannot answer A.
func TestAGUIA2UI_ParallelFormsResumeExactly(t *testing.T) {
	h := newA2UIHarness(t, parallelApprovals(), "model-a", "model-b")
	h.echoModels("model-a", "model-b")
	values, finished := h.pausedForm("t-par")
	forms := map[string]formFixture{}
	for _, v := range values {
		if envelopeKind(v) == "createSurface" {
			form := v["form"].(map[string]any)
			forms[form["title"].(string)] = formFixture{interruptID: form["interruptId"].(string), surfaceID: v["surfaceId"].(string), form: form}
		}
	}
	a, b := forms["Form A"], forms["Form B"]
	if a.interruptID == "" || b.interruptID == "" || a.interruptID == b.interruptID {
		t.Fatalf("forms = %+v", forms)
	}
	if n := len(finished["outcome"].(map[string]any)["interrupts"].([]any)); n != 2 {
		t.Fatalf("outcome interrupts = %d, want 2", n)
	}

	// B's binding cannot be aimed at A's Interrupt.
	w := h.post("approval", resumeBody("t-par", "run-x", a.interruptID, formSubmission(b.form, b.surfaceID, map[string]any{"answer": "b"})))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("cross-form submit status = %d, body = %s", w.Code, w.Body.String())
	}

	w = h.post("approval", resumeBody("t-par", "run-2", b.interruptID, formSubmission(b.form, b.surfaceID, map[string]any{"answer": "for b"})))
	if w.Code != http.StatusOK {
		t.Fatalf("submit B status = %d, body = %s", w.Code, w.Body.String())
	}
	if got := h.backend.LastInput("model-b"); got != `{"answer":"for b"}` {
		t.Errorf("pub_b input = %q", got)
	}
	if h.backend.CallCount("model-a") != 0 {
		t.Error("answering B ran A's successor")
	}
	events := sseEvents(t, w.Body.String())
	outcome := events[len(events)-1]["outcome"].(map[string]any)
	open := outcome["interrupts"].([]any)
	if outcome["type"] != "interrupt" || len(open) != 1 || open[0].(map[string]any)["id"] != a.interruptID {
		t.Fatalf("after answering B the outcome = %+v, want A still open", outcome)
	}
	snap := h.snapshotSurfaces("approval", "t-par")
	if len(snap) != 1 || snap[0]["surfaceId"] != a.surfaceID {
		t.Fatalf("snapshot after answering B = %+v", snap)
	}

	w = h.post("approval", resumeBody("t-par", "run-3", a.interruptID, formSubmission(a.form, a.surfaceID, map[string]any{"answer": "for a"})))
	if w.Code != http.StatusOK || h.backend.LastInput("model-a") != `{"answer":"for a"}` {
		t.Fatalf("submit A: status %d, pub_a input %q", w.Code, h.backend.LastInput("model-a"))
	}
}

// A client without A2UI gets no form events for a form node, reads the field
// instructions from the interrupt message, and can still answer with plain
// text — the successor receives it unchanged. A node without a form shows
// an A2UI client no form either.
func TestAGUIA2UI_FormNodeWithoutA2UIAndPlainNode(t *testing.T) {
	t.Run("client without A2UI", func(t *testing.T) {
		h := newA2UIHarness(t, approvalWorkflow(deployForm()), "drafter", "publisher")
		h.echoModels("drafter", "publisher")
		w := h.post("approval", minimalAGUIBody("t-plain-form", "write the release"))
		events := sseEvents(t, w.Body.String())
		if got := a2uiValues(events); len(got) != 0 {
			t.Fatalf("form events sent to a client without A2UI: %+v", got)
		}
		outcome := events[len(events)-1]["outcome"].(map[string]any)
		it := outcome["interrupts"].([]any)[0].(map[string]any)
		if !strings.Contains(it["message"].(string), "env: Environment") {
			t.Errorf("interrupt message lacks instructions: %q", it["message"])
		}
		body := resumeBody("t-plain-form", "run-2", it["id"].(string), "prod please, it is urgent")
		delete(body, "forwardedProps")
		if w := h.post("approval", body); w.Code != http.StatusOK {
			t.Fatalf("plain resume status = %d, body = %s", w.Code, w.Body.String())
		}
		if got := h.backend.LastInput("publisher"); got != "prod please, it is urgent" {
			t.Errorf("publisher input = %q, want the plain answer", got)
		}
	})
	t.Run("node without a form", func(t *testing.T) {
		h := newA2UIHarness(t, approvalWorkflow(nil), "drafter", "publisher")
		h.echoModels("drafter", "publisher")
		values, finished := h.pausedForm("t-noform")
		if len(values) != 0 {
			t.Fatalf("form events for a node without a form: %+v", values)
		}
		it := finished["outcome"].(map[string]any)["interrupts"].([]any)[0].(map[string]any)
		if it["message"] != "Approve this deploy?" {
			t.Errorf("plain node message = %q, want the bare question", it["message"])
		}
	})
}

// cancelled resume entries stay rejected even for a form's Interrupt.
func TestAGUIA2UI_FormCancelRejected(t *testing.T) {
	h := newA2UIHarness(t, approvalWorkflow(deployForm()), "drafter", "publisher")
	h.echoModels("drafter", "publisher")
	fx := h.pauseWithForm("t-cancel")
	body := resumeBody("t-cancel", "run-2", fx.interruptID, nil)
	body["resume"] = []any{map[string]any{"interruptId": fx.interruptID, "status": "cancelled"}}
	if w := h.post("approval", body); w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	if len(h.snapshotSurfaces("approval", "t-cancel")) != 1 {
		t.Fatal("cancel consumed the form")
	}
}

// --- #355: recovery ---

// After a restart the pending form comes back from the persisted session
// with the same binding, and submitting it resumes the workflow on the new
// instance without re-running anything before the Interrupt.
func TestAGUIA2UI_FormSurvivesRestart(t *testing.T) {
	h := newA2UIHarness(t, approvalWorkflow(deployForm()), "drafter", "publisher")
	h.echoModels("drafter", "publisher")
	fx := h.pauseWithForm("t-restart")
	drafts := h.backend.CallCount("drafter")

	h.restart(approvalWorkflow(deployForm()), "drafter", "publisher")
	snap := h.snapshotSurfaces("approval", "t-restart")
	if len(snap) != 1 || snap[0]["kind"] != "form" || snap[0]["surfaceId"] != fx.surfaceID {
		t.Fatalf("snapshot after restart = %+v", snap)
	}
	form := snap[0]["form"].(map[string]any)
	if form["token"] != fx.form["token"] || form["interruptId"] != fx.interruptID {
		t.Fatalf("restored binding = %+v, want %+v", form, fx.form)
	}
	if envs := snap[0]["envelopes"].([]any); len(envs) != 3 {
		t.Fatalf("restored envelopes = %+v", envs)
	}

	w := h.post("approval", resumeBody("t-restart", "run-2", fx.interruptID, formSubmission(form, fx.surfaceID, validDeployValues())))
	if w.Code != http.StatusOK || h.backend.CallCount("publisher") != 1 {
		t.Fatalf("submit after restart: status %d, publisher calls %d", w.Code, h.backend.CallCount("publisher"))
	}
	if h.backend.CallCount("drafter") != drafts {
		t.Error("resuming after a restart re-ran the drafter")
	}
}

// Editing the node's form after it was shown does not change the rules of
// the form already waiting: the frozen binding still validates.
func TestAGUIA2UI_FormRulesAreFrozen(t *testing.T) {
	h := newA2UIHarness(t, approvalWorkflow(deployForm()), "drafter", "publisher")
	h.echoModels("drafter", "publisher")
	fx := h.pauseWithForm("t-frozen")

	edited := deployForm()
	edited.Fields = edited.Fields[:1] // "reason" and "note" removed
	h.restart(approvalWorkflow(edited), "drafter", "publisher")

	w := h.post("approval", resumeBody("t-frozen", "run-2", fx.interruptID, formSubmission(fx.form, fx.surfaceID, map[string]any{"env": "prod"})))
	if w.Code != http.StatusUnprocessableEntity || !strings.Contains(w.Body.String(), "reason") {
		t.Fatalf("status = %d, body = %s; the shown form still requires reason", w.Code, w.Body.String())
	}
}

// A card persisted but never delivered (the client dropped mid-stream) is
// recovered by the snapshot, with no further agent run.
func TestAGUIA2UI_SnapshotRecoversUndeliveredCard(t *testing.T) {
	h := newA2UIHarness(t, []agentsv1.Agent{cardAgent()}, "card-model")
	h.scriptToolThenText("card-model", "render_ui", deployCardArgs(), "Deployed.")
	h.dropCustom = true
	w := h.post("carder", a2uiBody("t-drop", "deploy"))
	if got := a2uiValues(sseEvents(t, w.Body.String())); len(got) != 0 {
		t.Fatalf("setup: CUSTOM frames were delivered: %+v", got)
	}
	h.dropCustom = false
	calls := h.backend.CallCount("card-model")
	snap := h.snapshotSurfaces("carder", "t-drop")
	if len(snap) != 1 || snap[0]["kind"] != "card" || snap[0]["revision"] != float64(1) {
		t.Fatalf("snapshot = %+v", snap)
	}
	if h.backend.CallCount("card-model") != calls {
		t.Error("recovery ran the agent")
	}
}

// A form's answer is a JSON object text. ADK parses a JSON answer when the
// workflow resumes, so a Router after the form must still take it as text:
// it matches no route label here and the default edge fires.
func TestAGUIA2UI_FormAnswerFeedsRouter(t *testing.T) {
	form := &agentsv1.HumanInputForm{Fields: []*agentsv1.HumanInputFormField{{
		Name: "decision", Label: "Decision", Required: true,
		Type: agentsv1.HumanInputFormFieldType_HUMAN_INPUT_FORM_FIELD_TYPE_SINGLE_CHOICE,
		Options: []*agentsv1.HumanInputFormOption{
			{Value: "approve", Label: "Approve"},
			{Value: "reject", Label: "Reject"},
		},
	}}}
	agents := []agentsv1.Agent{{
		Name: "gate", AgentId: "gate", WorkspaceId: "ws-a",
		Type:          agentsv1.AgentType_AGENT_TYPE_WORKFLOW,
		ChildAgentIds: []string{"yes", "fallback"},
		Config: &agentsv1.AgentConfig{Workflow: &agentsv1.WorkflowConfig{
			Nodes: []*agentsv1.WorkflowNode{
				{Name: "ask", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_HUMAN_INPUT, Question: "Ship it?", Form: form},
				{Name: "route", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_ROUTER},
				{Name: "yes", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, AgentId: "yes"},
				{Name: "fallback", Kind: agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT, AgentId: "fallback"},
			},
			Edges: []*agentsv1.WorkflowEdge{
				{From: "START", To: "ask"},
				{From: "ask", To: "route"},
				{From: "route", To: "yes", Route: "approve"},
				{From: "route", To: "fallback", IsDefault: true},
			},
		}},
	},
		{Name: "yes", AgentId: "yes", WorkspaceId: "ws-a", Config: &agentsv1.AgentConfig{Model: "yes-model"}},
		{Name: "fallback", AgentId: "fallback", WorkspaceId: "ws-a", Config: &agentsv1.AgentConfig{Model: "fallback-model"}},
	}
	h := newA2UIHarness(t, agents, "yes-model", "fallback-model")
	h.echoModels("yes-model", "fallback-model")

	w := h.post("gate", a2uiBody("t-gate", "go"))
	values := a2uiValues(sseEvents(t, w.Body.String()))
	if len(values) == 0 {
		t.Fatalf("no form shown:\n%s", w.Body.String())
	}
	fx := values[0]["form"].(map[string]any)
	w = h.post("gate", resumeBody("t-gate", "run-2", fx["interruptId"].(string),
		formSubmission(fx, values[0]["surfaceId"].(string), map[string]any{"decision": "approve"})))
	if w.Code != http.StatusOK {
		t.Fatalf("submit status = %d, body = %s", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), `"type":"RUN_ERROR"`) {
		t.Fatalf("the router failed on the form answer:\n%s", w.Body.String())
	}
	if got := h.backend.LastInput("fallback-model"); got != `{"decision":"approve"}` {
		t.Errorf("default branch input = %q, want the answer as JSON text", got)
	}
}

// finishedOutcome returns the RUN_FINISHED outcome of a stream.
func finishedOutcome(t *testing.T, body string) map[string]any {
	t.Helper()
	events := sseEvents(t, body)
	for i := len(events) - 1; i >= 0; i-- {
		if events[i]["type"] == "RUN_FINISHED" {
			outcome, _ := events[i]["outcome"].(map[string]any)
			return outcome
		}
	}
	t.Fatalf("no RUN_FINISHED:\n%s", body)
	return nil
}

// For an A2UI client, RUN_FINISHED lists every Interrupt still open on the
// thread whenever one remains — also after a plain-text reply answered the
// oldest (ADR-0002), so the client's forms and pending set stay in step. A
// client without A2UI keeps the original contract: only Interrupts raised
// in the run are listed.
func TestAGUIA2UI_OutcomeListsOpenInterruptsForA2UIClients(t *testing.T) {
	t.Run("A2UI client, text reply answers the oldest", func(t *testing.T) {
		h := newA2UIHarness(t, parallelApprovals(), "model-a", "model-b")
		h.echoModels("model-a", "model-b")
		_, finished := h.pausedForm("t-fifo")
		open := finished["outcome"].(map[string]any)["interrupts"].([]any)
		if len(open) != 2 {
			t.Fatalf("setup: open interrupts = %d", len(open))
		}
		w := h.post("approval", a2uiBody("t-fifo", "my answer"))
		outcome := finishedOutcome(t, w.Body.String())
		left, _ := outcome["interrupts"].([]any)
		if outcome["type"] != "interrupt" || len(left) != 1 {
			t.Fatalf("outcome after a text reply = %+v, want the other interrupt still open", outcome)
		}
		if answered := a2uiValues(sseEvents(t, w.Body.String())); len(answered) != 1 || envelopeKind(answered[0]) != "updateDataModel" {
			t.Errorf("the answered form was not marked: %+v", answered)
		}
	})
	t.Run("client without A2UI", func(t *testing.T) {
		h := newA2UIHarness(t, parallelApprovals(), "model-a", "model-b")
		h.echoModels("model-a", "model-b")
		w := h.post("approval", minimalAGUIBody("t-plain-par", "go"))
		it := finishedOutcome(t, w.Body.String())["interrupts"].([]any)[0].(map[string]any)
		body := resumeBody("t-plain-par", "run-2", it["id"].(string), "done")
		delete(body, "forwardedProps")
		w = h.post("approval", body)
		if outcome := finishedOutcome(t, w.Body.String()); outcome["type"] != "success" {
			t.Fatalf("plain client outcome = %+v, want success (no interrupt raised in this run)", outcome)
		}
	})
}

// A form is submittable only from the context that owns its thread: the same
// caller reusing the threadId with another agent, or under another workspace
// whose agent shares the agent_id, is refused the thread itself. Either way
// nothing runs.
func TestAGUIA2UI_FormSubmissionFromAnotherContext(t *testing.T) {
	agents := approvalWorkflow(deployForm())
	agents = append(agents,
		agentsv1.Agent{Name: "other", AgentId: "other", WorkspaceId: "ws-a", Config: &agentsv1.AgentConfig{Model: "publisher"}},
		agentsv1.Agent{Name: "approvalB", AgentId: "approval", WorkspaceId: "ws-b", Config: &agentsv1.AgentConfig{Model: "publisher"}},
	)
	h := newA2UIHarness(t, agents, "drafter", "publisher")
	h.echoModels("drafter", "publisher")
	fx := h.pauseWithForm("t-ctx")
	calls := h.backend.CallCount("publisher")

	for name, tc := range map[string]struct {
		agentID string
		opts    []a2uiOpt
		status  int
		want    string
	}{
		"another agent":     {agentID: "other", status: http.StatusForbidden, want: "threadId belongs to another agent"},
		"another workspace": {agentID: "approval", opts: []a2uiOpt{inWorkspace("ws-b")}, status: http.StatusForbidden, want: "threadId is not available"},
	} {
		t.Run(name, func(t *testing.T) {
			w := h.post(tc.agentID, resumeBody("t-ctx", "run-x", fx.interruptID, formSubmission(fx.form, fx.surfaceID, validDeployValues())), tc.opts...)
			if w.Code != tc.status || !strings.Contains(w.Body.String(), tc.want) {
				t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
			}
		})
	}
	if h.backend.CallCount("publisher") != calls || len(h.storedAnswers("t-ctx")) != 0 {
		t.Fatal("a submission from another context ran or consumed the Interrupt")
	}
	if len(h.snapshotSurfaces("approval", "t-ctx")) != 1 {
		t.Fatal("the owner's form is gone")
	}
}

// --- #440: the Card Policy ---

// cardsOff is a Card Policy that turns Result Cards off for its agent and
// every agent below it in a run.
func cardsOff() *agentsv1.ResultCardConfig {
	return &agentsv1.ResultCardConfig{Generation: agentsv1.ResultCardGeneration_RESULT_CARD_GENERATION_DISABLED}
}

// An LLM agent whose policy turns cards off is not offered render_ui. A
// model that calls it anyway, as one copying an earlier call from the
// thread's history would, gets ADK's "not found" error back, and the run
// goes on to a text reply with no card.
func TestAGUIA2UI_CardPolicyDisabledRootOffersNoRenderUI(t *testing.T) {
	agents := []agentsv1.Agent{cardAgent()}
	agents[0].Config.ResultCards = cardsOff()
	h := newA2UIHarness(t, agents, "card-model")
	h.scriptToolThenText("card-model", "render_ui", deployCardArgs(), "Here it is in text.")

	w := h.post("carder", a2uiBody("t-off", "deploy please"))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	if tools := h.offeredTools("card-model"); containsString(tools, "render_ui") {
		t.Errorf("render_ui offered under a disabled policy: %v", tools)
	}
	if got := a2uiValues(sseEvents(t, w.Body.String())); len(got) != 0 {
		t.Errorf("a card was streamed under a disabled policy: %+v", got)
	}
	if msg, _ := h.lastToolResult("card-model")["error"].(string); !strings.Contains(msg, "not found") {
		t.Errorf("tool result = %q, want ADK's not-found error", msg)
	}
	if !strings.Contains(w.Body.String(), `"delta":"Here it is in text."`) {
		t.Errorf("the run did not go on to its text reply:\n%s", w.Body.String())
	}
	if got := h.snapshotSurfaces("carder", "t-off"); len(got) != 0 {
		t.Errorf("a card was persisted under a disabled policy: %+v", got)
	}
}

// pipeline is a Sequential root over the LLM agents first and second, each on
// its own model, with the given Card Policies.
func pipeline(root, first, second *agentsv1.ResultCardConfig) []agentsv1.Agent {
	return []agentsv1.Agent{
		{
			Name: "pipeline", AgentId: "pipeline", WorkspaceId: "ws-a",
			Type:          agentsv1.AgentType_AGENT_TYPE_SEQUENTIAL,
			ChildAgentIds: []string{"first", "second"},
			Config:        &agentsv1.AgentConfig{ResultCards: root},
		},
		{Name: "first", AgentId: "first", WorkspaceId: "ws-a", Config: &agentsv1.AgentConfig{Model: "first-model", ResultCards: first}},
		{Name: "second", AgentId: "second", WorkspaceId: "ws-a", Config: &agentsv1.AgentConfig{Model: "second-model", ResultCards: second}},
	}
}

// The Card Policy narrows down the agent tree: a disabled agent turns cards
// off for every agent below it and a disabled child only for itself, while
// an agent run directly answers only to its own policy.
func TestAGUIA2UI_CardPolicyNarrowsDownTheTree(t *testing.T) {
	cases := []struct {
		name                string
		root, first, second *agentsv1.ResultCardConfig
		run                 string
		offered             map[string]bool
	}{
		{name: "no policy", run: "pipeline", offered: map[string]bool{"first-model": true, "second-model": true}},
		{name: "composite root disabled", root: cardsOff(), run: "pipeline", offered: map[string]bool{"first-model": false, "second-model": false}},
		{name: "one child disabled", second: cardsOff(), run: "pipeline", offered: map[string]bool{"first-model": true, "second-model": false}},
		{name: "child of a disabled root, run directly", root: cardsOff(), run: "first", offered: map[string]bool{"first-model": true}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := newA2UIHarness(t, pipeline(tc.root, tc.first, tc.second), "first-model", "second-model")
			h.echoModels("first-model", "second-model")
			w := h.post(tc.run, a2uiBody("t-tree", "go"))
			if w.Code != http.StatusOK {
				t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
			}
			for model, want := range tc.offered {
				if got := containsString(h.offeredTools(model), "render_ui"); got != want {
					t.Errorf("%s offered render_ui = %v, want %v", model, got, want)
				}
			}
		})
	}
}

// Forms are outside the Card Policy: a Workflow root that turns cards off
// still opens its Human Input form, and submitting the form resumes the
// workflow. Its node agents, built as its children, get no render_ui.
func TestAGUIA2UI_CardPolicyLeavesFormsAlone(t *testing.T) {
	agents := approvalWorkflow(deployForm())
	agents[0].Config.ResultCards = cardsOff()
	h := newA2UIHarness(t, agents, "drafter", "publisher")
	h.echoModels("drafter", "publisher")

	fx := h.pauseWithForm("t-off-form")
	w := h.post("approval", resumeBody("t-off-form", "run-2", fx.interruptID, formSubmission(fx.form, fx.surfaceID, validDeployValues())))
	if w.Code != http.StatusOK {
		t.Fatalf("submit status = %d, body = %s", w.Code, w.Body.String())
	}
	if outcome := finishedOutcome(t, w.Body.String()); outcome["type"] != "success" {
		t.Fatalf("outcome = %+v, want the resumed workflow to finish", outcome)
	}
	if h.backend.CallCount("publisher") == 0 || len(h.storedAnswers("t-off-form")) != 1 {
		t.Fatal("the form's submission did not resume the workflow")
	}
	for _, model := range []string{"drafter", "publisher"} {
		if containsString(h.offeredTools(model), "render_ui") {
			t.Errorf("%s was offered render_ui under a disabled Workflow root", model)
		}
	}
}

// Turning cards off for an agent stops it changing cards; it does not hide
// the cards its threads already hold.
func TestAGUIA2UI_CardPolicyKeepsExistingCards(t *testing.T) {
	h := newA2UIHarness(t, []agentsv1.Agent{cardAgent()}, "card-model")
	h.scriptToolThenText("card-model", "render_ui", deployCardArgs(), "Deployed.")
	if got := a2uiValues(sseEvents(t, h.post("carder", a2uiBody("t-kept", "deploy")).Body.String())); len(got) != 3 {
		t.Fatalf("setup: card not rendered: %+v", got)
	}

	// The agent is switched to DISABLED; the runner rebuilds its trees.
	off := []agentsv1.Agent{cardAgent()}
	off[0].Config.ResultCards = cardsOff()
	h.restart(off, "card-model")
	h.backend.ScriptRequest("card-model", func(w http.ResponseWriter, req openaifake.ChatCompletionRequest) {
		openaifake.WriteReply(w, req, "noted")
	})
	w := h.post("carder", updateBody("t-kept", "run-2", "anything new?"))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	if containsString(h.offeredTools("card-model"), "render_ui") {
		t.Error("render_ui offered after the agent turned cards off")
	}
	if got := h.snapshotSurfaces("carder", "t-kept"); len(got) != 1 {
		t.Errorf("snapshot after turning cards off = %+v, want the thread's card", got)
	}
}

// toolDescription returns the description of the tool name that model was
// offered on its last call.
func (h *a2uiHarness) toolDescription(model, name string) string {
	h.t.Helper()
	req, ok := h.backend.LastRequest(model)
	if !ok {
		h.t.Fatalf("model %s was never called", model)
	}
	tools, _ := req.Decoded["tools"].([]any)
	for _, t := range tools {
		fn, _ := t.(map[string]any)["function"].(map[string]any)
		if fn["name"] == name {
			desc, _ := fn["description"].(string)
			return desc
		}
	}
	h.t.Fatalf("model %s was not offered %s", model, name)
	return ""
}

// --- #442: the presentation hint ---

// A composite root's PREFERRED reaches the render_ui description of the LLM
// agents below it, while a child that sets AUTO explicitly is not asked.
func TestAGUIA2UI_CardPolicyPreferredReachesTheTree(t *testing.T) {
	preferred := &agentsv1.ResultCardConfig{Presentation: agentsv1.ResultCardPresentation_RESULT_CARD_PRESENTATION_PREFERRED}
	auto := &agentsv1.ResultCardConfig{Presentation: agentsv1.ResultCardPresentation_RESULT_CARD_PRESENTATION_AUTO}
	h := newA2UIHarness(t, pipeline(preferred, nil, auto), "first-model", "second-model")
	h.echoModels("first-model", "second-model")

	w := h.post("pipeline", a2uiBody("t-prefer", "go"))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	if desc := h.toolDescription("first-model", "render_ui"); !strings.Contains(desc, a2uitool.PreferredHint) {
		t.Errorf("first, below a PREFERRED root, was not asked to prefer cards:\n%s", desc)
	}
	if desc := h.toolDescription("second-model", "render_ui"); strings.Contains(desc, a2uitool.PreferredHint) {
		t.Errorf("second, explicitly AUTO, was asked to prefer cards:\n%s", desc)
	}
}
