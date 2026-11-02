package runner

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	adkrunner "google.golang.org/adk/v2/runner"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/testsupport/openaifake"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestModelOverrideRetainsProviderProtocolForDuplicateModelIDs(t *testing.T) {
	chat := openaifake.New(t)
	responses := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/responses" {
			t.Errorf("wrong protocol: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":"resp_test","object":"response","model":"shared-model","status":"completed","output":[{"type":"message","id":"msg_test","role":"assistant","status":"completed","content":[{"type":"output_text","text":"responses-answer","annotations":[]}]}]}`)
	}))
	t.Cleanup(responses.Close)
	agents := []agentsv1.Agent{{
		Name: "protocol-selection", AgentId: "protocol-selection", WorkspaceId: "ws-a", Type: agentsv1.AgentType_AGENT_TYPE_LLM,
		Config: &agentsv1.AgentConfig{Model: "chat-alias"},
	}}
	root := &agents[0]
	providers := []agentsv1.ModelProvider{
		{Name: "chat", Type: "openai", ApiKey: "test-only", BaseUrl: chat.URL(), Models: []*agentsv1.ModelConfig{{Name: "shared-model", Alias: "chat-alias"}}},
		{Name: "responses", Type: "openai", ApiKey: "test-only", BaseUrl: responses.URL, OpenaiApi: "responses", Models: []*agentsv1.ModelConfig{{Name: "shared-model", Alias: "responses-alias"}}},
	}
	svc, err := NewService(t.Context(), agents, providers, nil, nil, nil, session.InMemoryService(), nil, nil, adkrunner.PluginConfig{})
	if err != nil {
		t.Fatal(err)
	}
	ci := turnCtxInfo(root)
	if _, err := svc.Run(t.Context(), root.Name, []*genai.Part{{Text: "warm"}}, "", ci, nil, nil); err != nil {
		t.Fatal(err)
	}
	for range 2 { // First build, then the warmed model-override cache.
		out, err := svc.Run(t.Context(), root.Name, []*genai.Part{{Text: "use the other provider"}}, "responses-alias", ci, nil, nil)
		if err != nil || out != "responses-answer" {
			t.Fatalf("selected Responses provider = (%q, %v), want responses-answer", out, err)
		}
	}
	if got := chat.CallCount("shared-model"); got != 1 {
		t.Fatalf("Chat provider called %d times, want only the initial warm turn", got)
	}
}
