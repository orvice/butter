package mongostore_test

import (
	"context"
	"fmt"
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
