// pkg/crypto/crypto.go
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
)

// Encrypt encrypts a plain text string using AES-256-GCM
//
// AES-GCM is authenticated encryption — it both encrypts AND
// detects if the ciphertext was tampered with.
// We use this to store LinkedIn/social access tokens in the DB.
//
// Flow:
//
//	plain text → AES-GCM encrypt (with random nonce) → base64 encode → stored in DB
func Encrypt(plainText string) (string, error) {
	key, err := getEncryptionKey()
	if err != nil {
		return "", err
	}

	// Create AES cipher block using our 32-byte key (AES-256)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher block: %w", err)
	}

	// GCM (Galois/Counter Mode) wraps the block cipher
	// It provides both encryption and authentication
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Nonce = "number used once" — a random value critical for AES-GCM
	// Never reuse the same nonce with the same key — we generate fresh each time
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Seal encrypts plainText and appends an authentication tag
	// We prepend the nonce to the ciphertext so Decrypt can extract it
	// Final format: [nonce][encrypted data][auth tag]
	cipherText := gcm.Seal(nonce, nonce, []byte(plainText), nil)

	// Base64 encode so we can safely store as a string in PostgreSQL TEXT column
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

// Decrypt reverses Encrypt — takes base64 encoded ciphertext, returns plain text
//
// Flow:
//
//	stored DB value → base64 decode → extract nonce → AES-GCM decrypt → plain text
func Decrypt(encoded string) (string, error) {
	key, err := getEncryptionKey()
	if err != nil {
		return "", err
	}

	// Reverse the base64 encoding
	cipherText, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher block: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Validate minimum length — nonce + at least 1 byte of data
	nonceSize := gcm.NonceSize()
	if len(cipherText) < nonceSize {
		return "", fmt.Errorf("ciphertext too short — data may be corrupted")
	}

	// Split: first nonceSize bytes = nonce, rest = actual ciphertext + auth tag
	nonce, cipherText := cipherText[:nonceSize], cipherText[nonceSize:]

	// Open decrypts AND verifies the authentication tag
	// If the ciphertext was tampered with, this returns an error
	plainText, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt — data may be tampered: %w", err)
	}

	return string(plainText), nil
}

// ─────────────────────────────────────────
// Helper
// ─────────────────────────────────────────

// getEncryptionKey reads and validates the encryption key from env
// AES-256 requires exactly 32 bytes
func getEncryptionKey() ([]byte, error) {
	key := os.Getenv("TOKEN_ENCRYPTION_KEY")
	if key == "" {
		return nil, fmt.Errorf("TOKEN_ENCRYPTION_KEY is not set in environment")
	}

	if len(key) != 32 {
		return nil, fmt.Errorf(
			"TOKEN_ENCRYPTION_KEY must be exactly 32 characters for AES-256, got %d",
			len(key),
		)
	}

	return []byte(key), nil
}
