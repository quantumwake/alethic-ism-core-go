package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
)

// AESGCMCipher implements Cipher using AES-256-GCM.
// The nonce is prepended to the ciphertext on Encrypt and stripped on Decrypt.
type AESGCMCipher struct {
	aead  cipher.AEAD
	keyID string
}

// NewAESGCMCipher creates a cipher from a 32-byte hex-encoded key.
func NewAESGCMCipher(hexKey string) (*AESGCMCipher, error) {
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("crypto: invalid hex key: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("crypto: key must be 32 bytes, got %d", len(key))
	}
	return newCipherFromKey(key)
}

// NewAESGCMCipherFromPassphrase derives a 32-byte key from a passphrase using Argon2id.
// Useful for development/testing; production should use a proper hex key.
func NewAESGCMCipherFromPassphrase(passphrase string) (*AESGCMCipher, error) {
	// Fixed salt so the same passphrase always produces the same key.
	// This is acceptable because Argon2id is designed for password hashing
	// and we rely on AES-GCM nonces for per-message uniqueness.
	salt := []byte("alethic-ism-vault-dev-salt")
	key := argon2.IDKey([]byte(passphrase), salt, 1, 64*1024, 4, 32)
	return newCipherFromKey(key)
}

func newCipherFromKey(key []byte) (*AESGCMCipher, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("crypto: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("crypto: %w", err)
	}

	// KeyID = first 8 hex chars of SHA-256(key)
	h := sha256.Sum256(key)
	keyID := hex.EncodeToString(h[:4])

	return &AESGCMCipher{aead: aead, keyID: keyID}, nil
}

func (c *AESGCMCipher) Encrypt(plaintext []byte) ([]byte, error) {
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("crypto: failed to generate nonce: %w", err)
	}
	// nonce is prepended to ciphertext
	return c.aead.Seal(nonce, nonce, plaintext, nil), nil
}

func (c *AESGCMCipher) Decrypt(ciphertext []byte) ([]byte, error) {
	nonceSize := c.aead.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("crypto: ciphertext too short")
	}
	nonce, ct := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := c.aead.Open(nil, nonce, ct, nil)
	if err != nil {
		return nil, fmt.Errorf("crypto: decryption failed: %w", err)
	}
	return plaintext, nil
}

func (c *AESGCMCipher) KeyID() string {
	return c.keyID
}
