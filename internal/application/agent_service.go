package application

import (
	"context"
	"encoding/base64"
	"errors"
	"sort"
	"strconv"
	"time"

	"butterfly.orx.me/core/log"
	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/genai"

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

func (s *AgentServiceServer) ListAgents(ctx context.Context, req *agentsv1.ListAgentsRequest) (*agentsv1.ListAgentsResponse, error) {
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
	pageSize := req.GetPageSize()
	if pageSize <= 0 {
		pageSize = 50
	}
	offset := 0
	if token := req.GetPageToken(); token != "" {
		if raw, err := base64.StdEncoding.DecodeString(token); err == nil {
			if n, err := strconv.Atoi(string(raw)); err == nil && n >= 0 {
				offset = n
			}
		}
	}
	if offset >= len(agents) {
		return &agentsv1.ListAgentsResponse{Total: total}, nil
	}
	end := offset + int(pageSize)
	if end > len(agents) {
		end = len(agents)
	}
	next := ""
	if end < len(agents) {
		next = base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(end)))
	}
	return &agentsv1.ListAgentsResponse{
		Agents:        agents[offset:end],
		NextPageToken: next,
		Total:         total,
	}, nil
}

func (s *AgentServiceServer) ReloadAgents(ctx context.Context, _ *agentsv1.ReloadAgentsRequest) (*agentsv1.ReloadAgentsResponse, error) {
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
	return &agentsv1.ReloadAgentsResponse{ReloadedAt: timestamppb.New(time.Now().UTC())}, nil
}

func (s *AgentServiceServer) GetAgent(ctx context.Context, req *agentsv1.GetAgentRequest) (*agentsv1.GetAgentResponse, error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	a, err := s.repo.GetAgent(ctx, wsID, req.GetName())
	if err != nil {
		return nil, toConnectError(err)
	}
	return &agentsv1.GetAgentResponse{Agent: a}, nil
}

func (s *AgentServiceServer) CreateAgent(ctx context.Context, req *agentsv1.CreateAgentRequest) (*agentsv1.CreateAgentResponse, error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	logger := log.FromContext(ctx)
	logger.Info("creating agent", "workspace_id", wsID, "agent", req.GetAgent().GetName(), "type", req.GetAgent().GetType().String())
	a, err := mutateWithRuntime(
		func() (*agentsv1.Agent, error) {
			return s.repo.CreateAgent(ctx, wsID, req.GetAgent())
		},
		func() error {
			return s.reloadRuntime(ctx)
		},
		func() error {
			if err := s.repo.DeleteAgent(ctx, wsID, req.GetAgent().GetName()); err != nil {
				return err
			}
			return s.reloadRuntime(ctx)
		},
	)
	if err != nil {
		logger.Error("create agent failed", "workspace_id", wsID, "agent", req.GetAgent().GetName(), "err", err)
		return nil, toConnectError(err)
	}
	logger.Info("agent created", "workspace_id", wsID, "agent", a.GetName(), "type", a.GetType().String())
	return &agentsv1.CreateAgentResponse{Agent: a}, nil
}

