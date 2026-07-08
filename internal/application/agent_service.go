package application

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	"butterfly.orx.me/core/log"
	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/genai"

	internalagent "go.orx.me/apps/butter/internal/agent"
	"go.orx.me/apps/butter/internal/repo/auth"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	"go.orx.me/apps/butter/internal/repo/invocation"
	"go.orx.me/apps/butter/internal/runtime/runner"
	"go.orx.me/apps/butter/internal/transport/connectx"
	"go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// maxInvokeAgentInputBytes caps the size of a single InvokeAgent input to
// protect the runner and session storage from oversized requests.
const maxInvokeAgentInputBytes = 1 << 20 // 1 MiB

type AgentServiceServer struct {
	repo      configrepo.AgentRepository
	runtime   ConfigRuntime
	runnerSvc *runner.Service
	invRepo   invocation.Repository
}

func NewAgentServiceServer(repo configrepo.AgentRepository) *AgentServiceServer {
	return &AgentServiceServer{repo: repo}
}

func (s *AgentServiceServer) SetRuntime(runtime ConfigRuntime) {
	s.runtime = runtime
}

// SetRunnerService wires the runner so InvokeAgent can execute agents.
func (s *AgentServiceServer) SetRunnerService(svc *runner.Service) {
	s.runnerSvc = svc
}

// SetInvocationRepo wires the invocation repository used by
// ListAgentInvocations.
func (s *AgentServiceServer) SetInvocationRepo(repo invocation.Repository) {
	s.invRepo = repo
}

func (s *AgentServiceServer) ListAgents(ctx context.Context, req *connect.Request[agentsv1.ListAgentsRequest]) (*connect.Response[agentsv1.ListAgentsResponse], error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	agents, err := s.repo.ListAgents(ctx, wsID)
	if err != nil {
		return nil, toConnectError(err)
	}

	sort.SliceStable(agents, func(i, j int) bool {
		return agents[i].GetName() < agents[j].GetName()
	})

	total := int32(len(agents))
	pageSize := req.Msg.GetPageSize()
	if pageSize <= 0 {
		pageSize = 50
	}
	if pageSize > 200 {
		pageSize = 200
	}
	offset := 0
	if token := req.Msg.GetPageToken(); token != "" {
		if raw, err := base64.StdEncoding.DecodeString(token); err == nil {
			if n, err := strconv.Atoi(string(raw)); err == nil && n >= 0 {
				offset = n
			}
		}
	}
	if offset >= len(agents) {
		return connect.NewResponse(&agentsv1.ListAgentsResponse{Total: total}), nil
	}
	end := offset + int(pageSize)
	if end > len(agents) {
		end = len(agents)
	}
	next := ""
	if end < len(agents) {
		next = base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(end)))
	}
	return connect.NewResponse(&agentsv1.ListAgentsResponse{
		Agents:        agents[offset:end],
		NextPageToken: next,
		Total:         total,
	}), nil
}

func (s *AgentServiceServer) ReloadAgents(ctx context.Context, _ *connect.Request[agentsv1.ReloadAgentsRequest]) (*connect.Response[agentsv1.ReloadAgentsResponse], error) {
	logger := log.FromContext(ctx)
	if s.runtime == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("config runtime not wired"))
	}
	logger.Info("reloading agent runtime")
	if err := s.runtime.ReloadRunner(ctx); err != nil {
		logger.Error("reload agent runtime failed", "err", err)
		return nil, toConnectError(err)
	}
	logger.Info("agent runtime reloaded")
	return connect.NewResponse(&agentsv1.ReloadAgentsResponse{ReloadedAt: timestamppb.New(time.Now().UTC())}), nil
}

func (s *AgentServiceServer) GetAgent(ctx context.Context, req *connect.Request[agentsv1.GetAgentRequest]) (*connect.Response[agentsv1.GetAgentResponse], error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	a, err := s.repo.GetAgent(ctx, wsID, req.Msg.GetName())
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&agentsv1.GetAgentResponse{Agent: a}), nil
}

