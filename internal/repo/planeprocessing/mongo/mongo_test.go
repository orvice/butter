package mongo

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	planeprocessing "go.orx.me/apps/butter/internal/repo/planeprocessing"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestRunClaimReturnsItsOwnAtomicFence(t *testing.T) {
	uri := os.Getenv("BUTTER_TEST_MONGO_URI")
	if uri == "" {
		t.Skip("BUTTER_TEST_MONGO_URI is required")
	}
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatal(err)
	}
	db := client.Database("butter_test_plane_claim_" + uuid.NewString())
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = db.Drop(ctx)
		_ = client.Disconnect(ctx)
	})
	store := New(db)
	if err := store.EnsureIndexes(t.Context()); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	type result struct {
		owner string
		claim planeprocessing.Claim
		err   error
	}
	results := make(chan result, 16)
	for range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			owner := uuid.NewString()
			claim, err := store.Claim(t.Context(), "ws", "app", "inst", "run", owner, time.Now().Add(time.Minute))
			results <- result{owner, claim, err}
		}()
	}
	wg.Wait()
	close(results)
	winners := 0
	var first planeprocessing.Claim
	for result := range results {
		if result.err == nil {
			winners++
			first = result.claim
			if result.owner != first.Owner || first.Fence == 0 || first.Revision == 0 {
				t.Fatalf("wrong claim returned: %+v", result)
			}
		} else if !errors.Is(result.err, planeprocessing.ErrClaimConflict) {
			t.Fatalf("unexpected contention error: %v", result.err)
		}
	}
	if winners != 1 {
		t.Fatalf("claim winners=%d", winners)
	}
	key := "ws:app:inst:run"
	if _, err := store.Renew(t.Context(), key, first, time.Now().Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	if err := store.Complete(t.Context(), key, first); err != nil {
		t.Fatal(err)
	}
	next, err := store.Claim(t.Context(), "ws", "app", "inst", "run", "next", time.Now().Add(time.Minute))
	if err != nil || next.Owner != "next" || next.Fence <= first.Fence || next.Revision <= first.Revision {
		t.Fatalf("next=%+v err=%v", next, err)
	}
	if err := store.Complete(t.Context(), key, first); !errors.Is(err, planeprocessing.ErrClaimExpired) {
		t.Fatalf("late completion=%v", err)
	}
	// Lease documents must not break the Dashboard processing-record list.
	_, err = store.Create(t.Context(), &agentsv1.PlaneProcessingRecord{Id: "record", WorkspaceId: "ws", AppId: "app",
		InstallationId: "inst", RunId: "run", ActivityId: "activity"})
	if err != nil {
		t.Fatal(err)
	}
	records, err := store.List(t.Context(), planeprocessing.Filter{WorkspaceID: "ws"})
	if err != nil || len(records) != 1 || records[0].GetId() != "record" {
		t.Fatalf("processing records=%v err=%v", records, err)
	}
}
