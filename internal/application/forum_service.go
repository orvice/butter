package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"butterfly.orx.me/core/log"
	"github.com/google/uuid"
	"github.com/twitchtv/twirp"
	"google.golang.org/genai"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/repo/auth"
	"go.orx.me/apps/butter/internal/repo/forum"
	"go.orx.me/apps/butter/internal/runtime/runner"
	wsctx "go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const (
	forumAppName       = "forum"
	forumStatusOpen    = "open"
	forumAuthorUser    = "user"
	forumAuthorAgent   = "agent"
	defaultRecentPosts = 20
)

type ForumServiceServer struct {
	mu        sync.RWMutex
	repo      forum.Repository
	runnerSvc *runner.Service
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

func (s *ForumServiceServer) getRunner() *runner.Service {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.runnerSvc
}

func (s *ForumServiceServer) ListThreads(ctx context.Context, req *agentsv1.ListForumThreadsRequest) (*agentsv1.ListForumThreadsResponse, error) {
	repo, workspaceID, err := s.requireRepoWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	threads, next, total, err := repo.ListThreads(ctx, forum.ThreadListFilter{WorkspaceID: workspaceID, Status: strings.TrimSpace(req.GetStatus())}, req.GetPageSize(), req.GetPageToken())
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}
	return &agentsv1.ListForumThreadsResponse{Threads: threads, NextPageToken: next, Total: total}, nil
}

func (s *ForumServiceServer) GetThread(ctx context.Context, req *agentsv1.GetForumThreadRequest) (*agentsv1.GetForumThreadResponse, error) {
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
		return nil, twirp.InternalErrorWith(err)
	}
	return &agentsv1.GetForumThreadResponse{Thread: thread, Posts: posts, NextPostPageToken: next, PostTotal: total}, nil
}

func (s *ForumServiceServer) CreateThread(ctx context.Context, req *agentsv1.CreateForumThreadRequest) (*agentsv1.CreateForumThreadResponse, error) {
	repo, workspaceID, err := s.requireRepoWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	userID := authorUserID(ctx)
	title := strings.TrimSpace(req.GetTitle())
	body := strings.TrimSpace(req.GetBody())
	if title == "" {
		return nil, twirp.RequiredArgumentError("title")
	}
	if body == "" {
		return nil, twirp.RequiredArgumentError("body")
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
		return nil, twirp.InternalErrorWith(err)
	}
	if err := repo.CreatePost(ctx, post); err != nil {
		_ = repo.DeleteThread(ctx, workspaceID, threadID)
		return nil, twirp.InternalErrorWith(err)
	}
	return &agentsv1.CreateForumThreadResponse{Thread: thread, FirstPost: post}, nil
}

func (s *ForumServiceServer) UpdateThread(ctx context.Context, req *agentsv1.UpdateForumThreadRequest) (*agentsv1.UpdateForumThreadResponse, error) {
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
	if req.GetMetadata() != nil {
		thread.Metadata = copyStringMap(req.GetMetadata())
	}
	thread.UpdatedAt = timestamppb.New(time.Now().UTC())
	if err := repo.UpdateThread(ctx, thread); err != nil {
		return nil, mapForumErr(err)
	}
	return &agentsv1.UpdateForumThreadResponse{Thread: thread}, nil
}

func (s *ForumServiceServer) DeleteThread(ctx context.Context, req *agentsv1.DeleteForumThreadRequest) (*agentsv1.DeleteForumThreadResponse, error) {
	repo, workspaceID, err := s.requireRepoWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if err := repo.DeleteThread(ctx, workspaceID, req.GetId()); err != nil {
		return nil, mapForumErr(err)
	}
	return &agentsv1.DeleteForumThreadResponse{}, nil
}

func (s *ForumServiceServer) CreatePost(ctx context.Context, req *agentsv1.CreateForumPostRequest) (*agentsv1.CreateForumPostResponse, error) {
	repo, workspaceID, err := s.requireRepoWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := repo.GetThread(ctx, workspaceID, req.GetThreadId()); err != nil {
		return nil, mapForumErr(err)
	}
	body := strings.TrimSpace(req.GetBody())
	if body == "" {
		return nil, twirp.RequiredArgumentError("body")
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
		return nil, twirp.InternalErrorWith(err)
	}
	return &agentsv1.CreateForumPostResponse{Post: post}, nil
}

