package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/argon2"
)

func DeriveKey(password, salt []byte) []byte {
	return argon2.IDKey(
		password,
		salt,
		1,
		64*1024,
		4,
		32)
}

func GenerateSalt() ([]byte, error) {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}
	return salt, nil
}

func Encrypt(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	g, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, g.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	return g.Seal(nonce, nonce, data, nil), nil
}

func Decrypt(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	g, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(data) < g.NonceSize() {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce := data[:g.NonceSize()]
	ciphertext := data[g.NonceSize():]

	return g.Open(nil, nonce, ciphertext, nil)
}

func NewVerifier(key []byte) ([]byte, error) {
	fixedValue := []byte("verification")
	return Encrypt(fixedValue, key)
}

func CheckVerifier(verifier, key []byte) bool {
	decrypted, err := Decrypt(verifier, key)
	if err != nil {
		return false
	}
	return string(decrypted) == "verification"
}
