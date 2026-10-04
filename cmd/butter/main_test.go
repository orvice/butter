package main

import (
	"context"
	"slices"
	"sync"
	"testing"
	"time"
)

// The first SIGTERM or SIGINT hands signals back to their default handling,
// runs the teardown, and only then exits.
func TestExitAfterTeardownRunsTheTeardownBeforeExiting(t *testing.T) {
	signals, signal := context.WithCancel(context.Background())
	var mu sync.Mutex
	var steps []string
	record := func(step string) {
		mu.Lock()
		defer mu.Unlock()
		steps = append(steps, step)
	}
	exited := make(chan int, 1)
	go exitAfterTeardown(signals,
		func() { record("stop signals") },
		func() error { record("teardown"); return nil },
		func(code int) { record("exit"); exited <- code })

	select {
	case <-exited:
		t.Fatal("exited before any signal")
	case <-time.After(20 * time.Millisecond):
	}
	signal()
	select {
	case code := <-exited:
		if code != 0 {
			t.Fatalf("exit code = %d, want 0", code)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("never exited after the signal")
	}
	mu.Lock()
	defer mu.Unlock()
	if want := []string{"stop signals", "teardown", "exit"}; !slices.Equal(steps, want) {
		t.Fatalf("steps = %v, want %v", steps, want)
	}
}
