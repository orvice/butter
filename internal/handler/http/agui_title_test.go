package http

import (
	"context"
	"encoding/json"
	"errors"
	"maps"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	adksession "google.golang.org/adk/v2/session"

	"go.orx.me/apps/butter/internal/application"
	"go.orx.me/apps/butter/internal/repo/auth"
	"go.orx.me/apps/butter/internal/testsupport/openaifake"
	wsctx "go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// fakeTitler records the threads the handler asks to title. When hold is set,
// each call waits for it to close, so a test can look at the request while
// the title is still being made.
type fakeTitler struct {
	hold chan struct{}

	mu    sync.Mutex
	calls []fakeTitleCall
}

type fakeTitleCall struct {
	session string // app/user/session
	// ctxErr and hasDeadline describe the call's context once it proceeds.
	ctxErr      error
	hasDeadline bool
}

func (f *fakeTitler) TitleSession(ctx context.Context, appName, userID, sessionID string) (*agentsv1.SessionInfo, bool, error) {
	if f.hold != nil {
		<-f.hold
	}
	_, hasDeadline := ctx.Deadline()
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, fakeTitleCall{
		session:     appName + "/" + userID + "/" + sessionID,
		ctxErr:      ctx.Err(),
		hasDeadline: hasDeadline,
	})
	return &agentsv1.SessionInfo{AppName: appName, UserId: userID, SessionId: sessionID, Title: "Trip plan"}, true, nil
}

func (f *fakeTitler) recorded() []fakeTitleCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]fakeTitleCall(nil), f.calls...)
}

// titledADKSession is a session its store reports a title for, as Mongo's
// sessions do.
type titledADKSession struct {
	fakeADKSession
	title string
}

func (s *titledADKSession) Title() string { return s.title }

// setupTitlingAGUIRouter serves the AG-UI endpoint to signed-in user u1 with
// a titler wired. It returns the handler too, so a test can wait for the
// background titles it started.
func setupTitlingAGUIRouter(runnerSvc AGUIRunnerService, sessions adksession.Service, titler AGUISessionTitler, guard *fakeSessionGuard) (*gin.Engine, *AGUIHandler) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		ctx := wsctx.WithID(c.Request.Context(), "test-workspace")
		ctx = auth.WithAuthenticated(ctx, &agentsv1.User{Id: "u1"}, nil)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	h := NewAGUIHandler(aguiEnabledRepo())
	h.SetRunnerService(runnerSvc)
	h.SetSessionService(sessions)
	h.SetSessionGuard(guard)
	h.SetSessionTitler(titler)
	h.Register(r)
	return r, h
}

func newAGUIRequest(t *testing.T, agentID string, body map[string]any) *http.Request {
	t.Helper()
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/agui/"+agentID, strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/json")
	return req
}

