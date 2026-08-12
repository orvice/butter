package secretbox

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/google/uuid"

	"go.orx.me/apps/butter/internal/repo/cryptokey"
)

// Keyring encrypts credential material under a database-backed master key
// (issue #264). Unlike NewCipher, which takes a key from configuration, the
// Keyring lazily initializes its key on first use and tags every ciphertext
// with the key ID that produced it, so a later rotation or Secret Manager
// migration can still read what this release wrote.
//
// A nil *Keyring fails closed, matching *Cipher, so services can hold an
// optional keyring without nil checks at every call site.
type Keyring struct {
	store cryptokey.Repository

	mu       sync.RWMutex
	activeID string
	ciphers  map[string]*Cipher
}

var errKeyringUnavailable = errors.New("credential encryption is not available")

// NewKeyring returns a keyring backed by store. No I/O happens until the
// first Encrypt or Decrypt.
func NewKeyring(store cryptokey.Repository) *Keyring {
	if store == nil {
		return nil
	}
	return &Keyring{store: store, ciphers: make(map[string]*Cipher)}
}

// Encrypt seals plaintext under the active key and returns the ciphertext
// together with the key ID needed to open it.
func (k *Keyring) Encrypt(ctx context.Context, plaintext []byte) (ciphertext, keyID string, err error) {
	if k == nil {
		return "", "", errKeyringUnavailable
	}
	id, cipher, err := k.active(ctx)
	if err != nil {
		return "", "", err
	}
	sealed, err := cipher.Encrypt(plaintext)
	if err != nil {
		return "", "", err
	}
	return sealed, id, nil
}

// Decrypt opens ciphertext that was sealed under keyID. An empty keyID is
// rejected rather than silently retried against the active key: guessing
// would turn a bookkeeping bug into an authentication failure that looks like
// a revoked credential.
func (k *Keyring) Decrypt(ctx context.Context, ciphertext, keyID string) ([]byte, error) {
	if k == nil {
		return nil, errKeyringUnavailable
	}
	if ciphertext == "" {
		return nil, errors.New("ciphertext is empty")
	}
	if keyID == "" {
		return nil, errors.New("ciphertext has no key id")
	}
	cipher, err := k.cipherFor(ctx, keyID)
	if err != nil {
		return nil, err
	}
	return cipher.Decrypt(ciphertext)
}

// active returns the active key ID and its cipher, initializing the master
// key on first use.
func (k *Keyring) active(ctx context.Context) (string, *Cipher, error) {
	k.mu.RLock()
	id := k.activeID
	cipher := k.ciphers[id]
	k.mu.RUnlock()
	if id != "" && cipher != nil {
		return id, cipher, nil
	}

	key, err := k.store.EnsureActive(ctx, generateMasterKey)
	if err != nil {
		return "", nil, fmt.Errorf("initialize master key: %w", err)
	}
	cipher, err = cipherFromMaterial(key.Material)
	if err != nil {
		return "", nil, err
	}

	k.mu.Lock()
	k.activeID = key.ID
	k.ciphers[key.ID] = cipher
	k.mu.Unlock()
	return key.ID, cipher, nil
}

func (k *Keyring) cipherFor(ctx context.Context, keyID string) (*Cipher, error) {
	k.mu.RLock()
	cipher := k.ciphers[keyID]
	k.mu.RUnlock()
	if cipher != nil {
		return cipher, nil
	}
	// A ciphertext may reference the active key we have not loaded yet, or a
	// superseded key; both are plain lookups by ID.
	if _, _, err := k.active(ctx); err != nil {
		return nil, err
	}
	k.mu.RLock()
	cipher = k.ciphers[keyID]
	k.mu.RUnlock()
	if cipher != nil {
		return cipher, nil
	}

	key, err := k.store.Get(ctx, keyID)
	if err != nil {
		return nil, fmt.Errorf("load master key %q: %w", keyID, err)
	}
	cipher, err = cipherFromMaterial(key.Material)
	if err != nil {
		return nil, err
	}
	k.mu.Lock()
	k.ciphers[key.ID] = cipher
	k.mu.Unlock()
	return cipher, nil
}

// generateMasterKey produces a fresh 256-bit key with a random ID.
func generateMasterKey() (cryptokey.MasterKey, error) {
	material := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, material); err != nil {
		return cryptokey.MasterKey{}, fmt.Errorf("read random key material: %w", err)
	}
	return cryptokey.MasterKey{ID: uuid.NewString(), Material: material}, nil
}

func cipherFromMaterial(material []byte) (*Cipher, error) {
	if !validAESKeyLen(len(material)) {
		return nil, fmt.Errorf("master key must be 16, 24, or 32 bytes, got %d", len(material))
	}
	return NewCipher(base64.StdEncoding.EncodeToString(material))
}
