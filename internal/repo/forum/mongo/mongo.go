package mongo

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"go.orx.me/apps/butter/internal/repo/forum"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const (
	threadsCollection = "forum_threads"
	postsCollection   = "forum_posts"
)

type Store struct {
	db      *mongo.Database
	threads *mongo.Collection
	posts   *mongo.Collection
}

func New(db *mongo.Database) *Store {
	return &Store{
		db:      db,
		threads: db.Collection(threadsCollection),
		posts:   db.Collection(postsCollection),
	}
}

func (s *Store) EnsureIndexes(ctx context.Context) error {
	if _, err := s.threads.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "workspace_id", Value: 1}, {Key: "updated_at", Value: -1}}},
		{Keys: bson.D{{Key: "workspace_id", Value: 1}, {Key: "status", Value: 1}}},
	}); err != nil {
		return fmt.Errorf("create forum thread indexes: %w", err)
	}
	if _, err := s.posts.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "workspace_id", Value: 1}, {Key: "thread_id", Value: 1}, {Key: "created_at", Value: 1}}},
	}); err != nil {
		return fmt.Errorf("create forum post indexes: %w", err)
	}
	return nil
}

func (s *Store) CreateThread(ctx context.Context, thread *agentsv1.ForumThread) error {
	if _, err := s.threads.InsertOne(ctx, thread); err != nil {
		return fmt.Errorf("insert forum thread: %w", err)
	}
	return nil
}

func (s *Store) UpdateThread(ctx context.Context, thread *agentsv1.ForumThread) error {
	res, err := s.threads.ReplaceOne(ctx, bson.M{"_id": thread.GetId(), "workspace_id": thread.GetWorkspaceId()}, thread)
	if err != nil {
		return fmt.Errorf("update forum thread: %w", err)
	}
	if res.MatchedCount == 0 {
		return forum.ErrThreadNotFound
	}
	return nil
}

func (s *Store) CreatePostAndMarkThreadProcessing(ctx context.Context, post *agentsv1.ForumPost, processing forum.ProcessingState) (*agentsv1.ForumThread, error) {
	session, err := s.db.Client().StartSession()
	if err != nil {
		return nil, fmt.Errorf("start forum processing transaction: %w", err)
	}
	defer session.EndSession(ctx)

	var thread *agentsv1.ForumThread
	_, err = session.WithTransaction(ctx, func(txCtx context.Context) (any, error) {
		update := bson.M{
			"$set": bson.M{
				"status":                            "processing",
				"updated_at":                        processing.StartedAt,
				"metadata.processing_agent":         processing.AgentName,
				"metadata.processing_invocation_id": processing.InvocationID,
				"metadata.processing_started_at":    processing.StartedAt.AsTime().Format(time.RFC3339),
			},
			"$unset": bson.M{"metadata.processing_error": ""},
		}
		res, err := s.threads.UpdateOne(txCtx,
			bson.M{"_id": post.GetThreadId(), "workspace_id": post.GetWorkspaceId(), "status": bson.M{"$ne": "processing"}},
			update,
		)
		if err != nil {
			return nil, fmt.Errorf("mark forum thread processing: %w", err)
		}
		if res.MatchedCount == 0 {
			state, err := s.GetThread(txCtx, post.GetWorkspaceId(), post.GetThreadId())
			if errors.Is(err, forum.ErrThreadNotFound) {
				return nil, err
			}
			if err != nil {
				return nil, err
			}
			if state.GetStatus() == "processing" {
				return nil, forum.ErrThreadProcessing
			}
			return nil, forum.ErrThreadNotFound
		}
		if _, err := s.posts.InsertOne(txCtx, post); err != nil {
			return nil, fmt.Errorf("insert forum processing post: %w", err)
		}
		var updated agentsv1.ForumThread
		if err := s.threads.FindOne(txCtx, bson.M{"_id": post.GetThreadId(), "workspace_id": post.GetWorkspaceId()}).Decode(&updated); err != nil {
			return nil, fmt.Errorf("get marked forum thread: %w", err)
		}
		thread = &updated
		return nil, nil
	})
	if err != nil {
		return nil, err
	}
	return thread, nil
}

func (s *Store) GetThread(ctx context.Context, workspaceID, id string) (*agentsv1.ForumThread, error) {
	var thread agentsv1.ForumThread
	err := s.threads.FindOne(ctx, bson.M{"_id": id, "workspace_id": workspaceID}).Decode(&thread)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, forum.ErrThreadNotFound
		}
		return nil, fmt.Errorf("get forum thread: %w", err)
	}
	return &thread, nil
}