func (s *AgentServiceServer) UpdateAgent(ctx context.Context, req *agentsv1.UpdateAgentRequest) (*agentsv1.UpdateAgentResponse, error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	logger := log.FromContext(ctx)
	prev, err := s.repo.GetAgent(ctx, wsID, req.GetAgent().GetName())
	if err != nil {
		return nil, toConnectError(err)
	}
	logger.Info("updating agent", "workspace_id", wsID, "agent", req.GetAgent().GetName())

	a, err := mutateWithRuntime(
		func() (*agentsv1.Agent, error) {
			return s.repo.UpdateAgent(ctx, wsID, req.GetAgent())
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
		logger.Error("update agent failed", "workspace_id", wsID, "agent", req.GetAgent().GetName(), "err", err)
		return nil, toConnectError(err)
	}
	logger.Info("agent updated", "workspace_id", wsID, "agent", a.GetName())
	return &agentsv1.UpdateAgentResponse{Agent: a}, nil
}

func (s *AgentServiceServer) DeleteAgent(ctx context.Context, req *agentsv1.DeleteAgentRequest) (*agentsv1.DeleteAgentResponse, error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	logger := log.FromContext(ctx)
	prev, err := s.repo.GetAgent(ctx, wsID, req.GetName())
	if err != nil {
		return nil, toConnectError(err)
	}
	logger.Info("deleting agent", "workspace_id", wsID, "agent", req.GetName())

	err = deleteWithRuntime(
		func() error {
			return s.repo.DeleteAgent(ctx, wsID, req.GetName())
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
		logger.Error("delete agent failed", "workspace_id", wsID, "agent", req.GetName(), "err", err)
		return nil, toConnectError(err)
	}
	logger.Info("agent deleted", "workspace_id", wsID, "agent", req.GetName())
	return &agentsv1.DeleteAgentResponse{}, nil
}

func (s *AgentServiceServer) InvokeAgent(ctx context.Context, req *agentsv1.InvokeAgentRequest) (*agentsv1.InvokeAgentResponse, error) {
	if s.runnerSvc == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("runner service not available"))
	}
	if req.GetAgentName() == "" {
		return nil, connectx.RequiredArgument("agent_name")
	}
	if req.GetInput() == "" {
		return nil, connectx.RequiredArgument("input")
	}
	if len(req.GetInput()) > maxInvokeAgentInputBytes {
		return nil, connectx.InvalidArgument("input",
			"exceeds maximum allowed size of "+strconv.Itoa(maxInvokeAgentInputBytes)+" bytes")
	}

	appName := req.GetAppName()
	if appName == "" {
		appName = "api"
	}
	userID := req.GetUserId()
	if userID == "" {
		userID = "api"
	}
	sessionID := req.GetSessionId()
	if sessionID == "" {
		sessionID = "invoke-" + uuid.NewString()
	}

	wsID, _ := workspace.FromContext(ctx)
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
		"agent", req.GetAgentName(),
		"app_name", appName,
		"user_id", userID,
		"session_id", sessionID,
		"model_override", req.GetModelOverride(),
		"input_len", len(req.GetInput()),
	)
	parts := []*genai.Part{{Text: req.GetInput()}}
	start := time.Now()
	response, err := s.runnerSvc.Run(ctx, req.GetAgentName(), parts, req.GetModelOverride(), ctxInfo, nil, nil)
	if err != nil {
		logger.Error("agent invocation failed",
			"workspace_id", wsID,
			"agent", req.GetAgentName(),
			"session_id", sessionID,
			"elapsed_ms", time.Since(start).Milliseconds(),
			"err", err,
		)
		return nil, connectx.InternalWith(err)
	}
	logger.Info("agent invocation completed",
		"workspace_id", wsID,
		"agent", req.GetAgentName(),
		"session_id", sessionID,
		"elapsed_ms", time.Since(start).Milliseconds(),
	)
	return &agentsv1.InvokeAgentResponse{SessionId: sessionID, Response: response}, nil
}

func (s *AgentServiceServer) ListAgentInvocations(ctx context.Context, req *agentsv1.ListAgentInvocationsRequest) (*agentsv1.ListAgentInvocationsResponse, error) {
	if s.invRepo == nil {
		return &agentsv1.ListAgentInvocationsResponse{}, nil
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	invs, next, total, err := s.invRepo.List(ctx, invocation.ListFilter{
		WorkspaceID: wsID,
		AgentName:   req.GetAgentName(),
		SessionID:   req.GetSessionId(),
	}, req.GetPageSize(), req.GetPageToken())
	if err != nil {
		return nil, connectx.InternalWith(err)
	}
	return &agentsv1.ListAgentInvocationsResponse{
		Invocations:   invs,
		NextPageToken: next,
		Total:         total,
	}, nil
}

func (s *AgentServiceServer) CancelAgentInvocation(ctx context.Context, req *agentsv1.CancelAgentInvocationRequest) (*agentsv1.CancelAgentInvocationResponse, error) {
	if s.runnerSvc == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("runner service not available"))
	}
	if req.GetInvocationId() == "" {
		return nil, connectx.RequiredArgument("invocation_id")
	}
	cancelled := s.runnerSvc.CancelInvocation(req.GetInvocationId())
	log.FromContext(ctx).Info("cancel agent invocation requested",
		"invocation_id", req.GetInvocationId(),
		"cancelled", cancelled,
	)
	return &agentsv1.CancelAgentInvocationResponse{Cancelled: cancelled}, nil
}

