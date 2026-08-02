package app

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/application"
	mongosession "go.orx.me/apps/butter/internal/runtime/session/mongo"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// mongoTitleStore adapts *mongosession.Service to application.SessionTitleStore.
type mongoTitleStore struct {
	svc *mongosession.Service
}

func newMongoTitleStore(svc *mongosession.Service) *mongoTitleStore {
	return &mongoTitleStore{svc: svc}
}

func (m *mongoTitleStore) SetSessionTitle(ctx context.Context, appName, userID, sessionID, title string) (*agentsv1.SessionInfo, error) {
	result, err := m.svc.SetSessionTitle(ctx, appName, userID, sessionID, title)
	if err != nil {
		if errors.Is(err, mongosession.ErrSessionNotFound) {
			return nil, fmt.Errorf("%w: %s/%s/%s", application.ErrSessionNotFound, appName, userID, sessionID)
		}
		return nil, err
	}
	return &agentsv1.SessionInfo{
		SessionId:      result.SessionID,
		AppName:        result.AppName,
		UserId:         result.UserID,
		Title:          result.Title,
		LastUpdateTime: timestamppb.New(result.LastUpdateTime),
	}, nil
}
