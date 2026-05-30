package memory

import (
	"context"
	"testing"
	"time"

	"go.orx.me/apps/butter/internal/repo/forum"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestCreatePostUpdatesThreadTimestampAndOrdering(t *testing.T) {
	ctx := context.Background()
	store := New()

	oldTime := time.Date(2026, 5, 21, 1, 0, 0, 0, time.UTC)
	newTime := oldTime.Add(time.Hour)
	replyTime := newTime.Add(time.Hour)

	for _, thread := range []*agentsv1.ForumThread{
		{
			Id:          "older-thread",
			WorkspaceId: "workspace-1",
			UpdatedAt:   timestamppb.New(oldTime),
		},
		{
			Id:          "newer-thread",
			WorkspaceId: "workspace-1",
			UpdatedAt:   timestamppb.New(newTime),
		},
	} {
		if err := store.CreateThread(ctx, thread); err != nil {
			t.Fatalf("CreateThread() error = %v", err)
		}
	}

	if err := store.CreatePost(ctx, &agentsv1.ForumPost{
		Id:          "reply-1",
		ThreadId:    "older-thread",
		WorkspaceId: "workspace-1",
		CreatedAt:   timestamppb.New(replyTime),
	}); err != nil {
		t.Fatalf("CreatePost() error = %v", err)
	}

	thread, err := store.GetThread(ctx, "workspace-1", "older-thread")
	if err != nil {
		t.Fatalf("GetThread() error = %v", err)
	}
	if got := thread.GetUpdatedAt().AsTime(); !got.Equal(replyTime) {
		t.Fatalf("thread UpdatedAt = %s, want %s", got, replyTime)
	}

	threads, _, total, err := store.ListThreads(ctx, forum.ThreadListFilter{WorkspaceID: "workspace-1"}, 10, "")
	if err != nil {
		t.Fatalf("ListThreads() error = %v", err)
	}
	if total != 2 {
		t.Fatalf("ListThreads() total = %d, want 2", total)
	}
	if got := threads[0].GetId(); got != "older-thread" {
		t.Fatalf("first listed thread = %q, want older-thread", got)
	}
}

func TestListThreadsFilterByLabel(t *testing.T) {
	ctx := context.Background()
	store := New()

	base := time.Date(2026, 5, 21, 1, 0, 0, 0, time.UTC)
	for _, thread := range []*agentsv1.ForumThread{
		{Id: "bug-thread", WorkspaceId: "workspace-1", Labels: []string{"bug", "urgent"}, UpdatedAt: timestamppb.New(base)},
		{Id: "question-thread", WorkspaceId: "workspace-1", Labels: []string{"question"}, UpdatedAt: timestamppb.New(base.Add(time.Hour))},
		{Id: "unlabeled-thread", WorkspaceId: "workspace-1", UpdatedAt: timestamppb.New(base.Add(2 * time.Hour))},
	} {
		if err := store.CreateThread(ctx, thread); err != nil {
			t.Fatalf("CreateThread() error = %v", err)
		}
	}

	threads, _, total, err := store.ListThreads(ctx, forum.ThreadListFilter{WorkspaceID: "workspace-1", Label: "bug"}, 10, "")
	if err != nil {
		t.Fatalf("ListThreads() error = %v", err)
	}
	if total != 1 {
		t.Fatalf("ListThreads() total = %d, want 1", total)
	}
	if got := threads[0].GetId(); got != "bug-thread" {
		t.Fatalf("filtered thread = %q, want bug-thread", got)
	}

	_, _, total, err = store.ListThreads(ctx, forum.ThreadListFilter{WorkspaceID: "workspace-1", Label: "missing"}, 10, "")
	if err != nil {
		t.Fatalf("ListThreads() error = %v", err)
	}
	if total != 0 {
		t.Fatalf("ListThreads() total = %d, want 0 for unknown label", total)
	}
}
