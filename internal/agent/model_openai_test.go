package agent

import (
	"encoding/json/v2"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"google.golang.org/adk/v2/model"
	"google.golang.org/genai"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Exercise Butter's real provider resolution, not just the ADK constructor.
// The first call asks for a tool, the second receives its result. Both calls
// must preserve an earlier image/file and assistant turn in the conversation.
func TestOpenAIProviderProtocolRoundTrip(t *testing.T) {
	for _, api := range []string{"", "chat_completions", "responses"} {
		for _, stream := range []bool{false, true} {
			for _, mime := range []string{"image/png", "application/pdf"} {
				t.Run(fmt.Sprintf("%s/stream=%t/%s", api, stream, mime), func(t *testing.T) {
					path := "/v1/chat/completions"
					if api == "responses" {
						path = "/v1/responses"
					}
					requests := make(chan []byte, 4)
					var calls atomic.Int32
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						if r.URL.Path != path || r.Method != http.MethodPost {
							t.Errorf("request = %s %s, want POST %s", r.Method, r.URL.Path, path)
							http.Error(w, "wrong API", http.StatusBadRequest)
							return
						}
						if got := r.Header.Get("Authorization"); got != "Bearer test-only" {
							t.Errorf("missing provider-scoped test credential")
						}
						body, err := io.ReadAll(r.Body)
						if err != nil {
							t.Error(err)
							return
						}
						requests <- body
						writeProtocolReply(t, w, api, stream, calls.Add(1) == 1)
					}))
					t.Cleanup(server.Close)
					llm, err := ResolveModelPtr(t.Context(), "alias", []*agentsv1.ModelProvider{{
						Name: "provider", Type: "openai", ApiKey: "test-only", BaseUrl: server.URL + "/v1", OpenaiApi: api,
						Models: []*agentsv1.ModelConfig{{Name: "actual-model", Alias: "alias"}},
					}})
					if err != nil {
						t.Fatal(err)
					}
					if llm.Name() != "actual-model" {
						t.Fatalf("model name = %q, want actual provider ID", llm.Name())
					}
					req := &model.LLMRequest{
						Contents: []*genai.Content{
							{Role: "user", Parts: []*genai.Part{{Text: "look"}, {InlineData: &genai.Blob{MIMEType: mime, Data: []byte{1, 2, 3}, DisplayName: "sample"}}}},
							genai.NewContentFromText("previous answer", genai.RoleModel),
							genai.NewContentFromText("continue", genai.RoleUser),
						},
						Config: &genai.GenerateContentConfig{
							SystemInstruction: genai.NewContentFromText("system-prompt", genai.RoleUser),
							Tools:             []*genai.Tool{{FunctionDeclarations: []*genai.FunctionDeclaration{{Name: "lookup", Parameters: &genai.Schema{Type: genai.TypeObject, Properties: map[string]*genai.Schema{"key": {Type: genai.TypeString}}}}}}},
						},
					}
					var call *genai.FunctionCall
					for res, err := range llm.GenerateContent(t.Context(), req, stream) {
						if err != nil {
							t.Fatal(err)
						}
						if res != nil && res.Content != nil {
							for _, part := range res.Content.Parts {
								if part.FunctionCall != nil {
									call = part.FunctionCall
								}
							}
						}
					}
					if call == nil || call.ID != "call_tool" || call.Name != "lookup" || call.Args["key"] != "value" {
						t.Fatalf("tool call = %+v, want preserved ID/name and complete arguments", call)
					}
					req.Contents = append(req.Contents,
						&genai.Content{Role: "model", Parts: []*genai.Part{{FunctionCall: call}}},
						&genai.Content{Role: "user", Parts: []*genai.Part{{FunctionResponse: &genai.FunctionResponse{ID: call.ID, Name: call.Name, Response: map[string]any{"result": "found"}}}}},
					)
					var final strings.Builder
					for res, err := range llm.GenerateContent(t.Context(), req, stream) {
						if err != nil {
							t.Fatal(err)
						}
						if res != nil && !res.Partial && res.Content != nil {
							for _, part := range res.Content.Parts {
								final.WriteString(part.Text)
							}
						}
					}
					if final.String() != "ok" || calls.Load() != 2 {
						t.Fatalf("final = %q, HTTP calls = %d, want ok after exactly two calls", final.String(), calls.Load())
					}
					for i := range 2 {
						body := <-requests
						var decoded map[string]any
						if err := json.Unmarshal(body, &decoded); err != nil {
							t.Fatal(err)
						}
						if decoded["model"] != "actual-model" {
							t.Fatalf("actual model ID not preserved: %s", body)
						}
						// Non-streaming requests may omit stream entirely.
						if (stream && decoded["stream"] != true) || (!stream && decoded["stream"] != nil && decoded["stream"] != false) {
							t.Fatalf("stream mode not preserved: %s", body)
						}
						for _, want := range []string{"system-prompt", "previous answer", "data:" + mime + ";base64,AQID"} {
							if !strings.Contains(string(body), want) {
								t.Errorf("request %d missing %q: %s", i+1, want, body)
							}
						}
						if api == "responses" {
							for _, want := range []string{`"output_text"`, `"strict":false`} {
								if !strings.Contains(string(body), want) {
									t.Errorf("Responses history/tool schema missing %s: %s", want, body)
								}
							}
						} else if strings.Contains(string(body), `"strict":true`) {
							t.Errorf("Chat tools must preserve optional arguments: %s", body)
						}
						if i == 1 && (strings.Count(string(body), "call_tool") != 2 || !strings.Contains(string(body), "found")) {
							t.Errorf("tool call/result correlation lost: %s", body)
						}
					}
				})
			}
		}
	}
}

