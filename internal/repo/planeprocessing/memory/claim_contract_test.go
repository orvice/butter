package memory

import (
	"context"
	planeprocessing "go.orx.me/apps/butter/internal/repo/planeprocessing"
	"testing"
	"time"
)

func TestClaimOwnerFenceExpiryAndIsolation(t *testing.T) {
	s := New()
	now := time.Now()
	a, e := s.Claim(context.Background(), "w", "a", "i", "r", "one", now.Add(time.Minute))
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.Claim(context.Background(), "w", "a", "i", "r", "two", now.Add(time.Minute)); e != planeprocessing.ErrClaimConflict {
		t.Fatalf("owner conflict=%v", e)
	}
	if e = s.Complete(context.Background(), "w:a:i:r", planeprocessing.Claim{Owner: "one", Revision: a.Revision, Fence: a.Fence + 1, ExpiresAt: a.ExpiresAt}); e != planeprocessing.ErrClaimExpired {
		t.Fatalf("old fence=%v", e)
	}
	if _, e = s.Claim(context.Background(), "other", "a", "i", "r", "two", now.Add(time.Minute)); e != nil {
		t.Fatal(e)
	}
	expired, _ := s.Claim(context.Background(), "w", "a", "x", "r", "one", now.Add(-time.Second))
	b, e := s.ReclaimExpired(context.Background(), "w", "a", "x", "r", "two", now.Add(time.Minute))
	if e != nil || b.Fence <= expired.Fence {
		t.Fatalf("reclaim=%+v err=%v", b, e)
	}
}
func TestClaimRecoveryAndCompletionRace(t *testing.T) {
	s := New()
	c, _ := s.Claim(context.Background(), "w", "a", "i", "r", "one", time.Now().Add(-time.Second))
	if e := s.RecoverExpired(context.Background(), time.Now()); e != nil {
		t.Fatal(e)
	}
	if e := s.Complete(context.Background(), "w:a:i:r", c); e != planeprocessing.ErrClaimExpired {
		t.Fatalf("stale complete=%v", e)
	}
	if _, e := s.Claim(context.Background(), "w", "a", "i", "r", "two", time.Now().Add(time.Minute)); e != nil {
		t.Fatal(e)
	}
}
