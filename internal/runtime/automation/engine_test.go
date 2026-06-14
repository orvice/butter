package automation

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"google.golang.org/genai"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/notify"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	"go.orx.me/apps/butter/internal/repo/forum"
	forummemory "go.orx.me/apps/butter/internal/repo/forum/memory"
	"go.orx.me/apps/butter/internal/runtime/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type engineRunner struct {
	mu      sync.Mutex
	outputs []string
	errs    []error
	block   chan struct{}
	calls   int
}

func (r *engineRunner) HasAgentInWorkspace(workspaceID, name string) bool {
	return workspaceID == "ws1" && name == "agent1"
}

func (r *engineRunner) RunSSE(ctx context.Context, _ string, _ []*genai.Part, _ string, _ *agentsv1.ContextInfo, _ runner.EventCallback, _ runner.CompactionCallback) (string, error) {
	r.mu.Lock()
	r.calls++
	call := r.calls
	r.mu.Unlock()
	if r.block != nil {
		select {
		case <-r.block:
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.errs) > 0 {
		err := r.errs[0]
		r.errs = r.errs[1:]
		if err != nil {
			return "", err
		}
	}
	if len(r.outputs) >= call {
		return r.outputs[call-1], nil
	}
	if len(r.outputs) > 0 {
		return r.outputs[len(r.outputs)-1], nil
	}
	return "ok", nil
}

type engineNotifyRepo struct {
	group *agentsv1.NotifyGroup
}

func (r *engineNotifyRepo) ListNotifyGroups(context.Context, string) ([]*agentsv1.NotifyGroup, error) {
	return []*agentsv1.NotifyGroup{r.group}, nil
}

func (r *engineNotifyRepo) GetNotifyGroup(_ context.Context, workspaceID, name string) (*agentsv1.NotifyGroup, error) {
	if r.group == nil || r.group.GetWorkspaceId() != workspaceID || r.group.GetName() != name {
		return nil, configrepo.ErrNotFound
	}
	return r.group, nil
}

func (r *engineNotifyRepo) CreateNotifyGroup(context.Context, string, *agentsv1.NotifyGroup) (*agentsv1.NotifyGroup, error) {
	return nil, errors.New("not implemented")
}

func (r *engineNotifyRepo) UpdateNotifyGroup(context.Context, string, *agentsv1.NotifyGroup) (*agentsv1.NotifyGroup, error) {
	return nil, errors.New("not implemented")
}

func (r *engineNotifyRepo) DeleteNotifyGroup(context.Context, string, string) error {
	return errors.New("not implemented")
}

func (r *engineNotifyRepo) ListNotifyGroupsAcrossWorkspaces(context.Context) ([]*agentsv1.NotifyGroup, error) {
	return []*agentsv1.NotifyGroup{r.group}, nil
}

type engineNotifier struct {
	mu    sync.Mutex
	calls int
}

func (n *engineNotifier) Send(context.Context, *agentsv1.NotifyTarget, notify.Message) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.calls++
	return nil
}

type engineHTTPClient struct {
	mu       sync.Mutex
	statuses []int
	bodies   []string
	errs     []error
	calls    int
}

func (c *engineHTTPClient) Do(req *http.Request) (*http.Response, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls++
	if len(c.errs) > 0 {
		err := c.errs[0]
		c.errs = c.errs[1:]
		if err != nil {
			return nil, err
		}
	}
	status := http.StatusNoContent
	if len(c.statuses) > 0 {
		status = c.statuses[0]
		c.statuses = c.statuses[1:]
	}
	body := ""
	if len(c.bodies) > 0 {
		body = c.bodies[0]
		c.bodies = c.bodies[1:]
	}
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
		Request:    req,
	}, nil
}

