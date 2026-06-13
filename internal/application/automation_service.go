package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"butterfly.orx.me/core/log"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	runtimeautomation "go.orx.me/apps/butter/internal/runtime/automation"
	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// AutomationServiceServer implements workspace-scoped automation CRUD,
// manual execution, and run history APIs.
type AutomationServiceServer struct {
	mu        sync.RWMutex
	defRepo   runtimeautomation.DefinitionRepo
	runRepo   runtimeautomation.RunRepo
	stepRepo  runtimeautomation.StepRunRepo
	engine    *runtimeautomation.Engine
	scheduler *runtimeautomation.Scheduler
}

func NewAutomationServiceServer() *AutomationServiceServer {
	return &AutomationServiceServer{}
}

func (s *AutomationServiceServer) SetRepos(defRepo runtimeautomation.DefinitionRepo, runRepo runtimeautomation.RunRepo, stepRepo runtimeautomation.StepRunRepo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.defRepo = defRepo
	s.runRepo = runRepo
	s.stepRepo = stepRepo
}

func (s *AutomationServiceServer) SetEngine(engine *runtimeautomation.Engine) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.engine = engine
}

func (s *AutomationServiceServer) SetScheduler(scheduler *runtimeautomation.Scheduler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.scheduler = scheduler
}

func (s *AutomationServiceServer) deps() (runtimeautomation.DefinitionRepo, runtimeautomation.RunRepo, runtimeautomation.StepRunRepo, *runtimeautomation.Engine, *runtimeautomation.Scheduler) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.defRepo, s.runRepo, s.stepRepo, s.engine, s.scheduler
}

func automationNotInitialized() error {
	return connect.NewError(connect.CodeFailedPrecondition, errors.New("automation service not initialized"))
}

func (s *AutomationServiceServer) ListAutomations(ctx context.Context, _ *connect.Request[agentsv1.ListAutomationsRequest]) (*connect.Response[agentsv1.ListAutomationsResponse], error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	defRepo, _, _, _, _ := s.deps()
	if defRepo == nil {
		return nil, automationNotInitialized()
	}
	automations, err := defRepo.List(ctx, wsID)
	if err != nil {
		return nil, automationConnectErr(err)
	}
	return connect.NewResponse(&agentsv1.ListAutomationsResponse{Automations: automations}), nil
}

func (s *AutomationServiceServer) GetAutomation(ctx context.Context, req *connect.Request[agentsv1.GetAutomationRequest]) (*connect.Response[agentsv1.GetAutomationResponse], error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	defRepo, _, _, _, _ := s.deps()
	if defRepo == nil {
		return nil, automationNotInitialized()
	}
	a, err := defRepo.Get(ctx, wsID, strings.TrimSpace(req.Msg.GetName()))
	if err != nil {
		return nil, automationConnectErr(err)
	}
	return connect.NewResponse(&agentsv1.GetAutomationResponse{Automation: a}), nil
}

func (s *AutomationServiceServer) CreateAutomation(ctx context.Context, req *connect.Request[agentsv1.CreateAutomationRequest]) (*connect.Response[agentsv1.CreateAutomationResponse], error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	defRepo, _, _, _, scheduler := s.deps()
	if defRepo == nil {
		return nil, automationNotInitialized()
	}
	a := req.Msg.GetAutomation()
	if a == nil {
		return nil, connectx.RequiredArgument("automation")
	}
	a.WorkspaceId = wsID
	now := timestamppb.New(time.Now().UTC())
	a.CreatedAt = now
	a.UpdatedAt = now
	if err := validateAutomation(a); err != nil {
		return nil, err
	}
	if err := defRepo.Create(ctx, a); err != nil {
		return nil, automationConnectErr(err)
	}
	if scheduler != nil {
		if err := scheduler.Reschedule(a); err != nil {
			log.FromContext(ctx).Error("failed to schedule automation after create", "workspace_id", wsID, "automation", a.GetName(), "err", err)
			return nil, connectx.InternalWith(err)
		}
	}
	return connect.NewResponse(&agentsv1.CreateAutomationResponse{Automation: a}), nil
}

