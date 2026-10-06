package application

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"butterfly.orx.me/core/log"
	"connectrpc.com/connect"
	"google.golang.org/adk/v2/model"
	"google.golang.org/genai"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/repo/auth"
	"go.orx.me/apps/butter/internal/repo/inputpart"
	"go.orx.me/apps/butter/internal/repo/invocation"
	workspacerepo "go.orx.me/apps/butter/internal/repo/workspace"
	"go.orx.me/apps/butter/internal/runtime/interrupt"
	"go.orx.me/apps/butter/internal/runtime/runner"
	"go.orx.me/apps/butter/internal/runtime/sessionshare"
	"go.orx.me/apps/butter/internal/transport/connectx"
	"go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/adk/v2/session"
)

// sessionReplyRunner is the subset of *runner.Service that ReplySession
// depends on; tests substitute a fake implementation.
type sessionReplyRunner interface {
	Run(ctx context.Context, agentName string, parts []*genai.Part, modelOverride string, ctxInfo *agentsv1.ContextInfo, onEvent runner.EventCallback, onCompaction runner.CompactionCallback) (string, error)
	ResolveAgentRef(workspaceID, agentID string) (string, bool)
}

// ErrSessionNotFound must be wrapped into the error a SessionTitleStore
// returns when the addressed session does not exist, so the RPC layer can map
// it to CodeNotFound with errors.Is instead of matching message text.
var ErrSessionNotFound = errors.New("session not found")

// SessionTitleStore is the narrow contract for Butter-owned session title
// metadata. The Mongo session service implements it via an adapter wired in
// internal/app; the ADK session.Service interface is not changed.
type SessionTitleStore interface {
	SetSessionTitle(ctx context.Context, appName, userID, sessionID, title string) (*agentsv1.SessionInfo, error)
	// SetSessionTitleIfEmpty atomically sets the title only when neither a
	// first-class title nor a legacy state["title"] is currently persisted.
	// Returns (info, true) when a new title was written, or (info, false)
	// when an existing title won.
	SetSessionTitleIfEmpty(ctx context.Context, appName, userID, sessionID, title string) (*agentsv1.SessionInfo, bool, error)
}

// WorkspaceSessionStore provides workspace-scoped session queries that go
// beyond the generic ADK session.Service interface.
type WorkspaceSessionStore interface {
	// ListByWorkspace returns the workspace's sessions newest first, for
	// one user and one app; an empty userID or appName matches any.
	ListByWorkspace(ctx context.Context, workspaceID, userID, appName string) ([]session.Session, error)
	GetWorkspaceID(ctx context.Context, appName, userID, sessionID string) (string, error)
}

// SessionReadStore persists and retrieves server-backed read state for a
// session. The Mongo session service implements it.
type SessionReadStore interface {
	MarkRead(ctx context.Context, appName, userID, sessionID string, readAt time.Time) (SessionReadResult, error)
}

// SessionReadResult carries the post-update read snapshot.
type SessionReadResult struct {
	SessionID      string
	AppName        string
	UserID         string
	Title          string
	LastUpdateTime time.Time
	LastReadAt     time.Time
	WorkspaceID    string
}

// aguiAppName is the app of AG-UI threads' sessions (the AG-UI handler's).
const aguiAppName = "agui"

// AGUIThreads holds an AG-UI thread while DeleteSession deletes it (ADR-0016
// decision 7). *httpHandler.AGUIHandler implements it.
type AGUIThreads interface {
	// HoldThreadForDelete stops the thread's Detached Run, on any Pod, then
	// takes the thread's lease, waiting with a bound, and drops the thread's
	// run state. The delete runs on holdCtx, and release ends the hold. held
	// is false when the lease did not come free in time: nothing was
	// changed.
	HoldThreadForDelete(ctx context.Context, userID, sessionID string) (holdCtx context.Context, release func(), held bool, err error)
}

// SessionServiceServer implements the generated SessionService ConnectRPC handler.
type SessionServiceServer struct {
	mu              sync.RWMutex
	sessionSvc      session.Service
	runnerSvc       sessionReplyRunner
	titleStore      SessionTitleStore
	wsStore         WorkspaceSessionStore
	wsRepo          workspacerepo.Repository
	readStore       SessionReadStore
	invRepo         invocation.Repository
	inputPartRepo   inputpart.Repository
	aguiThreads     AGUIThreads
	langfuseHost    string
	deleteListeners []SessionDeleteListener

	titleResolver       TitleModelResolver
	titleProviderLister WorkspaceModelProviderLister
	chatTitleModel      string
	// titleResolveModel is titleGenerator.resolveModel for this server: a
	// seam for tests, nil selecting the production resolver.
	titleResolveModel func(ctx context.Context, modelRef string, providers []*agentsv1.ModelProvider) (model.LLM, error)
}

// SessionDeleteListener observes successful session deletions with the
// deleted session's coordinates. The cron scheduler registers one to cancel
// WAITING_INPUT executions stranded on a deleted session (issue #132).
type SessionDeleteListener func(appName, userID, sessionID string)

// AddSessionDeleteListener registers a listener called after every
// successful DeleteSession.
func (s *SessionServiceServer) AddSessionDeleteListener(fn SessionDeleteListener) {
	if fn == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deleteListeners = append(s.deleteListeners, fn)
}

