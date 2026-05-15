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

	"github.com/twitchtv/twirp"
	"google.golang.org/genai"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/runtime/runner"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/adk/session"
)

// SessionServiceServer implements the generated SessionService Twirp interface.
type SessionServiceServer struct {
	mu           sync.RWMutex
	sessionSvc   session.Service
	runnerSvc    *runner.Service
	langfuseHost string
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

// SetRunnerService sets the runner service after bootstrap.
func (s *SessionServiceServer) SetRunnerService(svc *runner.Service) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.runnerSvc = svc
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

func (s *SessionServiceServer) getRunnerSvc() *runner.Service {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.runnerSvc
}

func (s *SessionServiceServer) CreateSession(ctx context.Context, req *agentsv1.CreateSessionRequest) (*agentsv1.CreateSessionResponse, error) {
	sessionSvc := s.getSessionSvc()
	if sessionSvc == nil {
		return nil, twirp.NewError(twirp.FailedPrecondition, "session service not available")
	}

	var state map[string]any
	if req.GetState() != nil {
		state = req.GetState().AsMap()
	}

	resp, err := sessionSvc.Create(ctx, &session.CreateRequest{
		AppName:   req.GetAppName(),
		UserID:    req.GetUserId(),
		SessionID: req.GetSessionId(),
		State:     state,
	})
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}

	return &agentsv1.CreateSessionResponse{Session: sessionToInfo(resp.Session)}, nil
}

func (s *SessionServiceServer) GetSession(ctx context.Context, req *agentsv1.GetSessionRequest) (*agentsv1.GetSessionResponse, error) {
	sessionSvc := s.getSessionSvc()
	if sessionSvc == nil {
		return nil, twirp.NewError(twirp.FailedPrecondition, "session service not available")
	}

	resp, err := sessionSvc.Get(ctx, &session.GetRequest{
		AppName:         req.GetAppName(),
		UserID:          req.GetUserId(),
		SessionID:       req.GetSessionId(),
		NumRecentEvents: int(req.GetNumRecentEvents()),
	})
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil, twirp.NewError(twirp.Canceled, err.Error())
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, twirp.NewError(twirp.DeadlineExceeded, err.Error())
		}
		if strings.Contains(strings.ToLower(err.Error()), "session not found") {
			return nil, twirp.NotFoundError(err.Error())
		}
		return nil, twirp.InternalErrorWith(err)
	}

	detail := &agentsv1.SessionDetail{
		Session: sessionToInfo(resp.Session),
	}

	host := s.getLangfuseHost()
	for evt := range resp.Session.Events().All() {
		detail.Events = append(detail.Events, eventToProtoWithTrace(evt, host))
	}

	return &agentsv1.GetSessionResponse{SessionDetail: detail}, nil
}

func (s *SessionServiceServer) ListSessions(ctx context.Context, req *agentsv1.ListSessionsRequest) (*agentsv1.ListSessionsResponse, error) {
	sessionSvc := s.getSessionSvc()
	if sessionSvc == nil {
		return nil, twirp.NewError(twirp.FailedPrecondition, "session service not available")
	}

	resp, err := sessionSvc.List(ctx, &session.ListRequest{
		AppName: req.GetAppName(),
		UserID:  req.GetUserId(),
	})
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}

	// Apply date-range filter at the service layer since ADK session.ListRequest
	// only supports AppName+UserID.
	startTs := req.GetStartTime()
	endTs := req.GetEndTime()
	infos := make([]*agentsv1.SessionInfo, 0, len(resp.Sessions))
	for _, sess := range resp.Sessions {
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
	page, next := paginateSessions(infos, req.GetPageSize(), req.GetPageToken())

	return &agentsv1.ListSessionsResponse{
		Sessions:      page,
		NextPageToken: next,
		Total:         total,
	}, nil
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

func (s *SessionServiceServer) DeleteSession(ctx context.Context, req *agentsv1.DeleteSessionRequest) (*agentsv1.DeleteSessionResponse, error) {
	sessionSvc := s.getSessionSvc()
	if sessionSvc == nil {
		return nil, twirp.NewError(twirp.FailedPrecondition, "session service not available")
	}

	err := sessionSvc.Delete(ctx, &session.DeleteRequest{
		AppName:   req.GetAppName(),
		UserID:    req.GetUserId(),
		SessionID: req.GetSessionId(),
	})
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}
	return &agentsv1.DeleteSessionResponse{}, nil
}

func (s *SessionServiceServer) ReplySession(ctx context.Context, req *agentsv1.ReplySessionRequest) (*agentsv1.ReplySessionResponse, error) {
	runnerSvc := s.getRunnerSvc()
	if runnerSvc == nil {
		return nil, twirp.NewError(twirp.FailedPrecondition, "runner service not available")
	}

	textPart := &genai.Part{Text: req.GetMessage()}
	ctxInfo := &agentsv1.ContextInfo{
		ChannelName: req.GetAppName(),
		SessionId:   req.GetSessionId(),
		UserId:      req.GetUserId(),
		Source:      agentsv1.ContextSource_CONTEXT_SOURCE_API,
	}

	response, err := runnerSvc.Run(ctx, req.GetAgentName(), []*genai.Part{textPart}, req.GetModelOverride(), ctxInfo, nil, nil)
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}

	return &agentsv1.ReplySessionResponse{Response: response}, nil
}

func sessionToInfo(sess session.Session) *agentsv1.SessionInfo {
	info := &agentsv1.SessionInfo{
		SessionId:      sess.ID(),
		AppName:        sess.AppName(),
		UserId:         sess.UserID(),
		LastUpdateTime: timestamppb.New(sess.LastUpdateTime()),
	}

	// Convert state to protobuf Struct.
	stateMap := make(map[string]any)
	for k, v := range sess.State().All() {
		stateMap[k] = v
	}
	if len(stateMap) > 0 {
		if st, err := structpb.NewStruct(stateMap); err == nil {
			info.State = st
		}
	}

	return info
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