func (s *Store) ListThreads(ctx context.Context, filter forum.ThreadListFilter, pageSize int32, pageToken string) ([]*agentsv1.ForumThread, string, int32, error) {
	q := bson.M{"workspace_id": filter.WorkspaceID}
	if filter.Status != "" {
		q["status"] = filter.Status
	}
	total, err := s.threads.CountDocuments(ctx, q)
	if err != nil {
		return nil, "", 0, fmt.Errorf("count forum threads: %w", err)
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := decodeToken(pageToken)
	cursor, err := s.threads.Find(ctx, q, options.Find().SetSort(bson.D{{Key: "updated_at", Value: -1}}).SetSkip(int64(offset)).SetLimit(int64(pageSize)))
	if err != nil {
		return nil, "", 0, fmt.Errorf("list forum threads: %w", err)
	}
	defer cursor.Close(ctx)
	var out []*agentsv1.ForumThread
	if err := cursor.All(ctx, &out); err != nil {
		return nil, "", 0, fmt.Errorf("decode forum threads: %w", err)
	}
	next := ""
	if int64(offset)+int64(len(out)) < total {
		next = encodeToken(offset + len(out))
	}
	return out, next, int32(total), nil
}

func (s *Store) DeleteThread(ctx context.Context, workspaceID, id string) error {
	res, err := s.threads.DeleteOne(ctx, bson.M{"_id": id, "workspace_id": workspaceID})
	if err != nil {
		return fmt.Errorf("delete forum thread: %w", err)
	}
	if res.DeletedCount == 0 {
		return forum.ErrThreadNotFound
	}
	return s.DeleteThreadPosts(ctx, workspaceID, id)
}

func (s *Store) CreatePost(ctx context.Context, post *agentsv1.ForumPost) error {
	if _, err := s.posts.InsertOne(ctx, post); err != nil {
		return fmt.Errorf("insert forum post: %w", err)
	}
	_, err := s.threads.UpdateOne(ctx,
		bson.M{"_id": post.GetThreadId(), "workspace_id": post.GetWorkspaceId()},
		bson.M{"$set": bson.M{"updated_at": post.GetCreatedAt()}},
	)
	if err != nil {
		return fmt.Errorf("touch forum thread: %w", err)
	}
	return nil
}

func (s *Store) GetPost(ctx context.Context, workspaceID, threadID, postID string) (*agentsv1.ForumPost, error) {
	var post agentsv1.ForumPost
	err := s.posts.FindOne(ctx, bson.M{"_id": postID, "workspace_id": workspaceID, "thread_id": threadID}).Decode(&post)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, forum.ErrPostNotFound
		}
		return nil, fmt.Errorf("get forum post: %w", err)
	}
	return &post, nil
}

func (s *Store) ListPosts(ctx context.Context, filter forum.PostListFilter, pageSize int32, pageToken string) ([]*agentsv1.ForumPost, string, int32, error) {
	q := bson.M{"workspace_id": filter.WorkspaceID, "thread_id": filter.ThreadID}
	total, err := s.posts.CountDocuments(ctx, q)
	if err != nil {
		return nil, "", 0, fmt.Errorf("count forum posts: %w", err)
	}
	if pageSize <= 0 {
		pageSize = 100
	}
	offset := decodeToken(pageToken)
	cursor, err := s.posts.Find(ctx, q, options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}}).SetSkip(int64(offset)).SetLimit(int64(pageSize)))
	if err != nil {
		return nil, "", 0, fmt.Errorf("list forum posts: %w", err)
	}
	defer cursor.Close(ctx)
	var out []*agentsv1.ForumPost
	if err := cursor.All(ctx, &out); err != nil {
		return nil, "", 0, fmt.Errorf("decode forum posts: %w", err)
	}
	next := ""
	if int64(offset)+int64(len(out)) < total {
		next = encodeToken(offset + len(out))
	}
	return out, next, int32(total), nil
}

func (s *Store) ListRecentPosts(ctx context.Context, workspaceID, threadID string, limit int32) ([]*agentsv1.ForumPost, error) {
	if limit <= 0 {
		limit = 20
	}
	cursor, err := s.posts.Find(ctx,
		bson.M{"workspace_id": workspaceID, "thread_id": threadID},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(int64(limit)),
	)
	if err != nil {
		return nil, fmt.Errorf("list recent forum posts: %w", err)
	}
	defer cursor.Close(ctx)
	var out []*agentsv1.ForumPost
	if err := cursor.All(ctx, &out); err != nil {
		return nil, fmt.Errorf("decode recent forum posts: %w", err)
	}
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out, nil
}

func (s *Store) DeletePost(ctx context.Context, workspaceID, threadID, postID string) error {
	res, err := s.posts.DeleteOne(ctx, bson.M{"_id": postID, "workspace_id": workspaceID, "thread_id": threadID})
	if err != nil {
		return fmt.Errorf("delete forum post: %w", err)
	}
	if res.DeletedCount == 0 {
		return forum.ErrPostNotFound
	}
	return nil
}

func (s *Store) DeleteThreadPosts(ctx context.Context, workspaceID, threadID string) error {
	_, err := s.posts.DeleteMany(ctx, bson.M{"workspace_id": workspaceID, "thread_id": threadID})
	if err != nil {
		return fmt.Errorf("delete forum thread posts: %w", err)
	}
	return nil
}

func decodeToken(token string) int {
	if token == "" {
		return 0
	}
	raw, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return 0
	}
	n, err := strconv.Atoi(string(raw))
	if err != nil || n < 0 {
		return 0
	}
	return n
}

func encodeToken(offset int) string {
	return base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(offset)))
}
