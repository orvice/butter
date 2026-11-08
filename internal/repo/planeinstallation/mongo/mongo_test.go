package mongo

import (
	"context"
	"errors"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	cryptomemory "go.orx.me/apps/butter/internal/repo/cryptokey/memory"
	repo "go.orx.me/apps/butter/internal/repo/planeinstallation"
	"go.orx.me/apps/butter/internal/secretbox"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestInstallationMongoStorageAndRecovery(t *testing.T) {
	uri := os.Getenv("BUTTER_TEST_MONGO_URI")
	if uri == "" {
		t.Skip("BUTTER_TEST_MONGO_URI is required")
	}
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatal(err)
	}
	db := client.Database("butter_test_plane_install_" + uuid.NewString())
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = db.Drop(ctx)
		_ = client.Disconnect(ctx)
	})
	keyring := secretbox.NewKeyring(cryptomemory.New())
	s := New(db, keyring)
	if err := s.EnsureIndexes(t.Context()); err != nil {
		t.Fatal(err)
	}
	in := &agentsv1.PlaneInstallation{
		AppId: "app", AppInstallationId: "ext-1", PlaneWorkspaceId: "plane-ws", WorkspaceSlug: "team",
		AppBotUserId: "bot", Status: "installed", CredentialState: "valid",
	}
	row, err := s.Upsert(t.Context(), "ws", "app", in, repo.Tokens{AccessToken: "private-token", ExpiresAt: time.Now().Add(time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	// A fresh store must resolve the same row/token through every public lookup.
	s = New(db, keyring)
	list, err := s.List(t.Context(), "ws", "app")
	if err != nil || len(list) != 1 {
		t.Fatalf("List=%v err=%v", list, err)
	}
	if found, err := s.Get(t.Context(), "ws", row.GetId()); err != nil || found.GetAppInstallationId() != "ext-1" {
		t.Fatalf("Get=%v err=%v", found, err)
	}
	if _, err := s.GetScoped(t.Context(), "ws", "app", "ext-1"); err != nil {
		t.Fatal(err)
	}
	for _, scope := range [][3]string{{"other", "app", "ext-1"}, {"ws", "other", "ext-1"}, {"ws", "app", "other"}} {
		if _, err := s.GetScoped(t.Context(), scope[0], scope[1], scope[2]); !errors.Is(err, repo.ErrNotFound) {
			t.Fatalf("cross-scope lookup err=%v", err)
		}
	}
	material, err := s.GetRuntimeToken(t.Context(), "ws", "app", "ext-1")
	if err != nil || material.AccessToken != "private-token" || material.Revision != 1 {
		t.Fatalf("runtime token err=%v revision=%d", err, material.Revision)
	}
	var stored doc
	if err := s.c.FindOne(t.Context(), bson.M{"id": row.GetId()}).Decode(&stored); err != nil {
		t.Fatal(err)
	}
	if stored.Token == "private-token" || stored.Token == "" || stored.TokenKey == "" {
		t.Fatal("installation token was not encrypted")
	}
	// Concurrent renewals have one winner, then reinstallation increments that
	// same fence so an older renewal cannot overwrite the new installation.
	var wins atomic.Int32
	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := s.CASRuntimeToken(t.Context(), "ws", "app", "ext-1", material.Revision,
				repo.Tokens{AccessToken: "renewed", ExpiresAt: time.Now().Add(time.Hour)}); err == nil {
				wins.Add(1)
			} else if !errors.Is(err, repo.ErrTokenCAS) {
				t.Errorf("CAS: %v", err)
			}
		}()
	}
	wg.Wait()
	if wins.Load() != 1 {
		t.Fatalf("renewal winners=%d", wins.Load())
	}
	before, _ := s.GetRuntimeToken(t.Context(), "ws", "app", "ext-1")
	in.AppInstallationId = "ext-2"
	reinstalled, err := s.Upsert(t.Context(), "ws", "app", in, repo.Tokens{AccessToken: "reinstalled", ExpiresAt: time.Now().Add(time.Hour)})
	if err != nil || reinstalled.GetId() != row.GetId() {
		t.Fatalf("reinstall=%v err=%v", reinstalled, err)
	}
	if _, err := s.GetScoped(t.Context(), "ws", "app", "ext-1"); !errors.Is(err, repo.ErrNotFound) {
		t.Fatalf("old installation remained live: %v", err)
	}
	if _, err := s.CASRuntimeToken(t.Context(), "ws", "app", "ext-2", before.Revision, repo.Tokens{AccessToken: "stale"}); !errors.Is(err, repo.ErrTokenCAS) {
		t.Fatalf("late renewal overwrote reinstall: %v", err)
	}
	in.AppId = "other-app"
	if _, err := s.Upsert(t.Context(), "other-ws", "other-app", in, repo.Tokens{AccessToken: "wrong-owner"}); !errors.Is(err, repo.ErrConflict) {
		t.Fatalf("cross-App adoption err=%v", err)
	}
}
