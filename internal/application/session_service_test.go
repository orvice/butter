package application

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/adk/v2/session"

	"go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// stubSessionService implements session.Service; only Get and Delete are scripted.
type stubSessionService struct {
	session.Service
	deleteErr error
	deleted   []*session.DeleteRequest
}

// Get answers that every addressed session exists, in the caller's
// workspace, so DeleteSession's access check finds it.
func (s *stubSessionService) Get(ctx context.Context, req *session.GetRequest) (*session.GetResponse, error) {
	wsID, _ := workspace.FromContext(ctx)
	return &session.GetResponse{Session: &fakeWSSession{id: req.SessionID, appName: req.AppName, userID: req.UserID, wsID: wsID}}, nil
}

func (s *stubSessionService) Delete(_ context.Context, req *session.DeleteRequest) error {
	if s.deleteErr != nil {
		return s.deleteErr
	}
	s.deleted = append(s.deleted, req)
	return nil
}

type deletedCoords struct{ appName, userID, sessionID string }

// TestDeleteSessionNotifiesDeleteListeners: issue #132 — a successful delete
// must fan out the deleted session's coordinates to registered listeners so
// the cron scheduler can cancel WAITING_INPUT executions stranded on the
// deleted session.
func TestDeleteSessionNotifiesDeleteListeners(t *testing.T) {
	svc := NewSessionServiceServer()
	svc.SetSessionService(&stubSessionService{})

	var got []deletedCoords
	svc.AddSessionDeleteListener(func(appName, userID, sessionID string) {
		got = append(got, deletedCoords{appName, userID, sessionID})
	})

	_, err := svc.DeleteSession(testContextWithUser("ws-test", "member-1"), connect.NewRequest(&agentsv1.DeleteSessionRequest{
		AppName:   "cron:approve-deploy",
		UserId:    "cron:approve-deploy",
		SessionId: "cron:approve-deploy:exec-1",
	}))
	if err != nil {
		t.Fatalf("DeleteSession: %v", err)
	}

	want := deletedCoords{"cron:approve-deploy", "cron:approve-deploy", "cron:approve-deploy:exec-1"}
	if len(got) != 1 || got[0] != want {
		t.Fatalf("listener calls = %+v, want exactly one with %+v", got, want)
	}
}

// TestDeleteSessionFailureDoesNotNotifyListeners: if the underlying delete
// fails, the session still exists — listeners must not see a deletion.
func TestDeleteSessionFailureDoesNotNotifyListeners(t *testing.T) {
	svc := NewSessionServiceServer()
	svc.SetSessionService(&stubSessionService{deleteErr: errors.New("mongo unavailable")})

	calls := 0
	svc.AddSessionDeleteListener(func(_, _, _ string) { calls++ })

	_, err := svc.DeleteSession(testContextWithUser("ws-test", "member-1"), connect.NewRequest(&agentsv1.DeleteSessionRequest{
		AppName:   "cron:approve-deploy",
		UserId:    "cron:approve-deploy",
		SessionId: "cron:approve-deploy:exec-1",
	}))
	if err == nil {
		t.Fatal("expected DeleteSession to fail")
	}
	if calls != 0 {
		t.Fatalf("listener calls = %d, want 0 on failed delete", calls)
	}
}
