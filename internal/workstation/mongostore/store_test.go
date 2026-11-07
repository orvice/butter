package mongostore_test

import (
	"context"
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

func TestMongoDurableBatchResume(t *testing.T) {
	uri := os.Getenv("BUTTER_TEST_MONGO_URI")
	if uri == "" {
		t.Skip("BUTTER_TEST_MONGO_URI not set")
	}
	ctx := context.Background()
	client, err := mongo.Connect(options.Client().ApplyURI(uri).SetTimeout(10 * time.Second))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Disconnect(ctx)
	db := client.Database(fmt.Sprintf("butter_batch_resume_%d", time.Now().UnixNano()))
	defer db.Drop(ctx)
	store := mongostore.New(db)
	if err := store.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	repotest.RunBatches(t, store, func(age time.Duration) {
		// Move due times and orphan ages, without substituting a fake clock for
		// Mongo's actual $$NOW publication fence or sleeping several minutes.
		records, err := store.List(ctx, "A")
		if err != nil {
			t.Fatal(err)
		}
		for _, r := range records {
			next := r
			next.Scopes = ws.CloneScopes(r.Scopes)
			for i := range next.Scopes {
				next.Scopes[i].Due = time.Now().Add(-time.Second)
			}
			next.Version++
			ok, err := store.CAS(ctx, r, &next, nil)
			if err != nil || !ok {
				t.Fatal("due injection", err)
			}
		}
		if _, err := db.Collection("work_station_observations").UpdateMany(ctx, bson.M{}, bson.M{"$set": bson.M{"observed": time.Now().Add(-age)}}); err != nil {
			t.Fatal(err)
		}
	})
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