func (s *AgentServiceServer) GetAgentRuntimeStatus(ctx context.Context, req *agentsv1.GetAgentRuntimeStatusRequest) (*agentsv1.GetAgentRuntimeStatusResponse, error) {
	if req.GetName() == "" {
		return nil, connectx.RequiredArgument("name")
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	status := s.runtimeStatusFor(ctx, wsID, req.GetName())
	return &agentsv1.GetAgentRuntimeStatusResponse{Status: status}, nil
}

func (s *AgentServiceServer) ListAgentRuntimeStatuses(ctx context.Context, req *agentsv1.ListAgentRuntimeStatusesRequest) (*agentsv1.ListAgentRuntimeStatusesResponse, error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	names := req.GetNames()
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
	out := make([]*agentsv1.AgentRuntimeStatus, 0, len(names))
	for _, name := range names {
		out = append(out, s.runtimeStatusFor(ctx, wsID, name))
	}
	return &agentsv1.ListAgentRuntimeStatusesResponse{Statuses: out}, nil
}

// runtimeStatusFor derives an AgentRuntimeStatus from the invocation repo. If
// no invocations exist (or the repo is not wired) the agent is reported as IDLE.
func (s *AgentServiceServer) runtimeStatusFor(ctx context.Context, workspaceID, name string) *agentsv1.AgentRuntimeStatus {
	out := &agentsv1.AgentRuntimeStatus{
		Name:  name,
		State: agentsv1.AgentRuntimeState_AGENT_RUNTIME_STATE_IDLE,
	}
	if s.invRepo == nil {
		return out
	}
	// Pull the most recent 100 invocations for this agent — enough to derive
	// last_run_at + in_flight count for the dashboard table.
	invs, _, _, err := s.invRepo.List(ctx, invocation.ListFilter{WorkspaceID: workspaceID, AgentName: name}, 100, "")
	if err != nil || len(invs) == 0 {
		return out
	}
	latest := invs[0]
	out.LastInvocationId = latest.GetId()
	if ts := latest.GetFinishedAt(); ts != nil {
		out.LastRunAt = ts
	} else if ts := latest.GetStartedAt(); ts != nil {
		out.LastRunAt = ts
	}
	switch latest.GetStatus() {
	case agentsv1.InvocationStatus_INVOCATION_STATUS_FAILED:
		out.State = agentsv1.AgentRuntimeState_AGENT_RUNTIME_STATE_FAILED
	case agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING:
		out.State = agentsv1.AgentRuntimeState_AGENT_RUNTIME_STATE_RUNNING
	default:
		out.State = agentsv1.AgentRuntimeState_AGENT_RUNTIME_STATE_IDLE
	}
	for _, inv := range invs {
		if inv.GetStatus() == agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING {
			out.InFlight++
		}
	}
	if out.InFlight > 0 {
		out.State = agentsv1.AgentRuntimeState_AGENT_RUNTIME_STATE_RUNNING
	}
	return out
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

type ConfigRuntime interface {
	ReloadRunner(ctx context.Context) error
	ReloadChannels(ctx context.Context) error
}

func toConnectError(err error) *connect.Error {
	var cerr *connect.Error
	if errors.As(err, &cerr) {
		return cerr
	}
	if errors.Is(err, configrepo.ErrNotFound) {
		return connectx.NotFound(err.Error())
	}
	if errors.Is(err, configrepo.ErrAlreadyExists) {
		return connect.NewError(connect.CodeAlreadyExists, errors.New(err.Error()))
	}
	return connectx.InternalWith(err)
}
