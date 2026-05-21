package memory

import (
	"context"
	"sort"
	"strconv"
	"sync"

	"google.golang.org/protobuf/proto"

	"go.orx.me/apps/butter/internal/repo/forum"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type Store struct {
	mu      sync.RWMutex
	threads map[string]*agentsv1.ForumThread
	posts   map[string]*agentsv1.ForumPost
}

func New() *Store {
	return &Store{
		threads: make(map[string]*agentsv1.ForumThread),
		posts:   make(map[string]*agentsv1.ForumPost),
	}
}

func (s *Store) EnsureIndexes(context.Context) error { return nil }

func (s *Store) CreateThread(_ context.Context, thread *agentsv1.ForumThread) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.threads[thread.GetId()] = proto.Clone(thread).(*agentsv1.ForumThread)
	return nil
}

func (s *Store) UpdateThread(_ context.Context, thread *agentsv1.ForumThread) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.threads[thread.GetId()]; !ok {
		return forum.ErrThreadNotFound
	}
	s.threads[thread.GetId()] = proto.Clone(thread).(*agentsv1.ForumThread)
	return nil
}

func (s *Store) GetThread(_ context.Context, workspaceID, id string) (*agentsv1.ForumThread, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	thread, ok := s.threads[id]
	if !ok || thread.GetWorkspaceId() != workspaceID {
		return nil, forum.ErrThreadNotFound
	}
	return proto.Clone(thread).(*agentsv1.ForumThread), nil
}

func (s *Store) ListThreads(_ context.Context, filter forum.ThreadListFilter, pageSize int32, pageToken string) ([]*agentsv1.ForumThread, string, int32, error) {
	s.mu.RLock()
	items := make([]*agentsv1.ForumThread, 0, len(s.threads))
	for _, thread := range s.threads {
		if filter.WorkspaceID != "" && thread.GetWorkspaceId() != filter.WorkspaceID {
			continue
		}
		if filter.Status != "" && thread.GetStatus() != filter.Status {
			continue
		}
		items = append(items, proto.Clone(thread).(*agentsv1.ForumThread))
	}
	s.mu.RUnlock()

	sort.SliceStable(items, func(i, j int) bool {
		return items[i].GetUpdatedAt().AsTime().After(items[j].GetUpdatedAt().AsTime())
	})
	page, next := paginate(items, pageSize, pageToken)
	return page, next, int32(len(items)), nil
}

func (s *Store) DeleteThread(_ context.Context, workspaceID, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	thread, ok := s.threads[id]
	if !ok || thread.GetWorkspaceId() != workspaceID {
		return forum.ErrThreadNotFound
	}
	delete(s.threads, id)
	for postID, post := range s.posts {
		if post.GetWorkspaceId() == workspaceID && post.GetThreadId() == id {
			delete(s.posts, postID)
		}
	}
	return nil
}

func (s *Store) CreatePost(_ context.Context, post *agentsv1.ForumPost) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.posts[post.GetId()] = proto.Clone(post).(*agentsv1.ForumPost)
	return nil
}

func (s *Store) GetPost(_ context.Context, workspaceID, threadID, postID string) (*agentsv1.ForumPost, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	post, ok := s.posts[postID]
	if !ok || post.GetWorkspaceId() != workspaceID || post.GetThreadId() != threadID {
		return nil, forum.ErrPostNotFound
	}
	return proto.Clone(post).(*agentsv1.ForumPost), nil
}

func (s *Store) ListPosts(_ context.Context, filter forum.PostListFilter, pageSize int32, pageToken string) ([]*agentsv1.ForumPost, string, int32, error) {
	items := s.matchPosts(filter.WorkspaceID, filter.ThreadID)
	page, next := paginate(items, pageSize, pageToken)
	return page, next, int32(len(items)), nil
}

func (s *Store) ListRecentPosts(_ context.Context, workspaceID, threadID string, limit int32) ([]*agentsv1.ForumPost, error) {
	items := s.matchPosts(workspaceID, threadID)
	if limit <= 0 || int(limit) > len(items) {
		return items, nil
	}
	return items[len(items)-int(limit):], nil
}

func (s *Store) matchPosts(workspaceID, threadID string) []*agentsv1.ForumPost {
	s.mu.RLock()
	items := make([]*agentsv1.ForumPost, 0)
	for _, post := range s.posts {
		if workspaceID != "" && post.GetWorkspaceId() != workspaceID {
			continue
		}
		if threadID != "" && post.GetThreadId() != threadID {
			continue
		}
		items = append(items, proto.Clone(post).(*agentsv1.ForumPost))
	}
	s.mu.RUnlock()
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].GetCreatedAt().AsTime().Before(items[j].GetCreatedAt().AsTime())
	})
	return items
}

func (s *Store) DeletePost(_ context.Context, workspaceID, threadID, postID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	post, ok := s.posts[postID]
	if !ok || post.GetWorkspaceId() != workspaceID || post.GetThreadId() != threadID {
		return forum.ErrPostNotFound
	}
	delete(s.posts, postID)
	return nil
}

func (s *Store) DeleteThreadPosts(_ context.Context, workspaceID, threadID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for postID, post := range s.posts {
		if post.GetWorkspaceId() == workspaceID && post.GetThreadId() == threadID {
			delete(s.posts, postID)
		}
	}
	return nil
}

func paginate[T any](items []T, pageSize int32, pageToken string) ([]T, string) {
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := 0
	if pageToken != "" {
		if n, err := strconv.Atoi(pageToken); err == nil && n >= 0 {
			offset = n
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
		next = strconv.Itoa(end)
	}
	return items[offset:end], next
}
