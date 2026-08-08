package secretbox

import (
	"encoding/base64"
	"strings"
	"testing"
)

const testKey = "0123456789abcdef0123456789abcdef" // raw 32-byte key

func TestEncryptDecryptRoundTrip(t *testing.T) {
	c, err := NewCipher(testKey)
	if err != nil {
		t.Fatalf("NewCipher: %v", err)
	}
	const pat = "glpat-supersecrettoken"
	sealed, err := c.Encrypt([]byte(pat))
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if strings.Contains(sealed, pat) {
		t.Fatalf("ciphertext contains plaintext: %s", sealed)
	}
	got, err := c.Decrypt(sealed)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if string(got) != pat {
		t.Fatalf("round trip mismatch: %q", got)
	}
}

func TestEncryptIsNonDeterministic(t *testing.T) {
	c, err := NewCipher(testKey)
	if err != nil {
		t.Fatalf("NewCipher: %v", err)
	}
	a, _ := c.Encrypt([]byte("tok"))
	b, _ := c.Encrypt([]byte("tok"))
	if a == b {
		t.Fatal("expected fresh nonce per encryption")
	}
}

func TestKeyEncodings(t *testing.T) {
	raw16 := "0123456789abcdef"
	b64 := base64.StdEncoding.EncodeToString([]byte(testKey))
	hexKey := "00112233445566778899aabbccddeeff"
	for _, key := range []string{raw16, testKey, b64, hexKey} {
		if _, err := NewCipher(key); err != nil {
			t.Errorf("NewCipher(%q): %v", key, err)
		}
	}
}

func TestInvalidKey(t *testing.T) {
	for _, key := range []string{"", "short", strings.Repeat("x", 33)} {
		if _, err := NewCipher(key); err == nil {
			t.Errorf("NewCipher(%q): expected error", key)
		}
	}
}

func TestDecryptRejectsGarbageAndWrongKey(t *testing.T) {
	c, _ := NewCipher(testKey)
	if _, err := c.Decrypt("not-base64!!!"); err == nil {
		t.Error("expected error for invalid encoding")
	}
	if _, err := c.Decrypt(base64.RawStdEncoding.EncodeToString([]byte("xx"))); err == nil {
		t.Error("expected error for too-short ciphertext")
	}
	sealed, _ := c.Encrypt([]byte("tok"))
	other, _ := NewCipher("ffffffffffffffffffffffffffffffff")
	if _, err := other.Decrypt(sealed); err == nil {
		t.Error("expected error decrypting with wrong key")
	}
}

func TestNilCipherFailsClosed(t *testing.T) {
	var c *Cipher
	if _, err := c.Encrypt([]byte("tok")); err == nil {
		t.Error("nil cipher Encrypt should error")
	}
	if _, err := c.Decrypt("abc"); err == nil {
		t.Error("nil cipher Decrypt should error")
	}
}
