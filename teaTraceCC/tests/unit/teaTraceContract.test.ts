import { Context, ChaincodeStub } from 'fabric-contract-api';
import { TeaTraceContract } from '../../src/teaTraceContract';
import { TeaBatch, TeaBatchStatus } from '../../src/models/teaBatch';
import { generateBatchHash } from '../../src/utils/hashUtils';

// Mock fabric-contract-api
jest.mock('fabric-contract-api');

describe('TeaTraceContract', () => {
  let contract: TeaTraceContract;
  let mockContext: jest.Mocked<Context>;
  let mockStub: jest.Mocked<ChaincodeStub>;
  let mockClientIdentity: any;

  const mockMSPConfig = {
    FARMER: 'Org1MSP',
    VERIFIER: 'Org2MSP',
    ADMIN: 'Org3MSP',
  };

  beforeEach(() => {
    // Reset mocks
    jest.clearAllMocks();

    // Mock ChaincodeStub
    mockStub = {
      getState: jest.fn(),
      putState: jest.fn(),
      getTxTimestamp: jest.fn(),
    } as any;

    // Mock ClientIdentity
    mockClientIdentity = {
      getMSPID: jest.fn().mockReturnValue('Org1MSP'),
      getAttributeValue: jest.fn().mockReturnValue(null),
    };

    // Mock Context
    mockContext = {
      stub: mockStub,
      clientIdentity: mockClientIdentity,
    } as any;

    // Create contract instance
    contract = new TeaTraceContract();

    // Mock getMSPIds
    jest.doMock('../../src/models/mspConfig', () => ({
      getMSPIds: () => mockMSPConfig,
    }));
  });

  describe('createBatch', () => {
    const batchId = 'BATCH001';
    const farmLocation = 'Moc Chau, Son La';
    const harvestDate = '2024-11-08';
    const processingInfo = 'Organic processing';
    const qualityCert = 'VN-ORG-2024';

    it('should create a new batch successfully', async () => {
      // Arrange
      mockStub.getState.mockResolvedValue(Buffer.from(''));
      mockStub.getTxTimestamp.mockReturnValue({
        seconds: { toNumber: () => 1700000000 },
        nanos: 0,
      });
      mockClientIdentity.getMSPID.mockReturnValue('Org1MSP');

      // Act
      const result = await contract.createBatch(
        mockContext,
        batchId,
        farmLocation,
        harvestDate,
        processingInfo,
        qualityCert
      );

      // Assert
      expect(result).toBeDefined();
      expect(result.batchId).toBe(batchId);
      expect(result.farmLocation).toBe(farmLocation);
      expect(result.harvestDate).toBe(harvestDate);
      expect(result.processingInfo).toBe(processingInfo);
      expect(result.qualityCert).toBe(qualityCert);
      expect(result.status).toBe('CREATED');
      expect(result.hashValue).toBeDefined();
      expect(result.owner).toBe('Org1MSP');
      expect(result.timestamp).toBeDefined();

      expect(mockStub.putState).toHaveBeenCalledWith(
        batchId,
        expect.any(Buffer)
      );
    });

    it('should throw error if batch already exists', async () => {
      // Arrange
      const existingBatch: TeaBatch = {
        batchId,
        farmLocation,
        harvestDate,
        processingInfo,
        qualityCert,
        hashValue: 'hash',
        owner: 'Org1MSP',
        timestamp: '2024-11-08T10:00:00.000Z',
        status: 'CREATED',
      };
      mockStub.getState.mockResolvedValue(
        Buffer.from(JSON.stringify(existingBatch))
      );
      mockClientIdentity.getMSPID.mockReturnValue('Org1MSP');

      // Act & Assert
      await expect(
        contract.createBatch(
          mockContext,
          batchId,
          farmLocation,
          harvestDate,
          processingInfo,
          qualityCert
        )
      ).rejects.toThrow('already exists');
    });

    it('should throw error if MSP is not authorized', async () => {
      // Arrange
      mockStub.getState.mockResolvedValue(Buffer.from(''));
      mockClientIdentity.getMSPID.mockReturnValue('UnauthorizedMSP');

      // Act & Assert
      await expect(
        contract.createBatch(
          mockContext,
          batchId,
          farmLocation,
          harvestDate,
          processingInfo,
          qualityCert
        )
      ).rejects.toThrow('not authorized');
    });

    it('should use owner attribute if available', async () => {
      // Arrange
      mockStub.getState.mockResolvedValue(Buffer.from(''));
      mockStub.getTxTimestamp.mockReturnValue({
        seconds: { toNumber: () => 1700000000 },
        nanos: 0,
      });
      mockClientIdentity.getMSPID.mockReturnValue('Org1MSP');
      mockClientIdentity.getAttributeValue.mockImplementation((key: string) => {
        if (key === 'owner') return 'CustomOwner';
        return null;
      });

      // Act
      const result = await contract.createBatch(
        mockContext,
        batchId,
        farmLocation,
        harvestDate,
        processingInfo,
        qualityCert
      );

      // Assert
      expect(result.owner).toBe('CustomOwner');
    });
  });

  describe('verifyBatch', () => {
    const batchId = 'BATCH001';
    const batch: TeaBatch = {
      batchId,
      farmLocation: 'Moc Chau',
      harvestDate: '2024-11-08',
      processingInfo: 'Organic',
      qualityCert: 'VN-ORG-2024',
      hashValue: generateBatchHash({
        batchId,
        farmLocation: 'Moc Chau',
        harvestDate: '2024-11-08',
        processingInfo: 'Organic',
        qualityCert: 'VN-ORG-2024',
      }),
      owner: 'Org1MSP',
      timestamp: '2024-11-08T10:00:00.000Z',
      status: 'CREATED',
    };

    it('should verify batch with correct hash', async () => {
      // Arrange
      mockStub.getState.mockResolvedValue(
        Buffer.from(JSON.stringify(batch))
      );
      mockClientIdentity.getMSPID.mockReturnValue('Org2MSP');
      const correctHash = batch.hashValue;

      // Act
      const result = await contract.verifyBatch(
        mockContext,
        batchId,
        correctHash
      );

      // Assert
      expect(result.isValid).toBe(true);
      expect(result.batch.status).toBe('VERIFIED');
      expect(mockStub.putState).toHaveBeenCalled();
    });

    it('should return false for incorrect hash', async () => {
      // Arrange
      mockStub.getState.mockResolvedValue(
        Buffer.from(JSON.stringify(batch))
      );
      mockClientIdentity.getMSPID.mockReturnValue('Org2MSP');
      const incorrectHash = 'wrong-hash';

      // Act
      const result = await contract.verifyBatch(
        mockContext,
        batchId,
        incorrectHash
      );

      // Assert
      expect(result.isValid).toBe(false);
      expect(result.batch.status).toBe('CREATED');
    });

    it('should throw error if batch does not exist', async () => {
      // Arrange
      mockStub.getState.mockResolvedValue(Buffer.from(''));
      mockClientIdentity.getMSPID.mockReturnValue('Org2MSP');

      // Act & Assert
      await expect(
        contract.verifyBatch(mockContext, batchId, 'hash')
      ).rejects.toThrow('does not exist');
    });

    it('should allow VERIFIER MSP to verify', async () => {
      // Arrange
      mockStub.getState.mockResolvedValue(
        Buffer.from(JSON.stringify(batch))
      );
      mockClientIdentity.getMSPID.mockReturnValue('Org2MSP');

      // Act
      const result = await contract.verifyBatch(
        mockContext,
        batchId,
        batch.hashValue
      );

      // Assert
      expect(result).toBeDefined();
    });

    it('should allow ADMIN MSP to verify', async () => {
      // Arrange
      mockStub.getState.mockResolvedValue(
        Buffer.from(JSON.stringify(batch))
      );
      mockClientIdentity.getMSPID.mockReturnValue('Org3MSP');

      // Act
      const result = await contract.verifyBatch(
        mockContext,
        batchId,
        batch.hashValue
      );

      // Assert
      expect(result).toBeDefined();
    });
  });

  describe('getBatchInfo', () => {
    const batchId = 'BATCH001';
    const batch: TeaBatch = {
      batchId,
      farmLocation: 'Moc Chau',
      harvestDate: '2024-11-08',
      processingInfo: 'Organic',
      qualityCert: 'VN-ORG-2024',
      hashValue: 'hash',
      owner: 'Org1MSP',
      timestamp: '2024-11-08T10:00:00.000Z',
      status: 'CREATED',
    };

    it('should return batch information', async () => {
      // Arrange
      mockStub.getState.mockResolvedValue(
        Buffer.from(JSON.stringify(batch))
      );

      // Act
      const result = await contract.getBatchInfo(mockContext, batchId);

      // Assert
      expect(result).toEqual(batch);
    });

    it('should throw error if batch does not exist', async () => {
      // Arrange
      mockStub.getState.mockResolvedValue(Buffer.from(''));

      // Act & Assert
      await expect(
        contract.getBatchInfo(mockContext, batchId)
      ).rejects.toThrow('does not exist');
    });
  });

  describe('updateBatchStatus', () => {
    const batchId = 'BATCH001';
    const batch: TeaBatch = {
      batchId,
      farmLocation: 'Moc Chau',
      harvestDate: '2024-11-08',
      processingInfo: 'Organic',
      qualityCert: 'VN-ORG-2024',
      hashValue: 'hash',
      owner: 'Org1MSP',
      timestamp: '2024-11-08T10:00:00.000Z',
      status: 'CREATED',
    };

    it('should update batch status to VERIFIED', async () => {
      // Arrange
      mockStub.getState.mockResolvedValue(
        Buffer.from(JSON.stringify(batch))
      );
      mockStub.getTxTimestamp.mockReturnValue({
        seconds: { toNumber: () => 1700000000 },
        nanos: 0,
      });
      mockClientIdentity.getMSPID.mockReturnValue('Org1MSP');

      // Act
      const result = await contract.updateBatchStatus(
        mockContext,
        batchId,
        'VERIFIED'
      );

      // Assert
      expect(result.status).toBe('VERIFIED');
      expect(mockStub.putState).toHaveBeenCalled();
    });

    it('should update batch status to EXPIRED', async () => {
      // Arrange
      mockStub.getState.mockResolvedValue(
        Buffer.from(JSON.stringify(batch))
      );
      mockStub.getTxTimestamp.mockReturnValue({
        seconds: { toNumber: () => 1700000000 },
        nanos: 0,
      });
      mockClientIdentity.getMSPID.mockReturnValue('Org1MSP');

      // Act
      const result = await contract.updateBatchStatus(
        mockContext,
        batchId,
        'expired'
      );

      // Assert
      expect(result.status).toBe('EXPIRED');
    });

    it('should throw error for invalid status', async () => {
      // Arrange
      mockStub.getState.mockResolvedValue(
        Buffer.from(JSON.stringify(batch))
      );
      mockClientIdentity.getMSPID.mockReturnValue('Org1MSP');

      // Act & Assert
      await expect(
        contract.updateBatchStatus(mockContext, batchId, 'INVALID')
      ).rejects.toThrow('Invalid status');
    });

    it('should throw error if MSP is not authorized', async () => {
      // Arrange
      mockStub.getState.mockResolvedValue(
        Buffer.from(JSON.stringify(batch))
      );
      mockClientIdentity.getMSPID.mockReturnValue('UnauthorizedMSP');

      // Act & Assert
      await expect(
        contract.updateBatchStatus(mockContext, batchId, 'VERIFIED')
      ).rejects.toThrow('not authorized');
    });

    it('should throw error if batch does not exist', async () => {
      // Arrange
      mockStub.getState.mockResolvedValue(Buffer.from(''));
      mockClientIdentity.getMSPID.mockReturnValue('Org1MSP');

      // Act & Assert
      await expect(
        contract.updateBatchStatus(mockContext, batchId, 'VERIFIED')
      ).rejects.toThrow('does not exist');
    });
  });
});

