package plane

import (
	"context"
	"errors"
	"testing"
)

type diagRetry struct{}

func (diagRetry) Error() string   { return "redis unavailable token=abc" }
func (diagRetry) Retryable() bool { return true }
func TestDiagnosticsClassifyAndRedact(t *testing.T) {
	r := BuildDiagnostic(&Event{WorkspaceID: "ws", AppID: "app", InstallationID: "i", AgentRunID: "r", ActivityID: "a"}, "wf", diagRetry{})
	if r.Class != FailureRetryable || !r.Recoverable {
		t.Fatalf("record=%+v", r)
	}
	if r.Summary != "redis unavailable token=[redacted]" {
		t.Fatalf("summary=%q", r.Summary)
	}
	if SafeDiagnosticSummary(errors.New("raw_payload=secret authorization=Bearer-abc")) != "raw_payload=[redacted] authorization=[redacted]" {
		t.Fatal("unsafe summary")
	}
}
func TestDiagnosticsWorkspaceIsolation(t *testing.T) {
	s := NewMemoryDiagnosticStore()
	r := &DiagnosticRecord{Key: "k", WorkspaceID: "ws", AppID: "app"}
	if err := s.Put(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Get(context.Background(), "other", "k"); err == nil {
		t.Fatal("cross workspace read")
	}
	rows, _ := s.List(context.Background(), "ws", "app")
	if len(rows) != 1 {
		t.Fatal(len(rows))
	}
}
