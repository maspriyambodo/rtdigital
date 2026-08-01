package auth_test

import (
	"crypto/rand"
	"testing"

	"github.com/maspriyambodo/rtdigital/services/api/internal/auth"
)

func TestAESCrypter(t *testing.T) {
	t.Parallel()

	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("generate key: %v", err)
	}
	crypter, err := auth.NewAESCrypter(key)
	if err != nil {
		t.Fatalf("NewAESCrypter(): %v", err)
	}

	plaintext := "the eagle flies at midnight"
	ciphertext, err := crypter.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt(): %v", err)
	}
	if ciphertext == plaintext || ciphertext == "" {
		t.Error("ciphertext must differ from plaintext")
	}

	ciphertext2, err := crypter.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() second call: %v", err)
	}
	if ciphertext == ciphertext2 {
		t.Error("Encrypt() must use a random nonce")
	}

	decrypted, err := crypter.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt(): %v", err)
	}
	if decrypted != plaintext {
		t.Errorf("Decrypt() = %q; want %q", decrypted, plaintext)
	}

	for _, invalid := range []string{"", "short", "invalid-base64-or-short", ciphertext + "tamper"} {
		if _, err := crypter.Decrypt(invalid); err != auth.ErrDecryptionFailed {
			t.Errorf("Decrypt(%q) = %v; want ErrDecryptionFailed", invalid, err)
		}
	}

	otherKey := make([]byte, 32)
	if _, err := rand.Read(otherKey); err != nil {
		t.Fatalf("generate other key: %v", err)
	}
	otherCrypter, err := auth.NewAESCrypter(otherKey)
	if err != nil {
		t.Fatalf("NewAESCrypter(other): %v", err)
	}
	if _, err := otherCrypter.Decrypt(ciphertext); err != auth.ErrDecryptionFailed {
		t.Errorf("Decrypt(wrong key) = %v; want ErrDecryptionFailed", err)
	}

	if _, err := auth.NewAESCrypter(make([]byte, 16)); err == nil {
		t.Error("NewAESCrypter(short key): expected error")
	}
}