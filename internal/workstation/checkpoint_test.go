package workstation

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestMemoryCheckpointCASAndSequenceContract(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryCheckpointStore()
	when := time.Date(2026, 11, 3, 0, 0, 0, 0, time.UTC)
	first, err := s.CompareAndSwap(ctx, "github:issues", 0, Checkpoint{Generation: "g1", Sequence: 1, Cursor: "page-2", ETag: `"v1"`, UpdatedAt: when})
	if err != nil || first.Revision != 1 {
		t.Fatalf("initial checkpoint: %+v %v", first, err)
	}
	if _, err := s.CompareAndSwap(ctx, "github:issues", 0, Checkpoint{Generation: "g1", Sequence: 2}); !errors.Is(err, ErrCheckpointConflict) {
		t.Fatalf("stale writer was accepted: %v", err)
	}
	if _, err := s.CompareAndSwap(ctx, "github:issues", 1, Checkpoint{Generation: "g1", Sequence: 0, Cursor: "page-0"}); !errors.Is(err, ErrCheckpointRegression) {
		t.Fatalf("sequence regression was accepted: %v", err)
	}
	reset, err := s.CompareAndSwap(ctx, "github:issues", 1, Checkpoint{Generation: "g2", Sequence: 0, Cursor: "page-1"})
	if err != nil || reset.Revision != 2 || reset.Generation != "g2" {
		t.Fatalf("generation reset: %+v %v", reset, err)
	}
	loaded, ok, err := s.Load(ctx, "github:issues")
	if err != nil || !ok || loaded.Cursor != "page-1" || loaded.ETag != "" {
		t.Fatalf("loaded checkpoint: %+v %v %v", loaded, ok, err)
	}
}

func TestMemoryCheckpointHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	store := NewMemoryCheckpointStore()
	if _, _, err := store.Load(ctx, "key"); !errors.Is(err, context.Canceled) {
		t.Fatalf("load cancellation: %v", err)
	}
	if _, err := store.CompareAndSwap(ctx, "key", 0, Checkpoint{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cas cancellation: %v", err)
	}
}
