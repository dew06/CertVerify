package services

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

type CardanoProductionService struct {
	network           string
	apiKey            string
	blockfrostClient  *BlockfrostClient
	useRealBlockchain bool
	meshEndpoint      string
}

type BlockfrostClient struct {
	apiKey  string
	network string
}

func NewCardanoProductionService(network, apiKey string) *CardanoProductionService {
	useReal := os.Getenv("USE_REAL_BLOCKCHAIN") == "true"

	var client *BlockfrostClient
	if useReal && apiKey != "" {
		client = &BlockfrostClient{
			apiKey:  apiKey,
			network: network,
		}
	}

	// MeshJS doesn't have a public API endpoint yet
	// We'll build transactions ourselves using their approach
	meshEndpoint := "https://api.meshjs.dev" // Placeholder

	return &CardanoProductionService{
		network:           network,
		apiKey:            apiKey,
		blockfrostClient:  client,
		useRealBlockchain: useReal,
		meshEndpoint:      meshEndpoint,
	}
}

func (s *CardanoProductionService) GenerateKeyPair() (address, privateKey, mnemonic string, err error) {
	return "", "", "", fmt.Errorf("use GenerateCardanoWallet from cardano_wallet.go")
}

func (s *CardanoProductionService) HashCertificate(certID, studentName, degree, salt string) string {
	data := fmt.Sprintf("%s|%s|%s|%s", certID, studentName, degree, salt)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// SubmitTransaction builds and submits transaction using MeshJS approach
func (s *CardanoProductionService) SubmitTransaction(
	walletAddress string,
	encryptedPrivKey string,
	password string,
	salt string,
	merkleRoot string,
	certCount int,
	universityName string,
) (txID string, err error) {
	log.Printf("🔗 Starting REAL blockchain transaction submission...")
	log.Printf("Network: %s", s.network)
	log.Printf("Merkle Root: %s", merkleRoot)
	log.Printf("Certificate Count: %d", certCount)

	if !s.useRealBlockchain || s.blockfrostClient == nil {
		log.Printf("⚠️  Real blockchain not enabled")
	}

	// Decrypt private key
	log.Printf("🔓 Decrypting private key...")
	privateKeyHex, err := DecryptPrivateKey(encryptedPrivKey, password, salt)
	if err != nil {
		log.Printf("❌ Failed to decrypt private key: %v", err)
	}

	log.Printf("💼 Wallet Address: %s", walletAddress)
	log.Printf("✅ Private key decrypted")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Get wallet info from Blockfrost
	log.Printf("💰 Checking wallet balance...")
	addressInfo, err := s.getAddressInfo(ctx, walletAddress)
	if err != nil {
		log.Printf("❌ Failed to get address info: %v", err)
	}

	var totalBalance int64
	for _, amount := range addressInfo.Amount {
		if amount.Unit == "lovelace" {
			fmt.Sscanf(amount.Quantity, "%d", &totalBalance)
			break
		}
	}

	log.Printf("💵 Balance: %d lovelace (%.2f ADA)", totalBalance, float64(totalBalance)/1000000)

	minRequired := int64(2000000)
	if totalBalance < minRequired {
		log.Printf("❌ Insufficient balance")
	}

	// Get UTXOs
	log.Printf("📦 Fetching UTXOs...")
	utxos, err := s.getAddressUTXOs(ctx, walletAddress)
	if err != nil || len(utxos) == 0 {
		log.Printf("❌ No UTXOs available")
	}

	log.Printf("✅ Found %d UTXOs", len(utxos))

	// Select UTXO
	var selectedUTXO *UTXO
	for i := range utxos {
		amount := int64(0)
		for _, amt := range utxos[i].Amount {
			if amt.Unit == "lovelace" {
				fmt.Sscanf(amt.Quantity, "%d", &amount)
				break
			}
		}

		if amount >= minRequired {
			selectedUTXO = &UTXO{
				TxHash:      utxos[i].TxHash,
				OutputIndex: utxos[i].OutputIndex,
				Amount:      amount,
			}
			break
		}
	}

	if selectedUTXO == nil {
		log.Printf("❌ No suitable UTXO found")
	}

	log.Printf("✅ Selected UTXO: %s#%d (%d lovelace)",
		selectedUTXO.TxHash[:10]+"...", selectedUTXO.OutputIndex, selectedUTXO.Amount)

	// Build transaction using MeshJS approach
	log.Printf("🔨 Building transaction with MeshJS approach...")

	txCBOR, err := s.buildTransactionMesh(
		walletAddress,
		privateKeyHex,
		merkleRoot,
		universityName,
		certCount,
		selectedUTXO,
	)

	if err != nil {
		log.Printf("❌ Failed to build transaction: %v", err)
	}

	log.Printf("✅ Transaction built and signed")

	// Submit to Blockfrost
	log.Printf("🚀 Submitting transaction to Cardano blockchain...")

	txHash, err := s.submitTransaction(ctx, txCBOR)
	if err != nil {
		log.Printf("❌ Transaction submission failed: %v", err)
	}

	log.Printf("✅ Transaction submitted successfully!")
	log.Printf("🎉 Transaction ID: %s", txHash)
	log.Printf("🔍 View on CardanoScan: https://%s.cardanoscan.io/transaction/%s", s.network, txHash)

	return txHash, nil
}

func (s *CardanoProductionService) buildTransactionMesh(
	address string,
	privateKeyHex string,
	merkleRoot string,
	universityName string,
	certCount int,
	utxo *UTXO,
) (string, error) {

	reqBody := map[string]interface{}{
		"network":          s.network,
		"walletAddress":    address,
		"privateKeyHex":    privateKeyHex,
		"merkleRoot":       merkleRoot,
		"universityName":   universityName,
		"certCount":        certCount,
		"blockfrostApiKey": s.apiKey,
	}

	reqJSON, _ := json.Marshal(reqBody)
	meshURL := "http://localhost:3001/build-transaction"

	log.Printf("🔨 Calling MeshJS + Blockfrost...")

	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Post(meshURL, "application/json", bytes.NewReader(reqJSON))
	if err != nil {
		return "", fmt.Errorf("MeshJS unreachable: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("MeshJS error: %s", string(body))
	}

	var result struct {
		Success bool   `json:"success"`
		TxCbor  string `json:"txCbor"`
		Error   string `json:"error"`
	}

	json.Unmarshal(body, &result)

	if !result.Success {
		return "", fmt.Errorf("MeshJS failed: %s", result.Error)
	}

	log.Printf("✅ MeshJS returned signed CBOR")
	return result.TxCbor, nil
}

func (s *CardanoProductionService) buildWithMeshJS(
	address string,
	privateKeyHex string,
	merkleRoot string,
	universityName string,
	certCount int,
	utxo *UTXO,
	fee int64,
) (string, error) {

	// ADD THESE DEBUG LOGS
	log.Printf("🔍 DEBUG: buildWithMeshJS called")
	log.Printf("🔍 Address: %s", address)
	log.Printf("🔍 Merkle Root: %s", merkleRoot)
	log.Printf("🔍 University: %s", universityName)
	log.Printf("🔍 Cert Count: %d", certCount)
	log.Printf("🔍 API Key: %s...", s.apiKey[:10])

	reqBody := map[string]interface{}{
		"network":          s.network,
		"walletAddress":    address,
		"privateKeyHex":    privateKeyHex,
		"merkleRoot":       merkleRoot,
		"universityName":   universityName,
		"certCount":        certCount,
		"blockfrostApiKey": s.apiKey,
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("❌ Failed to marshal JSON: %v", err)
		return "", err
	}

	log.Printf("🔍 Request JSON length: %d bytes", len(reqJSON))
	log.Printf("🔍 Request JSON: %s", string(reqJSON)[:200]+"...") // First 200 chars

	meshURL := "http://localhost:3001/build-transaction"

	log.Printf("🔨 Calling MeshJS at %s...", meshURL)

	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Post(meshURL, "application/json", bytes.NewReader(reqJSON))
	if err != nil {
		log.Printf("❌ HTTP POST failed: %v", err)
		return "", fmt.Errorf("MeshJS unreachable: %v", err)
	}
	defer resp.Body.Close()

	log.Printf("🔍 Response status: %d", resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)

	log.Printf("🔍 Response body: %s", string(body))

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("MeshJS error: %s", string(body))
	}

	var result struct {
		Success bool   `json:"success"`
		TxCbor  string `json:"txCbor"`
		Error   string `json:"error"`
	}

	json.Unmarshal(body, &result)

	if !result.Success {
		return "", fmt.Errorf("MeshJS failed: %s", result.Error)
	}

	log.Printf("✅ MeshJS returned signed CBOR")
	return result.TxCbor, nil
}

func (s *CardanoProductionService) buildWithCLI(
	address string,
	privateKeyHex string,
	utxo *UTXO,
	change int64,
	fee int64,
	metadata map[string]interface{},
) (string, error) {

	tmpDir := os.TempDir()
	timestamp := time.Now().UnixNano()
	keyFile := fmt.Sprintf("%s/key_%d.skey", tmpDir, timestamp)
	metadataFile := fmt.Sprintf("%s/metadata_%d.json", tmpDir, timestamp)
	txRawFile := fmt.Sprintf("%s/tx_%d.raw", tmpDir, timestamp)
	txSignedFile := fmt.Sprintf("%s/tx_%d.signed", tmpDir, timestamp)

	defer os.Remove(keyFile)
	defer os.Remove(metadataFile)
	defer os.Remove(txRawFile)
	defer os.Remove(txSignedFile)

	// The key from wallet generation is 64 bytes (Ed25519 private key)
	// But we need only the first 32 bytes (seed) for cardano-cli

	privKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("invalid private key hex: %v", err)
	}

	// Take only first 32 bytes (the seed)
	var seedBytes []byte
	if len(privKeyBytes) >= 32 {
		seedBytes = privKeyBytes[:32]
	} else {
		return "", fmt.Errorf("private key too short: %d bytes", len(privKeyBytes))
	}

	seedHex := hex.EncodeToString(seedBytes)

	// Use PaymentSigningKeyShelley_ed25519 (NOT Extended)
	// This format expects just the 32-byte seed with 5820 prefix
	keyJSON := fmt.Sprintf(`{
  "type": "PaymentSigningKeyShelley_ed25519",
  "description": "Payment Signing Key",
  "cborHex": "5820%s"
}`, seedHex)

	if err := os.WriteFile(keyFile, []byte(keyJSON), 0600); err != nil {
		return "", fmt.Errorf("failed to write key file: %v", err)
	}

	log.Printf("🔑 Created signing key file (32-byte seed)")

	// Write metadata
	metadataJSON, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(metadataFile, metadataJSON, 0600); err != nil {
		return "", err
	}

	// Network magic
	magic := "--testnet-magic 2"
	if s.network == "preprod" {
		magic = "--testnet-magic 1"
	} else if s.network == "mainnet" {
		magic = "--mainnet"
	}

	// Build transaction
	buildCmd := fmt.Sprintf(`cardano-cli transaction build-raw \
		--tx-in "%s#%d" \
		--tx-out "%s+%d" \
		--metadata-json-file "%s" \
		--fee %d \
		--out-file "%s"`,
		utxo.TxHash, utxo.OutputIndex,
		address, change,
		metadataFile,
		fee,
		txRawFile,
	)

	log.Printf("🔨 Building raw transaction...")
	output, err := exec.Command("bash", "-c", buildCmd).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("build failed: %v\nOutput: %s", err, string(output))
	}
	log.Printf("✅ Transaction body built")

	// Sign transaction
	signCmd := fmt.Sprintf(`cardano-cli transaction sign \
		%s \
		--tx-body-file "%s" \
		--signing-key-file "%s" \
		--out-file "%s"`,
		magic,
		txRawFile,
		keyFile,
		txSignedFile,
	)

	log.Printf("✍️  Signing transaction...")
	output, err = exec.Command("bash", "-c", signCmd).CombinedOutput()
	if err != nil {
		// Print the actual key file content for debugging
		keyContent, _ := os.ReadFile(keyFile)
		log.Printf("❌ Key file content:\n%s", string(keyContent))
		log.Printf("❌ Private key hex length: %d", len(privateKeyHex))
		log.Printf("❌ Seed hex: %s", seedHex[:20]+"...")
		return "", fmt.Errorf("sign failed: %v\nOutput: %s", err, string(output))
	}
	log.Printf("✅ Transaction signed successfully")

	// Read signed transaction
	txBytes, err := os.ReadFile(txSignedFile)
	if err != nil {
		return "", err
	}

	// Convert to hex
	return hex.EncodeToString(txBytes), nil
}

