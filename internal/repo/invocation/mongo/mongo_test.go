package mongo

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/repo/invocation"
	"go.orx.me/apps/butter/internal/repo/invocation/repotest"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// testDB connects to the MongoDB named by BUTTER_TEST_MONGO_URI and hands the
// test a throwaway database. Without the env var the test is skipped, so the
// default test run needs no infrastructure.
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

	db := client.Database(fmt.Sprintf("butter_invocation_test_%d", time.Now().UnixNano()))
	t.Cleanup(func() { _ = db.Drop(context.Background()) })
	return db
}

func testStore(t *testing.T) *Store {
	t.Helper()
	store := New(testDB(t))
	if err := store.EnsureIndexes(t.Context()); err != nil {
		t.Fatalf("EnsureIndexes: %v", err)
	}
	return store
}

func TestMongoRepositoryConformance(t *testing.T) {
	repotest.Run(t, func(t *testing.T) repotest.Open {
		store := testStore(t)
		return func(owner string) invocation.Repository { return store.WithOwner(owner) }
	})
}

// A sweep fails a record only as it read it. When the record's process turns
// out to be alive after all and saves its outcome between the sweep's read
// and its write, the outcome stands.
func TestMarkStaleRunningLeavesARecordSavedMeanwhile(t *testing.T) {
	store := testStore(t)
	owner := store.WithOwner("pod-a")
	inv := &agentsv1.Invocation{
		Id:          "a1",
		WorkspaceId: "ws-1",
		Status:      agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING,
		StartedAt:   timestamppb.Now(),
	}
	if err := owner.Save(t.Context(), inv); err != nil {
		t.Fatalf("Save: %v", err)
	}

	sweeper := store.WithOwner("pod-b")
	sweeper.beforeFail = func() {
		inv.Status = agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED
		inv.Output = "done"
		if err := owner.Save(context.Background(), inv); err != nil {
			t.Errorf("Save terminal: %v", err)
		}
	}
	n, err := sweeper.MarkStaleRunning(t.Context(), invocation.StaleSelection{
		LostOwners: []string{"pod-a"},
		Reason:     func(string) string { return "stale" },
	})
	if err != nil {
		t.Fatalf("MarkStaleRunning: %v", err)
	}
	if n != 0 {
		t.Fatalf("MarkStaleRunning failed %d records, want 0", n)
	}
	got, err := store.GetAcrossWorkspaces(t.Context(), "a1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.GetStatus() != agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED || got.GetOutput() != "done" {
		t.Fatalf("record = %v / %q, want the owner's SUCCEEDED outcome", got.GetStatus(), got.GetOutput())
	}
}

// The owner stamp is storage-only: it is a top-level field the API never
// sees, not part of the Invocation message.
func TestOwnerStampStaysOutOfTheInvocation(t *testing.T) {
	store := testStore(t)
	owner := store.WithOwner("pod-a")
	if err := owner.Save(t.Context(), &agentsv1.Invocation{
		Id:     "a1",
		Status: agentsv1.InvocationStatus_INVOCATION_STATUS_RUNNING,
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	var d doc
	if err := store.coll.FindOne(t.Context(), bson.M{"_id": "a1"}).Decode(&d); err != nil {
		t.Fatalf("FindOne: %v", err)
	}
	if d.Owner != "pod-a" {
		t.Fatalf("owner = %q, want pod-a", d.Owner)
	}
	if got := d.Fields.Spec; got == "" || strings.Contains(got, "pod-a") {
		t.Fatalf("spec = %q, want the Invocation without any owner", got)
	}
}
