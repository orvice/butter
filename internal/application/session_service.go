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

	"butterfly.orx.me/core/log"
	"connectrpc.com/connect"
	"google.golang.org/genai"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/runtime/runner"
	"go.orx.me/apps/butter/internal/transport/connectx"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/adk/v2/session"
)

// SessionServiceServer implements the generated SessionService ConnectRPC handler.
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

func (s *SessionServiceServer) CreateSession(ctx context.Context, req *connect.Request[agentsv1.CreateSessionRequest]) (*connect.Response[agentsv1.CreateSessionResponse], error) {
	sessionSvc := s.getSessionSvc()
	if sessionSvc == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("session service not available"))
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
			return nil, connectx.NotFound(err.Error())
		}
		return nil, connectx.InternalWith(err)
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

	return connect.NewResponse(&agentsv1.GetSessionResponse{SessionDetail: detail}), nil
}

func (s *SessionServiceServer) ListSessions(ctx context.Context, req *connect.Request[agentsv1.ListSessionsRequest]) (*connect.Response[agentsv1.ListSessionsResponse], error) {
	sessionSvc := s.getSessionSvc()
	if sessionSvc == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("session service not available"))
	}

	resp, err := sessionSvc.List(ctx, &session.ListRequest{
		AppName: req.Msg.GetAppName(),
		UserID:  req.Msg.GetUserId(),
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

	logger := log.FromContext(ctx)
	err := sessionSvc.Delete(ctx, &session.DeleteRequest{
		AppName:   req.Msg.GetAppName(),
		UserID:    req.Msg.GetUserId(),
		SessionID: req.Msg.GetSessionId(),
	})
	if err != nil {
		logger.Error("delete session failed",
			"app_name", req.Msg.GetAppName(),
			"user_id", req.Msg.GetUserId(),
			"session_id", req.Msg.GetSessionId(),
			"err", err,
		)
		return nil, connectx.InternalWith(err)
	}
	logger.Info("session deleted",
		"app_name", req.Msg.GetAppName(),
		"user_id", req.Msg.GetUserId(),
		"session_id", req.Msg.GetSessionId(),
	)
	return connect.NewResponse(&agentsv1.DeleteSessionResponse{}), nil
}

func (s *SessionServiceServer) ReplySession(ctx context.Context, req *connect.Request[agentsv1.ReplySessionRequest]) (*connect.Response[agentsv1.ReplySessionResponse], error) {
	runnerSvc := s.getRunnerSvc()
	if runnerSvc == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("runner service not available"))
	}

	textPart := &genai.Part{Text: req.Msg.GetMessage()}
	ctxInfo := &agentsv1.ContextInfo{
		ChannelName: req.Msg.GetAppName(),
		SessionId:   req.Msg.GetSessionId(),
		UserId:      req.Msg.GetUserId(),
		Source:      agentsv1.ContextSource_CONTEXT_SOURCE_API,
	}

	logger := log.FromContext(ctx)
	logger.Info("replying to session",
		"agent", req.Msg.GetAgentName(),
		"app_name", req.Msg.GetAppName(),
		"user_id", req.Msg.GetUserId(),
		"session_id", req.Msg.GetSessionId(),
		"message_len", len(req.Msg.GetMessage()),
	)
	start := time.Now()
	response, err := runnerSvc.Run(ctx, req.Msg.GetAgentName(), []*genai.Part{textPart}, req.Msg.GetModelOverride(), ctxInfo, nil, nil)
	if err != nil {
		logger.Error("session reply failed",
			"agent", req.Msg.GetAgentName(),
			"session_id", req.Msg.GetSessionId(),
			"elapsed_ms", time.Since(start).Milliseconds(),
			"err", err,
		)
		return nil, connectx.InternalWith(err)
	}
	logger.Info("session reply completed",
		"agent", req.Msg.GetAgentName(),
		"session_id", req.Msg.GetSessionId(),
		"elapsed_ms", time.Since(start).Milliseconds(),
	)
	return connect.NewResponse(&agentsv1.ReplySessionResponse{Response: response}), nil
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
