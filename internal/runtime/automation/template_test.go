package automation

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestRenderTemplate(t *testing.T) {
	roots := map[string]any{
		"payload": map[string]any{
			"kind":  "incident",
			"count": json.Number("3"),
			"tags":  []any{"a", "b"},
			"quote": `say "hi"`,
		},
		"steps": map[string]any{
			"summarize": map[string]any{"response": "all good"},
		},
	}

	cases := []struct {
		name    string
		in      string
		want    string
		wantErr string
	}{
		{name: "no placeholders", in: "static text", want: "static text"},
		{name: "string", in: "kind={{payload.kind}}", want: "kind=incident"},
		{name: "number", in: "n={{ payload.count }}", want: "n=3"},
		{name: "structured as json", in: "tags={{payload.tags}}", want: `tags=["a","b"]`},
		{name: "step output", in: "{{steps.summarize.response}}", want: "all good"},
		{name: "json filter quotes strings", in: `{"msg": {{payload.quote | json}}}`, want: `{"msg": "say \"hi\""}`},
		{name: "missing selector fails", in: "{{payload.absent}}", wantErr: "not found"},
		{name: "empty selector fails", in: "{{ }}", wantErr: "empty template selector"},
		{name: "unknown filter fails", in: "{{payload.kind | upper}}", wantErr: "unsupported template filter"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := renderTemplate(tc.in, roots)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("err = %v, want containing %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("renderTemplate: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// capturingHTTPClient records request bodies so a test can assert what a
// templated webhook step actually sent.
type capturingHTTPClient struct {
	mu     sync.Mutex
	bodies []string
	urls   []string
}

func (c *capturingHTTPClient) Do(req *http.Request) (*http.Response, error) {
	body, _ := io.ReadAll(req.Body)
	c.mu.Lock()
	c.bodies = append(c.bodies, string(body))
	c.urls = append(c.urls, req.URL.String())
	c.mu.Unlock()
	return &http.Response{
		StatusCode: http.StatusNoContent,
		Body:       io.NopCloser(strings.NewReader("")),
		Header:     make(http.Header),
		Request:    req,
	}, nil
}

func TestEngineTemplatesChainStepOutputs(t *testing.T) {
	ctx := context.Background()
	defRepo := NewMemoryDefinitionRepo()
	runRepo := NewMemoryRunRepo()
	stepRepo := NewMemoryStepRunRepo()
	runnerSvc := &engineRunner{outputs: []string{"all clear"}}
	httpClient := &capturingHTTPClient{}
	engine := NewEngine(defRepo, runRepo, stepRepo, EngineOptions{
		Runner:     runnerSvc,
		HTTPClient: httpClient,
	})

	run, err := engine.Execute(ctx, &agentsv1.Automation{
		Name:        "chained",
		Enabled:     true,
		WorkspaceId: "ws1",
		Steps: []*agentsv1.AutomationStep{
			{
				Name: "summarize",
				Type: agentsv1.AutomationStepType_AUTOMATION_STEP_TYPE_INVOKE_AGENT,
				InvokeAgent: &agentsv1.AutomationInvokeAgentStep{
					AgentId: "agent1-id",
					Input:   "summarize incident {{payload.id}}",
				},
			},
			{
				Name: "report",
				Type: agentsv1.AutomationStepType_AUTOMATION_STEP_TYPE_CALL_WEBHOOK,
				CallWebhook: &agentsv1.AutomationCallWebhookStep{
					Url:         "https://example.test/hook/{{payload.id}}",
					PayloadJson: `{"summary": {{steps.summarize.response | json}}}`,
				},
			},
		},
	}, agentsv1.AutomationTriggerType_AUTOMATION_TRIGGER_TYPE_MANUAL, `{"id":"inc-42"}`)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if run.GetStatus() != agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_SUCCEEDED {
		t.Fatalf("status = %s, want succeeded; err=%s", run.GetStatus(), run.GetError())
	}
	if len(httpClient.bodies) != 1 {
		t.Fatalf("webhook calls = %d, want 1", len(httpClient.bodies))
	}
	if got, want := httpClient.urls[0], "https://example.test/hook/inc-42"; got != want {
		t.Fatalf("webhook url = %q, want %q", got, want)
	}
	if got, want := httpClient.bodies[0], `{"summary": "all clear"}`; got != want {
		t.Fatalf("webhook body = %q, want %q", got, want)
	}
}

func TestEngineTemplateMissingSelectorFailsStep(t *testing.T) {
	ctx := context.Background()
	engine, _, _, stepRepo := newMinimalEngine()

	run, err := engine.Execute(ctx, &agentsv1.Automation{
		Name:        "broken-template",
		Enabled:     true,
		WorkspaceId: "ws1",
		Steps: []*agentsv1.AutomationStep{
			{
				Name: "invoke",
				Type: agentsv1.AutomationStepType_AUTOMATION_STEP_TYPE_INVOKE_AGENT,
				InvokeAgent: &agentsv1.AutomationInvokeAgentStep{
					AgentId: "agent1-id",
					Input:   "hello {{payload.nope}}",
				},
			},
		},
	}, agentsv1.AutomationTriggerType_AUTOMATION_TRIGGER_TYPE_MANUAL, `{}`)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if run.GetStatus() != agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_FAILED {
		t.Fatalf("status = %s, want failed", run.GetStatus())
	}
	if !strings.Contains(run.GetError(), "payload.nope") {
		t.Fatalf("run error = %q, want it to name the missing selector", run.GetError())
	}
	steps, err := stepRepo.ListByRun(ctx, "ws1", run.GetId())
	if err != nil || len(steps) != 1 {
		t.Fatalf("steps = %d (err %v), want 1", len(steps), err)
	}
	if steps[0].GetStatus() != agentsv1.AutomationStepRunStatus_AUTOMATION_STEP_RUN_STATUS_FAILED {
		t.Fatalf("step status = %s, want failed", steps[0].GetStatus())
	}
}