// notifySessionDeleted fans the deleted session's coordinates out to every
// registered listener. Dispatch is synchronous, matching the runner's turn
// listeners: the caller's DeleteSession response then confirms reconciliation
// has happened, at the cost of the RPC waiting on listener work (deliveries
// carry their own timeouts).
func (s *SessionServiceServer) notifySessionDeleted(appName, userID, sessionID string) {
	s.mu.RLock()
	listeners := make([]SessionDeleteListener, len(s.deleteListeners))
	copy(listeners, s.deleteListeners)
	s.mu.RUnlock()
	for _, fn := range listeners {
		fn(appName, userID, sessionID)
	}
}

func NewSessionServiceServer() *SessionServiceServer {
	return &SessionServiceServer{}
}

// SetSessionService sets the ADK session service after bootstrap.
func (s *SessionServiceServer) SetSessionService(svc session.Service) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessionSvc = svc
}

// SetRunnerService sets the runner service after bootstrap. A nil
// *runner.Service is ignored so the nil check in ReplySession keeps
// working against the interface-typed field.
func (s *SessionServiceServer) SetRunnerService(svc *runner.Service) {
	if svc == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.runnerSvc = svc
}

// SetTitleStore wires the Butter-owned session title persistence.
func (s *SessionServiceServer) SetTitleStore(store SessionTitleStore) {
	if store == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.titleStore = store
}

// SetWorkspaceSessionStore wires workspace-scoped session queries.
func (s *SessionServiceServer) SetWorkspaceSessionStore(store WorkspaceSessionStore) {
	if store == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.wsStore = store
}

// SetReadStore wires the server-backed read-state persistence.
func (s *SessionServiceServer) SetReadStore(store SessionReadStore) {
	if store == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.readStore = store
}

func (s *SessionServiceServer) getReadStore() SessionReadStore {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.readStore
}

// SetInvocationRepo wires the invocation repository for enriching session
// summaries with the active/latest invocation status.
func (s *SessionServiceServer) SetInvocationRepo(repo invocation.Repository) {
	if repo == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.invRepo = repo
}

func (s *SessionServiceServer) getInvRepo() invocation.Repository {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.invRepo
}

// SetInputPartRepo wires the input part repository for cleaning up stored
// multimodal input parts during session deletion.
func (s *SessionServiceServer) SetInputPartRepo(repo inputpart.Repository) {
	if repo == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.inputPartRepo = repo
}

func (s *SessionServiceServer) getInputPartRepo() inputpart.Repository {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.inputPartRepo
}

// SetAGUIThreads wires how DeleteSession deletes an AG-UI thread: stopping
// its Detached Run and holding the thread's lease while it deletes.
func (s *SessionServiceServer) SetAGUIThreads(threads AGUIThreads) {
	if threads == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.aguiThreads = threads
}

func (s *SessionServiceServer) getAGUIThreads() AGUIThreads {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.aguiThreads
}

func (s *SessionServiceServer) getWSStore() WorkspaceSessionStore {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.wsStore
}

// workspacedSession is implemented by session backends that carry workspace_id.
type workspacedSession interface {
	WorkspaceID() string
}

// enforceWorkspaceAccessByID looks up the session's workspace from the store
// and validates access. Used by mutations that load the session after the check.
func (s *SessionServiceServer) enforceWorkspaceAccessByID(ctx context.Context, appName, userID, sessionID string) error {
	if auth.IsAdmin(ctx) {
		return nil
	}
	wsID, ok := workspace.FromContext(ctx)
	if !ok || wsID == "" {
		return nil
	}
	wsStore := s.getWSStore()
	if wsStore == nil {
		return nil
	}
	sessWS, err := wsStore.GetWorkspaceID(ctx, appName, userID, sessionID)
	if err != nil {
		if strings.Contains(err.Error(), "session not found") {
			return connectx.NotFound("session not found")
		}
		return connectx.InternalWith(err)
	}
	if sessWS == "" || sessWS != wsID {
		return connectx.NotFound("session not found in this workspace")
	}
	return nil
}

func (s *SessionServiceServer) getTitleStore() SessionTitleStore {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.titleStore
}

// SetTitleModelResolver wires the model/agent resolver used for LLM title
// generation. Typically backed by runner.Service.
func (s *SessionServiceServer) SetTitleModelResolver(r TitleModelResolver) {
	if r == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.titleResolver = r
}

// SetTitleProviderLister wires the workspace-scoped model provider lister
// used for LLM title generation.
func (s *SessionServiceServer) SetTitleProviderLister(l WorkspaceModelProviderLister) {
	if l == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.titleProviderLister = l
}

// SetChatTitleModel configures the dedicated model alias for LLM title
// generation. Empty means fall back to the agent's configured model.
func (s *SessionServiceServer) SetChatTitleModel(model string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.chatTitleModel = model
}

func (s *SessionServiceServer) titleGen() titleGenerator {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return titleGenerator{
		resolver:       s.titleResolver,
		providerLister: s.titleProviderLister,
		chatTitleModel: s.chatTitleModel,
		resolveModel:   s.titleResolveModel,
	}
}

// SetLangfuseHost wires the Langfuse base URL used to render trace_url on
// SessionEvent. Empty disables trace_url emission (trace_id is still set).
func (s *SessionServiceServer) SetLangfuseHost(host string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.langfuseHost = strings.TrimRight(strings.TrimSpace(host), "/")
}

func (s *SessionServiceServer) getLangfuseHost() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.langfuseHost
}

