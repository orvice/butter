// Package secretbox provides AES-GCM encryption for credential material
// persisted by repositories (e.g. workspace repository binding PATs, issue
// #214). Ciphertext is base64(nonce||sealed); keys may be raw, hex, or
// base64 encoded 16/24/32-byte values. A nil *Cipher fails closed so callers
// can thread an optional cipher without nil checks at every call site.
package secretbox

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
)

// Cipher encrypts secrets before repository storage.
type Cipher struct {
	aead cipher.AEAD
}

var errNotConfigured = errors.New("encryption key is not configured")

func NewCipher(key string) (*Cipher, error) {
	raw, err := decodeKey(key)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(raw)
	if err != nil {
		return nil, fmt.Errorf("create aes cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create aes-gcm: %w", err)
	}
	return &Cipher{aead: aead}, nil
}

func (c *Cipher) Encrypt(plaintext []byte) (string, error) {
	if c == nil {
		return "", errNotConfigured
	}
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}
	sealed := c.aead.Seal(nil, nonce, plaintext, nil)
	return base64.RawStdEncoding.EncodeToString(append(nonce, sealed...)), nil
}

func (c *Cipher) Decrypt(encoded string) ([]byte, error) {
	if c == nil {
		return nil, errNotConfigured
	}
	raw, err := base64.RawStdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("decode ciphertext: %w", err)
	}
	if len(raw) < c.aead.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}
	nonce, ciphertext := raw[:c.aead.NonceSize()], raw[c.aead.NonceSize():]
	plaintext, err := c.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt ciphertext: %w", err)
	}
	return plaintext, nil
}

func decodeKey(key string) ([]byte, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, errNotConfigured
	}
	if raw, err := base64.StdEncoding.DecodeString(key); err == nil && validAESKeyLen(len(raw)) {
		return raw, nil
	}
	if raw, err := base64.RawStdEncoding.DecodeString(key); err == nil && validAESKeyLen(len(raw)) {
		return raw, nil
	}
	if raw, err := hex.DecodeString(key); err == nil && validAESKeyLen(len(raw)) {
		return raw, nil
	}
	if raw := []byte(key); validAESKeyLen(len(raw)) {
		return raw, nil
	}
	return nil, errors.New("encryption key must decode to 16, 24, or 32 bytes")
}

func validAESKeyLen(n int) bool {
	return n == 16 || n == 24 || n == 32
}
