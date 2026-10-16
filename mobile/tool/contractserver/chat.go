package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/gin-gonic/gin"
	adkrunner "google.golang.org/adk/v2/runner"
	adksession "google.golang.org/adk/v2/session"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/application"
	httpHandler "go.orx.me/apps/butter/internal/handler/http"
	configmemory "go.orx.me/apps/butter/internal/repo/config/memory"
	invocationmemory "go.orx.me/apps/butter/internal/repo/invocation/memory"
	skillrepo "go.orx.me/apps/butter/internal/repo/skill"
	"go.orx.me/apps/butter/internal/runtime/runlog"
	"go.orx.me/apps/butter/internal/runtime/runner"
	"go.orx.me/apps/butter/internal/runtime/runstate"
	"go.orx.me/apps/butter/internal/runtime/sessionguard"
	"go.orx.me/apps/butter/internal/runtime/sessionshare"
	"go.orx.me/apps/butter/internal/testsupport/openaifake"
	wsctx "go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"go.orx.me/apps/butter/pkg/proto/agents/v1/agentsv1connect"
)

// The chat fixture: the production AG-UI handler, runner, Detached Run
// machinery and SessionService, with only the model (a local echo) and the
// persistence (in memory) replaced. It lives in its own Workspace so the
// agents the other contracts count are untouched.
const (
	chatWorkspace = "chat"
	chatAgentID   = "talker"
	chatModel     = "fixture-model"
	// legacyThread is an AG-UI thread whose session has no UI binding, as one
	// created before bindings existed: a Detached Run on it must be refused.
	legacyThread = "legacy"
	chatUserID   = "mobile-test"
)

// chatMarkdownReply is what the model answers to a message that asks for it.
const chatMarkdownReply = "# Fixture\n\n| a | b |\n|---|---|\n| 1 | 2 |\n\n```go\nfmt.Println(1)\n```\n"

type sessionMeta struct {
	workspaceID string
	title       string
}

// chatSessions gives the in-memory ADK session service the behavior of the
// Mongo store that chat depends on: a session knows its Workspace and title,
// sessions list by Workspace, and a session ID another user holds is refused.
type chatSessions struct {
	adksession.Service
	mu      sync.Mutex
	metas   map[string]*sessionMeta // app|user|id
	holders map[string]map[string]bool
}

func newChatSessions() *chatSessions {
	return &chatSessions{
		Service: adksession.InMemoryService(),
		metas:   map[string]*sessionMeta{},
		holders: map[string]map[string]bool{},
	}
}

func metaKey(app, user, id string) string { return app + "|" + user + "|" + id }

type sessionView struct {
	adksession.Session
	workspaceID string
	title       string
}

func (v *sessionView) WorkspaceID() string { return v.workspaceID }
func (v *sessionView) Title() string       { return v.title }

func (s *chatSessions) view(sess adksession.Session) adksession.Session {
	s.mu.Lock()
	defer s.mu.Unlock()
	m := s.metas[metaKey(sess.AppName(), sess.UserID(), sess.ID())]
	if m == nil {
		return sess
	}
	return &sessionView{Session: sess, workspaceID: m.workspaceID, title: m.title}
}

func (s *chatSessions) Create(ctx context.Context, req *adksession.CreateRequest) (*adksession.CreateResponse, error) {
	holderKey := req.AppName + "/" + req.SessionID
	wsID, _ := wsctx.FromContext(ctx)
	s.mu.Lock()
	for user := range s.holders[holderKey] {
		if user != req.UserID && !sessionshare.Allowed(ctx) {
			s.mu.Unlock()
			return nil, fmt.Errorf("%w: %s", sessionshare.ErrIDTaken, holderKey)
		}
	}
	if s.holders[holderKey] == nil {
		s.holders[holderKey] = map[string]bool{}
	}
	s.holders[holderKey][req.UserID] = true
	s.metas[metaKey(req.AppName, req.UserID, req.SessionID)] = &sessionMeta{workspaceID: wsID}
	s.mu.Unlock()
	resp, err := s.Service.Create(ctx, req)
	if err != nil {
		return nil, err
	}
	return &adksession.CreateResponse{Session: s.view(resp.Session)}, nil
}

func (s *chatSessions) Get(ctx context.Context, req *adksession.GetRequest) (*adksession.GetResponse, error) {
	resp, err := s.Service.Get(ctx, req)
	if err != nil {
		return nil, err
	}
	return &adksession.GetResponse{Session: s.view(resp.Session)}, nil
}