func (s *SessionServiceServer) getSessionSvc() session.Service {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessionSvc
}

func (s *SessionServiceServer) getRunnerSvc() sessionReplyRunner {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.runnerSvc
}

func (s *SessionServiceServer) CreateSession(ctx context.Context, req *connect.Request[agentsv1.CreateSessionRequest]) (*connect.Response[agentsv1.CreateSessionResponse], error) {
	sessionSvc := s.getSessionSvc()
	if sessionSvc == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("session service not available"))
	}

	if err := s.authorizeNew(ctx, req.Msg.GetUserId()); err != nil {
		return nil, err
	}

	var state map[string]any
	if req.Msg.GetState() != nil {
		state = req.Msg.GetState().AsMap()
	}

	logger := log.FromContext(ctx)
	resp, err := sessionSvc.Create(ctx, &session.CreateRequest{
		AppName:   req.Msg.GetAppName(),
		UserID:    req.Msg.GetUserId(),
		SessionID: req.Msg.GetSessionId(),
		State:     state,
	})
	if err != nil {
		if errors.Is(err, sessionshare.ErrIDTaken) {
			return nil, errSessionIDTaken()
		}
		logger.Error("create session failed",
			"app_name", req.Msg.GetAppName(),
			"user_id", req.Msg.GetUserId(),
			"session_id", req.Msg.GetSessionId(),
			"err", err,
		)
		return nil, connectx.InternalWith(err)
	}

	logger.Info("session created",
		"app_name", req.Msg.GetAppName(),
		"user_id", req.Msg.GetUserId(),
		"session_id", resp.Session.ID(),
	)
	return connect.NewResponse(&agentsv1.CreateSessionResponse{Session: sessionToInfo(resp.Session)}), nil
}

func (s *SessionServiceServer) GetSession(ctx context.Context, req *connect.Request[agentsv1.GetSessionRequest]) (*connect.Response[agentsv1.GetSessionResponse], error) {
	sessionSvc := s.getSessionSvc()
	if sessionSvc == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("session service not available"))
	}

	resp, err := sessionSvc.Get(ctx, &session.GetRequest{
		AppName:         req.Msg.GetAppName(),
		UserID:          req.Msg.GetUserId(),
		SessionID:       req.Msg.GetSessionId(),
		NumRecentEvents: int(req.Msg.GetNumRecentEvents()),
	})
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil, connect.NewError(connect.CodeCanceled, errors.New(err.Error()))
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, connect.NewError(connect.CodeDeadlineExceeded, errors.New(err.Error()))
		}
		if strings.Contains(strings.ToLower(err.Error()), "session not found") {
			if _, authErr := sessionPrincipalFrom(ctx); authErr != nil {
				return nil, authErr
			}
			return nil, sessionNotFound()
		}
		return nil, connectx.InternalWith(err)
	}

	// Session access policy (session_access.go): a session the caller may
	// not see answers not-found, like a missing one.
	if err := s.authorizeExisting(ctx, req.Msg.GetAppName(), req.Msg.GetUserId(), sessionWorkspaceID(resp.Session)); err != nil {
		return nil, err
	}

	detail := &agentsv1.SessionDetail{
		Session: sessionToInfo(resp.Session),
	}

	host := s.getLangfuseHost()
	var firstTS, lastTS time.Time
	for evt := range resp.Session.Events().All() {
		detail.Events = append(detail.Events, eventToProtoWithTrace(evt, host))
		if firstTS.IsZero() || evt.Timestamp.Before(firstTS) {
			firstTS = evt.Timestamp
		}
		if evt.Timestamp.After(lastTS) {
			lastTS = evt.Timestamp
		}
	}
	if n := len(detail.GetEvents()); n > 0 {
		detail.Session.TurnCount = int32(n)
	}
	if !firstTS.IsZero() && !lastTS.IsZero() && lastTS.After(firstTS) {
		detail.Duration = durationpb.New(lastTS.Sub(firstTS))
	}

	// Enrich with invocation status. Events are already loaded, so derive
	// pending interrupts directly instead of loading them again.
	wsID, _ := workspace.FromContext(ctx)
	s.enrichInvocationStatus(ctx, wsID, detail.Session)
	detail.Session.HasPendingInterrupt = len(interrupt.Pending(resp.Session)) > 0

	return connect.NewResponse(&agentsv1.GetSessionResponse{SessionDetail: detail}), nil
}

