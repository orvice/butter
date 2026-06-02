package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"butterfly.orx.me/core/log"
	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/genai"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/repo/auth"
	"go.orx.me/apps/butter/internal/repo/forum"
	"go.orx.me/apps/butter/internal/runtime/runner"
	"go.orx.me/apps/butter/internal/transport/connectx"
	wsctx "go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const (
	forumAppName       = "forum"
	forumStatusOpen    = "open"
	forumStatusError   = "error"
	forumStatusRunning = "processing"
	forumAuthorUser    = "user"
	forumAuthorAgent   = "agent"
	forumAuthorSystem  = "system"
	defaultRecentPosts = 20
)

type ForumServiceServer struct {
	mu        sync.RWMutex
	repo      forum.Repository
	runnerSvc forumAgentRunner
}

type forumAgentRunner interface {
	Run(ctx context.Context, agentName string, parts []*genai.Part, modelOverride string, ctxInfo *agentsv1.ContextInfo, onEvent runner.EventCallback, onCompaction runner.CompactionCallback) (string, error)
}

func NewForumServiceServer(repo forum.Repository) *ForumServiceServer {
	return &ForumServiceServer{repo: repo}
}

func (s *ForumServiceServer) SetRepo(repo forum.Repository) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.repo = repo
}

func (s *ForumServiceServer) SetRunnerService(runnerSvc *runner.Service) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.runnerSvc = runnerSvc
}

func (s *ForumServiceServer) getRepo() forum.Repository {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.repo
}

func (s *ForumServiceServer) getRunner() forumAgentRunner {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.runnerSvc
}

func (s *ForumServiceServer) ListThreads(ctx context.Context, req *agentsv1.ListThreadsRequest) (*agentsv1.ListThreadsResponse, error) {
	repo, workspaceID, err := s.requireRepoWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	threads, next, total, err := repo.ListThreads(ctx, forum.ThreadListFilter{WorkspaceID: workspaceID, Status: strings.TrimSpace(req.GetStatus()), Label: strings.TrimSpace(req.GetLabel())}, req.GetPageSize(), req.GetPageToken())
	if err != nil {
		return nil, connectx.InternalWith(err)
	}
	return &agentsv1.ListThreadsResponse{Threads: threads, NextPageToken: next, Total: total}, nil
}

func (s *ForumServiceServer) ListThreadLabels(ctx context.Context, _ *agentsv1.ListThreadLabelsRequest) (*agentsv1.ListThreadLabelsResponse, error) {
	repo, workspaceID, err := s.requireRepoWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	labels, err := repo.ListThreadLabels(ctx, workspaceID)
	if err != nil {
		return nil, connectx.InternalWith(err)
	}
	return &agentsv1.ListThreadLabelsResponse{Labels: labels}, nil
}

func (s *ForumServiceServer) GetThread(ctx context.Context, req *agentsv1.GetThreadRequest) (*agentsv1.GetThreadResponse, error) {
	repo, workspaceID, err := s.requireRepoWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	thread, err := repo.GetThread(ctx, workspaceID, req.GetId())
	if err != nil {
		return nil, mapForumErr(err)
	}
	posts, next, total, err := repo.ListPosts(ctx, forum.PostListFilter{WorkspaceID: workspaceID, ThreadID: req.GetId()}, req.GetPostPageSize(), req.GetPostPageToken())
	if err != nil {
		return nil, connectx.InternalWith(err)
	}
	return &agentsv1.GetThreadResponse{Thread: thread, Posts: posts, NextPostPageToken: next, PostTotal: total}, nil
}

