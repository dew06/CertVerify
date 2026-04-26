package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/blockfrost/blockfrost-go"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	apiKey := os.Getenv("BLOCKFROST_API_KEY")
	network := os.Getenv("CARDANO_NETWORK")

	fmt.Printf("Testing Blockfrost Connection...\n")
	fmt.Printf("Network: %s\n", network)
	fmt.Printf("API Key: %s...\n", apiKey[:20])
	fmt.Println()

	// Create client with explicit server
	var server string
	switch network {
	case "preprod":
		server = blockfrost.CardanoPreProd
	case "preview":
		server = blockfrost.CardanoPreview
	case "mainnet":
		server = blockfrost.CardanoMainNet
	default:
		log.Fatal("Invalid network:", network)
	}

	client := blockfrost.NewAPIClient(
		blockfrost.APIClientOptions{
			ProjectID: apiKey,
			Server:    server,
		},
	)

	ctx := context.Background()

	// Test health
	health, err := client.Health(ctx)
	if err != nil {
		log.Fatalf("❌ Health check failed: %v", err)
	}

	fmt.Printf("✅ Health: %v\n", health.IsHealthy)

	// Test latest block
	block, err := client.BlockLatest(ctx)
	if err != nil {
		log.Fatalf("❌ Block fetch failed: %v", err)
	}

	fmt.Printf("✅ Latest Block: %d\n", block.Height)
	fmt.Printf("✅ Block Hash: %s\n", block.Hash[:20]+"...")

	fmt.Println("\n✅ Connection successful!")
}
