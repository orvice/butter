package plane

import (
	"context"
	"errors"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
	"testing"
)

func TestProcessorRecordsSafeDiagnosticsOnFailure(t *testing.T) {
	store := NewMemoryDiagnosticStore()
	p := NewProcessor(&fakeRunner{err: errors.New("authorization=secret")}, &fakePublisher{})
	p.SetDiagnosticStore(store)
	e := &Event{WorkspaceID: "ws", AppID: "app", InstallationID: "i", AgentRunID: "r", ActivityID: "a", Prompt: "x"}
	_, _ = p.Process(context.Background(), &agentsv1.PlaneApp{AgentId: "a"}, e)
	rows, err := store.List(context.Background(), "ws", "app")
	if err != nil || len(rows) != 1 {
		t.Fatalf("rows=%d err=%v", len(rows), err)
	}
	if rows[0].Summary != "authorization=[redacted]" {
		t.Fatalf("summary=%q", rows[0].Summary)
	}
}
