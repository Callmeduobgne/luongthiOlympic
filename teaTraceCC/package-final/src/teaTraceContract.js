"use strict";

var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.contracts = exports.TeaTraceContract = void 0;
const fabric_contract_api_1 = require("fabric-contract-api");
const crypto_1 = __importDefault(require("crypto"));
const teaBatch_1 = require("./models/teaBatch");
const teaPackage_1 = require("./models/teaPackage");
const hashUtils_1 = require("./utils/hashUtils");
const mspConfig_1 = require("./models/mspConfig");
const validation_1 = require("./utils/validation");
// Load MSP configuration
const MSP_CONFIG = (0, mspConfig_1.getMSPIds)();
/**
 * TeaTraceContract - Professional Chaincode Implementation
 *
 * Data Storage Strategy:
 * 1. CouchDB Indexes: All queries use CouchDB rich queries with pre-defined indexes
 *    - Indexes are defined in META-INF/statedb/couchdb/indexes/
 *    - Indexes: indexBatchStatus, indexBatchOwner, indexPackageBatch, indexPackageStatus, indexPackageOwner
 *
 * 2. Composite Keys: Packages use composite keys for efficient batch queries
 *    - Simple key: packageId (for direct access and backward compatibility)
 *    - Composite key: PACKAGE~batchId~packageId (for efficient getPackagesByBatch queries)
 *
 * 3. Document Type: All documents include docType field for CouchDB indexing
 *    - Batches: docType: "batch"
 *    - Packages: docType: "package"
 *
 * Performance Optimizations:
 * - getAllBatches/getBatchesByStatus: Uses CouchDB rich queries with indexes (O(log n))
 * - getPackagesByBatch: Uses composite key queries (O(k) where k = packages in batch)
 * - getPackagesByStatus: Uses CouchDB rich queries with indexes (O(log n))
 * - All queries support pagination with proper offset/limit handling
 *
 * Backward Compatibility:
 * - Packages are stored with both simple key (packageId) and composite key
 * - Direct access via packageId still works
 * - Existing data without docType will still work (but queries may be slower)
 */