func TestEngineRunNowExecutesAllStepTypes(t *testing.T) {
	ctx := context.Background()
	defRepo := NewMemoryDefinitionRepo()
	runRepo := NewMemoryRunRepo()
	stepRepo := NewMemoryStepRunRepo()
	runnerSvc := &engineRunner{outputs: []string{"agent output"}}
	notifier := &engineNotifier{}
	httpClient := &engineHTTPClient{statuses: []int{http.StatusAccepted}, bodies: []string{"accepted"}}
	forumRepo := forummemory.New()
	now := timestamppb.Now()
	if err := forumRepo.CreateThread(ctx, &agentsv1.ForumThread{
		Id:          "thread1",
		Title:       "Ops",
		Body:        "thread",
		Status:      "open",
		CreatedAt:   now,
		UpdatedAt:   now,
		WorkspaceId: "ws1",
	}); err != nil {
		t.Fatalf("seed forum thread: %v", err)
	}

	engine := NewEngine(defRepo, runRepo, stepRepo, EngineOptions{
		Runner: runnerSvc,
		NotifyGroupRepo: &engineNotifyRepo{group: &agentsv1.NotifyGroup{
			Name:        "ops",
			Enabled:     true,
			WorkspaceId: "ws1",
			Targets: []*agentsv1.NotifyTarget{
				{Name: "target1", Enabled: true, Type: agentsv1.NotifyTargetType_NOTIFY_TARGET_TYPE_LARK_WEBHOOK},
			},
		}},
		Notifier:   notifier,
		ForumRepo:  forumRepo,
		HTTPClient: httpClient,
	})

	automation := &agentsv1.Automation{
		Name:        "daily",
		Enabled:     true,
		WorkspaceId: "ws1",
		Trigger:     &agentsv1.AutomationTrigger{Type: agentsv1.AutomationTriggerType_AUTOMATION_TRIGGER_TYPE_MANUAL},
		Conditions: []*agentsv1.AutomationCondition{
			{Selector: "payload.kind", Operator: agentsv1.AutomationConditionOperator_AUTOMATION_CONDITION_OPERATOR_EQUALS, Value: "incident"},
		},
		Steps: []*agentsv1.AutomationStep{
			{Name: "summarize", Type: agentsv1.AutomationStepType_AUTOMATION_STEP_TYPE_INVOKE_AGENT, InvokeAgent: &agentsv1.AutomationInvokeAgentStep{AgentName: "agent1", Input: "summarize"}},
			{Name: "webhook", Type: agentsv1.AutomationStepType_AUTOMATION_STEP_TYPE_CALL_WEBHOOK, CallWebhook: &agentsv1.AutomationCallWebhookStep{Url: "https://example.test/hook", PayloadJson: `{"ok":true}`}},
			{Name: "notify", Type: agentsv1.AutomationStepType_AUTOMATION_STEP_TYPE_SEND_NOTIFY_GROUP, SendNotifyGroup: &agentsv1.AutomationSendNotifyGroupStep{NotifyGroupName: "ops", Title: "done", Message: "ok"}},
			{Name: "post", Type: agentsv1.AutomationStepType_AUTOMATION_STEP_TYPE_CREATE_FORUM_POST, CreateForumPost: &agentsv1.AutomationCreateForumPostStep{ThreadId: "thread1", Body: "posted"}},
		},
	}
	if err := defRepo.Create(ctx, automation); err != nil {
		t.Fatalf("create automation: %v", err)
	}

	run, err := engine.RunNow(ctx, "ws1", "daily", `{"kind":"incident"}`)
	if err != nil {
		t.Fatalf("RunNow: %v", err)
	}
	if run.GetStatus() != agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_SUCCEEDED {
		t.Fatalf("status = %s, want succeeded; err=%s", run.GetStatus(), run.GetError())
	}
	steps, err := stepRepo.ListByRun(ctx, "ws1", run.GetId())
	if err != nil {
		t.Fatalf("ListByRun: %v", err)
	}
	if len(steps) != 4 {
		t.Fatalf("step count = %d, want 4", len(steps))
	}
	for _, stepRun := range steps {
		if stepRun.GetStatus() != agentsv1.AutomationStepRunStatus_AUTOMATION_STEP_RUN_STATUS_SUCCEEDED {
			t.Fatalf("step %s status = %s, want succeeded: %s", stepRun.GetStepName(), stepRun.GetStatus(), stepRun.GetError())
		}
	}
	if notifier.calls != 1 {
		t.Fatalf("notifier calls = %d, want 1", notifier.calls)
	}
	posts, _, total, err := forumRepo.ListPosts(ctx, forum.PostListFilter{WorkspaceID: "ws1", ThreadID: "thread1"}, 10, "")
	if err != nil {
		t.Fatalf("ListPosts: %v", err)
	}
	if total != 1 || len(posts) != 1 || posts[0].GetBody() != "posted" {
		t.Fatalf("forum posts = len %d total %d, want one created post", len(posts), total)
	}
}

