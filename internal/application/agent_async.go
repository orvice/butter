package application

import (
	"context"
	"errors"

	"butterfly.orx.me/core/log"
	"connectrpc.com/connect"
	"google.golang.org/genai"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/repo/auth"
	"go.orx.me/apps/butter/internal/repo/inputpart"
	"go.orx.me/apps/butter/internal/repo/invocation"
	"go.orx.me/apps/butter/internal/runtime/asyncrun"
	"go.orx.me/apps/butter/internal/transport/connectx"
	"go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	adksession "google.golang.org/adk/v2/session"
)

// asyncCoordinator is the subset of *asyncrun.Coordinator used by the RPC
// handlers; allows test substitution.
type asyncCoordinator interface {
	Enqueue(inv *agentsv1.Invocation, agentName string, modelOverride string)
	Cancel(invocationID, workspaceID string) bool
	Watch(invocationID string) (<-chan asyncrun.Frame, func())
}

// SetAsyncCoordinator wires the background execution coordinator.
func (s *AgentServiceServer) SetAsyncCoordinator(c asyncCoordinator) {
	s.asyncCoord = c
}

// SetSessionService wires the ADK session service for creating workspace-
// owned sessions during async submission.
func (s *AgentServiceServer) SetSessionService(svc adksession.Service) {
	s.sessionSvc = svc
}

