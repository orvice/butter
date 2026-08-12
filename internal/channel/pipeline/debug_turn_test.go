package pipeline

import (
	"context"
	"sync"
	"testing"
	"time"

	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"
)

func TestDebugTurn_ThrottlesAndKeepsLatestSnapshot(t *testing.T) {
	const interval = 30 * time.Millisecond

	var (
		mu    sync.Mutex
		edits []DebugSummary
	)
	d := newDebugTurn(context.Background(), interval, func(summary DebugSummary) {
		mu.Lock()
		defer mu.Unlock()
		edits = append(edits, summary)
	})

	d.observeEvent(toolEvent(t, "search"))
	waitFor(t, 100*time.Millisecond, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(edits) == 1
	})

	d.observeEvent(toolEvent(t, "fetch"))
	d.observeEvent(toolEvent(t, "fetch"))
	d.observeCompaction("assistant")

	time.Sleep(interval / 3)
	mu.Lock()
	if len(edits) != 1 {
		t.Fatalf("edits before throttle window = %d, want 1", len(edits))
	}
	mu.Unlock()

	waitFor(t, 100*time.Millisecond, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(edits) == 2
	})

	final := d.finish()
	if final.ToolCalls != 3 || final.ToolCounts["search"] != 1 || final.ToolCounts["fetch"] != 2 {
		t.Errorf("final tools = %+v", final)
	}
	if final.Compactions != 1 || final.LatestCompaction != "assistant" || final.LatestEvent != nil {
		t.Errorf("final latest state = %+v", final)
	}

	mu.Lock()
	second := edits[1]
	mu.Unlock()
	if second.ToolCalls != 3 || second.Compactions != 1 || second.LatestCompaction != "assistant" {
		t.Errorf("second edit = %+v", second)
	}
}

func TestDebugTurn_IgnoresEventsAfterFinish(t *testing.T) {
	d := newDebugTurn(context.Background(), time.Millisecond, func(DebugSummary) {})
	d.observeEvent(toolEvent(t, "search"))
	final := d.finish()
	d.observeEvent(toolEvent(t, "late"))

	if final.ToolCalls != 1 || final.ToolCounts["late"] != 0 {
		t.Errorf("final summary = %+v", final)
	}
}

func toolEvent(t *testing.T, name string) *session.Event {
	t.Helper()
	evt := session.NewEvent(t.Context(), "inv-1")
	evt.Content = &genai.Content{Parts: []*genai.Part{
		{FunctionCall: &genai.FunctionCall{Name: name}},
	}}
	return evt
}

func waitFor(t *testing.T, timeout time.Duration, ready func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if ready() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("condition was not satisfied before timeout")
}
