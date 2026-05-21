package mcpoauth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
)

// Cipher encrypts token payloads and client secrets before repository storage.
type Cipher struct {
	aead cipher.AEAD
}

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
		return "", fmt.Errorf("mcp oauth encryption key is not configured")
	}
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}
	sealed := c.aead.Seal(nil, nonce, plaintext, nil)
	out := append(nonce, sealed...)
	return base64.RawStdEncoding.EncodeToString(out), nil
}

func (c *Cipher) Decrypt(encoded string) ([]byte, error) {
	if c == nil {
		return nil, fmt.Errorf("mcp oauth encryption key is not configured")
	}
	raw, err := base64.RawStdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("decode ciphertext: %w", err)
	}
	if len(raw) < c.aead.NonceSize() {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce := raw[:c.aead.NonceSize()]
	ciphertext := raw[c.aead.NonceSize():]
	plaintext, err := c.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt ciphertext: %w", err)
	}
	return plaintext, nil
}

func decodeKey(key string) ([]byte, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, fmt.Errorf("mcp oauth encryption key is required")
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
	raw := []byte(key)
	if validAESKeyLen(len(raw)) {
		return raw, nil
	}
	return nil, fmt.Errorf("mcp oauth encryption key must decode to 16, 24, or 32 bytes")
}

func validAESKeyLen(n int) bool {
	return n == 16 || n == 24 || n == 32
}
