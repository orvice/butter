package skill_test

import (
	"errors"
	"testing"

	skillrepo "go.orx.me/apps/butter/internal/repo/skill"
)

func TestMemoryContentStorePutGetRoundTrip(t *testing.T) {
	store := skillrepo.NewMemoryContentStore()

	key := skillrepo.ContentKey("ws-a", "pdf-report")
	if err := store.Put(t.Context(), key, "# PDF Report\n"); err != nil {
		t.Fatalf("Put: %v", err)
	}
	got, err := store.Get(t.Context(), key)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != "# PDF Report\n" {
		t.Fatalf("content did not round-trip, got %q", got)
	}
}

func TestMemoryContentStoreGetMissingIsNotFound(t *testing.T) {
	store := skillrepo.NewMemoryContentStore()

	_, err := store.Get(t.Context(), skillrepo.ContentKey("ws-a", "absent"))
	if !errors.Is(err, skillrepo.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestMemoryContentStoreDeleteRemovesKeys(t *testing.T) {
	store := skillrepo.NewMemoryContentStore()

	keyA := skillrepo.ContentKey("ws-a", "pdf-report")
	keyB := skillrepo.ContentKey("ws-a", "sql-helper")
	for _, key := range []string{keyA, keyB} {
		if err := store.Put(t.Context(), key, "body"); err != nil {
			t.Fatalf("Put %q: %v", key, err)
		}
	}
	if err := store.Delete(t.Context(), []string{keyA, keyB}); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	for _, key := range []string{keyA, keyB} {
		if _, err := store.Get(t.Context(), key); !errors.Is(err, skillrepo.ErrNotFound) {
			t.Fatalf("expected ErrNotFound for %q after delete, got %v", key, err)
		}
	}
}

func TestContentKeyShape(t *testing.T) {
	// ADR 0004 / issue #153: keys are <workspace_id>/<skill_name>/SKILL.md;
	// the configured key_prefix is applied by the S3 store, not the key.
	got := skillrepo.ContentKey("ws-a", "pdf-report")
	if got != "ws-a/pdf-report/SKILL.md" {
		t.Fatalf("unexpected content key %q", got)
	}
}