func (s *chatSessions) List(ctx context.Context, req *adksession.ListRequest) (*adksession.ListResponse, error) {
	resp, err := s.Service.List(ctx, req)
	if err != nil {
		return nil, err
	}
	out := make([]adksession.Session, 0, len(resp.Sessions))
	for _, sess := range resp.Sessions {
		out = append(out, s.view(sess))
	}
	return &adksession.ListResponse{Sessions: out}, nil
}

// AppendEvent unwraps the view: the in-memory service accepts only its own
// session type.
func (s *chatSessions) AppendEvent(ctx context.Context, sess adksession.Session, evt *adksession.Event) error {
	if v, ok := sess.(*sessionView); ok {
		sess = v.Session
	}
	return s.Service.AppendEvent(ctx, sess, evt)
}

func (s *chatSessions) ListByWorkspace(ctx context.Context, workspaceID, userID, appName string) ([]adksession.Session, error) {
	s.mu.Lock()
	type ref struct{ app, user, id string }
	var refs []ref
	for key, m := range s.metas {
		if m.workspaceID != workspaceID {
			continue
		}
		parts := strings.SplitN(key, "|", 3)
		if (userID != "" && parts[1] != userID) || (appName != "" && parts[0] != appName) {
			continue
		}
		refs = append(refs, ref{parts[0], parts[1], parts[2]})
	}
	s.mu.Unlock()
	out := make([]adksession.Session, 0, len(refs))
	for _, r := range refs {
		resp, err := s.Service.Get(ctx, &adksession.GetRequest{AppName: r.app, UserID: r.user, SessionID: r.id})
		if err != nil {
			continue
		}
		out = append(out, s.view(resp.Session))
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].LastUpdateTime().After(out[j].LastUpdateTime()) })
	return out, nil
}

func (s *chatSessions) GetWorkspaceID(_ context.Context, appName, userID, sessionID string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m := s.metas[metaKey(appName, userID, sessionID)]
	if m == nil {
		return "", fmt.Errorf("%w: %s", application.ErrSessionNotFound, sessionID)
	}
	return m.workspaceID, nil
}

func (s *chatSessions) info(ctx context.Context, appName, userID, sessionID string) (*agentsv1.SessionInfo, error) {
	resp, err := s.Service.Get(ctx, &adksession.GetRequest{AppName: appName, UserID: userID, SessionID: sessionID})
	if err != nil {
		return nil, fmt.Errorf("%w: %s", application.ErrSessionNotFound, sessionID)
	}
	s.mu.Lock()
	m := s.metas[metaKey(appName, userID, sessionID)]
	s.mu.Unlock()
	info := &agentsv1.SessionInfo{
		SessionId: sessionID, AppName: appName, UserId: userID,
		LastUpdateTime: timestamppb.New(resp.Session.LastUpdateTime()),
	}
	if m != nil {
		info.WorkspaceId, info.Title = m.workspaceID, m.title
	}
	return info, nil
}

func (s *chatSessions) SetSessionTitle(ctx context.Context, appName, userID, sessionID, title string) (*agentsv1.SessionInfo, error) {
	s.mu.Lock()
	m := s.metas[metaKey(appName, userID, sessionID)]
	if m == nil {
		s.mu.Unlock()
		return nil, fmt.Errorf("%w: %s", application.ErrSessionNotFound, sessionID)
	}
	m.title = title
	s.mu.Unlock()
	return s.info(ctx, appName, userID, sessionID)
}

func (s *chatSessions) SetSessionTitleIfEmpty(ctx context.Context, appName, userID, sessionID, title string) (*agentsv1.SessionInfo, bool, error) {
	s.mu.Lock()
	m := s.metas[metaKey(appName, userID, sessionID)]
	if m == nil {
		s.mu.Unlock()
		return nil, false, fmt.Errorf("%w: %s", application.ErrSessionNotFound, sessionID)
	}
	wrote := m.title == ""
	if wrote {
		m.title = title
	}
	s.mu.Unlock()
	info, err := s.info(ctx, appName, userID, sessionID)
	return info, wrote, err
}

