package pipeline

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	"go.orx.me/apps/butter/internal/runtime/runner"
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
	if len(tr.debugEdits) != 0 {
		t.Errorf("expected no debug edits, got %d", len(tr.debugEdits))
	}
	if len(tr.processingDebug) != 1 || tr.processingDebug[0] != nil {
		t.Errorf("expected processing message without debug summary, got %+v", tr.processingDebug)
	}
	if len(tr.editedMsgs) != 1 || tr.editedMsgs[0].debug != nil {
		t.Errorf("expected final message without debug summary, got %+v", tr.editedMsgs)
	}
}

// When debug is active, callbacks aggregate activity into edits of the same
// processing message and the final reply keeps the counts without latest-event
// detail.
func TestHandle_DebugActive_EditsProcessingAndFinalSummary(t *testing.T) {
	h, r, _, _, debug, tr := newHarness(Config{ChannelName: "tg", DefaultAgent: "assistant", DebugDefault: false})
	debug.override = boolPtr(true) // per-session override wins over channel default
	r.runResult = &runner.TurnResult{Output: "done"}
	r.runHook = func(call runCall) {
		evt := session.NewEvent(t.Context(), "inv-1")
		evt.Author = "router"
		evt.Actions.TransferToAgent = "assistant"
		evt.Content = &genai.Content{Parts: []*genai.Part{
			{FunctionCall: &genai.FunctionCall{Name: "search"}},
			{FunctionCall: &genai.FunctionCall{Name: "search"}},
			{FunctionCall: &genai.FunctionCall{Name: "fetch"}},
		}}
		call.onEvent(evt)
		call.onCompaction("assistant")
	}

	h.Handle(context.Background(), baseMsg())

	if len(r.runCalls) != 1 {
		t.Fatalf("expected 1 run, got %d", len(r.runCalls))
	}
	call := r.runCalls[0]
	if call.onEvent == nil || call.onCompaction == nil {
		t.Fatalf("expected non-nil debug callbacks when debug active")
	}

	if len(tr.processingDebug) != 1 || tr.processingDebug[0] == nil {
		t.Fatalf("expected initial debug summary, got %+v", tr.processingDebug)
	}
	initial := tr.processingDebug[0]
	if initial.ToolCalls != 0 || initial.Transfers != 0 || initial.Compactions != 0 {
		t.Errorf("initial summary = %+v, want zero counts", initial)
	}
	if len(tr.editedMsgs) != 1 || tr.editedMsgs[0].debug == nil {
		t.Fatalf("expected final debug summary, got %+v", tr.editedMsgs)
	}
	final := tr.editedMsgs[0].debug
	if final.ToolCalls != 3 || final.ToolCounts["search"] != 2 || final.ToolCounts["fetch"] != 1 {
		t.Errorf("tool summary = %+v", final)
	}
	if final.Transfers != 1 || final.Compactions != 1 {
		t.Errorf("final summary = %+v", final)
	}
	if final.LatestEvent != nil || final.LatestCompaction != "" {
		t.Errorf("final summary retained latest detail: %+v", final)
	}
	for _, edit := range tr.debugEdits {
		if edit.messageID != "proc-1" {
			t.Errorf("debug edited message %q, want proc-1", edit.messageID)
		}
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

func TestHandle_DebugActive_RunnerErrorKeepsSummary(t *testing.T) {
	h, r, _, _, debug, tr := newHarness(Config{ChannelName: "tg", DefaultAgent: "assistant"})
	debug.override = boolPtr(true)
	r.runErr = errors.New("boom")
	r.runHook = func(call runCall) {
		evt := session.NewEvent(t.Context(), "inv-1")
		evt.Content = &genai.Content{Parts: []*genai.Part{
			{FunctionCall: &genai.FunctionCall{Name: "search"}},
		}}
		call.onEvent(evt)
	}

	h.Handle(context.Background(), baseMsg())

	if len(tr.editedMsgs) != 1 || tr.editedMsgs[0].debug == nil {
		t.Fatalf("expected error reply with debug summary, got %+v", tr.editedMsgs)
	}
	if tr.editedMsgs[0].debug.ToolCalls != 1 {
		t.Errorf("tool calls = %d, want 1", tr.editedMsgs[0].debug.ToolCalls)
	}
}