func (s *ForumServiceServer) CreateThread(ctx context.Context, req *agentsv1.CreateThreadRequest) (*agentsv1.CreateThreadResponse, error) {
	repo, workspaceID, err := s.requireRepoWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	userID := authorUserID(ctx)
	title := strings.TrimSpace(req.GetTitle())
	body := strings.TrimSpace(req.GetBody())
	if title == "" {
		return nil, connectx.RequiredArgument("title")
	}
	if body == "" {
		return nil, connectx.RequiredArgument("body")
	}
	now := timestamppb.New(time.Now().UTC())
	threadID := uuid.NewString()
	thread := &agentsv1.ForumThread{
		Id:          threadID,
		Title:       title,
		Body:        body,
		CreatedBy:   userID,
		Status:      forumStatusOpen,
		AgentNames:  append([]string(nil), req.GetAgentNames()...),
		Labels:      normalizeLabels(req.GetLabels()),
		Metadata:    copyStringMap(req.GetMetadata()),
		CreatedAt:   now,
		UpdatedAt:   now,
		WorkspaceId: workspaceID,
	}
	post := &agentsv1.ForumPost{
		Id:           uuid.NewString(),
		ThreadId:     threadID,
		Body:         body,
		AuthorUserId: userID,
		AuthorKind:   forumAuthorUser,
		CreatedAt:    now,
		UpdatedAt:    now,
		WorkspaceId:  workspaceID,
	}
	if err := repo.CreateThread(ctx, thread); err != nil {
		return nil, connectx.InternalWith(err)
	}
	if err := repo.CreatePost(ctx, post); err != nil {
		_ = repo.DeleteThread(ctx, workspaceID, threadID)
		return nil, connectx.InternalWith(err)
	}
	if runnerSvc := s.getRunner(); runnerSvc != nil && len(thread.GetAgentNames()) > 0 {
		bgCtx := wsctx.WithID(context.Background(), workspaceID)
		for _, agentName := range thread.GetAgentNames() {
			agentName := agentName
			go s.invokeAgentPost(bgCtx, repo, runnerSvc, thread, []*agentsv1.ForumPost{post}, agentName, userID)
		}
	}
	return &agentsv1.CreateThreadResponse{Thread: thread, FirstPost: post}, nil
}

func (s *ForumServiceServer) invokeAgentPost(ctx context.Context, repo forum.Repository, runnerSvc forumAgentRunner, thread *agentsv1.ForumThread, posts []*agentsv1.ForumPost, agentName, userID string) {
	logger := log.FromContext(ctx)
	prompt := buildForumPrompt(thread, posts, userID, "")
	invocationID := uuid.NewString()
	if id, err := uuid.NewV7(); err == nil {
		invocationID = id.String()
	}
	ctxInfo := &agentsv1.ContextInfo{
		Uuid:        invocationID,
		ChannelName: forumAppName,
		SessionId:   thread.GetId(),
		UserId:      "forum:" + thread.GetWorkspaceId(),
		Source:      agentsv1.ContextSource_CONTEXT_SOURCE_API,
		ChatType:    agentsv1.ChatType_CHAT_TYPE_PRIVATE,
		WorkspaceId: thread.GetWorkspaceId(),
	}
	logger.Info("auto-invoking forum agent on thread create", "thread_id", thread.GetId(), "agent", agentName)
	response, err := runnerSvc.Run(ctx, agentName, []*genai.Part{genai.NewPartFromText(prompt)}, "", ctxInfo, nil, nil)
	if err != nil {
		logger.Error("auto-invoke agent failed", "thread_id", thread.GetId(), "agent", agentName, "error", err)
		return
	}
	now := timestamppb.New(time.Now().UTC())
	agentPost := &agentsv1.ForumPost{
		Id:              uuid.NewString(),
		ThreadId:        thread.GetId(),
		Body:            response,
		AuthorAgentName: agentName,
		AuthorKind:      forumAuthorAgent,
		InvocationId:    invocationID,
		CreatedAt:       now,
		UpdatedAt:       now,
		WorkspaceId:     thread.GetWorkspaceId(),
	}
	if err := repo.CreatePost(ctx, agentPost); err != nil {
		logger.Error("failed to save agent post", "thread_id", thread.GetId(), "agent", agentName, "error", err)
	}
}

