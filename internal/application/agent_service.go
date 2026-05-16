package application

import (
	"context"
	"encoding/base64"
	"errors"
	"sort"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/twitchtv/twirp"
	"google.golang.org/genai"

	configrepo "go.orx.me/apps/butter/internal/repo/config"
	"go.orx.me/apps/butter/internal/repo/invocation"
	"go.orx.me/apps/butter/internal/runtime/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

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
	agents, err := s.repo.ListAgents(ctx)
	if err != nil {
		return nil, toTwirpError(err)
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
	if s.runtime == nil {
		return nil, twirp.NewError(twirp.FailedPrecondition, "config runtime not wired")
	}
	if err := s.runtime.ReloadRunner(ctx); err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.ReloadAgentsResponse{ReloadedAt: timestamppb.New(time.Now().UTC())}, nil
}

func (s *AgentServiceServer) GetAgent(ctx context.Context, req *agentsv1.GetAgentRequest) (*agentsv1.GetAgentResponse, error) {
	a, err := s.repo.GetAgent(ctx, req.GetName())
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.GetAgentResponse{Agent: a}, nil
}

func (s *AgentServiceServer) CreateAgent(ctx context.Context, req *agentsv1.CreateAgentRequest) (*agentsv1.CreateAgentResponse, error) {
	a, err := mutateWithRuntime(
		func() (*agentsv1.Agent, error) {
			return s.repo.CreateAgent(ctx, req.GetAgent())
		},
		func() error {
			return s.reloadRuntime(ctx)
		},
		func() error {
			if err := s.repo.DeleteAgent(ctx, req.GetAgent().GetName()); err != nil {
				return err
			}
			return s.reloadRuntime(ctx)
		},
	)
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.CreateAgentResponse{Agent: a}, nil
}

func (s *AgentServiceServer) UpdateAgent(ctx context.Context, req *agentsv1.UpdateAgentRequest) (*agentsv1.UpdateAgentResponse, error) {
	prev, err := s.repo.GetAgent(ctx, req.GetAgent().GetName())
	if err != nil {
		return nil, toTwirpError(err)
	}

	a, err := mutateWithRuntime(
		func() (*agentsv1.Agent, error) {
			return s.repo.UpdateAgent(ctx, req.GetAgent())
		},
		func() error {
			return s.reloadRuntime(ctx)
		},
		func() error {
			if _, err := s.repo.UpdateAgent(ctx, proto.Clone(prev).(*agentsv1.Agent)); err != nil {
				return err
			}
			return s.reloadRuntime(ctx)
		},
	)
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.UpdateAgentResponse{Agent: a}, nil
}

func (s *AgentServiceServer) DeleteAgent(ctx context.Context, req *agentsv1.DeleteAgentRequest) (*agentsv1.DeleteAgentResponse, error) {
	prev, err := s.repo.GetAgent(ctx, req.GetName())
	if err != nil {
		return nil, toTwirpError(err)
	}

	err = deleteWithRuntime(
		func() error {
			return s.repo.DeleteAgent(ctx, req.GetName())
		},
		func() error {
			return s.reloadRuntime(ctx)
		},
		func() error {
			if _, err := s.repo.CreateAgent(ctx, proto.Clone(prev).(*agentsv1.Agent)); err != nil {
				return err
			}
			return s.reloadRuntime(ctx)
		},
	)
	if err != nil {
		return nil, toTwirpError(err)
	}
	return &agentsv1.DeleteAgentResponse{}, nil
}

func (s *AgentServiceServer) InvokeAgent(ctx context.Context, req *agentsv1.InvokeAgentRequest) (*agentsv1.InvokeAgentResponse, error) {
	if s.runnerSvc == nil {
		return nil, twirp.NewError(twirp.FailedPrecondition, "runner service not available")
	}
	if req.GetAgentName() == "" {
		return nil, twirp.RequiredArgumentError("agent_name")
	}
	if req.GetInput() == "" {
		return nil, twirp.RequiredArgumentError("input")
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

	ctxInfo := &agentsv1.ContextInfo{
		Uuid:        uuid.NewString(),
		SessionId:   sessionID,
		UserId:      userID,
		ChannelName: appName,
		Source:      agentsv1.ContextSource_CONTEXT_SOURCE_API,
	}

	parts := []*genai.Part{{Text: req.GetInput()}}
	response, err := s.runnerSvc.Run(ctx, req.GetAgentName(), parts, req.GetModelOverride(), ctxInfo, nil, nil)
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}
	return &agentsv1.InvokeAgentResponse{SessionId: sessionID, Response: response}, nil
}

func (s *AgentServiceServer) ListAgentInvocations(ctx context.Context, req *agentsv1.ListAgentInvocationsRequest) (*agentsv1.ListAgentInvocationsResponse, error) {
	if s.invRepo == nil {
		return &agentsv1.ListAgentInvocationsResponse{}, nil
	}
	invs, next, total, err := s.invRepo.List(ctx, invocation.ListFilter{
		AgentName: req.GetAgentName(),
		SessionID: req.GetSessionId(),
	}, req.GetPageSize(), req.GetPageToken())
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}
	return &agentsv1.ListAgentInvocationsResponse{
		Invocations:   invs,
		NextPageToken: next,
		Total:         total,
	}, nil
}

