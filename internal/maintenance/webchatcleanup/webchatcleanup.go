// Package webchatcleanup deletes the data the old dashboard Chat left behind
// (#411). Chat runs on AG-UI since #409, and the old Chat's "web-chat"
// sessions are not kept: no migration and no read-only view.
//
// Run deletes, in this order: the input parts of the Invocations below; the
// Invocations whose session_id is one of the web-chat sessions', except any
// whose record names another app; the events of the web-chat sessions; the
// adk_sessions with app_name "web-chat". It works on those MongoDB
// collections by name and depends on none of the packages that write them.
// Every step is a DeleteMany by ID, so a run that stops partway is finished
// by running it again, and a run that finds nothing is a quick no-op.
// Workspace Memory (mem0) entries and every other app's data are not touched.
//
// Two callers share it: the one-off command cmd/butter-delete-web-chat, a dry
// run unless --confirm is given, and the service's startup cleanup
// (internal/app/webchat_cleanup.go), which deletes on every startup unless
// maintenance.delete_web_chat is false.
//
// TEMPORARY (#411): the startup cleanup must be removed in a later release,
// once it has run in production.
package webchatcleanup

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

// App is the ADK app name the old dashboard Chat kept its sessions under.
const App = "web-chat"

// The collections the cleanup reads and deletes from, named as the stores
// that write them name them: adk_sessions and adk_events in
// internal/runtime/session/mongo, invocations in
// internal/repo/invocation/mongo, invocation_input_parts in
// internal/repo/inputpart/mongo. The cleanup addresses them by name, so it
// does not depend on those packages.
const (
	SessionsCollection    = "adk_sessions"
	EventsCollection      = "adk_events"
	InvocationsCollection = "invocations"
	InputPartsCollection  = "invocation_input_parts"
)

// idBatch is how many IDs one $in filter carries, so no query nears
// MongoDB's 16 MiB document limit however much data there is.
const idBatch = 1000

// NoWorkspace labels the data recorded without a workspace.
const NoWorkspace = "(none)"

// Counts is web-chat data per collection.
type Counts struct {
	Sessions    int64
	Events      int64
	Invocations int64
	InputParts  int64
}

func (c *Counts) add(o Counts) {
	c.Sessions += o.Sessions
	c.Events += o.Events
	c.Invocations += o.Invocations
	c.InputParts += o.InputParts
}

// WorkspaceCounts is the web-chat data of one workspace.
type WorkspaceCounts struct {
	// Workspace is the workspace ID, or NoWorkspace for the data recorded
	// without one.
	Workspace string
	Counts
}

// Found is the web-chat data in the database: what a confirmed run deletes.
type Found struct {
	// Workspaces counts the data per workspace, in workspace ID order, with
	// the data recorded without a workspace last, as NoWorkspace.
	Workspaces []WorkspaceCounts
	// OtherApps counts the Invocations that another app recorded under a
	// web-chat session ID. They are kept.
	OtherApps int64
	// FirstStart and LastStart span the found Invocations' started_at.
	FirstStart, LastStart time.Time

	// sessionIDs are the distinct IDs of the web-chat sessions.
	sessionIDs []string
	// invocationIDs are the Invocations recorded under those session IDs.
	invocationIDs []string
}

// Total is the data of all workspaces together.
func (f *Found) Total() Counts {
	var t Counts
	for _, w := range f.Workspaces {
		t.add(w.Counts)
	}
	return t
}

// listWorkspaces orders the counts per workspace ID by ID, with the data
// recorded without a workspace (the key "") last, as NoWorkspace.
func listWorkspaces(byWorkspace map[string]*Counts) []WorkspaceCounts {
	out := make([]WorkspaceCounts, 0, len(byWorkspace))
	for _, id := range slices.Sorted(maps.Keys(byWorkspace)) {
		if id != "" {
			out = append(out, WorkspaceCounts{Workspace: id, Counts: *byWorkspace[id]})
		}
	}
	if c, ok := byWorkspace[""]; ok {
		out = append(out, WorkspaceCounts{Workspace: NoWorkspace, Counts: *c})
	}
	return out
}

