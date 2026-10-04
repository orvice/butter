package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	adksession "google.golang.org/adk/v2/session"
	"google.golang.org/adk/v2/workflow"
	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/a2ui"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// historySession stores events in an in-memory session, which applies their
// state deltas as the real store does, and reads it back.
func historySession(t *testing.T, events ...*adksession.Event) adksession.Session {
	t.Helper()
	svc := adksession.InMemoryService()
	get := &adksession.GetRequest{AppName: "agui", UserID: "u1", SessionID: "agui-t1"}
	created, err := svc.Create(t.Context(), &adksession.CreateRequest{AppName: get.AppName, UserID: get.UserID, SessionID: get.SessionID})
	if err != nil {
		t.Fatal(err)
	}
	for _, ev := range events {
		if err := svc.AppendEvent(t.Context(), created.Session, ev); err != nil {
			t.Fatal(err)
		}
	}
	resp, err := svc.Get(t.Context(), get)
	if err != nil {
		t.Fatal(err)
	}
	return resp.Session
}

func userEvent(parts ...*genai.Part) *adksession.Event {
	ev := adksession.NewEvent(context.Background(), "inv-1")
	ev.Author = "user"
	ev.Content = &genai.Content{Role: genai.RoleUser, Parts: parts}
	return ev
}

func agentEvent(parts ...*genai.Part) *adksession.Event {
	ev := adksession.NewEvent(context.Background(), "inv-1")
	ev.Author = "carder"
	ev.Content = &genai.Content{Role: genai.RoleModel, Parts: parts}
	return ev
}

func callPart(id, name string, args map[string]any) *genai.Part {
	return &genai.Part{FunctionCall: &genai.FunctionCall{ID: id, Name: name, Args: args}}
}

func responsePart(id, name string, response map[string]any) *genai.Part {
	return &genai.Part{FunctionResponse: &genai.FunctionResponse{ID: id, Name: name, Response: response}}
}

func askPart(id, question string) *genai.Part {
	return callPart(id, workflow.WorkflowInputFunctionCallName, map[string]any{"message": question})
}

func answerPart(id string, payload any) *genai.Part {
	return responsePart(id, workflow.WorkflowInputFunctionCallName, map[string]any{aguiRequestInputPayloadKey: payload})
}

// newCard is a valid read-only card, as render_ui would store it.
func newCard(t *testing.T) *a2ui.Card {
	t.Helper()
	var messages []map[string]any
	raw := `[{"updateComponents": {"components": [
		{"id": "root", "component": "Card", "child": "t"},
		{"id": "t", "component": "Text", "text": "Healthy"}]}}]`
	if err := json.Unmarshal([]byte(raw), &messages); err != nil {
		t.Fatal(err)
	}
	res, err := a2ui.Apply(map[string]*a2ui.Card{}, a2ui.Batch{Messages: messages, Fallback: "Healthy"}, a2ui.Anchor{ThreadID: "t1"}, a2ui.DefaultLimits)
	if err != nil {
		t.Fatal(err)
	}
	return res.Card
}

// requireJSON compares got and want by their JSON, which is what a client
// receives.
func requireJSON(t *testing.T, label string, got, want any) {
	t.Helper()
	g, _ := json.Marshal(got)
	w, _ := json.Marshal(want)
	var gv, wv any
	_ = json.Unmarshal(g, &gv)
	_ = json.Unmarshal(w, &wv)
	if !reflect.DeepEqual(gv, wv) {
		t.Fatalf("%s:\n got: %s\nwant: %s", label, g, w)
	}
}

