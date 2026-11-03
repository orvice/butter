package mongostore_test

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	ws "go.orx.me/apps/butter/internal/workstation"
	"go.orx.me/apps/butter/internal/workstation/mongostore"
	"go.orx.me/apps/butter/internal/workstation/repotest"
	"os"
	"testing"
	"time"
)

func TestContract(t *testing.T) {
	uri := os.Getenv("BUTTER_TEST_MONGO_URI")
	if uri == "" {
		t.Skip("BUTTER_TEST_MONGO_URI not set")
	}
	client, err := mongo.Connect(options.Client().ApplyURI(uri).SetTimeout(10 * time.Second))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Disconnect(context.Background())
	repotest.Run(t, func(t *testing.T) ws.Repository {
		db := client.Database(fmt.Sprintf("butter_work_station_test_%d", time.Now().UnixNano()))
		t.Cleanup(func() { _ = db.Drop(context.Background()) })
		store := mongostore.New(db)
		if err := store.EnsureIndexes(context.Background()); err != nil {
			t.Fatal(err)
		}
		return store
	})
}

func TestCheckpointContract(t *testing.T) {
	uri := os.Getenv("BUTTER_TEST_MONGO_URI")
	if uri == "" {
		t.Skip("BUTTER_TEST_MONGO_URI not set")
	}
	client, err := mongo.Connect(options.Client().ApplyURI(uri).SetTimeout(10 * time.Second))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Disconnect(context.Background())
	db := client.Database(fmt.Sprintf("butter_work_station_checkpoint_test_%d", time.Now().UnixNano()))
	t.Cleanup(func() { _ = db.Drop(context.Background()) })
	store := mongostore.New(db)
	ctx := context.Background()
	first, err := store.CompareAndSwap(ctx, "github:issues", 0, ws.Checkpoint{Generation: "g1", Sequence: 1, Cursor: "page-2", ETag: `"v1"`})
	if err != nil || first.Revision != 1 {
		t.Fatalf("initial checkpoint: %+v %v", first, err)
	}
	if _, err := store.CompareAndSwap(ctx, "github:issues", 0, ws.Checkpoint{Generation: "g1", Sequence: 2}); !errors.Is(err, ws.ErrCheckpointConflict) {
		t.Fatalf("stale checkpoint writer accepted: %v", err)
	}
	if _, err := store.CompareAndSwap(ctx, "github:issues", 1, ws.Checkpoint{Generation: "g1", Sequence: 0}); !errors.Is(err, ws.ErrCheckpointRegression) {
		t.Fatalf("checkpoint regression accepted: %v", err)
	}
	loaded, ok, err := store.Load(ctx, "github:issues")
	if err != nil || !ok || loaded.Cursor != "page-2" || loaded.ETag != `"v1"` {
		t.Fatalf("loaded checkpoint: %+v %v %v", loaded, ok, err)
	}
	if err := store.Delete(ctx, "github:issues", 0); !errors.Is(err, ws.ErrCheckpointConflict) {
		t.Fatalf("stale checkpoint cleanup accepted: %v", err)
	}
	if err := store.Delete(ctx, "github:issues", 1); err != nil {
		t.Fatalf("checkpoint cleanup: %v", err)
	}
	if _, ok, err := store.Load(ctx, "github:issues"); err != nil || ok {
		t.Fatalf("checkpoint survived cleanup: ok=%v err=%v", ok, err)
	}
}
func TestReclamation(t *testing.T) {
	uri := os.Getenv("BUTTER_TEST_MONGO_URI")
	if uri == "" {
		t.Skip("BUTTER_TEST_MONGO_URI not set")
	}
	client, err := mongo.Connect(options.Client().ApplyURI(uri).SetTimeout(10 * time.Second))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Disconnect(context.Background())
	repotest.RunReclamation(t, func(t *testing.T) (ws.Repository, func() int64) {
		db := client.Database(fmt.Sprintf("butter_work_station_gc_test_%d", time.Now().UnixNano()))
		t.Cleanup(func() { _ = db.Drop(context.Background()) })
		store := mongostore.New(db)
		if err := store.EnsureIndexes(context.Background()); err != nil {
			t.Fatal(err)
		}
		return store, func() int64 {
			n, err := db.Collection("work_station_observations").CountDocuments(context.Background(), bson.M{})
			if err != nil {
				t.Fatal(err)
			}
			return n
		}
	})
}
