package application

import (
	"context"
	"encoding/json"
	"errors"

	"butterfly.orx.me/core/log"
	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/repo/auth"
	"go.orx.me/apps/butter/internal/transport/connectx"
	wsctx "go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// StreamAgent runs an agent and streams progress events to the caller.
// The dashboard chat UI uses this in place of the legacy
// POST /api/chat/stream Server-Sent Events handler.
//
// The stream's first message is always a StreamAgentStarted carrying the
// invocation_id; clients use that id with CancelAgentInvocation if the
// user hits "stop". Mid-stream messages are either StreamAgentTextDelta
// (partial assistant text) or StreamAgentRunEvent (a complete ADK
// session.Event that the dashboard parses to render tool calls etc.).
// The terminal message is a StreamAgentFinal carrying the final response.
// On error the stream is closed with a connect.Error — no payload event.
func (s *AgentServiceServer) StreamAgent(
	ctx context.Context,
	req *connect.Request[agentsv1.StreamAgentRequest],
	stream *connect.ServerStream[agentsv1.StreamAgentResponse],
) error {
	if s.runnerSvc == nil {
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("runner service not available"))
	}
	if req.Msg.GetAgentName() == "" {
		return connectx.RequiredArgument("agent_name")
	}
	if req.Msg.GetMessage() == "" {
		return connectx.RequiredArgument("message")
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
		sessionID = "chat-" + uuid.NewString()
	}
	invocationID := uuid.NewString()
	if id, err := uuid.NewV7(); err == nil {
		invocationID = id.String()
	}

	workspaceID, hasWorkspace := wsctx.FromContext(ctx)
	if !hasWorkspace && !auth.IsAdmin(ctx) {
		return connect.NewError(connect.CodeFailedPrecondition,
			errors.New("workspace required (set X-Workspace-ID header)"))
	}
	ctxInfo := &agentsv1.ContextInfo{
		Uuid:        invocationID,
		SessionId:   sessionID,
		UserId:      userID,
		ChannelName: appName,
		Source:      agentsv1.ContextSource_CONTEXT_SOURCE_API,
		ChatType:    agentsv1.ChatType_CHAT_TYPE_PRIVATE,
		WorkspaceId: workspaceID,
	}

	logger := log.FromContext(ctx)
	logger.Info("streaming agent started",
		"workspace_id", workspaceID,
		"agent", req.Msg.GetAgentName(),
		"session_id", sessionID,
		"invocation_id", invocationID,
		"message_len", len(req.Msg.GetMessage()),
	)

	if err := stream.Send(&agentsv1.StreamAgentResponse{
		Event: &agentsv1.StreamAgentResponse_Started{
			Started: &agentsv1.StreamAgentStarted{
				InvocationId: invocationID,
				SessionId:    sessionID,
				AgentName:    req.Msg.GetAgentName(),
			},
		},
	}); err != nil {
		return err
	}

	// connect.ServerStream is not safe for concurrent Send; serialize through
	// a single goroutine consuming an event channel.
	type pendingEvent struct {
		response *agentsv1.StreamAgentResponse
	}
	events := make(chan pendingEvent, 32)
	doneSending := make(chan error, 1)

	go func() {
		var sendErr error
		for evt := range events {
			if sendErr != nil {
				continue
			}
			if err := stream.Send(evt.response); err != nil {
				sendErr = err
			}
		}
		doneSending <- sendErr
	}()

	queue := func(resp *agentsv1.StreamAgentResponse) {
		select {
		case events <- pendingEvent{response: resp}:
		case <-ctx.Done():
		}
	}

	parts := []*genai.Part{genai.NewPartFromText(req.Msg.GetMessage())}
	response, runErr := s.runnerSvc.RunSSE(ctx, req.Msg.GetAgentName(), parts, req.Msg.GetModelOverride(), ctxInfo, func(evt *session.Event) {
		textParts := streamAgentTextParts(evt)
		for _, text := range textParts {
			queue(&agentsv1.StreamAgentResponse{
				Event: &agentsv1.StreamAgentResponse_TextDelta{
					TextDelta: &agentsv1.StreamAgentTextDelta{
						InvocationId: invocationID,
						SessionId:    sessionID,
						AgentName:    req.Msg.GetAgentName(),
						Text:         text,
					},
				},
			})
		}
		if len(textParts) > 0 && streamAgentEventHasOnlyTextParts(evt) {
			return
		}
		queue(&agentsv1.StreamAgentResponse{
			Event: &agentsv1.StreamAgentResponse_RunEvent{
				RunEvent: streamAgentRunEvent(evt, invocationID, sessionID, req.Msg.GetAgentName()),
			},
		})
	}, nil)

	if runErr == nil {
		queue(&agentsv1.StreamAgentResponse{
			Event: &agentsv1.StreamAgentResponse_Final{
				Final: &agentsv1.StreamAgentFinal{
					InvocationId: invocationID,
					SessionId:    sessionID,
					AgentName:    req.Msg.GetAgentName(),
					Response:     response,
				},
			},
		})
	}

	close(events)
	sendErr := <-doneSending

	if runErr != nil {
		logger.Error("streaming agent failed",
			"workspace_id", workspaceID,
			"agent", req.Msg.GetAgentName(),
			"session_id", sessionID,
			"invocation_id", invocationID,
			"err", runErr,
		)
		return streamAgentError(runErr)
	}
	if sendErr != nil {
		// Send failed (likely client disconnected); nothing to surface
		// because the connection is already gone.
		return sendErr
	}
	logger.Info("streaming agent finished",
		"workspace_id", workspaceID,
		"agent", req.Msg.GetAgentName(),
		"session_id", sessionID,
		"invocation_id", invocationID,
	)
	return nil
}