func (s *AgentServiceServer) CreateAgent(ctx context.Context, req *connect.Request[agentsv1.CreateAgentRequest]) (*connect.Response[agentsv1.CreateAgentResponse], error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	name := req.Msg.GetAgent().GetName()
	if s.runnerSvc != nil && s.runnerSvc.IsReservedAgentName(name) {
		return nil, connect.NewError(connect.CodeFailedPrecondition,
			fmt.Errorf("agent name %q is reserved by a built-in agent", name))
	}
	if err := internalagent.ValidateWorkflowAgent(req.Msg.GetAgent()); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	logger := log.FromContext(ctx)
	logger.Info("creating agent", "workspace_id", wsID, "agent", req.Msg.GetAgent().GetName(), "type", req.Msg.GetAgent().GetType().String())
	a, err := mutateWithRuntime(
		func() (*agentsv1.Agent, error) {
			return s.repo.CreateAgent(ctx, wsID, req.Msg.GetAgent())
		},
		func() error {
			return s.reloadRuntime(ctx)
		},
		func() error {
			if err := s.repo.DeleteAgent(ctx, wsID, req.Msg.GetAgent().GetName()); err != nil {
				return err
			}
			return s.reloadRuntime(ctx)
		},
	)
	if err != nil {
		logger.Error("create agent failed", "workspace_id", wsID, "agent", req.Msg.GetAgent().GetName(), "err", err)
		return nil, toConnectError(err)
	}
	logger.Info("agent created", "workspace_id", wsID, "agent", a.GetName(), "type", a.GetType().String())
	return connect.NewResponse(&agentsv1.CreateAgentResponse{Agent: a}), nil
}

func (s *AgentServiceServer) UpdateAgent(ctx context.Context, req *connect.Request[agentsv1.UpdateAgentRequest]) (*connect.Response[agentsv1.UpdateAgentResponse], error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if err := internalagent.ValidateWorkflowAgent(req.Msg.GetAgent()); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	logger := log.FromContext(ctx)
	prev, err := s.repo.GetAgent(ctx, wsID, req.Msg.GetAgent().GetName())
	if err != nil {
		return nil, toConnectError(err)
	}
	logger.Info("updating agent", "workspace_id", wsID, "agent", req.Msg.GetAgent().GetName())

	a, err := mutateWithRuntime(
		func() (*agentsv1.Agent, error) {
			return s.repo.UpdateAgent(ctx, wsID, req.Msg.GetAgent())
		},
		func() error {
			return s.reloadRuntime(ctx)
		},
		func() error {
			if _, err := s.repo.UpdateAgent(ctx, wsID, proto.Clone(prev).(*agentsv1.Agent)); err != nil {
				return err
			}
			return s.reloadRuntime(ctx)
		},
	)
	if err != nil {
		logger.Error("update agent failed", "workspace_id", wsID, "agent", req.Msg.GetAgent().GetName(), "err", err)
		return nil, toConnectError(err)
	}
	logger.Info("agent updated", "workspace_id", wsID, "agent", a.GetName())
	return connect.NewResponse(&agentsv1.UpdateAgentResponse{Agent: a}), nil
}

func (s *AgentServiceServer) DeleteAgent(ctx context.Context, req *connect.Request[agentsv1.DeleteAgentRequest]) (*connect.Response[agentsv1.DeleteAgentResponse], error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	logger := log.FromContext(ctx)
	prev, err := s.repo.GetAgent(ctx, wsID, req.Msg.GetName())
	if err != nil {
		return nil, toConnectError(err)
	}
	logger.Info("deleting agent", "workspace_id", wsID, "agent", req.Msg.GetName())

	err = deleteWithRuntime(
		func() error {
			return s.repo.DeleteAgent(ctx, wsID, req.Msg.GetName())
		},
		func() error {
			return s.reloadRuntime(ctx)
		},
		func() error {
			if _, err := s.repo.CreateAgent(ctx, wsID, proto.Clone(prev).(*agentsv1.Agent)); err != nil {
				return err
			}
			return s.reloadRuntime(ctx)
		},
	)
	if err != nil {
		logger.Error("delete agent failed", "workspace_id", wsID, "agent", req.Msg.GetName(), "err", err)
		return nil, toConnectError(err)
	}
	logger.Info("agent deleted", "workspace_id", wsID, "agent", req.Msg.GetName())
	return connect.NewResponse(&agentsv1.DeleteAgentResponse{}), nil
}