func (s *AutomationServiceServer) UpdateAutomation(ctx context.Context, req *connect.Request[agentsv1.UpdateAutomationRequest]) (*connect.Response[agentsv1.UpdateAutomationResponse], error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	defRepo, _, _, _, scheduler := s.deps()
	if defRepo == nil {
		return nil, automationNotInitialized()
	}
	a := req.Msg.GetAutomation()
	if a == nil {
		return nil, connectx.RequiredArgument("automation")
	}
	a.WorkspaceId = wsID
	existing, err := defRepo.Get(ctx, wsID, a.GetName())
	if err != nil {
		return nil, automationConnectErr(err)
	}
	if a.GetCreatedAt() == nil {
		a.CreatedAt = existing.GetCreatedAt()
	}
	a.UpdatedAt = timestamppb.New(time.Now().UTC())
	if err := validateAutomation(a); err != nil {
		return nil, err
	}
	if err := defRepo.Update(ctx, a); err != nil {
		return nil, automationConnectErr(err)
	}
	if scheduler != nil {
		if err := scheduler.Reschedule(a); err != nil {
			log.FromContext(ctx).Error("failed to reschedule automation after update", "workspace_id", wsID, "automation", a.GetName(), "err", err)
			return nil, connectx.InternalWith(err)
		}
	}
	return connect.NewResponse(&agentsv1.UpdateAutomationResponse{Automation: a}), nil
}

func (s *AutomationServiceServer) DeleteAutomation(ctx context.Context, req *connect.Request[agentsv1.DeleteAutomationRequest]) (*connect.Response[agentsv1.DeleteAutomationResponse], error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	defRepo, _, _, _, scheduler := s.deps()
	if defRepo == nil {
		return nil, automationNotInitialized()
	}
	name := strings.TrimSpace(req.Msg.GetName())
	a, err := defRepo.Delete(ctx, wsID, name)
	if err != nil {
		return nil, automationConnectErr(err)
	}
	if scheduler != nil {
		scheduler.Unregister(wsID, name)
	}
	return connect.NewResponse(&agentsv1.DeleteAutomationResponse{Automation: a}), nil
}

func (s *AutomationServiceServer) RunAutomationNow(ctx context.Context, req *connect.Request[agentsv1.RunAutomationNowRequest]) (*connect.Response[agentsv1.RunAutomationNowResponse], error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	_, _, _, engine, _ := s.deps()
	if engine == nil {
		return nil, automationNotInitialized()
	}
	if raw := strings.TrimSpace(req.Msg.GetTriggerPayloadJson()); raw != "" && !json.Valid([]byte(raw)) {
		return nil, connectx.InvalidArgument("trigger_payload_json", "must be valid JSON")
	}
	run, err := engine.RunNow(ctx, wsID, strings.TrimSpace(req.Msg.GetName()), req.Msg.GetTriggerPayloadJson())
	if err != nil {
		return nil, automationConnectErr(err)
	}
	return connect.NewResponse(&agentsv1.RunAutomationNowResponse{Run: run}), nil
}

func (s *AutomationServiceServer) ListAutomationRuns(ctx context.Context, req *connect.Request[agentsv1.ListAutomationRunsRequest]) (*connect.Response[agentsv1.ListAutomationRunsResponse], error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	_, runRepo, _, _, _ := s.deps()
	if runRepo == nil {
		return nil, automationNotInitialized()
	}
	runs, next, err := runRepo.List(ctx, wsID, strings.TrimSpace(req.Msg.GetAutomationName()), req.Msg.GetPageSize(), req.Msg.GetPageToken())
	if err != nil {
		return nil, automationConnectErr(err)
	}
	return connect.NewResponse(&agentsv1.ListAutomationRunsResponse{Runs: runs, NextPageToken: next}), nil
}

func (s *AutomationServiceServer) GetAutomationRun(ctx context.Context, req *connect.Request[agentsv1.GetAutomationRunRequest]) (*connect.Response[agentsv1.GetAutomationRunResponse], error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	_, runRepo, _, _, _ := s.deps()
	if runRepo == nil {
		return nil, automationNotInitialized()
	}
	run, err := runRepo.Get(ctx, wsID, strings.TrimSpace(req.Msg.GetId()))
	if err != nil {
		return nil, automationConnectErr(err)
	}
	return connect.NewResponse(&agentsv1.GetAutomationRunResponse{Run: run}), nil
}

