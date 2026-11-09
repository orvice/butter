package plane

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestRejectedWebhookReceiptIsScopedAndNeverWork(t *testing.T) {
	f := newReceiverFixture(t)
	store := NewMemoryDiagnosticStore()
	f.receiver.SetDiagnosticStore(store)
	body := []byte(`{"event":"unsupported_event","data":{"private":"prompt must not be stored"}}`)
	_, validation := f.deliverBody(t, body)
	if !errors.Is(validation, ErrIdentity) {
		t.Fatalf("expected authenticated rejection, got %v", validation)
	}
	receipt, err := f.receiver.RecordWebhookRejection(t.Context(), f.app, body, validation)
	if err != nil || receipt == "" {
		t.Fatal(err)
	}
	again, err := f.receiver.RecordWebhookRejection(t.Context(), f.app, body, validation)
	if err != nil || again != receipt {
		t.Fatal("retry did not retain its receipt key")
	}
	record, err := store.Get(t.Context(), "ws", receipt)
	if err != nil || record.AppID != "app" || record.Class != FailurePermanent || record.Recoverable || record.RunID != "" || record.InstallationID != "" {
		t.Fatalf("bad terminal receipt: %+v %v", record, err)
	}
	if record.ExpiresAt.Sub(record.CreatedAt) != 30*24*time.Hour {
		t.Fatal("receipt retention is not bounded")
	}
	if strings.Contains(record.Summary, "prompt must not be stored") || len(f.queue.events) != 0 {
		t.Fatal("rejection stored input or scheduled work")
	}
	if _, err := store.Get(t.Context(), "foreign", receipt); err == nil {
		t.Fatal("foreign workspace accessed receipt")
	}
	other, err := f.receiver.RecordWebhookRejection(t.Context(), &agentsv1.PlaneApp{Id: "app", WorkspaceId: "other-ws"}, body, validation)
	if err != nil || other == receipt {
		t.Fatal("receipt key ignored trusted workspace scope")
	}
	for _, forbidden := range []error{nil, ErrUnauthorized, context.DeadlineExceeded, errors.New("private infrastructure error")} {
		if _, err := f.receiver.RecordWebhookRejection(t.Context(), f.app, body, forbidden); err == nil {
			t.Fatal("unverified/transient failure recorded as a permanent receipt")
		}
	}
}

func TestDiagnosticHistoryIsBoundedWithoutDeletingExecutionRecords(t *testing.T) {
	store := NewMemoryDiagnosticStore()
	now := time.Now()
	for i := range 202 {
		if err := store.Put(t.Context(), &DiagnosticRecord{Key: fmt.Sprintf("row-%d", i), WorkspaceID: "ws", AppID: "app", CreatedAt: now.Add(time.Duration(i) * time.Minute)}); err != nil {
			t.Fatal(err)
		}
	}
	rows, err := store.List(t.Context(), "ws", "app")
	if err != nil || len(rows) != DiagnosticHistoryLimit || rows[0].Key != "row-201" || rows[len(rows)-1].Key != "row-2" {
		t.Fatal("history is not bounded newest-first")
	}
	if _, err := store.Get(t.Context(), "ws", "row-0"); err != nil {
		t.Fatal("list limit deleted old execution history")
	}
}

func TestRejectedWebhookRequiresDiagnosticPersistence(t *testing.T) {
	f := newReceiverFixture(t)
	if _, err := f.receiver.RecordWebhookRejection(t.Context(), f.app, []byte(`{}`), ErrIdentity); !errors.Is(err, ErrRuntimeDependency) {
		t.Fatalf("missing store acknowledged: %v", err)
	}
}