func (s *ForumServiceServer) UpdateThread(ctx context.Context, req *agentsv1.UpdateThreadRequest) (*agentsv1.UpdateThreadResponse, error) {
	repo, workspaceID, err := s.requireRepoWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	thread, err := repo.GetThread(ctx, workspaceID, req.GetId())
	if err != nil {
		return nil, mapForumErr(err)
	}
	if v := strings.TrimSpace(req.GetTitle()); v != "" {
		thread.Title = v
	}
	if req.GetBody() != "" {
		thread.Body = req.GetBody()
	}
	if v := strings.TrimSpace(req.GetStatus()); v != "" {
		thread.Status = v
	}
	if req.GetAgentNames() != nil {
		thread.AgentNames = append([]string(nil), req.GetAgentNames()...)
	}
	if req.GetLabels() != nil {
		thread.Labels = normalizeLabels(req.GetLabels())
	}
	if req.GetMetadata() != nil {
		thread.Metadata = copyStringMap(req.GetMetadata())
	}
	thread.UpdatedAt = timestamppb.New(time.Now().UTC())
	if err := repo.UpdateThread(ctx, thread); err != nil {
		return nil, mapForumErr(err)
	}
	return &agentsv1.UpdateThreadResponse{Thread: thread}, nil
}

func (s *ForumServiceServer) DeleteThread(ctx context.Context, req *agentsv1.DeleteThreadRequest) (*agentsv1.DeleteThreadResponse, error) {
	repo, workspaceID, err := s.requireRepoWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if err := repo.DeleteThread(ctx, workspaceID, req.GetId()); err != nil {
		return nil, mapForumErr(err)
	}
	return &agentsv1.DeleteThreadResponse{}, nil
}

func (s *ForumServiceServer) CreatePost(ctx context.Context, req *agentsv1.CreatePostRequest) (*agentsv1.CreatePostResponse, error) {
	repo, workspaceID, err := s.requireRepoWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := repo.GetThread(ctx, workspaceID, req.GetThreadId()); err != nil {
		return nil, mapForumErr(err)
	}
	body := strings.TrimSpace(req.GetBody())
	if body == "" {
		return nil, connectx.RequiredArgument("body")
	}
	now := timestamppb.New(time.Now().UTC())
	post := &agentsv1.ForumPost{
		Id:           uuid.NewString(),
		ThreadId:     req.GetThreadId(),
		Body:         body,
		AuthorUserId: authorUserID(ctx),
		AuthorKind:   forumAuthorUser,
		ParentPostId: req.GetParentPostId(),
		CreatedAt:    now,
		UpdatedAt:    now,
		WorkspaceId:  workspaceID,
	}
	if err := repo.CreatePost(ctx, post); err != nil {
		return nil, connectx.InternalWith(err)
	}
	return &agentsv1.CreatePostResponse{Post: post}, nil
}

func (s *ForumServiceServer) DeletePost(ctx context.Context, req *agentsv1.DeletePostRequest) (*agentsv1.DeletePostResponse, error) {
	repo, workspaceID, err := s.requireRepoWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if err := repo.DeletePost(ctx, workspaceID, req.GetThreadId(), req.GetPostId()); err != nil {
		return nil, mapForumErr(err)
	}
	return &agentsv1.DeletePostResponse{}, nil
}

func (s *ForumServiceServer) InvokeAgentInThread(ctx context.Context, req *agentsv1.InvokeAgentInThreadRequest) (*agentsv1.InvokeAgentInThreadResponse, error) {
	repo, workspaceID, err := s.requireRepoWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	runnerSvc := s.getRunner()
	if runnerSvc == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("runner service not available"))
	}
	agentName := strings.TrimSpace(req.GetAgentName())
	if agentName == "" {
		return nil, connectx.RequiredArgument("agent_name")
	}
	if _, err := repo.GetThread(ctx, workspaceID, req.GetThreadId()); err != nil {
		return nil, mapForumErr(err)
	}
	now := timestamppb.New(time.Now().UTC())
	userID := authorUserID(ctx)
	message := strings.TrimSpace(req.GetMessage())
	postBody := message
	if postBody == "" {
		postBody = fmt.Sprintf("@%s", agentName)
	}
	limit := req.GetRecentPostLimit()
	if limit <= 0 {
		limit = defaultRecentPosts
	}
	posts, err := repo.ListRecentPosts(ctx, workspaceID, req.GetThreadId(), limit)
	if err != nil {
		return nil, connectx.InternalWith(err)
	}
	userPost := &agentsv1.ForumPost{
		Id:           uuid.NewString(),
		ThreadId:     req.GetThreadId(),
		Body:         postBody,
		AuthorUserId: userID,
		AuthorKind:   forumAuthorUser,
		CreatedAt:    now,
		UpdatedAt:    now,
		WorkspaceId:  workspaceID,
	}
	invocationID := uuid.NewString()
	if id, err := uuid.NewV7(); err == nil {
		invocationID = id.String()
	}
	processing := forum.ProcessingState{AgentName: agentName, InvocationID: invocationID, StartedAt: now}
	thread, err := repo.CreatePostAndMarkThreadProcessing(ctx, userPost, processing)
	if err != nil {
		return nil, mapForumErr(err)
	}
	posts = appendRecentForumPost(posts, userPost, limit)
	logger := log.FromContext(ctx)
	logger.Info("queued forum agent invocation", "thread_id", req.GetThreadId(), "agent", agentName, "workspace_id", workspaceID, "invocation_id", invocationID)
	bgCtx := wsctx.WithID(context.Background(), workspaceID)
	go s.invokeAgentForUserPost(bgCtx, repo, runnerSvc, thread, posts, agentName, userID, message, req.GetModelOverride(), invocationID)
	return &agentsv1.InvokeAgentInThreadResponse{Post: userPost}, nil
}

