const bip39 = require('bip39');

async function generateWallet() {
    try {
        // Check if we need JSON output (called by Go)
        const outputJSON = process.argv.includes('--json');
        
        if (!outputJSON) {
            console.log('═══════════════════════════════════════');
            console.log('🎉 GENERATING WALLET');
            console.log('═══════════════════════════════════════\n');
        }
        
        // Generate 24-word mnemonic
        const mnemonic = bip39.generateMnemonic(256);
        
        // Create deterministic keys from mnemonic
        const seed = bip39.mnemonicToSeedSync(mnemonic);
        const privateKey = seed.slice(0, 32).toString('hex');
        const publicKey = seed.slice(32, 64).toString('hex');
        
        // Create mock testnet address
        const addressSuffix = seed.slice(64, 92).toString('hex');
        const address = 'addr_test1' + addressSuffix;
        
        if (outputJSON) {
            // JSON output for Go
            const result = {
                mnemonic: mnemonic,
                address: address,
                privateKey: privateKey,
                publicKey: publicKey
            };
            console.log(JSON.stringify(result));
        } else {
            // Human-readable output
            console.log('📝 Recovery Phrase (24 words):');
            console.log(mnemonic);
            console.log('\n⚠️  SAVE THESE WORDS!\n');
            console.log('💳 Wallet Address (PreProd Testnet):');
            console.log(address);
            console.log('\n🔑 Private Key:');
            console.log(privateKey);
            console.log('\n═══════════════════════════════════════\n');
        }
        
    } catch (error) {
        console.error('Error:', error.message);
        process.exit(1);
    }
}

generateWallet();