func (s *SessionServiceServer) ListSessions(ctx context.Context, req *connect.Request[agentsv1.ListSessionsRequest]) (*connect.Response[agentsv1.ListSessionsResponse], error) {
	sessionSvc := s.getSessionSvc()
	if sessionSvc == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("session service not available"))
	}

	// Workspace-scoped path: list only sessions owned by both the active
	// workspace and the authenticated user. Legacy sessions without
	// workspace_id are excluded.
	if req.Msg.GetWorkspaceScoped() {
		return s.listSessionsWorkspaceScoped(ctx, req)
	}

	// Session access policy (session_access.go).
	scope, err := s.listScope(ctx, req.Msg.GetAppName(), req.Msg.GetUserId())
	if err != nil {
		return nil, err
	}

	resp, err := sessionSvc.List(ctx, &session.ListRequest{
		AppName: req.Msg.GetAppName(),
		UserID:  scope.userID,
	})
	if err != nil {
		return nil, connectx.InternalWith(err)
	}

	// Apply date-range filter at the service layer since ADK session.ListRequest
	// only supports AppName+UserID.
	startTs := req.Msg.GetStartTime()
	endTs := req.Msg.GetEndTime()
	infos := make([]*agentsv1.SessionInfo, 0, len(resp.Sessions))
	for _, sess := range resp.Sessions {
		if scope.workspaceID != "" && sessionWorkspaceID(sess) != scope.workspaceID {
			continue
		}
		last := sess.LastUpdateTime()
		if startTs != nil && last.Before(startTs.AsTime()) {
			continue
		}
		if endTs != nil && last.After(endTs.AsTime()) {
			continue
		}
		infos = append(infos, sessionToInfo(sess))
	}

	// Newest first.
	sort.SliceStable(infos, func(i, j int) bool {
		return infos[i].GetLastUpdateTime().AsTime().After(infos[j].GetLastUpdateTime().AsTime())
	})

	total := int32(len(infos))
	page, next := paginateSessions(infos, req.Msg.GetPageSize(), req.Msg.GetPageToken())

	return connect.NewResponse(&agentsv1.ListSessionsResponse{
		Sessions:      page,
		NextPageToken: next,
		Total:         total,
	}), nil
}

// listSessionsWorkspaceScoped returns sessions scoped to the active workspace
// and authenticated user, and to app_name when the request sets one. Only
// workspace-owned sessions are included; legacy sessions without
// workspace_id are excluded.
func (s *SessionServiceServer) listSessionsWorkspaceScoped(ctx context.Context, req *connect.Request[agentsv1.ListSessionsRequest]) (*connect.Response[agentsv1.ListSessionsResponse], error) {
	wsID, _ := workspace.FromContext(ctx)
	if wsID == "" {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("workspace_scoped listing requires X-Workspace-ID header"))
	}

	wsStore := s.getWSStore()
	if wsStore == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("workspace session store not available"))
	}

	// Determine the user to filter on: a person sees only their own
	// sessions; admins and the workspace's API tokens may specify a user_id
	// or leave it empty for all users.
	userID := req.Msg.GetUserId()
	_, isAPIToken := auth.APITokenFromContext(ctx)
	if !auth.IsAdmin(ctx) && !isAPIToken {
		user, ok := auth.UserFromContext(ctx)
		if !ok {
			return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("authentication required"))
		}
		userID = user.GetId()
	}

	sessions, err := wsStore.ListByWorkspace(ctx, wsID, userID, req.Msg.GetAppName())
	if err != nil {
		return nil, connectx.InternalWith(err)
	}

	// Apply date-range filter.
	startTs := req.Msg.GetStartTime()
	endTs := req.Msg.GetEndTime()
	infos := make([]*agentsv1.SessionInfo, 0, len(sessions))
	for _, sess := range sessions {
		last := sess.LastUpdateTime()
		if startTs != nil && last.Before(startTs.AsTime()) {
			continue
		}
		if endTs != nil && last.After(endTs.AsTime()) {
			continue
		}
		infos = append(infos, sessionToInfo(sess))
	}

	total := int32(len(infos))
	page, next := paginateSessions(infos, req.Msg.GetPageSize(), req.Msg.GetPageToken())

	// Enrich the current page with invocation + interrupt state.
	s.enrichSessionInfos(ctx, wsID, page)

	return connect.NewResponse(&agentsv1.ListSessionsResponse{
		Sessions:      page,
		NextPageToken: next,
		Total:         total,
	}), nil
}

// enrichInvocationStatus populates the invocation ID and status on one
// SessionInfo. Tries the active (QUEUED/RUNNING) invocation first, then falls
// back to the latest regardless of status.
func (s *SessionServiceServer) enrichInvocationStatus(ctx context.Context, wsID string, info *agentsv1.SessionInfo) {
	invRepo := s.getInvRepo()
	if invRepo == nil {
		return
	}
	logger := log.FromContext(ctx)
	sid := info.GetSessionId()

	active, err := invRepo.FindActiveBySession(ctx, wsID, sid)
	if err == nil && active != nil {
		info.LatestInvocationId = active.GetId()
		info.LatestInvocationStatus = active.GetStatus()
		return
	}
	latest, latestErr := invRepo.FindLatestBySession(ctx, wsID, sid)
	if latestErr == nil && latest != nil {
		info.LatestInvocationId = latest.GetId()
		info.LatestInvocationStatus = latest.GetStatus()
	} else if latestErr != nil && !errors.Is(latestErr, invocation.ErrNotFound) {
		logger.Warn("enriching session invocation failed",
			"session_id", sid, "err", latestErr)
	}
}

// enrichSessionInfos populates invocation status and pending-interrupt state
// on the given session infos. Enrichment is best-effort: failures are logged
// and the fields are left at their zero values.
func (s *SessionServiceServer) enrichSessionInfos(ctx context.Context, wsID string, infos []*agentsv1.SessionInfo) {
	sessionSvc := s.getSessionSvc()

	for _, info := range infos {
		s.enrichInvocationStatus(ctx, wsID, info)

		// Pending-interrupt enrichment (ADR-0002): load session events and
		// call interrupt.Pending. Only for web-chat sessions to avoid loading
		// events for non-dashboard sessions.
		if sessionSvc != nil && info.GetAppName() == "web-chat" {
			sessResp, sessErr := sessionSvc.Get(ctx, &session.GetRequest{
				AppName:   info.GetAppName(),
				UserID:    info.GetUserId(),
				SessionID: info.GetSessionId(),
			})
			if sessErr == nil && sessResp.Session != nil {
				info.HasPendingInterrupt = len(interrupt.Pending(sessResp.Session)) > 0
			}
		}
	}
}