func (s *ForumServiceServer) invokeAgentForUserPost(ctx context.Context, repo forum.Repository, runnerSvc forumAgentRunner, thread *agentsv1.ForumThread, posts []*agentsv1.ForumPost, agentName, userID, message, modelOverride, invocationID string) {
	logger := log.FromContext(ctx).With("thread_id", thread.GetId(), "agent", agentName, "invocation_id", invocationID)
	prompt := buildForumPrompt(thread, posts, userID, message)
	ctxInfo := &agentsv1.ContextInfo{
		Uuid:        invocationID,
		ChannelName: forumAppName,
		SessionId:   thread.GetId(),
		UserId:      "forum:" + thread.GetWorkspaceId(),
		Source:      agentsv1.ContextSource_CONTEXT_SOURCE_API,
		ChatType:    agentsv1.ChatType_CHAT_TYPE_PRIVATE,
		WorkspaceId: thread.GetWorkspaceId(),
	}
	logger.Info("running queued forum agent invocation")
	response, err := runnerSvc.Run(ctx, agentName, []*genai.Part{genai.NewPartFromText(prompt)}, modelOverride, ctxInfo, nil, nil)
	if err != nil {
		logger.Error("queued forum agent invocation failed", "error", err)
		s.createAgentFailurePost(ctx, repo, thread, agentName, invocationID, err)
		s.finishThreadProcessing(ctx, repo, thread.GetWorkspaceId(), thread.GetId(), forumStatusError, err.Error())
		return
	}
	now := timestamppb.New(time.Now().UTC())
	agentPost := &agentsv1.ForumPost{
		Id:              uuid.NewString(),
		ThreadId:        thread.GetId(),
		Body:            response,
		AuthorAgentName: agentName,
		AuthorKind:      forumAuthorAgent,
		InvocationId:    invocationID,
		CreatedAt:       now,
		UpdatedAt:       now,
		WorkspaceId:     thread.GetWorkspaceId(),
	}
	if err := repo.CreatePost(ctx, agentPost); err != nil {
		logger.Error("failed to save queued forum agent post", "error", err)
		s.createAgentFailurePost(ctx, repo, thread, agentName, invocationID, err)
		s.finishThreadProcessing(ctx, repo, thread.GetWorkspaceId(), thread.GetId(), forumStatusError, err.Error())
		return
	}
	s.finishThreadProcessing(ctx, repo, thread.GetWorkspaceId(), thread.GetId(), forumStatusOpen, "")
}

func (s *ForumServiceServer) createAgentFailurePost(ctx context.Context, repo forum.Repository, thread *agentsv1.ForumThread, agentName, invocationID string, runErr error) {
	now := timestamppb.New(time.Now().UTC())
	post := &agentsv1.ForumPost{
		Id:              uuid.NewString(),
		ThreadId:        thread.GetId(),
		Body:            fmt.Sprintf("%s failed to process this request: %s", agentName, runErr.Error()),
		AuthorKind:      forumAuthorSystem,
		InvocationId:    invocationID,
		CreatedAt:       now,
		UpdatedAt:       now,
		WorkspaceId:     thread.GetWorkspaceId(),
		AuthorAgentName: agentName,
	}
	if err := repo.CreatePost(ctx, post); err != nil {
		log.FromContext(ctx).Warn("failed to save forum agent failure post", "thread_id", thread.GetId(), "agent", agentName, "error", err)
	}
}

