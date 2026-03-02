package crypto

// Cipher provides symmetric encryption and decryption.
type Cipher interface {
	// Encrypt encrypts plaintext and returns ciphertext.
	Encrypt(plaintext []byte) ([]byte, error)

	// Decrypt decrypts ciphertext and returns plaintext.
	Decrypt(ciphertext []byte) ([]byte, error)

	// KeyID returns a short fingerprint identifying the encryption key.
	KeyID() string
}
