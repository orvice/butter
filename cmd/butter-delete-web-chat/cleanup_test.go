package main

import (
	"bytes"
	"context"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"
	"google.golang.org/protobuf/types/known/timestamppb"

	invocationmongo "go.orx.me/apps/butter/internal/repo/invocation/mongo"
	mongosession "go.orx.me/apps/butter/internal/runtime/session/mongo"
	"go.orx.me/apps/butter/internal/workspace"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
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

	db := client.Database(fmt.Sprintf("butter_delete_web_chat_test_%d", time.Now().UnixNano()))
	t.Cleanup(func() { _ = db.Drop(context.Background()) })
	return db
}

// seedStart is when the first seeded web-chat Invocation started; the others
// start an hour apart.
var seedStart = time.Date(2026, 5, 1, 9, 0, 0, 0, time.UTC)

// webChatInvocations are the seeded Invocations a cleanup deletes.
var webChatInvocations = map[string]bool{
	"inv-a1": true, "inv-a2": true, "inv-a3": true, "inv-shared-web": true,
	"inv-b1": true, "inv-legacy": true, "inv-nows": true,
}

// seed writes web-chat data and other apps' data through the stores that
// write them in Butter, and input parts as their store lays them out. The
// web-chat data spans two workspaces, a session that keeps its workspace
// only in its state, and a session without one. The other apps' data
// includes an AG-UI session that reuses a web-chat session ID, with an
// Invocation of its own.
func seed(t *testing.T, db *mongo.Database) {
	t.Helper()
	ctx := t.Context()
	sessions, err := mongosession.New(ctx, db)
	if err != nil {
		t.Fatalf("session store: %v", err)
	}
	for _, s := range []struct {
		workspace, app, user, id string
		state                    map[string]any
		events                   int
	}{
		{"ws-a", webChatApp, "alice", "chat-a1", nil, 2},
		{"ws-a", webChatApp, "bob", "chat-a2", nil, 1},
		{"ws-a", webChatApp, "alice", "shared-1", nil, 1},
		{"ws-b", webChatApp, "carol", "chat-b1", nil, 3},
		{"", webChatApp, "dave", "chat-legacy", map[string]any{"workspace_id": "ws-b"}, 1},
		{"", webChatApp, "erin", "chat-nows", nil, 0},
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
		{"inv-a1", "ws-a", webChatApp, "chat-a1", seedStart, 2},
		// A record that names no app is matched by its session alone.
		{"inv-a2", "ws-a", "", "chat-a1", seedStart.Add(1 * time.Hour), 0},
		{"inv-a3", "ws-a", webChatApp, "chat-a2", seedStart.Add(2 * time.Hour), 1},
		{"inv-shared-web", "ws-a", webChatApp, "shared-1", seedStart.Add(3 * time.Hour), 1},
		{"inv-b1", "ws-b", webChatApp, "chat-b1", seedStart.Add(4 * time.Hour), 1},
		{"inv-legacy", "", webChatApp, "chat-legacy", seedStart.Add(5 * time.Hour), 1},
		{"inv-nows", "", webChatApp, "chat-nows", seedStart.Add(6 * time.Hour), 1},
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
			if _, err := db.Collection(inputPartsColl).InsertOne(ctx, bson.M{
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

// wantFound is what a run finds in the seeded database, per workspace.
var wantFound = map[string]counts{
	"ws-a": {Sessions: 3, Events: 4, Invocations: 4, InputParts: 4},
	// chat-legacy keeps ws-b in its state; inv-legacy takes its session's.
	"ws-b": {Sessions: 2, Events: 4, Invocations: 2, InputParts: 2},
	"":     {Sessions: 1, Events: 0, Invocations: 1, InputParts: 1},
}

var wantTotal = counts{Sessions: 6, Events: 8, Invocations: 7, InputParts: 7}

// contents reads every document of the four collections, in _id order.
func contents(t *testing.T, db *mongo.Database) map[string][]bson.Raw {
	t.Helper()
	out := map[string][]bson.Raw{}
	for _, coll := range []string{sessionsColl, eventsColl, invocationsColl, inputPartsColl} {
		cur, err := db.Collection(coll).Find(t.Context(), bson.M{}, options.Find().SetSort(bson.D{{Key: "_id", Value: 1}}))
		if err != nil {
			t.Fatalf("read %s: %v", coll, err)
		}
		var docs []bson.Raw
		if err := cur.All(t.Context(), &docs); err != nil {
			t.Fatalf("read %s: %v", coll, err)
		}
		out[coll] = docs
	}
	return out
}

// sameDocs reports a difference between two reads of a collection, comparing
// every field of every document.
func sameDocs(t *testing.T, coll string, got, want []bson.Raw) {
	t.Helper()
	asJSON := func(docs []bson.Raw) []string {
		out := make([]string, 0, len(docs))
		for _, d := range docs {
			js, err := bson.MarshalExtJSON(d, true, false)
			if err != nil {
				t.Fatalf("marshal %s document: %v", coll, err)
			}
			out = append(out, string(js))
		}
		return out
	}
	if g, w := asJSON(got), asJSON(want); !slices.Equal(g, w) {
		t.Errorf("%s:\n got %d documents %v\nwant %d documents %v", coll, len(g), g, len(w), w)
	}
}

func foundByWorkspace(f *found) map[string]counts {
	out := map[string]counts{}
	for id, c := range f.byWorkspace {
		out[id] = *c
	}
	return out
}

// tableRow returns the fields of the printed table row that starts with
// label.
func tableRow(t *testing.T, out, label string) string {
	t.Helper()
	for line := range strings.SplitSeq(out, "\n") {
		if fields := strings.Fields(line); len(fields) > 0 && fields[0] == label {
			return strings.Join(fields, " ")
		}
	}
	t.Fatalf("no %q row in:\n%s", label, out)
	return ""
}

func TestDryRunReportsTheCountsAndChangesNothing(t *testing.T) {
	db := testDB(t)
	seed(t, db)
	before := contents(t, db)

	var out bytes.Buffer
	rep, err := run(t.Context(), db, false, &out)
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	if got := foundByWorkspace(rep.found); !maps.Equal(got, wantFound) {
		t.Errorf("found per workspace = %+v, want %+v", got, wantFound)
	}
	if got := rep.found.total(); got != wantTotal {
		t.Errorf("found in total = %+v, want %+v", got, wantTotal)
	}
	if rep.found.otherApps != 1 {
		t.Errorf("kept Invocations of other apps = %d, want 1 (inv-shared-agui)", rep.found.otherApps)
	}
	if rep.deleted != (counts{}) {
		t.Errorf("deleted = %+v on a dry run", rep.deleted)
	}
	after := contents(t, db)
	for coll := range before {
		sameDocs(t, coll, after[coll], before[coll])
	}

	printed := out.String()
	for label, want := range map[string]string{
		"WORKSPACE": "WORKSPACE adk_sessions adk_events invocations invocation_input_parts",
		"ws-a":      "ws-a 3 4 4 4",
		"ws-b":      "ws-b 2 4 2 2",
		noWorkspace: "(none) 1 0 1 1",
		"total":     "total 6 8 7 7",
	} {
		if got := tableRow(t, printed, label); got != want {
			t.Errorf("printed row %q, want %q", got, want)
		}
	}
	for _, want := range []string{
		"DRY RUN: nothing was deleted",
		"Kept: 1 Invocation(s) that another app recorded under a web-chat session ID.",
		"Their started_at spans 2026-05-01T09:00:00Z to 2026-05-01T15:00:00Z.",
		"lowers past Activity counts",
		"Not touched: Workspace Memory (mem0) entries",
	} {
		if !strings.Contains(printed, want) {
			t.Errorf("output lacks %q:\n%s", want, printed)
		}
	}
	if strings.Contains(printed, ": deleted ") {
		t.Errorf("a dry run reports deletions:\n%s", printed)
	}
}

func TestConfirmDeletesOnlyTheWebChatDataAndARerunDeletesNothing(t *testing.T) {
	db := testDB(t)
	seed(t, db)
	before := contents(t, db)

	var out bytes.Buffer
	rep, err := run(t.Context(), db, true, &out)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if rep.deleted != wantTotal {
		t.Errorf("deleted = %+v, want %+v", rep.deleted, wantTotal)
	}
	for _, want := range []string{
		"1. invocation_input_parts: deleted 7",
		"2. invocations: deleted 7",
		"3. adk_events: deleted 8",
		"4. adk_sessions: deleted 6",
	} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("output lacks %q:\n%s", want, out.String())
		}
	}

	// Everything else stays exactly as it was.
	isWebChat := map[string]func(bson.Raw) bool{
		sessionsColl:    func(d bson.Raw) bool { return d.Lookup("app_name").StringValue() == webChatApp },
		eventsColl:      func(d bson.Raw) bool { return d.Lookup("app_name").StringValue() == webChatApp },
		invocationsColl: func(d bson.Raw) bool { return webChatInvocations[d.Lookup("_id").StringValue()] },
		inputPartsColl:  func(d bson.Raw) bool { return webChatInvocations[d.Lookup("invocation_id").StringValue()] },
	}
	wantKept := map[string]int{sessionsColl: 3, eventsColl: 5, invocationsColl: 3, inputPartsColl: 3}
	after := contents(t, db)
	for coll, docs := range before {
		var kept []bson.Raw
		for _, d := range docs {
			if !isWebChat[coll](d) {
				kept = append(kept, d)
			}
		}
		if len(kept) != wantKept[coll] {
			t.Fatalf("seed keeps %d %s documents, want %d", len(kept), coll, wantKept[coll])
		}
		sameDocs(t, coll, after[coll], kept)
	}

	out.Reset()
	rerun, err := run(t.Context(), db, true, &out)
	if err != nil {
		t.Fatalf("second run: %v", err)
	}
	if got := rerun.found.total(); got != (counts{}) {
		t.Errorf("second run found %+v, want nothing", got)
	}
	if rerun.deleted != (counts{}) {
		t.Errorf("second run deleted %+v, want nothing", rerun.deleted)
	}
	if got := tableRow(t, out.String(), "total"); got != "total 0 0 0 0" {
		t.Errorf("second run printed %q, want zero counts", got)
	}
	for _, want := range []string{
		"1. invocation_input_parts: deleted 0",
		"2. invocations: deleted 0",
		"3. adk_events: deleted 0",
		"4. adk_sessions: deleted 0",
	} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("second run output lacks %q:\n%s", want, out.String())
		}
	}
	again := contents(t, db)
	for coll := range after {
		sameDocs(t, coll, again[coll], after[coll])
	}
}

