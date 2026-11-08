package mongo

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.orx.me/apps/butter/internal/repo/planestate"
)

func TestStateRoundtripPreservesClientBindingAndConsumesOnce(t *testing.T) {
	uri := os.Getenv("BUTTER_TEST_MONGO_URI")
	if uri == "" {
		t.Skip("BUTTER_TEST_MONGO_URI is required")
	}
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatal(err)
	}
	db := client.Database("butter_test_plane_state_" + uuid.NewString())
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = db.Drop(ctx)
		_ = client.Disconnect(ctx)
	})
	s := New(db)
	if err := s.EnsureIndexes(t.Context()); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	entry := &planestate.Entry{
		State: "one-use-nonce", WorkspaceID: "ws", AppID: "app", ClientID: "client", UserID: "owner",
		RedirectURI: "https://backend.example/api/plane/oauth/callback", ReturnURL: "https://dashboard.example/plane-apps/app/edit",
		CreatedAt: now, ExpiresAt: now.Add(10 * time.Minute),
	}
	if err := s.Create(t.Context(), entry); err != nil {
		t.Fatal(err)
	}
	var doc bson.M
	if err := s.states.FindOne(t.Context(), bson.M{"_id": planestate.Hash(entry.State)}).Decode(&doc); err != nil {
		t.Fatal(err)
	}
	if _, exists := doc["state"]; exists {
		t.Fatal("state nonce stored in plaintext")
	}
	s = New(db)
	found, err := s.Consume(t.Context(), entry.State, now)
	if err != nil || found.ClientID != entry.ClientID || found.AppID != entry.AppID ||
		found.UserID != entry.UserID || found.ReturnURL != entry.ReturnURL || found.RedirectURI != entry.RedirectURI {
		t.Fatalf("state=%+v err=%v", found, err)
	}
	if _, err := s.Consume(t.Context(), entry.State, now); !errors.Is(err, planestate.ErrNotFound) {
		t.Fatalf("replayed state err=%v", err)
	}
	entry.State = "expired"
	entry.ExpiresAt = now.Add(-time.Second)
	if err := s.Create(t.Context(), entry); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Consume(t.Context(), entry.State, now); !errors.Is(err, planestate.ErrNotFound) {
		t.Fatalf("expired state err=%v", err)
	}
}