// Each run reads as one assistant message — its tool calls with their
// results, and its text — between the user turns, as the live stream showed
// it. Thoughts and the render_ui call stay hidden; the card is placed on the
// message of the run that created it.
func TestAGUIHistory_RebuildsRunsAsTheStreamShowedThem(t *testing.T) {
	card := newCard(t)
	ask := userEvent(&genai.Part{Text: "Summarize the deploy"})
	lookup := agentEvent(callPart("c1", "lookup", map[string]any{"svc": "api"}))
	looked := agentEvent(responsePart("c1", "lookup", map[string]any{"status": "ok"}))
	render := agentEvent(callPart("c2", "render_ui", map[string]any{"messages": []any{}}))
	rendered := agentEvent(responsePart("c2", "render_ui", map[string]any{"surface_id": card.ID}))
	rendered.Actions.StateDelta = map[string]any{a2ui.CardKey(card.ID): card.StateValue()}
	answer := agentEvent(&genai.Part{Text: "thinking it over", Thought: true}, &genai.Part{Text: "The deploy is healthy."})
	thanks := userEvent(&genai.Part{Text: "Thanks"})
	welcome := agentEvent(&genai.Part{Text: "You're welcome."})

	got := aguiHistory("t1", historySession(t, ask, lookup, looked, render, rendered, answer, thanks, welcome))

	requireJSON(t, "messages", got.Messages, []map[string]any{
		{"id": ask.ID, "role": "user", "content": "Summarize the deploy"},
		{"id": lookup.ID, "role": "assistant", "content": "The deploy is healthy.", "toolCalls": []map[string]any{
			{"id": "c1", "type": "function", "function": map[string]any{"name": "lookup", "arguments": `{"svc":"api"}`}},
		}},
		{"id": "result:c1", "role": "tool", "content": `{"status":"ok"}`, "toolCallId": "c1"},
		{"id": thanks.ID, "role": "user", "content": "Thanks"},
		{"id": welcome.ID, "role": "assistant", "content": "You're welcome."},
	})
	requireJSON(t, "surfaces", got.Surfaces, []aguiHistorySurface{{SurfaceID: card.ID, MessageID: lookup.ID}})
	requireJSON(t, "interrupts", got.Interrupts, []any{})
}

// An answered Human Input reads as a question and its answer, a form's answer
// as the dashboard showed the submission. An open one is reported as
// RUN_FINISHED reported it, and its form is placed on the run that raised it.
func TestAGUIHistory_HumanInputReadsAsQuestionAndAnswer(t *testing.T) {
	approval := a2ui.NewForm("ask-1", "Approve the deploy?", deployForm())
	region := a2ui.NewForm("ask-2", "Which region?", &agentsv1.HumanInputForm{
		Title: "Region",
		Fields: []*agentsv1.HumanInputFormField{{
			Name: "region", Label: "Region",
			Type: agentsv1.HumanInputFormFieldType_HUMAN_INPUT_FORM_FIELD_TYPE_TEXT,
		}},
	})

	start := userEvent(&genai.Part{Text: "Deploy please"})
	paused := agentEvent(askPart("ask-1", "Approve the deploy?\n\nFill in: Environment, Reason"))
	approval.Attach(paused)
	answered := userEvent(answerPart("ask-1", `{"env":"prod","reason":"hotfix","note":""}`))
	deploying := agentEvent(&genai.Part{Text: "Deploying."})
	pausedAgain := agentEvent(askPart("ask-2", "Which region?\n\nFill in: Region"))
	region.Attach(pausedAgain)

	got := aguiHistory("t1", historySession(t, start, paused, answered, deploying, pausedAgain))

	requireJSON(t, "messages", got.Messages, []map[string]any{
		{"id": start.ID, "role": "user", "content": "Deploy please"},
		{"id": paused.ID, "role": "assistant", "content": "Approve the deploy?"},
		{"id": answered.ID, "role": "user", "content": approval.ReadableAnswer(`{"env":"prod","reason":"hotfix","note":""}`)},
		{"id": deploying.ID, "role": "assistant", "content": "Deploying."},
	})
	if want := "Deploy approval\nEnvironment: Production\nReason: hotfix"; !strings.HasPrefix(got.Messages[2].Content.(string), want) {
		t.Errorf("form answer reads %q, want it to start with %q", got.Messages[2].Content, want)
	}
	requireJSON(t, "interrupts", got.Interrupts, []map[string]any{
		{"id": "ask-2", "reason": aguiInterruptReason, "message": "Which region?\n\nFill in: Region"},
	})
	requireJSON(t, "surfaces", got.Surfaces, []aguiHistorySurface{{SurfaceID: region.SurfaceID, MessageID: deploying.ID}})
}

// A typed answer to a Human Input, including an implicit resume, is the
// user's text.
func TestAGUIHistory_TypedAnswerIsTheUsersText(t *testing.T) {
	paused := agentEvent(askPart("ask-1", "Which region?"))
	answered := userEvent(answerPart("ask-1", "eu-west-1"))

	got := aguiHistory("t1", historySession(t, paused, answered))

	requireJSON(t, "messages", got.Messages, []map[string]any{
		{"id": paused.ID, "role": "assistant", "content": "Which region?"},
		{"id": answered.ID, "role": "user", "content": "eu-west-1"},
	})
}