func (s *AutomationServiceServer) ListAutomationStepRuns(ctx context.Context, req *connect.Request[agentsv1.ListAutomationStepRunsRequest]) (*connect.Response[agentsv1.ListAutomationStepRunsResponse], error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	_, _, stepRepo, _, _ := s.deps()
	if stepRepo == nil {
		return nil, automationNotInitialized()
	}
	stepRuns, err := stepRepo.ListByRun(ctx, wsID, strings.TrimSpace(req.Msg.GetRunId()))
	if err != nil {
		return nil, automationConnectErr(err)
	}
	return connect.NewResponse(&agentsv1.ListAutomationStepRunsResponse{StepRuns: stepRuns}), nil
}

func validateAutomation(a *agentsv1.Automation) error {
	if strings.TrimSpace(a.GetName()) == "" {
		return connectx.RequiredArgument("automation.name")
	}
	if a.GetTrigger() == nil || a.GetTrigger().GetType() == agentsv1.AutomationTriggerType_AUTOMATION_TRIGGER_TYPE_UNSPECIFIED {
		return connectx.RequiredArgument("automation.trigger.type")
	}
	if a.GetTrigger().GetType() == agentsv1.AutomationTriggerType_AUTOMATION_TRIGGER_TYPE_SCHEDULE {
		if strings.TrimSpace(a.GetTrigger().GetSchedule().GetSchedule()) == "" {
			return connectx.RequiredArgument("automation.trigger.schedule.schedule")
		}
	}
	if len(a.GetSteps()) == 0 {
		return connectx.InvalidArgument("automation.steps", "must contain at least one step")
	}
	for i, cond := range a.GetConditions() {
		if strings.TrimSpace(cond.GetSelector()) == "" {
			return connectx.RequiredArgument(fmt.Sprintf("automation.conditions[%d].selector", i))
		}
		switch cond.GetOperator() {
		case agentsv1.AutomationConditionOperator_AUTOMATION_CONDITION_OPERATOR_EQUALS,
			agentsv1.AutomationConditionOperator_AUTOMATION_CONDITION_OPERATOR_NOT_EQUALS,
			agentsv1.AutomationConditionOperator_AUTOMATION_CONDITION_OPERATOR_CONTAINS,
			agentsv1.AutomationConditionOperator_AUTOMATION_CONDITION_OPERATOR_REGEX_MATCH,
			agentsv1.AutomationConditionOperator_AUTOMATION_CONDITION_OPERATOR_EXISTS,
			agentsv1.AutomationConditionOperator_AUTOMATION_CONDITION_OPERATOR_NOT_EXISTS:
		default:
			return connectx.InvalidArgument(fmt.Sprintf("automation.conditions[%d].operator", i), "is unsupported")
		}
	}
	seenSteps := make(map[string]struct{}, len(a.GetSteps()))
	for i, step := range a.GetSteps() {
		if strings.TrimSpace(step.GetName()) == "" {
			return connectx.RequiredArgument(fmt.Sprintf("automation.steps[%d].name", i))
		}
		if _, ok := seenSteps[step.GetName()]; ok {
			return connectx.InvalidArgument(fmt.Sprintf("automation.steps[%d].name", i), "must be unique")
		}
		seenSteps[step.GetName()] = struct{}{}
		if err := validateAutomationPolicy(step.GetPolicy(), fmt.Sprintf("automation.steps[%d].policy", i)); err != nil {
			return err
		}
		switch step.GetType() {
		case agentsv1.AutomationStepType_AUTOMATION_STEP_TYPE_INVOKE_AGENT:
			if strings.TrimSpace(step.GetInvokeAgent().GetAgentName()) == "" {
				return connectx.RequiredArgument(fmt.Sprintf("automation.steps[%d].invoke_agent.agent_name", i))
			}
		case agentsv1.AutomationStepType_AUTOMATION_STEP_TYPE_CALL_WEBHOOK:
			if strings.TrimSpace(step.GetCallWebhook().GetUrl()) == "" {
				return connectx.RequiredArgument(fmt.Sprintf("automation.steps[%d].call_webhook.url", i))
			}
			if method := strings.TrimSpace(step.GetCallWebhook().GetMethod()); method != "" {
				switch strings.ToUpper(method) {
				case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
				default:
					return connectx.InvalidArgument(fmt.Sprintf("automation.steps[%d].call_webhook.method", i), "is unsupported")
				}
			}
			if payload := strings.TrimSpace(step.GetCallWebhook().GetPayloadJson()); payload != "" && !json.Valid([]byte(payload)) {
				return connectx.InvalidArgument(fmt.Sprintf("automation.steps[%d].call_webhook.payload_json", i), "must be valid JSON")
			}
		case agentsv1.AutomationStepType_AUTOMATION_STEP_TYPE_SEND_NOTIFY_GROUP:
			if strings.TrimSpace(step.GetSendNotifyGroup().GetNotifyGroupName()) == "" {
				return connectx.RequiredArgument(fmt.Sprintf("automation.steps[%d].send_notify_group.notify_group_name", i))
			}
		case agentsv1.AutomationStepType_AUTOMATION_STEP_TYPE_CREATE_FORUM_POST:
			if strings.TrimSpace(step.GetCreateForumPost().GetThreadId()) == "" {
				return connectx.RequiredArgument(fmt.Sprintf("automation.steps[%d].create_forum_post.thread_id", i))
			}
			if strings.TrimSpace(step.GetCreateForumPost().GetBody()) == "" {
				return connectx.RequiredArgument(fmt.Sprintf("automation.steps[%d].create_forum_post.body", i))
			}
		default:
			return connectx.InvalidArgument(fmt.Sprintf("automation.steps[%d].type", i), "is unsupported")
		}
	}
	return validateAutomationPolicy(a.GetPolicy(), "automation.policy")
}

