const express = require('express');
const { BlockfrostProvider } = require('@meshsdk/core');
const { exec } = require('child_process');
const fs = require('fs').promises;
const path = require('path');
const { promisify } = require('util');

const execAsync = promisify(exec);

const app = express();
app.use(express.json());

app.get('/health', (req, res) => {
  res.json({ status: 'ok' });
});

app.post('/build-transaction', async (req, res) => {
  let tempDir = null;
  
  try {
    const {
      network,
      walletAddress,
      privateKeyHex,
      merkleRoot,
      universityName,
      certCount,
      blockfrostApiKey
    } = req.body;

    console.log('═══════════════════════════════════');
    console.log('🔨 Building with cardano-cli');
    console.log('═══════════════════════════════════');
    console.log('Network:', network);
    console.log('Address:', walletAddress);
    console.log('University:', universityName);
    console.log('Cert Count:', certCount);

    if (!walletAddress || !privateKeyHex || !merkleRoot || !blockfrostApiKey) {
      throw new Error('Missing required fields');
    }

    // Create temp directory
    tempDir = path.join('/tmp', `cardano-tx-${Date.now()}`);
    await fs.mkdir(tempDir, { recursive: true });
    console.log('📁 Temp dir:', tempDir);

    // Fetch UTXOs using Blockfrost
    console.log('\n📦 Fetching UTXOs from Blockfrost...');
    const provider = new BlockfrostProvider(blockfrostApiKey);
    const utxos = await provider.fetchAddressUTxOs(walletAddress);

    if (!utxos || utxos.length === 0) {
      throw new Error('No UTXOs found for this address');
    }

    console.log(`✅ Found ${utxos.length} UTXOs`);

    // Select UTXO with sufficient balance
    let selectedUtxo = null;
    let utxoAmount = 0;

    for (const utxo of utxos) {
      const lovelaceAsset = utxo.output.amount.find(a => a.unit === 'lovelace');
      if (lovelaceAsset) {
        const amount = parseInt(lovelaceAsset.quantity);
        if (amount > 2000000) {
          selectedUtxo = utxo;
          utxoAmount = amount;
          break;
        }
      }
    }

    if (!selectedUtxo) {
      throw new Error('No UTXO with sufficient balance (need > 2 ADA)');
    }

    console.log('✅ Selected UTXO:');
    console.log('   TX Hash:', selectedUtxo.input.txHash.substring(0, 20) + '...');
    console.log('   Index:', selectedUtxo.input.outputIndex);
    console.log('   Amount:', utxoAmount, 'lovelace (', (utxoAmount / 1000000).toFixed(2), 'ADA)');

    // Calculate fee and change
    const fee = 200000;
    const change = utxoAmount - fee;

    if (change < 1000000) {
      throw new Error(`Insufficient funds. Have: ${(utxoAmount/1000000).toFixed(2)} ADA, Need: 1.2 ADA`);
    }

    console.log('\n💵 Fee:', fee, 'lovelace (', (fee / 1000000).toFixed(2), 'ADA)');
    console.log('   Change:', change, 'lovelace (', (change / 1000000).toFixed(2), 'ADA)');

    // Prepare metadata
    const merkleRoot1 = merkleRoot.substring(0, 32);
    const merkleRoot2 = merkleRoot.substring(32, 64);

    const metadata = {
      "674": {
        "msg": [
          `Uni: ${universityName}`,
          `Root1: ${merkleRoot1}`,
          `Root2: ${merkleRoot2}`,
          `Certs: ${certCount}`,
          `Time: ${Math.floor(Date.now() / 1000)}`
        ]
      }
    };

    console.log('\n📝 Metadata prepared');

    // Write metadata file
    const metadataFile = path.join(tempDir, 'metadata.json');
    await fs.writeFile(metadataFile, JSON.stringify(metadata, null, 2));

    // Write signing key file
    const keyFile = path.join(tempDir, 'payment.skey');
    
    let seedHex;
    if (privateKeyHex.length === 128) {
      seedHex = privateKeyHex.substring(0, 64);
    } else if (privateKeyHex.length === 64) {
      seedHex = privateKeyHex;
    } else {
      throw new Error(`Invalid private key length: ${privateKeyHex.length} chars (expected 64 or 128)`);
    }

    console.log('\n🔑 Key length:', privateKeyHex.length, 'hex chars');
    console.log('   Using seed:', seedHex.substring(0, 20) + '...');

    const keyJSON = {
      "type": "PaymentSigningKeyShelley_ed25519",
      "description": "Payment Signing Key",
      "cborHex": `5820${seedHex}`
    };

    await fs.writeFile(keyFile, JSON.stringify(keyJSON, null, 2));
    console.log('✅ Key file created');

    // Determine network magic
    const magic = network === 'mainnet' ? '--mainnet' : 
                  network === 'preprod' ? '--testnet-magic 1' : 
                  '--testnet-magic 2';

    // VERIFY KEY MATCHES ADDRESS
    console.log('\n🔍 Verifying key matches address...');
    
    try {
      // Generate verification key
      const vkeyFile = path.join(tempDir, 'payment.vkey');
      const vkeyCmd = `cardano-cli key verification-key --signing-key-file "${keyFile}" --verification-key-file "${vkeyFile}"`;
      await execAsync(vkeyCmd);
      
      // Generate address from verification key
      const addressFile = path.join(tempDir, 'payment.addr');
      const addressCmd = `cardano-cli address build --payment-verification-key-file "${vkeyFile}" ${magic} --out-file "${addressFile}"`;
      await execAsync(addressCmd);
      
      // Read generated address
      const generatedAddress = (await fs.readFile(addressFile, 'utf8')).trim();
      
      console.log('   Expected address:', walletAddress);
      console.log('   Generated address:', generatedAddress);
      
      if (generatedAddress === walletAddress) {
        console.log('✅ KEY MATCHES ADDRESS!');
      } else {
        console.log('❌ KEY MISMATCH!');
        console.log('   The private key does NOT generate this address!');
        console.log('   This means the wallet was generated with a different key.');
        throw new Error(`Key/Address mismatch! Expected ${walletAddress}, key generates ${generatedAddress}`);
      }
    } catch (err) {
      console.error('❌ Address verification failed:', err.message);
      throw new Error(`CRITICAL: ${err.message}. You need to re-register this university with a new wallet!`);
    }

    // Build raw transaction
    console.log('\n🔨 Building transaction with cardano-cli...');
    const txRawFile = path.join(tempDir, 'tx.raw');
    
    const buildCmd = `cardano-cli transaction build-raw \
      --tx-in "${selectedUtxo.input.txHash}#${selectedUtxo.input.outputIndex}" \
      --tx-out "${walletAddress}+${change}" \
      --metadata-json-file "${metadataFile}" \
      --fee ${fee} \
      --out-file "${txRawFile}"`;

    await execAsync(buildCmd);
    console.log('✅ Transaction body built');

    // Sign transaction
    console.log('\n✍️  Signing transaction with cardano-cli...');
    const txSignedFile = path.join(tempDir, 'tx.signed');
    
    const signCmd = `cardano-cli transaction sign \
      ${magic} \
      --tx-body-file "${txRawFile}" \
      --signing-key-file "${keyFile}" \
      --out-file "${txSignedFile}"`;

    await execAsync(signCmd);
    console.log('✅ Transaction signed');

    // Read signed transaction
    const signedTxEnvelope = await fs.readFile(txSignedFile, 'utf8');
    const envelope = JSON.parse(signedTxEnvelope);
    const signedTxHex = envelope.cborHex;

    if (!signedTxHex) {
      throw new Error('No cborHex in signed transaction file');
    }

    console.log('\n📦 Transaction CBOR:');
    console.log('   Length:', signedTxHex.length, 'characters');
    console.log('   Size:', Math.ceil(signedTxHex.length / 2), 'bytes');

    console.log('\n═══════════════════════════════════');
    console.log('✅ SUCCESS!');
    console.log('═══════════════════════════════════\n');

    res.json({
      success: true,
      txCbor: signedTxHex,
      txSize: Math.ceil(signedTxHex.length / 2)
    });

  } catch (error) {
    console.error('\n═══════════════════════════════════');
    console.error('❌ ERROR:', error.message);
    console.error('═══════════════════════════════════');
    
    if (error.stdout) console.error('stdout:', error.stdout);
    if (error.stderr) console.error('stderr:', error.stderr);
    
    console.error('\n');
    
    res.status(500).json({
      success: false,
      error: error.message,
      details: error.stderr || error.stdout
    });
  } finally {
    // Cleanup temp files
    if (tempDir) {
      try {
        await fs.rm(tempDir, { recursive: true, force: true });
        console.log('🧹 Cleaned up temp files\n');
      } catch (err) {
        console.error('⚠️  Cleanup error:', err.message);
      }
    }
  }
});

const PORT = process.env.PORT || 3001;
app.listen(PORT, () => {
  console.log('═══════════════════════════════════');
  console.log('🚀 Cardano Transaction Builder');
  console.log('═══════════════════════════════════');
  console.log(`📍 Port: ${PORT}`);
  console.log('🔧 Method: cardano-cli (simple & reliable)');
  console.log('📦 UTXO Fetcher: Blockfrost via MeshJS');
  console.log('🔍 Includes: Key/Address verification');
  console.log('✅ Ready to build and sign transactions');
  console.log('═══════════════════════════════════\n');
});