func (f *Found) started(at time.Time) {
	if at.IsZero() {
		return
	}
	if f.FirstStart.IsZero() || at.Before(f.FirstStart) {
		f.FirstStart = at
	}
	if at.After(f.LastStart) {
		f.LastStart = at
	}
}

// find reads the web-chat data in db and changes nothing: the sessions with
// app_name "web-chat", their events, the Invocations whose session_id is one
// of theirs, and those Invocations' input parts.
func find(ctx context.Context, db *mongo.Database) (*Found, error) {
	f := &Found{}
	// byWorkspace counts the data per workspace ID. The key "" holds what
	// was recorded without a workspace.
	byWorkspace := map[string]*Counts{}
	workspace := func(id string) *Counts {
		c, ok := byWorkspace[id]
		if !ok {
			c = &Counts{}
			byWorkspace[id] = c
		}
		return c
	}

	// Several users may hold one session ID, and its events belong to the
	// ID, so they count toward the workspace of its oldest holder.
	cur, err := db.Collection(SessionsCollection).Find(ctx,
		bson.M{"app_name": App},
		options.Find().
			SetProjection(bson.M{"session_id": 1, "workspace_id": 1, "state.workspace_id": 1}).
			SetSort(bson.D{{Key: "_id", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("find %s: %w", SessionsCollection, err)
	}
	var sessions []struct {
		SessionID   string `bson:"session_id"`
		WorkspaceID string `bson:"workspace_id"`
		State       bson.M `bson:"state"`
	}
	if err := cur.All(ctx, &sessions); err != nil {
		return nil, fmt.Errorf("read %s: %w", SessionsCollection, err)
	}
	sessionWorkspace := map[string]string{}
	for _, s := range sessions {
		ws := s.WorkspaceID
		if ws == "" {
			// The old Chat also wrote the workspace into the session state,
			// which is all a session stored without workspace_id has.
			ws, _ = s.State["workspace_id"].(string)
		}
		workspace(ws).Sessions++
		if _, seen := sessionWorkspace[s.SessionID]; !seen {
			sessionWorkspace[s.SessionID] = ws
			f.sessionIDs = append(f.sessionIDs, s.SessionID)
		}
	}

	for chunk := range slices.Chunk(f.sessionIDs, idBatch) {
		perSession, err := countBy(ctx, db.Collection(EventsCollection),
			bson.M{"app_name": App, "session_id": bson.M{"$in": chunk}}, "$session_id")
		if err != nil {
			return nil, err
		}
		for id, n := range perSession {
			workspace(sessionWorkspace[id]).Events += n
		}
	}

	// Invocation records keep their app only inside spec, so they are found
	// by session ID. A session ID is unique only within its app, so a record
	// that names another app is that app's and is kept.
	invocationWorkspace := map[string]string{}
	for chunk := range slices.Chunk(f.sessionIDs, idBatch) {
		cur, err := db.Collection(InvocationsCollection).Find(ctx,
			bson.M{"session_id": bson.M{"$in": chunk}},
			options.Find().SetProjection(bson.M{"workspace_id": 1, "session_id": 1, "started_at": 1, "spec": 1}))
		if err != nil {
			return nil, fmt.Errorf("find %s: %w", InvocationsCollection, err)
		}
		var invocations []struct {
			ID          string    `bson:"_id"`
			WorkspaceID string    `bson:"workspace_id"`
			SessionID   string    `bson:"session_id"`
			StartedAt   time.Time `bson:"started_at"`
			Spec        string    `bson:"spec"`
		}
		if err := cur.All(ctx, &invocations); err != nil {
			return nil, fmt.Errorf("read %s: %w", InvocationsCollection, err)
		}
		for _, inv := range invocations {
			if app := recordedApp(inv.Spec); app != "" && app != App {
				f.OtherApps++
				continue
			}
			ws := inv.WorkspaceID
			if ws == "" {
				ws = sessionWorkspace[inv.SessionID]
			}
			invocationWorkspace[inv.ID] = ws
			f.invocationIDs = append(f.invocationIDs, inv.ID)
			workspace(ws).Invocations++
			f.started(inv.StartedAt)
		}
	}

	for chunk := range slices.Chunk(f.invocationIDs, idBatch) {
		perInvocation, err := countBy(ctx, db.Collection(InputPartsCollection),
			bson.M{"invocation_id": bson.M{"$in": chunk}}, "$invocation_id")
		if err != nil {
			return nil, err
		}
		for id, n := range perInvocation {
			workspace(invocationWorkspace[id]).InputParts += n
		}
	}
	f.Workspaces = listWorkspaces(byWorkspace)
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
func remove(ctx context.Context, db *mongo.Database, f *Found, out io.Writer) (Counts, error) {
	var deleted Counts
	steps := []struct {
		coll  string
		match bson.M
		field string
		ids   []string
		n     *int64
	}{
		{InputPartsCollection, nil, "invocation_id", f.invocationIDs, &deleted.InputParts},
		{InvocationsCollection, nil, "_id", f.invocationIDs, &deleted.Invocations},
		{EventsCollection, bson.M{"app_name": App}, "session_id", f.sessionIDs, &deleted.Events},
		{SessionsCollection, bson.M{"app_name": App}, "session_id", f.sessionIDs, &deleted.Sessions},
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

// Report is what a run found and, when confirmed, deleted.
type Report struct {
	Found   *Found
	Deleted Counts
}

// Run finds the web-chat data in db and writes the counts to out. Only when
// confirm is set does it delete them, and then it writes what it deleted.
// When deleting fails, Run returns the Report with what it deleted before
// the error.
func Run(ctx context.Context, db *mongo.Database, confirm bool, out io.Writer) (*Report, error) {
	f, err := find(ctx, db)
	if err != nil {
		return nil, err
	}
	printFound(out, f)
	rep := &Report{Found: f}
	if !confirm {
		_, _ = fmt.Fprint(out, `
DRY RUN: nothing was deleted. Deleting needs --confirm. On production, run
it only with the project owner's explicit confirmation, given after they
reviewed this dry run, and after a backup or snapshot of the database.
`)
		return rep, nil
	}
	_, _ = fmt.Fprintln(out, "\nDeleting (--confirm):")
	rep.Deleted, err = remove(ctx, db, f, out)
	if err != nil {
		return rep, err
	}
	_, _ = fmt.Fprintln(out, "Done. Running it again deletes nothing more.")
	return rep, nil
}

func printFound(out io.Writer, f *Found) {
	_, _ = fmt.Fprintf(out, "web-chat data, per workspace:\n\n")
	tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintf(tw, "WORKSPACE\t%s\t%s\t%s\t%s\n", SessionsCollection, EventsCollection, InvocationsCollection, InputPartsCollection)
	row := func(label string, c Counts) {
		_, _ = fmt.Fprintf(tw, "%s\t%d\t%d\t%d\t%d\n", label, c.Sessions, c.Events, c.Invocations, c.InputParts)
	}
	for _, w := range f.Workspaces {
		row(w.Workspace, w.Counts)
	}
	total := f.Total()
	row("total", total)
	_ = tw.Flush()

	_, _ = fmt.Fprintln(out)
	if f.OtherApps > 0 {
		_, _ = fmt.Fprintf(out, "Kept: %d Invocation(s) that another app recorded under a web-chat session ID.\n", f.OtherApps)
	}
	if total.Invocations > 0 {
		_, _ = fmt.Fprintf(out, "Dashboard metrics: deleting the %d Invocation(s) lowers past Activity counts, which count\nInvocations by start time. Their started_at spans %s to %s.\n",
			total.Invocations, formatTime(f.FirstStart), formatTime(f.LastStart))
	}
	_, _ = fmt.Fprintln(out, "Not touched: Workspace Memory (mem0) entries, and every other app's data.")
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}
	return t.UTC().Format(time.RFC3339)
}
