package application

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/genai"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/repo/forum"
	forummemory "go.orx.me/apps/butter/internal/repo/forum/memory"
	"go.orx.me/apps/butter/internal/runtime/runner"
	"go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

type blockingForumRunner struct {
	release chan struct{}
	done    chan struct{}
}

type failingRecentPostsRepo struct {
	forum.Repository
	err error
}

func (r *failingRecentPostsRepo) ListRecentPosts(context.Context, string, string, int32) ([]*agentsv1.ForumPost, error) {
	return nil, r.err
}

func newBlockingForumRunner() *blockingForumRunner {
	return &blockingForumRunner{
		release: make(chan struct{}),
		done:    make(chan struct{}),
	}
}

func (r *blockingForumRunner) Run(context.Context, string, []*genai.Part, string, *agentsv1.ContextInfo, runner.EventCallback, runner.CompactionCallback) (string, error) {
	<-r.release
	close(r.done)
	return "agent response", nil
}

func TestForumServiceInvokeAgentInThreadRejectsConcurrentProcessing(t *testing.T) {
	ctx := workspace.WithID(context.Background(), "ws-forum")
	repo := forummemory.New()
	now := timestamppb.New(time.Now().UTC())
	if err := repo.CreateThread(ctx, &agentsv1.ForumThread{
		Id:          "thread-1",
		Title:       "Question",
		Body:        "Body",
		Status:      forumStatusOpen,
		CreatedAt:   now,
		UpdatedAt:   now,
		WorkspaceId: "ws-forum",
	}); err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}
	fakeRunner := newBlockingForumRunner()
	svc := NewForumServiceServer(repo)
	svc.runnerSvc = fakeRunner

	start := make(chan struct{})
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := svc.InvokeAgentInThread(ctx, connect.NewRequest(&agentsv1.InvokeAgentInThreadRequest{
				ThreadId:  "thread-1",
				AgentName: "agent-1",
				Message:   "please help",
			}))
			errs <- err
		}()
	}
	close(start)
	wg.Wait()
	close(errs)

	successes := 0
	preconditions := 0
	for err := range errs {
		if err == nil {
			successes++
			continue
		}
		var twerr *connect.Error
		if errors.As(err, &twerr) && twerr.Code() == connect.CodeFailedPrecondition {
			preconditions++
			continue
		}
		t.Fatalf("InvokeAgentInThread() unexpected error = %v", err)
	}
	if successes != 1 || preconditions != 1 {
		t.Fatalf("InvokeAgentInThread() successes=%d failed_preconditions=%d, want 1 and 1", successes, preconditions)
	}

	thread, err := repo.GetThread(ctx, "ws-forum", "thread-1")
	if err != nil {
		t.Fatalf("GetThread() error = %v", err)
	}
	if got := thread.GetStatus(); got != forumStatusRunning {
		t.Fatalf("thread status = %q, want %q", got, forumStatusRunning)
	}
	posts, _, total, err := repo.ListPosts(ctx, forum.PostListFilter{WorkspaceID: "ws-forum", ThreadID: "thread-1"}, 10, "")
	if err != nil {
		t.Fatalf("ListPosts() error = %v", err)
	}
	if total != 1 || len(posts) != 1 {
		t.Fatalf("posts total=%d len=%d, want one queued user post", total, len(posts))
	}

	close(fakeRunner.release)
	select {
	case <-fakeRunner.done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for queued runner to finish")
	}
}

func TestForumServiceInvokeAgentInThreadRecentPostsFailureDoesNotMarkProcessing(t *testing.T) {
	ctx := workspace.WithID(context.Background(), "ws-forum")
	baseRepo := forummemory.New()
	now := timestamppb.New(time.Now().UTC())
	if err := baseRepo.CreateThread(ctx, &agentsv1.ForumThread{
		Id:          "thread-1",
		Title:       "Question",
		Body:        "Body",
		Status:      forumStatusOpen,
		CreatedAt:   now,
		UpdatedAt:   now,
		WorkspaceId: "ws-forum",
	}); err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}
	repo := &failingRecentPostsRepo{
		Repository: baseRepo,
		err:        errors.New("list recent posts failed"),
	}
	svc := NewForumServiceServer(repo)
	svc.runnerSvc = newBlockingForumRunner()

	_, err := svc.InvokeAgentInThread(ctx, connect.NewRequest(&agentsv1.InvokeAgentInThreadRequest{
		ThreadId:  "thread-1",
		AgentName: "agent-1",
		Message:   "please help",
	}))
	var twerr *connect.Error
	if !errors.As(err, &twerr) || twerr.Code() != connect.CodeInternal {
		t.Fatalf("InvokeAgentInThread() error = %v, want internal connect error", err)
	}

	thread, err := baseRepo.GetThread(ctx, "ws-forum", "thread-1")
	if err != nil {
		t.Fatalf("GetThread() error = %v", err)
	}
	if got := thread.GetStatus(); got != forumStatusOpen {
		t.Fatalf("thread status = %q, want %q", got, forumStatusOpen)
	}
	_, _, total, err := baseRepo.ListPosts(ctx, forum.PostListFilter{WorkspaceID: "ws-forum", ThreadID: "thread-1"}, 10, "")
	if err != nil {
		t.Fatalf("ListPosts() error = %v", err)
	}
	if total != 0 {
		t.Fatalf("posts total=%d, want no queued user post", total)
	}
}
