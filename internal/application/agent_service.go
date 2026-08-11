package application

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"butterfly.orx.me/core/log"
	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/genai"

	internalagent "go.orx.me/apps/butter/internal/agent"
	agentoprepo "go.orx.me/apps/butter/internal/repo/agentop"
	"go.orx.me/apps/butter/internal/repo/auth"
	configrepo "go.orx.me/apps/butter/internal/repo/config"
	"go.orx.me/apps/butter/internal/repo/inputpart"
	"go.orx.me/apps/butter/internal/repo/invocation"
	workspacerepo "go.orx.me/apps/butter/internal/repo/workspace"
	"go.orx.me/apps/butter/internal/runtime/runner"
	"go.orx.me/apps/butter/internal/transport/connectx"
	"go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	adksession "google.golang.org/adk/v2/session"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// maxInvokeAgentInputBytes caps the size of a single InvokeAgent input to
// protect the runner and session storage from oversized requests.
const maxInvokeAgentInputBytes = 1 << 20 // 1 MiB

// agentRunner is the subset of *runner.Service the agent service depends
// on; tests substitute a fake implementation.
type agentRunner interface {
	IsReservedAgentName(name string) bool
	Run(ctx context.Context, agentName string, parts []*genai.Part, modelOverride string, ctxInfo *agentsv1.ContextInfo, onEvent runner.EventCallback, onCompaction runner.CompactionCallback) (string, error)
	RunSSE(ctx context.Context, agentName string, parts []*genai.Part, modelOverride string, ctxInfo *agentsv1.ContextInfo, onEvent runner.EventCallback, onCompaction runner.CompactionCallback) (string, error)
	CancelInvocation(id, workspaceID string) bool
	ResolveAgentRef(workspaceID, agentID string) (string, bool)
	GetAgentIdentity(name string) (agentID, displayName string, ok bool)
}

// resolveAgentRunnerRef maps a request's agent_id to the registered runtime
// agent name. agent_id is the sole agent reference on every interface
// (issue #213 contract step); an unknown id is NotFound.
func resolveAgentRunnerRef(r interface {
	ResolveAgentRef(workspaceID, agentID string) (string, bool)
}, workspaceID, agentID string) (string, error) {
	if agentID == "" {
		return "", connectx.RequiredArgument("agent_id")
	}
	name, ok := r.ResolveAgentRef(workspaceID, agentID)
	if !ok {
		return "", connectx.NotFound(fmt.Sprintf("agent %q not found", agentID))
	}
	return name, nil
}

type AgentServiceServer struct {
	repo          configrepo.AgentRepository
	runtime       ConfigRuntime
	runnerSvc     agentRunner
	invRepo       invocation.Repository
	inputPartRepo inputpart.Repository
	wsRepo        workspacerepo.Repository
	opRepo        agentoprepo.Repository
	content       agentContentCoordinator
	asyncCoord    asyncCoordinator
	sessionSvc    adksession.Service

	// asyncSubmitMu serializes the short accept transaction (idempotency
	// lookup, active-session check, optional Session creation, Invocation
	// persistence). Execution itself remains fully concurrent across Sessions.
	asyncSubmitMu sync.Mutex
}

func NewAgentServiceServer(repo configrepo.AgentRepository) *AgentServiceServer {
	return &AgentServiceServer{repo: repo}
}

// SetOperationRepo wires the durable Agent lifecycle operation store used by
// the create/save/delete/restore Sagas (issue #218).
func (s *AgentServiceServer) SetOperationRepo(repo agentoprepo.Repository) {
	s.opRepo = repo
}

// SetContentCoordinator wires the Agent Content write/publish seam (satisfied
// by *RepoBindingServiceServer) used by the lifecycle Sagas.
func (s *AgentServiceServer) SetContentCoordinator(c agentContentCoordinator) {
	s.content = c
}

// coordinator builds the Saga coordinator from the wired dependencies, or
// returns nil when lifecycle Sagas are unavailable (ops repo / content seam
// not wired) so callers fall back to the single-step path.
func (s *AgentServiceServer) coordinator() *agentOperationCoordinator {
	if s.opRepo == nil || s.content == nil {
		return nil
	}
	return newAgentOperationCoordinator(s.repo, s.opRepo, s.content, s.reloadRuntime)
}

func (s *AgentServiceServer) SetRuntime(runtime ConfigRuntime) {
	s.runtime = runtime
}

// SetRunnerService wires the runner so InvokeAgent can execute agents.
// A nil *runner.Service is ignored so the nil checks in the RPC methods
// keep working against the interface-typed field.
func (s *AgentServiceServer) SetRunnerService(svc *runner.Service) {
	if svc == nil {
		return
	}
	s.runnerSvc = svc
}

// SetInvocationRepo wires the invocation repository used by
// ListAgentInvocations.
func (s *AgentServiceServer) SetInvocationRepo(repo invocation.Repository) {
	s.invRepo = repo
}

// SetInputPartRepo wires the Input Part repository used to persist multimodal
// input for async invocations.
func (s *AgentServiceServer) SetInputPartRepo(repo inputpart.Repository) {
	s.inputPartRepo = repo
}

// SetWorkspaceRepo wires the workspace repository used by
// AssignAgentID for role-based authorization.
func (s *AgentServiceServer) SetWorkspaceRepo(repo workspacerepo.Repository) {
	s.wsRepo = repo
}

