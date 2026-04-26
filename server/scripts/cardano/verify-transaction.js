const axios = require('axios');

async function verifyTransaction() {
    try {
        const [txID, network, apiKey] = process.argv.slice(2);
        
        const response = await axios.get(
            `https://cardano-${network}.blockfrost.io/api/v0/txs/${txID}`,
            {
                headers: {
                    'project_id': apiKey
                }
            }
        );
        
        // Check if transaction exists and is confirmed
        const exists = response.data && response.data.block;
        console.log(exists);
        
    } catch (error) {
        if (error.response && error.response.status === 404) {
            console.log('false');
        } else {
            console.error('Verification error:', error.message);
            process.exit(1);
        }
    }
}

verifyTransaction();
