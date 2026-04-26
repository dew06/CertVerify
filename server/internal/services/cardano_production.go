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

	// Fix 1: Return error if tx-builder is not running
	if !s.useRealBlockchain || s.blockfrostClient == nil {
		return "", fmt.Errorf("real blockchain not enabled: set USE_REAL_BLOCKCHAIN=true and provide BLOCKFROST_API_KEY")
	}

	// Fix 2: Return error on decrypt failure
	log.Printf("🔓 Decrypting private key...")
	privateKeyHex, err := DecryptPrivateKey(encryptedPrivKey, password, salt)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt private key: %w", err)
	}
	log.Printf("✅ Private key decrypted")
	log.Printf("💼 Wallet Address: %s", walletAddress)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Fix 3: Return error on balance fetch failure
	log.Printf("💰 Checking wallet balance...")
	addressInfo, err := s.getAddressInfo(ctx, walletAddress)
	if err != nil {
		return "", fmt.Errorf("failed to get address info: %w", err)
	}

	var totalBalance int64
	for _, amount := range addressInfo.Amount {
		if amount.Unit == "lovelace" {
			fmt.Sscanf(amount.Quantity, "%d", &totalBalance)
			break
		}
	}
	log.Printf("💵 Balance: %d lovelace (%.2f ADA)", totalBalance, float64(totalBalance)/1000000)

	// Fix 4: Return error on insufficient balance
	minRequired := int64(2000000)
	if totalBalance < minRequired {
		return "", fmt.Errorf("insufficient balance: have %.2f ADA, need at least 2 ADA", float64(totalBalance)/1000000)
	}

	// Fix 5: Return error on UTXO fetch failure
	log.Printf("📦 Fetching UTXOs...")
	utxos, err := s.getAddressUTXOs(ctx, walletAddress)
	if err != nil {
		return "", fmt.Errorf("failed to fetch UTXOs: %w", err)
	}
	if len(utxos) == 0 {
		return "", fmt.Errorf("no UTXOs available for address: %s", walletAddress)
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

	// Fix 6: Return error if no suitable UTXO
	if selectedUTXO == nil {
		return "", fmt.Errorf("no UTXO with sufficient balance found (need > 2 ADA)")
	}
	log.Printf("✅ Selected UTXO: %s#%d (%d lovelace)",
		selectedUTXO.TxHash[:10]+"...", selectedUTXO.OutputIndex, selectedUTXO.Amount)

	// Fix 7: Return error on transaction build failure
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
		return "", fmt.Errorf("failed to build transaction: %w", err)
	}
	log.Printf("✅ Transaction built and signed")

	// Fix 8: Return error on submission failure
	log.Printf("🚀 Submitting transaction to Cardano blockchain...")
	txHash, err := s.submitTransaction(ctx, txCBOR)
	if err != nil {
		return "", fmt.Errorf("transaction submission failed: %w", err)
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
