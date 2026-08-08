package mongo_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"go.orx.me/apps/butter/internal/repo/repocache"
	repocachemongo "go.orx.me/apps/butter/internal/repo/repocache/mongo"
	"go.orx.me/apps/butter/internal/repo/repocache/repotest"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func testDB(t *testing.T) *mongo.Database {
	t.Helper()
	uri := os.Getenv("BUTTER_TEST_MONGO_URI")
	if uri == "" {
		t.Skip("BUTTER_TEST_MONGO_URI not set; skipping mongo integration test")
	}
	client, err := mongo.Connect(options.Client().ApplyURI(uri).SetTimeout(10 * time.Second))
	if err != nil {
		t.Fatalf("connect mongo: %v", err)
	}
	t.Cleanup(func() { _ = client.Disconnect(context.Background()) })
	db := client.Database(fmt.Sprintf("butter_repocache_test_%d", time.Now().UnixNano()))
	t.Cleanup(func() { _ = db.Drop(context.Background()) })
	return db
}

func newStore(t *testing.T, db *mongo.Database) *repocachemongo.Store {
	t.Helper()
	store := repocachemongo.New(db)
	if err := store.EnsureIndexes(context.Background()); err != nil {
		t.Fatalf("EnsureIndexes: %v", err)
	}
	return store
}

func TestMongoRepositoryConformance(t *testing.T) {
	repotest.Run(t, func(t *testing.T) repocache.Repository {
		return newStore(t, testDB(t))
	})
}

func TestMongoSnapshotPersistsAcrossStoreInstances(t *testing.T) {
	db := testDB(t)
	first := newStore(t, db)
	ctx := context.Background()
	metadata := repocache.SnapshotMetadata{BindingKey: "binding", CommitSHA: "sha"}
	if err := first.PutSnapshot(ctx, "ws-a", metadata, []*agentsv1.RepoCacheEntry{
		{Path: "prompt.md", Kind: agentsv1.RepoCacheEntryKind_REPO_CACHE_ENTRY_KIND_FILE},
	}, []repocache.CachedBlob{{Path: "prompt.md", Content: []byte("persisted")}}); err != nil {
		t.Fatalf("PutSnapshot: %v", err)
	}

	second := repocachemongo.New(db)
	gotMetadata, err := second.GetMetadata(ctx, "ws-a")
	if err != nil || gotMetadata.BindingKey != metadata.BindingKey || gotMetadata.CommitSHA != metadata.CommitSHA || gotMetadata.SnapshotID == "" {
		t.Fatalf("GetMetadata from second store = %#v, %v", gotMetadata, err)
	}
	content, err := second.GetBlob(ctx, "ws-a", gotMetadata.SnapshotID, "prompt.md")
	if err != nil || string(content) != "persisted" {
		t.Fatalf("GetBlob from second store = %q, %v", content, err)
	}
}

func TestMongoBlobContentIsChunkedBelowBSONLimit(t *testing.T) {
	db := testDB(t)
	store := newStore(t, db)
	ctx := context.Background()
	content := bytes.Repeat([]byte("x"), 17*1024*1024)
	if err := store.PutSnapshot(ctx, "ws-a", repocache.SnapshotMetadata{
		BindingKey: "binding", CommitSHA: "large",
	}, []*agentsv1.RepoCacheEntry{
		{Path: "large.md", Kind: agentsv1.RepoCacheEntryKind_REPO_CACHE_ENTRY_KIND_FILE, Size: int64(len(content))},
	}, []repocache.CachedBlob{{Path: "large.md", Content: content}}); err != nil {
		t.Fatalf("PutSnapshot large content: %v", err)
	}
	metadata, err := store.GetMetadata(ctx, "ws-a")
	if err != nil {
		t.Fatalf("GetMetadata: %v", err)
	}
	got, err := store.GetBlob(ctx, "ws-a", metadata.SnapshotID, "large.md")
	if err != nil || !bytes.Equal(got, content) {
		t.Fatalf("GetBlob large content: len=%d err=%v", len(got), err)
	}
	count, err := db.Collection("workspace_repo_cache_blob_chunks").CountDocuments(ctx, bson.M{"workspace_id": "ws-a"})
	if err != nil || count < 17 {
		t.Fatalf("blob chunks = %d, %v", count, err)
	}
}
