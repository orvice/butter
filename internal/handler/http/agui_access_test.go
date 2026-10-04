package http

import (
	"context"
	"net/http"
	"reflect"
	"strings"
	"testing"

	adksession "google.golang.org/adk/v2/session"

	"go.orx.me/apps/butter/internal/a2ui"
	"go.orx.me/apps/butter/internal/testsupport/openaifake"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Who reaches which agent on the AG-UI endpoint, and which agent a thread
// runs with (#394). A signed-in user reaches every agent the runner can run;
// enable_agui gates only API and root tokens. A thread bound to one agent is
// refused to every other one before anything runs.

func plainAgent(id, model string) agentsv1.Agent {
	return agentsv1.Agent{
		Name: "Agent-" + id, AgentId: id, WorkspaceId: "ws-a",
		Config: &agentsv1.AgentConfig{Model: model},
	}
}

func (h *a2uiHarness) answer(model, text string) {
	h.backend.ScriptRequest(model, func(w http.ResponseWriter, req openaifake.ChatCompletionRequest) {
		openaifake.WriteReply(w, req, text)
	})
}

// threadSession reads the thread's session as user u1 holds it.
func (h *a2uiHarness) threadSession(threadID string) (adksession.Session, bool) {
	return h.sessionOf("u1", threadID)
}

// sessionOf reads the thread's session as the given session user holds it;
// token callers have none of their own and run as agui-user.
func (h *a2uiHarness) sessionOf(userID, threadID string) (adksession.Session, bool) {
	h.t.Helper()
	resp, err := h.sessions.Get(context.Background(), &adksession.GetRequest{AppName: aguiAppName, UserID: userID, SessionID: aguiSessionPrefix + threadID})
	if err != nil {
		return nil, false
	}
	return resp.Session, true
}

func streamed(body string) bool {
	return strings.Contains(body, `"type":"RUN_STARTED"`)
}

// A signed-in user runs an agent that never enabled AG-UI, and reads its
// thread back: the flag does not gate the dashboard.
func TestAGUIAccess_SignedInUserReachesAgentWithoutEnableAGUI(t *testing.T) {
	h := newA2UIHarness(t, []agentsv1.Agent{plainAgent("plain", "plain-model")}, "plain-model")
	h.answer("plain-model", "plain answer")

	w := h.post("plain", minimalAGUIBody("t-plain", "hi"))
	if w.Code != http.StatusOK || !streamed(w.Body.String()) ||
		!strings.Contains(w.Body.String(), `"delta":"plain answer"`) ||
		!strings.Contains(w.Body.String(), `"type":"RUN_FINISHED"`) {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	if code, history := h.history("plain", "t-plain"); code != http.StatusOK || len(history.Messages) != 2 {
		t.Errorf("history: status %d, messages %+v", code, history.Messages)
	}
	if code, _ := h.snapshot("plain", "t-plain"); code != http.StatusOK {
		t.Errorf("UI snapshot: status %d", code)
	}
}

// API tokens and the root token still need enable_agui: without it the agent
// is not found, on runs and reads alike, and nothing runs.
func TestAGUIAccess_TokensNeedEnableAGUI(t *testing.T) {
	agents := []agentsv1.Agent{plainAgent("plain", "plain-model"), plainAgent("exposed", "exposed-model")}
	agents[1].EnableAgui = true
	h := newA2UIHarness(t, agents, "plain-model", "exposed-model")
	h.answer("plain-model", "should not run")
	h.answer("exposed-model", "exposed answer")

	for name, tc := range map[string]struct {
		token  a2uiOpt
		thread string
	}{
		"API token":  {token: asAPIToken(), thread: "t-api"},
		"root token": {token: asRootToken(), thread: "t-root"},
	} {
		t.Run(name, func(t *testing.T) {
			w := h.post("plain", minimalAGUIBody(tc.thread, "hi"), tc.token)
			if w.Code != http.StatusNotFound || !strings.Contains(w.Body.String(), "agent not found: plain") || streamed(w.Body.String()) {
				t.Fatalf("plain agent: status = %d, body = %s; want 404 before any stream", w.Code, w.Body.String())
			}
			if code, _ := h.history("plain", tc.thread, tc.token); code != http.StatusNotFound {
				t.Errorf("plain agent history: status %d, want 404", code)
			}
			if code, _ := h.snapshot("plain", tc.thread, tc.token); code != http.StatusNotFound {
				t.Errorf("plain agent UI snapshot: status %d, want 404", code)
			}

			w = h.post("exposed", minimalAGUIBody(tc.thread, "hi"), tc.token)
			if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"delta":"exposed answer"`) {
				t.Fatalf("exposed agent: status = %d, body = %s", w.Code, w.Body.String())
			}
		})
	}
	if calls := h.backend.CallCount("plain-model"); calls != 0 {
		t.Errorf("an agent without enable_agui ran %d times for a token", calls)
	}
}

// An agent the runner cannot run, here a deleted one the repository still
// returns, is refused before the stream opens, whoever asks and whatever its
// enable_agui. The thread gets no session.
func TestAGUIAccess_UnrunnableAgentIsRefusedBeforeTheStream(t *testing.T) {
	agents := []agentsv1.Agent{plainAgent("gone", "gone-model"), plainAgent("live", "live-model")}
	agents[0].EnableAgui = true
	agents[0].LifecycleStatus = agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_DELETED
	h := newA2UIHarness(t, agents, "gone-model", "live-model")

	for name, opts := range map[string][]a2uiOpt{"signed-in user": nil, "API token": {asAPIToken()}} {
		t.Run(name, func(t *testing.T) {
			w := h.post("gone", minimalAGUIBody("t-gone", "hi"), opts...)
			if w.Code != http.StatusNotFound || !strings.Contains(w.Body.String(), "agent not found: gone") || streamed(w.Body.String()) {
				t.Fatalf("status = %d, body = %s; want 404 before any stream", w.Code, w.Body.String())
			}
		})
	}
	for _, user := range []string{"u1", "agui-user"} {
		if _, ok := h.sessionOf(user, "t-gone"); ok {
			t.Errorf("a refused run created %s's session for the thread", user)
		}
	}
	if calls := h.backend.CallCount("gone-model"); calls != 0 {
		t.Errorf("the deleted agent's model was called %d times", calls)
	}
}

// A thread first run with one agent is refused to another before the stream
// opens: nothing is appended and its history reads back unchanged. It still
// runs with its own agent.
func TestAGUIAccess_ThreadStaysWithItsAgent(t *testing.T) {
	h := newA2UIHarness(t, []agentsv1.Agent{plainAgent("first", "first-model"), plainAgent("second", "second-model")},
		"first-model", "second-model")
	h.answer("first-model", "first answer")
	h.answer("second-model", "second answer")

	if w := h.post("first", a2uiBody("t-1", "hello")); w.Code != http.StatusOK || !streamed(w.Body.String()) {
		t.Fatalf("setup run: status = %d, body = %s", w.Code, w.Body.String())
	}
	before, ok := h.threadSession("t-1")
	if !ok {
		t.Fatal("setup run left no session")
	}
	eventsBefore := before.Events().Len()
	_, historyBefore := h.history("first", "t-1")
	if len(historyBefore.Messages) != 2 {
		t.Fatalf("setup history = %+v", historyBefore.Messages)
	}

	for name, body := range map[string]map[string]any{
		"text turn":   a2uiBody("t-1", "take over"),
		"resume turn": resumeBody("t-1", "run-2", "int-1", "yes"),
	} {
		t.Run(name, func(t *testing.T) {
			w := h.post("second", body)
			if w.Code != http.StatusForbidden || streamed(w.Body.String()) {
				t.Fatalf("status = %d, body = %s; want 403 before any stream", w.Code, w.Body.String())
			}
			if !strings.Contains(w.Body.String(), "threadId belongs to another agent; start a new thread") {
				t.Errorf("error body = %s", w.Body.String())
			}
		})
	}
	if calls := h.backend.CallCount("second-model"); calls != 0 {
		t.Errorf("the other agent ran %d times on the thread", calls)
	}
	after, _ := h.threadSession("t-1")
	if got := after.Events().Len(); got != eventsBefore {
		t.Errorf("session events = %d, want %d: a refused run appended to the thread", got, eventsBefore)
	}
	if binding, ok := a2ui.BindingOf(after.State()); !ok || binding.AgentID != "first" {
		t.Errorf("binding = %+v, %v; want the thread still bound to agent first", binding, ok)
	}
	if _, historyAfter := h.history("first", "t-1"); !reflect.DeepEqual(historyAfter, historyBefore) {
		t.Errorf("history changed:\nbefore %+v\nafter  %+v", historyBefore, historyAfter)
	}

	if w := h.post("first", a2uiBody("t-1", "still here?")); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"delta":"first answer"`) {
		t.Fatalf("own agent: status = %d, body = %s", w.Code, w.Body.String())
	}
}

// An agent_id names an agent only within its workspace. When the store does
// not report a session's workspace, the binding's still keeps the thread from
// running with the same agent_id in another workspace.
func TestAGUIAccess_BindingFromAnotherWorkspaceIsRefused(t *testing.T) {
	h := newA2UIHarness(t, []agentsv1.Agent{plainAgent("first", "first-model")}, "first-model")
	h.answer("first-model", "should not run")
	foreign := a2ui.Binding{Principal: "u1", WorkspaceID: "ws-b", AgentID: "first", ThreadID: "t-foreign"}
	// Created without a workspace in context, the seeded session reports none.
	if _, err := h.sessions.Create(context.Background(), &adksession.CreateRequest{
		AppName: aguiAppName, UserID: "u1", SessionID: aguiSessionPrefix + "t-foreign",
		State: map[string]any{a2ui.BindingKey: foreign.StateValue()},
	}); err != nil {
		t.Fatalf("seed session: %v", err)
	}

	w := h.post("first", minimalAGUIBody("t-foreign", "hi"))
	if w.Code != http.StatusForbidden || streamed(w.Body.String()) ||
		!strings.Contains(w.Body.String(), errThreadUnavailable.Error()) {
		t.Fatalf("status = %d, body = %s; want 403 before any stream", w.Code, w.Body.String())
	}
	if calls := h.backend.CallCount("first-model"); calls != 0 {
		t.Errorf("the agent ran %d times on another workspace's thread", calls)
	}
}

// A thread created before A2UI has no binding, so it names no agent: it runs
// with any agent, as before.
func TestAGUIAccess_UnboundThreadRunsWithAnyAgent(t *testing.T) {
	h := newA2UIHarness(t, []agentsv1.Agent{plainAgent("first", "first-model"), plainAgent("second", "second-model")},
		"first-model", "second-model")
	h.answer("first-model", "first answer")
	h.answer("second-model", "second answer")
	if _, err := h.sessions.Create(context.Background(), &adksession.CreateRequest{
		AppName: aguiAppName, UserID: "u1", SessionID: aguiSessionPrefix + "t-old",
	}); err != nil {
		t.Fatalf("seed session: %v", err)
	}

	for agentID, want := range map[string]string{"first": "first answer", "second": "second answer"} {
		w := h.post(agentID, a2uiBody("t-old", "hi"))
		if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"delta":"`+want+`"`) {
			t.Fatalf("%s: status = %d, body = %s", agentID, w.Code, w.Body.String())
		}
	}
	sess, _ := h.threadSession("t-old")
	if _, bound := a2ui.BindingOf(sess.State()); bound {
		t.Error("running an unbound thread bound it")
	}
}