// overlayActiveContent replaces description/instruction/global_instruction on
// agents whose content is Git-owned. When the workspace has no binding or no
// published Active Revision the agents pass through unchanged. A published
// revision whose exact snapshot is unavailable is an error, never a DB fallback.
func (s *AgentServiceServer) overlayActiveContent(ctx context.Context, wsID string, agents ...*agentsv1.Agent) error {
	if s.content == nil {
		return nil
	}
	snap, err := s.content.GetActiveSnapshot(ctx, wsID)
	if err != nil {
		return err
	}
	if snap == nil || len(snap.Entries) == 0 {
		return nil
	}
	for _, a := range agents {
		agentID := a.GetAgentId()
		if agentID == "" {
			continue
		}
		c, ok := snap.Entries[agentID]
		if !ok {
			continue
		}
		a.Description = c.Description
		if a.Config == nil {
			a.Config = &agentsv1.AgentConfig{}
		}
		a.Config.Instruction = c.Instruction
		a.Config.GlobalInstruction = c.GlobalInstruction
	}
	return nil
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

	if err := s.overlayActiveContent(ctx, wsID, agents...); err != nil {
		return nil, connectx.InternalWith(err)
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
	var a *agentsv1.Agent
	if aid := req.Msg.GetAgentId(); aid != "" {
		a, err = s.repo.GetAgentByID(ctx, wsID, aid)
	} else {
		a, err = s.repo.GetAgent(ctx, wsID, req.Msg.GetName())
	}
	if err != nil {
		return nil, toConnectError(err)
	}
	if err := s.overlayActiveContent(ctx, wsID, a); err != nil {
		return nil, connectx.InternalWith(err)
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

	agent := proto.Clone(req.Msg.GetAgent()).(*agentsv1.Agent)

	// V2 contract: every new agent is created with an immutable agent_id and
	// composes children via ID references — the embedded sub_agents write
	// path is gone. Legacy records with embedded children stay readable
	// until MigrateAgentsV2 expands them.
	if agent.GetAgentId() == "" {
		return nil, connectx.RequiredArgument("agent_id")
	}
	if len(agent.GetSubAgents()) > 0 {
		return nil, connectx.InvalidArgument("sub_agents",
			"is no longer writable; create child agents and reference them via child_agent_ids")
	}
	if err := internalagent.ValidateAgentID(agent.GetAgentId()); err != nil {
		return nil, connectx.InvalidArgument("agent_id", err.Error())
	}
	agent.LifecycleStatus = agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_ACTIVE
	agent.WorkspaceId = wsID
	if len(agent.GetChildAgentIds()) > 0 {
		pool, err := s.repo.ListAgents(ctx, wsID)
		if err != nil {
			return nil, toConnectError(err)
		}
		if err := internalagent.ValidateAgentRelationships(agent, pool); err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
	}

	logger := log.FromContext(ctx)
	coord := s.coordinator()
	bound := false
	if coord != nil {
		bound, err = s.content.HasBinding(ctx, wsID)
		if err != nil {
			return nil, connectx.InternalWith(err)
		}
		if bound {
			if err := s.requireOwnerOrAdmin(ctx, wsID); err != nil {
				return nil, err
			}
			initial := req.Msg.GetInitialContent()
			if agentTypeRequiresPrompt(agent.GetType()) && strings.TrimSpace(initial.GetPrompt()) == "" {
				return nil, connectx.RequiredArgument("initial_content.prompt")
			}
			reqJSON, err := createOperationRequestJSON(agent, initial)
			if err != nil {
				return nil, connectx.InternalWith(err)
			}
			created, op, vErrs, found, err := coord.resumeExisting(ctx, wsID, req.Msg.GetOperationId(),
				agentsv1.AgentOperationType_AGENT_OPERATION_TYPE_CREATE, agent.GetAgentId(), reqJSON, 0, "")
			if found {
				if err != nil {
					return nil, operationError(err, op)
				}
				if len(vErrs) > 0 {
					return nil, operationError(contentValidationError(vErrs), op)
				}
				return connect.NewResponse(&agentsv1.CreateAgentResponse{Agent: created, Operation: op}), nil
			}
		}
	}

	taken, err := s.repo.AgentIDExists(ctx, wsID, agent.GetAgentId())
	if err != nil {
		return nil, toConnectError(err)
	}
	if taken {
		// A DELETED tombstone keeps its ID reserved; steer the caller to
		// RestoreAgent rather than reporting a bare conflict (issue #218).
		if existing, gErr := s.repo.GetAgentByID(ctx, wsID, agent.GetAgentId()); gErr == nil &&
			existing.GetLifecycleStatus() == agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_DELETED {
			return nil, connect.NewError(connect.CodeFailedPrecondition,
				fmt.Errorf("agent_id %q belongs to a deleted agent; use RestoreAgent to reactivate it", agent.GetAgentId()))
		}
		return nil, connect.NewError(connect.CodeAlreadyExists,
			fmt.Errorf("agent_id %q is already in use in this workspace", agent.GetAgentId()))
	}

	// Git-bound workspace with a Saga coordinator: create through the durable
	// lifecycle Saga (PROVISIONING → commit content → publish → ACTIVE).
	if coord != nil {
		if bound {
			initial := req.Msg.GetInitialContent()
			logger.Info("creating bound agent via saga", "workspace_id", wsID, "agent", agent.GetName())
			created, op, vErrs, err := coord.RunCreate(ctx, wsID, agent, initial, req.Msg.GetOperationId())
			if err != nil {
				return nil, operationError(err, op)
			}
			if len(vErrs) > 0 {
				return nil, operationError(contentValidationError(vErrs), op)
			}
			return connect.NewResponse(&agentsv1.CreateAgentResponse{Agent: created, Operation: op}), nil
		}
	}

	logger.Info("creating agent", "workspace_id", wsID, "agent", agent.GetName(), "type", agent.GetType().String())
	a, err := mutateWithRuntime(
		func() (*agentsv1.Agent, error) {
			return s.repo.CreateAgent(ctx, wsID, agent)
		},
		func() error {
			return s.reloadRuntime(ctx)
		},
		func() error {
			if err := s.repo.DeleteAgent(ctx, wsID, agent.GetName()); err != nil {
				return err
			}
			return s.reloadRuntime(ctx)
		},
	)
	if err != nil {
		logger.Error("create agent failed", "workspace_id", wsID, "agent", agent.GetName(), "err", err)
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

	update := proto.Clone(req.Msg.GetAgent()).(*agentsv1.Agent)
	if update.GetAgentId() != "" && update.GetAgentId() != prev.GetAgentId() {
		return nil, connect.NewError(connect.CodeInvalidArgument,
			fmt.Errorf("agent_id cannot be set or changed via UpdateAgent; use AssignAgentID"))
	}
	update.AgentId = prev.GetAgentId()
	update.LifecycleStatus = prev.GetLifecycleStatus()
	update.LegacyName = prev.GetLegacyName()
	update.CreatedAt = prev.GetCreatedAt()
	update.DeletedAt = prev.GetDeletedAt()
	if s.content != nil {
		// Content is only read-only through this API once it is Git-owned (bound
		// AND an Active Revision published). A bound-but-not-yet-onboarded
		// workspace stays database-owned, so its content remains editable here —
		// which is also what lets an operator fix content and re-run a failed
		// EXPORT_CURRENT onboarding through the normal API (#219).
		gitOwned, bindErr := s.content.IsContentGitOwned(ctx, wsID)
		if bindErr != nil {
			return nil, connectx.InternalWith(bindErr)
		}
		if gitOwned {
			if _, snapshotErr := s.content.GetActiveSnapshot(ctx, wsID); snapshotErr != nil {
				return nil, connectx.InternalWith(snapshotErr)
			}
			update.Description = prev.GetDescription()
			if update.Config == nil && prev.GetConfig() != nil {
				update.Config = &agentsv1.AgentConfig{}
			}
			if update.Config != nil {
				update.Config.Instruction = prev.GetConfig().GetInstruction()
				update.Config.GlobalInstruction = prev.GetConfig().GetGlobalInstruction()
			}
		}
	}

	// V2 contract: embedded sub_agents are read-only legacy state. A legacy
	// agent can still be edited (its stored tree round-trips unchanged), but
	// mutating the embedded tree requires migrating to child_agent_ids.
	if !equalSubAgents(update.GetSubAgents(), prev.GetSubAgents()) {
		return nil, connectx.InvalidArgument("sub_agents",
			"is no longer writable; run MigrateAgentsV2 and compose via child_agent_ids")
	}

	if len(update.GetChildAgentIds()) > 0 {
		pool, err := s.repo.ListAgents(ctx, wsID)
		if err != nil {
			return nil, toConnectError(err)
		}
		update.WorkspaceId = wsID
		for i, a := range pool {
			if a.GetName() == update.GetName() {
				pool[i] = update
				break
			}
		}
		if err := internalagent.ValidateAgentRelationships(update, pool); err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
	}

	logger.Info("updating agent", "workspace_id", wsID, "agent", update.GetName())

	var applied *agentsv1.Agent
	a, err := mutateWithRuntime(
		func() (*agentsv1.Agent, error) {
			var applyErr error
			applied, applyErr = s.repo.UpdateAgentCAS(ctx, wsID, update, prev.GetVersion())
			return applied, applyErr
		},
		func() error {
			return s.reloadRuntime(ctx)
		},
		func() error {
			expected := prev.GetVersion() + 1
			if applied != nil {
				expected = applied.GetVersion()
			}
			if _, err := s.repo.UpdateAgentCAS(ctx, wsID, proto.Clone(prev).(*agentsv1.Agent), expected); err != nil {
				return err
			}
			return s.reloadRuntime(ctx)
		},
	)
	if err != nil {
		logger.Error("update agent failed", "workspace_id", wsID, "agent", req.Msg.GetAgent().GetName(), "err", err)
		if errors.Is(err, configrepo.ErrVersionConflict) {
			return nil, connect.NewError(connect.CodeAborted,
				errors.New("agent version changed since it was read; re-read and retry"))
		}
		return nil, toConnectError(err)
	}
	logger.Info("agent updated", "workspace_id", wsID, "agent", a.GetName())
	if err := s.overlayActiveContent(ctx, wsID, a); err != nil {
		return nil, connectx.InternalWith(err)
	}
	return connect.NewResponse(&agentsv1.UpdateAgentResponse{Agent: a}), nil
}

func (s *AgentServiceServer) DeleteAgent(ctx context.Context, req *connect.Request[agentsv1.DeleteAgentRequest]) (*connect.Response[agentsv1.DeleteAgentResponse], error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	logger := log.FromContext(ctx)

	var prev *agentsv1.Agent
	if aid := req.Msg.GetAgentId(); aid != "" {
		prev, err = s.repo.GetAgentByID(ctx, wsID, aid)
	} else {
		prev, err = s.repo.GetAgent(ctx, wsID, req.Msg.GetName())
	}
	if err != nil {
		return nil, toConnectError(err)
	}
	agentName := prev.GetName()
	coord := s.coordinator()
	if coord != nil && req.Msg.GetOperationId() != "" {
		reqJSON, jsonErr := deleteOperationRequestJSON(prev)
		if jsonErr != nil {
			return nil, connectx.InternalWith(jsonErr)
		}
		_, op, _, found, resumeErr := coord.resumeExisting(ctx, wsID, req.Msg.GetOperationId(),
			agentsv1.AgentOperationType_AGENT_OPERATION_TYPE_DELETE, prev.GetAgentId(), reqJSON, 0, "")
		if found {
			if resumeErr != nil {
				return nil, operationError(resumeErr, op)
			}
			return connect.NewResponse(&agentsv1.DeleteAgentResponse{Operation: op}), nil
		}
	}

	if agentID := prev.GetAgentId(); agentID != "" {
		pool, err := s.repo.ListAgents(ctx, wsID)
		if err != nil {
			return nil, toConnectError(err)
		}
		if err := internalagent.ValidateNoOrphanedReferences(agentID, pool); err != nil {
			return nil, connect.NewError(connect.CodeFailedPrecondition, err)
		}
	}

	// Already a tombstone: deleting again is a no-op (idempotent).
	if prev.GetLifecycleStatus() == agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_DELETED {
		return connect.NewResponse(&agentsv1.DeleteAgentResponse{}), nil
	}

	logger.Info("deleting (tombstoning) agent", "workspace_id", wsID, "agent", agentName)

	// Soft delete: with a Saga coordinator the delete is a durable TOMBSTONE
	// operation that flips lifecycle to DELETED while retaining the Agent ID
	// and any Git content (issue #218). Restore reactivates the same entity.
	if coord != nil {
		op, err := coord.RunDelete(ctx, wsID, prev, req.Msg.GetOperationId())
		if err != nil {
			logger.Error("tombstone agent failed", "workspace_id", wsID, "agent", agentName, "err", err)
			return nil, operationError(err, op)
		}
		logger.Info("agent tombstoned", "workspace_id", wsID, "agent", agentName)
		return connect.NewResponse(&agentsv1.DeleteAgentResponse{Operation: op}), nil
	}

	// No coordinator (unbound / legacy wiring): tombstone directly via the
	// single-step mutation seam.
	tombstone := proto.Clone(prev).(*agentsv1.Agent)
	tombstone.LifecycleStatus = agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_DELETED
	tombstone.DeletedAt = timestamppb.New(time.Now())
	var applied *agentsv1.Agent
	_, err = mutateWithRuntime(
		func() (*agentsv1.Agent, error) {
			var applyErr error
			applied, applyErr = s.repo.UpdateAgentCAS(ctx, wsID, tombstone, prev.GetVersion())
			return applied, applyErr
		},
		func() error {
			return s.reloadRuntime(ctx)
		},
		func() error {
			expected := prev.GetVersion() + 1
			if applied != nil {
				expected = applied.GetVersion()
			}
			if _, err := s.repo.UpdateAgentCAS(ctx, wsID, proto.Clone(prev).(*agentsv1.Agent), expected); err != nil {
				return err
			}
			return s.reloadRuntime(ctx)
		},
	)
	if err != nil {
		logger.Error("delete agent failed", "workspace_id", wsID, "agent", agentName, "err", err)
		if errors.Is(err, configrepo.ErrVersionConflict) {
			return nil, connect.NewError(connect.CodeAborted,
				errors.New("agent version changed since it was read; re-read and retry"))
		}
		return nil, toConnectError(err)
	}
	logger.Info("agent tombstoned", "workspace_id", wsID, "agent", agentName)
	return connect.NewResponse(&agentsv1.DeleteAgentResponse{}), nil
}

func (s *AgentServiceServer) InvokeAgent(ctx context.Context, req *connect.Request[agentsv1.InvokeAgentRequest]) (*connect.Response[agentsv1.InvokeAgentResponse], error) {
	if s.runnerSvc == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("runner service not available"))
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
	agentName, err := resolveAgentRunnerRef(s.runnerSvc, wsID, req.Msg.GetAgentId())
	if err != nil {
		return nil, err
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
		"agent", agentName,
		"agent_id", req.Msg.GetAgentId(),
		"app_name", appName,
		"user_id", userID,
		"session_id", sessionID,
		"model_override", req.Msg.GetModelOverride(),
		"input_len", len(req.Msg.GetInput()),
	)
	parts := []*genai.Part{{Text: req.Msg.GetInput()}}
	start := time.Now()
	response, err := s.runnerSvc.Run(ctx, agentName, parts, req.Msg.GetModelOverride(), ctxInfo, nil, nil)
	if err != nil {
		logger.Error("agent invocation failed",
			"workspace_id", wsID,
			"agent", agentName,
			"session_id", sessionID,
			"elapsed_ms", time.Since(start).Milliseconds(),
			"err", err,
		)
		return nil, connectx.InternalWith(err)
	}
	logger.Info("agent invocation completed",
		"workspace_id", wsID,
		"agent", agentName,
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
		AgentID:     req.Msg.GetAgentId(),
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

	var inv *agentsv1.Invocation
	if s.invRepo != nil {
		stored, err := getInvocation(ctx, s.invRepo, wsID, req.Msg.GetInvocationId())
		if err == nil {
			if err := authorizeInvocationAccess(ctx, wsID, stored); err != nil {
				return nil, err
			}
			inv = stored
		} else if !errors.Is(err, invocation.ErrNotFound) {
			return nil, connectx.InternalWith(err)
		}
	}

	var cancelled bool
	if inv != nil && inv.GetSource() == "dashboard-async" && s.asyncCoord != nil {
		// The coordinator owns the outer async context. Cancelling only the
		// runner's nested context would make the coordinator classify Stop as a
		// generic failure.
		cancelled = s.asyncCoord.Cancel(req.Msg.GetInvocationId(), wsID)
	}
	if !cancelled {
		cancelled = s.runnerSvc.CancelInvocation(req.Msg.GetInvocationId(), wsID)
	}
	if !cancelled && inv == nil && s.asyncCoord != nil {
		// Compatibility fallback for tests or deployments without an Invocation
		// recorder. Persisted dashboard async Invocations take the authorized path
		// above.
		cancelled = s.asyncCoord.Cancel(req.Msg.GetInvocationId(), wsID)
	}
	log.FromContext(ctx).Info("cancel agent invocation requested",
		"invocation_id", req.Msg.GetInvocationId(),
		"workspace_id", wsID,
		"cancelled", cancelled,
	)
	return connect.NewResponse(&agentsv1.CancelAgentInvocationResponse{Cancelled: cancelled}), nil
}

func (s *AgentServiceServer) GetAgentRuntimeStatus(ctx context.Context, req *connect.Request[agentsv1.GetAgentRuntimeStatusRequest]) (*connect.Response[agentsv1.GetAgentRuntimeStatusResponse], error) {
	if req.Msg.GetAgentId() == "" && req.Msg.GetName() == "" {
		return nil, connectx.RequiredArgument("agent_id")
	}
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	target := statusTarget{name: req.Msg.GetName()}
	if aid := req.Msg.GetAgentId(); aid != "" {
		a, err := s.repo.GetAgentByID(ctx, wsID, aid)
		if err != nil {
			return nil, toConnectError(err)
		}
		target = statusTarget{name: a.GetName(), agentID: aid}
	} else if a, err := s.repo.GetAgent(ctx, wsID, target.name); err == nil {
		// Legacy name lookup: backfill the agent_id when the agent has one.
		target.agentID = a.GetAgentId()
	}
	statuses, err := s.runtimeStatuses(ctx, wsID, []statusTarget{target})
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
	agents, err := s.repo.ListAgents(ctx, wsID)
	if err != nil {
		return nil, toConnectError(err)
	}
	byID := make(map[string]*agentsv1.Agent, len(agents))
	byName := make(map[string]*agentsv1.Agent, len(agents))
	for _, a := range agents {
		if id := a.GetAgentId(); id != "" {
			byID[id] = a
		}
		byName[a.GetName()] = a
	}

	var targets []statusTarget
	switch {
	case len(req.Msg.GetAgentIds()) > 0:
		for _, id := range req.Msg.GetAgentIds() {
			a, ok := byID[id]
			if !ok {
				return nil, connectx.NotFound(fmt.Sprintf("agent %q not found", id))
			}
			targets = append(targets, statusTarget{name: a.GetName(), agentID: id})
		}
	case len(req.Msg.GetNames()) > 0:
		for _, name := range req.Msg.GetNames() {
			t := statusTarget{name: name}
			if a, ok := byName[name]; ok {
				t.agentID = a.GetAgentId()
			}
			targets = append(targets, t)
		}
	default:
		for _, a := range agents {
			targets = append(targets, statusTarget{name: a.GetName(), agentID: a.GetAgentId()})
		}
	}
	statuses, err := s.runtimeStatuses(ctx, wsID, targets)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&agentsv1.ListAgentRuntimeStatusesResponse{Statuses: statuses}), nil
}

// statusTarget names one agent whose runtime status is being derived: the
// runtime (legacy) name keys the invocation summaries, the agent_id is echoed
// on the status row so clients can key rows by the immutable identifier.
type statusTarget struct {
	name    string
	agentID string
}

// runtimeStatuses derives AgentRuntimeStatus for the target agents from a
// single invocation repo query. Agents with no invocations (or when the repo
// is not wired) are reported as IDLE. The result preserves the target order.
func (s *AgentServiceServer) runtimeStatuses(ctx context.Context, workspaceID string, targets []statusTarget) ([]*agentsv1.AgentRuntimeStatus, error) {
	out := make([]*agentsv1.AgentRuntimeStatus, len(targets))
	names := make([]string, len(targets))
	for i, t := range targets {
		names[i] = t.name
		out[i] = &agentsv1.AgentRuntimeStatus{
			Name:    t.name,
			AgentId: t.agentID,
			State:   agentsv1.AgentRuntimeState_AGENT_RUNTIME_STATE_IDLE,
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

func (s *AgentServiceServer) AssignAgentID(ctx context.Context, req *connect.Request[agentsv1.AssignAgentIDRequest]) (*connect.Response[agentsv1.AssignAgentIDResponse], error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.requireOwnerOrAdmin(ctx, wsID); err != nil {
		return nil, err
	}
	name := req.Msg.GetName()
	if name == "" {
		return nil, connectx.RequiredArgument("name")
	}
	agentID := req.Msg.GetAgentId()
	if err := internalagent.ValidateAgentID(agentID); err != nil {
		return nil, connectx.InvalidArgument("agent_id", err.Error())
	}

	logger := log.FromContext(ctx)

	existing, err := s.repo.GetAgent(ctx, wsID, name)
	if err != nil {
		return nil, toConnectError(err)
	}
	if existing.GetAgentId() != "" {
		return nil, connect.NewError(connect.CodeFailedPrecondition,
			fmt.Errorf("agent %q already has agent_id %q; agent IDs are immutable once assigned", name, existing.GetAgentId()))
	}

	taken, err := s.repo.AgentIDExists(ctx, wsID, agentID)
	if err != nil {
		return nil, connectx.InternalWith(err)
	}
	if taken {
		return nil, connect.NewError(connect.CodeAlreadyExists,
			fmt.Errorf("agent_id %q is already in use in this workspace", agentID))
	}

	updated := proto.Clone(existing).(*agentsv1.Agent)
	updated.AgentId = agentID

	result, err := mutateWithRuntime(
		func() (*agentsv1.Agent, error) {
			return s.repo.UpdateAgent(ctx, wsID, updated)
		},
		func() error {
			return s.reloadRuntime(ctx)
		},
		func() error {
			rollback := proto.Clone(existing).(*agentsv1.Agent)
			if _, err := s.repo.UpdateAgent(ctx, wsID, rollback); err != nil {
				return err
			}
			return s.reloadRuntime(ctx)
		},
	)
	if err != nil {
		if errors.Is(err, configrepo.ErrAlreadyExists) {
			return nil, connect.NewError(connect.CodeAlreadyExists,
				fmt.Errorf("agent_id %q is already in use in this workspace", agentID))
		}
		logger.Error("assign agent id failed", "workspace_id", wsID, "agent", name, "agent_id", agentID, "err", err)
		return nil, toConnectError(err)
	}
	logger.Info("agent id assigned", "workspace_id", wsID, "agent", name, "agent_id", agentID)
	return connect.NewResponse(&agentsv1.AssignAgentIDResponse{Agent: result}), nil
}

func (s *AgentServiceServer) GetMigrationReadiness(ctx context.Context, _ *connect.Request[agentsv1.GetMigrationReadinessRequest]) (*connect.Response[agentsv1.GetMigrationReadinessResponse], error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}

	agents, err := s.repo.ListAgents(ctx, wsID)
	if err != nil {
		return nil, toConnectError(err)
	}

	idSet := make(map[string]int)
	for _, a := range agents {
		if id := a.GetAgentId(); id != "" {
			idSet[id]++
		}
	}

	agentsByName := make(map[string]*agentsv1.Agent, len(agents))
	for _, a := range agents {
		agentsByName[a.GetName()] = a
	}

	statuses := make([]*agentsv1.AgentMigrationStatus, 0, len(agents))
	for _, a := range agents {
		status := &agentsv1.AgentMigrationStatus{
			Name:    a.GetName(),
			AgentId: a.GetAgentId(),
		}

		switch {
		case a.GetAgentId() == "":
			status.Readiness = agentsv1.MigrationReadiness_MIGRATION_READINESS_MISSING_ID
			status.Detail = "agent has not been assigned an Agent ID"
		case idSet[a.GetAgentId()] > 1:
			status.Readiness = agentsv1.MigrationReadiness_MIGRATION_READINESS_CONFLICT
			status.Detail = fmt.Sprintf("agent_id %q is used by %d agents", a.GetAgentId(), idSet[a.GetAgentId()])
		default:
			if detail := checkSubAgentDeps(a, agentsByName); detail != "" {
				status.Readiness = agentsv1.MigrationReadiness_MIGRATION_READINESS_INCOMPLETE_DEPS
				status.Detail = detail
			} else {
				status.Readiness = agentsv1.MigrationReadiness_MIGRATION_READINESS_READY
			}
		}

		statuses = append(statuses, status)
	}

	return connect.NewResponse(&agentsv1.GetMigrationReadinessResponse{Statuses: statuses}), nil
}

// checkSubAgentDeps verifies that all sub-agents of a have agent IDs assigned
// and exist in the workspace. Returns a human-readable detail string if any
// dependency is unresolved, or empty string if all dependencies are ready.
func checkSubAgentDeps(a *agentsv1.Agent, byName map[string]*agentsv1.Agent) string {
	for _, sub := range a.GetSubAgents() {
		dep, ok := byName[sub.GetName()]
		if !ok {
			return fmt.Sprintf("sub-agent %q not found in workspace", sub.GetName())
		}
		if dep.GetAgentId() == "" {
			return fmt.Sprintf("sub-agent %q is missing an Agent ID", sub.GetName())
		}
	}
	if wf := a.GetConfig().GetWorkflow(); wf != nil {
		for _, node := range wf.GetNodes() {
			if node.GetKind() != agentsv1.WorkflowNodeKind_WORKFLOW_NODE_KIND_AGENT {
				continue
			}
			agentRef := node.GetAgent()
			dep, ok := byName[agentRef]
			if !ok {
				return fmt.Sprintf("workflow node %q references agent %q which is not found in workspace", node.GetName(), agentRef)
			}
			if dep.GetAgentId() == "" {
				return fmt.Sprintf("workflow node %q references agent %q which is missing an Agent ID", node.GetName(), agentRef)
			}
		}
	}
	return ""
}

// ── Agent lifecycle Saga RPCs (issue #218) ───────────────────────────────────

// agentTypeRequiresPrompt reports whether an agent type requires a non-empty
// prompt (LLM agents; UNSPECIFIED defaults to LLM).
func agentTypeRequiresPrompt(t agentsv1.AgentType) bool {
	return t == agentsv1.AgentType_AGENT_TYPE_LLM || t == agentsv1.AgentType_AGENT_TYPE_UNSPECIFIED
}

// contentValidationError wraps content validation messages as a
// FailedPrecondition Connect error.
func contentValidationError(vErrs []string) error {
	return connect.NewError(connect.CodeFailedPrecondition,
		fmt.Errorf("agent content invalid: %s", strings.Join(vErrs, "; ")))
}

func operationError(err error, op *agentsv1.AgentOperation) *connect.Error {
	cerr := toConnectError(err)
	if op == nil {
		return cerr
	}
	detail, detailErr := connect.NewErrorDetail(op)
	if detailErr == nil {
		cerr.AddDetail(detail)
	}
	return cerr
}

func (s *AgentServiceServer) requireCoordinator() (*agentOperationCoordinator, error) {
	coord := s.coordinator()
	if coord == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition,
			errors.New("agent lifecycle operations are not available (no repository binding configured)"))
	}
	return coord, nil
}

// UpdateAgentConfiguration is the composite-save command: it coordinates an
// operational config patch (DB) with Agent Content changes (Git) under one
// durable operation, with optimistic concurrency on the agent version.
func (s *AgentServiceServer) UpdateAgentConfiguration(ctx context.Context, req *connect.Request[agentsv1.UpdateAgentConfigurationRequest]) (*connect.Response[agentsv1.UpdateAgentConfigurationResponse], error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.requireOwnerOrAdmin(ctx, wsID); err != nil {
		return nil, err
	}
	patch := req.Msg.GetAgentPatch()
	if patch.GetAgentId() == "" {
		return nil, connectx.RequiredArgument("agent_patch.agent_id")
	}
	if err := internalagent.ValidateWorkflowAgent(patch); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	coord, err := s.requireCoordinator()
	if err != nil {
		return nil, err
	}
	reqJSON, err := updateOperationRequestJSON(patch, req.Msg.GetContentChanges(), req.Msg.GetExpectedAgentVersion(), req.Msg.GetBaseCommitSha())
	if err != nil {
		return nil, connectx.InternalWith(err)
	}
	agent, op, vErrs, found, err := coord.resumeExisting(ctx, wsID, req.Msg.GetOperationId(),
		agentsv1.AgentOperationType_AGENT_OPERATION_TYPE_UPDATE_CONFIGURATION, patch.GetAgentId(), reqJSON,
		req.Msg.GetExpectedAgentVersion(), req.Msg.GetBaseCommitSha())
	if found {
		if err != nil {
			return nil, operationError(err, op)
		}
		return connect.NewResponse(&agentsv1.UpdateAgentConfigurationResponse{
			Agent: agent, Operation: op, ValidationErrors: vErrs,
		}), nil
	}
	prev, err := s.repo.GetAgentByID(ctx, wsID, patch.GetAgentId())
	if err != nil {
		return nil, operationError(err, op)
	}
	if prev.GetLifecycleStatus() == agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_DELETED {
		return nil, connect.NewError(connect.CodeFailedPrecondition,
			errors.New("agent is deleted; restore it before updating its configuration"))
	}
	// Fast-fail on a stale version before doing any Git work; the DB patch step
	// re-checks atomically via UpdateAgentCAS.
	if prev.GetVersion() != req.Msg.GetExpectedAgentVersion() {
		return nil, connect.NewError(connect.CodeAborted,
			fmt.Errorf("agent version changed since it was read (have %d, expected %d); re-read and retry",
				prev.GetVersion(), req.Msg.GetExpectedAgentVersion()))
	}
	if len(patch.GetChildAgentIds()) > 0 {
		pool, err := s.repo.ListAgents(ctx, wsID)
		if err != nil {
			return nil, toConnectError(err)
		}
		check := proto.Clone(patch).(*agentsv1.Agent)
		check.Name = prev.GetName()
		check.WorkspaceId = wsID
		if err := internalagent.ValidateAgentRelationships(check, pool); err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
	}

	agent, op, vErrs, err = coord.RunUpdateConfiguration(ctx, wsID, prev, patch,
		req.Msg.GetContentChanges(), req.Msg.GetExpectedAgentVersion(), req.Msg.GetBaseCommitSha(), req.Msg.GetOperationId())
	if err != nil {
		return nil, operationError(err, op)
	}
	return connect.NewResponse(&agentsv1.UpdateAgentConfigurationResponse{
		Agent: agent, Operation: op, ValidationErrors: vErrs,
	}), nil
}

// RestoreAgent reactivates a tombstoned agent from its retained configuration
// and Agent Content.
func (s *AgentServiceServer) RestoreAgent(ctx context.Context, req *connect.Request[agentsv1.RestoreAgentRequest]) (*connect.Response[agentsv1.RestoreAgentResponse], error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.requireOwnerOrAdmin(ctx, wsID); err != nil {
		return nil, err
	}
	if req.Msg.GetAgentId() == "" {
		return nil, connectx.RequiredArgument("agent_id")
	}
	coord, err := s.requireCoordinator()
	if err != nil {
		return nil, err
	}
	reqJSON, err := restoreOperationRequestJSON(req.Msg.GetAgentId())
	if err != nil {
		return nil, connectx.InternalWith(err)
	}
	agent, op, vErrs, found, err := coord.resumeExisting(ctx, wsID, req.Msg.GetOperationId(),
		agentsv1.AgentOperationType_AGENT_OPERATION_TYPE_RESTORE, req.Msg.GetAgentId(), reqJSON, 0, "")
	if found {
		if err != nil {
			return nil, operationError(err, op)
		}
		return connect.NewResponse(&agentsv1.RestoreAgentResponse{
			Agent: agent, Operation: op, ValidationErrors: vErrs,
		}), nil
	}
	prev, err := s.repo.GetAgentByID(ctx, wsID, req.Msg.GetAgentId())
	if err != nil {
		return nil, operationError(err, op)
	}
	if prev.GetLifecycleStatus() != agentsv1.AgentLifecycleStatus_AGENT_LIFECYCLE_STATUS_DELETED {
		return nil, connect.NewError(connect.CodeFailedPrecondition,
			errors.New("agent is not deleted; nothing to restore"))
	}
	agent, op, vErrs, err = coord.RunRestore(ctx, wsID, prev, req.Msg.GetOperationId())
	if err != nil {
		return nil, operationError(err, op)
	}
	return connect.NewResponse(&agentsv1.RestoreAgentResponse{
		Agent: agent, Operation: op, ValidationErrors: vErrs,
	}), nil
}

// GetAgentOperation returns a durable lifecycle operation record.
func (s *AgentServiceServer) GetAgentOperation(ctx context.Context, req *connect.Request[agentsv1.GetAgentOperationRequest]) (*connect.Response[agentsv1.GetAgentOperationResponse], error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if s.opRepo == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("agent operations are not available"))
	}
	op, err := s.opRepo.Get(ctx, wsID, req.Msg.GetOperationId())
	if err != nil {
		if errors.Is(err, agentoprepo.ErrNotFound) {
			return nil, connectx.NotFound("operation not found")
		}
		return nil, connectx.InternalWith(err)
	}
	return connect.NewResponse(&agentsv1.GetAgentOperationResponse{Operation: op}), nil
}

