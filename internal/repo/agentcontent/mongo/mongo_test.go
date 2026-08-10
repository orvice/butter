package mongo_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"go.orx.me/apps/butter/internal/agentcontent"
	agentcontentrepo "go.orx.me/apps/butter/internal/repo/agentcontent"
	agentcontentmongo "go.orx.me/apps/butter/internal/repo/agentcontent/mongo"
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
	db := client.Database(fmt.Sprintf("butter_agentcontent_test_%d", time.Now().UnixNano()))
	t.Cleanup(func() { _ = db.Drop(context.Background()) })
	return db
}

func newStore(t *testing.T, db *mongo.Database) *agentcontentmongo.Store {
	t.Helper()
	store := agentcontentmongo.New(db)
	if err := store.EnsureIndexes(context.Background()); err != nil {
		t.Fatalf("EnsureIndexes: %v", err)
	}
	return store
}

func TestSnapshotPersistsAcrossStoreInstances(t *testing.T) {
	db := testDB(t)
	first := newStore(t, db)
	ctx := context.Background()
	snapshot := agentcontent.Snapshot{
		CommitSHA: "abc123",
		Entries: map[string]agentcontent.AgentContent{
			"agent-1": {
				AgentID:           "agent-1",
				Description:       "Persisted description",
				Instruction:       "Persisted prompt",
				GlobalInstruction: "Persisted global prompt",
			},
		},
	}
	if err := first.PutSnapshot(ctx, "ws-a", snapshot); err != nil {
		t.Fatalf("PutSnapshot: %v", err)
	}

	second := agentcontentmongo.New(db)
	got, err := second.GetSnapshot(ctx, "ws-a", "abc123")
	if err != nil {
		t.Fatalf("GetSnapshot from second store: %v", err)
	}
	if got.Entries["agent-1"] != snapshot.Entries["agent-1"] {
		t.Fatalf("persisted entry = %#v, want %#v", got.Entries["agent-1"], snapshot.Entries["agent-1"])
	}
}

func TestSnapshotsAreRevisionAddressedAndPruned(t *testing.T) {
	store := newStore(t, testDB(t))
	ctx := context.Background()
	for _, commitSHA := range []string{"v1", "v2"} {
		if err := store.PutSnapshot(ctx, "ws-a", agentcontent.Snapshot{
			CommitSHA: commitSHA,
			Entries: map[string]agentcontent.AgentContent{
				"agent-1": {AgentID: "agent-1", Instruction: commitSHA},
			},
		}); err != nil {
			t.Fatalf("PutSnapshot %s: %v", commitSHA, err)
		}
	}
	if _, err := store.GetSnapshot(ctx, "ws-a", "v1"); err != nil {
		t.Fatalf("GetSnapshot v1: %v", err)
	}
	if err := store.PruneSnapshots(ctx, "ws-a", "v2"); err != nil {
		t.Fatalf("PruneSnapshots: %v", err)
	}
	if _, err := store.GetSnapshot(ctx, "ws-a", "v1"); err != agentcontentrepo.ErrNotFound {
		t.Fatalf("GetSnapshot v1 after prune: %v", err)
	}
	if got, err := store.GetSnapshot(ctx, "ws-a", "v2"); err != nil || got.Entries["agent-1"].Instruction != "v2" {
		t.Fatalf("GetSnapshot v2 after prune = %#v, %v", got, err)
	}
}