func (s *ForumServiceServer) DeletePost(ctx context.Context, req *agentsv1.DeleteForumPostRequest) (*agentsv1.DeleteForumPostResponse, error) {
	repo, workspaceID, err := s.requireRepoWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	if err := repo.DeletePost(ctx, workspaceID, req.GetThreadId(), req.GetPostId()); err != nil {
		return nil, mapForumErr(err)
	}
	return &agentsv1.DeleteForumPostResponse{}, nil
}

func (s *ForumServiceServer) InvokeAgentInThread(ctx context.Context, req *agentsv1.InvokeAgentInThreadRequest) (*agentsv1.InvokeAgentInThreadResponse, error) {
	repo, workspaceID, err := s.requireRepoWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	runnerSvc := s.getRunner()
	if runnerSvc == nil {
		return nil, twirp.NewError(twirp.FailedPrecondition, "runner service not available")
	}
	agentName := strings.TrimSpace(req.GetAgentName())
	if agentName == "" {
		return nil, twirp.RequiredArgumentError("agent_name")
	}
	thread, err := repo.GetThread(ctx, workspaceID, req.GetThreadId())
	if err != nil {
		return nil, mapForumErr(err)
	}
	limit := req.GetRecentPostLimit()
	if limit <= 0 {
		limit = defaultRecentPosts
	}
	posts, err := repo.ListRecentPosts(ctx, workspaceID, req.GetThreadId(), limit)
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}
	prompt := buildForumPrompt(thread, posts, authorUserID(ctx), req.GetMessage())
	invocationID := uuid.NewString()
	if id, err := uuid.NewV7(); err == nil {
		invocationID = id.String()
	}
	ctxInfo := &agentsv1.ContextInfo{
		Uuid:        invocationID,
		ChannelName: forumAppName,
		SessionId:   req.GetThreadId(),
		UserId:      "forum:" + workspaceID,
		Source:      agentsv1.ContextSource_CONTEXT_SOURCE_API,
		ChatType:    agentsv1.ChatType_CHAT_TYPE_PRIVATE,
		WorkspaceId: workspaceID,
	}
	logger := log.FromContext(ctx)
	logger.Info("invoking forum agent", "thread_id", req.GetThreadId(), "agent", agentName, "workspace_id", workspaceID)
	response, err := runnerSvc.Run(ctx, agentName, []*genai.Part{genai.NewPartFromText(prompt)}, req.GetModelOverride(), ctxInfo, nil, nil)
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}
	now := timestamppb.New(time.Now().UTC())
	post := &agentsv1.ForumPost{
		Id:              uuid.NewString(),
		ThreadId:        req.GetThreadId(),
		Body:            response,
		AuthorAgentName: agentName,
		AuthorKind:      forumAuthorAgent,
		InvocationId:    invocationID,
		CreatedAt:       now,
		UpdatedAt:       now,
		WorkspaceId:     workspaceID,
	}
	if err := repo.CreatePost(ctx, post); err != nil {
		return nil, twirp.InternalErrorWith(err)
	}
	return &agentsv1.InvokeAgentInThreadResponse{Post: post, Response: response}, nil
}

func (s *ForumServiceServer) requireRepoWorkspace(ctx context.Context) (forum.Repository, string, error) {
	repo := s.getRepo()
	if repo == nil {
		return nil, "", twirp.NewError(twirp.FailedPrecondition, "forum repository not available")
	}
	workspaceID, ok := wsctx.FromContext(ctx)
	if !ok || strings.TrimSpace(workspaceID) == "" {
		return nil, "", twirp.NewError(twirp.FailedPrecondition, "workspace required")
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
		return twirp.NotFoundError(err.Error())
	default:
		return twirp.InternalErrorWith(err)
	}
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

var _ agentsv1.ForumService = (*ForumServiceServer)(nil)
var _ = proto.Clone
