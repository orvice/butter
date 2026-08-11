package application

import (
	"context"
	"errors"

	"butterfly.orx.me/core/log"
	"connectrpc.com/connect"

	"go.orx.me/apps/butter/internal/repo/auth"
	"go.orx.me/apps/butter/internal/repo/invocation"
	"go.orx.me/apps/butter/internal/runtime/asyncrun"
	"go.orx.me/apps/butter/internal/runtime/streamorch"
	"go.orx.me/apps/butter/internal/transport/connectx"
	"go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// sourceDashboardAsync marks invocations created by dashboard async chat.
const sourceDashboardAsync = "dashboard-async"

// WatchAgentInvocation streams read-only observer frames for one async
// invocation. The first frame is always the authoritative current state;
// live run events and text deltas follow; the stream ends after exactly one
// terminal state frame. Observation never owns execution: any number of
// watchers may attach, and disconnecting cancels nothing. A watcher that
// falls behind the live run is disconnected with RESOURCE_EXHAUSTED and is
// expected to reload persisted session state before re-watching.
func (s *AgentServiceServer) WatchAgentInvocation(
	ctx context.Context,
	req *connect.Request[agentsv1.WatchAgentInvocationRequest],
	stream *connect.ServerStream[agentsv1.WatchAgentInvocationResponse],
) error {
	if s.invRepo == nil {
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("invocation repository not available"))
	}
	if s.asyncCoord == nil {
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("async coordinator not available"))
	}
	invID := req.Msg.GetInvocationId()
	if invID == "" {
		return connectx.RequiredArgument("invocation_id")
	}

	wsID, hasWorkspace := workspace.FromContext(ctx)
	if !hasWorkspace && !auth.IsAdmin(ctx) {
		return connect.NewError(connect.CodeFailedPrecondition,
			errors.New("workspace required (set X-Workspace-ID header)"))
	}

	// Subscribe before reading the snapshot: the terminal frame is published
	// only after the terminal record is saved, so a snapshot that reads as
	// non-terminal is always followed by the terminal frame on this channel.
	frames, cancelWatch := s.asyncCoord.Watch(invID)
	defer cancelWatch()

	inv, err := getInvocation(ctx, s.invRepo, wsID, invID)
	if err != nil {
		if errors.Is(err, invocation.ErrNotFound) {
			return connectx.NotFound("invocation not found")
		}
		return connectx.InternalWith(err)
	}
	if err := authorizeInvocationAccess(ctx, wsID, inv); err != nil {
		return err
	}

	logger := log.FromContext(ctx)
	logger.Info("watch agent invocation attached",
		"invocation_id", invID,
		"workspace_id", wsID,
		"session_id", inv.GetSessionId(),
		"status", inv.GetStatus().String(),
	)

	if err := stream.Send(watchStateResponse(inv)); err != nil {
		return err
	}
	if asyncrun.IsTerminal(inv.GetStatus()) {
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			// Observer went away; the run continues unaffected.
			return nil
		case f, ok := <-frames:
			if !ok {
				// Closed without a terminal frame: this observer lagged and
				// was dropped by the hub (or the run ended abnormally).
				return connect.NewError(connect.CodeResourceExhausted,
					errors.New("observer lagged behind live stream; reload session state and re-watch"))
			}
			resp, terminal := watchFrameResponse(f)
			if resp == nil {
				continue
			}
			if err := stream.Send(resp); err != nil {
				return err
			}
			if terminal {
				return nil
			}
		}
	}
}

// watchStateResponse wraps an authoritative invocation snapshot.
func watchStateResponse(inv *agentsv1.Invocation) *agentsv1.WatchAgentInvocationResponse {
	return &agentsv1.WatchAgentInvocationResponse{
		Event: &agentsv1.WatchAgentInvocationResponse_State{
			State: &agentsv1.WatchAgentInvocationState{Invocation: inv},
		},
	}
}

// watchFrameResponse translates one hub frame into a wire response. The
// second return value reports whether this frame is the terminal state.
func watchFrameResponse(f asyncrun.Frame) (*agentsv1.WatchAgentInvocationResponse, bool) {
	switch f.Kind {
	case asyncrun.FrameState:
		if f.Invocation == nil {
			return nil, false
		}
		return watchStateResponse(f.Invocation), f.Terminal()
	case asyncrun.FrameTextDelta:
		return &agentsv1.WatchAgentInvocationResponse{
			Event: &agentsv1.WatchAgentInvocationResponse_TextDelta{
				TextDelta: watchTextDelta(f.Identity, f.Text),
			},
		}, false
	case asyncrun.FrameRunEvent:
		evt := streamAgentRunEvent(f.Identity, f.Event)
		if evt == nil {
			return nil, false
		}
		return &agentsv1.WatchAgentInvocationResponse{
			Event: &agentsv1.WatchAgentInvocationResponse_RunEvent{RunEvent: evt},
		}, false
	}
	return nil, false
}

func watchTextDelta(id streamorch.RunIdentity, text string) *agentsv1.StreamAgentTextDelta {
	return &agentsv1.StreamAgentTextDelta{
		InvocationId: id.InvocationID,
		SessionId:    id.SessionID,
		AgentName:    id.AgentName,
		AgentId:      id.AgentID,
		Text:         text,
	}
}
