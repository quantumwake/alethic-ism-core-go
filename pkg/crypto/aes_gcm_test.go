package crypto_test

import (
	"encoding/hex"
	"testing"

	"github.com/quantumwake/alethic-ism-core-go/pkg/crypto"
	"github.com/stretchr/testify/require"
)

func testKey(t *testing.T) string {
	t.Helper()
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	return hex.EncodeToString(key)
}

func TestAESGCM_RoundTrip(t *testing.T) {
	c, err := crypto.NewAESGCMCipher(testKey(t))
	require.NoError(t, err)

	plaintext := []byte(`{"api_key":"sk-secret-123","region":"us-west-2"}`)
	ciphertext, err := c.Encrypt(plaintext)
	require.NoError(t, err)
	require.NotEqual(t, plaintext, ciphertext)

	decrypted, err := c.Decrypt(ciphertext)
	require.NoError(t, err)
	require.Equal(t, plaintext, decrypted)
}

func TestAESGCM_TamperDetection(t *testing.T) {
	c, err := crypto.NewAESGCMCipher(testKey(t))
	require.NoError(t, err)

	ciphertext, err := c.Encrypt([]byte("sensitive data"))
	require.NoError(t, err)

	// Flip a byte in the ciphertext body (after nonce)
	ciphertext[len(ciphertext)-1] ^= 0xFF

	_, err = c.Decrypt(ciphertext)
	require.Error(t, err)
	require.Contains(t, err.Error(), "decryption failed")
}

func TestAESGCM_KeyIDStability(t *testing.T) {
	hexKey := testKey(t)

	c1, err := crypto.NewAESGCMCipher(hexKey)
	require.NoError(t, err)

	c2, err := crypto.NewAESGCMCipher(hexKey)
	require.NoError(t, err)

	require.Equal(t, c1.KeyID(), c2.KeyID())
	require.Len(t, c1.KeyID(), 8) // 4 bytes = 8 hex chars
}

func TestAESGCM_DifferentKeysDifferentIDs(t *testing.T) {
	key1 := make([]byte, 32)
	key2 := make([]byte, 32)
	key2[0] = 0xFF

	c1, err := crypto.NewAESGCMCipher(hex.EncodeToString(key1))
	require.NoError(t, err)

	c2, err := crypto.NewAESGCMCipher(hex.EncodeToString(key2))
	require.NoError(t, err)

	require.NotEqual(t, c1.KeyID(), c2.KeyID())
}

func TestAESGCM_InvalidKeyLength(t *testing.T) {
	_, err := crypto.NewAESGCMCipher("aabbccdd") // only 4 bytes
	require.Error(t, err)
	require.Contains(t, err.Error(), "32 bytes")
}

func TestAESGCM_InvalidHex(t *testing.T) {
	_, err := crypto.NewAESGCMCipher("not-hex-at-all!")
	require.Error(t, err)
}

func TestAESGCM_CiphertextTooShort(t *testing.T) {
	c, err := crypto.NewAESGCMCipher(testKey(t))
	require.NoError(t, err)

	_, err = c.Decrypt([]byte("short"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "too short")
}

func TestAESGCM_FromPassphrase_RoundTrip(t *testing.T) {
	c, err := crypto.NewAESGCMCipherFromPassphrase("my-dev-password")
	require.NoError(t, err)

	plaintext := []byte("hello vault")
	ct, err := c.Encrypt(plaintext)
	require.NoError(t, err)

	pt, err := c.Decrypt(ct)
	require.NoError(t, err)
	require.Equal(t, plaintext, pt)
}

func TestAESGCM_FromPassphrase_Deterministic(t *testing.T) {
	c1, err := crypto.NewAESGCMCipherFromPassphrase("same-password")
	require.NoError(t, err)

	c2, err := crypto.NewAESGCMCipherFromPassphrase("same-password")
	require.NoError(t, err)

	require.Equal(t, c1.KeyID(), c2.KeyID())

	// Encrypt with c1, decrypt with c2
	ct, err := c1.Encrypt([]byte("cross-cipher test"))
	require.NoError(t, err)

	pt, err := c2.Decrypt(ct)
	require.NoError(t, err)
	require.Equal(t, []byte("cross-cipher test"), pt)
}

func TestAESGCM_UniqueNonces(t *testing.T) {
	c, err := crypto.NewAESGCMCipher(testKey(t))
	require.NoError(t, err)

	plaintext := []byte("same data")
	ct1, err := c.Encrypt(plaintext)
	require.NoError(t, err)

	ct2, err := c.Encrypt(plaintext)
	require.NoError(t, err)

	// Same plaintext should produce different ciphertext due to random nonces
	require.NotEqual(t, ct1, ct2)
}
