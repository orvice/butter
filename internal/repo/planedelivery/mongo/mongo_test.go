package mongo

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	plane "go.orx.me/apps/butter/internal/runtime/plane"
)

func TestDeliveryStoreFencing(t *testing.T) {
	for _, backend := range []string{"memory", "mongo"} {
		t.Run(backend, func(t *testing.T) {
			var store plane.DeliveryStore = plane.NewMemoryDeliveryStore()
			if backend == "mongo" {
				uri := os.Getenv("BUTTER_TEST_MONGO_URI")
				if uri == "" {
					t.Skip("BUTTER_TEST_MONGO_URI is not set")
				}
				client, err := mongo.Connect(options.Client().ApplyURI(uri))
				if err != nil {
					t.Fatal(err)
				}
				db := client.Database("butter_test_plane_delivery")
				t.Cleanup(func() {
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()
					_ = db.Drop(ctx)
					_ = client.Disconnect(ctx)
				})
				if err := db.Drop(t.Context()); err != nil {
					t.Fatal(err)
				}
				s := New(db)
				if err := s.EnsureIndexes(t.Context()); err != nil {
					t.Fatal(err)
				}
				store = s
			}
			seed := &plane.DeliveryAttempt{Key: "reply", Status: plane.DeliveryRetryable}
			if err := store.Create(t.Context(), seed); err != nil {
				t.Fatal(err)
			}
			if err := store.Create(t.Context(), seed); !errors.Is(err, plane.ErrDeliveryAlreadyExists) {
				t.Fatalf("duplicate create: %v", err)
			}
			now := time.Now().UTC().Truncate(time.Millisecond)
			first, err := store.Claim(t.Context(), seed.Key, "first", now, time.Minute)
			if err != nil {
				t.Fatal(err)
			}
			if first.Attempts != 1 || first.ClaimedAt == nil || !first.ClaimedAt.Equal(now) {
				t.Fatalf("claim metadata: %+v", first)
			}
			// An expired pending POST is ambiguous, never an automatic retry.
			if _, err := store.Claim(t.Context(), seed.Key, "second", now.Add(2*time.Minute), time.Minute); !errors.Is(err, plane.ErrDeliveryInProgress) {
				t.Fatalf("expired pending claim: %v", err)
			}
			expired := *first
			expired.Status = plane.DeliveryUnknown
			expired.ClaimToken = ""
			expired.LeaseUntil = nil
			if err := store.CompareAndSwap(t.Context(), first, &expired); err != nil {
				t.Fatal(err)
			}
			late := *first
			late.Status = plane.DeliveryDelivered
			if err := store.PutOwned(t.Context(), &late, "first"); !errors.Is(err, plane.ErrDeliveryLeaseLost) {
				t.Fatalf("expired owner completion: %v", err)
			}
			retry := expired
			retry.Status = plane.DeliveryRetryable
			if err := store.CompareAndSwap(t.Context(), &expired, &retry); err != nil {
				t.Fatal(err)
			}
			second, err := store.Claim(t.Context(), seed.Key, "second", now.Add(3*time.Minute), time.Minute)
			if err != nil {
				t.Fatal(err)
			}
			if err := store.CompareAndSwap(t.Context(), first, &expired); !errors.Is(err, plane.ErrDeliveryInProgress) {
				t.Fatalf("stale transition: %v", err)
			}
			second.Status = plane.DeliveryDelivered
			second.ClaimToken = ""
			second.LeaseUntil = nil
			if err := store.PutOwned(t.Context(), second, "second"); err != nil {
				t.Fatal(err)
			}
			got, err := store.Get(t.Context(), seed.Key)
			if err != nil || got.Status != plane.DeliveryDelivered || got.Attempts != 2 {
				t.Fatalf("final record=%+v err=%v", got, err)
			}
		})
	}
}