func paginateSessions(items []*agentsv1.SessionInfo, pageSize int32, pageToken string) ([]*agentsv1.SessionInfo, string) {
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := 0
	if pageToken != "" {
		if raw, err := base64.StdEncoding.DecodeString(pageToken); err == nil {
			if n, err := strconv.Atoi(string(raw)); err == nil && n >= 0 {
				offset = n
			}
		}
	}
	if offset >= len(items) {
		return nil, ""
	}
	end := offset + int(pageSize)
	if end > len(items) {
		end = len(items)
	}
	next := ""
	if end < len(items) {
		next = base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(end)))
	}
	return items[offset:end], next
}

func (s *SessionServiceServer) DeleteSession(ctx context.Context, req *connect.Request[agentsv1.DeleteSessionRequest]) (*connect.Response[agentsv1.DeleteSessionResponse], error) {
	sessionSvc := s.getSessionSvc()
	if sessionSvc == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("session service not available"))
	}

	sessionID := req.Msg.GetSessionId()
	appName := req.Msg.GetAppName()
	userID := req.Msg.GetUserId()

	// Session access policy (session_access.go), checked before anything
	// is cancelled or redacted.
	if err := s.authorizeByID(ctx, appName, userID, sessionID); err != nil {
		return nil, err
	}

	logger := log.FromContext(ctx)
	wsID, _ := workspace.FromContext(ctx)
	invRepo := s.getInvRepo()

	if threads := s.getAGUIThreads(); appName == aguiAppName && threads != nil {
		// Step 1, for an AG-UI thread (ADR-0016 decision 7): its Detached
		// Run is stopped on whichever Pod runs it, then the thread's lease is
		// taken and held while the thread is deleted, so no run starts or
		// writes in between on any Pod. A lease that does not come free in
		// time, as under a run without the opt-in, fails the delete before
		// anything is deleted.
		holdCtx, release, held, err := threads.HoldThreadForDelete(ctx, userID, sessionID)
		if err != nil {
			logger.Error("holding the AG-UI thread for its delete failed", "session_id", sessionID, "err", err)
			return nil, connect.NewError(connect.CodeUnavailable, errors.New("the thread could not be held for its delete; retry later"))
		}
		if !held {
			return nil, connect.NewError(connect.CodeUnavailable, errors.New("a run still holds the thread; retry the delete once it ends"))
		}
		defer release()
		ctx = holdCtx
	}

	// Step 2: Delete input parts and redact invocation content for all
	// invocations associated with this session. Input parts are left only by
	// the retired dashboard async chat (#410).
	if invRepo != nil && wsID != "" {
		invocations, listErr := invRepo.ListBySession(ctx, wsID, sessionID)
		if listErr != nil {
			logger.Warn("listing invocations for session deletion failed",
				"session_id", sessionID, "err", listErr)
		}
		ipRepo := s.getInputPartRepo()
		for _, inv := range invocations {
			if ipRepo != nil {
				if delErr := ipRepo.Delete(ctx, inv.GetId()); delErr != nil {
					logger.Warn("deleting input parts failed",
						"invocation_id", inv.GetId(), "err", delErr)
				}
			}
			if redactErr := invRepo.RedactContent(ctx, wsID, inv.GetId()); redactErr != nil {
				logger.Warn("redacting invocation content failed",
					"invocation_id", inv.GetId(), "err", redactErr)
			}
		}
	}

	// Step 3: Delete the session metadata and events.
	err := sessionSvc.Delete(ctx, &session.DeleteRequest{
		AppName:   appName,
		UserID:    userID,
		SessionID: sessionID,
	})
	if err != nil {
		logger.Error("delete session failed",
			"app_name", appName,
			"user_id", userID,
			"session_id", sessionID,
			"err", err,
		)
		return nil, connectx.InternalWith(err)
	}
	logger.Info("session deleted",
		"app_name", appName,
		"user_id", userID,
		"session_id", sessionID,
	)

	// Step 4: Notify delete listeners (cron, automation).
	s.notifySessionDeleted(appName, userID, sessionID)
	return connect.NewResponse(&agentsv1.DeleteSessionResponse{}), nil
}