// ListAgentOperations lists lifecycle operations in the workspace.
func (s *AgentServiceServer) ListAgentOperations(ctx context.Context, req *connect.Request[agentsv1.ListAgentOperationsRequest]) (*connect.Response[agentsv1.ListAgentOperationsResponse], error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if s.opRepo == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("agent operations are not available"))
	}
	ops, next, err := s.opRepo.List(ctx, wsID, req.Msg.GetStatus(), req.Msg.GetPageSize(), req.Msg.GetPageToken())
	if err != nil {
		return nil, connectx.InternalWith(err)
	}
	return connect.NewResponse(&agentsv1.ListAgentOperationsResponse{Operations: ops, NextPageToken: next}), nil
}

// RetryAgentOperation resumes a FAILED lifecycle operation.
func (s *AgentServiceServer) RetryAgentOperation(ctx context.Context, req *connect.Request[agentsv1.RetryAgentOperationRequest]) (*connect.Response[agentsv1.RetryAgentOperationResponse], error) {
	wsID, err := requireWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.requireOwnerOrAdmin(ctx, wsID); err != nil {
		return nil, err
	}
	coord, err := s.requireCoordinator()
	if err != nil {
		return nil, err
	}
	agent, op, vErrs, err := coord.Retry(ctx, wsID, req.Msg.GetOperationId())
	if err != nil {
		return nil, operationError(err, op)
	}
	return connect.NewResponse(&agentsv1.RetryAgentOperationResponse{
		Agent: agent, Operation: op, ValidationErrors: vErrs,
	}), nil
}