func validateAutomationPolicy(policy *agentsv1.AutomationPolicy, prefix string) error {
	if policy == nil {
		return nil
	}
	if policy.GetTimeout() != nil && policy.GetTimeout().AsDuration() < 0 {
		return connectx.InvalidArgument(prefix+".timeout", "must be non-negative")
	}
	if policy.GetRetry() != nil {
		if policy.GetRetry().GetMaxAttempts() < 0 {
			return connectx.InvalidArgument(prefix+".retry.max_attempts", "must be non-negative")
		}
		if policy.GetRetry().GetBackoff() != nil && policy.GetRetry().GetBackoff().AsDuration() < 0 {
			return connectx.InvalidArgument(prefix+".retry.backoff", "must be non-negative")
		}
	}
	switch policy.GetConcurrency() {
	case agentsv1.AutomationConcurrencyPolicy_AUTOMATION_CONCURRENCY_POLICY_UNSPECIFIED,
		agentsv1.AutomationConcurrencyPolicy_AUTOMATION_CONCURRENCY_POLICY_SKIP,
		agentsv1.AutomationConcurrencyPolicy_AUTOMATION_CONCURRENCY_POLICY_QUEUE,
		agentsv1.AutomationConcurrencyPolicy_AUTOMATION_CONCURRENCY_POLICY_REPLACE,
		agentsv1.AutomationConcurrencyPolicy_AUTOMATION_CONCURRENCY_POLICY_ALLOW:
	default:
		return connectx.InvalidArgument(prefix+".concurrency", "is unsupported")
	}
	if policy.GetMaxOutputBytes() < 0 {
		return connectx.InvalidArgument(prefix+".max_output_bytes", "must be non-negative")
	}
	switch policy.GetNotifyOn() {
	case agentsv1.AutomationNotifyOn_AUTOMATION_NOTIFY_ON_UNSPECIFIED,
		agentsv1.AutomationNotifyOn_AUTOMATION_NOTIFY_ON_ALWAYS,
		agentsv1.AutomationNotifyOn_AUTOMATION_NOTIFY_ON_FAILURE,
		agentsv1.AutomationNotifyOn_AUTOMATION_NOTIFY_ON_SUCCESS:
	default:
		return connectx.InvalidArgument(prefix+".notify_on", "is unsupported")
	}
	return nil
}

func automationConnectErr(err error) error {
	var cerr *connect.Error
	if errors.As(err, &cerr) {
		return cerr
	}
	switch {
	case errors.Is(err, runtimeautomation.ErrAutomationNotFound), errors.Is(err, runtimeautomation.ErrRunNotFound), errors.Is(err, runtimeautomation.ErrStepRunNotFound):
		return connectx.NotFound(err.Error())
	case errors.Is(err, runtimeautomation.ErrAutomationAlreadyExists):
		return connect.NewError(connect.CodeAlreadyExists, errors.New(err.Error()))
	case errors.Is(err, runtimeautomation.ErrAutomationDisabled):
		return connect.NewError(connect.CodeFailedPrecondition, errors.New(err.Error()))
	default:
		return connectx.InternalWith(err)
	}
}
