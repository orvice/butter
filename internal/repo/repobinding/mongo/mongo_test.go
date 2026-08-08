package mongo_test

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

	repobindingrepo "go.orx.me/apps/butter/internal/repo/repobinding"
	repobindingmongo "go.orx.me/apps/butter/internal/repo/repobinding/mongo"
	"go.orx.me/apps/butter/internal/repo/repobinding/repotest"
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

	db := client.Database(fmt.Sprintf("butter_repobinding_test_%d", time.Now().UnixNano()))
	t.Cleanup(func() { _ = db.Drop(context.Background()) })
	return db
}

func newStore(t *testing.T, db *mongo.Database) *repobindingmongo.Store {
	t.Helper()
	store := repobindingmongo.New(db)
	if err := store.EnsureIndexes(context.Background()); err != nil {
		t.Fatalf("EnsureIndexes: %v", err)
	}
	return store
}

func TestMongoRepositoryConformance(t *testing.T) {
	repotest.Run(t, func(t *testing.T) repobindingrepo.Repository {
		return newStore(t, testDB(t))
	})
}

// TestCredentialNeverInsideSpec proves the persisted public model (the
// protojson spec field) cannot contain PAT ciphertext: the credential lives
// only in its dedicated document field (issue #214 secret handling).
func TestCredentialNeverInsideSpec(t *testing.T) {
	db := testDB(t)
	store := newStore(t, db)
	ctx := context.Background()

	if _, err := store.Put(ctx, "ws-a", &agentsv1.WorkspaceRepoBinding{
		GitHostId: "gh-1", Repository: "acme/agents", Branch: "main",
	}); err != nil {
		t.Fatalf("Put: %v", err)
	}
	const ciphertext = "sealed-pat-ciphertext"
	if err := store.SetCredential(ctx, "ws-a", ciphertext); err != nil {
		t.Fatalf("SetCredential: %v", err)
	}
	// Re-Put to force a spec re-marshal after the credential exists.
	if _, err := store.Put(ctx, "ws-a", &agentsv1.WorkspaceRepoBinding{
		GitHostId: "gh-1", Repository: "acme/agents", Branch: "develop",
	}); err != nil {
		t.Fatalf("second Put: %v", err)
	}

	var doc struct {
		Spec       string `bson:"spec"`
		Credential string `bson:"credential"`
	}
	if err := db.Collection("workspace_repo_bindings").
		FindOne(ctx, bson.M{"_id": "ws-a"}).Decode(&doc); err != nil {
		t.Fatalf("read raw doc: %v", err)
	}
	if strings.Contains(doc.Spec, ciphertext) {
		t.Fatalf("spec contains credential ciphertext: %s", doc.Spec)
	}
	if doc.Credential != ciphertext {
		t.Fatalf("credential field = %q, want %q", doc.Credential, ciphertext)
	}
	got, err := store.GetCredential(ctx, "ws-a")
	if err != nil || got != ciphertext {
		t.Fatalf("credential lost across Put: %q, %v", got, err)
	}
}