// fakeModel is a local OpenAI-compatible endpoint: it echoes the last user
// text, or answers Markdown when asked. It is the only stand-in on the run
// path.
func fakeModel(ctx context.Context) (string, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		var req openaifake.ChatCompletionRequest
		if json.Unmarshal(body, &req) != nil || json.Unmarshal(body, &req.Decoded) != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		text := lastUserText(req.Messages)
		if toolCommand(w, req, text) {
			return
		}
		reply := fmt.Sprintf("fixture(%s)", text)
		if strings.Contains(text, "markdown") {
			reply = chatMarkdownReply
		}
		openaifake.WriteReply(w, req, reply)
	})
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}
	server := &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go func() { _ = server.Serve(listener) }()
	go func() {
		<-ctx.Done()
		_ = server.Close()
	}()
	return "http://" + listener.Addr().String(), nil
}

func lastUserText(messages []openaifake.ChatCompletionMessage) string {
	last := ""
	for _, m := range messages {
		if m.Role != "user" {
			continue
		}
		var text string
		if json.Unmarshal(m.Content, &text) == nil {
			last = text
			continue
		}
		var parts []struct {
			Text string `json:"text"`
		}
		if json.Unmarshal(m.Content, &parts) == nil {
			var joined strings.Builder
			for _, p := range parts {
				joined.WriteString(p.Text)
			}
			last = joined.String()
		}
	}
	return last
}

// registerChat wires the chat fixture's routes onto r.
func registerChat(ctx context.Context, r *gin.Engine, agentRepo *configmemory.Store, opts []connect.HandlerOption, bx *boxFixture, backends *fileBackends, skills skillrepo.Repository) (*runner.Service, []agentsv1.ModelProvider, error) {
	modelURL, err := fakeModel(ctx)
	if err != nil {
		return nil, nil, err
	}
	sessions := newChatSessions()
	agents := []agentsv1.Agent{{
		Name: "chat-talker", AgentId: chatAgentID, WorkspaceId: chatWorkspace,
		Config: &agentsv1.AgentConfig{Model: chatModel},
	}}
	if _, err := agentRepo.CreateAgent(ctx, chatWorkspace, &agentsv1.Agent{
		Name: agents[0].Name, AgentId: agents[0].AgentId, Config: agents[0].Config,
	}); err != nil {
		return nil, nil, err
	}
	providers := chatProviders(modelURL)
	svc, err := runner.NewServiceWithMCPHTTPClientFactory(ctx, agents, providers,
		nil, nil, nil, sessions, nil, nil, backends.files, 0, skills, adkrunner.PluginConfig{}, nil, bx.builders)
	if err != nil {
		return nil, nil, err
	}
	invocations := invocationmemory.New().WithOwner("contract-fixture")
	svc.SetInvocationRecorder(invocations)

	agui := httpHandler.NewAGUIHandler(agentRepo)
	agui.SetRunnerService(svc)
	agui.SetSessionService(sessions)
	agui.SetSessionGuard(sessionguard.NewMemory())
	agui.SetInvocationRepo(invocations)
	agui.SetRunStateStore(runstate.NewMemory(time.Minute))
	agui.SetRunLogStore(runlog.NewMemory())
	agui.Register(r)

	// A session with no UI binding, as one from before bindings existed.
	legacyCtx := wsctx.WithID(ctx, chatWorkspace)
	if _, err := sessions.Create(legacyCtx, &adksession.CreateRequest{
		AppName: "agui", UserID: chatUserID, SessionID: "agui-" + legacyThread,
	}); err != nil {
		return nil, nil, err
	}

	sessionSvc := application.NewSessionServiceServer()
	sessionSvc.SetSessionService(sessions)
	sessionSvc.SetTitleStore(sessions)
	sessionSvc.SetWorkspaceSessionStore(sessions)
	sessionSvc.SetAGUIThreads(agui)
	_, handler := agentsv1connect.NewSessionServiceHandler(sessionSvc, opts...)
	for _, procedure := range []string{
		agentsv1connect.SessionServiceListSessionsProcedure,
		agentsv1connect.SessionServiceUpdateSessionTitleProcedure,
	} {
		r.POST("/api"+procedure, gin.WrapH(http.StripPrefix("/api", handler)))
	}
	return svc, providers, nil
}

// chatProviders is the one provider the fixture's runner serves, pointed at
// the fake model.
func chatProviders(modelURL string) []agentsv1.ModelProvider {
	return []agentsv1.ModelProvider{{
		Name: "fake", Type: "openai", BaseUrl: modelURL,
		Models: []*agentsv1.ModelConfig{{Name: chatModel}},
	}}
}
