// Package webchatcleanuptest seeds a throwaway MongoDB database with web-chat
// data and other apps' data, for the tests of internal/maintenance/webchatcleanup
// and of the command cmd/butter-delete-web-chat.
package webchatcleanuptest

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.orx.me/apps/butter/internal/maintenance/webchatcleanup"
	invocationmongo "go.orx.me/apps/butter/internal/repo/invocation/mongo"
	mongosession "go.orx.me/apps/butter/internal/runtime/session/mongo"
	"go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// URIEnv names the MongoDB the tests run against.
const URIEnv = "BUTTER_TEST_MONGO_URI"

// DB connects to the MongoDB named by BUTTER_TEST_MONGO_URI and hands the
// test a throwaway database. Without the env var the test is skipped, so the
// default test run needs no infrastructure.
func DB(t *testing.T) *mongo.Database {
	t.Helper()
	uri := os.Getenv(URIEnv)
	if uri == "" {
		t.Skip(URIEnv + " not set; skipping mongo integration test")
	}
	client, err := mongo.Connect(options.Client().ApplyURI(uri).SetTimeout(10 * time.Second))
	if err != nil {
		t.Fatalf("connect mongo: %v", err)
	}
	t.Cleanup(func() { _ = client.Disconnect(context.Background()) })

	db := client.Database(fmt.Sprintf("butter_delete_web_chat_test_%d", time.Now().UnixNano()))
	t.Cleanup(func() { _ = db.Drop(context.Background()) })
	return db
}

// seedStart is when the first seeded web-chat Invocation started; the others
// start an hour apart.
var seedStart = time.Date(2026, 5, 1, 9, 0, 0, 0, time.UTC)

// WebChatInvocations are the seeded Invocations a cleanup deletes.
var WebChatInvocations = map[string]bool{
	"inv-a1": true, "inv-a2": true, "inv-a3": true, "inv-shared-web": true,
	"inv-b1": true, "inv-legacy": true, "inv-nows": true,
}

// WantFound is what a run finds in the seeded database, per workspace, keyed
// as Found.Workspaces labels them.
var WantFound = map[string]webchatcleanup.Counts{
	"ws-a": {Sessions: 3, Events: 4, Invocations: 4, InputParts: 4},
	// chat-legacy keeps ws-b in its state; inv-legacy takes its session's.
	"ws-b":                     {Sessions: 2, Events: 4, Invocations: 2, InputParts: 2},
	webchatcleanup.NoWorkspace: {Sessions: 1, Events: 0, Invocations: 1, InputParts: 1},
}

// WantTotal is what a run finds in the seeded database in total.
var WantTotal = webchatcleanup.Counts{Sessions: 6, Events: 8, Invocations: 7, InputParts: 7}

// Seed writes web-chat data and other apps' data through the stores that
// write them in Butter, and input parts as their store lays them out. The
// web-chat data spans two workspaces, a session that keeps its workspace
// only in its state, and a session without one. The other apps' data
// includes an AG-UI session that reuses a web-chat session ID, with an
// Invocation of its own.
func Seed(t *testing.T, db *mongo.Database) {
	t.Helper()
	ctx := t.Context()
	sessions, err := mongosession.New(ctx, db)
	if err != nil {
		t.Fatalf("session store: %v", err)
	}
	const webChat = webchatcleanup.App
	for _, s := range []struct {
		workspace, app, user, id string
		state                    map[string]any
		events                   int
	}{
		{"ws-a", webChat, "alice", "chat-a1", nil, 2},
		{"ws-a", webChat, "bob", "chat-a2", nil, 1},
		{"ws-a", webChat, "alice", "shared-1", nil, 1},
		{"ws-b", webChat, "carol", "chat-b1", nil, 3},
		{"", webChat, "dave", "chat-legacy", map[string]any{"workspace_id": "ws-b"}, 1},
		{"", webChat, "erin", "chat-nows", nil, 0},
		// Kept.
		{"ws-a", "agui", "alice", "agui-t1", nil, 2},
		{"ws-a", "agui", "alice", "shared-1", nil, 2},
		{"ws-b", "tg-bot", "frank", "tg:1:2:3:helper", nil, 1},
	} {
		createCtx := ctx
		if s.workspace != "" {
			createCtx = workspace.WithID(ctx, s.workspace)
		}
		created, err := sessions.Create(createCtx, &session.CreateRequest{
			AppName: s.app, UserID: s.user, SessionID: s.id, State: s.state,
		})
		if err != nil {
			t.Fatalf("create session %s/%s: %v", s.app, s.id, err)
		}
		for i := range s.events {
			evt := session.NewEvent(ctx, "turn-1")
			evt.Author = "user"
			evt.Content = genai.NewContentFromText(fmt.Sprintf("message %d", i), genai.RoleUser)
			if err := sessions.AppendEvent(ctx, created.Session, evt); err != nil {
				t.Fatalf("append event to %s/%s: %v", s.app, s.id, err)
			}
		}
	}

	invocations := invocationmongo.New(db)
	for i, inv := range []struct {
		id, workspace, app, session string
		started                     time.Time
		parts                       int
	}{
		{"inv-a1", "ws-a", webChat, "chat-a1", seedStart, 2},
		// A record that names no app is matched by its session alone.
		{"inv-a2", "ws-a", "", "chat-a1", seedStart.Add(1 * time.Hour), 0},
		{"inv-a3", "ws-a", webChat, "chat-a2", seedStart.Add(2 * time.Hour), 1},
		{"inv-shared-web", "ws-a", webChat, "shared-1", seedStart.Add(3 * time.Hour), 1},
		{"inv-b1", "ws-b", webChat, "chat-b1", seedStart.Add(4 * time.Hour), 1},
		{"inv-legacy", "", webChat, "chat-legacy", seedStart.Add(5 * time.Hour), 1},
		{"inv-nows", "", webChat, "chat-nows", seedStart.Add(6 * time.Hour), 1},
		// Kept, and outside the web-chat Invocations' started_at span.
		{"inv-agui", "ws-a", "agui", "agui-t1", seedStart.Add(-24 * time.Hour), 1},
		{"inv-shared-agui", "ws-a", "agui", "shared-1", seedStart.Add(1000 * time.Hour), 1},
		{"inv-tg", "ws-b", "tg-bot", "tg:1:2:3:helper", seedStart.Add(1000 * time.Hour), 1},
	} {
		if err := invocations.Save(ctx, &agentsv1.Invocation{
			Id:          inv.id,
			WorkspaceId: inv.workspace,
			AppName:     inv.app,
			UserId:      "user",
			SessionId:   inv.session,
			AgentName:   "helper",
			Status:      agentsv1.InvocationStatus_INVOCATION_STATUS_SUCCEEDED,
			Input:       fmt.Sprintf("turn %d", i),
			StartedAt:   timestamppb.New(inv.started),
		}); err != nil {
			t.Fatalf("save invocation %s: %v", inv.id, err)
		}
		for p := range inv.parts {
			if _, err := db.Collection(webchatcleanup.InputPartsCollection).InsertOne(ctx, bson.M{
				"_id":           fmt.Sprintf("%s:%d", inv.id, p),
				"invocation_id": inv.id,
				"index":         p,
				"data":          []byte("part"),
			}); err != nil {
				t.Fatalf("insert input part of %s: %v", inv.id, err)
			}
		}
	}
}
