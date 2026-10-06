package main

import (
	"context"
	"fmt"
	"io"
	"maps"
	"slices"
	"text/tabwriter"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"google.golang.org/protobuf/encoding/protojson"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

// webChatApp is the ADK app name the old dashboard Chat kept its sessions
// under.
const webChatApp = "web-chat"

// The collections the cleanup reads and deletes from, named as the stores
// that write them name them: adk_sessions and adk_events in
// internal/runtime/session/mongo, invocations in
// internal/repo/invocation/mongo, invocation_input_parts in
// internal/repo/inputpart/mongo. The command addresses them by name, so it
// does not depend on those packages.
const (
	sessionsColl    = "adk_sessions"
	eventsColl      = "adk_events"
	invocationsColl = "invocations"
	inputPartsColl  = "invocation_input_parts"
)

// idBatch is how many IDs one $in filter carries, so no query nears
// MongoDB's 16 MiB document limit however much data there is.
const idBatch = 1000

// noWorkspace labels the data recorded without a workspace.
const noWorkspace = "(none)"

// counts is web-chat data per collection.
type counts struct {
	Sessions    int64
	Events      int64
	Invocations int64
	InputParts  int64
}

func (c *counts) add(o counts) {
	c.Sessions += o.Sessions
	c.Events += o.Events
	c.Invocations += o.Invocations
	c.InputParts += o.InputParts
}

// found is the web-chat data in the database: what a confirmed run deletes.
type found struct {
	// sessionIDs are the distinct IDs of the web-chat sessions.
	sessionIDs []string
	// invocationIDs are the Invocations recorded under those session IDs.
	invocationIDs []string
	// byWorkspace counts the data per workspace ID. The key "" holds what
	// was recorded without a workspace.
	byWorkspace map[string]*counts
	// otherApps counts the Invocations that another app recorded under a
	// web-chat session ID. They are kept.
	otherApps int64
	// firstStart and lastStart span the found Invocations' started_at.
	firstStart, lastStart time.Time
}

func (f *found) workspace(id string) *counts {
	c, ok := f.byWorkspace[id]
	if !ok {
		c = &counts{}
		f.byWorkspace[id] = c
	}
	return c
}

func (f *found) total() counts {
	var t counts
	for _, c := range f.byWorkspace {
		t.add(*c)
	}
	return t
}

func (f *found) started(at time.Time) {
	if at.IsZero() {
		return
	}
	if f.firstStart.IsZero() || at.Before(f.firstStart) {
		f.firstStart = at
	}
	if at.After(f.lastStart) {
		f.lastStart = at
	}
}

// find reads the web-chat data in db and changes nothing: the sessions with
// app_name "web-chat", their events, the Invocations whose session_id is one
// of theirs, and those Invocations' input parts.
func find(ctx context.Context, db *mongo.Database) (*found, error) {
	f := &found{byWorkspace: map[string]*counts{}}

	// Several users may hold one session ID, and its events belong to the
	// ID, so they count toward the workspace of its oldest holder.
	cur, err := db.Collection(sessionsColl).Find(ctx,
		bson.M{"app_name": webChatApp},
		options.Find().
			SetProjection(bson.M{"session_id": 1, "workspace_id": 1, "state.workspace_id": 1}).
			SetSort(bson.D{{Key: "_id", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("find %s: %w", sessionsColl, err)
	}
	var sessions []struct {
		SessionID   string `bson:"session_id"`
		WorkspaceID string `bson:"workspace_id"`
		State       bson.M `bson:"state"`
	}
	if err := cur.All(ctx, &sessions); err != nil {
		return nil, fmt.Errorf("read %s: %w", sessionsColl, err)
	}
	sessionWorkspace := map[string]string{}
	for _, s := range sessions {
		ws := s.WorkspaceID
		if ws == "" {
			// The old Chat also wrote the workspace into the session state,
			// which is all a session stored without workspace_id has.
			ws, _ = s.State["workspace_id"].(string)
		}
		f.workspace(ws).Sessions++
		if _, seen := sessionWorkspace[s.SessionID]; !seen {
			sessionWorkspace[s.SessionID] = ws
			f.sessionIDs = append(f.sessionIDs, s.SessionID)
		}
	}

	for chunk := range slices.Chunk(f.sessionIDs, idBatch) {
		perSession, err := countBy(ctx, db.Collection(eventsColl),
			bson.M{"app_name": webChatApp, "session_id": bson.M{"$in": chunk}}, "$session_id")
		if err != nil {
			return nil, err
		}
		for id, n := range perSession {
			f.workspace(sessionWorkspace[id]).Events += n
		}
	}

	// Invocation records keep their app only inside spec, so they are found
	// by session ID. A session ID is unique only within its app, so a record
	// that names another app is that app's and is kept.
	invocationWorkspace := map[string]string{}
	for chunk := range slices.Chunk(f.sessionIDs, idBatch) {
		cur, err := db.Collection(invocationsColl).Find(ctx,
			bson.M{"session_id": bson.M{"$in": chunk}},
			options.Find().SetProjection(bson.M{"workspace_id": 1, "session_id": 1, "started_at": 1, "spec": 1}))
		if err != nil {
			return nil, fmt.Errorf("find %s: %w", invocationsColl, err)
		}
		var invocations []struct {
			ID          string    `bson:"_id"`
			WorkspaceID string    `bson:"workspace_id"`
			SessionID   string    `bson:"session_id"`
			StartedAt   time.Time `bson:"started_at"`
			Spec        string    `bson:"spec"`
		}
		if err := cur.All(ctx, &invocations); err != nil {
			return nil, fmt.Errorf("read %s: %w", invocationsColl, err)
		}
		for _, inv := range invocations {
			if app := recordedApp(inv.Spec); app != "" && app != webChatApp {
				f.otherApps++
				continue
			}
			ws := inv.WorkspaceID
			if ws == "" {
				ws = sessionWorkspace[inv.SessionID]
			}
			invocationWorkspace[inv.ID] = ws
			f.invocationIDs = append(f.invocationIDs, inv.ID)
			f.workspace(ws).Invocations++
			f.started(inv.StartedAt)
		}
	}

	for chunk := range slices.Chunk(f.invocationIDs, idBatch) {
		perInvocation, err := countBy(ctx, db.Collection(inputPartsColl),
			bson.M{"invocation_id": bson.M{"$in": chunk}}, "$invocation_id")
		if err != nil {
			return nil, err
		}
		for id, n := range perInvocation {
			f.workspace(invocationWorkspace[id]).InputParts += n
		}
	}
	return f, nil
}

// recordedApp is the app an Invocation record names in its spec, or "" when
// it names none or the spec cannot be read.
func recordedApp(spec string) string {
	var inv agentsv1.Invocation
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal([]byte(spec), &inv); err != nil {
		return ""
	}
	return inv.GetAppName()
}

// countBy counts the documents of coll that match filter, per value of the
// field path key ("$field").
func countBy(ctx context.Context, coll *mongo.Collection, filter bson.M, key string) (map[string]int64, error) {
	cur, err := coll.Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$group", Value: bson.M{"_id": key, "n": bson.M{"$sum": 1}}}},
	})
	if err != nil {
		return nil, fmt.Errorf("count %s: %w", coll.Name(), err)
	}
	var groups []struct {
		ID string `bson:"_id"`
		N  int64  `bson:"n"`
	}
	if err := cur.All(ctx, &groups); err != nil {
		return nil, fmt.Errorf("count %s: %w", coll.Name(), err)
	}
	out := make(map[string]int64, len(groups))
	for _, g := range groups {
		out[g.ID] = g.N
	}
	return out, nil
}

// remove deletes what find found, in this order: the input parts, the
// Invocations, the events, the sessions. The sessions, through which find
// reaches everything else, go last, and every step is a DeleteMany by ID, so
// a run that stops partway is finished by running it again.
func remove(ctx context.Context, db *mongo.Database, f *found, out io.Writer) (counts, error) {
	var deleted counts
	steps := []struct {
		coll  string
		match bson.M
		field string
		ids   []string
		n     *int64
	}{
		{inputPartsColl, nil, "invocation_id", f.invocationIDs, &deleted.InputParts},
		{invocationsColl, nil, "_id", f.invocationIDs, &deleted.Invocations},
		{eventsColl, bson.M{"app_name": webChatApp}, "session_id", f.sessionIDs, &deleted.Events},
		{sessionsColl, bson.M{"app_name": webChatApp}, "session_id", f.sessionIDs, &deleted.Sessions},
	}
	for i, step := range steps {
		for chunk := range slices.Chunk(step.ids, idBatch) {
			filter := bson.M{step.field: bson.M{"$in": chunk}}
			maps.Copy(filter, step.match)
			res, err := db.Collection(step.coll).DeleteMany(ctx, filter)
			if err != nil {
				_, _ = fmt.Fprintf(out, "%d. %s: deleted %d, then failed\n", i+1, step.coll, *step.n)
				return deleted, fmt.Errorf("delete from %s: %w", step.coll, err)
			}
			*step.n += res.DeletedCount
		}
		_, _ = fmt.Fprintf(out, "%d. %s: deleted %d\n", i+1, step.coll, *step.n)
	}
	return deleted, nil
}

// report is what a run found and, when confirmed, deleted.
type report struct {
	found   *found
	deleted counts
}

// run finds the web-chat data in db and writes the counts to out. Only when
// confirm is set does it delete them, and then it writes what it deleted.
func run(ctx context.Context, db *mongo.Database, confirm bool, out io.Writer) (*report, error) {
	f, err := find(ctx, db)
	if err != nil {
		return nil, err
	}
	printFound(out, f)
	rep := &report{found: f}
	if !confirm {
		_, _ = fmt.Fprint(out, `
DRY RUN: nothing was deleted. Deleting needs --confirm. On production, run
it only with the project owner's explicit confirmation, given after they
reviewed this dry run, and after a backup or snapshot of the database.
`)
		return rep, nil
	}
	_, _ = fmt.Fprintln(out, "\nDeleting (--confirm):")
	rep.deleted, err = remove(ctx, db, f, out)
	if err != nil {
		return rep, err
	}
	_, _ = fmt.Fprintln(out, "Done. Running it again deletes nothing more.")
	return rep, nil
}

func printFound(out io.Writer, f *found) {
	_, _ = fmt.Fprintf(out, "web-chat data, per workspace:\n\n")
	tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintf(tw, "WORKSPACE\t%s\t%s\t%s\t%s\n", sessionsColl, eventsColl, invocationsColl, inputPartsColl)
	row := func(label string, c counts) {
		_, _ = fmt.Fprintf(tw, "%s\t%d\t%d\t%d\t%d\n", label, c.Sessions, c.Events, c.Invocations, c.InputParts)
	}
	for _, id := range slices.Sorted(maps.Keys(f.byWorkspace)) {
		if id != "" {
			row(id, *f.byWorkspace[id])
		}
	}
	if c, ok := f.byWorkspace[""]; ok {
		row(noWorkspace, *c)
	}
	total := f.total()
	row("total", total)
	_ = tw.Flush()

	_, _ = fmt.Fprintln(out)
	if f.otherApps > 0 {
		_, _ = fmt.Fprintf(out, "Kept: %d Invocation(s) that another app recorded under a web-chat session ID.\n", f.otherApps)
	}
	if total.Invocations > 0 {
		_, _ = fmt.Fprintf(out, "Dashboard metrics: deleting the %d Invocation(s) lowers past Activity counts, which count\nInvocations by start time. Their started_at spans %s to %s.\n",
			total.Invocations, formatTime(f.firstStart), formatTime(f.lastStart))
	}
	_, _ = fmt.Fprintln(out, "Not touched: Workspace Memory (mem0) entries, and every other app's data.")
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}
	return t.UTC().Format(time.RFC3339)
}
