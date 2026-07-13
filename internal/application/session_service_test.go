package application

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/adk/v2/session"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// stubSessionService implements session.Service; only Delete is scripted.
type stubSessionService struct {
	session.Service
	deleteErr error
	deleted   []*session.DeleteRequest
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

	_, err := svc.DeleteSession(context.Background(), connect.NewRequest(&agentsv1.DeleteSessionRequest{
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

	_, err := svc.DeleteSession(context.Background(), connect.NewRequest(&agentsv1.DeleteSessionRequest{
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
