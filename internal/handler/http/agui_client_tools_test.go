package http

import (
	"net/http"
	"testing"

	"go.orx.me/apps/butter/internal/testsupport/openaifake"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// A run that declares client tools goes through the real runner's request
// preprocessing, which refuses any tool that cannot pack itself into the model
// request: the model is offered the client tool, calls it, and the run ends
// with the call pending for the client to answer.
func TestAGUIClientTools_RunThroughTheRealRunner(t *testing.T) {
	h := newA2UIHarness(t, []agentsv1.Agent{cardAgent()}, "card-model")
	h.backend.ScriptRequest("card-model", func(w http.ResponseWriter, req openaifake.ChatCompletionRequest) {
		openaifake.WriteReply(w, req, "", openaifake.ToolCall{ID: "call-1", Name: "confirm", Arguments: `{"question":"Deploy now?"}`})
	})

	body := a2uiBody("t-client-tools", "please confirm")
	body["tools"] = []map[string]any{{
		"name":        "confirm",
		"description": "Ask the user to confirm",
		"parameters": map[string]any{
			"type":       "object",
			"properties": map[string]any{"question": map[string]any{"type": "string"}},
		},
	}}
	w := h.post("carder", body)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}

	if !containsString(h.offeredTools("card-model"), "confirm") {
		t.Fatalf("offered tools = %v, want the client tool", h.offeredTools("card-model"))
	}
	var started, errored bool
	for _, evt := range sseEvents(t, w.Body.String()) {
		switch evt["type"] {
		case "TOOL_CALL_START":
			started = started || evt["toolCallName"] == "confirm"
		case "RUN_ERROR":
			errored = true
			t.Errorf("run failed: %v", evt["message"])
		}
	}
	if !started {
		t.Fatalf("no TOOL_CALL_START for the client tool:\n%s", w.Body.String())
	}
	if errored {
		return
	}
	if outcome := finishedOutcome(t, w.Body.String()); outcome["type"] != "success" {
		t.Fatalf("outcome = %+v, want success with the call pending", outcome)
	}
}