func (s *SessionServiceServer) ReplySession(ctx context.Context, req *connect.Request[agentsv1.ReplySessionRequest]) (*connect.Response[agentsv1.ReplySessionResponse], error) {
	runnerSvc := s.getRunnerSvc()
	if runnerSvc == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("runner service not available"))
	}

	parts, err := resolveUserParts(req.Msg.GetParts(), req.Msg.GetMessage())
	if err != nil {
		return nil, err
	}

	// Session access policy (session_access.go): a turn runs with the
	// session's whole history, so replying is as privileged as reading it.
	if err := s.authorizeReply(ctx, req.Msg.GetAppName(), req.Msg.GetUserId(), req.Msg.GetSessionId()); err != nil {
		return nil, err
	}

	// The workspace header scopes agent resolution to the intended tenant.
	// It is also stamped onto ContextInfo so the runner re-checks the resolved
	// agent belongs to this workspace before executing — without it a request
	// carrying no X-Workspace-ID would resolve an agent_id across tenants
	// (agent_ids are only workspace-unique) and run it unscoped.
	wsID, _ := workspace.FromContext(ctx)
	agentName, err := resolveAgentRunnerRef(runnerSvc, wsID, req.Msg.GetAgentId())
	if err != nil {
		return nil, err
	}
	ctxInfo := &agentsv1.ContextInfo{
		ChannelName: req.Msg.GetAppName(),
		SessionId:   req.Msg.GetSessionId(),
		UserId:      req.Msg.GetUserId(),
		Source:      agentsv1.ContextSource_CONTEXT_SOURCE_API,
		WorkspaceId: wsID,
	}

	logger := log.FromContext(ctx)
	logger.Info("replying to session",
		"agent", agentName,
		"agent_id", req.Msg.GetAgentId(),
		"app_name", req.Msg.GetAppName(),
		"user_id", req.Msg.GetUserId(),
		"session_id", req.Msg.GetSessionId(),
		"message_len", len(req.Msg.GetMessage()),
		"parts", len(req.Msg.GetParts()),
	)
	start := time.Now()
	response, err := runnerSvc.Run(ctx, agentName, parts, req.Msg.GetModelOverride(), ctxInfo, nil, nil)
	if err != nil {
		logger.Error("session reply failed",
			"agent", agentName,
			"session_id", req.Msg.GetSessionId(),
			"elapsed_ms", time.Since(start).Milliseconds(),
			"err", err,
		)
		return nil, turnError(err)
	}
	logger.Info("session reply completed",
		"agent", agentName,
		"session_id", req.Msg.GetSessionId(),
		"elapsed_ms", time.Since(start).Milliseconds(),
	)
	return connect.NewResponse(&agentsv1.ReplySessionResponse{Response: response}), nil
}

const maxTitleCodePoints = 100

// normalizeTitle trims whitespace and collapses the value to a single line.
func normalizeTitle(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}

func (s *SessionServiceServer) UpdateSessionTitle(ctx context.Context, req *connect.Request[agentsv1.UpdateSessionTitleRequest]) (*connect.Response[agentsv1.UpdateSessionTitleResponse], error) {
	if req.Msg.GetAppName() == "" {
		return nil, connectx.RequiredArgument("app_name")
	}
	if req.Msg.GetUserId() == "" {
		return nil, connectx.RequiredArgument("user_id")
	}
	if req.Msg.GetSessionId() == "" {
		return nil, connectx.RequiredArgument("session_id")
	}

	title := normalizeTitle(req.Msg.GetTitle())
	if title == "" {
		return nil, connectx.InvalidArgument("title", "must not be blank")
	}
	if utf8.RuneCountInString(title) > maxTitleCodePoints {
		return nil, connectx.InvalidArgument("title", "must be at most 100 Unicode code points")
	}

	// Authorization: non-admin must match the requested user_id.
	if !auth.IsAdmin(ctx) {
		user, ok := auth.UserFromContext(ctx)
		if !ok {
			return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("authentication required"))
		}
		if user.GetId() != req.Msg.GetUserId() {
			return nil, connect.NewError(connect.CodePermissionDenied, errors.New("cannot update another user's session title"))
		}
	}

	titleStore := s.getTitleStore()
	if titleStore == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("session title store not available"))
	}

	info, err := titleStore.SetSessionTitle(ctx, req.Msg.GetAppName(), req.Msg.GetUserId(), req.Msg.GetSessionId(), title)
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			return nil, connectx.NotFound(err.Error())
		}
		return nil, connectx.InternalWith(err)
	}

	return connect.NewResponse(&agentsv1.UpdateSessionTitleResponse{Session: info}), nil
}

// titleFromSession returns the Butter-owned first-class title if the
// session implementation exposes one (mongoSession does).
type titledSession interface {
	Title() string
}

// effectiveTitle resolves the title according to precedence:
// first-class title → legacy state["title"] → empty (caller falls back).
func effectiveTitle(sess session.Session) string {
	if ts, ok := sess.(titledSession); ok {
		if t := ts.Title(); t != "" {
			return t
		}
	}
	if v, err := sess.State().Get("title"); err == nil {
		if s, ok := v.(string); ok {
			if trimmed := strings.TrimSpace(s); trimmed != "" {
				return trimmed
			}
		}
	}
	return ""
}

// readableSession is implemented by session backends that carry last_read_at.
type readableSession interface {
	LastReadAt() *time.Time
}

func sessionToInfo(sess session.Session) *agentsv1.SessionInfo {
	info := &agentsv1.SessionInfo{
		SessionId:      sess.ID(),
		AppName:        sess.AppName(),
		UserId:         sess.UserID(),
		LastUpdateTime: timestamppb.New(sess.LastUpdateTime()),
		Title:          effectiveTitle(sess),
	}

	if ws, ok := sess.(workspacedSession); ok {
		info.WorkspaceId = ws.WorkspaceID()
	}

	if rs, ok := sess.(readableSession); ok {
		if ra := rs.LastReadAt(); ra != nil {
			info.LastReadAt = timestamppb.New(*ra)
			info.Unread = sess.LastUpdateTime().After(*ra)
		} else {
			info.Unread = true
		}
	}

	// Convert state to protobuf Struct.
	stateMap := make(map[string]any)
	for k, v := range sess.State().All() {
		stateMap[k] = v
	}
	if len(stateMap) > 0 {
		info.State = sessionStateToProto(stateMap)
	}

	return info
}

