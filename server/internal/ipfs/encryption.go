package ipfs

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

// DeriveKey converts password into encryption key
func DeriveKey(password, salt string) []byte {
	return pbkdf2.Key(
		[]byte(password),
		[]byte(salt),
		100000, // Iterations (makes brute-force slow)
		32,     // Key size (256 bits)
		sha256.New,
	)
}

// Encrypt using AES-256-GCM
func Encrypt(plaintext, password, salt string) (string, error) {
	key := DeriveKey(password, salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Random nonce
	nonce := make([]byte, gcm.NonceSize())
	io.ReadFull(rand.Reader, nonce)

	// Encrypt
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt reverses encryption
func Decrypt(ciphertext64, password, salt string) (string, error) {
	key := DeriveKey(password, salt)

	ciphertext, err := base64.StdEncoding.DecodeString(ciphertext64)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// SecureWipe overwrites memory with zeros
func SecureWipe(data []byte) {
	for i := range data {
		data[i] = 0
	}
}
