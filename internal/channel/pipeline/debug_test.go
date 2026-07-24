package pipeline

import (
	"context"
	"testing"

	"google.golang.org/adk/v2/session"
)

func boolPtr(b bool) *bool { return &b }

// When debug is inactive, the runner must receive nil callbacks so no debug
// traffic is generated.
func TestHandle_DebugInactive_NoCallbacks(t *testing.T) {
	h, r, _, _, debug, tr := newHarness(Config{ChannelName: "tg", DefaultAgent: "assistant", DebugDefault: false})
	debug.override = nil // no override → falls back to DebugDefault=false

	h.Handle(context.Background(), baseMsg())

	if len(r.runCalls) != 1 {
		t.Fatalf("expected 1 run, got %d", len(r.runCalls))
	}
	if r.runCalls[0].onEvent != nil || r.runCalls[0].onCompaction != nil {
		t.Errorf("expected nil debug callbacks when debug inactive")
	}
	if tr.debugEvents != 0 || len(tr.compactions) != 0 {
		t.Errorf("expected no debug traffic, got events=%d compactions=%v", tr.debugEvents, tr.compactions)
	}
}

// When debug is active (per-session override), the runner receives live
// callbacks that stream events and compaction notices to the transport.
func TestHandle_DebugActive_StreamsToTransport(t *testing.T) {
	h, r, _, _, debug, tr := newHarness(Config{ChannelName: "tg", DefaultAgent: "assistant", DebugDefault: false})
	debug.override = boolPtr(true) // per-session override wins over channel default

	h.Handle(context.Background(), baseMsg())

	if len(r.runCalls) != 1 {
		t.Fatalf("expected 1 run, got %d", len(r.runCalls))
	}
	call := r.runCalls[0]
	if call.onEvent == nil || call.onCompaction == nil {
		t.Fatalf("expected non-nil debug callbacks when debug active")
	}

	// Simulate the runner firing an event and a compaction mid-run.
	call.onEvent(&session.Event{})
	call.onCompaction("assistant")

	if tr.debugEvents != 1 {
		t.Errorf("debug events streamed = %d, want 1", tr.debugEvents)
	}
	if len(tr.compactions) != 1 || tr.compactions[0] != "assistant" {
		t.Errorf("compactions = %v, want [assistant]", tr.compactions)
	}
}

// The channel default enables debug when there is no per-session override.
func TestHandle_DebugDefaultOn_StreamsToTransport(t *testing.T) {
	h, r, _, _, debug, _ := newHarness(Config{ChannelName: "tg", DefaultAgent: "assistant", DebugDefault: true})
	debug.override = nil

	h.Handle(context.Background(), baseMsg())

	if r.runCalls[0].onEvent == nil {
		t.Errorf("expected debug callbacks when channel default debug is on")
	}
}
