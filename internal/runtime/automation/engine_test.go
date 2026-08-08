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
	pending []runner.PendingInput
	block   chan struct{}
	calls   int
}

func (r *engineRunner) ResolveAgentRef(workspaceID, agentID string) (string, bool) {
	// The fake registry maps agent1's id to its name; unknown ids miss.
	if workspaceID == "ws1" && agentID == "agent1-id" {
		return "agent1", true
	}
	return "", false
}

func (r *engineRunner) RunTurnSSE(ctx context.Context, _ string, _ []*genai.Part, _ string, _ *agentsv1.ContextInfo, _ runner.EventCallback, _ runner.CompactionCallback) (*runner.TurnResult, error) {
	r.mu.Lock()
	r.calls++
	call := r.calls
	r.mu.Unlock()
	if r.block != nil {
		select {
		case <-r.block:
		case <-ctx.Done():
			return &runner.TurnResult{}, ctx.Err()
		}
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.errs) > 0 {
		err := r.errs[0]
		r.errs = r.errs[1:]
		if err != nil {
			return &runner.TurnResult{}, err
		}
	}
	turn := &runner.TurnResult{Pending: r.pending}
	switch {
	case len(r.outputs) >= call:
		turn.Output = r.outputs[call-1]
	case len(r.outputs) > 0:
		turn.Output = r.outputs[len(r.outputs)-1]
	default:
		turn.Output = "ok"
	}
	return turn, nil
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
			{Name: "summarize", Type: agentsv1.AutomationStepType_AUTOMATION_STEP_TYPE_INVOKE_AGENT, InvokeAgent: &agentsv1.AutomationInvokeAgentStep{AgentId: "agent1-id", Input: "summarize"}},
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

func TestEngineInvokeAgentPauseRecordsWaitingInput(t *testing.T) {
	ctx := context.Background()
	engine, runRepo, runnerSvc, stepRepo := newMinimalEngine()
	runnerSvc.outputs = []string{"Approve deploy to prod? (yes/no)"}
	runnerSvc.pending = []runner.PendingInput{{InterruptID: "call-1", Question: "Approve deploy to prod? (yes/no)"}}

	run, err := engine.Execute(ctx, &agentsv1.Automation{
		Name:        "approval",
		Enabled:     true,
		WorkspaceId: "ws1",
		Steps: []*agentsv1.AutomationStep{
			{Name: "ask", Type: agentsv1.AutomationStepType_AUTOMATION_STEP_TYPE_INVOKE_AGENT, InvokeAgent: &agentsv1.AutomationInvokeAgentStep{AgentId: "agent1-id", Input: "deploy"}},
			// A later step must NOT run while the workflow is paused (Option A).
			{Name: "after", Type: agentsv1.AutomationStepType_AUTOMATION_STEP_TYPE_INVOKE_AGENT, InvokeAgent: &agentsv1.AutomationInvokeAgentStep{AgentId: "agent1-id"}},
		},
	}, agentsv1.AutomationTriggerType_AUTOMATION_TRIGGER_TYPE_MANUAL, `{}`)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if run.GetStatus() != agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_WAITING_INPUT {
		t.Fatalf("status = %s, want WAITING_INPUT", run.GetStatus())
	}
	if got, want := run.GetSessionAppName(), "automation:approval"; got != want {
		t.Fatalf("session_app_name = %q, want %q", got, want)
	}
	if got, want := run.GetSessionUserId(), "automation:ws1"; got != want {
		t.Fatalf("session_user_id = %q, want %q", got, want)
	}
	if got, want := run.GetSessionId(), "automation:"+run.GetId(); got != want {
		t.Fatalf("session_id = %q, want %q", got, want)
	}
	if run.GetAgentName() != "agent1" {
		t.Fatalf("agent_name = %q, want agent1", run.GetAgentName())
	}
	if run.GetFinishedAt() != nil {
		t.Fatal("finished_at set, want nil for a waiting run")
	}

	// The run must be discoverable by its session coordinates for resume.
	waiting, err := runRepo.ListWaitingBySession(ctx, run.GetSessionAppName(), run.GetSessionUserId(), run.GetSessionId())
	if err != nil {
		t.Fatalf("ListWaitingBySession: %v", err)
	}
	if len(waiting) != 1 || waiting[0].GetId() != run.GetId() {
		t.Fatalf("waiting runs = %v, want [%s]", ids(waiting), run.GetId())
	}

	// Only the paused step ran; the later step is untouched.
	steps, _ := stepRepo.ListByRun(ctx, "ws1", run.GetId())
	if len(steps) != 1 || steps[0].GetStepName() != "ask" {
		t.Fatalf("steps = %v, want only [ask]", stepNames(steps))
	}
}

func stepNames(steps []*agentsv1.AutomationStepRun) []string {
	out := make([]string, len(steps))
	for i, s := range steps {
		out[i] = s.GetStepName()
	}
	return out
}

// pauseRun drives an automation to a WAITING_INPUT state and returns the run.
func pauseRun(t *testing.T, ctx context.Context, engine *Engine, runnerSvc *engineRunner) *agentsv1.AutomationRun {
	t.Helper()
	runnerSvc.outputs = []string{"Approve? (yes/no)"}
	runnerSvc.pending = []runner.PendingInput{{InterruptID: "call-1", Question: "Approve? (yes/no)"}}
	run, err := engine.Execute(ctx, &agentsv1.Automation{
		Name:        "approval",
		Enabled:     true,
		WorkspaceId: "ws1",
		Steps: []*agentsv1.AutomationStep{
			{Name: "ask", Type: agentsv1.AutomationStepType_AUTOMATION_STEP_TYPE_INVOKE_AGENT, InvokeAgent: &agentsv1.AutomationInvokeAgentStep{AgentId: "agent1-id", Input: "deploy"}},
		},
	}, agentsv1.AutomationTriggerType_AUTOMATION_TRIGGER_TYPE_MANUAL, `{}`)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if run.GetStatus() != agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_WAITING_INPUT {
		t.Fatalf("setup status = %s, want WAITING_INPUT", run.GetStatus())
	}
	return run
}

func TestEngineHandleTurnResumesWaitingRun(t *testing.T) {
	ctx := context.Background()
	engine, runRepo, runnerSvc, _ := newMinimalEngine()
	run := pauseRun(t, ctx, engine, runnerSvc)

	ctxInfo := &agentsv1.ContextInfo{
		ChannelName: run.GetSessionAppName(),
		UserId:      run.GetSessionUserId(),
		SessionId:   run.GetSessionId(),
	}

	// A turn that is still interrupted must leave the run waiting.
	engine.HandleTurn(ctxInfo, &runner.TurnResult{Output: "still asking", Pending: []runner.PendingInput{{InterruptID: "call-1"}}}, nil)
	if got, _ := runRepo.Get(ctx, "ws1", run.GetId()); got.GetStatus() != agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_WAITING_INPUT {
		t.Fatalf("after interrupted turn: status = %s, want WAITING_INPUT", got.GetStatus())
	}

	// A completed turn on the run's session finalizes it to SUCCEEDED.
	engine.HandleTurn(ctxInfo, &runner.TurnResult{Output: "Deploy approved."}, nil)
	got, err := runRepo.Get(ctx, "ws1", run.GetId())
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.GetStatus() != agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_SUCCEEDED {
		t.Fatalf("status = %s, want SUCCEEDED", got.GetStatus())
	}
	if got.GetFinishedAt() == nil {
		t.Fatal("finished_at nil, want set on terminal run")
	}
	// The run must no longer be discoverable as waiting.
	waiting, _ := runRepo.ListWaitingBySession(ctx, run.GetSessionAppName(), run.GetSessionUserId(), run.GetSessionId())
	if len(waiting) != 0 {
		t.Fatalf("still waiting = %v, want none", ids(waiting))
	}
}

func TestEngineHandleTurnIgnoresNonAutomationSession(t *testing.T) {
	ctx := context.Background()
	engine, runRepo, runnerSvc, _ := newMinimalEngine()
	run := pauseRun(t, ctx, engine, runnerSvc)

	// A reply on an ordinary chat session must not touch automation runs.
	engine.HandleTurn(&agentsv1.ContextInfo{ChannelName: "telegram", UserId: "u1", SessionId: "chat-42"}, &runner.TurnResult{Output: "hi"}, nil)
	if got, _ := runRepo.Get(ctx, "ws1", run.GetId()); got.GetStatus() != agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_WAITING_INPUT {
		t.Fatalf("status = %s, want WAITING_INPUT untouched", got.GetStatus())
	}
}

func TestEngineHandleSessionDeletedCancelsWaitingRun(t *testing.T) {
	ctx := context.Background()
	engine, runRepo, runnerSvc, _ := newMinimalEngine()
	run := pauseRun(t, ctx, engine, runnerSvc)

	engine.HandleSessionDeleted(run.GetSessionAppName(), run.GetSessionUserId(), run.GetSessionId())

	got, err := runRepo.Get(ctx, "ws1", run.GetId())
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.GetStatus() != agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_CANCELLED {
		t.Fatalf("status = %s, want CANCELLED", got.GetStatus())
	}
	if got.GetError() == "" {
		t.Fatal("error empty, want an abandonment reason")
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
		Steps: []*agentsv1.AutomationStep{{Name: "summarize", Type: agentsv1.AutomationStepType_AUTOMATION_STEP_TYPE_INVOKE_AGENT, InvokeAgent: &agentsv1.AutomationInvokeAgentStep{AgentId: "agent1-id"}}},
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
			{Name: "never", Type: agentsv1.AutomationStepType_AUTOMATION_STEP_TYPE_INVOKE_AGENT, InvokeAgent: &agentsv1.AutomationInvokeAgentStep{AgentId: "agent1-id"}},
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
		Steps:       []*agentsv1.AutomationStep{{Name: "summarize", Type: agentsv1.AutomationStepType_AUTOMATION_STEP_TYPE_INVOKE_AGENT, InvokeAgent: &agentsv1.AutomationInvokeAgentStep{AgentId: "agent1-id"}, Policy: &agentsv1.AutomationPolicy{Retry: &agentsv1.AutomationRetryPolicy{MaxAttempts: 1}}}},
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
		Steps:       []*agentsv1.AutomationStep{{Name: "summarize", Type: agentsv1.AutomationStepType_AUTOMATION_STEP_TYPE_INVOKE_AGENT, InvokeAgent: &agentsv1.AutomationInvokeAgentStep{AgentId: "agent1-id"}, Policy: &agentsv1.AutomationPolicy{Retry: &agentsv1.AutomationRetryPolicy{MaxAttempts: 1}}}},
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
		Steps:       []*agentsv1.AutomationStep{{Name: "slow", Type: agentsv1.AutomationStepType_AUTOMATION_STEP_TYPE_INVOKE_AGENT, InvokeAgent: &agentsv1.AutomationInvokeAgentStep{AgentId: "agent1-id"}, Policy: &agentsv1.AutomationPolicy{Timeout: durationpb.New(10 * time.Millisecond)}}},
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
		Steps:       []*agentsv1.AutomationStep{{Name: "big", Type: agentsv1.AutomationStepType_AUTOMATION_STEP_TYPE_INVOKE_AGENT, InvokeAgent: &agentsv1.AutomationInvokeAgentStep{AgentId: "agent1-id"}, Policy: &agentsv1.AutomationPolicy{MaxOutputBytes: 64}}},
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

func TestEngineInvokeAgentResolvesAgentID(t *testing.T) {
	ctx := context.Background()
	defRepo := NewMemoryDefinitionRepo()
	runRepo := NewMemoryRunRepo()
	stepRepo := NewMemoryStepRunRepo()
	runnerSvc := &engineRunner{outputs: []string{"agent output"}}
	engine := NewEngine(defRepo, runRepo, stepRepo, EngineOptions{Runner: runnerSvc})

	automation := &agentsv1.Automation{
		Name:        "by-id",
		Enabled:     true,
		WorkspaceId: "ws1",
		Trigger:     &agentsv1.AutomationTrigger{Type: agentsv1.AutomationTriggerType_AUTOMATION_TRIGGER_TYPE_MANUAL},
		Steps: []*agentsv1.AutomationStep{
			{Name: "summarize", Type: agentsv1.AutomationStepType_AUTOMATION_STEP_TYPE_INVOKE_AGENT, InvokeAgent: &agentsv1.AutomationInvokeAgentStep{AgentId: "agent1-id", Input: "go"}},
		},
	}
	if err := defRepo.Create(ctx, automation); err != nil {
		t.Fatalf("create automation: %v", err)
	}

	run, err := engine.RunNow(ctx, "ws1", "by-id", "")
	if err != nil {
		t.Fatalf("RunNow: %v", err)
	}
	if run.GetStatus() != agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_SUCCEEDED {
		t.Fatalf("status = %s, want succeeded; err=%s", run.GetStatus(), run.GetError())
	}
	if runnerSvc.calls != 1 {
		t.Fatalf("runner calls = %d, want 1 (resolved from agent_id)", runnerSvc.calls)
	}
}

func TestEngineInvokeAgentUnknownAgentIDFails(t *testing.T) {
	ctx := context.Background()
	defRepo := NewMemoryDefinitionRepo()
	runRepo := NewMemoryRunRepo()
	stepRepo := NewMemoryStepRunRepo()
	runnerSvc := &engineRunner{outputs: []string{"agent output"}}
	engine := NewEngine(defRepo, runRepo, stepRepo, EngineOptions{Runner: runnerSvc})

	automation := &agentsv1.Automation{
		Name:        "ghost-id",
		Enabled:     true,
		WorkspaceId: "ws1",
		Trigger:     &agentsv1.AutomationTrigger{Type: agentsv1.AutomationTriggerType_AUTOMATION_TRIGGER_TYPE_MANUAL},
		Steps: []*agentsv1.AutomationStep{
			// agent_name is valid but must not be used: agent_id wins.
			{Name: "summarize", Type: agentsv1.AutomationStepType_AUTOMATION_STEP_TYPE_INVOKE_AGENT, InvokeAgent: &agentsv1.AutomationInvokeAgentStep{AgentId: "ghost", Input: "go"}},
		},
	}
	if err := defRepo.Create(ctx, automation); err != nil {
		t.Fatalf("create automation: %v", err)
	}

	run, err := engine.RunNow(ctx, "ws1", "ghost-id", "")
	if err != nil {
		t.Fatalf("RunNow: %v", err)
	}
	if run.GetStatus() != agentsv1.AutomationRunStatus_AUTOMATION_RUN_STATUS_FAILED {
		t.Fatalf("status = %s, want failed", run.GetStatus())
	}
	if runnerSvc.calls != 0 {
		t.Fatalf("runner calls = %d, want 0 (unknown agent_id must not fall back)", runnerSvc.calls)
	}
}