class TeaTraceContract extends fabric_contract_api_1.Contract {
    constructor() {
        super("teaTraceContract");
    }
    async createBatch(ctx, batchId, farmLocation, harvestDate, processingInfo, qualityCert) {
        // Input validation
        (0, validation_1.validateBatchId)(batchId);
        (0, validation_1.validateString)(farmLocation, "Farm location", 200);
        (0, validation_1.validateDate)(harvestDate);
        (0, validation_1.validateString)(processingInfo, "Processing info", 1000);
        (0, validation_1.validateString)(qualityCert, "Quality certificate", 100);
        this.ensureOrg(ctx, [MSP_CONFIG.FARMER], "create batches");
        await this.assertBatchDoesNotExist(ctx, batchId);
        const owner = ctx.clientIdentity.getAttributeValue("owner") ||
            ctx.clientIdentity.getAttributeValue("organization") ||
            ctx.clientIdentity.getMSPID();
        const input = {
            batchId,
            farmLocation,
            harvestDate,
            processingInfo,
            qualityCert
        };
        const hashValue = (0, hashUtils_1.generateBatchHash)(input);
        const batch = {
            docType: "batch", // Required for CouchDB indexing
            ...input,
            hashValue,
            owner,
            timestamp: this.getCurrentTimestamp(ctx),
            status: "CREATED"
        };
        await ctx.stub.putState(batch.batchId, Buffer.from(JSON.stringify(batch)));
        return batch;
    }
    async verifyBatch(ctx, batchId, hashInput) {
        // Input validation
        (0, validation_1.validateBatchId)(batchId);
        if (!hashInput || hashInput.trim().length === 0) {
            throw new Error("Hash input cannot be empty");
        }
        this.ensureOrg(ctx, [MSP_CONFIG.VERIFIER, MSP_CONFIG.ADMIN, MSP_CONFIG.FARMER], "verify batches");
        const batch = await this.getBatchOrThrow(ctx, batchId);
        const isValid = (0, hashUtils_1.verifyHash)(batch.hashValue, hashInput);
        if (isValid && batch.status !== "VERIFIED") {
            batch.status = "VERIFIED";
            await ctx.stub.putState(batch.batchId, Buffer.from(JSON.stringify(batch)));
        }
        return { isValid, batch };
    }
    async getBatchInfo(ctx, batchId) {
        (0, validation_1.validateBatchId)(batchId);
        const buffer = await ctx.stub.getState(batchId);
        if (!buffer || buffer.length === 0) {
            // Return null instead of throwing error for better API compatibility
            return null;
        }
        return JSON.parse(this.bytesToString(buffer));
    }
    /**
     * Get all batches with pagination
     * Uses CouchDB rich query with index for optimal performance
     * Args: [limit?, offset?]
     */
    async getAllBatches(ctx, ...args) {
        const limitStr = args[0] || "100";
        const offsetStr = args[1] || "0";
        const limitNum = parseInt(limitStr, 10);
        const offsetNum = parseInt(offsetStr, 10);
        (0, validation_1.validatePagination)(limitNum, offsetNum);
        // Use CouchDB rich query with index for optimal performance
        const queryString = {
            selector: {
                docType: "batch"
            },
            sort: [{ timestamp: "desc" }],
            limit: limitNum + offsetNum, // Fetch more to handle offset
            skip: 0 // We'll handle offset manually to get accurate total
        };
        // Use CouchDB rich query with index for optimal performance
        // getQueryResult returns StateQueryIterator which supports async iteration
        const iterator = await ctx.stub.getQueryResult(JSON.stringify(queryString));
        const batches = [];
        let total = 0;
        let skipped = 0;
        try {
            // Use async iteration pattern compatible with fabric-contract-api
            while (true) {
                const result = await iterator.next();
                if (result.done) {
                    break;
                }
                total++;
                // Handle offset
                if (skipped < offsetNum) {
                    skipped++;
                    continue;
                }
                // Apply limit
                if (batches.length < limitNum) {
                    const batch = JSON.parse(this.bytesToString(result.value.value));
                    batches.push(batch);
                }
            }
        }
        finally {
            await iterator.close();
        }
        return { batches, total };
    }
    /**
     * Get batches by status
     * Uses CouchDB rich query with indexBatchStatus for optimal performance
     * Args: [status, limit?, offset?]
     */
    async getBatchesByStatus(ctx, status, ...args) {
        const normalizedStatus = status.toUpperCase();
        if (!(0, teaBatch_1.isTeaBatchStatus)(normalizedStatus)) {
            throw new Error(`Invalid status '${status}'. Allowed values: CREATED, VERIFIED, EXPIRED.`);
        }
        const limitStr = args[0] || "100";
        const offsetStr = args[1] || "0";
        const limitNum = parseInt(limitStr, 10);
        const offsetNum = parseInt(offsetStr, 10);
        (0, validation_1.validatePagination)(limitNum, offsetNum);
        // Use CouchDB rich query with indexBatchStatus index
        const queryString = {
            selector: {
                docType: "batch",
                status: normalizedStatus
            },
            sort: [{ timestamp: "desc" }],
            limit: limitNum + offsetNum, // Fetch more to handle offset
            skip: 0 // We'll handle offset manually to get accurate total
        };
        const iterator = await ctx.stub.getQueryResult(JSON.stringify(queryString));
        const batches = [];
        let total = 0;
        let skipped = 0;
        try {
            while (true) {
                const result = await iterator.next();
                if (result.done) {
                    break;
                }
                total++;
                // Handle offset
                if (skipped < offsetNum) {
                    skipped++;
                    continue;
                }
                // Apply limit
                if (batches.length < limitNum) {
                    const batch = JSON.parse(this.bytesToString(result.value.value));
                    batches.push(batch);
                }
            }
        }
        finally {
            await iterator.close();
        }
        return { batches, total };
    }
    /**
     * Get batches by owner
     * Uses CouchDB rich query with indexBatchOwner for optimal performance
     * Args: [owner, limit?, offset?]
     */
    async getBatchesByOwner(ctx, owner, ...args) {
        if (!owner || owner.trim().length === 0) {
            throw new Error("Owner cannot be empty");
        }
        const limitStr = args[0] || "100";
        const offsetStr = args[1] || "0";
        const limitNum = parseInt(limitStr, 10);
        const offsetNum = parseInt(offsetStr, 10);
        (0, validation_1.validatePagination)(limitNum, offsetNum);
        // Use CouchDB rich query with indexBatchOwner index
        const queryString = {
            selector: {
                docType: "batch",
                owner: owner
            },
            sort: [{ timestamp: "desc" }],
            limit: limitNum + offsetNum, // Fetch more to handle offset
            skip: 0 // We'll handle offset manually to get accurate total
        };
        const iterator = await ctx.stub.getQueryResult(JSON.stringify(queryString));
        const batches = [];
        let total = 0;
        let skipped = 0;
        try {
            while (true) {
                const result = await iterator.next();
                if (result.done) {
                    break;
                }
                total++;
                // Handle offset
                if (skipped < offsetNum) {
                    skipped++;
                    continue;
                }
                // Apply limit
                if (batches.length < limitNum) {
                    const batch = JSON.parse(this.bytesToString(result.value.value));
                    batches.push(batch);
                }
            }
        }
        finally {
            await iterator.close();
        }
        return { batches, total };
    }
    /**
     * Get batch history (all changes)
     */
    async getBatchHistory(ctx, batchId) {
        (0, validation_1.validateBatchId)(batchId);
        const historyIterator = await ctx.stub.getHistoryForKey(batchId);
        const history = [];
        while (true) {
            const result = await historyIterator.next();
            if (result.done) {
                await historyIterator.close();
                break;
            }
            if (result.value.isDelete) {
                continue;
            }
            const batch = JSON.parse(this.bytesToString(result.value.value));
            history.push(batch);
        }
        return history.reverse(); // Oldest first
    }
    async updateBatchStatus(ctx, batchId, status) {
        // Input validation
        (0, validation_1.validateBatchId)(batchId);
        if (!status || status.trim().length === 0) {
            throw new Error("Status cannot be empty");
        }
        this.ensureOrg(ctx, [MSP_CONFIG.FARMER, MSP_CONFIG.ADMIN], "update batch status");
        const normalizedStatus = status.toUpperCase();
        if (!(0, teaBatch_1.isTeaBatchStatus)(normalizedStatus)) {
            throw new Error(`Invalid status '${status}'. Allowed values: CREATED, VERIFIED, EXPIRED.`);
        }
        const batch = await this.getBatchOrThrow(ctx, batchId);
        batch.status = normalizedStatus;
        batch.timestamp = this.getCurrentTimestamp(ctx);
        await ctx.stub.putState(batch.batchId, Buffer.from(JSON.stringify(batch)));
        return batch;
    }
    // ==================== Package Functions ====================
    /**
     * Create a new tea package from a batch
     * Args: [packageId, batchId, weight, productionDate, expiryDate?, qrCode?]
     */
    async createPackage(ctx, packageId, batchId, weightStr, productionDate, expiryDate, qrCode) {
        // Input validation
        (0, validation_1.validatePackageId)(packageId);
        (0, validation_1.validateBatchId)(batchId);
        const weight = parseFloat(weightStr);
        if (isNaN(weight)) {
            throw new Error("Weight must be a valid number");
        }
        (0, validation_1.validateWeight)(weight);
        (0, validation_1.validateDate)(productionDate);
        if (expiryDate) {
            (0, validation_1.validateDate)(expiryDate);
            (0, validation_1.validateDateRange)(productionDate, expiryDate);
        }
        if (qrCode && qrCode.length > 500) {
            throw new Error("QR code must be less than 500 characters");
        }
        // Check authorization (Farmer, Admin can create packages)
        this.ensureOrg(ctx, [MSP_CONFIG.FARMER, MSP_CONFIG.ADMIN], "create packages");
        // Verify batch exists
        await this.getBatchOrThrow(ctx, batchId);
        // Check package doesn't exist
        await this.assertPackageDoesNotExist(ctx, packageId);
        // Get transaction ID
        const txId = ctx.stub.getTxID();
        // Get hash secret from environment (optional for backward compatibility)
        // If not set, will generate v1 hash (without secret)
        const hashSecret = process.env.HASH_SECRET || "";
        // Generate blockHash identifier from package data + txId + secret (if available)
        const blockHash = this.generatePackageBlockHash(packageId, batchId, weight, productionDate, txId, hashSecret // Pass secret for v2 hash, or empty string for v1 hash
        );
        const owner = ctx.clientIdentity.getAttributeValue("owner") ||
            ctx.clientIdentity.getAttributeValue("organization") ||
            ctx.clientIdentity.getMSPID();
        const pkg = {
            docType: "package", // Required for CouchDB indexing
            packageId,
            batchId,
            blockHash,
            hashVersion: hashSecret ? "v2" : "v1", // Track hash format for verification
            txId,
            weight,
            productionDate,
            expiryDate,
            qrCode,
            status: "CREATED",
            owner,
            timestamp: this.getCurrentTimestamp(ctx)
        };
        // Use composite key for efficient querying: PACKAGE~batchId~packageId
        // This allows efficient queries by batchId using getStateByPartialCompositeKey
        const packageKey = ctx.stub.createCompositeKey("PACKAGE", [batchId, packageId]);
        // Store with simple key for backward compatibility and direct access
        await ctx.stub.putState(packageId, Buffer.from(JSON.stringify(pkg)));
        // Store with composite key for efficient batch queries
        await ctx.stub.putState(packageKey, Buffer.from(JSON.stringify(pkg)));
        return pkg;
    }
    /**
     * Verify a package by comparing blockhash
     * Supports both v1 (no secret) and v2 (with secret) hash formats
     * Args: [packageId, blockHash?]
     */
    async verifyPackage(ctx, packageId, providedBlockHash) {
        (0, validation_1.validatePackageId)(packageId);
        // Public function - anyone can verify
        const pkg = await this.getPackageOrThrow(ctx, packageId);
        // If blockHash provided, compare
        if (providedBlockHash) {
            let isValid = false;
            // Check hash version to determine verification method
            const hashVersion = pkg.hashVersion || "v1"; // Default to v1 for backward compatibility
            if (hashVersion === "v2") {
                // V2 hash: Regenerate hash with secret and compare
                const hashSecret = process.env.HASH_SECRET || "";
                const regeneratedHash = this.generatePackageBlockHash(pkg.packageId, pkg.batchId, pkg.weight, pkg.productionDate, pkg.txId, hashSecret);
                isValid = regeneratedHash === providedBlockHash;
            }
            else {
                // V1 hash: Regenerate hash without secret and compare
                const regeneratedHash = this.generatePackageBlockHash(pkg.packageId, pkg.batchId, pkg.weight, pkg.productionDate, pkg.txId
                // No secret for v1
                );
                isValid = regeneratedHash === providedBlockHash;
            }
            // If valid and status is CREATED, update to VERIFIED
            if (isValid && pkg.status === "CREATED") {
                pkg.status = "VERIFIED";
                pkg.timestamp = this.getCurrentTimestamp(ctx);
                // Update both simple key and composite key
                await ctx.stub.putState(packageId, Buffer.from(JSON.stringify(pkg)));
                const packageKey = ctx.stub.createCompositeKey("PACKAGE", [pkg.batchId, packageId]);
                await ctx.stub.putState(packageKey, Buffer.from(JSON.stringify(pkg)));
            }
            return { isValid, package: pkg };
        }
        // If no blockHash provided, just return package info (exists check)
        return { isValid: true, package: pkg };
    }
    /**
     * Get package information
     * Args: [packageId]
     */
    async getPackageInfo(ctx, packageId) {
        (0, validation_1.validatePackageId)(packageId);
        const buffer = await ctx.stub.getState(packageId);
        if (!buffer || buffer.length === 0) {
            return null;
        }
        return JSON.parse(this.bytesToString(buffer));
    }
    /**
     * Get all packages with pagination
     * Uses CouchDB rich query with index for optimal performance
     * Args: [limit?, offset?]
     */
    async getAllPackages(ctx, ...args) {
        const limitStr = args[0] || "100";
        const offsetStr = args[1] || "0";
        const limitNum = parseInt(limitStr, 10);
        const offsetNum = parseInt(offsetStr, 10);
        (0, validation_1.validatePagination)(limitNum, offsetNum);
        // Use CouchDB rich query with index for optimal performance
        const queryString = {
            selector: {
                docType: "package"
            },
            sort: [{ timestamp: "desc" }],
            limit: limitNum + offsetNum, // Fetch more to handle offset
            skip: 0 // We'll handle offset manually to get accurate total
        };
        const iterator = await ctx.stub.getQueryResult(JSON.stringify(queryString));
        const packages = [];
        let total = 0;
        let skipped = 0;
        while (true) {
            const result = await iterator.next();
            if (result.done) {
                await iterator.close();
                break;
            }
            total++;
            // Handle offset
            if (skipped < offsetNum) {
                skipped++;
                continue;
            }
            // Apply limit
            if (packages.length < limitNum) {
                const pkg = JSON.parse(this.bytesToString(result.value.value));
                packages.push(pkg);
            }
        }
        return { packages, total };
    }
    /**
     * Get packages by batch ID
     * Uses composite key for optimal performance: PACKAGE~batchId~packageId
     * Args: [batchId, limit?, offset?]
     */
    async getPackagesByBatch(ctx, batchId, limitStr, offsetStr) {
        (0, validation_1.validateBatchId)(batchId);
        const limit = limitStr || "100";
        const offset = offsetStr || "0";
        const limitNum = parseInt(limit, 10);
        const offsetNum = parseInt(offset, 10);
        (0, validation_1.validatePagination)(limitNum, offsetNum);
        // Verify batch exists
        await this.getBatchOrThrow(ctx, batchId);
        // Use composite key for efficient querying: PACKAGE~batchId~packageId
        // This is much faster than scanning all state
        const iterator = await ctx.stub.getStateByPartialCompositeKey("PACKAGE", [batchId]);
        const packages = [];
        let total = 0;
        let skipped = 0;
        try {
            while (true) {
                const result = await iterator.next();
                if (result.done) {
                    break;
                }
                total++;
                // Handle offset
                if (skipped < offsetNum) {
                    skipped++;
                    continue;
                }
                // Apply limit
                if (packages.length < limitNum) {
                    const pkg = JSON.parse(this.bytesToString(result.value.value));
                    packages.push(pkg);
                }
            }
        }
        finally {
            await iterator.close();
        }
        return { packages, total };
    }
    /**
     * Get packages by status
     * Uses CouchDB rich query with indexPackageStatus for optimal performance
     * Args: [status, limit?, offset?]
     */
    async getPackagesByStatus(ctx, status, ...args) {
        const normalizedStatus = status.toUpperCase();
        if (!(0, teaPackage_1.isTeaPackageStatus)(normalizedStatus)) {
            throw new Error(`Invalid status '${status}'. Allowed values: CREATED, VERIFIED, SOLD, EXPIRED.`);
        }
        const limitStr = args[0] || "100";
        const offsetStr = args[1] || "0";
        const limitNum = parseInt(limitStr, 10);
        const offsetNum = parseInt(offsetStr, 10);
        (0, validation_1.validatePagination)(limitNum, offsetNum);
        // Use CouchDB rich query with indexPackageStatus index
        const queryString = {
            selector: {
                docType: "package",
                status: normalizedStatus
            },
            sort: [{ timestamp: "desc" }],
            limit: limitNum + offsetNum, // Fetch more to handle offset
            skip: 0 // We'll handle offset manually to get accurate total
        };
        const iterator = await ctx.stub.getQueryResult(JSON.stringify(queryString));
        const packages = [];
        let total = 0;
        let skipped = 0;
        try {
            while (true) {
                const result = await iterator.next();
                if (result.done) {
                    break;
                }
                total++;
                // Handle offset
                if (skipped < offsetNum) {
                    skipped++;
                    continue;
                }
                // Apply limit
                if (packages.length < limitNum) {
                    const pkg = JSON.parse(this.bytesToString(result.value.value));
                    packages.push(pkg);
                }
            }
        }
        finally {
            await iterator.close();
        }
        return { packages, total };
    }
    /**
     * Get package history (all changes)
     * Args: [packageId]
     */
    async getPackageHistory(ctx, packageId) {
        (0, validation_1.validatePackageId)(packageId);
        const historyIterator = await ctx.stub.getHistoryForKey(packageId);
        const history = [];
        while (true) {
            const result = await historyIterator.next();
            if (result.done) {
                await historyIterator.close();
                break;
            }
            if (result.value.isDelete) {
                continue;
            }
            const pkg = JSON.parse(this.bytesToString(result.value.value));
            history.push(pkg);
        }
        return history.reverse(); // Oldest first
    }
    /**
     * Update package status
     * Args: [packageId, status]
     */
    async updatePackageStatus(ctx, packageId, status) {
        (0, validation_1.validatePackageId)(packageId);
        if (!status || status.trim().length === 0) {
            throw new Error("Status cannot be empty");
        }
        this.ensureOrg(ctx, [MSP_CONFIG.FARMER, MSP_CONFIG.ADMIN], "update package status");
        const normalizedStatus = status.toUpperCase();
        if (!(0, teaPackage_1.isTeaPackageStatus)(normalizedStatus)) {
            throw new Error(`Invalid status '${status}'. Allowed values: CREATED, VERIFIED, SOLD, EXPIRED.`);
        }
        const pkg = await this.getPackageOrThrow(ctx, packageId);
        pkg.status = normalizedStatus;
        pkg.timestamp = this.getCurrentTimestamp(ctx);
        // Update both simple key and composite key
        await ctx.stub.putState(packageId, Buffer.from(JSON.stringify(pkg)));
        const packageKey = ctx.stub.createCompositeKey("PACKAGE", [pkg.batchId, packageId]);
        await ctx.stub.putState(packageKey, Buffer.from(JSON.stringify(pkg)));
        return pkg;
    }
    ensureOrg(ctx, allowedMsps, action) {
        const callerMsp = ctx.clientIdentity.getMSPID();
        if (!allowedMsps.includes(callerMsp)) {
            throw new Error(`MSP '${callerMsp}' is not authorized to ${action}. Allowed MSPs: ${allowedMsps.join(", ")}`);
        }
    }
    async getBatchOrThrow(ctx, batchId) {
        const buffer = await ctx.stub.getState(batchId);
        if (!buffer || buffer.length === 0) {
            throw new Error(`Batch with id '${batchId}' does not exist.`);
        }
        return JSON.parse(this.bytesToString(buffer));
    }
    async assertBatchDoesNotExist(ctx, batchId) {
        const buffer = await ctx.stub.getState(batchId);
        if (buffer && buffer.length > 0) {
            throw new Error(`Batch with id '${batchId}' already exists.`);
        }
    }
    async getPackageOrThrow(ctx, packageId) {
        const buffer = await ctx.stub.getState(packageId);
        if (!buffer || buffer.length === 0) {
            throw new Error(`Package with id '${packageId}' does not exist.`);
        }
        return JSON.parse(this.bytesToString(buffer));
    }
    async assertPackageDoesNotExist(ctx, packageId) {
        const buffer = await ctx.stub.getState(packageId);
        if (buffer && buffer.length > 0) {
            throw new Error(`Package with id '${packageId}' already exists.`);
        }
    }
    /**
     * Generate a unique blockHash identifier for package
     * This is a composite hash from package data + transaction ID
     * Supports v1 (no secret) and v2 (with secret) hash formats
     *
     * @param secret - Optional secret salt for enhanced security (v2 hash)
     */
    generatePackageBlockHash(packageId, batchId, weight, productionDate, txId, secret) {
        let payload = `${packageId}|${batchId}|${weight}|${productionDate}|${txId}`;
        // If secret provided, append it to payload (v2 hash)
        // This prevents rainbow table attacks while maintaining backward compatibility
        if (secret) {
            payload += `|${secret}`;
        }
        return crypto_1.default.createHash("sha256").update(payload).digest("hex");
    }
    getCurrentTimestamp(ctx) {
        const timestamp = ctx.stub.getTxTimestamp();
        const millis = timestamp.seconds.toNumber() * 1000 + Math.floor(timestamp.nanos / 1000000);
        return new Date(millis).toISOString();
    }
    bytesToString(bytes) {
        return Buffer.from(bytes).toString("utf8");
    }
}
exports.TeaTraceContract = TeaTraceContract;
exports.contracts = [TeaTraceContract];
//# sourceMappingURL=teaTraceContract.js.map