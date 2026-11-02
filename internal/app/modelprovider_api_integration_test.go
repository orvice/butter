package app

import (
	"encoding/json/v2"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"connectrpc.com/connect"
	adkrunner "google.golang.org/adk/v2/runner"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/application"
	"go.orx.me/apps/butter/internal/config"
	"go.orx.me/apps/butter/internal/runtime/runner"
	"go.orx.me/apps/butter/internal/testsupport/openaifake"
	"go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// Warm both runner caches, change only the provider's protocol via the real
// service/reload path, then continue the same sessions with their old images.
func TestModelProviderProtocolReloadPreservesSessions(t *testing.T) {
	type captured struct{ path, body string }
	requests := make(chan captured, 8)
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
			return
		}
		requests <- captured{r.URL.Path, string(body)}
		var req struct {
			Model string `json:"model"`
		}
		if err := json.Unmarshal(body, &req); err != nil {
			t.Error(err)
			return
		}
		switch r.URL.Path {
		case "/chat/completions":
			openaifake.WriteCompletion(w, req.Model, "chat-answer")
		case "/responses":
			w.Header().Set("Content-Type", "text/event-stream")
			fmt.Fprint(w, "data: {\"type\":\"response.output_text.delta\",\"delta\":\"responses-answer\"}\n\n")
			fmt.Fprintf(w, "data: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_test\",\"object\":\"response\",\"status\":\"completed\",\"model\":%q,\"output\":[{\"type\":\"message\",\"id\":\"msg_test\",\"role\":\"assistant\",\"status\":\"completed\",\"content\":[{\"type\":\"output_text\",\"text\":\"responses-answer\",\"annotations\":[]}]}]}}\n\n", req.Model)
		default:
			t.Errorf("unexpected endpoint %s", r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(backend.Close)
	ctx := workspace.WithID(t.Context(), "ws-test")
	store := NewConfigStore()
	root := &agentsv1.Agent{
		Name: "protocol-agent", AgentId: "protocol-agent", Type: agentsv1.AgentType_AGENT_TYPE_LLM, WorkspaceId: "ws-test",
		Config: &agentsv1.AgentConfig{Model: "default-alias"},
	}
	if _, err := store.CreateAgent(ctx, "ws-test", root); err != nil {
		t.Fatal(err)
	}
	provider := &agentsv1.ModelProvider{
		Name: "switchable", Type: "openai", ApiKey: "test-only", BaseUrl: backend.URL,
		Models: []*agentsv1.ModelConfig{{Name: "default-model", Alias: "default-alias"}, {Name: "override-model", Alias: "override-alias"}},
	}
	if _, err := store.CreateModelProvider(ctx, "ws-test", provider); err != nil {
		t.Fatal(err)
	}
	cfg := &config.AppConfig{}
	if err := store.SyncToConfig(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	runs, err := runner.NewService(ctx, cfg.Agents, cfg.ModelProviders, nil, nil, nil, session.InMemoryService(), nil, nil, adkrunner.PluginConfig{})
	if err != nil {
		t.Fatal(err)
	}
	runtime := NewConfigRuntime(store, cfg)
	runtime.SetRunnerService(runs)
	svc := application.NewModelProviderServiceServer(store)
	svc.SetRuntime(runtime)

	for _, override := range []string{"", "override-alias"} {
		ci := configRuntimeContextInfo("protocol-session-" + override)
		parts := []*genai.Part{{Text: "look"}, genai.NewPartFromBytes([]byte{1, 2, 3}, "image/png")}
		out, err := runs.Run(ctx, root.Name, parts, override, ci, nil, nil)
		if err != nil || out != "chat-answer" {
			t.Fatalf("warm run = (%q, %v)", out, err)
		}
		if got := <-requests; got.path != "/chat/completions" {
			t.Fatalf("legacy provider used %s", got.path)
		}
	}

	provider.OpenaiApi = "responses"
	if _, err := svc.UpdateModelProvider(ctx, connect.NewRequest(&agentsv1.UpdateModelProviderRequest{ModelProvider: provider})); err != nil {
		t.Fatalf("protocol update/reload: %v", err)
	}
	if cfg.ModelProviders[0].GetOpenaiApi() != "responses" {
		t.Fatal("runtime snapshot dropped the provider API")
	}
	for _, override := range []string{"", "override-alias"} {
		ci := configRuntimeContextInfo("protocol-session-" + override)
		out, err := runs.RunSSE(ctx, root.Name, []*genai.Part{{Text: "continue"}}, override, ci, nil, nil)
		if err != nil || out != "responses-answer" {
			t.Fatalf("resumed run = (%q, %v)", out, err)
		}
		got := <-requests
		if got.path != "/responses" {
			t.Fatalf("warmed runner kept the old protocol: %s", got.path)
		}
		for _, want := range []string{"chat-answer", `"output_text"`, "data:image/png;base64,AQID"} {
			if !strings.Contains(got.body, want) {
				t.Errorf("old session history missing %q: %s", want, got.body)
			}
		}
	}

	// The reverse edit must also keep the Responses turn in session history.
	provider.OpenaiApi = "chat_completions"
	if _, err := svc.UpdateModelProvider(ctx, connect.NewRequest(&agentsv1.UpdateModelProviderRequest{ModelProvider: provider})); err != nil {
		t.Fatal(err)
	}
	if _, err := runs.Run(ctx, root.Name, []*genai.Part{{Text: "again"}}, "", configRuntimeContextInfo("protocol-session-"), nil, nil); err != nil {
		t.Fatal(err)
	}
	if got := <-requests; got.path != "/chat/completions" || !strings.Contains(got.body, "responses-answer") {
		t.Fatalf("reverse protocol edit lost history: %+v", got)
	}
}
