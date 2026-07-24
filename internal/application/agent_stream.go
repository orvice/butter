package application

import (
	"context"
	"encoding/json"
	"errors"

	"butterfly.orx.me/core/log"
	"connectrpc.com/connect"
	"google.golang.org/adk/v2/session"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/repo/auth"
	"go.orx.me/apps/butter/internal/runtime/streamorch"
	"go.orx.me/apps/butter/internal/transport/connectx"
	wsctx "go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// StreamAgent runs an agent and streams progress events to the caller.
// The dashboard chat UI uses this Connect server-stream RPC (the legacy
// POST /api/chat/stream SSE handler was removed).
//
// The stream's first message is always a StreamAgentStarted carrying the
// invocation_id; clients use that id with CancelAgentInvocation if the
// user hits "stop". Mid-stream messages are either StreamAgentTextDelta
// (partial assistant text) or StreamAgentRunEvent (a complete ADK
// session.Event that the dashboard parses to render tool calls etc.).
// The terminal message is a StreamAgentFinal carrying the final response.
// On error the stream is closed with a connect.Error — no payload event.
//
// This handler is a thin adapter over streamorch: it builds the ContextInfo,
// wires a Sink that translates frames into StreamAgentResponse proto
// messages, and maps the orchestrator's raw error to a connect.Error.
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
	if len(req.Msg.GetParts()) == 0 && req.Msg.GetMessage() == "" {
		return connectx.RequiredArgument("message")
	}
	parts, err := resolveUserParts(req.Msg.GetParts(), req.Msg.GetMessage())
	if err != nil {
		return err
	}

	workspaceID, hasWorkspace := wsctx.FromContext(ctx)
	ctxInfo, err := streamorch.NewContextInfo(streamorch.ContextInfoInput{
		AppName:       req.Msg.GetAppName(),
		UserID:        req.Msg.GetUserId(),
		SessionID:     req.Msg.GetSessionId(),
		SessionPrefix: "chat-",
		WorkspaceID:   workspaceID,
		HasWorkspace:  hasWorkspace,
		IsAdmin:       auth.IsAdmin(ctx),
		Source:        agentsv1.ContextSource_CONTEXT_SOURCE_API,
		ChatType:      agentsv1.ChatType_CHAT_TYPE_PRIVATE,
	})
	if err != nil {
		return connect.NewError(connect.CodeFailedPrecondition, err)
	}

	logger := log.FromContext(ctx)
	logger.Info("streaming agent started",
		"workspace_id", ctxInfo.GetWorkspaceId(),
		"agent", req.Msg.GetAgentName(),
		"session_id", ctxInfo.GetSessionId(),
		"invocation_id", ctxInfo.GetUuid(),
		"message_len", len(req.Msg.GetMessage()),
		"parts", len(req.Msg.GetParts()),
	)

	sink := &streamAgentSink{ctx: ctx, stream: stream}
	sink.start()
	runErr := streamorch.Run(ctx, s.runnerSvc, req.Msg.GetAgentName(), parts, req.Msg.GetModelOverride(), ctxInfo, sink)
	sendErr := sink.finish()

	if runErr != nil {
		if !sink.started {
			// The Started frame itself failed to send (client already gone
			// before the first byte) — return the raw error unmapped, as
			// there is no connection left to report a status code over.
			return runErr
		}
		logger.Error("streaming agent failed",
			"workspace_id", ctxInfo.GetWorkspaceId(),
			"agent", req.Msg.GetAgentName(),
			"session_id", ctxInfo.GetSessionId(),
			"invocation_id", ctxInfo.GetUuid(),
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
		"workspace_id", ctxInfo.GetWorkspaceId(),
		"agent", req.Msg.GetAgentName(),
		"session_id", ctxInfo.GetSessionId(),
		"invocation_id", ctxInfo.GetUuid(),
	)
	return nil
}

// streamAgentSink implements streamorch.Sink by translating frames into
// StreamAgentResponse proto messages and sending them over the connect
// stream. connect.ServerStream is not safe for concurrent Send; sends after
// the initial Started frame are serialized through a single goroutine
// consuming an event channel, since streamorch.Run's callback runs
// synchronously inside the ADK runner's event loop.
type streamAgentSink struct {
	ctx    context.Context
	stream *connect.ServerStream[agentsv1.StreamAgentResponse]

	events      chan *agentsv1.StreamAgentResponse
	doneSending chan error
	started     bool
}

func (s *streamAgentSink) start() {
	s.events = make(chan *agentsv1.StreamAgentResponse, 32)
	s.doneSending = make(chan error, 1)
	go func() {
		var sendErr error
		for resp := range s.events {
			if sendErr != nil {
				continue
			}
			if err := s.stream.Send(resp); err != nil {
				sendErr = err
			}
		}
		s.doneSending <- sendErr
	}()
}

func (s *streamAgentSink) finish() error {
	close(s.events)
	return <-s.doneSending
}

func (s *streamAgentSink) queue(resp *agentsv1.StreamAgentResponse) {
	select {
	case s.events <- resp:
	case <-s.ctx.Done():
	}
}

func (s *streamAgentSink) Started(invocationID, sessionID, agentName string) error {
	// The Started frame must be sent (and observed by the client) before any
	// other frame, so it bypasses the queue and is sent directly — safe
	// because nothing else is sending concurrently yet.
	err := s.stream.Send(&agentsv1.StreamAgentResponse{
		Event: &agentsv1.StreamAgentResponse_Started{
			Started: &agentsv1.StreamAgentStarted{
				InvocationId: invocationID,
				SessionId:    sessionID,
				AgentName:    agentName,
			},
		},
	})
	s.started = err == nil
	return err
}

func (s *streamAgentSink) TextDelta(invocationID, sessionID, agentName, text string) error {
	s.queue(&agentsv1.StreamAgentResponse{
		Event: &agentsv1.StreamAgentResponse_TextDelta{
			TextDelta: &agentsv1.StreamAgentTextDelta{
				InvocationId: invocationID,
				SessionId:    sessionID,
				AgentName:    agentName,
				Text:         text,
			},
		},
	})
	return nil
}

func (s *streamAgentSink) RunEvent(evt *session.Event, invocationID, sessionID, agentName string) error {
	s.queue(&agentsv1.StreamAgentResponse{
		Event: &agentsv1.StreamAgentResponse_RunEvent{
			RunEvent: streamAgentRunEvent(evt, invocationID, sessionID, agentName),
		},
	})
	return nil
}

func (s *streamAgentSink) Final(invocationID, sessionID, agentName, response string) error {
	s.queue(&agentsv1.StreamAgentResponse{
		Event: &agentsv1.StreamAgentResponse_Final{
			Final: &agentsv1.StreamAgentFinal{
				InvocationId: invocationID,
				SessionId:    sessionID,
				AgentName:    agentName,
				Response:     response,
			},
		},
	})
	return nil
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

// streamAgentError maps a runner error to the appropriate Connect status code.
// Existing *connect.Error values are returned as-is. context.Canceled becomes
// CodeCanceled (e.g. user hit stop); context.DeadlineExceeded becomes
// CodeDeadlineExceeded. Everything else falls back to CodeInternal.
func streamAgentError(err error) *connect.Error {
	var connectErr *connect.Error
	if errors.As(err, &connectErr) {
		return connectErr
	}
	if errors.Is(err, context.Canceled) {
		return connect.NewError(connect.CodeCanceled, err)
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return connect.NewError(connect.CodeDeadlineExceeded, err)
	}
	return connect.NewError(connect.CodeInternal, err)
}
