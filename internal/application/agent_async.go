package application

import (
	"context"
	"errors"

	"butterfly.orx.me/core/log"
	"connectrpc.com/connect"
	"google.golang.org/genai"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/repo/auth"
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
	Enqueue(inv *agentsv1.Invocation, agentName string, parts []*genai.Part, modelOverride string)
	Cancel(invocationID, workspaceID string) bool
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

	// Validate/convert input parts.
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

	sessionID := req.Msg.GetSessionId()
	sessionCreated := false

	// Enforce single-active-invocation per session if session exists.
	if sessionID != "" {
		active, activeErr := s.invRepo.FindActiveBySession(ctx, wsID, sessionID)
		if activeErr == nil && active != nil {
			return nil, connect.NewError(connect.CodeFailedPrecondition,
				errors.New("session has an active invocation: "+active.GetId()))
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
		Source:           "dashboard-async",
		RequestId:        req.Msg.GetRequestId(),
		WorkspaceId:      wsID,
		StartedAt:        timestamppb.Now(),
	}

	if err := s.invRepo.Save(ctx, inv); err != nil {
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
	)

	// Enqueue for background execution.
	s.asyncCoord.Enqueue(inv, agentName, parts, req.Msg.GetModelOverride())

	return connect.NewResponse(&agentsv1.SubmitAgentInvocationResponse{
		SessionId:      sessionID,
		InvocationId:   invID,
		Status:         agentsv1.InvocationStatus_INVOCATION_STATUS_QUEUED,
		SessionCreated: sessionCreated,
	}), nil
}

// GetAgentInvocation returns the authoritative state of one invocation.
func (s *AgentServiceServer) GetAgentInvocation(ctx context.Context, req *connect.Request[agentsv1.GetAgentInvocationRequest]) (*connect.Response[agentsv1.GetAgentInvocationResponse], error) {
	if s.invRepo == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("invocation repository not available"))
	}
	if req.Msg.GetInvocationId() == "" {
		return nil, connectx.RequiredArgument("invocation_id")
	}

	wsID, hasWorkspace := workspace.FromContext(ctx)
	if !hasWorkspace && !auth.IsAdmin(ctx) {
		return nil, connect.NewError(connect.CodeFailedPrecondition,
			errors.New("workspace required (set X-Workspace-ID header)"))
	}

	inv, err := s.invRepo.Get(ctx, req.Msg.GetInvocationId())
	if err != nil {
		if errors.Is(err, invocation.ErrNotFound) {
			return nil, connectx.NotFound("invocation not found")
		}
		return nil, connectx.InternalWith(err)
	}

	// Workspace scope check: non-admin callers can only see invocations in
	// their workspace.
	if wsID != "" && inv.GetWorkspaceId() != wsID {
		return nil, connectx.NotFound("invocation not found")
	}

	return connect.NewResponse(&agentsv1.GetAgentInvocationResponse{
		Invocation: inv,
	}), nil
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
