package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/pbkdf2"
)

// GenerateCardanoWallet generates a proper Cardano wallet
func GenerateCardanoWallet(network string) (address, privateKeyHex, mnemonic string, err error) {
	log.Printf("🔑 Generating Cardano wallet for network: %s", network)

	// Generate 24-word mnemonic (256 bits of entropy)
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate entropy: %v", err)
	}

	mnemonic, err = bip39.NewMnemonic(entropy)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate mnemonic: %v", err)
	}

	log.Printf("✅ Generated 24-word mnemonic")

	// Generate seed from mnemonic
	seed := bip39.NewSeed(mnemonic, "")

	// Generate Ed25519 keypair from seed
	privateKey := ed25519.NewKeyFromSeed(seed[:32])
	publicKey := privateKey.Public().(ed25519.PublicKey)

	// Convert private key to hex
	privateKeyHex = hex.EncodeToString(seed[:32])

	// Generate Cardano address
	address = createCardanoAddress(publicKey, network)

	log.Printf("✅ Wallet generated successfully")
	log.Printf("📍 Address: %s", address)
	log.Printf("🔐 Network: %s", network)

	return address, privateKeyHex, mnemonic, nil
}

// createCardanoAddress creates a valid Cardano Bech32 address
func createCardanoAddress(publicKey ed25519.PublicKey, network string) string {
	// Use Blake2b-224 hash of the public key (Cardano standard)
	hasher, _ := blake2b.New(28, nil) // 28 bytes = 224 bits
	hasher.Write(publicKey)
	paymentHash := hasher.Sum(nil)

	// Determine network and address type
	var header byte
	var hrp string

	switch network {
	case "mainnet":
		header = 0x61 // Enterprise address, mainnet
		hrp = "addr"
	case "preprod", "preview":
		header = 0x60 // Enterprise address, testnet
		hrp = "addr_test"
	default:
		header = 0x60
		hrp = "addr_test"
	}

	// Build address payload: [header] + [payment credential]
	addressBytes := append([]byte{header}, paymentHash...)

	// Encode to Bech32
	address := encodeBech32(hrp, addressBytes)

	return address
}

// encodeBech32 encodes data to Bech32 format
func encodeBech32(hrp string, data []byte) string {
	// Convert 8-bit bytes to 5-bit groups
	converted := convertBits(data, 8, 5, true)

	// Compute checksum
	checksum := bech32Checksum(hrp, converted)

	// Combine data and checksum
	combined := append(converted, checksum...)

	// Encode to Bech32 string
	var result string
	result += hrp + "1" // Separator

	for _, value := range combined {
		result += string(charset[value])
	}

	return result
}

// Bech32 character set
const charset = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"

// convertBits converts between bit groups
func convertBits(data []byte, fromBits, toBits uint, pad bool) []byte {
	var result []byte
	var acc uint32
	var bits uint

	maxv := uint32((1 << toBits) - 1)
	maxAcc := uint32((1 << (fromBits + toBits - 1)) - 1)

	for _, value := range data {
		acc = ((acc << fromBits) | uint32(value)) & maxAcc
		bits += fromBits

		for bits >= toBits {
			bits -= toBits
			result = append(result, byte((acc>>bits)&maxv))
		}
	}

	if pad {
		if bits > 0 {
			result = append(result, byte((acc<<(toBits-bits))&maxv))
		}
	}

	return result
}

// bech32Checksum computes the Bech32 checksum
func bech32Checksum(hrp string, data []byte) []byte {
	values := bech32HrpExpand(hrp)
	values = append(values, data...)
	values = append(values, []byte{0, 0, 0, 0, 0, 0}...)

	polymod := bech32Polymod(values) ^ 1

	var checksum []byte
	for i := 0; i < 6; i++ {
		checksum = append(checksum, byte((polymod>>(5*(5-i)))&31))
	}

	return checksum
}

// bech32HrpExpand expands the HRP for checksum computation
func bech32HrpExpand(hrp string) []byte {
	var result []byte

	for _, c := range hrp {
		result = append(result, byte(c>>5))
	}

	result = append(result, 0)

	for _, c := range hrp {
		result = append(result, byte(c&31))
	}

	return result
}

// bech32Polymod computes the Bech32 polymod
func bech32Polymod(values []byte) uint32 {
	generator := []uint32{0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3}
	chk := uint32(1)

	for _, value := range values {
		top := chk >> 25
		chk = (chk&0x1ffffff)<<5 ^ uint32(value)

		for i := 0; i < 5; i++ {
			if (top>>uint(i))&1 == 1 {
				chk ^= generator[i]
			}
		}
	}

	return chk
}

// EncryptPrivateKey encrypts a private key with password using AES-GCM
func EncryptPrivateKey(privateKey, password, salt string) (string, error) {
	// Decode salt
	saltBytes, err := hex.DecodeString(salt)
	if err != nil {
		return "", fmt.Errorf("invalid salt: %v", err)
	}

	// Derive encryption key from password using PBKDF2
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

	// Generate random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %v", err)
	}

	// Encrypt the private key
	ciphertext := gcm.Seal(nonce, nonce, []byte(privateKey), nil)

	// Return as hex string
	return hex.EncodeToString(ciphertext), nil
}

// DecryptPrivateKey decrypts an encrypted private key
func DecryptPrivateKey(encryptedKey, password, salt string) (string, error) {
	// Decode encrypted data from hex
	encryptedData, err := hex.DecodeString(encryptedKey)
	if err != nil {
		return "", fmt.Errorf("invalid encrypted key format: %v", err)
	}

	// Decode salt from hex
	saltBytes, err := hex.DecodeString(salt)
	if err != nil {
		return "", fmt.Errorf("invalid salt format: %v", err)
	}

	// Derive decryption key from password using PBKDF2
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

	// Extract nonce from encrypted data
	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return "", fmt.Errorf("encrypted data too short")
	}

	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decryption failed - invalid password or corrupted data: %v", err)
	}

	return string(plaintext), nil
}

// GetAddressFromPrivateKey derives the Cardano address from a private key
func GetAddressFromPrivateKey(privateKeyHex string, network string) (string, error) {
	// Decode private key
	privKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("invalid private key: %v", err)
	}

	// Create Ed25519 private key
	privateKey := ed25519.PrivateKey(privKeyBytes)

	// Get public key
	publicKey := privateKey.Public().(ed25519.PublicKey)

	// Generate address
	address := createCardanoAddress(publicKey, network)

	return address, nil
}
