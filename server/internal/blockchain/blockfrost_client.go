package blockchain

import (
	"context"
	"log"

	"github.com/blockfrost/blockfrost-go"
)

type BlockfrostClient struct {
	client  blockfrost.APIClient
	network string
}

func NewBlockfrostClient(apiKey, network string) *BlockfrostClient {
	// Map network to Blockfrost server
	var server string
	switch network {
	case "preprod":
		server = blockfrost.CardanoPreProd
	case "preview":
		server = blockfrost.CardanoPreview
	case "mainnet":
		server = blockfrost.CardanoMainNet
	default:
		server = blockfrost.CardanoPreProd // Default to preprod
	}

	client := blockfrost.NewAPIClient(
		blockfrost.APIClientOptions{
			ProjectID: apiKey,
			Server:    server,
		},
	)

	return &BlockfrostClient{
		client:  client,
		network: network,
	}
}

// GetHealth checks if Blockfrost API is accessible
func (c *BlockfrostClient) GetHealth(ctx context.Context) (bool, error) {
	health, err := c.client.Health(ctx)
	if err != nil {
		return false, err
	}
	return health.IsHealthy, nil
}

// GetAddressInfo gets information about a Cardano address
func (c *BlockfrostClient) GetAddressInfo(ctx context.Context, address string) (blockfrost.Address, error) {
	addr, err := c.client.Address(ctx, address)
	return addr, err
}

// GetLatestBlock gets the latest block information
func (c *BlockfrostClient) GetLatestBlock(ctx context.Context) (blockfrost.Block, error) {
	block, err := c.client.BlockLatest(ctx)
	return block, err
}

// GetTransaction gets transaction details
func (c *BlockfrostClient) GetTransaction(ctx context.Context, hash string) (blockfrost.TransactionContent, error) {
	tx, err := c.client.Transaction(ctx, hash)
	return tx, err
}

// SubmitTransaction submits a signed transaction
func (c *BlockfrostClient) SubmitTransaction(ctx context.Context, txData []byte) (string, error) {
	txHash, err := c.client.TransactionSubmit(ctx, txData)
	if err != nil {
		return "", err
	}
	return txHash, nil
}

// GetAddressUTXOs gets all UTXOs for an address
func (c *BlockfrostClient) GetAddressUTXOs(ctx context.Context, address string) ([]blockfrost.AddressUTXO, error) {
	utxos, err := c.client.AddressUTXOs(ctx, address, blockfrost.APIQueryParams{})
	return utxos, err
}

// GetProtocolParameters gets current protocol parameters
func (c *BlockfrostClient) GetProtocolParameters(ctx context.Context) (blockfrost.EpochParameters, error) {
	epoch, err := c.client.EpochLatest(ctx)
	if err != nil {
		return blockfrost.EpochParameters{}, err
	}

	params, err := c.client.EpochParameters(ctx, epoch.Epoch)
	if err != nil {
		return blockfrost.EpochParameters{}, err
	}

	return params, nil
}

// LogInfo logs blockchain information
func (c *BlockfrostClient) LogInfo(ctx context.Context) {
	health, err := c.GetHealth(ctx)
	if err != nil {
		log.Printf("❌ Blockfrost health check failed: %v", err)
		return
	}

	if !health {
		log.Printf("⚠️  Blockfrost is not healthy")
		return
	}

	block, err := c.GetLatestBlock(ctx)
	if err != nil {
		log.Printf("❌ Failed to get latest block: %v", err)
		return
	}

	log.Printf("✅ Blockfrost connected successfully")
	log.Printf("📦 Latest block: %d", block.Height)
	log.Printf("🌐 Network: %s", c.network)
}
