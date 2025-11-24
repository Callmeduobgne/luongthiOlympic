/**
 * Copyright 2024 IBN Network (ICTU Blockchain Network)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import { Context, Contract } from "fabric-contract-api";
import crypto from "crypto";

import {
  CreateTeaBatchInput,
  TeaBatch,
  TeaBatchStatus,
  isTeaBatchStatus
} from "./models/teaBatch";
import {
  TeaPackage,
  TeaPackageStatus,
  isTeaPackageStatus,
  CreateTeaPackageInput
} from "./models/teaPackage";
import { generateBatchHash, verifyHash } from "./utils/hashUtils";
import { getMSPIds } from "./models/mspConfig";
import {
  validateBatchId,
  validateDate,
  validateString,
  validatePagination,
  validatePackageId,
  validateWeight,
  validateDateRange
} from "./utils/validation";

// Load MSP configuration
const MSP_CONFIG = getMSPIds();

export class TeaTraceContract extends Contract {
  constructor() {
    super("teaTraceContract");
  }

  public async createBatch(
    ctx: Context,
    batchId: string,
    farmLocation: string,
    harvestDate: string,
    processingInfo: string,
    qualityCert: string
  ): Promise<TeaBatch> {
    // Input validation
    validateBatchId(batchId);
    validateString(farmLocation, "Farm location", 200);
    validateDate(harvestDate);
    validateString(processingInfo, "Processing info", 1000);
    validateString(qualityCert, "Quality certificate", 100);

    this.ensureOrg(ctx, [MSP_CONFIG.FARMER], "create batches");

    await this.assertBatchDoesNotExist(ctx, batchId);

    const owner =
      ctx.clientIdentity.getAttributeValue("owner") ||
      ctx.clientIdentity.getAttributeValue("organization") ||
      ctx.clientIdentity.getMSPID();

    const input: CreateTeaBatchInput = {
      batchId,
      farmLocation,
      harvestDate,
      processingInfo,
      qualityCert
    };

    const hashValue = generateBatchHash(input);

    const batch: TeaBatch = {
      ...input,
      hashValue,
      owner,
      timestamp: this.getCurrentTimestamp(ctx),
      status: "CREATED"
    };

    await ctx.stub.putState(batch.batchId, Buffer.from(JSON.stringify(batch)));
    return batch;
  }

  public async verifyBatch(
    ctx: Context,
    batchId: string,
    hashInput: string
  ): Promise<{ isValid: boolean; batch: TeaBatch }> {
    // Input validation
    validateBatchId(batchId);
    if (!hashInput || hashInput.trim().length === 0) {
      throw new Error("Hash input cannot be empty");
    }

    this.ensureOrg(ctx, [MSP_CONFIG.VERIFIER, MSP_CONFIG.ADMIN, MSP_CONFIG.FARMER], "verify batches");

    const batch = await this.getBatchOrThrow(ctx, batchId);
    const isValid = verifyHash(batch.hashValue, hashInput);

    if (isValid && batch.status !== "VERIFIED") {
      batch.status = "VERIFIED";
      await ctx.stub.putState(batch.batchId, Buffer.from(JSON.stringify(batch)));
    }

    return { isValid, batch };
  }

  public async getBatchInfo(ctx: Context, batchId: string): Promise<TeaBatch | null> {
    validateBatchId(batchId);
    const buffer = await ctx.stub.getState(batchId);
    if (!buffer || buffer.length === 0) {
      // Return null instead of throwing error for better API compatibility
      return null;
    }
    return JSON.parse(this.bytesToString(buffer)) as TeaBatch;
  }

  /**
   * Get all batches with pagination
   * Args: [limit?, offset?]
   */
  public async getAllBatches(
    ctx: Context,
    ...args: string[]
  ): Promise<{ batches: TeaBatch[]; total: number }> {
    const limitStr = args[0] || "100";
    const offsetStr = args[1] || "0";
    const limitNum = parseInt(limitStr, 10);
    const offsetNum = parseInt(offsetStr, 10);
    validatePagination(limitNum, offsetNum);

    const iterator = await ctx.stub.getStateByRange("", "");
    const batches: TeaBatch[] = [];
    let total = 0;

    while (true) {
      const result = await iterator.next();
      if (result.done) {
        await iterator.close();
        break;
      }

      total++;
      if (total > offsetNum && batches.length < limitNum) {
        const batch = JSON.parse(this.bytesToString(result.value.value)) as TeaBatch;
        batches.push(batch);
      }
    }

    return { batches, total };
  }

  /**
   * Get batches by status
   * Args: [status, limit?, offset?]
   */
  public async getBatchesByStatus(
    ctx: Context,
    status: string,
    ...args: string[]
  ): Promise<{ batches: TeaBatch[]; total: number }> {
    const normalizedStatus = status.toUpperCase();
    if (!isTeaBatchStatus(normalizedStatus)) {
      throw new Error(
        `Invalid status '${status}'. Allowed values: CREATED, VERIFIED, EXPIRED.`
      );
    }

    const limitStr = args[0] || "100";
    const offsetStr = args[1] || "0";
    const limitNum = parseInt(limitStr, 10);
    const offsetNum = parseInt(offsetStr, 10);
    validatePagination(limitNum, offsetNum);

    const iterator = await ctx.stub.getStateByRange("", "");
    const batches: TeaBatch[] = [];
    let total = 0;

    while (true) {
      const result = await iterator.next();
      if (result.done) {
        await iterator.close();
        break;
      }

      const batch = JSON.parse(this.bytesToString(result.value.value)) as TeaBatch;
      if (batch.status === normalizedStatus) {
        total++;
        if (total > offsetNum && batches.length < limitNum) {
          batches.push(batch);
        }
      }
    }

    return { batches, total };
  }

  /**
   * Get batches by owner
   * Args: [owner, limit?, offset?]
   */
  public async getBatchesByOwner(
    ctx: Context,
    owner: string,
    ...args: string[]
  ): Promise<{ batches: TeaBatch[]; total: number }> {
    if (!owner || owner.trim().length === 0) {
      throw new Error("Owner cannot be empty");
    }

    const limitStr = args[0] || "100";
    const offsetStr = args[1] || "0";
    const limitNum = parseInt(limitStr, 10);
    const offsetNum = parseInt(offsetStr, 10);
    validatePagination(limitNum, offsetNum);

    const iterator = await ctx.stub.getStateByRange("", "");
    const batches: TeaBatch[] = [];
    let total = 0;

    while (true) {
      const result = await iterator.next();
      if (result.done) {
        await iterator.close();
        break;
      }

      const batch = JSON.parse(this.bytesToString(result.value.value)) as TeaBatch;
      if (batch.owner === owner) {
        total++;
        if (total > offsetNum && batches.length < limitNum) {
          batches.push(batch);
        }
      }
    }

    return { batches, total };
  }

  /**
   * Get batch history (all changes)
   */
  public async getBatchHistory(ctx: Context, batchId: string): Promise<TeaBatch[]> {
    validateBatchId(batchId);

    const historyIterator = await ctx.stub.getHistoryForKey(batchId);
    const history: TeaBatch[] = [];

    while (true) {
      const result = await historyIterator.next();
      if (result.done) {
        await historyIterator.close();
        break;
      }

      if (result.value.isDelete) {
        continue;
      }

      const batch = JSON.parse(this.bytesToString(result.value.value)) as TeaBatch;
      history.push(batch);
    }

    return history.reverse(); // Oldest first
  }

  public async updateBatchStatus(
    ctx: Context,
    batchId: string,
    status: string
  ): Promise<TeaBatch> {
    // Input validation
    validateBatchId(batchId);
    if (!status || status.trim().length === 0) {
      throw new Error("Status cannot be empty");
    }

    this.ensureOrg(ctx, [MSP_CONFIG.FARMER, MSP_CONFIG.ADMIN], "update batch status");

    const normalizedStatus = status.toUpperCase();
    if (!isTeaBatchStatus(normalizedStatus)) {
      throw new Error(
        `Invalid status '${status}'. Allowed values: CREATED, VERIFIED, EXPIRED.`
      );
    }

    const batch = await this.getBatchOrThrow(ctx, batchId);
    batch.status = normalizedStatus as TeaBatchStatus;
    batch.timestamp = this.getCurrentTimestamp(ctx);

    await ctx.stub.putState(batch.batchId, Buffer.from(JSON.stringify(batch)));
    return batch;
  }

  // ==================== Package Functions ====================

  /**
   * Create a new tea package from a batch
   * Args: [packageId, batchId, weight, productionDate, expiryDate?, qrCode?]
   */
  public async createPackage(
    ctx: Context,
    packageId: string,
    batchId: string,
    weightStr: string,
    productionDate: string,
    expiryDate?: string,
    qrCode?: string
  ): Promise<TeaPackage> {
    // Input validation
    validatePackageId(packageId);
    validateBatchId(batchId);
    const weight = parseFloat(weightStr);
    if (isNaN(weight)) {
      throw new Error("Weight must be a valid number");
    }
    validateWeight(weight);
    validateDate(productionDate);
    if (expiryDate) {
      validateDate(expiryDate);
      validateDateRange(productionDate, expiryDate);
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
    
    // Generate blockHash identifier from package data + txId
    // Backend can update with actual blockhash later
    const blockHash = this.generatePackageBlockHash(packageId, batchId, weight, productionDate, txId);

    const owner =
      ctx.clientIdentity.getAttributeValue("owner") ||
      ctx.clientIdentity.getAttributeValue("organization") ||
      ctx.clientIdentity.getMSPID();

    const pkg: TeaPackage = {
      packageId,
      batchId,
      blockHash,
      txId,
      weight,
      productionDate,
      expiryDate,
      qrCode,
      status: "CREATED",
      owner,
      timestamp: this.getCurrentTimestamp(ctx)
    };

    await ctx.stub.putState(packageId, Buffer.from(JSON.stringify(pkg)));
    return pkg;
  }

  /**
   * Verify a package by comparing blockhash
   * Args: [packageId, blockHash?]
   */
  public async verifyPackage(
    ctx: Context,
    packageId: string,
    providedBlockHash?: string
  ): Promise<{ isValid: boolean; package: TeaPackage }> {
    validatePackageId(packageId);

    // Public function - anyone can verify
    const pkg = await this.getPackageOrThrow(ctx, packageId);

    // If blockHash provided, compare
    if (providedBlockHash) {
      const isValid = pkg.blockHash === providedBlockHash;
      
      // If valid and status is CREATED, update to VERIFIED
      if (isValid && pkg.status === "CREATED") {
        pkg.status = "VERIFIED";
        pkg.timestamp = this.getCurrentTimestamp(ctx);
        await ctx.stub.putState(packageId, Buffer.from(JSON.stringify(pkg)));
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
  public async getPackageInfo(ctx: Context, packageId: string): Promise<TeaPackage | null> {
    validatePackageId(packageId);
    const buffer = await ctx.stub.getState(packageId);
    if (!buffer || buffer.length === 0) {
      return null;
    }
    return JSON.parse(this.bytesToString(buffer)) as TeaPackage;
  }

  /**
   * Get all packages with pagination
   * Args: [limit?, offset?]
   */
  public async getAllPackages(
    ctx: Context,
    ...args: string[]
  ): Promise<{ packages: TeaPackage[]; total: number }> {
    const limitStr = args[0] || "100";
    const offsetStr = args[1] || "0";
    const limitNum = parseInt(limitStr, 10);
    const offsetNum = parseInt(offsetStr, 10);
    validatePagination(limitNum, offsetNum);

    // Use composite key pattern: PACKAGE_{packageId}
    // Or iterate all and filter by prefix
    const iterator = await ctx.stub.getStateByRange("", "");
    const packages: TeaPackage[] = [];
    let total = 0;

    while (true) {
      const result = await iterator.next();
      if (result.done) {
        await iterator.close();
        break;
      }

      try {
        const pkg = JSON.parse(this.bytesToString(result.value.value)) as TeaPackage;
        // Check if it's a package (has packageId field)
        if (pkg.packageId) {
          total++;
          if (total > offsetNum && packages.length < limitNum) {
            packages.push(pkg);
          }
        }
      } catch (e) {
        // Skip non-package entries
        continue;
      }
    }

    return { packages, total };
  }

  /**
   * Get packages by batch ID
   * Args: [batchId, limit?, offset?]
   */
  public async getPackagesByBatch(
    ctx: Context,
    batchId: string,
    ...args: string[]
  ): Promise<{ packages: TeaPackage[]; total: number }> {
    validateBatchId(batchId);

    const limitStr = args[0] || "100";
    const offsetStr = args[1] || "0";
    const limitNum = parseInt(limitStr, 10);
    const offsetNum = parseInt(offsetStr, 10);
    validatePagination(limitNum, offsetNum);

    // Verify batch exists
    await this.getBatchOrThrow(ctx, batchId);

    const iterator = await ctx.stub.getStateByRange("", "");
    const packages: TeaPackage[] = [];
    let total = 0;

    while (true) {
      const result = await iterator.next();
      if (result.done) {
        await iterator.close();
        break;
      }

      try {
        const pkg = JSON.parse(this.bytesToString(result.value.value)) as TeaPackage;
        if (pkg.packageId && pkg.batchId === batchId) {
          total++;
          if (total > offsetNum && packages.length < limitNum) {
            packages.push(pkg);
          }
        }
      } catch (e) {
        continue;
      }
    }

    return { packages, total };
  }

  /**
   * Get packages by status
   * Args: [status, limit?, offset?]
   */
  public async getPackagesByStatus(
    ctx: Context,
    status: string,
    ...args: string[]
  ): Promise<{ packages: TeaPackage[]; total: number }> {
    const normalizedStatus = status.toUpperCase();
    if (!isTeaPackageStatus(normalizedStatus)) {
      throw new Error(
        `Invalid status '${status}'. Allowed values: CREATED, VERIFIED, SOLD, EXPIRED.`
      );
    }

    const limitStr = args[0] || "100";
    const offsetStr = args[1] || "0";
    const limitNum = parseInt(limitStr, 10);
    const offsetNum = parseInt(offsetStr, 10);
    validatePagination(limitNum, offsetNum);

    const iterator = await ctx.stub.getStateByRange("", "");
    const packages: TeaPackage[] = [];
    let total = 0;

    while (true) {
      const result = await iterator.next();
      if (result.done) {
        await iterator.close();
        break;
      }

      try {
        const pkg = JSON.parse(this.bytesToString(result.value.value)) as TeaPackage;
        if (pkg.packageId && pkg.status === normalizedStatus) {
          total++;
          if (total > offsetNum && packages.length < limitNum) {
            packages.push(pkg);
          }
        }
      } catch (e) {
        continue;
      }
    }

    return { packages, total };
  }

  /**
   * Get package history (all changes)
   * Args: [packageId]
   */
  public async getPackageHistory(ctx: Context, packageId: string): Promise<TeaPackage[]> {
    validatePackageId(packageId);

    const historyIterator = await ctx.stub.getHistoryForKey(packageId);
    const history: TeaPackage[] = [];

    while (true) {
      const result = await historyIterator.next();
      if (result.done) {
        await historyIterator.close();
        break;
      }

      if (result.value.isDelete) {
        continue;
      }

      const pkg = JSON.parse(this.bytesToString(result.value.value)) as TeaPackage;
      history.push(pkg);
    }

    return history.reverse(); // Oldest first
  }

  /**
   * Update package status
   * Args: [packageId, status]
   */
  public async updatePackageStatus(
    ctx: Context,
    packageId: string,
    status: string
  ): Promise<TeaPackage> {
    validatePackageId(packageId);
    if (!status || status.trim().length === 0) {
      throw new Error("Status cannot be empty");
    }

    this.ensureOrg(ctx, [MSP_CONFIG.FARMER, MSP_CONFIG.ADMIN], "update package status");

    const normalizedStatus = status.toUpperCase();
    if (!isTeaPackageStatus(normalizedStatus)) {
      throw new Error(
        `Invalid status '${status}'. Allowed values: CREATED, VERIFIED, SOLD, EXPIRED.`
      );
    }

    const pkg = await this.getPackageOrThrow(ctx, packageId);
    pkg.status = normalizedStatus as TeaPackageStatus;
    pkg.timestamp = this.getCurrentTimestamp(ctx);

    await ctx.stub.putState(packageId, Buffer.from(JSON.stringify(pkg)));
    return pkg;
  }

  private ensureOrg(ctx: Context, allowedMsps: string[], action: string): void {
    const callerMsp = ctx.clientIdentity.getMSPID();
    if (!allowedMsps.includes(callerMsp)) {
      throw new Error(
        `MSP '${callerMsp}' is not authorized to ${action}. Allowed MSPs: ${allowedMsps.join(", ")}`
      );
    }
  }

  private async getBatchOrThrow(ctx: Context, batchId: string): Promise<TeaBatch> {
    const buffer = await ctx.stub.getState(batchId);
    if (!buffer || buffer.length === 0) {
      throw new Error(`Batch with id '${batchId}' does not exist.`);
    }

    return JSON.parse(this.bytesToString(buffer)) as TeaBatch;
  }

  private async assertBatchDoesNotExist(ctx: Context, batchId: string): Promise<void> {
    const buffer = await ctx.stub.getState(batchId);
    if (buffer && buffer.length > 0) {
      throw new Error(`Batch with id '${batchId}' already exists.`);
    }
  }

  private async getPackageOrThrow(ctx: Context, packageId: string): Promise<TeaPackage> {
    const buffer = await ctx.stub.getState(packageId);
    if (!buffer || buffer.length === 0) {
      throw new Error(`Package with id '${packageId}' does not exist.`);
    }

    return JSON.parse(this.bytesToString(buffer)) as TeaPackage;
  }

  private async assertPackageDoesNotExist(ctx: Context, packageId: string): Promise<void> {
    const buffer = await ctx.stub.getState(packageId);
    if (buffer && buffer.length > 0) {
      throw new Error(`Package with id '${packageId}' already exists.`);
    }
  }

  /**
   * Generate a unique blockHash identifier for package
   * This is a composite hash from package data + transaction ID
   * Backend can update with actual blockhash from blockchain later
   */
  private generatePackageBlockHash(
    packageId: string,
    batchId: string,
    weight: number,
    productionDate: string,
    txId: string
  ): string {
    const payload = `${packageId}|${batchId}|${weight}|${productionDate}|${txId}`;
    return crypto.createHash("sha256").update(payload).digest("hex");
  }

  private getCurrentTimestamp(ctx: Context): string {
    const timestamp = ctx.stub.getTxTimestamp();
    const millis = timestamp.seconds.toNumber() * 1000 + Math.floor(timestamp.nanos / 1_000_000);
    return new Date(millis).toISOString();
  }

  private bytesToString(bytes: Uint8Array): string {
    return Buffer.from(bytes).toString("utf8");
  }
}

export const contracts = [TeaTraceContract];


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
    
    // Generate blockHash identifier from package data + txId
    // Backend can update with actual blockhash later
    const blockHash = this.generatePackageBlockHash(packageId, batchId, weight, productionDate, txId);

    const owner =
      ctx.clientIdentity.getAttributeValue("owner") ||
      ctx.clientIdentity.getAttributeValue("organization") ||
      ctx.clientIdentity.getMSPID();

    const pkg: TeaPackage = {
      packageId,
      batchId,
      blockHash,
      txId,
      weight,
      productionDate,
      expiryDate,
      qrCode,
      status: "CREATED",
      owner,
      timestamp: this.getCurrentTimestamp(ctx)
    };

    await ctx.stub.putState(packageId, Buffer.from(JSON.stringify(pkg)));
    return pkg;
  }

  /**
   * Verify a package by comparing blockhash
   * Args: [packageId, blockHash?]
   */
  public async verifyPackage(
    ctx: Context,
    packageId: string,
    providedBlockHash?: string
  ): Promise<{ isValid: boolean; package: TeaPackage }> {
    validatePackageId(packageId);

    // Public function - anyone can verify
    const pkg = await this.getPackageOrThrow(ctx, packageId);

    // If blockHash provided, compare
    if (providedBlockHash) {
      const isValid = pkg.blockHash === providedBlockHash;
      
      // If valid and status is CREATED, update to VERIFIED
      if (isValid && pkg.status === "CREATED") {
        pkg.status = "VERIFIED";
        pkg.timestamp = this.getCurrentTimestamp(ctx);
        await ctx.stub.putState(packageId, Buffer.from(JSON.stringify(pkg)));
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
  public async getPackageInfo(ctx: Context, packageId: string): Promise<TeaPackage | null> {
    validatePackageId(packageId);
    const buffer = await ctx.stub.getState(packageId);
    if (!buffer || buffer.length === 0) {
      return null;
    }
    return JSON.parse(this.bytesToString(buffer)) as TeaPackage;
  }

  /**
   * Get all packages with pagination
   * Args: [limit?, offset?]
   */
  public async getAllPackages(
    ctx: Context,
    ...args: string[]
  ): Promise<{ packages: TeaPackage[]; total: number }> {
    const limitStr = args[0] || "100";
    const offsetStr = args[1] || "0";
    const limitNum = parseInt(limitStr, 10);
    const offsetNum = parseInt(offsetStr, 10);
    validatePagination(limitNum, offsetNum);

    // Use composite key pattern: PACKAGE_{packageId}
    // Or iterate all and filter by prefix
    const iterator = await ctx.stub.getStateByRange("", "");
    const packages: TeaPackage[] = [];
    let total = 0;

    while (true) {
      const result = await iterator.next();
      if (result.done) {
        await iterator.close();
        break;
      }

      try {
        const pkg = JSON.parse(this.bytesToString(result.value.value)) as TeaPackage;
        // Check if it's a package (has packageId field)
        if (pkg.packageId) {
          total++;
          if (total > offsetNum && packages.length < limitNum) {
            packages.push(pkg);
          }
        }
      } catch (e) {
        // Skip non-package entries
        continue;
      }
    }

    return { packages, total };
  }

  /**
   * Get packages by batch ID
   * Args: [batchId, limit?, offset?]
   */
  public async getPackagesByBatch(
    ctx: Context,
    batchId: string,
    ...args: string[]
  ): Promise<{ packages: TeaPackage[]; total: number }> {
    validateBatchId(batchId);

    const limitStr = args[0] || "100";
    const offsetStr = args[1] || "0";
    const limitNum = parseInt(limitStr, 10);
    const offsetNum = parseInt(offsetStr, 10);
    validatePagination(limitNum, offsetNum);

    // Verify batch exists
    await this.getBatchOrThrow(ctx, batchId);

    const iterator = await ctx.stub.getStateByRange("", "");
    const packages: TeaPackage[] = [];
    let total = 0;

    while (true) {
      const result = await iterator.next();
      if (result.done) {
        await iterator.close();
        break;
      }

      try {
        const pkg = JSON.parse(this.bytesToString(result.value.value)) as TeaPackage;
        if (pkg.packageId && pkg.batchId === batchId) {
          total++;
          if (total > offsetNum && packages.length < limitNum) {
            packages.push(pkg);
          }
        }
      } catch (e) {
        continue;
      }
    }

    return { packages, total };
  }

  /**
   * Get packages by status
   * Args: [status, limit?, offset?]
   */
  public async getPackagesByStatus(
    ctx: Context,
    status: string,
    ...args: string[]
  ): Promise<{ packages: TeaPackage[]; total: number }> {
    const normalizedStatus = status.toUpperCase();
    if (!isTeaPackageStatus(normalizedStatus)) {
      throw new Error(
        `Invalid status '${status}'. Allowed values: CREATED, VERIFIED, SOLD, EXPIRED.`
      );
    }

    const limitStr = args[0] || "100";
    const offsetStr = args[1] || "0";
    const limitNum = parseInt(limitStr, 10);
    const offsetNum = parseInt(offsetStr, 10);
    validatePagination(limitNum, offsetNum);

    const iterator = await ctx.stub.getStateByRange("", "");
    const packages: TeaPackage[] = [];
    let total = 0;

    while (true) {
      const result = await iterator.next();
      if (result.done) {
        await iterator.close();
        break;
      }

      try {
        const pkg = JSON.parse(this.bytesToString(result.value.value)) as TeaPackage;
        if (pkg.packageId && pkg.status === normalizedStatus) {
          total++;
          if (total > offsetNum && packages.length < limitNum) {
            packages.push(pkg);
          }
        }
      } catch (e) {
        continue;
      }
    }

    return { packages, total };
  }

  /**
   * Get package history (all changes)
   * Args: [packageId]
   */
  public async getPackageHistory(ctx: Context, packageId: string): Promise<TeaPackage[]> {
    validatePackageId(packageId);

    const historyIterator = await ctx.stub.getHistoryForKey(packageId);
    const history: TeaPackage[] = [];

    while (true) {
      const result = await historyIterator.next();
      if (result.done) {
        await historyIterator.close();
        break;
      }

      if (result.value.isDelete) {
        continue;
      }

      const pkg = JSON.parse(this.bytesToString(result.value.value)) as TeaPackage;
      history.push(pkg);
    }

    return history.reverse(); // Oldest first
  }

  /**
   * Update package status
   * Args: [packageId, status]
   */
  public async updatePackageStatus(
    ctx: Context,
    packageId: string,
    status: string
  ): Promise<TeaPackage> {
    validatePackageId(packageId);
    if (!status || status.trim().length === 0) {
      throw new Error("Status cannot be empty");
    }

    this.ensureOrg(ctx, [MSP_CONFIG.FARMER, MSP_CONFIG.ADMIN], "update package status");

    const normalizedStatus = status.toUpperCase();
    if (!isTeaPackageStatus(normalizedStatus)) {
      throw new Error(
        `Invalid status '${status}'. Allowed values: CREATED, VERIFIED, SOLD, EXPIRED.`
      );
    }

    const pkg = await this.getPackageOrThrow(ctx, packageId);
    pkg.status = normalizedStatus as TeaPackageStatus;
    pkg.timestamp = this.getCurrentTimestamp(ctx);

    await ctx.stub.putState(packageId, Buffer.from(JSON.stringify(pkg)));
    return pkg;
  }

  private ensureOrg(ctx: Context, allowedMsps: string[], action: string): void {
    const callerMsp = ctx.clientIdentity.getMSPID();
    if (!allowedMsps.includes(callerMsp)) {
      throw new Error(
        `MSP '${callerMsp}' is not authorized to ${action}. Allowed MSPs: ${allowedMsps.join(", ")}`
      );
    }
  }

  private async getBatchOrThrow(ctx: Context, batchId: string): Promise<TeaBatch> {
    const buffer = await ctx.stub.getState(batchId);
    if (!buffer || buffer.length === 0) {
      throw new Error(`Batch with id '${batchId}' does not exist.`);
    }

    return JSON.parse(this.bytesToString(buffer)) as TeaBatch;
  }

  private async assertBatchDoesNotExist(ctx: Context, batchId: string): Promise<void> {
    const buffer = await ctx.stub.getState(batchId);
    if (buffer && buffer.length > 0) {
      throw new Error(`Batch with id '${batchId}' already exists.`);
    }
  }

  private async getPackageOrThrow(ctx: Context, packageId: string): Promise<TeaPackage> {
    const buffer = await ctx.stub.getState(packageId);
    if (!buffer || buffer.length === 0) {
      throw new Error(`Package with id '${packageId}' does not exist.`);
    }

    return JSON.parse(this.bytesToString(buffer)) as TeaPackage;
  }

  private async assertPackageDoesNotExist(ctx: Context, packageId: string): Promise<void> {
    const buffer = await ctx.stub.getState(packageId);
    if (buffer && buffer.length > 0) {
      throw new Error(`Package with id '${packageId}' already exists.`);
    }
  }

  /**
   * Generate a unique blockHash identifier for package
   * This is a composite hash from package data + transaction ID
   * Backend can update with actual blockhash from blockchain later
   */
  private generatePackageBlockHash(
    packageId: string,
    batchId: string,
    weight: number,
    productionDate: string,
    txId: string
  ): string {
    const payload = `${packageId}|${batchId}|${weight}|${productionDate}|${txId}`;
    return crypto.createHash("sha256").update(payload).digest("hex");
  }

  private getCurrentTimestamp(ctx: Context): string {
    const timestamp = ctx.stub.getTxTimestamp();
    const millis = timestamp.seconds.toNumber() * 1000 + Math.floor(timestamp.nanos / 1_000_000);
    return new Date(millis).toISOString();
  }

  private bytesToString(bytes: Uint8Array): string {
    return Buffer.from(bytes).toString("utf8");
  }
}

export const contracts = [TeaTraceContract];

