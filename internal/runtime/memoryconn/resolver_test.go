package memoryconn

import (
	"errors"
	"testing"

	cryptokeymemory "go.orx.me/apps/butter/internal/repo/cryptokey/memory"
	memoryconfigrepo "go.orx.me/apps/butter/internal/repo/memoryconfig"
	memoryconfigmem "go.orx.me/apps/butter/internal/repo/memoryconfig/memory"
	"go.orx.me/apps/butter/internal/secretbox"
	agentsv1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

func TestResolverWithoutConfigIsNotConfigured(t *testing.T) {
	r := NewResolver(memoryconfigmem.New(), nil)
	if _, err := r.Resolve(t.Context(), "ws1"); !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("Resolve = %v, want ErrNotConfigured", err)
	}
	if _, err := r.Load(t.Context(), "ws1"); !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("Load = %v, want ErrNotConfigured", err)
	}
	var nilResolver *Resolver
	if _, err := nilResolver.Resolve(t.Context(), "ws1"); !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("nil Resolve = %v, want ErrNotConfigured", err)
	}
}

func TestResolverDecryptsKeyAndHonorsEnabled(t *testing.T) {
	ctx := t.Context()
	repo := memoryconfigmem.New()
	keyring := secretbox.NewKeyring(cryptokeymemory.New())
	ciphertext, keyID, err := keyring.Encrypt(ctx, []byte("m0sk_secret"))
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	cfg := &agentsv1.WorkspaceMemoryConfig{BaseUrl: "https://mem0.example.com", Enabled: true}
	if _, err := repo.Put(ctx, "ws1", cfg, &memoryconfigrepo.Credential{Ciphertext: ciphertext, KeyID: keyID}); err != nil {
		t.Fatalf("Put: %v", err)
	}

	r := NewResolver(repo, keyring)
	conn, err := r.Resolve(ctx, "ws1")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if conn.BaseURL != "https://mem0.example.com" || conn.APIKey != "m0sk_secret" || !conn.Enabled {
		t.Fatalf("conn = %+v", conn)
	}

	cfg.Enabled = false
	if _, err := repo.Put(ctx, "ws1", cfg, nil); err != nil {
		t.Fatalf("Put disabled: %v", err)
	}
	if _, err := r.Resolve(ctx, "ws1"); !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("Resolve disabled = %v, want ErrNotConfigured", err)
	}
	conn, err = r.Load(ctx, "ws1")
	if err != nil || conn.Enabled || conn.APIKey != "m0sk_secret" {
		t.Fatalf("Load disabled = %+v, %v", conn, err)
	}
}

func TestResolverWithoutKeyringFailsOnStoredKey(t *testing.T) {
	ctx := t.Context()
	repo := memoryconfigmem.New()
	cfg := &agentsv1.WorkspaceMemoryConfig{BaseUrl: "https://mem0.example.com", Enabled: true}
	if _, err := repo.Put(ctx, "ws1", cfg, &memoryconfigrepo.Credential{Ciphertext: "x", KeyID: "k"}); err != nil {
		t.Fatalf("Put: %v", err)
	}
	_, err := NewResolver(repo, nil).Resolve(ctx, "ws1")
	if err == nil || errors.Is(err, ErrNotConfigured) {
		t.Fatalf("Resolve = %v, want a decrypt failure", err)
	}
}
