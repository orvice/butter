package http

import (
	"context"
	"fmt"
	"sync"

	adksession "google.golang.org/adk/v2/session"

	"go.orx.me/apps/butter/internal/runtime/sessionshare"
	wsctx "go.orx.me/apps/butter/internal/workspace"
)

// storeLikeSessions gives the in-memory session service the two behaviors of
// the Mongo store an AG-UI run depends on: a session reports the workspace it
// was created in, and a session ID another user holds is refused unless the
// creating context allows sharing.
type storeLikeSessions struct {
	adksession.Service
	mu      sync.Mutex
	holders map[string]map[string]string // app/session ID -> user -> workspace
}

func newStoreLikeSessions() *storeLikeSessions {
	return &storeLikeSessions{Service: adksession.InMemoryService(), holders: map[string]map[string]string{}}
}

func (s *storeLikeSessions) Create(ctx context.Context, req *adksession.CreateRequest) (*adksession.CreateResponse, error) {
	key := req.AppName + "/" + req.SessionID
	wsID, _ := wsctx.FromContext(ctx)
	s.mu.Lock()
	for user := range s.holders[key] {
		if user != req.UserID && !sessionshare.Allowed(ctx) {
			s.mu.Unlock()
			return nil, fmt.Errorf("%w: %s", sessionshare.ErrIDTaken, key)
		}
	}
	if s.holders[key] == nil {
		s.holders[key] = map[string]string{}
	}
	s.holders[key][req.UserID] = wsID
	s.mu.Unlock()

	resp, err := s.Service.Create(ctx, req)
	if err != nil {
		return nil, err
	}
	return &adksession.CreateResponse{Session: &workspacedView{Session: resp.Session, workspaceID: wsID}}, nil
}

func (s *storeLikeSessions) Get(ctx context.Context, req *adksession.GetRequest) (*adksession.GetResponse, error) {
	resp, err := s.Service.Get(ctx, req)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	wsID := s.holders[req.AppName+"/"+req.SessionID][req.UserID]
	s.mu.Unlock()
	return &adksession.GetResponse{Session: &workspacedView{Session: resp.Session, workspaceID: wsID}}, nil
}

// AppendEvent unwraps the view: the in-memory service only accepts its own
// session type.
func (s *storeLikeSessions) AppendEvent(ctx context.Context, sess adksession.Session, evt *adksession.Event) error {
	if v, ok := sess.(*workspacedView); ok {
		sess = v.Session
	}
	return s.Service.AppendEvent(ctx, sess, evt)
}

// workspacedView is a session that knows its workspace, as Mongo's does.
type workspacedView struct {
	adksession.Session
	workspaceID string
}

func (v *workspacedView) WorkspaceID() string { return v.workspaceID }