// sessionStateToProto preserves state loaded from stores whose nested
// documents use named map or slice types (for example MongoDB's bson.D).
// structpb accepts only its fixed set of JSON-shaped Go types, while a JSON
// round trip normalizes those store-specific representations.
func sessionStateToProto(stateMap map[string]any) *structpb.Struct {
	if state, err := structpb.NewStruct(stateMap); err == nil {
		return state
	}

	raw, err := json.Marshal(stateMap)
	if err != nil {
		return nil
	}
	var normalized map[string]any
	if err := json.Unmarshal(raw, &normalized); err != nil {
		return nil
	}
	state, err := structpb.NewStruct(normalized)
	if err != nil {
		return nil
	}
	return state
}

func (s *SessionServiceServer) MarkSessionRead(ctx context.Context, req *connect.Request[agentsv1.MarkSessionReadRequest]) (*connect.Response[agentsv1.MarkSessionReadResponse], error) {
	if req.Msg.GetAppName() == "" {
		return nil, connectx.RequiredArgument("app_name")
	}
	if req.Msg.GetUserId() == "" {
		return nil, connectx.RequiredArgument("user_id")
	}
	if req.Msg.GetSessionId() == "" {
		return nil, connectx.RequiredArgument("session_id")
	}

	// Authorization: only the session owner may mark it read.
	if !auth.IsAdmin(ctx) {
		user, ok := auth.UserFromContext(ctx)
		if !ok {
			return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("authentication required"))
		}
		if user.GetId() != req.Msg.GetUserId() {
			return nil, connectx.NotFound("session not found")
		}
	}

	// Workspace guard.
	if err := s.enforceWorkspaceAccessByID(ctx, req.Msg.GetAppName(), req.Msg.GetUserId(), req.Msg.GetSessionId()); err != nil {
		return nil, err
	}

	readStore := s.getReadStore()
	if readStore == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("session read store not available"))
	}

	result, err := readStore.MarkRead(ctx, req.Msg.GetAppName(), req.Msg.GetUserId(), req.Msg.GetSessionId(), time.Now())
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			return nil, connectx.NotFound(err.Error())
		}
		return nil, connectx.InternalWith(err)
	}

	info := &agentsv1.SessionInfo{
		SessionId:      result.SessionID,
		AppName:        result.AppName,
		UserId:         result.UserID,
		Title:          result.Title,
		LastUpdateTime: timestamppb.New(result.LastUpdateTime),
		LastReadAt:     timestamppb.New(result.LastReadAt),
		Unread:         result.LastUpdateTime.After(result.LastReadAt),
		WorkspaceId:    result.WorkspaceID,
	}

	return connect.NewResponse(&agentsv1.MarkSessionReadResponse{Session: info}), nil
}

const maxAutoTitleCodePoints = 30

// truncateCodePoints returns s truncated to at most n Unicode code points.
func truncateCodePoints(s string, n int) string {
	count := 0
	for i := range s {
		if count >= n {
			return s[:i]
		}
		count++
	}
	return s
}

// normalizeAutoTitle collapses whitespace, trims, and limits to
// maxAutoTitleCodePoints Unicode code points.
func normalizeAutoTitle(s string) string {
	s = normalizeTitle(s)
	if s == "" {
		return ""
	}
	return truncateCodePoints(s, maxAutoTitleCodePoints)
}

// firstEventText extracts the first non-blank text from a genai.Content's
// Parts, ignoring function calls and function responses.
func firstEventText(content *genai.Content) string {
	if content == nil {
		return ""
	}
	for _, part := range content.Parts {
		if part == nil {
			continue
		}
		if s := strings.TrimSpace(part.Text); s != "" {
			return s
		}
	}
	return ""
}

// hasOnlyNonTextParts returns true when the content contains at least one
// part, but none of them are usable text (e.g. all images/inline-data).
func hasOnlyNonTextParts(content *genai.Content) bool {
	if content == nil || len(content.Parts) == 0 {
		return false
	}
	for _, part := range content.Parts {
		if part == nil {
			continue
		}
		if strings.TrimSpace(part.Text) != "" {
			return false
		}
	}
	return true
}

// deriveAutoTitle walks session events and returns a deterministic title:
//  1. First user text (truncated)
//  2. First assistant text (truncated)
//  3. "Image chat" when input contained non-text parts but no usable text
//  4. "" when no useful content was found
func deriveAutoTitle(events []*session.Event) string {
	var firstUserText, firstAssistantText string
	hasUserNonTextParts := false

	for _, evt := range events {
		if evt.Content == nil {
			continue
		}
		// Skip tool calls and tool responses: events whose only parts
		// are FunctionCall or FunctionResponse.
		if isToolOnlyEvent(evt) {
			continue
		}

		text := firstEventText(evt.Content)
		if evt.Author == "user" {
			if firstUserText == "" {
				firstUserText = text
				if text == "" && hasOnlyNonTextParts(evt.Content) {
					hasUserNonTextParts = true
				}
			}
		} else if firstAssistantText == "" {
			firstAssistantText = text
		}

		// User text always wins, so once it is found there is nothing
		// left to look for.
		if firstUserText != "" {
			break
		}
	}

	if firstUserText != "" {
		return normalizeAutoTitle(firstUserText)
	}
	if firstAssistantText != "" {
		return normalizeAutoTitle(firstAssistantText)
	}
	if hasUserNonTextParts {
		return "Image chat"
	}
	return ""
}

