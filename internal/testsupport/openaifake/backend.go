package openaifake

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Backend is an OpenAI-compatible chat completions endpoint for tests. It
// records complete requests and derived user input by actual model ID.
type Backend struct {
	server *httptest.Server

	mu sync.Mutex
	// scripts answer a model's calls in place of the echo default.
	scripts           map[string]script
	inputsByModelID   map[string][]string
	requestsByModelID map[string][]ChatCompletionRequest
}

// script answers one call: the HTTP request, whose body is already read, and
// the decoded request.
type script func(http.ResponseWriter, *http.Request, ChatCompletionRequest)

type ChatCompletionRequest struct {
	Model    string                  `json:"model"`
	Messages []ChatCompletionMessage `json:"messages"`
	Decoded  map[string]any          `json:"-"`
}

type ChatCompletionMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

func New(t testing.TB) *Backend {
	t.Helper()
	b := &Backend{
		scripts:           make(map[string]script),
		inputsByModelID:   make(map[string][]string),
		requestsByModelID: make(map[string][]ChatCompletionRequest),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/chat/completions", b.handleCompletion)
	b.server = httptest.NewServer(mux)
	t.Cleanup(b.server.Close)
	return b
}

func (b *Backend) URL() string {
	return b.server.URL
}

func (b *Backend) handleCompletion(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var req ChatCompletionRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(body, &req.Decoded); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	lastUser := lastUserInput(req.Messages)
	b.mu.Lock()
	b.inputsByModelID[req.Model] = append(b.inputsByModelID[req.Model], lastUser)
	b.requestsByModelID[req.Model] = append(b.requestsByModelID[req.Model], req)
	answer := b.scripts[req.Model]
	b.mu.Unlock()

	if answer != nil {
		answer(w, r, req)
		return
	}
	WriteCompletion(w, req.Model, fmt.Sprintf("%s(%s)", req.Model, lastUser))
}

func (b *Backend) setScript(model string, answer script) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.scripts[model] = answer
}

func lastUserInput(messages []ChatCompletionMessage) string {
	lastUser := ""
	for _, message := range messages {
		if message.Role != "user" {
			continue
		}
		var text string
		if json.Unmarshal(message.Content, &text) == nil {
			lastUser = text
			continue
		}
		var parts []struct {
			Text string `json:"text"`
		}
		if json.Unmarshal(message.Content, &parts) == nil {
			var joined strings.Builder
			for _, part := range parts {
				joined.WriteString(part.Text)
			}
			lastUser = joined.String()
		}
	}
	return lastUser
}

// Answer scripts a fixed reply for a model, overriding the echo default.
func (b *Backend) Answer(model, reply string) {
	b.Script(model, func(w http.ResponseWriter, _ *http.Request) {
		WriteCompletion(w, model, reply)
	})
}

// Script installs a handler for a model, replacing the echo default.
func (b *Backend) Script(model string, handler http.HandlerFunc) {
	b.setScript(model, func(w http.ResponseWriter, r *http.Request, _ ChatCompletionRequest) {
		handler(w, r)
	})
}

// RequireConcurrent blocks completions until n requests are in flight.
func (b *Backend) RequireConcurrent(model string, n int32) {
	var inFlight atomic.Int32
	proceed := make(chan struct{})
	var once sync.Once
	b.Script(model, func(w http.ResponseWriter, _ *http.Request) {
		if inFlight.Add(1) >= n {
			once.Do(func() { close(proceed) })
		}
		select {
		case <-proceed:
			WriteCompletion(w, model, "done")
		case <-time.After(3 * time.Second):
			http.Error(w, `{"error": {"message": "items were not processed concurrently"}}`, http.StatusBadRequest)
		}
	})
}

// FailFirstCall makes a model fail once with an HTTP 400, then echo normally.
func (b *Backend) FailFirstCall(model string) {
	failed := false
	b.Script(model, func(w http.ResponseWriter, _ *http.Request) {
		b.mu.Lock()
		first := !failed
		failed = true
		input := ""
		if inputs := b.inputsByModelID[model]; len(inputs) > 0 {
			input = inputs[len(inputs)-1]
		}
		b.mu.Unlock()
		if first {
			http.Error(w, `{"error": {"message": "transient failure"}}`, http.StatusBadRequest)
			return
		}
		WriteCompletion(w, model, fmt.Sprintf("%s(%s)", model, input))
	})
}

