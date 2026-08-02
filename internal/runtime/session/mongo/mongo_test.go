package mongo_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"google.golang.org/adk/v2/session"

	mongosession "go.orx.me/apps/butter/internal/runtime/session/mongo"
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

	db := client.Database(fmt.Sprintf("butter_session_test_%d", time.Now().UnixNano()))
	t.Cleanup(func() { _ = db.Drop(context.Background()) })
	return db
}

func newService(t *testing.T, db *mongo.Database) *mongosession.Service {
	t.Helper()
	svc, err := mongosession.New(context.Background(), db)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return svc
}

func getSession(t *testing.T, svc *mongosession.Service, appName, userID, sessionID string) session.Session {
	t.Helper()
	resp, err := svc.Get(context.Background(), &session.GetRequest{
		AppName:   appName,
		UserID:    userID,
		SessionID: sessionID,
	})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	return resp.Session
}

func TestSetSessionTitle_PersistsWithoutTouchingLastUpdateTime(t *testing.T) {
	svc := newService(t, testDB(t))
	ctx := context.Background()

	created, err := svc.Create(ctx, &session.CreateRequest{
		AppName: "web", UserID: "u1", SessionID: "s1",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	sid := created.Session.ID()

	before := getSession(t, svc, "web", "u1", sid).LastUpdateTime()

	result, err := svc.SetSessionTitle(ctx, "web", "u1", sid, "My Chat")
	if err != nil {
		t.Fatalf("SetSessionTitle: %v", err)
	}
	if result.Title != "My Chat" {
		t.Fatalf("expected returned title %q, got %q", "My Chat", result.Title)
	}

	after := getSession(t, svc, "web", "u1", sid)
	titled, ok := after.(interface{ Title() string })
	if !ok {
		t.Fatal("session does not expose Title()")
	}
	if titled.Title() != "My Chat" {
		t.Fatalf("Get did not surface persisted title: got %q", titled.Title())
	}
	if !after.LastUpdateTime().Equal(before) {
		t.Fatalf("rename must not change last_update_time: before=%v after=%v", before, after.LastUpdateTime())
	}
}

func TestSetSessionTitle_ManualUpdateReplacesExistingTitle(t *testing.T) {
	svc := newService(t, testDB(t))
	ctx := context.Background()

	if _, err := svc.Create(ctx, &session.CreateRequest{AppName: "web", UserID: "u1", SessionID: "s1"}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := svc.SetSessionTitle(ctx, "web", "u1", "s1", "First"); err != nil {
		t.Fatalf("first SetSessionTitle: %v", err)
	}
	result, err := svc.SetSessionTitle(ctx, "web", "u1", "s1", "Second")
	if err != nil {
		t.Fatalf("second SetSessionTitle: %v", err)
	}
	if result.Title != "Second" {
		t.Fatalf("expected replaced title %q, got %q", "Second", result.Title)
	}
}

func TestSetSessionTitle_ListSurfacesTitle(t *testing.T) {
	svc := newService(t, testDB(t))
	ctx := context.Background()

	if _, err := svc.Create(ctx, &session.CreateRequest{AppName: "web", UserID: "u1", SessionID: "s1"}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := svc.SetSessionTitle(ctx, "web", "u1", "s1", "Listed"); err != nil {
		t.Fatalf("SetSessionTitle: %v", err)
	}

	resp, err := svc.List(ctx, &session.ListRequest{AppName: "web", UserID: "u1"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(resp.Sessions))
	}
	titled, ok := resp.Sessions[0].(interface{ Title() string })
	if !ok {
		t.Fatal("listed session does not expose Title()")
	}
	if titled.Title() != "Listed" {
		t.Fatalf("List did not surface persisted title: got %q", titled.Title())
	}
}

func TestSetSessionTitle_MissingSessionReturnsErrSessionNotFound(t *testing.T) {
	svc := newService(t, testDB(t))

	_, err := svc.SetSessionTitle(context.Background(), "web", "u1", "does-not-exist", "x")
	if !errors.Is(err, mongosession.ErrSessionNotFound) {
		t.Fatalf("expected ErrSessionNotFound, got %v", err)
	}
}