// SubmitAgentInvocation durably accepts one dashboard chat turn and returns
// without waiting for execution. The runner executes independently.
func (s *AgentServiceServer) SubmitAgentInvocation(ctx context.Context, req *connect.Request[agentsv1.SubmitAgentInvocationRequest]) (*connect.Response[agentsv1.SubmitAgentInvocationResponse], error) {
	if s.runnerSvc == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("runner service not available"))
	}
	if s.asyncCoord == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("async coordinator not available"))
	}
	if s.invRepo == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("invocation repository not available"))
	}
	if s.inputPartRepo == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("input part repository not available"))
	}

	// Validate required fields.
	if req.Msg.GetRequestId() == "" {
		return nil, connectx.RequiredArgument("request_id")
	}
	if req.Msg.GetAgentId() == "" {
		return nil, connectx.RequiredArgument("agent_id")
	}
	if len(req.Msg.GetParts()) == 0 && req.Msg.GetMessage() == "" {
		return nil, connectx.RequiredArgument("message")
	}

	// Require workspace context.
	wsID, hasWorkspace := workspace.FromContext(ctx)
	if !hasWorkspace && !auth.IsAdmin(ctx) {
		return nil, connect.NewError(connect.CodeFailedPrecondition,
			errors.New("workspace required (set X-Workspace-ID header)"))
	}

	// Resolve agent.
	agentName, err := resolveAgentRunnerRef(s.runnerSvc, wsID, req.Msg.GetAgentId())
	if err != nil {
		return nil, err
	}
	agentID, displayName, _ := s.runnerSvc.GetAgentIdentity(agentName)

	// Validate/convert input parts (validates limits and MIME types).
	parts, err := resolveUserParts(req.Msg.GetParts(), req.Msg.GetMessage())
	if err != nil {
		return nil, err
	}

	// Determine the authenticated user.
	user, hasUser := auth.UserFromContext(ctx)
	if !hasUser {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("authentication required"))
	}
	userID := user.GetId()

	// The first release is explicitly single-instance. Serialize only the
	// durable accept transaction so two browser tabs cannot both pass the
	// active-Invocation check before either QUEUED record is visible.
	s.asyncSubmitMu.Lock()
	defer s.asyncSubmitMu.Unlock()

	// Idempotency check: if this request_id already exists, return original.
	existing, findErr := s.invRepo.FindByRequestID(ctx, wsID, req.Msg.GetRequestId())
	if findErr == nil && existing != nil {
		return connect.NewResponse(&agentsv1.SubmitAgentInvocationResponse{
			SessionId:      existing.GetSessionId(),
			InvocationId:   existing.GetId(),
			Status:         existing.GetStatus(),
			SessionCreated: false,
		}), nil
	}
	if findErr != nil && !errors.Is(findErr, invocation.ErrNotFound) {
		return nil, connectx.InternalWith(findErr)
	}

	sessionID := req.Msg.GetSessionId()
	sessionCreated := false

	// Validate an existing private Session before checking its active
	// Invocation. The authenticated user supplies no user_id in this command,
	// so another user's Session is intentionally indistinguishable from a
	// missing one.
	if sessionID != "" {
		if s.sessionSvc == nil {
			return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("session service not available"))
		}
		sessResp, sessErr := s.sessionSvc.Get(ctx, &adksession.GetRequest{
			AppName:   "web-chat",
			UserID:    userID,
			SessionID: sessionID,
		})
		if sessErr != nil || sessResp == nil || sessResp.Session == nil {
			return nil, connectx.NotFound("session not found")
		}
		if value, stateErr := sessResp.Session.State().Get("workspace_id"); stateErr != nil || value != wsID {
			return nil, connectx.NotFound("session not found")
		}
		if value, stateErr := sessResp.Session.State().Get("agent_id"); stateErr == nil && value != "" && value != agentID {
			return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("session is bound to another agent"))
		}
		if value, stateErr := sessResp.Session.State().Get("agent_name"); stateErr == nil && value != "" && value != agentName {
			return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("session is bound to another agent"))
		}

		active, activeErr := s.invRepo.FindActiveBySession(ctx, wsID, sessionID)
		if activeErr == nil && active != nil {
			return nil, activeInvocationError(active.GetId())
		}
		if activeErr != nil && !errors.Is(activeErr, invocation.ErrNotFound) {
			return nil, connectx.InternalWith(activeErr)
		}
	}

	// Create session if needed.
	if sessionID == "" {
		if s.sessionSvc == nil {
			return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("session service not available"))
		}
		sessResp, sessErr := s.sessionSvc.Create(ctx, &adksession.CreateRequest{
			AppName:   "web-chat",
			UserID:    userID,
			SessionID: "chat-" + asyncrun.NewInvocationID(),
			State: map[string]any{
				"agent_name":   agentName,
				"agent_id":     agentID,
				"workspace_id": wsID,
			},
		})
		if sessErr != nil {
			return nil, connectx.InternalWith(sessErr)
		}
		sessionID = sessResp.Session.ID()
		sessionCreated = true
	}

	// Build and persist the QUEUED invocation.
	invID := asyncrun.NewInvocationID()
	inv := &agentsv1.Invocation{
		Id:               invID,
		AgentName:        agentName,
		AgentId:          agentID,
		AgentDisplayName: displayName,
		AppName:          "web-chat",
		UserId:           userID,
		SessionId:        sessionID,
		Status:           agentsv1.InvocationStatus_INVOCATION_STATUS_QUEUED,
		Input:            extractTextInput(parts),
		Source:           sourceDashboardAsync,
		RequestId:        req.Msg.GetRequestId(),
		WorkspaceId:      wsID,
		StartedAt:        timestamppb.Now(),
	}

	if err := s.invRepo.Save(ctx, inv); err != nil {
		return nil, connectx.InternalWith(err)
	}

	// Persist input parts durably before enqueuing. This guarantees that the
	// response means every validated input part is recoverable even if the
	// process crashes before execution starts.
	protoParts := buildProtoParts(req.Msg.GetParts(), req.Msg.GetMessage())
	if err := s.inputPartRepo.SaveAll(ctx, invID, protoParts); err != nil {
		return nil, connectx.InternalWith(err)
	}

	logger := log.FromContext(ctx)
	logger.Info("async invocation submitted",
		"invocation_id", invID,
		"workspace_id", wsID,
		"agent", agentName,
		"agent_id", agentID,
		"session_id", sessionID,
		"request_id", req.Msg.GetRequestId(),
		"session_created", sessionCreated,
		"input_parts", len(protoParts),
	)

	// Enqueue for background execution.
	s.asyncCoord.Enqueue(inv, agentName, req.Msg.GetModelOverride())

	return connect.NewResponse(&agentsv1.SubmitAgentInvocationResponse{
		SessionId:      sessionID,
		InvocationId:   invID,
		Status:         agentsv1.InvocationStatus_INVOCATION_STATUS_QUEUED,
		SessionCreated: sessionCreated,
	}), nil
}

// buildProtoParts constructs the canonical InputPart slice for persistence.
// When parts is non-empty it is used directly; otherwise the text message is
// wrapped as a single text part.
func buildProtoParts(parts []*agentsv1.InputPart, message string) []*agentsv1.InputPart {
	if len(parts) > 0 {
		return parts
	}
	return []*agentsv1.InputPart{
		{Part: &agentsv1.InputPart_Text{Text: message}},
	}
}

