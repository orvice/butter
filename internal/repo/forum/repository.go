package forum

import (
	"context"
	"errors"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	ErrThreadNotFound   = errors.New("forum thread not found")
	ErrThreadProcessing = errors.New("forum thread already has an agent processing")
	ErrPostNotFound     = errors.New("forum post not found")
)

type ThreadListFilter struct {
	WorkspaceID string
	Status      string
	Label       string
}

type PostListFilter struct {
	WorkspaceID string
	ThreadID    string
}

type ProcessingState struct {
	AgentName    string
	InvocationID string
	StartedAt    *timestamppb.Timestamp
}

type Repository interface {
	EnsureIndexes(ctx context.Context) error
	CreateThread(ctx context.Context, thread *agentsv1.ForumThread) error
	UpdateThread(ctx context.Context, thread *agentsv1.ForumThread) error
	CreatePostAndMarkThreadProcessing(ctx context.Context, post *agentsv1.ForumPost, processing ProcessingState) (*agentsv1.ForumThread, error)
	GetThread(ctx context.Context, workspaceID, id string) (*agentsv1.ForumThread, error)
	ListThreads(ctx context.Context, filter ThreadListFilter, pageSize int32, pageToken string) ([]*agentsv1.ForumThread, string, int32, error)
	DeleteThread(ctx context.Context, workspaceID, id string) error

	CreatePost(ctx context.Context, post *agentsv1.ForumPost) error
	GetPost(ctx context.Context, workspaceID, threadID, postID string) (*agentsv1.ForumPost, error)
	ListPosts(ctx context.Context, filter PostListFilter, pageSize int32, pageToken string) ([]*agentsv1.ForumPost, string, int32, error)
	ListRecentPosts(ctx context.Context, workspaceID, threadID string, limit int32) ([]*agentsv1.ForumPost, error)
	DeletePost(ctx context.Context, workspaceID, threadID, postID string) error
	DeleteThreadPosts(ctx context.Context, workspaceID, threadID string) error
}