// isToolOnlyEvent returns true when the event contains at least one
// FunctionCall or FunctionResponse and no other meaningful content.
func isToolOnlyEvent(evt *session.Event) bool {
	if evt.Content == nil || len(evt.Content.Parts) == 0 {
		return false
	}
	hasTool := false
	for _, part := range evt.Content.Parts {
		if part == nil {
			continue
		}
		if part.FunctionCall != nil || part.FunctionResponse != nil {
			hasTool = true
			continue
		}
		if strings.TrimFunc(part.Text, unicode.IsSpace) != "" ||
			part.InlineData != nil ||
			part.FileData != nil {
			return false
		}
	}
	return hasTool
}

func (s *SessionServiceServer) GenerateSessionTitle(ctx context.Context, req *connect.Request[agentsv1.GenerateSessionTitleRequest]) (*connect.Response[agentsv1.GenerateSessionTitleResponse], error) {
	if req.Msg.GetAppName() == "" {
		return nil, connectx.RequiredArgument("app_name")
	}
	if req.Msg.GetUserId() == "" {
		return nil, connectx.RequiredArgument("user_id")
	}
	if req.Msg.GetSessionId() == "" {
		return nil, connectx.RequiredArgument("session_id")
	}

	// Same self-only non-admin authorization policy as UpdateSessionTitle.
	if !auth.IsAdmin(ctx) {
		user, ok := auth.UserFromContext(ctx)
		if !ok {
			return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("authentication required"))
		}
		if user.GetId() != req.Msg.GetUserId() {
			return nil, connect.NewError(connect.CodePermissionDenied, errors.New("cannot generate title for another user's session"))
		}
	}

	info, generated, err := s.TitleSession(ctx, req.Msg.GetAppName(), req.Msg.GetUserId(), req.Msg.GetSessionId())
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&agentsv1.GenerateSessionTitleResponse{
		Session:   info,
		Generated: generated,
	}), nil
}

// TitleSession gives an untitled session a title derived from its first turn
// and stores it with SetSessionTitleIfEmpty, so a title that already exists
// (manual, legacy, or written concurrently) always wins. generated reports
// whether this call stored a new title; info is the session as it stands
// afterwards. Errors are Connect errors.
//
// It authorizes no one. GenerateSessionTitle checks the caller first; the
// AG-UI handler's turn hook calls it on the server's own behalf after a
// successful run, for a turn that was authorized when it started.
func (s *SessionServiceServer) TitleSession(ctx context.Context, appName, userID, sessionID string) (info *agentsv1.SessionInfo, generated bool, err error) {
	sessionSvc := s.getSessionSvc()
	if sessionSvc == nil {
		return nil, false, connect.NewError(connect.CodeFailedPrecondition, errors.New("session service not available"))
	}
	titleStore := s.getTitleStore()
	if titleStore == nil {
		return nil, false, connect.NewError(connect.CodeFailedPrecondition, errors.New("session title store not available"))
	}

	// Load session with all events.
	sessResp, err := sessionSvc.Get(ctx, &session.GetRequest{
		AppName:   appName,
		UserID:    userID,
		SessionID: sessionID,
	})
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "session not found") {
			return nil, false, connectx.NotFound(err.Error())
		}
		return nil, false, connectx.InternalWith(err)
	}

	// If an effective title already exists, return it without generating.
	if effectiveTitle(sessResp.Session) != "" {
		return sessionToInfo(sessResp.Session), false, nil
	}

	// Collect events for title derivation.
	var events []*session.Event
	for evt := range sessResp.Session.Events().All() {
		events = append(events, evt)
	}

	// Try LLM-based title generation first.
	llmTitle, llmOK := s.titleGen().generate(ctx, events, sessionID)
	var title string
	if llmOK {
		title = llmTitle
	} else {
		title = deriveAutoTitle(events)
	}

	if title == "" {
		return sessionToInfo(sessResp.Session), false, nil
	}

	// Atomic CAS: write only if no first-class title exists yet.
	info, generated, err = titleStore.SetSessionTitleIfEmpty(ctx, appName, userID, sessionID, title)
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			return nil, false, connectx.NotFound(err.Error())
		}
		return nil, false, connectx.InternalWith(err)
	}

	if generated {
		method := "deterministic"
		if llmOK {
			method = "llm"
		}
		log.FromContext(ctx).Info("auto-generated session title",
			"app_name", appName,
			"session_id", sessionID,
			"method", method,
		)
	}
	return info, generated, nil
}

func eventToProtoWithTrace(evt *session.Event, langfuseHost string) *agentsv1.SessionEvent {
	pe := &agentsv1.SessionEvent{
		EventId:      evt.ID,
		InvocationId: evt.InvocationID,
		Author:       evt.Author,
		Branch:       evt.Branch,
		Timestamp:    timestamppb.New(evt.Timestamp),
		TraceId:      evt.InvocationID,
	}

	if evt.Content != nil {
		if data, err := json.Marshal(evt.Content); err == nil {
			pe.ContentJson = string(data)
		}
	}

	if langfuseHost != "" && evt.InvocationID != "" {
		pe.TraceUrl = langfuseHost + "/trace/" + evt.InvocationID
	}

	return pe
}