// A successful run on a thread without a title has the server title it, once.
// The title is made after the response is complete and the lease released, on
// a context that outlives the request but has a deadline.
func TestAGUIRun_TitlesAnUntitledThreadInTheBackground(t *testing.T) {
	cases := map[string]func() adksession.Service{
		// The run creates the thread's session, untitled.
		"new thread": func() adksession.Service { return newStoreLikeSessions() },
		// An earlier run left the thread without a title.
		"existing untitled thread": func() adksession.Service {
			return &fakeSessionService{sess: &fakeADKSession{}}
		},
	}
	for name, sessions := range cases {
		t.Run(name, func(t *testing.T) {
			titler := &fakeTitler{hold: make(chan struct{})}
			guard := &fakeSessionGuard{}
			router, h := setupTitlingAGUIRouter(&mockRunner{runResult: "Here is a plan."}, sessions(), titler, guard)

			reqCtx, endRequest := context.WithCancel(context.Background())
			req := newAGUIRequest(t, "writer", minimalAGUIBody("t-1", "Plan a trip to Kyoto")).WithContext(reqCtx)
			w := httptest.NewRecorder()
			served := make(chan struct{})
			go func() {
				defer close(served)
				router.ServeHTTP(w, req)
			}()
			// The response completes and the lease is released while the title
			// is still being made: titling holds up neither.
			select {
			case <-served:
			case <-time.After(5 * time.Second):
				close(titler.hold)
				t.Fatal("the response waited for the title")
			}
			if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"type":"RUN_FINISHED"`) {
				t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
			}
			if guard.releases != 1 {
				t.Fatalf("lease releases = %d while the title was pending, want 1", guard.releases)
			}

			endRequest()
			close(titler.hold)
			h.titles.Wait()

			calls := titler.recorded()
			if len(calls) != 1 {
				t.Fatalf("title calls = %d, want 1", len(calls))
			}
			if calls[0].session != "agui/u1/agui-t-1" {
				t.Errorf("titled %q, want the run's session agui/u1/agui-t-1", calls[0].session)
			}
			if calls[0].ctxErr != nil {
				t.Errorf("the title's context ended with the request: %v", calls[0].ctxErr)
			}
			if !calls[0].hasDeadline {
				t.Error("the title's context has no deadline")
			}
		})
	}
}

// memoryTitleStore keeps session titles with the compare-and-set the Mongo
// store implements: a title is written only while none is stored.
type memoryTitleStore struct {
	mu     sync.Mutex
	titles map[string]string // app/user/session -> title
}

func (s *memoryTitleStore) SetSessionTitle(_ context.Context, appName, userID, sessionID, title string) (*agentsv1.SessionInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.titles[appName+"/"+userID+"/"+sessionID] = title
	return &agentsv1.SessionInfo{AppName: appName, UserId: userID, SessionId: sessionID, Title: title}, nil
}

func (s *memoryTitleStore) SetSessionTitleIfEmpty(_ context.Context, appName, userID, sessionID, title string) (*agentsv1.SessionInfo, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := appName + "/" + userID + "/" + sessionID
	info := &agentsv1.SessionInfo{AppName: appName, UserId: userID, SessionId: sessionID, Title: s.titles[key]}
	if info.Title != "" {
		return info, false, nil
	}
	s.titles[key] = title
	info.Title = title
	return info, true, nil
}

func (s *memoryTitleStore) stored() map[string]string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return maps.Clone(s.titles)
}

// With the real titler reading the session store the run writes to, a
// successful run on a new thread stores a title derived from its first
// message, and a failed run stores none.
func TestAGUIRun_StoresATitleOnlyAfterASuccessfulRun(t *testing.T) {
	cases := map[string]struct {
		fail bool
		want map[string]string
	}{
		"successful run": {want: map[string]string{"agui/u1/agui-t-1": "Plan a trip to Kyoto"}},
		"failed run":     {fail: true, want: map[string]string{}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			sessions := newStoreLikeSessions()
			store := &memoryTitleStore{titles: map[string]string{}}
			titler := application.NewSessionServiceServer()
			titler.SetSessionService(sessions)
			titler.SetTitleStore(store)

			h := &a2uiHarness{t: t, backend: openaifake.New(t), sessions: sessions, guard: &fakeSessionGuard{}, titler: titler}
			h.router = h.build([]agentsv1.Agent{cardAgent()}, []string{"card-model"})
			if tc.fail {
				h.backend.Script("card-model", func(w http.ResponseWriter, _ *http.Request) {
					http.Error(w, `{"error": {"message": "model exploded"}}`, http.StatusBadRequest)
				})
			}

			w := h.post("carder", minimalAGUIBody("t-1", "Plan a trip to Kyoto"))
			if w.Code != http.StatusOK {
				t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
			}
			if failed := strings.Contains(w.Body.String(), `"type":"RUN_ERROR"`); failed != tc.fail {
				t.Fatalf("run failed = %v, want %v:\n%s", failed, tc.fail, w.Body.String())
			}
			h.handler.titles.Wait()
			if got := store.stored(); !maps.Equal(got, tc.want) {
				t.Fatalf("stored titles = %v, want %v", got, tc.want)
			}
		})
	}
}

// Only a successful run on a thread without a title asks for one: a titled
// thread keeps its title, and a failed run titles nothing.
func TestAGUIRun_NoTitleForATitledThreadOrAFailedRun(t *testing.T) {
	cases := []struct {
		name     string
		runner   AGUIRunnerService
		sessions adksession.Service
		wantType string
	}{
		{
			name:     "titled thread",
			runner:   &mockRunner{runResult: "Here is a plan."},
			sessions: &fakeSessionService{sess: &titledADKSession{title: "Trip plan"}},
			wantType: "RUN_FINISHED",
		},
		{
			name:     "failed run",
			runner:   &mockRunner{runErr: errors.New("model exploded")},
			sessions: newStoreLikeSessions(),
			wantType: "RUN_ERROR",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			titler := &fakeTitler{}
			router, h := setupTitlingAGUIRouter(tc.runner, tc.sessions, titler, &fakeSessionGuard{})

			w := httptest.NewRecorder()
			router.ServeHTTP(w, newAGUIRequest(t, "writer", minimalAGUIBody("t-1", "Plan a trip to Kyoto")))
			if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"type":"`+tc.wantType+`"`) {
				t.Fatalf("status = %d, want a stream ending in %s: %s", w.Code, tc.wantType, w.Body.String())
			}
			h.titles.Wait()
			if calls := titler.recorded(); len(calls) != 0 {
				t.Fatalf("title calls = %+v, want none", calls)
			}
		})
	}
}