// A user turn that carried images reads back as AG-UI content parts, its text
// and its images inline in the order sent, so an image-only turn is kept. A
// turn of only text stays a string.
func TestAGUIHistory_UserImagesComeBackInline(t *testing.T) {
	jpeg := []byte{0xff, 0xd8, 0xff}
	ask := userEvent(&genai.Part{Text: "What is this?"}, genai.NewPartFromBytes(testPNG, "image/png"))
	cat := agentEvent(&genai.Part{Text: "A cat."})
	look := userEvent(genai.NewPartFromBytes(jpeg, "image/jpeg"))
	dog := agentEvent(&genai.Part{Text: "A dog."})
	thanks := userEvent(&genai.Part{Text: "Thanks"})

	got := aguiHistory("t1", historySession(t, ask, cat, look, dog, thanks))

	requireJSON(t, "messages", got.Messages, []map[string]any{
		{"id": ask.ID, "role": "user", "content": []map[string]any{
			textContent("What is this?"), imageContent("image/png", testPNG),
		}},
		{"id": cat.ID, "role": "assistant", "content": "A cat."},
		{"id": look.ID, "role": "user", "content": []map[string]any{imageContent("image/jpeg", jpeg)}},
		{"id": dog.ID, "role": "assistant", "content": "A dog."},
		{"id": thanks.ID, "role": "user", "content": "Thanks"},
	})
}

// A typed answer that carried an image, as an implicit resume stores it,
// keeps both: the answer's text, then the image.
func TestAGUIHistory_AnswerWithAnImageKeepsBoth(t *testing.T) {
	paused := agentEvent(askPart("ask-1", "Which photo?"))
	answered := userEvent(answerPart("ask-1", "this one"), genai.NewPartFromBytes(testPNG, "image/png"))

	got := aguiHistory("t1", historySession(t, paused, answered))

	requireJSON(t, "messages", got.Messages, []map[string]any{
		{"id": paused.ID, "role": "assistant", "content": "Which photo?"},
		{"id": answered.ID, "role": "user", "content": []map[string]any{
			textContent("this one"), imageContent("image/png", testPNG),
		}},
	})
}

// A tool call stays only with its result, or while the session still awaits
// one from the client. A call nobody will answer is dropped: restoring it
// would make a client cancel it, and the server rejects a result for a call
// it does not consider pending.
func TestAGUIHistory_KeepsOnlyCallsTheServerAccepts(t *testing.T) {
	ask := userEvent(&genai.Part{Text: "Go"})
	calls := agentEvent(
		callPart("client-1", "confirm", map[string]any{"q": "ok?"}),
		callPart("lost-1", "lookup", nil),
	)
	calls.LongRunningToolIDs = []string{"client-1"}

	got := aguiHistory("t1", historySession(t, ask, calls))
	requireJSON(t, "pending client call", got.Messages, []map[string]any{
		{"id": ask.ID, "role": "user", "content": "Go"},
		{"id": calls.ID, "role": "assistant", "toolCalls": []map[string]any{
			{"id": "client-1", "type": "function", "function": map[string]any{"name": "confirm", "arguments": `{"q":"ok?"}`}},
		}},
	})

	result := userEvent(responsePart("client-1", "confirm", map[string]any{"result": "yes"}))
	done := agentEvent(&genai.Part{Text: "Confirmed."})
	got = aguiHistory("t1", historySession(t, ask, calls, result, done))
	requireJSON(t, "answered client call", got.Messages, []map[string]any{
		{"id": ask.ID, "role": "user", "content": "Go"},
		{"id": calls.ID, "role": "assistant", "toolCalls": []map[string]any{
			{"id": "client-1", "type": "function", "function": map[string]any{"name": "confirm", "arguments": `{"q":"ok?"}`}},
		}},
		{"id": "result:client-1", "role": "tool", "content": `{"result":"yes"}`, "toolCallId": "client-1"},
		{"id": done.ID, "role": "assistant", "content": "Confirmed."},
	})
}