func (s *ForumServiceServer) finishThreadProcessing(ctx context.Context, repo forum.Repository, workspaceID, threadID, status, errText string) {
	thread, err := repo.GetThread(ctx, workspaceID, threadID)
	if err != nil {
		log.FromContext(ctx).Warn("failed to load forum thread while finishing agent invocation", "thread_id", threadID, "error", err)
		return
	}
	thread.Status = status
	thread.UpdatedAt = timestamppb.New(time.Now().UTC())
	if thread.Metadata == nil {
		thread.Metadata = map[string]string{}
	}
	delete(thread.Metadata, "processing_agent")
	delete(thread.Metadata, "processing_invocation_id")
	delete(thread.Metadata, "processing_started_at")
	if errText == "" {
		delete(thread.Metadata, "processing_error")
	} else {
		thread.Metadata["processing_error"] = errText
	}
	if err := repo.UpdateThread(ctx, thread); err != nil {
		log.FromContext(ctx).Warn("failed to update forum thread processing status", "thread_id", threadID, "error", err)
	}
}

func (s *ForumServiceServer) requireRepoWorkspace(ctx context.Context) (forum.Repository, string, error) {
	repo := s.getRepo()
	if repo == nil {
		return nil, "", connect.NewError(connect.CodeFailedPrecondition, errors.New("forum repository not available"))
	}
	workspaceID, ok := wsctx.FromContext(ctx)
	if !ok || strings.TrimSpace(workspaceID) == "" {
		return nil, "", connect.NewError(connect.CodeFailedPrecondition, errors.New("workspace required"))
	}
	return repo, workspaceID, nil
}

func authorUserID(ctx context.Context) string {
	if user, ok := auth.UserFromContext(ctx); ok && user.GetId() != "" {
		return user.GetId()
	}
	return "api"
}

func mapForumErr(err error) error {
	switch {
	case errors.Is(err, forum.ErrThreadNotFound), errors.Is(err, forum.ErrPostNotFound):
		return connectx.NotFound(err.Error())
	case errors.Is(err, forum.ErrThreadProcessing):
		return connect.NewError(connect.CodeFailedPrecondition, errors.New(err.Error()))
	default:
		return connectx.InternalWith(err)
	}
}

func normalizeLabels(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, label := range in {
		label = strings.TrimSpace(label)
		if label == "" {
			continue
		}
		if _, ok := seen[label]; ok {
			continue
		}
		seen[label] = struct{}{}
		out = append(out, label)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func copyStringMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func appendRecentForumPost(posts []*agentsv1.ForumPost, post *agentsv1.ForumPost, limit int32) []*agentsv1.ForumPost {
	if limit > 0 && int(limit) <= len(posts) {
		posts = posts[len(posts)-int(limit)+1:]
	}
	return append(posts, post)
}

func buildForumPrompt(thread *agentsv1.ForumThread, posts []*agentsv1.ForumPost, currentUserID, message string) string {
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, "You are participating in a Butter forum thread. Reply as the selected agent.\n\nThread title: %s\nThread body: %s\nCurrent user id: %s\n\nRecent posts:\n", thread.GetTitle(), thread.GetBody(), currentUserID)
	for _, post := range posts {
		author := post.GetAuthorUserId()
		if post.GetAuthorKind() == forumAuthorAgent || post.GetAuthorAgentName() != "" {
			author = "agent:" + post.GetAuthorAgentName()
		}
		if author == "" {
			author = post.GetAuthorKind()
		}
		_, _ = fmt.Fprintf(&b, "- %s: %s\n", author, post.GetBody())
	}
	if strings.TrimSpace(message) != "" {
		_, _ = fmt.Fprintf(&b, "\nNew instruction/message from %s:\n%s\n", currentUserID, strings.TrimSpace(message))
	}
	return b.String()
}

var _ = proto.Clone
