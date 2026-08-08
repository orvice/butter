package mcpoauth

import "go.orx.me/apps/butter/internal/secretbox"

// Cipher encrypts token payloads and client secrets before repository
// storage. It is the shared AES-GCM cipher from internal/secretbox; the
// alias keeps this package's existing call sites and public surface stable.
type Cipher = secretbox.Cipher

func NewCipher(key string) (*Cipher, error) {
	return secretbox.NewCipher(key)
}