func (s *AgentServiceServer) InvokeAgent(ctx context.Context, req *connect.Request[agentsv1.InvokeAgentRequest]) (*connect.Response[agentsv1.InvokeAgentResponse], error) {
	if s.runnerSvc == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("runner service not available"))
	}
	if req.Msg.GetAgentName() == "" {
		return nil, connectx.RequiredArgument("agent_name")
	}
	if req.Msg.GetInput() == "" {
		return nil, connectx.RequiredArgument("input")
	}
	if len(req.Msg.GetInput()) > maxInvokeAgentInputBytes {
		return nil, connectx.InvalidArgument("input",
			"exceeds maximum allowed size of "+strconv.Itoa(maxInvokeAgentInputBytes)+" bytes")
	}

	appName := req.Msg.GetAppName()
	if appName == "" {
		appName = "api"
	}
	userID := req.Msg.GetUserId()
	if userID == "" {
		userID = "api"
	}
	sessionID := req.Msg.GetSessionId()
	if sessionID == "" {
		sessionID = "invoke-" + uuid.NewString()
	}

	// An empty workspace on ctxInfo makes the runner treat the call as a
	// system path and skip the tenant boundary check, so non-admin callers
	// must always carry a validated workspace.
	wsID, hasWorkspace := workspace.FromContext(ctx)
	if !hasWorkspace && !auth.IsAdmin(ctx) {
		return nil, connect.NewError(connect.CodeFailedPrecondition,
			errors.New("workspace required (set X-Workspace-ID header)"))
	}
	ctxInfo := &agentsv1.ContextInfo{
		Uuid:        uuid.NewString(),
		SessionId:   sessionID,
		UserId:      userID,
		ChannelName: appName,
		Source:      agentsv1.ContextSource_CONTEXT_SOURCE_API,
		WorkspaceId: wsID,
	}

	logger := log.FromContext(ctx)
	logger.Info("invoking agent",
		"workspace_id", wsID,
		"agent", req.Msg.GetAgentName(),
		"app_name", appName,
		"user_id", userID,
		"session_id", sessionID,
		"model_override", req.Msg.GetModelOverride(),
		"input_len", len(req.Msg.GetInput()),
	)
	parts := []*genai.Part{{Text: req.Msg.GetInput()}}
	start := time.Now()
	response, err := s.runnerSvc.Run(ctx, req.Msg.GetAgentName(), parts, req.Msg.GetModelOverride(), ctxInfo, nil, nil)
	if err != nil {
		logger.Error("agent invocation failed",
			"workspace_id", wsID,
			"agent", req.Msg.GetAgentName(),
			"session_id", sessionID,
			"elapsed_ms", time.Since(start).Milliseconds(),
			"err", err,
		)
		return nil, connectx.InternalWith(err)
	}
	logger.Info("agent invocation completed",
		"workspace_id", wsID,
		"agent", req.Msg.GetAgentName(),
		"session_id", sessionID,
		"elapsed_ms", time.Since(start).Milliseconds(),
	)
	return connect.NewResponse(&agentsv1.InvokeAgentResponse{SessionId: sessionID, Response: response}), nil
}

func (s *AgentServiceServer) ListAgentInvocations(ctx context.Context, req *connect.Request[agentsv1.ListAgentInvocationsRequest]) (*connect.Response[agentsv1.ListAgentInvocationsResponse], error) {
	if s.invRepo == nil {
		return connect.NewResponse(&agentsv1.ListAgentInvocationsResponse{}), nil
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	invs, next, total, err := s.invRepo.List(ctx, invocation.ListFilter{
		WorkspaceID: wsID,
		AgentName:   req.Msg.GetAgentName(),
		SessionID:   req.Msg.GetSessionId(),
	}, req.Msg.GetPageSize(), req.Msg.GetPageToken())
	if err != nil {
		return nil, connectx.InternalWith(err)
	}
	return connect.NewResponse(&agentsv1.ListAgentInvocationsResponse{
		Invocations:   invs,
		NextPageToken: next,
		Total:         total,
	}), nil
}

func (s *AgentServiceServer) CancelAgentInvocation(ctx context.Context, req *connect.Request[agentsv1.CancelAgentInvocationRequest]) (*connect.Response[agentsv1.CancelAgentInvocationResponse], error) {
	if s.runnerSvc == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("runner service not available"))
	}
	if req.Msg.GetInvocationId() == "" {
		return nil, connectx.RequiredArgument("invocation_id")
	}
	// Scope the cancel to the caller's workspace; admins without a workspace
	// header keep the global (system) cancel path.
	wsID, hasWorkspace := workspace.FromContext(ctx)
	if !hasWorkspace && !auth.IsAdmin(ctx) {
		return nil, connect.NewError(connect.CodeFailedPrecondition,
			errors.New("workspace required (set X-Workspace-ID header)"))
	}
	cancelled := s.runnerSvc.CancelInvocation(req.Msg.GetInvocationId(), wsID)
	log.FromContext(ctx).Info("cancel agent invocation requested",
		"invocation_id", req.Msg.GetInvocationId(),
		"workspace_id", wsID,
		"cancelled", cancelled,
	)
	return connect.NewResponse(&agentsv1.CancelAgentInvocationResponse{Cancelled: cancelled}), nil
}