// requireOwnerOrAdmin checks that the caller is a global admin or holds
// the "owner" or "admin" role in the workspace.
func (s *AgentServiceServer) requireOwnerOrAdmin(ctx context.Context, workspaceID string) error {
	if auth.IsAdmin(ctx) {
		return nil
	}
	if s.wsRepo == nil {
		return connect.NewError(connect.CodeFailedPrecondition,
			errors.New("workspace store not available"))
	}
	user, ok := auth.UserFromContext(ctx)
	if !ok {
		return connect.NewError(connect.CodeUnauthenticated, errors.New("authentication required"))
	}
	member, err := s.wsRepo.GetMember(ctx, workspaceID, user.GetId())
	if err != nil {
		if errors.Is(err, workspacerepo.ErrNotFound) {
			return connect.NewError(connect.CodePermissionDenied, errors.New("insufficient workspace role"))
		}
		return connectx.InternalWith(err)
	}
	role := member.GetRole()
	if role == "owner" || role == "admin" {
		return nil
	}
	return connect.NewError(connect.CodePermissionDenied, errors.New("only workspace owners and administrators can assign Agent IDs"))
}

func (s *AgentServiceServer) MigrateAgentsV2(ctx context.Context, req *connect.Request[agentsv1.MigrateAgentsV2Request]) (*connect.Response[agentsv1.MigrateAgentsV2Response], error) {
	return migrateAgentsV2(ctx, s, req)
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
