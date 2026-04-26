package blockchain

import (
	"context"
	"fmt"
	"sort"

	"github.com/blockfrost/blockfrost-go"
	"github.com/fxamacker/cbor/v2"
)

type TransactionBuilder struct {
	client *BlockfrostClient
}

func NewTransactionBuilder(client *BlockfrostClient) *TransactionBuilder {
	return &TransactionBuilder{
		client: client,
	}
}

// UTXO represents an unspent transaction output
type UTXO struct {
	TxHash  string
	TxIndex int
	Amount  int64
	Address string
}

// BuildMetadataTransaction builds a transaction with metadata
func (tb *TransactionBuilder) BuildMetadataTransaction(
	ctx context.Context,
	fromAddress string,
	metadata map[string]interface{},
) (map[string]interface{}, error) {
	// Get UTXOs for the address
	utxos, err := tb.client.GetAddressUTXOs(ctx, fromAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get UTXOs: %v", err)
	}

	if len(utxos) == 0 {
		return nil, fmt.Errorf("no UTXOs available for address: %s", fromAddress)
	}

	// Sort UTXOs by amount (largest first) for optimal selection
	sort.Slice(utxos, func(i, j int) bool {
		return parseAmount(utxos[i].Amount[0].Quantity) > parseAmount(utxos[j].Amount[0].Quantity)
	})

	// Get protocol parameters
	params, err := tb.client.GetProtocolParameters(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get protocol parameters: %v", err)
	}

	// FIX: Compare against an empty struct literal instead of nil
	if params == (blockfrost.EpochParameters{}) {
		return nil, fmt.Errorf("protocol parameters were empty")
	}

	minFee := int64(200000)

	// Select UTXOs to cover fees
	var selectedUTXOs []blockfrost.AddressUTXO
	var totalInput int64

	for _, utxo := range utxos {
		amount := parseAmount(utxo.Amount[0].Quantity)
		selectedUTXOs = append(selectedUTXOs, utxo)
		totalInput += amount

		if totalInput >= minFee {
			break
		}
	}

	if totalInput < minFee {
		return nil, fmt.Errorf("insufficient funds: need %d lovelace, have %d", minFee, totalInput)
	}

	// Build transaction structure
	tx := map[string]interface{}{
		"inputs": buildInputs(selectedUTXOs),
		"outputs": []map[string]interface{}{
			{
				"address": fromAddress,
				"amount":  totalInput - minFee, // Return change minus fee
			},
		},
		"metadata": metadata,
		"fee":      minFee,
	}

	return tx, nil
}

func buildInputs(utxos []blockfrost.AddressUTXO) []map[string]interface{} {
	inputs := make([]map[string]interface{}, len(utxos))

	for i, utxo := range utxos {
		inputs[i] = map[string]interface{}{
			"transaction_id": utxo.TxHash,
			"index":          utxo.OutputIndex,
		}
	}

	return inputs
}

func parseAmount(amount string) int64 {
	var value int64
	fmt.Sscanf(amount, "%d", &value)
	return value
}

// SerializeTransaction converts transaction to CBOR
func (tb *TransactionBuilder) SerializeTransaction(tx map[string]interface{}) ([]byte, error) {
	return cbor.Marshal(tx)
}