func TestEngineConditionFailureSkipsRun(t *testing.T) {
	ctx := context.Background()
	engine, _, _, stepRepo := newMinimalEngine()
	run, err := engine.Execute(ctx, &agentsv1.Automation{
		Name:        "conditional",
		Enabled:     true,
		WorkspaceId: "ws1",
		Conditions: []*agentsv1.AutomationCondition{
			{Selector: "payload.kind", Operator: agentsv1.AutomationConditionOperator_AUTOMATION_CONDITION_OPERATOR_EQUALS, Value: "incident"},
		},
		Steps: []*agentsv1.AutomationStep{{Name: "summarize", Type: agentsv1.AutomationStepType_AUTOMATION_STEP_TYPE_INVOKE_AGENT, InvokeAgent: &agentsv1.AutomationInvokeAgentStep{AgentName: "agent1"}}},
	}, agentsv1.AutomationTriggerType_AUTOMATION_TRIGGER_TYPE_MANUAL, `{"kind":"note"}`)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if run.GetStatus() != agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_SKIPPED {
		t.Fatalf("status = %s, want skipped", run.GetStatus())
	}
	steps, err := stepRepo.ListByRun(ctx, "ws1", run.GetId())
	if err != nil {
		t.Fatalf("ListByRun: %v", err)
	}
	if len(steps) != 0 {
		t.Fatalf("steps executed = %d, want 0", len(steps))
	}
}

func TestEngineStepFailureStopsLaterSteps(t *testing.T) {
	ctx := context.Background()
	engine, _, _, stepRepo := newMinimalEngine()
	run, err := engine.Execute(ctx, &agentsv1.Automation{
		Name:        "failure",
		Enabled:     true,
		WorkspaceId: "ws1",
		Steps: []*agentsv1.AutomationStep{
			{Name: "bad-webhook", Type: agentsv1.AutomationStepType_AUTOMATION_STEP_TYPE_CALL_WEBHOOK, CallWebhook: &agentsv1.AutomationCallWebhookStep{Url: "https://example.test/fail"}},
			{Name: "never", Type: agentsv1.AutomationStepType_AUTOMATION_STEP_TYPE_INVOKE_AGENT, InvokeAgent: &agentsv1.AutomationInvokeAgentStep{AgentName: "agent1"}},
		},
	}, agentsv1.AutomationTriggerType_AUTOMATION_TRIGGER_TYPE_MANUAL, `{}`)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if run.GetStatus() != agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_FAILED {
		t.Fatalf("status = %s, want failed", run.GetStatus())
	}
	steps, _ := stepRepo.ListByRun(ctx, "ws1", run.GetId())
	if len(steps) != 1 {
		t.Fatalf("steps executed = %d, want 1", len(steps))
	}
}

func TestEngineStepRetrySuccessAndExhaustion(t *testing.T) {
	ctx := context.Background()
	engine, _, runnerSvc, stepRepo := newMinimalEngine()
	runnerSvc.errs = []error{errors.New("temporary"), nil}
	run, err := engine.Execute(ctx, &agentsv1.Automation{
		Name:        "retry-success",
		Enabled:     true,
		WorkspaceId: "ws1",
		Steps:       []*agentsv1.AutomationStep{{Name: "summarize", Type: agentsv1.AutomationStepType_AUTOMATION_STEP_TYPE_INVOKE_AGENT, InvokeAgent: &agentsv1.AutomationInvokeAgentStep{AgentName: "agent1"}, Policy: &agentsv1.AutomationPolicy{Retry: &agentsv1.AutomationRetryPolicy{MaxAttempts: 1}}}},
	}, agentsv1.AutomationTriggerType_AUTOMATION_TRIGGER_TYPE_MANUAL, `{}`)
	if err != nil {
		t.Fatalf("Execute retry success: %v", err)
	}
	if run.GetStatus() != agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_SUCCEEDED {
		t.Fatalf("status = %s, want succeeded", run.GetStatus())
	}
	steps, _ := stepRepo.ListByRun(ctx, "ws1", run.GetId())
	if steps[0].GetAttemptCount() != 2 {
		t.Fatalf("attempt count = %d, want 2", steps[0].GetAttemptCount())
	}

	engine, _, runnerSvc, stepRepo = newMinimalEngine()
	runnerSvc.errs = []error{errors.New("one"), errors.New("two")}
	run, err = engine.Execute(ctx, &agentsv1.Automation{
		Name:        "retry-exhausted",
		Enabled:     true,
		WorkspaceId: "ws1",
		Steps:       []*agentsv1.AutomationStep{{Name: "summarize", Type: agentsv1.AutomationStepType_AUTOMATION_STEP_TYPE_INVOKE_AGENT, InvokeAgent: &agentsv1.AutomationInvokeAgentStep{AgentName: "agent1"}, Policy: &agentsv1.AutomationPolicy{Retry: &agentsv1.AutomationRetryPolicy{MaxAttempts: 1}}}},
	}, agentsv1.AutomationTriggerType_AUTOMATION_TRIGGER_TYPE_MANUAL, `{}`)
	if err != nil {
		t.Fatalf("Execute retry exhausted: %v", err)
	}
	if run.GetStatus() != agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_FAILED {
		t.Fatalf("status = %s, want failed", run.GetStatus())
	}
	steps, _ = stepRepo.ListByRun(ctx, "ws1", run.GetId())
	if steps[0].GetAttemptCount() != 2 {
		t.Fatalf("attempt count = %d, want 2", steps[0].GetAttemptCount())
	}
}

