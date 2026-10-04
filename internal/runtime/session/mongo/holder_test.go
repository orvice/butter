package mongo_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/runtime/sessionshare"
)

func create(ctx context.Context, svc interface {
	Create(context.Context, *session.CreateRequest) (*session.CreateResponse, error)
}, appName, userID, sessionID string) (session.Session, error) {
	resp, err := svc.Create(ctx, &session.CreateRequest{AppName: appName, UserID: userID, SessionID: sessionID})
	if err != nil {
		return nil, err
	}
	return resp.Session, nil
}

// Events live under (app, session ID), so a second user creating a session
// under someone else's ID would read their conversation: it is refused, and
// the would-be intruder still has no session to read.
func TestCreate_RefusesASessionIDAnotherUserHolds(t *testing.T) {
	db := testDB(t)
	svc := newService(t, db)
	ctx := context.Background()

	alice, err := create(ctx, svc, "agui", "alice", "agui-t1")
	if err != nil {
		t.Fatalf("alice: %v", err)
	}
	if err := svc.AppendEvent(ctx, alice, textEvent("secret")); err != nil {
		t.Fatalf("AppendEvent: %v", err)
	}

	_, err = create(ctx, svc, "agui", "bob", "agui-t1")
	if !errors.Is(err, sessionshare.ErrIDTaken) {
		t.Fatalf("bob: err = %v, want ErrIDTaken", err)
	}
	if _, err := svc.Get(ctx, &session.GetRequest{AppName: "agui", UserID: "bob", SessionID: "agui-t1"}); err == nil {
		t.Fatal("bob has a session under alice's ID")
	}
	// The same ID in another app is a different session.
	if _, err := create(ctx, svc, "web-chat", "bob", "agui-t1"); err != nil {
		t.Fatalf("another app: %v", err)
	}
	// Recreating your own session is not "taken by another user".
	if _, err := create(ctx, svc, "agui", "alice", "agui-t1"); err == nil || errors.Is(err, sessionshare.ErrIDTaken) {
		t.Fatalf("alice again: err = %v, want a duplicate that is not ErrIDTaken", err)
	}
}

// A Telegram Destination session is held by every member of the chat: each
// member's own document joins the shared conversation.
func TestCreate_SharedHoldersShareOneConversation(t *testing.T) {
	db := testDB(t)
	svc := newService(t, db)
	shared := sessionshare.Allow(context.Background())

	first, err := create(shared, svc, "telegram", "111", "tg:c:d:d1:a")
	if err != nil {
		t.Fatalf("first member: %v", err)
	}
	if err := svc.AppendEvent(shared, first, textEvent("hello group")); err != nil {
		t.Fatalf("AppendEvent: %v", err)
	}
	if _, err := create(shared, svc, "telegram", "222", "tg:c:d:d1:a"); err != nil {
		t.Fatalf("second member: %v", err)
	}
	got := getSession(t, svc, "telegram", "222", "tg:c:d:d1:a")
	if got.Events().Len() != 1 {
		t.Fatalf("second member sees %d events, want the shared one", got.Events().Len())
	}

	// A caller-chosen ID still cannot join the shared conversation.
	if _, err := create(context.Background(), svc, "telegram", "intruder", "tg:c:d:d1:a"); !errors.Is(err, sessionshare.ErrIDTaken) {
		t.Fatalf("intruder: err = %v, want ErrIDTaken", err)
	}
}

// A session created exclusive cannot be joined later, not even by a shared
// path: claiming a predictable shared ID early must not open the
// conversation that later forms under it.
func TestCreate_SharedPathRefusesAnExclusiveHolder(t *testing.T) {
	db := testDB(t)
	svc := newService(t, db)

	if _, err := create(context.Background(), svc, "telegram", "intruder", "tg:c:d:d2:a"); err != nil {
		t.Fatalf("intruder: %v", err)
	}
	shared := sessionshare.Allow(context.Background())
	if _, err := create(shared, svc, "telegram", "111", "tg:c:d:d2:a"); !errors.Is(err, sessionshare.ErrIDTaken) {
		t.Fatalf("member: err = %v, want ErrIDTaken", err)
	}
}

// Documents written before holders were tracked carry no exclusive mark:
// shared paths keep joining them, everyone else is refused.
func TestCreate_UnmarkedHoldersFromBeforeTheMark(t *testing.T) {
	db := testDB(t)
	svc := newService(t, db)
	ctx := context.Background()
	if _, err := db.Collection("adk_sessions").InsertOne(ctx, bson.M{
		"app_name": "telegram", "user_id": "111", "session_id": "tg:c:d:d3:a", "state": bson.M{},
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}

	if _, err := create(sessionshare.Allow(ctx), svc, "telegram", "222", "tg:c:d:d3:a"); err != nil {
		t.Fatalf("member joining a legacy holder: %v", err)
	}
	if _, err := create(ctx, svc, "telegram", "intruder", "tg:c:d:d3:a"); !errors.Is(err, sessionshare.ErrIDTaken) {
		t.Fatalf("intruder: err = %v, want ErrIDTaken", err)
	}
}

// Users racing for one fresh ID: exactly one holds it afterwards, however
// the inserts interleave.
func TestCreate_RacingUsersLeaveOneHolder(t *testing.T) {
	db := testDB(t)
	svc := newService(t, db)
	const racers = 8
	var wg sync.WaitGroup
	errs := make([]error, racers)
	for i := range racers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, errs[i] = create(context.Background(), svc, "agui", fmt.Sprintf("user-%d", i), "agui-race")
		}()
	}
	wg.Wait()

	won := 0
	for i, err := range errs {
		switch {
		case err == nil:
			won++
		case !errors.Is(err, sessionshare.ErrIDTaken):
			t.Errorf("racer %d: err = %v, want ErrIDTaken", i, err)
		}
	}
	if won != 1 {
		t.Fatalf("%d racers hold the ID, want exactly one", won)
	}
}

// The race guard is a partial unique index over exclusive holders; a store
// that cannot build it only logs, so make sure a real MongoDB has it.
func TestNew_BuildsTheExclusiveHolderIndex(t *testing.T) {
	db := testDB(t)
	newService(t, db)
	cursor, err := db.Collection("adk_sessions").Indexes().List(context.Background())
	if err != nil {
		t.Fatalf("list indexes: %v", err)
	}
	var indexes []bson.M
	if err := cursor.All(context.Background(), &indexes); err != nil {
		t.Fatalf("decode indexes: %v", err)
	}
	for _, idx := range indexes {
		if idx["name"] == "app_name_session_id_exclusive" {
			if idx["unique"] != true || idx["partialFilterExpression"] == nil {
				t.Fatalf("guard index = %v, want unique and partial", idx)
			}
			return
		}
	}
	t.Fatalf("no exclusive holder index among %v", indexes)
}

func textEvent(text string) *session.Event {
	evt := session.NewEvent(context.Background(), "inv-1")
	evt.Author = "user"
	evt.Content = genai.NewContentFromText(text, genai.RoleUser)
	return evt
}
