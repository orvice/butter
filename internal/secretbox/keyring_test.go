package secretbox

import (
	"errors"
	"sync"
	"testing"

	"go.orx.me/apps/butter/internal/repo/cryptokey"
	cryptokeymemory "go.orx.me/apps/butter/internal/repo/cryptokey/memory"
)

func TestKeyringRoundTripsUnderTheActiveKey(t *testing.T) {
	ring := NewKeyring(cryptokeymemory.New())

	ciphertext, keyID, err := ring.Encrypt(t.Context(), []byte("123456:bot-token"))
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if keyID == "" {
		t.Fatal("expected a key id to be recorded with the ciphertext")
	}
	if ciphertext == "123456:bot-token" {
		t.Fatal("ciphertext must not equal the plaintext")
	}

	plaintext, err := ring.Decrypt(t.Context(), ciphertext, keyID)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if string(plaintext) != "123456:bot-token" {
		t.Errorf("plaintext = %q", plaintext)
	}
}

// A ciphertext with no key id is a bookkeeping bug. Guessing the active key
// would turn it into an authentication failure that looks like a revoked
// credential, so it must fail loudly instead.
func TestKeyringRejectsCiphertextWithoutAKeyID(t *testing.T) {
	ring := NewKeyring(cryptokeymemory.New())
	ciphertext, _, err := ring.Encrypt(t.Context(), []byte("secret"))
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if _, err := ring.Decrypt(t.Context(), ciphertext, ""); err == nil {
		t.Fatal("expected decryption without a key id to fail")
	}
}

func TestNilKeyringFailsClosed(t *testing.T) {
	var ring *Keyring
	if _, _, err := ring.Encrypt(t.Context(), []byte("secret")); err == nil {
		t.Fatal("expected a nil keyring to refuse to encrypt")
	}
	if _, err := ring.Decrypt(t.Context(), "x", "k"); err == nil {
		t.Fatal("expected a nil keyring to refuse to decrypt")
	}
	if NewKeyring(nil) != nil {
		t.Fatal("expected NewKeyring(nil) to produce a nil keyring")
	}
}

// Two Pods starting at once must converge on one master key; otherwise each
// encrypts under its own and neither can read the other's ciphertext.
func TestConcurrentInitializationConvergesOnOneKey(t *testing.T) {
	store := cryptokeymemory.New()

	const goroutines = 16
	var wg sync.WaitGroup
	ids := make([]string, goroutines)
	texts := make([]string, goroutines)
	for i := range goroutines {
		wg.Go(func() {
			// A fresh keyring per goroutine models a separate process: no
			// in-memory cache is shared, only the store.
			ring := NewKeyring(store)
			ciphertext, keyID, err := ring.Encrypt(t.Context(), []byte("shared"))
			if err != nil {
				t.Errorf("Encrypt: %v", err)
				return
			}
			ids[i] = keyID
			texts[i] = ciphertext
		})
	}
	wg.Wait()

	for i, id := range ids {
		if id != ids[0] {
			t.Fatalf("goroutine %d used key %q, goroutine 0 used %q", i, id, ids[0])
		}
	}

	// Every ciphertext must be readable by any other participant.
	reader := NewKeyring(store)
	for i, text := range texts {
		plaintext, err := reader.Decrypt(t.Context(), text, ids[i])
		if err != nil {
			t.Fatalf("Decrypt ciphertext %d: %v", i, err)
		}
		if string(plaintext) != "shared" {
			t.Fatalf("ciphertext %d decrypted to %q", i, plaintext)
		}
	}
}

func TestKeyringReportsUnknownKeyID(t *testing.T) {
	ring := NewKeyring(cryptokeymemory.New())
	ciphertext, _, err := ring.Encrypt(t.Context(), []byte("secret"))
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	_, err = ring.Decrypt(t.Context(), ciphertext, "no-such-key")
	if !errors.Is(err, cryptokey.ErrNotFound) {
		t.Fatalf("err = %v, want cryptokey.ErrNotFound", err)
	}
}
