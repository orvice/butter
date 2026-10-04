package application

import (
	"context"

	"butterfly.orx.me/core/log"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// AsyncTurnComplete is called by the async coordinator after a successful
// invocation, on a context that carries no caller. It titles an untitled
// session through TitleSession, which needs none: the invocation was
// authorized when it was submitted. A session that already has a title keeps
// it (the SetSessionTitleIfEmpty CAS keeps concurrent calls safe). Failures
// are logged, never propagated.
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

	info, generated, err := s.TitleSession(ctx, appName, userID, sessionID)
	if err != nil {
		logger.Warn("async title generation failed",
			"session_id", sessionID,
			"err", err,
		)
		return
	}
	if generated {
		logger.Info("async title generated",
			"session_id", sessionID,
			"title", info.GetTitle(),
		)
	}
}