// Blockfrost API calls
func (s *CardanoProductionService) getAddressInfo(ctx context.Context, address string) (*AddressInfo, error) {
	url := fmt.Sprintf("https://cardano-%s.blockfrost.io/api/v0/addresses/%s", s.network, address)

	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.Header.Set("project_id", s.apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("blockfrost error: %s", string(body))
	}

	var info AddressInfo
	json.NewDecoder(resp.Body).Decode(&info)
	return &info, nil
}

func (s *CardanoProductionService) getAddressUTXOs(ctx context.Context, address string) ([]UTXOResponse, error) {
	url := fmt.Sprintf("https://cardano-%s.blockfrost.io/api/v0/addresses/%s/utxos", s.network, address)

	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.Header.Set("project_id", s.apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var utxos []UTXOResponse
	json.NewDecoder(resp.Body).Decode(&utxos)
	return utxos, nil
}

func (s *CardanoProductionService) getProtocolParameters(ctx context.Context) (*ProtocolParams, error) {
	url := fmt.Sprintf("https://cardano-%s.blockfrost.io/api/v0/epochs/latest/parameters", s.network)

	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.Header.Set("project_id", s.apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var params ProtocolParams
	json.NewDecoder(resp.Body).Decode(&params)
	return &params, nil
}

func (s *CardanoProductionService) submitTransaction(ctx context.Context, txCBORHex string) (string, error) {
	txBytes, err := hex.DecodeString(txCBORHex)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("https://cardano-%s.blockfrost.io/api/v0/tx/submit", s.network)

	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(txBytes))
	req.Header.Set("Content-Type", "application/cbor")
	req.Header.Set("project_id", s.apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("submission failed: %s", string(body))
	}

	txHash := string(body)
	txHash = strings.Trim(txHash, "\"")
	return txHash, nil
}

func (s *CardanoProductionService) VerifyTransaction(ctx context.Context, txHash string) (bool, error) {
	if !s.useRealBlockchain {
		return false, fmt.Errorf("real blockchain not enabled")
	}

	url := fmt.Sprintf("https://cardano-%s.blockfrost.io/api/v0/txs/%s", s.network, txHash)

	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.Header.Set("project_id", s.apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200, nil
}

// Types
type AddressInfo struct {
	Address string `json:"address"`
	Amount  []struct {
		Unit     string `json:"unit"`
		Quantity string `json:"quantity"`
	} `json:"amount"`
}

type UTXOResponse struct {
	TxHash      string `json:"tx_hash"`
	OutputIndex int    `json:"output_index"`
	Amount      []struct {
		Unit     string `json:"unit"`
		Quantity string `json:"quantity"`
	} `json:"amount"`
}

type UTXO struct {
	TxHash      string
	OutputIndex int
	Amount      int64
}

type ProtocolParams struct {
	MinFeeA int64 `json:"min_fee_a"`
	MinFeeB int64 `json:"min_fee_b"`
}