func (s *AgentServiceServer) CancelAgentInvocation(ctx context.Context, req *agentsv1.CancelAgentInvocationRequest) (*agentsv1.CancelAgentInvocationResponse, error) {
	if s.runnerSvc == nil {
		return nil, twirp.NewError(twirp.FailedPrecondition, "runner service not available")
	}
	if req.GetInvocationId() == "" {
		return nil, twirp.RequiredArgumentError("invocation_id")
	}
	cancelled := s.runnerSvc.CancelInvocation(req.GetInvocationId())
	return &agentsv1.CancelAgentInvocationResponse{Cancelled: cancelled}, nil
}

func (s *AgentServiceServer) GetAgentRuntimeStatus(ctx context.Context, req *agentsv1.GetAgentRuntimeStatusRequest) (*agentsv1.GetAgentRuntimeStatusResponse, error) {
	if req.GetName() == "" {
		return nil, twirp.RequiredArgumentError("name")
	}
	status := s.runtimeStatusFor(ctx, req.GetName())
	return &agentsv1.GetAgentRuntimeStatusResponse{Status: status}, nil
}

func (s *AgentServiceServer) ListAgentRuntimeStatuses(ctx context.Context, req *agentsv1.ListAgentRuntimeStatusesRequest) (*agentsv1.ListAgentRuntimeStatusesResponse, error) {
	names := req.GetNames()
	if len(names) == 0 {
		agents, err := s.repo.ListAgents(ctx)
		if err != nil {
			return nil, toTwirpError(err)
		}
		names = make([]string, 0, len(agents))
		for _, a := range agents {
			names = append(names, a.GetName())
		}
	}
	out := make([]*agentsv1.AgentRuntimeStatus, 0, len(names))
	for _, name := range names {
		out = append(out, s.runtimeStatusFor(ctx, name))
	}
	return &agentsv1.ListAgentRuntimeStatusesResponse{Statuses: out}, nil
}

// runtimeStatusFor derives an AgentRuntimeStatus from the invocation repo. If
// no invocations exist (or the repo is not wired) the agent is reported as IDLE.
func (s *AgentServiceServer) runtimeStatusFor(ctx context.Context, name string) *agentsv1.AgentRuntimeStatus {
	out := &agentsv1.AgentRuntimeStatus{
		Name:  name,
		State: agentsv1.AgentRuntimeState_AGENT_RUNTIME_STATE_IDLE,
	}
	if s.invRepo == nil {
		return out
	}
	// Pull the most recent 100 invocations for this agent — enough to derive
	// last_run_at + in_flight count for the dashboard table.
	invs, _, _, err := s.invRepo.List(ctx, invocation.ListFilter{AgentName: name}, 100, "")
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
		return toTwirpError(err)
	}
	return nil
}

type ConfigRuntime interface {
	ReloadRunner(ctx context.Context) error
	ReloadChannels(ctx context.Context) error
}

func toTwirpError(err error) twirp.Error {
	if twerr, ok := err.(twirp.Error); ok {
		return twerr
	}
	if errors.Is(err, configrepo.ErrNotFound) {
		return twirp.NotFoundError(err.Error())
	}
	if errors.Is(err, configrepo.ErrAlreadyExists) {
		return twirp.NewError(twirp.AlreadyExists, err.Error())
	}
	return twirp.InternalErrorWith(err)
}