func (b *Backend) LastInput(model string) string {
	b.mu.Lock()
	defer b.mu.Unlock()
	inputs := b.inputsByModelID[model]
	if len(inputs) == 0 {
		return ""
	}
	return inputs[len(inputs)-1]
}

func (b *Backend) LastRequest(model string) (ChatCompletionRequest, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	requests := b.requestsByModelID[model]
	if len(requests) == 0 {
		return ChatCompletionRequest{}, false
	}
	return requests[len(requests)-1], true
}

func (b *Backend) CallCount(model string) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.requestsByModelID[model])
}

func WriteCompletion(w http.ResponseWriter, model, reply string) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{
		"id": "cmpl-test",
		"object": "chat.completion",
		"created": 1,
		"model": %q,
		"choices": [{"index": 0, "message": {"role": "assistant", "content": %q}, "finish_reason": "stop"}]
	}`, model, reply)
}

// ToolCall is one function call a scripted model reply asks the agent to make.
type ToolCall struct {
	ID        string
	Name      string
	Arguments string
}

// ScriptRequest installs a handler that sees the decoded request — which the
// plain Script handler cannot, because the body is already consumed — so a
// script can answer differently per turn (e.g. call a tool, then reply once
// the tool result is in the conversation).
func (b *Backend) ScriptRequest(model string, handler func(w http.ResponseWriter, req ChatCompletionRequest)) {
	b.setScript(model, func(w http.ResponseWriter, _ *http.Request, req ChatCompletionRequest) {
		handler(w, req)
	})
}

// ScriptCall is ScriptRequest with the HTTP request as well, whose context
// ends when the caller gives up on the call — so a script can hold a call
// open and tell a cancelled call from one it answered.
func (b *Backend) ScriptCall(model string, handler func(w http.ResponseWriter, r *http.Request, req ChatCompletionRequest)) {
	b.setScript(model, handler)
}

// Streaming reports whether the request asked for an SSE chunk stream.
func (r ChatCompletionRequest) Streaming() bool {
	stream, _ := r.Decoded["stream"].(bool)
	return stream
}

// LastRole returns the role of the conversation's final message.
func (r ChatCompletionRequest) LastRole() string {
	if len(r.Messages) == 0 {
		return ""
	}
	return r.Messages[len(r.Messages)-1].Role
}

// WriteReply answers in the shape the request asked for — an SSE chunk
// stream when it set "stream": true, one JSON completion otherwise — with
// either text or tool calls.
func WriteReply(w http.ResponseWriter, req ChatCompletionRequest, text string, calls ...ToolCall) {
	finish := "stop"
	if len(calls) > 0 {
		finish = "tool_calls"
	}
	if !req.Streaming() {
		message := map[string]any{"role": "assistant", "content": text}
		if len(calls) > 0 {
			message["tool_calls"] = toolCallsJSON(calls, false)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": "cmpl-test", "object": "chat.completion", "created": 1, "model": req.Model,
			"choices": []any{map[string]any{"index": 0, "message": message, "finish_reason": finish}},
		})
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	chunk := func(delta map[string]any, finishReason any) {
		raw, _ := json.Marshal(map[string]any{
			"id": "cmpl-test", "object": "chat.completion.chunk", "created": 1, "model": req.Model,
			"choices": []any{map[string]any{"index": 0, "delta": delta, "finish_reason": finishReason}},
		})
		fmt.Fprintf(w, "data: %s\n\n", raw)
	}
	if text != "" {
		chunk(map[string]any{"role": "assistant", "content": text}, nil)
	}
	if len(calls) > 0 {
		chunk(map[string]any{"role": "assistant", "tool_calls": toolCallsJSON(calls, true)}, nil)
	}
	chunk(map[string]any{}, finish)
	fmt.Fprint(w, "data: [DONE]\n\n")
}

func toolCallsJSON(calls []ToolCall, streaming bool) []any {
	out := make([]any, 0, len(calls))
	for i, c := range calls {
		call := map[string]any{
			"id": c.ID, "type": "function",
			"function": map[string]any{"name": c.Name, "arguments": c.Arguments},
		}
		if streaming {
			call["index"] = i
		}
		out = append(out, call)
	}
	return out
}
