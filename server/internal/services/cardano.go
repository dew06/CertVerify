package services

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

type CardanoService struct{}

func (s *CardanoService) GenerateKeyPair() (publicKey, privateKey string, err error) {
	// Get current working directory
	cwd, _ := os.Getwd()
	log.Printf("Current working directory: %s", cwd)

	scriptPath := filepath.Join("scripts", "generate-test-wallet.js")
	fullPath := filepath.Join(cwd, scriptPath)
	log.Printf("Full script path: %s", fullPath)

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return "", "", fmt.Errorf("script not found at: %s", fullPath)
	}
	log.Printf("Script file exists: ✓")

	cmd := exec.Command("node", scriptPath, "--json")
	cmd.Dir = cwd

	// Capture both stdout and stderr separately
	output, err := cmd.CombinedOutput()

	log.Printf("Script exit code: %v", err)
	log.Printf("Script output length: %d bytes", len(output))
	log.Printf("Script output: %s", string(output))

	if err != nil {
		return "", "", fmt.Errorf("script failed: %v - output: %s", err, string(output))
	}

	var result struct {
		Mnemonic   string `json:"mnemonic"`
		Address    string `json:"address"`
		PrivateKey string `json:"privateKey"`
		PublicKey  string `json:"publicKey"`
	}

	err = json.Unmarshal(output, &result)
	if err != nil {
		log.Printf("JSON parse error: %v", err)
		return "", "", fmt.Errorf("parse failed: %v - output: %s", err, string(output))
	}

	log.Printf("Wallet generated successfully!")
	return result.Address, result.PrivateKey, nil
}

func (s *CardanoService) HashCertificate(certID, studentName, degree, salt string) string {
	data := fmt.Sprintf("%s|%s|%s|%s", certID, studentName, degree, salt)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func (s *CardanoService) EphemeralSign(
	encryptedPrivKey string,
	password string,
	salt string,
	certificateHash string,
) (txID string, err error) {
	txHash := sha256.Sum256([]byte(certificateHash))
	return hex.EncodeToString(txHash[:]), nil
}
