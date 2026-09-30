package mongo_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	linearrepo "go.orx.me/apps/butter/internal/repo/linear"
	linearmongo "go.orx.me/apps/butter/internal/repo/linear/mongo"
	"go.orx.me/apps/butter/internal/repo/linear/repotest"
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

	db := client.Database(fmt.Sprintf("butter_linear_test_%d", time.Now().UnixNano()))
	t.Cleanup(func() { _ = db.Drop(context.Background()) })
	return db
}

func TestMongoRepositoryConformance(t *testing.T) {
	repotest.Run(t, func(t *testing.T) linearrepo.Repository {
		store := linearmongo.New(testDB(t))
		if err := store.EnsureIndexes(context.Background()); err != nil {
			t.Fatalf("EnsureIndexes: %v", err)
		}
		return store
	})
}
