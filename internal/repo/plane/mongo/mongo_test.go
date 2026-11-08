package mongo

import (
	"context"
	"os"
	"testing"
	"time"

	mongodriver "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	planerepo "go.orx.me/apps/butter/internal/repo/plane"
	"go.orx.me/apps/butter/internal/repo/plane/repotest"
)

// TestConformance runs the shared repository suite against a real MongoDB.
// It deliberately skips (rather than silently passing) when no test database
// is configured: uniqueness is enforced by a MongoDB index, so the memory
// suite alone cannot prove it.
func TestConformance(t *testing.T) {
	uri := os.Getenv("BUTTER_TEST_MONGO_URI")
	if uri == "" {
		t.Skip("BUTTER_TEST_MONGO_URI is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	client, err := mongodriver.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("connect mongo: %v", err)
	}
	defer client.Disconnect(ctx)
	db := client.Database("butter_test_plane")
	if err := db.Drop(ctx); err != nil {
		t.Fatalf("drop test database: %v", err)
	}
	repotest.Run(t, func(t *testing.T) planerepo.Repository {
		t.Helper()
		if err := db.Drop(t.Context()); err != nil {
			t.Fatalf("drop test database: %v", err)
		}
		store := New(db)
		if err := store.EnsureIndexes(t.Context()); err != nil {
			t.Fatalf("ensure indexes: %v", err)
		}
		return store
	})
}