// GetAgentInvocation returns the authoritative state of one invocation.
// When invocation_id is empty and session_id is set, it looks up the
// session's active (QUEUED/RUNNING) invocation instead — the reconnect path
// clients use to decide whether to attach a WatchAgentInvocation observer —
// or, with `latest`, the session's most recent invocation regardless of
// status (inline failed/stopped turn rendering after reload). With
// `include_input_parts` the retained Input Parts of a FAILED/CANCELLED
// invocation are returned so the original input can be restored for an
// explicit resubmission.
func (s *AgentServiceServer) GetAgentInvocation(ctx context.Context, req *connect.Request[agentsv1.GetAgentInvocationRequest]) (*connect.Response[agentsv1.GetAgentInvocationResponse], error) {
	if s.invRepo == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("invocation repository not available"))
	}
	if req.Msg.GetInvocationId() == "" && req.Msg.GetSessionId() == "" {
		return nil, connectx.RequiredArgument("invocation_id")
	}

	wsID, hasWorkspace := workspace.FromContext(ctx)
	if !hasWorkspace && !auth.IsAdmin(ctx) {
		return nil, connect.NewError(connect.CodeFailedPrecondition,
			errors.New("workspace required (set X-Workspace-ID header)"))
	}

	var inv *agentsv1.Invocation
	var err error
	switch {
	case req.Msg.GetInvocationId() != "":
		inv, err = getInvocation(ctx, s.invRepo, wsID, req.Msg.GetInvocationId())
	case req.Msg.GetLatest():
		inv, err = s.invRepo.FindLatestBySession(ctx, wsID, req.Msg.GetSessionId())
	default:
		inv, err = s.invRepo.FindActiveBySession(ctx, wsID, req.Msg.GetSessionId())
	}
	if err != nil {
		if errors.Is(err, invocation.ErrNotFound) {
			return nil, connectx.NotFound("invocation not found")
		}
		return nil, connectx.InternalWith(err)
	}
	if inv == nil {
		return nil, connectx.NotFound("invocation not found")
	}

	if err := authorizeInvocationAccess(ctx, wsID, inv); err != nil {
		return nil, err
	}

	resp := &agentsv1.GetAgentInvocationResponse{Invocation: inv}
	if req.Msg.GetIncludeInputParts() && s.inputPartRepo != nil {
		// Parts are retained for FAILED/CANCELLED invocations so the user can
		// restore the original input for explicit resubmission; a successful
		// run's parts were cleaned up and simply come back empty.
		parts, partsErr := s.inputPartRepo.Load(ctx, inv.GetId())
		if partsErr == nil {
			resp.InputParts = parts
		} else if !errors.Is(partsErr, inputpart.ErrNotFound) {
			return nil, connectx.InternalWith(partsErr)
		}
	}
	return connect.NewResponse(resp), nil
}

func getInvocation(ctx context.Context, repo invocation.Repository, workspaceID, invocationID string) (*agentsv1.Invocation, error) {
	if workspaceID == "" {
		return repo.GetAcrossWorkspaces(ctx, invocationID)
	}
	return repo.Get(ctx, workspaceID, invocationID)
}

func activeInvocationError(invocationID string) *connect.Error {
	err := connect.NewError(connect.CodeFailedPrecondition,
		errors.New("session has an active invocation: "+invocationID))
	err.Meta().Set("active-invocation-id", invocationID)
	return err
}

func authorizeInvocationAccess(ctx context.Context, workspaceID string, inv *agentsv1.Invocation) error {
	if workspaceID != "" && inv.GetWorkspaceId() != workspaceID {
		return connectx.NotFound("invocation not found")
	}
	if auth.IsAdmin(ctx) || inv.GetAppName() != "web-chat" {
		return nil
	}
	user, ok := auth.UserFromContext(ctx)
	if !ok {
		return connect.NewError(connect.CodeUnauthenticated, errors.New("authentication required"))
	}
	if inv.GetUserId() != user.GetId() {
		return connectx.NotFound("invocation not found")
	}
	return nil
}

// extractTextInput returns the text content from parts for persisting as the
// invocation's input field (a summary — not the full multimodal payload).
func extractTextInput(parts []*genai.Part) string {
	for _, p := range parts {
		if p != nil && p.Text != "" {
			return p.Text
		}
	}
	return ""
}
