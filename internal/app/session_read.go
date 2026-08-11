package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.orx.me/apps/butter/internal/application"
	mongosession "go.orx.me/apps/butter/internal/runtime/session/mongo"
)

// mongoReadStore adapts *mongosession.Service to application.SessionReadStore.
type mongoReadStore struct {
	svc *mongosession.Service
}

func newMongoReadStore(svc *mongosession.Service) *mongoReadStore {
	return &mongoReadStore{svc: svc}
}

func (m *mongoReadStore) MarkRead(ctx context.Context, appName, userID, sessionID string, readAt time.Time) (application.SessionReadResult, error) {
	result, err := m.svc.MarkRead(ctx, appName, userID, sessionID, readAt)
	if err != nil {
		if errors.Is(err, mongosession.ErrSessionNotFound) {
			return application.SessionReadResult{}, fmt.Errorf("%w: %s/%s/%s", application.ErrSessionNotFound, appName, userID, sessionID)
		}
		return application.SessionReadResult{}, err
	}
	return application.SessionReadResult{
		SessionID:      result.SessionID,
		AppName:        result.AppName,
		UserID:         result.UserID,
		Title:          result.Title,
		LastUpdateTime: result.LastUpdateTime,
		LastReadAt:     result.LastReadAt,
		WorkspaceID:    result.WorkspaceID,
	}, nil
}