func TestEngineTimeoutAndOutputTruncation(t *testing.T) {
	ctx := context.Background()
	engine, _, runnerSvc, _ := newMinimalEngine()
	runnerSvc.block = make(chan struct{})
	run, err := engine.Execute(ctx, &agentsv1.Automation{
		Name:        "timeout",
		Enabled:     true,
		WorkspaceId: "ws1",
		Steps:       []*agentsv1.AutomationStep{{Name: "slow", Type: agentsv1.AutomationStepType_AUTOMATION_STEP_TYPE_INVOKE_AGENT, InvokeAgent: &agentsv1.AutomationInvokeAgentStep{AgentName: "agent1"}, Policy: &agentsv1.AutomationPolicy{Timeout: durationpb.New(10 * time.Millisecond)}}},
	}, agentsv1.AutomationTriggerType_AUTOMATION_TRIGGER_TYPE_MANUAL, `{}`)
	if err != nil {
		t.Fatalf("Execute timeout: %v", err)
	}
	if run.GetStatus() != agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_FAILED {
		t.Fatalf("status = %s, want failed", run.GetStatus())
	}

	engine, _, runnerSvc, stepRepo := newMinimalEngine()
	runnerSvc.outputs = []string{strings.Repeat("x", 200)}
	run, err = engine.Execute(ctx, &agentsv1.Automation{
		Name:        "truncate",
		Enabled:     true,
		WorkspaceId: "ws1",
		Steps:       []*agentsv1.AutomationStep{{Name: "big", Type: agentsv1.AutomationStepType_AUTOMATION_STEP_TYPE_INVOKE_AGENT, InvokeAgent: &agentsv1.AutomationInvokeAgentStep{AgentName: "agent1"}, Policy: &agentsv1.AutomationPolicy{MaxOutputBytes: 64}}},
	}, agentsv1.AutomationTriggerType_AUTOMATION_TRIGGER_TYPE_MANUAL, `{}`)
	if err != nil {
		t.Fatalf("Execute truncate: %v", err)
	}
	steps, _ := stepRepo.ListByRun(ctx, "ws1", run.GetId())
	if !steps[0].GetTruncated() {
		t.Fatal("expected truncated step output")
	}
	if len(steps[0].GetOutputJson()) > 64 {
		t.Fatalf("output len = %d, want <=64", len(steps[0].GetOutputJson()))
	}
}

func newMinimalEngine() (*Engine, *MemoryRunRepo, *engineRunner, *MemoryStepRunRepo) {
	runRepo := NewMemoryRunRepo()
	stepRepo := NewMemoryStepRunRepo()
	runnerSvc := &engineRunner{outputs: []string{"ok"}}
	engine := NewEngine(NewMemoryDefinitionRepo(), runRepo, stepRepo, EngineOptions{
		Runner:     runnerSvc,
		HTTPClient: &engineHTTPClient{statuses: []int{http.StatusInternalServerError}},
	})
	return engine, runRepo, runnerSvc, stepRepo
}
