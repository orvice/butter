package http

import (
	"reflect"
	"testing"

	aguievents "github.com/ag-ui-protocol/ag-ui/sdks/community/go/pkg/core/events"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/runtime/streamorch"
)

func TestAGUIStateDiffOps(t *testing.T) {
	prev := map[string]any{"kept": 1, "changed": "old", "dropped": true}
	next := map[string]any{"kept": 1, "changed": "new", "added": []any{"x"}}

	got := aguiStateDiffOps(prev, next)
	want := []aguievents.JSONPatchOperation{
		{Op: "add", Path: "/added", Value: []any{"x"}},
		{Op: "replace", Path: "/changed", Value: "new"},
		{Op: "remove", Path: "/dropped"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("diff ops\n got: %+v\nwant: %+v", got, want)
	}

	if ops := aguiStateDiffOps(prev, prev); ops != nil {
		t.Fatalf("identical states produced ops: %+v", ops)
	}
}

// Numeric representation must not produce phantom diffs: bson decodes numbers
// as int32/int64 where JSON uses float64.
func TestAGUIStateDiffNormalizesRepresentations(t *testing.T) {
	prev := map[string]any{"count": int64(3)}
	next := map[string]any{"count": float64(3)}
	if ops := aguiStateDiffOps(prev, next); ops != nil {
		t.Fatalf("equivalent numbers diffed: %+v", ops)
	}
	if !aguiStatesEqual(prev, next) {
		t.Fatal("equivalent states compared unequal")
	}
}

func TestJSONPointerEscaping(t *testing.T) {
	if got := jsonPointerToken("a/b~c"); got != "/a~1b~0c" {
		t.Fatalf("pointer = %q", got)
	}
}

func TestAGUIVisibleStateFiltersScopes(t *testing.T) {
	got := aguiVisibleState(map[string]any{
		"draft":            "v1",
		"app:flag":         true,
		"user:theme":       "dark",
		"temp:cursor":      2,
		"temporary_public": "kept", // only the exact temp: prefix is hidden
	})
	want := map[string]any{"draft": "v1", "temporary_public": "kept"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("visible state = %+v", got)
	}
}

// A tool-written state delta streams as STATE_DELTA on the event that
// carries it, patching only the client-visible keys.
func TestAGUISink_InStreamStateDelta(t *testing.T) {
	rec := &aguiEventRecorder{}
	sink := newAGUISink("thread-1", "run-1", "msg-1", rec.emit)
	sink.setSharedState(map[string]any{"draft": "v1"}, false, nil)

	evt := &session.Event{ID: "e1", Author: "chat"}
	evt.Actions.StateDelta = map[string]any{"draft": "v2", "temp:cursor": 9}
	evt.Content = &genai.Content{Parts: []*genai.Part{{Text: "updated"}}}

	id := streamorch.RunIdentity{}
	if err := sink.Started(id); err != nil {
		t.Fatal(err)
	}
	if err := sink.RunEvent(id, evt); err != nil {
		t.Fatal(err)
	}
	if err := sink.Final(id, ""); err != nil {
		t.Fatal(err)
	}

	assertAGUISequence(t, rec.types, []string{
		"RUN_STARTED", "STATE_DELTA", "RUN_FINISHED",
	})
	delta := rec.events[1].(*aguievents.StateDeltaEvent)
	want := []aguievents.JSONPatchOperation{{Op: "replace", Path: "/draft", Value: "v2"}}
	if !reflect.DeepEqual(delta.Delta, want) {
		t.Fatalf("delta = %+v, want %+v", delta.Delta, want)
	}
}

// The initial snapshot is emitted immediately after RUN_STARTED so every
// later delta has a baseline.
func TestAGUISink_SnapshotFollowsRunStarted(t *testing.T) {
	rec := &aguiEventRecorder{}
	sink := newAGUISink("thread-1", "run-1", "msg-1", rec.emit)
	sink.setSharedState(map[string]any{"draft": "v1"}, true, nil)

	id := streamorch.RunIdentity{}
	if err := sink.Started(id); err != nil {
		t.Fatal(err)
	}
	if err := sink.Final(id, "hi"); err != nil {
		t.Fatal(err)
	}
	assertAGUISequence(t, rec.types, []string{
		"RUN_STARTED", "STATE_SNAPSHOT",
		"TEXT_MESSAGE_START", "TEXT_MESSAGE_CONTENT", "TEXT_MESSAGE_END",
		"RUN_FINISHED",
	})
	snap := rec.events[1].(*aguievents.StateSnapshotEvent)
	if !reflect.DeepEqual(snap.Snapshot, map[string]any{"draft": "v1"}) {
		t.Fatalf("snapshot = %+v", snap.Snapshot)
	}
}
