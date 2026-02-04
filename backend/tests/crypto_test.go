package tests

import (
	"testing"

	"github.com/messenger/backend/pkg/crypto"
)

func TestEncryptionDecryption(t *testing.T) {
	key := []byte("thisis32byteslongsecretkey123456") // 32 bytes
	plaintext := []byte("Hello, E2E encryption!")

	ciphertext, err := crypto.Encrypt(key, plaintext)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	decrypted, err := crypto.Decrypt(key, ciphertext)
	if err != nil {
		t.Fatalf("Failed to decrypt: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("Expected %s, got %s", string(plaintext), string(decrypted))
	}
}

func TestEncryptionWithInvalidKey(t *testing.T) {
	key := []byte("shortkey")
	plaintext := []byte("Hello")

	_, err := crypto.Encrypt(key, plaintext)
	if err == nil {
		t.Error("Expected error with short key, got nil")
	}
}
