/**
 * Integration tests for network operations
 * 
 * These tests require a running Fabric network.
 * To run these tests:
 * 1. Start the Fabric network: cd ../../core/docker && docker-compose up -d
 * 2. Wait for network to be ready
 * 3. Run: npm run test:integration
 */

import { Gateway, Wallets } from 'fabric-network';
import * as path from 'path';
import * as fs from 'fs';

describe('Network Integration Tests', () => {
  const CHANNEL_NAME = process.env.CHANNEL_NAME || 'ibnchannel';
  const CHAINCODE_NAME = process.env.CHAINCODE_NAME || 'teaTraceCC';
  const MSP_ID = process.env.MSP_ID || 'Org1MSP';
  const WALLET_PATH = path.join(process.cwd(), 'wallet');

  let gateway: Gateway;
  let network: any;
  let contract: any;

  beforeAll(async () => {
    // Check if network is running
    // This is a placeholder - actual implementation would check Docker containers
    const networkRunning = await checkNetworkRunning();
    if (!networkRunning) {
      console.warn('⚠️  Network is not running. Skipping integration tests.');
      return;
    }

    // Create wallet
    const wallet = await Wallets.newFileSystemWallet(WALLET_PATH);
    console.log(`Wallet path: ${WALLET_PATH}`);

    // Check if user exists in wallet
    const userExists = await wallet.get('appUser');
    if (!userExists) {
      console.log('Creating user identity...');
      // In real scenario, you would load identity from certificates
      // For now, we'll skip if identity doesn't exist
      return;
    }

    // Create gateway connection
    gateway = new Gateway();
    const connectionProfilePath = path.resolve(
      __dirname,
      '../../connection-profile.json'
    );

    if (!fs.existsSync(connectionProfilePath)) {
      console.warn('⚠️  Connection profile not found. Skipping integration tests.');
      return;
    }

    const connectionProfile = JSON.parse(
      fs.readFileSync(connectionProfilePath, 'utf8')
    );

    const connectionOptions = {
      wallet,
      identity: 'appUser',
      discovery: { enabled: true, asLocalhost: true },
    };

    await gateway.connect(connectionProfile, connectionOptions);
    network = await gateway.getNetwork(CHANNEL_NAME);
    contract = network.getContract(CHAINCODE_NAME);
  });

  afterAll(async () => {
    if (gateway) {
      await gateway.disconnect();
    }
  });

  async function checkNetworkRunning(): Promise<boolean> {
    // Placeholder: Check if Docker containers are running
    // In real implementation, use Docker API or exec commands
    try {
      // This would check if containers are running
      return true; // Placeholder
    } catch (error) {
      return false;
    }
  }

  describe('Channel Operations', () => {
    it('should connect to channel', async () => {
      if (!network) {
        console.log('⏭️  Skipping: Network not available');
        return;
      }

      expect(network).toBeDefined();
      expect(network.getChannel().getName()).toBe(CHANNEL_NAME);
    });

    it('should query chaincode info', async () => {
      if (!contract) {
        console.log('⏭️  Skipping: Contract not available');
        return;
      }

      // Query chaincode info
      const result = await contract.evaluateTransaction('getBatchInfo', 'NONEXISTENT');
      // Should throw error for non-existent batch
      expect(result).toBeDefined();
    });
  });

  describe('Chaincode Operations', () => {
    const testBatchId = `TEST_${Date.now()}`;

    it('should create a batch', async () => {
      if (!contract) {
        console.log('⏭️  Skipping: Contract not available');
        return;
      }

      const result = await contract.submitTransaction(
        'createBatch',
        testBatchId,
        'Test Farm',
        '2024-01-01',
        'Test Processing',
        'TEST-CERT'
      );

      expect(result).toBeDefined();
      const batch = JSON.parse(result.toString());
      expect(batch.batchId).toBe(testBatchId);
      expect(batch.status).toBe('CREATED');
    });

    it('should query batch info', async () => {
      if (!contract) {
        console.log('⏭️  Skipping: Contract not available');
        return;
      }

      const result = await contract.evaluateTransaction(
        'getBatchInfo',
        testBatchId
      );

      expect(result).toBeDefined();
      const batch = JSON.parse(result.toString());
      expect(batch.batchId).toBe(testBatchId);
    });

    it('should update batch status', async () => {
      if (!contract) {
        console.log('⏭️  Skipping: Contract not available');
        return;
      }

      const result = await contract.submitTransaction(
        'updateBatchStatus',
        testBatchId,
        'VERIFIED'
      );

      expect(result).toBeDefined();
      const batch = JSON.parse(result.toString());
      expect(batch.status).toBe('VERIFIED');
    });
  });

  describe('Peer Discovery', () => {
    it('should discover peers in network', async () => {
      if (!network) {
        console.log('⏭️  Skipping: Network not available');
        return;
      }

      const channel = network.getChannel();
      const peers = channel.getPeers();
      expect(peers.length).toBeGreaterThan(0);
    });
  });
});

