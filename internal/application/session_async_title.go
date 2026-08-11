package application

import (
	"context"

	"butterfly.orx.me/core/log"
	"connectrpc.com/connect"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// AsyncTurnComplete is called by the async coordinator after a successful
// invocation. It triggers best-effort title generation without blocking
// the invocation. Only the first successful turn in a session generates a
// title (the SetSessionTitleIfEmpty CAS ensures concurrent calls are safe).
func (s *SessionServiceServer) AsyncTurnComplete(ctx context.Context, inv *agentsv1.Invocation) {
	if inv == nil {
		return
	}

	logger := log.FromContext(ctx)
	appName := inv.GetAppName()
	userID := inv.GetUserId()
	sessionID := inv.GetSessionId()

	if appName == "" || userID == "" || sessionID == "" {
		return
	}

	// Fire-and-forget title generation; errors are logged, never propagated.
	resp, err := s.GenerateSessionTitle(ctx, connect.NewRequest(&agentsv1.GenerateSessionTitleRequest{
		AppName:   appName,
		UserId:    userID,
		SessionId: sessionID,
	}))
	if err != nil {
		logger.Debug("async title generation skipped or failed",
			"session_id", sessionID,
			"err", err,
		)
		return
	}
	if resp.Msg.GetGenerated() {
		logger.Info("async title generated",
			"session_id", sessionID,
			"title", resp.Msg.GetSession().GetTitle(),
		)
	}
}