func (s *AgentServiceServer) GetAgentRuntimeStatus(ctx context.Context, req *connect.Request[agentsv1.GetAgentRuntimeStatusRequest]) (*connect.Response[agentsv1.GetAgentRuntimeStatusResponse], error) {
	if req.Msg.GetName() == "" {
		return nil, connectx.RequiredArgument("name")
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	statuses, err := s.runtimeStatuses(ctx, wsID, []string{req.Msg.GetName()})
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&agentsv1.GetAgentRuntimeStatusResponse{Status: statuses[0]}), nil
}

func (s *AgentServiceServer) ListAgentRuntimeStatuses(ctx context.Context, req *connect.Request[agentsv1.ListAgentRuntimeStatusesRequest]) (*connect.Response[agentsv1.ListAgentRuntimeStatusesResponse], error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	names := req.Msg.GetNames()
	if len(names) == 0 {
		agents, err := s.repo.ListAgents(ctx, wsID)
		if err != nil {
			return nil, toConnectError(err)
		}
		names = make([]string, 0, len(agents))
		for _, a := range agents {
			names = append(names, a.GetName())
		}
	}
	statuses, err := s.runtimeStatuses(ctx, wsID, names)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&agentsv1.ListAgentRuntimeStatusesResponse{Statuses: statuses}), nil
}

// runtimeStatuses derives AgentRuntimeStatus for the named agents from a single
// invocation repo query. Agents with no invocations (or when the repo is not
// wired) are reported as IDLE. The result preserves the order of names.
func (s *AgentServiceServer) runtimeStatuses(ctx context.Context, workspaceID string, names []string) ([]*agentsv1.AgentRuntimeStatus, error) {
	out := make([]*agentsv1.AgentRuntimeStatus, len(names))
	for i, name := range names {
		out[i] = &agentsv1.AgentRuntimeStatus{
			Name:  name,
			State: agentsv1.AgentRuntimeState_AGENT_RUNTIME_STATE_IDLE,
		}
	}
	if s.invRepo == nil || len(names) == 0 {
		return out, nil
	}
	summaries, err := s.invRepo.StatusSummaries(ctx, workspaceID, names)
	if err != nil {
		return nil, connectx.InternalWith(fmt.Errorf("query invocation status: %w", err))
	}
	for _, status := range out {
		sum, ok := summaries[status.GetName()]
		if !ok || sum.Latest == nil {
			continue
		}
		latest := sum.Latest
		status.LastInvocationId = latest.GetId()
		if ts := latest.GetFinishedAt(); ts != nil {
			status.LastRunAt = ts
		} else if ts := latest.GetStartedAt(); ts != nil {
			status.LastRunAt = ts
		}
		switch latest.GetStatus() {
		case agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED:
			status.State = agentsv1.AgentRuntimeState_AGENT_RUNTIME_STATE_FAILED
		case agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING:
			status.State = agentsv1.AgentRuntimeState_AGENT_RUNTIME_STATE_RUNNING
		default:
			status.State = agentsv1.AgentRuntimeState_AGENT_RUNTIME_STATE_IDLE
		}
		status.InFlight = sum.Running
		if status.InFlight > 0 {
			status.State = agentsv1.AgentRuntimeState_AGENT_RUNTIME_STATE_RUNNING
		}
	}
	return out, nil
}

func (s *AgentServiceServer) reloadRuntime(ctx context.Context) error {
	if s.runtime == nil {
		return nil
	}
	if err := s.runtime.ReloadRunner(ctx); err != nil {
		return toConnectError(err)
	}
	return nil
}
