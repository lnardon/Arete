package tokencrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

// Cipher encrypts/decrypts secrets (e.g. OAuth tokens) at rest with
// AES-256-GCM, so a database leak alone can't be used to access a connected
// external account.
type Cipher struct {
	gcm cipher.AEAD
}

// New builds a cipher from a 32-byte, base64-encoded key. An empty key is
// accepted (the integration it protects may be optional/unconfigured in this
// environment) and yields a cipher whose Encrypt/Decrypt calls fail clearly
// rather than panicking, so callers don't need to nil-check this at every
// construction site.
func New(base64Key string) (*Cipher, error) {
	if base64Key == "" {
		return &Cipher{}, nil
	}

	key, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return nil, fmt.Errorf("decode token encryption key: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("token encryption key must be 32 bytes, got %d", len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("build cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("build GCM: %w", err)
	}
	return &Cipher{gcm: gcm}, nil
}

func (c *Cipher) Encrypt(plaintext string) (string, error) {
	if c.gcm == nil {
		return "", errors.New("google token encryption not configured (set GOOGLE_TOKEN_ENCRYPTION_KEY)")
	}

	nonce := make([]byte, c.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}
	ciphertext := c.gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (c *Cipher) Decrypt(encoded string) (string, error) {
	if c.gcm == nil {
		return "", errors.New("google token encryption not configured (set GOOGLE_TOKEN_ENCRYPTION_KEY)")
	}

	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("decode ciphertext: %w", err)
	}
	nonceSize := c.gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := c.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}
	return string(plaintext), nil
}
