#!/bin/bash

# Cardano Transaction Submission Script (Offline Mode - No Node Required)
# Uses transaction build-raw instead of build

set -e

# Arguments
WALLET_ADDRESS=$1
PRIVATE_KEY_HEX=$2
MERKLE_ROOT=$3
UNIVERSITY_NAME=$4
CERT_COUNT=$5
UTXO_HASH=$6
UTXO_INDEX=$7
UTXO_AMOUNT=$8
NETWORK=$9

# Determine network parameters
if [ "$NETWORK" = "preview" ]; then
    MAGIC="--testnet-magic 2"
elif [ "$NETWORK" = "preprod" ]; then
    MAGIC="--testnet-magic 1"
else
    MAGIC="--mainnet"
fi

# Create temp directory
WORK_DIR=$(mktemp -d)
cd $WORK_DIR

# Create signing key file from hex private key
echo "{" > payment.skey
echo "  \"type\": \"PaymentSigningKeyShelley_ed25519\"," >> payment.skey
echo "  \"description\": \"Payment Signing Key\"," >> payment.skey
echo "  \"cborHex\": \"5820${PRIVATE_KEY_HEX}\"" >> payment.skey
echo "}" >> payment.skey

# Create metadata JSON (CIP-20 format)
cat > metadata.json << EOF
{
  "674": {
    "msg": [
      "University: $UNIVERSITY_NAME",
      "Merkle Root: $MERKLE_ROOT",
      "Certificates: $CERT_COUNT",
      "Timestamp: $(date +%s)"
    ]
  }
}
EOF

# Calculate transaction fee (estimated)
# Base fee: ~170,000 lovelace (0.17 ADA)
FEE=170000

# Calculate change (return amount)
CHANGE=$((UTXO_AMOUNT - FEE))

# Build raw transaction (works offline, no node needed!)
cardano-cli transaction build-raw \
  --tx-in "${UTXO_HASH}#${UTXO_INDEX}" \
  --tx-out "${WALLET_ADDRESS}+${CHANGE}" \
  --metadata-json-file metadata.json \
  --fee ${FEE} \
  --out-file tx.raw

# Sign transaction
cardano-cli transaction sign \
  $MAGIC \
  --tx-body-file tx.raw \
  --signing-key-file payment.skey \
  --out-file tx.signed

# Output the signed transaction as hex for Blockfrost submission
xxd -p -c 1000000 tx.signed | tr -d '\n'

# Cleanup
cd /
rm -rf $WORK_DIR    