// Without text, a workflow's Output stands in, as it does for the live turn.
func TestAGUIHistory_WorkflowOutputStandsInForText(t *testing.T) {
	ask := userEvent(&genai.Part{Text: "Classify"})
	node := adksession.NewEvent(context.Background(), "inv-1")
	node.Author = "classifier"
	node.Output = map[string]any{"label": "urgent"}

	got := aguiHistory("t1", historySession(t, ask, node))
	requireJSON(t, "messages", got.Messages, []map[string]any{
		{"id": ask.ID, "role": "user", "content": "Classify"},
		{"id": node.ID, "role": "assistant", "content": `{"label":"urgent"}`},
	})
}

func TestAGUIHistory_NoSessionIsEmpty(t *testing.T) {
	requireJSON(t, "empty", aguiHistory("t1", nil), map[string]any{
		"threadId": "t1", "messages": []any{}, "interrupts": []any{}, "surfaces": []any{},
	})
}

func (h *a2uiHarness) history(agentID, threadID string, opts ...a2uiOpt) (int, aguiThreadHistory) {
	h.t.Helper()
	var ro a2uiRequest
	for _, o := range opts {
		o(&ro)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/agui/"+agentID+"/threads/"+threadID+"/messages", nil)
	ro.apply(req)
	w := httptest.NewRecorder()
	h.router.ServeHTTP(w, req)
	var body aguiThreadHistory
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	return w.Code, body
}

// The endpoint reads the thread back without running the agent, places the
// card on its reply, and reveals nothing to another caller or agent.
func TestAGUIHistory_EndpointRestoresTheThread(t *testing.T) {
	h := newA2UIHarness(t, []agentsv1.Agent{
		cardAgent(),
		{Name: "Other", AgentId: "other", WorkspaceId: "ws-a", Config: &agentsv1.AgentConfig{Model: "card-model"}},
	}, "card-model")
	h.scriptToolThenText("card-model", "render_ui", deployCardArgs(), "Deployed.")
	if w := h.post("carder", a2uiBody("t-hist", "deploy")); w.Code != http.StatusOK {
		t.Fatalf("run status = %d: %s", w.Code, w.Body.String())
	}
	calls := h.backend.CallCount("card-model")

	code, got := h.history("carder", "t-hist")
	if code != http.StatusOK {
		t.Fatalf("status = %d", code)
	}
	if len(got.Messages) != 2 || got.Messages[0].Role != "user" || got.Messages[0].Content != "deploy" ||
		got.Messages[1].Role != "assistant" || got.Messages[1].Content != "Deployed." || len(got.Messages[1].ToolCalls) != 0 {
		t.Fatalf("messages = %+v", got.Messages)
	}
	surfaces := h.snapshotSurfaces("carder", "t-hist")
	if len(surfaces) != 1 || len(got.Surfaces) != 1 ||
		got.Surfaces[0].SurfaceID != surfaces[0]["surfaceId"] || got.Surfaces[0].MessageID != got.Messages[1].ID {
		t.Fatalf("surfaces = %+v, snapshot = %+v", got.Surfaces, surfaces)
	}
	if h.backend.CallCount("card-model") != calls {
		t.Error("reading the history ran the agent")
	}

	for name, read := range map[string]func() (int, aguiThreadHistory){
		"another caller": func() (int, aguiThreadHistory) { return h.history("carder", "t-hist", asUser("intruder")) },
		"another agent":  func() (int, aguiThreadHistory) { return h.history("other", "t-hist") },
		"unknown thread": func() (int, aguiThreadHistory) { return h.history("carder", "t-none") },
	} {
		if code, body := read(); code != http.StatusOK || len(body.Messages) != 0 || len(body.Surfaces) != 0 {
			t.Errorf("%s: status %d, body %+v; want an empty history", name, code, body)
		}
	}

	// Reads take no lease (ADR-0016 decision 6): a thread whose lease is held
	// reads the same, and no read acquired it.
	h.guard.busy = true
	if code, again := h.history("carder", "t-hist"); code != http.StatusOK || !reflect.DeepEqual(again, got) {
		t.Errorf("held lease: status %d, history %+v; want the same history", code, again)
	}
	if acquired := h.guard.acquisitions(); acquired != 1 {
		t.Errorf("lease acquisitions = %d, want only the run's", acquired)
	}
}

// acquisitions counts the guard's Acquire calls, held or not.
func (g *fakeSessionGuard) acquisitions() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return len(g.keys)
}
