package http

import (
	"context"
	"net/http"
	"testing"

	configmemory "go.orx.me/apps/butter/internal/repo/config/memory"
	"go.orx.me/apps/butter/internal/runtime/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/genai"
)

func TestModelOverrideProp(t *testing.T) {
	for _, tt := range []struct {
		name    string
		props   any
		want    string
		invalid bool
	}{
		{name: "absent"},
		{name: "other extensions", props: map[string]any{"butterRun": map[string]any{"detach": true}}},
		{name: "empty object", props: map[string]any{"butterModel": map[string]any{}}},
		{name: "default", props: map[string]any{"butterModel": map[string]any{"override": ""}}},
		{name: "alias", props: map[string]any{"butterModel": map[string]any{"override": " fast "}}, want: "fast"},
		{name: "null", props: map[string]any{"butterModel": nil}, invalid: true},
		{name: "wrong object", props: map[string]any{"butterModel": "fast"}, invalid: true},
		{name: "wrong value", props: map[string]any{"butterModel": map[string]any{"override": 42}}, invalid: true},
		{name: "unknown field", props: map[string]any{"butterModel": map[string]any{"provider": "p"}}, invalid: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := modelOverrideProp(tt.props)
			if got != tt.want || (err != nil) != tt.invalid {
				t.Fatalf("got %q, %v", got, err)
			}
		})
	}
}

func TestPrepareModelOverride(t *testing.T) {
	store := configmemory.New()
	_, err := store.CreateModelProvider(t.Context(), "ws-a", &agentsv1.ModelProvider{Name: "provider", Models: []*agentsv1.ModelConfig{{Name: "model-a", Alias: "fast"}}})
	if err != nil {
		t.Fatal(err)
	}
	h := NewAGUIHandler(store)
	h.SetModelProviderRepo(store)
	for _, tt := range []struct {
		name, workspace, model string
		kind                   agentsv1.AgentType
		status                 int
	}{
		{name: "alias", workspace: "ws-a", model: "fast", kind: agentsv1.AgentType_AGENT_TYPE_LLM},
		{name: "name", workspace: "ws-a", model: "model-a", kind: agentsv1.AgentType_AGENT_TYPE_LLM},
		{name: "legacy LLM", workspace: "ws-a", model: "fast"},
		{name: "unknown", workspace: "ws-a", model: "missing", status: http.StatusBadRequest},
		{name: "other workspace", workspace: "ws-b", model: "fast", status: http.StatusBadRequest},
		{name: "pi", workspace: "ws-a", model: "fast", kind: agentsv1.AgentType_AGENT_TYPE_PI, status: http.StatusBadRequest},
		{name: "cursor", workspace: "ws-a", model: "fast", kind: agentsv1.AgentType_AGENT_TYPE_CURSOR, status: http.StatusBadRequest},
		{name: "workflow", workspace: "ws-a", model: "fast", kind: agentsv1.AgentType_AGENT_TYPE_WORKFLOW, status: http.StatusBadRequest},
		{name: "pi default", workspace: "ws-a", kind: agentsv1.AgentType_AGENT_TYPE_PI},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, status, err := h.prepareModelOverride(t.Context(), aguiTarget{workspaceID: tt.workspace, agent: &agentsv1.Agent{Type: tt.kind}}, map[string]any{"butterModel": map[string]any{"override": tt.model}})
			if status != tt.status || (err != nil) != (tt.status != 0) {
				t.Fatalf("got %q, %d, %v", got, status, err)
			}
			if err == nil && got != tt.model {
				t.Fatalf("override = %q", got)
			}
		})
	}
	h.SetModelProviderRepo(nil)
	_, status, _ := h.prepareModelOverride(t.Context(), aguiTarget{agent: &agentsv1.Agent{}}, map[string]any{"butterModel": map[string]any{"override": "fast"}})
	if status != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", status)
	}
}

type modelCaptureRunner struct {
	mockRunner
	model string
}

func (r *modelCaptureRunner) RunSSE(ctx context.Context, name string, parts []*genai.Part, model string, info *agentsv1.ContextInfo, event runner.EventCallback, compact runner.CompactionCallback) (string, error) {
	r.model = model
	return r.mockRunner.RunSSE(ctx, name, parts, model, info, event, compact)
}

func TestAGUIRunTurnPassesModelOverride(t *testing.T) {
	for _, model := range []string{"", "fast"} {
		svc := &modelCaptureRunner{}
		run := &aguiRun{rc: &aguiRunContext{svc: svc, agent: &agentsv1.Agent{AgentId: "a"}, agentName: "a", modelOverride: model, ctxInfo: &agentsv1.ContextInfo{}}, sink: newAGUISink("t", "r", "m", (&aguiEventRecorder{}).emit)}
		if err := run.runTurn(t.Context()); err != nil {
			t.Fatal(err)
		}
		if svc.model != model {
			t.Fatalf("model = %q, want %q", svc.model, model)
		}
	}
}

func TestAGUIDetachedRecordsModelOverride(t *testing.T) {
	d := newDetachHarness(t)
	store := configmemory.New()
	_, err := store.CreateModelProvider(t.Context(), "ws-a", &agentsv1.ModelProvider{Name: "p", Models: []*agentsv1.ModelConfig{{Name: "card-model"}}})
	if err != nil {
		t.Fatal(err)
	}
	d.handler.SetModelProviderRepo(store)
	body := detachBody("t-model", "r-model", "hi")
	body["forwardedProps"].(map[string]any)["butterModel"] = map[string]any{"override": "card-model"}
	post := d.start("carder", body)
	w := post.wait(t)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d: %s", w.Code, w.Body.String())
	}
	d.waitForRuns()
	inv, ok := d.record("t-model", "r-model")
	if !ok || inv.GetModelOverride() != "card-model" || inv.GetStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED {
		t.Fatalf("record = %v", inv)
	}
}