func writeProtocolReply(t *testing.T, w http.ResponseWriter, api string, stream, tool bool) {
	t.Helper()
	function := map[string]any{"name": "lookup", "arguments": `{"key":"value"}`}
	if api == "responses" {
		item := map[string]any{"type": "message", "id": "msg_test", "role": "assistant", "status": "completed", "content": []any{map[string]any{"type": "output_text", "text": "ok", "annotations": []any{}}}}
		if tool {
			item = map[string]any{"type": "function_call", "id": "fc_test", "call_id": "call_tool", "name": "lookup", "arguments": function["arguments"]}
		}
		response := map[string]any{"id": "resp_test", "object": "response", "model": "actual-model", "status": "completed", "output": []any{item}}
		if !stream {
			writeProtocolJSON(t, w, response, false)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		if tool {
			writeProtocolJSON(t, w, map[string]any{"type": "response.output_item.added", "item": item}, true)
			for _, delta := range []string{`{"key":`, `"value"}`} {
				writeProtocolJSON(t, w, map[string]any{"type": "response.function_call_arguments.delta", "item_id": "fc_test", "delta": delta}, true)
			}
			writeProtocolJSON(t, w, map[string]any{"type": "response.function_call_arguments.done", "item_id": "fc_test", "arguments": ""}, true)
		} else {
			writeProtocolJSON(t, w, map[string]any{"type": "response.output_text.delta", "delta": "ok", "output_index": 0, "content_index": 0}, true)
		}
		writeProtocolJSON(t, w, map[string]any{"type": "response.completed", "response": response}, true)
		return
	}
	message := map[string]any{"role": "assistant", "content": "ok"}
	reason := "stop"
	if tool {
		message = map[string]any{"role": "assistant", "tool_calls": []any{map[string]any{"id": "call_tool", "type": "function", "function": function}}}
		reason = "tool_calls"
	}
	if !stream {
		writeProtocolJSON(t, w, map[string]any{"id": "chat_test", "object": "chat.completion", "model": "actual-model", "choices": []any{map[string]any{"index": 0, "message": message, "finish_reason": reason}}}, false)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	if tool {
		for i, args := range []string{`{"key":`, `"value"}`} {
			delta := map[string]any{"index": 0, "function": map[string]any{"arguments": args}}
			if i == 0 {
				delta["id"], delta["type"] = "call_tool", "function"
				delta["function"] = map[string]any{"name": "lookup", "arguments": args}
			}
			writeProtocolJSON(t, w, map[string]any{"id": "chat_test", "object": "chat.completion.chunk", "choices": []any{map[string]any{"index": 0, "delta": map[string]any{"tool_calls": []any{delta}}}}}, true)
		}
	} else {
		writeProtocolJSON(t, w, map[string]any{"id": "chat_test", "object": "chat.completion.chunk", "choices": []any{map[string]any{"index": 0, "delta": message}}}, true)
	}
	writeProtocolJSON(t, w, map[string]any{"id": "chat_test", "object": "chat.completion.chunk", "choices": []any{map[string]any{"index": 0, "delta": map[string]any{}, "finish_reason": reason}}}, true)
	fmt.Fprint(w, "data: [DONE]\n\n")
}

func writeProtocolJSON(t *testing.T, w http.ResponseWriter, value any, stream bool) {
	t.Helper()
	body, err := json.Marshal(value)
	if err != nil {
		t.Error(err)
		return
	}
	if stream {
		fmt.Fprintf(w, "data: %s\n\n", body)
	} else {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}
}

func TestOpenAIProviderDoesNotFallBackToAnotherProtocol(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.URL.Path != "/responses" {
			t.Errorf("unexpected fallback request to %s", r.URL.Path)
		}
		http.Error(w, "unsupported", http.StatusNotFound)
	}))
	t.Cleanup(server.Close)
	llm, err := createModelFromProvider(t.Context(), "test", &agentsv1.ModelProvider{Type: "openai", OpenaiApi: "responses", ApiKey: "test-only", BaseUrl: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	var callErr error
	for _, err := range llm.GenerateContent(t.Context(), &model.LLMRequest{Contents: []*genai.Content{genai.NewContentFromText("hello", genai.RoleUser)}}, false) {
		if err != nil {
			callErr = err
		}
	}
	if callErr == nil || calls.Load() != 1 {
		t.Fatalf("error = %v, calls = %d, want one failed Responses request", callErr, calls.Load())
	}
}

func TestProviderRejectsInvalidAPIAtBuild(t *testing.T) {
	for _, p := range []*agentsv1.ModelProvider{
		{Type: "openai", OpenaiApi: "invalid"},
		{Type: "gemini", OpenaiApi: "responses"},
	} {
		if _, err := createModelFromProvider(t.Context(), "test", p); err == nil || !strings.Contains(err.Error(), "openai_api") {
			t.Errorf("provider %+v should fail with an openai_api validation error, got %v", p, err)
		}
	}
}
