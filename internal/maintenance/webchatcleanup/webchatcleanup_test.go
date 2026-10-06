package webchatcleanup_test

import (
	"bytes"
	"maps"
	"slices"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"go.orx.me/apps/butter/internal/maintenance/webchatcleanup"
	"go.orx.me/apps/butter/internal/maintenance/webchatcleanup/webchatcleanuptest"
)

// The four collections the cleanup reads and deletes from.
var collections = []string{
	webchatcleanup.SessionsCollection,
	webchatcleanup.EventsCollection,
	webchatcleanup.InvocationsCollection,
	webchatcleanup.InputPartsCollection,
}

// contents reads every document of the four collections, in _id order.
func contents(t *testing.T, db *mongo.Database) map[string][]bson.Raw {
	t.Helper()
	out := map[string][]bson.Raw{}
	for _, coll := range collections {
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

func foundByWorkspace(f *webchatcleanup.Found) map[string]webchatcleanup.Counts {
	out := map[string]webchatcleanup.Counts{}
	for _, w := range f.Workspaces {
		out[w.Workspace] = w.Counts
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
	db := webchatcleanuptest.DB(t)
	webchatcleanuptest.Seed(t, db)
	before := contents(t, db)

	var out bytes.Buffer
	rep, err := webchatcleanup.Run(t.Context(), db, false, &out)
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	if got := foundByWorkspace(rep.Found); !maps.Equal(got, webchatcleanuptest.WantFound) {
		t.Errorf("found per workspace = %+v, want %+v", got, webchatcleanuptest.WantFound)
	}
	if got := rep.Found.Total(); got != webchatcleanuptest.WantTotal {
		t.Errorf("found in total = %+v, want %+v", got, webchatcleanuptest.WantTotal)
	}
	if rep.Found.OtherApps != 1 {
		t.Errorf("kept Invocations of other apps = %d, want 1 (inv-shared-agui)", rep.Found.OtherApps)
	}
	if rep.Deleted != (webchatcleanup.Counts{}) {
		t.Errorf("deleted = %+v on a dry run", rep.Deleted)
	}
	after := contents(t, db)
	for coll := range before {
		sameDocs(t, coll, after[coll], before[coll])
	}

	printed := out.String()
	for label, want := range map[string]string{
		"WORKSPACE":                "WORKSPACE adk_sessions adk_events invocations invocation_input_parts",
		"ws-a":                     "ws-a 3 4 4 4",
		"ws-b":                     "ws-b 2 4 2 2",
		webchatcleanup.NoWorkspace: "(none) 1 0 1 1",
		"total":                    "total 6 8 7 7",
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
	db := webchatcleanuptest.DB(t)
	webchatcleanuptest.Seed(t, db)
	before := contents(t, db)

	var out bytes.Buffer
	rep, err := webchatcleanup.Run(t.Context(), db, true, &out)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if rep.Deleted != webchatcleanuptest.WantTotal {
		t.Errorf("deleted = %+v, want %+v", rep.Deleted, webchatcleanuptest.WantTotal)
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
	webChatInvocation := webchatcleanuptest.WebChatInvocations
	isWebChat := map[string]func(bson.Raw) bool{
		webchatcleanup.SessionsCollection: func(d bson.Raw) bool {
			return d.Lookup("app_name").StringValue() == webchatcleanup.App
		},
		webchatcleanup.EventsCollection: func(d bson.Raw) bool {
			return d.Lookup("app_name").StringValue() == webchatcleanup.App
		},
		webchatcleanup.InvocationsCollection: func(d bson.Raw) bool {
			return webChatInvocation[d.Lookup("_id").StringValue()]
		},
		webchatcleanup.InputPartsCollection: func(d bson.Raw) bool {
			return webChatInvocation[d.Lookup("invocation_id").StringValue()]
		},
	}
	wantKept := map[string]int{
		webchatcleanup.SessionsCollection:    3,
		webchatcleanup.EventsCollection:      5,
		webchatcleanup.InvocationsCollection: 3,
		webchatcleanup.InputPartsCollection:  3,
	}
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
	rerun, err := webchatcleanup.Run(t.Context(), db, true, &out)
	if err != nil {
		t.Fatalf("second run: %v", err)
	}
	if got := rerun.Found.Total(); got != (webchatcleanup.Counts{}) {
		t.Errorf("second run found %+v, want nothing", got)
	}
	if rerun.Deleted != (webchatcleanup.Counts{}) {
		t.Errorf("second run deleted %+v, want nothing", rerun.Deleted)
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