// The command itself: a dry run by default, and --confirm to delete.
func TestCommandIsADryRunUnlessConfirmed(t *testing.T) {
	db := testDB(t)
	seed(t, db)
	uri := os.Getenv("BUTTER_TEST_MONGO_URI")
	t.Setenv(envConfigPath, "")
	webChatSessions := func() int64 {
		t.Helper()
		n, err := db.Collection(sessionsColl).CountDocuments(t.Context(), bson.M{"app_name": webChatApp})
		if err != nil {
			t.Fatalf("count sessions: %v", err)
		}
		return n
	}

	var stdout, stderr bytes.Buffer
	if code := cli([]string{"--mongo-uri", uri, "--mongo-db", db.Name()}, &stdout, &stderr); code != 0 {
		t.Fatalf("dry run exit code %d, stderr:\n%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), fmt.Sprintf("DRY RUN on database %q", db.Name())) {
		t.Errorf("dry run does not name its database:\n%s", stdout.String())
	}
	if n := webChatSessions(); n != wantTotal.Sessions {
		t.Fatalf("after a dry run %d web-chat sessions remain, want %d", n, wantTotal.Sessions)
	}

	stdout.Reset()
	if code := cli([]string{"--mongo-uri", uri, "--mongo-db", db.Name(), "--confirm"}, &stdout, &stderr); code != 0 {
		t.Fatalf("confirmed run exit code %d, stderr:\n%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "4. adk_sessions: deleted 6") {
		t.Errorf("confirmed run output:\n%s", stdout.String())
	}
	if n := webChatSessions(); n != 0 {
		t.Fatalf("after --confirm %d web-chat sessions remain", n)
	}
}
