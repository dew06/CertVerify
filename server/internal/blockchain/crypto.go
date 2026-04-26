package blockchain

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/pbkdf2"
)

// DecryptPrivateKey decrypts an encrypted private key
func DecryptPrivateKey(encryptedKey, password, salt string) (string, error) {
	// Decode hex-encoded encrypted data
	encryptedData, err := hex.DecodeString(encryptedKey)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted key: %v", err)
	}

	// Decode salt
	saltBytes, err := hex.DecodeString(salt)
	if err != nil {
		return "", fmt.Errorf("failed to decode salt: %v", err)
	}

	// Derive key from password using PBKDF2
	key := pbkdf2.Key([]byte(password), saltBytes, 100000, 32, sha256.New)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %v", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %v", err)
	}

	// Extract nonce
	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return "", fmt.Errorf("encrypted data too short")
	}

	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %v", err)
	}

	return string(plaintext), nil
}

// SignTransaction signs a transaction with a private key
func SignTransaction(txHash string, privateKey string) (string, error) {
	// This is a simplified version
	// In production, you would use proper Cardano signing with Ed25519

	// For now, we'll create a simple signature
	combined := txHash + privateKey
	hash := sha256.Sum256([]byte(combined))
	signature := hex.EncodeToString(hash[:])

	return signature, nil
}

// GenerateTransactionHash generates a hash for a transaction
func GenerateTransactionHash(tx []byte) string {
	hash := sha256.Sum256(tx)
	return hex.EncodeToString(hash[:])
}