func streamAgentError(err error) error {
	if err == nil {
		return nil
	}
	var connectErr *connect.Error
	if errors.As(err, &connectErr) {
		return connectErr
	}
	if errors.Is(err, context.Canceled) {
		return connect.NewError(connect.CodeCanceled, err)
	}
	return connectx.InternalWith(err)
}

func streamAgentRunEvent(evt *session.Event, invocationID, sessionID, agentName string) *agentsv1.StreamAgentRunEvent {
	if evt == nil {
		return nil
	}
	out := &agentsv1.StreamAgentRunEvent{
		InvocationId:  invocationID,
		SessionId:     sessionID,
		AgentName:     agentName,
		EventId:       evt.ID,
		Author:        evt.Author,
		Branch:        evt.Branch,
		Partial:       evt.Partial,
		FinalResponse: evt.IsFinalResponse(),
		Timestamp:     timestamppb.New(evt.Timestamp.UTC()),
	}
	// Override invocation_id with the ADK runtime's value when it
	// surfaces one — the runner records its own per-event invocation id.
	if evt.InvocationID != "" {
		out.InvocationId = evt.InvocationID
	}
	if evt.Content != nil {
		if data, err := json.Marshal(evt.Content); err == nil {
			out.ContentJson = string(data)
		}
	}
	return out
}

func streamAgentEventHasOnlyTextParts(evt *session.Event) bool {
	if evt == nil || evt.Content == nil || len(evt.Content.Parts) == 0 {
		return false
	}
	for _, part := range evt.Content.Parts {
		if part == nil {
			continue
		}
		if part.Text == "" && !part.Thought {
			return false
		}
		if part.FunctionCall != nil || part.FunctionResponse != nil ||
			part.CodeExecutionResult != nil || part.ExecutableCode != nil ||
			part.InlineData != nil || part.FileData != nil {
			return false
		}
	}
	return true
}

func streamAgentTextParts(evt *session.Event) []string {
	if evt == nil || !evt.Partial || evt.Content == nil {
		return nil
	}
	out := make([]string, 0, len(evt.Content.Parts))
	for _, part := range evt.Content.Parts {
		if part == nil || part.Text == "" || part.Thought {
			continue
		}
		out = append(out, part.Text)
	}
	return out
}
