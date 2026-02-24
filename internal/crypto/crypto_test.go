package crypto_test

import (
	"bytes"
	"testing"

	"github.com/stickpro/kyp/internal/crypto"
)

func TestDeriveKey(t *testing.T) {
	password := []byte("master-password")
	salt := []byte("1234567890123456")

	key := crypto.DeriveKey(password, salt)

	if len(key) != 32 {
		t.Fatalf("expected key length 32, got %d", len(key))
	}

	key2 := crypto.DeriveKey(password, salt)
	if !bytes.Equal(key, key2) {
		t.Fatal("same inputs must produce same key")
	}

	otherKey := crypto.DeriveKey(password, []byte("other-salt-12345"))
	if bytes.Equal(key, otherKey) {
		t.Fatal("different salt must produce different key")
	}
}

func TestGenerateSalt(t *testing.T) {
	salt, err := crypto.GenerateSalt()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(salt) != 16 {
		t.Fatalf("expected salt length 16, got %d", len(salt))
	}

	// два вызова должны давать разные соли
	salt2, _ := crypto.GenerateSalt()
	if bytes.Equal(salt, salt2) {
		t.Fatal("two salts must not be equal")
	}
}

func TestEncryptDecrypt(t *testing.T) {
	key := bytes.Repeat([]byte("k"), 32)
	plaintext := []byte("super-secret-password")

	ciphertext, err := crypto.Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("encrypt error: %v", err)
	}

	// зашифрованное не должно совпадать с исходным
	if bytes.Equal(ciphertext, plaintext) {
		t.Fatal("ciphertext must differ from plaintext")
	}

	decrypted, err := crypto.Decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("decrypt error: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Fatalf("expected %q, got %q", plaintext, decrypted)
	}
}

func TestEncryptProducesUniqueOutput(t *testing.T) {
	key := bytes.Repeat([]byte("k"), 32)
	plaintext := []byte("same-data")

	ct1, _ := crypto.Encrypt(plaintext, key)
	ct2, _ := crypto.Encrypt(plaintext, key)

	// каждый вызов генерирует новый nonce → разный ciphertext
	if bytes.Equal(ct1, ct2) {
		t.Fatal("two encryptions of same data must produce different ciphertext")
	}
}

func TestDecryptWithWrongKey(t *testing.T) {
	key := bytes.Repeat([]byte("k"), 32)
	wrongKey := bytes.Repeat([]byte("x"), 32)

	ciphertext, _ := crypto.Encrypt([]byte("secret"), key)

	_, err := crypto.Decrypt(ciphertext, wrongKey)
	if err == nil {
		t.Fatal("expected error when decrypting with wrong key")
	}
}

func TestDecryptTooShort(t *testing.T) {
	key := bytes.Repeat([]byte("k"), 32)

	_, err := crypto.Decrypt([]byte("short"), key)
	if err == nil {
		t.Fatal("expected error for too short ciphertext")
	}
}

func TestVerifier(t *testing.T) {
	key := bytes.Repeat([]byte("k"), 32)

	verifier, err := crypto.NewVerifier(key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !crypto.CheckVerifier(verifier, key) {
		t.Fatal("expected true with correct key")
	}

	wrongKey := bytes.Repeat([]byte("x"), 32)
	if crypto.CheckVerifier(verifier, wrongKey) {
		t.Fatal("expected false with wrong key")
	}
}
