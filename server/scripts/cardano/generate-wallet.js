const { AppWallet } = require('@meshsdk/core');

async function generateWallet() {
    try {
        // Generate 24-word mnemonic
        const mnemonic = AppWallet.brew(256);
        
        // Create wallet
        const wallet = new AppWallet({
            networkId: 0, // 0 = testnet (preprod)
            key: {
                type: 'mnemonic',
                words: mnemonic
            }
        });
        
        // Get address and keys
        const address = await wallet.getPaymentAddress();
        const privateKey = await wallet.getPrivateKey();
        
        // Output as JSON
        const result = {
            mnemonic: mnemonic,
            address: address,
            privateKey: privateKey
        };
        
        console.log(JSON.stringify(result));
    } catch (error) {
        console.error('Error:', error.message);
        process.exit(1);
    }
}

generateWallet();
